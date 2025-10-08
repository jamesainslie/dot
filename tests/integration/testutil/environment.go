package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestEnvironment provides an isolated environment for integration tests.
type TestEnvironment struct {
	t          testing.TB
	tmpDir     string
	PackageDir string
	TargetDir  string
	ctx        context.Context
	cancel     context.CancelFunc
	cleanupFns []func()
}

// NewTestEnvironment creates a new isolated test environment.
func NewTestEnvironment(t testing.TB) *TestEnvironment {
	t.Helper()

	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	env := &TestEnvironment{
		t:          t,
		tmpDir:     tmpDir,
		PackageDir: packageDir,
		TargetDir:  targetDir,
		ctx:        ctx,
		cancel:     cancel,
		cleanupFns: make([]func(), 0),
	}

	t.Cleanup(func() {
		env.Cleanup()
	})

	return env
}

// Context returns the test context.
func (te *TestEnvironment) Context() context.Context {
	return te.ctx
}

// AddCleanup registers a cleanup function to be called during cleanup.
func (te *TestEnvironment) AddCleanup(fn func()) {
	te.t.Helper()
	te.cleanupFns = append(te.cleanupFns, fn)
}

// Cleanup performs cleanup operations.
func (te *TestEnvironment) Cleanup() {
	te.t.Helper()

	// Run cleanup functions in reverse order
	for i := len(te.cleanupFns) - 1; i >= 0; i-- {
		te.cleanupFns[i]()
	}

	// Cancel context
	te.cancel()
}

// FixtureBuilder returns a fixture builder for this environment.
func (te *TestEnvironment) FixtureBuilder() *FixtureBuilder {
	te.t.Helper()
	return NewFixtureBuilder(te.t, te.PackageDir)
}

// ManifestPath returns the path where manifest would be stored.
func (te *TestEnvironment) ManifestPath() string {
	return filepath.Join(te.tmpDir, "manifest.json")
}

// CreatePackage creates a simple test package with files.
func (te *TestEnvironment) CreatePackage(name string, files map[string]string) string {
	te.t.Helper()
	packagePath := filepath.Join(te.PackageDir, name)
	require.NoError(te.t, os.MkdirAll(packagePath, 0755))

	for path, content := range files {
		fullPath := filepath.Join(packagePath, path)
		dirPath := filepath.Dir(fullPath)
		require.NoError(te.t, os.MkdirAll(dirPath, 0755))
		require.NoError(te.t, os.WriteFile(fullPath, []byte(content), 0644)) //nolint:gosec // Test fixtures
	}

	return packagePath
}

// CreateSimplePackage creates a package with a single dotfile.
func (te *TestEnvironment) CreateSimplePackage(name, filename, content string) string {
	te.t.Helper()
	return te.CreatePackage(name, map[string]string{
		filename: content,
	})
}

// WithTimeout creates a new environment with custom timeout.
func WithTimeout(t testing.TB, timeout time.Duration) *TestEnvironment {
	t.Helper()

	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	env := &TestEnvironment{
		t:          t,
		tmpDir:     tmpDir,
		PackageDir: packageDir,
		TargetDir:  targetDir,
		ctx:        ctx,
		cancel:     cancel,
		cleanupFns: make([]func(), 0),
	}

	t.Cleanup(func() {
		env.Cleanup()
	})

	return env
}
