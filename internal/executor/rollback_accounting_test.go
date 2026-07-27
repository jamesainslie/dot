package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

// TestRollback_FailedRollbacksAreNotCountedAsRestored verifies that operations
// whose rollback fails (or is impossible) are reported separately instead of
// being silently counted as restored.
func TestRollback_FailedRollbacksAreNotCountedAsRestored(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// DirRemoveAll rollback is impossible: it must be reported as failed.
	dirPath := domain.MustParsePath("/packages/pkg")
	removeOp := domain.NewDirRemoveAll("purge1", dirPath)

	// LinkCreate rollback succeeds.
	source := domain.MustParsePath("/packages/other/file")
	target := domain.MustParseTargetPath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/other", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))
	require.NoError(t, fs.Symlink(ctx, source.String(), target.String()))
	linkOp := domain.NewLinkCreate("link1", source, target)

	checkpoint := exec.checkpoint.Create(ctx)
	checkpoint.Record("purge1", removeOp)
	checkpoint.Record("link1", linkOp)

	executed := []domain.OperationID{"purge1", "link1"}
	rolledBack, rollbackFailed := exec.rollback(ctx, executed, checkpoint)

	assert.Len(t, rolledBack, 1, "only the link rollback actually restored state")
	assert.Contains(t, rolledBack, domain.OperationID("link1"))
	assert.Len(t, rollbackFailed, 1, "the impossible rollback must be reported as failed")
	assert.Contains(t, rollbackFailed, domain.OperationID("purge1"))
}

// TestExecute_SurfacesRollbackFailuresInError verifies that ErrExecutionFailed
// reports how many operations could not be rolled back.
func TestExecute_SurfacesRollbackFailuresInError(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Op 1: DirRemoveAll succeeds but cannot be rolled back.
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/packages/pkg/file", []byte("content"), 0644))
	removeOp := domain.NewDirRemoveAll("purge1", domain.MustParsePath("/packages/pkg"))

	// Op 2: LinkCreate fails during execution (source does not exist),
	// triggering rollback of op 1.
	missingSource := domain.MustParsePath("/packages/missing/file")
	target := domain.MustParseTargetPath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/packages/missing", 0755))
	require.NoError(t, fs.WriteFile(ctx, missingSource.String(), []byte("content"), 0644))
	linkOp := failingOperation{inner: domain.NewLinkCreate("link1", missingSource, target)}

	plan := domain.Plan{Operations: []domain.Operation{removeOp, linkOp}}
	result := exec.Execute(ctx, plan)
	require.False(t, result.IsOk())

	var execErr domain.ErrExecutionFailed
	require.ErrorAs(t, result.UnwrapErr(), &execErr)
	assert.Equal(t, 1, execErr.RollbackFailed, "the impossible rollback must be surfaced")
	assert.Equal(t, 0, execErr.RolledBack)
	assert.Contains(t, execErr.Error(), "could not be rolled back")
}

// failingOperation wraps an operation and always fails at execute time.
type failingOperation struct {
	inner domain.Operation
}

func (f failingOperation) ID() domain.OperationID           { return f.inner.ID() }
func (f failingOperation) Kind() domain.OperationKind       { return f.inner.Kind() }
func (f failingOperation) Validate() error                  { return nil }
func (f failingOperation) Dependencies() []domain.Operation { return nil }
func (f failingOperation) Execute(ctx context.Context, fs domain.FS) error {
	return assert.AnError
}
func (f failingOperation) Rollback(ctx context.Context, fs domain.FS) error { return nil }
func (f failingOperation) String() string                                   { return "always fails" }
func (f failingOperation) Equals(other domain.Operation) bool               { return false }

// TestExecute_BackupTripleRollbackRestoresOriginal verifies that rolling back
// the planner backup triple (FileBackup, FileDelete, LinkCreate) restores the
// original file from the backup and then removes the backup.
func TestExecute_BackupTripleRollbackRestoresOriginal(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/.dot-backup", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/packages/bash", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/home/.bashrc", []byte("original"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/packages/bash/dot-bashrc", []byte("packaged"), 0644))

	conflictPath := domain.MustParsePath("/home/.bashrc")
	backupPath := domain.MustParsePath("/home/.dot-backup/.bashrc.bak")
	source := domain.MustParsePath("/packages/bash/dot-bashrc")
	target := domain.MustParseTargetPath("/home/.bashrc")

	backupOp := domain.NewFileBackup("bak1", conflictPath, backupPath)
	deleteOp := domain.NewFileDeleteWithBackup("del1", conflictPath, backupPath)
	linkOp := domain.NewLinkCreate("link1", source, target)
	failOp := failingOperation{inner: domain.NewLinkCreate("fail1", source, target)}

	plan := domain.Plan{Operations: []domain.Operation{backupOp, deleteOp, linkOp, failOp}}
	result := exec.Execute(ctx, plan)
	require.False(t, result.IsOk(), "the final operation must fail the plan")

	var execErr domain.ErrExecutionFailed
	require.ErrorAs(t, result.UnwrapErr(), &execErr)
	assert.Equal(t, 3, execErr.RolledBack, "all three triple operations must roll back")
	assert.Equal(t, 0, execErr.RollbackFailed)

	// The original file must be restored, not the symlink.
	isLink, err := fs.IsSymlink(ctx, "/home/.bashrc")
	require.NoError(t, err)
	assert.False(t, isLink, "rollback must leave the original regular file")

	data, err := fs.ReadFile(ctx, "/home/.bashrc")
	require.NoError(t, err)
	assert.Equal(t, []byte("original"), data, "the original content must be restored from backup")

	assert.False(t, fs.Exists(ctx, "/home/.dot-backup/.bashrc.bak"), "the backup must be cleaned up after restore")
}

// TestPrepare_LinkDeleteOverRegularFileFails verifies that the prepare phase
// refuses a plan containing a LinkDelete whose target is a regular file.
func TestPrepare_LinkDeleteOverRegularFileFails(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/home/file", []byte("user data"), 0644))

	target := domain.MustParseTargetPath("/home/file")
	op := domain.NewLinkDelete("del1", target)

	plan := domain.Plan{Operations: []domain.Operation{op}}
	result := exec.Execute(ctx, plan)
	require.False(t, result.IsOk(), "prepare must reject deleting a regular file as a link")

	var unsafeErr domain.ErrUnsafeDelete
	assert.ErrorAs(t, result.UnwrapErr(), &unsafeErr)

	// The file must be untouched.
	data, err := fs.ReadFile(ctx, "/home/file")
	require.NoError(t, err)
	assert.Equal(t, []byte("user data"), data)
}

// TestPrepare_LinkDeleteWrongDestinationFails verifies the prepare phase
// rejects a LinkDelete whose symlink no longer points at the recorded
// destination.
func TestPrepare_LinkDeleteWrongDestinationFails(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/elsewhere", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/elsewhere/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/elsewhere/file", "/home/link"))

	target := domain.MustParseTargetPath("/home/link")
	op := domain.NewLinkDeleteWithDestination("del1", target, "/packages/pkg/file")

	plan := domain.Plan{Operations: []domain.Operation{op}}
	result := exec.Execute(ctx, plan)
	require.False(t, result.IsOk(), "prepare must reject deleting a repointed symlink")

	var unsafeErr domain.ErrUnsafeDelete
	assert.ErrorAs(t, result.UnwrapErr(), &unsafeErr)

	isLink, err := fs.IsSymlink(ctx, "/home/link")
	require.NoError(t, err)
	assert.True(t, isLink, "the symlink must be untouched")
}
