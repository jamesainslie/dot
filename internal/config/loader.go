package config

import (
	"fmt"
	"reflect"
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
	// Load from config file if it exists
	if fileExists(l.configPath) {
		fileCfg, err := LoadExtendedFromFile(l.configPath)
		if err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
		// Use file config directly to preserve explicit false values
		return fileCfg, nil
	}

	// No file - return defaults
	cfg := DefaultExtended()

	// Validate configuration
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

	// Apply environment overrides directly onto the loaded configuration,
	// keyed on viper's IsSet. A sparse overlay merged by zero-value cannot
	// represent "explicitly set to false", which silently discarded env
	// overrides like DOT_DOTFILE_PACKAGE_NAME_MAPPING=false (issue #86).
	if err := applyEnvOverrides(l.newEnvViper(), cfg); err != nil {
		return nil, fmt.Errorf("apply environment configuration: %w", err)
	}

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

	// Flags normally arrive already expanded by the shell, but a quoted or
	// script-supplied value reaches us verbatim. Expand the overlay before
	// merging, for the same reason the environment overlay is expanded first.
	if err := flagCfg.ExpandPaths(); err != nil {
		return nil, fmt.Errorf("expand flag configuration: %w", err)
	}

	cfg = mergeConfigsWithVerbosity(cfg, flagCfg, verbositySet)

	// Validate again after flag overrides
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// newEnvViper returns a viper instance with every configuration key bound
// to its DOT_-prefixed environment variable.
func (l *Loader) newEnvViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(strings.ToUpper(l.appName))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind every key, derived from the config struct's mapstructure tags so
	// a new field cannot be forgotten here (issue #86: a hand-written bind
	// list silently ignored the keys it missed).
	for _, key := range configKeys() {
		_ = v.BindEnv(key) // BindEnv only errors when called with no key
	}

	return v
}

// configKeys returns every section.field key in ExtendedConfig, derived by
// reflection from the mapstructure tags. Only tagged struct sections with
// tagged fields produce keys, so the walk tolerates the struct growing a
// scalar or untagged field without panicking on every command start.
func configKeys() []string {
	var keys []string
	root := reflect.TypeOf(ExtendedConfig{})
	for i := 0; i < root.NumField(); i++ {
		sectionType := root.Field(i).Type
		if sectionType.Kind() != reflect.Struct {
			continue
		}
		section := root.Field(i).Tag.Get("mapstructure")
		if section == "" {
			continue
		}
		for j := 0; j < sectionType.NumField(); j++ {
			field := sectionType.Field(j).Tag.Get("mapstructure")
			if field == "" {
				continue
			}
			keys = append(keys, section+"."+field)
		}
	}
	return keys
}

// applyEnvOverrides applies every environment-set key directly onto cfg.
// Path-typed values are expanded here, mirroring the expansion file values
// receive on load. The per-section lists below are kept total by
// TestLoader_EnvOverridesEveryConfigKey, which walks the config struct.
func applyEnvOverrides(v *viper.Viper, cfg *ExtendedConfig) error {
	if err := applyPathEnv(v, cfg); err != nil {
		return err
	}

	applyEnv(v, v.GetString, map[string]*string{
		"logging.level":          &cfg.Logging.Level,
		"logging.format":         &cfg.Logging.Format,
		"logging.destination":    &cfg.Logging.Destination,
		"symlinks.mode":          &cfg.Symlinks.Mode,
		"symlinks.backup_suffix": &cfg.Symlinks.BackupSuffix,
		"dotfile.prefix":         &cfg.Dotfile.Prefix,
		"output.format":          &cfg.Output.Format,
		"output.color":           &cfg.Output.Color,
		"output.table_style":     &cfg.Output.TableStyle,
		"packages.sort_by":       &cfg.Packages.SortBy,
		"update.package_manager": &cfg.Update.PackageManager,
		"update.repository":      &cfg.Update.Repository,
		"network.http_proxy":     &cfg.Network.HTTPProxy,
		"network.https_proxy":    &cfg.Network.HTTPSProxy,
		"network.no_proxy":       &cfg.Network.NoProxy,
	})

	applyEnv(v, v.GetBool, map[string]*bool{
		"symlinks.folding":               &cfg.Symlinks.Folding,
		"symlinks.overwrite":             &cfg.Symlinks.Overwrite,
		"symlinks.backup":                &cfg.Symlinks.Backup,
		"ignore.use_defaults":            &cfg.Ignore.UseDefaults,
		"ignore.per_package_ignore":      &cfg.Ignore.PerPackageIgnore,
		"ignore.interactive_large_files": &cfg.Ignore.InteractiveLargeFiles,
		"dotfile.translate":              &cfg.Dotfile.Translate,
		"dotfile.package_name_mapping":   &cfg.Dotfile.PackageNameMapping,
		"output.progress":                &cfg.Output.Progress,
		"operations.dry_run":             &cfg.Operations.DryRun,
		"operations.atomic":              &cfg.Operations.Atomic,
		"packages.auto_discover":         &cfg.Packages.AutoDiscover,
		"packages.validate_names":        &cfg.Packages.ValidateNames,
		"doctor.auto_fix":                &cfg.Doctor.AutoFix,
		"doctor.check_manifest":          &cfg.Doctor.CheckManifest,
		"doctor.check_broken_links":      &cfg.Doctor.CheckBrokenLinks,
		"doctor.check_orphaned":          &cfg.Doctor.CheckOrphaned,
		"doctor.check_permissions":       &cfg.Doctor.CheckPermissions,
		"update.check_on_startup":        &cfg.Update.CheckOnStartup,
		"update.include_prerelease":      &cfg.Update.IncludePrerelease,
		"experimental.parallel":          &cfg.Experimental.Parallel,
		"experimental.profiling":         &cfg.Experimental.Profiling,
	})

	applyEnv(v, v.GetInt, map[string]*int{
		"output.verbosity":        &cfg.Output.Verbosity,
		"output.width":            &cfg.Output.Width,
		"operations.max_parallel": &cfg.Operations.MaxParallel,
		"update.check_frequency":  &cfg.Update.CheckFrequency,
		"network.timeout":         &cfg.Network.Timeout,
		"network.connect_timeout": &cfg.Network.ConnectTimeout,
		"network.tls_timeout":     &cfg.Network.TLSTimeout,
	})

	applyEnv(v, v.GetInt64, map[string]*int64{
		"ignore.max_file_size": &cfg.Ignore.MaxFileSize,
	})

	// Lists are comma-separated in the environment, matching the documented
	// contract and the comma convention config set already uses. Viper's own
	// GetStringSlice would whitespace-split instead.
	applyEnv(v, func(key string) []string { return splitEnvList(v.GetString(key)) }, map[string]*[]string{
		"ignore.patterns":  &cfg.Ignore.Patterns,
		"ignore.overrides": &cfg.Ignore.Overrides,
	})

	return nil
}

// applyPathEnv sets path-typed values from the environment, expanding "~"
// and "$VAR" references the way file values are expanded on load. It walks
// pathFields in order, so with several bad values the reported key is
// deterministic.
func applyPathEnv(v *viper.Viper, cfg *ExtendedConfig) error {
	for _, field := range cfg.pathFields() {
		if !v.IsSet(field.key) {
			continue
		}
		expanded, err := expandPath(v.GetString(field.key))
		if err != nil {
			return fmt.Errorf("%s: %w", field.key, err)
		}
		*field.value = expanded
	}
	return nil
}

// applyEnv copies every environment-set key in targets onto its destination.
func applyEnv[T any](v *viper.Viper, get func(string) T, targets map[string]*T) {
	for key, dst := range targets {
		if v.IsSet(key) {
			*dst = get(key)
		}
	}
}

// splitEnvList splits a comma-separated environment value into its items,
// trimming surrounding whitespace and dropping empty entries.
func splitEnvList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

// configFromFlags creates partial config from flag map.
func (l *Loader) configFromFlags(flags map[string]interface{}) (*ExtendedConfig, bool) {
	cfg := createSparseConfig()

	verbositySet := applyFlagsToConfig(cfg, flags)

	return cfg, verbositySet
}

// createSparseConfig creates an empty config for flag merging. Verbosity -1
// is the "not set" sentinel.
func createSparseConfig() *ExtendedConfig {
	return &ExtendedConfig{Output: OutputConfig{Verbosity: -1}}
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
		cfg.Directories.Package = val
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
	if override.Directories.Package != "" {
		merged.Directories.Package = override.Directories.Package
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
