package domain

// UnwrapResult extracts the value from a Result or returns an error with context.
// This helper simplifies the common pattern of checking Result.IsOk() and unwrapping.
//
// Returns:
//   - (value, nil) if Result is Ok
//   - (zero value, wrapped error) if Result is Err
//
// Example usage:
//
//	// Old pattern (5 lines):
//	result := NewFilePath(path)
//	if !result.IsOk() {
//	    return WrapError(result.UnwrapErr(), "invalid path")
//	}
//	value := result.Unwrap()
//
//	// New pattern (2 lines):
//	value, err := UnwrapResult(NewFilePath(path), "invalid path")
//	if err != nil { return err }
//
// The error chain is preserved, allowing errors.Is() and errors.As() to work correctly.
func UnwrapResult[T any](r Result[T], context string) (T, error) {
	if !r.IsOk() {
		var zero T
		return zero, WrapError(r.UnwrapErr(), context)
	}
	return r.Unwrap(), nil
}
