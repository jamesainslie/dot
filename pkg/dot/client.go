package dot

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/cli/selector"
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
	cloneSvc    *CloneService
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
		FS:                 cfg.FS,
		IgnoreSet:          ignoreSet,
		Policies:           policies,
		BackupDir:          cfg.BackupDir,
		PackageNameMapping: cfg.PackageNameMapping,
	})

	// Create executor
	exec := executor.New(executor.Opts{
		FS:     cfg.FS,
		Logger: cfg.Logger,
		Tracer: cfg.Tracer,
	})

	// Create manifest store and service
	var manifestStore *manifest.FSManifestStore
	if cfg.ManifestDir != "" {
		manifestStore = manifest.NewFSManifestStoreWithDir(cfg.FS, cfg.ManifestDir)
	} else {
		manifestStore = manifest.NewFSManifestStore(cfg.FS)
	}
	manifestSvc := newManifestService(cfg.FS, cfg.Logger, manifestStore)

	// Create specialized services (unmanageSvc first since manageSvc depends on it)
	unmanageSvc := newUnmanageService(cfg.FS, cfg.Logger, exec, manifestSvc, cfg.PackageDir, cfg.TargetDir, cfg.DryRun)
	manageSvc := newManageService(cfg.FS, cfg.Logger, managePipe, exec, manifestSvc, unmanageSvc, cfg.PackageDir, cfg.TargetDir, cfg.DryRun)
	statusSvc := newStatusService(manifestSvc, cfg.TargetDir)
	doctorSvc := newDoctorService(cfg.FS, cfg.Logger, manifestSvc, cfg.TargetDir)
	adoptSvc := newAdoptService(cfg.FS, cfg.Logger, exec, manifestSvc, cfg.PackageDir, cfg.TargetDir, cfg.DryRun)

	// Create git cloner and package selector for clone service
	gitCloner := adapters.NewGoGitCloner()
	packageSelector := selector.NewInteractiveSelector(os.Stdin, os.Stdout)
	cloneSvc := newCloneService(cfg.FS, cfg.Logger, manageSvc, gitCloner, packageSelector, cfg.PackageDir, cfg.TargetDir, cfg.DryRun)

	return &Client{
		config:      cfg,
		manageSvc:   manageSvc,
		unmanageSvc: unmanageSvc,
		statusSvc:   statusSvc,
		doctorSvc:   doctorSvc,
		adoptSvc:    adoptSvc,
		cloneSvc:    cloneSvc,
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

// === Methods from unmanage.go ===

// Unmanage removes the specified packages by deleting symlinks.
// Adopted packages are automatically restored unless disabled.
func (c *Client) Unmanage(ctx context.Context, packages ...string) error {
	return c.unmanageSvc.Unmanage(ctx, packages...)
}

// UnmanageWithOptions removes packages with specified options.
func (c *Client) UnmanageWithOptions(ctx context.Context, opts UnmanageOptions, packages ...string) error {
	return c.unmanageSvc.UnmanageWithOptions(ctx, opts, packages...)
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

// Clone clones a dotfiles repository and installs packages.
//
// Workflow:
//  1. Validates package directory is empty (unless Force=true)
//  2. Clones repository to package directory
//  3. Loads optional bootstrap configuration
//  4. Selects packages (via profile, interactive, or all)
//  5. Filters packages by current platform
//  6. Installs selected packages
//  7. Updates manifest with repository tracking
//
// Returns an error if:
//   - Package directory is not empty (and Force=false)
//   - Authentication fails
//   - Clone operation fails
//   - Bootstrap config is invalid
//   - Package installation fails
func (c *Client) Clone(ctx context.Context, repoURL string, opts CloneOptions) error {
	return c.cloneSvc.Clone(ctx, repoURL, opts)
}

// === Methods from helpers.go ===

// isManifestNotFoundError checks if an error represents a missing manifest file.
func isManifestNotFoundError(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
