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
	// HealthOK indicates no issues found.
	HealthOK HealthStatus = iota
	// HealthWarnings indicates non-critical issues found.
	HealthWarnings
	// HealthErrors indicates critical issues found.
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
	// SeverityInfo represents informational messages.
	SeverityInfo IssueSeverity = iota
	// SeverityWarning represents non-critical issues.
	SeverityWarning
	// SeverityError represents critical issues requiring attention.
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
	// IssueBrokenLink indicates a symlink pointing to a non-existent target.
	IssueBrokenLink IssueType = iota
	// IssueOrphanedLink indicates a symlink not managed by any package.
	IssueOrphanedLink
	// IssueWrongTarget indicates a symlink pointing to an unexpected target.
	IssueWrongTarget
	// IssuePermission indicates insufficient permissions for an operation.
	IssuePermission
	// IssueCircular indicates a circular symlink reference.
	IssueCircular
	// IssueManifestInconsistency indicates mismatch between manifest and filesystem.
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

// ScanMode controls orphaned link detection behavior.
type ScanMode int

const (
	// ScanOff disables orphaned link detection (fastest, use --scan-mode=off to enable).
	ScanOff ScanMode = iota
	// ScanScoped only scans directories containing managed links (default, recommended).
	ScanScoped
	// ScanDeep performs full recursive scan with depth limits (slowest, thorough).
	ScanDeep
)

// String returns the string representation of scan mode.
func (s ScanMode) String() string {
	switch s {
	case ScanOff:
		return "off"
	case ScanScoped:
		return "scoped"
	case ScanDeep:
		return "deep"
	default:
		return "unknown"
	}
}

// ScanConfig controls orphaned link detection behavior.
type ScanConfig struct {
	// Mode determines the scanning strategy.
	Mode ScanMode

	// MaxDepth limits directory recursion depth.
	// Values <= 0 are normalized to 3 (scoped) or 10 (deep) by constructor helpers.
	// Default: 3 (scoped), 10 (deep)
	MaxDepth int

	// ScopeToDirs limits scanning to specific directories.
	// Empty means auto-detect from manifest (for ScanScoped) or target dir (for ScanDeep).
	ScopeToDirs []string

	// SkipPatterns are directory names/patterns to skip during scanning.
	// Default constructors include common large directories unlikely to contain dotfile symlinks.
	SkipPatterns []string

	// MaxWorkers limits parallel directory scanning goroutines.
	// Values <= 0 default to runtime.NumCPU().
	// Set to 1 to disable parallel scanning.
	// Default: 0 (use NumCPU)
	MaxWorkers int

	// MaxIssues limits the number of issues to collect before stopping scan.
	// Values <= 0 mean unlimited (scan everything).
	// Useful for fast health checks without full enumeration.
	// Default: 0 (unlimited)
	MaxIssues int
}

// defaultSkipPatterns returns common directories to skip during scanning.
// These are large directories unlikely to contain dotfile symlinks.
func defaultSkipPatterns() []string {
	return []string{
		// Version control
		".git", ".svn", ".hg",
		// Build artifacts and dependencies
		"node_modules", "vendor", "__pycache__", ".terraform",
		// Package manager caches
		".cache", ".npm", ".cargo", ".rustup", ".pyenv", ".rbenv",
		".gradle", ".m2", "go/pkg",
		// Application state and data
		".docker", ".rd", ".local/share", ".kube/cache",
		// Editor caches
		".vscode", ".config/Code",
		// System directories (macOS)
		"Library", ".Trash",
	}
}

// DefaultScanConfig returns the default scan configuration (scoped).
// Scoped scanning checks directories containing managed links for orphaned symlinks.
func DefaultScanConfig() ScanConfig {
	return ScanConfig{
		Mode:         ScanScoped,
		MaxDepth:     3,
		ScopeToDirs:  nil,
		SkipPatterns: defaultSkipPatterns(),
		MaxWorkers:   0, // Use NumCPU
		MaxIssues:    0, // Unlimited
	}
}

// ScopedScanConfig returns a scan configuration for scoped scanning.
func ScopedScanConfig() ScanConfig {
	return ScanConfig{
		Mode:         ScanScoped,
		MaxDepth:     3,
		ScopeToDirs:  nil,
		SkipPatterns: defaultSkipPatterns(),
		MaxWorkers:   0,
		MaxIssues:    0,
	}
}

// DeepScanConfig returns a scan configuration for deep scanning.
func DeepScanConfig(maxDepth int) ScanConfig {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	return ScanConfig{
		Mode:         ScanDeep,
		MaxDepth:     maxDepth,
		ScopeToDirs:  nil,
		SkipPatterns: defaultSkipPatterns(),
		MaxWorkers:   0,
		MaxIssues:    0,
	}
}
