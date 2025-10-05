package api

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Doctor performs comprehensive health checks on the installation.
func (c *client) Doctor(ctx context.Context) (dot.DiagnosticReport, error) {
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

	// TODO: Implement orphaned link detection with depth limiting
	// Scanning entire home directory is too slow for now
	// Future enhancement: only scan directories containing managed links

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

// scanForOrphanedLinks recursively scans for symlinks not in the manifest.
func (c *client) scanForOrphanedLinks(ctx context.Context, dir string, m *manifest.Manifest, issues *[]dot.Issue, stats *dot.DiagnosticStats) error {
	entries, err := c.config.FS.ReadDir(ctx, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		relPath, _ := filepath.Rel(c.config.TargetDir, fullPath)

		// Skip the manifest file itself
		if entry.Name() == ".dot-manifest.json" {
			continue
		}

		if entry.IsDir() {
			// Recurse into subdirectories
			if err := c.scanForOrphanedLinks(ctx, fullPath, m, issues, stats); err != nil {
				// Continue on error
				continue
			}
		} else {
			// Check if it's a symlink
			isLink, err := c.config.FS.IsSymlink(ctx, fullPath)
			if err != nil {
				continue
			}

			if isLink {
				// It's a symlink - check if it's in any package's links
				managed := false
				for _, pkgInfo := range m.Packages {
					for _, link := range pkgInfo.Links {
						if link == relPath || link == fullPath {
							managed = true
							break
						}
					}
					if managed {
						break
					}
				}

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
