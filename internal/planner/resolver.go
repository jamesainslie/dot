package planner

import (
	"fmt"

	"github.com/jamesainslie/dot/pkg/dot"
)

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
	Path        dot.FilePath
	Details     string
	Context     map[string]string // Additional context
	Suggestions []Suggestion
}

// NewConflict creates a new Conflict with the given type, path, and details
func NewConflict(ct ConflictType, path dot.FilePath, details string) Conflict {
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

// ResolveResult contains all resolved operations, conflicts, and warnings
type ResolveResult struct {
	Operations []dot.Operation
	Conflicts  []Conflict
	Warnings   []Warning
}

// NewResolveResult creates a new ResolveResult with the given operations
func NewResolveResult(ops []dot.Operation) ResolveResult {
	if ops == nil {
		ops = []dot.Operation{}
	}
	return ResolveResult{
		Operations: ops,
		Conflicts:  []Conflict{},
		Warnings:   []Warning{},
	}
}

// WithConflict adds a conflict to the result
func (r ResolveResult) WithConflict(c Conflict) ResolveResult {
	r.Conflicts = append(r.Conflicts, c)
	return r
}

// WithWarning adds a warning to the result
func (r ResolveResult) WithWarning(w Warning) ResolveResult {
	r.Warnings = append(r.Warnings, w)
	return r
}

// HasConflicts returns true if there are any conflicts
func (r ResolveResult) HasConflicts() bool {
	return len(r.Conflicts) > 0
}

// ConflictCount returns the number of conflicts
func (r ResolveResult) ConflictCount() int {
	return len(r.Conflicts)
}

// WarningCount returns the number of warnings
func (r ResolveResult) WarningCount() int {
	return len(r.Warnings)
}

// FileInfo represents basic file information
type FileInfo struct {
	Size int64
	Mode uint32
}

// LinkTarget represents a symlink target
type LinkTarget struct {
	Target string
}

// CurrentState represents the current filesystem state
type CurrentState struct {
	Files map[string]FileInfo   // Regular files at target paths
	Links map[string]LinkTarget // Existing symlinks
	Dirs  map[string]bool       // Existing directories
}

// detectLinkCreateConflicts checks for conflicts when creating a symlink
func detectLinkCreateConflicts(op dot.LinkCreate, current CurrentState) ResolutionOutcome {
	targetKey := op.Target.String()

	// Check if symlink already exists and points to the correct location
	if link, exists := current.Links[targetKey]; exists {
		if link.Target == op.Source.String() {
			// Link already correct, skip
			return ResolutionOutcome{
				Status: ResolveSkip,
			}
		}
		// Symlink exists but points elsewhere
		conflict := NewConflict(
			ConflictWrongLink,
			op.Target,
			fmt.Sprintf("Symlink points to %s, expected %s", link.Target, op.Source.String()),
		)
		return ResolutionOutcome{
			Status:   ResolveConflict,
			Conflict: &conflict,
		}
	}

	// Check if regular file exists at target
	if fileInfo, exists := current.Files[targetKey]; exists {
		conflict := NewConflict(
			ConflictFileExists,
			op.Target,
			fmt.Sprintf("File exists at target (size=%d)", fileInfo.Size),
		)
		return ResolutionOutcome{
			Status:   ResolveConflict,
			Conflict: &conflict,
		}
	}

	// No conflict
	return ResolutionOutcome{
		Status:     ResolveOK,
		Operations: []dot.Operation{op},
	}
}

// detectDirCreateConflicts checks for conflicts when creating a directory
func detectDirCreateConflicts(op dot.DirCreate, current CurrentState) ResolutionOutcome {
	pathKey := op.Path.String()

	// Check if directory already exists
	if current.Dirs[pathKey] {
		// Directory already exists, skip
		return ResolutionOutcome{
			Status: ResolveSkip,
		}
	}

	// Check if file exists where directory is expected
	if _, exists := current.Files[pathKey]; exists {
		conflict := NewConflict(
			ConflictFileExpected,
			op.Path,
			"File exists where directory expected",
		)
		return ResolutionOutcome{
			Status:   ResolveConflict,
			Conflict: &conflict,
		}
	}

	// No conflict
	return ResolutionOutcome{
		Status:     ResolveOK,
		Operations: []dot.Operation{op},
	}
}
