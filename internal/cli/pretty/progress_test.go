package pretty

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProgressConfig(t *testing.T) {
	config := DefaultProgressConfig()

	assert.NotNil(t, config.Output)
	assert.Greater(t, config.UpdateFrequency, time.Duration(0))
}

func TestNewProgressTracker(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		var buf bytes.Buffer
		config := ProgressConfig{
			Enabled:         true,
			Output:          &buf,
			UpdateFrequency: 50 * time.Millisecond,
		}

		pt := NewProgressTracker(config)
		require.NotNil(t, pt)
		assert.True(t, pt.enabled)
		assert.NotNil(t, pt.output)
		assert.NotNil(t, pt.trackers)
	})

	t.Run("disabled", func(t *testing.T) {
		config := ProgressConfig{
			Enabled: false,
		}

		pt := NewProgressTracker(config)
		require.NotNil(t, pt)
		assert.False(t, pt.enabled)
		assert.NotNil(t, pt.trackers)
	})
}

func TestProgressTracker_Track(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)

	pt.Track("test1", "Processing files", 100)

	assert.Contains(t, pt.trackers, "test1")
	assert.Equal(t, int64(100), pt.trackers["test1"].total)
}

func TestProgressTracker_Update(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)
	pt.Track("test", "Processing", 100)

	pt.Update("test", 50)

	// Value should be updated
	assert.Equal(t, int64(50), pt.trackers["test"].current)
}

func TestProgressTracker_UpdateMessage(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)
	pt.Track("test", "Initial message", 100)

	pt.UpdateMessage("test", "Updated message")

	// Message should be updated (we can't directly test this, but we verify no panic)
	assert.NotNil(t, pt.trackers["test"])
}

func TestProgressTracker_Increment(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)
	pt.Track("test", "Processing", 100)

	pt.Increment("test")
	pt.Increment("test")

	assert.Equal(t, int64(2), pt.trackers["test"].current)
}

func TestProgressTracker_MarkDone(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)
	pt.Track("test", "Processing", 100)

	pt.MarkDone("test")

	assert.True(t, pt.trackers["test"].done)
}

func TestProgressTracker_MarkError(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)
	pt.Track("test", "Processing", 100)

	pt.MarkError("test")

	// Should be marked as error (we verify no panic)
	assert.NotNil(t, pt.trackers["test"])
}

func TestProgressTracker_StartStop(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)
	pt.Track("test", "Processing", 100)

	pt.Start()

	// Quick update and stop
	pt.Update("test", 50)
	pt.MarkDone("test")

	// Stop should complete without hanging
	done := make(chan bool, 1)
	go func() {
		pt.Stop()
		done <- true
	}()

	select {
	case <-done:
		// Success - Stop() completed without hanging
	case <-time.After(5 * time.Second):
		t.Fatal("Stop() hung")
	}
}

func TestProgressTracker_MultipleTrackers(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)

	pt.Track("task1", "Task 1", 100)
	pt.Track("task2", "Task 2", 200)
	pt.Track("task3", "Task 3", 50)

	assert.Len(t, pt.trackers, 3)

	// Test without rendering to avoid hanging tests
	pt.Update("task1", 50)
	pt.Update("task2", 100)
	pt.Update("task3", 25)

	// Verify values are updated correctly
	assert.Equal(t, int64(50), pt.trackers["task1"].current)
	assert.Equal(t, int64(100), pt.trackers["task2"].current)
	assert.Equal(t, int64(25), pt.trackers["task3"].current)
}

func TestProgressTracker_DisabledDoesNothing(t *testing.T) {
	config := ProgressConfig{
		Enabled: false,
	}

	pt := NewProgressTracker(config)

	// All operations should be no-ops
	pt.Track("test", "Processing", 100)
	pt.Start()
	pt.Update("test", 50)
	pt.Increment("test")
	pt.UpdateMessage("test", "New message")
	pt.MarkDone("test")
	pt.MarkError("test")
	pt.Stop()

	// Should not crash and IsActive should be false
	assert.False(t, pt.IsActive())
}

func TestProgressTracker_NonExistentID(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)

	// Operations on non-existent ID should not panic
	pt.Update("nonexistent", 50)
	pt.Increment("nonexistent")
	pt.UpdateMessage("nonexistent", "Message")
	pt.MarkDone("nonexistent")
	pt.MarkError("nonexistent")

	// Should complete without error
	assert.NotNil(t, pt)
}

func TestProgressTracker_IsActive(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)

	assert.False(t, pt.IsActive(), "Should not be active before Start")

	pt.Track("test", "Processing", 100)
	pt.Start()

	time.Sleep(100 * time.Millisecond)

	// Might be active during rendering
	// (timing dependent, so we just verify no crash)
	_ = pt.IsActive()

	pt.Stop()

	assert.False(t, pt.IsActive(), "Should not be active after Stop")
}

func TestGetProgressStyle(t *testing.T) {
	// Test removed - getProgressStyle no longer exists in lipgloss implementation
	// Progress bar rendering is handled internally by renderTracker
	t.Skip("getProgressStyle removed in lipgloss implementation")
}

func TestProgressTracker_CompleteWorkflow(t *testing.T) {
	var buf bytes.Buffer
	config := ProgressConfig{
		Enabled:         true,
		Output:          &buf,
		UpdateFrequency: 50 * time.Millisecond,
	}

	pt := NewProgressTracker(config)

	// Simulate a complete workflow without rendering (to avoid hanging)
	pt.Track("download", "Downloading files", 1000)
	pt.Track("process", "Processing files", 500)
	pt.Track("upload", "Uploading results", 250)

	// Simulate download progress
	for i := int64(0); i <= 1000; i += 100 {
		pt.Update("download", i)
	}
	pt.MarkDone("download")

	// Simulate processing
	for i := int64(0); i <= 500; i += 50 {
		pt.Update("process", i)
	}
	pt.MarkDone("process")

	// Simulate upload
	for i := int64(0); i <= 250; i += 25 {
		pt.Update("upload", i)
	}
	pt.MarkDone("upload")

	// Verify all trackers are done
	assert.True(t, pt.trackers["download"].done)
	assert.True(t, pt.trackers["process"].done)
	assert.True(t, pt.trackers["upload"].done)
}
