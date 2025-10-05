package api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Unmanage removes the specified packages by deleting symlinks.
func (c *client) Unmanage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanUnmanage(ctx, packages...)
	if err != nil {
		return err
	}

	// Handle empty plan (nothing to unmanage)
	if len(plan.Operations) == 0 {
		c.config.Logger.Info(ctx, "nothing_to_unmanage", "packages", packages)
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

	// Update manifest to remove packages
	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	manifestResult := c.manifest.Load(ctx, targetPath)
	if manifestResult.IsOk() {
		m := manifestResult.Unwrap()
		for _, pkg := range packages {
			m.RemovePackage(pkg)
		}
		if err := c.manifest.Save(ctx, targetPath, m); err != nil {
			c.config.Logger.Warn(ctx, "failed_to_update_manifest", "error", err)
			return err
		}
	}

	return nil
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (c *client) PlanUnmanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		// Check if this is a "file not found" error (no manifest = nothing installed)
		if isManifestNotFoundError(err) {
			// No manifest means nothing is installed - return empty plan
			return dot.Plan{
				Operations: []dot.Operation{},
				Metadata:   dot.PlanMetadata{},
			}, nil
		}
		// Other errors (corrupt manifest, permission errors, etc.) should propagate
		return dot.Plan{}, err
	}

	m := manifestResult.Unwrap()

	// Build delete operations for all links in specified packages
	var operations []dot.Operation
	for _, pkg := range packages {
		pkgInfo, exists := m.Packages[pkg]
		if !exists {
			c.config.Logger.Warn(ctx, "package_not_installed", "package", pkg)
			continue
		}

		// Create delete operations for each link
		for _, link := range pkgInfo.Links {
			targetFilePath := filepath.Join(c.config.TargetDir, link)
			filePathResult := dot.NewFilePath(targetFilePath)
			if !filePathResult.IsOk() {
				continue
			}

			id := dot.OperationID(fmt.Sprintf("unmanage-link-%s", link))
			operations = append(operations, dot.NewLinkDelete(id, filePathResult.Unwrap()))
		}
	}

	return dot.Plan{
		Operations: operations,
		Metadata: dot.PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(operations),
		},
	}, nil
}
