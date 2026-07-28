package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// remanageDuplicateClient builds an in-memory client whose packages each
// contain the given relative files. Nothing outside the MemFS is touched.
func remanageDuplicateClient(t *testing.T, packages map[string][]string) (*dot.Client, *adapters.MemFS, context.Context) {
	t.Helper()

	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	for pkg, files := range packages {
		require.NoError(t, fs.MkdirAll(ctx, "/test/packages/"+pkg, 0755))
		for _, file := range files {
			require.NoError(t, fs.WriteFile(ctx, "/test/packages/"+pkg+"/"+file, []byte(pkg), 0644))
		}
	}

	client, err := dot.NewClient(dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	})
	require.NoError(t, err)

	return client, fs, ctx
}

// TestPlanRemanageRejectsDuplicateTargets covers the remanage path, which plans
// each package on its own rather than through one shared desired state. The
// per-package plans must still be checked against each other.
func TestPlanRemanageRejectsDuplicateTargets(t *testing.T) {
	tests := []struct {
		name         string
		packages     map[string][]string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name: "two packages claiming the same file collide",
			packages: map[string][]string{
				"base":    {"dot-vimrc"},
				"overlay": {"dot-vimrc"},
			},
			args:         []string{"base", "overlay"},
			wantErr:      true,
			wantContains: []string{"/test/target/.vimrc", "base", "overlay"},
		},
		{
			name: "distinct files in the same directory do not collide",
			packages: map[string][]string{
				"base":    {"dot-vimrc"},
				"overlay": {"dot-zshrc"},
			},
			args:    []string{"base", "overlay"},
			wantErr: false,
		},
		{
			name: "a single package is unaffected",
			packages: map[string][]string{
				"base": {"dot-vimrc"},
			},
			args:    []string{"base"},
			wantErr: false,
		},
		{
			name: "the same package named twice is not a collision",
			packages: map[string][]string{
				"base": {"dot-vimrc"},
			},
			args:    []string{"base", "base"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _, ctx := remanageDuplicateClient(t, tt.packages)

			_, err := client.PlanRemanage(ctx, tt.args...)

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			var duplicate dot.ErrDuplicateTarget
			require.ErrorAs(t, err, &duplicate)
			for _, want := range tt.wantContains {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}

// TestRemanageRejectsDuplicateTargets checks the executing path, not just the
// planning path, so no link is created for a colliding pair.
func TestRemanageRejectsDuplicateTargets(t *testing.T) {
	client, _, ctx := remanageDuplicateClient(t, map[string][]string{
		"base":    {"dot-vimrc"},
		"overlay": {"dot-vimrc"},
	})

	err := client.Remanage(ctx, "base", "overlay")

	require.Error(t, err)
	var duplicate dot.ErrDuplicateTarget
	assert.ErrorAs(t, err, &duplicate)
}

// TestRemanageRejectsCollisionWithManagedPackage covers the first-time
// collision against a package that is already installed: vim owns .rc on disk
// and in the manifest, then emacs gains a colliding dot-rc. The per-package
// remanage plan for emacs resolves the collision as a wrong-link conflict and
// omits the operation, so the claim guard over planned operations never sees
// it. The conflict must surface as an error naming both packages, not vanish
// into an exit-0 remanage that records emacs with zero links.
func TestRemanageRejectsCollisionWithManagedPackage(t *testing.T) {
	client, _, ctx := remanageDuplicateClient(t, map[string][]string{
		"vim":   {"dot-rc"},
		"emacs": {"dot-erc", "dot-rc"},
	})

	require.NoError(t, client.Manage(ctx, "vim"))

	err := client.Remanage(ctx, "vim", "emacs")

	require.Error(t, err)
	var duplicate dot.ErrDuplicateTarget
	require.ErrorAs(t, err, &duplicate)
	assert.Contains(t, err.Error(), "/test/target/.rc")
	assert.Contains(t, err.Error(), "vim")
	assert.Contains(t, err.Error(), "emacs")

	// The failed remanage must not have half-registered emacs.
	status, statusErr := client.Status(ctx, "emacs")
	if statusErr == nil {
		for _, pkg := range status.Packages {
			if pkg.Name == "emacs" {
				assert.NotEmpty(t, pkg.Links,
					"emacs must not be recorded with zero links after a rejected remanage")
			}
		}
	}
}

// TestRemanageForeignConflictStillErrors ensures the conflict fallback holds
// when the blocking target is not owned by any package: a plain file the user
// created. The error is the generic conflict error, but it must be an error.
func TestRemanageForeignConflictStillErrors(t *testing.T) {
	client, fs, ctx := remanageDuplicateClient(t, map[string][]string{
		"vim": {"dot-rc"},
	})

	require.NoError(t, fs.WriteFile(ctx, "/test/target/.rc", []byte("user file"), 0644))

	err := client.Remanage(ctx, "vim")

	require.Error(t, err)
}
