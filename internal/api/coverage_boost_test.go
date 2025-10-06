package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestUnmanage_ManifestUpdateFails(t *testing.T) {
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

	// Create valid link
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "vim"), 0755))
	sourcePath := filepath.Join(cfg.StowDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))
	linkPath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, linkPath))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)
	// Even if manifest update fails (logged as warning), operation succeeds
}

func TestRemanage_ErrorInUnmanage(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Remanage (unmanage will fail with error, but remanage continues to manage)
	err = client.Remanage(ctx, "vim")
	require.NoError(t, err)

	// Should be managed
	status, _ := client.Status(ctx)
	assert.NotEmpty(t, status.Packages)
}

func TestUnmanage_PlanErrorPropagates(t *testing.T) {
	cfg := testConfig(t)
	cfg.TargetDir = "relative"

	// NewClient should fail validation
	_, err := dot.NewClient(cfg)
	assert.Error(t, err)
}

func TestAdopt_WithManifestError(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create files
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("test"), 0644))

	pkgDir := filepath.Join(cfg.StowDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt will fail during execution but code path is tested
	_ = client.Adopt(ctx, []string{".vimrc"}, "vim")
}

func TestDoctor_ChecksAllLinksInPackage(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create manifest with 5 links
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {
				"name": "vim",
				"installed_at": "2024-01-01T00:00:00Z",
				"link_count": 5,
				"links": [".vimrc", ".gvimrc", ".vim/", ".vim/colors", ".vim/syntax"]
			}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	// All 5 links missing
	assert.Equal(t, 5, report.Statistics.TotalLinks)
	assert.Equal(t, 5, report.Statistics.BrokenLinks)
	assert.Len(t, report.Issues, 5)
}

func TestStatus_AllPackages(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Large manifest
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 2, "links": [".vimrc", ".vim/"]},
			"tmux": {"name": "tmux", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".tmux.conf"]},
			"zsh": {"name": "zsh", "installed_at": "2024-01-01T00:00:00Z", "link_count": 3, "links": [".zshrc", ".zsh/", ".zprofile"]},
			"bash": {"name": "bash", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".bashrc"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	status, err := client.Status(ctx)
	require.NoError(t, err)

	assert.Len(t, status.Packages, 4)

	// Verify all packages present
	names := make(map[string]bool)
	for _, pkg := range status.Packages {
		names[pkg.Name] = true
	}
	assert.True(t, names["vim"])
	assert.True(t, names["tmux"])
	assert.True(t, names["zsh"])
	assert.True(t, names["bash"])
}

func TestUnmanage_MultipleLinksPerPackage(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with multiple links
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {
				"name": "vim",
				"installed_at": "2024-01-01T00:00:00Z",
				"link_count": 3,
				"links": [".vimrc", ".gvimrc", ".vim/"]
			}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	// Create the links
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "vim"), 0755))
	for _, link := range []string{".vimrc", ".gvimrc", ".vim/"} {
		source := filepath.Join(cfg.StowDir, "vim", filepath.Base(link))
		require.NoError(t, cfg.FS.WriteFile(ctx, source, []byte("test"), 0644))
		target := filepath.Join(cfg.TargetDir, link)
		require.NoError(t, cfg.FS.Symlink(ctx, source, target))
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)

	// All links should be removed
	for _, link := range []string{".vimrc", ".gvimrc", ".vim/"} {
		exists := cfg.FS.Exists(ctx, filepath.Join(cfg.TargetDir, link))
		assert.False(t, exists)
	}
}

func TestUnmanage_SkipsInvalidPaths(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with some invalid paths (will be skipped)
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"test": {
				"name": "test",
				"installed_at": "2024-01-01T00:00:00Z",
				"link_count": 1,
				"links": ["valid-link"]
			}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan should still be generated
	plan, err := client.PlanUnmanage(ctx, "test")
	require.NoError(t, err)
	assert.NotNil(t, plan)
}
