package gitsync

import (
	"fmt"
	"strings"
)

// ErrGitNotFound reports that the git executable is not installed or not on PATH.
type ErrGitNotFound struct {
	Err error
}

func (e ErrGitNotFound) Error() string {
	return "git executable not found in PATH: dot sync requires a working git installation"
}

func (e ErrGitNotFound) Unwrap() error {
	return e.Err
}

// ErrNotRepository reports that a directory is not inside a git working tree.
type ErrNotRepository struct {
	Dir string
}

func (e ErrNotRepository) Error() string {
	return fmt.Sprintf("package directory is not a git repository: %s", e.Dir)
}

// ErrRebaseConflict reports that `git pull --rebase` stopped with conflicts.
// Paths lists the conflicted files relative to the repository root.
type ErrRebaseConflict struct {
	Dir    string
	Paths  []string
	Output string
	// Err is the failing git invocation, kept so callers can inspect it.
	Err error
}

func (e ErrRebaseConflict) Error() string {
	return fmt.Sprintf("rebase stopped with conflicts in %s: %s", e.Dir, strings.Join(e.Paths, ", "))
}

func (e ErrRebaseConflict) Unwrap() error {
	return e.Err
}

// ErrCommand reports a git invocation that exited non-zero or failed to start.
type ErrCommand struct {
	Args     []string
	ExitCode int
	Stderr   string
	Err      error
}

func (e ErrCommand) Error() string {
	msg := fmt.Sprintf("git %s failed with exit code %d", strings.Join(e.Args, " "), e.ExitCode)
	if e.Stderr != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.Stderr)
	}
	return msg
}

func (e ErrCommand) Unwrap() error {
	return e.Err
}
