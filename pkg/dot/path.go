package dot

import "github.com/jamesainslie/dot/internal/domain"

// PathKind is a marker interface for phantom type parameters.
// Re-exported from internal/domain.
type PathKind = domain.PathKind

// PackageDirKind marks paths pointing to package directories.
type PackageDirKind = domain.PackageDirKind

// TargetDirKind marks paths pointing to target directories.
type TargetDirKind = domain.TargetDirKind

// FileDirKind marks paths pointing to file directories.
type FileDirKind = domain.FileDirKind

// Path represents a filesystem path with phantom typing for compile-time safety.
// Re-exported from internal/domain.
type Path[K PathKind] domain.Path[K]

// String returns the string representation of the path.
func (p Path[K]) String() string {
	return domain.Path[K](p).String()
}

// Join appends a path component, returning a FilePath.
func (p Path[K]) Join(elem string) FilePath {
	return domain.Path[K](p).Join(elem)
}

// Parent returns the parent directory of this path.
func (p Path[K]) Parent() Result[Path[K]] {
	r := domain.Path[K](p).Parent()
	if r.IsErr() {
		return Err[Path[K]](r.UnwrapErr())
	}
	return Ok(Path[K](r.Unwrap()))
}

// Equals checks if two paths are equal.
func (p Path[K]) Equals(other Path[K]) bool {
	return domain.Path[K](p).Equals(domain.Path[K](other))
}

// PackagePath is a path to a package directory.
type PackagePath = domain.PackagePath

// TargetPath is a path to a target directory.
type TargetPath = domain.TargetPath

// FilePath is a path to a file or directory within a package.
type FilePath = domain.FilePath

// NewPackagePath creates a new package path with validation.
func NewPackagePath(s string) Result[PackagePath] {
	r := domain.NewPackagePath(s)
	return Result[PackagePath](r)
}

// NewTargetPath creates a new target path with validation.
func NewTargetPath(s string) Result[TargetPath] {
	r := domain.NewTargetPath(s)
	return Result[TargetPath](r)
}

// NewFilePath creates a new file path with validation.
func NewFilePath(s string) Result[FilePath] {
	r := domain.NewFilePath(s)
	return Result[FilePath](r)
}

// MustParsePath creates a FilePath from a string, panicking on error.
func MustParsePath(s string) FilePath {
	return domain.MustParsePath(s)
}
