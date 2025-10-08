package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags at build time)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := NewRootCommand(version, commit, date)

	// Execute command
	executedCmd, err := executeCommand(rootCmd)
	if err != nil {
		// Show usage for argument validation errors
		// (Flag errors are handled by SetFlagErrorFunc in root.go)
		if executedCmd != nil && isArgValidationError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			_ = executedCmd.Usage()
		}

		os.Exit(1)
	}
}

// executeCommand executes the root command and returns the executed command and any error.
func executeCommand(rootCmd *cobra.Command) (*cobra.Command, error) {
	var executedCmd *cobra.Command

	// Use PreRun hook to capture the executed command
	originalPreRun := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		executedCmd = cmd
		if originalPreRun != nil {
			return originalPreRun(cmd, args)
		}
		return nil
	}

	err := rootCmd.Execute()
	return executedCmd, err
}

// isArgValidationError determines if an error is from argument validation.
func isArgValidationError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Common argument validation error patterns from Cobra
	argPatterns := []string{
		"accepts",
		"requires",
		"requires at least",
		"requires at most",
		"accepts at most",
		"too many arguments",
		"unknown command",
	}

	for _, pattern := range argPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}
