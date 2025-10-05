package dot_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	require.NoError(t, err)
}

func TestConfig_Validate_EmptyStowDir(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "",
		TargetDir: "/test/target",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "stowDir")
}

func TestConfig_Validate_RelativeStowDir(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "relative/path",
		TargetDir: "/test/target",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "stowDir")
	require.Contains(t, err.Error(), "absolute")
}

func TestConfig_Validate_EmptyTargetDir(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "targetDir")
}

func TestConfig_Validate_RelativeTargetDir(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "relative/path",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "targetDir")
	require.Contains(t, err.Error(), "absolute")
}

func TestConfig_Validate_MissingFS(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        nil,
		Logger:    adapters.NewNoopLogger(),
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "FS")
}

func TestConfig_Validate_MissingLogger(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        adapters.NewMemFS(),
		Logger:    nil,
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Logger")
}

func TestConfig_Validate_NegativeVerbosity(t *testing.T) {
	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
		Verbosity: -1,
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "verbosity")
}

func TestConfig_Validate_NegativeConcurrency(t *testing.T) {
	cfg := dot.Config{
		StowDir:     "/test/stow",
		TargetDir:   "/test/target",
		FS:          adapters.NewMemFS(),
		Logger:      adapters.NewNoopLogger(),
		Concurrency: -1,
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "concurrency")
}

