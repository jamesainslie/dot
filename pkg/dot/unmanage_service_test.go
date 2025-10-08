package dot

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmanageService_Unmanage(t *testing.T) {
	t.Run("unmanages package successfully", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup and manage package first
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		// Manage first
		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:        fs,
			IgnoreSet: ignore.NewDefaultIgnoreSet(),
			Policies:  planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, packageDir, targetDir, false)

		err := manageSvc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link created
		assert.True(t, fs.Exists(ctx, targetDir+"/.vimrc"))

		// Now unmanage
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, targetDir, false)
		err = unmanageSvc.Unmanage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link removed
		assert.False(t, fs.Exists(ctx, targetDir+"/.vimrc"))
	})

	t.Run("handles non-existent package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)

		svc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, targetDir, false)
		err := svc.Unmanage(ctx, "non-existent")
		require.NoError(t, err) // Should not error, just no-op
	})
}

func TestUnmanageService_PlanUnmanage(t *testing.T) {
	t.Run("creates delete operations for installed package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup and manage package first
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:        fs,
			IgnoreSet: ignore.NewDefaultIgnoreSet(),
			Policies:  planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, packageDir, targetDir, false)

		err := manageSvc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Plan unmanage
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, targetDir, false)
		plan, err := unmanageSvc.PlanUnmanage(ctx, "test-pkg")
		require.NoError(t, err)
		assert.Greater(t, len(plan.Operations), 0)
	})

	t.Run("returns empty plan when no manifest exists", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)

		svc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, targetDir, false)
		plan, err := svc.PlanUnmanage(ctx, "test-pkg")
		require.NoError(t, err)
		assert.Len(t, plan.Operations, 0)
	})
}
