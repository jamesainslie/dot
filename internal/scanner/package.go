package scanner

import (
	"context"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/pkg/dot"
)

// ScanPackage scans a single package directory.
// Returns a Package containing the package metadata and file tree.
//
// The scanner:
// 1. Verifies package directory exists
// 2. Scans the directory tree
// 3. Applies ignore patterns (filtered during tree scan)
// 4. Returns Package with tree
func ScanPackage(ctx context.Context, fs dot.FS, path dot.PackagePath, name string, ignoreSet *ignore.IgnoreSet) dot.Result[dot.Package] {
	// Check if package exists
	if !fs.Exists(ctx, path.String()) {
		return dot.Err[dot.Package](dot.ErrPackageNotFound{
			Package: name,
		})
	}

	// Scan the package directory tree
	pkgFilePath := dot.NewFilePath(path.String()).Unwrap()
	treeResult := ScanTree(ctx, fs, pkgFilePath)
	if treeResult.IsErr() {
		return dot.Err[dot.Package](treeResult.UnwrapErr())
	}

	tree := treeResult.Unwrap()

	// Filter tree based on ignore patterns
	filtered := filterTree(tree, ignoreSet)

	return dot.Ok(dot.Package{
		Name: name,
		Path: path,
		Tree: &filtered,
	})
}

// filterTree removes ignored files from a tree.
// Returns a new tree with ignored nodes filtered out.
func filterTree(node dot.Node, ignoreSet *ignore.IgnoreSet) dot.Node {
	// Check if this node should be ignored
	if ignoreSet.ShouldIgnore(node.Path.String()) {
		// Return empty node to be filtered by parent
		return dot.Node{}
	}

	// If directory, filter children
	if node.Type == dot.NodeDir {
		var filteredChildren []dot.Node
		for _, child := range node.Children {
			filtered := filterTree(child, ignoreSet)
			// Skip empty nodes (ignored)
			if filtered.Path.String() != "" {
				filteredChildren = append(filteredChildren, filtered)
			}
		}

		return dot.Node{
			Path:     node.Path,
			Type:     node.Type,
			Children: filteredChildren,
		}
	}

	// File or symlink - return as-is
	return node
}
