package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/spf13/cobra"
)

// newCloneCommand creates the clone command.
func newCloneCommand() *cobra.Command {
	var (
		cloneProfile     string
		cloneInteractive bool
		cloneForce       bool
		cloneBranch      string
	)

	cmd := &cobra.Command{
		Use:   "clone <repository-url>",
		Short: "Clone dotfiles repository and install packages",
		Long: `Clone a dotfiles repository and install packages.

The clone command performs the following steps:
  1. Validates package directory is empty (unless --force is used)
  2. Clones the repository to the configured package directory
  3. Detects and uses repository configuration (.config/dot/config.yaml)
  4. Loads optional .dotbootstrap.yaml for package selection
  5. Selects packages to install:
     - Via named profile (--profile)
     - Interactively (--interactive or automatic terminal detection)
     - All packages (non-interactive mode)
  6. Filters packages by current platform
  7. Installs selected packages
  8. Updates manifest with repository tracking

Repository Configuration:
  If the repository contains .config/dot/config.yaml, it will be used
  automatically for all subsequent dot commands. This allows repositories
  to define their own management configuration without circular dependency.

  Example: ~/.dotfiles/.config/dot/config.yaml defines how the repository
  should be managed, and dot uses it automatically after clone.

Authentication:
  The command automatically resolves authentication in this order:
  1. GITHUB_TOKEN environment variable (for GitHub repositories)
  2. GIT_TOKEN environment variable (for general git repositories)
  3. SSH keys in ~/.ssh/ directory (id_rsa, id_ed25519)
  4. No authentication (public repositories only)

Bootstrap Configuration:
  If .dotbootstrap.yaml exists in the repository root, it defines:
  - Available packages with platform requirements
  - Named installation profiles
  - Default profile and conflict resolution policies

  Without bootstrap configuration, all discovered packages are offered.

Examples:
  # Clone and install all packages
  dot clone https://github.com/user/dotfiles

  # Clone specific branch
  dot clone https://github.com/user/dotfiles --branch develop

  # Use named profile from bootstrap config
  dot clone https://github.com/user/dotfiles --profile minimal

  # Force interactive selection
  dot clone https://github.com/user/dotfiles --interactive

  # Overwrite existing package directory
  dot clone https://github.com/user/dotfiles --force

  # Clone via SSH
  dot clone git@github.com:user/dotfiles.git`,
		Args: argsWithUsage(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(cmd, args, cloneProfile, cloneInteractive, cloneForce, cloneBranch)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVar(&cloneProfile, "profile", "", "installation profile from bootstrap config")
	cmd.Flags().BoolVar(&cloneInteractive, "interactive", false, "interactively select packages")
	cmd.Flags().BoolVar(&cloneForce, "force", false, "overwrite package directory if exists")
	cmd.Flags().StringVar(&cloneBranch, "branch", "", "branch to clone (defaults to repository default)")

	// Add bootstrap subcommand
	cmd.AddCommand(newCloneBootstrapCommand())

	return cmd
}

// runClone handles the clone command execution.
func runClone(cmd *cobra.Command, args []string, profile string, interactive bool, force bool, branch string) error {
	repoURL := args[0]

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

	// Build clone options
	opts := dot.CloneOptions{
		Profile:     profile,
		Interactive: interactive,
		Force:       force,
		Branch:      branch,
	}

	// Execute clone
	if err := client.Clone(ctx, repoURL, opts); err != nil {
		return formatCloneError(err)
	}

	return nil
}

// formatCloneError formats clone-specific errors with helpful messages.
func formatCloneError(err error) error {
	var packageDirNotEmpty dot.ErrPackageDirNotEmpty
	if errors.As(err, &packageDirNotEmpty) {
		return fmt.Errorf("%w\n\nUse --force to overwrite the existing directory", packageDirNotEmpty)
	}

	var bootstrapNotFound dot.ErrBootstrapNotFound
	if errors.As(err, &bootstrapNotFound) {
		return fmt.Errorf("%w\n\nThe repository may not have been properly cloned", bootstrapNotFound)
	}

	var invalidBootstrap dot.ErrInvalidBootstrap
	if errors.As(err, &invalidBootstrap) {
		return fmt.Errorf("%w\n\nCheck the .dotbootstrap.yaml syntax and validation rules", invalidBootstrap)
	}

	var authFailed dot.ErrAuthFailed
	if errors.As(err, &authFailed) {
		return fmt.Errorf("%w\n\nTry:\n  - Setting GITHUB_TOKEN environment variable\n  - Setting GIT_TOKEN environment variable\n  - Configuring SSH keys in ~/.ssh/", authFailed)
	}

	var cloneFailed dot.ErrCloneFailed
	if errors.As(err, &cloneFailed) {
		return fmt.Errorf("%w\n\nEnsure:\n  - URL is correct\n  - Repository is accessible\n  - Network connection is available\n  - Authentication is configured (for private repos)", cloneFailed)
	}

	var profileNotFound dot.ErrProfileNotFound
	if errors.As(err, &profileNotFound) {
		return fmt.Errorf("%w\n\nCheck available profiles in .dotbootstrap.yaml", profileNotFound)
	}

	return err
}
