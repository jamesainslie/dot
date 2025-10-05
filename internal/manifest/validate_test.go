package manifest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate_EmptyManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	m := New()
	validator := NewValidator(fs)

	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_ValidManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")

	// Create source file
	vimrcSrc := "/stow/vim/dot-vimrc"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(vimrcSrc), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), vimrcSrc, []byte("content"), 0644))

	// Create target link
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
	require.NoError(t, fs.Symlink(context.Background(), vimrcSrc, vimrcTarget))

	// Create manifest
	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_BrokenLink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")

	// Create broken symlink
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
	require.NoError(t, fs.Symlink(context.Background(), "/nonexistent", vimrcTarget))

	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueBrokenLink, result.Issues[0].Type)
	assert.Equal(t, "vim", result.Issues[0].Package)
}

func TestValidator_Validate_MissingLink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueMissingLink, result.Issues[0].Type)
	assert.Contains(t, result.Issues[0].Path, ".vimrc")
}

func TestValidator_Validate_NotSymlink(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")

	// Create regular file where symlink should be
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), vimrcTarget, []byte("content"), 0644))

	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
	})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, IssueNotSymlink, result.Issues[0].Type)
}

func TestValidator_Validate_MultiplePackages(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create vim package and link
	vimrcSrc := "/stow/vim/dot-vimrc"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(vimrcSrc), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), vimrcSrc, []byte("vim"), 0644))
	vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
	require.NoError(t, fs.Symlink(context.Background(), vimrcSrc, vimrcTarget))

	// Create zsh package and link
	zshrcSrc := "/stow/zsh/dot-zshrc"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(zshrcSrc), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), zshrcSrc, []byte("zsh"), 0644))
	zshrcTarget := filepath.Join(targetDir.String(), ".zshrc")
	require.NoError(t, fs.Symlink(context.Background(), zshrcSrc, zshrcTarget))

	m := New()
	m.AddPackage(PackageInfo{Name: "vim", Links: []string{".vimrc"}})
	m.AddPackage(PackageInfo{Name: "zsh", Links: []string{".zshrc"}})

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.True(t, result.IsValid)
	assert.Empty(t, result.Issues)
}

func TestValidator_Validate_MultipleIssues(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create source file and valid link
	sourcePath := "/some/path"
	require.NoError(t, fs.MkdirAll(context.Background(), filepath.Dir(sourcePath), 0755))
	require.NoError(t, fs.WriteFile(context.Background(), sourcePath, []byte("content"), 0644))
	validTarget := filepath.Join(targetDir.String(), ".valid")
	require.NoError(t, fs.Symlink(context.Background(), sourcePath, validTarget))

	m := New()
	m.AddPackage(PackageInfo{
		Name:  "pkg",
		Links: []string{".valid", ".missing", ".broken"},
	})

	// Create broken link
	brokenTarget := filepath.Join(targetDir.String(), ".broken")
	require.NoError(t, fs.Symlink(context.Background(), "/nonexistent", brokenTarget))

	validator := NewValidator(fs)
	result := validator.Validate(context.Background(), targetDir, m)

	assert.False(t, result.IsValid)
	assert.GreaterOrEqual(t, len(result.Issues), 2) // At least missing and broken
}
