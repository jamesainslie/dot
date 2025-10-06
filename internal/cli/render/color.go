package render

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// Color represents a terminal color.
type Color struct {
	ANSI string
}

// ColorScheme defines a color palette.
type ColorScheme struct {
	Error   Color
	Warning Color
	Success Color
	Info    Color
	Dim     Color
}

// Predefined colors using ANSI codes.
var (
	// Basic colors
	colorRed    = Color{ANSI: "\033[31m"}
	colorYellow = Color{ANSI: "\033[33m"}
	colorGreen  = Color{ANSI: "\033[32m"}
	colorBlue   = Color{ANSI: "\033[34m"}
	colorGray   = Color{ANSI: "\033[90m"}
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"

	// Default color scheme with accessible colors
	DefaultScheme = ColorScheme{
		Error:   colorRed,
		Warning: colorYellow,
		Success: colorGreen,
		Info:    colorBlue,
		Dim:     colorGray,
	}

	// No-color scheme for plain text
	NoColorScheme = ColorScheme{
		Error:   Color{ANSI: ""},
		Warning: Color{ANSI: ""},
		Success: Color{ANSI: ""},
		Info:    Color{ANSI: ""},
		Dim:     Color{ANSI: ""},
	}
)

// Apply applies the color to text.
func (c Color) Apply(text string) string {
	if c.ANSI == "" {
		return text
	}
	return c.ANSI + text + colorReset
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

	// Check TERM environment variable
	termEnv := os.Getenv("TERM")
	if termEnv == "" || termEnv == "dumb" {
		return false
	}

	// Check for color support
	if strings.Contains(termEnv, "color") || strings.Contains(termEnv, "256") || strings.Contains(termEnv, "xterm") {
		return true
	}

	return true
}

// GetScheme returns the appropriate color scheme based on environment.
func GetScheme() ColorScheme {
	if ShouldUseColor() {
		return DefaultScheme
	}
	return NoColorScheme
}
