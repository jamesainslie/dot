package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
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
			name:  "undefined variable expands to empty",
			value: "$DOT_TEST_UNDEFINED/packages",
			want:  "/packages",
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

			cfg.ExpandPaths()

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

	cfg.ExpandPaths()

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

	cfg.ExpandPaths()
	first := *cfg
	cfg.ExpandPaths()

	assert.Equal(t, first.Directories.Target, cfg.Directories.Target)
	assert.Equal(t, first.Directories.Package, cfg.Directories.Package)
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
