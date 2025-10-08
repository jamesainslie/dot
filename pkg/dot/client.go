package dot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/internal/scanner"
)

// Client provides the high-level API for dot operations.
//
// Client directly uses internal packages (pipeline, executor, manifest)
// to provide a clean public API. No import cycles occur because domain types
// are now in internal/domain.
//
// All operations are safe for concurrent use from multiple goroutines.
type Client struct {
	config     Config
	managePipe *pipeline.ManagePipeline
	executor   *executor.Executor
	manifest   manifest.ManifestStore
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

	// Create manifest store
	manifestStore := manifest.NewFSManifestStore(cfg.FS)

	return &Client{
		config:     cfg,
		managePipe: managePipe,
		executor:   exec,
		manifest:   manifestStore,
	}, nil
}

// Config returns the client's configuration.
func (c *Client) Config() Config {
	return c.config
}

// === Methods from manage.go ===

// Manage installs the specified packages by creating symlinks.
func (c *Client) Manage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanManage(ctx, packages...)
	if err != nil {
		return err
	}
	if c.config.DryRun {
		// In dry-run mode, return early without executing.
		// The CLI layer will handle rendering the plan.
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
func (c *Client) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
	packagePathResult := NewPackagePath(c.config.PackageDir)
	if !packagePathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid package directory: %w", packagePathResult.UnwrapErr())
	}
	packagePath := packagePathResult.Unwrap()
	targetPathResult := NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, fmt.Errorf("invalid target directory: %w", targetPathResult.UnwrapErr())
	}
	targetPath := targetPathResult.Unwrap()
	input := pipeline.ManageInput{
		PackageDir: packagePath,
		TargetDir:  targetPath,
		Packages:   packages,
	}
	planResult := c.managePipe.Execute(ctx, input)
	if !planResult.IsOk() {
		return Plan{}, planResult.UnwrapErr()
	}
	return planResult.Unwrap(), nil
}

// updateManifest updates the manifest with installed packages.
func (c *Client) updateManifest(ctx context.Context, packages []string, plan Plan) error {
	targetPathResult := NewTargetPath(c.config.TargetDir)
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
	hasher := manifest.NewContentHasher(c.config.FS)

	// If packages slice is empty, populate from plan
	packagesToUpdate := packages
	if len(packagesToUpdate) == 0 && plan.PackageOperations != nil {
		packagesToUpdate = plan.PackageNames()
	}

	for _, pkg := range packagesToUpdate {
		// Extract links from package operations
		ops := plan.OperationsForPackage(pkg)
		links := extractLinksFromOperations(ops, c.config.TargetDir)
		m.AddPackage(manifest.PackageInfo{
			Name:        pkg,
			InstalledAt: time.Now(),
			LinkCount:   len(links),
			Links:       links,
		})
		// Compute and store package hash for incremental remanage
		pkgPathStr := filepath.Join(c.config.PackageDir, pkg)
		pkgPathResult := NewPackagePath(pkgPathStr)
		if pkgPathResult.IsOk() {
			pkgPath := pkgPathResult.Unwrap()
			hash, err := hasher.HashPackage(ctx, pkgPath)
			if err != nil {
				c.config.Logger.Warn(ctx, "failed_to_compute_hash", "package", pkg, "error", err)
			} else {
				m.SetHash(pkg, hash)
			}
		}
	}
	// Save manifest
	return c.manifest.Save(ctx, targetPath, m)
}

// extractLinksFromOperations extracts link paths from LinkCreate operations.
// Returns relative paths from the target directory for manifest storage.
func extractLinksFromOperations(ops []Operation, targetDir string) []string {
	links := make([]string, 0, len(ops))
	for _, op := range ops {
		// Only track LinkCreate operations
		if linkOp, ok := op.(LinkCreate); ok {
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
		return ErrMultiple{Errors: execResult.Errors}
	}
	// Update manifest to remove packages
	targetPathResult := NewTargetPath(c.config.TargetDir)
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
func (c *Client) PlanUnmanage(ctx context.Context, packages ...string) (Plan, error) {
	targetPathResult := NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()
	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		// Check if this is a "file not found" error (no manifest = nothing installed)
		if isManifestNotFoundError(err) {
			// No manifest means nothing is installed - return empty plan
			return Plan{
				Operations: []Operation{},
				Metadata:   PlanMetadata{},
			}, nil
		}
		// Other errors (corrupt manifest, permission errors, etc.) should propagate
		return Plan{}, err
	}
	m := manifestResult.Unwrap()
	// Build delete operations for all links in specified packages
	var operations []Operation
	for _, pkg := range packages {
		pkgInfo, exists := m.Packages[pkg]
		if !exists {
			c.config.Logger.Warn(ctx, "package_not_installed", "package", pkg)
			continue
		}
		// Create delete operations for each link
		for _, link := range pkgInfo.Links {
			targetFilePath := filepath.Join(c.config.TargetDir, link)
			filePathResult := NewFilePath(targetFilePath)
			if !filePathResult.IsOk() {
				continue
			}
			id := OperationID(fmt.Sprintf("unmanage-link-%s", link))
			operations = append(operations, NewLinkDelete(id, filePathResult.Unwrap()))
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

// === Methods from remanage.go ===

// Remanage reinstalls packages using incremental hash-based change detection.
// Only processes files that have changed since last installation.
func (c *Client) Remanage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanRemanage(ctx, packages...)
	if err != nil {
		return err
	}
	// Check if no changes detected
	if len(plan.Operations) == 0 {
		c.config.Logger.Info(ctx, "no_changes_detected", "packages", packages)
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
		return ErrMultiple{Errors: execResult.Errors}
	}
	// Update manifest with new hashes
	if err := c.updateManifest(ctx, packages, plan); err != nil {
		c.config.Logger.Warn(ctx, "manifest_update_failed", "error", err)
	}
	return nil
}

// PlanRemanage computes incremental execution plan using hash-based change detection.
// Only includes operations for files that have changed since last installation.
func (c *Client) PlanRemanage(ctx context.Context, packages ...string) (Plan, error) {
	targetPathResult := NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()
	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest - fall back to full manage
		return c.PlanManage(ctx, packages...)
	}
	m := manifestResult.Unwrap()
	hasher := manifest.NewContentHasher(c.config.FS)
	allOperations := make([]Operation, 0)
	packageOps := make(map[string][]OperationID)
	for _, pkg := range packages {
		ops, pkgOpsMap, err := c.planSinglePackageRemanage(ctx, pkg, m, hasher)
		if err != nil {
			return Plan{}, err
		}
		allOperations = append(allOperations, ops...)
		for k, v := range pkgOpsMap {
			packageOps[k] = v
		}
	}
	return Plan{
		Operations: allOperations,
		Metadata: PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(allOperations),
		},
		PackageOperations: packageOps,
	}, nil
}

// planSinglePackageRemanage plans remanage for a single package using hash comparison.
func (c *Client) planSinglePackageRemanage(
	ctx context.Context,
	pkg string,
	m manifest.Manifest,
	hasher *manifest.ContentHasher,
) ([]Operation, map[string][]OperationID, error) {
	_, exists := m.GetPackage(pkg)
	if !exists {
		// Package not installed - plan as new install
		return c.planNewPackageInstall(ctx, pkg)
	}
	// Compute current hash
	pkgPath, err := c.getPackagePath(pkg)
	if err != nil {
		return nil, nil, err
	}
	currentHash, err := hasher.HashPackage(ctx, pkgPath)
	if err != nil {
		// Can't compute hash - fall back to full remanage
		c.config.Logger.Warn(ctx, "hash_computation_failed", "package", pkg, "error", err)
		return c.planFullRemanage(ctx, pkg)
	}
	// Compare with stored hash
	storedHash, hasHash := m.GetHash(pkg)
	if !hasHash || storedHash != currentHash {
		// Package changed - do full unmanage + manage
		return c.planFullRemanage(ctx, pkg)
	}
	// Package unchanged - no operations needed
	c.config.Logger.Info(ctx, "package_unchanged", "package", pkg)
	return []Operation{}, map[string][]OperationID{}, nil
}

// planNewPackageInstall plans installation of a package not yet in manifest.
func (c *Client) planNewPackageInstall(ctx context.Context, pkg string) ([]Operation, map[string][]OperationID, error) {
	pkgPlan, err := c.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	packageOps := make(map[string][]OperationID)
	if pkgPlan.PackageOperations != nil {
		if pkgOps, hasPkg := pkgPlan.PackageOperations[pkg]; hasPkg {
			packageOps[pkg] = pkgOps
		}
	}
	return pkgPlan.Operations, packageOps, nil
}

// planFullRemanage plans full unmanage + manage for a package.
func (c *Client) planFullRemanage(ctx context.Context, pkg string) ([]Operation, map[string][]OperationID, error) {
	unmanagePlan, err := c.PlanUnmanage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	managePlan, err := c.PlanManage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	allOps := make([]Operation, 0, len(unmanagePlan.Operations)+len(managePlan.Operations))
	allOps = append(allOps, unmanagePlan.Operations...)
	allOps = append(allOps, managePlan.Operations...)
	packageOps := make(map[string][]OperationID)
	if managePlan.PackageOperations != nil {
		if pkgOps, hasPkg := managePlan.PackageOperations[pkg]; hasPkg {
			packageOps[pkg] = pkgOps
		}
	}
	return allOps, packageOps, nil
}

// getPackagePath constructs and validates package path.
func (c *Client) getPackagePath(pkg string) (PackagePath, error) {
	pkgPathStr := filepath.Join(c.config.PackageDir, pkg)
	pkgPathResult := NewPackagePath(pkgPathStr)
	if !pkgPathResult.IsOk() {
		return PackagePath{}, pkgPathResult.UnwrapErr()
	}
	return pkgPathResult.Unwrap(), nil
}

// === Methods from adopt.go ===

// Adopt moves existing files from target into package then creates symlinks.
func (c *Client) Adopt(ctx context.Context, files []string, pkg string) error {
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
		return ErrMultiple{Errors: execResult.Errors}
	}
	// Update manifest to track adopted files
	if err := c.updateManifest(ctx, []string{pkg}, plan); err != nil {
		c.config.Logger.Warn(ctx, "failed_to_update_manifest", "error", err)
	}
	return nil
}

// PlanAdopt computes the execution plan for adopting files.
func (c *Client) PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error) {
	packagePathResult := NewPackagePath(c.config.PackageDir)
	if !packagePathResult.IsOk() {
		return Plan{}, packagePathResult.UnwrapErr()
	}
	targetPathResult := NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return Plan{}, targetPathResult.UnwrapErr()
	}
	// Verify package directory exists
	pkgPath := filepath.Join(c.config.PackageDir, pkg)
	exists := c.config.FS.Exists(ctx, pkgPath)
	if !exists {
		return Plan{}, ErrPackageNotFound{Package: pkg}
	}
	operations := make([]Operation, 0, len(files)*2) // 2 ops per file (move + link)
	// For each file to adopt
	for _, file := range files {
		sourceFile := filepath.Join(c.config.TargetDir, file)
		// Verify source file exists
		if !c.config.FS.Exists(ctx, sourceFile) {
			return Plan{}, ErrSourceNotFound{Path: sourceFile}
		}
		// Translate filename (e.g., .vimrc -> dot-vimrc)
		adoptedName := scanner.UntranslateDotfile(filepath.Base(file))
		destFile := filepath.Join(pkgPath, adoptedName)
		sourcePathResult := NewFilePath(sourceFile)
		if !sourcePathResult.IsOk() {
			return Plan{}, sourcePathResult.UnwrapErr()
		}
		destPathResult := NewFilePath(destFile)
		if !destPathResult.IsOk() {
			return Plan{}, destPathResult.UnwrapErr()
		}
		targetLinkPathResult := NewFilePath(sourceFile)
		if !targetLinkPathResult.IsOk() {
			return Plan{}, targetLinkPathResult.UnwrapErr()
		}
		// Create FileMove operation
		moveID := OperationID(fmt.Sprintf("adopt-move-%s", file))
		operations = append(operations, FileMove{
			OpID:   moveID,
			Source: sourcePathResult.Unwrap(),
			Dest:   destPathResult.Unwrap(),
		})
		// Create LinkCreate operation for symlink
		linkID := OperationID(fmt.Sprintf("adopt-link-%s", file))
		operations = append(operations, NewLinkCreate(linkID, destPathResult.Unwrap(), targetLinkPathResult.Unwrap()))
	}
	return Plan{
		Operations: operations,
		Metadata: PlanMetadata{
			PackageCount:   1,
			OperationCount: len(operations),
		},
	}, nil
}

// === Methods from status.go ===

// Status reports the current installation state for packages.
func (c *Client) Status(ctx context.Context, packages ...string) (Status, error) {
	targetPathResult := NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return Status{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()
	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest means nothing installed
		return Status{Packages: []PackageInfo{}}, nil
	}
	m := manifestResult.Unwrap()
	// Filter to requested packages if specified
	pkgInfos := make([]PackageInfo, 0)
	if len(packages) == 0 {
		// Return all packages
		for _, info := range m.Packages {
			pkgInfos = append(pkgInfos, PackageInfo{
				Name:        info.Name,
				InstalledAt: info.InstalledAt,
				LinkCount:   info.LinkCount,
				Links:       info.Links,
			})
		}
	} else {
		// Return only specified packages
		for _, pkg := range packages {
			if info, exists := m.Packages[pkg]; exists {
				pkgInfos = append(pkgInfos, PackageInfo{
					Name:        info.Name,
					InstalledAt: info.InstalledAt,
					LinkCount:   info.LinkCount,
					Links:       info.Links,
				})
			}
		}
	}
	return Status{
		Packages: pkgInfos,
	}, nil
}

// List returns all installed packages from the manifest.
func (c *Client) List(ctx context.Context) ([]PackageInfo, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return nil, err
	}
	return status.Packages, nil
}

// === Methods from doctor.go ===

// Doctor performs health checks with default scan configuration.
// Uses DefaultScanConfig() which performs no orphan scanning for
// backward compatibility and performance.
func (c *Client) Doctor(ctx context.Context) (DiagnosticReport, error) {
	return c.DoctorWithScan(ctx, DefaultScanConfig())
}

// DoctorWithScan performs health checks with explicit scan configuration.
func (c *Client) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
	targetPathResult := NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return DiagnosticReport{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()
	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	issues := make([]Issue, 0)
	stats := DiagnosticStats{}
	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		// Check if this is a not-found error vs other errors
		if isManifestNotFoundError(err) {
			// No manifest - report as info
			issues = append(issues, Issue{
				Severity:   SeverityInfo,
				Type:       IssueManifestInconsistency,
				Message:    "No manifest found - no packages are currently managed",
				Suggestion: "Run 'dot manage' to install packages",
			})
			return DiagnosticReport{
				OverallHealth: HealthOK,
				Issues:        issues,
				Statistics:    stats,
			}, nil
		}
		// Other errors (corrupt manifest, permission errors, etc.) should propagate
		return DiagnosticReport{}, err
	}
	m := manifestResult.Unwrap()
	// Check each package in the manifest
	for pkgName, pkgInfo := range m.Packages {
		stats.ManagedLinks += pkgInfo.LinkCount
		// Check each link
		for _, linkPath := range pkgInfo.Links {
			stats.TotalLinks++
			c.checkLink(ctx, pkgName, linkPath, &issues, &stats)
		}
	}
	// Orphaned link detection (if enabled)
	if scanCfg.Mode != ScanOff {
		c.performOrphanScan(ctx, &m, scanCfg, &issues, &stats)
	}
	// Determine overall health
	health := HealthOK
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			health = HealthErrors
			break
		}
		if issue.Severity == SeverityWarning && health == HealthOK {
			health = HealthWarnings
		}
	}
	return DiagnosticReport{
		OverallHealth: health,
		Issues:        issues,
		Statistics:    stats,
	}, nil
}

// performOrphanScan executes orphaned link scanning based on configuration.
func (c *Client) performOrphanScan(
	ctx context.Context,
	m *manifest.Manifest,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	// Determine scan directories
	var scanDirs []string
	if len(scanCfg.ScopeToDirs) > 0 {
		// Use explicitly provided directories
		scanDirs = scanCfg.ScopeToDirs
	} else if scanCfg.Mode == ScanScoped {
		// Auto-detect from manifest
		scanDirs = extractManagedDirectories(m)
	} else {
		// Deep scan - use target directory
		scanDirs = []string{c.config.TargetDir}
	}
	// Normalize to absolute paths and deduplicate
	absScanDirs := make([]string, 0, len(scanDirs))
	for _, dir := range scanDirs {
		// For scoped mode, dir is relative; for deep mode, it's absolute
		fullPath := dir
		if scanCfg.Mode == ScanScoped {
			fullPath = filepath.Join(c.config.TargetDir, dir)
		}
		absScanDirs = append(absScanDirs, fullPath)
	}
	// Remove descendants to avoid rescanning subdirectories
	rootDirs := filterDescendants(absScanDirs)
	// Build link set for O(1) lookup
	linkSet := buildManagedLinkSet(m)
	// Scan each root directory (recursion will cover descendants)
	for _, dir := range rootDirs {
		err := c.scanForOrphanedLinksWithLimits(ctx, dir, m, linkSet, scanCfg, issues, stats)
		if err != nil {
			// Log but continue - orphan detection is best-effort
			continue
		}
	}
}

// extractManagedDirectories returns unique directories containing managed links.
func extractManagedDirectories(m *manifest.Manifest) []string {
	dirSet := make(map[string]bool)
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			// Extract all parent directories
			dir := filepath.Dir(link)
			for dir != "." && dir != "/" && dir != "" {
				dirSet[dir] = true
				dir = filepath.Dir(dir)
			}
			// Always include root
			dirSet["."] = true
		}
	}
	// Convert set to slice
	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs
}

// filterDescendants removes directories that are descendants of other directories in the list.
// This prevents rescanning the same subtrees multiple times.
func filterDescendants(dirs []string) []string {
	if len(dirs) <= 1 {
		return dirs
	}
	// Clean all paths
	cleaned := make([]string, len(dirs))
	for i, dir := range dirs {
		cleaned[i] = filepath.Clean(dir)
	}
	// Filter out descendants
	roots := make([]string, 0, len(cleaned))
	for _, dir := range cleaned {
		isDescendant := false
		// Check if this dir is a descendant of any other dir
		for _, other := range cleaned {
			if dir == other {
				continue
			}
			// Check if dir is under other
			rel, err := filepath.Rel(other, dir)
			if err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
				// dir is a descendant of other
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
// Normalizes paths to forward slashes for cross-platform compatibility.
func buildManagedLinkSet(m *manifest.Manifest) map[string]bool {
	linkSet := make(map[string]bool)
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			// Normalize to forward slashes for consistent lookup on all platforms
			normalized := filepath.ToSlash(link)
			linkSet[normalized] = true
		}
	}
	return linkSet
}

// calculateDepth returns the directory depth relative to target directory.
func calculateDepth(path, targetDir string) int {
	// Clean both paths
	path = filepath.Clean(path)
	targetDir = filepath.Clean(targetDir)
	// If same directory, depth is 0
	if path == targetDir {
		return 0
	}
	// Get relative path
	rel, err := filepath.Rel(targetDir, path)
	if err != nil || rel == "." {
		return 0
	}
	// Count separators in relative path
	depth := 0
	for _, c := range rel {
		if c == filepath.Separator {
			depth++
		}
	}
	// If path doesn't end with separator, add 1
	if rel != "" && rel != "." {
		depth++
	}
	return depth
}

// shouldSkipDirectory checks if a directory should be skipped based on patterns.
func shouldSkipDirectory(path string, skipPatterns []string) bool {
	// Get basename for matching
	base := filepath.Base(path)
	for _, pattern := range skipPatterns {
		if base == pattern {
			return true
		}
		// Also check if any component in path matches
		if strings.Contains(path, string(filepath.Separator)+pattern+string(filepath.Separator)) ||
			strings.HasSuffix(path, string(filepath.Separator)+pattern) {
			return true
		}
	}
	return false
}

// scanForOrphanedLinksWithLimits wraps scanForOrphanedLinks with depth and skip checks.
func (c *Client) scanForOrphanedLinksWithLimits(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) error {
	// Check context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Check depth limit
	depth := calculateDepth(dir, c.config.TargetDir)
	if scanCfg.MaxDepth > 0 && depth > scanCfg.MaxDepth {
		return nil // Skip too-deep directories
	}
	// Check skip patterns
	if shouldSkipDirectory(dir, scanCfg.SkipPatterns) {
		return nil
	}
	// Scan this directory, passing config for recursive depth/context checks
	return c.scanForOrphanedLinks(ctx, dir, m, linkSet, scanCfg, issues, stats)
}

// scanForOrphanedLinks recursively scans for symlinks not in the manifest.
func (c *Client) scanForOrphanedLinks(ctx context.Context, dir string, m *manifest.Manifest, linkSet map[string]bool, scanCfg ScanConfig, issues *[]Issue, stats *DiagnosticStats) error {
	entries, err := c.config.FS.ReadDir(ctx, dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(c.config.TargetDir, fullPath)
		if err != nil {
			// Use fullPath as fallback if relative path cannot be determined
			relPath = fullPath
		}
		// Skip the manifest file itself
		if entry.Name() == ".dot-manifest.json" {
			continue
		}
		if entry.IsDir() {
			// Recurse into subdirectories using wrapper to enforce all safety checks
			if err := c.scanForOrphanedLinksWithLimits(ctx, fullPath, m, linkSet, scanCfg, issues, stats); err != nil {
				// Continue on error - best effort scanning
				continue
			}
		} else {
			// Check if it's a symlink
			isLink, err := c.config.FS.IsSymlink(ctx, fullPath)
			if err != nil {
				continue
			}
			if isLink {
				// It's a symlink - check if it's managed using O(1) set lookup
				// Normalize paths to forward slashes for cross-platform compatibility
				normalizedRel := filepath.ToSlash(relPath)
				normalizedFull := filepath.ToSlash(fullPath)
				managed := linkSet[normalizedRel] || linkSet[normalizedFull]
				if !managed {
					stats.OrphanedLinks++
					*issues = append(*issues, Issue{
						Severity:   SeverityWarning,
						Type:       IssueOrphanedLink,
						Path:       relPath,
						Message:    "Symlink not managed by dot",
						Suggestion: "Remove manually or use 'dot adopt' to bring under management",
					})
				}
			}
		}
	}
	return nil
}

// checkLink validates a single link from the manifest.
func (c *Client) checkLink(ctx context.Context, pkgName string, linkPath string, issues *[]Issue, stats *DiagnosticStats) {
	fullPath := filepath.Join(c.config.TargetDir, linkPath)
	// Check if link exists
	_, err := c.config.FS.Stat(ctx, fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			stats.BrokenLinks++
			*issues = append(*issues, Issue{
				Severity:   SeverityError,
				Type:       IssueBrokenLink,
				Path:       linkPath,
				Message:    "Link does not exist",
				Suggestion: "Run 'dot remanage " + pkgName + "' to restore link",
			})
		} else {
			*issues = append(*issues, Issue{
				Severity:   SeverityError,
				Type:       IssuePermission,
				Path:       linkPath,
				Message:    "Cannot access link: " + err.Error(),
				Suggestion: "Check filesystem permissions",
			})
		}
		return
	}
	// Check if it's actually a symlink
	isLink, err := c.config.FS.IsSymlink(ctx, fullPath)
	if err != nil {
		*issues = append(*issues, Issue{
			Severity:   SeverityError,
			Type:       IssuePermission,
			Path:       linkPath,
			Message:    "Cannot check if path is symlink: " + err.Error(),
			Suggestion: "Check filesystem permissions",
		})
		return
	}
	if !isLink {
		*issues = append(*issues, Issue{
			Severity:   SeverityError,
			Type:       IssueWrongTarget,
			Path:       linkPath,
			Message:    "Expected symlink but found regular file",
			Suggestion: "Run 'dot unmanage " + pkgName + "' then 'dot manage " + pkgName + "'",
		})
		return
	}
	// Check where the link points
	target, err := c.config.FS.ReadLink(ctx, fullPath)
	if err != nil {
		*issues = append(*issues, Issue{
			Severity:   SeverityError,
			Type:       IssuePermission,
			Path:       linkPath,
			Message:    "Cannot read link target: " + err.Error(),
			Suggestion: "Check filesystem permissions",
		})
		return
	}
	// Resolve to absolute path
	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = target
	} else {
		absTarget = filepath.Join(filepath.Dir(fullPath), target)
	}
	// Check if target exists
	_, err = c.config.FS.Stat(ctx, absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			stats.BrokenLinks++
			*issues = append(*issues, Issue{
				Severity:   SeverityError,
				Type:       IssueBrokenLink,
				Path:       linkPath,
				Message:    "Link target does not exist: " + target,
				Suggestion: "Run 'dot remanage " + pkgName + "' to fix broken link",
			})
		}
	}
}

// === Methods from helpers.go ===

// isManifestNotFoundError checks if an error represents a missing manifest file.
func isManifestNotFoundError(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
