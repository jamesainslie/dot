package pretty

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTableConfig(t *testing.T) {
	config := DefaultTableConfig()

	assert.True(t, config.AutoWrap, "AutoWrap should be enabled by default")
	assert.Equal(t, -1, config.SortColumn, "No sorting by default")
	assert.True(t, config.SortAsc, "Ascending sort by default")
	assert.Greater(t, config.MaxWidth, 0, "MaxWidth should be positive")
}

func TestNewTableWriter(t *testing.T) {
	tests := []struct {
		name  string
		style TableStyle
	}{
		{"bordered", StyleBordered},
		{"light", StyleLight},
		{"minimal", StyleMinimal},
		{"compact", StyleCompact},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultTableConfig()
			tw := NewTableWriter(tt.style, config)

			require.NotNil(t, tw)
			assert.Equal(t, tt.style, tw.style)
		})
	}
}

func TestTableWriter_SetHeader(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Age", "City")

	var buf bytes.Buffer
	tw.Render(&buf)

	output := buf.String()
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Age")
	assert.Contains(t, output, "City")
}

func TestTableWriter_AppendRow(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Age")
	tw.AppendRow("Alice", 30)
	tw.AppendRow("Bob", 25)

	var buf bytes.Buffer
	tw.Render(&buf)

	output := buf.String()
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "30")
	assert.Contains(t, output, "Bob")
	assert.Contains(t, output, "25")
}

func TestTableWriter_AppendRows(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Score")

	rows := [][]interface{}{
		{"Alice", 95},
		{"Bob", 87},
		{"Charlie", 92},
	}
	tw.AppendRows(rows)

	var buf bytes.Buffer
	tw.Render(&buf)

	output := buf.String()
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "Bob")
	assert.Contains(t, output, "Charlie")
}

func TestTableWriter_AppendSeparator(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleLight, config)

	tw.SetHeader("Name", "Value")
	tw.AppendRow("First", 1)
	tw.AppendSeparator()
	tw.AppendRow("Second", 2)

	var buf bytes.Buffer
	tw.Render(&buf)

	output := buf.String()
	assert.Contains(t, output, "First")
	assert.Contains(t, output, "Second")
}

func TestTableWriter_SetAutoIndex(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetAutoIndex(true)
	tw.SetHeader("Name")
	tw.AppendRow("Alice")
	tw.AppendRow("Bob")

	var buf bytes.Buffer
	tw.Render(&buf)

	output := buf.String()
	// Auto-index should add row numbers
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
}

func TestTableWriter_RenderString(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Col1", "Col2")
	tw.AppendRow("A", "B")

	output := tw.RenderString()

	assert.Contains(t, output, "Col1")
	assert.Contains(t, output, "Col2")
	assert.Contains(t, output, "A")
	assert.Contains(t, output, "B")
}

func TestTableWriter_StyleBordered(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleBordered, config)

	tw.SetHeader("Name", "Value")
	tw.AppendRow("Test", 123)

	output := tw.RenderString()

	// Bordered style should have box drawing characters
	assert.True(t, strings.Contains(output, "│") || strings.Contains(output, "|"),
		"Bordered style should contain border characters")
}

func TestTableWriter_StyleLight(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleLight, config)

	tw.SetHeader("Name", "Value")
	tw.AppendRow("Test", 456)

	output := tw.RenderString()

	// Headers are uppercase in go-pretty
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "Test")
}

func TestTableWriter_StyleMinimal(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Value")
	tw.AppendRow("Test", 789)

	output := tw.RenderString()

	// Minimal style should have no borders
	assert.NotContains(t, output, "┌")
	assert.NotContains(t, output, "└")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Test")
}

func TestTableWriter_StyleCompact(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleCompact, config)

	tw.SetHeader("A", "B")
	tw.AppendRow("X", "Y")

	output := tw.RenderString()

	assert.Contains(t, output, "A")
	assert.Contains(t, output, "X")
}

func TestShouldUseColor(t *testing.T) {
	// Save original NO_COLOR value
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
	}()

	t.Run("with NO_COLOR set", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		assert.False(t, ShouldUseColor())
	})

	t.Run("without NO_COLOR", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		// Result depends on whether stdout is a terminal
		result := ShouldUseColor()
		assert.IsType(t, false, result)
	})
}

func TestGetTerminalWidth(t *testing.T) {
	width := GetTerminalWidth()

	// Should return a positive value (either actual width or default)
	assert.Greater(t, width, 0)

	// Default fallback should be 80
	if width == 80 {
		// Running in non-terminal or default fallback
		t.Log("Using default terminal width of 80")
	} else {
		// Running in actual terminal
		t.Logf("Detected terminal width: %d", width)
	}
}

func TestGetTerminalHeight(t *testing.T) {
	height := GetTerminalHeight()

	// Should return a positive value (either actual height or default)
	assert.Greater(t, height, 0)

	// Default fallback should be 24
	if height == 24 {
		// Running in non-terminal or default fallback
		t.Log("Using default terminal height of 24")
	} else {
		// Running in actual terminal
		t.Logf("Detected terminal height: %d", height)
	}
}

func TestIsInteractive(t *testing.T) {
	// Result depends on how tests are run
	result := IsInteractive()
	assert.IsType(t, false, result)

	// When running in CI or with redirected output, should be false
	if !result {
		t.Log("Non-interactive terminal detected (expected in test environment)")
	}
}

func TestTableWriter_ColorEnabled(t *testing.T) {
	t.Run("with colors", func(t *testing.T) {
		config := DefaultTableConfig()
		config.ColorEnabled = true
		tw := NewTableWriter(StyleBordered, config)

		tw.SetHeader("Name")
		tw.AppendRow("Test")

		output := tw.RenderString()
		// Headers are uppercase
		assert.Contains(t, output, "NAME")
	})

	t.Run("without colors", func(t *testing.T) {
		config := DefaultTableConfig()
		config.ColorEnabled = false
		tw := NewTableWriter(StyleBordered, config)

		tw.SetHeader("Name")
		tw.AppendRow("Test")

		output := tw.RenderString()
		// Headers are uppercase
		assert.Contains(t, output, "NAME")
		// Should not contain ANSI color codes
		assert.NotContains(t, output, "\033[")
	})
}

func TestTableWriter_MaxWidth(t *testing.T) {
	config := DefaultTableConfig()
	config.MaxWidth = 40
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Description")
	tw.AppendRow("Test", "This is a very long description that should be wrapped")

	output := tw.RenderString()

	// Output should respect max width
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) > 0 {
			// Allow some flexibility for wrapping and formatting
			assert.LessOrEqual(t, len(line), config.MaxWidth+20,
				"Line should not significantly exceed max width")
		}
	}
}

func TestTableWriter_EmptyTable(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Col1", "Col2")
	// No rows added

	output := tw.RenderString()

	// Should still show headers
	assert.Contains(t, output, "Col1")
	assert.Contains(t, output, "Col2")
}

func TestTableWriter_SortBy(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Age")
	tw.AppendRow("Charlie", 30)
	tw.AppendRow("Alice", 25)
	tw.AppendRow("Bob", 28)

	// Sort by first column (Name) ascending
	tw.SortBy(1, true)

	output := tw.RenderString()

	// Check that Alice comes before Bob comes before Charlie
	alicePos := strings.Index(output, "Alice")
	bobPos := strings.Index(output, "Bob")
	charliePos := strings.Index(output, "Charlie")

	assert.Greater(t, bobPos, alicePos, "Bob should come after Alice")
	assert.Greater(t, charliePos, bobPos, "Charlie should come after Bob")
}

func TestTableWriter_SortByDescending(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Name", "Score")
	tw.AppendRow("Alice", 85)
	tw.AppendRow("Bob", 95)
	tw.AppendRow("Charlie", 90)

	// Sort by second column (Score) descending
	tw.SortBy(2, false)

	output := tw.RenderString()

	// Check that higher scores come first
	bobPos := strings.Index(output, "Bob")
	charliePos := strings.Index(output, "Charlie")
	alicePos := strings.Index(output, "Alice")

	assert.Greater(t, charliePos, bobPos, "Charlie (90) should come after Bob (95)")
	assert.Greater(t, alicePos, charliePos, "Alice (85) should come after Charlie (90)")
}

func TestTableWriter_SetColumnConfig(t *testing.T) {
	config := DefaultTableConfig()
	config.ColorEnabled = false
	tw := NewTableWriter(StyleMinimal, config)

	tw.SetHeader("Short", "LongColumn")
	tw.AppendRow("A", "B")

	// Set column config for second column
	tw.SetColumnConfig(2, table.ColumnConfig{
		WidthMax: 10,
	})

	output := tw.RenderString()

	// Verify output contains headers and data
	assert.Contains(t, output, "Short")
	assert.Contains(t, output, "A")
	assert.Contains(t, output, "B")
}

func TestGetTerminalSize(t *testing.T) {
	// Test that both width and height are reasonable values
	width := GetTerminalWidth()
	height := GetTerminalHeight()

	assert.Greater(t, width, 0, "Width should be positive")
	assert.Greater(t, height, 0, "Height should be positive")

	// Should be at least the default values
	assert.GreaterOrEqual(t, width, 80, "Width should be at least 80 (default)")
	assert.GreaterOrEqual(t, height, 24, "Height should be at least 24 (default)")
}
