package pretty

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPagerConfig(t *testing.T) {
	config := DefaultPagerConfig()

	assert.Greater(t, config.PageSize, 0, "PageSize should be positive")
	assert.NotNil(t, config.Output, "Output should not be nil")
}

func TestNewPager(t *testing.T) {
	t.Run("with custom page size", func(t *testing.T) {
		var buf bytes.Buffer
		config := PagerConfig{
			PageSize: 10,
			Output:   &buf,
		}

		pager := NewPager(config)
		require.NotNil(t, pager)
		assert.Equal(t, 10, pager.pageSize)
	})

	t.Run("with zero page size auto-detects", func(t *testing.T) {
		var buf bytes.Buffer
		config := PagerConfig{
			PageSize: 0,
			Output:   &buf,
		}

		pager := NewPager(config)
		require.NotNil(t, pager)
		assert.Greater(t, pager.pageSize, 0, "Should auto-detect page size")
	})

	t.Run("enforces minimum page size", func(t *testing.T) {
		var buf bytes.Buffer
		config := PagerConfig{
			PageSize: 2,
			Output:   &buf,
		}

		pager := NewPager(config)
		require.NotNil(t, pager)
		assert.GreaterOrEqual(t, pager.pageSize, 5, "Should enforce minimum page size")
	})
}

func TestPager_Page_ShortContent(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Content shorter than page size should be printed without pagination
	content := "Line 1\nLine 2\nLine 3"
	err := pager.Page(content)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Line 1")
	assert.Contains(t, buf.String(), "Line 2")
	assert.Contains(t, buf.String(), "Line 3")
}

func TestPager_Page_ExactPageSize(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 3,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Content exactly matching page size should be printed without pagination
	content := "Line 1\nLine 2\nLine 3"
	err := pager.Page(content)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 2")
	assert.Contains(t, output, "Line 3")
	// Should not contain pagination prompt
	assert.NotContains(t, output, "More")
}

func TestPager_Page_EmptyContent(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	err := pager.Page("")

	assert.NoError(t, err)
	// Empty content should produce minimal output
}

func TestPager_Page_SingleLine(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	content := "Single line of text"
	err := pager.Page(content)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Single line of text")
}

func TestPager_PageLines(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	lines := []string{"Line 1", "Line 2", "Line 3"}
	err := pager.PageLines(lines)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 2")
	assert.Contains(t, output, "Line 3")
}

func TestPager_PageLines_Empty(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	lines := []string{}
	err := pager.PageLines(lines)

	assert.NoError(t, err)
}

func TestPager_PageLines_SingleLine(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	lines := []string{"Only one line"}
	err := pager.PageLines(lines)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Only one line")
}

func TestPager_MultiplePages_NonInteractive(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 3,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Create content that would normally require pagination
	lines := []string{}
	for i := 1; i <= 10; i++ {
		lines = append(lines, "Line "+string(rune('0'+i)))
	}
	content := strings.Join(lines, "\n")

	err := pager.Page(content)

	assert.NoError(t, err)

	// In non-interactive mode (which test environment is), all content should be printed
	// without pagination prompts
	output := buf.String()
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 5")

	// When running in test environment (non-interactive), pagination is bypassed
	// So we verify content is present but don't check for pagination prompts
}

func TestPager_PageSize_Configuration(t *testing.T) {
	tests := []struct {
		name            string
		configPageSize  int
		expectedMinimum int
	}{
		{"positive page size", 15, 15},
		{"zero uses default", 0, 5},
		{"negative uses default", -1, 5},
		{"small page size gets minimum", 2, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := PagerConfig{
				PageSize: tt.configPageSize,
				Output:   &buf,
			}

			pager := NewPager(config)
			assert.GreaterOrEqual(t, pager.pageSize, tt.expectedMinimum,
				"Page size should meet minimum requirement")
		})
	}
}

func TestPager_ContentWithVariousLineEndings(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Test with different line content
	content := "Line 1\n\nLine 3\n\n\nLine 6"
	err := pager.Page(content)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 3")
	assert.Contains(t, output, "Line 6")
}

func TestPager_LongSingleLine(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Very long single line (no newlines)
	content := strings.Repeat("This is a very long line of text. ", 50)
	err := pager.Page(content)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "This is a very long line")
}

func TestPager_MultipleEmptyLines(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Content with multiple consecutive empty lines
	content := "Line 1\n\n\n\nLine 5"
	err := pager.Page(content)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 5")
}

func TestPager_OutputWriter(t *testing.T) {
	var buf bytes.Buffer
	config := PagerConfig{
		PageSize: 10,
		Output:   &buf,
	}

	pager := NewPager(config)

	// Verify output goes to configured writer
	assert.Equal(t, &buf, pager.output)
}
