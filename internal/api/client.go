// Package api provides the internal implementation of the public Client interface.
// This package is internal to prevent direct use - consumers should use pkg/dot.
package api

import (
	"fmt"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
)

func init() {
	// Register our implementation with pkg/dot
	dot.RegisterClientImpl(newClient)
}

// client implements the dot.Client interface.
type client struct {
	config     dot.Config
	managePipe *pipeline.ManagePipeline
	executor   *executor.Executor
	manifest   manifest.ManifestStore
}

// newClient creates a new client implementation.
func newClient(cfg dot.Config) (dot.Client, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply defaults
	cfg = cfg.WithDefaults()

	// Create default ignore set
	ignoreSet := ignore.NewDefaultIgnoreSet()

	// Create default resolution policies
	policies := planner.ResolutionPolicies{
		OnFileExists: planner.PolicyFail, // Safe default
	}

	// Convert config types to domain types
	// Since dot types are defined as domain types, we can type assert
	domainFS := cfg.FS.(domain.FS)
	domainLogger := cfg.Logger.(domain.Logger)
	domainTracer := cfg.Tracer.(domain.Tracer)

	// Create manage pipeline
	managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
		FS:        domainFS,
		IgnoreSet: ignoreSet,
		Policies:  policies,
		BackupDir: cfg.BackupDir,
	})

	// Create executor
	exec := executor.New(executor.Opts{
		FS:     domainFS,
		Logger: domainLogger,
		Tracer: domainTracer,
	})

	// Create manifest store
	manifestStore := manifest.NewFSManifestStore(domainFS)

	return &client{
		config:     cfg,
		managePipe: managePipe,
		executor:   exec,
		manifest:   manifestStore,
	}, nil
}

// Config returns the client's configuration.
func (c *client) Config() dot.Config {
	return c.config
}
