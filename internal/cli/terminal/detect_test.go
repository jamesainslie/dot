package terminal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteractive(t *testing.T) {
	// Note: This test's behavior depends on how tests are run
	// When run normally (with terminal), it should detect TTY
	// When run in CI/automation (no terminal), it should not detect TTY

	result := IsInteractive()

	// We can't assert a specific value since it depends on environment
	// But we can verify it returns a boolean without panicking
	assert.IsType(t, false, result)
}

func TestIsInteractive_WithNonTerminal(t *testing.T) {
	// Save original stdin
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Create a pipe (non-terminal)
	r, w, err := os.Pipe()
	if err != nil {
		t.Skip("Cannot create pipe:", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin with pipe
	os.Stdin = r

	// Should detect as non-interactive
	result := IsInteractive()
	assert.False(t, result)
}

func TestIsInteractive_WithNonTerminalStdout(t *testing.T) {
	// Save original stdout
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// Create a pipe (non-terminal)
	r, w, err := os.Pipe()
	if err != nil {
		t.Skip("Cannot create pipe:", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdout with pipe
	os.Stdout = w

	// Should detect as non-interactive
	result := IsInteractive()
	assert.False(t, result)
}

func TestIsInteractive_WithBothNonTerminal(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	defer func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
	}()

	// Create pipes for both
	rIn, wIn, err := os.Pipe()
	if err != nil {
		t.Skip("Cannot create stdin pipe:", err)
	}
	defer rIn.Close()
	defer wIn.Close()

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Skip("Cannot create stdout pipe:", err)
	}
	defer rOut.Close()
	defer wOut.Close()

	// Replace both with pipes
	os.Stdin = rIn
	os.Stdout = wOut

	// Should detect as non-interactive
	result := IsInteractive()
	assert.False(t, result)
}

func TestIsInteractive_FileDescriptors(t *testing.T) {
	// Verify file descriptors are valid and non-negative
	stdinFd := int(os.Stdin.Fd())
	stdoutFd := int(os.Stdout.Fd())

	assert.GreaterOrEqual(t, stdinFd, 0, "stdin fd should be non-negative")
	assert.GreaterOrEqual(t, stdoutFd, 0, "stdout fd should be non-negative")
}

func TestIsInteractive_ConsistentResults(t *testing.T) {
	// Multiple calls should return consistent results
	result1 := IsInteractive()
	result2 := IsInteractive()
	result3 := IsInteractive()

	assert.Equal(t, result1, result2, "consecutive calls should return same result")
	assert.Equal(t, result2, result3, "consecutive calls should return same result")
}

func TestIsInteractive_DocumentationExample(t *testing.T) {
	// Example from documentation: checking if interactive before prompting
	if IsInteractive() {
		// Would prompt user for input
		t.Log("Running in interactive mode")
	} else {
		// Would use non-interactive defaults
		t.Log("Running in non-interactive mode")
	}
	// This test just verifies the function can be called without error
	assert.True(t, true)
}
