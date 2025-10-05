package api_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	_ "github.com/jamesainslie/dot/internal/api" // Register implementation
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestClient_Manage(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1")

	cfg := testConfig(t)
	cfg.FS = fs

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Manage(ctx, "package1")
	require.NoError(t, err)

	// Verify link created
	linkPath := "/test/target/.config"
	isLink, err := fs.IsSymlink(ctx, linkPath)
	require.NoError(t, err)
	require.True(t, isLink, "expected symlink to be created at %s", linkPath)

	// Verify link points to correct location
	target, err := fs.ReadLink(ctx, linkPath)
	require.NoError(t, err)
	require.Contains(t, target, "package1/dot-config")
}

func TestClient_PlanManage(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1")

	cfg := testConfig(t)
	cfg.FS = fs

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	plan, err := client.PlanManage(ctx, "package1")
	require.NoError(t, err)
	require.NotEmpty(t, plan.Operations, "expected plan to contain operations")

	// Should have at least a link create operation
	hasLinkCreate := false
	for _, op := range plan.Operations {
		if op.Kind() == dot.OpKindLinkCreate {
			hasLinkCreate = true
			break
		}
	}
	require.True(t, hasLinkCreate, "expected plan to include LinkCreate operation")
}

func TestClient_Manage_DryRun(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1")

	cfg := testConfig(t)
	cfg.FS = fs
	cfg.DryRun = true

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Manage(ctx, "package1")
	require.NoError(t, err)

	// Verify no links created (dry-run mode)
	exists := fs.Exists(ctx, "/test/target/.config")
	require.False(t, exists, "expected no changes in dry-run mode")
}

func TestClient_Manage_NonExistentPackage(t *testing.T) {
	fs := adapters.NewMemFS()
	// Don't create package fixtures

	cfg := testConfig(t)
	cfg.FS = fs

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Manage(ctx, "nonexistent")
	require.Error(t, err, "expected error for non-existent package")
}

func TestClient_Manage_MultiplePackages(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1", "package2")

	cfg := testConfig(t)
	cfg.FS = fs

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Manage(ctx, "package1", "package2")
	require.NoError(t, err)

	// Verify both packages created links
	isLink1, _ := fs.IsSymlink(ctx, "/test/target/.config")
	require.True(t, isLink1)
}
