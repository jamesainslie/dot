package render

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColor_Apply(t *testing.T) {
	color := Color{ANSI: "\033[31m"}
	result := color.Apply("test")
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "\033[31m")
	assert.Contains(t, result, colorReset)
}

func TestColor_Apply_Empty(t *testing.T) {
	color := Color{ANSI: ""}
	result := color.Apply("test")
	assert.Equal(t, "test", result)
}

func TestDefaultScheme(t *testing.T) {
	assert.NotEmpty(t, DefaultScheme.Error.ANSI)
	assert.NotEmpty(t, DefaultScheme.Warning.ANSI)
	assert.NotEmpty(t, DefaultScheme.Success.ANSI)
	assert.NotEmpty(t, DefaultScheme.Info.ANSI)
	assert.NotEmpty(t, DefaultScheme.Dim.ANSI)
}

func TestNoColorScheme(t *testing.T) {
	assert.Empty(t, NoColorScheme.Error.ANSI)
	assert.Empty(t, NoColorScheme.Warning.ANSI)
	assert.Empty(t, NoColorScheme.Success.ANSI)
	assert.Empty(t, NoColorScheme.Info.ANSI)
	assert.Empty(t, NoColorScheme.Dim.ANSI)
}

func TestShouldUseColor_NoColor(t *testing.T) {
	// Save original
	orig := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", orig)
		}
	}()

	os.Setenv("NO_COLOR", "1")
	assert.False(t, ShouldUseColor())
}

func TestShouldUseColor_DumbTerm(t *testing.T) {
	// Save original
	orig := os.Getenv("TERM")
	noColor := os.Getenv("NO_COLOR")
	defer func() {
		if orig == "" {
			os.Unsetenv("TERM")
		} else {
			os.Setenv("TERM", orig)
		}
		if noColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", noColor)
		}
	}()

	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "dumb")
	assert.False(t, ShouldUseColor())
}

func TestGetScheme(t *testing.T) {
	scheme := GetScheme()
	// Just verify it returns a valid scheme
	assert.NotNil(t, scheme)
}

func TestColorScheme_AllColors(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		text  string
	}{
		{"error", DefaultScheme.Error, "error message"},
		{"warning", DefaultScheme.Warning, "warning message"},
		{"success", DefaultScheme.Success, "success message"},
		{"info", DefaultScheme.Info, "info message"},
		{"dim", DefaultScheme.Dim, "dim message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.color.Apply(tt.text)
			assert.Contains(t, result, tt.text)
		})
	}
}

func TestColor_Apply_Consistency(t *testing.T) {
	color := Color{ANSI: "\033[32m"}
	text := "consistent text"

	// Applying color multiple times should produce same result
	result1 := color.Apply(text)
	result2 := color.Apply(text)

	assert.Equal(t, result1, result2)
}

func TestNoColorScheme_NoANSI(t *testing.T) {
	text := "plain text"

	tests := []struct {
		name  string
		color Color
	}{
		{"error", NoColorScheme.Error},
		{"warning", NoColorScheme.Warning},
		{"success", NoColorScheme.Success},
		{"info", NoColorScheme.Info},
		{"dim", NoColorScheme.Dim},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.color.Apply(text)
			assert.Equal(t, text, result)
			assert.NotContains(t, result, "\033")
		})
	}
}
