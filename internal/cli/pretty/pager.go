package pretty

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Pager handles paginated output for long content.
type Pager struct {
	output   io.Writer
	pageSize int
}

// PagerConfig holds configuration for the pager.
type PagerConfig struct {
	// PageSize is the number of lines per page (0 = auto-detect from terminal height)
	PageSize int
	// Output is where paginated content is written
	Output io.Writer
}

// DefaultPagerConfig returns sensible defaults for pagination.
func DefaultPagerConfig() PagerConfig {
	return PagerConfig{
		PageSize: GetTerminalHeight() - 2, // Leave room for prompt
		Output:   os.Stdout,
	}
}

// NewPager creates a new pager with the given configuration.
func NewPager(config PagerConfig) *Pager {
	pageSize := config.PageSize
	if pageSize <= 0 {
		pageSize = GetTerminalHeight() - 2
	}

	// Ensure minimum page size
	if pageSize < 5 {
		pageSize = 20
	}

	return &Pager{
		output:   config.Output,
		pageSize: pageSize,
	}
}

// Page displays content with pagination if in an interactive terminal.
// If not interactive (piped or redirected), content is displayed without pagination.
func (p *Pager) Page(content string) error {
	lines := strings.Split(content, "\n")

	// If not interactive (piped output or not a terminal), just print everything
	if !IsInteractive() {
		_, err := fmt.Fprint(p.output, content)
		return err
	}

	// If content fits on screen, just print it
	if len(lines) <= p.pageSize {
		_, err := fmt.Fprint(p.output, content)
		return err
	}

	// Otherwise, paginate
	for i := 0; i < len(lines); i += p.pageSize {
		end := i + p.pageSize
		if end > len(lines) {
			end = len(lines)
		}

		// Print this page
		pageContent := strings.Join(lines[i:end], "\n")
		fmt.Fprint(p.output, pageContent)

		// If this is the last page, we're done
		if end >= len(lines) {
			fmt.Fprintln(p.output)
			break
		}

		// Show continuation prompt
		remaining := len(lines) - end
		if !p.promptContinue(remaining) {
			fmt.Fprintln(p.output)
			return nil
		}
		fmt.Fprintln(p.output)
	}

	return nil
}

// promptContinue shows a continuation prompt and waits for user input.
// Returns true if user wants to continue, false otherwise.
func (p *Pager) promptContinue(remainingLines int) bool {
	// Use dim color for subtle prompt
	prompt := Dim(fmt.Sprintf("\n--- More (%d lines remaining, press Enter to continue or q to quit) ---", remainingLines))
	fmt.Fprint(p.output, prompt)

	// Read single character from stdin
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	// Check if user wants to quit
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "q" || input == "quit" {
		return false
	}

	return true
}

// PageLines is a convenience method for paging a slice of strings.
func (p *Pager) PageLines(lines []string) error {
	return p.Page(strings.Join(lines, "\n"))
}
