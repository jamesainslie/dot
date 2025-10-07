package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestDoctor_OrphanDetection_ScanOff(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))

	// Create an orphaned symlink
	sourcePath := filepath.Join(cfg.PackageDir, "orphan")
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.PackageDir, 0755))
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("data"), 0644))
	orphanLink := filepath.Join(cfg.TargetDir, ".orphan")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, orphanLink))

	// Create empty manifest
	manifestContent := []byte(`{"version": "1.0", "updated_at": "2024-01-01T00:00:00Z", "packages": {}, "hashes": {}}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// ScanOff should NOT detect orphans
	report, err := client.DoctorWithScan(ctx, dot.DefaultScanConfig())
	require.NoError(t, err)

	assert.Equal(t, 0, report.Statistics.OrphanedLinks)
}

func TestDoctor_OrphanDetection_Scoped(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, filepath.Join(cfg.PackageDir, "vim"), 0755))

	// Create orphaned symlink in root
	sourcePath := filepath.Join(cfg.PackageDir, "vim", "orphan")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("data"), 0644))
	orphanLink := filepath.Join(cfg.TargetDir, ".orphan")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, orphanLink))

	// Create managed link
	managedSource := filepath.Join(cfg.PackageDir, "vim", "vimrc")
	require.NoError(t, cfg.FS.WriteFile(ctx, managedSource, []byte("config"), 0644))
	managedLink := filepath.Join(cfg.TargetDir, ".vimrc")
	require.NoError(t, cfg.FS.Symlink(ctx, managedSource, managedLink))

	// Create manifest with managed link
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

	// Scoped scan should detect orphan in scanned directory
	report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())
	require.NoError(t, err)

	assert.Equal(t, 1, report.Statistics.OrphanedLinks)
	assert.Equal(t, dot.HealthWarnings, report.OverallHealth)
}

func TestDoctor_OrphanDetection_Deep(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create nested directory structure
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	nestedDir := filepath.Join(cfg.TargetDir, ".config", "app")
	require.NoError(t, cfg.FS.MkdirAll(ctx, nestedDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.PackageDir, 0755))

	// Create orphaned link in nested directory
	sourcePath := filepath.Join(cfg.PackageDir, "orphan-file")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("data"), 0644))
	orphanLink := filepath.Join(nestedDir, "config.json")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, orphanLink))

	// Create empty manifest
	manifestContent := []byte(`{"version": "1.0", "updated_at": "2024-01-01T00:00:00Z", "packages": {}, "hashes": {}}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Deep scan should find nested orphan
	report, err := client.DoctorWithScan(ctx, dot.DeepScanConfig(5))
	require.NoError(t, err)

	assert.Greater(t, report.Statistics.OrphanedLinks, 0)
}

func TestDoctor_OrphanDetection_SkipsGit(t *testing.T) {
	cfg := testConfig(t)
	ctx := context.Background()

	// Create .git directory with symlink
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.TargetDir, 0755))
	gitDir := filepath.Join(cfg.TargetDir, ".git", "hooks")
	require.NoError(t, cfg.FS.MkdirAll(ctx, gitDir, 0755))
	require.NoError(t, cfg.FS.MkdirAll(ctx, cfg.PackageDir, 0755))

	// Create symlink in .git (should be skipped)
	sourcePath := filepath.Join(cfg.PackageDir, "hook")
	require.NoError(t, cfg.FS.WriteFile(ctx, sourcePath, []byte("hook"), 0755))
	hookLink := filepath.Join(gitDir, "pre-commit")
	require.NoError(t, cfg.FS.Symlink(ctx, sourcePath, hookLink))

	// Empty manifest
	manifestContent := []byte(`{"version": "1.0", "updated_at": "2024-01-01T00:00:00Z", "packages": {}, "hashes": {}}`)
	manifestPath := filepath.Join(cfg.TargetDir, ".dot-manifest.json")
	require.NoError(t, cfg.FS.WriteFile(ctx, manifestPath, manifestContent, 0644))

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Deep scan should skip .git directory
	report, err := client.DoctorWithScan(ctx, dot.DeepScanConfig(10))
	require.NoError(t, err)

	// Should not detect link in .git (skip patterns working)
	// Note: If this detects the link, skip logic may need refinement
	assert.Equal(t, 0, report.Statistics.OrphanedLinks, "links in .git should be skipped")
}
