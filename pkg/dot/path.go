// Package dot provides the public API for the dot dotfile manager.
package dot

import (
	"path/filepath"
	"strings"
)

// PathKind is a marker interface for phantom type parameters.
// Implementations exist only at compile time to enforce type safety.
type PathKind interface {
	pathKind()
}

// PackageDirKind marks paths pointing to package directories.
type PackageDirKind struct{}

func (PackageDirKind) pathKind() {}

// TargetDirKind marks paths pointing to target directories.
type TargetDirKind struct{}

func (TargetDirKind) pathKind() {}

// FileDirKind marks paths pointing to file directories.
type FileDirKind struct{}

func (FileDirKind) pathKind() {}

// Path represents a filesystem path with phantom typing for compile-time safety.
// The type parameter K ensures paths of different kinds cannot be mixed accidentally.
type Path[K PathKind] struct {
	path string
}

// PackagePath is a path to a package directory.
type PackagePath = Path[PackageDirKind]

// TargetPath is a path to a target directory.
type TargetPath = Path[TargetDirKind]

// FilePath is a path to a file or directory within a package.
type FilePath = Path[FileDirKind]

// NewPackagePath creates a new package path with validation.
// Returns error if path is not absolute.
func NewPackagePath(s string) Result[PackagePath] {
	if !filepath.IsAbs(s) {
		return Err[PackagePath](ErrInvalidPath{Path: s, Reason: "path must be absolute"})
	}
	return Ok(Path[PackageDirKind]{path: clean(s)})
}

// NewTargetPath creates a new target path with validation.
// Returns error if path is not absolute.
func NewTargetPath(s string) Result[TargetPath] {
	if !filepath.IsAbs(s) {
		return Err[TargetPath](ErrInvalidPath{Path: s, Reason: "path must be absolute"})
	}
	return Ok(Path[TargetDirKind]{path: clean(s)})
}

// NewFilePath creates a new file path with validation.
// Returns error if path is not absolute.
func NewFilePath(s string) Result[FilePath] {
	if !filepath.IsAbs(s) {
		return Err[FilePath](ErrInvalidPath{Path: s, Reason: "path must be absolute"})
	}
	return Ok(Path[FileDirKind]{path: clean(s)})
}

// String returns the string representation of the path.
func (p Path[K]) String() string {
	return p.path
}

// Join appends a path component, returning a FilePath.
func (p Path[K]) Join(elem string) FilePath {
	joined := filepath.Join(p.path, elem)
	return Path[FileDirKind]{path: joined}
}

// Parent returns the parent directory of this path.
func (p Path[K]) Parent() Result[Path[K]] {
	parent := filepath.Dir(p.path)
	if parent == p.path {
		return Err[Path[K]](ErrInvalidPath{Path: p.path, Reason: "path has no parent"})
	}
	return Ok(Path[K]{path: parent})
}

// Equals checks if two paths are equal.
func (p Path[K]) Equals(other Path[K]) bool {
	return p.path == other.path
}

// clean normalizes a path by removing redundant separators and resolving dots.
func clean(path string) string {
	cleaned := filepath.Clean(path)
	// Remove trailing slash except for root
	if len(cleaned) > 1 && strings.HasSuffix(cleaned, string(filepath.Separator)) {
		cleaned = strings.TrimSuffix(cleaned, string(filepath.Separator))
	}
	return cleaned
}

