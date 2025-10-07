package executor

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	tracer := adapters.NewNoopTracer()

	exec := New(Opts{
		FS:     fs,
		Logger: logger,
		Tracer: tracer,
	})

	require.NotNil(t, exec)
}

func TestNewExecutor_DefaultCheckpoint(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	tracer := adapters.NewNoopTracer()

	exec := New(Opts{
		FS:     fs,
		Logger: logger,
		Tracer: tracer,
		// No checkpoint store provided - should use default
	})

	require.NotNil(t, exec)
	require.NotNil(t, exec.checkpoint)
}

func TestExecute_EmptyPlan(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Empty plan
	plan := dot.Plan{
		Operations: []dot.Operation{},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr())
	require.IsType(t, dot.ErrEmptyPlan{}, result.UnwrapErr())
}

func TestExecute_SingleOperation_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create source file with parent directories
	source := dot.MustParsePath("/packages/pkg/file")
	target := dot.MustParsePath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	// Create operation
	op := dot.NewLinkCreate("link1", source, target)

	plan := dot.Plan{
		Operations: []dot.Operation{op},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk(), "execution should succeed")

	// Verify symlink created
	exists := fs.Exists(ctx, target.String())
	require.True(t, exists, "symlink should be created")
}

func TestExecute_OperationFailure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create operation that will fail (source doesn't exist)
	source := dot.MustParsePath("/nonexistent")
	target := dot.MustParsePath("/home/file")
	op := dot.NewLinkCreate("link1", source, target)

	plan := dot.Plan{
		Operations: []dot.Operation{op},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")
}

func TestExecute_MultipleOperations_PartialFailure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// First operation succeeds
	source1 := dot.MustParsePath("/packages/pkg/file1")
	target1 := dot.MustParsePath("/home/file1")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	op1 := dot.NewLinkCreate("link1", source1, target1)

	// Second operation fails (source doesn't exist)
	source2 := dot.MustParsePath("/nonexistent")
	target2 := dot.MustParsePath("/home/file2")
	op2 := dot.NewLinkCreate("link2", source2, target2)

	plan := dot.Plan{
		Operations: []dot.Operation{op1, op2},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail due to second operation")

	// First operation should have been executed (and then rolled back)
	// We'll verify rollback behavior in later tests
}
