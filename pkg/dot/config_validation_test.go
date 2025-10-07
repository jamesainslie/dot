package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
)

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/stow",
		TargetDir:  "/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_EmptyPackageDir(t *testing.T) {
	cfg := dot.Config{
		TargetDir: "/target",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "packageDir")
}

func TestConfig_Validate_RelativePackageDir(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "relative",
		TargetDir:  "/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")
}

func TestConfig_Validate_MissingFS(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/stow",
		TargetDir:  "/target",
		Logger:     adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FS")
}

func TestConfig_Validate_NegativeVerbosity(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/stow",
		TargetDir:  "/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
		Verbosity:  -1,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verbosity")
}
