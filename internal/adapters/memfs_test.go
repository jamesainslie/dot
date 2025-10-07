package adapters

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMemFS(t *testing.T) {
	mfs := NewMemFS()
	require.NotNil(t, mfs)
	require.NotNil(t, mfs.files)
}

func TestMemFS_WriteFile_ReadFile(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	// Create parent directory
	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))

	// Write file
	data := []byte("test content")
	err := mfs.WriteFile(ctx, "/home/test.txt", data, 0644)
	require.NoError(t, err)

	// Read file
	read, err := mfs.ReadFile(ctx, "/home/test.txt")
	require.NoError(t, err)
	require.Equal(t, data, read)
}

func TestMemFS_WriteFile_ParentNotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	// Try to write without parent directory
	err := mfs.WriteFile(ctx, "/home/test.txt", []byte("data"), 0644)
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_ReadFile_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	_, err := mfs.ReadFile(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_ReadFile_IsDirectory(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))

	_, err := mfs.ReadFile(ctx, "/home")
	require.Error(t, err)
	require.Contains(t, err.Error(), "directory")
}

func TestMemFS_Mkdir(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	err := mfs.Mkdir(ctx, "/home", 0755)
	require.NoError(t, err)

	exists := mfs.Exists(ctx, "/home")
	require.True(t, exists)

	isDir, err := mfs.IsDir(ctx, "/home")
	require.NoError(t, err)
	require.True(t, isDir)
}

func TestMemFS_Mkdir_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.Mkdir(ctx, "/home", 0755))

	// Try to create again
	err := mfs.Mkdir(ctx, "/home", 0755)
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrExist)
}

func TestMemFS_MkdirAll(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	// Create nested directories
	err := mfs.MkdirAll(ctx, "/home/user/.config", 0755)
	require.NoError(t, err)

	// Verify final directory created
	require.True(t, mfs.Exists(ctx, "/home/user/.config"))
}

func TestMemFS_MkdirAll_Idempotent(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	// Create once
	require.NoError(t, mfs.MkdirAll(ctx, "/home/user", 0755))

	// Create again - should succeed
	err := mfs.MkdirAll(ctx, "/home/user", 0755)
	require.NoError(t, err)
}

func TestMemFS_Remove(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("data"), 0644))

	// Remove file
	err := mfs.Remove(ctx, "/home/file")
	require.NoError(t, err)

	require.False(t, mfs.Exists(ctx, "/home/file"))
}

func TestMemFS_Remove_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	err := mfs.Remove(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_RemoveAll(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	// Create simple directory with file
	require.NoError(t, mfs.MkdirAll(ctx, "/home/user", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/user/file1", []byte("data"), 0644))

	// Remove all
	err := mfs.RemoveAll(ctx, "/home/user")
	require.NoError(t, err)

	// Verify removed
	require.False(t, mfs.Exists(ctx, "/home/user"))
}

func TestMemFS_Symlink(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.MkdirAll(ctx, "/packages", 0755))

	// Create symlink
	err := mfs.Symlink(ctx, "/packages/file", "/home/link")
	require.NoError(t, err)

	// Verify it's a symlink
	isSymlink, err := mfs.IsSymlink(ctx, "/home/link")
	require.NoError(t, err)
	require.True(t, isSymlink)

	// Read link target
	target, err := mfs.ReadLink(ctx, "/home/link")
	require.NoError(t, err)
	require.Equal(t, "/packages/file", target)
}

func TestMemFS_Symlink_ParentNotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	err := mfs.Symlink(ctx, "/source", "/nonexistent/link")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_ReadLink_NotSymlink(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("data"), 0644))

	_, err := mfs.ReadLink(ctx, "/home/file")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a symlink")
}

func TestMemFS_ReadLink_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	_, err := mfs.ReadLink(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_Rename(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/old", []byte("data"), 0644))

	// Rename
	err := mfs.Rename(ctx, "/home/old", "/home/new")
	require.NoError(t, err)

	// Verify old gone, new exists
	require.False(t, mfs.Exists(ctx, "/home/old"))
	require.True(t, mfs.Exists(ctx, "/home/new"))

	// Verify content preserved
	data, err := mfs.ReadFile(ctx, "/home/new")
	require.NoError(t, err)
	require.Equal(t, []byte("data"), data)
}

func TestMemFS_Rename_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	err := mfs.Rename(ctx, "/nonexistent", "/new")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_Exists(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.False(t, mfs.Exists(ctx, "/nonexistent"))

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.True(t, mfs.Exists(ctx, "/home"))
}

func TestMemFS_IsDir(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("data"), 0644))

	// Directory
	isDir, err := mfs.IsDir(ctx, "/home")
	require.NoError(t, err)
	require.True(t, isDir)

	// File
	isDir, err = mfs.IsDir(ctx, "/home/file")
	require.NoError(t, err)
	require.False(t, isDir)
}

func TestMemFS_IsDir_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	_, err := mfs.IsDir(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_IsSymlink(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("data"), 0644))
	require.NoError(t, mfs.Symlink(ctx, "/target", "/home/link"))

	// Symlink
	isSymlink, err := mfs.IsSymlink(ctx, "/home/link")
	require.NoError(t, err)
	require.True(t, isSymlink)

	// File
	isSymlink, err = mfs.IsSymlink(ctx, "/home/file")
	require.NoError(t, err)
	require.False(t, isSymlink)
}

func TestMemFS_IsSymlink_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	_, err := mfs.IsSymlink(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_Stat(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("test"), 0644))

	info, err := mfs.Stat(ctx, "/home/file")
	require.NoError(t, err)
	require.Equal(t, "file", info.Name())
	require.Equal(t, int64(4), info.Size())
	require.False(t, info.IsDir())
}

func TestMemFS_Stat_Directory(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home/user", 0755))

	info, err := mfs.Stat(ctx, "/home/user")
	require.NoError(t, err)
	require.Equal(t, "user", info.Name())
	require.True(t, info.IsDir())
}

func TestMemFS_Stat_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	_, err := mfs.Stat(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_ReadDir(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	// Create directory with files
	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file1", []byte("data1"), 0644))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file2", []byte("data2"), 0644))
	require.NoError(t, mfs.MkdirAll(ctx, "/home/subdir", 0755))

	entries, err := mfs.ReadDir(ctx, "/home")
	require.NoError(t, err)
	require.Len(t, entries, 3)

	// Check entries
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	require.Contains(t, names, "file1")
	require.Contains(t, names, "file2")
	require.Contains(t, names, "subdir")
}

func TestMemFS_ReadDir_NotExist(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	_, err := mfs.ReadDir(ctx, "/nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMemFS_ReadDir_NotDirectory(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("data"), 0644))

	_, err := mfs.ReadDir(ctx, "/home/file")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")
}

func TestMemFS_FileInfo(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("test"), 0644))

	info, err := mfs.Stat(ctx, "/home/file")
	require.NoError(t, err)

	require.Equal(t, "file", info.Name())
	require.Equal(t, int64(4), info.Size())
	require.Equal(t, fs.FileMode(0644), info.Mode())
	require.NotNil(t, info.ModTime())
	require.False(t, info.IsDir())
	require.Nil(t, info.Sys())
}

func TestMemFS_DirEntry(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, mfs.WriteFile(ctx, "/home/file", []byte("data"), 0644))

	entries, err := mfs.ReadDir(ctx, "/home")
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	// Check file entry
	found := false
	for _, e := range entries {
		if e.Name() == "file" {
			found = true
			require.False(t, e.IsDir())
			info, err := e.Info()
			require.NoError(t, err)
			require.Equal(t, "file", info.Name())
			break
		}
	}
	require.True(t, found, "file entry should be found")
}

func TestMemFS_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	mfs := NewMemFS()

	require.NoError(t, mfs.MkdirAll(ctx, "/home", 0755))

	// Concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			path := "/home/file" + string(rune('0'+n))
			_ = mfs.WriteFile(ctx, path, []byte("data"), 0644)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = mfs.Exists(ctx, "/home")
			done <- true
		}()
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}

	// No test assertion - just verify no panics
}
