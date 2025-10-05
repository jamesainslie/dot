package dot

// DiagnosticReport contains health check results.
type DiagnosticReport struct {
	OverallHealth HealthStatus
	Issues        []Issue
	Statistics    DiagnosticStats
}

// HealthStatus represents the overall health of the installation.
type HealthStatus int

const (
	HealthOK HealthStatus = iota
	HealthWarnings
	HealthErrors
)

// String returns the string representation of health status.
func (h HealthStatus) String() string {
	switch h {
	case HealthOK:
		return "healthy"
	case HealthWarnings:
		return "warnings"
	case HealthErrors:
		return "errors"
	default:
		return "unknown"
	}
}

// Issue represents a single diagnostic issue.
type Issue struct {
	Severity   IssueSeverity
	Type       IssueType
	Path       string
	Message    string
	Suggestion string
}

// IssueSeverity indicates the severity of an issue.
type IssueSeverity int

const (
	SeverityInfo IssueSeverity = iota
	SeverityWarning
	SeverityError
)

// String returns the string representation of severity.
func (s IssueSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// IssueType categorizes the type of issue.
type IssueType int

const (
	IssueBrokenLink IssueType = iota
	IssueOrphanedLink
	IssueWrongTarget
	IssuePermission
	IssueCircular
	IssueManifestInconsistency
)

// String returns the string representation of issue type.
func (t IssueType) String() string {
	switch t {
	case IssueBrokenLink:
		return "broken_link"
	case IssueOrphanedLink:
		return "orphaned_link"
	case IssueWrongTarget:
		return "wrong_target"
	case IssuePermission:
		return "permission"
	case IssueCircular:
		return "circular"
	case IssueManifestInconsistency:
		return "manifest_inconsistency"
	default:
		return "unknown"
	}
}

// DiagnosticStats contains summary statistics.
type DiagnosticStats struct {
	TotalLinks    int
	BrokenLinks   int
	OrphanedLinks int
	ManagedLinks  int
}
