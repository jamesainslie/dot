package api

import (
	"context"
	"fmt"
	"time"

	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Manage installs the specified packages by creating symlinks.
func (c *client) Manage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanManage(ctx, packages...)
	if err != nil {
		return err
	}

	if c.config.DryRun {
		c.config.Logger.Info(ctx, "dry_run_mode", "operations", len(plan.Operations))
		return nil
	}

	result := c.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}

	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}

	// Update manifest
	if err := c.updateManifest(ctx, packages, plan); err != nil {
		c.config.Logger.Warn(ctx, "manifest_update_failed", "error", err)
		// Don't fail the operation if manifest update fails
	}

	return nil
}

// PlanManage computes the execution plan for managing packages without applying changes.
func (c *client) PlanManage(ctx context.Context, packages ...string) (dot.Plan, error) {
	stowPathResult := dot.NewPackagePath(c.config.StowDir)
	if !stowPathResult.IsOk() {
		return dot.Plan{}, fmt.Errorf("invalid stow directory: %w", stowPathResult.UnwrapErr())
	}
	stowPath := stowPathResult.Unwrap()

	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()

	input := pipeline.StowInput{
		StowDir:   stowPath,
		TargetDir: targetPath,
		Packages:  packages,
	}

	planResult := c.stowPipe.Execute(ctx, input)
	if !planResult.IsOk() {
		return dot.Plan{}, planResult.UnwrapErr()
	}

	return planResult.Unwrap(), nil
}

// updateManifest updates the manifest with installed packages.
func (c *client) updateManifest(ctx context.Context, packages []string, plan dot.Plan) error {
	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load existing manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	var m manifest.Manifest
	if manifestResult.IsOk() {
		m = manifestResult.Unwrap()
	} else {
		// Create new manifest
		m = manifest.New()
	}

	// Update package entries
	for _, pkg := range packages {
		// Count links created for this package
		linkCount := 0
		var links []string
		for _, op := range plan.Operations {
			if linkCreate, ok := op.(dot.LinkCreate); ok {
				// Check if this link belongs to this package
				// (simplified - full implementation would track this better)
				linkCount++
				links = append(links, linkCreate.Target.String())
			}
		}

		m.AddPackage(manifest.PackageInfo{
			Name:        pkg,
			InstalledAt: time.Now(),
			LinkCount:   linkCount,
			Links:       links,
		})
	}

	// Save manifest
	return c.manifest.Save(ctx, targetPath, m)
}
