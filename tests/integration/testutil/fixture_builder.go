package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// FixtureBuilder provides methods for building test fixtures.
type FixtureBuilder struct {
	t       testing.TB
	baseDir string
}

// NewFixtureBuilder creates a new fixture builder.
func NewFixtureBuilder(t testing.TB, baseDir string) *FixtureBuilder {
	t.Helper()
	return &FixtureBuilder{
		t:       t,
		baseDir: baseDir,
	}
}

// PackageBuilder builds test packages.
type PackageBuilder struct {
	fb          *FixtureBuilder
	packageName string
	files       map[string]string
	dirs        []string
}

// Package starts building a new package.
func (fb *FixtureBuilder) Package(name string) *PackageBuilder {
	fb.t.Helper()
	return &PackageBuilder{
		fb:          fb,
		packageName: name,
		files:       make(map[string]string),
		dirs:        make([]string, 0),
	}
}

// WithFile adds a file to the package with given content.
func (pb *PackageBuilder) WithFile(path, content string) *PackageBuilder {
	pb.fb.t.Helper()
	pb.files[path] = content
	return pb
}

// WithDir adds an empty directory to the package.
func (pb *PackageBuilder) WithDir(path string) *PackageBuilder {
	pb.fb.t.Helper()
	pb.dirs = append(pb.dirs, path)
	return pb
}

// Create creates the package on the filesystem.
func (pb *PackageBuilder) Create() string {
	pb.fb.t.Helper()
	packagePath := filepath.Join(pb.fb.baseDir, pb.packageName)
	require.NoError(pb.fb.t, os.MkdirAll(packagePath, 0755))

	// Create directories
	for _, dir := range pb.dirs {
		dirPath := filepath.Join(packagePath, dir)
		require.NoError(pb.fb.t, os.MkdirAll(dirPath, 0755))
	}

	// Create files
	for path, content := range pb.files {
		fullPath := filepath.Join(packagePath, path)
		dirPath := filepath.Dir(fullPath)
		require.NoError(pb.fb.t, os.MkdirAll(dirPath, 0755))
		require.NoError(pb.fb.t, os.WriteFile(fullPath, []byte(content), 0644)) //nolint:gosec // Test fixtures
	}

	return packagePath
}

// FileTreeBuilder builds arbitrary directory trees.
type FileTreeBuilder struct {
	fb   *FixtureBuilder
	base string
}

// FileTree starts building a file tree at the given base path.
func (fb *FixtureBuilder) FileTree(base string) *FileTreeBuilder {
	fb.t.Helper()
	return &FileTreeBuilder{
		fb:   fb,
		base: base,
	}
}

// File creates a file with given content.
func (ftb *FileTreeBuilder) File(path, content string) *FileTreeBuilder {
	ftb.fb.t.Helper()
	fullPath := filepath.Join(ftb.base, path)
	dirPath := filepath.Dir(fullPath)
	require.NoError(ftb.fb.t, os.MkdirAll(dirPath, 0755))
	require.NoError(ftb.fb.t, os.WriteFile(fullPath, []byte(content), 0644)) //nolint:gosec // Test fixtures
	return ftb
}

// Dir creates a directory.
func (ftb *FileTreeBuilder) Dir(path string) *FileTreeBuilder {
	ftb.fb.t.Helper()
	fullPath := filepath.Join(ftb.base, path)
	require.NoError(ftb.fb.t, os.MkdirAll(fullPath, 0755))
	return ftb
}

// Symlink creates a symlink.
func (ftb *FileTreeBuilder) Symlink(oldname, newname string) *FileTreeBuilder {
	ftb.fb.t.Helper()
	newPath := filepath.Join(ftb.base, newname)
	dirPath := filepath.Dir(newPath)
	require.NoError(ftb.fb.t, os.MkdirAll(dirPath, 0755))
	require.NoError(ftb.fb.t, os.Symlink(oldname, newPath))
	return ftb
}

// FileWithMode creates a file with specific permissions.
func (ftb *FileTreeBuilder) FileWithMode(path, content string, mode os.FileMode) *FileTreeBuilder {
	ftb.fb.t.Helper()
	fullPath := filepath.Join(ftb.base, path)
	dirPath := filepath.Dir(fullPath)
	require.NoError(ftb.fb.t, os.MkdirAll(dirPath, 0755))
	require.NoError(ftb.fb.t, os.WriteFile(fullPath, []byte(content), mode))
	return ftb
}
