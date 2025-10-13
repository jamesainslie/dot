package updater

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStartupChecker(t *testing.T) {
	cfg := config.DefaultExtended()
	var buf bytes.Buffer

	sc := NewStartupChecker("1.0.0", cfg, "/test/config", &buf)
	require.NotNil(t, sc)
	assert.Equal(t, "1.0.0", sc.currentVersion)
	assert.NotNil(t, sc.config)
	assert.NotNil(t, sc.stateManager)
	assert.NotNil(t, sc.checker)
	assert.Equal(t, &buf, sc.output)
}

func TestStartupChecker_Check_Disabled(t *testing.T) {
	cfg := config.DefaultExtended()
	cfg.Update.CheckOnStartup = false

	tmpDir := t.TempDir()
	var buf bytes.Buffer
	sc := NewStartupChecker("1.0.0", cfg, tmpDir, &buf)

	result, err := sc.Check()
	require.NoError(t, err)
	assert.True(t, result.SkipCheck)
	assert.False(t, result.UpdateAvailable)
}

func TestStartupChecker_Check_ZeroFrequency(t *testing.T) {
	cfg := config.DefaultExtended()
	cfg.Update.CheckOnStartup = true
	cfg.Update.CheckFrequency = 0 // Disabled

	tmpDir := t.TempDir()
	var buf bytes.Buffer
	sc := NewStartupChecker("1.0.0", cfg, tmpDir, &buf)

	result, err := sc.Check()
	require.NoError(t, err)
	assert.True(t, result.SkipCheck)
}

func TestStartupChecker_Check_NotDueYet(t *testing.T) {
	cfg := config.DefaultExtended()
	cfg.Update.CheckOnStartup = true
	cfg.Update.CheckFrequency = 24

	tmpDir := t.TempDir()
	var buf bytes.Buffer
	sc := NewStartupChecker("1.0.0", cfg, tmpDir, &buf)

	// Record a recent check
	err := sc.stateManager.RecordCheck()
	require.NoError(t, err)

	result, err := sc.Check()
	require.NoError(t, err)
	assert.True(t, result.SkipCheck)
}

func TestStartupChecker_Check_NetworkError(t *testing.T) {
	cfg := config.DefaultExtended()
	cfg.Update.CheckOnStartup = true
	cfg.Update.CheckFrequency = 24
	cfg.Update.Repository = "invalid/repo-xyz-123"

	tmpDir := t.TempDir()
	var buf bytes.Buffer
	sc := NewStartupChecker("1.0.0", cfg, tmpDir, &buf)
	sc.checker.httpClient.Timeout = 100 * time.Millisecond

	// Set last check to old time so check is due
	state := &CheckState{
		LastCheck: time.Now().Add(-25 * time.Hour),
	}
	err := sc.stateManager.Save(state)
	require.NoError(t, err)

	// Should not fail, just skip silently
	result, err := sc.Check()
	require.NoError(t, err)
	assert.True(t, result.SkipCheck)
}

func TestStartupChecker_ShowNotification(t *testing.T) {
	cfg := config.DefaultExtended()
	var buf bytes.Buffer

	sc := NewStartupChecker("1.0.0", cfg, "/test/config", &buf)

	t.Run("skip when no update", func(t *testing.T) {
		buf.Reset()
		result := &CheckResult{
			UpdateAvailable: false,
			SkipCheck:       false,
		}

		sc.ShowNotification(result)
		assert.Empty(t, buf.String())
	})

	t.Run("skip when check skipped", func(t *testing.T) {
		buf.Reset()
		result := &CheckResult{
			UpdateAvailable: true,
			SkipCheck:       true,
		}

		sc.ShowNotification(result)
		assert.Empty(t, buf.String())
	})

	t.Run("show when update available", func(t *testing.T) {
		buf.Reset()
		result := &CheckResult{
			UpdateAvailable: true,
			LatestVersion:   "v2.0.0",
			ReleaseURL:      "https://github.com/owner/repo/releases/tag/v2.0.0",
			SkipCheck:       false,
		}

		sc.ShowNotification(result)
		output := buf.String()

		assert.NotEmpty(t, output)
		assert.Contains(t, output, "new version")
		assert.Contains(t, output, "1.0.0")  // current version
		assert.Contains(t, output, "v2.0.0") // latest version
		assert.Contains(t, output, "dot upgrade")
		assert.Contains(t, output, "┌") // box drawing characters
		assert.Contains(t, output, "└")
	})
}

func TestCheckResult_Structure(t *testing.T) {
	result := &CheckResult{
		UpdateAvailable: true,
		LatestVersion:   "1.2.3",
		ReleaseURL:      "https://example.com",
		SkipCheck:       false,
	}

	assert.True(t, result.UpdateAvailable)
	assert.Equal(t, "1.2.3", result.LatestVersion)
	assert.Equal(t, "https://example.com", result.ReleaseURL)
	assert.False(t, result.SkipCheck)
}

func TestStartupChecker_ShowNotification_Format(t *testing.T) {
	cfg := config.DefaultExtended()
	var buf bytes.Buffer

	sc := NewStartupChecker("v1.0.0", cfg, "/test/config", &buf)

	result := &CheckResult{
		UpdateAvailable: true,
		LatestVersion:   "v2.5.10",
		ReleaseURL:      "https://github.com/test/repo/releases/tag/v2.5.10",
		SkipCheck:       false,
	}

	sc.ShowNotification(result)
	output := buf.String()

	// Verify box drawing characters are present
	assert.Contains(t, output, "┌")
	assert.Contains(t, output, "└")
	assert.Contains(t, output, "│")

	// Verify content is aligned
	lines := strings.Split(output, "\n")
	assert.True(t, len(lines) > 5, "should have multiple lines")
}

func TestStartupChecker_ShowNotification_LongVersions(t *testing.T) {
	cfg := config.DefaultExtended()
	var buf bytes.Buffer

	// Test with very long version strings
	sc := NewStartupChecker("v1.0.0-very-long-version-string-that-exceeds-limit", cfg, "/test/config", &buf)

	result := &CheckResult{
		UpdateAvailable: true,
		LatestVersion:   "v2.0.0-another-very-long-version-string",
		ReleaseURL:      "https://github.com/test/repo/releases",
		SkipCheck:       false,
	}

	sc.ShowNotification(result)
	output := buf.String()

	// Verify truncation occurred
	assert.Contains(t, output, "...")
	
	// Verify box is properly formed
	assert.Contains(t, output, "┌")
	assert.Contains(t, output, "└")
	assert.Contains(t, output, "│")
	
	// Verify content is present
	assert.Contains(t, output, "new version")
	assert.Contains(t, output, "Current:")
	assert.Contains(t, output, "Latest:")
	assert.Contains(t, output, "dot upgrade")
	
	// Verify no lines are excessively long
	// Note: UTF-8 box drawing characters may have byte lengths different from display width
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) > 0 {
			// Allow for UTF-8 encoding overhead
			assert.LessOrEqual(t, len(line), 200, "line should not be excessively long: %q", line)
		}
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"no escape codes",
			"plain text",
			"plain text",
		},
		{
			"with color code",
			"\033[38;5;71mgreen text\033[0m",
			"green text",
		},
		{
			"with bold",
			"\033[1mbold text\033[0m",
			"bold text",
		},
		{
			"mixed codes",
			"\033[1m\033[38;5;109mcyan bold\033[0m text",
			"cyan bold text",
		},
		{
			"empty string",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestColorize(t *testing.T) {
	// Save and restore NO_COLOR
	oldNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if oldNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", oldNoColor)
		}
	}()

	t.Run("with NO_COLOR set", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		result := colorize(colorCyan, "test")
		assert.Equal(t, "test", result, "should not add color when NO_COLOR is set")
	})

	t.Run("with NO_COLOR unset", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		// Note: This will still return plain text if not running in a terminal
		// but we're testing the NO_COLOR logic works
		result := colorize(colorCyan, "test")
		// Either colored or plain depending on terminal
		assert.NotEmpty(t, result)
	})
}

func TestShouldUseColor(t *testing.T) {
	// Save and restore NO_COLOR
	oldNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if oldNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", oldNoColor)
		}
	}()

	t.Run("respects NO_COLOR", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		assert.False(t, shouldUseColor())
	})

	t.Run("without NO_COLOR", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		// Result depends on whether stdout is a terminal
		// Just verify it doesn't panic
		_ = shouldUseColor()
	})
}
