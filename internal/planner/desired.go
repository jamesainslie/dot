// Package planner provides pure planning logic for computing operations.
package planner

import (
	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// LinkSpec specifies a desired symbolic link.
type LinkSpec struct {
	Source dot.FilePath // Source file in package
	Target dot.FilePath // Target location
}

// DirSpec specifies a desired directory.
type DirSpec struct {
	Path dot.FilePath
}

// DesiredState represents the desired filesystem state.
type DesiredState struct {
	Links map[string]LinkSpec // Key: target path
	Dirs  map[string]DirSpec  // Key: directory path
}

// PlanResult contains planning results with optional conflict resolution
type PlanResult struct {
	Desired  DesiredState
	Resolved *ResolveResult // Optional resolution results
}

// HasConflicts returns true if there are unresolved conflicts
func (pr PlanResult) HasConflicts() bool {
	return pr.Resolved != nil && pr.Resolved.HasConflicts()
}

// ComputeDesiredState computes desired state from packages.
// This is a pure function that determines what links and directories
// should exist based on the package contents.
//
// For each file in each package:
// 1. Compute relative path from package root
// 2. Apply dotfile translation (dot-vimrc -> .vimrc)
// 3. Join with target to get target path
// 4. Create LinkSpec (source -> target)
// 5. Create DirSpec for parent directories
func ComputeDesiredState(packages []dot.Package, target dot.TargetPath) dot.Result[DesiredState] {
	state := DesiredState{
		Links: make(map[string]LinkSpec),
		Dirs:  make(map[string]DirSpec),
	}

	for _, pkg := range packages {
		// Skip packages without trees
		if pkg.Tree == nil {
			continue
		}

		// Process all files in the package tree
		if err := processPackageTree(pkg, target, &state); err != nil {
			return dot.Err[DesiredState](err)
		}
	}

	return dot.Ok(state)
}

// processPackageTree walks a package tree and adds link/dir specs to state.
func processPackageTree(pkg dot.Package, target dot.TargetPath, state *DesiredState) error {
	return walkPackageFiles(*pkg.Tree, pkg.Path, target, state)
}

// walkPackageFiles recursively processes files in a package tree.
func walkPackageFiles(node dot.Node, pkgRoot dot.PackagePath, target dot.TargetPath, state *DesiredState) error {
	// Process files only (not directories or symlinks)
	if node.Type == dot.NodeFile {
		// Compute relative path from package root
		relPathResult := relativePath(pkgRoot, node.Path)
		if relPathResult.IsErr() {
			return relPathResult.UnwrapErr()
		}
		relPath := relPathResult.Unwrap()

		// Apply dotfile translation to the relative path
		translated := translatePath(relPath)

		// Compute target path
		targetPath := target.Join(translated)

		// Add link spec
		state.Links[targetPath.String()] = LinkSpec{
			Source: node.Path,
			Target: targetPath,
		}

		// Add parent directory specs
		if err := addParentDirs(targetPath, target, state); err != nil {
			return err
		}
	}

	// Recurse on children
	for _, child := range node.Children {
		if err := walkPackageFiles(child, pkgRoot, target, state); err != nil {
			return err
		}
	}

	return nil
}

// addParentDirs adds directory specs for all parent directories of path.
func addParentDirs(path dot.FilePath, target dot.TargetPath, state *DesiredState) error {
	current := path
	targetStr := target.String()

	for {
		parentResult := current.Parent()
		if parentResult.IsErr() {
			break
		}

		parent := parentResult.Unwrap()
		parentStr := parent.String()

		// Stop when we reach the target directory
		if parentStr == targetStr {
			break
		}

		// Add directory spec if not already present
		if _, exists := state.Dirs[parentStr]; !exists {
			state.Dirs[parentStr] = DirSpec{Path: parent}
		}

		current = parent
	}

	return nil
}

// Helper functions that will be moved to scanner package

func relativePath(base dot.PackagePath, target dot.FilePath) dot.Result[string] {
	// Simple relative path computation
	baseStr := base.String()
	targetStr := target.String()

	// If target doesn't start with base, error
	if len(targetStr) <= len(baseStr) {
		return dot.Err[string](dot.ErrInvalidPath{Path: targetStr, Reason: "not under base"})
	}

	// Strip base path and leading slash
	rel := targetStr[len(baseStr):]
	if len(rel) > 0 && rel[0] == '/' {
		rel = rel[1:]
	}

	return dot.Ok(rel)
}

func translatePath(path string) string {
	return scanner.TranslatePath(path)
}

// ComputeOperationsFromDesiredState converts desired state into operations
func ComputeOperationsFromDesiredState(desired DesiredState) []dot.Operation {
	// Preallocate slice for directories and links
	ops := make([]dot.Operation, 0, len(desired.Dirs)+len(desired.Links))

	// Create directory operations
	for _, dirSpec := range desired.Dirs {
		ops = append(ops, dot.NewDirCreate(dirSpec.Path))
	}

	// Create link operations
	for _, linkSpec := range desired.Links {
		ops = append(ops, dot.NewLinkCreate(linkSpec.Source, linkSpec.Target))
	}

	return ops
}
