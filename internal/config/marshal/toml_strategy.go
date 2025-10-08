package marshal

import (
	"errors"

	"github.com/jamesainslie/dot/internal/config"
)

// TOMLStrategy implements Strategy for TOML format.
type TOMLStrategy struct{}

// NewTOMLStrategy creates a new TOML marshaling strategy.
func NewTOMLStrategy() *TOMLStrategy {
	return &TOMLStrategy{}
}

// Name returns "toml".
func (s *TOMLStrategy) Name() string {
	return "toml"
}

// Marshal converts configuration to TOML bytes.
func (s *TOMLStrategy) Marshal(cfg *config.ExtendedConfig, opts MarshalOptions) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("cannot marshal nil config")
	}
	// TODO: Implement TOML marshaling
	return nil, errors.New("not implemented")
}

// Unmarshal converts TOML bytes to configuration.
func (s *TOMLStrategy) Unmarshal(data []byte) (*config.ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot unmarshal empty data")
	}
	// TODO: Implement TOML unmarshaling
	return nil, errors.New("not implemented")
}
