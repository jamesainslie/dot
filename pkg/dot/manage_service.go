package dot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
)

// ManageService handles package installation (manage and remanage operations).
type ManageService struct {
	fs          FS
	logger      Logger
	managePipe  *pipeline.ManagePipeline
	executor    *executor.Executor
	manifestSvc *ManifestService
	unmanageSvc *UnmanageService
	packageDir  string
	targetDir   string
	dryRun      bool
}

// newManageService creates a new manage service.
func newManageService(
	fs FS,
	logger Logger,
	managePipe *pipeline.ManagePipeline,
	exec *executor.Executor,
	manifestSvc *ManifestService,
	unmanageSvc *UnmanageService,
	packageDir string,
	targetDir string,
	dryRun bool,
) *ManageService {
	return &ManageService{
		fs:          fs,
		logger:      logger,
		managePipe:  managePipe,
		executor:    exec,
		manifestSvc: manifestSvc,
		unmanageSvc: unmanageSvc,
		packageDir:  packageDir,
		targetDir:   targetDir,
		dryRun:      dryRun,
	}
}

// Manage installs the specified packages by creating symlinks.
func (s *ManageService) Manage(ctx context.Context, packages ...string) error {
	plan, err := s.PlanManage(ctx, packages...)
	if err != nil {
		return err
	}
	if s.dryRun {
		return nil
	}
	result := s.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}
	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}
	// Update manifest
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	if err := s.manifestSvc.Update(ctx, targetPathResult.Unwrap(), s.packageDir, packages, plan); err != nil {
		s.logger.Warn(ctx, "manifest_update_failed", "error", err)
	}
	return nil
}

// PlanManage computes the execution plan for managing packages without applying changes.
func (s *ManageService) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
	packagePathResult := NewPackagePath(s.packageDir)
	if !packagePathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid package directory: %w", packagePathResult.UnwrapErr())
	}
	packagePath := packagePathResult.Unwrap()

	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()

	input := pipeline.ManageInput{
		PackageDir: packagePath,
		TargetDir:  targetPath,
		Packages:   packages,
	}
	planResult := s.managePipe.Execute(ctx, input)
	if !planResult.IsOk() {
		return Plan{}, planResult.UnwrapErr()
	}
	return planResult.Unwrap(), nil
}

// Remanage reinstalls packages using incremental hash-based change detection.
func (s *ManageService) Remanage(ctx context.Context, packages ...string) error {
	plan, err := s.PlanRemanage(ctx, packages...)
	if err != nil {
		return err
	}
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "no_changes_detected", "packages", packages)
		return nil
	}
	if s.dryRun {
		s.logger.Info(ctx, "dry_run_plan", "operations", len(plan.Operations))
		return nil
	}
	result := s.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}
	execResult := result.Unwrap()
	if !execResult.Success() {
		return ErrMultiple{Errors: execResult.Errors}
	}
	// Update manifest
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	if err := s.manifestSvc.Update(ctx, targetPathResult.Unwrap(), s.packageDir, packages, plan); err != nil {
		s.logger.Warn(ctx, "manifest_update_failed", "error", err)
	}
	return nil
}

// PlanRemanage computes incremental execution plan using hash-based change detection.
func (s *ManageService) PlanRemanage(ctx context.Context, packages ...string) (Plan, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest - fall back to full manage
		return s.PlanManage(ctx, packages...)
	}

	m := manifestResult.Unwrap()
	hasher := manifest.NewContentHasher(s.fs)
	allOperations := make([]Operation, 0)
	packageOps := make(map[string][]OperationID)

	for _, pkg := range packages {
		ops, pkgOpsMap, err := s.planSinglePackageRemanage(ctx, pkg, &m, hasher)
		if err != nil {
			return Plan{}, err
		}
		allOperations = append(allOperations, ops...)
		for k, v := range pkgOpsMap {
			packageOps[k] = v
		}
	}

	return Plan{
		Operations: allOperations,
		Metadata: PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(allOperations),
		},
		PackageOperations: packageOps,
	}, nil
}

// planSinglePackageRemanage plans remanage for a single package using hash comparison.
func (s *ManageService) planSinglePackageRemanage(
	ctx context.Context,
	pkg string,
	m *manifest.Manifest,
	hasher *manifest.ContentHasher,
) ([]Operation, map[string][]OperationID, error) {
	_, exists := m.GetPackage(pkg)
	if !exists {
		return s.planNewPackageInstall(ctx, pkg)
	}

	pkgPath, err := s.getPackagePath(pkg)
	if err != nil {
		return nil, nil, err
	}
	currentHash, err := hasher.HashPackage(ctx, pkgPath)
	if err != nil {
		s.logger.Warn(ctx, "hash_computation_failed", "package", pkg, "error", err)
		return s.planFullRemanage(ctx, pkg)
	}

	storedHash, hasHash := m.GetHash(pkg)
	if !hasHash || storedHash != currentHash {
		return s.planFullRemanage(ctx, pkg)
	}

	s.logger.Info(ctx, "package_unchanged", "package", pkg)
	return []Operation{}, map[string][]OperationID{}, nil
}

// planNewPackageInstall plans installation of a package not yet in manifest.
func (s *ManageService) planNewPackageInstall(ctx context.Context, pkg string) ([]Operation, map[string][]OperationID, error) {
	pkgPlan, err := s.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	packageOps := make(map[string][]OperationID)
	if pkgPlan.PackageOperations != nil {
		if pkgOps, hasPkg := pkgPlan.PackageOperations[pkg]; hasPkg {
			packageOps[pkg] = pkgOps
		}
	}
	return pkgPlan.Operations, packageOps, nil
}

// planFullRemanage plans full unmanage + manage for a package.
func (s *ManageService) planFullRemanage(ctx context.Context, pkg string) ([]Operation, map[string][]OperationID, error) {
	// Get unmanage operations first
	unmanagePlan, err := s.unmanageSvc.PlanUnmanage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	// Get manage operations
	managePlan, err := s.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	// Concatenate operations (unmanage first, then manage)
	ops := make([]Operation, 0, len(unmanagePlan.Operations)+len(managePlan.Operations))
	ops = append(ops, unmanagePlan.Operations...)
	ops = append(ops, managePlan.Operations...)

	// Merge package operations
	packageOps := make(map[string][]OperationID)
	unmanageOps := unmanagePlan.PackageOperations[pkg]
	manageOps := managePlan.PackageOperations[pkg]
	mergedOps := make([]OperationID, 0, len(unmanageOps)+len(manageOps))
	mergedOps = append(mergedOps, unmanageOps...)
	mergedOps = append(mergedOps, manageOps...)
	packageOps[pkg] = mergedOps

	return ops, packageOps, nil
}

// getPackagePath constructs and validates package path.
func (s *ManageService) getPackagePath(pkg string) (PackagePath, error) {
	pkgPathStr := filepath.Join(s.packageDir, pkg)
	pkgPathResult := NewPackagePath(pkgPathStr)
	if !pkgPathResult.IsOk() {
		return PackagePath{}, pkgPathResult.UnwrapErr()
	}
	return pkgPathResult.Unwrap(), nil
}
