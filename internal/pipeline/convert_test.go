package pipeline

import (
	"testing"

	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertConflicts(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		conflicts := []planner.Conflict{}
		result := convertConflicts(conflicts)
		assert.Nil(t, result)
	})

	t.Run("nil slice", func(t *testing.T) {
		result := convertConflicts(nil)
		assert.Nil(t, result)
	})

	t.Run("single conflict", func(t *testing.T) {
		path := dot.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := planner.NewConflict(
			planner.ConflictFileExists,
			path,
			"File exists at target",
		).WithContext("package", "bash")

		result := convertConflicts([]planner.Conflict{conflict})

		require.Len(t, result, 1)
		assert.Equal(t, "file_exists", result[0].Type)
		assert.Equal(t, "/home/user/.bashrc", result[0].Path)
		assert.Equal(t, "File exists at target", result[0].Details)
		assert.Equal(t, "bash", result[0].Context["package"])
	})

	t.Run("multiple conflicts", func(t *testing.T) {
		path1 := dot.NewFilePath("/home/user/.bashrc").Unwrap()
		path2 := dot.NewFilePath("/home/user/.vimrc").Unwrap()

		conflicts := []planner.Conflict{
			planner.NewConflict(planner.ConflictFileExists, path1, "File 1 exists"),
			planner.NewConflict(planner.ConflictWrongLink, path2, "Wrong link"),
		}

		result := convertConflicts(conflicts)

		require.Len(t, result, 2)
		assert.Equal(t, "file_exists", result[0].Type)
		assert.Equal(t, "wrong_link", result[1].Type)
	})
}

func TestConvertWarnings(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		warnings := []planner.Warning{}
		result := convertWarnings(warnings)
		assert.Nil(t, result)
	})

	t.Run("nil slice", func(t *testing.T) {
		result := convertWarnings(nil)
		assert.Nil(t, result)
	})

	t.Run("single warning", func(t *testing.T) {
		warning := planner.Warning{
			Message:  "Backup created",
			Severity: planner.WarnCaution,
			Context: map[string]string{
				"path": "/home/user/.bashrc",
			},
		}

		result := convertWarnings([]planner.Warning{warning})

		require.Len(t, result, 1)
		assert.Equal(t, "Backup created", result[0].Message)
		assert.Equal(t, "caution", result[0].Severity)
		assert.Equal(t, "/home/user/.bashrc", result[0].Context["path"])
	})

	t.Run("multiple warnings with different severities", func(t *testing.T) {
		warnings := []planner.Warning{
			{Message: "Info message", Severity: planner.WarnInfo},
			{Message: "Caution message", Severity: planner.WarnCaution},
			{Message: "Danger message", Severity: planner.WarnDanger},
		}

		result := convertWarnings(warnings)

		require.Len(t, result, 3)
		assert.Equal(t, "info", result[0].Severity)
		assert.Equal(t, "caution", result[1].Severity)
		assert.Equal(t, "danger", result[2].Severity)
	})
}
