package pretty

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultListConfig(t *testing.T) {
	config := DefaultListConfig()

	assert.Equal(t, 2, config.Indent, "Default indent should be 2")
}

func TestNewListWriter(t *testing.T) {
	tests := []struct {
		name  string
		style ListStyle
	}{
		{"bullet", StyleBullet},
		{"tree", StyleTree},
		{"numbered", StyleNumbered},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultListConfig()
			lw := NewListWriter(tt.style, config)

			require.NotNil(t, lw)
			assert.Equal(t, tt.style, lw.style)
		})
	}
}

func TestListWriter_AppendItem(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleBullet, config)

	lw.AppendItem("Item 1")
	lw.AppendItem("Item 2")

	output := lw.RenderString()

	assert.Contains(t, output, "Item 1")
	assert.Contains(t, output, "Item 2")
}

func TestListWriter_AppendItems(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleBullet, config)

	items := []interface{}{"First", "Second", "Third"}
	lw.AppendItems(items)

	output := lw.RenderString()

	assert.Contains(t, output, "First")
	assert.Contains(t, output, "Second")
	assert.Contains(t, output, "Third")
}

func TestListWriter_IndentUnIndent(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleBullet, config)

	lw.AppendItem("Level 1")
	lw.Indent()
	lw.AppendItem("Level 2")
	lw.Indent()
	lw.AppendItem("Level 3")
	lw.UnIndent()
	lw.AppendItem("Back to Level 2")
	lw.UnIndent()
	lw.AppendItem("Back to Level 1")

	output := lw.RenderString()

	// Check all items are present
	assert.Contains(t, output, "Level 1")
	assert.Contains(t, output, "Level 2")
	assert.Contains(t, output, "Level 3")
	assert.Contains(t, output, "Back to Level 2")
	assert.Contains(t, output, "Back to Level 1")
}

func TestListWriter_Render(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleBullet, config)

	lw.AppendItem("Test Item")

	var buf bytes.Buffer
	lw.Render(&buf)

	output := buf.String()
	assert.Contains(t, output, "Test Item")
}

func TestListWriter_StyleBullet(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleBullet, config)

	lw.AppendItem("Bullet Item")

	output := lw.RenderString()

	assert.Contains(t, output, "Bullet Item")
	// Verify output is non-empty (bullet character may vary)
	assert.NotEmpty(t, output)
}

func TestListWriter_StyleTree(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleTree, config)

	lw.AppendItem("Root")
	lw.Indent()
	lw.AppendItem("Child")

	output := lw.RenderString()

	assert.Contains(t, output, "Root")
	assert.Contains(t, output, "Child")
}

func TestListWriter_StyleNumbered(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleNumbered, config)

	lw.AppendItem("First Item")
	lw.AppendItem("Second Item")
	lw.AppendItem("Third Item")

	output := lw.RenderString()

	assert.Contains(t, output, "First Item")
	assert.Contains(t, output, "Second Item")
	assert.Contains(t, output, "Third Item")
}

func TestListWriter_ColorEnabled(t *testing.T) {
	t.Run("with colors", func(t *testing.T) {
		config := DefaultListConfig()
		config.ColorEnabled = true
		lw := NewListWriter(StyleBullet, config)

		lw.AppendItem("Colored Item")

		output := lw.RenderString()
		assert.Contains(t, output, "Colored Item")
	})

	t.Run("without colors", func(t *testing.T) {
		config := DefaultListConfig()
		config.ColorEnabled = false
		lw := NewListWriter(StyleBullet, config)

		lw.AppendItem("Plain Item")

		output := lw.RenderString()
		assert.Contains(t, output, "Plain Item")
		// Should not contain ANSI codes
		assert.NotContains(t, output, "\033[")
	})
}

func TestListWriter_CustomIndent(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	config.Indent = 4
	lw := NewListWriter(StyleBullet, config)

	lw.AppendItem("Root")
	lw.Indent()
	lw.AppendItem("Child")

	output := lw.RenderString()

	assert.Contains(t, output, "Root")
	assert.Contains(t, output, "Child")
}

func TestListWriter_EmptyList(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleBullet, config)

	// Don't add any items
	output := lw.RenderString()

	// Empty list should produce empty or minimal output
	assert.Equal(t, "", strings.TrimSpace(output))
}

func TestListWriter_MultipleIndentLevels(t *testing.T) {
	config := DefaultListConfig()
	config.ColorEnabled = false
	lw := NewListWriter(StyleTree, config)

	lw.AppendItem("Level 0")
	lw.Indent()
	lw.AppendItem("Level 1-A")
	lw.Indent()
	lw.AppendItem("Level 2")
	lw.UnIndent()
	lw.AppendItem("Level 1-B")
	lw.UnIndent()
	lw.AppendItem("Back to Level 0")

	output := lw.RenderString()

	assert.Contains(t, output, "Level 0")
	assert.Contains(t, output, "Level 1-A")
	assert.Contains(t, output, "Level 2")
	assert.Contains(t, output, "Level 1-B")
	assert.Contains(t, output, "Back to Level 0")
}
