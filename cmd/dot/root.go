package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/spf13/cobra"
)

// Global configuration shared across commands
type globalConfig struct {
	packageDir string
	targetDir  string
	backupDir  string
	dryRun     bool
	verbose    int
	quiet      bool
	logJSON    bool
}

var globalCfg globalConfig

// NewRootCommand creates the root cobra command.
func NewRootCommand(version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dot",
		Short: "Modern symlink manager for dotfiles",
		Long: `dot is a type-safe dotfile manager written in Go.

dot manages dotfiles by creating symlinks from a source directory 
(package directory) to a target directory. It provides atomic operations,
comprehensive conflict detection, and incremental updates.`,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Set up flag error function to show usage on flag parsing errors
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
		_ = cmd.Usage()
		return err
	})

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&globalCfg.packageDir, "dir", "d", ".",
		"Source directory containing packages")

	// Compute cross-platform home directory default
	defaultTarget, err := os.UserHomeDir()
	if err != nil || defaultTarget == "" {
		// Fall back to current working directory
		defaultTarget, err = os.Getwd()
		if err != nil || defaultTarget == "" {
			defaultTarget = "."
		}
	}

	rootCmd.PersistentFlags().StringVarP(&globalCfg.targetDir, "target", "t", defaultTarget,
		"Target directory for symlinks")
	rootCmd.PersistentFlags().StringVar(&globalCfg.backupDir, "backup-dir", "",
		"Directory for backup files (default: <target>/.dot-backup)")
	rootCmd.PersistentFlags().BoolVarP(&globalCfg.dryRun, "dry-run", "n", false,
		"Show what would be done without applying changes")
	rootCmd.PersistentFlags().CountVarP(&globalCfg.verbose, "verbose", "v",
		"Increase verbosity (repeatable: -v, -vv, -vvv)")
	rootCmd.PersistentFlags().BoolVarP(&globalCfg.quiet, "quiet", "q", false,
		"Suppress all non-error output")
	rootCmd.PersistentFlags().BoolVar(&globalCfg.logJSON, "log-json", false,
		"Output logs in JSON format")

	// Add subcommands
	rootCmd.AddCommand(
		newManageCommand(),
		newUnmanageCommand(),
		newRemanageCommand(),
		newAdoptCommand(),
		newStatusCommand(),
		newListCommand(),
		newDoctorCommand(),
		newConfigCommand(),
	)

	return rootCmd
}

// buildConfig creates a dot.Config from global flags and adapters.
func buildConfig() (dot.Config, error) {
	// Make paths absolute
	packageDir, err := filepath.Abs(globalCfg.packageDir)
	if err != nil {
		return dot.Config{}, fmt.Errorf("invalid package directory: %w", err)
	}

	targetDir, err := filepath.Abs(globalCfg.targetDir)
	if err != nil {
		return dot.Config{}, fmt.Errorf("invalid target directory: %w", err)
	}

	// Create adapters
	fs := adapters.NewOSFilesystem()
	logger := createLogger()

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		BackupDir:  globalCfg.backupDir,
		DryRun:     globalCfg.dryRun,
		Verbosity:  globalCfg.verbose,
		FS:         fs,
		Logger:     logger,
	}

	return cfg.WithDefaults(), nil
}

// createLogger creates appropriate logger based on flags.
func createLogger() dot.Logger {
	if globalCfg.quiet {
		return adapters.NewNoopLogger()
	}

	level := verbosityToLevel(globalCfg.verbose)

	if globalCfg.logJSON {
		return adapters.NewSlogLogger(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})))
	}

	return adapters.NewSlogLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

// verbosityToLevel converts verbosity count to log level.
func verbosityToLevel(v int) slog.Level {
	switch {
	case v == 0:
		return slog.LevelInfo
	case v == 1:
		return slog.LevelDebug
	default:
		// Even more verbose
		return slog.LevelDebug - slog.Level(v-1)
	}
}

// formatError converts domain errors to user-friendly messages.
func formatError(err error) error {
	// For now, just return the error
	// In the future, this can be enhanced to provide better error messages
	return err
}

// argsWithUsage wraps a Cobra Args validator to show usage on validation errors.
func argsWithUsage(validator cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := validator(cmd, args)
		if err != nil {
			// Print error and usage
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
			_ = cmd.Usage()
		}
		return err
	}
}

// shouldColorize determines if output should be colorized based on the color flag.
func shouldColorize(color string) bool {
	// Respect NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	switch color {
	case "always":
		return true
	case "never":
		return false
	case "auto":
		// Check if stdout is a terminal using portable detection
		return term.IsTerminal(int(os.Stdout.Fd()))
	default:
		return false
	}
}
