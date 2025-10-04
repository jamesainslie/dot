package dot

import (
	"fmt"
	"strings"
)

// Domain Errors

// ErrInvalidPath indicates a path failed validation.
type ErrInvalidPath struct {
	Path   string
	Reason string
}

func (e ErrInvalidPath) Error() string {
	return fmt.Sprintf("invalid path %q: %s", e.Path, e.Reason)
}

// ErrPackageNotFound indicates a requested package does not exist.
type ErrPackageNotFound struct {
	Package string
}

func (e ErrPackageNotFound) Error() string {
	return fmt.Sprintf("package %q not found", e.Package)
}

// ErrConflict indicates a conflict that prevents an operation.
type ErrConflict struct {
	Path   string
	Reason string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("conflict at %q: %s", e.Path, e.Reason)
}

// ErrCyclicDependency indicates a circular dependency in operations.
type ErrCyclicDependency struct {
	Cycle []string
}

func (e ErrCyclicDependency) Error() string {
	return fmt.Sprintf("cyclic dependency detected: %s", strings.Join(e.Cycle, " -> "))
}

// Infrastructure Errors

// ErrFilesystemOperation indicates a filesystem operation failed.
type ErrFilesystemOperation struct {
	Operation string
	Path      string
	Err       error
}

func (e ErrFilesystemOperation) Error() string {
	return fmt.Sprintf("filesystem operation %q failed at %q: %v", e.Operation, e.Path, e.Err)
}

func (e ErrFilesystemOperation) Unwrap() error {
	return e.Err
}

// ErrPermissionDenied indicates insufficient permissions for an operation.
type ErrPermissionDenied struct {
	Path      string
	Operation string
}

func (e ErrPermissionDenied) Error() string {
	return fmt.Sprintf("permission denied: cannot %s %q", e.Operation, e.Path)
}

// Error Aggregation

// ErrMultiple aggregates multiple errors into one.
type ErrMultiple struct {
	Errors []error
}

func (e ErrMultiple) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d errors occurred:\n", len(e.Errors))
	for i, err := range e.Errors {
		fmt.Fprintf(&b, "  %d. %v\n", i+1, err)
	}
	return b.String()
}

// Unwrap returns the underlying errors for errors.Is and errors.As support.
func (e ErrMultiple) Unwrap() []error {
	return e.Errors
}

// User-Facing Error Messages

// UserFacingError converts an error into a user-friendly message.
// Removes technical jargon and provides actionable information.
func UserFacingError(err error) string {
	switch e := err.(type) {
	case ErrPackageNotFound:
		return fmt.Sprintf("Package %q not found. Check that the package exists in your package directory.", e.Package)

	case ErrInvalidPath:
		return fmt.Sprintf("Invalid path %q: %s", e.Path, e.Reason)

	case ErrConflict:
		return fmt.Sprintf("Cannot proceed: conflict at %q\n%s", e.Path, e.Reason)

	case ErrCyclicDependency:
		return fmt.Sprintf("Circular dependency detected in operations: %s", strings.Join(e.Cycle, " → "))

	case ErrFilesystemOperation:
		return fmt.Sprintf("Failed to %s: %v", e.Operation, e.Err)

	case ErrPermissionDenied:
		return fmt.Sprintf("Permission denied: cannot %s %q\nCheck file permissions and try again.", e.Operation, e.Path)

	case ErrMultiple:
		if len(e.Errors) == 1 {
			return UserFacingError(e.Errors[0])
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Multiple errors occurred:\n")
		for i, subErr := range e.Errors {
			fmt.Fprintf(&b, "%d. %s\n", i+1, UserFacingError(subErr))
		}
		return b.String()

	default:
		return err.Error()
	}
}
