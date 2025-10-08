package marshal

import (
	"fmt"
	"strings"

	"github.com/jamesainslie/dot/internal/config"
)

// Strategy defines the interface for configuration marshaling and unmarshaling.
// Each format (YAML, JSON, TOML) implements this interface.
type Strategy interface {
	// Name returns the format name (e.g., "yaml", "json", "toml")
	Name() string

	// Marshal converts configuration to bytes in the strategy's format
	Marshal(cfg *config.ExtendedConfig, opts MarshalOptions) ([]byte, error)

	// Unmarshal converts bytes to configuration from the strategy's format
	Unmarshal(data []byte) (*config.ExtendedConfig, error)
}

// MarshalOptions controls marshaling behavior.
type MarshalOptions struct {
	// IncludeComments adds explanatory comments to output (format-dependent)
	IncludeComments bool

	// Indent specifies indentation size (spaces)
	Indent int
}

// DefaultMarshalOptions returns sensible default marshaling options.
func DefaultMarshalOptions() MarshalOptions {
	return MarshalOptions{
		IncludeComments: false,
		Indent:          2,
	}
}

// GetStrategy returns the appropriate strategy for the given format.
// Format strings are case-insensitive.
// Supported formats: yaml, yml, json, toml
func GetStrategy(format string) (Strategy, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))

	switch normalized {
	case "yaml", "yml":
		return NewYAMLStrategy(), nil
	case "json":
		return NewJSONStrategy(), nil
	case "toml":
		return NewTOMLStrategy(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: yaml, json, toml)", format)
	}
}
