package dot

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/pipeline"
	"github.com/yaklabco/dot/internal/planner"
)

// newRepoConfigCloneService builds a clone service whose pre-clone config has
// package name mapping ENABLED (the shipped default), with a rebuild hook
// mirroring the one the client wires, so a cloned repo carrying
// .config/dot/config.yaml can override the dotfile options for the install.
func newRepoConfigCloneService(t *testing.T, fs FS, repoFiles map[string]string) *CloneService {
	t.Helper()

	logger := adapters.NewNoopLogger()
	exec := executor.New(executor.Opts{FS: fs, Logger: logger, Tracer: adapters.NewNoopTracer()})
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, store)
	unmanageSvc := newUnmanageService(fs, logger, exec, manifestSvc, "/packages", "/home", false)

	buildManage := func(mapping bool, translate *bool) *ManageService {
		pipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicySkip},
			PackageNameMapping: mapping,
			Translate:          translate,
		})
		return newManageService(fs, logger, pipe, exec, manifestSvc, unmanageSvc, "/packages", "/home", false)
	}

	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			for path, content := range repoFiles {
				full := dest + "/" + path
				if err := fs.MkdirAll(ctx, filepath.Dir(full), 0755); err != nil {
					return err
				}
				if err := fs.WriteFile(ctx, full, []byte(content), 0644); err != nil {
					return err
				}
			}
			return nil
		},
	}

	// Pre-clone configuration: mapping enabled (the default on a machine with
	// no user config).
	svc := newCloneService(fs, logger, buildManage(true, nil), cloner, &mockPackageSelector{}, store, "/packages", "/home", false)
	svc.rebuildManageSvc = func(mapping *bool, translate *bool) *ManageService {
		m := true
		if mapping != nil {
			m = *mapping
		}
		return buildManage(m, translate)
	}
	return svc
}

// TestCloneAppliesRepoDotfileConfig is the zero-touch bootstrap contract: a
// repo that commits .config/dot/config.yaml with package_name_mapping: false
// must install stow-style full-tree links in the SAME clone invocation, on a
// machine with no pre-existing user config.
func TestCloneAppliesRepoDotfileConfig(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	svc := newRepoConfigCloneService(t, fs, map[string]string{
		".config/dot/config.yaml": "dotfile:\n  package_name_mapping: false\n",
		"mypkg/dot-rc":            "content",
	})

	require.NoError(t, svc.Clone(ctx, "https://example.com/dots.git", CloneOptions{}))

	assert.True(t, fs.Exists(ctx, "/home/.rc"),
		"full-tree layout expected: /home/.rc must exist")
	assert.False(t, fs.Exists(ctx, "/home/mypkg"),
		"package-name-prefixed layout must not appear when the repo config disables mapping")
}

// TestCloneWithoutRepoConfigKeepsDefaults: no repo config means the pre-clone
// configuration governs, unchanged behavior.
func TestCloneWithoutRepoConfigKeepsDefaults(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	svc := newRepoConfigCloneService(t, fs, map[string]string{
		"mypkg/dot-rc": "content",
	})

	require.NoError(t, svc.Clone(ctx, "https://example.com/dots.git", CloneOptions{}))

	assert.True(t, fs.Exists(ctx, "/home/mypkg/.rc"),
		"mapping enabled pre-clone config must still apply without a repo config")
}

// TestCloneRejectsBrokenRepoConfig: a present but unparseable repo config is
// a loud error, matching the CLI loader's behavior, never a silent fallback.
func TestCloneRejectsBrokenRepoConfig(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	svc := newRepoConfigCloneService(t, fs, map[string]string{
		".config/dot/config.yaml": "dotfile: [broken",
		"mypkg/dot-rc":            "content",
	})

	err := svc.Clone(ctx, "https://example.com/dots.git", CloneOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository config")
}
