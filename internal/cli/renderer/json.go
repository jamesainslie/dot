package renderer

import (
	"encoding/json"
	"io"

	"github.com/jamesainslie/dot/pkg/dot"
)

// JSONRenderer renders output as JSON.
type JSONRenderer struct {
	pretty bool
}

// RenderStatus renders installation status as JSON.
func (r *JSONRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	encoder := json.NewEncoder(w)
	if r.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(status)
}

// RenderDiagnostics renders diagnostic report as JSON.
func (r *JSONRenderer) RenderDiagnostics(w io.Writer, report dot.DiagnosticReport) error {
	encoder := json.NewEncoder(w)
	if r.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(report)
}
