package dot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/manifest"
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
	// Update manifest with source=adopted
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	if err := s.manifestSvc.UpdateWithSource(ctx, targetPathResult.Unwrap(), s.packageDir, []string{pkg}, plan, manifest.SourceAdopted); err != nil {
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

	// Check if package directory exists, create if not
	pkgPath := filepath.Join(s.packageDir, pkg)
	operations := make([]Operation, 0, len(files)*2+1)

	if !s.fs.Exists(ctx, pkgPath) {
		// Add operation to create package directory
		pkgPathResult := NewFilePath(pkgPath)
		if pkgPathResult.IsErr() {
			return Plan{}, fmt.Errorf("invalid package path %s: %w", pkgPath, pkgPathResult.UnwrapErr())
		}
		dirID := OperationID(fmt.Sprintf("adopt-create-pkg-%s", pkg))
		operations = append(operations, NewDirCreate(dirID, pkgPathResult.Unwrap()))
	}

	for _, file := range files {
		sourceFile := filepath.Join(s.targetDir, file)
		if !s.fs.Exists(ctx, sourceFile) {
			return Plan{}, ErrSourceNotFound{Path: sourceFile}
		}

		// Check if source is a directory
		isDir, err := s.fs.IsDir(ctx, sourceFile)
		if err != nil {
			return Plan{}, fmt.Errorf("failed to check if directory: %w", err)
		}

		if isDir {
			// For directories: move CONTENTS into package root (flat structure)
			dirOps, err := s.createDirectoryAdoptOperations(ctx, sourceFile, pkgPath, file)
			if err != nil {
				return Plan{}, err
			}
			operations = append(operations, dirOps...)
		} else {
			// For files: move file into package directory with translation
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
	}

	// Build PackageOperations map for manifest tracking
	packageOps := make(map[string][]OperationID)
	opIDs := make([]OperationID, 0, len(operations))
	for _, op := range operations {
		opIDs = append(opIDs, op.ID())
	}
	packageOps[pkg] = opIDs

	return Plan{
		Operations:        operations,
		PackageOperations: packageOps,
		Metadata: PlanMetadata{
			PackageCount:   1,
			OperationCount: len(operations),
		},
	}, nil
}

// createDirectoryAdoptOperations creates operations to adopt a directory's contents.
// Moves directory CONTENTS into package root (flat structure), not the directory itself.
func (s *AdoptService) createDirectoryAdoptOperations(ctx context.Context, sourceDir, pkgPath, originalPath string) ([]Operation, error) {
	var operations []Operation

	// Recursively collect all files in the directory
	filesToMove, err := s.collectDirectoryFiles(ctx, sourceDir, "")
	if err != nil {
		return nil, fmt.Errorf("failed to collect directory files: %w", err)
	}

	// First pass: Create all directories
	for _, relPath := range filesToMove {
		sourcePath := filepath.Join(sourceDir, relPath)

		isDir, _ := s.fs.IsDir(ctx, sourcePath)
		if isDir {
			translatedPath := translatePathComponents(relPath)
			destPath := filepath.Join(pkgPath, translatedPath)

			destResult := NewFilePath(destPath)
			if !destResult.IsOk() {
				continue
			}

			dirID := OperationID(fmt.Sprintf("adopt-create-dir-%s", translatedPath))
			operations = append(operations, NewDirCreate(dirID, destResult.Unwrap()))
		}
	}

	// Second pass: Move all files and track subdirectories
	var subdirs []string
	for _, relPath := range filesToMove {
		sourcePath := filepath.Join(sourceDir, relPath)

		isDir, _ := s.fs.IsDir(ctx, sourcePath)
		if isDir {
			// Track subdirectories for deletion later
			subdirs = append(subdirs, relPath)
		} else {
			translatedPath := translatePathComponents(relPath)
			destPath := filepath.Join(pkgPath, translatedPath)

			sourceResult := NewTargetPath(sourcePath)
			if !sourceResult.IsOk() {
				continue
			}

			destResult := NewFilePath(destPath)
			if !destResult.IsOk() {
				continue
			}

			moveID := OperationID(fmt.Sprintf("adopt-move-content-%s", relPath))
			operations = append(operations, FileMove{
				OpID:   moveID,
				Source: sourceResult.Unwrap(),
				Dest:   destResult.Unwrap(),
			})
		}
	}

	// Third pass: Delete subdirectories in reverse order (deepest first)
	// This ensures child directories are deleted before parents
	for i := len(subdirs) - 1; i >= 0; i-- {
		subdirPath := filepath.Join(sourceDir, subdirs[i])
		subdirResult := NewFilePath(subdirPath)
		if subdirResult.IsOk() {
			delID := OperationID(fmt.Sprintf("adopt-remove-subdir-%s", subdirs[i]))
			operations = append(operations, NewDirDelete(delID, subdirResult.Unwrap()))
		}
	}

	// Remove the now-empty source directory
	sourceDirResult := NewTargetPath(sourceDir)
	if !sourceDirResult.IsOk() {
		return nil, sourceDirResult.UnwrapErr()
	}
	sourceDirPath := sourceDirResult.Unwrap()

	// Delete the original directory (now empty after moving contents)
	delID := OperationID(fmt.Sprintf("adopt-remove-empty-%s", originalPath))
	sourceDirFilePath := NewFilePath(sourceDir)
	if sourceDirFilePath.IsOk() {
		operations = append(operations, NewDirDelete(delID, sourceDirFilePath.Unwrap()))
	}

	// Create symlink from original location to package root
	pkgRootResult := NewFilePath(pkgPath)
	if !pkgRootResult.IsOk() {
		return nil, pkgRootResult.UnwrapErr()
	}

	linkID := OperationID(fmt.Sprintf("adopt-link-%s", originalPath))
	operations = append(operations, NewLinkCreate(linkID, pkgRootResult.Unwrap(), sourceDirPath))

	return operations, nil
}

// collectDirectoryFiles recursively collects all file paths in a directory.
// Returns paths relative to the root directory.
func (s *AdoptService) collectDirectoryFiles(ctx context.Context, dir, prefix string) ([]string, error) {
	var files []string

	entries, err := s.fs.ReadDir(ctx, dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		relPath := entry.Name()
		if prefix != "" {
			relPath = filepath.Join(prefix, entry.Name())
		}

		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Add directory itself (will be created as empty dir in package)
			files = append(files, relPath)

			// Recursively collect files in subdirectory
			subFiles, err := s.collectDirectoryFiles(ctx, fullPath, relPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			// Regular file
			files = append(files, relPath)
		}
	}

	return files, nil
}

// translatePathComponents applies dotfile translation to each component of a path.
// ".cache/data" → "dot-cache/data"
// "regular/.hidden" → "regular/dot-hidden"
func translatePathComponents(path string) string {
	if path == "" || path == "." {
		return path
	}

	components := filepath.SplitList(path)
	if len(components) == 0 {
		// Not a list, it's a regular path - split by separator
		components = splitPath(path)
	}

	for i, comp := range components {
		components[i] = scanner.UntranslateDotfile(comp)
	}

	return filepath.Join(components...)
}

// splitPath splits a file path into components.
func splitPath(path string) []string {
	var components []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			components = append([]string{file}, components...)
		}
		if dir == "" || dir == "/" {
			break
		}
		path = filepath.Clean(dir)
	}
	return components
}
