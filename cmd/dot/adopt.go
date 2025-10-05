package main

import (
	"github.com/spf13/cobra"
)

// newAdoptCommand creates the adopt command.
func newAdoptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt PACKAGE FILE [FILE...]",
		Short: "Move existing files into package then link",
		Long: `Move one or more existing files from the target directory into 
a package, then create symlinks back to the original locations.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement
			return nil
		},
	}

	return cmd
}
