package marshal

import (
	"errors"

	"github.com/jamesainslie/dot/internal/config"
)

// JSONStrategy implements Strategy for JSON format.
type JSONStrategy struct{}

// NewJSONStrategy creates a new JSON marshaling strategy.
func NewJSONStrategy() *JSONStrategy {
	return &JSONStrategy{}
}

// Name returns "json".
func (s *JSONStrategy) Name() string {
	return "json"
}

// Marshal converts configuration to JSON bytes.
func (s *JSONStrategy) Marshal(cfg *config.ExtendedConfig, opts MarshalOptions) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("cannot marshal nil config")
	}
	// TODO: Implement JSON marshaling
	return nil, errors.New("not implemented")
}

// Unmarshal converts JSON bytes to configuration.
func (s *JSONStrategy) Unmarshal(data []byte) (*config.ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot unmarshal empty data")
	}
	// TODO: Implement JSON unmarshaling
	return nil, errors.New("not implemented")
}
