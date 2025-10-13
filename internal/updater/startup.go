package updater

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jamesainslie/dot/internal/config"
	"golang.org/x/term"
)

// StartupChecker performs update checks at application startup.
type StartupChecker struct {
	currentVersion string
	config         *config.ExtendedConfig
	stateManager   *StateManager
	checker        *VersionChecker
	output         io.Writer
	useColor       bool
}

// NewStartupChecker creates a new startup checker.
func NewStartupChecker(currentVersion string, cfg *config.ExtendedConfig, configDir string, output io.Writer) *StartupChecker {
	return &StartupChecker{
		currentVersion: currentVersion,
		config:         cfg,
		stateManager:   NewStateManager(configDir),
		checker:        NewVersionChecker(cfg.Update.Repository),
		output:         output,
		useColor:       detectColor(output),
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
	cf := sc.config.Update.CheckFrequency
	if cf < 0 {
		// Disabled via frequency (-1)
		return &CheckResult{SkipCheck: true}, nil
	}

	var shouldCheck bool
	if cf == 0 {
		// Always check (0 means check every time)
		shouldCheck = true
	} else {
		// Use frequency-based checking
		frequency := time.Duration(cf) * time.Hour
		var err error
		shouldCheck, err = sc.stateManager.ShouldCheck(frequency)
		if err != nil {
			// Don't fail startup on state file errors
			return &CheckResult{SkipCheck: true}, nil
		}
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

// Color codes for terminal output
const (
	colorCyan   = "\033[38;5;109m" // Muted cyan for accents
	colorGreen  = "\033[38;5;71m"  // Muted green for success
	colorYellow = "\033[38;5;179m" // Muted yellow for version highlight
	colorGray   = "\033[38;5;245m" // Muted gray for box
	colorBold   = "\033[1m"
	colorReset  = "\033[0m"
)

// detectColor determines if color output should be enabled for the given writer
func detectColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// Check if the writer has an Fd() (e.g., *os.File)
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// colorize applies color if enabled
func (sc *StartupChecker) colorize(color, text string) string {
	if !sc.useColor {
		return text
	}
	return color + text + colorReset
}

// ShowNotification displays an update notification to the user.
func (sc *StartupChecker) ShowNotification(result *CheckResult) {
	if result.SkipCheck || !result.UpdateAvailable {
		return
	}

	// Truncate version strings if too long (use rune-based slicing for UTF-8 safety)
	current := sc.currentVersion
	if len([]rune(current)) > 20 {
		runes := []rune(current)
		current = string(runes[:17]) + "..."
	}
	latest := result.LatestVersion
	if len([]rune(latest)) > 20 {
		runes := []rune(latest)
		latest = string(runes[:17]) + "..."
	}

	// Box drawing characters (always visible)
	boxColor := colorGray
	topLeft := sc.colorize(boxColor, "┌")
	topRight := sc.colorize(boxColor, "┐")
	bottomLeft := sc.colorize(boxColor, "└")
	bottomRight := sc.colorize(boxColor, "┘")
	vertical := sc.colorize(boxColor, "│")
	horizontal := sc.colorize(boxColor, "─────────────────────────────────────────────────────────")

	fmt.Fprintf(sc.output, "\n")
	fmt.Fprintf(sc.output, "%s%s%s\n", topLeft, horizontal, topRight)

	// Title line with bold and cyan
	title := sc.colorize(colorBold+colorCyan, "A new version of dot is available!")
	titleLen := len(stripANSI(title))
	titlePad := 57 - titleLen - 2 // -2 for "  " prefix
	fmt.Fprintf(sc.output, "%s  %s%*s%s\n", vertical, title, titlePad, "", vertical)

	// Empty line
	fmt.Fprintf(sc.output, "%s%-57s%s\n", vertical, "", vertical)

	// Current version line
	currentLabel := sc.colorize(colorGray, "Current:")
	currentVer := sc.colorize(colorYellow, current)
	currentLine := fmt.Sprintf("  %s %s", currentLabel, currentVer)
	currentLen := len(stripANSI(currentLine))
	currentPad := 57 - currentLen
	fmt.Fprintf(sc.output, "%s%s%*s%s\n", vertical, currentLine, currentPad, "", vertical)

	// Latest version line
	latestLabel := sc.colorize(colorGray, "Latest:")
	latestVer := sc.colorize(colorGreen, latest)
	latestLine := fmt.Sprintf("  %s  %s", latestLabel, latestVer)
	latestLen := len(stripANSI(latestLine))
	latestPad := 57 - latestLen
	fmt.Fprintf(sc.output, "%s%s%*s%s\n", vertical, latestLine, latestPad, "", vertical)

	// Empty line
	fmt.Fprintf(sc.output, "%s%-57s%s\n", vertical, "", vertical)

	// Upgrade message
	upgradeCmd := sc.colorize(colorCyan, "'dot upgrade'")
	upgradeMsg := fmt.Sprintf("  Run %s to update", upgradeCmd)
	upgradeMsgLen := len(stripANSI(upgradeMsg))
	upgradePad := 57 - upgradeMsgLen
	fmt.Fprintf(sc.output, "%s%s%*s%s\n", vertical, upgradeMsg, upgradePad, "", vertical)

	fmt.Fprintf(sc.output, "%s%s%s\n", bottomLeft, horizontal, bottomRight)
	fmt.Fprintf(sc.output, "\n")
}

// stripANSI removes ANSI escape codes for length calculation
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
