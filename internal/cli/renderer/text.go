package renderer

import (
	"fmt"
	"io"

	"github.com/jamesainslie/dot/pkg/dot"
)

// TextRenderer renders output as human-readable plain text.
type TextRenderer struct {
	colorize bool
	scheme   ColorScheme
	width    int
}

// RenderStatus renders installation status as plain text.
func (r *TextRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	if len(status.Packages) == 0 {
		fmt.Fprintln(w, "No packages installed")
		return nil
	}

	for _, pkg := range status.Packages {
		fmt.Fprintf(w, "%s%s%s\n", r.colorText(r.scheme.Info), pkg.Name, r.resetColor())
		fmt.Fprintf(w, "  Links: %d\n", pkg.LinkCount)
		fmt.Fprintf(w, "  Installed: %s\n", formatDuration(pkg.InstalledAt))

		if len(pkg.Links) > 0 {
			fmt.Fprintf(w, "  Files:\n")
			displayCount := 5
			for i, link := range pkg.Links {
				if i >= displayCount {
					remaining := len(pkg.Links) - displayCount
					fmt.Fprintf(w, "    ... and %d more\n", remaining)
					break
				}
				fmt.Fprintf(w, "    %s\n", link)
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}

func (r *TextRenderer) colorText(color string) string {
	if r.colorize && color != "" {
		return color
	}
	return ""
}

func (r *TextRenderer) resetColor() string {
	if r.colorize {
		return "\033[0m"
	}
	return ""
}
