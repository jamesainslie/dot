package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

// LinkDelete execute-time safety: the operation must never delete anything
// that is not a symlink, and when a destination was recorded at plan time it
// must only delete a symlink that still points there.

func TestLinkDelete_Execute_RefusesRegularFile(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/target/link", []byte("precious user data"), 0644))

	target := domain.MustParseTargetPath("/target/link")
	op := domain.NewLinkDelete("del1", target)

	err := op.Execute(ctx, fs)
	require.Error(t, err, "deleting a regular file must be refused")

	var unsafeErr domain.ErrUnsafeDelete
	require.ErrorAs(t, err, &unsafeErr)
	assert.Equal(t, "/target/link", unsafeErr.Path)

	// The file must be untouched
	data, readErr := fs.ReadFile(ctx, "/target/link")
	require.NoError(t, readErr)
	assert.Equal(t, []byte("precious user data"), data)
}

func TestLinkDelete_Execute_RefusesUnexpectedDestination(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/elsewhere", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/elsewhere/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/elsewhere/file", "/target/link"))

	target := domain.MustParseTargetPath("/target/link")
	op := domain.NewLinkDeleteWithDestination("del1", target, "/packages/pkg/file")

	err := op.Execute(ctx, fs)
	require.Error(t, err, "deleting a symlink pointing somewhere unexpected must be refused")

	var unsafeErr domain.ErrUnsafeDelete
	require.ErrorAs(t, err, &unsafeErr)

	// The symlink must be untouched
	isLink, linkErr := fs.IsSymlink(ctx, "/target/link")
	require.NoError(t, linkErr)
	assert.True(t, isLink)
}

func TestLinkDelete_Execute_DeletesMatchingLink(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/packages/pkg/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/packages/pkg/file", "/target/link"))

	target := domain.MustParseTargetPath("/target/link")
	op := domain.NewLinkDeleteWithDestination("del1", target, "/packages/pkg/file")

	require.NoError(t, op.Execute(ctx, fs))
	assert.False(t, fs.Exists(ctx, "/target/link"))
}

func TestLinkDelete_Execute_MissingTargetIsNoOp(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))

	target := domain.MustParseTargetPath("/target/link")
	op := domain.NewLinkDeleteWithDestination("del1", target, "/packages/pkg/file")

	assert.NoError(t, op.Execute(ctx, fs))
}

func TestLinkDelete_Rollback_RecreatesRecordedLink(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/packages/pkg/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/packages/pkg/file", "/target/link"))

	target := domain.MustParseTargetPath("/target/link")
	op := domain.NewLinkDeleteWithDestination("del1", target, "/packages/pkg/file")

	require.NoError(t, op.Execute(ctx, fs))
	require.False(t, fs.Exists(ctx, "/target/link"))

	require.NoError(t, op.Rollback(ctx, fs))

	isLink, err := fs.IsSymlink(ctx, "/target/link")
	require.NoError(t, err)
	assert.True(t, isLink, "rollback must recreate the deleted symlink")

	dest, err := fs.ReadLink(ctx, "/target/link")
	require.NoError(t, err)
	assert.Equal(t, "/packages/pkg/file", dest)
}

func TestLinkDelete_Rollback_WithoutDestinationReportsImpossible(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))

	target := domain.MustParseTargetPath("/target/link")
	op := domain.NewLinkDelete("del1", target)

	err := op.Rollback(ctx, fs)
	require.Error(t, err, "rollback without a recorded destination must report failure")

	var impossibleErr domain.ErrRollbackImpossible
	assert.ErrorAs(t, err, &impossibleErr)
}

// FileDelete rollback: must restore the deleted file from a recorded backup.

func TestFileDelete_Rollback_RestoresFromBackup(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/backups", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/target/file", []byte("original"), 0640))

	path := domain.MustParsePath("/target/file")
	backup := domain.MustParsePath("/backups/file.bak")

	// Simulate the planner backup triple: back up, then delete
	backupOp := domain.NewFileBackup("bak1", path, backup)
	require.NoError(t, backupOp.Execute(ctx, fs))

	op := domain.NewFileDeleteWithBackup("del1", path, backup)
	require.NoError(t, op.Execute(ctx, fs))
	require.False(t, fs.Exists(ctx, "/target/file"))

	require.NoError(t, op.Rollback(ctx, fs))

	data, err := fs.ReadFile(ctx, "/target/file")
	require.NoError(t, err)
	assert.Equal(t, []byte("original"), data, "rollback must restore the file content from backup")
}

func TestFileDelete_Rollback_WithoutBackupReportsImpossible(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	path := domain.MustParsePath("/target/file")
	op := domain.NewFileDelete("del1", path)

	err := op.Rollback(ctx, fs)
	require.Error(t, err, "rollback without a recorded backup must report failure")

	var impossibleErr domain.ErrRollbackImpossible
	assert.ErrorAs(t, err, &impossibleErr)
}

// FileBackup rollback: must never delete the backup while the original is gone.

func TestFileBackup_Rollback_PreservesBackupWhenOriginalMissing(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.bak", []byte("backup"), 0644))

	source := domain.MustParsePath("/test/file")
	backup := domain.MustParsePath("/test/file.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	err := op.Rollback(ctx, fs)
	require.Error(t, err, "rollback must refuse to delete the only remaining copy")
	assert.True(t, fs.Exists(ctx, "/test/file.bak"), "backup must be preserved")
}

func TestFileBackup_Rollback_RemovesBackupWhenOriginalPresent(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/file", []byte("original"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/test/file.bak", []byte("original"), 0644))

	source := domain.MustParsePath("/test/file")
	backup := domain.MustParsePath("/test/file.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	require.NoError(t, op.Rollback(ctx, fs))
	assert.False(t, fs.Exists(ctx, "/test/file.bak"))
	assert.True(t, fs.Exists(ctx, "/test/file"))
}

// DirRemoveAll rollback is genuinely impossible and must say so.

func TestDirRemoveAll_Rollback_ReportsImpossible(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	path := domain.MustParsePath("/parent/deleteddir")
	op := domain.NewDirRemoveAll("del1", path)

	err := op.Rollback(ctx, fs)
	require.Error(t, err, "rollback of a recursive delete must report that it cannot restore")

	var impossibleErr domain.ErrRollbackImpossible
	assert.ErrorAs(t, err, &impossibleErr)
}

// DirDelete rollback must restore the recorded permissions.

func TestDirDelete_Rollback_RestoresRecordedMode(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/parent", 0755))
	require.NoError(t, fs.Mkdir(ctx, "/parent/secret", 0700))

	path := domain.MustParsePath("/parent/secret")
	op := domain.NewDirDeleteWithMode("del1", path, 0700)

	require.NoError(t, op.Execute(ctx, fs))
	require.False(t, fs.Exists(ctx, "/parent/secret"))

	require.NoError(t, op.Rollback(ctx, fs))

	info, err := fs.Stat(ctx, "/parent/secret")
	require.NoError(t, err)
	assert.Equal(t, uint32(0700), uint32(info.Mode().Perm()), "rollback must restore the recorded permissions")
}
