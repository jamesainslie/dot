package dot

import (
	"context"
	"os"
	"path/filepath"

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
	targetPath, err := s.getTargetPath()
	if err != nil {
		return DiagnosticReport{}, err
	}

	m, issues, stats, err := s.loadManifestOrCreateDefault(ctx, targetPath)
	if err != nil {
		return DiagnosticReport{}, err
	}
	// If manifest doesn't exist, return early with info issue
	if m == nil {
		return DiagnosticReport{
			OverallHealth: HealthOK,
			Issues:        issues,
			Statistics:    stats,
		}, nil
	}

	s.checkManagedPackages(ctx, m, &issues, &stats)

	if scanCfg.Mode != ScanOff {
		s.performOrphanScan(ctx, m, scanCfg, &issues, &stats)
	}

	health := s.determineOverallHealth(issues)

	return DiagnosticReport{
		OverallHealth: health,
		Issues:        issues,
		Statistics:    stats,
	}, nil
}

// getTargetPath constructs and validates target path.
func (s *DoctorService) getTargetPath() (TargetPath, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return TargetPath{}, targetPathResult.UnwrapErr()
	}
	return targetPathResult.Unwrap(), nil
}

// loadManifestOrCreateDefault loads manifest or returns default state if not found.
func (s *DoctorService) loadManifestOrCreateDefault(ctx context.Context, targetPath TargetPath) (*manifest.Manifest, []Issue, DiagnosticStats, error) {
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
			return nil, issues, stats, nil
		}
		return nil, nil, stats, err
	}

	m := manifestResult.Unwrap()
	return &m, issues, stats, nil
}

// checkManagedPackages validates all packages in the manifest.
func (s *DoctorService) checkManagedPackages(ctx context.Context, m *manifest.Manifest, issues *[]Issue, stats *DiagnosticStats) {
	for pkgName, pkgInfo := range m.Packages {
		stats.ManagedLinks += pkgInfo.LinkCount
		for _, linkPath := range pkgInfo.Links {
			stats.TotalLinks++
			s.checkLink(ctx, pkgName, linkPath, issues, stats)
		}
	}
}

// determineOverallHealth computes health status from issues.
func (s *DoctorService) determineOverallHealth(issues []Issue) HealthStatus {
	health := HealthOK
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return HealthErrors
		}
		if issue.Severity == SeverityWarning && health == HealthOK {
			health = HealthWarnings
		}
	}
	return health
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
	scanDirs := s.determineScanDirectories(m, scanCfg)
	rootDirs := s.normalizeAndDeduplicateDirs(scanDirs, scanCfg.Mode)
	linkSet := buildManagedLinkSet(m)

	for _, dir := range rootDirs {
		s.scanDirectory(ctx, dir, m, linkSet, scanCfg, issues, stats)
	}
}

// determineScanDirectories determines which directories to scan based on configuration.
func (s *DoctorService) determineScanDirectories(m *manifest.Manifest, scanCfg ScanConfig) []string {
	if len(scanCfg.ScopeToDirs) > 0 {
		return scanCfg.ScopeToDirs
	}
	if scanCfg.Mode == ScanScoped {
		return extractManagedDirectories(m)
	}
	return []string{s.targetDir}
}

// normalizeAndDeduplicateDirs converts scan directories to absolute paths and removes descendants.
func (s *DoctorService) normalizeAndDeduplicateDirs(dirs []string, mode ScanMode) []string {
	absDirs := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		fullPath := dir
		if mode == ScanScoped {
			fullPath = filepath.Join(s.targetDir, dir)
		}
		absDirs = append(absDirs, fullPath)
	}
	return filterDescendants(absDirs)
}

// scanDirectory scans a single directory for orphaned links with limit checks.
func (s *DoctorService) scanDirectory(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	err := s.scanForOrphanedLinksWithLimits(ctx, dir, m, linkSet, scanCfg, issues, stats)
	if err != nil {
		// Log but continue - orphan detection is best-effort
		s.logger.Warn(ctx, "scan_directory_failed", "dir", dir, "error", err)
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
		if s.shouldSkipEntry(entry) {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			s.scanDirectoryRecursive(ctx, fullPath, m, linkSet, scanCfg, issues, stats)
		} else {
			s.checkForOrphanedLink(ctx, fullPath, linkSet, issues, stats)
		}
	}
	return nil
}

// shouldSkipEntry checks if directory entry should be skipped.
func (s *DoctorService) shouldSkipEntry(entry DirEntry) bool {
	return entry.Name() == ".dot-manifest.json"
}

// scanDirectoryRecursive recursively scans subdirectory.
func (s *DoctorService) scanDirectoryRecursive(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	err := s.scanForOrphanedLinksWithLimits(ctx, dir, m, linkSet, scanCfg, issues, stats)
	if err != nil {
		// Continue on error - best effort scanning
		s.logger.Warn(ctx, "recursive_scan_failed", "dir", dir, "error", err)
	}
}

// checkForOrphanedLink checks if symlink is orphaned (not in manifest).
func (s *DoctorService) checkForOrphanedLink(
	ctx context.Context,
	fullPath string,
	linkSet map[string]bool,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	relPath, err := filepath.Rel(s.targetDir, fullPath)
	if err != nil {
		relPath = fullPath
	}

	isLink, err := s.fs.IsSymlink(ctx, fullPath)
	if err != nil || !isLink {
		return
	}

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
