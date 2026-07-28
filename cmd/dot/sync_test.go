package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// syncSandbox holds the paths of an isolated sync test environment.
type syncSandbox struct {
	origin     string
	packageDir string
	targetDir  string
}

// requireGitBinary skips the test when git is not installed.
func requireGitBinary(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// gitInDir runs a git command and fails the test when it errors.
func gitInDir(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %v failed: %s", args, out)
	return string(out)
}

// configureSyncRepo applies the local git settings needed for hermetic commits.
func configureSyncRepo(t *testing.T, dir string) {
	t.Helper()
	gitInDir(t, dir, "config", "user.name", "dot test")
	gitInDir(t, dir, "config", "user.email", "dot@example.test")
	gitInDir(t, dir, "config", "commit.gpgsign", "false")
	gitInDir(t, dir, "config", "core.hooksPath", t.TempDir())
}

// newSyncSandbox builds a bare origin repository seeded with a dot-vim package,
// clones it as the package directory, and points every ambient path at
// temporary directories so the real HOME is never touched.
func newSyncSandbox(t *testing.T) syncSandbox {
	t.Helper()
	requireGitBinary(t)

	root := t.TempDir()
	home := filepath.Join(root, "home")
	require.NoError(t, os.MkdirAll(home, 0o755))

	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(home, ".gitconfig-system"))
	t.Setenv("GIT_AUTHOR_NAME", "dot test")
	t.Setenv("GIT_AUTHOR_EMAIL", "dot@example.test")
	t.Setenv("GIT_COMMITTER_NAME", "dot test")
	t.Setenv("GIT_COMMITTER_EMAIL", "dot@example.test")

	origin := filepath.Join(root, "origin.git")
	seed := filepath.Join(root, "seed")
	packageDir := filepath.Join(root, "dotfiles")
	targetDir := filepath.Join(root, "target")

	require.NoError(t, os.MkdirAll(origin, 0o755))
	gitInDir(t, origin, "init", "--bare", "-b", "main")

	require.NoError(t, os.MkdirAll(filepath.Join(seed, "dot-vim"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(seed, "dot-vim", "dot-vimrc"), []byte("set number\n"), 0o644))
	gitInDir(t, seed, "init", "-b", "main")
	configureSyncRepo(t, seed)
	gitInDir(t, seed, "add", ".")
	gitInDir(t, seed, "commit", "-m", "seed: initial packages")
	gitInDir(t, seed, "remote", "add", "origin", origin)
	gitInDir(t, seed, "push", "-u", "origin", "main")

	gitInDir(t, root, "clone", origin, packageDir)
	configureSyncRepo(t, packageDir)
	require.NoError(t, os.MkdirAll(targetDir, 0o755))

	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
	})

	return syncSandbox{origin: origin, packageDir: packageDir, targetDir: targetDir}
}

// cloneOrigin makes an independent clone of the sandbox origin.
func (s syncSandbox) cloneOrigin(t *testing.T, name string) string {
	t.Helper()
	parent := t.TempDir()
	gitInDir(t, parent, "clone", s.origin, name)
	dir := filepath.Join(parent, name)
	configureSyncRepo(t, dir)
	return dir
}

// publish commits a change from an independent clone and pushes it to origin.
func (s syncSandbox) publish(t *testing.T, relPath, content, message string) {
	t.Helper()
	other := s.cloneOrigin(t, "publisher")
	full := filepath.Join(other, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	gitInDir(t, other, "add", "--all")
	gitInDir(t, other, "commit", "-m", message)
	gitInDir(t, other, "push")
}

// remove deletes a path from an independent clone and pushes the deletion.
func (s syncSandbox) remove(t *testing.T, relPath, message string) {
	t.Helper()
	other := s.cloneOrigin(t, "remover")
	require.NoError(t, os.RemoveAll(filepath.Join(other, relPath)))
	gitInDir(t, other, "add", "--all")
	gitInDir(t, other, "commit", "-m", message)
	gitInDir(t, other, "push")
}

// manage installs packages through the manage command in the sandbox.
func manageInSandbox(t *testing.T, packages ...string) {
	t.Helper()
	cmd := newManageCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(context.Background())
	cmd.SetArgs(packages)
	require.NoError(t, cmd.Execute())
}

// runSyncCommand executes the sync command and returns stdout, stderr, and error.
func runSyncCommand(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := newSyncCommand()
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// listLinks returns the sorted symlink paths under dir.
func listLinks(t *testing.T, dir string) []string {
	t.Helper()
	links := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			links = append(links, path)
		}
		return nil
	})
	require.NoError(t, err)
	return links
}

func TestSyncCommand_ContentOnlyPull(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	linksBefore := listLinks(t, sandbox.targetDir)
	require.NotEmpty(t, linksBefore)

	sandbox.publish(t, "dot-vim/dot-vimrc", "set number\nset ruler\n", "feat(vim): enable ruler")

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Pulled 1 commit")
	assert.Contains(t, stdout, "feat(vim): enable ruler")
	// The relink pass still runs and reports itself, as the docs describe.
	assert.Contains(t, stdout, "Remanaged 1 package")

	// A content-only change does not alter the link set.
	assert.Equal(t, linksBefore, listLinks(t, sandbox.targetDir))

	content, err := os.ReadFile(filepath.Join(sandbox.targetDir, ".vim", ".vimrc"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "set ruler")
}

func TestSyncCommand_StructuralPullCreatesLink(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	newLink := filepath.Join(sandbox.targetDir, ".vim", "colors", "solarized.vim")
	require.NoFileExists(t, newLink)

	sandbox.publish(t, "dot-vim/colors/solarized.vim", "\" solarized\n", "feat(vim): add solarized colors")

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Pulled 1 commit")
	assert.FileExists(t, newLink)

	info, err := os.Lstat(newLink)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&os.ModeSymlink, "new package file is linked into the target")
}

func TestSyncCommand_UpToDateIsQuiet(t *testing.T) {
	newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Already up to date")
	assert.NotContains(t, stdout, "Uncommitted changes")
	assert.Contains(t, stdout, "1/1")
}

func TestSyncCommand_PackageRemovedUpstreamKeepsGoing(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	// A local edit keeps the tree dirty so the warning block is observable.
	require.NoError(t, os.WriteFile(filepath.Join(sandbox.packageDir, "README.md"), []byte("notes\n"), 0o644))

	sandbox.remove(t, "dot-vim", "chore: retire the vim package")

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Pulled 1 commit")
	assert.Contains(t, stdout, "dot-vim", "the package that vanished upstream is named")
	assert.Contains(t, stdout, "dot unmanage", "the warning explains how to finish the removal")
	assert.Contains(t, stdout, "Packages:", "the health summary still runs")
	assert.Contains(t, stdout, "Uncommitted changes", "the dirty-tree warning still runs")
}

func TestSyncCommand_DryRunDoesNotClaimRemanage(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	// Removing a link gives remanage real work to do on a live run.
	link := filepath.Join(sandbox.targetDir, ".vim", ".vimrc")
	require.NoError(t, os.Remove(link))

	cliFlags.dryRun = true
	t.Cleanup(func() { cliFlags.dryRun = false })

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.NotContains(t, stdout, "Remanaged", "a dry run never claims work it did not do")
	assert.NoFileExists(t, link, "a dry run does not restore the link")
}

func TestSyncCommand_DirtyTreeWarning(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	// Simulate an edit made through a symlink: the package clone goes dirty.
	require.NoError(t, os.WriteFile(filepath.Join(sandbox.packageDir, "dot-vim", "dot-vimrc"), []byte("set number\nset ruler\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(sandbox.packageDir, "dot-zsh"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sandbox.packageDir, "dot-zsh", "dot-zshrc"), []byte("alias l=ls\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sandbox.packageDir, "README.md"), []byte("notes\n"), 0o644))

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Uncommitted changes")
	assert.Contains(t, stdout, sandbox.packageDir)
	assert.Contains(t, stdout, "dot-vim")
	assert.Contains(t, stdout, "dot-vim/dot-vimrc")
	assert.Contains(t, stdout, "dot-zsh")
	assert.Contains(t, stdout, "README.md")
	assert.Contains(t, stdout, "dot sync --push")
}

func TestSyncCommand_CleanTreePrintsNoWarning(t *testing.T) {
	newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.NotContains(t, stdout, "Uncommitted changes")
	assert.NotContains(t, stdout, "dot sync --push")
}

func TestSyncCommand_PushRoundTrip(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	require.NoError(t, os.WriteFile(filepath.Join(sandbox.packageDir, "dot-vim", "dot-vimrc"), []byte("set number\nset ruler\n"), 0o644))

	stdout, _, err := runSyncCommand(t, "--push")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Committed")
	assert.Contains(t, stdout, "Pushed")

	verify := sandbox.cloneOrigin(t, "verify")
	content, err := os.ReadFile(filepath.Join(verify, "dot-vim", "dot-vimrc"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "set ruler")

	subject := gitInDir(t, verify, "log", "-1", "--pretty=%s")
	assert.Contains(t, subject, "local edits (dot-vim)")
	assert.True(t, strings.HasPrefix(strings.TrimSpace(subject), "sync: "), "subject was %q", subject)
}

func TestSyncCommand_PushWithCleanTree(t *testing.T) {
	newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	stdout, _, err := runSyncCommand(t, "--push")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Nothing to commit")
	assert.NotContains(t, stdout, "Pushed")
}

func TestSyncCommand_RebaseConflictStops(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")

	sandbox.publish(t, "dot-vim/dot-vimrc", "remote change\n", "feat(vim): remote change")

	require.NoError(t, os.WriteFile(filepath.Join(sandbox.packageDir, "dot-vim", "dot-vimrc"), []byte("local change\n"), 0o644))
	gitInDir(t, sandbox.packageDir, "commit", "-am", "feat(vim): local change")

	_, stderr, err := runSyncCommand(t)
	require.Error(t, err)

	assert.Contains(t, stderr, "dot-vim/dot-vimrc")
	assert.Contains(t, stderr, sandbox.packageDir)
	assert.Contains(t, stderr, "dot sync")
}

func TestSyncCommand_NotAGitRepository(t *testing.T) {
	requireGitBinary(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))

	plain := t.TempDir()
	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: plain,
		targetDir:  t.TempDir(),
	})

	_, _, err := runSyncCommand(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
	assert.Contains(t, err.Error(), plain)
}

func TestSyncCommand_RejectsPositionalArguments(t *testing.T) {
	newSyncSandbox(t)

	_, _, err := runSyncCommand(t, "dot-vim")
	require.Error(t, err)
}

func TestSyncCommand_DryRunSkipsGitWrites(t *testing.T) {
	sandbox := newSyncSandbox(t)
	manageInSandbox(t, "dot-vim")
	sandbox.publish(t, "dot-vim/dot-vimrc", "set number\nset ruler\n", "feat(vim): enable ruler")

	cliFlags.dryRun = true
	t.Cleanup(func() { cliFlags.dryRun = false })

	stdout, _, err := runSyncCommand(t)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Dry run")

	content, err := os.ReadFile(filepath.Join(sandbox.packageDir, "dot-vim", "dot-vimrc"))
	require.NoError(t, err)
	assert.NotContains(t, string(content), "set ruler", "dry run does not pull")
}

func TestNewSyncCommand_Metadata(t *testing.T) {
	cmd := newSyncCommand()

	assert.Equal(t, "sync", cmd.Name())
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Flags().Lookup("push"))
}
