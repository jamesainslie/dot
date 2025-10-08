package planner_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDesiredState_EmptyPackage(t *testing.T) {
	packages := []domain.Package{}
	target := domain.NewTargetPath("/home/user").Unwrap()

	result := planner.ComputeDesiredState(packages, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()
	assert.Empty(t, state.Links)
	assert.Empty(t, state.Dirs)
}

func TestComputeDesiredState_SingleFile(t *testing.T) {
	// Package with single file: vim/vimrc -> ~/.vimrc
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := domain.NewTargetPath("/home/user").Unwrap()

	// Create a file node representing vim/vimrc
	fileNode := domain.Node{
		Path: domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap(),
		Type: domain.NodeFile,
	}

	pkg := domain.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &fileNode,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target)
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
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := domain.NewTargetPath("/home/user").Unwrap()

	fileNode := domain.Node{
		Path: domain.NewFilePath("/home/user/.dotfiles/vim/dot-vimrc").Unwrap(),
		Type: domain.NodeFile,
	}

	pkg := domain.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &fileNode,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// dot-vimrc should translate to .vimrc
	linkSpec, exists := state.Links["/home/user/.vimrc"]
	require.True(t, exists, "Expected link at /home/user/.vimrc (translated)")
	assert.Equal(t, "/home/user/.dotfiles/vim/dot-vimrc", linkSpec.Source.String())
}

func TestComputeDesiredState_NestedFiles(t *testing.T) {
	// Package with nested structure: vim/colors/desert.vim
	pkgPath := domain.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	target := domain.NewTargetPath("/home/user").Unwrap()

	// Build tree: vim/ -> colors/ -> desert.vim
	fileNode := domain.Node{
		Path: domain.NewFilePath("/home/user/.dotfiles/vim/colors/desert.vim").Unwrap(),
		Type: domain.NodeFile,
	}

	colorsDir := domain.Node{
		Path:     domain.NewFilePath("/home/user/.dotfiles/vim/colors").Unwrap(),
		Type:     domain.NodeDir,
		Children: []domain.Node{fileNode},
	}

	rootNode := domain.Node{
		Path:     domain.NewFilePath("/home/user/.dotfiles/vim").Unwrap(),
		Type:     domain.NodeDir,
		Children: []domain.Node{colorsDir},
	}

	pkg := domain.Package{
		Name: "vim",
		Path: pkgPath,
		Tree: &rootNode,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target)
	require.True(t, result.IsOk())

	state := result.Unwrap()

	// Should create: /home/user/colors/.vim/colors/desert.vim -> source
	// Or more likely: /home/user/.vim/colors/desert.vim -> source
	assert.NotEmpty(t, state.Links)

	// Should create parent directory: /home/user/.vim/colors
	assert.NotEmpty(t, state.Dirs)
}

func TestLinkSpec(t *testing.T) {
	source := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

	spec := planner.LinkSpec{
		Source: source,
		Target: target,
	}

	assert.Equal(t, source, spec.Source)
	assert.Equal(t, target, spec.Target)
}

func TestDirSpec(t *testing.T) {
	path := domain.NewFilePath("/home/user/.vim").Unwrap()

	spec := planner.DirSpec{
		Path: path,
	}

	assert.Equal(t, path, spec.Path)
}

func TestDesiredState(t *testing.T) {
	source := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	dirPath := domain.NewFilePath("/home/user/.vim").Unwrap()

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

// Task 7.5: Test Integration with Planner
func TestPlanResult(t *testing.T) {
	t.Run("without resolution", func(t *testing.T) {
		desired := planner.DesiredState{
			Links: make(map[string]planner.LinkSpec),
			Dirs:  make(map[string]planner.DirSpec),
		}

		result := planner.PlanResult{
			Desired: desired,
		}

		assert.NotNil(t, result.Desired)
		assert.Nil(t, result.Resolved)
		assert.False(t, result.HasConflicts())
	})

	t.Run("with resolution", func(t *testing.T) {
		desired := planner.DesiredState{
			Links: make(map[string]planner.LinkSpec),
			Dirs:  make(map[string]planner.DirSpec),
		}

		targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := planner.NewConflict(planner.ConflictFileExists, targetPath, "File exists")

		resolved := planner.NewResolveResult(nil).WithConflict(conflict)

		result := planner.PlanResult{
			Desired:  desired,
			Resolved: &resolved,
		}

		assert.NotNil(t, result.Resolved)
		assert.True(t, result.HasConflicts())
	})
}

func TestComputeOperationsFromDesiredState(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

	desired := planner.DesiredState{
		Links: map[string]planner.LinkSpec{
			targetPath.String(): {
				Source: sourcePath,
				Target: targetPath,
			},
		},
		Dirs: make(map[string]planner.DirSpec),
	}

	ops := planner.ComputeOperationsFromDesiredState(desired)

	assert.Len(t, ops, 1)
	linkOp, ok := ops[0].(domain.LinkCreate)
	assert.True(t, ok)
	assert.Equal(t, sourcePath, linkOp.Source)
	assert.Equal(t, targetPath, linkOp.Target)
}

func TestComputeOperationsFromDesiredStateWithDirs(t *testing.T) {
	dirPath := domain.NewFilePath("/home/user/.config").Unwrap()
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.config/bash").Unwrap()

	desired := planner.DesiredState{
		Links: map[string]planner.LinkSpec{
			targetPath.String(): {
				Source: sourcePath,
				Target: targetPath,
			},
		},
		Dirs: map[string]planner.DirSpec{
			dirPath.String(): {Path: dirPath},
		},
	}

	ops := planner.ComputeOperationsFromDesiredState(desired)

	assert.Len(t, ops, 2) // One dir + one link

	// Should have both operation types
	hasDirCreate := false
	hasLinkCreate := false
	for _, op := range ops {
		switch op.Kind() {
		case domain.OpKindDirCreate:
			hasDirCreate = true
		case domain.OpKindLinkCreate:
			hasLinkCreate = true
		}
	}
	assert.True(t, hasDirCreate)
	assert.True(t, hasLinkCreate)
}

func TestComputeDesiredStateWithMultipleFiles(t *testing.T) {
	targetDir := domain.NewTargetPath("/home/user").Unwrap()

	// Create package with multiple files
	pkgPath := domain.NewPackagePath("/packages/bash").Unwrap()
	pkgRoot := domain.NewFilePath("/packages/bash").Unwrap()
	file1Path := pkgPath.Join("dot-bashrc")
	file2Path := pkgPath.Join("dot-profile")
	file1 := domain.NewFilePath(file1Path.String()).Unwrap()
	file2 := domain.NewFilePath(file2Path.String()).Unwrap()

	tree := &domain.Node{
		Path: pkgRoot,
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: file1,
				Type: domain.NodeFile,
			},
			{
				Path: file2,
				Type: domain.NodeFile,
			},
		},
	}

	pkg := domain.Package{
		Name: "bash",
		Path: pkgPath,
		Tree: tree,
	}

	result := planner.ComputeDesiredState([]domain.Package{pkg}, targetDir)

	assert.True(t, result.IsOk())
	state := result.Unwrap()

	// Should have 2 links
	assert.Len(t, state.Links, 2)
}
