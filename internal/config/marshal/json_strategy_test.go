package marshal

import (
	"encoding/json"
	"testing"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONStrategy_Name(t *testing.T) {
	strategy := NewJSONStrategy()
	assert.Equal(t, "json", strategy.Name())
}

func TestJSONStrategy_Marshal(t *testing.T) {
	t.Run("marshals configuration to JSON", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Logging.Level = "DEBUG"
		cfg.Output.Verbosity = 2

		strategy := NewJSONStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify JSON is valid and contains expected values
		var decoded config.ExtendedConfig
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, "DEBUG", decoded.Logging.Level)
		assert.Equal(t, 2, decoded.Output.Verbosity)
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		strategy := NewJSONStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(nil, opts)

		require.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "nil config")
	})

	t.Run("respects indent option", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewJSONStrategy()

		// Test with custom indent
		opts := MarshalOptions{
			Indent: 4,
		}

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)
		// JSON should be pretty-printed with 4 spaces
		assert.Contains(t, string(data), "    ") // 4 spaces
	})

	t.Run("handles all config sections", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewJSONStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)

		jsonStr := string(data)
		sections := []string{
			`"directories"`,
			`"logging"`,
			`"symlinks"`,
			`"ignore"`,
			`"dotfile"`,
			`"output"`,
			`"operations"`,
			`"packages"`,
			`"doctor"`,
			`"experimental"`,
		}

		for _, section := range sections {
			assert.Contains(t, jsonStr, section,
				"JSON should contain %s section", section)
		}
	})
}

func TestJSONStrategy_Unmarshal(t *testing.T) {
	t.Run("unmarshals valid JSON to configuration", func(t *testing.T) {
		jsonData := `{
  "logging": {
    "level": "DEBUG",
    "format": "json"
  },
  "output": {
    "verbosity": 3,
    "progress": false
  }
}`
		strategy := NewJSONStrategy()
		cfg, err := strategy.Unmarshal([]byte(jsonData))

		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "DEBUG", cfg.Logging.Level)
		assert.Equal(t, "json", cfg.Logging.Format)
		assert.Equal(t, 3, cfg.Output.Verbosity)
		assert.False(t, cfg.Output.Progress)
	})

	t.Run("returns error for empty data", func(t *testing.T) {
		strategy := NewJSONStrategy()
		cfg, err := strategy.Unmarshal([]byte{})

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{this is not valid json}`)
		strategy := NewJSONStrategy()

		cfg, err := strategy.Unmarshal(invalidJSON)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestJSONStrategy_RoundTrip(t *testing.T) {
	t.Run("marshal then unmarshal preserves data", func(t *testing.T) {
		original := config.DefaultExtended()
		original.Logging.Level = "WARN"
		original.Output.Verbosity = 2
		original.Symlinks.Folding = false
		original.Operations.DryRun = true

		strategy := NewJSONStrategy()
		opts := DefaultMarshalOptions()

		// Marshal
		data, err := strategy.Marshal(original, opts)
		require.NoError(t, err)

		// Unmarshal
		restored, err := strategy.Unmarshal(data)
		require.NoError(t, err)

		// Verify key fields preserved
		assert.Equal(t, original.Logging.Level, restored.Logging.Level)
		assert.Equal(t, original.Output.Verbosity, restored.Output.Verbosity)
		assert.Equal(t, original.Symlinks.Folding, restored.Symlinks.Folding)
		assert.Equal(t, original.Operations.DryRun, restored.Operations.DryRun)
	})
}

func TestJSONStrategy_FormatValidation(t *testing.T) {
	t.Run("output is valid JSON", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewJSONStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		// Should be parseable as JSON
		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err, "marshaled JSON should be valid and parseable")
	})

	t.Run("output is pretty-printed", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewJSONStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		// Should contain newlines and indentation
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "\n", "JSON should be pretty-printed with newlines")
		assert.Contains(t, jsonStr, "  ", "JSON should be indented")
	})
}
