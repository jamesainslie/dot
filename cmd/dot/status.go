package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/cli/renderer"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newStatusCommand creates the status command with configuration from global flags.
func newStatusCommand() *cobra.Command {
	cmd := NewStatusCommand(&dot.Config{})

	// Override RunE to build config from global flags
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfigWithCmd(cmd)
		if err != nil {
			return err
		}

		// Get format and color from local flags
		format, _ := cmd.Flags().GetString("format")
		color, _ := cmd.Flags().GetString("color")

		// Create client
		client, err := dot.NewClient(cfg)
		if err != nil {
			return formatError(err)
		}

		// Get status
		status, err := client.Status(cmd.Context(), args...)
		if err != nil {
			return formatError(err)
		}

		// Determine colorization
		colorize := shouldColorize(color)

		// Create renderer
		// TODO: Get table_style from config
		r, err := renderer.NewRenderer(format, colorize, "")
		if err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}

		// Render status
		if err := r.RenderStatus(cmd.OutOrStdout(), status); err != nil {
			return fmt.Errorf("render failed: %w", err)
		}

		return nil
	}

	return cmd
}

// NewStatusCommand creates the status command.
func NewStatusCommand(cfg *dot.Config) *cobra.Command {
	var format string
	var color string

	cmd := &cobra.Command{
		Use:   "status [PACKAGE...]",
		Short: "Show installation status for packages",
		Long: `Display the current installation state for specified packages.

If no packages are specified, shows status for all installed packages.
The status includes installation timestamp, number of links, and link paths.`,
		Example: `  # Show status for all packages
  dot status

  # Show status for specific packages
  dot status vim tmux

  # Show status in JSON format
  dot status --format=json

  # Show status with colors disabled
  dot status --color=never`,
		ValidArgsFunction: packageCompletion(true), // Complete with installed packages
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create client
			client, err := dot.NewClient(*cfg)
			if err != nil {
				return formatError(err)
			}

			// Get status
			status, err := client.Status(cmd.Context(), args...)
			if err != nil {
				return formatError(err)
			}

			// Determine colorization
			colorize := shouldColorize(color)

			// Create renderer
			// TODO: Get table_style from config
			r, err := renderer.NewRenderer(format, colorize, "")
			if err != nil {
				return fmt.Errorf("invalid format: %w", err)
			}

			// Render status
			if err := r.RenderStatus(cmd.OutOrStdout(), status); err != nil {
				return fmt.Errorf("render failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json, yaml, table)")
	cmd.Flags().StringVar(&color, "color", "auto", "Colorize output (auto, always, never)")

	return cmd
}
