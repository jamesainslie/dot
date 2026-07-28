package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneBootstrapCommand_Structure(t *testing.T) {
	cmd := newCloneBootstrapCommand()

	assert.Equal(t, "bootstrap", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestCloneBootstrapCommand_Flags(t *testing.T) {
	cmd := newCloneBootstrapCommand()

	// Verify flags exist
	outputFlag := cmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "o", outputFlag.Shorthand)

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	require.NotNil(t, dryRunFlag)

	manifestFlag := cmd.Flags().Lookup("from-manifest")
	require.NotNil(t, manifestFlag)

	policyFlag := cmd.Flags().Lookup("conflict-policy")
	require.NotNil(t, policyFlag)

	forceFlag := cmd.Flags().Lookup("force")
	require.NotNil(t, forceFlag)
}

func TestCloneBootstrapCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	// Setup package directories
	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "vim"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "zsh"), 0755))

	// Create output path
	outputPath := filepath.Join(tmpDir, ".dotbootstrap.yaml")

	// Create root command with test configuration
	rootCmd := NewRootCommand("test", "none", "unknown")
	rootCmd.SetContext(context.Background())

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Set args
	rootCmd.SetArgs([]string{
		"--dir", packageDir,
		"--target", targetDir,
		"clone", "bootstrap",
		"--output", outputPath,
	})

	// Execute
	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Verify file content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "version:")
	assert.Contains(t, content, "packages:")
	assert.Contains(t, content, "name: vim")
	assert.Contains(t, content, "name: zsh")
	assert.Contains(t, content, "Bootstrap configuration")
}

func TestCloneBootstrapCommand_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	// Setup package directories
	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "vim"), 0755))

	// Create root command with test configuration
	rootCmd := NewRootCommand("test", "none", "unknown")
	rootCmd.SetContext(context.Background())

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Set args for dry-run
	rootCmd.SetArgs([]string{
		"--dir", packageDir,
		"--target", targetDir,
		"clone", "bootstrap",
		"--dry-run",
	})

	// Execute
	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := buf.String()
	assert.Contains(t, output, "version:")
	assert.Contains(t, output, "packages:")
	assert.Contains(t, output, "name: vim")

	// Verify no file was created
	outputPath := filepath.Join(packageDir, ".dotbootstrap.yaml")
	_, err = os.Stat(outputPath)
	assert.True(t, os.IsNotExist(err))
}

func TestCloneBootstrapCommand_Force(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	// Setup package directories
	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "vim"), 0755))

	outputPath := filepath.Join(packageDir, ".dotbootstrap.yaml")

	// Create existing file
	require.NoError(t, os.WriteFile(outputPath, []byte("existing content"), 0644))

	// Create root command with test configuration
	rootCmd := NewRootCommand("test", "none", "unknown")
	rootCmd.SetContext(context.Background())

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Set args with force flag
	rootCmd.SetArgs([]string{
		"--dir", packageDir,
		"--target", targetDir,
		"clone", "bootstrap",
		"--force",
	})

	// Execute
	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify file was overwritten
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "version:")
	assert.Contains(t, content, "packages:")
	assert.NotContains(t, content, "existing content")
}

func TestCloneBootstrapCommand_NoPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory with no packages
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	// Create root command
	rootCmd := NewRootCommand("test", "none", "unknown")
	rootCmd.SetContext(context.Background())

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Set args
	rootCmd.SetArgs([]string{
		"--dir", packageDir,
		"--target", targetDir,
		"clone", "bootstrap",
	})

	// Execute - should fail
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no packages")
}

// runBootstrapForRepo runs "clone bootstrap" against a prepared repository.
func runBootstrapForRepo(t *testing.T, packageDir, targetDir string) string {
	t.Helper()

	rootCmd := NewRootCommand("test", "none", "unknown")
	rootCmd.SetContext(context.Background())

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"--dir", packageDir,
		"--target", targetDir,
		"clone", "bootstrap",
	})

	require.NoError(t, rootCmd.Execute())
	return buf.String()
}

func TestCloneBootstrapCommand_FullTreeRepoConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		name          string
		contents      map[string]string
		existing      string
		wantRepoCfg   bool
		wantUntouched bool
	}{
		{
			name: "full tree layout writes repo config",
			contents: map[string]string{
				"nvim": ".config",
				"zsh":  ".zshrc",
			},
			wantRepoCfg: true,
		},
		{
			name: "prefixed layout writes nothing",
			contents: map[string]string{
				"dot-vim": "dot-vimrc",
			},
			wantRepoCfg: false,
		},
		{
			name: "existing repo config is left alone",
			contents: map[string]string{
				"nvim": ".config",
			},
			existing:      "dotfile:\n  package_name_mapping: true\n",
			wantRepoCfg:   true,
			wantUntouched: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			packageDir := filepath.Join(tmpDir, "packages")
			targetDir := filepath.Join(tmpDir, "target")
			require.NoError(t, os.MkdirAll(packageDir, 0755))
			require.NoError(t, os.MkdirAll(targetDir, 0755))

			for pkg, entry := range tt.contents {
				require.NoError(t, os.MkdirAll(filepath.Join(packageDir, pkg, entry), 0755))
			}

			repoConfigPath := filepath.Join(packageDir, ".config", "dot", "config.yaml")
			if tt.existing != "" {
				require.NoError(t, os.MkdirAll(filepath.Dir(repoConfigPath), 0755))
				require.NoError(t, os.WriteFile(repoConfigPath, []byte(tt.existing), 0644))
			}

			runBootstrapForRepo(t, packageDir, targetDir)

			data, err := os.ReadFile(repoConfigPath)
			if !tt.wantRepoCfg {
				assert.True(t, os.IsNotExist(err), "repo config should not be created")
				return
			}

			require.NoError(t, err)
			if tt.wantUntouched {
				assert.Equal(t, tt.existing, string(data))
				return
			}
			assert.Contains(t, string(data), "package_name_mapping: false")
		})
	}
}
