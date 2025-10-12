package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorize(t *testing.T) {
	// Save and restore NO_COLOR
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()

	t.Run("with NO_COLOR set", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		result := colorize(mutedGreen, "test")
		assert.Equal(t, "test", result)
	})

	t.Run("without NO_COLOR returns string", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := colorize(mutedGreen, "test")
		// Just verify it returns a string (may or may not have colors depending on terminal)
		assert.Contains(t, result, "test")
	})
}

func TestColorHelpers(t *testing.T) {
	// Save and restore NO_COLOR
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()

	os.Setenv("NO_COLOR", "1")

	t.Run("success", func(t *testing.T) {
		result := success("test")
		assert.Equal(t, "test", result)
	})

	t.Run("warning", func(t *testing.T) {
		result := warning("test")
		assert.Equal(t, "test", result)
	})

	t.Run("errorText", func(t *testing.T) {
		result := errorText("test")
		assert.Equal(t, "test", result)
	})

	t.Run("info", func(t *testing.T) {
		result := info("test")
		assert.Equal(t, "test", result)
	})

	t.Run("dim", func(t *testing.T) {
		result := dim("test")
		assert.Equal(t, "test", result)
	})

	t.Run("accent", func(t *testing.T) {
		result := accent("test")
		assert.Equal(t, "test", result)
	})

	t.Run("bold", func(t *testing.T) {
		result := bold("test")
		assert.Equal(t, "test", result)
	})
}

func TestColorHelpers_WithColors(t *testing.T) {
	// Save and restore NO_COLOR
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()

	os.Unsetenv("NO_COLOR")

	t.Run("bold with colors enabled", func(t *testing.T) {
		result := bold("test")
		assert.Contains(t, result, "test")
	})

	t.Run("colorize with colors enabled", func(t *testing.T) {
		result := colorize(mutedGreen, "test")
		assert.Contains(t, result, "test")
	})
}

func TestShouldUseColor(t *testing.T) {
	// Save and restore NO_COLOR
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()

	t.Run("returns false with NO_COLOR", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		result := shouldUseColor()
		assert.False(t, result)
	})

	t.Run("returns boolean without NO_COLOR", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := shouldUseColor()
		// Just verify it returns a boolean (depends on terminal)
		assert.IsType(t, false, result)
	})
}
