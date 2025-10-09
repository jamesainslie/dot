package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/cli/renderer"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newDoctorCommand creates the doctor command with configuration from global flags.
func newDoctorCommand() *cobra.Command {
	cmd := NewDoctorCommand(&dot.Config{})

	// Override RunE to build config from global flags
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfig()
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
		case "off", "":
			scanCfg = dot.DefaultScanConfig()
		case "scoped":
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

		// Render diagnostics
		if err := r.RenderDiagnostics(cmd.OutOrStdout(), report); err != nil {
			return fmt.Errorf("render failed: %w", err)
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

// NewDoctorCommand creates the doctor command.
func NewDoctorCommand(cfg *dot.Config) *cobra.Command {
	var format string
	var color string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Perform health checks on the installation",
		Long: `Run comprehensive health checks on the dot installation.

Checks for:
  - Broken symlinks (links pointing to non-existent targets)
  - Orphaned links (links not managed by dot)
  - Permission issues
  - Manifest inconsistencies

Exit codes:
  0 - Healthy (no issues found)
  1 - Warnings detected
  2 - Errors detected`,
		Example: `  # Run health check
  dot doctor

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
	cmd.Flags().String("scan-mode", "off", "Orphan detection mode (off, scoped, deep)")
	cmd.Flags().Int("max-depth", 10, "Maximum recursion depth for deep scan")

	return cmd
}
