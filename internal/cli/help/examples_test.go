package help

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManageExamples(t *testing.T) {
	assert.Greater(t, len(ManageExamples), 0)
	for _, ex := range ManageExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot manage")
	}
}

func TestUnmanageExamples(t *testing.T) {
	assert.Greater(t, len(UnmanageExamples), 0)
	for _, ex := range UnmanageExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot unmanage")
	}
}

func TestRemanageExamples(t *testing.T) {
	assert.Greater(t, len(RemanageExamples), 0)
	for _, ex := range RemanageExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot remanage")
	}
}

func TestAdoptExamples(t *testing.T) {
	assert.Greater(t, len(AdoptExamples), 0)
	for _, ex := range AdoptExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot adopt")
	}
}

func TestStatusExamples(t *testing.T) {
	assert.Greater(t, len(StatusExamples), 0)
	for _, ex := range StatusExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot status")
	}
}

func TestDoctorExamples(t *testing.T) {
	assert.Greater(t, len(DoctorExamples), 0)
	for _, ex := range DoctorExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot doctor")
	}
}

func TestListExamples(t *testing.T) {
	assert.Greater(t, len(ListExamples), 0)
	for _, ex := range ListExamples {
		assert.NotEmpty(t, ex.Description)
		assert.NotEmpty(t, ex.Command)
		assert.Contains(t, ex.Command, "dot list")
	}
}

func TestFormatExamples_Empty(t *testing.T) {
	result := FormatExamples([]Example{})
	assert.Equal(t, "", result)
}

func TestFormatExamples_Single(t *testing.T) {
	examples := []Example{
		{
			Description: "Test example",
			Command:     "dot test",
		},
	}

	result := FormatExamples(examples)
	assert.Contains(t, result, "Examples:")
	assert.Contains(t, result, "Test example")
	assert.Contains(t, result, "$ dot test")
}

func TestFormatExamples_Multiple(t *testing.T) {
	examples := []Example{
		{
			Description: "First example",
			Command:     "dot first",
		},
		{
			Description: "Second example",
			Command:     "dot second",
		},
	}

	result := FormatExamples(examples)
	assert.Contains(t, result, "First example")
	assert.Contains(t, result, "Second example")
	assert.Contains(t, result, "$ dot first")
	assert.Contains(t, result, "$ dot second")
}

func TestFormatExamples_WithOutput(t *testing.T) {
	examples := []Example{
		{
			Description: "Test with output",
			Command:     "dot test",
			Output:      "Success",
		},
	}

	result := FormatExamples(examples)
	assert.Contains(t, result, "Test with output")
	assert.Contains(t, result, "$ dot test")
	assert.Contains(t, result, "Success")
}

func TestExample_AllFields(t *testing.T) {
	ex := Example{
		Description: "Test description",
		Command:     "test command",
		Output:      "test output",
	}

	assert.Equal(t, "Test description", ex.Description)
	assert.Equal(t, "test command", ex.Command)
	assert.Equal(t, "test output", ex.Output)
}

func TestAllCommandsHaveExamples(t *testing.T) {
	commands := map[string][]Example{
		"manage":   ManageExamples,
		"unmanage": UnmanageExamples,
		"remanage": RemanageExamples,
		"adopt":    AdoptExamples,
		"status":   StatusExamples,
		"doctor":   DoctorExamples,
		"list":     ListExamples,
	}

	for name, examples := range commands {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, examples, "command %s should have examples", name)
		})
	}
}

func TestExampleCommands_NoEmojis(t *testing.T) {
	allExamples := [][]Example{
		ManageExamples,
		UnmanageExamples,
		RemanageExamples,
		AdoptExamples,
		StatusExamples,
		DoctorExamples,
		ListExamples,
	}

	for _, examples := range allExamples {
		for _, ex := range examples {
			// Check for common emoji ranges
			for _, r := range ex.Description {
				assert.False(t, r >= 0x1F300 && r <= 0x1F9FF, "description should not contain emojis")
			}
			for _, r := range ex.Command {
				assert.False(t, r >= 0x1F300 && r <= 0x1F9FF, "command should not contain emojis")
			}
		}
	}
}

func TestFormatExamples_Structure(t *testing.T) {
	examples := []Example{
		{Description: "Test", Command: "dot test"},
	}

	result := FormatExamples(examples)

	// Verify structure
	lines := strings.Split(result, "\n")
	assert.True(t, strings.Contains(lines[0], "Examples:"))

	// Find comment line
	foundComment := false
	for _, line := range lines {
		if strings.Contains(line, "#") {
			foundComment = true
			break
		}
	}
	assert.True(t, foundComment, "should have comment line with #")

	// Find command line
	foundCommand := false
	for _, line := range lines {
		if strings.Contains(line, "$") {
			foundCommand = true
			break
		}
	}
	assert.True(t, foundCommand, "should have command line with $")
}
