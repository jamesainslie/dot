package domain

// Result represents a value or an error, implementing a Result monad for error handling.
// This provides a functional approach to error handling with composition support.
type Result[T any] struct {
	value T
	err   error
	isOk  bool
}

// Ok creates a successful Result containing a value.
func Ok[T any](value T) Result[T] {
	return Result[T]{
		value: value,
		isOk:  true,
	}
}

// Err creates a failed Result containing an error.
func Err[T any](err error) Result[T] {
	return Result[T]{
		err:  err,
		isOk: false,
	}
}

// IsOk returns true if the Result contains a value.
func (r Result[T]) IsOk() bool {
	return r.isOk
}

// IsErr returns true if the Result contains an error.
func (r Result[T]) IsErr() bool {
	return !r.isOk
}

// Unwrap returns the contained value.
//
// Panics if the Result contains an error. This is by design to enforce
// proper error handling. In production code, prefer UnwrapOr() to provide
// a safe default, or check IsOk() before calling Unwrap().
//
// Unwrap is appropriate when you have proven the Result must be Ok,
// such as in tests or after explicit IsOk() checks.
//
// Example:
//
//	// Safe - explicit check
//	if result.IsOk() {
//	    value := result.Unwrap()
//	}
//
//	// Safer - provides default
//	value := result.UnwrapOr(defaultValue)
func (r Result[T]) Unwrap() T {
	if !r.isOk {
		panic("called Unwrap on an Err Result")
	}
	return r.value
}

// UnwrapErr returns the contained error.
//
// Panics if the Result contains a value. Use IsErr() to check before
// calling, or use pattern matching with IsOk()/IsErr() branches.
//
// Appropriate for error handling paths where you've confirmed the
// Result is Err.
//
// Example:
//
//	// Safe - explicit check
//	if result.IsErr() {
//	    err := result.UnwrapErr()
//	    // handle error
//	}
func (r Result[T]) UnwrapErr() error {
	if r.isOk {
		panic("called UnwrapErr on an Ok Result")
	}
	return r.err
}

// UnwrapOr returns the contained value or a default if Result is Err.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.isOk {
		return r.value
	}
	return defaultValue
}

// Map applies a function to the contained value if Ok, otherwise propagates the error.
// This is the functorial map operation.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if !r.isOk {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}

// FlatMap applies a function that returns a Result to the contained value if Ok.
// This is the monadic bind operation, enabling composition of Result-returning functions.
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if !r.isOk {
		return Err[U](r.err)
	}
	return fn(r.value)
}

// Collect aggregates a slice of Results into a Result containing a slice.
// Returns Err if any Result is Err, otherwise returns Ok with all values.
func Collect[T any](results []Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))
	for _, r := range results {
		if !r.isOk {
			return Err[[]T](r.err)
		}
		values = append(values, r.value)
	}
	return Ok(values)
}
