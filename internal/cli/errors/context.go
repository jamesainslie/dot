package errors

import (
	"time"

	"github.com/spf13/cobra"
)

// ErrorContext provides additional information for error rendering.
type ErrorContext struct {
	Command   string
	Arguments []string
	Config    ConfigSummary
	Timestamp time.Time
}

// ConfigSummary contains relevant configuration information for error context.
type ConfigSummary struct {
	PackageDir string
	TargetDir  string
	DryRun     bool
	Verbose    int
}

// Extract pulls context from various sources.
func Extract(cmd *cobra.Command, cfg *ConfigSummary) ErrorContext {
	ctx := ErrorContext{
		Timestamp: time.Now(),
	}

	if cmd != nil {
		ctx.Command = cmd.CommandPath()
		ctx.Arguments = cmd.Flags().Args()
	}

	if cfg != nil {
		ctx.Config = *cfg
	}

	return ctx
}

// ExtractCommand creates context from just a command.
func ExtractCommand(cmd *cobra.Command) ErrorContext {
	return Extract(cmd, nil)
}

// ExtractConfig creates context from just config.
func ExtractConfig(cfg *ConfigSummary) ErrorContext {
	return Extract(nil, cfg)
}
