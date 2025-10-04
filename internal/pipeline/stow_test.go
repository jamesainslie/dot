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

func TestNewStowPipeline(t *testing.T) {
	fs := adapters.NewOSFilesystem()
	ignoreSet := ignore.NewIgnoreSet()

	pipeline := NewStowPipeline(StowPipelineOpts{
		FS:        fs,
		IgnoreSet: ignoreSet,
		Policies:  planner.DefaultPolicies(),
	})

	assert.NotNil(t, pipeline)
}

func TestStowPipeline_Execute(t *testing.T) {
	t.Run("empty package list", func(t *testing.T) {
		fs := adapters.NewOSFilesystem()
		ignoreSet := ignore.NewIgnoreSet()

		pipeline := NewStowPipeline(StowPipelineOpts{
			FS:        fs,
			IgnoreSet: ignoreSet,
			Policies:  planner.DefaultPolicies(),
		})

		stowPath := dot.NewPackagePath("/stow").Unwrap()
		targetPath := dot.NewTargetPath("/target").Unwrap()

		result := pipeline.Execute(context.Background(), StowInput{
			StowDir:   stowPath,
			TargetDir: targetPath,
			Packages:  []string{},
		})

		require.True(t, result.IsOk())
		plan := result.Unwrap()
		assert.Empty(t, plan.Operations)
	})

	t.Run("package not found", func(t *testing.T) {
		fs := adapters.NewOSFilesystem()
		ignoreSet := ignore.NewIgnoreSet()

		pipeline := NewStowPipeline(StowPipelineOpts{
			FS:        fs,
			IgnoreSet: ignoreSet,
			Policies:  planner.DefaultPolicies(),
		})

		stowPath := dot.NewPackagePath("/stow").Unwrap()
		targetPath := dot.NewTargetPath("/target").Unwrap()

		result := pipeline.Execute(context.Background(), StowInput{
			StowDir:   stowPath,
			TargetDir: targetPath,
			Packages:  []string{"nonexistent"},
		})

		require.False(t, result.IsOk())
		err := result.UnwrapErr()

		var pkgErr dot.ErrPackageNotFound
		assert.ErrorAs(t, err, &pkgErr)
	})
}

func TestStowPipeline_Composition(t *testing.T) {
	t.Run("stages compose correctly", func(t *testing.T) {
		// Test individual stages exist
		scanStage := ScanStage()
		planStage := PlanStage()
		resolveStage := ResolveStage()
		sortStage := SortStage()

		assert.NotNil(t, scanStage)
		assert.NotNil(t, planStage)
		assert.NotNil(t, resolveStage)
		assert.NotNil(t, sortStage)
	})
}
