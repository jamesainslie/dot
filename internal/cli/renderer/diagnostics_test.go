package renderer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/internal/domain"
)

func TestJSONRenderer_RenderDiagnostics(t *testing.T) {
	r := &JSONRenderer{pretty: true}

	report := domain.DiagnosticReport{
		OverallHealth: domain.HealthOK,
		Issues:        []domain.Issue{},
		Statistics: domain.DiagnosticStats{
			TotalLinks: 10,
		},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"overall_health"`)
	assert.Contains(t, output, `"healthy"`)
	assert.Contains(t, output, `"total_links"`)
}

func TestYAMLRenderer_RenderDiagnostics(t *testing.T) {
	r := &YAMLRenderer{indent: 2}

	report := domain.DiagnosticReport{
		OverallHealth: domain.HealthWarnings,
		Issues: []domain.Issue{
			{
				Severity:   domain.SeverityWarning,
				Type:       domain.IssueOrphanedLink,
				Path:       "/test",
				Message:    "Test",
				Suggestion: "Fix it",
			},
		},
		Statistics: domain.DiagnosticStats{},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "overall_health")
	assert.Contains(t, output, "warnings")
	assert.Contains(t, output, "issues")
}

func TestTableRenderer_RenderDiagnostics(t *testing.T) {
	r := &TableRenderer{
		colorize: false,
		scheme:   ColorScheme{},
		width:    80,
	}

	report := domain.DiagnosticReport{
		OverallHealth: domain.HealthErrors,
		Issues: []domain.Issue{
			{
				Severity: domain.SeverityError,
				Type:     domain.IssueBrokenLink,
				Path:     "/test/path",
				Message:  "Broken",
			},
		},
		Statistics: domain.DiagnosticStats{
			TotalLinks:  5,
			BrokenLinks: 1,
		},
	}

	var buf bytes.Buffer
	err := r.RenderDiagnostics(&buf, report)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "errors")
	assert.Contains(t, output, "Total Links: 5")
	assert.Contains(t, output, "Broken Links: 1")
}
