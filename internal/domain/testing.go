package domain

// MustParsePath creates a FilePath from a string, panicking on error.
// This function is intended for use in tests only and should not be used in production code.
func MustParsePath(s string) FilePath {
	result := NewFilePath(s)
	if result.IsErr() {
		panic(result.UnwrapErr())
	}
	return result.Unwrap()
}
