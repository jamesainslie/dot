package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestUnmanage_RemovesFromManifest(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Manually create manifest (simpler than full manage)
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

	// Create the symlink
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.PackageDir, "vim"), 0755))
	sourcePath := filepath.Join(cfg.PackageDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("test"), 0644))
	linkPath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, linkPath))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)

	// Link should be removed
	exists := cfg.FS.Exists(ctx, linkPath)
	assert.False(t, exists)

	// Manifest should be updated
	status, _ := client.Status(ctx)
	assert.Empty(t, status.Packages)
}

func TestUnmanage_WithPlanError(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage without manifest returns empty plan
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err) // Succeeds with empty plan
}

func TestUnmanage_DryRunLogsPlan(t *testing.T) {
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

	cfg.DryRun = true
	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Dry run unmanage
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err) // Dry-run just logs
}

func TestRemanage_WithManifest(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Remanage (will unmanage then manage)
	err = client.Remanage(ctx, "vim")
	require.NoError(t, err)

	// Should be managed
	status, _ := client.Status(ctx)
	assert.NotEmpty(t, status.Packages)
}

func TestRemanage_NoManifest(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Remanage without manifest (unmanage will succeed with empty plan, then manage)
	err = client.Remanage(ctx, "vim")
	require.NoError(t, err)

	// Should be managed
	status, _ := client.Status(ctx)
	assert.Len(t, status.Packages, 1)
}

func TestAdopt_PlansCorrectly(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create file to adopt
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("content"), 0644))

	// Create package directory
	pkgDir := filepath.Join(cfg.PackageDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan adopt
	plan, err := client.PlanAdopt(ctx, []string{".vimrc"}, "vim")
	require.NoError(t, err)

	// Should have 2 operations (move + link)
	assert.Len(t, plan.Operations, 2)
	assert.Equal(t, dot.OpKindFileMove, plan.Operations[0].Kind())
	assert.Equal(t, dot.OpKindLinkCreate, plan.Operations[1].Kind())
}

func TestAdopt_DryRunDoesNotModify(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create file
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".bashrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("test"), 0644))

	pkgDir := filepath.Join(cfg.PackageDir, "bash")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	cfg.DryRun = true
	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt in dry-run
	err = client.Adopt(ctx, []string{".bashrc"}, "bash")
	require.NoError(t, err)

	// File should still be regular file (not symlink)
	isLink, _ := cfg.FS.IsSymlink(ctx, filePath)
	assert.False(t, isLink)
}

func TestList_ReturnsAllPackages(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim", "tmux")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage packages
	err = client.Manage(ctx, "vim", "tmux")
	require.NoError(t, err)

	// List
	packages, err := client.List(ctx)
	require.NoError(t, err)

	assert.Len(t, packages, 2)
}

func TestStatus_ReturnsCorrectLinkCount(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with known link counts
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 5, "links": [".vimrc", ".vim/", ".gvimrc", ".vim/colors", ".vim/syntax"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	status, err := client.Status(ctx, "vim")
	require.NoError(t, err)

	assert.Len(t, status.Packages, 1)
	assert.Equal(t, 5, status.Packages[0].LinkCount)
	assert.Len(t, status.Packages[0].Links, 5)
}
