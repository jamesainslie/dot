package domain_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkCreate_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))

	source := domain.MustParsePath("/source/file")
	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkCreate("link1", source, target)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify link was created
	isLink, _ := fs.IsSymlink(ctx, "/target/link")
	assert.True(t, isLink)
}

func TestLinkCreate_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/source/file", "/target/link"))

	source := domain.MustParsePath("/source/file")
	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkCreate("link1", source, target)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify link was removed
	assert.False(t, fs.Exists(ctx, "/target/link"))
}

func TestLinkDelete_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/source/file", "/target/link"))

	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkDelete("del1", target)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify link was deleted
	assert.False(t, fs.Exists(ctx, "/target/link"))
}

func TestLinkDelete_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file", []byte("data"), 0644))

	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	// LinkDelete rollback needs the original source to recreate the link
	// Since we don't store that, rollback returns ErrNotImplemented
	op := domain.NewLinkDelete("del1", target)

	err := op.Rollback(ctx, fs)
	// LinkDelete rollback returns nil (cannot restore without knowing source)
	assert.NoError(t, err)
}

func TestDirCreate_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent", 0755))

	path := domain.MustParsePath("/parent/newdir")
	op := domain.NewDirCreate("dir1", path)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify directory was created
	isDir, _ := fs.IsDir(ctx, "/parent/newdir")
	assert.True(t, isDir)
}

func TestDirCreate_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent/dir", 0755))

	path := domain.MustParsePath("/parent/dir")
	op := domain.NewDirCreate("dir1", path)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify directory was removed
	assert.False(t, fs.Exists(ctx, "/parent/dir"))
}

func TestDirDelete_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent/dir", 0755))

	path := domain.MustParsePath("/parent/dir")
	op := domain.NewDirDelete("del1", path)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify directory was deleted
	assert.False(t, fs.Exists(ctx, "/parent/dir"))
}

func TestDirDelete_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent", 0755))

	path := domain.MustParsePath("/parent/restoreddir")
	op := domain.NewDirDelete("del1", path)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify directory was recreated
	exists := fs.Exists(ctx, "/parent/restoreddir")
	assert.True(t, exists)
}

func TestFileBackup_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.txt", []byte("original"), 0644))

	source := domain.MustParsePath("/test/file.txt")
	backup := domain.MustParsePath("/test/file.txt.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify backup was created
	assert.True(t, fs.Exists(ctx, "/test/file.txt.bak"))

	// Verify content
	data, _ := fs.ReadFile(ctx, "/test/file.txt.bak")
	assert.Equal(t, []byte("original"), data)
}

func TestFileBackup_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.bak", []byte("backup"), 0644))

	source := domain.MustParsePath("/test/file")
	backup := domain.MustParsePath("/test/file.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify backup was deleted
	assert.False(t, fs.Exists(ctx, "/test/file.bak"))
}
