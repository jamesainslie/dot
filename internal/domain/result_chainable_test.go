package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResult_OrElse(t *testing.T) {
	t.Run("returns value for Ok result", func(t *testing.T) {
		result := Ok(42)
		value := result.OrElse(func() int { return 99 })
		assert.Equal(t, 42, value)
	})

	t.Run("executes fallback for Err result", func(t *testing.T) {
		result := Err[int](errors.New("failed"))
		value := result.OrElse(func() int { return 99 })
		assert.Equal(t, 99, value)
	})

	t.Run("fallback is not called for Ok result", func(t *testing.T) {
		result := Ok(42)
		called := false
		value := result.OrElse(func() int {
			called = true
			return 99
		})
		assert.Equal(t, 42, value)
		assert.False(t, called, "fallback should not be called for Ok result")
	})
}

func TestResult_OrDefault(t *testing.T) {
	t.Run("returns value for Ok result", func(t *testing.T) {
		result := Ok(42)
		value := result.OrDefault()
		assert.Equal(t, 42, value)
	})

	t.Run("returns zero value for Err result", func(t *testing.T) {
		result := Err[int](errors.New("failed"))
		value := result.OrDefault()
		assert.Equal(t, 0, value)
	})

	t.Run("returns empty string for string type", func(t *testing.T) {
		result := Err[string](errors.New("failed"))
		value := result.OrDefault()
		assert.Equal(t, "", value)
	})
}

func TestResult_AndThen(t *testing.T) {
	t.Run("chains Ok results", func(t *testing.T) {
		result1 := Ok(10)
		result2 := FlatMap(result1, func(n int) Result[string] {
			return Ok("value:" + string(rune(n+'0')))
		})

		require.True(t, result2.IsOk())
		assert.Contains(t, result2.Unwrap(), "value")
	})

	t.Run("short-circuits on Err", func(t *testing.T) {
		result1 := Err[int](errors.New("first failed"))
		called := false
		result2 := FlatMap(result1, func(n int) Result[string] {
			called = true
			return Ok("should not reach")
		})

		assert.True(t, result2.IsErr())
		assert.False(t, called, "function should not be called for Err result")
	})
}

func TestResult_Map(t *testing.T) {
	t.Run("transforms Ok value", func(t *testing.T) {
		result1 := Ok(42)
		result2 := Map(result1, func(n int) string {
			return "number:42"
		})

		require.True(t, result2.IsOk())
		assert.Contains(t, result2.Unwrap(), "number")
	})

	t.Run("preserves Err", func(t *testing.T) {
		baseErr := errors.New("failed")
		result1 := Err[int](baseErr)
		result2 := Map(result1, func(n int) string {
			return "should not reach"
		})

		require.True(t, result2.IsErr())
		assert.True(t, errors.Is(result2.UnwrapErr(), baseErr))
	})
}

func TestResultChaining(t *testing.T) {
	t.Run("chain multiple operations", func(t *testing.T) {
		initial := Ok(5)

		final := FlatMap(initial, func(n int) Result[int] {
			return Ok(n * 2)
		})
		final = FlatMap(final, func(n int) Result[int] {
			return Ok(n + 3)
		})

		require.True(t, final.IsOk())
		assert.Equal(t, 13, final.Unwrap()) // (5 * 2) + 3
	})

	t.Run("stops at first error", func(t *testing.T) {
		initial := Ok(5)

		step1 := FlatMap(initial, func(n int) Result[int] {
			return Err[int](errors.New("step 1 failed"))
		})
		step2Called := false
		final := FlatMap(step1, func(n int) Result[int] {
			step2Called = true
			return Ok(n * 2)
		})

		assert.True(t, final.IsErr())
		assert.False(t, step2Called)
	})
}
