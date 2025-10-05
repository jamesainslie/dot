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

// RenderDiagnostics renders diagnostic report as plain text.
func (r *TextRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	// Show overall health
	healthColor := r.scheme.Success
	healthSymbol := "✓"
	if report.OverallHealth == dot.HealthWarnings {
		healthColor = r.scheme.Warning
		healthSymbol = "⚠"
	} else if report.OverallHealth == dot.HealthErrors {
		healthColor = r.scheme.Error
		healthSymbol = "✗"
	}

	fmt.Fprintf(w, "%s%s Health Status: %s%s\n\n", r.colorText(healthColor), healthSymbol, report.OverallHealth.String(), r.resetColor())

	// Show statistics
	fmt.Fprintf(w, "Statistics:\n")
	fmt.Fprintf(w, "  Total Links: %d\n", report.Statistics.TotalLinks)
	fmt.Fprintf(w, "  Managed Links: %d\n", report.Statistics.ManagedLinks)
	if report.Statistics.BrokenLinks > 0 {
		fmt.Fprintf(w, "  %sBroken Links: %d%s\n", r.colorText(r.scheme.Error), report.Statistics.BrokenLinks, r.resetColor())
	}
	if report.Statistics.OrphanedLinks > 0 {
		fmt.Fprintf(w, "  %sOrphaned Links: %d%s\n", r.colorText(r.scheme.Warning), report.Statistics.OrphanedLinks, r.resetColor())
	}
	fmt.Fprintln(w)

	// Show issues
	if len(report.Issues) == 0 {
		fmt.Fprintf(w, "%sNo issues found%s\n", r.colorText(r.scheme.Success), r.resetColor())
		return nil
	}

	fmt.Fprintf(w, "Issues Found: %d\n\n", len(report.Issues))

	for i, issue := range report.Issues {
		severityColor := r.scheme.Info
		severitySymbol := "ℹ"
		if issue.Severity == dot.SeverityWarning {
			severityColor = r.scheme.Warning
			severitySymbol = "⚠"
		} else if issue.Severity == dot.SeverityError {
			severityColor = r.scheme.Error
			severitySymbol = "✗"
		}

		fmt.Fprintf(w, "%d. %s%s %s%s\n", i+1, r.colorText(severityColor), severitySymbol, issue.Severity.String(), r.resetColor())
		fmt.Fprintf(w, "   Type: %s\n", issue.Type.String())
		if issue.Path != "" {
			fmt.Fprintf(w, "   Path: %s\n", issue.Path)
		}
		fmt.Fprintf(w, "   %s\n", issue.Message)
		if issue.Suggestion != "" {
			fmt.Fprintf(w, "   %sSuggestion:%s %s\n", r.colorText(r.scheme.Info), r.resetColor(), issue.Suggestion)
		}
		fmt.Fprintln(w)
	}

	return nil
}
