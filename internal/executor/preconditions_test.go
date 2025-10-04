package executor

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestCheckDirCreatePreconditions_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directory
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	dirPath := dot.MustParsePath("/home/subdir")
	op := dot.NewDirCreate("dir1", dirPath)

	err := exec.checkDirCreatePreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckDirCreatePreconditions_ParentNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Parent doesn't exist
	dirPath := dot.MustParsePath("/nonexistent/subdir")
	op := dot.NewDirCreate("dir1", dirPath)

	err := exec.checkDirCreatePreconditions(ctx, op)
	require.Error(t, err)
	require.IsType(t, dot.ErrParentNotFound{}, err)
}

func TestCheckDirCreatePreconditions_TopLevelDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create root so top-level directory can be created
	require.NoError(t, fs.MkdirAll(ctx, "/", 0755))

	dirPath := dot.MustParsePath("/toplevel")
	op := dot.NewDirCreate("dir1", dirPath)

	err := exec.checkDirCreatePreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckFileMovePreconditions_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Set up source and destination parent
	source := dot.MustParsePath("/home/file")
	dest := dot.MustParsePath("/stow/pkg/file")
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/stow/pkg", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	op := dot.NewFileMove("move1", source, dest)

	err := exec.checkFileMovePreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckFileMovePreconditions_SourceNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source := dot.MustParsePath("/nonexistent")
	dest := dot.MustParsePath("/stow/pkg/file")
	require.NoError(t, fs.MkdirAll(ctx, "/stow/pkg", 0755))

	op := dot.NewFileMove("move1", source, dest)

	err := exec.checkFileMovePreconditions(ctx, op)
	require.Error(t, err)
	require.IsType(t, dot.ErrSourceNotFound{}, err)
}

func TestCheckFileMovePreconditions_DestParentNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source := dot.MustParsePath("/home/file")
	dest := dot.MustParsePath("/nonexistent/file")
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	op := dot.NewFileMove("move1", source, dest)

	err := exec.checkFileMovePreconditions(ctx, op)
	require.Error(t, err)
	require.IsType(t, dot.ErrParentNotFound{}, err)
}

func TestCheckPreconditions_LinkDelete(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	target := dot.MustParsePath("/home/file")
	op := dot.NewLinkDelete("link1", target)

	// LinkDelete has no preconditions - should return nil
	err := exec.checkPreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckPreconditions_DirDelete(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	dirPath := dot.MustParsePath("/home/dir")
	op := dot.NewDirDelete("dir1", dirPath)

	// DirDelete has no preconditions - should return nil
	err := exec.checkPreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckPreconditions_FileBackup(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source := dot.MustParsePath("/home/file")
	backup := dot.MustParsePath("/home/file.bak")
	op := dot.NewFileBackup("backup1", source, backup)

	// FileBackup has no preconditions - should return nil
	err := exec.checkPreconditions(ctx, op)
	require.NoError(t, err)
}
