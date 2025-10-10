package dot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/bootstrap"
	"github.com/jamesainslie/dot/internal/cli/selector"
	"github.com/jamesainslie/dot/internal/cli/terminal"
	"github.com/jamesainslie/dot/internal/manifest"
)

// CloneService handles repository cloning and package installation.
type CloneService struct {
	fs         FS
	logger     Logger
	manageSvc  *ManageService
	cloner     adapters.GitCloner
	selector   selector.PackageSelector
	packageDir string
	targetDir  string
	dryRun     bool
}

// newCloneService creates a new clone service.
func newCloneService(
	fs FS,
	logger Logger,
	manageSvc *ManageService,
	cloner adapters.GitCloner,
	sel selector.PackageSelector,
	packageDir string,
	targetDir string,
	dryRun bool,
) *CloneService {
	return &CloneService{
		fs:         fs,
		logger:     logger,
		manageSvc:  manageSvc,
		cloner:     cloner,
		selector:   sel,
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     dryRun,
	}
}

// CloneOptions configures repository cloning behavior.
type CloneOptions struct {
	// Profile specifies which bootstrap profile to use.
	// If empty, uses default profile or interactive selection.
	Profile string

	// Interactive forces interactive package selection.
	// If false, uses profile or installs all packages.
	Interactive bool

	// Force allows cloning into non-empty packageDir.
	Force bool

	// Branch specifies which branch to clone.
	// If empty, clones default branch.
	Branch string
}

// Clone clones a repository and installs packages.
//
// Workflow:
//  1. Validate packageDir is empty (unless Force=true)
//  2. Resolve authentication from environment
//  3. Clone repository to packageDir
//  4. Load bootstrap config if present
//  5. Select packages (profile, interactive, or all)
//  6. Filter packages by current platform
//  7. Install selected packages via ManageService
//  8. Update manifest with repository information
func (s *CloneService) Clone(ctx context.Context, repoURL string, opts CloneOptions) error {
	// Validate package directory
	if err := validatePackageDir(ctx, s.fs, s.packageDir, opts.Force); err != nil {
		return err
	}

	// Resolve authentication
	auth, err := adapters.ResolveAuth(ctx, repoURL)
	if err != nil {
		return ErrAuthFailed{Cause: err}
	}

	s.logger.Info(ctx, "cloning_repository", "url", repoURL, "packageDir", s.packageDir)

	// Clone repository
	cloneOpts := adapters.CloneOptions{
		Auth:   auth,
		Branch: opts.Branch,
		Depth:  1, // Shallow clone for faster cloning
	}

	if err := s.cloner.Clone(ctx, repoURL, s.packageDir, cloneOpts); err != nil {
		return ErrCloneFailed{URL: repoURL, Cause: err}
	}

	s.logger.Info(ctx, "clone_successful", "path", s.packageDir)

	// Load bootstrap configuration if present
	bootstrapConfig, hasBootstrap, err := loadBootstrapConfig(ctx, s.fs, s.packageDir)
	if err != nil {
		return err
	}

	// Select packages to install
	var packagesToInstall []string
	if hasBootstrap {
		packagesToInstall, err = s.selectPackagesWithBootstrap(ctx, bootstrapConfig, opts)
	} else {
		packagesToInstall, err = s.selectPackagesWithoutBootstrap(ctx, opts)
	}
	if err != nil {
		return err
	}

	if len(packagesToInstall) == 0 {
		s.logger.Info(ctx, "no_packages_selected")
		return nil
	}

	s.logger.Info(ctx, "installing_packages", "count", len(packagesToInstall), "packages", packagesToInstall)

	// Install packages
	if s.dryRun {
		s.logger.Info(ctx, "dry_run_complete", "would_install", packagesToInstall)
		return nil
	}

	if err := s.manageSvc.Manage(ctx, packagesToInstall...); err != nil {
		return fmt.Errorf("install packages: %w", err)
	}

	// Update manifest with repository information
	branch := opts.Branch
	if branch == "" {
		branch = "main" // Default branch assumption
	}

	commitSHA, _ := getCommitSHA(s.packageDir) // Best effort, ignore errors
	repoInfo := buildRepositoryInfo(repoURL, branch, commitSHA)

	if err := s.updateManifestRepository(ctx, repoInfo); err != nil {
		s.logger.Warn(ctx, "failed_to_update_manifest_repository", "error", err)
	}

	s.logger.Info(ctx, "clone_complete", "packages_installed", len(packagesToInstall))
	return nil
}

// selectPackagesWithBootstrap selects packages using bootstrap configuration.
func (s *CloneService) selectPackagesWithBootstrap(ctx context.Context, config bootstrap.Config, opts CloneOptions) ([]string, error) {
	// Filter packages by platform
	filtered := bootstrap.FilterPackagesByPlatform(config.Packages, runtime.GOOS)
	allPackages := extractPackageNames(filtered)

	// If profile specified, use it
	if opts.Profile != "" {
		return selectPackagesFromProfile(config, opts.Profile)
	}

	// If interactive flag set, prompt user
	if opts.Interactive || terminal.IsInteractive() {
		return s.selector.Select(ctx, allPackages)
	}

	// Use default profile if configured
	if config.Defaults.Profile != "" {
		return selectPackagesFromProfile(config, config.Defaults.Profile)
	}

	// Install all packages
	return allPackages, nil
}

// selectPackagesWithoutBootstrap selects packages when no bootstrap config exists.
func (s *CloneService) selectPackagesWithoutBootstrap(ctx context.Context, opts CloneOptions) ([]string, error) {
	// Discover packages in directory
	packages, err := discoverPackages(ctx, s.fs, s.packageDir)
	if err != nil {
		return nil, fmt.Errorf("discover packages: %w", err)
	}

	if len(packages) == 0 {
		s.logger.Warn(ctx, "no_packages_found", "packageDir", s.packageDir)
		return []string{}, nil
	}

	// If interactive flag or terminal is interactive, prompt user
	if opts.Interactive || terminal.IsInteractive() {
		return s.selector.Select(ctx, packages)
	}

	// Install all discovered packages
	return packages, nil
}

// updateManifestRepository updates the manifest with repository information.
func (s *CloneService) updateManifestRepository(ctx context.Context, info manifest.RepositoryInfo) error {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}

	// Load existing manifest
	manifestStore := manifest.NewFSManifestStore(s.fs)
	manifestResult := manifestStore.Load(ctx, targetPathResult.Unwrap())
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}

	// Update repository info
	m := manifestResult.Unwrap()
	m.SetRepository(info)

	// Save manifest
	return manifestStore.Save(ctx, targetPathResult.Unwrap(), m)
}

// validatePackageDir checks if the package directory is suitable for cloning.
func validatePackageDir(ctx context.Context, fs FS, path string, force bool) error {
	// Check if directory exists
	exists := fs.Exists(ctx, path)
	if !exists {
		return nil // Non-existent directory is fine
	}

	// Check if it's a directory
	isDir, err := fs.IsDir(ctx, path)
	if err != nil {
		return fmt.Errorf("check packageDir: %w", err)
	}
	if !isDir {
		return ErrPackageDirNotEmpty{Path: path, Cause: fmt.Errorf("path exists but is not a directory")}
	}

	// If force flag is set, allow non-empty directory
	if force {
		return nil
	}

	// Check if directory is empty
	entries, err := fs.ReadDir(ctx, path)
	if err != nil {
		return fmt.Errorf("read packageDir: %w", err)
	}

	if len(entries) > 0 {
		return ErrPackageDirNotEmpty{Path: path}
	}

	return nil
}

// loadBootstrapConfig loads the bootstrap configuration if it exists.
func loadBootstrapConfig(ctx context.Context, fs FS, packageDir string) (bootstrap.Config, bool, error) {
	bootstrapPath := filepath.Join(packageDir, ".dotbootstrap.yaml")

	// Check if bootstrap file exists
	if !fs.Exists(ctx, bootstrapPath) {
		return bootstrap.Config{}, false, nil
	}

	// Load and parse bootstrap config
	config, err := bootstrap.Load(ctx, fs, bootstrapPath)
	if err != nil {
		return bootstrap.Config{}, false, ErrInvalidBootstrap{
			Reason: "failed to parse bootstrap configuration",
			Cause:  err,
		}
	}

	return config, true, nil
}

// selectPackagesFromProfile selects packages from a named profile.
func selectPackagesFromProfile(config bootstrap.Config, profileName string) ([]string, error) {
	packages, err := bootstrap.GetProfile(config, profileName)
	if err != nil {
		return nil, ErrProfileNotFound{Profile: profileName}
	}
	return packages, nil
}

// discoverPackages discovers package directories in the package directory.
func discoverPackages(ctx context.Context, fs FS, packageDir string) ([]string, error) {
	entries, err := fs.ReadDir(ctx, packageDir)
	if err != nil {
		return nil, fmt.Errorf("read packageDir: %w", err)
	}

	packages := make([]string, 0)
	for _, entry := range entries {
		// Only include directories, skip files and hidden directories
		if entry.IsDir() && !isHiddenFile(entry.Name()) {
			packages = append(packages, entry.Name())
		}
	}

	return packages, nil
}

// isHiddenFile checks if a filename is hidden (starts with dot).
func isHiddenFile(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// extractPackageNames extracts package names from package specs.
func extractPackageNames(packages []bootstrap.PackageSpec) []string {
	names := make([]string, len(packages))
	for i, pkg := range packages {
		names[i] = pkg.Name
	}
	return names
}

// buildRepositoryInfo constructs repository information.
func buildRepositoryInfo(url, branch, commitSHA string) manifest.RepositoryInfo {
	return manifest.RepositoryInfo{
		URL:       url,
		Branch:    branch,
		ClonedAt:  time.Now(),
		CommitSHA: commitSHA,
	}
}

// getCommitSHA attempts to get the current commit SHA from a git repository.
// Returns empty string if unable to determine (best effort).
func getCommitSHA(repoPath string) (string, error) {
	// Read the HEAD file to get current ref
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	headData, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}

	headRef := string(headData)
	// If HEAD contains a ref like "ref: refs/heads/main", extract the ref
	if len(headRef) > 5 && headRef[:5] == "ref: " {
		refPath := filepath.Join(repoPath, ".git", headRef[5:len(headRef)-1]) // Trim "ref: " and newline
		shaData, err := os.ReadFile(refPath)
		if err != nil {
			return "", err
		}
		return string(shaData[:40]), nil // SHA is first 40 chars
	}

	// HEAD directly contains SHA (detached HEAD)
	if len(headRef) >= 40 {
		return headRef[:40], nil
	}

	return "", fmt.Errorf("unable to parse HEAD")
}
