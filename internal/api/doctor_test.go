package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestDoctor_NoManifest(t *testing.T) {
	cfg := testConfig(t)
	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(context.Background(), dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Equal(t, 0, report.Statistics.TotalLinks)
}

func TestDoctor_WithValidLinks(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest and actual symlinks
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.PackageDir, "vim"), 0755))

	// Create source file
	sourcePath := filepath.Join(cfg.PackageDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))

	// Create symlink
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

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Equal(t, 1, report.Statistics.TotalLinks)
	assert.Equal(t, 1, report.Statistics.ManagedLinks)
	assert.Equal(t, 0, report.Statistics.BrokenLinks)
	assert.Empty(t, report.Issues)
}

func TestDoctor_BrokenLink_DoesNotExist(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create manifest with link that doesn't exist
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

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.Equal(t, 1, report.Statistics.BrokenLinks)
	assert.NotEmpty(t, report.Issues)
	assert.Equal(t, dot.IssueBrokenLink, report.Issues[0].Type)
}

func TestDoctor_NotSymlink(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create regular file where symlink expected
	linkPath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, linkPath, []byte("test"), 0644))

	// Create manifest expecting it to be a symlink
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

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.NotEmpty(t, report.Issues)
	assert.Equal(t, dot.IssueWrongTarget, report.Issues[0].Type)
}

func TestDoctor_BrokenSymlinkTarget(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create symlink pointing to non-existent target
	linkPath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, "/nonexistent/file", linkPath))

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

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.Equal(t, 1, report.Statistics.BrokenLinks)
	assert.NotEmpty(t, report.Issues)
}

func TestDoctor_MultiplePackages(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create manifest with multiple packages
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 2, "links": [".vimrc", ".vim/"]},
			"tmux": {"name": "tmux", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".tmux.conf"]},
			"zsh": {"name": "zsh", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".zshrc"]}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	// All links missing, should have errors
	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.Equal(t, 4, report.Statistics.TotalLinks)
	assert.Equal(t, 4, report.Statistics.ManagedLinks)
	assert.Greater(t, report.Statistics.BrokenLinks, 0)
}

func TestDoctor_MixedHealthyAndBroken(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.PackageDir, "vim"), 0755))

	// Create one valid link
	sourcePath := filepath.Join(cfg.PackageDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))
	linkPath1 := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, linkPath1))

	// Create manifest with 2 links (one valid, one broken)
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 2, "links": [".vimrc", ".gvimrc"]}
		},
		"hashes": {}
	}`)

	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.Equal(t, 2, report.Statistics.TotalLinks)
	assert.Equal(t, 1, report.Statistics.BrokenLinks) // .gvimrc is missing
}

func TestDoctor_InvalidTargetPath(t *testing.T) {
	cfg := testConfig(t)
	cfg.TargetDir = "relative/path" // Invalid

	// NewClient should fail validation
	_, err := dot.NewClient(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")
}

func TestStatus_NoPackages(t *testing.T) {
	cfg := testConfig(t)
	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	status, err := client.Status(context.Background())
	require.NoError(t, err)
	assert.Empty(t, status.Packages)
}

func TestList_Empty(t *testing.T) {
	cfg := testConfig(t)
	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	packages, err := client.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, packages)
}
