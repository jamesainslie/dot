package main

import (
	"github.com/spf13/cobra"
)

// newRemanageCommand creates the remanage command.
func newRemanageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remanage PACKAGE [PACKAGE...]",
		Short: "Reinstall packages with incremental updates",
		Long: `Reinstall one or more packages by removing old symlinks and 
creating new ones.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement
			return nil
		},
	}

	return cmd
}
