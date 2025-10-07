package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
)

func TestConfig_Validate_EmptyTargetDir(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/stow",
		// TargetDir is empty
		FS:     adapters.NewMemFS(),
		Logger: adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "targetDir")
}

func TestConfig_Validate_RelativeTargetDir(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/stow",
		TargetDir:  "relative/path",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")
}

func TestConfig_Validate_MissingLogger(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/stow",
		TargetDir:  "/target",
		FS:         adapters.NewMemFS(),
		// Logger is nil
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Logger")
}

func TestConfig_Validate_NegativeConcurrency(t *testing.T) {
	cfg := dot.Config{
		PackageDir:  "/stow",
		TargetDir:   "/target",
		FS:          adapters.NewMemFS(),
		Logger:      adapters.NewNoopLogger(),
		Concurrency: -1,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "concurrency")
}
