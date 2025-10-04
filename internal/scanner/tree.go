// Package scanner provides pure scanning logic for filesystem traversal.
// All functions in this package are side-effect free, accepting FS interface
// for I/O operations.
package scanner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/pkg/dot"
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
func ScanTree(ctx context.Context, fs dot.FS, path dot.FilePath) dot.Result[dot.Node] {
	// Check for symlinks first (symlinks are always leaves)
	isLink, err := fs.IsSymlink(ctx, path.String())
	if err != nil {
		return dot.Err[dot.Node](fmt.Errorf("check symlink %s: %w", path.String(), err))
	}
	
	if isLink {
		return dot.Ok(dot.Node{
			Path:     path,
			Type:     dot.NodeSymlink,
			Children: nil,
		})
	}
	
	// Check if directory
	isDir, err := fs.IsDir(ctx, path.String())
	if err != nil {
		return dot.Err[dot.Node](fmt.Errorf("check directory %s: %w", path.String(), err))
	}
	
	if !isDir {
		// Regular file
		return dot.Ok(dot.Node{
			Path:     path,
			Type:     dot.NodeFile,
			Children: nil,
		})
	}
	
	// Directory - scan children
	entries, err := fs.ReadDir(ctx, path.String())
	if err != nil {
		return dot.Err[dot.Node](fmt.Errorf("read directory %s: %w", path.String(), err))
	}
	
	// Recursively scan each child
	children := make([]dot.Node, 0, len(entries))
	for _, entry := range entries {
		childPath := path.Join(entry.Name())
		
		childResult := ScanTree(ctx, fs, childPath)
		if childResult.IsErr() {
			return dot.Err[dot.Node](childResult.UnwrapErr())
		}
		
		children = append(children, childResult.Unwrap())
	}
	
	return dot.Ok(dot.Node{
		Path:     path,
		Type:     dot.NodeDir,
		Children: children,
	})
}

// Walk traverses a Node tree, calling fn for each node.
// Traversal is depth-first pre-order.
//
// If fn returns an error, traversal stops and the error is returned.
func Walk(node dot.Node, fn func(dot.Node) error) error {
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
func CollectFiles(node dot.Node) []dot.FilePath {
	var files []dot.FilePath
	
	Walk(node, func(n dot.Node) error {
		if n.Type == dot.NodeFile {
			files = append(files, n.Path)
		}
		return nil
	})
	
	return files
}

// CountNodes returns the total number of nodes in a tree.
func CountNodes(node dot.Node) int {
	count := 1 // Count this node
	
	for _, child := range node.Children {
		count += CountNodes(child)
	}
	
	return count
}

// RelativePath computes the relative path from base to target.
// Both paths must be absolute. Returns error if target is not under base.
func RelativePath(base, target dot.FilePath) dot.Result[string] {
	rel, err := filepath.Rel(base.String(), target.String())
	if err != nil {
		return dot.Err[string](fmt.Errorf("compute relative path: %w", err))
	}
	return dot.Ok(rel)
}

