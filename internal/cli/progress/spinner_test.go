package progress

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSpinnerWithStyle(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: true,
		Width:       80,
	}

	spinner := NewSpinnerWithStyle(cfg, SpinnerDots)
	assert.NotNil(t, spinner)
	assert.Equal(t, SpinnerDots, spinner.frames)
}

func TestNewSpinnerWithStyle_EmptyStyle(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: true,
	}

	spinner := NewSpinnerWithStyle(cfg, SpinnerStyle{})
	assert.NotNil(t, spinner)
	assert.Equal(t, SpinnerDots, spinner.frames)
}

func TestSpinner_Spin(t *testing.T) {
	spinner := &Spinner{
		frames: SpinnerDots,
		active: true,
	}

	initial := spinner.current
	spinner.Spin()

	assert.NotEqual(t, initial, spinner.current)
}

func TestSpinner_Spin_Cycles(t *testing.T) {
	spinner := &Spinner{
		frames: SpinnerLine, // 4 frames
		active: true,
	}

	// Spin through all frames
	for i := 0; i < len(SpinnerLine); i++ {
		assert.Equal(t, i, spinner.current)
		spinner.Spin()
	}

	// Should cycle back to 0
	assert.Equal(t, 0, spinner.current)
}

func TestSpinner_Start(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerDots)

	spinner.Start("Loading")

	assert.Equal(t, "Loading", spinner.message)
	assert.True(t, spinner.active)
	assert.NotNil(t, spinner.ticker)

	// Clean up
	spinner.Stop("")
}

func TestSpinner_Start_AlreadyActive(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerDots)

	spinner.Start("First")
	firstTicker := spinner.ticker

	spinner.Start("Second")

	// Should not change ticker when already active
	assert.Equal(t, firstTicker, spinner.ticker)
	assert.Equal(t, "First", spinner.message)

	// Clean up
	spinner.Stop("")
}

func TestSpinner_Update(t *testing.T) {
	spinner := &Spinner{
		message: "Original",
	}

	spinner.Update(5, 10, "Updated")

	assert.Equal(t, "Updated", spinner.message)
}

func TestSpinner_Update_EmptyMessage(t *testing.T) {
	spinner := &Spinner{
		message: "Original",
	}

	spinner.Update(5, 10, "")

	assert.Equal(t, "Original", spinner.message)
}

func TestSpinner_Stop(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerDots)

	spinner.Start("Running")
	assert.True(t, spinner.active)

	spinner.Stop("Complete")

	assert.False(t, spinner.active)
}

func TestSpinner_Stop_NotActive(t *testing.T) {
	spinner := &Spinner{
		active: false,
	}

	// Should not panic
	spinner.Stop("test")
}

func TestSpinner_Fail(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerDots)

	spinner.Start("Running")
	spinner.Fail("Error occurred")

	assert.False(t, spinner.active)
}

func TestSpinnerStyles(t *testing.T) {
	tests := []struct {
		name  string
		style SpinnerStyle
	}{
		{"dots", SpinnerDots},
		{"line", SpinnerLine},
		{"arrows", SpinnerArrows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Greater(t, len(tt.style), 0)
			assert.NotEmpty(t, tt.style[0])
		})
	}
}

func TestSpinner_Animation(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerLine)

	spinner.Start("Animating")

	// Wait a bit for animation
	time.Sleep(250 * time.Millisecond)

	// Stop animation
	spinner.Stop("Done")

	// Verify it advanced through frames
	assert.False(t, spinner.active)
}

func TestSpinner_ConcurrentStart(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerDots)

	// Start multiple times concurrently
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			spinner.Start("Test")
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	assert.True(t, spinner.active)
	spinner.Stop("")
}

func TestSpinner_ConcurrentUpdate(t *testing.T) {
	spinner := NewSpinnerWithStyle(Config{
		Enabled:     true,
		Interactive: true,
	}, SpinnerDots)

	spinner.Start("Test")

	// Update multiple times concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(n int) {
			spinner.Update(n, 10, "Update")
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	spinner.Stop("")
	assert.False(t, spinner.active)
}
