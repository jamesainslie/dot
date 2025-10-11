package pretty

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

// Color palette using 256-color codes for subtle, professional output.
const (
	// Muted green for success states
	colorSuccess = "\033[38;5;71m"
	// Muted gold for warnings
	colorWarning = "\033[38;5;179m"
	// Muted red for errors
	colorError = "\033[38;5;167m"
	// Muted blue for informational messages
	colorInfo = "\033[38;5;110m"
	// Muted cyan for accents
	colorAccent = "\033[38;5;109m"
	// Gray for dimmed text
	colorDim = "\033[38;5;245m"
	// Reset to default
	colorReset = "\033[0m"
)

// Success colors text in muted green.
func Success(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return colorSuccess + s + colorReset
}

// Warning colors text in muted gold.
func Warning(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return colorWarning + s + colorReset
}

// Error colors text in muted red.
func Error(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return colorError + s + colorReset
}

// Info colors text in muted blue.
func Info(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return colorInfo + s + colorReset
}

// Accent colors text in muted cyan.
func Accent(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return colorAccent + s + colorReset
}

// Dim colors text in gray.
func Dim(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return colorDim + s + colorReset
}

// Bold makes text bold.
func Bold(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return text.Bold.Sprint(s)
}

// Underline underlines text.
func Underline(s string) string {
	if !ShouldUseColor() {
		return s
	}
	return text.Underline.Sprint(s)
}

// Truncate shortens text to maxLen, adding ellipsis if needed.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// AlignLeft pads string to width with spaces on the right.
func AlignLeft(s string, width int) string {
	return text.AlignLeft.Apply(s, width)
}

// AlignRight pads string to width with spaces on the left.
func AlignRight(s string, width int) string {
	return text.AlignRight.Apply(s, width)
}

// AlignCenter centers string within width.
func AlignCenter(s string, width int) string {
	return text.AlignCenter.Apply(s, width)
}

// WrapText wraps text to specified width.
func WrapText(s string, width int) string {
	return text.WrapText(s, width)
}

// Box draws a simple box around text with optional title.
func Box(content string, title string) string {
	if title == "" {
		return fmt.Sprintf("┌─────────────────────────────────────────┐\n%s\n└─────────────────────────────────────────┘", content)
	}
	return fmt.Sprintf("┌─ %s ──────────────────────────────────┐\n%s\n└─────────────────────────────────────────┘", title, content)
}

// Indent adds leading spaces to each line.
func Indent(s string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if len(line) > 0 {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
