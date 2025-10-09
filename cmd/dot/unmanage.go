package main

import (
	"context"

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
		Args: argsWithUsage(cobra.MinimumNArgs(1)),
		RunE: runUnmanage,
	}

	return cmd
}

// runUnmanage handles the unmanage command execution.
func runUnmanage(cmd *cobra.Command, args []string) error {
	return executePackageCommand(cmd, args, func(client *dot.Client, ctx context.Context, packages []string) error {
		return client.Unmanage(ctx, packages...)
	}, "unmanaged")
}
