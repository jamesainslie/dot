package dot

// DiagnosticReport contains health check results.
type DiagnosticReport struct {
	OverallHealth HealthStatus    `json:"overall_health" yaml:"overall_health"`
	Issues        []Issue         `json:"issues" yaml:"issues"`
	Statistics    DiagnosticStats `json:"statistics" yaml:"statistics"`
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

// MarshalJSON marshals HealthStatus as a string.
func (h HealthStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + h.String() + `"`), nil
}

// MarshalYAML marshals HealthStatus as a string.
func (h HealthStatus) MarshalYAML() (interface{}, error) {
	return h.String(), nil
}

// Issue represents a single diagnostic issue.
type Issue struct {
	Severity   IssueSeverity `json:"severity" yaml:"severity"`
	Type       IssueType     `json:"type" yaml:"type"`
	Path       string        `json:"path,omitempty" yaml:"path,omitempty"`
	Message    string        `json:"message" yaml:"message"`
	Suggestion string        `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
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

// MarshalJSON marshals IssueSeverity as a string.
func (s IssueSeverity) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// MarshalYAML marshals IssueSeverity as a string.
func (s IssueSeverity) MarshalYAML() (interface{}, error) {
	return s.String(), nil
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

// MarshalJSON marshals IssueType as a string.
func (t IssueType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

// MarshalYAML marshals IssueType as a string.
func (t IssueType) MarshalYAML() (interface{}, error) {
	return t.String(), nil
}

// DiagnosticStats contains summary statistics.
type DiagnosticStats struct {
	TotalLinks    int `json:"total_links" yaml:"total_links"`
	BrokenLinks   int `json:"broken_links" yaml:"broken_links"`
	OrphanedLinks int `json:"orphaned_links" yaml:"orphaned_links"`
	ManagedLinks  int `json:"managed_links" yaml:"managed_links"`
}
