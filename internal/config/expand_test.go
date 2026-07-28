package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
	"gopkg.in/yaml.v3"
)

func TestExtendedConfig_ExpandPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOT_TEST_ROOT", "/srv/dotfiles")
	t.Setenv("DOT_TEST_TILDE", "~/outer")

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "bare tilde becomes home",
			value: "~",
			want:  home,
		},
		{
			name:  "leading tilde is joined with home",
			value: "~/.dotfiles",
			want:  filepath.Join(home, ".dotfiles"),
		},
		{
			name:  "nested tilde path is joined with home",
			value: "~/.local/share/dot/manifest",
			want:  filepath.Join(home, ".local/share/dot/manifest"),
		},
		{
			name:  "bare environment variable is expanded",
			value: "$DOT_TEST_ROOT",
			want:  "/srv/dotfiles",
		},
		{
			name:  "environment variable with suffix is expanded",
			value: "$DOT_TEST_ROOT/packages",
			want:  "/srv/dotfiles/packages",
		},
		{
			name:  "braced environment variable is expanded",
			value: "${DOT_TEST_ROOT}/packages",
			want:  "/srv/dotfiles/packages",
		},
		{
			name:  "HOME variable is expanded",
			value: "$HOME/.dotfiles",
			want:  filepath.Join(home, ".dotfiles"),
		},
		{
			name:  "environment variable expanding to a tilde path is expanded twice",
			value: "$DOT_TEST_TILDE/inner",
			want:  filepath.Join(home, "outer/inner"),
		},
		{
			name:  "absolute path is unchanged",
			value: "/etc/dot",
			want:  "/etc/dot",
		},
		{
			name:  "relative path is unchanged",
			value: ".",
			want:  ".",
		},
		{
			name:  "empty value stays empty",
			value: "",
			want:  "",
		},
		{
			name:  "tilde in the middle is not a home reference",
			value: "/opt/~/cache",
			want:  "/opt/~/cache",
		},
		{
			name:  "named tilde is left alone",
			value: "~alice/.dotfiles",
			want:  "~alice/.dotfiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Directories.Package = tt.value
			cfg.Directories.Target = tt.value
			cfg.Directories.Manifest = tt.value
			cfg.Symlinks.BackupDir = tt.value
			cfg.Logging.File = tt.value

			require.NoError(t, cfg.ExpandPaths())

			assert.Equal(t, tt.want, cfg.Directories.Package, "directories.package")
			assert.Equal(t, tt.want, cfg.Directories.Target, "directories.target")
			assert.Equal(t, tt.want, cfg.Directories.Manifest, "directories.manifest")
			assert.Equal(t, tt.want, cfg.Symlinks.BackupDir, "symlinks.backup_dir")
			assert.Equal(t, tt.want, cfg.Logging.File, "logging.file")
		})
	}
}

func TestExtendedConfig_ExpandPathsLeavesNonPathValues(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := config.DefaultExtended()
	cfg.Dotfile.Prefix = "~dot-"
	cfg.Symlinks.BackupSuffix = "$HOME.bak"
	cfg.Ignore.Patterns = []string{"~/*.local", "$HOME/secret"}
	cfg.Update.Repository = "yaklabco/dot"

	require.NoError(t, cfg.ExpandPaths())

	assert.Equal(t, "~dot-", cfg.Dotfile.Prefix)
	assert.Equal(t, "$HOME.bak", cfg.Symlinks.BackupSuffix)
	assert.Equal(t, []string{"~/*.local", "$HOME/secret"}, cfg.Ignore.Patterns)
	assert.Equal(t, "yaklabco/dot", cfg.Update.Repository)
}

func TestExtendedConfig_ExpandPathsIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := config.DefaultExtended()
	cfg.Directories.Target = "~"
	cfg.Directories.Package = "~/.dotfiles"

	require.NoError(t, cfg.ExpandPaths())
	first := *cfg
	require.NoError(t, cfg.ExpandPaths())

	assert.Equal(t, first.Directories.Target, cfg.Directories.Target)
	assert.Equal(t, first.Directories.Package, cfg.Directories.Package)
}

func TestExtendedConfig_ExpandPathsRejectsUndefinedVariables(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("DOT_TEST_ROOT", "/srv/dotfiles")

	tests := []struct {
		name    string
		value   string
		wantVar string
	}{
		{
			name:    "bare undefined variable",
			value:   "$DOT_TEST_UNDEFINED",
			wantVar: "$DOT_TEST_UNDEFINED",
		},
		{
			name:    "undefined variable with a suffix",
			value:   "$DOT_TEST_UNDEFINED/packages",
			wantVar: "$DOT_TEST_UNDEFINED",
		},
		{
			name:    "braced undefined variable",
			value:   "${DOT_TEST_UNDEFINED}/packages",
			wantVar: "$DOT_TEST_UNDEFINED",
		},
		{
			name:    "undefined variable alongside a defined one",
			value:   "$DOT_TEST_ROOT/$DOT_TEST_UNDEFINED",
			wantVar: "$DOT_TEST_UNDEFINED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Directories.Package = tt.value

			err := cfg.ExpandPaths()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "directories.package")
			assert.Contains(t, err.Error(), tt.wantVar)
			assert.Equal(t, tt.value, cfg.Directories.Package, "value is left as written when expansion fails")
		})
	}
}

func TestExtendedConfig_ExpandPathsReportsUnresolvableHome(t *testing.T) {
	t.Setenv("HOME", "")

	cfg := config.DefaultExtended()
	cfg.Directories.Target = "~"

	err := cfg.ExpandPaths()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "directories.target")
}

func TestLoadExtendedFromFile_RejectsUndefinedVariables(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
directories:
  package: $DOT_TEST_UNDEFINED/repo
  target: "~"
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	_, err := config.LoadExtendedFromFile(configPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "$DOT_TEST_UNDEFINED")
}

func TestLoadExtendedFromFile_ExpandsPathValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOT_TEST_ROOT", "/srv/dotfiles")

	tests := []struct {
		name         string
		content      string
		wantPackage  string
		wantTarget   string
		wantManifest string
		wantBackup   string
		wantLogFile  string
	}{
		{
			name: "tilde paths are expanded to home",
			content: `
directories:
  package: ~/.dotfiles
  target: "~"
  manifest: ~/.local/share/dot/manifest
symlinks:
  backup_dir: ~/.dotfiles.backup
logging:
  file: ~/.local/state/dot/dot.log
`,
			wantPackage:  filepath.Join(home, ".dotfiles"),
			wantTarget:   home,
			wantManifest: filepath.Join(home, ".local/share/dot/manifest"),
			wantBackup:   filepath.Join(home, ".dotfiles.backup"),
			wantLogFile:  filepath.Join(home, ".local/state/dot/dot.log"),
		},
		{
			name: "environment variables are expanded",
			content: `
directories:
  package: $DOT_TEST_ROOT/packages
  target: $HOME
  manifest: ${DOT_TEST_ROOT}/manifest
symlinks:
  backup_dir: $DOT_TEST_ROOT/backup
logging:
  file: $DOT_TEST_ROOT/dot.log
`,
			wantPackage:  "/srv/dotfiles/packages",
			wantTarget:   home,
			wantManifest: "/srv/dotfiles/manifest",
			wantBackup:   "/srv/dotfiles/backup",
			wantLogFile:  "/srv/dotfiles/dot.log",
		},
		{
			name: "absolute paths are preserved verbatim",
			content: `
directories:
  package: /test/dotfiles
  target: /test/home
  manifest: /test/manifest
symlinks:
  backup_dir: /test/backup
logging:
  file: /test/dot.log
`,
			wantPackage:  "/test/dotfiles",
			wantTarget:   "/test/home",
			wantManifest: "/test/manifest",
			wantBackup:   "/test/backup",
			wantLogFile:  "/test/dot.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(configPath, []byte(tt.content), 0o600))

			cfg, err := config.LoadExtendedFromFile(configPath)
			require.NoError(t, err)

			assert.Equal(t, tt.wantPackage, cfg.Directories.Package)
			assert.Equal(t, tt.wantTarget, cfg.Directories.Target)
			assert.Equal(t, tt.wantManifest, cfg.Directories.Manifest)
			assert.Equal(t, tt.wantBackup, cfg.Symlinks.BackupDir)
			assert.Equal(t, tt.wantLogFile, cfg.Logging.File)
		})
	}
}

// TestLoadExtendedFromFile_PortableRepositoryConfig covers the use case from
// issue #80: a repository config committed to a dotfiles repository can now
// name paths relative to the home directory and stay portable across machines.
func TestLoadExtendedFromFile_PortableRepositoryConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoConfigDir := filepath.Join(home, ".dotfiles", ".config", "dot")
	require.NoError(t, os.MkdirAll(repoConfigDir, 0o755))

	configPath := filepath.Join(repoConfigDir, "config.yaml")
	content := `
directories:
  package: ~/.dotfiles
  target: "~"
  manifest: ~/.local/share/dot/manifest

symlinks:
  backup: true
  backup_dir: ~/.dotfiles.backup

dotfile:
  translate: true
  package_name_mapping: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	cfg, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, ".dotfiles"), cfg.Directories.Package)
	assert.Equal(t, home, cfg.Directories.Target)
	assert.Equal(t, filepath.Join(home, ".local/share/dot/manifest"), cfg.Directories.Manifest)
	assert.Equal(t, filepath.Join(home, ".dotfiles.backup"), cfg.Symlinks.BackupDir)
	assert.True(t, cfg.Dotfile.PackageNameMapping)
}

func TestLoader_LoadExpandsPathValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
directories:
  package: ~/.dotfiles
  target: "~"
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	cfg, err := config.NewLoader("dot", configPath).Load()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, ".dotfiles"), cfg.Directories.Package)
	assert.Equal(t, home, cfg.Directories.Target)
}

func TestLoader_LoadWithEnvExpandsPathValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOT_DIRECTORIES_PACKAGE", "~/env-dotfiles")
	t.Setenv("DOT_DIRECTORIES_TARGET", "$HOME/env-target")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
directories:
  package: ~/.dotfiles
  target: "~"
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	cfg, err := config.NewLoader("dot", configPath).LoadWithEnv()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "env-dotfiles"), cfg.Directories.Package)
	assert.Equal(t, filepath.Join(home, "env-target"), cfg.Directories.Target)
}

func TestLoader_LoadWithFlagsExpandsPathValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
directories:
  package: /absolute/dotfiles
  target: /absolute/home
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	flags := map[string]interface{}{
		"dir":    "~/flag-dotfiles",
		"target": "$HOME/flag-target",
	}

	cfg, err := config.NewLoader("dot", configPath).LoadWithFlags(flags)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "flag-dotfiles"), cfg.Directories.Package)
	assert.Equal(t, filepath.Join(home, "flag-target"), cfg.Directories.Target)
}

// writtenPaths mirrors the path-typed subset of a written configuration file.
type writtenPaths struct {
	Directories struct {
		Package  string `yaml:"package"`
		Target   string `yaml:"target"`
		Manifest string `yaml:"manifest"`
	} `yaml:"directories"`
	Symlinks struct {
		BackupDir string `yaml:"backup_dir"`
	} `yaml:"symlinks"`
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
}

func readWrittenPaths(t *testing.T, path string) writtenPaths {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var written writtenPaths
	require.NoError(t, yaml.Unmarshal(data, &written))

	return written
}

const portableConfigContent = `
directories:
  package: ~/.dotfiles
  target: "~"
  manifest: ~/.local/share/dot/manifest

symlinks:
  backup: true
  backup_dir: ~/.dotfiles.backup

logging:
  level: INFO
  file: ~/.local/state/dot/dot.log
`

// TestWriter_UpdatePreservesUnexpandedPathValues guards the round trip: a
// hand-written "~" must survive "dot config set" instead of being frozen into
// the absolute path of whichever machine happened to run the command.
func TestWriter_UpdatePreservesUnexpandedPathValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(portableConfigContent), 0o600))

	require.NoError(t, config.NewWriter(configPath).Update("logging.level", "DEBUG"))

	written := readWrittenPaths(t, configPath)
	assert.Equal(t, "~/.dotfiles", written.Directories.Package)
	assert.Equal(t, "~", written.Directories.Target)
	assert.Equal(t, "~/.local/share/dot/manifest", written.Directories.Manifest)
	assert.Equal(t, "~/.dotfiles.backup", written.Symlinks.BackupDir)
	assert.Equal(t, "~/.local/state/dot/dot.log", written.Logging.File)
	assert.Equal(t, "DEBUG", written.Logging.Level, "the requested update is applied")

	cfg, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".dotfiles"), cfg.Directories.Package)
	assert.Equal(t, home, cfg.Directories.Target)
}

func TestWriter_UpdatePreservesEnvironmentReferences(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOT_TEST_ROOT", filepath.Join(home, "srv"))

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
directories:
  package: $DOT_TEST_ROOT/packages
  target: "~"
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	require.NoError(t, config.NewWriter(configPath).Update("logging.level", "DEBUG"))

	written := readWrittenPaths(t, configPath)
	assert.Equal(t, "$DOT_TEST_ROOT/packages", written.Directories.Package)
}

// TestUpgradeConfig_PreservesUnexpandedPathValues covers the same round trip
// through "dot config upgrade", which rewrites the whole file.
func TestUpgradeConfig_PreservesUnexpandedPathValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(portableConfigContent), 0o600))

	_, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)

	written := readWrittenPaths(t, configPath)
	assert.Equal(t, "~/.dotfiles", written.Directories.Package)
	assert.Equal(t, "~", written.Directories.Target)
	assert.Equal(t, "~/.local/share/dot/manifest", written.Directories.Manifest)
	assert.Equal(t, "~/.dotfiles.backup", written.Symlinks.BackupDir)
	assert.Equal(t, "~/.local/state/dot/dot.log", written.Logging.File)
}
