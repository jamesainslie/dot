package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles loading configuration from multiple sources.
type Loader struct {
	appName    string
	configPath string
}

// NewLoader creates a configuration loader.
func NewLoader(appName string, configPath string) *Loader {
	return &Loader{
		appName:    appName,
		configPath: configPath,
	}
}

// Load loads configuration from file with proper precedence.
// Precedence: file > defaults
func (l *Loader) Load() (*ExtendedConfig, error) {
	// Start with defaults
	cfg := DefaultExtended()

	// Load from config file if it exists
	if fileExists(l.configPath) {
		fileCfg, err := LoadExtendedFromFile(l.configPath)
		if err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
		cfg = mergeConfigs(cfg, fileCfg)
	}

	// Validate merged configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadWithEnv loads configuration from file and applies environment variable overrides.
// Precedence: env > file > defaults
func (l *Loader) LoadWithEnv() (*ExtendedConfig, error) {
	// Start with file load
	cfg, err := l.Load()
	if err != nil {
		return nil, err
	}

	// Load from environment (sparse config with only env-set values)
	envCfg := l.loadFromEnv()
	// Use simple merge for env (only strings, no booleans unless tracked)
	cfg = mergeConfigs(cfg, envCfg)

	// Validate merged configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadWithFlags loads configuration and applies flag overrides.
// Precedence: flags > env > file > defaults
func (l *Loader) LoadWithFlags(flags map[string]interface{}) (*ExtendedConfig, error) {
	// Load with env
	cfg, err := l.LoadWithEnv()
	if err != nil {
		return nil, err
	}

	// Apply flag overrides
	flagCfg, verbositySet := l.configFromFlags(flags)
	cfg = mergeConfigsWithVerbosity(cfg, flagCfg, verbositySet)

	// Validate again after flag overrides
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables.
// Returns a sparse config with only explicitly set environment values.
func (l *Loader) loadFromEnv() *ExtendedConfig {
	cfg := &ExtendedConfig{
		Directories:  DirectoriesConfig{},
		Logging:      LoggingConfig{},
		Symlinks:     SymlinksConfig{},
		Ignore:       IgnoreConfig{},
		Dotfile:      DotfileConfig{},
		Output:       OutputConfig{},
		Operations:   OperationsConfig{},
		Packages:     PackagesConfig{},
		Doctor:       DoctorConfig{},
		Experimental: ExperimentalConfig{},
	}

	// Only set values that have corresponding environment variables
	prefix := strings.ToUpper(l.appName) + "_"

	// Directories
	if val := getEnvWithPrefix(prefix, "DIRECTORIES_STOW"); val != "" {
		cfg.Directories.Stow = val
	}
	if val := getEnvWithPrefix(prefix, "DIRECTORIES_TARGET"); val != "" {
		cfg.Directories.Target = val
	}
	if val := getEnvWithPrefix(prefix, "DIRECTORIES_MANIFEST"); val != "" {
		cfg.Directories.Manifest = val
	}

	// Logging
	if val := getEnvWithPrefix(prefix, "LOGGING_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}
	if val := getEnvWithPrefix(prefix, "LOGGING_FORMAT"); val != "" {
		cfg.Logging.Format = val
	}
	if val := getEnvWithPrefix(prefix, "LOGGING_DESTINATION"); val != "" {
		cfg.Logging.Destination = val
	}
	if val := getEnvWithPrefix(prefix, "LOGGING_FILE"); val != "" {
		cfg.Logging.File = val
	}

	// Symlinks
	if val := getEnvWithPrefix(prefix, "SYMLINKS_MODE"); val != "" {
		cfg.Symlinks.Mode = val
	}

	// Output
	if val := getEnvWithPrefix(prefix, "OUTPUT_FORMAT"); val != "" {
		cfg.Output.Format = val
	}
	if val := getEnvWithPrefix(prefix, "OUTPUT_COLOR"); val != "" {
		cfg.Output.Color = val
	}

	// Packages
	if val := getEnvWithPrefix(prefix, "PACKAGES_SORT_BY"); val != "" {
		cfg.Packages.SortBy = val
	}

	return cfg
}

// getEnvWithPrefix gets an environment variable with the given prefix.
func getEnvWithPrefix(prefix, key string) string {
	return os.Getenv(prefix + key)
}

// bindEnvKeys binds all configuration keys to environment variables.
func (l *Loader) bindEnvKeys(v *viper.Viper) {
	v.BindEnv("directories.stow")
	v.BindEnv("directories.target")
	v.BindEnv("directories.manifest")

	v.BindEnv("logging.level")
	v.BindEnv("logging.format")
	v.BindEnv("logging.destination")
	v.BindEnv("logging.file")

	v.BindEnv("symlinks.mode")
	v.BindEnv("symlinks.folding")
	v.BindEnv("symlinks.overwrite")
	v.BindEnv("symlinks.backup")
	v.BindEnv("symlinks.backup_suffix")

	v.BindEnv("ignore.use_defaults")
	v.BindEnv("ignore.patterns")
	v.BindEnv("ignore.overrides")

	v.BindEnv("dotfile.translate")
	v.BindEnv("dotfile.prefix")

	v.BindEnv("output.format")
	v.BindEnv("output.color")
	v.BindEnv("output.progress")
	v.BindEnv("output.verbosity")
	v.BindEnv("output.width")

	v.BindEnv("operations.dry_run")
	v.BindEnv("operations.atomic")
	v.BindEnv("operations.max_parallel")

	v.BindEnv("packages.sort_by")
	v.BindEnv("packages.auto_discover")
	v.BindEnv("packages.validate_names")

	v.BindEnv("doctor.auto_fix")
	v.BindEnv("doctor.check_manifest")
	v.BindEnv("doctor.check_broken_links")
	v.BindEnv("doctor.check_orphaned")
	v.BindEnv("doctor.check_permissions")

	v.BindEnv("experimental.parallel")
	v.BindEnv("experimental.profiling")
}

// configFromFlags creates partial config from flag map.
func (l *Loader) configFromFlags(flags map[string]interface{}) (*ExtendedConfig, bool) {
	cfg := createSparseConfig()

	verbositySet := applyFlagsToConfig(cfg, flags)

	return cfg, verbositySet
}

// createSparseConfig creates an empty config for flag/env merging.
func createSparseConfig() *ExtendedConfig {
	return &ExtendedConfig{
		Directories:  DirectoriesConfig{},
		Logging:      LoggingConfig{},
		Symlinks:     SymlinksConfig{},
		Ignore:       IgnoreConfig{},
		Dotfile:      DotfileConfig{},
		Output:       OutputConfig{Verbosity: -1}, // Use -1 as sentinel for "not set"
		Operations:   OperationsConfig{},
		Packages:     PackagesConfig{},
		Doctor:       DoctorConfig{},
		Experimental: ExperimentalConfig{},
	}
}

// applyFlagsToConfig maps command-line flags to configuration fields.
func applyFlagsToConfig(cfg *ExtendedConfig, flags map[string]interface{}) bool {
	verbositySet := false

	applyDirectoryFlags(cfg, flags)
	applyLoggingFlags(cfg, flags)
	applyOperationFlags(cfg, flags)
	verbositySet = applyOutputFlags(cfg, flags)

	return verbositySet
}

// applyDirectoryFlags applies directory-related flags.
func applyDirectoryFlags(cfg *ExtendedConfig, flags map[string]interface{}) {
	if val, ok := flags["dir"].(string); ok && val != "" {
		cfg.Directories.Stow = val
	}
	if val, ok := flags["target"].(string); ok && val != "" {
		cfg.Directories.Target = val
	}
}

// applyLoggingFlags applies logging-related flags.
func applyLoggingFlags(cfg *ExtendedConfig, flags map[string]interface{}) {
	if val, ok := flags["log-json"].(bool); ok && val {
		cfg.Logging.Format = "json"
	}
}

// applyOperationFlags applies operation-related flags.
func applyOperationFlags(cfg *ExtendedConfig, flags map[string]interface{}) {
	if val, ok := flags["dry-run"].(bool); ok && val {
		cfg.Operations.DryRun = val
	}
}

// applyOutputFlags applies output-related flags and returns if verbosity was set.
func applyOutputFlags(cfg *ExtendedConfig, flags map[string]interface{}) bool {
	verbositySet := false

	if val, ok := flags["verbose"].(int); ok {
		cfg.Output.Verbosity = val
		verbositySet = true
	}
	if val, ok := flags["quiet"].(bool); ok && val {
		cfg.Output.Verbosity = 0
		verbositySet = true
	}
	if val, ok := flags["color"].(string); ok && val != "" {
		cfg.Output.Color = val
	}
	if val, ok := flags["format"].(string); ok && val != "" {
		cfg.Output.Format = val
	}

	return verbositySet
}

// mergeConfigs merges two configs, with override taking precedence for non-zero values.
// Only merges fields that are explicitly set in override (non-empty strings, non-zero lists).
func mergeConfigs(base, override *ExtendedConfig) *ExtendedConfig {
	return mergeConfigsWithVerbosity(base, override, false)
}

// mergeConfigsWithVerbosity merges configs with special handling for verbosity.
func mergeConfigsWithVerbosity(base, override *ExtendedConfig, verbosityExplicit bool) *ExtendedConfig {
	merged := *base

	mergeDirectories(&merged, override)
	mergeLogging(&merged, override)
	mergeSymlinks(&merged, override)
	mergeIgnore(&merged, override)
	mergeDotfile(&merged, override)
	mergeOutput(&merged, override, verbosityExplicit)
	mergeOperations(&merged, override)
	mergePackages(&merged, override)
	mergeDoctor(&merged, override)
	mergeExperimental(&merged, override)

	return &merged
}

// mergeDirectories merges directory configuration.
func mergeDirectories(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Directories.Stow != "" {
		merged.Directories.Stow = override.Directories.Stow
	}
	if override.Directories.Target != "" {
		merged.Directories.Target = override.Directories.Target
	}
	if override.Directories.Manifest != "" {
		merged.Directories.Manifest = override.Directories.Manifest
	}
}

// mergeLogging merges logging configuration.
func mergeLogging(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Logging.Level != "" {
		merged.Logging.Level = override.Logging.Level
	}
	if override.Logging.Format != "" {
		merged.Logging.Format = override.Logging.Format
	}
	if override.Logging.Destination != "" {
		merged.Logging.Destination = override.Logging.Destination
	}
	if override.Logging.File != "" {
		merged.Logging.File = override.Logging.File
	}
}

// mergeSymlinks merges symlink configuration.
func mergeSymlinks(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Symlinks.Mode != "" {
		merged.Symlinks.Mode = override.Symlinks.Mode
	}
	if override.Symlinks.BackupSuffix != "" {
		merged.Symlinks.BackupSuffix = override.Symlinks.BackupSuffix
	}
	if override.Symlinks.Overwrite {
		merged.Symlinks.Overwrite = true
	}
	if override.Symlinks.Backup {
		merged.Symlinks.Backup = true
	}
}

// mergeIgnore merges ignore pattern configuration.
func mergeIgnore(merged *ExtendedConfig, override *ExtendedConfig) {
	if len(override.Ignore.Patterns) > 0 {
		merged.Ignore.Patterns = override.Ignore.Patterns
	}
	if len(override.Ignore.Overrides) > 0 {
		merged.Ignore.Overrides = override.Ignore.Overrides
	}
}

// mergeDotfile merges dotfile translation configuration.
func mergeDotfile(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Dotfile.Prefix != "" {
		merged.Dotfile.Prefix = override.Dotfile.Prefix
	}
}

// mergeOutput merges output configuration with special verbosity handling.
func mergeOutput(merged *ExtendedConfig, override *ExtendedConfig, verbosityExplicit bool) {
	if override.Output.Format != "" {
		merged.Output.Format = override.Output.Format
	}
	if override.Output.Color != "" {
		merged.Output.Color = override.Output.Color
	}
	if verbosityExplicit || override.Output.Verbosity >= 0 {
		merged.Output.Verbosity = override.Output.Verbosity
	}
	if override.Output.Width > 0 {
		merged.Output.Width = override.Output.Width
	}
}

// mergeOperations merges operation configuration.
func mergeOperations(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Operations.DryRun {
		merged.Operations.DryRun = true
	}
	if override.Operations.MaxParallel > 0 {
		merged.Operations.MaxParallel = override.Operations.MaxParallel
	}
}

// mergePackages merges package management configuration.
func mergePackages(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Packages.SortBy != "" {
		merged.Packages.SortBy = override.Packages.SortBy
	}
}

// mergeDoctor merges doctor configuration.
func mergeDoctor(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Doctor.AutoFix {
		merged.Doctor.AutoFix = true
	}
}

// mergeExperimental merges experimental feature configuration.
func mergeExperimental(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Experimental.Parallel {
		merged.Experimental.Parallel = true
	}
	if override.Experimental.Profiling {
		merged.Experimental.Profiling = true
	}
}
