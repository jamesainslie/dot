// Package terminal provides terminal detection utilities.
package terminal

import (
	"os"

	"golang.org/x/term"
)

// IsInteractive determines if the current process is running in an interactive terminal.
//
// Returns true if both stdin and stdout are connected to a terminal (TTY).
// Returns false if either is redirected to a file or pipe.
//
// This is useful for deciding whether to prompt the user for input or
// fall back to non-interactive behavior.
func IsInteractive() bool {
	// Check if stdin is a terminal
	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		return false
	}

	// Check if stdout is a terminal
	stdoutFd := int(os.Stdout.Fd())
	if !term.IsTerminal(stdoutFd) {
		return false
	}

	return true
}
