package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertLink verifies a symlink exists and points to the expected target.
func AssertLink(t *testing.T, linkPath, expectedTarget string) {
	t.Helper()

	// Check link exists
	info, err := os.Lstat(linkPath)
	require.NoError(t, err, "symlink should exist: %s", linkPath)
	assert.True(t, info.Mode()&os.ModeSymlink != 0, "path should be a symlink: %s", linkPath)

	// Check target
	target, err := os.Readlink(linkPath)
	require.NoError(t, err, "should read symlink: %s", linkPath)
	assert.Equal(t, expectedTarget, target, "symlink target mismatch")
}

// AssertLinkContains verifies a symlink exists and its target contains the expected substring.
func AssertLinkContains(t *testing.T, linkPath, targetSubstring string) {
	t.Helper()

	info, err := os.Lstat(linkPath)
	require.NoError(t, err, "symlink should exist: %s", linkPath)
	assert.True(t, info.Mode()&os.ModeSymlink != 0, "path should be a symlink: %s", linkPath)

	target, err := os.Readlink(linkPath)
	require.NoError(t, err, "should read symlink: %s", linkPath)
	assert.Contains(t, target, targetSubstring, "symlink target should contain substring")
}

// AssertFile verifies a file exists and has expected content.
func AssertFile(t *testing.T, path, expectedContent string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "file should exist: %s", path)
	assert.False(t, info.IsDir(), "path should be a file, not directory: %s", path)

	content, err := os.ReadFile(path)
	require.NoError(t, err, "should read file: %s", path)
	assert.Equal(t, expectedContent, string(content), "file content mismatch")
}

// AssertFileContains verifies a file exists and contains expected substring.
func AssertFileContains(t *testing.T, path, expectedSubstring string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "file should exist: %s", path)
	assert.False(t, info.IsDir(), "path should be a file, not directory: %s", path)

	content, err := os.ReadFile(path)
	require.NoError(t, err, "should read file: %s", path)
	assert.Contains(t, string(content), expectedSubstring, "file should contain substring")
}

// AssertDir verifies a directory exists.
func AssertDir(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "directory should exist: %s", path)
	assert.True(t, info.IsDir(), "path should be a directory: %s", path)
}

// AssertNotExists verifies a path does not exist.
func AssertNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Lstat(path)
	assert.True(t, os.IsNotExist(err), "path should not exist: %s", path)
}

// AssertFileMode verifies a file has expected permissions.
func AssertFileMode(t *testing.T, path string, expectedMode os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "file should exist: %s", path)
	assert.Equal(t, expectedMode, info.Mode().Perm(), "file mode mismatch")
}

// AssertDirEmpty verifies a directory is empty.
func AssertDirEmpty(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(path)
	require.NoError(t, err, "should read directory: %s", path)
	assert.Empty(t, entries, "directory should be empty: %s", path)
}

// AssertDirHasEntries verifies a directory has expected number of entries.
func AssertDirHasEntries(t *testing.T, path string, count int) {
	t.Helper()

	entries, err := os.ReadDir(path)
	require.NoError(t, err, "should read directory: %s", path)
	assert.Len(t, entries, count, "directory entry count mismatch")
}

// AssertSymlinkChain verifies a chain of symlinks.
func AssertSymlinkChain(t *testing.T, linkPath string, expectedChain []string) {
	t.Helper()

	current := linkPath
	for i, expected := range expectedChain {
		target, err := os.Readlink(current)
		require.NoError(t, err, "should read symlink at chain position %d: %s", i, current)
		assert.Equal(t, expected, target, "symlink target mismatch at chain position %d", i)

		// Prepare for next iteration if not at end
		if i < len(expectedChain)-1 {
			if filepath.IsAbs(target) {
				current = target
			} else {
				current = filepath.Join(filepath.Dir(current), target)
			}
		}
	}
}
