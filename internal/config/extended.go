package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ExtendedConfig contains all application configuration with comprehensive settings.
type ExtendedConfig struct {
	Directories  DirectoriesConfig  `mapstructure:"directories" json:"directories" yaml:"directories" toml:"directories"`
	Logging      LoggingConfig      `mapstructure:"logging" json:"logging" yaml:"logging" toml:"logging"`
	Symlinks     SymlinksConfig     `mapstructure:"symlinks" json:"symlinks" yaml:"symlinks" toml:"symlinks"`
	Ignore       IgnoreConfig       `mapstructure:"ignore" json:"ignore" yaml:"ignore" toml:"ignore"`
	Dotfile      DotfileConfig      `mapstructure:"dotfile" json:"dotfile" yaml:"dotfile" toml:"dotfile"`
	Output       OutputConfig       `mapstructure:"output" json:"output" yaml:"output" toml:"output"`
	Operations   OperationsConfig   `mapstructure:"operations" json:"operations" yaml:"operations" toml:"operations"`
	Packages     PackagesConfig     `mapstructure:"packages" json:"packages" yaml:"packages" toml:"packages"`
	Doctor       DoctorConfig       `mapstructure:"doctor" json:"doctor" yaml:"doctor" toml:"doctor"`
	Experimental ExperimentalConfig `mapstructure:"experimental" json:"experimental" yaml:"experimental" toml:"experimental"`
}

// DirectoriesConfig contains directory path configuration.
type DirectoriesConfig struct {
	// Package directory containing packages
	Package string `mapstructure:"package" json:"package" yaml:"package" toml:"package"`

	// Target directory for symlinks
	Target string `mapstructure:"target" json:"target" yaml:"target" toml:"target"`

	// Manifest directory for tracking
	Manifest string `mapstructure:"manifest" json:"manifest" yaml:"manifest" toml:"manifest"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	// Log level: DEBUG, INFO, WARN, ERROR
	Level string `mapstructure:"level" json:"level" yaml:"level" toml:"level"`

	// Log format: text, json
	Format string `mapstructure:"format" json:"format" yaml:"format" toml:"format"`

	// Log destination: stderr, stdout, file
	Destination string `mapstructure:"destination" json:"destination" yaml:"destination" toml:"destination"`

	// Log file path (only used if destination is "file")
	File string `mapstructure:"file" json:"file" yaml:"file" toml:"file"`
}

// SymlinksConfig contains symlink behavior configuration.
type SymlinksConfig struct {
	// Link mode: relative, absolute
	Mode string `mapstructure:"mode" json:"mode" yaml:"mode" toml:"mode"`

	// Enable directory folding optimization
	Folding bool `mapstructure:"folding" json:"folding" yaml:"folding" toml:"folding"`

	// Overwrite existing files when conflicts occur
	Overwrite bool `mapstructure:"overwrite" json:"overwrite" yaml:"overwrite" toml:"overwrite"`

	// Create backup of overwritten files
	Backup bool `mapstructure:"backup" json:"backup" yaml:"backup" toml:"backup"`

	// Backup suffix when backups enabled
	BackupSuffix string `mapstructure:"backup_suffix" json:"backup_suffix" yaml:"backup_suffix" toml:"backup_suffix"`

	// Directory for backup files (default: <target>/.dot-backup)
	BackupDir string `mapstructure:"backup_dir" json:"backup_dir" yaml:"backup_dir" toml:"backup_dir"`
}

// IgnoreConfig contains ignore pattern configuration.
type IgnoreConfig struct {
	// Use default ignore patterns
	UseDefaults bool `mapstructure:"use_defaults" json:"use_defaults" yaml:"use_defaults" toml:"use_defaults"`

	// Additional patterns to ignore (glob format)
	Patterns []string `mapstructure:"patterns" json:"patterns" yaml:"patterns" toml:"patterns"`

	// Patterns to override (force include even if ignored)
	Overrides []string `mapstructure:"overrides" json:"overrides" yaml:"overrides" toml:"overrides"`
}

// DotfileConfig contains dotfile translation configuration.
type DotfileConfig struct {
	// Enable dot- to . translation
	Translate bool `mapstructure:"translate" json:"translate" yaml:"translate" toml:"translate"`

	// Prefix for dotfile translation
	Prefix string `mapstructure:"prefix" json:"prefix" yaml:"prefix" toml:"prefix"`

	// PackageNameMapping enables package name to target directory mapping.
	// When enabled, package "dot-gnupg" targets ~/.gnupg/ instead of ~/.
	// Default: true (project is pre-1.0, breaking change acceptable)
	PackageNameMapping bool `mapstructure:"package_name_mapping" json:"package_name_mapping" yaml:"package_name_mapping" toml:"package_name_mapping"`
}

// OutputConfig contains output formatting configuration.
type OutputConfig struct {
	// Default output format: text, json, yaml, table
	Format string `mapstructure:"format" json:"format" yaml:"format" toml:"format"`

	// Enable colored output: auto, always, never
	Color string `mapstructure:"color" json:"color" yaml:"color" toml:"color"`

	// Show progress indicators
	Progress bool `mapstructure:"progress" json:"progress" yaml:"progress" toml:"progress"`

	// Verbosity level: 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)
	Verbosity int `mapstructure:"verbosity" json:"verbosity" yaml:"verbosity" toml:"verbosity"`

	// Terminal width for text wrapping (0 = auto-detect)
	Width int `mapstructure:"width" json:"width" yaml:"width" toml:"width"`
}

// OperationsConfig contains operation behavior configuration.
type OperationsConfig struct {
	// Enable dry-run mode by default
	DryRun bool `mapstructure:"dry_run" json:"dry_run" yaml:"dry_run" toml:"dry_run"`

	// Enable atomic operations with rollback
	Atomic bool `mapstructure:"atomic" json:"atomic" yaml:"atomic" toml:"atomic"`

	// Maximum number of parallel operations (0 = auto-detect CPU count)
	MaxParallel int `mapstructure:"max_parallel" json:"max_parallel" yaml:"max_parallel" toml:"max_parallel"`
}

// PackagesConfig contains package management configuration.
type PackagesConfig struct {
	// Default sort order for list command: name, links, date
	SortBy string `mapstructure:"sort_by" json:"sort_by" yaml:"sort_by" toml:"sort_by"`

	// Automatically scan for new packages
	AutoDiscover bool `mapstructure:"auto_discover" json:"auto_discover" yaml:"auto_discover" toml:"auto_discover"`

	// Package naming convention validation
	ValidateNames bool `mapstructure:"validate_names" json:"validate_names" yaml:"validate_names" toml:"validate_names"`
}

// DoctorConfig contains doctor command configuration.
type DoctorConfig struct {
	// Auto-fix issues when possible
	AutoFix bool `mapstructure:"auto_fix" json:"auto_fix" yaml:"auto_fix" toml:"auto_fix"`

	// Check manifest integrity
	CheckManifest bool `mapstructure:"check_manifest" json:"check_manifest" yaml:"check_manifest" toml:"check_manifest"`

	// Check for broken symlinks
	CheckBrokenLinks bool `mapstructure:"check_broken_links" json:"check_broken_links" yaml:"check_broken_links" toml:"check_broken_links"`

	// Check for orphaned links
	CheckOrphaned bool `mapstructure:"check_orphaned" json:"check_orphaned" yaml:"check_orphaned" toml:"check_orphaned"`

	// Check file permissions
	CheckPermissions bool `mapstructure:"check_permissions" json:"check_permissions" yaml:"check_permissions" toml:"check_permissions"`
}

// ExperimentalConfig contains experimental feature flags.
type ExperimentalConfig struct {
	// Enable parallel operations
	Parallel bool `mapstructure:"parallel" json:"parallel" yaml:"parallel" toml:"parallel"`

	// Enable performance profiling
	Profiling bool `mapstructure:"profiling" json:"profiling" yaml:"profiling" toml:"profiling"`
}

// DefaultExtended returns extended configuration with sensible defaults.
func DefaultExtended() *ExtendedConfig {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "."
	}

	return &ExtendedConfig{
		Directories: DirectoriesConfig{
			Package:  ".",
			Target:   homeDir,
			Manifest: getXDGDataPath("dot/manifest"),
		},
		Logging: LoggingConfig{
			Level:       "INFO",
			Format:      "text",
			Destination: "stderr",
			File:        getXDGStatePath("dot/dot.log"),
		},
		Symlinks: SymlinksConfig{
			Mode:         "relative",
			Folding:      true,
			Overwrite:    false,
			Backup:       false,
			BackupSuffix: ".bak",
		},
		Ignore: IgnoreConfig{
			UseDefaults: true,
			Patterns:    []string{},
			Overrides:   []string{},
		},
		Dotfile: DotfileConfig{
			Translate:          true,
			Prefix:             "dot-",
			PackageNameMapping: true,
		},
		Output: OutputConfig{
			Format:    "text",
			Color:     "auto",
			Progress:  true,
			Verbosity: 1,
			Width:     0,
		},
		Operations: OperationsConfig{
			DryRun:      false,
			Atomic:      true,
			MaxParallel: 0,
		},
		Packages: PackagesConfig{
			SortBy:        "name",
			AutoDiscover:  true,
			ValidateNames: true,
		},
		Doctor: DoctorConfig{
			AutoFix:          false,
			CheckManifest:    true,
			CheckBrokenLinks: true,
			CheckOrphaned:    true,
			CheckPermissions: true,
		},
		Experimental: ExperimentalConfig{
			Parallel:  false,
			Profiling: false,
		},
	}
}

// LoadExtendedFromFile loads extended configuration from specified file.
func LoadExtendedFromFile(path string) (*ExtendedConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultExtended()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// Validate checks configuration for errors.
func (c *ExtendedConfig) Validate() error {
	if err := c.validateDirectories(); err != nil {
		return err
	}
	if err := c.validateLogging(); err != nil {
		return err
	}
	if err := c.validateSymlinks(); err != nil {
		return err
	}
	if err := c.validateIgnore(); err != nil {
		return err
	}
	if err := c.validateDotfile(); err != nil {
		return err
	}
	if err := c.validateOutput(); err != nil {
		return err
	}
	if err := c.validateOperations(); err != nil {
		return err
	}
	if err := c.validatePackages(); err != nil {
		return err
	}

	return nil
}

func (c *ExtendedConfig) validateDirectories() error {
	if c.Directories.Package == "" {
		return fmt.Errorf("directories.package: package directory cannot be empty")
	}

	if c.Directories.Target == "" {
		return fmt.Errorf("directories.target: target directory cannot be empty")
	}

	return nil
}

func (c *ExtendedConfig) validateLogging() error {
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if !contains(validLevels, c.Logging.Level) {
		return fmt.Errorf("logging.level: invalid log level %q (must be one of: %s)",
			c.Logging.Level, strings.Join(validLevels, ", "))
	}

	validFormats := []string{"text", "json"}
	if !contains(validFormats, c.Logging.Format) {
		return fmt.Errorf("logging.format: invalid log format %q (must be one of: %s)",
			c.Logging.Format, strings.Join(validFormats, ", "))
	}

	validDestinations := []string{"stderr", "stdout", "file"}
	if !contains(validDestinations, c.Logging.Destination) {
		return fmt.Errorf("logging.destination: invalid log destination %q (must be one of: %s)",
			c.Logging.Destination, strings.Join(validDestinations, ", "))
	}

	if c.Logging.Destination == "file" && c.Logging.File == "" {
		return fmt.Errorf("logging.file: log file must be specified when destination is 'file'")
	}

	return nil
}

func (c *ExtendedConfig) validateSymlinks() error {
	validModes := []string{"relative", "absolute"}
	if !contains(validModes, c.Symlinks.Mode) {
		return fmt.Errorf("symlinks.mode: invalid symlink mode %q (must be one of: %s)",
			c.Symlinks.Mode, strings.Join(validModes, ", "))
	}

	if c.Symlinks.Backup && c.Symlinks.BackupSuffix == "" {
		return fmt.Errorf("symlinks.backup_suffix: backup suffix cannot be empty when backup is enabled")
	}

	return nil
}

func (c *ExtendedConfig) validateIgnore() error {
	// Validate ignore patterns are valid globs
	for i, pattern := range c.Ignore.Patterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("ignore.patterns[%d]: invalid glob pattern %q: %w", i, pattern, err)
		}
	}

	// Validate override patterns
	for i, pattern := range c.Ignore.Overrides {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("ignore.overrides[%d]: invalid glob pattern %q: %w", i, pattern, err)
		}
	}

	return nil
}

func (c *ExtendedConfig) validateDotfile() error {
	if c.Dotfile.Translate && c.Dotfile.Prefix == "" {
		return fmt.Errorf("dotfile.prefix: dotfile prefix cannot be empty when translate is enabled")
	}

	return nil
}

func (c *ExtendedConfig) validateOutput() error {
	validFormats := []string{"text", "json", "yaml", "table"}
	if !contains(validFormats, c.Output.Format) {
		return fmt.Errorf("output.format: invalid output format %q (must be one of: %s)",
			c.Output.Format, strings.Join(validFormats, ", "))
	}

	validColors := []string{"auto", "always", "never"}
	if !contains(validColors, c.Output.Color) {
		return fmt.Errorf("output.color: invalid color mode %q (must be one of: %s)",
			c.Output.Color, strings.Join(validColors, ", "))
	}

	if c.Output.Verbosity < 0 || c.Output.Verbosity > 3 {
		return fmt.Errorf("output.verbosity: verbosity must be between 0 and 3, got %d", c.Output.Verbosity)
	}

	if c.Output.Width < 0 {
		return fmt.Errorf("output.width: width cannot be negative (use 0 for auto-detect), got %d", c.Output.Width)
	}

	return nil
}

func (c *ExtendedConfig) validateOperations() error {
	if c.Operations.MaxParallel < 0 {
		return fmt.Errorf("operations.max_parallel: max_parallel cannot be negative (use 0 for auto-detect), got %d",
			c.Operations.MaxParallel)
	}

	return nil
}

func (c *ExtendedConfig) validatePackages() error {
	validSortBy := []string{"name", "links", "date"}
	if !contains(validSortBy, c.Packages.SortBy) {
		return fmt.Errorf("packages.sort_by: invalid sort field %q (must be one of: %s)",
			c.Packages.SortBy, strings.Join(validSortBy, ", "))
	}

	return nil
}

// getXDGDataPath returns XDG data directory path.
func getXDGDataPath(suffix string) string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, suffix)
	}
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".local", "share", suffix)
}

// getXDGStatePath returns XDG state directory path.
func getXDGStatePath(suffix string) string {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, suffix)
	}
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".local", "state", suffix)
}
