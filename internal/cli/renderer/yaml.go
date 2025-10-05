package renderer

import (
	"io"

	"gopkg.in/yaml.v3"

	"github.com/jamesainslie/dot/pkg/dot"
)

// YAMLRenderer renders output as YAML.
type YAMLRenderer struct {
	indent int
}

// RenderStatus renders installation status as YAML.
func (r *YAMLRenderer) RenderStatus(w io.Writer, status dot.Status) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(r.indent)
	defer encoder.Close()
	return encoder.Encode(status)
}
