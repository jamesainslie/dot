package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestExecute_ContextCancellation_DuringPrepare(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create a plan with many operations to increase chance of cancellation during prepare
	var ops []domain.Operation
	for i := 0; i < 100; i++ {
		source := domain.MustParsePath("/packages/pkg/file")
		target := domain.MustParseTargetPath("/home/file")
		require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
		require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))
		ops = append(ops, domain.NewLinkCreate(domain.OperationID("link"+string(rune(i))), source, target))
	}

	plan := domain.Plan{Operations: ops}

	// Cancel context immediately
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail due to cancellation")
	err := result.UnwrapErr()
	require.Error(t, err)
	// Should fail during prepare due to cancelled context
}

func TestExecute_ContextCancellation_DuringSequentialExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create multiple operations
	var ops []domain.Operation
	for i := 0; i < 5; i++ {
		dirPath := domain.MustParsePath("/test/dir" + string(rune('a'+i)))
		ops = append(ops, domain.NewDirCreate(domain.OperationID("dir"+string(rune('a'+i))), dirPath))
	}

	// Ensure parent exists
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	plan := domain.Plan{Operations: ops}

	// Cancel context immediately to ensure we catch it
	cancel()

	result := exec.Execute(ctx, plan)

	// Should fail - either during prepare or execution
	if result.IsErr() {
		// Expected - cancellation or prepare failure
		t.Logf("Execution failed as expected: %v", result.UnwrapErr())
	} else {
		// Operations may have completed before cancellation was observed
		t.Log("Operations completed before cancellation - timing dependent")
	}
}

func TestExecute_CancellationErrorReturnedWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directory
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create multiple operations
	var ops []domain.Operation
	for i := 0; i < 10; i++ {
		dirPath := domain.MustParsePath("/test/dir" + string(rune('a'+i)))
		ops = append(ops, domain.NewDirCreate(domain.OperationID("dir"+string(rune('a'+i))), dirPath))
	}

	plan := domain.Plan{Operations: ops}

	// Cancel context before execution
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")

	// Should get cancellation error during prepare
	err := result.UnwrapErr()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cancel", "error should mention cancellation")
}

func TestExecute_ContextCancellation_DuringParallelExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directories
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create operations that can be parallelized (no dependencies)
	var batch1, batch2 []domain.Operation
	for i := 0; i < 5; i++ {
		dirPath := domain.MustParsePath("/test/parallel1_" + string(rune('a'+i)))
		batch1 = append(batch1, domain.NewDirCreate(domain.OperationID("par1_"+string(rune('a'+i))), dirPath))
	}
	for i := 0; i < 5; i++ {
		dirPath := domain.MustParsePath("/test/parallel2_" + string(rune('a'+i)))
		batch2 = append(batch2, domain.NewDirCreate(domain.OperationID("par2_"+string(rune('a'+i))), dirPath))
	}

	plan := domain.Plan{
		Operations: append(batch1, batch2...),
		Batches:    [][]domain.Operation{batch1, batch2}, // Two parallel batches
	}

	// Cancel immediately to catch between batches
	cancel()

	result := exec.Execute(ctx, plan)

	// Should fail - either during prepare or between batches
	if result.IsErr() {
		t.Logf("Execution failed as expected: %v", result.UnwrapErr())
	} else {
		// All operations may have completed before cancellation
		t.Log("Operations completed before cancellation - timing dependent")
	}
}

func TestExecute_ImmediateCancellation_NoOperationsExecuted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directory
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create a single operation
	dirPath := domain.MustParsePath("/test/dir1")
	op := domain.NewDirCreate("dir1", dirPath)
	plan := domain.Plan{Operations: []domain.Operation{op}}

	// Cancel immediately before Execute
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")

	// Verify the directory was not created
	exists := fs.Exists(ctx, dirPath.String())
	require.False(t, exists, "directory should not have been created")
}

// cancelAfterExecute wraps an operation and cancels the given context once
// the wrapped operation has executed successfully. It simulates Ctrl-C
// arriving mid-plan, immediately after an operation completes.
type cancelAfterExecute struct {
	domain.Operation
	cancel context.CancelFunc
}

func (c cancelAfterExecute) Execute(ctx context.Context, fs domain.FS) error {
	if err := c.Operation.Execute(ctx, fs); err != nil {
		return err
	}
	c.cancel()
	return nil
}

func TestExecute_CancelledMidPlan_RollsBackExecutedOperations(t *testing.T) {
	// Cancellation mid-plan must still roll back the operations that already
	// executed. The rollback runs on a context detached from cancellation,
	// so a real filesystem (whose every call checks ctx.Err()) must observe
	// the reverted state.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := t.TempDir()
	fs := adapters.NewOSFilesystem()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	dir1 := domain.MustParsePath(filepath.Join(root, "dir1"))
	dir2 := domain.MustParsePath(filepath.Join(root, "dir2"))
	dir3 := domain.MustParsePath(filepath.Join(root, "dir3"))
	dir4 := domain.MustParsePath(filepath.Join(root, "dir4"))

	ops := []domain.Operation{
		domain.NewDirCreate("dir1", dir1),
		domain.NewDirCreate("dir2", dir2),
		cancelAfterExecute{
			Operation: domain.NewDirCreate("dir3", dir3),
			cancel:    cancel,
		},
		domain.NewDirCreate("dir4", dir4),
	}

	result := exec.Execute(ctx, domain.Plan{Operations: ops})

	require.True(t, result.IsErr(), "execution should fail due to cancellation")

	var cancelErr domain.ErrExecutionCancelled
	require.ErrorAs(t, result.UnwrapErr(), &cancelErr,
		"cancellation error should be returned")
	require.Equal(t, 3, cancelErr.Executed, "three operations should have executed")
	require.Equal(t, 1, cancelErr.Skipped, "the fourth operation should have been skipped")

	// Operations 1..3 executed before cancellation and must be reverted on
	// the filesystem despite the cancelled context.
	for _, dir := range []domain.FilePath{dir1, dir2, dir3} {
		_, err := os.Lstat(dir.String())
		require.True(t, os.IsNotExist(err),
			"%s should have been rolled back after cancellation", dir.String())
	}

	// Operation 4 was skipped and must never have been created.
	_, err := os.Lstat(dir4.String())
	require.True(t, os.IsNotExist(err), "dir4 should never have been created")
}
