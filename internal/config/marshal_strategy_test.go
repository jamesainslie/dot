package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrategyInterface(t *testing.T) {
	t.Run("Strategy interface exists", func(t *testing.T) {
		// This test verifies the Strategy interface can be used as a type
		var _ Strategy = (*mockStrategy)(nil)
	})

	t.Run("MarshalOptions has expected fields", func(t *testing.T) {
		opts := MarshalOptions{
			IncludeComments: true,
			Indent:          2,
		}

		assert.True(t, opts.IncludeComments)
		assert.Equal(t, 2, opts.Indent)
	})
}

func TestMarshalOptionsDefaults(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := DefaultMarshalOptions()

		assert.False(t, opts.IncludeComments)
		assert.Equal(t, 2, opts.Indent)
	})

	t.Run("options are independent", func(t *testing.T) {
		opts1 := DefaultMarshalOptions()
		opts2 := DefaultMarshalOptions()

		opts1.IncludeComments = true
		opts1.Indent = 4

		assert.False(t, opts2.IncludeComments, "modifying opts1 should not affect opts2")
		assert.Equal(t, 2, opts2.Indent, "modifying opts1 should not affect opts2")
	})
}

func TestStrategySelection(t *testing.T) {
	t.Run("GetStrategy returns correct strategy for format", func(t *testing.T) {
		tests := []struct {
			format   string
			expected string
		}{
			{"yaml", "yaml"},
			{"yml", "yaml"},
			{"json", "json"},
			{"toml", "toml"},
		}

		for _, tt := range tests {
			t.Run(tt.format, func(t *testing.T) {
				strategy, err := GetStrategy(tt.format)
				require.NoError(t, err)
				assert.NotNil(t, strategy)
				assert.Equal(t, tt.expected, strategy.Name())
			})
		}
	})

	t.Run("GetStrategy returns error for unknown format", func(t *testing.T) {
		strategy, err := GetStrategy("unknown")

		require.Error(t, err)
		assert.Nil(t, strategy)
		assert.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("GetStrategy is case insensitive", func(t *testing.T) {
		formats := []string{"YAML", "Json", "ToMl"}

		for _, format := range formats {
			strategy, err := GetStrategy(format)
			require.NoError(t, err)
			assert.NotNil(t, strategy)
		}
	})
}

func TestStrategyRoundTrip(t *testing.T) {
	t.Run("strategy preserves configuration data", func(t *testing.T) {
		cfg := DefaultExtended()
		cfg.Logging.Level = "DEBUG"
		cfg.Output.Verbosity = 2

		// Test each strategy
		formats := []string{"yaml", "json", "toml"}

		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				strategy, err := GetStrategy(format)
				require.NoError(t, err)

				// Marshal
				data, err := strategy.Marshal(cfg, DefaultMarshalOptions())
				require.NoError(t, err)
				assert.NotEmpty(t, data)

				// Unmarshal
				decoded, err := strategy.Unmarshal(data)
				require.NoError(t, err)
				assert.NotNil(t, decoded)

				// Verify key fields preserved
				assert.Equal(t, "DEBUG", decoded.Logging.Level)
				assert.Equal(t, 2, decoded.Output.Verbosity)
			})
		}
	})
}

func TestStrategyErrorHandling(t *testing.T) {
	t.Run("Marshal handles nil config", func(t *testing.T) {
		strategy, _ := GetStrategy("yaml")

		data, err := strategy.Marshal(nil, DefaultMarshalOptions())

		require.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "nil config")
	})

	t.Run("Unmarshal handles empty data", func(t *testing.T) {
		strategy, _ := GetStrategy("yaml")

		cfg, err := strategy.Unmarshal([]byte{})

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Unmarshal handles invalid data", func(t *testing.T) {
		strategy, _ := GetStrategy("yaml")
		invalidData := []byte("this is not valid yaml: {{{{")

		cfg, err := strategy.Unmarshal(invalidData)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})
}

// mockStrategy is a test implementation of Strategy
type mockStrategy struct {
	name string
}

func (m *mockStrategy) Name() string {
	return m.name
}

func (m *mockStrategy) Marshal(cfg *ExtendedConfig, opts MarshalOptions) ([]byte, error) {
	if cfg == nil {
		return nil, assert.AnError
	}
	return []byte("mock marshaled data"), nil
}

func (m *mockStrategy) Unmarshal(data []byte) (*ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, assert.AnError
	}
	return DefaultExtended(), nil
}
