package api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Adopt moves existing files from target into package then creates symlinks.
func (c *client) Adopt(ctx context.Context, files []string, pkg string) error {
	plan, err := c.PlanAdopt(ctx, files, pkg)
	if err != nil {
		return err
	}

	if len(plan.Operations) == 0 {
		c.config.Logger.Info(ctx, "nothing_to_adopt")
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

	// Update manifest to track adopted files
	if err := c.updateManifest(ctx, []string{pkg}, plan); err != nil {
		c.config.Logger.Warn(ctx, "failed_to_update_manifest", "error", err)
	}

	return nil
}

// PlanAdopt computes the execution plan for adopting files.
func (c *client) PlanAdopt(ctx context.Context, files []string, pkg string) (dot.Plan, error) {
	packagePathResult := dot.NewPackagePath(c.config.PackageDir)
	if !packagePathResult.IsOk() {
		return dot.Plan{}, packagePathResult.UnwrapErr()
	}

	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.Plan{}, targetPathResult.UnwrapErr()
	}

	// Verify package directory exists
	pkgPath := filepath.Join(c.config.PackageDir, pkg)
	exists := c.config.FS.Exists(ctx, pkgPath)
	if !exists {
		return dot.Plan{}, dot.ErrPackageNotFound{Package: pkg}
	}

	operations := make([]dot.Operation, 0, len(files)*2) // 2 ops per file (move + link)

	// For each file to adopt
	for _, file := range files {
		sourceFile := filepath.Join(c.config.TargetDir, file)

		// Verify source file exists
		if !c.config.FS.Exists(ctx, sourceFile) {
			return dot.Plan{}, dot.ErrSourceNotFound{Path: sourceFile}
		}

		// Translate filename (e.g., .vimrc -> dot-vimrc)
		adoptedName := scanner.UntranslateDotfile(filepath.Base(file))
		destFile := filepath.Join(pkgPath, adoptedName)

		sourcePathResult := dot.NewFilePath(sourceFile)
		if !sourcePathResult.IsOk() {
			return dot.Plan{}, sourcePathResult.UnwrapErr()
		}

		destPathResult := dot.NewFilePath(destFile)
		if !destPathResult.IsOk() {
			return dot.Plan{}, destPathResult.UnwrapErr()
		}

		targetLinkPathResult := dot.NewFilePath(sourceFile)
		if !targetLinkPathResult.IsOk() {
			return dot.Plan{}, targetLinkPathResult.UnwrapErr()
		}

		// Create FileMove operation
		moveID := dot.OperationID(fmt.Sprintf("adopt-move-%s", file))
		operations = append(operations, dot.FileMove{
			OpID:   moveID,
			Source: sourcePathResult.Unwrap(),
			Dest:   destPathResult.Unwrap(),
		})

		// Create LinkCreate operation for symlink
		linkID := dot.OperationID(fmt.Sprintf("adopt-link-%s", file))
		operations = append(operations, dot.NewLinkCreate(linkID, destPathResult.Unwrap(), targetLinkPathResult.Unwrap()))
	}

	return dot.Plan{
		Operations: operations,
		Metadata: dot.PlanMetadata{
			PackageCount:   1,
			OperationCount: len(operations),
		},
	}, nil
}
