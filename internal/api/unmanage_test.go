package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestPlanUnmanage_WithManifest(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest with package
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 2, "links": [".vimrc", ".vim/"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan unmanage
	plan, err := client.PlanUnmanage(ctx, "vim")
	require.NoError(t, err)

	// Should have 2 delete operations
	assert.Len(t, plan.Operations, 2)
	assert.Equal(t, dot.OpKindLinkDelete, plan.Operations[0].Kind())
}

func TestUnmanage_NotInstalled(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to unmanage package that was never installed (no manifest)
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err) // Succeeds with empty plan
}

func TestPlanUnmanage_NoManifest(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to plan unmanage without manifest - should return empty plan
	plan, err := client.PlanUnmanage(ctx, "vim")
	require.NoError(t, err)
	assert.Empty(t, plan.Operations)
}

func TestPlanUnmanage_PackageNotInstalled(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create manifest without the requested package
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"tmux": {"name": "tmux", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".tmux.conf"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to unmanage package not in manifest
	plan, err := client.PlanUnmanage(ctx, "vim")
	require.NoError(t, err)

	// Should have no operations (package wasn't installed)
	assert.Empty(t, plan.Operations)
}
