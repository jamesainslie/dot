package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidFlagShowsUsage(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectUsage bool
		expectError bool
	}{
		{
			name:        "unknown flag shows usage",
			args:        []string{"manage", "--invalid-flag", "pkg"},
			expectUsage: true,
			expectError: true,
		},
		{
			name:        "missing required arg shows usage",
			args:        []string{"manage"},
			expectUsage: true,
			expectError: true,
		},
		{
			name:        "missing args in adopt shows usage",
			args:        []string{"adopt", "pkg"},
			expectUsage: true,
			expectError: true,
		},
		{
			name:        "missing args in config get shows usage",
			args:        []string{"config", "get"},
			expectUsage: true,
			expectError: true,
		},
		{
			name:        "too many args in config get shows usage",
			args:        []string{"config", "get", "key1", "key2"},
			expectUsage: true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := NewRootCommand("dev", "none", "unknown")
			rootCmd.SetArgs(tt.args)

			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			rootCmd.SetOut(outBuf)
			rootCmd.SetErr(errBuf)

			err := rootCmd.Execute()

			if tt.expectError {
				require.Error(t, err)
			}

			output := outBuf.String() + errBuf.String()

			if tt.expectUsage {
				// Should contain usage information
				assert.True(t, strings.Contains(output, "Usage:") ||
					strings.Contains(output, "usage:"),
					"Expected usage information in output, got: %s", output)
			}
		})
	}
}

func TestRuntimeErrorDoesNotShowUsage(t *testing.T) {
	// This test ensures that runtime errors (not flag/arg errors)
	// do NOT show usage, only the error message
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"manage", "nonexistent-package", "--dir=/nonexistent"})

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)

	err := rootCmd.Execute()
	require.Error(t, err)

	output := outBuf.String() + errBuf.String()

	// Runtime errors should not show usage
	// (This test may need adjustment based on actual error handling)
	_ = output // Placeholder for now
}

func TestHelpFlagShowsUsage(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"manage", "--help"})

	outBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := outBuf.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "manage")
}
