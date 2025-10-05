package errors

import (
	"strings"
)

// Template for structured error messages.
type Template struct {
	Title       string
	Description string
	Details     []string
	Suggestions []string
	Footer      string
}

// Render applies template to produce final message.
func (t *Template) Render(colorEnabled bool, width int) string {
	var b strings.Builder

	t.renderTitle(&b, colorEnabled)
	t.renderDescription(&b, width)
	t.renderDetails(&b, colorEnabled, width)
	t.renderSuggestions(&b, colorEnabled, width)
	t.renderFooter(&b, colorEnabled, width)

	return strings.TrimRight(b.String(), "\n")
}

func (t *Template) renderTitle(b *strings.Builder, colorEnabled bool) {
	if colorEnabled {
		b.WriteString(colorRed + colorBold)
	}
	b.WriteString(t.Title)
	if colorEnabled {
		b.WriteString(colorReset)
	}
	b.WriteString("\n")
}

func (t *Template) renderDescription(b *strings.Builder, width int) {
	if t.Description != "" {
		b.WriteString(wrapText(t.Description, width, 0))
		b.WriteString("\n")
	}
}

func (t *Template) renderDetails(b *strings.Builder, colorEnabled bool, width int) {
	if len(t.Details) == 0 {
		return
	}

	b.WriteString("\n")
	if colorEnabled {
		b.WriteString(colorGray)
	}
	b.WriteString("Details:")
	if colorEnabled {
		b.WriteString(colorReset)
	}
	b.WriteString("\n")

	for _, detail := range t.Details {
		b.WriteString("  ")
		b.WriteString(wrapText(detail, width-2, 2))
		b.WriteString("\n")
	}
}

func (t *Template) renderSuggestions(b *strings.Builder, colorEnabled bool, width int) {
	if len(t.Suggestions) == 0 {
		return
	}

	b.WriteString("\n")
	if colorEnabled {
		b.WriteString(colorBlue + colorBold)
	}
	b.WriteString("Suggestions:")
	if colorEnabled {
		b.WriteString(colorReset)
	}
	b.WriteString("\n")

	for _, suggestion := range t.Suggestions {
		if colorEnabled {
			b.WriteString(colorBlue)
		}
		b.WriteString("  • ")
		if colorEnabled {
			b.WriteString(colorReset)
		}
		b.WriteString(wrapText(suggestion, width-4, 4))
		b.WriteString("\n")
	}
}

func (t *Template) renderFooter(b *strings.Builder, colorEnabled bool, width int) {
	if t.Footer == "" {
		return
	}

	b.WriteString("\n")
	if colorEnabled {
		b.WriteString(colorGray)
	}
	b.WriteString(wrapText(t.Footer, width, 0))
	if colorEnabled {
		b.WriteString(colorReset)
	}
	b.WriteString("\n")
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
			// First word
			b.WriteString(word)
			currentLineLen = wordLen
		} else if currentLineLen+1+wordLen <= effectiveWidth {
			// Word fits on current line
			b.WriteString(" ")
			b.WriteString(word)
			currentLineLen += 1 + wordLen
		} else {
			// Word needs new line
			b.WriteString("\n")
			b.WriteString(indentStr)
			b.WriteString(word)
			currentLineLen = wordLen
		}
	}

	return b.String()
}
