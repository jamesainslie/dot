package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestAdopt_InvalidStowDir(t *testing.T) {
	cfg := testConfig(t)
	cfg.StowDir = "relative/path"

	// NewClient should fail validation
	_, err := dot.NewClient(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")
}

func TestAdopt_InvalidTargetDir(t *testing.T) {
	cfg := testConfig(t)
	cfg.TargetDir = "relative/path"

	// NewClient should fail validation
	_, err := dot.NewClient(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")
}

func TestAdopt_DryRunMode(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Setup file
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	filePath := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, filePath, []byte("test"), 0644))

	pkgDir := filepath.Join(cfg.StowDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	cfg.DryRun = true
	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Dry-run adopt
	err = client.Adopt(ctx, []string{".vimrc"}, "vim")
	require.NoError(t, err)

	// File should not be moved
	exists := cfg.FS.Exists(ctx, filePath)
	assert.True(t, exists)

	// Should not be symlink
	isLink, _ := cfg.FS.IsSymlink(ctx, filePath)
	assert.False(t, isLink)
}

func TestAdopt_EmptyFileList(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	pkgDir := filepath.Join(cfg.StowDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt with no files
	err = client.Adopt(ctx, []string{}, "vim")
	require.NoError(t, err) // Should succeed with nothing to do
}

func TestAdopt_MultipleFilesPlan(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create multiple files
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	files := []string{".vimrc", ".gvimrc", ".vim/colors"}
	for _, file := range files {
		dir := filepath.Dir(filepath.Join(cfg.TargetDir, file))
		require.NoError(t, cfg.FS.MkdirAll(ctx, dir, 0755))
		path := filepath.Join(cfg.TargetDir, file)
		require.NoError(t, cfg.FS.WriteFile(ctx, path, []byte("test"), 0644))
	}

	pkgDir := filepath.Join(cfg.StowDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan adopt for multiple files
	plan, err := client.PlanAdopt(ctx, files, "vim")
	require.NoError(t, err)

	// Should have 2 operations per file (move + link)
	assert.Equal(t, 6, len(plan.Operations))
}

func TestAdopt_InvalidFilePath(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	pkgDir := filepath.Join(cfg.StowDir, "vim")
	require.NoError(t, cfg.FS.MkdirAll(ctx, pkgDir, 0755))

	// Create file with problematic path
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.WriteFile(ctx, filepath.Join(cfg.TargetDir, "file"), []byte("test"), 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Plan adopt
	plan, err := client.PlanAdopt(ctx, []string{"file"}, "vim")
	require.NoError(t, err)
	assert.NotZero(t, len(plan.Operations))
}
