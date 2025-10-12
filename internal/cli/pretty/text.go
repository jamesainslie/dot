package pretty

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// Define lipgloss styles for consistent, professional output.
var (
	// Muted green for success states
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("71"))
	// Muted gold for warnings
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("179"))
	// Muted red for errors
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("167"))
	// Muted blue for informational messages
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	// Muted cyan for accents
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("109"))
	// Gray for dimmed text
	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	// Bold style
	boldStyle = lipgloss.NewStyle().Bold(true)
	// Underline style
	underlineStyle = lipgloss.NewStyle().Underline(true)
)

// Success colors text in muted green.
func Success(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return successStyle.Render(s)
}

// Warning colors text in muted gold.
func Warning(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return warningStyle.Render(s)
}

// Error colors text in muted red.
func Error(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return errorStyle.Render(s)
}

// Info colors text in muted blue.
func Info(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return infoStyle.Render(s)
}

// Accent colors text in muted cyan.
func Accent(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return accentStyle.Render(s)
}

// Dim colors text in gray.
func Dim(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return dimStyle.Render(s)
}

// Bold makes text bold.
func Bold(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return boldStyle.Render(s)
}

// Underline underlines text.
func Underline(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return underlineStyle.Render(s)
}

// Truncate shortens text to maxLen, adding ellipsis if needed.
func Truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		runes := []rune(s)
		if len(runes) > maxLen {
			return string(runes[:maxLen])
		}
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

// AlignLeft pads string to width with spaces on the right.
func AlignLeft(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Left).Render(s)
}

// AlignRight pads string to width with spaces on the left.
func AlignRight(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(s)
}

// AlignCenter centers string within width.
func AlignCenter(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}

// WrapText wraps text to specified width.
func WrapText(s string, width int) string {
	if width <= 0 {
		return s
	}

	var result strings.Builder
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if utf8.RuneCountInString(currentLine)+utf8.RuneCountInString(word)+1 <= width {
			currentLine += " " + word
		} else {
			result.WriteString(currentLine)
			result.WriteString("\n")
			currentLine = word
		}
	}
	result.WriteString(currentLine)
	return result.String()
}

// Box draws a simple box around text with optional title.
func Box(content string, title string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if title != "" {
		boxStyle = boxStyle.BorderTop(true).BorderBottom(true).BorderLeft(true).BorderRight(true)
		return lipgloss.JoinVertical(lipgloss.Left,
			boxStyle.Copy().BorderBottom(false).Render(title),
			boxStyle.Copy().BorderTop(false).Render(content),
		)
	}

	return boxStyle.Render(content)
}

// Indent adds leading spaces to each line.
func Indent(s string, spaces int) string {
	return lipgloss.NewStyle().PaddingLeft(spaces).Render(s)
}
