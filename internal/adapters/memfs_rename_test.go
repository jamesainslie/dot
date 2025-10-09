package adapters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemFS_Rename_Directory_WithChildren(t *testing.T) {
	fs := NewMemFS()
	ctx := context.Background()

	// Create directory with files
	require.NoError(t, fs.MkdirAll(ctx, "/old/dir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/old/dir/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/old/dir/file2.txt", []byte("content2"), 0644))
	require.NoError(t, fs.MkdirAll(ctx, "/old/dir/subdir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/old/dir/subdir/file3.txt", []byte("content3"), 0644))

	require.NoError(t, fs.MkdirAll(ctx, "/new", 0755))

	// Rename directory
	err := fs.Rename(ctx, "/old/dir", "/new/dir")
	require.NoError(t, err)

	// Verify old location is gone
	assert.False(t, fs.Exists(ctx, "/old/dir"))
	assert.False(t, fs.Exists(ctx, "/old/dir/file1.txt"))

	// Verify new location has all files
	assert.True(t, fs.Exists(ctx, "/new/dir"))
	assert.True(t, fs.Exists(ctx, "/new/dir/file1.txt"))
	assert.True(t, fs.Exists(ctx, "/new/dir/file2.txt"))
	assert.True(t, fs.Exists(ctx, "/new/dir/subdir/file3.txt"))

	// Verify contents
	data, err := fs.ReadFile(ctx, "/new/dir/file1.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), data)

	data, err = fs.ReadFile(ctx, "/new/dir/subdir/file3.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("content3"), data)
}

func TestMemFS_Rename_File(t *testing.T) {
	fs := NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/dir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/dir/oldname.txt", []byte("content"), 0644))

	err := fs.Rename(ctx, "/dir/oldname.txt", "/dir/newname.txt")
	require.NoError(t, err)

	// Old should not exist
	assert.False(t, fs.Exists(ctx, "/dir/oldname.txt"))

	// New should exist with same content
	assert.True(t, fs.Exists(ctx, "/dir/newname.txt"))
	data, _ := fs.ReadFile(ctx, "/dir/newname.txt")
	assert.Equal(t, []byte("content"), data)
}

func TestMemFS_Rename_NonExistent(t *testing.T) {
	fs := NewMemFS()
	ctx := context.Background()

	err := fs.Rename(ctx, "/nonexistent", "/other")
	assert.Error(t, err)
}
