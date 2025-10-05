package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVerboseLogger(t *testing.T) {
	logger := NewVerboseLogger(2, true, false)
	assert.NotNil(t, logger)
	assert.Equal(t, 2, logger.level)
	assert.True(t, logger.colorEnabled)
	assert.False(t, logger.quiet)
}

func TestVerboseLogger_Debug(t *testing.T) {
	logger := NewVerboseLogger(3, false, false)
	// Should not panic
	logger.Debug("debug message: %d", 42)
}

func TestVerboseLogger_Debug_InsufficientLevel(t *testing.T) {
	logger := NewVerboseLogger(2, false, false)
	// Should not print but also not panic
	logger.Debug("debug message")
}

func TestVerboseLogger_Info(t *testing.T) {
	logger := NewVerboseLogger(2, false, false)
	// Should not panic
	logger.Info("info message: %s", "test")
}

func TestVerboseLogger_Info_InsufficientLevel(t *testing.T) {
	logger := NewVerboseLogger(1, false, false)
	// Should not print but also not panic
	logger.Info("info message")
}

func TestVerboseLogger_Summary(t *testing.T) {
	logger := NewVerboseLogger(1, false, false)
	// Should not panic
	logger.Summary("summary: %d items", 5)
}

func TestVerboseLogger_Summary_InsufficientLevel(t *testing.T) {
	logger := NewVerboseLogger(0, false, false)
	// Should not print but also not panic
	logger.Summary("summary")
}

func TestVerboseLogger_Always(t *testing.T) {
	logger := NewVerboseLogger(0, false, false)
	// Should always print (unless quiet)
	logger.Always("always printed")
}

func TestVerboseLogger_Always_Quiet(t *testing.T) {
	logger := NewVerboseLogger(0, false, true)
	// Should not print in quiet mode
	logger.Always("should be suppressed")
}

func TestVerboseLogger_QuietMode(t *testing.T) {
	logger := NewVerboseLogger(3, false, true)

	// None of these should print in quiet mode
	logger.Debug("debug")
	logger.Info("info")
	logger.Summary("summary")
	logger.Always("always")
}

func TestVerboseLogger_IsQuiet(t *testing.T) {
	tests := []struct {
		name  string
		quiet bool
	}{
		{"quiet", true},
		{"not quiet", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewVerboseLogger(1, false, tt.quiet)
			assert.Equal(t, tt.quiet, logger.IsQuiet())
		})
	}
}

func TestVerboseLogger_Level(t *testing.T) {
	tests := []struct {
		name  string
		level int
	}{
		{"level 0", 0},
		{"level 1", 1},
		{"level 2", 2},
		{"level 3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewVerboseLogger(tt.level, false, false)
			assert.Equal(t, tt.level, logger.Level())
		})
	}
}

func TestVerboseLogger_WithColor(t *testing.T) {
	logger := NewVerboseLogger(3, true, false)

	// Should not panic with color enabled
	logger.Debug("colored debug")
	logger.Info("colored info")
	logger.Summary("colored summary")
	logger.Always("colored always")
}

func TestVerboseLogger_WithoutColor(t *testing.T) {
	logger := NewVerboseLogger(3, false, false)

	// Should not panic without color
	logger.Debug("plain debug")
	logger.Info("plain info")
	logger.Summary("plain summary")
	logger.Always("plain always")
}

func TestVerboseLogger_AllLevels(t *testing.T) {
	// Test that higher levels enable lower level messages
	tests := []struct {
		level         int
		debugPrints   bool
		infoPrints    bool
		summaryPrints bool
	}{
		{level: 0, debugPrints: false, infoPrints: false, summaryPrints: false},
		{level: 1, debugPrints: false, infoPrints: false, summaryPrints: true},
		{level: 2, debugPrints: false, infoPrints: true, summaryPrints: true},
		{level: 3, debugPrints: true, infoPrints: true, summaryPrints: true},
	}

	for _, tt := range tests {
		logger := NewVerboseLogger(tt.level, false, false)
		assert.Equal(t, tt.level, logger.Level())

		// These should all not panic regardless of level
		logger.Debug("test")
		logger.Info("test")
		logger.Summary("test")
	}
}
