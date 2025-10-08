package scanner

import (
	"context"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/domain"
)

// ScanPackage scans a single package directory.
// Returns a Package containing the package metadata and file tree.
//
// The scanner:
// 1. Verifies package directory exists
// 2. Scans the directory tree
// 3. Applies ignore patterns (filtered during tree scan)
// 4. Returns Package with tree
func ScanPackage(ctx context.Context, fs domain.FS, path domain.PackagePath, name string, ignoreSet *ignore.IgnoreSet) domain.Result[domain.Package] {
	// Check if package exists
	if !fs.Exists(ctx, path.String()) {
		return domain.Err[domain.Package](domain.ErrPackageNotFound{
			Package: name,
		})
	}

	// Scan the package directory tree
	pkgFilePath := domain.NewFilePath(path.String()).Unwrap()
	treeResult := ScanTree(ctx, fs, pkgFilePath)
	if treeResult.IsErr() {
		return domain.Err[domain.Package](treeResult.UnwrapErr())
	}

	tree := treeResult.Unwrap()

	// Filter tree based on ignore patterns
	filtered := filterTree(tree, ignoreSet)

	return domain.Ok(domain.Package{
		Name: name,
		Path: path,
		Tree: &filtered,
	})
}

// filterTree removes ignored files from a tree.
// Returns a new tree with ignored nodes filtered out.
func filterTree(node domain.Node, ignoreSet *ignore.IgnoreSet) domain.Node {
	// Check if this node should be ignored
	if ignoreSet.ShouldIgnore(node.Path.String()) {
		// Return empty node to be filtered by parent
		return domain.Node{}
	}

	// If directory, filter children
	if node.Type == domain.NodeDir {
		var filteredChildren []domain.Node
		for _, child := range node.Children {
			filtered := filterTree(child, ignoreSet)
			// Skip empty nodes (ignored)
			if filtered.Path.String() != "" {
				filteredChildren = append(filteredChildren, filtered)
			}
		}

		return domain.Node{
			Path:     node.Path,
			Type:     node.Type,
			Children: filteredChildren,
		}
	}

	// File or symlink - return as-is
	return node
}
