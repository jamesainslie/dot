package progress

import (
	"fmt"
	"strings"
	"time"
)

// Bar displays a progress bar.
type Bar struct {
	width     int
	current   int
	total     int
	message   string
	startTime time.Time
	started   bool
}

// Start begins progress display.
func (b *Bar) Start(message string) {
	b.message = message
	b.current = 0
	b.total = 0
	b.startTime = time.Now()
	b.started = true
	b.render()
}

// Update updates progress.
func (b *Bar) Update(current, total int, message string) {
	b.current = current
	b.total = total
	if message != "" {
		b.message = message
	}
	if !b.started {
		b.started = true
		b.startTime = time.Now()
	}
	b.render()
}

// Stop completes progress display.
func (b *Bar) Stop(message string) {
	if message != "" {
		b.message = message
	}
	b.current = b.total
	b.render()
	fmt.Println() // Move to next line
	b.started = false
}

// Fail displays failure message.
func (b *Bar) Fail(message string) {
	if message != "" {
		b.message = message
	}
	b.render()
	fmt.Println() // Move to next line
	b.started = false
}

// Render generates progress bar string.
func (b *Bar) Render() string {
	if b.total == 0 {
		return fmt.Sprintf("%s...", b.message)
	}

	percentage := (b.current * 100) / b.total
	barWidth := 20 // Fixed bar width
	if b.width > 0 && b.width < 80 {
		barWidth = b.width / 4
	}
	if barWidth < 10 {
		barWidth = 10
	}

	filled := (barWidth * b.current) / b.total
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Calculate ETA
	etaStr := ""
	if b.started && b.current > 0 && b.current < b.total {
		elapsed := time.Since(b.startTime)
		rate := float64(b.current) / elapsed.Seconds()
		remaining := float64(b.total-b.current) / rate
		etaStr = fmt.Sprintf(" ETA %ds", int(remaining))
	}

	return fmt.Sprintf("%s [%s] %d%% (%d/%d)%s",
		b.message, bar, percentage, b.current, b.total, etaStr)
}

func (b *Bar) render() {
	output := b.Render()
	// Clear line and print progress
	fmt.Printf("\r\033[K%s", output)
}
