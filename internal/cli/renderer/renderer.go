package renderer

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Renderer defines the interface for output formatting.
type Renderer interface {
	RenderStatus(w io.Writer, status dot.Status) error
	RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error
}

// ColorScheme defines semantic colors for terminal output.
type ColorScheme struct {
	Success string
	Warning string
	Error   string
	Info    string
	Muted   string
}

// DefaultColorScheme returns the default color scheme.
// Colors are disabled if NO_COLOR environment variable is set.
func DefaultColorScheme() ColorScheme {
	if os.Getenv("NO_COLOR") != "" {
		return ColorScheme{}
	}

	return ColorScheme{
		Success: "\033[32m", // Green
		Warning: "\033[33m", // Yellow
		Error:   "\033[31m", // Red
		Info:    "\033[36m", // Cyan
		Muted:   "\033[90m", // Gray
	}
}

// NewRenderer creates a new renderer based on the specified format.
func NewRenderer(format string, colorize bool) (Renderer, error) {
	width := getTerminalWidth()
	scheme := DefaultColorScheme()

	if !colorize {
		scheme = ColorScheme{} // Disable colors
	}

	switch format {
	case "text":
		return &TextRenderer{
			colorize: colorize,
			scheme:   scheme,
			width:    width,
		}, nil
	case "json":
		return &JSONRenderer{
			pretty: true,
		}, nil
	case "yaml":
		return &YAMLRenderer{
			indent: 2,
		}, nil
	case "table":
		return &TableRenderer{
			colorize: colorize,
			scheme:   scheme,
			width:    width,
		}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: text, json, yaml, table)", format)
	}
}

// getTerminalWidth returns the width of the terminal, or a default if not available.
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		return 80 // Default fallback
	}
	return width
}

// formatBytes converts bytes to human-readable format.
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	base := 1024.0

	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	exp := int(math.Log(float64(bytes)) / math.Log(base))
	if exp >= len(units) {
		exp = len(units) - 1
	}

	value := float64(bytes) / math.Pow(base, float64(exp))
	return fmt.Sprintf("%.1f %s", value, units[exp])
}

// formatDuration converts a time to a human-readable relative duration.
func formatDuration(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		if duration < 10*time.Second {
			return "just now"
		}
		seconds := int(duration.Seconds())
		return fmt.Sprintf("%d seconds ago", seconds)
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d %s ago", minutes, pluralize(minutes, "minute"))
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d %s ago", hours, pluralize(hours, "hour"))
	}

	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d %s ago", days, pluralize(days, "day"))
}

// truncatePath truncates a path to fit within maxLen, preserving beginning and end.
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	if maxLen < 10 {
		return path // Too short to meaningfully truncate
	}

	// Split path into parts
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		// Not enough parts to truncate meaningfully
		if len(path) > maxLen {
			return path[:maxLen-3] + "..."
		}
		return path
	}

	// Keep first and last parts, truncate middle
	first := parts[0]
	if first == "" && len(parts) > 1 {
		first = "/" + parts[1] // Absolute path
	}
	last := parts[len(parts)-1]

	truncated := first + "/.../" + last
	if len(truncated) <= maxLen {
		return truncated
	}

	// Still too long, truncate the last part
	// Account for "/.../" (5 chars) + "..." (3 chars) = 8 total
	availableLen := maxLen - len(first) - 8
	if availableLen > 0 && len(last) > availableLen {
		result := first + "/.../" + last[:availableLen] + "..."
		if len(result) <= maxLen {
			return result
		}
	}

	// Fallback: return prefix that fits
	if maxLen > 3 {
		return path[:maxLen-3] + "..."
	}
	return path[:maxLen]
}

// pluralize returns the plural form of a word based on count.
func pluralize(count int, word string) string {
	if count == 1 {
		return word
	}

	// Simple pluralization rules
	if strings.HasSuffix(word, "y") && !strings.HasSuffix(word, "ay") && !strings.HasSuffix(word, "ey") {
		return word[:len(word)-1] + "ies"
	}

	return word + "s"
}

// indent adds indentation to each line of text.
func indent(text string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// wrapText wraps text to fit within the specified width.
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				result.WriteString(currentLine)
				result.WriteString("\n")
			}
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}

	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}
