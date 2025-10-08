package dot

import "github.com/jamesainslie/dot/internal/domain"

// Concrete path types re-exported from internal/domain.
// These use proper type aliases (=) and include all methods from domain types.
//
// Note: The generic Path[K PathKind] type is NOT re-exported to avoid
// Go 1.25.1 generic type alias limitations. Users should use the concrete
// types (PackagePath, TargetPath, FilePath) which work perfectly as aliases.

// PackagePath is a path to a package directory.
// Includes methods: String(), Join(), Parent(), Equals()
type PackagePath = domain.PackagePath

// TargetPath is a path to a target directory.
// Includes methods: String(), Join(), Parent(), Equals()
type TargetPath = domain.TargetPath

// FilePath is a path to a file or directory within a package.
// Includes methods: String(), Join(), Parent(), Equals()
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

// MustParseTargetPath creates a TargetPath from a string, panicking on error.
func MustParseTargetPath(s string) TargetPath {
	r := domain.NewTargetPath(s)
	if !r.IsOk() {
		panic(r.UnwrapErr())
	}
	return r.Unwrap()
}
