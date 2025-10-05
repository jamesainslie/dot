package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand_Structure(t *testing.T) {
	cmd := NewRootCommand("v1.0.0", "abc123", "2024-01-01")

	assert.Equal(t, "dot", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Version, "v1.0.0")
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)
}

func TestNewRootCommand_HasSubcommands(t *testing.T) {
	cmd := NewRootCommand("dev", "none", "unknown")

	commands := cmd.Commands()
	assert.GreaterOrEqual(t, len(commands), 7) // manage, unmanage, remanage, adopt, status, list, doctor
}

func TestNewRootCommand_HasGlobalFlags(t *testing.T) {
	cmd := NewRootCommand("dev", "none", "unknown")

	assert.NotNil(t, cmd.PersistentFlags().Lookup("dir"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("target"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("dry-run"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("verbose"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("quiet"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("log-json"))
}
