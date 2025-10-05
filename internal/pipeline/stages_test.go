package pipeline

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before start", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		scanStage := ScanStage()
		input := ScanInput{
			StowDir:   dot.NewPackagePath("/stow").Unwrap(),
			TargetDir: dot.NewTargetPath("/target").Unwrap(),
			Packages:  []string{"vim"},
			IgnoreSet: ignore.NewIgnoreSet(),
			FS:        adapters.NewOSFilesystem(),
		}

		result := scanStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})

	t.Run("empty package list with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		scanStage := ScanStage()
		input := ScanInput{
			StowDir:   dot.NewPackagePath("/stow").Unwrap(),
			TargetDir: dot.NewTargetPath("/target").Unwrap(),
			Packages:  []string{}, // Empty list
			IgnoreSet: ignore.NewIgnoreSet(),
			FS:        adapters.NewOSFilesystem(),
		}

		result := scanStage(ctx, input)

		// Should catch cancellation at early check
		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestPlanStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before planning", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		planStage := PlanStage()
		input := PlanInput{
			Packages:  []dot.Package{},
			TargetDir: dot.NewTargetPath("/target").Unwrap(),
		}

		result := planStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestResolveStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before resolution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		resolveStage := ResolveStage()
		input := ResolveInput{
			Desired: planner.DesiredState{
				Links: make(map[string]planner.LinkSpec),
				Dirs:  make(map[string]planner.DirSpec),
			},
			FS:        adapters.NewOSFilesystem(),
			Policies:  planner.DefaultPolicies(),
			BackupDir: "",
		}

		result := resolveStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestSortStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before sorting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		sortStage := SortStage()
		input := SortInput{
			Operations: []dot.Operation{},
		}

		result := sortStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})

	t.Run("cancelled with operations", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		sortStage := SortStage()

		source := dot.NewFilePath("/stow/vim/vimrc").Unwrap()
		target := dot.NewFilePath("/home/user/.vimrc").Unwrap()

		input := SortInput{
			Operations: []dot.Operation{
				dot.NewLinkCreate("link1", source, target),
			},
		}

		result := sortStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestStages_ValidContextPropagation(t *testing.T) {
	t.Run("all stages respect valid context", func(t *testing.T) {
		ctx := context.Background()

		// Scan stage with empty packages
		scanResult := ScanStage()(ctx, ScanInput{
			StowDir:   dot.NewPackagePath("/stow").Unwrap(),
			TargetDir: dot.NewTargetPath("/target").Unwrap(),
			Packages:  []string{},
			IgnoreSet: ignore.NewIgnoreSet(),
			FS:        adapters.NewOSFilesystem(),
		})
		require.True(t, scanResult.IsOk())

		// Plan stage with empty packages
		planResult := PlanStage()(ctx, PlanInput{
			Packages:  []dot.Package{},
			TargetDir: dot.NewTargetPath("/target").Unwrap(),
		})
		require.True(t, planResult.IsOk())

		// Resolve stage with empty state
		resolveResult := ResolveStage()(ctx, ResolveInput{
			Desired: planner.DesiredState{
				Links: make(map[string]planner.LinkSpec),
				Dirs:  make(map[string]planner.DirSpec),
			},
			FS:        adapters.NewOSFilesystem(),
			Policies:  planner.DefaultPolicies(),
			BackupDir: "",
		})
		require.True(t, resolveResult.IsOk())

		// Sort stage with empty operations
		sortResult := SortStage()(ctx, SortInput{
			Operations: []dot.Operation{},
		})
		require.True(t, sortResult.IsOk())
	})
}
