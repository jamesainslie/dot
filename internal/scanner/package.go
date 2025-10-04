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

	// For now, create basic package
	// Full tree scanning with ignore support will be added
	// when we integrate with the tree scanner

	return dot.Ok(dot.Package{
		Name: name,
		Path: path,
	})
}
