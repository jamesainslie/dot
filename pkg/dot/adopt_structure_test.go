package dot_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdopt_DirectoryStructure(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Create directories
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.ssh", 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/config", []byte("ssh config"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/known_hosts", []byte("hosts"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt .ssh directory
	err = client.Adopt(ctx, []string{".ssh"}, "ssh")
	require.NoError(t, err)

	// Check where the directory was stored
	t.Log("Checking package structure after adopt...")

	// Debug: List what's in the package directory
	if fs.Exists(ctx, packageDir+"/ssh") {
		entries, _ := fs.ReadDir(ctx, packageDir+"/ssh")
		t.Logf("Package /ssh contains %d entries:", len(entries))
		for _, e := range entries {
			t.Logf("  - %s (isDir: %v)", e.Name(), e.IsDir())
		}
	}

	// Check different possible structures
	exists := fs.Exists(ctx, packageDir+"/ssh/dot-ssh")
	t.Logf("packageDir/ssh/dot-ssh exists: %v", exists)

	exists = fs.Exists(ctx, packageDir+"/ssh/.ssh")
	t.Logf("packageDir/ssh/.ssh exists: %v", exists)

	// The directory should be at packageDir/ssh/dot-ssh
	assert.True(t, fs.Exists(ctx, packageDir+"/ssh/dot-ssh"), "Package directory should exist at /ssh/dot-ssh")

	// Verify original was replaced with symlink
	isLink, _ := fs.IsSymlink(ctx, targetDir+"/.ssh")
	assert.True(t, isLink, "Target should be a symlink")
}
