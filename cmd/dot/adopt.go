package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/pkg/dot"
)

// newAdoptCommand creates the adopt command.
func newAdoptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt PACKAGE FILE [FILE...]",
		Short: "Move existing files into package then link",
		Long: `Move one or more existing files from the target directory into 
a package, then create symlinks back to the original locations.`,
		Args: argsWithUsage(cobra.MinimumNArgs(2)),
		RunE: runAdopt,
	}

	return cmd
}

// runAdopt handles the adopt command execution.
func runAdopt(cmd *cobra.Command, args []string) error {
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

	// First arg is package, rest are files
	pkg := args[0]
	files := args[1:]

	if err := client.Adopt(ctx, files, pkg); err != nil {
		return formatError(err)
	}

	if !cfg.DryRun {
		fmt.Printf("Successfully adopted %d file(s) into %s\n", len(files), pkg)
	}

	return nil
}
