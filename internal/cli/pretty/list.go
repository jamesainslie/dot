package pretty

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ListStyle defines the visual style for lists.
type ListStyle int

const (
	// StyleBullet uses simple bullet points.
	StyleBullet ListStyle = iota
	// StyleTree uses tree structure with branches.
	StyleTree
	// StyleNumbered uses numbered items.
	StyleNumbered
)

// ListWriter provides list rendering with lipgloss styling.
type ListWriter struct {
	items       []listItem
	style       ListStyle
	config      ListConfig
	indentLevel int
}

// listItem represents a single list item with its indentation level.
type listItem struct {
	content string
	level   int
}

// ListConfig holds configuration for list rendering.
type ListConfig struct {
	// ColorEnabled controls whether to use colors in output.
	ColorEnabled bool
	// Indent is the number of spaces per indentation level.
	Indent int
}

// DefaultListConfig returns sensible defaults for list rendering.
func DefaultListConfig() ListConfig {
	return ListConfig{
		ColorEnabled: ShouldUseColor(),
		Indent:       2,
	}
}

// NewListWriter creates a new list writer with the given style and config.
func NewListWriter(style ListStyle, config ListConfig) *ListWriter {
	return &ListWriter{
		items:       []listItem{},
		style:       style,
		config:      config,
		indentLevel: 0,
	}
}

// AppendItem adds an item to the list.
func (w *ListWriter) AppendItem(item interface{}) {
	w.items = append(w.items, listItem{
		content: fmt.Sprintf("%v", item),
		level:   w.indentLevel,
	})
}

// AppendItems adds multiple items to the list.
func (w *ListWriter) AppendItems(items []interface{}) {
	for _, item := range items {
		w.AppendItem(item)
	}
}

// Indent increases the indentation level for subsequent items.
func (w *ListWriter) Indent() {
	w.indentLevel++
}

// UnIndent decreases the indentation level.
func (w *ListWriter) UnIndent() {
	if w.indentLevel > 0 {
		w.indentLevel--
	}
}

// Render outputs the list to the given writer.
func (w *ListWriter) Render(out io.Writer) {
	fmt.Fprint(out, w.RenderString())
}

// RenderString returns the list as a string.
func (w *ListWriter) RenderString() string {
	var result strings.Builder

	itemStyle := lipgloss.NewStyle()
	if w.config.ColorEnabled {
		itemStyle = itemStyle.Foreground(lipgloss.Color("252"))
	}

	for i, item := range w.items {
		// Render indentation
		indent := strings.Repeat(" ", item.level*w.config.Indent)
		result.WriteString(indent)

		// Render prefix based on style
		prefix := w.getPrefix(item.level, i)
		result.WriteString(prefix)

		// Render content
		result.WriteString(itemStyle.Render(item.content))
		result.WriteString("\n")
	}

	return result.String()
}

// getPrefix returns the appropriate prefix for the given item.
func (w *ListWriter) getPrefix(level, index int) string {
	prefixStyle := lipgloss.NewStyle()
	if w.config.ColorEnabled {
		prefixStyle = prefixStyle.Foreground(lipgloss.Color("240"))
	}

	switch w.style {
	case StyleTree:
		if level == 0 {
			return prefixStyle.Render("├─ ")
		}
		return prefixStyle.Render("├─ ")
	case StyleNumbered:
		return prefixStyle.Render(fmt.Sprintf("%d. ", index+1))
	default: // StyleBullet
		return prefixStyle.Render("• ")
	}
}
