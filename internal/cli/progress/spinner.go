package progress

import (
	"fmt"
	"sync"
	"time"
)

// SpinnerStyle defines the animation frames for a spinner.
type SpinnerStyle []string

// Predefined spinner styles.
var (
	SpinnerDots   = SpinnerStyle{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine   = SpinnerStyle{"-", "\\", "|", "/"}
	SpinnerArrows = SpinnerStyle{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
)

// Spinner displays an indeterminate progress spinner.
type Spinner struct {
	frames  SpinnerStyle
	current int
	message string
	ticker  *time.Ticker
	done    chan struct{}
	mu      sync.Mutex
	active  bool
}

// NewSpinnerWithStyle creates a spinner with the specified style.
func NewSpinnerWithStyle(cfg Config, style SpinnerStyle) *Spinner {
	if len(style) == 0 {
		style = SpinnerDots
	}
	return &Spinner{
		frames: style,
		done:   make(chan struct{}),
	}
}

// Start begins spinner animation.
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return
	}

	s.message = message
	s.current = 0
	s.active = true
	s.ticker = time.NewTicker(100 * time.Millisecond)

	go s.animate()
	s.render()
}

// Update updates the message.
func (s *Spinner) Update(current, total int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if message != "" {
		s.message = message
	}
}

// Stop completes spinner display.
func (s *Spinner) Stop(message string) {
	s.mu.Lock()

	if !s.active {
		s.mu.Unlock()
		return
	}

	s.active = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.mu.Unlock()

	// Signal animation goroutine to stop
	select {
	case s.done <- struct{}{}:
	default:
	}

	// Clear spinner and print final message
	if message != "" {
		fmt.Printf("\r\033[K%s\n", message)
	} else {
		fmt.Printf("\r\033[K%s\n", s.message)
	}
}

// Fail displays failure message.
func (s *Spinner) Fail(message string) {
	s.Stop(message)
}

// Spin advances to next frame.
func (s *Spinner) Spin() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.current = (s.current + 1) % len(s.frames)
	s.render()
}

func (s *Spinner) animate() {
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			s.Spin()
		}
	}
}

func (s *Spinner) render() {
	if !s.active {
		return
	}

	frame := s.frames[s.current]
	output := fmt.Sprintf("%s %s", frame, s.message)
	fmt.Printf("\r\033[K%s", output)
}
