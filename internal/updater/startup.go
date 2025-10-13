package updater

import (
	"fmt"
	"io"
	"time"

	"github.com/jamesainslie/dot/internal/config"
)

// StartupChecker performs update checks at application startup.
type StartupChecker struct {
	currentVersion string
	config         *config.ExtendedConfig
	stateManager   *StateManager
	checker        *VersionChecker
	output         io.Writer
}

// NewStartupChecker creates a new startup checker.
func NewStartupChecker(currentVersion string, cfg *config.ExtendedConfig, configDir string, output io.Writer) *StartupChecker {
	return &StartupChecker{
		currentVersion: currentVersion,
		config:         cfg,
		stateManager:   NewStateManager(configDir),
		checker:        NewVersionChecker(cfg.Update.Repository),
		output:         output,
	}
}

// CheckResult contains the result of an update check.
type CheckResult struct {
	UpdateAvailable bool
	LatestVersion   string
	ReleaseURL      string
	SkipCheck       bool
}

// Check performs a startup update check if configured and due.
func (sc *StartupChecker) Check() (*CheckResult, error) {
	// If checking is disabled, skip
	if !sc.config.Update.CheckOnStartup {
		return &CheckResult{SkipCheck: true}, nil
	}

	// Check if we should perform a check based on frequency
	frequency := time.Duration(sc.config.Update.CheckFrequency) * time.Hour
	if frequency <= 0 {
		// Disabled via frequency
		return &CheckResult{SkipCheck: true}, nil
	}

	shouldCheck, err := sc.stateManager.ShouldCheck(frequency)
	if err != nil {
		// Don't fail startup on state file errors
		return &CheckResult{SkipCheck: true}, nil
	}

	if !shouldCheck {
		return &CheckResult{SkipCheck: true}, nil
	}

	// Perform the check
	latestRelease, hasUpdate, err := sc.checker.CheckForUpdate(
		sc.currentVersion,
		sc.config.Update.IncludePrerelease,
	)
	if err != nil {
		// Don't fail startup on check errors - just skip silently
		return &CheckResult{SkipCheck: true}, nil
	}

	// Record that we checked
	if err := sc.stateManager.RecordCheck(); err != nil {
		// Non-fatal error
		_ = err
	}

	if !hasUpdate {
		return &CheckResult{
			UpdateAvailable: false,
			SkipCheck:       false,
		}, nil
	}

	return &CheckResult{
		UpdateAvailable: true,
		LatestVersion:   latestRelease.TagName,
		ReleaseURL:      latestRelease.HTMLURL,
		SkipCheck:       false,
	}, nil
}

// ShowNotification displays an update notification to the user.
func (sc *StartupChecker) ShowNotification(result *CheckResult) {
	if result.SkipCheck || !result.UpdateAvailable {
		return
	}

	// Truncate version strings if too long
	current := sc.currentVersion
	if len(current) > 20 {
		current = current[:17] + "..."
	}
	latest := result.LatestVersion
	if len(latest) > 20 {
		latest = latest[:17] + "..."
	}

	const boxWidth = 57 // Interior width
	
	fmt.Fprintf(sc.output, "\n")
	fmt.Fprintf(sc.output, "┌─────────────────────────────────────────────────────────┐\n")
	fmt.Fprintf(sc.output, "│  A new version of dot is available!                    │\n")
	fmt.Fprintf(sc.output, "│                                                         │\n")
	
	// Format version lines with proper padding
	currentLine := fmt.Sprintf("  Current: %-20s", current)
	fmt.Fprintf(sc.output, "│%-57s│\n", currentLine)
	
	latestLine := fmt.Sprintf("  Latest:  %-20s", latest)
	fmt.Fprintf(sc.output, "│%-57s│\n", latestLine)
	
	fmt.Fprintf(sc.output, "│                                                         │\n")
	fmt.Fprintf(sc.output, "│  Run 'dot upgrade' to update                            │\n")
	fmt.Fprintf(sc.output, "└─────────────────────────────────────────────────────────┘\n")
	fmt.Fprintf(sc.output, "\n")
}

