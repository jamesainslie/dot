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
	configPath := getConfigFilePath()
	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	if err != nil {
		// Use defaults if config load fails
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
	fmt.Printf("\n%s A new version is available!\n\n", info("ⓘ"))
	fmt.Printf("  Current version:  %s\n", accent(currentVersion))
	fmt.Printf("  Latest version:   %s\n", accent(latestRelease.TagName))
	fmt.Printf("  Release URL:      %s\n\n", dim(latestRelease.HTMLURL))

	if latestRelease.Body != "" {
		fmt.Println(bold("Release Notes:"))
		// Show first few lines of release notes
		lines := strings.Split(latestRelease.Body, "\n")
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
		fmt.Println(bold("Manual Upgrade Instructions:"))
		fmt.Printf("\n  Visit the release page to download the latest version:\n")
		fmt.Printf("  %s\n\n", accent(latestRelease.HTMLURL))
		return nil
	}

	// Show upgrade command
	cmd := pkgMgr.UpgradeCommand()
	fmt.Printf("Package manager: %s\n", accent(pkgMgr.Name()))
	fmt.Printf("Upgrade command: %s\n\n", dim(strings.Join(cmd, " ")))

	// Confirm upgrade
	if !yes {
		fmt.Printf("Do you want to upgrade now? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	// Execute upgrade
	fmt.Printf("\n%s Upgrading...\n\n", info("→"))

	// Handle compound commands (with &&)
	if len(cmd) > 1 && contains(cmd, "&&") {
		// Execute as shell command
		shellCmd := exec.Command("sh", "-c", strings.Join(cmd, " "))
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr
		shellCmd.Stdin = os.Stdin

		if err := shellCmd.Run(); err != nil {
			return fmt.Errorf("upgrade failed: %w", err)
		}
	} else {
		// Execute as direct command
		upgradeCmd := exec.Command(cmd[0], cmd[1:]...)
		upgradeCmd.Stdout = os.Stdout
		upgradeCmd.Stderr = os.Stderr
		upgradeCmd.Stdin = os.Stdin

		if err := upgradeCmd.Run(); err != nil {
			return fmt.Errorf("upgrade failed: %w", err)
		}
	}

	fmt.Printf("\n%s Upgrade completed successfully!\n", success("✓"))
	fmt.Printf("Run %s to verify the new version.\n", accent("dot --version"))

	return nil
}

// contains checks if a slice contains a string.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

