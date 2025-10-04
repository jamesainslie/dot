// Package config provides centralized configuration management with Viper isolation.
//
// All Viper usage is contained within this package. Other packages must access
// configuration through this package's functions, never directly via Viper.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config contains all application configuration.
type Config struct {
	// LogLevel specifies logging level: DEBUG, INFO, WARN, ERROR
	LogLevel string `mapstructure:"log_level" json:"log_level" yaml:"log_level" toml:"log_level"`

	// LogFormat specifies log output format: json, text
	LogFormat string `mapstructure:"log_format" json:"log_format" yaml:"log_format" toml:"log_format"`
}

// Default returns configuration with secure defaults.
func Default() *Config {
	return &Config{
		LogLevel:  "INFO",
		LogFormat: "json",
	}
}

// LoadFromFile loads configuration from specified file.
// Supports YAML, JSON, and TOML formats based on file extension.
func LoadFromFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := Default()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// LoadWithEnv loads configuration from file and applies environment variable overrides.
// Environment variables use DOT_ prefix and replace dots with underscores.
func LoadWithEnv(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	// Set up environment variable handling
	v.SetEnvPrefix("DOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		// File not found is acceptable, use defaults with env overrides
	}

	cfg := Default()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// Validate checks configuration for errors.
func (c *Config) Validate() error {
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if !contains(validLevels, c.LogLevel) {
		return fmt.Errorf("invalid log level: %s (must be DEBUG, INFO, WARN, or ERROR)", c.LogLevel)
	}

	validFormats := []string{"json", "text"}
	if !contains(validFormats, c.LogFormat) {
		return fmt.Errorf("invalid log format: %s (must be json or text)", c.LogFormat)
	}

	return nil
}

// GetConfigPath returns XDG-compliant configuration directory path.
// Uses XDG_CONFIG_HOME if set, otherwise falls back to ~/.config on Unix systems.
func GetConfigPath(appName string) string {
	// Use XDG_CONFIG_HOME if set
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, appName)
	}

	// Try os.UserConfigDir (cross-platform)
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, appName)
	}

	// Fallback to ~/.config on Unix
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".config", appName)
	}

	// Last resort fallback
	return filepath.Join(".", appName)
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

