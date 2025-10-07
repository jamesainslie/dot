package pipeline

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUnderPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		basePath string
		want     bool
	}{
		{
			name:     "direct child",
			path:     "/packages/vim/dot-vimrc",
			basePath: "/packages/vim",
			want:     true,
		},
		{
			name:     "nested child",
			path:     "/packages/vim/colors/theme.vim",
			basePath: "/packages/vim",
			want:     true,
		},
		{
			name:     "sibling",
			path:     "/packages/zsh/dot-zshrc",
			basePath: "/packages/vim",
			want:     false,
		},
		{
			name:     "parent",
			path:     "/packages",
			basePath: "/packages/vim",
			want:     false,
		},
		{
			name:     "same path",
			path:     "/packages/vim",
			basePath: "/packages/vim",
			want:     false,
		},
		{
			name:     "prefix match but different",
			path:     "/packages/vimrc",
			basePath: "/packages/vim",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnderPath(tt.path, tt.basePath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOperationBelongsToPackage(t *testing.T) {
	vimPkgPath := "/packages/vim"

	tests := []struct {
		name    string
		op      dot.Operation
		pkgPath string
		want    bool
	}{
		{
			name: "LinkCreate from package",
			op: dot.NewLinkCreate(
				dot.OperationID("test-1"),
				mustFilePath("/packages/vim/dot-vimrc"),
				mustFilePath("/home/user/.vimrc"),
			),
			pkgPath: vimPkgPath,
			want:    true,
		},
		{
			name: "LinkCreate from different package",
			op: dot.NewLinkCreate(
				dot.OperationID("test-2"),
				mustFilePath("/packages/zsh/dot-zshrc"),
				mustFilePath("/home/user/.zshrc"),
			),
			pkgPath: vimPkgPath,
			want:    false,
		},
		{
			name: "DirCreate",
			op: dot.NewDirCreate(
				dot.OperationID("test-3"),
				mustFilePath("/home/user/.vim"),
			),
			pkgPath: vimPkgPath,
			want:    false,
		},
		{
			name: "LinkDelete",
			op: dot.NewLinkDelete(
				dot.OperationID("test-4"),
				mustFilePath("/home/user/.vimrc"),
			),
			pkgPath: vimPkgPath,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := operationBelongsToPackage(tt.op, tt.pkgPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildPackageOperationMapping(t *testing.T) {
	// Create test packages
	packages := []dot.Package{
		{
			Name: "vim",
			Path: mustPackagePath("/packages/vim"),
		},
		{
			Name: "zsh",
			Path: mustPackagePath("/packages/zsh"),
		},
	}

	// Create test operations
	ops := []dot.Operation{
		dot.NewLinkCreate(
			dot.OperationID("vim-link-1"),
			mustFilePath("/packages/vim/dot-vimrc"),
			mustFilePath("/home/user/.vimrc"),
		),
		dot.NewLinkCreate(
			dot.OperationID("vim-link-2"),
			mustFilePath("/packages/vim/dot-vim-colors"),
			mustFilePath("/home/user/.vim-colors"),
		),
		dot.NewLinkCreate(
			dot.OperationID("zsh-link-1"),
			mustFilePath("/packages/zsh/dot-zshrc"),
			mustFilePath("/home/user/.zshrc"),
		),
		dot.NewDirCreate(
			dot.OperationID("dir-1"),
			mustFilePath("/home/user/.config"),
		),
	}

	// Build mapping
	mapping := buildPackageOperationMapping(packages, ops)

	// Verify vim operations
	require.Contains(t, mapping, "vim")
	assert.Len(t, mapping["vim"], 2)
	assert.Contains(t, mapping["vim"], dot.OperationID("vim-link-1"))
	assert.Contains(t, mapping["vim"], dot.OperationID("vim-link-2"))

	// Verify zsh operations
	require.Contains(t, mapping, "zsh")
	assert.Len(t, mapping["zsh"], 1)
	assert.Contains(t, mapping["zsh"], dot.OperationID("zsh-link-1"))

	// Verify dir operation not assigned to any package
	for _, opIDs := range mapping {
		assert.NotContains(t, opIDs, dot.OperationID("dir-1"))
	}
}

func TestBuildPackageOperationMapping_EmptyPackages(t *testing.T) {
	ops := []dot.Operation{
		dot.NewLinkCreate(
			dot.OperationID("link-1"),
			mustFilePath("/packages/vim/dot-vimrc"),
			mustFilePath("/home/user/.vimrc"),
		),
	}

	mapping := buildPackageOperationMapping([]dot.Package{}, ops)
	assert.Len(t, mapping, 0)
}

func TestBuildPackageOperationMapping_EmptyOperations(t *testing.T) {
	packages := []dot.Package{
		{
			Name: "vim",
			Path: mustPackagePath("/packages/vim"),
		},
	}

	mapping := buildPackageOperationMapping(packages, []dot.Operation{})
	assert.Len(t, mapping, 0)
}

func TestBuildPackageOperationMapping_NoMatchingOperations(t *testing.T) {
	packages := []dot.Package{
		{
			Name: "vim",
			Path: mustPackagePath("/packages/vim"),
		},
	}

	ops := []dot.Operation{
		dot.NewDirCreate(
			dot.OperationID("dir-1"),
			mustFilePath("/home/user/.config"),
		),
	}

	mapping := buildPackageOperationMapping(packages, ops)
	assert.Len(t, mapping, 0, "should not create mapping for packages with no matching operations")
}

// Test helpers

func mustFilePath(path string) dot.FilePath {
	result := dot.NewFilePath(path)
	if !result.IsOk() {
		panic("invalid file path: " + path)
	}
	return result.Unwrap()
}

func mustPackagePath(path string) dot.PackagePath {
	result := dot.NewPackagePath(path)
	if !result.IsOk() {
		panic("invalid package path: " + path)
	}
	return result.Unwrap()
}
