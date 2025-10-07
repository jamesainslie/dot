package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestPlanAdopt_CreatesOperations(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Setup: create file in target directory
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("set number"), 0644))

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
}

func TestPlanAdopt_GeneratesPlan(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Setup file to adopt
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".bashrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("export PATH"), 0644))

	// Create package directory
	pkgDir := filepath.Join(cfg.PackageDir, "bash")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan adopt
	plan, err := client.PlanAdopt(ctx, []string{".bashrc"}, "bash")
	require.NoError(t, err)

	// Should have operations (move + link)
	assert.NotZero(t, len(plan.Operations))
}

func TestAdopt_FileNotFound(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	pkgDir := filepath.Join(cfg.PackageDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to adopt non-existent file
	err = client.Adopt(ctx, []string{".nonexistent"}, "vim")
	assert.Error(t, err)
}

func TestAdopt_PackageNotFound(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Setup file but not package directory
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("test"), 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Try to adopt into non-existent package
	err = client.Adopt(ctx, []string{".vimrc"}, "nonexistent")
	assert.Error(t, err)
}
