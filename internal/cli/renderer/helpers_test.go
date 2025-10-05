package renderer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndent(t *testing.T) {
	text := "line1\nline2\nline3"
	result := indent(text, 2)

	assert.Contains(t, result, "  line1")
	assert.Contains(t, result, "  line2")
	assert.Contains(t, result, "  line3")
}

func TestIndent_EmptyLines(t *testing.T) {
	text := "line1\n\nline3"
	result := indent(text, 2)

	lines := []string{"  line1", "", "  line3"}
	for _, line := range lines {
		assert.Contains(t, result, line)
	}
}

func TestWrapText(t *testing.T) {
	text := "this is a very long line that should be wrapped at a certain width to fit better"
	result := wrapText(text, 20)

	lines := countLines(result)
	assert.Greater(t, lines, 1)
}

func TestWrapText_ShortText(t *testing.T) {
	text := "short"
	result := wrapText(text, 20)

	assert.Equal(t, "short", result)
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
