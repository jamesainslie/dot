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
	plan, err := s.PlanUnmanage(ctx, packages...)
	if err != nil {
		return err
	}
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "nothing_to_unmanage", "packages", packages)
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
	// Update manifest to remove packages
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	for _, pkg := range packages {
		if err := s.manifestSvc.RemovePackage(ctx, targetPath, pkg); err != nil {
			s.logger.Warn(ctx, "failed_to_update_manifest", "error", err)
			return err
		}
	}
	return nil
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (s *UnmanageService) PlanUnmanage(ctx context.Context, packages ...string) (Plan, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		// Check if this is a "file not found" error
		if isManifestNotFoundError(err) {
			return Plan{
				Operations: []Operation{},
				Metadata:   PlanMetadata{},
			}, nil
		}
		return Plan{}, err
	}

	m := manifestResult.Unwrap()

	// Build delete operations for all links in specified packages
	var operations []Operation
	for _, pkg := range packages {
		pkgInfo, exists := m.GetPackage(pkg)
		if !exists {
			s.logger.Warn(ctx, "package_not_installed", "package", pkg)
			continue
		}
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
	return Plan{
		Operations: operations,
		Metadata: PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(operations),
		},
	}, nil
}
