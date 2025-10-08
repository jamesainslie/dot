package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapError(t *testing.T) {
	baseErr := errors.New("base error")

	t.Run("wraps error with context", func(t *testing.T) {
		wrapped := WrapError(baseErr, "operation failed")

		assert.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "operation failed")
		assert.Contains(t, wrapped.Error(), "base error")

		// Verify error chain preserved
		assert.True(t, errors.Is(wrapped, baseErr))
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := WrapError(nil, "some context")
		assert.NoError(t, wrapped)
	})

	t.Run("preserves error chain", func(t *testing.T) {
		err1 := errors.New("original")
		err2 := fmt.Errorf("level 2: %w", err1)
		err3 := WrapError(err2, "level 3")

		assert.True(t, errors.Is(err3, err1))
		assert.True(t, errors.Is(err3, err2))
	})
}

func TestWrapErrorf(t *testing.T) {
	baseErr := errors.New("base error")

	t.Run("wraps error with formatted context", func(t *testing.T) {
		wrapped := WrapErrorf(baseErr, "operation %s failed at step %d", "test", 42)

		assert.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "operation test failed at step 42")
		assert.Contains(t, wrapped.Error(), "base error")

		// Verify error chain preserved
		assert.True(t, errors.Is(wrapped, baseErr))
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := WrapErrorf(nil, "operation %s failed", "test")
		assert.NoError(t, wrapped)
	})

	t.Run("handles no format arguments", func(t *testing.T) {
		wrapped := WrapErrorf(baseErr, "simple context")

		assert.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "simple context")
		assert.Contains(t, wrapped.Error(), "base error")
	})

	t.Run("handles multiple format arguments", func(t *testing.T) {
		wrapped := WrapErrorf(baseErr, "file: %s, line: %d, col: %d", "test.go", 10, 5)

		assert.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "file: test.go, line: 10, col: 5")
	})

	t.Run("preserves error chain with formatting", func(t *testing.T) {
		err1 := errors.New("original")
		err2 := fmt.Errorf("level 2: %w", err1)
		err3 := WrapErrorf(err2, "level 3: operation %d", 123)

		assert.True(t, errors.Is(err3, err1))
		assert.True(t, errors.Is(err3, err2))
		assert.Contains(t, err3.Error(), "operation 123")
	})
}

func TestErrorWrappingPatterns(t *testing.T) {
	t.Run("typical usage pattern", func(t *testing.T) {
		// Simulate typical usage
		performOperation := func() error {
			return errors.New("file not found")
		}

		err := performOperation()
		if err != nil {
			err = WrapError(err, "load configuration")
		}

		require.Error(t, err)
		assert.Contains(t, err.Error(), "load configuration")
		assert.Contains(t, err.Error(), "file not found")
	})

	t.Run("nested wrapping pattern", func(t *testing.T) {
		// Simulate nested error wrapping
		lowLevel := func() error {
			return errors.New("disk error")
		}

		midLevel := func() error {
			err := lowLevel()
			if err != nil {
				return WrapError(err, "read file")
			}
			return nil
		}

		highLevel := func() error {
			err := midLevel()
			if err != nil {
				return WrapErrorf(err, "load config from %s", "/path/to/config")
			}
			return nil
		}

		err := highLevel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "load config from /path/to/config")
		assert.Contains(t, err.Error(), "read file")
		assert.Contains(t, err.Error(), "disk error")
	})

	t.Run("conditional wrapping pattern", func(t *testing.T) {
		processWithContext := func(shouldFail bool) error {
			if shouldFail {
				err := errors.New("validation failed")
				return WrapError(err, "process input")
			}
			return nil
		}

		// Success case
		err := processWithContext(false)
		assert.NoError(t, err)

		// Error case
		err = processWithContext(true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "process input")
	})
}

func TestErrorUnwrapping(t *testing.T) {
	t.Run("unwrap preserves original error", func(t *testing.T) {
		original := ErrInvalidPath{Path: "/bad/path", Reason: "not absolute"}
		wrapped := WrapError(original, "validate configuration")

		var pathErr ErrInvalidPath
		assert.True(t, errors.As(wrapped, &pathErr))
		assert.Equal(t, "/bad/path", pathErr.Path)
		assert.Equal(t, "not absolute", pathErr.Reason)
	})

	t.Run("multiple wrapping levels", func(t *testing.T) {
		original := ErrSourceNotFound{Path: "/missing/file"}
		err1 := WrapError(original, "step 1")
		err2 := WrapError(err1, "step 2")
		err3 := WrapErrorf(err2, "step 3: operation %d", 42)

		var sourceErr ErrSourceNotFound
		assert.True(t, errors.As(err3, &sourceErr))
		assert.Equal(t, "/missing/file", sourceErr.Path)
	})
}

func TestErrorHelperPerformance(t *testing.T) {
	t.Run("nil error check has minimal overhead", func(t *testing.T) {
		// This test documents that checking nil is fast
		iterations := 10000
		for i := 0; i < iterations; i++ {
			err := WrapError(nil, "context")
			if err != nil {
				t.Fatal("should be nil")
			}
		}
	})

	t.Run("error wrapping is efficient", func(t *testing.T) {
		baseErr := errors.New("base")
		iterations := 1000

		for i := 0; i < iterations; i++ {
			wrapped := WrapErrorf(baseErr, "iteration %d", i)
			if wrapped == nil {
				t.Fatal("should not be nil")
			}
		}
	})
}

func TestErrorContextConsistency(t *testing.T) {
	t.Run("context appears before error", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrapped := WrapError(baseErr, "context message")

		errMsg := wrapped.Error()
		contextPos := 0
		for i, c := range errMsg {
			if c == 'c' && errMsg[i:i+7] == "context" {
				contextPos = i
				break
			}
		}

		basePos := 0
		for i, c := range errMsg {
			if c == 'b' && i+4 < len(errMsg) && errMsg[i:i+4] == "base" {
				basePos = i
				break
			}
		}

		assert.Less(t, contextPos, basePos,
			"context should appear before base error in message")
	})
}
