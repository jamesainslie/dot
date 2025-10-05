# Phase 12: Public Library API - Implementation Plan (Option 4: Interface Pattern)

## Overview

Phase 12 delivers a clean, embeddable Go API for the dot library using an interface-based pattern to avoid import cycles. The Client interface lives in `pkg/dot/` while the implementation resides in `internal/api/`, allowing the implementation to import all necessary internal packages without creating circular dependencies.

**Architecture Strategy**: Interface in public package, implementation in internal package.

**Dependencies**: Phases 1-11 must be complete (domain model, ports, adapters, scanner, planner, resolver, sorter, pipeline orchestration, executor, manifest management).

**Deliverable**: Clean, tested public library API suitable for embedding in other tools.

## Architectural Pattern

### Import Cycle Resolution

**The Problem**: Domain types in `pkg/dot/` are imported by `internal/*` packages, preventing `pkg/dot/` from importing those internal packages.

**The Solution**: Interface segregation pattern
```
pkg/dot/client.go           # Client interface definition
    ↓ depends on
internal/api/client.go      # Client implementation
    ↓ imports (no cycle!)
internal/executor           # Can import pkg/dot for domain types
internal/pipeline           # Can import pkg/dot for domain types
```

**Key Insight**: Interfaces can be defined without importing implementations. The constructor returns the interface, hiding the concrete type.

### Package Structure

```
pkg/dot/
├── client.go          # Client interface + constructor
├── client_test.go     # Client interface tests
├── config.go          # Configuration (existing)
├── config_test.go     # Configuration tests (existing)
├── ports.go           # Port interfaces (existing)
├── execution.go       # ExecutionResult (existing)
├── checkpoint.go      # Checkpoint types (existing)
├── types.go           # Domain type exports (existing)
├── doc.go             # Package documentation
└── examples_test.go   # Example tests for godoc

internal/api/
├── client.go          # Client implementation
├── client_test.go     # Implementation tests
├── manage.go          # Manage operations
├── manage_test.go     # Manage tests
├── status.go          # Status operations  
└── status_test.go     # Status tests
```

## Design Principles

- **Interface Segregation**: Public interface, private implementation
- **Zero Breaking Changes**: No modifications to Phases 1-11
- **Type Safety**: Leverage existing domain types
- **Context-Aware**: All operations accept context.Context
- **Testability**: Interface enables easy mocking
- **Future-Proof**: Implementation can evolve without API changes

---

## Task Breakdown

### 12.1: Client Interface Definition (Priority: Critical)

#### 12.1.1: Core Client Interface

**File**: `pkg/dot/client.go`

**Test-First Approach**:
```go
// pkg/dot/client_test.go
package dot_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func testConfig() dot.Config {
	return dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        adapters.NewMemFS(),
		Logger:    adapters.NewNoopLogger(),
	}
}

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := testConfig()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient_InvalidConfig(t *testing.T) {
	cfg := dot.Config{
		StowDir: "relative/path", // Invalid
	}

	client, err := dot.NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "invalid configuration")
}
```

**Implementation**:
```go
// pkg/dot/client.go
package dot

import "context"

// Client provides the high-level API for dot operations.
// This interface abstracts the internal pipeline and executor orchestration,
// providing a simple facade for library consumers.
type Client interface {
	// Manage operations
	Manage(ctx context.Context, packages ...string) error
	PlanManage(ctx context.Context, packages ...string) (Plan, error)

	// Unmanage operations
	Unmanage(ctx context.Context, packages ...string) error
	PlanUnmanage(ctx context.Context, packages ...string) (Plan, error)

	// Remanage operations
	Remanage(ctx context.Context, packages ...string) error
	PlanRemanage(ctx context.Context, packages ...string) (Plan, error)

	// Adopt operations
	Adopt(ctx context.Context, files []string, pkg string) error
	PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error)

	// Query operations
	Status(ctx context.Context, packages ...string) (Status, error)
	List(ctx context.Context) ([]PackageInfo, error)

	// Configuration
	Config() Config
}

// NewClient creates a new Client with the given configuration.
// Returns an error if configuration is invalid.
//
// The returned Client is safe for concurrent use from multiple goroutines.
func NewClient(cfg Config) (Client, error) {
	return newClientImpl(cfg)
}

// newClientImpl is implemented in internal/api to avoid import cycles.
var newClientImpl func(Config) (Client, error)

// RegisterClientImpl is called by internal/api during initialization.
// This allows the implementation to be in internal/api while the interface
// remains in pkg/dot.
func RegisterClientImpl(fn func(Config) (Client, error)) {
	newClientImpl = fn
}
```

**Tasks**:
- [ ] Define Client interface with all operations
- [ ] Add NewClient constructor function
- [ ] Add registration mechanism for implementation
- [ ] Write tests for constructor with valid config
- [ ] Write tests for constructor with invalid config
- [ ] Document interface methods

**Commit**: `feat(api): define Client interface for public API`

---

### 12.2: Client Implementation (Priority: Critical)

#### 12.2.1: Implementation Structure

**File**: `internal/api/client.go`

**Implementation**:
```go
// Package api provides the internal implementation of the public Client interface.
// This package is internal to prevent direct use - consumers should use pkg/dot.
package api

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/internal/executor"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/pkg/dot"
)

func init() {
	// Register our implementation with pkg/dot
	dot.RegisterClientImpl(newClient)
}

// client implements the dot.Client interface.
type client struct {
	config   dot.Config
	stowPipe *pipeline.StowPipeline
	executor *executor.Executor
	manifest manifest.Store
}

// newClient creates a new client implementation.
func newClient(cfg dot.Config) (dot.Client, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply defaults
	cfg = cfg.WithDefaults()

	// Create stow pipeline
	stowPipe := pipeline.NewStowPipeline(pipeline.StowPipelineOpts{
		FS:     cfg.FS,
		Logger: cfg.Logger,
		Tracer: cfg.Tracer,
	})

	// Create executor
	exec := executor.New(executor.Opts{
		FS:     cfg.FS,
		Logger: cfg.Logger,
		Tracer: cfg.Tracer,
	})

	// Create manifest store
	manifestStore := manifest.NewFSStore(cfg.FS)

	return &client{
		config:   cfg,
		stowPipe: stowPipe,
		executor: exec,
		manifest: manifestStore,
	}, nil
}

// Config returns the client's configuration.
func (c *client) Config() dot.Config {
	return c.config
}
```

**Tasks**:
- [ ] Create internal/api package
- [ ] Implement client struct
- [ ] Implement newClient constructor
- [ ] Add init() to register implementation
- [ ] Implement Config() accessor
- [ ] Write tests for client creation
- [ ] Verify registration mechanism works

**Commit**: `feat(api): implement Client interface in internal package`

---

#### 12.2.2: Manage Operations

**Test-First Approach**:
```go
// internal/api/manage_test.go
package api

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func setupTestFixtures(t *testing.T, fs dot.FS, packages ...string) {
	t.Helper()
	ctx := context.Background()

	// Create stow directory structure
	for _, pkg := range packages {
		pkgDir := filepath.Join("/test/stow", pkg)
		require.NoError(t, fs.MkdirAll(ctx, pkgDir, 0755))

		// Create sample dotfile
		dotfile := filepath.Join(pkgDir, "dot-config")
		require.NoError(t, fs.WriteFile(ctx, dotfile, []byte("test content"), 0644))
	}

	// Create target directory
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
}

func TestClient_Manage(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1")

	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        fs,
		Logger:    adapters.NewNoopLogger(),
	}

	client, err := newClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Manage(ctx, "package1")
	require.NoError(t, err)

	// Verify link created
	isLink, err := fs.IsSymlink(ctx, "/test/target/.config")
	require.NoError(t, err)
	require.True(t, isLink)
}

func TestClient_PlanManage(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1")

	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        fs,
		Logger:    adapters.NewNoopLogger(),
	}

	client, err := newClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	plan, err := client.PlanManage(ctx, "package1")
	require.NoError(t, err)
	require.NotEmpty(t, plan.Operations)
}

func TestClient_Manage_DryRun(t *testing.T) {
	fs := adapters.NewMemFS()
	setupTestFixtures(t, fs, "package1")

	cfg := dot.Config{
		StowDir:   "/test/stow",
		TargetDir: "/test/target",
		FS:        fs,
		Logger:    adapters.NewNoopLogger(),
		DryRun:    true,
	}

	client, err := newClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Manage(ctx, "package1")
	require.NoError(t, err)

	// Verify no links created (dry-run)
	exists := fs.Exists(ctx, "/test/target/.config")
	require.False(t, exists)
}
```

**Implementation**:
```go
// internal/api/manage.go
package api

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Manage installs the specified packages by creating symlinks.
func (c *client) Manage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanManage(ctx, packages...)
	if err != nil {
		return err
	}

	if c.config.DryRun {
		c.config.Logger.Info(ctx, "dry_run_mode", "operations", len(plan.Operations))
		return nil
	}

	result := c.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}

	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}

	return nil
}

// PlanManage computes the execution plan for managing packages without applying changes.
func (c *client) PlanManage(ctx context.Context, packages ...string) (dot.Plan, error) {
	pkgPath, err := dot.NewPackagePath(c.config.StowDir)
	if err != nil {
		return dot.Plan{}, fmt.Errorf("invalid stow directory: %w", err)
	}

	targetPath, err := dot.NewTargetPath(c.config.TargetDir)
	if err != nil {
		return dot.Plan{}, fmt.Errorf("invalid target directory: %w", err)
	}

	input := pipeline.StowInput{
		StowDir:  pkgPath,
		TargetDir: targetPath,
		Packages: packages,
	}

	planResult := c.stowPipe.Execute(ctx, input)
	if !planResult.IsOk() {
		return dot.Plan{}, planResult.UnwrapErr()
	}

	return planResult.Unwrap(), nil
}
```

**Tasks**:
- [ ] Implement Manage() method
- [ ] Implement PlanManage() method
- [ ] Write tests for successful manage
- [ ] Write tests for manage with conflicts
- [ ] Write tests for dry-run mode
- [ ] Write tests for empty package list
- [ ] Write tests for non-existent package
- [ ] Test context cancellation

**Commit**: `feat(api): implement Manage and PlanManage operations`

---

#### 12.2.3: Unmanage Operations

**Implementation**:
```go
// internal/api/unmanage.go
package api

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Unmanage removes the specified packages by deleting symlinks.
func (c *client) Unmanage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanUnmanage(ctx, packages...)
	if err != nil {
		return err
	}

	if c.config.DryRun {
		c.config.Logger.Info(ctx, "dry_run_mode", "operations", len(plan.Operations))
		return nil
	}

	result := c.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}

	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}

	return nil
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (c *client) PlanUnmanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	// Load manifest to find installed packages
	targetPath, err := dot.NewTargetPath(c.config.TargetDir)
	if err != nil {
		return dot.Plan{}, fmt.Errorf("invalid target directory: %w", err)
	}

	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return dot.Plan{}, fmt.Errorf("failed to load manifest: %w", manifestResult.UnwrapErr())
	}

	manifest := manifestResult.Unwrap()

	// Create unmanage operations for specified packages
	var operations []dot.Operation
	for _, pkg := range packages {
		pkgInfo, exists := manifest.Packages[pkg]
		if !exists {
			c.config.Logger.Warn(ctx, "package_not_installed", "package", pkg)
			continue
		}

		// Create delete operations for all package links
		for _, link := range pkgInfo.Links {
			operations = append(operations, dot.LinkDelete{
				ID:     dot.NewOperationID(),
				Target: link,
			})
		}
	}

	// Build plan from operations
	return dot.NewPlan(operations, nil)
}
```

**Tasks**:
- [ ] Implement Unmanage() method
- [ ] Implement PlanUnmanage() method
- [ ] Write tests for successful unmanage
- [ ] Write tests for unmanage non-installed package
- [ ] Write tests for dry-run mode
- [ ] Test context cancellation

**Commit**: `feat(api): implement Unmanage and PlanUnmanage operations`

---

#### 12.2.4: Remanage Operations

**Implementation**:
```go
// internal/api/remanage.go
package api

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/internal/pipeline"
	"github.com/jamesainslie/dot/pkg/dot"
)

// Remanage reinstalls packages by unmanaging then managing.
// Uses incremental planning to skip unchanged packages when possible.
func (c *client) Remanage(ctx context.Context, packages ...string) error {
	plan, err := c.PlanRemanage(ctx, packages...)
	if err != nil {
		return err
	}

	if c.config.DryRun {
		c.config.Logger.Info(ctx, "dry_run_mode", "operations", len(plan.Operations))
		return nil
	}

	result := c.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}

	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}

	return nil
}

// PlanRemanage computes an incremental remanage plan.
// Only processes packages that changed since last operation.
func (c *client) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	// For now, remanage is implemented as unmanage + manage
	// Future: Add incremental planning using content hashes

	// Get current plan
	managePlan, err := c.PlanManage(ctx, packages...)
	if err != nil {
		return dot.Plan{}, err
	}

	// For remanage, we want to remove existing then reinstall
	// This is a simplified implementation
	// Full incremental planning would check hashes and skip unchanged packages

	return managePlan, nil
}
```

**Tasks**:
- [ ] Implement Remanage() method
- [ ] Implement PlanRemanage() method  
- [ ] Write tests for successful remanage
- [ ] Write tests for remanage with changes
- [ ] Write tests for dry-run mode
- [ ] Test context cancellation

**Commit**: `feat(api): implement Remanage and PlanRemanage operations`

---

#### 12.2.5: Adopt Operations

**Implementation**:
```go
// internal/api/adopt.go
package api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Adopt moves existing files from target into package then creates symlinks.
func (c *client) Adopt(ctx context.Context, files []string, pkg string) error {
	plan, err := c.PlanAdopt(ctx, files, pkg)
	if err != nil {
		return err
	}

	if c.config.DryRun {
		c.config.Logger.Info(ctx, "dry_run_mode", "operations", len(plan.Operations))
		return nil
	}

	result := c.executor.Execute(ctx, plan)
	if !result.IsOk() {
		return result.UnwrapErr()
	}

	execResult := result.Unwrap()
	if !execResult.Success() {
		return fmt.Errorf("execution failed: %d operations failed", len(execResult.Failed))
	}

	return nil
}

// PlanAdopt computes the execution plan for adopting files.
func (c *client) PlanAdopt(ctx context.Context, files []string, pkg string) (dot.Plan, error) {
	var operations []dot.Operation

	for _, file := range files {
		// Build paths
		targetFile := filepath.Join(c.config.TargetDir, file)
		
		// Determine package destination (with dot- prefix translation)
		pkgFile := translateToDotfile(file)
		pkgPath := filepath.Join(c.config.StowDir, pkg, pkgFile)

		// Create operations: move file, then create link
		operations = append(operations,
			dot.FileMove{
				ID:   dot.NewOperationID(),
				From: targetFile,
				To:   pkgPath,
			},
			dot.LinkCreate{
				ID:     dot.NewOperationID(),
				Source: pkgPath,
				Target: targetFile,
				Mode:   c.config.LinkMode,
			},
		)
	}

	return dot.NewPlan(operations, nil)
}

// translateToDotfile converts a filename to its package representation.
// Example: ".bashrc" -> "dot-bashrc"
func translateToDotfile(name string) string {
	if len(name) > 0 && name[0] == '.' {
		return "dot-" + name[1:]
	}
	return name
}
```

**Tasks**:
- [ ] Implement Adopt() method
- [ ] Implement PlanAdopt() method
- [ ] Implement translateToDotfile() helper
- [ ] Write tests for successful adoption
- [ ] Write tests for multiple files
- [ ] Write tests for dry-run mode
- [ ] Test content preservation

**Commit**: `feat(api): implement Adopt and PlanAdopt operations`

---

#### 12.2.6: Query Operations

**Implementation**:
```go
// internal/api/status.go
package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Status reports the current installation state for packages.
func (c *client) Status(ctx context.Context, packages ...string) (dot.Status, error) {
	targetPath, err := dot.NewTargetPath(c.config.TargetDir)
	if err != nil {
		return dot.Status{}, fmt.Errorf("invalid target directory: %w", err)
	}

	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return dot.Status{}, fmt.Errorf("failed to load manifest: %w", manifestResult.UnwrapErr())
	}

	manifest := manifestResult.Unwrap()

	// Filter to requested packages if specified
	var pkgInfos []dot.PackageInfo
	if len(packages) == 0 {
		// Return all packages
		for _, info := range manifest.Packages {
			pkgInfos = append(pkgInfos, info)
		}
	} else {
		// Return only specified packages
		for _, pkg := range packages {
			if info, exists := manifest.Packages[pkg]; exists {
				pkgInfos = append(pkgInfos, info)
			}
		}
	}

	return dot.Status{
		Packages: pkgInfos,
	}, nil
}

// List returns all installed packages from the manifest.
func (c *client) List(ctx context.Context) ([]dot.PackageInfo, error) {
	return c.Status(ctx)
}
```

**Tasks**:
- [ ] Implement Status() method
- [ ] Implement List() method
- [ ] Write tests for Status with packages
- [ ] Write tests for Status all packages
- [ ] Write tests for List
- [ ] Test with no manifest

**Commit**: `feat(api): implement Status and List query operations`

---

### 12.3: Status Types (Priority: High)

Since Status and PackageInfo are referenced by the Client interface, we need to define them:

**File**: `pkg/dot/status.go`

**Implementation**:
```go
package dot

import "time"

// Status represents the installation state of packages.
type Status struct {
	Packages []PackageInfo
}

// PackageInfo contains metadata about an installed package.
type PackageInfo struct {
	Name        string
	InstalledAt time.Time
	LinkCount   int
	Links       []string
}
```

**Test**:
```go
// pkg/dot/status_test.go
package dot_test

import (
	"testing"
	"time"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	status := dot.Status{
		Packages: []dot.PackageInfo{
			{
				Name:        "vim",
				InstalledAt: time.Now(),
				LinkCount:   3,
				Links:       []string{".vimrc", ".vim/"},
			},
		},
	}

	require.Len(t, status.Packages, 1)
	require.Equal(t, "vim", status.Packages[0].Name)
}
```

**Tasks**:
- [ ] Define Status struct
- [ ] Define PackageInfo struct
- [ ] Write tests for types
- [ ] Document fields

**Commit**: `feat(types): add Status and PackageInfo types`

---

### 12.4: Package Documentation (Priority: Medium)

**File**: `pkg/dot/doc.go`

**Implementation**:
```go
// Package dot provides a modern, type-safe symlink manager for dotfiles.
//
// dot is a GNU Stow replacement written in Go 1.25.1, following strict
// constitutional principles: test-driven development, atomic operations,
// functional programming, and comprehensive error handling.
//
// # Architecture
//
// The library uses an interface-based Client pattern to provide a clean
// public API while keeping internal implementation details hidden:
//
//   - Client interface in pkg/dot (stable public API)
//   - Implementation in internal/api (can evolve freely)
//   - Domain types in pkg/dot (shared between public and internal)
//
// # Basic Usage
//
// Create a client and manage packages:
//
//	cfg := dot.Config{
//		StowDir:   "/home/user/dotfiles",
//		TargetDir: "/home/user",
//		FS:        osfs.New(),
//		Logger:    slogger.New(),
//	}
//
//	client, err := dot.NewClient(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	if err := client.Manage(ctx, "vim", "zsh", "git"); err != nil {
//		log.Fatal(err)
//	}
//
// # Dry Run Mode
//
// Preview operations without applying changes:
//
//	cfg.DryRun = true
//	client, _ := dot.NewClient(cfg)
//
//	// Shows what would be done without executing
//	if err := client.Manage(ctx, "vim"); err != nil {
//		log.Fatal(err)
//	}
//
// # Query Operations
//
// Check installation status:
//
//	status, err := client.Status(ctx, "vim")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, pkg := range status.Packages {
//		fmt.Printf("%s: %d links\n", pkg.Name, pkg.LinkCount)
//	}
//
// List all installed packages:
//
//	packages, err := client.List(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Configuration
//
// The Config struct controls all dot behavior:
//
//   - StowDir: Source directory containing packages (required, absolute)
//   - TargetDir: Destination directory for symlinks (required, absolute)
//   - FS: Filesystem implementation (required)
//   - Logger: Logger implementation (required)
//   - Tracer: Distributed tracing (optional, defaults to noop)
//   - Metrics: Metrics collection (optional, defaults to noop)
//   - LinkMode: Relative or absolute symlinks (default: relative)
//   - Folding: Enable directory folding (default: true)
//   - DryRun: Preview mode (default: false)
//   - Verbosity: Logging level (default: 0)
//
// # Observability
//
// The library provides first-class observability through injected ports:
//
//   - Structured logging via Logger interface
//   - Distributed tracing via Tracer interface (OpenTelemetry compatible)
//   - Metrics collection via Metrics interface (Prometheus compatible)
//
// # Testing
//
// The library is designed for testability:
//
//   - All operations accept context.Context for cancellation
//   - Filesystem abstraction enables testing without disk I/O
//   - Pure functional core enables property-based testing
//   - Interface-based Client enables mocking
//
// For examples, see the examples_test.go file and the examples/ directory.
package dot
```

**Tasks**:
- [ ] Write comprehensive package documentation
- [ ] Add usage examples
- [ ] Document architecture pattern
- [ ] Document configuration options
- [ ] Add links to detailed docs

**Commit**: `docs(api): add comprehensive package documentation`

---

### 12.5: Example Tests (Priority: Medium)

**File**: `pkg/dot/examples_test.go`

**Implementation**:
```go
// pkg/dot/examples_test.go
package dot_test

import (
	"context"
	"fmt"
	"log"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
)

func ExampleNewClient() {
	cfg := dot.Config{
		StowDir:   "/home/user/dotfiles",
		TargetDir: "/home/user",
		FS:        adapters.NewOSFilesystem(),
		Logger:    adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client created: %v\n", client != nil)
	// Output: Client created: true
}

func ExampleClient_Manage() {
	cfg := dot.Config{
		StowDir:   "/home/user/dotfiles",
		TargetDir: "/home/user",
		FS:        adapters.NewOSFilesystem(),
		Logger:    adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := client.Manage(ctx, "vim", "zsh"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Packages installed successfully")
	// Output: Packages installed successfully
}

func ExampleClient_PlanManage() {
	cfg := dot.Config{
		StowDir:   "/home/user/dotfiles",
		TargetDir: "/home/user",
		FS:        adapters.NewOSFilesystem(),
		Logger:    adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	plan, err := client.PlanManage(ctx, "vim")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Plan contains %d operations\n", len(plan.Operations))
	// Output: Plan contains 3 operations
}

func ExampleConfig_validation() {
	cfg := dot.Config{
		StowDir:   "/home/user/dotfiles",
		TargetDir: "/home/user",
		FS:        adapters.NewOSFilesystem(),
		Logger:    adapters.NewNoopLogger(),
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Configuration is valid")
	// Output: Configuration is valid
}
```

**Tasks**:
- [ ] Write ExampleNewClient
- [ ] Write ExampleClient_Manage
- [ ] Write ExampleClient_PlanManage
- [ ] Write ExampleClient_Status
- [ ] Write ExampleConfig_validation
- [ ] Verify examples appear in godoc

**Commit**: `docs(api): add example tests for godoc`

---

## Quality Gates

### Definition of Done

Each task is complete when:
- [ ] Implementation follows test-first approach
- [ ] All tests pass
- [ ] Test coverage ≥ 80% for new code
- [ ] All linters pass (golangci-lint)
- [ ] No changes to Phases 1-11 code
- [ ] Documentation updated
- [ ] Atomic commit created

### Phase Completion Criteria

Phase 12 is complete when:
- [ ] Client interface fully defined in pkg/dot
- [ ] Client implementation complete in internal/api
- [ ] All operations functional (Manage, Unmanage, Remanage, Adopt)
- [ ] Query operations working (Status, List)
- [ ] Registration mechanism working
- [ ] Documentation comprehensive
- [ ] Examples demonstrate usage
- [ ] Test coverage ≥ 80%
- [ ] All linters pass
- [ ] No import cycles
- [ ] API suitable for library embedding

---

## Development Workflow

### For Each Task

1. **Write Test**: Create failing test in appropriate package
2. **Run Test**: Verify test fails (red)
3. **Implement**: Write minimum code to pass (green)
4. **Refactor**: Improve while maintaining green
5. **Lint**: Run `make check`
6. **Commit**: Create atomic commit

### Testing Strategy

```bash
# Test specific package
go test ./pkg/dot -v
go test ./internal/api -v

# Test with race detector
go test ./pkg/dot -race
go test ./internal/api -race

# Run all tests
make test

# Check coverage
go test ./pkg/dot -coverprofile=coverage.out
go tool cover -func=coverage.out
```

---

## Import Cycle Resolution Summary

### The Pattern

```
pkg/dot/
  client.go          ← Interface definition (no imports of internal/*)
  config.go          ← Config struct
  <domain types>     ← Operation, Plan, Result, etc.

internal/api/
  client.go          ← Implementation (imports pkg/dot + internal/*)
  manage.go          ← Can import internal/pipeline
  status.go          ← Can import internal/manifest

internal/executor/   ← Imports pkg/dot for domain types (NO CYCLE)
internal/pipeline/   ← Imports pkg/dot for domain types (NO CYCLE)
```

### Key Insight

**Interface definitions don't create dependencies on implementations.**

- `pkg/dot/client.go` defines `type Client interface { ... }`
- `internal/api/client.go` implements it
- `pkg/dot` never imports `internal/api`
- `internal/api` imports both `pkg/dot` (for interface) and `internal/*` (for implementation)
- **Result**: No cycle!

### Registration Mechanism

```go
// pkg/dot/client.go
var newClientImpl func(Config) (Client, error)

func NewClient(cfg Config) (Client, error) {
    return newClientImpl(cfg)
}

// internal/api/client.go
func init() {
    dot.RegisterClientImpl(newClient)
}
```

This allows the constructor to live in `pkg/dot` while implementation is in `internal/api`.

---

## Success Metrics

- [ ] Client interface accessible via `import "pkg/dot"`
- [ ] All management operations work (Manage, Unmanage, Remanage, Adopt)
- [ ] Query operations work (Status, List)
- [ ] Configuration system flexible and type-safe
- [ ] Documentation comprehensive
- [ ] Examples demonstrate common use cases
- [ ] Test coverage ≥ 80%
- [ ] No linter warnings
- [ ] No changes to Phases 1-11
- [ ] Library embeddable in other tools
- [ ] No import cycles

---

## Timeline Estimate

**Total Effort**: 6-8 hours (reduced from original 8-12 due to simpler initial implementation)

- 12.1 Client Interface: 1 hour
- 12.2 Client Implementation: 3-4 hours
- 12.3 Status Types: 30 minutes
- 12.4 Documentation: 1-1.5 hours
- 12.5 Examples: 1 hour

**Note**: This implements core functionality. Advanced features (streaming API, etc.) deferred to future phases or Phase 12b refactoring.

---

## What's Deferred

These features from the original Phase 12 plan are deferred pending Phase 12b refactoring:

- **Streaming API**: Requires more complex pipeline integration
- **ConfigBuilder**: Can be added incrementally
- **Doctor command**: Requires additional pipeline support
- **Advanced query operations**: Build on basic Status/List

These can be added incrementally once the basic Client interface is working.

---

## Next Steps After Phase 12

1. **Verify Integration**: Test Client with actual use cases
2. **Document Patterns**: Add examples of embedding in other tools
3. **Phase 13**: CLI Layer can now use Client interface
4. **Phase 12b** (optional): Refactor to Option 1 for cleaner architecture

---

## References

- [Implementation Plan](./Implementation-Plan.md)
- [Architecture Documentation](./Architecture.md)
- [Features Specification](./Features.md)
- [Phase 12b Refactoring Plan](./Phase-12b-Refactor-Plan.md) (for future work)
- [Go Interfaces Design](https://go.dev/doc/effective_go#interfaces)
