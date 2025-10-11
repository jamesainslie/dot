package pretty

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/list"
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

// ListWriter wraps go-pretty list.Writer with consistent styling.
type ListWriter struct {
	writer list.Writer
	style  ListStyle
	config ListConfig
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
	l := list.NewWriter()

	// Apply style
	applyListStyle(l, style, config.ColorEnabled)

	// Set indent
	if config.Indent > 0 {
		l.SetStyle(getListStyle(style, config.ColorEnabled, config.Indent))
	}

	return &ListWriter{
		writer: l,
		style:  style,
		config: config,
	}
}

// AppendItem adds an item to the list.
func (w *ListWriter) AppendItem(item interface{}) {
	w.writer.AppendItem(item)
}

// AppendItems adds multiple items to the list.
func (w *ListWriter) AppendItems(items []interface{}) {
	w.writer.AppendItems(items)
}

// Indent increases the indentation level for subsequent items.
func (w *ListWriter) Indent() {
	w.writer.Indent()
}

// UnIndent decreases the indentation level.
func (w *ListWriter) UnIndent() {
	w.writer.UnIndent()
}

// Render outputs the list to the given writer.
func (w *ListWriter) Render(out io.Writer) {
	w.writer.SetOutputMirror(out)
	w.writer.Render()
}

// RenderString returns the list as a string.
func (w *ListWriter) RenderString() string {
	return w.writer.Render()
}

// applyListStyle applies the selected style to the list writer.
func applyListStyle(l list.Writer, style ListStyle, colorEnabled bool) {
	l.SetStyle(getListStyle(style, colorEnabled, 2))
}

// getListStyle returns the appropriate list style.
func getListStyle(style ListStyle, colorEnabled bool, indent int) list.Style {
	var baseStyle list.Style

	switch style {
	case StyleTree:
		baseStyle = list.StyleConnectedRounded
	case StyleNumbered:
		baseStyle = list.StyleDefault
		// Style numbered doesn't need format customization - uses default numbering
	default:
		baseStyle = list.StyleBulletCircle
	}

	// Customize indentation
	baseStyle.LinePrefix = ""

	// Note: Color customization for list items would require direct manipulation
	// of the list output, which we'll skip for now to keep the API simple

	return baseStyle
}
