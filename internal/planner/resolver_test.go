package planner

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

// Task 7.1.1: Test ConflictType enumeration
func TestConflictTypeString(t *testing.T) {
	tests := []struct {
		name string
		ct   ConflictType
		want string
	}{
		{"file exists", ConflictFileExists, "file_exists"},
		{"wrong link", ConflictWrongLink, "wrong_link"},
		{"permission", ConflictPermission, "permission"},
		{"circular", ConflictCircular, "circular"},
		{"dir expected", ConflictDirExpected, "dir_expected"},
		{"file expected", ConflictFileExpected, "file_expected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Task 7.1.2: Test Conflict value object
func TestConflictCreation(t *testing.T) {
	targetPath := dot.NewTargetPath("/home/user/.bashrc").Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists at target location",
	)

	assert.Equal(t, ConflictFileExists, conflict.Type)
	assert.Equal(t, targetPath, conflict.Path)
	assert.Equal(t, "File exists at target location", conflict.Details)
	assert.NotNil(t, conflict.Context)
	assert.Empty(t, conflict.Suggestions)
}

func TestConflictWithContext(t *testing.T) {
	targetPath := dot.NewTargetPath("/home/user/.bashrc").Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists",
	)

	conflict = conflict.WithContext("size", "1024")
	conflict = conflict.WithContext("mode", "0644")

	assert.Equal(t, "1024", conflict.Context["size"])
	assert.Equal(t, "0644", conflict.Context["mode"])
}

func TestConflictWithSuggestion(t *testing.T) {
	targetPath := dot.NewTargetPath("/home/user/.bashrc").Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists",
	)

	suggestion := Suggestion{
		Action:      "Use --backup flag",
		Explanation: "Preserves existing file",
	}

	conflict = conflict.WithSuggestion(suggestion)

	assert.Len(t, conflict.Suggestions, 1)
	assert.Equal(t, "Use --backup flag", conflict.Suggestions[0].Action)
}

// Task 7.1.3: Test Resolution Status Types
func TestResolutionStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status ResolutionStatus
		want   string
	}{
		{"ok", ResolveOK, "ok"},
		{"conflict", ResolveConflict, "conflict"},
		{"warning", ResolveWarning, "warning"},
		{"skip", ResolveSkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolutionOutcomeCreation(t *testing.T) {
	t.Run("ok status", func(t *testing.T) {
		outcome := ResolutionOutcome{
			Status:     ResolveOK,
			Operations: []dot.Operation{},
		}
		assert.Equal(t, ResolveOK, outcome.Status)
		assert.NotNil(t, outcome.Operations)
		assert.Nil(t, outcome.Conflict)
		assert.Nil(t, outcome.Warning)
	})

	t.Run("conflict status", func(t *testing.T) {
		targetPath := dot.NewTargetPath("/home/user/.bashrc").Unwrap()
		conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

		outcome := ResolutionOutcome{
			Status:   ResolveConflict,
			Conflict: &conflict,
		}
		assert.Equal(t, ResolveConflict, outcome.Status)
		assert.NotNil(t, outcome.Conflict)
		assert.Equal(t, ConflictFileExists, outcome.Conflict.Type)
	})
}

// Task 7.1.4: Test ResolveResult Type
func TestResolveResultConstruction(t *testing.T) {
	t.Run("with operations", func(t *testing.T) {
		ops := []dot.Operation{}
		result := NewResolveResult(ops)
		assert.Len(t, result.Operations, 0)
		assert.Empty(t, result.Conflicts)
		assert.Empty(t, result.Warnings)
	})

	t.Run("with conflicts", func(t *testing.T) {
		targetPath := dot.NewTargetPath("/home/user/.bashrc").Unwrap()
		conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

		result := NewResolveResult(nil)
		result = result.WithConflict(conflict)

		assert.Len(t, result.Conflicts, 1)
		assert.Equal(t, ConflictFileExists, result.Conflicts[0].Type)
	})

	t.Run("with warnings", func(t *testing.T) {
		warning := Warning{
			Message:  "File backed up",
			Severity: WarnInfo,
		}

		result := NewResolveResult(nil)
		result = result.WithWarning(warning)

		assert.Len(t, result.Warnings, 1)
		assert.Equal(t, "File backed up", result.Warnings[0].Message)
	})
}

func TestResolveResultQueries(t *testing.T) {
	t.Run("HasConflicts", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.False(t, result.HasConflicts())

		targetPath := dot.NewTargetPath("/home/user/.bashrc").Unwrap()
		conflict := NewConflict(ConflictFileExists, targetPath, "File exists")
		result = result.WithConflict(conflict)

		assert.True(t, result.HasConflicts())
	})

	t.Run("ConflictCount", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.Equal(t, 0, result.ConflictCount())

		targetPath1 := dot.NewTargetPath("/home/user/.bashrc").Unwrap()
		conflict1 := NewConflict(ConflictFileExists, targetPath1, "File exists")
		result = result.WithConflict(conflict1)

		targetPath2 := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
		conflict2 := NewConflict(ConflictWrongLink, targetPath2, "Wrong link")
		result = result.WithConflict(conflict2)

		assert.Equal(t, 2, result.ConflictCount())
	})

	t.Run("WarningCount", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.Equal(t, 0, result.WarningCount())

		warning := Warning{Message: "Test warning"}
		result = result.WithWarning(warning)

		assert.Equal(t, 1, result.WarningCount())
	})
}

