package help

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	require.NotNil(t, gen)
	assert.Greater(t, gen.width, 0)
}

func TestGenerator_GenerateUsage(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use:   "test [flags] ARGS",
		Short: "Test command",
	}

	usage := gen.GenerateUsage(cmd)
	assert.Contains(t, usage, "test")
}

func TestGenerator_GenerateExamples(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use:     "test",
		Example: "  $ test example",
	}

	examples := gen.GenerateExamples(cmd)
	assert.Contains(t, examples, "test example")
}

func TestGenerator_GenerateExamples_Empty(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use: "test",
	}

	examples := gen.GenerateExamples(cmd)
	assert.Empty(t, examples)
}

func TestGenerator_GenerateSeeAlso_WithParent(t *testing.T) {
	gen := NewGenerator()
	parent := &cobra.Command{
		Use: "parent",
	}
	child := &cobra.Command{
		Use: "child",
	}
	parent.AddCommand(child)

	seeAlso := gen.GenerateSeeAlso(child)
	assert.Contains(t, seeAlso, "parent")
}

func TestGenerator_GenerateSeeAlso_WithSubcommands(t *testing.T) {
	gen := NewGenerator()
	parent := &cobra.Command{
		Use: "parent",
	}
	child1 := &cobra.Command{
		Use: "child1",
	}
	child2 := &cobra.Command{
		Use: "child2",
	}
	parent.AddCommand(child1)
	parent.AddCommand(child2)

	seeAlso := gen.GenerateSeeAlso(parent)
	assert.Contains(t, seeAlso, "child1")
	assert.Contains(t, seeAlso, "child2")
}

func TestGenerator_GenerateSeeAlso_Empty(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use: "standalone",
	}

	seeAlso := gen.GenerateSeeAlso(cmd)
	assert.Empty(t, seeAlso)
}

func TestGenerator_Generate_Basic(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use:   "test [flags] ARGS",
		Short: "Test command",
		Long:  "This is a longer description of the test command",
	}

	result := gen.Generate(cmd)
	assert.Contains(t, result, "Test command")
	assert.Contains(t, result, "longer description")
	assert.Contains(t, result, "Usage:")
}

func TestGenerator_Generate_WithFlags(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	cmd.Flags().StringP("name", "n", "", "Name flag")

	result := gen.Generate(cmd)
	assert.Contains(t, result, "Flags:")
	assert.Contains(t, result, "name")
}

func TestGenerator_Generate_WithExample(t *testing.T) {
	gen := NewGenerator()
	cmd := &cobra.Command{
		Use:     "test",
		Short:   "Test command",
		Example: "  $ test example",
	}

	result := gen.Generate(cmd)
	assert.Contains(t, result, "test example")
}

func TestGenerator_Generate_WithSubcommands(t *testing.T) {
	gen := NewGenerator()
	parent := &cobra.Command{
		Use:   "parent",
		Short: "Parent command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	child := &cobra.Command{
		Use:   "child",
		Short: "Child command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	parent.AddCommand(child)

	result := gen.Generate(parent)

	// Only check if subcommands section exists when there are available subcommands
	if parent.HasAvailableSubCommands() {
		assert.Contains(t, result, "Available Commands:")
		assert.Contains(t, result, "child")
	}
}

func TestGenerator_Wrap(t *testing.T) {
	gen := &Generator{width: 40}
	text := "This is a long text that should wrap to multiple lines when rendered"

	result := gen.wrap(text, 0)
	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 1)

	for _, line := range lines {
		assert.LessOrEqual(t, len(line), 40)
	}
}

func TestGenerator_Wrap_WithIndent(t *testing.T) {
	gen := &Generator{width: 40}
	text := "This text should be indented when wrapped"

	result := gen.wrap(text, 4)
	lines := strings.Split(result, "\n")

	if len(lines) > 1 {
		for i := 1; i < len(lines); i++ {
			assert.True(t, strings.HasPrefix(lines[i], "    "))
		}
	}
}

func TestGenerator_Wrap_Empty(t *testing.T) {
	gen := NewGenerator()
	result := gen.wrap("", 0)
	assert.Empty(t, result)
}

func TestGenerator_Generate_Complete(t *testing.T) {
	gen := NewGenerator()

	parent := &cobra.Command{
		Use:   "parent",
		Short: "Parent command",
		Long:  "This is a detailed description of the parent command",
	}

	child := &cobra.Command{
		Use:   "child",
		Short: "Child command",
	}
	parent.AddCommand(child)

	parent.Flags().StringP("flag1", "f", "", "First flag")
	parent.PersistentFlags().StringP("global", "g", "", "Global flag")

	result := gen.Generate(child)

	assert.NotEmpty(t, result)
	// Should include parent in see also
	assert.Contains(t, result, "See Also:")
}
