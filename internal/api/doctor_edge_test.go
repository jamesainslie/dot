package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestDoctor_EmptyPackageLinks(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create manifest with package but no links
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"empty": {"name": "empty", "installed_at": "2024-01-01T00:00:00Z", "link_count": 0, "links": []}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// No links to check - should be healthy
	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Equal(t, 0, report.Statistics.TotalLinks)
}

func TestDoctor_AbsoluteSymlinkPath(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "vim"), 0755))

	// Create source file
	sourcePath := filepath.Join(cfg.StowDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))

	// Create absolute symlink
	linkPath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, linkPath))

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

	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// Absolute symlink with valid target - should be healthy
	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Equal(t, 0, report.Statistics.BrokenLinks)
}

func TestDoctor_MultipleIssuesSeverities(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "vim"), 0755))

	// Create one valid link
	sourcePath := filepath.Join(cfg.StowDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))
	validLink := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, validLink))

	// Create regular file where symlink expected
	wrongFile := filepath.Join(cfg.TargetDir, ".gvimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, wrongFile, []byte("test"), 0644))

	// Create manifest
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 3, "links": [".vimrc", ".gvimrc", ".missing"]}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// Should have multiple errors
	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.Greater(t, len(report.Issues), 1)
}

func TestStatus_AllPackagesInManifest(t *testing.T) {
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

	// Get all packages (no filter)
	status, err := client.Status(ctx)
	require.NoError(t, err)

	assert.Len(t, status.Packages, 1)
	assert.Equal(t, "vim", status.Packages[0].Name)
}
