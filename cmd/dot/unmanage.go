package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/pkg/dot"
)

// newUnmanageCommand creates the unmanage command.
func newUnmanageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmanage PACKAGE [PACKAGE...]",
		Short: "Remove packages by deleting symlinks",
		Long: `Remove one or more packages by deleting their symlinks from 
the target directory.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runUnmanage,
	}

	return cmd
}

// runUnmanage handles the unmanage command execution.
func runUnmanage(cmd *cobra.Command, args []string) error {
	cfg, err := buildConfig()
	if err != nil {
		return formatError(err)
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		return formatError(err)
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	packages := args

	if err := client.Unmanage(ctx, packages...); err != nil {
		return formatError(err)
	}

	if !cfg.DryRun {
		fmt.Printf("Successfully unmanaged %d package(s)\n", len(packages))
	}

	return nil
}
