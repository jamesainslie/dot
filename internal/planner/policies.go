package planner

import (
	"github.com/jamesainslie/dot/pkg/dot"
)

// ResolutionPolicy defines how to handle conflicts
type ResolutionPolicy int

const (
	// PolicyFail stops and reports conflict (default, safest)
	PolicyFail ResolutionPolicy = iota
	// PolicyBackup backs up conflicting file before linking
	PolicyBackup
	// PolicyOverwrite replaces conflicting file with link
	PolicyOverwrite
	// PolicySkip skips conflicting operation
	PolicySkip
)

// String returns the string representation of ResolutionPolicy
func (rp ResolutionPolicy) String() string {
	switch rp {
	case PolicyFail:
		return "fail"
	case PolicyBackup:
		return "backup"
	case PolicyOverwrite:
		return "overwrite"
	case PolicySkip:
		return "skip"
	default:
		return "unknown"
	}
}

// ResolutionPolicies configures conflict resolution behavior per conflict type
type ResolutionPolicies struct {
	OnFileExists    ResolutionPolicy
	OnWrongLink     ResolutionPolicy
	OnPermissionErr ResolutionPolicy
	OnCircular      ResolutionPolicy
	OnTypeMismatch  ResolutionPolicy
}

// DefaultPolicies returns safe default policies (all fail)
func DefaultPolicies() ResolutionPolicies {
	return ResolutionPolicies{
		OnFileExists:    PolicyFail,
		OnWrongLink:     PolicyFail,
		OnPermissionErr: PolicyFail,
		OnCircular:      PolicyFail,
		OnTypeMismatch:  PolicyFail,
	}
}

// applyFailPolicy returns unresolved conflict
func applyFailPolicy(c Conflict) ResolutionOutcome {
	return ResolutionOutcome{
		Status:   ResolveConflict,
		Conflict: &c,
	}
}

// applySkipPolicy skips operation with warning
func applySkipPolicy(op dot.LinkCreate, c Conflict) ResolutionOutcome {
	warning := Warning{
		Message:  "Skipping due to conflict: " + op.Target.String(),
		Severity: WarnInfo,
	}

	return ResolutionOutcome{
		Status:  ResolveSkip,
		Warning: &warning,
	}
}
