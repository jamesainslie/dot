package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/jamesainslie/dot/internal/updater"
	"github.com/spf13/cobra"
)

// newUpgradeCommand creates the upgrade command.
func newUpgradeCommand(version string) *cobra.Command {
	var yes bool
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade dot to the latest version",
		Long: `Upgrade dot to the latest version using the configured package manager.

The upgrade command checks for new versions on GitHub and uses your system's
package manager to perform the upgrade. If no package manager is configured,
it will provide instructions for manual upgrade.

Configuration (in ~/.config/dot/config.yaml):
  update:
    package_manager: auto    # auto, brew, apt, yum, pacman, dnf, zypper, manual
    repository: jamesainslie/dot
    include_prerelease: false`,
		Example: `  # Check for and install updates
  dot upgrade

  # Check for updates without installing
  dot upgrade --check-only

  # Skip confirmation prompt
  dot upgrade --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(version, yes, checkOnly)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "Check for updates without installing")

	return cmd
}

// runUpgrade handles the upgrade command execution.
func runUpgrade(currentVersion string, yes, checkOnly bool) error {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		cfg = config.DefaultExtended()
	}

	fmt.Println("Checking for updates...")

	// Check for updates
	checker := updater.NewVersionChecker(cfg.Update.Repository)
	latestRelease, hasUpdate, err := checker.CheckForUpdate(currentVersion, cfg.Update.IncludePrerelease)
	if err != nil {
		return fmt.Errorf("check for updates: %w", err)
	}

	if !hasUpdate {
		fmt.Printf("%s You are already running the latest version (%s)\n",
			success("✓"), currentVersion)
		return nil
	}

	// Display update information
	displayUpdateInfo(currentVersion, latestRelease)

	if checkOnly {
		fmt.Printf("Run %s to upgrade.\n", accent("dot upgrade"))
		return nil
	}

	// Resolve package manager
	pkgMgr, err := updater.ResolvePackageManager(cfg.Update.PackageManager)
	if err != nil {
		return fmt.Errorf("resolve package manager: %w", err)
	}

	// Handle manual upgrade
	if pkgMgr.Name() == "manual" {
		displayManualInstructions(latestRelease.HTMLURL)
		return nil
	}

	// Show upgrade command
	cmd := pkgMgr.UpgradeCommand()
	fmt.Printf("Package manager: %s\n", accent(pkgMgr.Name()))
	fmt.Printf("Upgrade command: %s\n\n", dim(strings.Join(cmd, " ")))

	// Confirm upgrade
	if !yes && !confirmUpgrade() {
		fmt.Println("Upgrade cancelled.")
		return nil
	}

	// Execute upgrade
	fmt.Printf("\n%s Upgrading...\n\n", info("→"))
	if err := executeUpgradeCommand(cmd); err != nil {
		return err
	}

	fmt.Printf("\n%s Upgrade completed successfully!\n", success("✓"))
	fmt.Printf("Run %s to verify the new version.\n", accent("dot --version"))

	return nil
}

// loadConfig loads the configuration from the config file.
func loadConfig() (*config.ExtendedConfig, error) {
	configPath := getConfigFilePath()
	loader := config.NewLoader("dot", configPath)
	return loader.LoadWithEnv()
}

// displayUpdateInfo shows update information and release notes.
func displayUpdateInfo(currentVersion string, release *updater.GitHubRelease) {
	fmt.Printf("\n%s A new version is available!\n\n", info("ⓘ"))
	fmt.Printf("  Current version:  %s\n", accent(currentVersion))
	fmt.Printf("  Latest version:   %s\n", accent(release.TagName))
	fmt.Printf("  Release URL:      %s\n\n", dim(release.HTMLURL))

	if release.Body == "" {
		return
	}

	fmt.Println(bold("Release Notes:"))
	lines := strings.Split(release.Body, "\n")
	maxLines := 10
	if len(lines) > maxLines {
		for i := 0; i < maxLines; i++ {
			fmt.Printf("  %s\n", dim(lines[i]))
		}
		fmt.Printf("  %s\n\n", dim("..."))
	} else {
		for _, line := range lines {
			fmt.Printf("  %s\n", dim(line))
		}
		fmt.Println()
	}
}

// displayManualInstructions shows manual upgrade instructions.
func displayManualInstructions(releaseURL string) {
	fmt.Println(bold("Manual Upgrade Instructions:"))
	fmt.Printf("\n  Visit the release page to download the latest version:\n")
	fmt.Printf("  %s\n\n", accent(releaseURL))
}

// confirmUpgrade prompts the user for upgrade confirmation.
func confirmUpgrade() bool {
	fmt.Printf("Do you want to upgrade now? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// executeUpgradeCommand executes the upgrade command directly without shell invocation.
func executeUpgradeCommand(cmd []string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty upgrade command")
	}

	// Execute command directly (no shell invocation)
	// Command comes from trusted PackageManager interface, not user input
	// #nosec G204 -- Command source is PackageManager interface, not user-controlled
	upgradeCmd := exec.Command(cmd[0], cmd[1:]...)
	upgradeCmd.Stdout = os.Stdout
	upgradeCmd.Stderr = os.Stderr
	upgradeCmd.Stdin = os.Stdin

	if err := upgradeCmd.Run(); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	return nil
}
