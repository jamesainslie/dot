package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Version(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"version", []string{"dot", "version"}},
		{"--version", []string{"dot", "--version"}},
		{"-v", []string{"dot", "-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := run(tt.args, &stdout, &stderr)

			assert.Equal(t, 0, exitCode, "Expected exit code 0")
			assert.Empty(t, stderr.String(), "Expected no stderr output")

			output := stdout.String()
			assert.Contains(t, output, "dot version")
			assert.Contains(t, output, "commit:")
			assert.Contains(t, output, "built:")
		})
	}
}

func TestRun_Help(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"help", []string{"dot", "help"}},
		{"--help", []string{"dot", "--help"}},
		{"-h", []string{"dot", "-h"}},
		{"no args", []string{"dot"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := run(tt.args, &stdout, &stderr)

			assert.Equal(t, 0, exitCode, "Expected exit code 0")
			assert.Empty(t, stderr.String(), "Expected no stderr output")

			output := stdout.String()
			assert.Contains(t, output, "dot - dotfile manager")
			assert.Contains(t, output, "Usage:")
			assert.Contains(t, output, "version")
			assert.Contains(t, output, "help")
			assert.Contains(t, output, "manage")
			assert.Contains(t, output, "unmanage")
			assert.Contains(t, output, "remanage")
		})
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"dot", "unknown-command"}, &stdout, &stderr)

	assert.Equal(t, 1, exitCode, "Expected exit code 1 for unknown command")

	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "Error: unknown command")
	assert.Contains(t, stderrOutput, "unknown-command")

	// Help should be printed to stdout after error
	stdoutOutput := stdout.String()
	assert.Contains(t, stdoutOutput, "dot - dotfile manager")
	assert.Contains(t, stdoutOutput, "Usage:")
}

func TestPrintVersion(t *testing.T) {
	var buf bytes.Buffer

	printVersion(&buf)

	output := buf.String()
	assert.Contains(t, output, "dot version")
	assert.Contains(t, output, "commit:")
	assert.Contains(t, output, "built:")

	// Check default values
	assert.Contains(t, output, "dev")
	assert.Contains(t, output, "unknown")
}

func TestPrintHelp(t *testing.T) {
	var buf bytes.Buffer

	printHelp(&buf)

	output := buf.String()

	// Verify key components of help output
	assert.Contains(t, output, "dot - dotfile manager")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Available commands:")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "help")
	assert.Contains(t, output, "Commands under development:")
	assert.Contains(t, output, "manage")
	assert.Contains(t, output, "unmanage")
	assert.Contains(t, output, "remanage")

	// Check structure
	lines := strings.Split(output, "\n")
	require.Greater(t, len(lines), 5, "Help output should have multiple lines")
	assert.True(t, strings.HasPrefix(lines[0], "dot - dotfile manager"))
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables exist and have default values
	require.NotEmpty(t, version)
	require.NotEmpty(t, commit)
	require.NotEmpty(t, date)

	// Default values should be set
	assert.Equal(t, "dev", version)
	assert.Equal(t, "unknown", commit)
	assert.Equal(t, "unknown", date)
}

func TestRun_MultipleCommands(t *testing.T) {
	// Test that we handle each command type correctly
	commands := []struct {
		name       string
		args       []string
		expectExit int
		checkOut   string
	}{
		{"version command", []string{"dot", "version"}, 0, "dot version"},
		{"help command", []string{"dot", "help"}, 0, "dot - dotfile manager"},
		{"no args", []string{"dot"}, 0, "Usage:"},
		{"unknown command", []string{"dot", "invalid"}, 1, ""},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := run(tt.args, &stdout, &stderr)

			assert.Equal(t, tt.expectExit, exitCode)
			if tt.checkOut != "" {
				assert.Contains(t, stdout.String(), tt.checkOut)
			}
		})
	}
}

func TestRun_StderrOnlyForErrors(t *testing.T) {
	// Test that valid commands don't write to stderr
	validCommands := [][]string{
		{"dot", "version"},
		{"dot", "help"},
		{"dot"},
	}

	for _, args := range validCommands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			run(args, &stdout, &stderr)

			assert.Empty(t, stderr.String(), "Valid commands should not write to stderr")
			assert.NotEmpty(t, stdout.String(), "Valid commands should write to stdout")
		})
	}

	// Test that invalid commands write to stderr
	var stdout, stderr bytes.Buffer
	run([]string{"dot", "invalid"}, &stdout, &stderr)

	assert.NotEmpty(t, stderr.String(), "Invalid commands should write to stderr")
	assert.NotEmpty(t, stdout.String(), "Invalid commands should also show help on stdout")
}
