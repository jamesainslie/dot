package dot

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/bootstrap"
	"github.com/jamesainslie/dot/internal/cli/selector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCloneService(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	manageSvc := &ManageService{}
	cloner := adapters.NewGoGitCloner()
	sel := selector.NewInteractiveSelector(os.Stdin, os.Stdout)

	svc := newCloneService(fs, logger, manageSvc, cloner, sel, "/packages", "/home", false)

	assert.NotNil(t, svc)
	assert.Equal(t, "/packages", svc.packageDir)
	assert.Equal(t, "/home", svc.targetDir)
}

func TestCloneService_ValidatePackageDir(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	t.Run("empty directory is valid", func(t *testing.T) {
		err := fs.MkdirAll(ctx, "/packages", 0755)
		require.NoError(t, err)

		err = validatePackageDir(ctx, fs, "/packages", false)
		assert.NoError(t, err)
	})

	t.Run("non-existent directory is valid", func(t *testing.T) {
		err := validatePackageDir(ctx, fs, "/nonexistent", false)
		assert.NoError(t, err)
	})

	t.Run("directory with files fails", func(t *testing.T) {
		err := fs.MkdirAll(ctx, "/packages2", 0755)
		require.NoError(t, err)
		err = fs.WriteFile(ctx, "/packages2/file.txt", []byte("test"), 0644)
		require.NoError(t, err)

		err = validatePackageDir(ctx, fs, "/packages2", false)
		assert.Error(t, err)
		assert.IsType(t, ErrPackageDirNotEmpty{}, err)
	})

	t.Run("force flag allows non-empty directory", func(t *testing.T) {
		err := fs.MkdirAll(ctx, "/packages3", 0755)
		require.NoError(t, err)
		err = fs.WriteFile(ctx, "/packages3/file.txt", []byte("test"), 0644)
		require.NoError(t, err)

		err = validatePackageDir(ctx, fs, "/packages3", true)
		assert.NoError(t, err)
	})
}

func TestCloneService_SelectPackages_WithProfile(t *testing.T) {
	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim", "dot-zsh"},
			},
		},
	}

	packages, err := selectPackagesFromProfile(config, "minimal")
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, packages)
}

func TestCloneService_SelectPackages_ProfileNotFound(t *testing.T) {
	config := bootstrap.Config{
		Version:  "1.0",
		Packages: []bootstrap.PackageSpec{{Name: "dot-vim"}},
	}

	_, err := selectPackagesFromProfile(config, "nonexistent")
	assert.Error(t, err)
	assert.IsType(t, ErrProfileNotFound{}, err)
}

func TestCloneService_SelectPackages_Interactive(t *testing.T) {
	input := strings.NewReader("1,2\n")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	packages := []string{"dot-vim", "dot-zsh", "dot-tmux"}
	selected, err := sel.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, selected)
}

func TestCloneService_FilterByPlatform(t *testing.T) {
	packages := []bootstrap.PackageSpec{
		{Name: "all-platforms"},
		{Name: "linux-only", Platform: []string{"linux"}},
		{Name: "darwin-only", Platform: []string{"darwin"}},
	}

	currentPlatform := runtime.GOOS

	filtered := bootstrap.FilterPackagesByPlatform(packages, currentPlatform)

	// Verify "all-platforms" is always included
	names := make([]string, 0, len(filtered))
	for _, p := range filtered {
		names = append(names, p.Name)
	}
	assert.Contains(t, names, "all-platforms")

	// Verify platform-specific packages are filtered correctly
	if currentPlatform == "linux" {
		assert.Contains(t, names, "linux-only")
		assert.NotContains(t, names, "darwin-only")
	} else if currentPlatform == "darwin" {
		assert.Contains(t, names, "darwin-only")
		assert.NotContains(t, names, "linux-only")
	}
}

func TestCloneService_BuildRepositoryInfo(t *testing.T) {
	url := "https://github.com/user/dotfiles"
	branch := "main"
	beforeClone := time.Now()

	info := buildRepositoryInfo(url, branch, "abc123def456")

	assert.Equal(t, url, info.URL)
	assert.Equal(t, branch, info.Branch)
	assert.Equal(t, "abc123def456", info.CommitSHA)
	assert.True(t, info.ClonedAt.After(beforeClone.Add(-time.Second)))
	assert.True(t, info.ClonedAt.Before(time.Now().Add(time.Second)))
}

func TestCloneService_ExtractPackageNames(t *testing.T) {
	packages := []bootstrap.PackageSpec{
		{Name: "dot-vim"},
		{Name: "dot-zsh"},
		{Name: "dot-tmux"},
	}

	names := extractPackageNames(packages)
	assert.Equal(t, []string{"dot-vim", "dot-zsh", "dot-tmux"}, names)
}

func TestCloneService_LoadBootstrapConfig_Found(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory first
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Create bootstrap config
	configContent := `version: "1.0"
packages:
  - name: dot-vim
    required: true
`
	err = fs.WriteFile(ctx, "/packages/.dotbootstrap.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	config, found, err := loadBootstrapConfig(ctx, fs, "/packages")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "1.0", config.Version)
	assert.Len(t, config.Packages, 1)
}

func TestCloneService_LoadBootstrapConfig_NotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	config, found, err := loadBootstrapConfig(ctx, fs, "/packages")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, bootstrap.Config{}, config)
}

func TestCloneService_LoadBootstrapConfig_Invalid(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory first
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Create invalid config
	invalidConfig := `this is not valid yaml: [unclosed`
	err = fs.WriteFile(ctx, "/packages/.dotbootstrap.yaml", []byte(invalidConfig), 0644)
	require.NoError(t, err)

	_, _, err = loadBootstrapConfig(ctx, fs, "/packages")
	assert.Error(t, err)
	assert.IsType(t, ErrInvalidBootstrap{}, err)
}

func TestCloneService_DiscoverPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directories
	err := fs.MkdirAll(ctx, "/packages/dot-vim", 0755)
	require.NoError(t, err)
	err = fs.MkdirAll(ctx, "/packages/dot-zsh", 0755)
	require.NoError(t, err)
	err = fs.WriteFile(ctx, "/packages/README.md", []byte("test"), 0644)
	require.NoError(t, err)

	packages, err := discoverPackages(ctx, fs, "/packages")
	require.NoError(t, err)

	// Should only find directories, not files
	assert.Contains(t, packages, "dot-vim")
	assert.Contains(t, packages, "dot-zsh")
	assert.NotContains(t, packages, "README.md")
}

func TestCloneService_SelectPackagesWithBootstrap_DefaultProfile(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Defaults: bootstrap.Defaults{
			Profile: "minimal",
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim", "dot-zsh"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{})
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, packages)
}

func TestCloneService_SelectPackagesWithBootstrap_ExplicitProfile(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{Profile: "minimal"})
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-vim"}, packages)
}

func TestCloneService_SelectPackagesWithoutBootstrap_AllPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create package directories
	err := fs.MkdirAll(ctx, "/packages/dot-vim", 0755)
	require.NoError(t, err)
	err = fs.MkdirAll(ctx, "/packages/dot-zsh", 0755)
	require.NoError(t, err)

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	// Non-interactive should install all
	packages, err := svc.selectPackagesWithoutBootstrap(ctx, CloneOptions{})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dot-vim", "dot-zsh"}, packages)
}

func TestCloneService_SelectPackagesWithoutBootstrap_NoPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create empty directory
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithoutBootstrap(ctx, CloneOptions{})
	require.NoError(t, err)
	assert.Empty(t, packages)
}

func TestCloneOptions_Defaults(t *testing.T) {
	opts := CloneOptions{}

	assert.Empty(t, opts.Profile)
	assert.False(t, opts.Interactive)
	assert.False(t, opts.Force)
	assert.Empty(t, opts.Branch)
}

func TestCloneOptions_WithValues(t *testing.T) {
	opts := CloneOptions{
		Profile:     "minimal",
		Interactive: true,
		Force:       true,
		Branch:      "develop",
	}

	assert.Equal(t, "minimal", opts.Profile)
	assert.True(t, opts.Interactive)
	assert.True(t, opts.Force)
	assert.Equal(t, "develop", opts.Branch)
}

func TestCloneService_Clone_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create managed package directories to simulate a successful manage operation
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Mock git cloner
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			// Simulate cloning by creating a package directory
			return fs.MkdirAll(ctx, dest+"/dot-vim", 0755)
		},
	}

	// Mock selector
	selector := &mockPackageSelector{
		selectFn: func(ctx context.Context, packages []string) ([]string, error) {
			return []string{"dot-vim"}, nil
		},
	}

	// Create a simple ManageService that doesn't do anything
	manageSvc := &ManageService{
		fs:         fs,
		logger:     logger,
		packageDir: "/packages",
		targetDir:  "/home",
		dryRun:     true, // Dry run to avoid actual file operations
	}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", true)

	err = svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{
		Branch: "main",
	})

	require.NoError(t, err)
}

func TestCloneService_Clone_PackageDirNotEmpty(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create non-empty package directory
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.WriteFile(ctx, "/packages/existing-file.txt", []byte("test"), 0644)
	require.NoError(t, err)

	cloner := &mockGitCloner{}
	selector := &mockPackageSelector{}
	manageSvc := &ManageService{}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", false)

	err = svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{})

	assert.Error(t, err)
	assert.IsType(t, ErrPackageDirNotEmpty{}, err)
}

func TestCloneService_Clone_CloneFails(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Mock git cloner that fails
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			return assert.AnError
		},
	}

	selector := &mockPackageSelector{}
	manageSvc := &ManageService{}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", false)

	err := svc.Clone(ctx, "https://github.com/user/invalid", CloneOptions{})

	assert.Error(t, err)
	assert.IsType(t, ErrCloneFailed{}, err)
}

func TestCloneService_Clone_WithBootstrap(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create package directory
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Mock git cloner that creates bootstrap config
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			// Create package directories
			_ = fs.MkdirAll(ctx, dest+"/dot-vim", 0755)
			_ = fs.MkdirAll(ctx, dest+"/dot-zsh", 0755)

			// Create bootstrap config
			bootstrapContent := `version: "1.0"
packages:
  - name: dot-vim
    required: true
  - name: dot-zsh
    required: false
profiles:
  minimal:
    description: "Minimal setup"
    packages:
      - dot-vim
`
			return fs.WriteFile(ctx, dest+"/.dotbootstrap.yaml", []byte(bootstrapContent), 0644)
		},
	}

	// Mock selector (shouldn't be called because profile is specified)
	selector := &mockPackageSelector{}

	manageSvc := &ManageService{
		fs:         fs,
		logger:     logger,
		packageDir: "/packages",
		targetDir:  "/home",
		dryRun:     true,
	}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", true)

	err = svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{
		Profile: "minimal",
		Branch:  "main",
	})

	require.NoError(t, err)
}

func TestCloneService_Clone_NoPackagesSelected(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Mock git cloner
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			// Create empty clone
			return fs.MkdirAll(ctx, dest, 0755)
		},
	}

	// Mock selector that returns no packages
	selector := &mockPackageSelector{
		selectFn: func(ctx context.Context, packages []string) ([]string, error) {
			return []string{}, nil
		},
	}

	manageSvc := &ManageService{}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", false)

	err := svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{
		Interactive: true,
	})

	require.NoError(t, err) // Should succeed even with no packages
}

func TestCloneService_GetCommitSHA(t *testing.T) {
	t.Skip("getCommitSHA requires git repository - tested in integration tests")
}
