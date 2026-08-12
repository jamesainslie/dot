package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
)

func TestLoadFromFile_WithYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
directories:
  package: /test/dotfiles
  target: /test/home

logging:
  level: DEBUG
  format: json
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)

	assert.Equal(t, "/test/dotfiles", cfg.Directories.Package)
	assert.Equal(t, "/test/home", cfg.Directories.Target)
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestNewLoader(t *testing.T) {
	loader := config.NewLoader("dot", "/path/to/config.yaml")
	assert.NotNil(t, loader)
}

func TestLoader_LoadWithMissingFile(t *testing.T) {
	loader := config.NewLoader("dot", "/nonexistent/config.yaml")
	cfg, err := loader.Load()
	require.NoError(t, err)

	// Should return defaults
	assert.NotNil(t, cfg)
	assert.Equal(t, "INFO", cfg.Logging.Level)
}

func TestLoader_LoadWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
logging:
  level: INFO
  format: text
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("DOT_LOGGING_LEVEL", "DEBUG")
	os.Setenv("DOT_LOGGING_FORMAT", "json")
	defer os.Unsetenv("DOT_LOGGING_LEVEL")
	defer os.Unsetenv("DOT_LOGGING_FORMAT")

	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	require.NoError(t, err)

	// Environment should override file
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoader_LoadWithFlags(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
directories:
  package: /file/dotfiles
  target: /file/home

output:
  verbosity: 1
  color: auto
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	loader := config.NewLoader("dot", configPath)
	flags := map[string]interface{}{
		"dir":     "/flag/dotfiles",
		"verbose": 2,
		"color":   "always",
	}

	cfg, err := loader.LoadWithFlags(flags)
	require.NoError(t, err)

	// Flags should override file
	assert.Equal(t, "/flag/dotfiles", cfg.Directories.Package)
	assert.Equal(t, 2, cfg.Output.Verbosity)
	assert.Equal(t, "always", cfg.Output.Color)
	// File value for non-overridden
	assert.Equal(t, "/file/home", cfg.Directories.Target)
}

func TestLoader_Precedence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
directories:
  package: /file/dotfiles

logging:
  level: INFO
  format: text

output:
  verbosity: 1
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("DOT_LOGGING_LEVEL", "WARN")
	defer os.Unsetenv("DOT_LOGGING_LEVEL")

	loader := config.NewLoader("dot", configPath)
	flags := map[string]interface{}{
		"verbose": 2,
	}

	cfg, err := loader.LoadWithFlags(flags)
	require.NoError(t, err)

	// Verify precedence: flags > env > file > default
	assert.Equal(t, "/file/dotfiles", cfg.Directories.Package) // from file
	assert.Equal(t, "WARN", cfg.Logging.Level)                 // from env (overrides file)
	assert.Equal(t, 2, cfg.Output.Verbosity)                   // from flags (highest)
	assert.Equal(t, "text", cfg.Logging.Format)                // from file (no override)
}

func TestLoader_ValidateOnLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create invalid config file
	configContent := `
logging:
  level: INVALID_LEVEL
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	loader := config.NewLoader("dot", configPath)
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestLoader_FlagMapping(t *testing.T) {
	loader := config.NewLoader("dot", "/nonexistent/config.yaml")

	tests := []struct {
		name     string
		flags    map[string]interface{}
		validate func(*testing.T, *config.ExtendedConfig)
	}{
		{
			name: "dir flag",
			flags: map[string]interface{}{
				"dir": "/custom/dotfiles",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "/custom/dotfiles", cfg.Directories.Package)
			},
		},
		{
			name: "target flag",
			flags: map[string]interface{}{
				"target": "/custom/target",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "/custom/target", cfg.Directories.Target)
			},
		},
		{
			name: "dry-run flag",
			flags: map[string]interface{}{
				"dry-run": true,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.True(t, cfg.Operations.DryRun)
			},
		},
		{
			name: "verbose flag",
			flags: map[string]interface{}{
				"verbose": 2,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, 2, cfg.Output.Verbosity)
			},
		},
		{
			name: "quiet flag",
			flags: map[string]interface{}{
				"quiet": true,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, 0, cfg.Output.Verbosity)
			},
		},
		{
			name: "log-json flag",
			flags: map[string]interface{}{
				"log-json": true,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "json", cfg.Logging.Format)
			},
		},
		{
			name: "color flag",
			flags: map[string]interface{}{
				"color": "never",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "never", cfg.Output.Color)
			},
		},
		{
			name: "format flag",
			flags: map[string]interface{}{
				"format": "json",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "json", cfg.Output.Format)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loader.LoadWithFlags(tt.flags)
			require.NoError(t, err)
			tt.validate(t, cfg)
		})
	}
}

func TestLoader_MultipleSourcesIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file with baseline values
	configContent := `
directories:
  package: /file/dotfiles
  target: /file/home

logging:
  level: INFO
  format: text

symlinks:
  mode: relative
  folding: true

output:
  verbosity: 1
  color: auto
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variables to override some file values
	os.Setenv("DOT_LOGGING_LEVEL", "WARN")
	os.Setenv("DOT_SYMLINKS_MODE", "absolute")
	defer os.Unsetenv("DOT_LOGGING_LEVEL")
	defer os.Unsetenv("DOT_SYMLINKS_MODE")

	loader := config.NewLoader("dot", configPath)

	// Load with flags to override env and file
	flags := map[string]interface{}{
		"verbose": 3,
		"color":   "never",
	}

	cfg, err := loader.LoadWithFlags(flags)
	require.NoError(t, err)

	// Verify precedence: flags > env > file > defaults
	assert.Equal(t, "/file/dotfiles", cfg.Directories.Package) // from file (no override)
	assert.Equal(t, "/file/home", cfg.Directories.Target)      // from file (no override)
	assert.Equal(t, "WARN", cfg.Logging.Level)                 // from env (overrides file)
	assert.Equal(t, "text", cfg.Logging.Format)                // from file (no override)
	assert.Equal(t, "absolute", cfg.Symlinks.Mode)             // from env (overrides file)
	assert.True(t, cfg.Symlinks.Folding)                       // from file (no override)
	assert.Equal(t, 3, cfg.Output.Verbosity)                   // from flags (highest priority)
	assert.Equal(t, "never", cfg.Output.Color)                 // from flags (highest priority)
}

func TestLoader_AutoDetectFormat(t *testing.T) {
	tmpDir := t.TempDir()

	formats := []struct {
		ext     string
		content string
	}{
		{
			ext: ".yaml",
			content: `
directories:
  package: /test/dotfiles
`,
		},
		{
			ext: ".json",
			content: `{
  "directories": {
    "package": "/test/dotfiles"
  }
}`,
		},
		{
			ext: ".toml",
			content: `[directories]
package = "/test/dotfiles"
`,
		},
	}

	for _, fmt := range formats {
		t.Run("load from "+fmt.ext, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, "config"+fmt.ext)
			err := os.WriteFile(configPath, []byte(fmt.content), 0600)
			require.NoError(t, err)

			loader := config.NewLoader("dot", configPath)
			cfg, err := loader.Load()
			require.NoError(t, err)
			assert.Equal(t, "/test/dotfiles", cfg.Directories.Package)
		})
	}
}

// envSampleValues holds one valid, non-default sample per configuration key,
// in the string form an operator would export. TestLoader_EnvOverridesEveryConfigKey
// walks the ExtendedConfig struct, so a new key without a sample here fails
// loudly instead of shipping as a silently ignored env var (issue #86).
var envSampleValues = map[string]string{
	"directories.package":            "/env/pkgs",
	"directories.target":             "/env/target",
	"directories.manifest":           "/env/manifest",
	"logging.level":                  "DEBUG",
	"logging.format":                 "json",
	"logging.destination":            "stdout",
	"logging.file":                   "/env/dot.log",
	"symlinks.mode":                  "absolute",
	"symlinks.folding":               "false",
	"symlinks.overwrite":             "true",
	"symlinks.backup":                "true",
	"symlinks.backup_suffix":         ".orig",
	"symlinks.backup_dir":            "/env/backups",
	"ignore.use_defaults":            "false",
	"ignore.patterns":                "*.log",
	"ignore.overrides":               "keep.me",
	"ignore.per_package_ignore":      "false",
	"ignore.max_file_size":           "1024",
	"ignore.interactive_large_files": "false",
	"dotfile.translate":              "false",
	"dotfile.prefix":                 "custom-",
	"dotfile.package_name_mapping":   "false",
	"output.format":                  "json",
	"output.color":                   "always",
	"output.table_style":             "simple",
	"output.progress":                "false",
	"output.verbosity":               "2",
	"output.width":                   "120",
	"operations.dry_run":             "true",
	"operations.atomic":              "false",
	"operations.max_parallel":        "4",
	"packages.sort_by":               "links",
	"packages.auto_discover":         "false",
	"packages.validate_names":        "false",
	"doctor.auto_fix":                "true",
	"doctor.check_manifest":          "false",
	"doctor.check_broken_links":      "false",
	"doctor.check_orphaned":          "false",
	"doctor.check_permissions":       "false",
	"update.check_on_startup":        "false",
	"update.check_frequency":         "48",
	"update.package_manager":         "brew",
	"update.repository":              "acme/dot",
	"update.include_prerelease":      "true",
	"network.http_proxy":             "http://proxy.env:3128",
	"network.https_proxy":            "http://proxy.env:3129",
	"network.no_proxy":               "localhost",
	"network.timeout":                "30",
	"network.connect_timeout":        "9",
	"network.tls_timeout":            "7",
	"experimental.parallel":          "true",
	"experimental.profiling":         "true",
}

// configLeafKeys walks the two-level ExtendedConfig shape and returns every
// section.field key together with the reflect.Value accessor for a loaded
// config, so the test cannot drift from the struct.
func configLeafKeys(t *testing.T, cfg *config.ExtendedConfig) map[string]reflect.Value {
	t.Helper()

	keys := make(map[string]reflect.Value)
	root := reflect.ValueOf(cfg).Elem()
	rootType := root.Type()

	for i := 0; i < rootType.NumField(); i++ {
		section := rootType.Field(i).Tag.Get("mapstructure")
		require.NotEmpty(t, section, "field %s has no mapstructure tag", rootType.Field(i).Name)

		sectionValue := root.Field(i)
		sectionType := sectionValue.Type()
		for j := 0; j < sectionType.NumField(); j++ {
			field := sectionType.Field(j).Tag.Get("mapstructure")
			require.NotEmpty(t, field, "field %s.%s has no mapstructure tag", section, sectionType.Field(j).Name)
			keys[section+"."+field] = sectionValue.Field(j)
		}
	}

	return keys
}

func TestLoader_EnvOverridesEveryConfigKey(t *testing.T) {
	// Issue #86: DOT_ env overrides were driven by hand-lists, so keys
	// outside them (dotfile.package_name_mapping among others) were
	// silently ignored. Every key in the config struct must be settable
	// from the environment.
	for key := range configLeafKeys(t, config.DefaultExtended()) {
		t.Run(key, func(t *testing.T) {
			sample, ok := envSampleValues[key]
			require.True(t, ok, "config key %s has no env sample; add one to envSampleValues", key)

			envVar := "DOT_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
			t.Setenv(envVar, sample)

			loader := config.NewLoader("dot", filepath.Join(t.TempDir(), "missing.yaml"))
			cfg, err := loader.LoadWithEnv()
			require.NoError(t, err)

			got := configLeafKeys(t, cfg)[key]

			var want interface{}
			switch got.Kind() {
			case reflect.Bool:
				b, perr := strconv.ParseBool(sample)
				require.NoError(t, perr)
				want = b
			case reflect.Int:
				n, perr := strconv.Atoi(sample)
				require.NoError(t, perr)
				want = n
			case reflect.Int64:
				n, perr := strconv.ParseInt(sample, 10, 64)
				require.NoError(t, perr)
				want = n
			case reflect.String:
				want = sample
			case reflect.Slice:
				want = []string{sample}
			default:
				t.Fatalf("unhandled kind %s for %s", got.Kind(), key)
			}

			// A sample equal to the default would make this subtest pass
			// even if the override were silently dropped.
			def := configLeafKeys(t, config.DefaultExtended())[key]
			require.NotEqual(t, def.Interface(), want,
				"sample for %s equals the default; pick a distinct value in envSampleValues", key)

			assert.Equal(t, want, got.Interface(), "%s not overridden by %s", key, envVar)
		})
	}
}

func TestLoader_EnvListValuesAreCommaSeparated(t *testing.T) {
	// The documented contract (and the config set convention) is
	// comma-separated lists; viper's own env handling would split on
	// whitespace instead, turning "*.log,*.tmp" into one bogus pattern.
	t.Setenv("DOT_IGNORE_PATTERNS", "*.log, *.tmp,cache")

	loader := config.NewLoader("dot", filepath.Join(t.TempDir(), "missing.yaml"))
	cfg, err := loader.LoadWithEnv()
	require.NoError(t, err)

	assert.Equal(t, []string{"*.log", "*.tmp", "cache"}, cfg.Ignore.Patterns)
}

func TestLoader_EnvFalseOverridesFileTrue(t *testing.T) {
	// The old sparse-config merge could only propagate true booleans, so an
	// env var set to false could never override a file value of true. That
	// is the exact shape of the issue #86 report.
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
dotfile:
  translate: true
  package_name_mapping: true
symlinks:
  folding: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	t.Setenv("DOT_DOTFILE_PACKAGE_NAME_MAPPING", "false")
	t.Setenv("DOT_DOTFILE_TRANSLATE", "false")
	t.Setenv("DOT_SYMLINKS_FOLDING", "false")

	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	require.NoError(t, err)

	assert.False(t, cfg.Dotfile.PackageNameMapping, "env false must override file true")
	assert.False(t, cfg.Dotfile.Translate, "env false must override file true")
	assert.False(t, cfg.Symlinks.Folding, "env false must override file true")
}
