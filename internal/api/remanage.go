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
		_, exists := m.GetPackage(pkg)
		if !exists {
			// Package not installed - plan as new install
			pkgPlan, err := c.PlanManage(ctx, pkg)
			if err != nil {
				return dot.Plan{}, err
			}
			allOperations = append(allOperations, pkgPlan.Operations...)
			// Merge package operations
			if pkgPlan.PackageOperations != nil {
				if pkgOps, hasPkg := pkgPlan.PackageOperations[pkg]; hasPkg {
					packageOps[pkg] = pkgOps
				}
			}
			continue
		}

		// Compute current hash
		pkgPathStr := filepath.Join(c.config.PackageDir, pkg)
		pkgPathResult := dot.NewPackagePath(pkgPathStr)
		if !pkgPathResult.IsOk() {
			return dot.Plan{}, pkgPathResult.UnwrapErr()
		}
		pkgPath := pkgPathResult.Unwrap()

		currentHash, err := hasher.HashPackage(ctx, pkgPath)
		if err != nil {
			// Can't compute hash - fall back to full remanage for this package
			c.config.Logger.Warn(ctx, "hash_computation_failed", "package", pkg, "error", err)
			unmanagePlan, _ := c.PlanUnmanage(ctx, pkg)
			managePlan, err := c.PlanManage(ctx, pkg)
			if err != nil {
				return dot.Plan{}, err
			}
			allOperations = append(allOperations, unmanagePlan.Operations...)
			allOperations = append(allOperations, managePlan.Operations...)
			continue
		}

		// Compare with stored hash
		storedHash, hasHash := m.GetHash(pkg)
		if !hasHash || storedHash != currentHash {
			// Package changed - do full unmanage + manage
			unmanagePlan, _ := c.PlanUnmanage(ctx, pkg)
			managePlan, err := c.PlanManage(ctx, pkg)
			if err != nil {
				return dot.Plan{}, err
			}
			allOperations = append(allOperations, unmanagePlan.Operations...)
			allOperations = append(allOperations, managePlan.Operations...)

			// Merge package operations
			if managePlan.PackageOperations != nil {
				if pkgOps, hasPkg := managePlan.PackageOperations[pkg]; hasPkg {
					packageOps[pkg] = pkgOps
				}
			}
		} else {
			// Package unchanged - no operations needed
			c.config.Logger.Info(ctx, "package_unchanged", "package", pkg)
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
