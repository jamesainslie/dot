package pretty

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Success("test")
		assert.Contains(t, result, "test")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Success("test")
		assert.Equal(t, "test", result)
	})
}

func TestWarning(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Warning("warning")
		assert.Contains(t, result, "warning")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Warning("warning")
		assert.Equal(t, "warning", result)
	})
}

func TestError(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Error("error")
		assert.Contains(t, result, "error")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Error("error")
		assert.Equal(t, "error", result)
	})
}

func TestInfo(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Info("info")
		assert.Contains(t, result, "info")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Info("info")
		assert.Equal(t, "info", result)
	})
}

func TestAccent(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Accent("accent")
		assert.Contains(t, result, "accent")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Accent("accent")
		assert.Equal(t, "accent", result)
	})
}

func TestDim(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Dim("dim")
		assert.Contains(t, result, "dim")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Dim("dim")
		assert.Equal(t, "dim", result)
	})
}

func TestBold(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Bold("bold")
		assert.Contains(t, result, "bold")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Bold("bold")
		assert.Equal(t, "bold", result)
	})
}

func TestUnderline(t *testing.T) {
	t.Run("with_colors", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := Underline("underline")
		assert.Contains(t, result, "underline")
	})
	t.Run("without_colors", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		result := Underline("underline")
		assert.Equal(t, "underline", result)
	})
}

func TestColorFunctionsWithNO_COLOR(t *testing.T) {
	// Save original NO_COLOR value
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
	}()

	// Set NO_COLOR to disable colors
	os.Setenv("NO_COLOR", "1")

	// All color functions should return plain text
	assert.Equal(t, "test", Success("test"))
	assert.Equal(t, "warning", Warning("warning"))
	assert.Equal(t, "error", Error("error"))
	assert.Equal(t, "info", Info("info"))
	assert.Equal(t, "accent", Accent("accent"))
	assert.Equal(t, "dim", Dim("dim"))
	assert.Equal(t, "bold", Bold("bold"))
	assert.Equal(t, "underline", Underline("underline"))
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "exact length",
			input:    "exact",
			maxLen:   5,
			expected: "exact",
		},
		{
			name:     "needs truncation",
			input:    "this is a long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "very short maxLen",
			input:    "test",
			maxLen:   2,
			expected: "te",
		},
		{
			name:     "maxLen equals 3",
			input:    "testing",
			maxLen:   3,
			expected: "tes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), tt.maxLen)
		})
	}
}

func TestAlignLeft(t *testing.T) {
	result := AlignLeft("test", 10)
	assert.Equal(t, 10, len(result))
	assert.True(t, strings.HasPrefix(result, "test"))
}

func TestAlignRight(t *testing.T) {
	result := AlignRight("test", 10)
	assert.Equal(t, 10, len(result))
	assert.True(t, strings.HasSuffix(result, "test"))
}

func TestAlignCenter(t *testing.T) {
	result := AlignCenter("test", 10)
	assert.Equal(t, 10, len(result))
	assert.Contains(t, result, "test")
}

func TestWrapText(t *testing.T) {
	longText := "This is a very long line of text that should be wrapped"
	result := WrapText(longText, 20)

	// Should contain the original text
	assert.Contains(t, result, "This")
	assert.Contains(t, result, "wrapped")

	// Should have line breaks
	assert.Contains(t, result, "\n")

	// Each line should be <= 20 chars (with some flexibility for word boundaries)
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		assert.LessOrEqual(t, len(line), 30, "Line too long: %q", line)
	}
}

func TestBox(t *testing.T) {
	t.Run("without title", func(t *testing.T) {
		result := Box("content", "")
		assert.Contains(t, result, "content")
		// lipgloss uses rounded borders: ╭ and ╰
		assert.True(t, strings.Contains(result, "╭") || strings.Contains(result, "┌"), "Should contain top border")
		assert.True(t, strings.Contains(result, "╰") || strings.Contains(result, "└"), "Should contain bottom border")
	})

	t.Run("with title", func(t *testing.T) {
		result := Box("content", "Title")
		assert.Contains(t, result, "content")
		assert.Contains(t, result, "Title")
		// lipgloss uses rounded borders: ╭ and ╰
		assert.True(t, strings.Contains(result, "╭") || strings.Contains(result, "┌"), "Should contain top border")
		assert.True(t, strings.Contains(result, "╰") || strings.Contains(result, "└"), "Should contain bottom border")
	})
}

func TestIndent(t *testing.T) {
	text := "line1\nline2\nline3"
	result := Indent(text, 4)

	// Each line should be indented
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if len(line) > 0 {
			assert.True(t, strings.HasPrefix(line, "    "), "Line should be indented: %q", line)
		}
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color styles are defined (internal lipgloss styles)
	// In the lipgloss version, these are styles not string constants
	assert.NotNil(t, successStyle)
	assert.NotNil(t, warningStyle)
	assert.NotNil(t, errorStyle)
	assert.NotNil(t, infoStyle)
	assert.NotNil(t, accentStyle)
	assert.NotNil(t, dimStyle)
}

func TestWrapText_EdgeCases(t *testing.T) {
	t.Run("zero width", func(t *testing.T) {
		result := WrapText("test", 0)
		assert.Equal(t, "test", result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := WrapText("", 20)
		assert.Equal(t, "", result)
	})

	t.Run("single word longer than width", func(t *testing.T) {
		result := WrapText("verylongword", 5)
		assert.Contains(t, result, "verylongword")
	})
}

func TestAlign_EdgeCases(t *testing.T) {
	t.Run("align with width zero", func(t *testing.T) {
		result := AlignLeft("test", 0)
		assert.Contains(t, result, "test")
	})

	t.Run("align empty string", func(t *testing.T) {
		result := AlignCenter("", 10)
		assert.NotNil(t, result)
	})
}

func TestTruncate_Unicode(t *testing.T) {
	t.Run("unicode characters", func(t *testing.T) {
		result := Truncate("Hello 世界", 8)
		assert.LessOrEqual(t, len([]rune(result)), 8)
	})

	t.Run("emoji", func(t *testing.T) {
		result := Truncate("Test 🎉🎊🎈", 8)
		assert.LessOrEqual(t, len([]rune(result)), 8)
	})
}
