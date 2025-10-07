package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/pkg/dot"
)

// newRemanageCommand creates the remanage command.
func newRemanageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remanage PACKAGE [PACKAGE...]",
		Short: "Reinstall packages with incremental updates",
		Long: `Reinstall one or more packages by removing old symlinks and 
creating new ones.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runRemanage,
	}

	return cmd
}

// runRemanage handles the remanage command execution.
func runRemanage(cmd *cobra.Command, args []string) error {
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

	if err := client.Remanage(ctx, packages...); err != nil {
		return formatError(err)
	}

	if !cfg.DryRun {
		fmt.Printf("Successfully remanaged %d package(s)\n", len(packages))
	}

	return nil
}
