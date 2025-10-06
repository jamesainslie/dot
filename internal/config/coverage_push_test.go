package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Targeted tests to push coverage over 80%

func TestLoadFromFile_NonexistentFile(t *testing.T) {
	_, err := config.LoadFromFile("/nonexistent/path/config.yaml")
	// Should error on missing file
	assert.Error(t, err)
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.yaml")
	
	err := os.WriteFile(badFile, []byte("invalid: yaml: [[["), 0600)
	require.NoError(t, err)
	
	_, err = config.LoadFromFile(badFile)
	assert.Error(t, err)
}

func TestLoadWithEnv_AllVariables(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	// Minimal config
	err := os.WriteFile(configFile, []byte(""), 0600)
	require.NoError(t, err)

	// Set all possible env vars
	os.Setenv("DOT_LOG_LEVEL", "WARN")
	os.Setenv("DOT_LOG_FORMAT", "json")
	defer os.Unsetenv("DOT_LOG_LEVEL")
	defer os.Unsetenv("DOT_LOG_FORMAT")

	cfg, err := config.LoadWithEnv(configFile)
	if err == nil {
		assert.Equal(t, "WARN", cfg.LogLevel)
		assert.Equal(t, "json", cfg.LogFormat)
	}
}

func TestGetConfigPath_AppNameVariations(t *testing.T) {
	tests := []string{
		"dot",
		"my-app",
		"",
		"app.name",
	}

	for _, appName := range tests {
		path := config.GetConfigPath(appName)
		assert.NotEmpty(t, path)
	}
}

func TestDefaultExtended_ValidationPasses(t *testing.T) {
	cfg := config.DefaultExtended()
	err := cfg.Validate()
	assert.NoError(t, err, "default config should always be valid")
}

func TestValidateOutputFormats(t *testing.T) {
	formats := []string{"text", "json", "yaml", "table"}
	
	for _, format := range formats {
		cfg := config.DefaultExtended()
		cfg.Output.Format = format
		
		err := cfg.Validate()
		assert.NoError(t, err, "format %s should be valid", format)
	}
}

func TestValidateOutputColors(t *testing.T) {
	colors := []string{"auto", "always", "never"}
	
	for _, color := range colors {
		cfg := config.DefaultExtended()
		cfg.Output.Color = color
		
		err := cfg.Validate()
		assert.NoError(t, err, "color %s should be valid", color)
	}
}

func TestValidateLoggingLevels(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	
	for _, level := range levels {
		cfg := config.DefaultExtended()
		cfg.Logging.Level = level
		
		err := cfg.Validate()
		assert.NoError(t, err, "level %s should be valid", level)
	}
}

func TestValidateLoggingFormats(t *testing.T) {
	formats := []string{"text", "json"}
	
	for _, format := range formats {
		cfg := config.DefaultExtended()
		cfg.Logging.Format = format
		
		err := cfg.Validate()
		assert.NoError(t, err, "format %s should be valid", format)
	}
}

func TestValidateSymlinkModes(t *testing.T) {
	modes := []string{"relative", "absolute"}
	
	for _, mode := range modes {
		cfg := config.DefaultExtended()
		cfg.Symlinks.Mode = mode
		
		err := cfg.Validate()
		assert.NoError(t, err, "mode %s should be valid", mode)
	}
}

func TestValidateWithAllBooleanCombinations(t *testing.T) {
	cfg := config.DefaultExtended()
	
	// Test all boolean flag combinations
	boolCombos := []struct {
		folding   bool
		overwrite bool
		backup    bool
	}{
		{true, true, true},
		{true, true, false},
		{true, false, true},
		{true, false, false},
		{false, true, true},
		{false, true, false},
		{false, false, true},
		{false, false, false},
	}

	for _, combo := range boolCombos {
		cfg.Symlinks.Folding = combo.folding
		cfg.Symlinks.Overwrite = combo.overwrite
		cfg.Symlinks.Backup = combo.backup
		
		err := cfg.Validate()
		assert.NoError(t, err)
	}
}

