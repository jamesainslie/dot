package marshal

import (
	"errors"

	"github.com/jamesainslie/dot/internal/config"
)

// YAMLStrategy implements Strategy for YAML format.
type YAMLStrategy struct{}

// NewYAMLStrategy creates a new YAML marshaling strategy.
func NewYAMLStrategy() *YAMLStrategy {
	return &YAMLStrategy{}
}

// Name returns "yaml".
func (s *YAMLStrategy) Name() string {
	return "yaml"
}

// Marshal converts configuration to YAML bytes.
func (s *YAMLStrategy) Marshal(cfg *config.ExtendedConfig, opts MarshalOptions) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("cannot marshal nil config")
	}
	// TODO: Implement YAML marshaling
	return nil, errors.New("not implemented")
}

// Unmarshal converts YAML bytes to configuration.
func (s *YAMLStrategy) Unmarshal(data []byte) (*config.ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot unmarshal empty data")
	}
	// TODO: Implement YAML unmarshaling
	return nil, errors.New("not implemented")
}
