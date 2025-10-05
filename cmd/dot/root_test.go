package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommand_Version(t *testing.T) {
	rootCmd := NewRootCommand("1.0.0", "abc123", "2025-01-01")
	rootCmd.SetArgs([]string{"--version"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "1.0.0")
	require.Contains(t, out.String(), "abc123")
	require.Contains(t, out.String(), "2025-01-01")
}

func TestRootCommand_Help(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "dot")
	require.Contains(t, out.String(), "manage")
	require.Contains(t, out.String(), "GNU Stow replacement")
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")

	// Verify global flags exist
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("dir"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("target"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("dry-run"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("quiet"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("log-json"))
}

func TestRootCommand_ShortFlags(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")

	// Verify short flag aliases exist
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("d"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("t"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("n"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("v"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("q"))
}
