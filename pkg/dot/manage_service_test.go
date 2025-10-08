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

func TestManageService_Manage(t *testing.T) {
	t.Run("manages package successfully", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		// Create dependencies
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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, packageDir, targetDir, false)

		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link created
		linkExists := fs.Exists(ctx, targetDir+"/.vimrc")
		assert.True(t, linkExists)
	})

	t.Run("dry run does not execute", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, packageDir, targetDir, true)

		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link NOT created (dry run)
		linkExists := fs.Exists(ctx, targetDir+"/.vimrc")
		assert.False(t, linkExists)
	})
}

func TestManageService_PlanManage(t *testing.T) {
	t.Run("creates execution plan", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:        fs,
			IgnoreSet: ignore.NewDefaultIgnoreSet(),
			Policies:  planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
		})
		exec := executor.New(executor.Opts{FS: fs, Logger: adapters.NewNoopLogger()})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, packageDir, targetDir, false)

		plan, err := svc.PlanManage(ctx, "test-pkg")
		require.NoError(t, err)
		assert.Greater(t, len(plan.Operations), 0)
	})
}

func TestManageService_Remanage(t *testing.T) {
	t.Run("skips unchanged packages", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, packageDir, targetDir, false)

		// Initial manage
		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Remanage without changes
		err = svc.Remanage(ctx, "test-pkg")
		require.NoError(t, err)
	})
}
