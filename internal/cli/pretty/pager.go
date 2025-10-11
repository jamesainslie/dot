package pretty

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
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
// Supports spacebar/Enter for next page, up/down arrows for line scrolling, and 'q' to quit.
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

	// Interactive pagination with keyboard controls
	return p.pageInteractive(lines)
}

// pageInteractive handles interactive pagination with keyboard controls.
func (p *Pager) pageInteractive(lines []string) error {
	position := 0
	maxPos := len(lines)

	for position < maxPos {
		// Calculate end position for current view
		end := position + p.pageSize
		if end > maxPos {
			end = maxPos
		}

		// Display current page
		pageContent := strings.Join(lines[position:end], "\n")
		fmt.Fprint(p.output, pageContent)

		// Show status line
		remaining := maxPos - end
		if remaining > 0 {
			p.showStatusLine(position, end, maxPos)

			// Get next action from user
			action := p.getKeyPress()
			
			// Clear status line before showing next content
			p.clearLine()
			fmt.Fprintln(p.output)

			switch action {
			case actionQuit:
				return nil
			case actionPageDown:
				position = end
			case actionLineDown:
				if position < maxPos-p.pageSize {
					position++
				} else {
					// Can't scroll down further, treat as page down
					position = end
				}
			case actionLineUp:
				if position > 0 {
					position--
				}
				// If can't scroll up, just stay at current position
			}
		} else {
			// Last page, just display and exit
			fmt.Fprintln(p.output)
			break
		}
	}

	return nil
}

// Action represents user input action
type pagerAction int

const (
	actionQuit pagerAction = iota
	actionPageDown
	actionLineUp
	actionLineDown
)

// clearLine clears the current line (used to remove status line).
func (p *Pager) clearLine() {
	// ANSI escape code to move cursor to start of line and clear it
	fmt.Fprint(p.output, "\r\033[K")
}

// showStatusLine displays the pagination status and controls hint.
func (p *Pager) showStatusLine(start, end, total int) {
	percent := (end * 100) / total
	status := fmt.Sprintf("\n%s [%d-%d/%d %d%%] Space/Enter: page down | ↑↓: scroll | q: quit %s",
		Dim("───"),
		start+1,
		end,
		total,
		percent,
		Dim("───"),
	)
	fmt.Fprint(p.output, status)
}

// getKeyPress reads a single keypress from stdin in raw mode.
func (p *Pager) getKeyPress() pagerAction {
	// Get file descriptor for stdin
	fd := int(os.Stdin.Fd())

	// Save current terminal state
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		// Fallback to Enter-only mode if raw mode fails
		return actionPageDown
	}
	defer term.Restore(fd, oldState)

	// Read single key
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return actionPageDown
	}

	// Handle key presses
	switch {
	case buf[0] == 'q' || buf[0] == 'Q':
		return actionQuit
	case buf[0] == ' ' || buf[0] == '\r' || buf[0] == '\n':
		return actionPageDown
	case n == 3 && buf[0] == 27 && buf[1] == 91:
		// Arrow key escape sequence: ESC [ [A-D]
		switch buf[2] {
		case 65: // Up arrow
			return actionLineUp
		case 66: // Down arrow
			return actionLineDown
		}
	}

	return actionPageDown
}

// PageLines is a convenience method for paging a slice of strings.
func (p *Pager) PageLines(lines []string) error {
	return p.Page(strings.Join(lines, "\n"))
}
