package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestPlanRemanage_CombinesPlans(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan remanage (will create manage plan since no manifest)
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)

	// Should have operations
	assert.NotZero(t, len(plan.Operations))
}

func TestPlanRemanage_GeneratesPlan(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage first
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Plan remanage without changes
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)

	// With incremental remanage, unchanged packages should have no operations
	assert.NotNil(t, plan)
	assert.Equal(t, 0, len(plan.Operations), "unchanged package should have no operations")
}

func TestPlanRemanage_CombinesUnmanageAndManage(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	// Create manifest
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	manifestContent := []byte(`{
		"version": "1.0",
		"updated_at": "2024-01-01T00:00:00Z",
		"packages": {
			"vim": {"name": "vim", "installed_at": "2024-01-01T00:00:00Z", "link_count": 1, "links": [".config"]}
		},
		"hashes": {}
	}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan remanage
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)

	// Should have both unmanage and manage operations
	assert.NotZero(t, len(plan.Operations))
}

func TestPlanRemanage_NoManifest(t *testing.T) {
	cfg := testConfig(t)
	setupTestFixtures(t, cfg.FS, "vim")
	ctx := context.Background()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan remanage without manifest - should fall back to plan manage
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)
	assert.NotZero(t, len(plan.Operations))
}
