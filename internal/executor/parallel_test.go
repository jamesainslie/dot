package executor

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestExecuteBatch_Concurrent(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create independent operations (no dependencies)
	require.NoError(t, fs.MkdirAll(ctx, "/stow/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	source1 := dot.MustParsePath("/stow/pkg/file1")
	target1 := dot.MustParsePath("/home/file1")
	source2 := dot.MustParsePath("/stow/pkg/file2")
	target2 := dot.MustParsePath("/home/file2")
	source3 := dot.MustParsePath("/stow/pkg/file3")
	target3 := dot.MustParsePath("/home/file3")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source3.String(), []byte("content3"), 0644))

	ops := []dot.Operation{
		dot.NewLinkCreate("link1", source1, target1),
		dot.NewLinkCreate("link2", source2, target2),
		dot.NewLinkCreate("link3", source3, target3),
	}

	checkpoint := exec.checkpoint.Create(ctx)
	result := exec.executeBatch(ctx, ops, checkpoint)

	require.Len(t, result.Executed, 3)
	require.Empty(t, result.Failed)

	// Verify all links created
	require.True(t, fs.Exists(ctx, target1.String()))
	require.True(t, fs.Exists(ctx, target2.String()))
	require.True(t, fs.Exists(ctx, target3.String()))
}

func TestExecuteBatch_PartialFailure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/stow/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	// Mix of success and failure
	source1 := dot.MustParsePath("/stow/pkg/file1")
	target1 := dot.MustParsePath("/home/file1")
	source3 := dot.MustParsePath("/stow/pkg/file3")
	target3 := dot.MustParsePath("/home/file3")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source3.String(), []byte("content3"), 0644))

	ops := []dot.Operation{
		dot.NewLinkCreate("link1", source1, target1),
		// This will fail because parent directory /nonexistent doesn't exist
		dot.NewLinkCreate("link2", dot.MustParsePath("/stow/pkg/file3"), dot.MustParsePath("/nonexistent/file2")),
		dot.NewLinkCreate("link3", source3, target3),
	}

	checkpoint := exec.checkpoint.Create(ctx)
	result := exec.executeBatch(ctx, ops, checkpoint)

	require.Len(t, result.Executed, 2, "two operations should succeed")
	require.Len(t, result.Failed, 1, "one operation should fail")
	require.Contains(t, result.Failed, dot.OperationID("link2"))
}

func TestExecute_ParallelBatches(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create plan with parallelizable operations
	require.NoError(t, fs.MkdirAll(ctx, "/stow/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/dir1", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/dir2", 0755))

	source1 := dot.MustParsePath("/stow/pkg/file1")
	target1 := dot.MustParsePath("/home/dir1/file1")
	source2 := dot.MustParsePath("/stow/pkg/file2")
	target2 := dot.MustParsePath("/home/dir2/file2")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("c1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("c2"), 0644))

	ops := []dot.Operation{
		dot.NewLinkCreate("link1", source1, target1),
		dot.NewLinkCreate("link2", source2, target2),
	}

	// Create plan with parallel batches
	plan := dot.Plan{
		Operations: ops,
		Batches: [][]dot.Operation{
			{ops[0], ops[1]}, // Both in same batch (can run in parallel)
		},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk(), "execution should succeed")
	execResult := result.Unwrap()
	require.Len(t, execResult.Executed, 2)
	require.Empty(t, execResult.Failed)
}

func TestExecuteParallel_Internal_MultipleBatches(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/stow/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	source1 := dot.MustParsePath("/stow/pkg/file1")
	target1 := dot.MustParsePath("/home/file1")
	source2 := dot.MustParsePath("/stow/pkg/file2")
	target2 := dot.MustParsePath("/home/file2")
	source3 := dot.MustParsePath("/stow/pkg/file3")
	target3 := dot.MustParsePath("/home/file3")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source3.String(), []byte("content3"), 0644))

	// Batch 1: two operations in parallel
	batch1Op1 := dot.NewLinkCreate("link1", source1, target1)
	batch1Op2 := dot.NewLinkCreate("link2", source2, target2)

	// Batch 2: depends on batch 1 completing
	batch2Op := dot.NewLinkCreate("link3", source3, target3)

	plan := dot.Plan{
		Operations: []dot.Operation{batch1Op1, batch1Op2, batch2Op},
		Batches: [][]dot.Operation{
			{batch1Op1, batch1Op2}, // Batch 1
			{batch2Op},             // Batch 2
		},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk(), "execution should succeed")
	execResult := result.Unwrap()

	require.Len(t, execResult.Executed, 3, "all operations should execute")
	require.Empty(t, execResult.Failed)

	// Verify all links created
	require.True(t, fs.Exists(ctx, target1.String()))
	require.True(t, fs.Exists(ctx, target2.String()))
	require.True(t, fs.Exists(ctx, target3.String()))
}
