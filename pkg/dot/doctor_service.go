package dot

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/manifest"
)

// DoctorService handles health check and diagnostic operations.
type DoctorService struct {
	fs          FS
	logger      Logger
	manifestSvc *ManifestService
	targetDir   string
}

// newDoctorService creates a new doctor service.
func newDoctorService(
	fs FS,
	logger Logger,
	manifestSvc *ManifestService,
	targetDir string,
) *DoctorService {
	return &DoctorService{
		fs:          fs,
		logger:      logger,
		manifestSvc: manifestSvc,
		targetDir:   targetDir,
	}
}

// Doctor performs health checks with default scan configuration.
func (s *DoctorService) Doctor(ctx context.Context) (DiagnosticReport, error) {
	return s.DoctorWithScan(ctx, DefaultScanConfig())
}

// DoctorWithScan performs health checks with explicit scan configuration.
func (s *DoctorService) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return DiagnosticReport{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	issues := make([]Issue, 0)
	stats := DiagnosticStats{}

	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if isManifestNotFoundError(err) {
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
		return DiagnosticReport{}, err
	}

	m := manifestResult.Unwrap()

	// Check each package in the manifest
	for pkgName, pkgInfo := range m.Packages {
		stats.ManagedLinks += pkgInfo.LinkCount
		for _, linkPath := range pkgInfo.Links {
			stats.TotalLinks++
			s.checkLink(ctx, pkgName, linkPath, &issues, &stats)
		}
	}

	// Orphaned link detection (if enabled)
	if scanCfg.Mode != ScanOff {
		s.performOrphanScan(ctx, &m, scanCfg, &issues, &stats)
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

// checkLink validates a single link from the manifest.
func (s *DoctorService) checkLink(ctx context.Context, pkgName string, linkPath string, issues *[]Issue, stats *DiagnosticStats) {
	fullPath := filepath.Join(s.targetDir, linkPath)
	_, err := s.fs.Stat(ctx, fullPath)
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

	isLink, err := s.fs.IsSymlink(ctx, fullPath)
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

	target, err := s.fs.ReadLink(ctx, fullPath)
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

	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = target
	} else {
		absTarget = filepath.Join(filepath.Dir(fullPath), target)
	}

	_, err = s.fs.Stat(ctx, absTarget)
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

// performOrphanScan executes orphaned link scanning based on configuration.
func (s *DoctorService) performOrphanScan(
	ctx context.Context,
	m *manifest.Manifest,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	var scanDirs []string
	if len(scanCfg.ScopeToDirs) > 0 {
		scanDirs = scanCfg.ScopeToDirs
	} else if scanCfg.Mode == ScanScoped {
		scanDirs = extractManagedDirectories(m)
	} else {
		scanDirs = []string{s.targetDir}
	}

	absScanDirs := make([]string, 0, len(scanDirs))
	for _, dir := range scanDirs {
		fullPath := dir
		if scanCfg.Mode == ScanScoped {
			fullPath = filepath.Join(s.targetDir, dir)
		}
		absScanDirs = append(absScanDirs, fullPath)
	}

	rootDirs := filterDescendants(absScanDirs)
	linkSet := buildManagedLinkSet(m)

	for _, dir := range rootDirs {
		err := s.scanForOrphanedLinksWithLimits(ctx, dir, m, linkSet, scanCfg, issues, stats)
		if err != nil {
			continue
		}
	}
}

// scanForOrphanedLinksWithLimits wraps scanForOrphanedLinks with depth and skip checks.
func (s *DoctorService) scanForOrphanedLinksWithLimits(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	depth := calculateDepth(dir, s.targetDir)
	if scanCfg.MaxDepth > 0 && depth > scanCfg.MaxDepth {
		return nil
	}

	if shouldSkipDirectory(dir, scanCfg.SkipPatterns) {
		return nil
	}

	return s.scanForOrphanedLinks(ctx, dir, m, linkSet, scanCfg, issues, stats)
}

// scanForOrphanedLinks recursively scans for symlinks not in the manifest.
func (s *DoctorService) scanForOrphanedLinks(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) error {
	entries, err := s.fs.ReadDir(ctx, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(s.targetDir, fullPath)
		if err != nil {
			relPath = fullPath
		}

		if entry.Name() == ".dot-manifest.json" {
			continue
		}

		if entry.IsDir() {
			if err := s.scanForOrphanedLinksWithLimits(ctx, fullPath, m, linkSet, scanCfg, issues, stats); err != nil {
				continue
			}
		} else {
			isLink, err := s.fs.IsSymlink(ctx, fullPath)
			if err != nil {
				continue
			}
			if isLink {
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
