package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnwrapResult(t *testing.T) {
	t.Run("unwraps Ok result successfully", func(t *testing.T) {
		result := Ok(42)
		value, err := UnwrapResult(result, "get value")

		assert.NoError(t, err)
		assert.Equal(t, 42, value)
	})

	t.Run("returns error with context for Err result", func(t *testing.T) {
		baseErr := errors.New("base error")
		result := Err[int](baseErr)

		value, err := UnwrapResult(result, "operation failed")

		require.Error(t, err)
		assert.Equal(t, 0, value) // Zero value for int
		assert.Contains(t, err.Error(), "operation failed")
		assert.Contains(t, err.Error(), "base error")

		// Verify error chain preserved
		assert.True(t, errors.Is(err, baseErr))
	})

	t.Run("works with string type", func(t *testing.T) {
		result := Ok("success")
		value, err := UnwrapResult(result, "get string")

		assert.NoError(t, err)
		assert.Equal(t, "success", value)
	})

	t.Run("returns empty string for Err result", func(t *testing.T) {
		result := Err[string](errors.New("failed"))
		value, err := UnwrapResult(result, "get string")

		require.Error(t, err)
		assert.Equal(t, "", value) // Zero value for string
	})

	t.Run("works with custom types", func(t *testing.T) {
		type customType struct {
			ID   int
			Name string
		}

		result := Ok(customType{ID: 1, Name: "test"})
		value, err := UnwrapResult(result, "get custom")

		assert.NoError(t, err)
		assert.Equal(t, 1, value.ID)
		assert.Equal(t, "test", value.Name)
	})

	t.Run("returns zero value struct for Err result", func(t *testing.T) {
		type customType struct {
			ID   int
			Name string
		}

		result := Err[customType](errors.New("failed"))
		value, err := UnwrapResult(result, "get custom")

		require.Error(t, err)
		assert.Equal(t, 0, value.ID)    // Zero value
		assert.Equal(t, "", value.Name) // Zero value
	})
}

func TestUnwrapResultPatterns(t *testing.T) {
	t.Run("typical usage pattern simplification", func(t *testing.T) {
		// Before: verbose pattern
		createPath := func() Result[FilePath] {
			return NewFilePath("/absolute/path")
		}

		result := createPath()
		var path FilePath
		var err error

		// Old pattern (what we're replacing)
		if !result.IsOk() {
			t.Log("Would have been: return Plan{}, WrapError(result.UnwrapErr(), \"invalid path\")")
			err = WrapError(result.UnwrapErr(), "invalid path")
		} else {
			path = result.Unwrap()
		}

		// New simplified pattern
		path2, err2 := UnwrapResult(result, "invalid path")

		// Both should produce same results
		assert.NoError(t, err)
		assert.NoError(t, err2)
		assert.Equal(t, path, path2)
	})

	t.Run("chaining result operations", func(t *testing.T) {
		step1 := func() Result[int] {
			return Ok(10)
		}

		step2 := func(n int) Result[string] {
			if n > 0 {
				return Ok("positive")
			}
			return Err[string](errors.New("not positive"))
		}

		// Chain operations with helpers
		num, err := UnwrapResult(step1(), "step 1 failed")
		require.NoError(t, err)

		str, err := UnwrapResult(step2(num), "step 2 failed")
		require.NoError(t, err)
		assert.Equal(t, "positive", str)
	})

	t.Run("error propagation in function returns", func(t *testing.T) {
		processData := func() error {
			// Simulate Result-based operation
			result := NewFilePath("/test/path")
			_, err := UnwrapResult(result, "validate path")
			if err != nil {
				return err // Direct return, no additional wrapping needed
			}
			return nil
		}

		err := processData()
		assert.NoError(t, err)
	})
}

func TestUnwrapResultErrorChain(t *testing.T) {
	t.Run("preserves error chain through Result", func(t *testing.T) {
		original := ErrInvalidPath{Path: "/bad", Reason: "not absolute"}
		result := Err[FilePath](original)

		_, err := UnwrapResult(result, "process path")

		require.Error(t, err)

		// Should be able to unwrap to original error type
		var pathErr ErrInvalidPath
		assert.True(t, errors.As(err, &pathErr))
		assert.Equal(t, "/bad", pathErr.Path)
		assert.Equal(t, "not absolute", pathErr.Reason)
	})

	t.Run("works with nested errors", func(t *testing.T) {
		err1 := errors.New("root cause")
		err2 := WrapError(err1, "level 2")
		result := Err[int](err2)

		_, err3 := UnwrapResult(result, "level 3")

		require.Error(t, err3)
		assert.True(t, errors.Is(err3, err1))
		assert.True(t, errors.Is(err3, err2))
	})
}

func TestUnwrapResultWithDifferentErrorTypes(t *testing.T) {
	t.Run("ErrSourceNotFound", func(t *testing.T) {
		err := ErrSourceNotFound{Path: "/missing"}
		result := Err[FilePath](err)

		_, unwrapped := UnwrapResult(result, "load source")

		require.Error(t, unwrapped)

		var sourceErr ErrSourceNotFound
		assert.True(t, errors.As(unwrapped, &sourceErr))
		assert.Equal(t, "/missing", sourceErr.Path)
	})

	t.Run("ErrPermissionDenied", func(t *testing.T) {
		err := ErrPermissionDenied{Path: "/protected", Operation: "write"}
		result := Err[FilePath](err)

		_, unwrapped := UnwrapResult(result, "check permissions")

		require.Error(t, unwrapped)

		var permErr ErrPermissionDenied
		assert.True(t, errors.As(unwrapped, &permErr))
		assert.Equal(t, "/protected", permErr.Path)
		assert.Equal(t, "write", permErr.Operation)
	})
}

func TestUnwrapResultComparisonWithOldPattern(t *testing.T) {
	t.Run("line count reduction", func(t *testing.T) {
		// This test documents the boilerplate reduction

		// Old pattern: 5 lines
		oldPatternLines := `
result := NewFilePath(path)
if !result.IsOk() {
    return WrapError(result.UnwrapErr(), "invalid path")
}
value := result.Unwrap()`

		// New pattern: 3 lines
		newPatternLines := `
result := NewFilePath(path)
value, err := UnwrapResult(result, "invalid path")
return err`

		t.Logf("Old pattern:\n%s\n", oldPatternLines)
		t.Logf("New pattern:\n%s\n", newPatternLines)
		t.Log("Reduction: 2 lines saved per Result unwrap")
	})
}

func TestUnwrapResultPerformance(t *testing.T) {
	t.Run("minimal overhead for Ok results", func(t *testing.T) {
		iterations := 10000
		for i := 0; i < iterations; i++ {
			result := Ok(i)
			val, err := UnwrapResult(result, "iteration")
			if err != nil || val != i {
				t.Fatalf("iteration %d failed", i)
			}
		}
	})

	t.Run("efficient error path", func(t *testing.T) {
		baseErr := errors.New("base")
		iterations := 1000

		for i := 0; i < iterations; i++ {
			result := Err[int](baseErr)
			_, err := UnwrapResult(result, "iteration")
			if err == nil {
				t.Fatal("should have error")
			}
		}
	})
}
