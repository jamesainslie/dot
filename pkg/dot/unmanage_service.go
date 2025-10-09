package dot

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/internal/executor"
)

// UnmanageService handles package removal (unmanage operations).
type UnmanageService struct {
	fs          FS
	logger      Logger
	executor    *executor.Executor
	manifestSvc *ManifestService
	targetDir   string
	dryRun      bool
}

// newUnmanageService creates a new unmanage service.
func newUnmanageService(
	fs FS,
	logger Logger,
	exec *executor.Executor,
	manifestSvc *ManifestService,
	targetDir string,
	dryRun bool,
) *UnmanageService {
	return &UnmanageService{
		fs:          fs,
		logger:      logger,
		executor:    exec,
		manifestSvc: manifestSvc,
		targetDir:   targetDir,
		dryRun:      dryRun,
	}
}

// Unmanage removes the specified packages by deleting symlinks.
func (s *UnmanageService) Unmanage(ctx context.Context, packages ...string) error {
	s.logger.Info(ctx, "unmanaging_packages", "count", len(packages), "packages", packages)

	s.logger.Debug(ctx, "planning_unmanage", "packages", packages)
	plan, err := s.PlanUnmanage(ctx, packages...)
	if err != nil {
		s.logger.Error(ctx, "plan_failed", "error", err)
		return err
	}
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "nothing_to_unmanage", "packages", packages)
		return nil
	}

	s.logger.Info(ctx, "plan_created", "operations", len(plan.Operations))
	s.logger.Debug(ctx, "plan_details", "link_deletions", len(plan.Operations))

	if s.dryRun {
		s.logger.Info(ctx, "dry_run_plan", "operations", len(plan.Operations))
		return nil
	}

	s.logger.Debug(ctx, "executing_plan", "operation_count", len(plan.Operations))
	result := s.executor.Execute(ctx, plan)
	if !result.IsOk() {
		s.logger.Error(ctx, "execution_error", "error", result.UnwrapErr())
		return result.UnwrapErr()
	}
	execResult := result.Unwrap()
	if !execResult.Success() {
		s.logger.Error(ctx, "execution_failed", "failed_count", len(execResult.Failed))
		return ErrMultiple{Errors: execResult.Errors}
	}

	s.logger.Info(ctx, "execution_successful", "operations", len(execResult.Executed))

	// Update manifest to remove packages
	s.logger.Debug(ctx, "removing_packages_from_manifest", "packages", packages)
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	for _, pkg := range packages {
		if err := s.manifestSvc.RemovePackage(ctx, targetPath, pkg); err != nil {
			s.logger.Warn(ctx, "failed_to_update_manifest", "package", pkg, "error", err)
			return err
		}
		s.logger.Debug(ctx, "package_removed_from_manifest", "package", pkg)
	}

	s.logger.Debug(ctx, "manifest_updated")
	return nil
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (s *UnmanageService) PlanUnmanage(ctx context.Context, packages ...string) (Plan, error) {
	s.logger.Debug(ctx, "plan_unmanage_started", "packages", packages)

	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	s.logger.Debug(ctx, "loading_manifest")
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		// Check if this is a "file not found" error
		if isManifestNotFoundError(err) {
			s.logger.Debug(ctx, "no_manifest_found_nothing_to_unmanage")
			return Plan{
				Operations: []Operation{},
				Metadata:   PlanMetadata{},
			}, nil
		}
		return Plan{}, err
	}

	m := manifestResult.Unwrap()
	s.logger.Debug(ctx, "manifest_loaded", "installed_packages", len(m.Packages))

	// Build delete operations for all links in specified packages
	var operations []Operation
	for _, pkg := range packages {
		pkgInfo, exists := m.GetPackage(pkg)
		if !exists {
			s.logger.Warn(ctx, "package_not_installed", "package", pkg)
			continue
		}

		s.logger.Debug(ctx, "creating_delete_operations", "package", pkg, "links", len(pkgInfo.Links))

		// Create delete operations for each link
		for _, link := range pkgInfo.Links {
			targetFilePath := s.targetDir + "/" + link
			targetPathResult := NewTargetPath(targetFilePath)
			if !targetPathResult.IsOk() {
				continue
			}
			id := OperationID(fmt.Sprintf("unmanage-link-%s", link))
			operations = append(operations, NewLinkDelete(id, targetPathResult.Unwrap()))
		}
	}

	s.logger.Debug(ctx, "plan_unmanage_completed", "operations", len(operations))

	return Plan{
		Operations: operations,
		Metadata: PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(operations),
		},
	}, nil
}
