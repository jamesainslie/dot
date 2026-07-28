// Package planner provides pure planning logic for computing operations.
package planner

import (
	"fmt"
	"path/filepath"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/scanner"
)

// LinkSpec specifies a desired symbolic link.
type LinkSpec struct {
	Source  domain.FilePath   // Source file in package
	Target  domain.TargetPath // Target location
	Package string            // Name of the package that claimed this target
}

// DirSpec specifies a desired directory.
type DirSpec struct {
	Path    domain.FilePath
	Package string // Name of the package that first required this directory
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
//  1. Compute relative path from package root
//  2. Apply dotfile translation (dot-vimrc -> .vimrc)
//  3. If packageNameMapping is enabled, prepend the translated package name;
//     otherwise use the stow-style full-tree layout, where the package tree
//     maps straight onto the target
//  4. Join with target to get target path
//  5. Create LinkSpec (source -> target)
//  6. Create DirSpec for parent directories
//
// Every target path is owned by exactly one package. If a second, different
// package claims a path that is already claimed, planning fails with
// domain.ErrDuplicateTarget (same path, both as links) or
// domain.ErrTargetKindConflict (one package links a file where another needs a
// directory) rather than letting the last package walked silently win.
func ComputeDesiredState(packages []domain.Package, target domain.TargetPath, packageNameMapping bool, translate ...bool) domain.Result[DesiredState] {
	// Default translate to true for backward compatibility
	doTranslate := true
	if len(translate) > 0 {
		doTranslate = translate[0]
	}

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
		if err := processPackageTree(pkg, target, packageNameMapping, doTranslate, &state); err != nil {
			return domain.Err[DesiredState](err)
		}
	}

	return domain.Ok(state)
}

// processPackageTree walks a package tree and adds link/dir specs to state.
func processPackageTree(pkg domain.Package, target domain.TargetPath, packageNameMapping bool, translate bool, state *DesiredState) error {
	return walkPackageFiles(*pkg.Tree, pkg.Path, pkg.Name, target, packageNameMapping, translate, state)
}

// walkPackageFiles recursively processes files in a package tree.
func walkPackageFiles(node domain.Node, pkgRoot domain.PackagePath, pkgName string, target domain.TargetPath, packageNameMapping bool, translate bool, state *DesiredState) error {
	// Process files only (not directories or symlinks)
	if node.Type == domain.NodeFile {
		// Compute relative path from package root
		relPathResult := relativePath(pkgRoot, node.Path)
		if relPathResult.IsErr() {
			return relPathResult.UnwrapErr()
		}
		relPath := relPathResult.Unwrap()

		// Apply dotfile translation to the relative path (only if enabled)
		translated := relPath
		if translate {
			translated = translatePath(relPath)
		}

		// Compute target path
		var targetPath domain.TargetPath
		if packageNameMapping {
			// Apply package name translation and prepend to path.
			// Note: TranslatePackageName is intentionally not gated by the translate flag.
			// packageNameMapping controls directory structure (dot-gnupg -> .gnupg/),
			// while translate controls file-level dot- prefix rewriting (dot-vimrc -> .vimrc).
			translatedPkgName := scanner.TranslatePackageName(pkgName)
			combinedPath := filepath.Join(translatedPkgName, translated)
			targetPath = target.Join(combinedPath)
		} else {
			// Stow-style full-tree layout: the package tree maps straight onto
			// the target, so paths are relative to the package root and the
			// package name never appears in the target path.
			targetPath = target.Join(translated)
		}

		// Claim the target path for this package, rejecting cross-package collisions
		if err := claimLinkTarget(state, targetPath, node.Path, pkgName); err != nil {
			return err
		}

		// Add parent directory specs
		if err := addParentDirs(targetPath, target, pkgName, state); err != nil {
			return err
		}
	}

	// Recurse on children
	for _, child := range node.Children {
		if err := walkPackageFiles(child, pkgRoot, pkgName, target, packageNameMapping, translate, state); err != nil {
			return err
		}
	}

	return nil
}

// claimLinkTarget records a link spec for pkgName at targetPath.
//
// A target path may only be claimed by one package. Re-walking the same package,
// or a package claiming the same path twice, is a no-op rather than an error.
func claimLinkTarget(state *DesiredState, targetPath domain.TargetPath, source domain.FilePath, pkgName string) error {
	key := targetPath.String()

	if existing, claimed := state.Links[key]; claimed && existing.Package != pkgName {
		return domain.ErrDuplicateTarget{
			TargetPath:    key,
			FirstPackage:  existing.Package,
			SecondPackage: pkgName,
		}
	}

	// Another package already needs this path as a directory holding its links.
	if dir, claimed := state.Dirs[key]; claimed && dir.Package != pkgName {
		return domain.ErrTargetKindConflict{
			TargetPath:  key,
			FilePackage: pkgName,
			DirPackage:  dir.Package,
		}
	}

	state.Links[key] = LinkSpec{
		Source:  source,
		Target:  targetPath,
		Package: pkgName,
	}

	return nil
}

// addParentDirs adds directory specs for all parent directories of path.
// Directories may be shared between packages; only the first claimant is
// recorded, so that two packages placing distinct files in the same directory
// create it once.
func addParentDirs(path domain.TargetPath, target domain.TargetPath, pkgName string, state *DesiredState) error {
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

		// Another package already links this exact path as a file
		if link, claimed := state.Links[parentStr]; claimed && link.Package != pkgName {
			return domain.ErrTargetKindConflict{
				TargetPath:  parentStr,
				FilePackage: link.Package,
				DirPackage:  pkgName,
			}
		}

		// Add directory spec if not already present
		if _, exists := state.Dirs[parentStr]; !exists {
			// Convert TargetPath to FilePath for DirSpec storage
			dirPathResult := domain.NewFilePath(parentStr)
			if dirPathResult.IsErr() {
				return fmt.Errorf("invalid path %s: %w", parentStr, dirPathResult.UnwrapErr())
			}
			dirPath := dirPathResult.Unwrap()
			state.Dirs[parentStr] = DirSpec{Path: dirPath, Package: pkgName}
		}

		current = parent
	}

	return nil
}

// Helper functions that will be moved to scanner package

func relativePath(base domain.PackagePath, target domain.FilePath) domain.Result[string] {
	// Simple relative path computation
	baseStr := base.String()
	targetStr := target.String()

	// If target doesn't start with base, error
	if len(targetStr) <= len(baseStr) {
		return domain.Err[string](domain.ErrInvalidPath{Path: targetStr, Reason: "not under base"})
	}

	// Strip base path and leading slash
	rel := targetStr[len(baseStr):]
	if len(rel) > 0 && rel[0] == '/' {
		rel = rel[1:]
	}

	return domain.Ok(rel)
}

func translatePath(path string) string {
	return scanner.TranslatePathAll(path)
}

// ComputeOperationsFromDesiredState converts desired state into operations
func ComputeOperationsFromDesiredState(desired DesiredState) []domain.Operation {
	// Preallocate slice for directories and links
	ops := make([]domain.Operation, 0, len(desired.Dirs)+len(desired.Links))

	// Create directory operations with content-based IDs for determinism
	for _, dirSpec := range desired.Dirs {
		id := domain.OperationID(fmt.Sprintf("dir-%s", dirSpec.Path.String()))
		ops = append(ops, domain.NewDirCreate(id, dirSpec.Path))
	}

	// Create link operations with content-based IDs for determinism
	for _, linkSpec := range desired.Links {
		id := domain.OperationID(fmt.Sprintf("link-%s->%s", linkSpec.Source.String(), linkSpec.Target.String()))
		ops = append(ops, domain.NewLinkCreate(id, linkSpec.Source, linkSpec.Target))
	}

	return ops
}
