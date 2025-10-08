package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/spf13/cobra"
)

// newConfigCommand creates the config command.
func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage dot configuration",
		Long: `View and modify dot configuration settings.

Configuration is loaded from multiple sources in order of precedence:
  1. Command-line flags (highest)
  2. Environment variables (DOT_* prefix)
  3. Configuration file (~/.config/dot/config.yaml)
  4. Built-in defaults (lowest)

The config command allows viewing current settings, modifying configuration
files, and managing configuration across sources.`,
		Example: `  # Show current configuration
  dot config list

  # Initialize configuration file
  dot config init

  # Get specific value
  dot config get directories.package

  # Set configuration value
  dot config set directories.package ~/dotfiles

  # Show configuration file path
  dot config path`,
		RunE: runConfigList,
	}

	cmd.AddCommand(
		newConfigInitCommand(),
		newConfigGetCommand(),
		newConfigSetCommand(),
		newConfigListCommand(),
		newConfigPathCommand(),
	)

	return cmd
}

// runConfigList is the default action (list config).
func runConfigList(cmd *cobra.Command, args []string) error {
	return runConfigListCmd(cmd, args)
}

// getConfigFilePath returns the configuration file path for the app.
func getConfigFilePath() string {
	// Check for explicit config file path
	if path := os.Getenv("DOT_CONFIG"); path != "" {
		return path
	}

	// Use XDG config directory with default filename
	configDir := config.GetConfigPath("dot")
	return filepath.Join(configDir, "config.yaml")
}

// newConfigInitCommand creates the init subcommand.
func newConfigInitCommand() *cobra.Command {
	var force bool
	var format string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create initial configuration file",
		Long: `Create a new configuration file with default values.

The configuration file is created in the XDG config directory:
  ~/.config/dot/config.yaml (default)

Use --force to overwrite existing configuration.`,
		Example: `  # Create config with defaults
  dot config init

  # Force overwrite existing config
  dot config init --force

  # Create in JSON format
  dot config init --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(force, format)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing config")
	cmd.Flags().StringVar(&format, "format", "yaml", "Config format (yaml, json, toml)")

	return cmd
}

// runConfigInit handles the init subcommand.
func runConfigInit(force bool, format string) error {
	configPath := getConfigFilePath()

	// If format was not explicitly specified (still default "yaml"),
	// detect format from configPath extension
	if format == "yaml" {
		ext := filepath.Ext(configPath)
		if ext != "" {
			// Normalize extension
			ext = strings.TrimPrefix(ext, ".")
			if ext == "yml" {
				ext = "yaml"
			}
			// Use detected format if recognized
			if ext == "json" || ext == "toml" {
				format = ext
			}
		}
	} else {
		// Format explicitly specified - adjust path extension
		dir := filepath.Dir(configPath)
		base := filepath.Base(configPath)
		// Strip existing extension and add new one
		ext := filepath.Ext(base)
		if ext != "" {
			base = base[:len(base)-len(ext)]
		}
		configPath = filepath.Join(dir, base+"."+format)
	}

	// Check if exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
	}

	// Create writer and write default config
	writer := config.NewWriter(configPath)
	if err := writer.WriteDefault(config.WriteOptions{
		Format:          format,
		IncludeComments: format == "yaml",
	}); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Printf("Configuration file created: %s\n", configPath)
	if editor := os.Getenv("EDITOR"); editor != "" {
		fmt.Printf("Edit with: %s %s\n", editor, configPath)
	} else {
		fmt.Printf("Edit with your preferred editor\n")
	}

	return nil
}

// newConfigGetCommand creates the get subcommand.
func newConfigGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get configuration value",
		Long: `Retrieve configuration value by key path.

Keys use dot notation: section.field
For example: directories.package, logging.level`,
		Example: `  # Get package directory
  dot config get directories.package

  # Get logging level
  dot config get logging.level`,
		Args: argsWithUsage(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGet(args[0])
		},
	}

	return cmd
}

// runConfigGet handles the get subcommand.
func runConfigGet(key string) error {
	configPath := getConfigFilePath()

	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	value, err := getConfigValue(cfg, key)
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}

// getConfigValue retrieves a value from config by key path.
func getConfigValue(cfg *config.ExtendedConfig, key string) (string, error) {
	switch key {
	case "directories.package":
		return cfg.Directories.Package, nil
	case "directories.target":
		return cfg.Directories.Target, nil
	case "directories.manifest":
		return cfg.Directories.Manifest, nil
	case "logging.level":
		return cfg.Logging.Level, nil
	case "logging.format":
		return cfg.Logging.Format, nil
	case "logging.destination":
		return cfg.Logging.Destination, nil
	case "symlinks.mode":
		return cfg.Symlinks.Mode, nil
	case "symlinks.backup_suffix":
		return cfg.Symlinks.BackupSuffix, nil
	case "symlinks.backup_dir":
		return cfg.Symlinks.BackupDir, nil
	case "dotfile.prefix":
		return cfg.Dotfile.Prefix, nil
	case "output.format":
		return cfg.Output.Format, nil
	case "output.color":
		return cfg.Output.Color, nil
	case "packages.sort_by":
		return cfg.Packages.SortBy, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// newConfigSetCommand creates the set subcommand.
func newConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration value",
		Long: `Set configuration value by key path.

Keys use dot notation: section.field
Values are automatically type-converted based on the field.`,
		Example: `  # Set package directory
  dot config set directories.package ~/dotfiles

  # Set logging level
  dot config set logging.level DEBUG

  # Set symlink mode
  dot config set symlinks.mode absolute`,
		Args: argsWithUsage(cobra.ExactArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSet(args[0], args[1])
		},
	}

	return cmd
}

// runConfigSet handles the set subcommand.
func runConfigSet(key, value string) error {
	configPath := getConfigFilePath()

	writer := config.NewWriter(configPath)
	if err := writer.Update(key, value); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	fmt.Printf("Updated configuration: %s\n", configPath)
	fmt.Printf("  %s: %s\n", key, value)

	return nil
}

// newConfigListCommand creates the list subcommand.
func newConfigListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration settings",
		Long: `Display all configuration settings with their current values.

Shows the final merged configuration from all sources.`,
		Example: `  # List all settings
  dot config list

  # List in JSON format
  dot config list --format json`,
		RunE: runConfigListCmd,
	}

	return cmd
}

// runConfigListCmd handles the list subcommand.
func runConfigListCmd(cmd *cobra.Command, args []string) error {
	configPath := getConfigFilePath()

	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Display configuration
	fmt.Printf("Configuration (from %s):\n\n", configPath)

	fmt.Println("Directories:")
	fmt.Printf("  package: %s\n", cfg.Directories.Package)
	fmt.Printf("  target: %s\n", cfg.Directories.Target)
	fmt.Printf("  manifest: %s\n\n", cfg.Directories.Manifest)

	fmt.Println("Logging:")
	fmt.Printf("  level: %s\n", cfg.Logging.Level)
	fmt.Printf("  format: %s\n", cfg.Logging.Format)
	fmt.Printf("  destination: %s\n\n", cfg.Logging.Destination)

	fmt.Println("Symlinks:")
	fmt.Printf("  mode: %s\n", cfg.Symlinks.Mode)
	fmt.Printf("  folding: %t\n", cfg.Symlinks.Folding)
	fmt.Printf("  backup: %t\n\n", cfg.Symlinks.Backup)

	fmt.Println("Output:")
	fmt.Printf("  format: %s\n", cfg.Output.Format)
	fmt.Printf("  color: %s\n", cfg.Output.Color)
	fmt.Printf("  verbosity: %d\n", cfg.Output.Verbosity)

	return nil
}

// newConfigPathCommand creates the path subcommand.
func newConfigPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  `Display the path to the configuration file.`,
		Example: `  # Show config path
  dot config path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigPath()
		},
	}

	return cmd
}

// runConfigPath handles the path subcommand.
func runConfigPath() error {
	configPath := getConfigFilePath()

	exists := "✗ (not created)"
	if _, err := os.Stat(configPath); err == nil {
		exists = "✓"
	}

	fmt.Printf("Configuration file: %s %s\n", configPath, exists)

	// Show XDG directories
	fmt.Println("\nXDG directories:")
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		fmt.Printf("  XDG_CONFIG_HOME: %s\n", xdgConfig)
	}
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		fmt.Printf("  XDG_DATA_HOME: %s\n", xdgData)
	}
	if xdgState := os.Getenv("XDG_STATE_HOME"); xdgState != "" {
		fmt.Printf("  XDG_STATE_HOME: %s\n", xdgState)
	}

	return nil
}
