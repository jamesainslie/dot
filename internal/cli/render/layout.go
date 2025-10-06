package render

import (
	"fmt"
	"strings"

	"golang.org/x/term"
)

// Layout provides text layout utilities.
type Layout struct {
	width int
}

// NewLayout creates a layout with the given width.
func NewLayout(width int) *Layout {
	return &Layout{width: width}
}

// NewLayoutAuto creates a layout with automatic width detection.
func NewLayoutAuto() *Layout {
	width := 80 // Default
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		width = w
	}
	return &Layout{width: width}
}

// Width returns the layout width.
func (l *Layout) Width() int {
	return l.width
}

// Wrap wraps text to terminal width.
func (l *Layout) Wrap(text string, indent int) string {
	return wrapText(text, l.width, indent)
}

// Indent adds indentation to text.
func (l *Layout) Indent(text string, level int) string {
	indentStr := strings.Repeat(" ", level*2)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indentStr + line
		}
	}
	return strings.Join(lines, "\n")
}

// Box draws a box around text.
func (l *Layout) Box(text string, title string) string {
	lines := strings.Split(text, "\n")
	maxLen := len(title) + 4

	for _, line := range lines {
		lineLen := stripANSI(line)
		if len(lineLen) > maxLen {
			maxLen = len(lineLen)
		}
	}

	if maxLen > l.width-4 {
		maxLen = l.width - 4
	}

	var b strings.Builder

	// Top border
	b.WriteString("┌")
	if title != "" {
		b.WriteString("─ ")
		b.WriteString(title)
		b.WriteString(" ")
		pad := max(0, maxLen-len(title)-2)
		b.WriteString(strings.Repeat("─", pad))
	} else {
		b.WriteString(strings.Repeat("─", maxLen))
	}
	b.WriteString("┐\n")

	// Content
	for _, line := range lines {
		b.WriteString("│ ")
		b.WriteString(line)
		// Pad to max length (accounting for ANSI codes)
		plainLen := len(stripANSI(line))
		if plainLen < maxLen {
			b.WriteString(strings.Repeat(" ", maxLen-plainLen))
		}
		b.WriteString(" │\n")
	}

	// Bottom border
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", maxLen))
	b.WriteString("┘")

	return b.String()
}

// Table formats data as a table.
func (l *Layout) Table(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	colWidths := l.calculateColumnWidths(headers, rows)
	colWidths = l.adjustColumnWidths(colWidths)

	var b strings.Builder
	l.writeTableHeader(&b, headers, colWidths)
	l.writeTableSeparator(&b, headers, colWidths)
	l.writeTableRows(&b, rows, colWidths)

	return strings.TrimRight(b.String(), "\n")
}

func (l *Layout) calculateColumnWidths(headers []string, rows [][]string) []int {
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellLen := len(stripANSI(cell))
				if cellLen > colWidths[i] {
					colWidths[i] = cellLen
				}
			}
		}
	}

	return colWidths
}

func (l *Layout) adjustColumnWidths(colWidths []int) []int {
	// Calculate separators correctly: n columns need n-1 separators
	separators := (len(colWidths) - 1) * 3
	totalWidth := separators
	for _, w := range colWidths {
		totalWidth += w
	}

	if totalWidth <= l.width {
		return colWidths
	}

	// Proportionally reduce column widths
	excess := totalWidth - l.width
	adjusted := make([]int, len(colWidths))
	for i, w := range colWidths {
		reduction := (excess * w) / totalWidth
		adjusted[i] = w - reduction
		if adjusted[i] < 3 {
			adjusted[i] = 3
		}
	}
	return adjusted
}

func (l *Layout) writeTableHeader(b *strings.Builder, headers []string, colWidths []int) {
	for i, header := range headers {
		if i > 0 {
			b.WriteString(" │ ")
		}
		b.WriteString(padRight(header, colWidths[i]))
	}
	b.WriteString("\n")
}

func (l *Layout) writeTableSeparator(b *strings.Builder, headers []string, colWidths []int) {
	for i := range headers {
		if i > 0 {
			b.WriteString("─┼─")
		}
		b.WriteString(strings.Repeat("─", colWidths[i]))
	}
	b.WriteString("\n")
}

func (l *Layout) writeTableRows(b *strings.Builder, rows [][]string, colWidths []int) {
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(colWidths) {
				break
			}
			if i > 0 {
				b.WriteString(" │ ")
			}
			l.writeTableCell(b, cell, colWidths[i])
		}
		b.WriteString("\n")
	}
}

func (l *Layout) writeTableCell(b *strings.Builder, cell string, width int) {
	plainCell := stripANSI(cell)
	padding := width - len(plainCell)
	b.WriteString(cell)
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
}

// wrapText wraps text to the specified width with indentation.
func wrapText(text string, width, indent int) string {
	if width <= 0 {
		width = 80
	}

	effectiveWidth := width - indent
	if effectiveWidth < 20 {
		effectiveWidth = 20
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	indentStr := strings.Repeat(" ", indent)
	currentLineLen := 0

	for i, word := range words {
		wordLen := len(word)

		if i == 0 {
			b.WriteString(word)
			currentLineLen = wordLen
		} else if currentLineLen+1+wordLen <= effectiveWidth {
			b.WriteString(" ")
			b.WriteString(word)
			currentLineLen += 1 + wordLen
		} else {
			b.WriteString("\n")
			b.WriteString(indentStr)
			b.WriteString(word)
			currentLineLen = wordLen
		}
	}

	return b.String()
}

// padRight pads string to width (right-aligned).
func padRight(s string, width int) string {
	sLen := len(stripANSI(s))
	if sLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sLen)
}

// stripANSI removes ANSI escape codes from string for length calculation.
func stripANSI(s string) string {
	// Simple ANSI stripping - matches \033[...m patterns
	result := ""
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // Skip the '['
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(s[i])
	}
	return result
}

// Divider returns a horizontal divider line.
func (l *Layout) Divider(char string) string {
	if char == "" {
		char = "─"
	}
	return strings.Repeat(char, l.width)
}

// Center centers text within the layout width.
func (l *Layout) Center(text string) string {
	textLen := len(stripANSI(text))
	if textLen >= l.width {
		return text
	}
	padding := (l.width - textLen) / 2
	return strings.Repeat(" ", padding) + text
}

// List formats items as a bulleted list.
func (l *Layout) List(items []string, bullet string) string {
	if bullet == "" {
		bullet = "•"
	}

	var b strings.Builder
	for i, item := range items {
		b.WriteString(bullet)
		b.WriteString(" ")
		b.WriteString(item)
		if i < len(items)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// NumberedList formats items as a numbered list.
func (l *Layout) NumberedList(items []string) string {
	var b strings.Builder
	for i, item := range items {
		b.WriteString(fmt.Sprintf("%d. %s", i+1, item))
		if i < len(items)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
