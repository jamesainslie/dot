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

func TestGetConfigValue(t *testing.T) {
	cfg := &config.ExtendedConfig{
		Directories: config.DirectoriesConfig{
			Package:  "/test/package",
			Target:   "/test/target",
			Manifest: "/test/manifest",
		},
		Logging: config.LoggingConfig{
			Level:       "INFO",
			Format:      "text",
			Destination: "stderr",
		},
		Symlinks: config.SymlinksConfig{
			Mode:         "relative",
			BackupSuffix: ".bak",
			BackupDir:    "/test/backup",
		},
		Dotfile: config.DotfileConfig{
			Prefix: "dot-",
		},
		Output: config.OutputConfig{
			Format: "text",
			Color:  "auto",
		},
		Packages: config.PackagesConfig{
			SortBy: "name",
		},
	}

	tests := []struct {
		key      string
		expected string
		wantErr  bool
	}{
		{"directories.package", "/test/package", false},
		{"directories.target", "/test/target", false},
		{"directories.manifest", "/test/manifest", false},
		{"logging.level", "INFO", false},
		{"logging.format", "text", false},
		{"logging.destination", "stderr", false},
		{"symlinks.mode", "relative", false},
		{"symlinks.backup_suffix", ".bak", false},
		{"symlinks.backup_dir", "/test/backup", false},
		{"dotfile.prefix", "dot-", false},
		{"output.format", "text", false},
		{"output.color", "auto", false},
		{"packages.sort_by", "name", false},
		{"unknown.key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, err := getConfigValue(cfg, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
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

func TestConfigCommand_List_DisplaysAllSections(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Create a config file with custom values
	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/dotfiles"
	cfg.Logging.Level = "DEBUG"
	cfg.Symlinks.Mode = "absolute"
	cfg.Ignore.UseDefaults = false
	cfg.Dotfile.Translate = false
	cfg.Output.Format = "json"
	cfg.Operations.DryRun = true
	cfg.Packages.SortBy = "date"
	cfg.Doctor.AutoFix = true
	cfg.Experimental.Parallel = true

	writer := config.NewWriter(configPath)
	err := writer.Write(cfg, config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	// Load the config using the list command
	loader := config.NewLoader("dot", configPath)
	loadedCfg, err := loader.LoadWithEnv()
	require.NoError(t, err)

	// Verify all sections are present in loaded config
	assert.Equal(t, "/test/dotfiles", loadedCfg.Directories.Package)
	assert.Equal(t, "DEBUG", loadedCfg.Logging.Level)
	assert.Equal(t, "absolute", loadedCfg.Symlinks.Mode)
	assert.False(t, loadedCfg.Ignore.UseDefaults)
	assert.False(t, loadedCfg.Dotfile.Translate)
	assert.Equal(t, "json", loadedCfg.Output.Format)
	assert.True(t, loadedCfg.Operations.DryRun)
	assert.Equal(t, "date", loadedCfg.Packages.SortBy)
	assert.True(t, loadedCfg.Doctor.AutoFix)
	assert.True(t, loadedCfg.Experimental.Parallel)
}

func TestConfigCommand_List_DisplaysDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	os.Setenv("DOT_CONFIG", configPath)
	defer os.Unsetenv("DOT_CONFIG")

	// Create config with defaults
	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	require.NoError(t, err)

	// Verify defaults are loaded correctly
	assert.NotNil(t, cfg)
	assert.Equal(t, ".", cfg.Directories.Package)
	assert.Equal(t, "INFO", cfg.Logging.Level)
	assert.Equal(t, "relative", cfg.Symlinks.Mode)
	assert.True(t, cfg.Ignore.UseDefaults)
	assert.True(t, cfg.Dotfile.Translate)
	assert.Equal(t, "text", cfg.Output.Format)
	assert.False(t, cfg.Operations.DryRun)
	assert.Equal(t, "name", cfg.Packages.SortBy)
	assert.False(t, cfg.Doctor.AutoFix)
	assert.False(t, cfg.Experimental.Parallel)
}

func TestGetValidConfigKeys(t *testing.T) {
	keys := getValidConfigKeys()

	// Verify keys are returned
	assert.NotEmpty(t, keys)

	// Verify specific keys are present
	expectedKeys := []string{
		"directories.package",
		"directories.target",
		"directories.manifest",
		"logging.level",
		"logging.format",
		"logging.destination",
		"symlinks.mode",
		"symlinks.backup_suffix",
		"symlinks.backup_dir",
		"dotfile.prefix",
		"output.format",
		"output.color",
		"packages.sort_by",
	}

	for _, expected := range expectedKeys {
		assert.Contains(t, keys, expected)
	}
}

func TestConfigGetCommand_Completion(t *testing.T) {
	cmd := newConfigGetCommand()

	// Verify ValidArgsFunction is set
	assert.NotNil(t, cmd.ValidArgsFunction)

	// Test completion function
	completions, directive := cmd.ValidArgsFunction(cmd, []string{}, "")

	// Should return all valid config keys
	assert.NotEmpty(t, completions)
	assert.Contains(t, completions, "directories.package")
	assert.Contains(t, completions, "logging.level")
	assert.Contains(t, completions, "output.format")

	// Should use NoFileComp directive (don't complete file names)
	assert.Equal(t, 4, int(directive))
}

func TestConfigSetCommand_Completion(t *testing.T) {
	cmd := newConfigSetCommand()

	// Verify ValidArgsFunction is set
	assert.NotNil(t, cmd.ValidArgsFunction)

	// Test completion for first argument (key)
	completions, directive := cmd.ValidArgsFunction(cmd, []string{}, "")

	// Should return all valid config keys
	assert.NotEmpty(t, completions)
	assert.Contains(t, completions, "directories.package")
	assert.Contains(t, completions, "logging.level")

	// Should use NoFileComp directive
	assert.Equal(t, 4, int(directive))

	// Test completion for second argument (value) - should not suggest anything
	completions, directive = cmd.ValidArgsFunction(cmd, []string{"logging.level"}, "")

	// Should return no completions for value
	assert.Empty(t, completions)
	assert.Equal(t, 4, int(directive)) // NoFileComp
}
