package api

import (
	"context"
	"fmt"
	"path/filepath"
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
	packagePathResult := dot.NewPackagePath(c.config.PackageDir)
	if !packagePathResult.IsOk() {
		return dot.Plan{}, fmt.Errorf("invalid package directory: %w", packagePathResult.UnwrapErr())
	}
	packagePath := packagePathResult.Unwrap()

	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()

	input := pipeline.ManageInput{
		PackageDir: packagePath,
		TargetDir:  targetPath,
		Packages:   packages,
	}

	planResult := c.managePipe.Execute(ctx, input)
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

	// Update package entries using package-operation mapping from plan
	for _, pkg := range packages {
		// Extract links from package operations
		ops := plan.OperationsForPackage(pkg)
		links := extractLinksFromOperations(ops, c.config.TargetDir)

		m.AddPackage(manifest.PackageInfo{
			Name:        pkg,
			InstalledAt: time.Now(),
			LinkCount:   len(links),
			Links:       links,
		})
	}

	// Save manifest
	return c.manifest.Save(ctx, targetPath, m)
}

// extractLinksFromOperations extracts link paths from LinkCreate operations.
// Returns relative paths from the target directory for manifest storage.
func extractLinksFromOperations(ops []dot.Operation, targetDir string) []string {
	links := make([]string, 0, len(ops))

	for _, op := range ops {
		// Only track LinkCreate operations
		if linkOp, ok := op.(dot.LinkCreate); ok {
			// Get relative path from target directory
			targetPath := linkOp.Target.String()
			relPath, err := filepath.Rel(targetDir, targetPath)
			if err != nil {
				// If we can't compute relative path, use absolute path
				relPath = targetPath
			}
			links = append(links, relPath)
		}
	}

	return links
}

// countLinksInPlan returns the number of LinkCreate operations in a plan.
func countLinksInPlan(plan dot.Plan) int {
	count := 0
	for _, op := range plan.Operations {
		if op.Kind() == dot.OpKindLinkCreate {
			count++
		}
	}
	return count
}
