package dot

import "github.com/jamesainslie/dot/internal/domain"

// Error types re-exported from internal/domain

// ErrInvalidPath represents a path validation error.
type ErrInvalidPath = domain.ErrInvalidPath

// ErrPackageNotFound represents a missing package error.
type ErrPackageNotFound = domain.ErrPackageNotFound

// ErrConflict represents a conflict during installation.
type ErrConflict = domain.ErrConflict

// ErrCyclicDependency represents a dependency cycle error.
type ErrCyclicDependency = domain.ErrCyclicDependency

// ErrFilesystemOperation represents a filesystem operation error.
type ErrFilesystemOperation = domain.ErrFilesystemOperation

// ErrPermissionDenied represents a permission denied error.
type ErrPermissionDenied = domain.ErrPermissionDenied

// ErrMultiple represents multiple aggregated errors.
type ErrMultiple = domain.ErrMultiple

// ErrEmptyPlan represents an empty plan error.
type ErrEmptyPlan = domain.ErrEmptyPlan

// ErrSourceNotFound represents a missing source file error.
type ErrSourceNotFound = domain.ErrSourceNotFound

// ErrExecutionFailed represents an execution failure error.
type ErrExecutionFailed = domain.ErrExecutionFailed

// ErrParentNotFound represents a missing parent directory error.
type ErrParentNotFound = domain.ErrParentNotFound

// ErrCheckpointNotFound represents a missing checkpoint error.
type ErrCheckpointNotFound = domain.ErrCheckpointNotFound

// ErrNotImplemented represents a not implemented error.
type ErrNotImplemented = domain.ErrNotImplemented

// UserFacingError converts an error into a user-friendly message.
func UserFacingError(err error) string {
	return domain.UserFacingError(err)
}