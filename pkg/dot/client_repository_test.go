package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestClient_RepositoryInfo(t *testing.T) {
	tests := []struct {
		name        string
		manifestDir string
		// storeDir is where the repository info is written before reading.
		storeDir    string
		recorded    bool
		wantExists  bool
		wantProfile string
	}{
		{
			name:        "reads repository info from the target directory",
			manifestDir: "",
			storeDir:    "",
			recorded:    true,
			wantExists:  true,
			wantProfile: "work",
		},
		{
			name:        "reads repository info from a custom manifest directory",
			manifestDir: "/state/dot",
			storeDir:    "/state/dot",
			recorded:    true,
			wantExists:  true,
			wantProfile: "work",
		},
		{
			name:        "reports no repository info when none is recorded",
			manifestDir: "/state/dot",
			storeDir:    "/state/dot",
			recorded:    false,
			wantExists:  false,
		},
		{
			name:        "ignores a manifest written outside the configured store",
			manifestDir: "/state/dot",
			storeDir:    "",
			recorded:    true,
			wantExists:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := adapters.NewMemFS()
			require.NoError(t, fs.MkdirAll(ctx, "/packages", 0755))
			require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

			store := manifest.NewFSManifestStore(fs)
			if tt.storeDir != "" {
				store = manifest.NewFSManifestStoreWithDir(fs, tt.storeDir)
			}

			targetPath := dot.NewTargetPath("/home")
			require.True(t, targetPath.IsOk())

			m := manifest.New()
			if tt.recorded {
				m.SetRepository(manifest.RepositoryInfo{
					URL:     "https://github.com/user/dotfiles",
					Branch:  "main",
					Profile: tt.wantProfile,
				})
			}
			require.NoError(t, store.Save(ctx, targetPath.Unwrap(), m))

			client, err := dot.NewClient(dot.Config{
				PackageDir:  "/packages",
				TargetDir:   "/home",
				ManifestDir: tt.manifestDir,
				FS:          fs,
				Logger:      adapters.NewNoopLogger(),
			})
			require.NoError(t, err)

			info, exists, err := client.RepositoryInfo(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
			if tt.wantExists {
				assert.Equal(t, tt.wantProfile, info.Profile)
				assert.Equal(t, "https://github.com/user/dotfiles", info.URL)
			}
		})
	}
}
