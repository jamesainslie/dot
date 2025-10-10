package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_New(t *testing.T) {
	gen := NewGenerator()

	assert.NotNil(t, gen)
}

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name      string
		packages  []string
		installed []string
		opts      GenerateOptions
		wantErr   bool
	}{
		{
			name:      "empty packages",
			packages:  []string{},
			installed: []string{},
			opts:      GenerateOptions{},
			wantErr:   true, // Need at least one package
		},
		{
			name:      "single package",
			packages:  []string{"vim"},
			installed: []string{},
			opts:      GenerateOptions{},
			wantErr:   false,
		},
		{
			name:      "multiple packages",
			packages:  []string{"vim", "zsh", "git"},
			installed: []string{},
			opts:      GenerateOptions{},
			wantErr:   false,
		},
		{
			name:      "packages with installed subset",
			packages:  []string{"vim", "zsh", "git"},
			installed: []string{"vim", "zsh"},
			opts:      GenerateOptions{},
			wantErr:   false,
		},
		{
			name:      "from manifest only",
			packages:  []string{"vim", "zsh", "git"},
			installed: []string{"vim"},
			opts: GenerateOptions{
				FromManifest: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()

			cfg, err := gen.Generate(tt.packages, tt.installed, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "1.0", cfg.Version)
			assert.NotEmpty(t, cfg.Packages)

			if tt.opts.FromManifest {
				// Should only include installed packages
				assert.Len(t, cfg.Packages, len(tt.installed))
			} else {
				// Should include all discovered packages
				assert.Len(t, cfg.Packages, len(tt.packages))
			}
		})
	}
}

func TestGenerator_Generate_PackageDetails(t *testing.T) {
	gen := NewGenerator()
	packages := []string{"vim", "zsh"}
	installed := []string{"vim"}

	cfg, err := gen.Generate(packages, installed, GenerateOptions{})

	require.NoError(t, err)
	require.Len(t, cfg.Packages, 2)

	// Verify package structure
	for _, pkg := range cfg.Packages {
		assert.NotEmpty(t, pkg.Name)
		assert.False(t, pkg.Required) // Default should be false
		assert.Empty(t, pkg.Platform) // Default should be no restrictions
	}
}

func TestGenerator_Generate_ConflictPolicy(t *testing.T) {
	tests := []struct {
		name           string
		conflictPolicy string
		wantPolicy     string
		wantErr        bool
	}{
		{
			name:           "valid backup policy",
			conflictPolicy: "backup",
			wantPolicy:     "backup",
			wantErr:        false,
		},
		{
			name:           "valid fail policy",
			conflictPolicy: "fail",
			wantPolicy:     "fail",
			wantErr:        false,
		},
		{
			name:           "valid overwrite policy",
			conflictPolicy: "overwrite",
			wantPolicy:     "overwrite",
			wantErr:        false,
		},
		{
			name:           "valid skip policy",
			conflictPolicy: "skip",
			wantPolicy:     "skip",
			wantErr:        false,
		},
		{
			name:           "empty uses default",
			conflictPolicy: "",
			wantPolicy:     "backup", // Default
			wantErr:        false,
		},
		{
			name:           "invalid policy",
			conflictPolicy: "invalid",
			wantPolicy:     "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			packages := []string{"vim"}

			opts := GenerateOptions{
				ConflictPolicy: tt.conflictPolicy,
			}

			cfg, err := gen.Generate(packages, []string{}, opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPolicy, cfg.Defaults.ConflictPolicy)
		})
	}
}

func TestGenerator_Generate_Validation(t *testing.T) {
	gen := NewGenerator()

	// Generate should produce valid config
	cfg, err := gen.Generate([]string{"vim", "zsh"}, []string{}, GenerateOptions{})

	require.NoError(t, err)

	// Config should pass validation
	err = cfg.Validate()
	assert.NoError(t, err)
}

func TestGenerator_MarshalYAML(t *testing.T) {
	gen := NewGenerator()

	cfg, err := gen.Generate([]string{"vim"}, []string{}, GenerateOptions{})
	require.NoError(t, err)

	// Should be able to marshal to YAML without error
	data, err := gen.MarshalYAML(cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Should contain expected structure
	assert.Contains(t, string(data), "version:")
	assert.Contains(t, string(data), "packages:")
	assert.Contains(t, string(data), "name: vim")
}

func TestGenerator_MarshalYAML_WithComments(t *testing.T) {
	gen := NewGenerator()

	cfg, err := gen.Generate([]string{"vim"}, []string{"vim"}, GenerateOptions{})
	require.NoError(t, err)

	// Use MarshalYAMLWithComments instead of MarshalYAML
	data, err := gen.MarshalYAMLWithComments(cfg, []string{"vim"})
	require.NoError(t, err)

	// Should include helpful comments
	content := string(data)
	assert.Contains(t, content, "Bootstrap configuration")
	assert.Contains(t, content, "Generated")
}

func TestGenerateOptions_Defaults(t *testing.T) {
	opts := GenerateOptions{}

	// Test that zero values are sensible
	assert.False(t, opts.FromManifest)
	assert.False(t, opts.IncludeComments)
	assert.Empty(t, opts.ConflictPolicy)
}
