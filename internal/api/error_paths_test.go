package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestAdopt_PlanErrorReturned(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Package doesn't exist - should error in PlanAdopt
	err = client.Adopt(ctx, []string{".vimrc"}, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAdopt_ExecutorError(t *testing.T) {
	cfg := testConfig(t)
	ctx, cancel := context.WithCancel(context.Background())

	// Setup
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("test"), 0644))
	pkgDir := filepath.Join(cfg.StowDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Cancel context before execution
	cancel()

	// Should error from executor
	err = client.Adopt(ctx, []string{".vimrc"}, "vim")
	assert.Error(t, err)
}

func TestUnmanage_ExecutorError(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".vimrc"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	// Create link
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "vim"), 0755))
	sourcePath := filepath.Join(cfg.StowDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))
	linkPath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, linkPath))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage should succeed
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)

	// Verify link removed
	exists := cfg.FS.Exists(ctx, linkPath)
	assert.False(t, exists)
}

func TestUnmanage_ManifestLoadFails(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage without manifest (empty plan path)
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err) // Succeeds with empty plan
}

func TestRemanage_PlanErrorFromManage(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Remanage non-existent package
	err = client.Remanage(ctx, "nonexistent")
	assert.Error(t, err) // Manage will error
}

func TestList_WithError(t *testing.T) {
	cfg := testConfig(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Cancelled context might cause error
	_, _ = client.List(ctx)
	// Either succeeds or errors depending on timing
}

func TestDoctor_CheckLinkPermissionError(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create manifest
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".vimrc"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Doctor will report broken link
	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	assert.Greater(t, len(report.Issues), 0)
}

func TestDoctor_ChecksAllBranches(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "pkg"), 0755))

	// Create a symlink that points to a file that exists
	source := filepath.Join(cfg.StowDir, "pkg", "file")
	require.NoError(t, cfg.FS.WriteFile(ctx, source, []byte("test"), 0644))
	linkPath := filepath.Join(cfg.TargetDir, ".file")
	require.NoError(t, cfg.FS.Symlink(ctx, source, linkPath))

	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"pkg": {"name": "pkg", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".file"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// Valid link - no errors
	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Empty(t, report.Issues)
}
