package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/cli/renderer"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newListCommand creates the list command with configuration from global flags.
func newListCommand() *cobra.Command {
	cmd := NewListCommand(&dot.Config{})

	// Override RunE to build config from global flags
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfigWithCmd(cmd)
		if err != nil {
			return err
		}

		// Get flags
		format, _ := cmd.Flags().GetString("format")
		color, _ := cmd.Flags().GetString("color")
		sortBy, _ := cmd.Flags().GetString("sort")

		// Create client
		client, err := dot.NewClient(cfg)
		if err != nil {
			return formatError(err)
		}

		// Get list of packages
		packages, err := client.List(cmd.Context())
		if err != nil {
			return formatError(err)
		}

		// Sort packages
		sortPackages(packages, sortBy)

		// Create status from packages
		status := dot.Status{
			Packages: packages,
		}

		// Determine colorization
		colorize := shouldColorize(color)

		// Print context header (only for text/table formats)
		if format == "text" || format == "table" {
			fmt.Fprintf(cmd.OutOrStdout(), "Package directory: %s\n", cfg.PackageDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Target directory:  %s\n", cfg.TargetDir)
			if cfg.ManifestDir != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Manifest:          %s\n", cfg.ManifestDir)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Manifest:          %s/.dot-manifest.json\n", cfg.TargetDir)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		// Create renderer
		// TODO: Get table_style from config
		r, err := renderer.NewRenderer(format, colorize, "")
		if err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}

		// Render list
		if err := r.RenderStatus(cmd.OutOrStdout(), status); err != nil {
			return fmt.Errorf("render failed: %w", err)
		}

		return nil
	}

	return cmd
}

// NewListCommand creates the list command.
func NewListCommand(cfg *dot.Config) *cobra.Command {
	var format string
	var color string
	var sortBy string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed packages",
		Long: `Display information about all installed packages.

Shows package name, link count, and installation timestamp for all
packages currently managed by dot. The list can be sorted by various
fields and displayed in multiple output formats.`,
		Example: `  # List all packages
  dot list

  # List packages sorted by link count
  dot list --sort=links

  # List packages in JSON format
  dot list --format=json

  # List packages without colors
  dot list --color=never`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder - will be overridden by newListCommand
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (text, json, yaml, table)")
	cmd.Flags().StringVar(&color, "color", "auto", "Colorize output (auto, always, never)")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort by field (name, links, date)")

	return cmd
}

// sortPackages sorts packages by the specified field.
func sortPackages(packages []dot.PackageInfo, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
	case "links":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].LinkCount > packages[j].LinkCount // Descending
		})
	case "date":
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].InstalledAt.After(packages[j].InstalledAt) // Most recent first
		})
	default:
		// Default to name sorting
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
	}
}
