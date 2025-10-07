package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	_ "github.com/jamesainslie/dot/internal/api" // Register implementation
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifest_AccurateLinksPerPackage(t *testing.T) {
	// Setup test directories
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package with 3 files
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("set nocompatible"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-colors"), []byte("color theme"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-plugins"), []byte("plugins"), 0644))

	// Create zsh package with 2 files
	zshPackage := filepath.Join(packageDir, "zsh")
	require.NoError(t, os.MkdirAll(zshPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(zshPackage, "dot-zshrc"), []byte("zsh config"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(zshPackage, "dot-zshenv"), []byte("zsh env"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Manage vim package
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Load manifest and verify vim package
	manifestStore := manifest.NewFSManifestStore(adapters.NewOSFilesystem())
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())

	manifestResult := manifestStore.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())

	m := manifestResult.Unwrap()

	// Check vim package
	vimInfo, hasVim := m.GetPackage("vim")
	require.True(t, hasVim, "vim package should be in manifest")
	assert.Equal(t, 3, vimInfo.LinkCount, "vim should have 3 links")
	assert.Len(t, vimInfo.Links, 3, "vim Links array should have 3 entries")
	assert.Contains(t, vimInfo.Links, ".vimrc")
	assert.Contains(t, vimInfo.Links, ".vim-colors")
	assert.Contains(t, vimInfo.Links, ".vim-plugins")

	// Manage zsh package
	err = client.Manage(ctx, "zsh")
	require.NoError(t, err)

	// Reload manifest and verify both packages
	manifestResult = manifestStore.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m = manifestResult.Unwrap()

	// Check vim package (should still be there)
	vimInfo, hasVim = m.GetPackage("vim")
	require.True(t, hasVim)
	assert.Equal(t, 3, vimInfo.LinkCount)

	// Check zsh package
	zshInfo, hasZsh := m.GetPackage("zsh")
	require.True(t, hasZsh, "zsh package should be in manifest")
	assert.Equal(t, 2, zshInfo.LinkCount, "zsh should have 2 links")
	assert.Len(t, zshInfo.Links, 2, "zsh Links array should have 2 entries")
	assert.Contains(t, zshInfo.Links, ".zshrc")
	assert.Contains(t, zshInfo.Links, ".zshenv")
}

func TestManifest_MultiplePackagesSingleCommand(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create packages
	for _, pkg := range []string{"vim", "zsh", "git"} {
		pkgPath := filepath.Join(packageDir, pkg)
		require.NoError(t, os.MkdirAll(pkgPath, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(pkgPath, "dot-"+pkg+"rc"), []byte("config"), 0644))
	}

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Manage all packages at once
	err = client.Manage(ctx, "vim", "zsh", "git")
	require.NoError(t, err)

	// Load manifest
	manifestStore := manifest.NewFSManifestStore(adapters.NewOSFilesystem())
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())

	manifestResult := manifestStore.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m := manifestResult.Unwrap()

	// Verify all packages tracked correctly
	for _, pkg := range []string{"vim", "zsh", "git"} {
		info, has := m.GetPackage(pkg)
		require.True(t, has, "%s package should be in manifest", pkg)
		assert.Equal(t, 1, info.LinkCount, "%s should have 1 link", pkg)
		assert.Len(t, info.Links, 1, "%s Links array should have 1 entry", pkg)
		assert.Contains(t, info.Links, "."+pkg+"rc")
	}
}

func TestManifest_RemanageUpdatesLinks(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package with 2 files initially
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("config"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-colors"), []byte("colors"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial manage
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Verify initial state
	manifestStore := manifest.NewFSManifestStore(adapters.NewOSFilesystem())
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())

	manifestResult := manifestStore.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m := manifestResult.Unwrap()

	vimInfo, _ := m.GetPackage("vim")
	assert.Equal(t, 2, vimInfo.LinkCount)

	// Add a new file to vim package
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-plugins"), []byte("plugins"), 0644))

	// Remanage
	err = client.Remanage(ctx, "vim")
	require.NoError(t, err)

	// Verify updated state
	manifestResult = manifestStore.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m = manifestResult.Unwrap()

	vimInfo, has := m.GetPackage("vim")
	require.True(t, has)
	assert.Equal(t, 3, vimInfo.LinkCount, "vim should now have 3 links after remanage")
	assert.Len(t, vimInfo.Links, 3)
	assert.Contains(t, vimInfo.Links, ".vimrc")
	assert.Contains(t, vimInfo.Links, ".vim-colors")
	assert.Contains(t, vimInfo.Links, ".vim-plugins")
}

func TestManifest_StatusShowsCorrectCounts(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create packages with different file counts
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-colors"), []byte("2"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vim-plugins"), []byte("3"), 0644))

	zshPackage := filepath.Join(packageDir, "zsh")
	require.NoError(t, os.MkdirAll(zshPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(zshPackage, "dot-zshrc"), []byte("1"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Manage packages
	err = client.Manage(ctx, "vim", "zsh")
	require.NoError(t, err)

	// Get status
	status, err := client.Status(ctx)
	require.NoError(t, err)

	// Verify counts
	require.Len(t, status.Packages, 2)

	// Find vim in status
	var vimStatus *dot.PackageInfo
	for i := range status.Packages {
		if status.Packages[i].Name == "vim" {
			vimStatus = &status.Packages[i]
			break
		}
	}
	require.NotNil(t, vimStatus, "vim should be in status")
	assert.Equal(t, 3, vimStatus.LinkCount, "vim status should show 3 links")

	// Find zsh in status
	var zshStatus *dot.PackageInfo
	for i := range status.Packages {
		if status.Packages[i].Name == "zsh" {
			zshStatus = &status.Packages[i]
			break
		}
	}
	require.NotNil(t, zshStatus, "zsh should be in status")
	assert.Equal(t, 1, zshStatus.LinkCount, "zsh status should show 1 link")
}

