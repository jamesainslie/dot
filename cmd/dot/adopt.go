package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// newAdoptCommand creates the adopt command.
func newAdoptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt [PACKAGE] FILE [FILE...]",
		Short: "Move existing files into package then link",
		Long: `Move one or more existing files from the target directory into 
a package, then create symlinks back to the original locations.

Package name can be auto-derived from the file name:
  dot adopt .ssh              # Auto-creates package "ssh"
  dot adopt .vimrc            # Auto-creates package "vimrc"

Or explicitly specified:
  dot adopt dot-ssh .ssh      # Use package "dot-ssh"
  dot adopt vim .vimrc .vim   # Adopt multiple files to "vim"`,
		Args: argsWithUsage(cobra.MinimumNArgs(1)),
		RunE: runAdopt,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// For auto-naming mode, complete with files
			// For explicit mode, first arg is package, rest are files
			if len(args) == 0 {
				// Could be package name or file - suggest both packages and files
				return getAvailablePackages(), cobra.ShellCompDirectiveDefault
			}
			// Subsequent arguments: complete with files
			return nil, cobra.ShellCompDirectiveDefault
		},
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

	var pkg string
	var files []string

	// Determine if using auto-naming or explicit package name
	if len(args) == 1 {
		// Auto-naming mode: derive package from file name
		files = []string{args[0]}
		pkg = derivePackageName(args[0])
		if pkg == "" {
			return fmt.Errorf("cannot derive package name from: %s", args[0])
		}
		// Apply dotfile translation to package name
		// ".ssh" → "dot-ssh", "README.md" → "README.md"
		pkg = scanner.UntranslateDotfile(pkg)
	} else {
		// Explicit mode: first arg is package, rest are files
		pkg = args[0]
		files = args[1:]
	}

	if err := client.Adopt(ctx, files, pkg); err != nil {
		return formatError(err)
	}

	if !cfg.DryRun {
		fmt.Printf("Successfully adopted %d file(s) into %s\n", len(files), pkg)
	}

	return nil
}
