package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestNewListCommand(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewListCommand(cfg)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "list")
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestListCommand_Flags(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewListCommand(cfg)

	// Check that format flag exists (default is table for list)
	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "table", formatFlag.DefValue)

	// Check that color flag exists
	colorFlag := cmd.Flags().Lookup("color")
	require.NotNil(t, colorFlag)
	assert.Equal(t, "auto", colorFlag.DefValue)

	// Check that sort flag exists
	sortFlag := cmd.Flags().Lookup("sort")
	require.NotNil(t, sortFlag)
	assert.Equal(t, "name", sortFlag.DefValue)
}

func TestListCommand_Help(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewListCommand(cfg)

	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}
