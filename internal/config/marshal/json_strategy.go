package marshal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

	indent := opts.Indent
	if indent == 0 {
		indent = 2
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", strings.Repeat(" ", indent))

	if err := encoder.Encode(cfg); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return buf.Bytes(), nil
}

// Unmarshal converts JSON bytes to configuration.
func (s *JSONStrategy) Unmarshal(data []byte) (*config.ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot unmarshal empty data")
	}

	var cfg config.ExtendedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	return &cfg, nil
}
