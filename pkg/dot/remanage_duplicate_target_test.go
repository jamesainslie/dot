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
func remanageDuplicateClient(t *testing.T, packages map[string][]string) (*dot.Client, context.Context) {
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

	return client, ctx
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
			client, ctx := remanageDuplicateClient(t, tt.packages)

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
	client, ctx := remanageDuplicateClient(t, map[string][]string{
		"base":    {"dot-vimrc"},
		"overlay": {"dot-vimrc"},
	})

	err := client.Remanage(ctx, "base", "overlay")

	require.Error(t, err)
	var duplicate dot.ErrDuplicateTarget
	assert.ErrorAs(t, err, &duplicate)
}
