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
		} else {
			// Print all other errors to stderr
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		// Handle doctor-specific exit codes
		exitCode := getDoctorExitCode(err)
		os.Exit(exitCode)
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

// getDoctorExitCode returns the appropriate exit code for doctor command errors.
func getDoctorExitCode(err error) int {
	if err == nil {
		return 0
	}

	errMsg := err.Error()

	// Doctor command uses specific error messages for different health states
	if strings.Contains(errMsg, "health check detected errors") {
		return 2
	}
	if strings.Contains(errMsg, "health check detected warnings") {
		return 1
	}

	// Default error exit code
	return 1
}
