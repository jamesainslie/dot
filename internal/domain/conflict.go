package domain

// ConflictInfo represents conflict information in plan metadata.
// This is a simplified view of conflicts for plan consumers.
type ConflictInfo struct {
	Type    string            `json:"type"`
	Path    string            `json:"path"`
	Details string            `json:"details"`
	Context map[string]string `json:"context,omitempty"`
}

// WarningInfo represents warning information in plan metadata.
// This is a simplified view of warnings for plan consumers.
type WarningInfo struct {
	Message  string            `json:"message"`
	Severity string            `json:"severity"`
	Context  map[string]string `json:"context,omitempty"`
}

