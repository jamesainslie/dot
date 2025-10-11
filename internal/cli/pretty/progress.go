package pretty

import (
	"io"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

// ProgressTracker manages multiple progress indicators using go-pretty.
type ProgressTracker struct {
	writer   progress.Writer
	trackers map[string]*progress.Tracker
	enabled  bool
}

// ProgressConfig holds configuration for progress tracking.
type ProgressConfig struct {
	// Enabled controls whether progress is shown
	Enabled bool
	// Output is where progress is written (usually os.Stderr)
	Output io.Writer
	// UpdateFrequency controls how often progress updates
	UpdateFrequency time.Duration
}

// DefaultProgressConfig returns sensible defaults for progress tracking.
func DefaultProgressConfig() ProgressConfig {
	return ProgressConfig{
		Enabled:         IsInteractive(),
		Output:          os.Stderr,
		UpdateFrequency: 100 * time.Millisecond,
	}
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker(config ProgressConfig) *ProgressTracker {
	if !config.Enabled {
		return &ProgressTracker{
			enabled:  false,
			trackers: make(map[string]*progress.Tracker),
		}
	}

	pw := progress.NewWriter()
	pw.SetOutputWriter(config.Output)
	pw.SetAutoStop(false)
	pw.SetTrackerLength(25)
	pw.SetMessageLength(40)
	pw.SetNumTrackersExpected(0)
	pw.SetSortBy(progress.SortByNone)
	pw.SetStyle(getProgressStyle())
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(config.UpdateFrequency)
	pw.Style().Visibility.TrackerOverall = false
	pw.Style().Visibility.Time = true
	pw.Style().Visibility.Value = true

	return &ProgressTracker{
		writer:   pw,
		trackers: make(map[string]*progress.Tracker),
		enabled:  true,
	}
}

// Track adds a new progress tracker with the given ID and message.
func (pt *ProgressTracker) Track(id, message string, total int64) {
	if !pt.enabled {
		return
	}

	tracker := &progress.Tracker{
		Message: message,
		Total:   total,
		Units:   progress.UnitsDefault,
	}

	pt.trackers[id] = tracker
	pt.writer.AppendTracker(tracker)
}

// Update updates the progress for a specific tracker.
func (pt *ProgressTracker) Update(id string, current int64) {
	if !pt.enabled {
		return
	}

	if tracker, ok := pt.trackers[id]; ok {
		tracker.SetValue(current)
	}
}

// UpdateMessage updates the message for a specific tracker.
func (pt *ProgressTracker) UpdateMessage(id, message string) {
	if !pt.enabled {
		return
	}

	if tracker, ok := pt.trackers[id]; ok {
		tracker.UpdateMessage(message)
	}
}

// Increment increments the progress for a specific tracker.
func (pt *ProgressTracker) Increment(id string) {
	if !pt.enabled {
		return
	}

	if tracker, ok := pt.trackers[id]; ok {
		tracker.Increment(1)
	}
}

// MarkDone marks a tracker as complete.
func (pt *ProgressTracker) MarkDone(id string) {
	if !pt.enabled {
		return
	}

	if tracker, ok := pt.trackers[id]; ok {
		tracker.MarkAsDone()
	}
}

// MarkError marks a tracker as failed.
func (pt *ProgressTracker) MarkError(id string) {
	if !pt.enabled {
		return
	}

	if tracker, ok := pt.trackers[id]; ok {
		tracker.MarkAsErrored()
	}
}

// Start starts rendering progress.
func (pt *ProgressTracker) Start() {
	if !pt.enabled {
		return
	}
	go pt.writer.Render()
}

// Stop stops rendering progress.
func (pt *ProgressTracker) Stop() {
	if !pt.enabled {
		return
	}

	// Mark any unfinished trackers as done
	for _, tracker := range pt.trackers {
		if !tracker.IsDone() {
			tracker.MarkAsDone()
		}
	}

	// Stop the writer (this will wait for render to finish)
	pt.writer.Stop()

	// Give a small grace period for goroutines to cleanup
	time.Sleep(50 * time.Millisecond)
}

// IsActive returns whether the tracker is currently active.
func (pt *ProgressTracker) IsActive() bool {
	return pt.enabled && pt.writer.IsRenderInProgress()
}

// getProgressStyle returns a subtle, professional progress style.
func getProgressStyle() progress.Style {
	style := progress.StyleDefault

	// Use subtle colors
	colorEnabled := ShouldUseColor()
	if colorEnabled {
		// Muted colors for professional look
		style.Colors = progress.StyleColorsExample
		style.Colors.Message = style.Colors.Message // Keep default
		style.Colors.Tracker = style.Colors.Tracker // Keep default
	} else {
		// Disable colors
		style.Colors = progress.StyleColorsDefault
	}

	// Customize characters for cleaner look
	style.Chars.BoxLeft = "["
	style.Chars.BoxRight = "]"
	style.Chars.Finished = "="
	style.Chars.Unfinished = " "

	// Options for better UX
	style.Options.DoneString = "Done"
	style.Options.ErrorString = "Failed"
	style.Options.PercentFormat = "%5.1f%%"
	style.Options.Separator = " "
	style.Options.SpeedPosition = progress.PositionRight

	// Visibility settings
	style.Visibility.ETA = true
	style.Visibility.Percentage = true
	style.Visibility.Speed = false
	style.Visibility.SpeedOverall = false
	style.Visibility.Time = false
	style.Visibility.TrackerOverall = false
	style.Visibility.Value = true

	return style
}
