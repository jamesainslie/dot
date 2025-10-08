// Package scanner provides pure scanning logic for filesystem traversal.
// All functions in this package are side-effect free, accepting FS interface
// for I/O operations.
package scanner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/domain"
)

// ScanTree recursively scans a filesystem tree starting at path.
// Returns a Node representing the tree structure.
//
// The scanning logic:
// 1. Check if path is a symlink (symlinks are leaf nodes)
// 2. Check if path is a directory
// 3. If directory, recursively scan children
// 4. If file, return file node
//
// This is a pure function - all I/O goes through the FS interface.
func ScanTree(ctx context.Context, fs domain.FS, path domain.FilePath) domain.Result[domain.Node] {
	// Check for symlinks first (symlinks are always leaves)
	isLink, err := fs.IsSymlink(ctx, path.String())
	if err != nil {
		return domain.Err[domain.Node](fmt.Errorf("check symlink %s: %w", path.String(), err))
	}

	if isLink {
		return domain.Ok(domain.Node{
			Path:     path,
			Type:     domain.NodeSymlink,
			Children: nil,
		})
	}

	// Check if directory
	isDir, err := fs.IsDir(ctx, path.String())
	if err != nil {
		return domain.Err[domain.Node](fmt.Errorf("check directory %s: %w", path.String(), err))
	}

	if !isDir {
		// Regular file
		return domain.Ok(domain.Node{
			Path:     path,
			Type:     domain.NodeFile,
			Children: nil,
		})
	}

	// Directory - scan children
	entries, err := fs.ReadDir(ctx, path.String())
	if err != nil {
		return domain.Err[domain.Node](fmt.Errorf("read directory %s: %w", path.String(), err))
	}

	// Recursively scan each child
	children := make([]domain.Node, 0, len(entries))
	for _, entry := range entries {
		childPath := path.Join(entry.Name())

		childResult := ScanTree(ctx, fs, childPath)
		if childResult.IsErr() {
			return domain.Err[domain.Node](childResult.UnwrapErr())
		}

		children = append(children, childResult.Unwrap())
	}

	return domain.Ok(domain.Node{
		Path:     path,
		Type:     domain.NodeDir,
		Children: children,
	})
}

// Walk traverses a Node tree, calling fn for each node.
// Traversal is depth-first pre-order.
//
// If fn returns an error, traversal stops and the error is returned.
func Walk(node domain.Node, fn func(domain.Node) error) error {
	// Visit current node
	if err := fn(node); err != nil {
		return err
	}

	// Visit children
	for _, child := range node.Children {
		if err := Walk(child, fn); err != nil {
			return err
		}
	}

	return nil
}

// CollectFiles returns all file paths in a tree.
// Useful for collecting all files in a package.
func CollectFiles(node domain.Node) []domain.FilePath {
	var files []domain.FilePath

	Walk(node, func(n domain.Node) error {
		if n.Type == domain.NodeFile {
			files = append(files, n.Path)
		}
		return nil
	})

	return files
}

// CountNodes returns the total number of nodes in a tree.
func CountNodes(node domain.Node) int {
	count := 1 // Count this node

	for _, child := range node.Children {
		count += CountNodes(child)
	}

	return count
}

// RelativePath computes the relative path from base to target.
// Both paths must be absolute. Returns error if target is not under base.
func RelativePath(base, target domain.FilePath) domain.Result[string] {
	rel, err := filepath.Rel(base.String(), target.String())
	if err != nil {
		return domain.Err[string](fmt.Errorf("compute relative path: %w", err))
	}
	return domain.Ok(rel)
}
