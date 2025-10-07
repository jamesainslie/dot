package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Doctor performs health checks with default scan configuration.
// Uses DefaultScanConfig() which performs no orphan scanning for
// backward compatibility and performance.
func (c *client) Doctor(ctx context.Context) (dot.DiagnosticReport, error) {
	return c.DoctorWithScan(ctx, dot.DefaultScanConfig())
}

// DoctorWithScan performs health checks with explicit scan configuration.
func (c *client) DoctorWithScan(ctx context.Context, scanCfg dot.ScanConfig) (dot.DiagnosticReport, error) {
	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.DiagnosticReport{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)

	issues := make([]dot.Issue, 0)
	stats := dot.DiagnosticStats{}

	if !manifestResult.IsOk() {
		// No manifest - report as info
		issues = append(issues, dot.Issue{
			Severity:   dot.SeverityInfo,
			Type:       dot.IssueManifestInconsistency,
			Message:    "No manifest found - no packages are currently managed",
			Suggestion: "Run 'dot manage' to install packages",
		})

		return dot.DiagnosticReport{
			OverallHealth: dot.HealthOK,
			Issues:        issues,
			Statistics:    stats,
		}, nil
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
	if scanCfg.Mode != dot.ScanOff {
		c.performOrphanScan(ctx, &m, scanCfg, &issues, &stats)
	}

	// Determine overall health
	health := dot.HealthOK
	for _, issue := range issues {
		if issue.Severity == dot.SeverityError {
			health = dot.HealthErrors
			break
		}
		if issue.Severity == dot.SeverityWarning && health == dot.HealthOK {
			health = dot.HealthWarnings
		}
	}

	return dot.DiagnosticReport{
		OverallHealth: health,
		Issues:        issues,
		Statistics:    stats,
	}, nil
}

// performOrphanScan executes orphaned link scanning based on configuration.
func (c *client) performOrphanScan(
	ctx context.Context,
	m *manifest.Manifest,
	scanCfg dot.ScanConfig,
	issues *[]dot.Issue,
	stats *dot.DiagnosticStats,
) {
	// Determine scan directories
	var scanDirs []string
	if len(scanCfg.ScopeToDirs) > 0 {
		// Use explicitly provided directories
		scanDirs = scanCfg.ScopeToDirs
	} else if scanCfg.Mode == dot.ScanScoped {
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
		if scanCfg.Mode == dot.ScanScoped {
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
func (c *client) scanForOrphanedLinksWithLimits(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	scanCfg dot.ScanConfig,
	issues *[]dot.Issue,
	stats *dot.DiagnosticStats,
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
func (c *client) scanForOrphanedLinks(ctx context.Context, dir string, m *manifest.Manifest, linkSet map[string]bool, scanCfg dot.ScanConfig, issues *[]dot.Issue, stats *dot.DiagnosticStats) error {
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
					*issues = append(*issues, dot.Issue{
						Severity:   dot.SeverityWarning,
						Type:       dot.IssueOrphanedLink,
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
func (c *client) checkLink(ctx context.Context, pkgName string, linkPath string, issues *[]dot.Issue, stats *dot.DiagnosticStats) {
	fullPath := filepath.Join(c.config.TargetDir, linkPath)

	// Check if link exists
	_, err := c.config.FS.Stat(ctx, fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			stats.BrokenLinks++
			*issues = append(*issues, dot.Issue{
				Severity:   dot.SeverityError,
				Type:       dot.IssueBrokenLink,
				Path:       linkPath,
				Message:    "Link does not exist",
				Suggestion: "Run 'dot remanage " + pkgName + "' to restore link",
			})
		} else {
			*issues = append(*issues, dot.Issue{
				Severity:   dot.SeverityError,
				Type:       dot.IssuePermission,
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
		*issues = append(*issues, dot.Issue{
			Severity:   dot.SeverityError,
			Type:       dot.IssuePermission,
			Path:       linkPath,
			Message:    "Cannot check if path is symlink: " + err.Error(),
			Suggestion: "Check filesystem permissions",
		})
		return
	}

	if !isLink {
		*issues = append(*issues, dot.Issue{
			Severity:   dot.SeverityError,
			Type:       dot.IssueWrongTarget,
			Path:       linkPath,
			Message:    "Expected symlink but found regular file",
			Suggestion: "Run 'dot unmanage " + pkgName + "' then 'dot manage " + pkgName + "'",
		})
		return
	}

	// Check where the link points
	target, err := c.config.FS.ReadLink(ctx, fullPath)
	if err != nil {
		*issues = append(*issues, dot.Issue{
			Severity:   dot.SeverityError,
			Type:       dot.IssuePermission,
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
			*issues = append(*issues, dot.Issue{
				Severity:   dot.SeverityError,
				Type:       dot.IssueBrokenLink,
				Path:       linkPath,
				Message:    "Link target does not exist: " + target,
				Suggestion: "Run 'dot remanage " + pkgName + "' to fix broken link",
			})
		}
	}
}
