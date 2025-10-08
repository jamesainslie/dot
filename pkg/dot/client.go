package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/internal/planner"
)

// Client provides the high-level API for dot operations.
//
// Client acts as a facade that delegates operations to specialized services.
// This design provides clean separation of concerns while maintaining a simple
// public API.
//
// All operations are safe for concurrent use from multiple goroutines.
type Client struct {
	config      Config
	manageSvc   *ManageService
	unmanageSvc *UnmanageService
	statusSvc   *StatusService
	doctorSvc   *DoctorService
	adoptSvc    *AdoptService
}

// NewClient creates a new Client with the given configuration.
//
// Returns an error if:
//   - Configuration is invalid (see Config.Validate)
//   - Required dependencies are missing (FS, Logger)
//
// The returned Client is safe for concurrent use from multiple goroutines.
func NewClient(cfg Config) (*Client, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply defaults
	cfg = cfg.WithDefaults()

	// Create default ignore set
	ignoreSet := ignore.NewDefaultIgnoreSet()

	// Create default resolution policies
	policies := planner.ResolutionPolicies{
		OnFileExists: planner.PolicyFail, // Safe default
	}

	// Create manage pipeline
	managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
		FS:        cfg.FS,
		IgnoreSet: ignoreSet,
		Policies:  policies,
		BackupDir: cfg.BackupDir,
	})

	// Create executor
	exec := executor.New(executor.Opts{
		FS:     cfg.FS,
		Logger: cfg.Logger,
		Tracer: cfg.Tracer,
	})

	// Create manifest store and service
	manifestStore := manifest.NewFSManifestStore(cfg.FS)
	manifestSvc := newManifestService(cfg.FS, cfg.Logger, manifestStore)

	// Create specialized services (unmanageSvc first since manageSvc depends on it)
	unmanageSvc := newUnmanageService(cfg.FS, cfg.Logger, exec, manifestSvc, cfg.TargetDir, cfg.DryRun)
	manageSvc := newManageService(cfg.FS, cfg.Logger, managePipe, exec, manifestSvc, unmanageSvc, cfg.PackageDir, cfg.TargetDir, cfg.DryRun)
	statusSvc := newStatusService(manifestSvc, cfg.TargetDir)
	doctorSvc := newDoctorService(cfg.FS, cfg.Logger, manifestSvc, cfg.TargetDir)
	adoptSvc := newAdoptService(cfg.FS, cfg.Logger, exec, manifestSvc, cfg.PackageDir, cfg.TargetDir, cfg.DryRun)

	return &Client{
		config:      cfg,
		manageSvc:   manageSvc,
		unmanageSvc: unmanageSvc,
		statusSvc:   statusSvc,
		doctorSvc:   doctorSvc,
		adoptSvc:    adoptSvc,
	}, nil
}

// Config returns the client's configuration.
func (c *Client) Config() Config {
	return c.config
}

// === Methods from manage.go ===

// Manage installs the specified packages by creating symlinks.
func (c *Client) Manage(ctx context.Context, packages ...string) error {
	return c.manageSvc.Manage(ctx, packages...)
}

// PlanManage computes the execution plan for managing packages without applying changes.
func (c *Client) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
	return c.manageSvc.PlanManage(ctx, packages...)
}

// These helper functions are now in ManifestService
// Kept here for backward compatibility (not exported)

// extractLinksFromOperations extracts link paths from LinkCreate operations.
func extractLinksFromOperations(ops []Operation, targetDir string) []string {
	links := make([]string, 0, len(ops))
	for _, op := range ops {
		if linkOp, ok := op.(LinkCreate); ok {
			targetPath := linkOp.Target.String()
			relPath, err := filepath.Rel(targetDir, targetPath)
			if err != nil {
				relPath = targetPath
			}
			links = append(links, relPath)
		}
	}
	return links
}

// countLinksInPlan returns the number of LinkCreate operations in a plan.
func countLinksInPlan(plan Plan) int {
	count := 0
	for _, op := range plan.Operations {
		if op.Kind() == OpKindLinkCreate {
			count++
		}
	}
	return count
}

// === Methods from unmanage.go ===

// Unmanage removes the specified packages by deleting symlinks.
func (c *Client) Unmanage(ctx context.Context, packages ...string) error {
	return c.unmanageSvc.Unmanage(ctx, packages...)
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (c *Client) PlanUnmanage(ctx context.Context, packages ...string) (Plan, error) {
	return c.unmanageSvc.PlanUnmanage(ctx, packages...)
}

// === Methods from remanage.go ===

// Remanage reinstalls packages using incremental hash-based change detection.
func (c *Client) Remanage(ctx context.Context, packages ...string) error {
	return c.manageSvc.Remanage(ctx, packages...)
}

// PlanRemanage computes incremental execution plan using hash-based change detection.
func (c *Client) PlanRemanage(ctx context.Context, packages ...string) (Plan, error) {
	return c.manageSvc.PlanRemanage(ctx, packages...)
}

// === Methods from adopt.go ===

// Adopt moves existing files from target into package then creates symlinks.
func (c *Client) Adopt(ctx context.Context, files []string, pkg string) error {
	return c.adoptSvc.Adopt(ctx, files, pkg)
}

// PlanAdopt computes the execution plan for adopting files.
func (c *Client) PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error) {
	return c.adoptSvc.PlanAdopt(ctx, files, pkg)
}

// === Methods from status.go ===

// Status reports the current installation state for packages.
func (c *Client) Status(ctx context.Context, packages ...string) (Status, error) {
	return c.statusSvc.Status(ctx, packages...)
}

// List returns all installed packages from the manifest.
func (c *Client) List(ctx context.Context) ([]PackageInfo, error) {
	return c.statusSvc.List(ctx)
}

// === Methods from doctor.go ===

// Doctor performs health checks with default scan configuration.
func (c *Client) Doctor(ctx context.Context) (DiagnosticReport, error) {
	return c.doctorSvc.Doctor(ctx)
}

// DoctorWithScan performs health checks with explicit scan configuration.
func (c *Client) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
	return c.doctorSvc.DoctorWithScan(ctx, scanCfg)
}

// Helper functions for DoctorService (kept for backward compatibility)
// These are now primarily in DoctorService but kept here for any existing references

// extractManagedDirectories returns unique directories containing managed links.
func extractManagedDirectories(m *manifest.Manifest) []string {
	dirSet := make(map[string]bool)
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			dir := filepath.Dir(link)
			for dir != "." && dir != "/" && dir != "" {
				dirSet[dir] = true
				dir = filepath.Dir(dir)
			}
			dirSet["."] = true
		}
	}
	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs
}

// filterDescendants removes directories that are descendants of other directories.
func filterDescendants(dirs []string) []string {
	if len(dirs) <= 1 {
		return dirs
	}
	cleaned := make([]string, len(dirs))
	for i, dir := range dirs {
		cleaned[i] = filepath.Clean(dir)
	}
	roots := make([]string, 0, len(cleaned))
	for _, dir := range cleaned {
		isDescendant := false
		for _, other := range cleaned {
			if dir == other {
				continue
			}
			rel, err := filepath.Rel(other, dir)
			if err == nil && rel != "." && !filepath.IsAbs(rel) && rel[0] != '.' {
				isDescendant = true
				break
			}
		}
		if !isDescendant {
			roots = append(roots, dir)
		}
	}
	return roots
}

// buildManagedLinkSet creates a set for O(1) link lookup.
func buildManagedLinkSet(m *manifest.Manifest) map[string]bool {
	linkSet := make(map[string]bool)
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			normalized := filepath.ToSlash(link)
			linkSet[normalized] = true
		}
	}
	return linkSet
}

// calculateDepth returns the directory depth relative to target directory.
func calculateDepth(path, targetDir string) int {
	path = filepath.Clean(path)
	targetDir = filepath.Clean(targetDir)
	if path == targetDir {
		return 0
	}
	rel, err := filepath.Rel(targetDir, path)
	if err != nil || rel == "." {
		return 0
	}
	depth := 0
	for _, c := range rel {
		if c == filepath.Separator {
			depth++
		}
	}
	if rel != "" && rel != "." {
		depth++
	}
	return depth
}

// shouldSkipDirectory checks if a directory should be skipped based on patterns.
func shouldSkipDirectory(path string, skipPatterns []string) bool {
	base := filepath.Base(path)
	for _, pattern := range skipPatterns {
		if base == pattern {
			return true
		}
		if filepath.Base(filepath.Dir(path)) == pattern {
			return true
		}
	}
	return false
}

// === Methods from helpers.go ===

// isManifestNotFoundError checks if an error represents a missing manifest file.
func isManifestNotFoundError(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
