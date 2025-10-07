package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/jamesainslie/dot/pkg/dot"
)

// TableRenderer renders output as tables.
type TableRenderer struct {
	colorize bool
	scheme   ColorScheme
	width    int
}

// RenderStatus renders installation status as a table.
func (r *TableRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	if len(status.Packages) == 0 {
		fmt.Fprintln(w, "No packages installed")
		return nil
	}

	// Build table
	headers := []string{"Package", "Links", "Installed"}
	rows := make([][]string, 0, len(status.Packages))

	for _, pkg := range status.Packages {
		row := []string{
			pkg.Name,
			fmt.Sprintf("%d", pkg.LinkCount),
			formatDuration(pkg.InstalledAt),
		}
		rows = append(rows, row)
	}

	return r.renderTable(w, headers, rows)
}

func (r *TableRenderer) renderTable(w io.Writer, headers []string, rows [][]string) error {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Render header
	r.renderRow(w, headers, widths, true)
	r.renderSeparator(w, widths)

	// Render rows
	for _, row := range rows {
		r.renderRow(w, row, widths, false)
	}

	return nil
}

func (r *TableRenderer) renderRow(w io.Writer, cells []string, widths []int, header bool) {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		width := widths[i]
		if header && r.colorize {
			parts[i] = fmt.Sprintf("%s%-*s%s", r.scheme.Info, width, cell, r.resetColor())
		} else {
			parts[i] = fmt.Sprintf("%-*s", width, cell)
		}
	}
	fmt.Fprintf(w, "  %s  \n", strings.Join(parts, "  "))
}

func (r *TableRenderer) renderSeparator(w io.Writer, widths []int) {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	fmt.Fprintf(w, "  %s  \n", strings.Join(parts, "  "))
}

func (r *TableRenderer) resetColor() string {
	if r.colorize {
		return "\033[0m"
	}
	return ""
}

// RenderDiagnostics renders diagnostic report as a table.
func (r *TableRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	// Show overall health
	healthColor := r.scheme.Success
	if report.OverallHealth == dot.HealthWarnings {
		healthColor = r.scheme.Warning
	} else if report.OverallHealth == dot.HealthErrors {
		healthColor = r.scheme.Error
	}

	fmt.Fprintf(w, "%sHealth Status: %s%s\n\n", r.colorText(healthColor), report.OverallHealth.String(), r.resetColor())

	// Show statistics
	fmt.Fprintln(w, "Statistics:")
	fmt.Fprintf(w, "  Total Links: %d\n", report.Statistics.TotalLinks)
	fmt.Fprintf(w, "  Managed Links: %d\n", report.Statistics.ManagedLinks)
	fmt.Fprintf(w, "  Broken Links: %d\n", report.Statistics.BrokenLinks)
	fmt.Fprintf(w, "  Orphaned Links: %d\n\n", report.Statistics.OrphanedLinks)

	// Show issues in a table
	if len(report.Issues) == 0 {
		fmt.Fprintln(w, "No issues found")
		return nil
	}

	headers := []string{"#", "Severity", "Type", "Path", "Message"}
	rows := make([][]string, 0, len(report.Issues))

	for i, issue := range report.Issues {
		pathDisplay := issue.Path
		if len(pathDisplay) > 30 {
			pathDisplay = pathDisplay[:27] + "..."
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			issue.Severity.String(),
			issue.Type.String(),
			pathDisplay,
			issue.Message,
		})
	}

	return r.renderTable(w, headers, rows)
}

func (r *TableRenderer) colorText(color string) string {
	if r.colorize && color != "" {
		return color
	}
	return ""
}

// operationDisplay holds display information for an operation.
type operationDisplay struct {
	Action  string
	Type    string
	Details string
}

// formatOperationForTable extracts display information from an operation.
func formatOperationForTable(op dot.Operation) operationDisplay {
	display := operationDisplay{Action: "Create"}

	switch typed := op.(type) {
	case dot.DirCreate:
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case *dot.DirCreate:
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case dot.LinkCreate:
		display.Type = "Symlink"
		display.Details = fmt.Sprintf("%s -> %s", typed.Target.String(), typed.Source.String())

	case *dot.LinkCreate:
		display.Type = "Symlink"
		display.Details = fmt.Sprintf("%s -> %s", typed.Target.String(), typed.Source.String())

	case dot.FileMove:
		display.Action = "Move"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Dest.String())

	case *dot.FileMove:
		display.Action = "Move"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Dest.String())

	case dot.FileBackup:
		display.Action = "Backup"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Backup.String())

	case *dot.FileBackup:
		display.Action = "Backup"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Backup.String())

	case dot.DirDelete:
		display.Action = "Delete"
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case *dot.DirDelete:
		display.Action = "Delete"
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case dot.LinkDelete:
		display.Action = "Delete"
		display.Type = "Symlink"
		display.Details = typed.Target.String()

	case *dot.LinkDelete:
		display.Action = "Delete"
		display.Type = "Symlink"
		display.Details = typed.Target.String()
	}

	return display
}

// RenderPlan renders an execution plan as a table.
func (r *TableRenderer) RenderPlan(w io.Writer, plan dot.Plan) error {
	fmt.Fprintf(w, "%sDry run mode - no changes will be applied%s\n\n", r.colorText(r.scheme.Warning), r.resetColor())

	if len(plan.Operations) == 0 {
		fmt.Fprintln(w, "No operations required")
		return nil
	}

	// Build table of operations
	headers := []string{"#", "Action", "Type", "Details"}
	rows := make([][]string, 0, len(plan.Operations))

	for i, op := range plan.Operations {
		display := formatOperationForTable(op)

		// Truncate details if too long
		if len(display.Details) > 60 {
			display.Details = display.Details[:57] + "..."
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			display.Action,
			display.Type,
			display.Details,
		})
	}

	if err := r.renderTable(w, headers, rows); err != nil {
		return err
	}

	// Summary
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary:")

	dirCount := 0
	linkCount := 0
	for _, op := range plan.Operations {
		if op.Kind() == dot.OpKindDirCreate {
			dirCount++
		} else if op.Kind() == dot.OpKindLinkCreate {
			linkCount++
		}
	}

	fmt.Fprintf(w, "  Directories: %d\n", dirCount)
	fmt.Fprintf(w, "  Symlinks: %d\n", linkCount)
	fmt.Fprintf(w, "  Conflicts: %d\n", len(plan.Metadata.Conflicts))

	return nil
}
