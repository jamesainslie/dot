package api_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	_ "github.com/jamesainslie/dot/internal/api" // Register implementation
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) dot.Config {
	t.Helper()
	return dot.Config{
		PackageDir: "/test/stow",
		TargetDir:  "/test/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}
}

func setupTestFixtures(t *testing.T, fs dot.FS, packages ...string) {
	t.Helper()
	ctx := context.Background()

	// Create stow directory structure
	for _, pkg := range packages {
		pkgDir := filepath.Join("/test/stow", pkg)
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))

		// Create sample dotfile
		dotfile := filepath.Join(pkgDir, "dot-config")
		require.NoError(t, fs.WriteFile(ctx, dotfile, []byte("test content"), 0644))
	}

	// Create target directory
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
}

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := testConfig(t)

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient_InvalidConfig(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "relative/path", // Invalid
	}

	client, err := dot.NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "invalid configuration")
}

func TestNewClient_AppliesDefaults(t *testing.T) {
	cfg := testConfig(t)
	// Don't set Tracer or Metrics - they should be defaulted

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify defaults applied
	resultCfg := client.Config()
	require.NotNil(t, resultCfg.Tracer)
	require.NotNil(t, resultCfg.Metrics)
}

func TestClient_Config(t *testing.T) {
	cfg := testConfig(t)

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	resultCfg := client.Config()
	require.Equal(t, cfg.PackageDir, resultCfg.PackageDir)
	require.Equal(t, cfg.TargetDir, resultCfg.TargetDir)
}
