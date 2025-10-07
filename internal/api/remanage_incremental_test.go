package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	_ "github.com/jamesainslie/dot/internal/api" // Register implementation
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemanage_Incremental_NoChanges(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("set nocompatible"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial manage
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Remanage without changes
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)

	// Should have no operations (package unchanged)
	assert.Equal(t, 0, len(plan.Operations), "should have no operations when package unchanged")
}

func TestRemanage_Incremental_WithChanges(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("original"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial manage
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Modify the package file
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("modified"), 0644))

	// Remanage with changes
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)

	// Should have operations (package changed)
	assert.Greater(t, len(plan.Operations), 0, "should have operations when package changed")
}

func TestRemanage_Incremental_NewPackage(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package (but don't install it)
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vimPackage, "dot-vimrc"), []byte("config"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Remanage package that's not installed yet
	plan, err := client.PlanRemanage(ctx, "vim")
	require.NoError(t, err)

	// Should plan as new install
	assert.Greater(t, len(plan.Operations), 0, "should have operations for new package")
	assert.True(t, plan.HasPackage("vim"), "plan should include vim package")
}

func TestRemanage_Incremental_Execute(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create vim package
	vimPackage := filepath.Join(packageDir, "vim")
	require.NoError(t, os.MkdirAll(vimPackage, 0755))
	vimrcPath := filepath.Join(vimPackage, "dot-vimrc")
	require.NoError(t, os.WriteFile(vimrcPath, []byte("version1"), 0644))

	// Create client
	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial manage
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	linkPath := filepath.Join(targetDir, ".vimrc")
	require.FileExists(t, linkPath)

	// Verify link content
	target, err := os.Readlink(linkPath)
	require.NoError(t, err)
	content1, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "version1", string(content1))

	// Modify package file
	require.NoError(t, os.WriteFile(vimrcPath, []byte("version2"), 0644))

	// Remanage (should detect change and update)
	err = client.Remanage(ctx, "vim")
	require.NoError(t, err)

	// Verify link still exists and points to updated content
	require.FileExists(t, linkPath)
	target, err = os.Readlink(linkPath)
	require.NoError(t, err)
	content2, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "version2", string(content2), "link should point to updated content")
}

