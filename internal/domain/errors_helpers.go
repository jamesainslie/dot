package domain

import "fmt"

// WrapError wraps an error with contextual message while preserving the error chain.
// Returns nil if the provided error is nil, allowing for safe usage in conditional checks.
//
// Example usage:
//
//	err := someOperation()
//	if err != nil {
//	    return WrapError(err, "operation failed")
//	}
//
// The resulting error message format is: "context: original error"
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorf wraps an error with formatted contextual message while preserving the error chain.
// Returns nil if the provided error is nil, allowing for safe usage in conditional checks.
//
// Example usage:
//
//	err := loadFile(path)
//	if err != nil {
//	    return WrapErrorf(err, "load config from %s", path)
//	}
//
// The resulting error message format is: "formatted context: original error"
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, err)
}
