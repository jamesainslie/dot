package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssertLink(t *testing.T) {
	tmpDir := t.TempDir()
	linkPath := filepath.Join(tmpDir, "link")
	target := "/some/target"

	require.NoError(t, os.Symlink(target, linkPath))

	AssertLink(t, linkPath, target)
}

func TestAssertLinkContains(t *testing.T) {
	tmpDir := t.TempDir()
	linkPath := filepath.Join(tmpDir, "link")
	target := "/some/target/file"

	require.NoError(t, os.Symlink(target, linkPath))

	AssertLinkContains(t, linkPath, "target")
}

func TestAssertFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	content := "test content"

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	AssertFile(t, filePath, content)
}

func TestAssertFileContains(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	content := "test content with substring"

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	AssertFileContains(t, filePath, "substring")
}

func TestAssertDir(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "subdir")

	require.NoError(t, os.MkdirAll(dirPath, 0755))

	AssertDir(t, dirPath)
}

func TestAssertNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	AssertNotExists(t, nonExistent)
}

func TestAssertFileMode(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "script.sh")

	require.NoError(t, os.WriteFile(filePath, []byte("#!/bin/bash"), 0755))

	AssertFileMode(t, filePath, 0755)
}

func TestAssertDirEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")

	require.NoError(t, os.MkdirAll(emptyDir, 0755))

	AssertDirEmpty(t, emptyDir)
}

func TestAssertDirHasEntries(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "dir")
	require.NoError(t, os.MkdirAll(dir, 0755))

	// Create some entries
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("b"), 0644))

	AssertDirHasEntries(t, dir, 2)
}
