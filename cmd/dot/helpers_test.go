package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatError(t *testing.T) {
	err := errors.New("test error")
	result := formatError(err)
	assert.Equal(t, err, result)
}

func TestShouldColorize(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{"always", "always", true},
		{"never", "never", false},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldColorize(tt.color)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestShouldColorize_Auto(t *testing.T) {
	// Auto detection depends on if stdout is a TTY
	// In tests, it's typically not a TTY, so should be false
	result := shouldColorize("auto")
	// Just verify it doesn't panic, actual result depends on environment
	_ = result
}

func TestBuildConfig_ValidatesPackageDir(t *testing.T) {
	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: ".",
		targetDir:  ".",
	}

	cfg, err := buildConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.PackageDir)
}

func TestCreateLogger_AllModes(t *testing.T) {
	tests := []struct {
		name    string
		quiet   bool
		logJSON bool
		verbose int
	}{
		{"quiet", true, false, 0},
		{"json", false, true, 0},
		{"text", false, false, 0},
		{"verbose", false, false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := globalCfg
			t.Cleanup(func() {
				globalCfg = previous
			})

			globalCfg = globalConfig{
				quiet:   tt.quiet,
				logJSON: tt.logJSON,
				verbose: tt.verbose,
			}

			logger := createLogger()
			assert.NotNil(t, logger)
		})
	}
}
