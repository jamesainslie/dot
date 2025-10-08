package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	tmpDir := t.TempDir()

	// Create some files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir/file2.txt"), []byte("content2"), 0644))
	require.NoError(t, os.Symlink("/some/target", filepath.Join(tmpDir, "link")))

	snapshot := CaptureState(t, tmpDir)

	assert.Equal(t, tmpDir, snapshot.Root)
	assert.Len(t, snapshot.Files, 4) // file1.txt, subdir, subdir/file2.txt, link

	// Verify symlink captured
	assert.True(t, snapshot.HasPath("link"))
	state, found := snapshot.GetState("link")
	require.True(t, found)
	assert.True(t, state.IsSymlink)
	assert.Equal(t, "/some/target", state.Target)
}

func TestCompareStates_NoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644))

	before := CaptureState(t, tmpDir)
	after := CaptureState(t, tmpDir)

	diffs := CompareStates(t, before, after)
	assert.Empty(t, diffs)
}

func TestCompareStates_Added(t *testing.T) {
	tmpDir := t.TempDir()

	before := CaptureState(t, tmpDir)

	// Add file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("content"), 0644))

	after := CaptureState(t, tmpDir)

	diffs := CompareStates(t, before, after)
	assert.Contains(t, diffs, "+ file new.txt")
}

func TestCompareStates_Removed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file
	filePath := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))

	before := CaptureState(t, tmpDir)

	// Remove file
	require.NoError(t, os.Remove(filePath))

	after := CaptureState(t, tmpDir)

	diffs := CompareStates(t, before, after)
	assert.Contains(t, diffs, "- file file.txt")
}

func TestCompareStates_SymlinkAdded(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	tmpDir := t.TempDir()

	before := CaptureState(t, tmpDir)

	// Add symlink
	require.NoError(t, os.Symlink("/target", filepath.Join(tmpDir, "link")))

	after := CaptureState(t, tmpDir)

	diffs := CompareStates(t, before, after)
	assert.Contains(t, diffs, "+ symlink link -> /target")
}

func TestStateSnapshot_Counts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	tmpDir := t.TempDir()

	// Create various items
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("b"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755))
	require.NoError(t, os.Symlink("/target", filepath.Join(tmpDir, "link1")))

	snapshot := CaptureState(t, tmpDir)

	assert.Equal(t, 3, snapshot.CountFiles()) // 2 regular files + 1 symlink
	assert.Equal(t, 1, snapshot.CountSymlinks())
	assert.Equal(t, 1, snapshot.CountDirs())
}

func TestStateSnapshot_HasPath(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644))

	snapshot := CaptureState(t, tmpDir)

	assert.True(t, snapshot.HasPath("file.txt"))
	assert.False(t, snapshot.HasPath("nonexistent.txt"))
}

func TestStateSnapshot_GetState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	tmpDir := t.TempDir()
	require.NoError(t, os.Symlink("/target", filepath.Join(tmpDir, "link")))

	snapshot := CaptureState(t, tmpDir)

	state, found := snapshot.GetState("link")
	require.True(t, found)
	assert.True(t, state.IsSymlink)
	assert.Equal(t, "/target", state.Target)

	_, found = snapshot.GetState("nonexistent")
	assert.False(t, found)
}
