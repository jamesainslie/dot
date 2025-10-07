package help

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompletionGenerator(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "root",
	}

	gen := NewCompletionGenerator(rootCmd)
	require.NotNil(t, gen)
	assert.Equal(t, rootCmd, gen.rootCmd)
}

func TestCompletionGenerator_GenerateBash(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	gen := NewCompletionGenerator(rootCmd)
	output, err := gen.GenerateBash()

	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestCompletionGenerator_GenerateZsh(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	gen := NewCompletionGenerator(rootCmd)
	output, err := gen.GenerateZsh()

	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestCompletionGenerator_GenerateFish(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	gen := NewCompletionGenerator(rootCmd)
	output, err := gen.GenerateFish()

	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestCompletionGenerator_GeneratePowerShell(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	gen := NewCompletionGenerator(rootCmd)
	output, err := gen.GeneratePowerShell()

	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestCompletionGenerator_AllShells(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "dot",
		Short: "Dotfile manager",
	}

	// Add some subcommands
	rootCmd.AddCommand(&cobra.Command{Use: "manage", Short: "Manage packages"})
	rootCmd.AddCommand(&cobra.Command{Use: "unmanage", Short: "Unmanage packages"})

	gen := NewCompletionGenerator(rootCmd)

	tests := []struct {
		name string
		fn   func() (string, error)
	}{
		{"bash", gen.GenerateBash},
		{"zsh", gen.GenerateZsh},
		{"fish", gen.GenerateFish},
		{"powershell", gen.GeneratePowerShell},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.fn()
			assert.NoError(t, err)
			assert.NotEmpty(t, output)
		})
	}
}

func TestCompletionGenerator_WithComplexCommand(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "dot",
		Short: "Dotfile manager",
	}

	manageCmd := &cobra.Command{
		Use:   "manage",
		Short: "Manage packages",
	}
	manageCmd.Flags().StringP("dir", "d", "", "Package directory")
	manageCmd.Flags().StringP("target", "t", "", "Target directory")

	rootCmd.AddCommand(manageCmd)

	gen := NewCompletionGenerator(rootCmd)

	output, err := gen.GenerateBash()
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	// Cobra's bash completion generates generic completion script
	assert.Contains(t, output, "bash completion")
}
