package help

import (
	"bytes"

	"github.com/spf13/cobra"
)

// CompletionGenerator creates shell completion scripts.
type CompletionGenerator struct {
	rootCmd *cobra.Command
}

// NewCompletionGenerator creates a new completion generator.
func NewCompletionGenerator(rootCmd *cobra.Command) *CompletionGenerator {
	return &CompletionGenerator{
		rootCmd: rootCmd,
	}
}

// GenerateBash creates bash completion.
func (g *CompletionGenerator) GenerateBash() (string, error) {
	var buf bytes.Buffer
	err := g.rootCmd.GenBashCompletion(&buf)
	return buf.String(), err
}

// GenerateZsh creates zsh completion.
func (g *CompletionGenerator) GenerateZsh() (string, error) {
	var buf bytes.Buffer
	err := g.rootCmd.GenZshCompletion(&buf)
	return buf.String(), err
}

// GenerateFish creates fish completion.
func (g *CompletionGenerator) GenerateFish() (string, error) {
	var buf bytes.Buffer
	err := g.rootCmd.GenFishCompletion(&buf, true)
	return buf.String(), err
}

// GeneratePowerShell creates PowerShell completion.
func (g *CompletionGenerator) GeneratePowerShell() (string, error) {
	var buf bytes.Buffer
	err := g.rootCmd.GenPowerShellCompletionWithDesc(&buf)
	return buf.String(), err
}
