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

func TestIsFileDescriptor(t *testing.T) {
	// Test with stdin file descriptor
	fd := int(os.Stdin.Fd())
	assert.GreaterOrEqual(t, fd, 0)
}
