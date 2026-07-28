// Package gitsync wraps the system git executable with the small set of
// operations dot needs to reconcile a package directory with its remote:
// rebasing pulls, working-tree inspection, and commit and push of local edits.
package gitsync

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
)

// Output holds the captured result of a git invocation.
type Output struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Combined returns stdout and stderr joined, in that order, skipping empties.
func (o Output) Combined() string {
	parts := make([]string, 0, 2)
	if o.Stdout != "" {
		parts = append(parts, o.Stdout)
	}
	if o.Stderr != "" {
		parts = append(parts, o.Stderr)
	}
	return strings.Join(parts, "\n")
}

// Runner executes a git subcommand in a working directory.
type Runner interface {
	Run(ctx context.Context, dir string, args ...string) (Output, error)
}

// ExecRunner runs git through os/exec.
type ExecRunner struct {
	// GitPath is the resolved path to the git executable.
	GitPath string
	// Env overrides the environment passed to git. A nil Env inherits the
	// process environment.
	Env []string
}

// NewExecRunner locates git on PATH and returns a runner for it.
// It returns ErrGitNotFound when git is not installed.
func NewExecRunner() (*ExecRunner, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return nil, ErrGitNotFound{Err: err}
	}
	return &ExecRunner{GitPath: path}, nil
}

// Run executes git with the given arguments in dir and captures its output.
// Trailing newlines are trimmed from both streams.
func (r *ExecRunner) Run(ctx context.Context, dir string, args ...string) (Output, error) {
	cmd := exec.CommandContext(ctx, r.GitPath, args...) // #nosec G204 -- arguments are constructed by dot, not user input
	cmd.Dir = dir
	if r.Env != nil {
		cmd.Env = r.Env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := Output{
		Stdout:   strings.TrimRight(stdout.String(), "\n"),
		Stderr:   strings.TrimRight(stderr.String(), "\n"),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			out.ExitCode = -1
		}
		return out, ErrCommand{
			Args:     args,
			ExitCode: out.ExitCode,
			Stderr:   out.Stderr,
			Err:      err,
		}
	}

	return out, nil
}

// Commit is a single commit summary produced by a pull.
type Commit struct {
	Hash    string
	Subject string
}

// PullResult describes the effect of a rebasing pull.
type PullResult struct {
	OldHead string
	NewHead string
	// Commits are the commits reachable from the new head but not the old
	// one, newest first.
	Commits []Commit
}

// UpToDate reports whether the pull left the branch unchanged.
func (r PullResult) UpToDate() bool {
	return len(r.Commits) == 0
}

// Repo is a git working tree that dot manages.
type Repo struct {
	dir    string
	runner Runner
}

// NewRepo returns a Repo for dir that uses the given runner.
func NewRepo(dir string, runner Runner) *Repo {
	return &Repo{dir: dir, runner: runner}
}

// Open returns a Repo for dir backed by the system git executable.
func Open(dir string) (*Repo, error) {
	runner, err := NewExecRunner()
	if err != nil {
		return nil, err
	}
	return NewRepo(dir, runner), nil
}

// Dir returns the working directory the repository operates on.
func (r *Repo) Dir() string {
	return r.dir
}

// EnsureRepository verifies that the directory is inside a git working tree.
func (r *Repo) EnsureRepository(ctx context.Context) error {
	out, err := r.runner.Run(ctx, r.dir, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(out.Stdout) != "true" {
		return ErrNotRepository{Dir: r.dir}
	}
	return nil
}

// Head returns the current commit hash.
func (r *Repo) Head(ctx context.Context) (string, error) {
	out, err := r.runner.Run(ctx, r.dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Stdout), nil
}

// Pull runs `git pull --rebase --autostash` and reports the commits gained.
// It returns ErrRebaseConflict when the rebase stops with conflicting paths;
// callers are expected to stop and let the user resolve them by hand.
func (r *Repo) Pull(ctx context.Context) (PullResult, error) {
	oldHead, _ := r.Head(ctx)

	out, err := r.runner.Run(ctx, r.dir, "pull", "--rebase", "--autostash")
	if err != nil {
		if paths := r.conflictedPaths(ctx); len(paths) > 0 {
			return PullResult{}, ErrRebaseConflict{
				Dir:    r.dir,
				Paths:  paths,
				Output: out.Combined(),
			}
		}
		return PullResult{}, err
	}

	newHead, headErr := r.Head(ctx)
	if headErr != nil {
		return PullResult{}, headErr
	}

	commits, err := r.commitsBetween(ctx, oldHead, newHead)
	if err != nil {
		return PullResult{}, err
	}

	return PullResult{OldHead: oldHead, NewHead: newHead, Commits: commits}, nil
}

// Status returns the working-tree entries reported by git status --porcelain.
func (r *Repo) Status(ctx context.Context) ([]StatusEntry, error) {
	out, err := r.runner.Run(ctx, r.dir, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	return ParsePorcelain(out.Stdout), nil
}

// CommitAll stages every change in the working tree and commits it.
// Commit hooks run normally; dot never bypasses them.
func (r *Repo) CommitAll(ctx context.Context, message string) error {
	if _, err := r.runner.Run(ctx, r.dir, "add", "--all"); err != nil {
		return err
	}
	_, err := r.runner.Run(ctx, r.dir, "commit", "-m", message)
	return err
}

// Push publishes the current branch to its configured remote.
func (r *Repo) Push(ctx context.Context) error {
	_, err := r.runner.Run(ctx, r.dir, "push")
	return err
}

// conflictedPaths returns the unmerged paths in the working tree, if any.
func (r *Repo) conflictedPaths(ctx context.Context) []string {
	out, err := r.runner.Run(ctx, r.dir, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil
	}

	paths := make([]string, 0)
	for _, line := range strings.Split(out.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

// commitsBetween lists the commits reachable from newHead but not oldHead.
// After a rebase this includes locally rewritten commits, which are new to
// this branch as well.
func (r *Repo) commitsBetween(ctx context.Context, oldHead, newHead string) ([]Commit, error) {
	if oldHead == "" || oldHead == newHead {
		return nil, nil
	}

	out, err := r.runner.Run(ctx, r.dir, "log", "--pretty=format:%h\t%s", oldHead+".."+newHead)
	if err != nil {
		return nil, err
	}

	commits := make([]Commit, 0)
	for _, line := range strings.Split(out.Stdout, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		hash, subject, _ := strings.Cut(line, "\t")
		commits = append(commits, Commit{Hash: hash, Subject: subject})
	}
	return commits, nil
}
