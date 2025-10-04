package scanner_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanPackage(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	packagePath := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	ignoreSet := ignore.NewIgnoreSet()

	// Mock: package directory exists and is empty
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/vim").Return(true)
	mockFS.On("IsSymlink", ctx, "/home/user/.dotfiles/vim").Return(false, nil)
	mockFS.On("IsDir", ctx, "/home/user/.dotfiles/vim").Return(true, nil)
	mockFS.On("ReadDir", ctx, "/home/user/.dotfiles/vim").Return([]dot.DirEntry{}, nil)

	result := scanner.ScanPackage(ctx, mockFS, packagePath, "vim", ignoreSet)
	require.True(t, result.IsOk())

	pkg := result.Unwrap()
	assert.Equal(t, "vim", pkg.Name)
	assert.Equal(t, packagePath, pkg.Path)
	require.NotNil(t, pkg.Tree, "Tree should be populated")
	assert.Equal(t, dot.NodeDir, pkg.Tree.Type)

	mockFS.AssertExpectations(t)
}

func TestScanPackage_PackageNotFound(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	packagePath := dot.NewPackagePath("/home/user/.dotfiles/missing").Unwrap()
	ignoreSet := ignore.NewIgnoreSet()

	// Mock: package directory does not exist
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/missing").Return(false)

	result := scanner.ScanPackage(ctx, mockFS, packagePath, "missing", ignoreSet)
	assert.True(t, result.IsErr())

	// Should return ErrPackageNotFound
	err := result.UnwrapErr()
	_, ok := err.(dot.ErrPackageNotFound)
	assert.True(t, ok, "Expected ErrPackageNotFound")

	mockFS.AssertExpectations(t)
}

func TestScanPackage_WithIgnorePatterns(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	packagePath := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	ignoreSet := ignore.NewIgnoreSet()
	ignoreSet.Add(".git")

	// Mock: package exists and is a directory
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/vim").Return(true)
	mockFS.On("IsSymlink", ctx, "/home/user/.dotfiles/vim").Return(false, nil)
	mockFS.On("IsDir", ctx, "/home/user/.dotfiles/vim").Return(true, nil)
	mockFS.On("ReadDir", ctx, "/home/user/.dotfiles/vim").Return([]dot.DirEntry{}, nil)

	result := scanner.ScanPackage(ctx, mockFS, packagePath, "vim", ignoreSet)
	require.True(t, result.IsOk())

	pkg := result.Unwrap()
	assert.Equal(t, "vim", pkg.Name)
	require.NotNil(t, pkg.Tree, "Tree should be scanned")

	// Tree filtering is applied during scan
	// With empty directory, tree has no children (nothing to filter)
	assert.Empty(t, pkg.Tree.Children)

	mockFS.AssertExpectations(t)
}
