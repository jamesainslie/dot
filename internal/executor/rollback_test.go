package executor

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestRollback_SingleOperation(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Set up filesystem and create link
	source := dot.MustParsePath("/packages/pkg/file")
	target := dot.MustParsePath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))
	require.NoError(t, fs.Symlink(ctx, source.String(), target.String()))

	// Create checkpoint with the operation
	checkpoint := exec.checkpoint.Create(ctx)
	op := dot.NewLinkCreate("link1", source, target)
	checkpoint.Record("link1", op)

	// Rollback
	rolledBack := exec.rollback(ctx, []dot.OperationID{"link1"}, checkpoint)

	require.Len(t, rolledBack, 1)
	require.Contains(t, rolledBack, dot.OperationID("link1"))

	// Verify link was removed
	exists := fs.Exists(ctx, target.String())
	require.False(t, exists, "link should be removed after rollback")
}

func TestRollback_ReverseOrder(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create operations in order: DirCreate, then LinkCreate
	dirPath := dot.MustParsePath("/home/subdir")
	source := dot.MustParsePath("/packages/pkg/file")
	target := dot.MustParsePath("/home/subdir/file")

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))
	require.NoError(t, fs.MkdirAll(ctx, dirPath.String(), 0755))
	require.NoError(t, fs.Symlink(ctx, source.String(), target.String()))

	checkpoint := exec.checkpoint.Create(ctx)

	dirOp := dot.NewDirCreate("dir1", dirPath)
	linkOp := dot.NewLinkCreate("link1", source, target)

	checkpoint.Record("dir1", dirOp)
	checkpoint.Record("link1", linkOp)

	// Rollback should happen in reverse order: link first, then dir
	executed := []dot.OperationID{"dir1", "link1"}
	rolledBack := exec.rollback(ctx, executed, checkpoint)

	require.Len(t, rolledBack, 2)

	// Verify both were removed
	require.False(t, fs.Exists(ctx, target.String()), "link should be removed")
	require.False(t, fs.Exists(ctx, dirPath.String()), "directory should be removed")
}

func TestRollback_PartialRollbackOnError(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create two links
	source1 := dot.MustParsePath("/packages/pkg/file1")
	target1 := dot.MustParsePath("/home/file1")
	source2 := dot.MustParsePath("/packages/pkg/file2")
	target2 := dot.MustParsePath("/home/file2")

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))
	require.NoError(t, fs.Symlink(ctx, source1.String(), target1.String()))
	// Don't create second link - rollback will fail for it

	checkpoint := exec.checkpoint.Create(ctx)
	op1 := dot.NewLinkCreate("link1", source1, target1)
	op2 := dot.NewLinkCreate("link2", source2, target2)

	checkpoint.Record("link1", op1)
	checkpoint.Record("link2", op2)

	// Rollback both - first should succeed, second should fail (doesn't exist)
	executed := []dot.OperationID{"link1", "link2"}
	rolledBack := exec.rollback(ctx, executed, checkpoint)

	// Should have rolled back link1 even though link2 failed
	require.Len(t, rolledBack, 1)
	require.Contains(t, rolledBack, dot.OperationID("link1"))
	require.False(t, fs.Exists(ctx, target1.String()), "link1 should be removed")
}

func TestExecute_AutomaticRollback(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create a scenario where prepare passes but execute fails
	// We'll test by directly calling executeSequential with a checkpoint
	source1 := dot.MustParsePath("/packages/pkg/file1")
	target1 := dot.MustParsePath("/home/file1")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))

	op1 := dot.NewLinkCreate("link1", source1, target1)

	// Second operation will fail during execution (parent doesn't exist)
	source2 := dot.MustParsePath("/packages/pkg/file2")
	target2 := dot.MustParsePath("/nonexistent/file2")
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))

	op2 := dot.NewLinkCreate("link2", source2, target2)

	// Create checkpoint and execute manually (bypassing prepare)
	checkpoint := exec.checkpoint.Create(ctx)
	execResult := exec.executeSequential(ctx, dot.Plan{Operations: []dot.Operation{op1, op2}}, checkpoint)

	require.Len(t, execResult.Executed, 1, "first operation should execute")
	require.Len(t, execResult.Failed, 1, "second operation should fail")

	// Now rollback
	rolledBack := exec.rollback(ctx, execResult.Executed, checkpoint)
	require.Len(t, rolledBack, 1, "first operation should be rolled back")

	// Verify first operation was rolled back
	exists := fs.Exists(ctx, target1.String())
	require.False(t, exists, "rolled back operation should be undone")
}
