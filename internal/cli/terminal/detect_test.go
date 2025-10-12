package terminal

import (
	"io"
	"os"
	"testing"

	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIsInteractive_WithPseudoTerminalStdin(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	defer func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
	}()

	// Create a pseudo-terminal for stdin
	ptyStdin, ttyStdin, err := pty.Open()
	if err != nil {
		t.Skip("Cannot create pty:", err)
	}
	defer ptyStdin.Close()
	defer ttyStdin.Close()

	// Replace stdin with pty (is a terminal)
	os.Stdin = ttyStdin

	// Create a pipe for stdout (not a terminal)
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()
	os.Stdout = w

	// stdin is terminal, stdout is not → should return false
	result := IsInteractive()
	assert.False(t, result, "should return false when stdout is not a terminal")
}

func TestIsInteractive_WithBothPseudoTerminals(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	defer func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
	}()

	// Create pseudo-terminals for both stdin and stdout
	ptyStdin, ttyStdin, err := pty.Open()
	if err != nil {
		t.Skip("Cannot create stdin pty:", err)
	}
	defer ptyStdin.Close()
	defer ttyStdin.Close()

	ptyStdout, ttyStdout, err := pty.Open()
	if err != nil {
		t.Skip("Cannot create stdout pty:", err)
	}
	defer ptyStdout.Close()
	defer ttyStdout.Close()

	// Replace both with ptys (both are terminals)
	os.Stdin = ttyStdin
	os.Stdout = ttyStdout

	// Both are terminals → should return true
	result := IsInteractive()
	assert.True(t, result, "should return true when both stdin and stdout are terminals")
}

func TestIsInteractive_WithPseudoTerminalStdout(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	defer func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
	}()

	// Create a pipe for stdin (not a terminal)
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()
	os.Stdin = r

	// Create a pseudo-terminal for stdout
	ptyStdout, ttyStdout, err := pty.Open()
	if err != nil {
		t.Skip("Cannot create pty:", err)
	}
	defer ptyStdout.Close()
	defer ttyStdout.Close()

	os.Stdout = ttyStdout

	// stdin is not terminal, stdout is → should return false (checks stdin first)
	result := IsInteractive()
	assert.False(t, result, "should return false when stdin is not a terminal")
}

func TestIsInteractive_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupStdin  func() (*os.File, func())
		setupStdout func() (*os.File, func())
		expected    bool
		skipReason  string
		description string
	}{
		{
			name: "stdin_pty_stdout_pty",
			setupStdin: func() (*os.File, func()) {
				pty, tty, err := pty.Open()
				if err != nil {
					return nil, nil
				}
				return tty, func() { pty.Close(); tty.Close() }
			},
			setupStdout: func() (*os.File, func()) {
				pty, tty, err := pty.Open()
				if err != nil {
					return nil, nil
				}
				return tty, func() { pty.Close(); tty.Close() }
			},
			expected:    true,
			description: "both stdin and stdout are terminals",
		},
		{
			name: "stdin_pipe_stdout_pty",
			setupStdin: func() (*os.File, func()) {
				r, w, err := os.Pipe()
				if err != nil {
					return nil, nil
				}
				return r, func() { r.Close(); w.Close() }
			},
			setupStdout: func() (*os.File, func()) {
				pty, tty, err := pty.Open()
				if err != nil {
					return nil, nil
				}
				return tty, func() { pty.Close(); tty.Close() }
			},
			expected:    false,
			description: "stdin is not a terminal",
		},
		{
			name: "stdin_pty_stdout_pipe",
			setupStdin: func() (*os.File, func()) {
				pty, tty, err := pty.Open()
				if err != nil {
					return nil, nil
				}
				return tty, func() { pty.Close(); tty.Close() }
			},
			setupStdout: func() (*os.File, func()) {
				r, w, err := os.Pipe()
				if err != nil {
					return nil, nil
				}
				return w, func() { r.Close(); w.Close() }
			},
			expected:    false,
			description: "stdout is not a terminal",
		},
		{
			name: "stdin_pipe_stdout_pipe",
			setupStdin: func() (*os.File, func()) {
				r, w, err := os.Pipe()
				if err != nil {
					return nil, nil
				}
				return r, func() { r.Close(); w.Close() }
			},
			setupStdout: func() (*os.File, func()) {
				r, w, err := os.Pipe()
				if err != nil {
					return nil, nil
				}
				return w, func() { r.Close(); w.Close() }
			},
			expected:    false,
			description: "neither stdin nor stdout are terminals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save originals
			originalStdin := os.Stdin
			originalStdout := os.Stdout
			defer func() {
				os.Stdin = originalStdin
				os.Stdout = originalStdout
			}()

			// Setup stdin
			stdin, cleanupStdin := tt.setupStdin()
			if stdin == nil {
				t.Skip("Cannot setup stdin:", tt.skipReason)
			}
			defer cleanupStdin()
			os.Stdin = stdin

			// Setup stdout
			stdout, cleanupStdout := tt.setupStdout()
			if stdout == nil {
				t.Skip("Cannot setup stdout:", tt.skipReason)
			}
			defer cleanupStdout()
			os.Stdout = stdout

			// Test
			result := IsInteractive()
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestIsInteractive_NilChecks(t *testing.T) {
	// Verify that os.Stdin and os.Stdout are not nil
	require.NotNil(t, os.Stdin, "os.Stdin should not be nil")
	require.NotNil(t, os.Stdout, "os.Stdout should not be nil")

	// Call IsInteractive to ensure it doesn't panic with valid inputs
	result := IsInteractive()
	assert.IsType(t, false, result)
}

func TestIsInteractive_MultipleCalls(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	defer func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
	}()

	// Setup with pipes
	rIn, wIn, err := os.Pipe()
	require.NoError(t, err)
	defer rIn.Close()
	defer wIn.Close()

	rOut, wOut, err := os.Pipe()
	require.NoError(t, err)
	defer rOut.Close()
	defer wOut.Close()

	os.Stdin = rIn
	os.Stdout = wOut

	// Multiple calls should return the same result
	results := make([]bool, 10)
	for i := 0; i < 10; i++ {
		results[i] = IsInteractive()
	}

	// All results should be false and consistent
	for i, result := range results {
		assert.False(t, result, "call %d should return false", i)
	}
}

func TestIsInteractive_ConcurrentCalls(t *testing.T) {
	// Test that concurrent calls don't cause issues
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			_ = IsInteractive()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// If we get here without panic, the test passes
	assert.True(t, true)
}

func TestIsInteractive_WithClosedPipe(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	defer func() {
		os.Stdin = originalStdin
	}()

	// Create and immediately close a pipe
	r, w, err := os.Pipe()
	require.NoError(t, err)
	w.Close() // Close write end

	os.Stdin = r
	defer r.Close()

	// Should still return false without panicking
	result := IsInteractive()
	assert.False(t, result)
}

func BenchmarkIsInteractive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsInteractive()
	}
}

func BenchmarkIsInteractive_WithPipes(b *testing.B) {
	// Save originals
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	defer func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
	}()

	// Setup pipes
	rIn, wIn, _ := os.Pipe()
	defer rIn.Close()
	defer wIn.Close()

	rOut, wOut, _ := os.Pipe()
	defer rOut.Close()
	defer wOut.Close()

	os.Stdin = rIn
	os.Stdout = wOut

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsInteractive()
	}
}

func ExampleIsInteractive() {
	interactive := IsInteractive()

	// Check interactivity before deciding how to proceed
	_ = interactive

	// Output:
}

func ExampleIsInteractive_conditionalPrompt() {
	var input string
	if IsInteractive() {
		// Prompt user
		input = "user-provided-value"
	} else {
		// Use default
		input = "default-value"
	}
	_ = input
}

// TestIsInteractive_DevNull tests behavior with /dev/null
func TestIsInteractive_DevNull(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	defer func() {
		os.Stdin = originalStdin
	}()

	// Open /dev/null
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Skip("Cannot open /dev/null:", err)
	}
	defer devNull.Close()

	os.Stdin = devNull

	// /dev/null is not a terminal
	result := IsInteractive()
	assert.False(t, result, "/dev/null should not be detected as a terminal")
}

// TestIsInteractive_FileDescriptorReuse tests with file descriptor reuse
func TestIsInteractive_FileDescriptorReuse(t *testing.T) {
	// Create multiple pipes to ensure file descriptors are reused properly
	for i := 0; i < 10; i++ {
		r, w, err := os.Pipe()
		require.NoError(t, err, "iteration %d", i)

		originalStdin := os.Stdin
		os.Stdin = r

		result := IsInteractive()
		assert.False(t, result, "iteration %d should return false", i)

		os.Stdin = originalStdin
		r.Close()
		w.Close()
	}
}

// TestIsInteractive_ValidFileDescriptors verifies that file descriptors are valid
func TestIsInteractive_ValidFileDescriptors(t *testing.T) {
	// Get file descriptors
	stdinFd := int(os.Stdin.Fd())
	stdoutFd := int(os.Stdout.Fd())

	// Verify they are valid (non-negative)
	assert.GreaterOrEqual(t, stdinFd, 0, "stdin fd should be valid")
	assert.GreaterOrEqual(t, stdoutFd, 0, "stdout fd should be valid")

	// Typically stdin=0, stdout=1, but this may vary
	t.Logf("stdin fd: %d, stdout fd: %d", stdinFd, stdoutFd)
}

// TestIsInteractive_ReadmeExample tests the example from package documentation
func TestIsInteractive_ReadmeExample(t *testing.T) {
	// This demonstrates the intended usage pattern
	interactive := IsInteractive()

	if interactive {
		t.Log("Terminal detected: would enable interactive prompts")
	} else {
		t.Log("Non-terminal detected: would use non-interactive defaults")
	}

	// The function should return a valid boolean
	assert.IsType(t, false, interactive)
}

// TestIsInteractive_StdinOnly tests that stdin check happens first
func TestIsInteractive_StdinOnly(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	defer func() {
		os.Stdin = originalStdin
	}()

	// Create a non-terminal stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	os.Stdin = r

	// Don't modify stdout
	// Since stdin is not a terminal, should return false immediately
	result := IsInteractive()
	assert.False(t, result, "should return false based on stdin alone")
}

// TestIsInteractive_WithFile tests behavior with a regular file
func TestIsInteractive_WithFile(t *testing.T) {
	// Save originals
	originalStdin := os.Stdin
	defer func() {
		os.Stdin = originalStdin
	}()

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "terminal-test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	// Write some data
	_, err = io.WriteString(tmpfile, "test data\n")
	require.NoError(t, err)

	// Seek back to beginning
	_, err = tmpfile.Seek(0, 0)
	require.NoError(t, err)

	// Replace stdin with file
	os.Stdin = tmpfile

	// Regular file is not a terminal
	result := IsInteractive()
	assert.False(t, result, "regular file should not be detected as terminal")
}
