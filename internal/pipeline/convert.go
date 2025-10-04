package pipeline

import (
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// convertConflicts converts planner.Conflict to dot.ConflictInfo for plan metadata.
// Creates shallow copies of context maps to prevent shared mutation.
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
			Context: copyContext(c.Context),
		})
	}
	return infos
}

// convertWarnings converts planner.Warning to dot.WarningInfo for plan metadata.
// Creates shallow copies of context maps to prevent shared mutation.
func convertWarnings(warnings []planner.Warning) []dot.WarningInfo {
	if len(warnings) == 0 {
		return nil
	}

	infos := make([]dot.WarningInfo, 0, len(warnings))
	for _, w := range warnings {
		infos = append(infos, dot.WarningInfo{
			Message:  w.Message,
			Severity: w.Severity.String(),
			Context:  copyContext(w.Context),
		})
	}
	return infos
}

// copyContext creates a shallow copy of a context map.
// Returns nil if the input is nil, otherwise returns a new map with copied entries.
// This prevents shared mutation between planner structures and public API metadata.
func copyContext(ctx map[string]string) map[string]string {
	if ctx == nil {
		return nil
	}

	copied := make(map[string]string, len(ctx))
	for k, v := range ctx {
		copied[k] = v
	}
	return copied
}
