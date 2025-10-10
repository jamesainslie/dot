package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig_UsesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")

	// Create config file with custom values
	configContent := `directories:
  package: /custom/packages
  target: /custom/target
  manifest: /custom/manifest
`
	require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

	previous := globalCfg
	err := os.Setenv("DOT_CONFIG", tmpConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		globalCfg = previous
		os.Unsetenv("DOT_CONFIG")
	})

	// Set flags to defaults (should use config values)
	globalCfg = globalConfig{
		packageDir: ".",
		targetDir:  "",
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should use config file values
	assert.Contains(t, cfg.PackageDir, "/custom/packages")
	assert.Contains(t, cfg.TargetDir, "/custom/target")
	assert.Equal(t, "/custom/manifest", cfg.ManifestDir)
}

func TestBuildConfig_FlagsOverrideConfig(t *testing.T) {
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")
	flagPkgDir := tmpDir + "/flag-packages"
	flagTargetDir := tmpDir + "/flag-target"

	// Create config file
	configContent := `directories:
  package: /config/packages
  target: /config/target
`
	require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

	previous := globalCfg
	err := os.Setenv("DOT_CONFIG", tmpConfig)
	if err != nil {
		t.Fatalf("os.Setenv DOT_CONFIG=%s: %v", tmpConfig, err)
	}
	t.Cleanup(func() {
		globalCfg = previous
		os.Unsetenv("DOT_CONFIG")
	})

	// Set flags explicitly (not defaults)
	globalCfg = globalConfig{
		packageDir: flagPkgDir,
		targetDir:  flagTargetDir,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should use flag values, not config
	assert.Contains(t, cfg.PackageDir, "flag-packages")
	assert.Contains(t, cfg.TargetDir, "flag-target")
}

func TestBuildConfig_AppliesDefaults(t *testing.T) {
	tmpConfig := filepath.Join(t.TempDir(), "nonexistent.yaml")

	previous := globalCfg
	err := os.Setenv("DOT_CONFIG", tmpConfig)
	if err != nil {
		t.Fatalf("os.Setenv DOT_CONFIG=%s: %v", tmpConfig, err)
	}
	t.Cleanup(func() {
		globalCfg = previous
		os.Unsetenv("DOT_CONFIG")
	})

	globalCfg = globalConfig{
		packageDir: ".",
		targetDir:  "",
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should have defaults applied
	assert.NotEmpty(t, cfg.PackageDir)
	assert.NotEmpty(t, cfg.TargetDir)
	assert.NotNil(t, cfg.FS)
	assert.NotNil(t, cfg.Logger)
}

func TestBuildConfig_BackupDirFlag(t *testing.T) {
	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	tmpBackup := t.TempDir() + "/backups"
	globalCfg = globalConfig{
		packageDir: ".",
		targetDir:  t.TempDir(),
		backupDir:  tmpBackup,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	assert.Contains(t, cfg.BackupDir, "backups")
}
