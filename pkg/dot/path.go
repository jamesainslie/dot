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
