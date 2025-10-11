package renderer

import (
	"fmt"
	"io"

	"github.com/jamesainslie/dot/internal/cli/pretty"
	"github.com/jamesainslie/dot/internal/domain"
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

	// Create table with Light style for clean, professional look
	table := pretty.NewTableWriter(pretty.StyleLight, pretty.TableConfig{
		ColorEnabled: r.colorize,
		AutoWrap:     true,
		MaxWidth:     0, // Auto-detect terminal width
	})

	// Set header
	table.SetHeader("Package", "Links", "Installed")

	// Add rows
	for _, pkg := range status.Packages {
		table.AppendRow(
			pkg.Name,
			fmt.Sprintf("%d", pkg.LinkCount),
			formatDuration(pkg.InstalledAt),
		)
	}

	// Render
	table.Render(w)
	return nil
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

	// Create table with Light style
	table := pretty.NewTableWriter(pretty.StyleLight, pretty.TableConfig{
		ColorEnabled: r.colorize,
		AutoWrap:     true,
		MaxWidth:     0, // Auto-detect terminal width
	})

	// Set header
	table.SetHeader("#", "Severity", "Type", "Path", "Message")

	// Add rows
	for i, issue := range report.Issues {
		table.AppendRow(
			fmt.Sprintf("%d", i+1),
			issue.Severity.String(),
			issue.Type.String(),
			issue.Path, // Let TableWriter handle truncation/wrapping
			issue.Message,
		)
	}

	// Render
	table.Render(w)
	return nil
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
func formatOperationForTable(op domain.Operation) operationDisplay {
	// Normalize: dereference pointers to get value type for switching
	normalized := normalizeOperation(op)

	display := operationDisplay{Action: "Create"}

	switch typed := normalized.(type) {
	case domain.DirCreate:
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case domain.LinkCreate:
		display.Type = "Symlink"
		display.Details = fmt.Sprintf("%s -> %s", typed.Target.String(), typed.Source.String())

	case domain.FileMove:
		display.Action = "Move"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Dest.String())

	case domain.FileBackup:
		display.Action = "Backup"
		display.Type = "File"
		display.Details = fmt.Sprintf("%s -> %s", typed.Source.String(), typed.Backup.String())

	case domain.DirDelete:
		display.Action = "Delete"
		display.Type = "Directory"
		display.Details = typed.Path.String()

	case domain.LinkDelete:
		display.Action = "Delete"
		display.Type = "Symlink"
		display.Details = typed.Target.String()

	default:
		// Handle unknown operation types with clear, informative display
		display.Action = "Unknown"
		display.Type = fmt.Sprintf("%T", op)
		display.Details = op.String()
	}

	return display
}

// RenderPlan renders an execution plan as a table.
func (r *TableRenderer) RenderPlan(w io.Writer, plan domain.Plan) error {
	fmt.Fprintf(w, "%sDry run mode - no changes will be applied%s\n\n", r.colorText(r.scheme.Warning), r.resetColor())

	if len(plan.Operations) == 0 {
		fmt.Fprintln(w, "No operations required")
		return nil
	}

	// Create table with Light style
	table := pretty.NewTableWriter(pretty.StyleLight, pretty.TableConfig{
		ColorEnabled: r.colorize,
		AutoWrap:     true,
		MaxWidth:     0, // Auto-detect terminal width
	})

	// Set header
	table.SetHeader("#", "Action", "Type", "Details")

	// Add rows
	for i, op := range plan.Operations {
		display := formatOperationForTable(op)

		table.AppendRow(
			fmt.Sprintf("%d", i+1),
			display.Action,
			display.Type,
			display.Details, // Let TableWriter handle truncation/wrapping
		)
	}

	// Render
	table.Render(w)

	// Summary
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary:")

	// Count all operation kinds in a single pass
	counts := make(map[domain.OperationKind]int)
	for _, op := range plan.Operations {
		counts[op.Kind()]++
	}

	// Display counts with semantic labels for each operation kind
	if count := counts[domain.OpKindDirCreate]; count > 0 {
		fmt.Fprintf(w, "  Directories created: %d\n", count)
	}
	if count := counts[domain.OpKindLinkCreate]; count > 0 {
		fmt.Fprintf(w, "  Symlinks created: %d\n", count)
	}
	if count := counts[domain.OpKindFileMove]; count > 0 {
		fmt.Fprintf(w, "  Files moved: %d\n", count)
	}
	if count := counts[domain.OpKindFileBackup]; count > 0 {
		fmt.Fprintf(w, "  Backups created: %d\n", count)
	}
	if count := counts[domain.OpKindDirDelete]; count > 0 {
		fmt.Fprintf(w, "  Directories deleted: %d\n", count)
	}
	if count := counts[domain.OpKindLinkDelete]; count > 0 {
		fmt.Fprintf(w, "  Symlinks deleted: %d\n", count)
	}

	// Always show conflicts count
	fmt.Fprintf(w, "  Conflicts: %d\n", len(plan.Metadata.Conflicts))

	return nil
}
