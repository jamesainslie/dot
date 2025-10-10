package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestRepoURL returns a file:// URL to the local test repository fixture.
//
// The test repository contains Unix-focused dotfile packages (zsh, git, vim, ssh, tmux).
// TODO: Add Windows-specific packages (PowerShell, Windows Terminal, WSL) to test
// cross-platform scenarios with appropriate file paths and line endings.
func getTestRepoURL(t *testing.T) string {
	t.Helper()

	// Get absolute path to the test fixture repository
	testRepoPath, err := filepath.Abs("testdata/test-repo")
	require.NoError(t, err, "failed to get absolute path to test repository")

	// Verify the test repository directory exists
	require.DirExists(t, testRepoPath, "test repository fixture not found")

	// Initialize git repository if not already initialized
	gitDir := filepath.Join(testRepoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		initializeTestRepo(t, testRepoPath)
	}

	require.DirExists(t, gitDir, "test repository is not a git repository after initialization")

	return "file://" + testRepoPath
}

// initializeTestRepo initializes the test repository with git and commits the fixture files.
func initializeTestRepo(t *testing.T, repoPath string) {
	t.Helper()

	// Initialize git repository
	repo, err := git.PlainInit(repoPath, false)
	require.NoError(t, err, "failed to initialize test repository")

	// Get working tree
	worktree, err := repo.Worktree()
	require.NoError(t, err, "failed to get worktree")

	// Add all files
	err = worktree.AddGlob(".")
	require.NoError(t, err, "failed to add files")

	// Create initial commit on master branch
	_, err = worktree.Commit("Initial commit with dotfile packages", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err, "failed to create commit")

	// Rename master to main (modern git convention)
	// Get the current HEAD reference
	ref, err := repo.Head()
	require.NoError(t, err, "failed to get HEAD")

	// Create main branch reference pointing to same commit
	mainRef := plumbing.NewHashReference("refs/heads/main", ref.Hash())
	err = repo.Storer.SetReference(mainRef)
	require.NoError(t, err, "failed to create main branch")

	// Update HEAD to point to main
	symbolicRef := plumbing.NewSymbolicReference("HEAD", "refs/heads/main")
	err = repo.Storer.SetReference(symbolicRef)
	require.NoError(t, err, "failed to update HEAD to main")

	t.Logf("initialized test git repository at %s", repoPath)
}

func TestNewGoGitCloner(t *testing.T) {
	cloner := NewGoGitCloner()
	assert.NotNil(t, cloner)
}

func TestGoGitCloner_Clone_PublicRepo(t *testing.T) {
	ctx := context.Background()
	cloner := NewGoGitCloner()

	// Create temp directory for clone
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "repo")

	// Clone local test repository
	opts := CloneOptions{
		Auth:   NoAuth{},
		Depth:  1, // Shallow clone for speed
		Branch: "main",
	}

	url := getTestRepoURL(t)
	err := cloner.Clone(ctx, url, targetPath, opts)
	require.NoError(t, err)

	// Verify repository was cloned
	assert.DirExists(t, targetPath)
	assert.DirExists(t, filepath.Join(targetPath, ".git"))

	// Verify expected files and packages exist
	assert.FileExists(t, filepath.Join(targetPath, "README.md"))
	assert.FileExists(t, filepath.Join(targetPath, ".dotbootstrap.yaml"))
	assert.DirExists(t, filepath.Join(targetPath, "dot-zsh"))
	assert.FileExists(t, filepath.Join(targetPath, "dot-zsh", "zshrc"))
}

func TestGoGitCloner_Clone_InvalidURL(t *testing.T) {
	ctx := context.Background()
	cloner := NewGoGitCloner()
	tempDir := t.TempDir()

	opts := CloneOptions{
		Auth: NoAuth{},
	}

	err := cloner.Clone(ctx, "not-a-valid-url", tempDir, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "clone repository")
}

func TestGoGitCloner_Clone_NonExistentRepo(t *testing.T) {
	ctx := context.Background()
	cloner := NewGoGitCloner()
	tempDir := t.TempDir()

	opts := CloneOptions{
		Auth: NoAuth{},
	}

	// Use a non-existent local path to test error handling
	url := "file:///nonexistent/repo/path/that/does/not/exist"
	err := cloner.Clone(ctx, url, tempDir, opts)
	assert.Error(t, err)
}

func TestGoGitCloner_Clone_ExistingDirectory(t *testing.T) {
	ctx := context.Background()
	cloner := NewGoGitCloner()

	// Create and use existing directory
	tempDir := t.TempDir()
	existingPath := filepath.Join(tempDir, "existing")
	err := os.MkdirAll(existingPath, 0755)
	require.NoError(t, err)

	// Create a file in the directory
	testFile := filepath.Join(existingPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	opts := CloneOptions{
		Auth: NoAuth{},
	}

	url := getTestRepoURL(t)
	err = cloner.Clone(ctx, url, existingPath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGoGitCloner_Clone_WithBranch(t *testing.T) {
	ctx := context.Background()
	cloner := NewGoGitCloner()
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "repo")

	opts := CloneOptions{
		Auth:   NoAuth{},
		Depth:  1,
		Branch: "main",
	}

	url := getTestRepoURL(t)
	err := cloner.Clone(ctx, url, targetPath, opts)
	require.NoError(t, err)

	assert.DirExists(t, targetPath)
	assert.FileExists(t, filepath.Join(targetPath, "README.md"))
}

func TestGoGitCloner_Clone_ContextCancellation(t *testing.T) {
	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cloner := NewGoGitCloner()
	tempDir := t.TempDir()

	opts := CloneOptions{
		Auth: NoAuth{},
	}

	url := getTestRepoURL(t)
	err := cloner.Clone(ctx, url, tempDir, opts)
	assert.Error(t, err)
}
