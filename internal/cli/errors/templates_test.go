package errors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_Render_Basic(t *testing.T) {
	tmpl := &Template{
		Title:       "Test Error",
		Description: "This is a test error",
	}

	result := tmpl.Render(false, 80)

	assert.Contains(t, result, "Test Error")
	assert.Contains(t, result, "This is a test error")
	assert.NotContains(t, result, "\033[") // No color codes
}

func TestTemplate_Render_WithDetails(t *testing.T) {
	tmpl := &Template{
		Title:       "Test Error",
		Description: "This is a test error",
		Details:     []string{"detail one", "detail two"},
	}

	result := tmpl.Render(false, 80)

	assert.Contains(t, result, "Test Error")
	assert.Contains(t, result, "Details:")
	assert.Contains(t, result, "detail one")
	assert.Contains(t, result, "detail two")
}

func TestTemplate_Render_WithSuggestions(t *testing.T) {
	tmpl := &Template{
		Title:       "Test Error",
		Description: "This is a test error",
		Suggestions: []string{"try this", "or this"},
	}

	result := tmpl.Render(false, 80)

	assert.Contains(t, result, "Test Error")
	assert.Contains(t, result, "Suggestions:")
	assert.Contains(t, result, "try this")
	assert.Contains(t, result, "or this")
	assert.Contains(t, result, "•")
}

func TestTemplate_Render_WithFooter(t *testing.T) {
	tmpl := &Template{
		Title:       "Test Error",
		Description: "This is a test error",
		Footer:      "See docs for more info",
	}

	result := tmpl.Render(false, 80)

	assert.Contains(t, result, "Test Error")
	assert.Contains(t, result, "See docs for more info")
}

func TestTemplate_Render_Complete(t *testing.T) {
	tmpl := &Template{
		Title:       "Complete Error",
		Description: "This error has all fields",
		Details:     []string{"detail one", "detail two"},
		Suggestions: []string{"suggestion one", "suggestion two"},
		Footer:      "Footer text",
	}

	result := tmpl.Render(false, 80)

	assert.Contains(t, result, "Complete Error")
	assert.Contains(t, result, "This error has all fields")
	assert.Contains(t, result, "Details:")
	assert.Contains(t, result, "detail one")
	assert.Contains(t, result, "Suggestions:")
	assert.Contains(t, result, "suggestion one")
	assert.Contains(t, result, "Footer text")
}

func TestTemplate_Render_WithColor(t *testing.T) {
	tmpl := &Template{
		Title:       "Colored Error",
		Description: "This should have colors",
		Suggestions: []string{"colorful suggestion"},
	}

	result := tmpl.Render(true, 80)

	assert.Contains(t, result, "Colored Error")
	assert.Contains(t, result, "\033[") // Should have color codes
	assert.Contains(t, result, colorRed)
	assert.Contains(t, result, colorReset)
}

func TestTemplate_Render_EmptyTemplate(t *testing.T) {
	tmpl := &Template{}

	result := tmpl.Render(false, 80)

	// Should not crash, just return empty/minimal output
	assert.NotContains(t, result, "panic")
}

func TestWrapText_Basic(t *testing.T) {
	text := "This is a simple text that should wrap"
	result := wrapText(text, 20, 0)

	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 1) // Should wrap to multiple lines

	for _, line := range lines {
		assert.LessOrEqual(t, len(line), 20)
	}
}

func TestWrapText_WithIndent(t *testing.T) {
	text := "This text should be indented when wrapped"
	result := wrapText(text, 20, 2)

	lines := strings.Split(result, "\n")
	if len(lines) > 1 {
		// Lines after first should be indented
		for i := 1; i < len(lines); i++ {
			assert.True(t, strings.HasPrefix(lines[i], "  "))
		}
	}
}

func TestWrapText_NoWrapNeeded(t *testing.T) {
	text := "Short"
	result := wrapText(text, 80, 0)

	assert.Equal(t, "Short", result)
	assert.NotContains(t, result, "\n")
}

func TestWrapText_LongWord(t *testing.T) {
	text := "ThisIsAReallyLongWordThatExceedsTheWidth"
	result := wrapText(text, 20, 0)

	// Should handle long words gracefully
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "ThisIsAReallyLongWordThatExceedsTheWidth")
}

func TestWrapText_EmptyText(t *testing.T) {
	result := wrapText("", 80, 0)
	assert.Equal(t, "", result)
}

func TestWrapText_OnlyWhitespace(t *testing.T) {
	result := wrapText("   ", 80, 0)
	assert.Equal(t, "", result)
}

func TestWrapText_ZeroWidth(t *testing.T) {
	text := "This should not crash with zero width"
	result := wrapText(text, 0, 0)

	// Should use default width
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "This")
}

func TestWrapText_NegativeWidth(t *testing.T) {
	text := "This should not crash with negative width"
	result := wrapText(text, -10, 0)

	// Should use default width
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "This")
}

func TestWrapText_VerySmallEffectiveWidth(t *testing.T) {
	text := "Text with tiny width"
	result := wrapText(text, 10, 8)

	// Effective width would be 2, should use minimum
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Text")
}

func TestWrapText_MultipleSpaces(t *testing.T) {
	text := "Word1    Word2     Word3"
	result := wrapText(text, 80, 0)

	// Multiple spaces should be normalized to single space
	assert.NotContains(t, result, "    ")
	assert.Contains(t, result, "Word1")
	assert.Contains(t, result, "Word2")
	assert.Contains(t, result, "Word3")
}

func TestWrapText_PreservesContent(t *testing.T) {
	text := "Alpha Beta Gamma Delta Epsilon"
	result := wrapText(text, 15, 0)

	// All words should be preserved
	assert.Contains(t, result, "Alpha")
	assert.Contains(t, result, "Beta")
	assert.Contains(t, result, "Gamma")
	assert.Contains(t, result, "Delta")
	assert.Contains(t, result, "Epsilon")
}

func TestTemplate_Render_LongDescription(t *testing.T) {
	tmpl := &Template{
		Title: "Error",
		Description: "This is a very long description that should wrap across multiple lines " +
			"when rendered to ensure that the output fits within the terminal width and " +
			"remains readable for the user",
	}

	result := tmpl.Render(false, 40)

	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 2) // Should wrap to multiple lines

	// Check that lines are reasonably short
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "•") {
			assert.LessOrEqual(t, len(line), 80)
		}
	}
}

func TestTemplate_Render_MultipleSuggestions(t *testing.T) {
	tmpl := &Template{
		Title:       "Error",
		Description: "Multiple suggestions",
		Suggestions: []string{
			"First suggestion",
			"Second suggestion",
			"Third suggestion",
			"Fourth suggestion",
		},
	}

	result := tmpl.Render(false, 80)

	// All suggestions should be present
	assert.Contains(t, result, "First suggestion")
	assert.Contains(t, result, "Second suggestion")
	assert.Contains(t, result, "Third suggestion")
	assert.Contains(t, result, "Fourth suggestion")

	// Should have bullet points
	bulletCount := strings.Count(result, "•")
	assert.Equal(t, 4, bulletCount)
}

func TestTemplate_Render_NoTrailingNewline(t *testing.T) {
	tmpl := &Template{
		Title:       "Error",
		Description: "Should not have trailing newline",
	}

	result := tmpl.Render(false, 80)

	assert.NotEmpty(t, result)
	assert.False(t, strings.HasSuffix(result, "\n\n"))
}
