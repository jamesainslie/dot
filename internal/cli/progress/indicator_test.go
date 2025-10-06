package progress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Disabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
	}
	indicator := New(cfg)
	require.NotNil(t, indicator)
	_, ok := indicator.(*NoOpIndicator)
	assert.True(t, ok, "should return NoOpIndicator when disabled")
}

func TestNew_NonInteractive(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: false,
	}
	indicator := New(cfg)
	require.NotNil(t, indicator)
	_, ok := indicator.(*NoOpIndicator)
	assert.True(t, ok, "should return NoOpIndicator when non-interactive")
}

func TestNew_Interactive(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: true,
		Width:       80,
	}
	indicator := New(cfg)
	require.NotNil(t, indicator)
	_, ok := indicator.(*Spinner)
	assert.True(t, ok, "should return Spinner for interactive terminals")
}

func TestNewBar_Disabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
	}
	indicator := NewBar(cfg)
	_, ok := indicator.(*NoOpIndicator)
	assert.True(t, ok)
}

func TestNewBar_Enabled(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: true,
		Width:       80,
	}
	indicator := NewBar(cfg)
	_, ok := indicator.(*Bar)
	assert.True(t, ok)
}

func TestNewSpinner_Disabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
	}
	indicator := NewSpinner(cfg)
	_, ok := indicator.(*NoOpIndicator)
	assert.True(t, ok)
}

func TestNewSpinner_Enabled(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: true,
	}
	indicator := NewSpinner(cfg)
	_, ok := indicator.(*Spinner)
	assert.True(t, ok)
}

func TestNoOpIndicator_Methods(t *testing.T) {
	indicator := &NoOpIndicator{}

	// Should not panic
	indicator.Start("test")
	indicator.Update(5, 10, "update")
	indicator.Stop("done")
	indicator.Fail("failed")
}

func TestConfig_AllFields(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Interactive: true,
		Width:       100,
	}

	assert.True(t, cfg.Enabled)
	assert.True(t, cfg.Interactive)
	assert.Equal(t, 100, cfg.Width)
}

func TestIndicator_Interface(t *testing.T) {
	var _ Indicator = (*NoOpIndicator)(nil)
	var _ Indicator = (*Bar)(nil)
	var _ Indicator = (*Spinner)(nil)
}
