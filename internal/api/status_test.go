package api_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestStatus_WithManualManifest(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Manually create manifest without using Manage
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {
				"name": "vim",
				"installed_at": "2024-01-01T00:00:00Z",
				"link_count": 2,
				"links": [".vimrc", ".vim/colors"]
			}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Now check status
	status, err := client.Status(ctx)
	require.NoError(t, err)

	assert.Len(t, status.Packages, 1)
	assert.Equal(t, "vim", status.Packages[0].Name)
	assert.Equal(t, 2, status.Packages[0].LinkCount)
}

func TestStatus_FilteredPackages(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with multiple packages
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".vimrc"]},
			"tmux": {"name": "tmux", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".tmux.conf"]},
			"zsh": {"name": "zsh", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".zshrc"]}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Request status for only vim and tmux
	status, err := client.Status(ctx, "vim", "tmux")
	require.NoError(t, err)

	assert.Len(t, status.Packages, 2)

	names := make(map[string]bool)
	for _, pkg := range status.Packages {
		names[pkg.Name] = true
	}
	assert.True(t, names["vim"])
	assert.True(t, names["tmux"])
	assert.False(t, names["zsh"])
}

func TestStatus_PackageNotInManifest(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage vim
	err = client.Manage(context.Background(), "vim")
	require.NoError(t, err)

	// Request status for tmux (not installed)
	status, err := client.Status(context.Background(), "tmux")
	require.NoError(t, err)

	// Should return empty (tmux not in manifest)
	assert.Empty(t, status.Packages)
}

func TestList_WithMultiplePackages(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with multiple packages
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 2, "links": [".vimrc", ".vim/"]},
			"tmux": {"name": "tmux", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".tmux.conf"]},
			"zsh": {"name": "zsh", "installed_at": "2024-01-01T00:00:00Z", "link_count": 3, "links": [".zshrc", ".zsh/", ".zprofile"]}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// List all
	packages, err := client.List(ctx)
	require.NoError(t, err)

	assert.Len(t, packages, 3)

	// Verify package info
	for _, pkg := range packages {
		assert.NotEmpty(t, pkg.Name)
		assert.NotZero(t, pkg.LinkCount)
		assert.False(t, pkg.InstalledAt.IsZero())
		assert.NotEmpty(t, pkg.Links)
	}
}

func TestStatus_CorruptManifest(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create corrupt manifest
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, []byte("invalid json"), 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Should handle gracefully
	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.Empty(t, status.Packages)
}

func TestList_NoManifest(t *testing.T) {
	cfg := testConfig(t)

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	packages, err := client.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, packages)
}

func TestStatus_PreservesTimestamps(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Get status
	status1, err := client.Status(ctx)
	require.NoError(t, err)
	require.Len(t, status1.Packages, 1)

	firstInstallTime := status1.Packages[0].InstalledAt

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Get status again
	status2, err := client.Status(ctx)
	require.NoError(t, err)
	require.Len(t, status2.Packages, 1)

	// Timestamp should be the same (not updated on status query)
	assert.Equal(t, firstInstallTime, status2.Packages[0].InstalledAt)
}
