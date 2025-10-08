package dot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/scanner"
)

// AdoptService handles file adoption operations.
type AdoptService struct {
	fs          FS
	logger      Logger
	executor    *executor.Executor
	manifestSvc *ManifestService
	packageDir  string
	targetDir   string
	dryRun      bool
}

// newAdoptService creates a new adopt service.
func newAdoptService(
	fs FS,
	logger Logger,
	exec *executor.Executor,
	manifestSvc *ManifestService,
	packageDir string,
	targetDir string,
	dryRun bool,
) *AdoptService {
	return &AdoptService{
		fs:          fs,
		logger:      logger,
		executor:    exec,
		manifestSvc: manifestSvc,
		packageDir:  packageDir,
		targetDir:   targetDir,
		dryRun:      dryRun,
	}
}

// Adopt moves existing files from target into package then creates symlinks.
func (s *AdoptService) Adopt(ctx context.Context, files []string, pkg string) error {
	plan, err := s.PlanAdopt(ctx, files, pkg)
	if err != nil {
		return err
	}
	if len(plan.Operations) == 0 {
		s.logger.Info(ctx, "nothing_to_adopt")
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
	if err := s.manifestSvc.Update(ctx, targetPathResult.Unwrap(), s.packageDir, []string{pkg}, plan); err != nil {
		s.logger.Warn(ctx, "failed_to_update_manifest", "error", err)
	}
	return nil
}

// PlanAdopt computes the execution plan for adopting files.
func (s *AdoptService) PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error) {
	packagePathResult := NewPackagePath(s.packageDir)
	if !packagePathResult.IsOk() {
		return Plan{}, packagePathResult.UnwrapErr()
	}
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}

	// Verify package directory exists
	pkgPath := filepath.Join(s.packageDir, pkg)
	exists := s.fs.Exists(ctx, pkgPath)
	if !exists {
		return Plan{}, ErrPackageNotFound{Package: pkg}
	}

	operations := make([]Operation, 0, len(files)*2)
	for _, file := range files {
		sourceFile := filepath.Join(s.targetDir, file)
		if !s.fs.Exists(ctx, sourceFile) {
			return Plan{}, ErrSourceNotFound{Path: sourceFile}
		}

		adoptedName := scanner.UntranslateDotfile(filepath.Base(file))
		destFile := filepath.Join(pkgPath, adoptedName)

		sourceLinkPathResult := NewTargetPath(sourceFile)
		if !sourceLinkPathResult.IsOk() {
			return Plan{}, sourceLinkPathResult.UnwrapErr()
		}
		destPathResult := NewFilePath(destFile)
		if !destPathResult.IsOk() {
			return Plan{}, destPathResult.UnwrapErr()
		}

		moveID := OperationID(fmt.Sprintf("adopt-move-%s", file))
		operations = append(operations, FileMove{
			OpID:   moveID,
			Source: sourceLinkPathResult.Unwrap(),
			Dest:   destPathResult.Unwrap(),
		})

		linkID := OperationID(fmt.Sprintf("adopt-link-%s", file))
		operations = append(operations, NewLinkCreate(linkID, destPathResult.Unwrap(), sourceLinkPathResult.Unwrap()))
	}
	return Plan{
		Operations: operations,
		Metadata: PlanMetadata{
			PackageCount:   1,
			OperationCount: len(operations),
		},
	}, nil
}
