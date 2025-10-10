package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/pkg/dot"
)

// newUnmanageCommand creates the unmanage command.
func newUnmanageCommand() *cobra.Command {
	var purge bool
	var noRestore bool
	var cleanup bool

	cmd := &cobra.Command{
		Use:   "unmanage PACKAGE [PACKAGE...]",
		Short: "Remove packages by deleting symlinks",
		Long: `Remove one or more packages by deleting their symlinks from 
the target directory.

By default, adopted packages (created via 'dot adopt') are restored to 
their original locations. Managed packages only have their symlinks removed.

Cleanup mode removes orphaned packages from the manifest without modifying 
the filesystem - useful when packages no longer exist.`,
		Example: `  # Remove package and restore adopted files
  dot unmanage ssh

  # Remove package and delete package directory
  dot unmanage ssh --purge

  # Remove package without restoring (leave in package dir)
  dot unmanage ssh --no-restore

  # Clean up orphaned manifest entry (no filesystem changes)
  dot unmanage old-package --cleanup`,
		Args: argsWithUsage(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnmanage(cmd, args, purge, noRestore, cleanup)
		},
		ValidArgsFunction: packageCompletion(true), // Complete with installed packages
	}

	cmd.Flags().BoolVar(&purge, "purge", false, "Delete package directory instead of restoring files")
	cmd.Flags().BoolVar(&noRestore, "no-restore", false, "Don't restore adopted files (leave in package directory)")
	cmd.Flags().BoolVar(&cleanup, "cleanup", false, "Remove orphaned manifest entries (packages with missing links/directories)")

	return cmd
}

// runUnmanage handles the unmanage command execution.
func runUnmanage(cmd *cobra.Command, args []string, purge, noRestore, cleanup bool) error {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return err
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	packages := args

	// Build options
	opts := dot.UnmanageOptions{
		Purge:   purge,
		Restore: !noRestore, // Default is true unless --no-restore
		Cleanup: cleanup,
	}

	// Execute unmanage with options
	if err := client.UnmanageWithOptions(ctx, opts, packages...); err != nil {
		return err
	}

	if !cfg.DryRun {
		if cleanup {
			fmt.Printf("Cleaned up %d orphaned package(s) from manifest\n", len(packages))
		} else if purge {
			fmt.Printf("Successfully unmanaged and purged %d package(s)\n", len(packages))
		} else if opts.Restore {
			fmt.Printf("Successfully unmanaged and restored %d package(s)\n", len(packages))
		} else {
			fmt.Printf("Successfully unmanaged %d package(s)\n", len(packages))
		}
	}

	return nil
}
