package render

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayout(t *testing.T) {
	layout := NewLayout(100)
	assert.Equal(t, 100, layout.Width())
}

func TestNewLayoutAuto(t *testing.T) {
	layout := NewLayoutAuto()
	assert.Greater(t, layout.Width(), 0)
}

func TestLayout_Wrap(t *testing.T) {
	layout := NewLayout(20)
	text := "This is a long text that should wrap"
	result := layout.Wrap(text, 0)

	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 1)
	for _, line := range lines {
		assert.LessOrEqual(t, len(line), 20)
	}
}

func TestLayout_Wrap_WithIndent(t *testing.T) {
	layout := NewLayout(20)
	text := "This text should be indented"
	result := layout.Wrap(text, 2)

	lines := strings.Split(result, "\n")
	if len(lines) > 1 {
		for i := 1; i < len(lines); i++ {
			assert.True(t, strings.HasPrefix(lines[i], "  "))
		}
	}
}

func TestLayout_Indent(t *testing.T) {
	layout := NewLayout(80)
	text := "line1\nline2\nline3"
	result := layout.Indent(text, 1)

	lines := strings.Split(result, "\n")
	assert.Equal(t, 3, len(lines))
	for _, line := range lines {
		assert.True(t, strings.HasPrefix(line, "  "))
	}
}

func TestLayout_Indent_EmptyLines(t *testing.T) {
	layout := NewLayout(80)
	text := "line1\n\nline3"
	result := layout.Indent(text, 1)

	lines := strings.Split(result, "\n")
	assert.Equal(t, 3, len(lines))
	assert.True(t, strings.HasPrefix(lines[0], "  "))
	assert.Equal(t, "", lines[1]) // Empty line stays empty
	assert.True(t, strings.HasPrefix(lines[2], "  "))
}

func TestLayout_Box(t *testing.T) {
	layout := NewLayout(80)
	text := "Hello\nWorld"
	result := layout.Box(text, "Title")

	assert.Contains(t, result, "Title")
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "World")
	assert.Contains(t, result, "┌")
	assert.Contains(t, result, "└")
	assert.Contains(t, result, "│")
}

func TestLayout_Box_NoTitle(t *testing.T) {
	layout := NewLayout(80)
	text := "Content"
	result := layout.Box(text, "")

	assert.Contains(t, result, "Content")
	assert.Contains(t, result, "┌")
	assert.Contains(t, result, "└")
}

func TestLayout_Table(t *testing.T) {
	layout := NewLayout(80)
	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"Item1", "Value1"},
		{"Item2", "Value2"},
	}

	result := layout.Table(headers, rows)

	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "Value")
	assert.Contains(t, result, "Item1")
	assert.Contains(t, result, "Item2")
	assert.Contains(t, result, "│")
	assert.Contains(t, result, "─")
}

func TestLayout_Table_Empty(t *testing.T) {
	layout := NewLayout(80)
	result := layout.Table([]string{}, [][]string{})
	assert.Empty(t, result)
}

func TestLayout_Divider(t *testing.T) {
	layout := NewLayout(10)
	result := layout.Divider("─")
	// "─" is a 3-byte UTF-8 character, so 10 repetitions = 30 bytes
	assert.Equal(t, 30, len(result))
	assert.Equal(t, "──────────", result)
}

func TestLayout_Divider_DefaultChar(t *testing.T) {
	layout := NewLayout(5)
	result := layout.Divider("")
	// Default "─" is a 3-byte UTF-8 character, so 5 repetitions = 15 bytes
	assert.Equal(t, 15, len(result))
}

func TestLayout_Center(t *testing.T) {
	layout := NewLayout(20)
	result := layout.Center("test")

	// "test" is 4 chars, centered in 20 should have 8 spaces before
	spaces := strings.Count(result[:9], " ")
	assert.Equal(t, 8, spaces)
	assert.Contains(t, result, "test")
}

func TestLayout_Center_TooLong(t *testing.T) {
	layout := NewLayout(10)
	text := "very long text that exceeds width"
	result := layout.Center(text)
	assert.Equal(t, text, result)
}

func TestLayout_List(t *testing.T) {
	layout := NewLayout(80)
	items := []string{"item1", "item2", "item3"}
	result := layout.List(items, "•")

	assert.Contains(t, result, "• item1")
	assert.Contains(t, result, "• item2")
	assert.Contains(t, result, "• item3")
}

func TestLayout_List_DefaultBullet(t *testing.T) {
	layout := NewLayout(80)
	items := []string{"item1"}
	result := layout.List(items, "")

	assert.Contains(t, result, "• item1")
}

func TestLayout_NumberedList(t *testing.T) {
	layout := NewLayout(80)
	items := []string{"first", "second", "third"}
	result := layout.NumberedList(items)

	assert.Contains(t, result, "1. first")
	assert.Contains(t, result, "2. second")
	assert.Contains(t, result, "3. third")
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "with color",
			input:    "\033[31mred text\033[0m",
			expected: "red text",
		},
		{
			name:     "multiple codes",
			input:    "\033[1m\033[31mbold red\033[0m",
			expected: "bold red",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPadRight(t *testing.T) {
	result := padRight("test", 10)
	assert.Equal(t, 10, len(result))
	assert.Equal(t, "test      ", result)
}

func TestPadRight_ExactWidth(t *testing.T) {
	result := padRight("test", 4)
	assert.Equal(t, "test", result)
}

func TestPadRight_TooLong(t *testing.T) {
	result := padRight("very long text", 5)
	assert.Equal(t, "very long text", result)
}

func TestWrapText_Basic(t *testing.T) {
	result := wrapText("one two three four five", 10, 0)
	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 1)
}

func TestWrapText_ZeroWidth(t *testing.T) {
	result := wrapText("test text", 0, 0)
	assert.Contains(t, result, "test")
}

func TestWrapText_Empty(t *testing.T) {
	result := wrapText("", 80, 0)
	assert.Equal(t, "", result)
}

func TestLayout_Table_WidthAdjustment(t *testing.T) {
	layout := NewLayout(30) // Narrow layout
	headers := []string{"Very Long Header Name", "Another Long Header"}
	rows := [][]string{
		{"Short", "Data"},
	}

	result := layout.Table(headers, rows)

	// Just verify it doesn't panic and produces output
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Short")
	assert.Contains(t, result, "Data")
}

func TestLayout_Box_WithANSI(t *testing.T) {
	layout := NewLayout(80)
	text := "\033[31mcolored text\033[0m"
	result := layout.Box(text, "")

	// Box should handle ANSI codes correctly
	assert.Contains(t, result, "colored text")
	assert.Contains(t, result, "┌")
	assert.Contains(t, result, "└")
}

func TestLayout_Table_WithANSI(t *testing.T) {
	layout := NewLayout(80)
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"Item", "\033[32mOK\033[0m"},
	}

	result := layout.Table(headers, rows)

	// Table should handle ANSI codes correctly
	assert.Contains(t, result, "Item")
	assert.Contains(t, result, "OK")
}
