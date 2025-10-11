package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/cli/renderer"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newDoctorCommand creates the doctor command with configuration from global flags.
func newDoctorCommand() *cobra.Command {
	cmd := NewDoctorCommand(&dot.Config{})

	// Override RunE to build config from global flags
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfigWithCmd(cmd)
		if err != nil {
			return err
		}

		// Get flags
		format, _ := cmd.Flags().GetString("format")
		color, _ := cmd.Flags().GetString("color")
		scanMode, _ := cmd.Flags().GetString("scan-mode")
		maxDepth, _ := cmd.Flags().GetInt("max-depth")

		// Create client
		client, err := dot.NewClient(cfg)
		if err != nil {
			return formatError(err)
		}

		// Build scan config based on flags
		var scanCfg dot.ScanConfig
		switch scanMode {
		case "off":
			scanCfg = dot.ScanConfig{
				Mode:         dot.ScanOff,
				MaxDepth:     10,
				ScopeToDirs:  nil,
				SkipPatterns: []string{".git", "node_modules", ".cache", ".npm", ".cargo", ".rustup"},
			}
		case "scoped", "":
			scanCfg = dot.ScopedScanConfig()
		case "deep":
			scanCfg = dot.DeepScanConfig(maxDepth)
		default:
			return fmt.Errorf("invalid scan-mode: %s (must be off, scoped, or deep)", scanMode)
		}

		// Run diagnostics
		report, err := client.DoctorWithScan(cmd.Context(), scanCfg)
		if err != nil {
			return formatError(err)
		}

		// Determine colorization
		colorize := shouldColorize(color)

		// Create renderer
		r, err := renderer.NewRenderer(format, colorize)
		if err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}

		// Render diagnostics - use succinct output for text format
		if format == "text" {
			renderSuccinctDiagnostics(cmd.OutOrStdout(), report)
		} else {
			if err := r.RenderDiagnostics(cmd.OutOrStdout(), report); err != nil {
				return fmt.Errorf("render failed: %w", err)
			}
		}

		// Return error to set exit code based on health status
		// The main function will handle converting this to an exit code
		if report.OverallHealth == dot.HealthErrors {
			return fmt.Errorf("health check detected errors")
		} else if report.OverallHealth == dot.HealthWarnings {
			return fmt.Errorf("health check detected warnings")
		}

		return nil
	}

	return cmd
}

// renderSuccinctDiagnostics outputs diagnostics in a succinct, colorized format.
func renderSuccinctDiagnostics(w io.Writer, report dot.DiagnosticReport) {
	// Health status header
	healthIcon, healthText, healthColor := getHealthDisplay(report.OverallHealth)
	fmt.Fprintf(w, "%s %s\n",
		healthIcon,
		healthColor(healthText),
	)

	// Statistics summary if available
	if report.Statistics.TotalLinks > 0 {
		fmt.Fprintf(w, "  %s %s\n",
			dim("•"),
			dim(fmt.Sprintf("%d total links (%d managed, %d broken, %d orphaned)",
				report.Statistics.TotalLinks,
				report.Statistics.ManagedLinks,
				report.Statistics.BrokenLinks,
				report.Statistics.OrphanedLinks)),
		)
	}

	// Issues grouped by severity
	errors := filterIssuesBySeverity(report.Issues, dot.SeverityError)
	warnings := filterIssuesBySeverity(report.Issues, dot.SeverityWarning)
	infos := filterIssuesBySeverity(report.Issues, dot.SeverityInfo)

	if len(errors) > 0 {
		fmt.Fprintf(w, "\n%s %s\n", errorText("✗"), errorText(fmt.Sprintf("%d errors:", len(errors))))
		renderIssueList(w, errors, errorText)
	}

	if len(warnings) > 0 {
		fmt.Fprintf(w, "\n%s %s\n", warning("⚠"), warning(fmt.Sprintf("%d warnings:", len(warnings))))
		renderIssueList(w, warnings, warning)
	}

	if len(infos) > 0 {
		fmt.Fprintf(w, "\n%s %s\n", info("ℹ"), info(fmt.Sprintf("%d info:", len(infos))))
		renderIssueList(w, infos, dim)
	}

	// Clean summary if no issues
	if len(report.Issues) == 0 {
		fmt.Fprintf(w, "  %s\n", success("No issues found"))
	}
}

// getHealthDisplay returns icon, text, and color for health status
func getHealthDisplay(health dot.HealthStatus) (string, string, func(string) string) {
	switch health {
	case dot.HealthOK:
		return success("✓"), "Healthy", success
	case dot.HealthWarnings:
		return warning("⚠"), "Warnings detected", warning
	case dot.HealthErrors:
		return errorText("✗"), "Errors detected", errorText
	default:
		return dim("?"), "Unknown status", dim
	}
}

// filterIssuesBySeverity returns issues matching the given severity
func filterIssuesBySeverity(issues []dot.Issue, severity dot.IssueSeverity) []dot.Issue {
	var filtered []dot.Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// renderIssueList renders a list of issues succinctly
func renderIssueList(w io.Writer, issues []dot.Issue, colorFunc func(string) string) {
	for _, issue := range issues {
		fmt.Fprintf(w, "  %s %s",
			dim("•"),
			bold(issue.Path),
		)
		if issue.Message != "" {
			fmt.Fprintf(w, " %s %s",
				dim("—"),
				dim(issue.Message),
			)
		}
		fmt.Fprintln(w)
	}
}

// NewDoctorCommand creates the doctor command.
func NewDoctorCommand(cfg *dot.Config) *cobra.Command {
	var format string
	var color string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Perform health checks on the installation",
		Long: `Run comprehensive health checks on the dot installation.

Checks for:
  - Broken symlinks in managed packages (links pointing to non-existent targets)
  - Orphaned symlinks not in manifest (unmanaged links in target directory)
  - Broken unmanaged symlinks (orphaned links with non-existent targets)
  - Permission issues
  - Manifest inconsistencies

Orphan Detection:
  By default, doctor uses scoped scanning to find unmanaged symlinks in
  directories containing managed links. This efficiently detects leftover
  symlinks from previously managed packages.

  Use --scan-mode=off to disable orphan detection for faster checks.
  Use --scan-mode=deep for thorough scanning of entire target directory.

Exit codes:
  0 - Healthy (no issues found)
  1 - Warnings detected (e.g., orphaned links)
  2 - Errors detected (e.g., broken links)`,
		Example: `  # Run health check with default scoped scanning
  dot doctor

  # Run health check without orphan detection (faster)
  dot doctor --scan-mode=off

  # Run thorough scan of entire home directory
  dot doctor --scan-mode=deep

  # Run health check with JSON output
  dot doctor --format=json

  # Run health check without colors
  dot doctor --color=never`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder - will be overridden by newDoctorCommand
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json, yaml, table)")
	cmd.Flags().StringVar(&color, "color", "auto", "Colorize output (auto, always, never)")
	cmd.Flags().String("scan-mode", "scoped", "Orphan detection mode (off, scoped, deep)")
	cmd.Flags().Int("max-depth", 10, "Maximum recursion depth for deep scan")

	return cmd
}
