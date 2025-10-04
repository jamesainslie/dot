package pipeline

import (
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// convertConflicts converts planner.Conflict to dot.ConflictInfo for plan metadata.
func convertConflicts(conflicts []planner.Conflict) []dot.ConflictInfo {
	if len(conflicts) == 0 {
		return nil
	}

	infos := make([]dot.ConflictInfo, 0, len(conflicts))
	for _, c := range conflicts {
		infos = append(infos, dot.ConflictInfo{
			Type:    c.Type.String(),
			Path:    c.Path.String(),
			Details: c.Details,
			Context: c.Context,
		})
	}
	return infos
}

// convertWarnings converts planner.Warning to dot.WarningInfo for plan metadata.
func convertWarnings(warnings []planner.Warning) []dot.WarningInfo {
	if len(warnings) == 0 {
		return nil
	}

	infos := make([]dot.WarningInfo, 0, len(warnings))
	for _, w := range warnings {
		infos = append(infos, dot.WarningInfo{
			Message:  w.Message,
			Severity: w.Severity.String(),
			Context:  w.Context,
		})
	}
	return infos
}
