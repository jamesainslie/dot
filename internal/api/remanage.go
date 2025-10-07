package api

import (
	"context"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Remanage reinstalls packages using incremental hash-based change detection.
// Only processes files that have changed since last installation.
func (c *client) Remanage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanRemanage(ctx, packages...)
	if err != nil {
		return err
	}

	// Check if no changes detected
	if len(plan.Operations) == 0 {
		c.config.Logger.Info(ctx, "no_changes_detected", "packages", packages)
		return nil
	}

	if c.config.DryRun {
		c.config.Logger.Info(ctx, "dry_run_plan", "operations", len(plan.Operations))
		return nil
	}

	// Execute the plan
	result := c.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}

	execResult := result.Unwrap()
	if !execResult.Success() {
		return dot.ErrMultiple{Errors: execResult.Errors}
	}

	// Update manifest with new hashes
	if err := c.updateManifest(ctx, packages, plan); err != nil {
		c.config.Logger.Warn(ctx, "manifest_update_failed", "error", err)
	}

	return nil
}

// PlanRemanage computes incremental execution plan using hash-based change detection.
// Only includes operations for files that have changed since last installation.
func (c *client) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest - fall back to full manage
		return c.PlanManage(ctx, packages...)
	}

	m := manifestResult.Unwrap()
	hasher := manifest.NewContentHasher(c.config.FS)

	allOperations := make([]dot.Operation, 0)
	packageOps := make(map[string][]dot.OperationID)

	for _, pkg := range packages {
		ops, pkgOpsMap, err := c.planSinglePackageRemanage(ctx, pkg, m, hasher)
		if err != nil {
			return dot.Plan{}, err
		}
		allOperations = append(allOperations, ops...)
		for k, v := range pkgOpsMap {
			packageOps[k] = v
		}
	}

	return dot.Plan{
		Operations: allOperations,
		Metadata: dot.PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(allOperations),
		},
		PackageOperations: packageOps,
	}, nil
}

// planSinglePackageRemanage plans remanage for a single package using hash comparison.
func (c *client) planSinglePackageRemanage(
	ctx context.Context,
	pkg string,
	m manifest.Manifest,
	hasher *manifest.ContentHasher,
) ([]dot.Operation, map[string][]dot.OperationID, error) {
	_, exists := m.GetPackage(pkg)
	if !exists {
		// Package not installed - plan as new install
		return c.planNewPackageInstall(ctx, pkg)
	}

	// Compute current hash
	pkgPath, err := c.getPackagePath(pkg)
	if err != nil {
		return nil, nil, err
	}

	currentHash, err := hasher.HashPackage(ctx, pkgPath)
	if err != nil {
		// Can't compute hash - fall back to full remanage
		c.config.Logger.Warn(ctx, "hash_computation_failed", "package", pkg, "error", err)
		return c.planFullRemanage(ctx, pkg)
	}

	// Compare with stored hash
	storedHash, hasHash := m.GetHash(pkg)
	if !hasHash || storedHash != currentHash {
		// Package changed - do full unmanage + manage
		return c.planFullRemanage(ctx, pkg)
	}

	// Package unchanged - no operations needed
	c.config.Logger.Info(ctx, "package_unchanged", "package", pkg)
	return []dot.Operation{}, map[string][]dot.OperationID{}, nil
}

// planNewPackageInstall plans installation of a package not yet in manifest.
func (c *client) planNewPackageInstall(ctx context.Context, pkg string) ([]dot.Operation, map[string][]dot.OperationID, error) {
	pkgPlan, err := c.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	packageOps := make(map[string][]dot.OperationID)
	if pkgPlan.PackageOperations != nil {
		if pkgOps, hasPkg := pkgPlan.PackageOperations[pkg]; hasPkg {
			packageOps[pkg] = pkgOps
		}
	}

	return pkgPlan.Operations, packageOps, nil
}

// planFullRemanage plans full unmanage + manage for a package.
func (c *client) planFullRemanage(ctx context.Context, pkg string) ([]dot.Operation, map[string][]dot.OperationID, error) {
	unmanagePlan, _ := c.PlanUnmanage(ctx, pkg)
	managePlan, err := c.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	allOps := make([]dot.Operation, 0, len(unmanagePlan.Operations)+len(managePlan.Operations))
	allOps = append(allOps, unmanagePlan.Operations...)
	allOps = append(allOps, managePlan.Operations...)

	packageOps := make(map[string][]dot.OperationID)
	if managePlan.PackageOperations != nil {
		if pkgOps, hasPkg := managePlan.PackageOperations[pkg]; hasPkg {
			packageOps[pkg] = pkgOps
		}
	}

	return allOps, packageOps, nil
}

// getPackagePath constructs and validates package path.
func (c *client) getPackagePath(pkg string) (dot.PackagePath, error) {
	pkgPathStr := filepath.Join(c.config.PackageDir, pkg)
	pkgPathResult := dot.NewPackagePath(pkgPathStr)
	if !pkgPathResult.IsOk() {
		return dot.PackagePath{}, pkgPathResult.UnwrapErr()
	}
	return pkgPathResult.Unwrap(), nil
}
