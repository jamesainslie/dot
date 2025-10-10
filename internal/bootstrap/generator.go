package bootstrap

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Generator creates bootstrap configurations from package information.
type Generator struct{}

// NewGenerator creates a new bootstrap configuration generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateOptions configures bootstrap generation behavior.
type GenerateOptions struct {
	// FromManifest only includes packages present in manifest
	FromManifest bool

	// ConflictPolicy sets default conflict resolution policy
	ConflictPolicy string

	// IncludeComments adds helpful comments to generated config
	IncludeComments bool
}

// Generate creates a bootstrap configuration from package information.
//
// Parameters:
//   - packages: All discovered package names
//   - installed: Package names that are currently installed
//   - opts: Generation options
//
// Returns a validated bootstrap configuration or an error.
func (g *Generator) Generate(packages []string, installed []string, opts GenerateOptions) (Config, error) {
	// Validate inputs
	if len(packages) == 0 {
		return Config{}, fmt.Errorf("no packages provided")
	}

	// Determine which packages to include
	var pkgNames []string
	if opts.FromManifest {
		pkgNames = installed
	} else {
		pkgNames = packages
	}

	if len(pkgNames) == 0 {
		return Config{}, fmt.Errorf("no packages to include in configuration")
	}

	// Build package specs
	pkgSpecs := make([]PackageSpec, 0, len(pkgNames))

	for _, name := range pkgNames {
		spec := PackageSpec{
			Name:     name,
			Required: false, // User must explicitly mark packages as required
			Platform: nil,   // User must explicitly set platform restrictions
		}
		pkgSpecs = append(pkgSpecs, spec)
	}

	// Determine conflict policy
	conflictPolicy := opts.ConflictPolicy
	if conflictPolicy == "" {
		conflictPolicy = "backup" // Safe default
	}

	// Validate conflict policy
	if !isValidConflictPolicy(conflictPolicy) {
		return Config{}, fmt.Errorf("invalid conflict policy: %s", conflictPolicy)
	}

	// Build configuration
	cfg := Config{
		Version:  "1.0",
		Packages: pkgSpecs,
		Profiles: nil, // User must define profiles
		Defaults: Defaults{
			ConflictPolicy: conflictPolicy,
			Profile:        "", // No default profile
		},
	}

	// Validate generated configuration
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("generated invalid configuration: %w", err)
	}

	return cfg, nil
}

// MarshalYAML converts configuration to YAML bytes.
//
// The output is formatted for human readability with proper indentation
// and ordering of fields.
func (g *Generator) MarshalYAML(cfg Config) ([]byte, error) {
	// Create YAML encoder with custom settings
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal YAML: %w", err)
	}

	return data, nil
}

// MarshalYAMLWithComments converts configuration to YAML with helpful comments.
func (g *Generator) MarshalYAMLWithComments(cfg Config, installed []string) ([]byte, error) {
	// Build header comment
	header := fmt.Sprintf(`# Bootstrap configuration for dotfiles repository
# Generated: %s
#
# This configuration defines packages and installation profiles.
# Review and customize before committing to your repository.
#
# Documentation: https://github.com/jamesainslie/dot/docs/user/bootstrap-config-spec.md

`, time.Now().Format(time.RFC3339))

	// Marshal basic config
	data, err := g.MarshalYAML(cfg)
	if err != nil {
		return nil, err
	}

	// Add comments for each package
	// Note: This is a basic implementation. A more sophisticated approach
	// would use yaml.Node to insert comments directly into the YAML structure.
	result := []byte(header)
	result = append(result, data...)

	// Add example profiles at the end
	profileExample := `
# Profiles define named package sets for different use cases
# Uncomment and customize as needed:
#
# profiles:
#   minimal:
#     description: Minimal shell setup
#     packages:
#       - zsh
#       - git
#
#   full:
#     description: Complete development environment
#     packages:
`
	for _, pkg := range cfg.Packages {
		profileExample += fmt.Sprintf("#       - %s\n", pkg.Name)
	}

	result = append(result, []byte(profileExample)...)

	return result, nil
}

// makeSet creates a set from a slice for O(1) membership testing.
func makeSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}
