package progress

import "fmt"

// Tracker manages progress for multi-stage operations.
type Tracker struct {
	stages    []Stage
	current   int
	indicator Indicator
}

// Stage represents one stage of a multi-stage operation.
type Stage struct {
	Name    string
	Total   int
	Current int
}

// NewTracker creates a new progress tracker.
func NewTracker(stages []Stage, indicator Indicator) *Tracker {
	return &Tracker{
		stages:    stages,
		current:   0,
		indicator: indicator,
	}
}

// Start begins tracking.
func (t *Tracker) Start() {
	if t.current < len(t.stages) {
		stage := t.stages[t.current]
		msg := fmt.Sprintf("[%d/%d] %s", t.current+1, len(t.stages), stage.Name)
		t.indicator.Start(msg)
	}
}

// Advance moves to next stage.
func (t *Tracker) Advance() {
	if t.current < len(t.stages) {
		t.current++
		if t.current < len(t.stages) {
			stage := t.stages[t.current]
			msg := fmt.Sprintf("[%d/%d] %s", t.current+1, len(t.stages), stage.Name)
			t.indicator.Update(0, stage.Total, msg)
		}
	}
}

// UpdateCurrent updates current stage progress.
func (t *Tracker) UpdateCurrent(current int, message string) {
	if t.current < len(t.stages) {
		t.stages[t.current].Current = current

		msg := message
		if msg == "" {
			stage := t.stages[t.current]
			msg = fmt.Sprintf("[%d/%d] %s", t.current+1, len(t.stages), stage.Name)
		}

		t.indicator.Update(current, t.stages[t.current].Total, msg)
	}
}

// Complete completes tracking.
func (t *Tracker) Complete(message string) {
	t.indicator.Stop(message)
}

// Fail fails tracking.
func (t *Tracker) Fail(message string) {
	t.indicator.Fail(message)
}

// CurrentStage returns the current stage index.
func (t *Tracker) CurrentStage() int {
	return t.current
}

// TotalStages returns the total number of stages.
func (t *Tracker) TotalStages() int {
	return len(t.stages)
}
