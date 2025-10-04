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

