package config

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
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
func (s *TOMLStrategy) Marshal(cfg *ExtendedConfig, opts MarshalOptions) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("cannot marshal nil config")
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)

	if err := encoder.Encode(cfg); err != nil {
		return nil, fmt.Errorf("encode toml: %w", err)
	}

	return buf.Bytes(), nil
}

// Unmarshal converts TOML bytes to configuration.
func (s *TOMLStrategy) Unmarshal(data []byte) (*ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot unmarshal empty data")
	}

	var cfg ExtendedConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal toml: %w", err)
	}

	return &cfg, nil
}
