package marshal

import (
	"strings"
	"testing"

	"github.com/jamesainslie/dot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLStrategy_Name(t *testing.T) {
	strategy := NewYAMLStrategy()
	assert.Equal(t, "yaml", strategy.Name())
}

func TestYAMLStrategy_Marshal(t *testing.T) {
	t.Run("marshals configuration to YAML", func(t *testing.T) {
		cfg := config.DefaultExtended()
		cfg.Logging.Level = "DEBUG"
		cfg.Output.Verbosity = 2

		strategy := NewYAMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify YAML contains expected values
		yamlStr := string(data)
		assert.Contains(t, yamlStr, "level: DEBUG")
		assert.Contains(t, yamlStr, "verbosity: 2")
	})

	t.Run("marshals with comments when requested", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewYAMLStrategy()
		opts := MarshalOptions{
			IncludeComments: true,
			Indent:          2,
		}

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)
		assert.NotEmpty(t, data)

		yamlStr := string(data)
		// Should contain comments
		assert.Contains(t, yamlStr, "# Dot Configuration File")
		assert.Contains(t, yamlStr, "# Core Directories")
		assert.Contains(t, yamlStr, "# Logging Configuration")
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		strategy := NewYAMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(nil, opts)

		require.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "nil config")
	})

	t.Run("handles all config sections", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewYAMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)

		require.NoError(t, err)

		yamlStr := string(data)
		sections := []string{
			"directories:",
			"logging:",
			"symlinks:",
			"ignore:",
			"dotfile:",
			"output:",
			"operations:",
			"packages:",
			"doctor:",
			"experimental:",
		}

		for _, section := range sections {
			assert.Contains(t, yamlStr, section,
				"YAML should contain %s section", section)
		}
	})
}

func TestYAMLStrategy_Unmarshal(t *testing.T) {
	t.Run("unmarshals valid YAML to configuration", func(t *testing.T) {
		yamlData := `
logging:
  level: DEBUG
  format: json
output:
  verbosity: 3
  progress: false
`
		strategy := NewYAMLStrategy()
		cfg, err := strategy.Unmarshal([]byte(yamlData))

		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "DEBUG", cfg.Logging.Level)
		assert.Equal(t, "json", cfg.Logging.Format)
		assert.Equal(t, 3, cfg.Output.Verbosity)
		assert.False(t, cfg.Output.Progress)
	})

	t.Run("returns error for empty data", func(t *testing.T) {
		strategy := NewYAMLStrategy()
		cfg, err := strategy.Unmarshal([]byte{})

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		invalidYAML := []byte("this is not valid: {{{{ yaml")
		strategy := NewYAMLStrategy()

		cfg, err := strategy.Unmarshal(invalidYAML)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("handles partial configuration", func(t *testing.T) {
		// Only some fields specified - YAML unmarshal sets zero values for unspecified fields
		partialYAML := `
logging:
  level: ERROR
`
		strategy := NewYAMLStrategy()
		cfg, err := strategy.Unmarshal([]byte(partialYAML))

		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Specified value
		assert.Equal(t, "ERROR", cfg.Logging.Level)
		// Unspecified fields have zero values (empty string for string fields)
		// Note: Applying defaults is the responsibility of the caller, not the strategy
		assert.Equal(t, "", cfg.Logging.Format)
	})
}

func TestYAMLStrategy_RoundTrip(t *testing.T) {
	t.Run("marshal then unmarshal preserves data", func(t *testing.T) {
		original := config.DefaultExtended()
		original.Logging.Level = "WARN"
		original.Output.Verbosity = 2
		original.Symlinks.Folding = false
		original.Operations.DryRun = true

		strategy := NewYAMLStrategy()
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

	t.Run("round trip with comments preserves data", func(t *testing.T) {
		original := config.DefaultExtended()
		original.Logging.Level = "DEBUG"

		strategy := NewYAMLStrategy()
		opts := MarshalOptions{
			IncludeComments: true,
			Indent:          2,
		}

		// Marshal with comments
		data, err := strategy.Marshal(original, opts)
		require.NoError(t, err)

		// Unmarshal (comments should be ignored)
		restored, err := strategy.Unmarshal(data)
		require.NoError(t, err)

		assert.Equal(t, original.Logging.Level, restored.Logging.Level)
	})
}

func TestYAMLStrategy_Comments(t *testing.T) {
	t.Run("comments include all major sections", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewYAMLStrategy()
		opts := MarshalOptions{IncludeComments: true}

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		yamlStr := string(data)
		commentSections := []string{
			"# Dot Configuration File",
			"# Core Directories",
			"# Logging Configuration",
			"# Symlink Behavior",
			"# Ignore Patterns",
			"# Dotfile Translation",
			"# Output Configuration",
			"# Operation Defaults",
			"# Package Management",
			"# Doctor Configuration",
			"# Experimental Features",
		}

		for _, section := range commentSections {
			assert.Contains(t, yamlStr, section,
				"commented YAML should include section: %s", section)
		}
	})

	t.Run("comments include field descriptions", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewYAMLStrategy()
		opts := MarshalOptions{IncludeComments: true}

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		yamlStr := string(data)
		descriptions := []string{
			"# Log level: DEBUG, INFO, WARN, ERROR",
			"# Link mode: relative, absolute",
			"# Enable directory folding optimization",
			"# Verbosity level: 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)",
		}

		for _, desc := range descriptions {
			assert.Contains(t, yamlStr, desc,
				"commented YAML should include description")
		}
	})
}

func TestYAMLStrategy_FormatValidation(t *testing.T) {
	t.Run("output is valid YAML", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewYAMLStrategy()
		opts := DefaultMarshalOptions()

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		// Should be parseable as YAML
		_, err = strategy.Unmarshal(data)
		assert.NoError(t, err, "marshaled YAML should be valid and parseable")
	})

	t.Run("no trailing whitespace on lines", func(t *testing.T) {
		cfg := config.DefaultExtended()
		strategy := NewYAMLStrategy()
		opts := MarshalOptions{IncludeComments: true}

		data, err := strategy.Marshal(cfg, opts)
		require.NoError(t, err)

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			assert.Equal(t, strings.TrimRight(line, " \t"), line,
				"line %d should not have trailing whitespace", i+1)
		}
	})
}
