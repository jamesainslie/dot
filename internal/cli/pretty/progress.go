package pretty

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ProgressTracker manages multiple progress indicators using lipgloss.
type ProgressTracker struct {
	output          io.Writer
	trackers        map[string]*tracker
	enabled         bool
	updateFrequency time.Duration
	ticker          *time.Ticker
	done            chan bool
	mu              sync.RWMutex
	isRendering     bool
}

// tracker represents a single progress tracker.
type tracker struct {
	message string
	current int64
	total   int64
	done    bool
	errored bool
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
	return &ProgressTracker{
		output:          config.Output,
		trackers:        make(map[string]*tracker),
		enabled:         config.Enabled,
		updateFrequency: config.UpdateFrequency,
		done:            make(chan bool),
		isRendering:     false,
	}
}

// Track adds a new progress tracker with the given ID and message.
func (pt *ProgressTracker) Track(id, message string, total int64) {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.trackers[id] = &tracker{
		message: message,
		total:   total,
		current: 0,
		done:    false,
		errored: false,
	}
}

// Update updates the progress for a specific tracker.
func (pt *ProgressTracker) Update(id string, current int64) {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	if t, ok := pt.trackers[id]; ok {
		t.current = current
	}
}

// UpdateMessage updates the message for a specific tracker.
func (pt *ProgressTracker) UpdateMessage(id, message string) {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	if t, ok := pt.trackers[id]; ok {
		t.message = message
	}
}

// Increment increments the progress for a specific tracker.
func (pt *ProgressTracker) Increment(id string) {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	if t, ok := pt.trackers[id]; ok {
		t.current++
	}
}

// MarkDone marks a tracker as complete.
func (pt *ProgressTracker) MarkDone(id string) {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	if t, ok := pt.trackers[id]; ok {
		t.done = true
		t.current = t.total
	}
}

// MarkError marks a tracker as failed.
func (pt *ProgressTracker) MarkError(id string) {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	if t, ok := pt.trackers[id]; ok {
		t.errored = true
	}
}

// Start starts rendering progress.
func (pt *ProgressTracker) Start() {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	if pt.isRendering {
		pt.mu.Unlock()
		return
	}
	pt.isRendering = true

	// Use configured update frequency, fallback to 100ms if zero or negative
	freq := pt.updateFrequency
	if freq <= 0 {
		freq = 100 * time.Millisecond
	}
	pt.ticker = time.NewTicker(freq)
	pt.mu.Unlock()

	go pt.render()
}

// Stop stops rendering progress.
func (pt *ProgressTracker) Stop() {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	if !pt.isRendering {
		pt.mu.Unlock()
		return
	}
	pt.mu.Unlock()

	// Signal done
	pt.done <- true

	// Wait for ticker to stop
	time.Sleep(50 * time.Millisecond)

	// Final render
	pt.renderOnce()
}

// IsActive returns whether the tracker is currently active.
func (pt *ProgressTracker) IsActive() bool {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.enabled && pt.isRendering
}

// render continuously renders progress updates.
func (pt *ProgressTracker) render() {
	for {
		select {
		case <-pt.done:
			pt.mu.Lock()
			pt.ticker.Stop()
			pt.isRendering = false
			pt.mu.Unlock()
			return
		case <-pt.ticker.C:
			pt.renderOnce()
		}
	}
}

// renderOnce renders the current state once.
func (pt *ProgressTracker) renderOnce() {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if len(pt.trackers) == 0 {
		return
	}

	// Clear previous lines (simple approach)
	fmt.Fprint(pt.output, "\r")

	for _, t := range pt.trackers {
		line := pt.renderTracker(t)
		fmt.Fprintln(pt.output, line)
	}

	// Move cursor up to overwrite next time
	if len(pt.trackers) > 0 {
		fmt.Fprintf(pt.output, "\033[%dA", len(pt.trackers))
	}
}

// renderTracker renders a single tracker.
func (pt *ProgressTracker) renderTracker(t *tracker) string {
	var result strings.Builder

	// Message
	msgStyle := lipgloss.NewStyle()
	if ShouldUseColor() {
		msgStyle = msgStyle.Foreground(lipgloss.Color("252"))
	}
	result.WriteString(msgStyle.Width(40).Render(Truncate(t.message, 40)))
	result.WriteString(" ")

	// Progress bar
	barWidth := 25
	var pct float64
	if t.total > 0 {
		pct = float64(t.current) / float64(t.total)
	}
	filled := int(pct * float64(barWidth))

	barStyle := lipgloss.NewStyle()
	if ShouldUseColor() {
		if t.errored {
			barStyle = barStyle.Foreground(lipgloss.Color("167"))
		} else if t.done {
			barStyle = barStyle.Foreground(lipgloss.Color("71"))
		} else {
			barStyle = barStyle.Foreground(lipgloss.Color("110"))
		}
	}

	result.WriteString("[")
	result.WriteString(barStyle.Render(strings.Repeat("=", filled)))
	result.WriteString(strings.Repeat(" ", barWidth-filled))
	result.WriteString("]")

	// Percentage
	result.WriteString(fmt.Sprintf(" %5.1f%% ", pct*100))

	// Status
	if t.errored {
		result.WriteString(Error("Failed"))
	} else if t.done {
		result.WriteString(Success("Done"))
	} else {
		result.WriteString(fmt.Sprintf("%d/%d", t.current, t.total))
	}

	return result.String()
}
