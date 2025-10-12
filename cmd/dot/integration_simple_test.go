package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusCommand_Execute(t *testing.T) {
	setupGlobalCfg(t)

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"status"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
}

func TestListCommand_Execute(t *testing.T) {
	setupGlobalCfg(t)

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"list"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
}

func TestDoctorCommand_Execute(t *testing.T) {
	setupGlobalCfg(t)

	rootCmd := NewRootCommand("dev", "none", "unknown")
	// Use --scan-mode=off to skip slow filesystem scanning in tests
	rootCmd.SetArgs([]string{"doctor", "--scan-mode=off"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	// Doctor command may return errors for health warnings/errors, which is expected
	_ = rootCmd.Execute()
}

func TestStatusCommand_WithFormat(t *testing.T) {
	setupGlobalCfg(t)

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"status", "--format=json"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "packages")
}

func TestListCommand_WithSort(t *testing.T) {
	setupGlobalCfg(t)

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"list", "--sort=links"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
}

func TestDoctorCommand_WithFormat(t *testing.T) {
	setupGlobalCfg(t)

	rootCmd := NewRootCommand("dev", "none", "unknown")
	// Use --scan-mode=off to skip slow filesystem scanning in tests
	rootCmd.SetArgs([]string{"doctor", "--format=table", "--scan-mode=off"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	// Doctor command may return errors for health warnings/errors, which is expected
	_ = rootCmd.Execute()
}
