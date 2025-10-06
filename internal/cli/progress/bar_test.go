package progress

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBar_Render_Zero(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 0,
		total:   0,
		message: "Loading",
	}

	result := bar.Render()
	assert.Contains(t, result, "Loading")
	assert.Contains(t, result, "...")
}

func TestBar_Render_Percentage(t *testing.T) {
	tests := []struct {
		name       string
		current    int
		total      int
		percentage int
	}{
		{"0%", 0, 10, 0},
		{"50%", 5, 10, 50},
		{"100%", 10, 10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := &Bar{
				width:   80,
				current: tt.current,
				total:   tt.total,
				message: "Progress",
			}

			result := bar.Render()
			assert.Contains(t, result, "Progress")
			assert.Contains(t, result, "[")
			assert.Contains(t, result, "]")
			assert.Contains(t, result, "%")
		})
	}
}

func TestBar_Render_Counter(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 5,
		total:   10,
		message: "Installing",
	}

	result := bar.Render()
	assert.Contains(t, result, "(5/10)")
}

func TestBar_Render_FilledBar(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 7,
		total:   10,
		message: "Test",
	}

	result := bar.Render()
	// Should have filled and unfilled characters
	assert.Contains(t, result, "█")
	assert.Contains(t, result, "░")
}

func TestBar_Render_ETA(t *testing.T) {
	bar := &Bar{
		width:     80,
		current:   5,
		total:     10,
		message:   "Test",
		startTime: time.Now().Add(-5 * time.Second),
		started:   true,
	}

	result := bar.Render()
	// Should include ETA
	assert.Contains(t, result, "ETA")
}

func TestBar_Render_NoETA_NotStarted(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 5,
		total:   10,
		message: "Test",
		started: false,
	}

	result := bar.Render()
	assert.NotContains(t, result, "ETA")
}

func TestBar_Render_NoETA_ZeroProgress(t *testing.T) {
	bar := &Bar{
		width:     80,
		current:   0,
		total:     10,
		message:   "Test",
		startTime: time.Now(),
		started:   true,
	}

	result := bar.Render()
	assert.NotContains(t, result, "ETA")
}

func TestBar_Render_NoETA_Complete(t *testing.T) {
	bar := &Bar{
		width:     80,
		current:   10,
		total:     10,
		message:   "Test",
		startTime: time.Now(),
		started:   true,
	}

	result := bar.Render()
	assert.NotContains(t, result, "ETA")
}

func TestBar_Update(t *testing.T) {
	bar := &Bar{
		width: 80,
	}

	bar.Update(3, 10, "Updated")

	assert.Equal(t, 3, bar.current)
	assert.Equal(t, 10, bar.total)
	assert.Equal(t, "Updated", bar.message)
	assert.True(t, bar.started)
}

func TestBar_Update_PreservesMessage(t *testing.T) {
	bar := &Bar{
		width:   80,
		message: "Original",
	}

	bar.Update(3, 10, "")

	assert.Equal(t, "Original", bar.message)
}

func TestBar_Start(t *testing.T) {
	bar := &Bar{
		width: 80,
	}

	bar.Start("Starting")

	assert.Equal(t, "Starting", bar.message)
	assert.Equal(t, 0, bar.current)
	assert.Equal(t, 0, bar.total)
	assert.True(t, bar.started)
	assert.False(t, bar.startTime.IsZero())
}

func TestBar_Stop(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 5,
		total:   10,
		started: true,
	}

	bar.Stop("Complete")

	assert.Equal(t, 10, bar.current)
	assert.Equal(t, "Complete", bar.message)
	assert.False(t, bar.started)
}

func TestBar_Fail(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 5,
		total:   10,
		started: true,
	}

	bar.Fail("Failed")

	assert.Equal(t, "Failed", bar.message)
	assert.False(t, bar.started)
}

func TestBar_Render_NarrowTerminal(t *testing.T) {
	bar := &Bar{
		width:   40,
		current: 5,
		total:   10,
		message: "Narrow",
	}

	result := bar.Render()
	assert.Contains(t, result, "Narrow")
	// Bar should still render with reduced width
	assert.Contains(t, result, "[")
	assert.Contains(t, result, "]")
}

func TestBar_Render_VeryNarrowTerminal(t *testing.T) {
	bar := &Bar{
		width:   20,
		current: 5,
		total:   10,
		message: "Tiny",
	}

	result := bar.Render()
	assert.Contains(t, result, "Tiny")
	// Should use minimum bar width
	barStart := strings.Index(result, "[")
	barEnd := strings.Index(result, "]")
	if barStart >= 0 && barEnd > barStart {
		barContent := result[barStart+1 : barEnd]
		// Bar should have at least minimal width
		assert.GreaterOrEqual(t, len(barContent), 10)
	}
}

func TestBar_Render_OverflowCurrent(t *testing.T) {
	bar := &Bar{
		width:   80,
		current: 15,
		total:   10,
		message: "Overflow",
	}

	result := bar.Render()
	// Should handle overflow gracefully
	assert.Contains(t, result, "Overflow")
	assert.NotContains(t, result, "panic")
}
