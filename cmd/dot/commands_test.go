package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupGlobalCfg initializes globalCfg with deterministic test values and registers cleanup.
func setupGlobalCfg(t *testing.T) {
	t.Helper()

	// Save previous globalCfg
	previous := globalCfg

	// Set globalCfg to use temporary directories
	globalCfg = globalConfig{
		packageDir: t.TempDir(),
		targetDir:  t.TempDir(),
		dryRun:     true, // Always dry-run in tests to avoid side effects
		verbose:    0,
		quiet:      false,
		logJSON:    false,
	}

	// Restore previous globalCfg on cleanup
	t.Cleanup(func() {
		globalCfg = previous
	})
}

func TestManageCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	// Commands now actually execute, so we expect error for non-existent package
	cmd := newManageCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestManageCommand_NoPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newManageCommand()
	cmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestManageCommand_Metadata(t *testing.T) {
	cmd := newManageCommand()

	require.Equal(t, "manage PACKAGE [PACKAGE...]", cmd.Use)
	require.Equal(t, "Install packages by creating symlinks", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestUnmanageCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestUnmanageCommand_NoPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestUnmanageCommand_Metadata(t *testing.T) {
	cmd := newUnmanageCommand()

	require.Equal(t, "unmanage PACKAGE [PACKAGE...]", cmd.Use)
	require.Equal(t, "Remove packages by deleting symlinks", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestRemanageCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	// Remanage tries to unmanage then manage, so will error on manage phase
	cmd := newRemanageCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestRemanageCommand_NoPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newRemanageCommand()
	cmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestRemanageCommand_Metadata(t *testing.T) {
	cmd := newRemanageCommand()

	require.Equal(t, "remanage PACKAGE [PACKAGE...]", cmd.Use)
	require.Equal(t, "Reinstall packages with incremental updates", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestAdoptCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	// Adopt tries to verify package exists, so will error
	cmd := newAdoptCommand()
	cmd.SetArgs([]string{"package1", "file1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestAdoptCommand_NotEnoughArgs(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newAdoptCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestAdoptCommand_Metadata(t *testing.T) {
	cmd := newAdoptCommand()

	require.Equal(t, "adopt PACKAGE FILE [FILE...]", cmd.Use)
	require.Equal(t, "Move existing files into package then link", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestAdoptCommand_MultipleFiles(t *testing.T) {
	setupGlobalCfg(t)

	// Multiple files with non-existent package will error
	cmd := newAdoptCommand()
	cmd.SetArgs([]string{"package1", "file1", "file2", "file3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestManageCommand_MultiplePackages(t *testing.T) {
	setupGlobalCfg(t)

	// Multiple packages that don't exist will error
	cmd := newManageCommand()
	cmd.SetArgs([]string{"package1", "package2", "package3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when packages do not exist")
}

func TestUnmanageCommand_MultiplePackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{"package1", "package2", "package3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestRemanageCommand_MultiplePackages(t *testing.T) {
	setupGlobalCfg(t)

	// Multiple packages that don't exist will error on manage phase
	cmd := newRemanageCommand()
	cmd.SetArgs([]string{"package1", "package2", "package3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when packages do not exist")
}

func TestRootCommand_NoArgs(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "dot")
}

func TestRootCommand_WithManageCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"manage", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "manage")
}

func TestRootCommand_WithUnmanageCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"unmanage", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "unmanage")
}

func TestRootCommand_WithRemanageCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"remanage", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "remanage")
}

func TestRootCommand_WithAdoptCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"adopt", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "adopt")
}

func TestRootCommand_GlobalFlagsWithCommand(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--dir", tmpDir, "--target", tmpDir, "--dry-run", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_DryRunFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_VerboseFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "-vvv", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_QuietFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "--quiet", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_LogJSONFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "--log-json", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}
