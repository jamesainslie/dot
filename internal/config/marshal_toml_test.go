package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOMLStrategy_Name(t *testing.T) {
	strategy := NewTOMLStrategy()
	assert.Equal(t, "toml", strategy.Name())
}

func TestTOMLStrategy_Marshal(t *testing.T) {
	t.Run("marshals configuration to TOML", func(t *testing.T) {
		cfg := DefaultExtended()
		cfg.Logging.Level = "DEBUG"
		cfg.Output.Verbosity = 2

		strategy := NewTOMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify TOML is valid and contains expected values
		var decoded ExtendedConfig
		err = toml.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, "DEBUG", decoded.Logging.Level)
		assert.Equal(t, 2, decoded.Output.Verbosity)
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		strategy := NewTOMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(nil, opts)

		require.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "nil config")
	})

	t.Run("handles all config sections", func(t *testing.T) {
		cfg := DefaultExtended()
		strategy := NewTOMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)

		tomlStr := string(data)
		sections := []string{
			"[directories]",
			"[logging]",
			"[symlinks]",
			"[ignore]",
			"[dotfile]",
			"[output]",
			"[operations]",
			"[packages]",
			"[doctor]",
			"[experimental]",
		}

		for _, section := range sections {
			assert.Contains(t, tomlStr, section,
				"TOML should contain %s section", section)
		}
	})
}

func TestTOMLStrategy_Unmarshal(t *testing.T) {
	t.Run("unmarshals valid TOML to configuration", func(t *testing.T) {
		tomlData := `
[logging]
level = "DEBUG"
format = "json"

[output]
verbosity = 3
progress = false
`
		strategy := NewTOMLStrategy()
		cfg, err := strategy.Unmarshal([]byte(tomlData))

		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "DEBUG", cfg.Logging.Level)
		assert.Equal(t, "json", cfg.Logging.Format)
		assert.Equal(t, 3, cfg.Output.Verbosity)
		assert.False(t, cfg.Output.Progress)
	})

	t.Run("returns error for empty data", func(t *testing.T) {
		strategy := NewTOMLStrategy()
		cfg, err := strategy.Unmarshal([]byte{})

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		invalidTOML := []byte(`[this is not valid toml`)
		strategy := NewTOMLStrategy()

		cfg, err := strategy.Unmarshal(invalidTOML)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestTOMLStrategy_RoundTrip(t *testing.T) {
	t.Run("marshal then unmarshal preserves data", func(t *testing.T) {
		original := DefaultExtended()
		original.Logging.Level = "WARN"
		original.Output.Verbosity = 2
		original.Symlinks.Folding = false
		original.Operations.DryRun = true

		strategy := NewTOMLStrategy()
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

func TestTOMLStrategy_FormatValidation(t *testing.T) {
	t.Run("output is valid TOML", func(t *testing.T) {
		cfg := DefaultExtended()
		strategy := NewTOMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		// Should be parseable as TOML
		var decoded map[string]interface{}
		err = toml.Unmarshal(data, &decoded)
		assert.NoError(t, err, "marshaled TOML should be valid and parseable")
	})
}
