package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/cli/renderer"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newManageCommand creates the manage command.
func newManageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage PACKAGE [PACKAGE...]",
		Short: "Install packages by creating symlinks",
		Long: `Install one or more packages by creating symlinks from the package 
directory to the target directory.`,
		Args: argsWithUsage(cobra.MinimumNArgs(1)),
		RunE: runManage,
	}

	return cmd
}

// runManage handles the manage command execution.
func runManage(cmd *cobra.Command, args []string) error {
	cfg, err := buildConfig()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	packages := args

	// If dry-run mode, render the plan instead of executing
	if cfg.DryRun {
		plan, err := client.PlanManage(ctx, packages...)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}

		// Create renderer and render the plan
		rend, err := renderer.NewRenderer("text", true)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}

		if err := rend.RenderPlan(os.Stdout, plan); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}

		return nil
	}

	// Normal execution
	if err := client.Manage(ctx, packages...); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	fmt.Printf("Successfully managed %d package(s)\n", len(packages))

	return nil
}
