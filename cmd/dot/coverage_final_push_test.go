package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig_WithManifestDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")
	manifestDir := filepath.Join(tmpDir, "manifest")

	configContent := `directories:
  package: ` + tmpDir + `/packages
  target: ` + tmpDir + `/target
  manifest: ` + manifestDir + `
`
	require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))
	require.NoError(t, os.MkdirAll(tmpDir+"/packages", 0755))
	require.NoError(t, os.MkdirAll(tmpDir+"/target", 0755))

	previous := globalCfg
	os.Setenv("DOT_CONFIG", tmpConfig)
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

	assert.Equal(t, manifestDir, cfg.ManifestDir)
}

func TestBuildConfig_WithBackupDir(t *testing.T) {
	tmpDir := t.TempDir()

	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	backupDir := filepath.Join(tmpDir, "backups")
	globalCfg = globalConfig{
		packageDir: tmpDir,
		targetDir:  tmpDir,
		backupDir:  backupDir,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	assert.Equal(t, backupDir, cfg.BackupDir)
}

func TestBuildConfig_FlagPrecedence_PackageDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")
	configPackageDir := filepath.Join(tmpDir, "config-packages")
	flagPackageDir := filepath.Join(tmpDir, "flag-packages")

	configContent := `directories:
  package: ` + configPackageDir + `
  target: ` + tmpDir + `
`
	require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))
	require.NoError(t, os.MkdirAll(configPackageDir, 0755))
	require.NoError(t, os.MkdirAll(flagPackageDir, 0755))

	previous := globalCfg
	os.Setenv("DOT_CONFIG", tmpConfig)
	t.Cleanup(func() {
		globalCfg = previous
		os.Unsetenv("DOT_CONFIG")
	})

	// Set explicit flag value (not default ".")
	globalCfg = globalConfig{
		packageDir: flagPackageDir,
		targetDir:  tmpDir,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Flag should override config
	assert.Contains(t, cfg.PackageDir, "flag-packages")
}

func TestIsHiddenOrIgnored_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{".git", true},
		{".svn", true},
		{"node_modules", true},
		{"vendor", true},
		{"normal", false},
		{"", true},
	}

	for _, tt := range tests {
		result := isHiddenOrIgnored(tt.name)
		assert.Equal(t, tt.expected, result, "Name: %s", tt.name)
	}
}

func TestPackageCompletion_BothModes(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "pkg1"), 0755))

	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: tmpDir,
		targetDir:  tmpDir,
	}

	// Test available packages
	availableFn := packageCompletion(false)
	available, _ := availableFn(nil, []string{}, "")
	assert.Contains(t, available, "pkg1")

	// Test installed packages
	installedFn := packageCompletion(true)
	installed, _ := installedFn(nil, []string{}, "")
	assert.NotNil(t, installed) // May be empty but should not panic
}
