package main

import (
	"github.com/spf13/cobra"
)

// newManageCommand creates the manage command.
func newManageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage PACKAGE [PACKAGE...]",
		Short: "Install packages by creating symlinks",
		Long: `Install one or more packages by creating symlinks from the stow 
directory to the target directory.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement
			return nil
		},
	}

	return cmd
}
