package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaklabco/dot/internal/bootstrap"
	"github.com/yaklabco/dot/pkg/dot"
)

// newCloneBootstrapCommand creates the clone bootstrap subcommand.
func newCloneBootstrapCommand() *cobra.Command {
	var (
		outputPath     string
		dryRun         bool
		fromManifest   bool
		conflictPolicy string
		force          bool
	)

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Generate bootstrap configuration from installation",
		Long: `Generate a .dotbootstrap.yaml configuration file from current dotfiles installation.

This command analyzes the packages in your package directory and creates a
bootstrap configuration file that can be committed to your repository. The
generated configuration allows others to clone your dotfiles with predefined
package selections and profiles.

The command discovers all packages in the package directory and creates
a bootstrap configuration with sensible defaults. You should review and
customize the generated file before committing it.

Output:
  The generated configuration includes:
  - All discovered packages with required: false
  - Default conflict resolution policy
  - Example profile structures
  - Helpful comments for customization

Repository layout:
  If package contents are stored as real dotfile paths (for example
  nvim/.config/nvim) rather than under dot- prefixed names, the command
  also writes .config/dot/config.yaml in the package directory with
  dotfile.package_name_mapping set to false. An existing file at that
  path is left untouched.

Examples:
  # Generate bootstrap config in package directory
  dot clone bootstrap

  # Specify custom output location
  dot clone bootstrap --output ~/dotfiles/.dotbootstrap.yaml

  # Preview without writing file
  dot clone bootstrap --dry-run

  # Only include packages from manifest
  dot clone bootstrap --from-manifest

  # Set default conflict policy
  dot clone bootstrap --conflict-policy backup

  # Overwrite existing file
  dot clone bootstrap --force`,
		Args: argsWithUsage(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCloneBootstrap(cmd, outputPath, dryRun, fromManifest, conflictPolicy, force)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path (default: .dotbootstrap.yaml in package dir)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print config to stdout instead of writing file")
	cmd.Flags().BoolVar(&fromManifest, "from-manifest", false, "only include packages from manifest")
	cmd.Flags().StringVar(&conflictPolicy, "conflict-policy", "", "default conflict policy (backup, fail, overwrite, skip)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing bootstrap file")

	return cmd
}

// runCloneBootstrap executes the bootstrap generation command.
func runCloneBootstrap(cmd *cobra.Command, outputPath string, dryRun bool, fromManifest bool, conflictPolicy string, force bool) error {
	// Build config
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return formatError(err)
	}

	// Create client
	client, err := dot.NewClient(cfg)
	if err != nil {
		return formatError(err)
	}

	// Get context
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Build generation options
	opts := dot.GenerateBootstrapOptions{
		FromManifest:    fromManifest,
		ConflictPolicy:  conflictPolicy,
		IncludeComments: true, // Always include helpful comments
		Force:           force,
	}

	// Generate bootstrap configuration
	result, err := client.GenerateBootstrap(ctx, opts)
	if err != nil {
		return formatBootstrapError(err)
	}

	// Dry run mode - print to command output
	if dryRun {
		cmd.Print(string(result.YAML))
		return nil
	}

	// Determine output path
	if outputPath == "" {
		outputPath = filepath.Join(cfg.PackageDir, ".dotbootstrap.yaml")
	}

	// Handle force flag - delete existing file
	if force {
		// Check if file exists
		if cfg.FS.Exists(ctx, outputPath) {
			if err := cfg.FS.Remove(ctx, outputPath); err != nil {
				return fmt.Errorf("cannot remove existing file: %w", err)
			}
		}
	}

	// Write configuration
	if err := client.WriteBootstrap(ctx, result.YAML, outputPath); err != nil {
		return formatBootstrapError(err)
	}

	// Full-tree repositories need package name mapping turned off
	repoConfigPath, err := writeRepoConfigForFullTree(ctx, cfg)
	if err != nil {
		return err
	}
	if repoConfigPath != "" {
		cmd.Printf("Repository configuration written to: %s\n", repoConfigPath)
	}

	// Success message
	cmd.Printf("Bootstrap configuration written to: %s\n", outputPath)
	cmd.Printf("  Packages: %d\n", result.PackageCount)
	if result.InstalledCount > 0 {
		cmd.Printf("  Installed: %d\n", result.InstalledCount)
	}
	cmd.Println("\nReview and customize the configuration before committing.")

	return nil
}

// writeRepoConfigForFullTree writes a repository configuration when the
// package directory uses the full-tree layout.
//
// Full-tree repositories store package contents as real dotfile paths, so
// package name mapping has to be off for every dot command run against them.
// The setting is recorded in <repo>/.config/dot/config.yaml. An existing file
// is never modified.
//
// Returns the path written, or an empty string when nothing was written.
func writeRepoConfigForFullTree(ctx context.Context, cfg dot.Config) (string, error) {
	layout, err := detectRepoLayout(ctx, cfg)
	if err != nil {
		return "", err
	}
	if layout != bootstrap.LayoutFullTree {
		return "", nil
	}

	repoConfigPath := filepath.Join(cfg.PackageDir, ".config", "dot", "config.yaml")
	if cfg.FS.Exists(ctx, repoConfigPath) {
		return "", nil
	}

	if err := cfg.FS.MkdirAll(ctx, filepath.Dir(repoConfigPath), 0755); err != nil {
		return "", fmt.Errorf("create repository config directory: %w", err)
	}
	if err := cfg.FS.WriteFile(ctx, repoConfigPath, bootstrap.RepoConfigYAML(), 0644); err != nil {
		return "", fmt.Errorf("write repository config: %w", err)
	}

	return repoConfigPath, nil
}

// detectRepoLayout classifies the package directory by reading each package's
// top-level entries.
func detectRepoLayout(ctx context.Context, cfg dot.Config) (bootstrap.Layout, error) {
	entries, err := cfg.FS.ReadDir(ctx, cfg.PackageDir)
	if err != nil {
		return "", fmt.Errorf("read package directory: %w", err)
	}

	packages := make(map[string][]string, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		contents, err := cfg.FS.ReadDir(ctx, filepath.Join(cfg.PackageDir, entry.Name()))
		if err != nil {
			return "", fmt.Errorf("read package %s: %w", entry.Name(), err)
		}

		names := make([]string, 0, len(contents))
		for _, item := range contents {
			names = append(names, item.Name())
		}
		packages[entry.Name()] = names
	}

	return bootstrap.DetectLayout(packages), nil
}

// formatBootstrapError formats bootstrap-specific errors with helpful messages.
func formatBootstrapError(err error) error {
	var bootstrapExists dot.ErrBootstrapExists
	if errors.As(err, &bootstrapExists) {
		return fmt.Errorf("%w\n\nUse --force to overwrite the existing file", bootstrapExists)
	}

	var pkgNotFound dot.ErrPackageNotFound
	if errors.As(err, &pkgNotFound) {
		return fmt.Errorf("%w\n\nEnsure packages exist in the package directory", pkgNotFound)
	}

	return err
}
