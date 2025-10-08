package dot_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

// TestNewClient verifies Client creation.
func TestNewClient(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify config is stored
	clientCfg := client.Config()
	require.Equal(t, cfg.PackageDir, clientCfg.PackageDir)
	require.Equal(t, cfg.TargetDir, clientCfg.TargetDir)
}

// TestNewClient_InvalidConfig verifies validation errors.
func TestNewClient_InvalidConfig(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "relative/path", // Invalid - not absolute
	}

	client, err := dot.NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "invalid configuration")
}
