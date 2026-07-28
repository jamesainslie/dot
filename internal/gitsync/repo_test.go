package gitsync

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireGit skips the test when git is not installed on the machine.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// runGit runs a git command in dir and fails the test when it errors.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv(t)
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %v failed: %s", args, out)
	return string(out)
}

// gitEnv returns a hermetic environment for git invocations in tests.
func gitEnv(t *testing.T) []string {
	t.Helper()
	home := t.TempDir()
	return append(os.Environ(),
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, "config"),
		"GIT_CONFIG_GLOBAL="+filepath.Join(home, "gitconfig"),
		"GIT_CONFIG_SYSTEM="+filepath.Join(home, "gitconfig-system"),
		"GIT_AUTHOR_NAME=dot test",
		"GIT_AUTHOR_EMAIL=dot@example.test",
		"GIT_COMMITTER_NAME=dot test",
		"GIT_COMMITTER_EMAIL=dot@example.test",
	)
}

// configureRepo applies the local settings needed for hermetic commits.
func configureRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "config", "user.name", "dot test")
	runGit(t, dir, "config", "user.email", "dot@example.test")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	runGit(t, dir, "config", "core.hooksPath", t.TempDir())
}

// newOriginWithClone creates a bare origin repo seeded with one package and
// returns the origin path plus a working clone path.
func newOriginWithClone(t *testing.T) (origin, clone string) {
	t.Helper()
	requireGit(t)

	root := t.TempDir()
	origin = filepath.Join(root, "origin.git")
	seed := filepath.Join(root, "seed")
	clone = filepath.Join(root, "clone")

	require.NoError(t, os.MkdirAll(origin, 0o755))
	runGit(t, origin, "init", "--bare", "-b", "main")

	require.NoError(t, os.MkdirAll(filepath.Join(seed, "dot-vim"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(seed, "dot-vim", "dot-vimrc"), []byte("set number\n"), 0o644))
	runGit(t, seed, "init", "-b", "main")
	configureRepo(t, seed)
	runGit(t, seed, "add", ".")
	runGit(t, seed, "commit", "-m", "seed: initial packages")
	runGit(t, seed, "remote", "add", "origin", origin)
	runGit(t, seed, "push", "-u", "origin", "main")

	runGit(t, root, "clone", origin, clone)
	configureRepo(t, clone)

	return origin, clone
}

// openTestRepo opens a Repo whose git invocations use the hermetic test env.
func openTestRepo(t *testing.T, dir string) *Repo {
	t.Helper()
	runner, err := NewExecRunner()
	require.NoError(t, err)
	runner.Env = gitEnv(t)
	return NewRepo(dir, runner)
}

func TestExecRunner_RunCapturesOutput(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)

	runner, err := NewExecRunner()
	require.NoError(t, err)
	runner.Env = gitEnv(t)

	out, err := runner.Run(context.Background(), clone, "rev-parse", "--abbrev-ref", "HEAD")
	require.NoError(t, err)
	assert.Equal(t, "main", out.Stdout)
}

func TestExecRunner_RunReturnsCommandError(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)

	runner, err := NewExecRunner()
	require.NoError(t, err)
	runner.Env = gitEnv(t)

	_, err = runner.Run(context.Background(), clone, "rev-parse", "--verify", "does-not-exist")
	require.Error(t, err)

	var cmdErr ErrCommand
	require.ErrorAs(t, err, &cmdErr)
	assert.NotZero(t, cmdErr.ExitCode)
	assert.Contains(t, cmdErr.Error(), "git rev-parse")
}

func TestRepo_EnsureRepository(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)

	t.Run("git repository passes", func(t *testing.T) {
		repo := openTestRepo(t, clone)
		require.NoError(t, repo.EnsureRepository(context.Background()))
	})

	t.Run("plain directory reports not a repository", func(t *testing.T) {
		plain := t.TempDir()
		repo := openTestRepo(t, plain)

		err := repo.EnsureRepository(context.Background())
		require.Error(t, err)

		var notRepo ErrNotRepository
		require.ErrorAs(t, err, &notRepo)
		assert.Equal(t, plain, notRepo.Dir)
		assert.Contains(t, err.Error(), plain)
	})

	t.Run("missing directory reports not a repository", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "nope")
		repo := openTestRepo(t, missing)

		var notRepo ErrNotRepository
		require.ErrorAs(t, repo.EnsureRepository(context.Background()), &notRepo)
	})
}

// stubRunner returns canned results, so tests can exercise git failures that
// are awkward to provoke with a real repository.
type stubRunner struct {
	out Output
	err error
}

func (r stubRunner) Run(_ context.Context, _ string, _ ...string) (Output, error) {
	return r.out, r.err
}

func TestRepo_EnsureRepositoryPropagatesUnexpectedFailure(t *testing.T) {
	inner := ErrCommand{
		Args:     []string{"rev-parse", "--is-inside-work-tree"},
		ExitCode: 128,
		Stderr:   "fatal: detected dubious ownership in repository",
		Err:      errors.New("exit status 128"),
	}
	// The directory exists, so the failure cannot be a missing package dir.
	repo := NewRepo(t.TempDir(), stubRunner{err: inner})

	err := repo.EnsureRepository(context.Background())
	require.Error(t, err)

	var notRepo ErrNotRepository
	assert.NotErrorAs(t, err, &notRepo, "an unrelated git failure is not reported as a missing repository")

	var cmdErr ErrCommand
	require.ErrorAs(t, err, &cmdErr)
	assert.Contains(t, cmdErr.Stderr, "dubious ownership")
}

func TestRepo_HeadOnUnbornBranch(t *testing.T) {
	requireGit(t)

	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	configureRepo(t, dir)

	repo := openTestRepo(t, dir)
	head, err := repo.Head(context.Background())
	require.NoError(t, err, "an unborn HEAD is a normal state, not an error")
	assert.Empty(t, head)
}

func TestRepo_PullIntoUnbornBranchReportsEveryCommit(t *testing.T) {
	requireGit(t)
	origin, _ := newOriginWithClone(t)

	// A repository with no commits of its own, tracking the seeded origin.
	fresh := filepath.Join(t.TempDir(), "fresh")
	require.NoError(t, os.MkdirAll(fresh, 0o755))
	runGit(t, fresh, "init", "-b", "main")
	configureRepo(t, fresh)
	runGit(t, fresh, "remote", "add", "origin", origin)
	runGit(t, fresh, "config", "branch.main.remote", "origin")
	runGit(t, fresh, "config", "branch.main.merge", "refs/heads/main")

	repo := openTestRepo(t, fresh)
	result, err := repo.Pull(context.Background())
	require.NoError(t, err)

	assert.Empty(t, result.OldHead)
	assert.NotEmpty(t, result.NewHead)
	require.Len(t, result.Commits, 1, "the imported history counts as pulled commits")
	assert.Equal(t, "seed: initial packages", result.Commits[0].Subject)
	assert.False(t, result.UpToDate())
}

func TestRepo_PullReportsNewCommits(t *testing.T) {
	requireGit(t)
	origin, clone := newOriginWithClone(t)

	// Second clone publishes a new commit.
	other := filepath.Join(t.TempDir(), "other")
	runGit(t, filepath.Dir(other), "clone", origin, other)
	configureRepo(t, other)
	require.NoError(t, os.WriteFile(filepath.Join(other, "dot-vim", "dot-vimrc"), []byte("set number\nset ruler\n"), 0o644))
	runGit(t, other, "commit", "-am", "feat(vim): enable ruler")
	runGit(t, other, "push")

	repo := openTestRepo(t, clone)
	result, err := repo.Pull(context.Background())
	require.NoError(t, err)

	require.Len(t, result.Commits, 1)
	assert.Equal(t, "feat(vim): enable ruler", result.Commits[0].Subject)
	assert.NotEmpty(t, result.Commits[0].Hash)
	assert.NotEqual(t, result.OldHead, result.NewHead)

	content, err := os.ReadFile(filepath.Join(clone, "dot-vim", "dot-vimrc"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "set ruler")
}

func TestRepo_PullUpToDateReportsNoCommits(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)

	repo := openTestRepo(t, clone)
	result, err := repo.Pull(context.Background())
	require.NoError(t, err)

	assert.Empty(t, result.Commits)
	assert.Equal(t, result.OldHead, result.NewHead)
}

func TestRepo_PullConflictReportsPaths(t *testing.T) {
	requireGit(t)
	origin, clone := newOriginWithClone(t)

	other := filepath.Join(t.TempDir(), "other")
	runGit(t, filepath.Dir(other), "clone", origin, other)
	configureRepo(t, other)
	require.NoError(t, os.WriteFile(filepath.Join(other, "dot-vim", "dot-vimrc"), []byte("remote change\n"), 0o644))
	runGit(t, other, "commit", "-am", "feat(vim): remote change")
	runGit(t, other, "push")

	// Local commit touching the same line causes a rebase conflict.
	require.NoError(t, os.WriteFile(filepath.Join(clone, "dot-vim", "dot-vimrc"), []byte("local change\n"), 0o644))
	runGit(t, clone, "commit", "-am", "feat(vim): local change")

	repo := openTestRepo(t, clone)
	_, err := repo.Pull(context.Background())
	require.Error(t, err)

	var conflict ErrRebaseConflict
	require.ErrorAs(t, err, &conflict)
	assert.Contains(t, conflict.Paths, "dot-vim/dot-vimrc")
	assert.Contains(t, conflict.Error(), "dot-vim/dot-vimrc")

	// The failing git invocation stays reachable for exit-code mapping.
	var cmdErr ErrCommand
	require.ErrorAs(t, err, &cmdErr)
	assert.Equal(t, []string{"pull", "--rebase", "--autostash"}, cmdErr.Args)
}

func TestRepo_StatusReportsWorkingTree(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)
	repo := openTestRepo(t, clone)
	ctx := context.Background()

	entries, err := repo.Status(ctx)
	require.NoError(t, err)
	assert.Empty(t, entries, "freshly cloned tree is clean")

	require.NoError(t, os.WriteFile(filepath.Join(clone, "dot-vim", "dot-vimrc"), []byte("edited\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(clone, "dot-zsh"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(clone, "dot-zsh", "dot-zshrc"), []byte("new\n"), 0o644))

	entries, err = repo.Status(ctx)
	require.NoError(t, err)

	groups := GroupByPackage(entries)
	require.Len(t, groups, 2)
	assert.Equal(t, "dot-vim", groups[0].Package)
	assert.Equal(t, "dot-zsh", groups[1].Package)
}

func TestRepo_CommitAllAndPush(t *testing.T) {
	requireGit(t)
	origin, clone := newOriginWithClone(t)
	repo := openTestRepo(t, clone)
	ctx := context.Background()

	require.NoError(t, os.WriteFile(filepath.Join(clone, "dot-vim", "dot-vimrc"), []byte("edited\n"), 0o644))

	require.NoError(t, repo.CommitAll(ctx, "sync: testhost local edits (dot-vim)"))
	require.NoError(t, repo.Push(ctx))

	entries, err := repo.Status(ctx)
	require.NoError(t, err)
	assert.Empty(t, entries)

	// The commit is visible from an independent clone.
	verify := filepath.Join(t.TempDir(), "verify")
	runGit(t, filepath.Dir(verify), "clone", origin, verify)
	content, err := os.ReadFile(filepath.Join(verify, "dot-vim", "dot-vimrc"))
	require.NoError(t, err)
	assert.Equal(t, "edited\n", string(content))

	log := runGit(t, verify, "log", "-1", "--pretty=%s")
	assert.Contains(t, log, "sync: testhost local edits (dot-vim)")
}

func TestRepo_CommitAllWithNothingStagedFails(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)
	repo := openTestRepo(t, clone)

	err := repo.CommitAll(context.Background(), "sync: testhost local edits")
	require.Error(t, err)
}

func TestOpenAndDir(t *testing.T) {
	requireGit(t)
	_, clone := newOriginWithClone(t)

	repo, err := Open(clone)
	require.NoError(t, err)
	assert.Equal(t, clone, repo.Dir())
}

func TestOpenWithoutGitOnPath(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := Open(t.TempDir())
	require.Error(t, err)

	var notFound ErrGitNotFound
	require.ErrorAs(t, err, &notFound)
	assert.Contains(t, err.Error(), "git")
}

func TestPullResult_UpToDate(t *testing.T) {
	tests := []struct {
		name   string
		result PullResult
		want   bool
	}{
		{"no commits", PullResult{OldHead: "abc", NewHead: "abc"}, true},
		{"one commit", PullResult{Commits: []Commit{{Hash: "abc", Subject: "feat: x"}}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.UpToDate())
		})
	}
}

func TestOutput_Combined(t *testing.T) {
	tests := []struct {
		name string
		out  Output
		want string
	}{
		{"both streams", Output{Stdout: "out", Stderr: "err"}, "out\nerr"},
		{"stdout only", Output{Stdout: "out"}, "out"},
		{"stderr only", Output{Stderr: "err"}, "err"},
		{"neither", Output{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.out.Combined())
		})
	}
}

func TestErrCommand_Unwrap(t *testing.T) {
	inner := errors.New("exit status 1")
	err := ErrCommand{Args: []string{"push"}, ExitCode: 1, Err: inner}

	assert.ErrorIs(t, err, inner)
}

func TestErrGitNotFound(t *testing.T) {
	err := ErrGitNotFound{Err: errors.New("boom")}
	assert.Contains(t, err.Error(), "git")
	assert.Equal(t, "boom", errors.Unwrap(err).Error())
}

func TestErrCommand_Error(t *testing.T) {
	err := ErrCommand{Args: []string{"status"}, ExitCode: 128, Stderr: "fatal: bad"}
	assert.Contains(t, err.Error(), "git status")
	assert.Contains(t, err.Error(), "fatal: bad")
	assert.Contains(t, err.Error(), "128")
}
