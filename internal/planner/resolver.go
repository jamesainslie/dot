package planner

import "github.com/jamesainslie/dot/pkg/dot"

// ConflictType categorizes conflicts by their nature
type ConflictType int

const (
	// ConflictFileExists indicates a file exists at the link target location
	ConflictFileExists ConflictType = iota
	// ConflictWrongLink indicates a symlink points to the wrong source
	ConflictWrongLink
	// ConflictPermission indicates permission denied for operation
	ConflictPermission
	// ConflictCircular indicates a circular symlink dependency
	ConflictCircular
	// ConflictDirExpected indicates a directory was expected but file found
	ConflictDirExpected
	// ConflictFileExpected indicates a file was expected but directory found
	ConflictFileExpected
)

// String returns the string representation of ConflictType
func (ct ConflictType) String() string {
	switch ct {
	case ConflictFileExists:
		return "file_exists"
	case ConflictWrongLink:
		return "wrong_link"
	case ConflictPermission:
		return "permission"
	case ConflictCircular:
		return "circular"
	case ConflictDirExpected:
		return "dir_expected"
	case ConflictFileExpected:
		return "file_expected"
	default:
		return "unknown"
	}
}

// Suggestion provides actionable resolution guidance
type Suggestion struct {
	Action      string // What to do
	Explanation string // Why this helps
	Example     string // Example command (optional)
}

// Conflict represents a detected conflict during planning
type Conflict struct {
	Type        ConflictType
	Path        dot.TargetPath
	Details     string
	Context     map[string]string // Additional context
	Suggestions []Suggestion
}

// NewConflict creates a new Conflict with the given type, path, and details
func NewConflict(ct ConflictType, path dot.TargetPath, details string) Conflict {
	return Conflict{
		Type:        ct,
		Path:        path,
		Details:     details,
		Context:     make(map[string]string),
		Suggestions: []Suggestion{},
	}
}

// WithContext adds a context key-value pair to the conflict
func (c Conflict) WithContext(key, value string) Conflict {
	c.Context[key] = value
	return c
}

// WithSuggestion adds a suggestion to the conflict
func (c Conflict) WithSuggestion(s Suggestion) Conflict {
	c.Suggestions = append(c.Suggestions, s)
	return c
}

// WarningSeverity indicates the severity level of a warning
type WarningSeverity int

const (
	// WarnInfo is informational only
	WarnInfo WarningSeverity = iota
	// WarnCaution requires attention
	WarnCaution
	// WarnDanger indicates potentially destructive operation
	WarnDanger
)

// String returns the string representation of WarningSeverity
func (ws WarningSeverity) String() string {
	switch ws {
	case WarnInfo:
		return "info"
	case WarnCaution:
		return "caution"
	case WarnDanger:
		return "danger"
	default:
		return "unknown"
	}
}

// Warning represents a non-fatal issue
type Warning struct {
	Message  string
	Severity WarningSeverity
	Context  map[string]string
}

// ResolutionStatus indicates the outcome of conflict resolution
type ResolutionStatus int

const (
	// ResolveOK indicates no conflict, proceed with operation
	ResolveOK ResolutionStatus = iota
	// ResolveConflict indicates unresolved conflict, operation fails
	ResolveConflict
	// ResolveWarning indicates resolved with warning
	ResolveWarning
	// ResolveSkip indicates operation was skipped
	ResolveSkip
)

// String returns the string representation of ResolutionStatus
func (rs ResolutionStatus) String() string {
	switch rs {
	case ResolveOK:
		return "ok"
	case ResolveConflict:
		return "conflict"
	case ResolveWarning:
		return "warning"
	case ResolveSkip:
		return "skip"
	default:
		return "unknown"
	}
}

// ResolutionOutcome captures the result of resolving a single operation
type ResolutionOutcome struct {
	Status     ResolutionStatus
	Operations []dot.Operation // Modified operations after resolution
	Conflict   *Conflict       // If status is ResolveConflict
	Warning    *Warning        // If status is ResolveWarning
}

