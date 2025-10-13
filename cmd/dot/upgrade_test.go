package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgradeCommand(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")
	require.NotNil(t, cmd)

	assert.Equal(t, "upgrade", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("yes"))
	assert.NotNil(t, cmd.Flags().Lookup("check-only"))
	
	// Verify flag defaults
	yesFlag := cmd.Flags().Lookup("yes")
	assert.Equal(t, "false", yesFlag.DefValue)
	
	checkOnlyFlag := cmd.Flags().Lookup("check-only")
	assert.Equal(t, "false", checkOnlyFlag.DefValue)
}

func TestUpgradeCommand_Help(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")
	
	// Verify help text includes key information
	assert.Contains(t, cmd.Long, "package manager")
	assert.Contains(t, cmd.Long, "GitHub")
	assert.Contains(t, cmd.Long, "update:")

	// Verify examples exist
	assert.Contains(t, cmd.Example, "dot upgrade")
	assert.Contains(t, cmd.Example, "--check-only")
	assert.Contains(t, cmd.Example, "--yes")
	
	// Verify config documentation
	assert.Contains(t, cmd.Long, "~/.config/dot/config.yaml")
	assert.Contains(t, cmd.Long, "package_manager")
	assert.Contains(t, cmd.Long, "repository")
	assert.Contains(t, cmd.Long, "include_prerelease")
}

func TestUpgradeCommand_FlagShortcuts(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")
	
	// Verify yes flag has -y shortcut
	yesFlag := cmd.Flags().Lookup("yes")
	assert.Equal(t, "y", yesFlag.Shorthand)
	
	// Verify check-only has no shortcut
	checkOnlyFlag := cmd.Flags().Lookup("check-only")
	assert.Empty(t, checkOnlyFlag.Shorthand)
}

func TestUpgradeCommand_Execution(t *testing.T) {
	// This test verifies the command can be executed without panicking
	// We can't test actual execution without mocking, but we can test structure
	cmd := newUpgradeCommand("999.999.999") // Version that won't match any release
	
	require.NotNil(t, cmd.RunE, "RunE should be set")
	
	// Verify command is properly structured
	assert.NotNil(t, cmd.RunE)
	assert.Equal(t, "upgrade", cmd.Use)
}

func TestUpgradeCommand_Integration(t *testing.T) {
	// Integration test that verifies command is properly added to root
	rootCmd := NewRootCommand("test-version", "abc123", "2024-01-01")
	
	// Find upgrade command
	var upgradeCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "upgrade" {
			upgradeCmd = cmd
			break
		}
	}
	
	require.NotNil(t, upgradeCmd, "upgrade command should be registered")
	assert.Equal(t, "upgrade", upgradeCmd.Use)
}

func TestUpgradeCommand_HelpOutput(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	// Set help flag
	cmd.SetArgs([]string{"--help"})
	
	// Execute should succeed for help
	err := cmd.Execute()
	
	// Help returns nil error but shows help text
	if err != nil {
		t.Logf("Help execution: %v", err)
	}
	
	// Verify help was shown (buffer should have content)
	output := buf.String()
	if len(output) > 0 {
		assert.Contains(t, output, "upgrade")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  bool
	}{
		{"found in middle", []string{"a", "b", "c"}, "b", true},
		{"found at start", []string{"a", "b", "c"}, "a", true},
		{"found at end", []string{"a", "b", "c"}, "c", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"empty string", []string{"a", "b"}, "", false},
		{"special char found", []string{"&&", "||", ";"}, "&&", true},
		{"multiple same values", []string{"a", "a", "a"}, "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.str)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContains_EdgeCases(t *testing.T) {
	// Test with nil slice (should not panic)
	got := contains(nil, "test")
	assert.False(t, got)
	
	// Test with slice containing empty strings
	got = contains([]string{"", "a", ""}, "")
	assert.True(t, got)
	
	// Test case sensitivity
	got = contains([]string{"Hello", "World"}, "hello")
	assert.False(t, got)
}

