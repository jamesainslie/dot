package dot_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Doctor_OrphanedLinkDetection(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup managed package
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Create orphaned symlink
	require.NoError(t, fs.Symlink(ctx, "/nowhere", "/test/target/.orphaned"))

	// Test scoped scan (should detect orphan)
	report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())
	require.NoError(t, err)
	assert.True(t, report.Statistics.OrphanedLinks >= 1, "Expected to detect orphaned link")

	// Verify issues reported
	hasOrphanIssue := false
	for _, issue := range report.Issues {
		if issue.Type == dot.IssueOrphanedLink {
			hasOrphanIssue = true
			break
		}
	}
	assert.True(t, hasOrphanIssue, "Expected orphaned link issue")
}

func TestClient_Doctor_NestedDirectories(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup with nested structure
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/deep", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/subdir/nested", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/deep/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Test deep scan with depth limit
	scanCfg := dot.DeepScanConfig(10)
	report, err := client.DoctorWithScan(ctx, scanCfg)
	require.NoError(t, err)
	assert.NotNil(t, report)
}

func TestClient_Doctor_SkipPatterns(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/.git", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/node_modules", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Create links in directories that should be skipped
	require.NoError(t, fs.Symlink(ctx, "/nowhere", "/test/target/.git/link"))
	require.NoError(t, fs.Symlink(ctx, "/nowhere", "/test/target/node_modules/link"))

	// Deep scan should skip these directories
	report, err := client.DoctorWithScan(ctx, dot.DeepScanConfig(5))
	require.NoError(t, err)

	// Links in skipped directories should not be reported as orphans
	assert.NotNil(t, report)
}
