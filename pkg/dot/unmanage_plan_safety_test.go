package dot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/pipeline"
	"github.com/yaklabco/dot/internal/planner"
)

// TestPlanUnmanage_RecordsLinkDestination verifies that LinkDelete operations
// planned for unmanage record the destination the symlink pointed at when the
// plan was computed, so execution can verify it and rollback can restore it.
func TestPlanUnmanage_RecordsLinkDestination(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()
	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

	managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
		FS:                 fs,
		IgnoreSet:          ignore.NewDefaultIgnoreSet(),
		Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
		PackageNameMapping: false,
	})
	exec := executor.New(executor.Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
	unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
	manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

	require.NoError(t, manageSvc.Manage(ctx, "test-pkg"))
	require.True(t, fs.Exists(ctx, targetDir+"/.vimrc"))

	expectedDest, err := fs.ReadLink(ctx, targetDir+"/.vimrc")
	require.NoError(t, err)

	plan, err := unmanageSvc.PlanUnmanage(ctx, "test-pkg")
	require.NoError(t, err)

	var linkDeletes []domain.LinkDelete
	for _, op := range plan.Operations {
		if del, ok := op.(domain.LinkDelete); ok {
			linkDeletes = append(linkDeletes, del)
		}
	}
	require.NotEmpty(t, linkDeletes, "unmanage plan must contain LinkDelete operations")

	for _, del := range linkDeletes {
		assert.Equal(t, expectedDest, del.Destination,
			"LinkDelete must record the symlink destination observed at plan time")
	}
}

// TestUnmanage_RefusesToDeleteRegularFileAtLinkPath verifies the end-to-end
// safety behavior: if a manifest-listed link path holds a regular file, the
// unmanage execution refuses to delete it.
func TestUnmanage_RefusesToDeleteRegularFileAtLinkPath(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()
	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

	managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
		FS:                 fs,
		IgnoreSet:          ignore.NewDefaultIgnoreSet(),
		Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
		PackageNameMapping: false,
	})
	exec := executor.New(executor.Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
	unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
	manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

	require.NoError(t, manageSvc.Manage(ctx, "test-pkg"))
	require.True(t, fs.Exists(ctx, targetDir+"/.vimrc"))

	// A user replaces the symlink with a regular file after managing.
	require.NoError(t, fs.Remove(ctx, targetDir+"/.vimrc"))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("hand-edited config"), 0644))

	err := unmanageSvc.Unmanage(ctx, "test-pkg")
	require.Error(t, err, "unmanage must refuse to delete a regular file at the link path")

	data, readErr := fs.ReadFile(ctx, targetDir+"/.vimrc")
	require.NoError(t, readErr)
	assert.Equal(t, []byte("hand-edited config"), data, "the user's file must be untouched")
}
