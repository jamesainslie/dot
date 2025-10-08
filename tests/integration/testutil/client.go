package testutil

import (
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

// NewTestClient creates a dot client configured for the test environment.
func NewTestClient(t *testing.T, env *TestEnvironment) *dot.Client {
	t.Helper()

	cfg := dot.Config{
		PackageDir: env.PackageDir,
		TargetDir:  env.TargetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
		LinkMode:   dot.LinkRelative,
		Folding:    true,
		DryRun:     false,
		Verbosity:  0,
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err, "failed to create client")

	return client
}

// NewTestClientWithOptions creates a client with custom configuration.
func NewTestClientWithOptions(t *testing.T, env *TestEnvironment, opts ClientOptions) *dot.Client {
	t.Helper()

	cfg := dot.Config{
		PackageDir:  env.PackageDir,
		TargetDir:   env.TargetDir,
		FS:          opts.FS,
		Logger:      opts.Logger,
		LinkMode:    opts.LinkMode,
		Folding:     opts.Folding,
		DryRun:      opts.DryRun,
		Verbosity:   opts.Verbosity,
		Concurrency: opts.Concurrency,
	}

	// Apply defaults
	if cfg.FS == nil {
		cfg.FS = adapters.NewOSFilesystem()
	}
	if cfg.Logger == nil {
		cfg.Logger = adapters.NewNoopLogger()
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err, "failed to create client")

	return client
}

// ClientOptions holds custom options for client creation.
type ClientOptions struct {
	FS          dot.FS
	Logger      dot.Logger
	LinkMode    dot.LinkMode
	Folding     bool
	DryRun      bool
	Verbosity   int
	Concurrency int
}

// DefaultClientOptions returns default client options.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		FS:        adapters.NewOSFilesystem(),
		Logger:    adapters.NewNoopLogger(),
		LinkMode:  dot.LinkRelative,
		Folding:   true,
		DryRun:    false,
		Verbosity: 0,
	}
}
