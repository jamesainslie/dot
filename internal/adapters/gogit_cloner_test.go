package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGoGitCloner(t *testing.T) {
	cloner := NewGoGitCloner()
	assert.NotNil(t, cloner)
}

func TestGoGitCloner_Clone_PublicRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	cloner := NewGoGitCloner()

	// Create temp directory for clone
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "repo")

	// Clone a small public repository
	opts := CloneOptions{
		Auth:   NoAuth{},
		Depth:  1, // Shallow clone for speed
		Branch: "main",
	}

	// Use a test repository (dot's own repo or a small test fixture)
	url := "https://github.com/jamesainslie/dot"
	err := cloner.Clone(ctx, url, targetPath, opts)
	require.NoError(t, err)

	// Verify repository was cloned
	assert.DirExists(t, targetPath)
	assert.DirExists(t, filepath.Join(targetPath, ".git"))
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
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	cloner := NewGoGitCloner()
	tempDir := t.TempDir()

	opts := CloneOptions{
		Auth: NoAuth{},
	}

	url := "https://github.com/nonexistent/repo-that-does-not-exist-12345"
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

	url := "https://github.com/jamesainslie/dot"
	err = cloner.Clone(ctx, url, existingPath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGoGitCloner_Clone_WithBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	cloner := NewGoGitCloner()
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "repo")

	opts := CloneOptions{
		Auth:   NoAuth{},
		Depth:  1,
		Branch: "main",
	}

	url := "https://github.com/jamesainslie/dot"
	err := cloner.Clone(ctx, url, targetPath, opts)
	require.NoError(t, err)

	assert.DirExists(t, targetPath)
}

func TestGoGitCloner_Clone_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cloner := NewGoGitCloner()
	tempDir := t.TempDir()

	opts := CloneOptions{
		Auth: NoAuth{},
	}

	url := "https://github.com/jamesainslie/dot"
	err := cloner.Clone(ctx, url, tempDir, opts)
	assert.Error(t, err)
}
