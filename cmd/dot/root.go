package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/spf13/cobra"
)

// Global configuration shared across commands
type globalConfig struct {
	stowDir   string
	targetDir string
	dryRun    bool
	verbose   int
	quiet     bool
	logJSON   bool
}

var globalCfg globalConfig

// NewRootCommand creates the root cobra command.
func NewRootCommand(version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dot",
		Short: "Modern symlink manager for dotfiles",
		Long: `dot is a type-safe GNU Stow replacement written in Go.

dot manages dotfiles by creating symlinks from a source directory 
(stow directory) to a target directory. It provides atomic operations,
comprehensive conflict detection, and incremental updates.`,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&globalCfg.stowDir, "dir", "d", ".",
		"Stow directory containing packages")
	rootCmd.PersistentFlags().StringVarP(&globalCfg.targetDir, "target", "t", os.Getenv("HOME"),
		"Target directory for symlinks")
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
	)

	return rootCmd
}

// buildConfig creates a dot.Config from global flags and adapters.
func buildConfig() (dot.Config, error) {
	// Make paths absolute
	stowDir, err := filepath.Abs(globalCfg.stowDir)
	if err != nil {
		return dot.Config{}, fmt.Errorf("invalid stow directory: %w", err)
	}

	targetDir, err := filepath.Abs(globalCfg.targetDir)
	if err != nil {
		return dot.Config{}, fmt.Errorf("invalid target directory: %w", err)
	}

	// Create adapters
	fs := adapters.NewOSFilesystem()
	logger := createLogger()

	cfg := dot.Config{
		StowDir:   stowDir,
		TargetDir: targetDir,
		DryRun:    globalCfg.dryRun,
		Verbosity: globalCfg.verbose,
		FS:        fs,
		Logger:    logger,
	}

	return cfg, nil
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

// shouldColorize determines if output should be colorized based on the color flag.
func shouldColorize(color string) bool {
	switch color {
	case "always":
		return true
	case "never":
		return false
	case "auto":
		// Check if stdout is a terminal
		fileInfo, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	default:
		return false
	}
}
