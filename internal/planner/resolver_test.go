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

