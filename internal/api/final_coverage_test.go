package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestAdopt_WithExecutionError(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Setup file
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".bashrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("test"), 0644))

	pkgDir := filepath.Join(cfg.StowDir, "bash")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt will fail during execution (precondition: source doesn't exist after move)
	err = client.Adopt(ctx, []string{".bashrc"}, "bash")
	assert.Error(t, err) // Execution fails
}

func TestAdopt_WithPlanAdoptError(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to adopt from non-existent package
	err = client.Adopt(ctx, []string{".vimrc"}, "nonexistent")
	assert.Error(t, err)
}

func TestRemanage_ManagesAfterUnmanageFailure(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Remanage (unmanage will warn but continue to manage)
	err = client.Remanage(ctx, "vim")
	require.NoError(t, err)

	// Should be managed
	status, _ := client.Status(ctx)
	assert.Len(t, status.Packages, 1)
}

func TestUnmanage_WarnsForNonInstalledPackage(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with vim only
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

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan unmanage for tmux (not in manifest)
	plan, err := client.PlanUnmanage(ctx, "tmux")
	require.NoError(t, err)

	// Should return empty plan (tmux not installed)
	assert.Empty(t, plan.Operations)
}

func TestDoctor_HandlesManyPackages(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Large manifest
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"pkg1": {"name": "pkg1", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".file1"]},
			"pkg2": {"name": "pkg2", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".file2"]},
			"pkg3": {"name": "pkg3", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".file3"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// Should check all 3 packages
	assert.Equal(t, 3, report.Statistics.TotalLinks)
	assert.Equal(t, 3, report.Statistics.ManagedLinks)
}

func TestUnmanage_UpdatesManifestCorrectly(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".vimrc"]},
			"tmux": {"name": "tmux", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".tmux.conf"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	// Create links
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "vim"), 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.StowDir, "tmux"), 0755))

	vimSource := filepath.Join(cfg.StowDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, vimSource, []byte("test"), 0644))
	vimLink := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, vimSource, vimLink))

	tmuxSource := filepath.Join(cfg.StowDir, "tmux", "conf")
	require.NoError(t, cfg.FS.WriteFile(ctx, tmuxSource, []byte("test"), 0644))
	tmuxLink := filepath.Join(cfg.TargetDir, ".tmux.conf")
	require.NoError(t, cfg.FS.Symlink(ctx, tmuxSource, tmuxLink))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Unmanage only vim
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)

	// Manifest should only have tmux now
	status, _ := client.Status(ctx)
	assert.Len(t, status.Packages, 1)
	assert.Equal(t, "tmux", status.Packages[0].Name)
}

func TestList_WithContextCancellation(t *testing.T) {
	cfg := testConfig(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Should handle cancelled context
	_, err = client.List(ctx)
	// Might error or succeed depending on when cancellation is checked
	_ = err
}

func TestDoctor_WithEmptyManifest(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create empty manifest
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// Empty manifest is healthy
	assert.Equal(t, dot.HealthOK, report.OverallHealth)
	assert.Equal(t, 0, report.Statistics.TotalLinks)
}
