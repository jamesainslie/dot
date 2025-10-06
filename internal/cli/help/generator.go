package help

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Generator creates help text for commands.
type Generator struct {
	width int
}

// NewGenerator creates a new help generator.
func NewGenerator() *Generator {
	width := 80 // Default
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		width = w
	}
	return &Generator{
		width: width,
	}
}

// Generate creates complete help text.
func (g *Generator) Generate(cmd *cobra.Command) string {
	var b strings.Builder

	// Command name and description
	if cmd.Short != "" {
		b.WriteString(cmd.Short)
		b.WriteString("\n\n")
	}

	if cmd.Long != "" {
		b.WriteString(g.wrap(cmd.Long, 0))
		b.WriteString("\n\n")
	}

	// Usage
	usage := g.GenerateUsage(cmd)
	if usage != "" {
		b.WriteString("Usage:\n")
		b.WriteString("  ")
		b.WriteString(usage)
		b.WriteString("\n\n")
	}

	// Flags
	if cmd.HasAvailableFlags() {
		b.WriteString("Flags:\n")
		b.WriteString(cmd.LocalFlags().FlagUsages())
		b.WriteString("\n")
	}

	// Global flags
	if cmd.HasAvailableInheritedFlags() {
		b.WriteString("Global Flags:\n")
		b.WriteString(cmd.InheritedFlags().FlagUsages())
		b.WriteString("\n")
	}

	// Examples
	if cmd.Example != "" {
		b.WriteString(cmd.Example)
		b.WriteString("\n")
	}

	// Subcommands
	if cmd.HasAvailableSubCommands() {
		b.WriteString("Available Commands:\n")
		for _, subCmd := range cmd.Commands() {
			if !subCmd.Hidden {
				b.WriteString(fmt.Sprintf("  %-15s %s\n", subCmd.Name(), subCmd.Short))
			}
		}
		b.WriteString("\n")
	}

	// See also
	if cmd.HasParent() || cmd.HasAvailableSubCommands() {
		seeAlso := g.GenerateSeeAlso(cmd)
		if seeAlso != "" {
			b.WriteString("See Also:\n")
			b.WriteString(seeAlso)
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

// GenerateUsage creates usage string.
func (g *Generator) GenerateUsage(cmd *cobra.Command) string {
	usage := cmd.UseLine()
	return usage
}

// GenerateExamples creates examples section.
func (g *Generator) GenerateExamples(cmd *cobra.Command) string {
	if cmd.Example == "" {
		return ""
	}
	return cmd.Example
}

// GenerateSeeAlso creates see also section.
func (g *Generator) GenerateSeeAlso(cmd *cobra.Command) string {
	var related []string

	if cmd.HasParent() {
		related = append(related, cmd.Parent().CommandPath())
	}

	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden {
			related = append(related, subCmd.CommandPath())
		}
	}

	if len(related) == 0 {
		return ""
	}

	return "  " + strings.Join(related, ", ")
}

// wrap wraps text to terminal width.
func (g *Generator) wrap(text string, indent int) string {
	effectiveWidth := g.width - indent
	if effectiveWidth < 40 {
		effectiveWidth = 40
	}

	lines := strings.Split(text, "\n")
	var b strings.Builder
	indentStr := strings.Repeat(" ", indent)

	for lineIdx, line := range lines {
		if lineIdx > 0 {
			b.WriteString("\n")
		}

		// Empty line - preserve it
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Process this line with word wrapping
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLen := 0
		for i, word := range words {
			wordLen := len(word)

			if i == 0 {
				b.WriteString(word)
				currentLen = wordLen
			} else if currentLen+1+wordLen <= effectiveWidth {
				b.WriteString(" ")
				b.WriteString(word)
				currentLen += 1 + wordLen
			} else {
				b.WriteString("\n")
				b.WriteString(indentStr)
				b.WriteString(word)
				currentLen = wordLen
			}
		}
	}

	return b.String()
}
