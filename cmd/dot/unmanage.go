package main

import (
	"github.com/spf13/cobra"
)

// newUnmanageCommand creates the unmanage command.
func newUnmanageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmanage PACKAGE [PACKAGE...]",
		Short: "Remove packages by deleting symlinks",
		Long: `Remove one or more packages by deleting their symlinks from 
the target directory.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement
			return nil
		},
	}

	return cmd
}
