// Package pretty provides consistent, professional CLI output formatting using go-pretty libraries.
package pretty

import (
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

// TableStyle defines the visual style for tables.
type TableStyle int

const (
	// StyleBordered uses rounded borders with full table structure.
	StyleBordered TableStyle = iota
	// StyleLight uses light borders for a clean look.
	StyleLight
	// StyleMinimal uses no borders, just spacing (default).
	StyleMinimal
	// StyleCompact uses dense formatting for large datasets.
	StyleCompact
)

// TableConfig holds configuration for table rendering.
type TableConfig struct {
	// MaxWidth is the maximum table width (0 = auto-detect from terminal).
	MaxWidth int
	// ColorEnabled controls whether to use colors in output.
	ColorEnabled bool
	// AutoWrap enables automatic text wrapping in columns.
	AutoWrap bool
	// SortColumn is the column index to sort by (-1 = no sorting).
	SortColumn int
	// SortAsc controls sort direction (true = ascending).
	SortAsc bool
}

// DefaultTableConfig returns sensible defaults for table rendering.
func DefaultTableConfig() TableConfig {
	return TableConfig{
		MaxWidth:     GetTerminalWidth(),
		ColorEnabled: ShouldUseColor(),
		AutoWrap:     true,
		SortColumn:   -1,
		SortAsc:      true,
	}
}

// TableWriter wraps go-pretty table.Writer with consistent styling.
type TableWriter struct {
	writer table.Writer
	config TableConfig
	style  TableStyle
}

// NewTableWriter creates a new table writer with the given style and config.
func NewTableWriter(style TableStyle, config TableConfig) *TableWriter {
	t := table.NewWriter()

	// Apply style
	applyTableStyle(t, style, config.ColorEnabled)

	// Configure column constraints based on terminal width
	if config.MaxWidth > 0 {
		t.SetAllowedRowLength(config.MaxWidth)
	}

	return &TableWriter{
		writer: t,
		config: config,
		style:  style,
	}
}

// SetHeader sets the table header row.
func (w *TableWriter) SetHeader(headers ...interface{}) {
	w.writer.AppendHeader(table.Row(headers))
}

// AppendRow adds a data row to the table.
func (w *TableWriter) AppendRow(row ...interface{}) {
	w.writer.AppendRow(table.Row(row))
}

// AppendRows adds multiple data rows.
func (w *TableWriter) AppendRows(rows [][]interface{}) {
	for _, row := range rows {
		w.writer.AppendRow(table.Row(row))
	}
}

// AppendSeparator adds a visual separator line.
func (w *TableWriter) AppendSeparator() {
	w.writer.AppendSeparator()
}

// SetAutoIndex enables row numbering.
func (w *TableWriter) SetAutoIndex(enabled bool) {
	w.writer.SetAutoIndex(enabled)
}

// SetColumnConfig sets configuration for a specific column.
func (w *TableWriter) SetColumnConfig(columnNumber int, config table.ColumnConfig) {
	configs := []table.ColumnConfig{config}
	configs[0].Number = columnNumber
	w.writer.SetColumnConfigs(configs)
}

// SortBy sorts the table by the configured column.
func (w *TableWriter) SortBy(columnNumber int, ascending bool) {
	mode := table.Asc
	if !ascending {
		mode = table.Dsc
	}
	w.writer.SortBy([]table.SortBy{{Number: columnNumber, Mode: mode}})
}

// Render outputs the table to the given writer.
func (w *TableWriter) Render(out io.Writer) {
	w.writer.SetOutputMirror(out)
	w.writer.Render()
}

// RenderString returns the table as a string.
func (w *TableWriter) RenderString() string {
	return w.writer.Render()
}

// applyTableStyle applies the selected style to the table writer.
func applyTableStyle(t table.Writer, style TableStyle, colorEnabled bool) {
	var baseStyle table.Style

	switch style {
	case StyleBordered:
		baseStyle = createBorderedStyle(colorEnabled)
	case StyleLight:
		baseStyle = createLightStyle(colorEnabled)
	case StyleCompact:
		baseStyle = createCompactStyle(colorEnabled)
	default:
		baseStyle = createMinimalStyle(colorEnabled)
	}

	t.SetStyle(baseStyle)
}

// createBorderedStyle creates a rounded border style with full table structure.
func createBorderedStyle(colorEnabled bool) table.Style {
	style := table.StyleRounded

	if colorEnabled {
		style.Color.Header = text.Colors{text.FgHiBlack, text.Bold}
		style.Color.Border = text.Colors{text.FgHiBlack}
		style.Color.Separator = text.Colors{text.FgHiBlack}
	} else {
		style.Color = table.ColorOptions{}
	}

	style.Options.SeparateRows = false
	style.Options.SeparateColumns = true
	style.Options.DrawBorder = true

	return style
}

// createLightStyle creates a light border style for clean look.
func createLightStyle(colorEnabled bool) table.Style {
	style := table.StyleLight

	if colorEnabled {
		style.Color.Header = text.Colors{text.FgHiBlack, text.Bold}
		style.Color.Border = text.Colors{text.FgHiBlack}
		style.Color.Separator = text.Colors{text.FgHiBlack}
	} else {
		style.Color = table.ColorOptions{}
	}

	style.Options.SeparateRows = false
	style.Options.SeparateColumns = true
	style.Options.DrawBorder = true

	return style
}

// createMinimalStyle creates a borderless style with just spacing.
func createMinimalStyle(colorEnabled bool) table.Style {
	style := table.Style{
		Name: "Minimal",
		Box:  table.BoxStyle{},
		Format: table.FormatOptions{
			Header: text.FormatDefault,
			Row:    text.FormatDefault,
		},
		Options: table.Options{
			SeparateRows:    false,
			SeparateColumns: true,
			SeparateHeader:  true,
			DrawBorder:      false,
		},
	}

	// Add subtle header separator
	style.Box.MiddleHorizontal = "-"
	style.Box.MiddleVertical = " "
	style.Box.MiddleSeparator = " "

	if colorEnabled {
		style.Color.Header = text.Colors{text.FgHiBlack, text.Bold}
		style.Color.Row = text.Colors{}
	} else {
		style.Color = table.ColorOptions{}
	}

	return style
}

// createCompactStyle creates a dense style for large datasets.
func createCompactStyle(colorEnabled bool) table.Style {
	style := table.StyleLight
	style.Name = "Compact"

	// Remove all spacing
	style.Box.PaddingLeft = ""
	style.Box.PaddingRight = " "

	if colorEnabled {
		style.Color.Header = text.Colors{text.FgHiBlack, text.Bold}
	} else {
		style.Color = table.ColorOptions{}
	}

	style.Options.SeparateRows = false
	style.Options.SeparateColumns = true
	style.Options.DrawBorder = false

	return style
}

// ShouldUseColor determines if color output should be enabled.
func ShouldUseColor() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	return true
}

// GetTerminalWidth returns the terminal width in columns.
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // Default fallback
	}
	return width
}

// GetTerminalHeight returns the terminal height in lines.
func GetTerminalHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || height <= 0 {
		return 24 // Default fallback
	}
	return height
}

// IsInteractive checks if we're in an interactive terminal session.
func IsInteractive() bool {
	// Check if both stdin and stdout are terminals
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
