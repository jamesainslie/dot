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
	
	// Mock: package directory exists
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/vim").Return(true)
	
	result := scanner.ScanPackage(ctx, mockFS, packagePath, "vim", ignoreSet)
	require.True(t, result.IsOk())
	
	pkg := result.Unwrap()
	assert.Equal(t, "vim", pkg.Name)
	assert.Equal(t, packagePath, pkg.Path)
	
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
	
	// Mock: package exists
	mockFS.On("Exists", ctx, "/home/user/.dotfiles/vim").Return(true)
	
	result := scanner.ScanPackage(ctx, mockFS, packagePath, "vim", ignoreSet)
	require.True(t, result.IsOk())
	
	pkg := result.Unwrap()
	assert.Equal(t, "vim", pkg.Name)
	
	mockFS.AssertExpectations(t)
}

