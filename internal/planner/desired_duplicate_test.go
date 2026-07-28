package planner_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/planner"
)

// buildTestPackage builds a package whose tree contains the given files,
// expressed as slash separated paths relative to the package root.
func buildTestPackage(t *testing.T, root, name string, files ...string) domain.Package {
	t.Helper()

	rootNode := domain.Node{
		Path: domain.NewFilePath(root).Unwrap(),
		Type: domain.NodeDir,
	}

	for _, rel := range files {
		insertTestFile(t, &rootNode, root, rel)
	}

	pkgPath := domain.NewPackagePath(root)
	require.True(t, pkgPath.IsOk(), "package path %s must be valid", root)

	return domain.Package{
		Name: name,
		Path: pkgPath.Unwrap(),
		Tree: &rootNode,
	}
}

// insertTestFile inserts rel into the tree rooted at parent, creating
// intermediate directory nodes as needed.
func insertTestFile(t *testing.T, parent *domain.Node, parentPath, rel string) {
	t.Helper()

	segments := strings.Split(rel, "/")
	current := parent
	currentPath := parentPath

	for i, segment := range segments {
		currentPath = filepath.Join(currentPath, segment)

		nodeType := domain.NodeDir
		if i == len(segments)-1 {
			nodeType = domain.NodeFile
		}

		child := findTestChild(current, currentPath)
		if child == nil {
			current.Children = append(current.Children, domain.Node{
				Path: domain.NewFilePath(currentPath).Unwrap(),
				Type: nodeType,
			})
			child = &current.Children[len(current.Children)-1]
		}

		current = child
	}
}

func findTestChild(parent *domain.Node, path string) *domain.Node {
	for i := range parent.Children {
		if parent.Children[i].Path.String() == path {
			return &parent.Children[i]
		}
	}
	return nil
}

func TestComputeDesiredState_DuplicateTargetAcrossPackages(t *testing.T) {
	target := domain.NewTargetPath("/home/user").Unwrap()

	tests := []struct {
		name          string
		packages      []domain.Package
		wantPath      string
		wantFirstPkg  string
		wantSecondPkg string
	}{
		{
			name: "two packages claim the same top level file",
			packages: []domain.Package{
				buildTestPackage(t, "/dotfiles/base", "base", "dot-vimrc"),
				buildTestPackage(t, "/dotfiles/overlay", "overlay", "dot-vimrc"),
			},
			wantPath:      "/home/user/.vimrc",
			wantFirstPkg:  "base",
			wantSecondPkg: "overlay",
		},
		{
			name: "two packages claim the same nested file",
			packages: []domain.Package{
				buildTestPackage(t, "/dotfiles/base", "base", "dot-config/nvim/init.lua"),
				buildTestPackage(t, "/dotfiles/overlay", "overlay", "dot-config/nvim/init.lua"),
			},
			wantPath:      "/home/user/.config/nvim/init.lua",
			wantFirstPkg:  "base",
			wantSecondPkg: "overlay",
		},
		{
			name: "collision only after dotfile translation",
			packages: []domain.Package{
				buildTestPackage(t, "/dotfiles/base", "base", "dot-gitconfig"),
				buildTestPackage(t, "/dotfiles/overlay", "overlay", ".gitconfig"),
			},
			wantPath:      "/home/user/.gitconfig",
			wantFirstPkg:  "base",
			wantSecondPkg: "overlay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.ComputeDesiredState(tt.packages, target, false)
			require.True(t, result.IsErr(), "expected duplicate target to be rejected")

			err := result.UnwrapErr()

			var dup domain.ErrDuplicateTarget
			require.True(t, errors.As(err, &dup), "expected ErrDuplicateTarget, got %T", err)
			assert.Equal(t, tt.wantPath, dup.TargetPath)
			assert.Equal(t, tt.wantFirstPkg, dup.FirstPackage)
			assert.Equal(t, tt.wantSecondPkg, dup.SecondPackage)

			msg := err.Error()
			assert.Contains(t, msg, tt.wantPath)
			assert.Contains(t, msg, tt.wantFirstPkg)
			assert.Contains(t, msg, tt.wantSecondPkg)
		})
	}
}

func TestComputeDesiredState_SamePackageClaimingTwiceIsAllowed(t *testing.T) {
	target := domain.NewTargetPath("/home/user").Unwrap()
	pkg := buildTestPackage(t, "/dotfiles/base", "base", "dot-vimrc")

	tests := []struct {
		name     string
		packages []domain.Package
	}{
		{
			name:     "package walked once",
			packages: []domain.Package{pkg},
		},
		{
			name:     "identical package walked twice",
			packages: []domain.Package{pkg, pkg},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.ComputeDesiredState(tt.packages, target, false)
			require.True(t, result.IsOk(), "expected no error")

			state := result.Unwrap()
			require.Len(t, state.Links, 1)

			link, ok := state.Links["/home/user/.vimrc"]
			require.True(t, ok)
			assert.Equal(t, "base", link.Package)
		})
	}
}

func TestComputeDesiredState_DistinctFilesInSharedDirectory(t *testing.T) {
	target := domain.NewTargetPath("/home/user").Unwrap()

	packages := []domain.Package{
		buildTestPackage(t, "/dotfiles/base", "base", "dot-config/shell/aliases"),
		buildTestPackage(t, "/dotfiles/overlay", "overlay", "dot-config/shell/functions"),
	}

	result := planner.ComputeDesiredState(packages, target, false)
	require.True(t, result.IsOk(), "packages sharing a directory must not collide")

	state := result.Unwrap()
	require.Len(t, state.Links, 2)

	aliases, ok := state.Links["/home/user/.config/shell/aliases"]
	require.True(t, ok)
	assert.Equal(t, "base", aliases.Package)

	functions, ok := state.Links["/home/user/.config/shell/functions"]
	require.True(t, ok)
	assert.Equal(t, "overlay", functions.Package)

	// The shared parent directories are recorded exactly once each.
	require.Len(t, state.Dirs, 2)
	assert.Contains(t, state.Dirs, "/home/user/.config")
	assert.Contains(t, state.Dirs, "/home/user/.config/shell")
	assert.Equal(t, "base", state.Dirs["/home/user/.config/shell"].Package)
}

func TestComputeDesiredState_DirectoryFileCollisionAcrossPackages(t *testing.T) {
	target := domain.NewTargetPath("/home/user").Unwrap()

	tests := []struct {
		name        string
		packages    []domain.Package
		wantPath    string
		wantFilePkg string
		wantDirPkg  string
	}{
		{
			name: "file claimed before directory",
			packages: []domain.Package{
				buildTestPackage(t, "/dotfiles/base", "base", "dot-config"),
				buildTestPackage(t, "/dotfiles/overlay", "overlay", "dot-config/nvim/init.lua"),
			},
			wantPath:    "/home/user/.config",
			wantFilePkg: "base",
			wantDirPkg:  "overlay",
		},
		{
			name: "directory claimed before file",
			packages: []domain.Package{
				buildTestPackage(t, "/dotfiles/overlay", "overlay", "dot-config/nvim/init.lua"),
				buildTestPackage(t, "/dotfiles/base", "base", "dot-config"),
			},
			wantPath:    "/home/user/.config",
			wantFilePkg: "base",
			wantDirPkg:  "overlay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := planner.ComputeDesiredState(tt.packages, target, false)
			require.True(t, result.IsErr(), "expected directory/file collision to be rejected")

			err := result.UnwrapErr()

			var collision domain.ErrTargetKindConflict
			require.True(t, errors.As(err, &collision), "expected ErrTargetKindConflict, got %T", err)
			assert.Equal(t, tt.wantPath, collision.TargetPath)
			assert.Equal(t, tt.wantFilePkg, collision.FilePackage)
			assert.Equal(t, tt.wantDirPkg, collision.DirPackage)

			msg := err.Error()
			assert.Contains(t, msg, tt.wantPath)
			assert.Contains(t, msg, tt.wantFilePkg)
			assert.Contains(t, msg, tt.wantDirPkg)
		})
	}
}

func TestComputeDesiredState_PackageNameMappingKeepsPackagesDisjoint(t *testing.T) {
	target := domain.NewTargetPath("/home/user").Unwrap()

	packages := []domain.Package{
		buildTestPackage(t, "/dotfiles/base", "base", "dot-vimrc"),
		buildTestPackage(t, "/dotfiles/overlay", "overlay", "dot-vimrc"),
	}

	result := planner.ComputeDesiredState(packages, target, true)
	require.True(t, result.IsOk(), "package name mapping gives each package its own subtree")

	state := result.Unwrap()
	require.Len(t, state.Links, 2)
	assert.Contains(t, state.Links, "/home/user/base/.vimrc")
	assert.Contains(t, state.Links, "/home/user/overlay/.vimrc")
}

func TestComputeDesiredState_RecordsOwningPackage(t *testing.T) {
	target := domain.NewTargetPath("/home/user").Unwrap()
	pkg := buildTestPackage(t, "/dotfiles/vim", "vim", "dot-vimrc", "dot-vim/colors/desert.vim")

	result := planner.ComputeDesiredState([]domain.Package{pkg}, target, false)
	require.True(t, result.IsOk())

	state := result.Unwrap()
	for path, link := range state.Links {
		assert.Equal(t, "vim", link.Package, "link %s must record its owning package", path)
	}
	for path, dir := range state.Dirs {
		assert.Equal(t, "vim", dir.Package, "dir %s must record its owning package", path)
	}
}
