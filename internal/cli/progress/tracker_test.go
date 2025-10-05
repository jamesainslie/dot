package progress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTracker(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
		{Name: "Stage 2", Total: 20},
	}
	indicator := &NoOpIndicator{}

	tracker := NewTracker(stages, indicator)

	assert.NotNil(t, tracker)
	assert.Equal(t, 2, len(tracker.stages))
	assert.Equal(t, 0, tracker.current)
}

func TestTracker_Start(t *testing.T) {
	stages := []Stage{
		{Name: "First Stage", Total: 10},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()

	assert.Equal(t, 0, tracker.current)
}

func TestTracker_Advance(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
		{Name: "Stage 2", Total: 20},
		{Name: "Stage 3", Total: 30},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()
	assert.Equal(t, 0, tracker.current)

	tracker.Advance()
	assert.Equal(t, 1, tracker.current)

	tracker.Advance()
	assert.Equal(t, 2, tracker.current)

	tracker.Advance()
	assert.Equal(t, 3, tracker.current)

	// Should not advance beyond stages
	tracker.Advance()
	assert.Equal(t, 3, tracker.current)
}

func TestTracker_UpdateCurrent(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()
	tracker.UpdateCurrent(5, "Progress update")

	assert.Equal(t, 5, tracker.stages[0].Current)
}

func TestTracker_UpdateCurrent_DefaultMessage(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()
	tracker.UpdateCurrent(5, "")

	assert.Equal(t, 5, tracker.stages[0].Current)
}

func TestTracker_Complete(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()
	tracker.Complete("All done")

	// Should not panic
}

func TestTracker_Fail(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()
	tracker.Fail("Operation failed")

	// Should not panic
}

func TestTracker_CurrentStage(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
		{Name: "Stage 2", Total: 20},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	assert.Equal(t, 0, tracker.CurrentStage())

	tracker.Advance()
	assert.Equal(t, 1, tracker.CurrentStage())
}

func TestTracker_TotalStages(t *testing.T) {
	stages := []Stage{
		{Name: "Stage 1", Total: 10},
		{Name: "Stage 2", Total: 20},
		{Name: "Stage 3", Total: 30},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	assert.Equal(t, 3, tracker.TotalStages())
}

func TestTracker_MultiStageProgress(t *testing.T) {
	stages := []Stage{
		{Name: "Scanning", Total: 10},
		{Name: "Planning", Total: 20},
		{Name: "Executing", Total: 30},
	}
	indicator := &NoOpIndicator{}
	tracker := NewTracker(stages, indicator)

	tracker.Start()

	// Stage 1
	tracker.UpdateCurrent(5, "")
	assert.Equal(t, 5, tracker.stages[0].Current)

	tracker.UpdateCurrent(10, "")
	assert.Equal(t, 10, tracker.stages[0].Current)

	// Advance to stage 2
	tracker.Advance()
	assert.Equal(t, 1, tracker.CurrentStage())

	tracker.UpdateCurrent(10, "")
	assert.Equal(t, 10, tracker.stages[1].Current)

	// Advance to stage 3
	tracker.Advance()
	assert.Equal(t, 2, tracker.CurrentStage())

	tracker.UpdateCurrent(30, "")
	assert.Equal(t, 30, tracker.stages[2].Current)

	tracker.Complete("All stages complete")
}

func TestTracker_EmptyStages(t *testing.T) {
	indicator := &NoOpIndicator{}
	tracker := NewTracker([]Stage{}, indicator)

	assert.Equal(t, 0, tracker.TotalStages())
	assert.Equal(t, 0, tracker.CurrentStage())

	// Should not panic
	tracker.Start()
	tracker.UpdateCurrent(5, "test")
	tracker.Advance()
	tracker.Complete("done")
}

func TestStage_AllFields(t *testing.T) {
	stage := Stage{
		Name:    "Test Stage",
		Total:   100,
		Current: 50,
	}

	assert.Equal(t, "Test Stage", stage.Name)
	assert.Equal(t, 100, stage.Total)
	assert.Equal(t, 50, stage.Current)
}
