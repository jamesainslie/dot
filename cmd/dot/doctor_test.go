package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestNewDoctorCommand(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewDoctorCommand(cfg)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "doctor")
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestDoctorCommand_Flags(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewDoctorCommand(cfg)

	// Check that format flag exists
	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Check that color flag exists
	colorFlag := cmd.Flags().Lookup("color")
	require.NotNil(t, colorFlag)
	assert.Equal(t, "auto", colorFlag.DefValue)
}

func TestDoctorCommand_Help(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewDoctorCommand(cfg)

	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Long, "Exit codes")
}
