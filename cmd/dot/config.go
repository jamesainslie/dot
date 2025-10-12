package main

import (
	"bytes"
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Only complete the first argument (key)
			if len(args) == 0 {
				return getValidConfigKeys(), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
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

// getValidConfigKeys returns all valid configuration keys for completion.
func getValidConfigKeys() []string {
	return []string{
		"directories.package",
		"directories.target",
		"directories.manifest",
		"logging.level",
		"logging.format",
		"logging.destination",
		"symlinks.mode",
		"symlinks.backup_suffix",
		"symlinks.backup_dir",
		"dotfile.prefix",
		"output.format",
		"output.color",
		"packages.sort_by",
	}
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Only complete the first argument (key)
			if len(args) == 0 {
				return getValidConfigKeys(), cobra.ShellCompDirectiveNoFileComp
			}
			// Don't complete the second argument (value)
			return nil, cobra.ShellCompDirectiveNoFileComp
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

	// Build the configuration table output
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s\n\n", dim("Configuration from: "+configPath))

	// Create table for each section
	sections := []struct {
		name   string
		render func(*bytes.Buffer, *config.ExtendedConfig)
	}{
		{"Directories", renderDirectoriesSection},
		{"Logging", renderLoggingSection},
		{"Symlinks", renderSymlinksSection},
		{"Ignore", renderIgnoreSection},
		{"Dotfile", renderDotfileSection},
		{"Output", renderOutputSection},
		{"Operations", renderOperationsSection},
		{"Packages", renderPackagesSection},
		{"Doctor", renderDoctorSection},
		{"Experimental", renderExperimentalSection},
	}

	for i, section := range sections {
		if i > 0 {
			buf.WriteString("\n")
		}
		section.render(&buf, cfg)
	}

	// Output directly (pagination will be added in future commit)
	fmt.Fprint(cmd.OutOrStdout(), buf.String())
	return nil
}

// renderDirectoriesSection renders the directories configuration table.
func renderDirectoriesSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Directories"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("package:"), cfg.Directories.Package)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("target:"), cfg.Directories.Target)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("manifest:"), cfg.Directories.Manifest)
}

// renderLoggingSection renders the logging configuration.
func renderLoggingSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Logging"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("level:"), cfg.Logging.Level)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("format:"), cfg.Logging.Format)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("destination:"), cfg.Logging.Destination)
	if cfg.Logging.File != "" {
		fmt.Fprintf(buf, "  %-20s %s\n", dim("file:"), cfg.Logging.File)
	}
}

// renderSymlinksSection renders the symlinks configuration.
func renderSymlinksSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Symlinks"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("mode:"), cfg.Symlinks.Mode)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("folding:"), formatBool(cfg.Symlinks.Folding))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("overwrite:"), formatBool(cfg.Symlinks.Overwrite))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("backup:"), formatBool(cfg.Symlinks.Backup))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("backup_suffix:"), cfg.Symlinks.BackupSuffix)
	if cfg.Symlinks.BackupDir != "" {
		fmt.Fprintf(buf, "  %-20s %s\n", dim("backup_dir:"), cfg.Symlinks.BackupDir)
	}
}

// renderIgnoreSection renders the ignore configuration section.
func renderIgnoreSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Ignore"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("use_defaults:"), formatBool(cfg.Ignore.UseDefaults))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("patterns:"), formatSlice(cfg.Ignore.Patterns))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("overrides:"), formatSlice(cfg.Ignore.Overrides))
}

// renderDotfileSection renders the dotfile configuration section.
func renderDotfileSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Dotfile"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("translate:"), formatBool(cfg.Dotfile.Translate))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("prefix:"), cfg.Dotfile.Prefix)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("package_name_mapping:"), formatBool(cfg.Dotfile.PackageNameMapping))
}

// renderOutputSection renders the output configuration section.
func renderOutputSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Output"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("format:"), cfg.Output.Format)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("color:"), cfg.Output.Color)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("progress:"), formatBool(cfg.Output.Progress))
	fmt.Fprintf(buf, "  %-20s %d\n", dim("verbosity:"), cfg.Output.Verbosity)
	fmt.Fprintf(buf, "  %-20s %d\n", dim("width:"), cfg.Output.Width)
}

// renderOperationsSection renders the operations configuration section.
func renderOperationsSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Operations"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("dry_run:"), formatBool(cfg.Operations.DryRun))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("atomic:"), formatBool(cfg.Operations.Atomic))
	fmt.Fprintf(buf, "  %-20s %d\n", dim("max_parallel:"), cfg.Operations.MaxParallel)
}

// renderPackagesSection renders the packages configuration section.
func renderPackagesSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Packages"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("sort_by:"), cfg.Packages.SortBy)
	fmt.Fprintf(buf, "  %-20s %s\n", dim("auto_discover:"), formatBool(cfg.Packages.AutoDiscover))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("validate_names:"), formatBool(cfg.Packages.ValidateNames))
}

// renderDoctorSection renders the doctor configuration section.
func renderDoctorSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Doctor"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("auto_fix:"), formatBool(cfg.Doctor.AutoFix))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("check_manifest:"), formatBool(cfg.Doctor.CheckManifest))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("check_broken_links:"), formatBool(cfg.Doctor.CheckBrokenLinks))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("check_orphaned:"), formatBool(cfg.Doctor.CheckOrphaned))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("check_permissions:"), formatBool(cfg.Doctor.CheckPermissions))
}

// renderExperimentalSection renders the experimental configuration section.
func renderExperimentalSection(buf *bytes.Buffer, cfg *config.ExtendedConfig) {
	fmt.Fprintf(buf, "%s\n", bold("Experimental"))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("parallel:"), formatBool(cfg.Experimental.Parallel))
	fmt.Fprintf(buf, "  %-20s %s\n", dim("profiling:"), formatBool(cfg.Experimental.Profiling))
}

// formatBool formats a boolean value for display.
func formatBool(b bool) string {
	if b {
		return success("true")
	}
	return dim("false")
}

// formatSlice formats a string slice for display.
func formatSlice(s []string) string {
	if len(s) == 0 {
		return dim("(none)")
	}
	if len(s) <= 3 {
		return strings.Join(s, ", ")
	}
	return strings.Join(s[:3], ", ") + dim(fmt.Sprintf(" (+%d more)", len(s)-3))
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
