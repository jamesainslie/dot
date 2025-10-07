package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCommand_Init(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set DOT_CONFIG to use temp directory
	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Run init
	err := runConfigInit(false, "yaml")
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, configPath)

	// Verify file has correct permissions
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content is valid
	cfg, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestConfigCommand_InitForce(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Create initial config
	err := runConfigInit(false, "yaml")
	require.NoError(t, err)

	// Try to init again without force - should fail
	err = runConfigInit(false, "yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Init with force - should succeed
	err = runConfigInit(true, "yaml")
	assert.NoError(t, err)
}

func TestConfigCommand_Get(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/dotfiles"
	cfg.Logging.Level = "DEBUG"

	writer := config.NewWriter(configPath)
	err := writer.Write(cfg, config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Test get
	value, err := getConfigValue(cfg, "directories.package")
	require.NoError(t, err)
	assert.Equal(t, "/test/dotfiles", value)

	value, err = getConfigValue(cfg, "logging.level")
	require.NoError(t, err)
	assert.Equal(t, "DEBUG", value)
}

func TestConfigCommand_GetUnknownKey(t *testing.T) {
	cfg := config.DefaultExtended()

	_, err := getConfigValue(cfg, "invalid.key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigCommand_Set(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Create initial config
	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	// Set a value
	err = runConfigSet("directories.package", "/new/dotfiles")
	require.NoError(t, err)

	// Verify value was set
	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, "/new/dotfiles", cfg.Directories.Package)
}

func TestConfigCommand_Path(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Should not error even if file doesn't exist
	err := runConfigPath()
	assert.NoError(t, err)
}

func TestConfigCommand_Structure(t *testing.T) {
	cmd := newConfigCommand()

	assert.Equal(t, "config", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Verify subcommands exist
	subcommands := cmd.Commands()
	assert.GreaterOrEqual(t, len(subcommands), 5) // init, get, set, list, path
}

func TestConfigCommand_HasRequiredSubcommands(t *testing.T) {
	cmd := newConfigCommand()

	subcommandNames := make([]string, 0)
	for _, subcmd := range cmd.Commands() {
		subcommandNames = append(subcommandNames, subcmd.Name())
	}

	assert.Contains(t, subcommandNames, "init")
	assert.Contains(t, subcommandNames, "get")
	assert.Contains(t, subcommandNames, "set")
	assert.Contains(t, subcommandNames, "list")
	assert.Contains(t, subcommandNames, "path")
}
