package planner_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDesiredState_EmptyPackage(t *testing.T) {
	packages := []dot.Package{}
	target := dot.NewTargetPath("/home/user").Unwrap()

	result := planner.ComputeDesiredState(packages, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()
	assert.Empty(t, state.Links)
	assert.Empty(t, state.Dirs)
}

func TestComputeDesiredState_SingleFile(t *testing.T) {
	// Package with single file: vim/vimrc -> ~/.vimrc
	pkgPath := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := dot.NewTargetPath("/home/user").Unwrap()

	// Create a file node representing vim/vimrc
	fileNode := dot.Node{
		Path: dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap(),
		Type: dot.NodeFile,
	}

	pkg := dot.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &fileNode,
	}

	result := planner.ComputeDesiredState([]dot.Package{pkg}, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// File "vimrc" (no dot- prefix) should create /home/user/vimrc
	assert.Len(t, state.Links, 1)

	linkSpec, exists := state.Links["/home/user/vimrc"]
	require.True(t, exists, "Expected link at /home/user/vimrc")
	assert.Equal(t, "/home/user/.dotfiles/vim/vimrc", linkSpec.Source.String())
	assert.Equal(t, "/home/user/vimrc", linkSpec.Target.String())
}

func TestComputeDesiredState_DotfileTranslation(t *testing.T) {
	// Package with dot-vimrc -> should become .vimrc in target
	pkgPath := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := dot.NewTargetPath("/home/user").Unwrap()

	fileNode := dot.Node{
		Path: dot.NewFilePath("/home/user/.dotfiles/vim/dot-vimrc").Unwrap(),
		Type: dot.NodeFile,
	}

	pkg := dot.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &fileNode,
	}

	result := planner.ComputeDesiredState([]dot.Package{pkg}, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// dot-vimrc should translate to .vimrc
	linkSpec, exists := state.Links["/home/user/.vimrc"]
	require.True(t, exists, "Expected link at /home/user/.vimrc (translated)")
	assert.Equal(t, "/home/user/.dotfiles/vim/dot-vimrc", linkSpec.Source.String())
}

func TestComputeDesiredState_NestedFiles(t *testing.T) {
	// Package with nested structure: vim/colors/desert.vim
	pkgPath := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := dot.NewTargetPath("/home/user").Unwrap()

	// Build tree: vim/ -> colors/ -> desert.vim
	fileNode := dot.Node{
		Path: dot.NewFilePath("/home/user/.dotfiles/vim/colors/desert.vim").Unwrap(),
		Type: dot.NodeFile,
	}

	colorsDir := dot.Node{
		Path:     dot.NewFilePath("/home/user/.dotfiles/vim/colors").Unwrap(),
		Type:     dot.NodeDir,
		Children: []dot.Node{fileNode},
	}

	rootNode := dot.Node{
		Path:     dot.NewFilePath("/home/user/.dotfiles/vim").Unwrap(),
		Type:     dot.NodeDir,
		Children: []dot.Node{colorsDir},
	}

	pkg := dot.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &rootNode,
	}

	result := planner.ComputeDesiredState([]dot.Package{pkg}, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// Should create: /home/user/colors/.vim/colors/desert.vim -> source
	// Or more likely: /home/user/.vim/colors/desert.vim -> source
	assert.NotEmpty(t, state.Links)

	// Should create parent directory: /home/user/.vim/colors
	assert.NotEmpty(t, state.Dirs)
}

func TestLinkSpec(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewFilePath("/home/user/.vimrc").Unwrap()

	spec := planner.LinkSpec{
		Source: source,
		Target: target,
	}

	assert.Equal(t, source, spec.Source)
	assert.Equal(t, target, spec.Target)
}

func TestDirSpec(t *testing.T) {
	path := dot.NewFilePath("/home/user/.vim").Unwrap()

	spec := planner.DirSpec{
		Path: path,
	}

	assert.Equal(t, path, spec.Path)
}

func TestDesiredState(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	dirPath := dot.NewFilePath("/home/user/.vim").Unwrap()

	state := planner.DesiredState{
		Links: map[string]planner.LinkSpec{
			target.String(): {Source: source, Target: target},
		},
		Dirs: map[string]planner.DirSpec{
			dirPath.String(): {Path: dirPath},
		},
	}

	assert.Len(t, state.Links, 1)
	assert.Len(t, state.Dirs, 1)
	assert.Contains(t, state.Links, target.String())
	assert.Contains(t, state.Dirs, dirPath.String())
}
