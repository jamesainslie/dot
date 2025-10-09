package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/pkg/dot"
)

// packageCommandFunc is a function that executes a package operation.
type packageCommandFunc func(*dot.Client, context.Context, []string) error

// executePackageCommand is a helper that handles the common pattern for package commands.
// It builds the config, creates a client, executes the provided function, and prints success message.
func executePackageCommand(cmd *cobra.Command, args []string, fn packageCommandFunc, actionVerb string) error {
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

	if err := fn(client, ctx, packages); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	if !cfg.DryRun {
		fmt.Printf("Successfully %s %d package(s)\n", actionVerb, len(packages))
	}

	return nil
}
