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
	// Package with single file: vim/vimrc
	pkgPath := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := dot.NewTargetPath("/home/user").Unwrap()

	pkg := dot.Package{
		Name: "vim",
		Path: pkgPath,
	}

	result := planner.ComputeDesiredState([]dot.Package{pkg}, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()
	assert.NotNil(t, state)
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
