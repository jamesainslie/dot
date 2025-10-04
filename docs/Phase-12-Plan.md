# Phase 12: Public Library API - Implementation Plan

## Overview

Phase 12 delivers a clean, embeddable Go API for the dot library. This phase creates the public interface in `pkg/dot/` that wraps the internal functional core, providing a simple facade for library consumers.

**Dependencies**: Phases 1-11 must be complete (domain model, ports, adapters, scanner, planner, resolver, sorter, pipeline orchestration, executor, manifest management).

**Deliverable**: Clean, tested public library API suitable for embedding in other tools.

## Design Principles

- **Library First**: Core has zero CLI dependencies, fully embeddable
- **Type Safety**: Leverage domain types with compile-time guarantees
- **Context-Aware**: All operations accept context.Context for cancellation
- **Immutability**: Configuration objects are immutable
- **Explicit Dependencies**: Dependency injection for all infrastructure
- **No Global State**: Thread-safe by design
- **Streaming Support**: Memory-efficient APIs for large operations

## Package Structure

```
pkg/dot/
├── client.go          # Client facade with main API
├── client_test.go     # Client tests
├── config.go          # Configuration types and validation
├── config_test.go     # Configuration tests
├── streaming.go       # Streaming API and operators
├── streaming_test.go  # Streaming tests
├── types.go           # Re-exported domain types
├── types_test.go      # Type export tests
├── doc.go             # Package documentation
└── examples_test.go   # Example tests for godoc

examples/
├── basic/
│   └── main.go        # Basic usage example
├── streaming/
│   └── main.go        # Streaming API example
├── custom-fs/
│   └── main.go        # Custom filesystem example
└── embedded/
    └── main.go        # Embedding in another tool
```

---

## Task Breakdown

### 12.1: Client Facade (Priority: High)

#### 12.1.1: Core Client Structure

**File**: `pkg/dot/client.go`

**Test-First Approach**:
```go
// pkg/dot/client_test.go
func TestNew_ValidConfig(t *testing.T) {
    cfg := Config{
        StowDir:   "/test/stow",
        TargetDir: "/test/target",
        FS:        memfs.New(),
        Logger:    noop.NewLogger(),
    }
    
    client, err := New(cfg)
    require.NoError(t, err)
    require.NotNil(t, client)
}

func TestNew_InvalidConfig(t *testing.T) {
    cfg := Config{
        StowDir: "relative/path", // Invalid
    }
    
    client, err := New(cfg)
    require.Error(t, err)
    require.Nil(t, client)
}
```

**Implementation**:
```go
// pkg/dot/client.go
package dot

import (
    "context"
    "fmt"
    
    "github.com/yourorg/dot/internal/executor"
    "github.com/yourorg/dot/internal/pipeline"
    "github.com/yourorg/dot/pkg/dot/ports"
)

// Client provides the main API for dot operations.
type Client struct {
    config   Config
    pipeline *pipeline.Engine
    executor *executor.Executor
}

// New creates a new Client with the given configuration.
// Returns an error if configuration is invalid.
func New(cfg Config) (*Client, error) {
    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Apply defaults
    cfg = cfg.withDefaults()
    
    // Build pipeline engine
    pipe := pipeline.New(pipeline.Opts{
        FS:       cfg.FS,
        Logger:   cfg.Logger,
        Tracer:   cfg.Tracer,
        Metrics:  cfg.Metrics,
        LinkMode: cfg.LinkMode,
        Folding:  cfg.Folding,
        Ignore:   cfg.Ignore,
    })
    
    // Build executor
    exec := executor.New(executor.Opts{
        FS:     cfg.FS,
        Logger: cfg.Logger,
        Tracer: cfg.Tracer,
    })
    
    return &Client{
        config:   cfg,
        pipeline: pipe,
        executor: exec,
    }, nil
}
```

**Tasks**:
- [ ] Write Client struct definition
- [ ] Write New() constructor with validation
- [ ] Add Config struct (see 12.2)
- [ ] Write tests for valid configuration
- [ ] Write tests for invalid configuration
- [ ] Test default value application

**Commit**: `feat(api): add Client facade with constructor`

---

#### 12.1.2: Stow Operations

**Test-First Approach**:
```go
func TestClient_Stow(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    err = client.Stow(ctx, "package1")
    require.NoError(t, err)
    
    // Verify links created
    assertLinkExists(t, fs, "/test/target/.bashrc")
}

func TestClient_PlanStow(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    plan, err := client.PlanStow(ctx, "package1")
    require.NoError(t, err)
    require.NotEmpty(t, plan.Operations())
}

func TestClient_Stow_DryRun(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    cfg.DryRun = true
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    err = client.Stow(ctx, "package1")
    require.NoError(t, err)
    
    // Verify no changes made
    assertLinkNotExists(t, fs, "/test/target/.bashrc")
}
```

**Implementation**:
```go
// Stow installs the specified packages by creating symlinks.
// If DryRun is enabled, returns the plan without executing.
func (c *Client) Stow(ctx context.Context, packages ...string) error {
    plan, err := c.PlanStow(ctx, packages...)
    if err != nil {
        return err
    }
    
    if c.config.DryRun {
        return c.renderPlan(plan)
    }
    
    result := c.executor.Execute(ctx, plan)
    _, err = result.Unwrap()
    return err
}

// PlanStow computes the execution plan for stowing packages
// without applying changes.
func (c *Client) PlanStow(ctx context.Context, packages ...string) (Plan, error) {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    result := c.pipeline.Stow(ctx, input)
    return result.Unwrap()
}
```

**Tasks**:
- [ ] Implement Stow() method
- [ ] Implement PlanStow() method
- [ ] Write tests for successful stow
- [ ] Write tests for stow with conflicts
- [ ] Write tests for dry-run mode
- [ ] Write tests for empty package list
- [ ] Write tests for non-existent package
- [ ] Test context cancellation

**Commit**: `feat(api): implement Stow and PlanStow methods`

---

#### 12.1.3: Unstow Operations

**Test-First Approach**:
```go
func TestClient_Unstow(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Stow first
    err = client.Stow(ctx, "package1")
    require.NoError(t, err)
    
    // Then unstow
    err = client.Unstow(ctx, "package1")
    require.NoError(t, err)
    
    // Verify links removed
    assertLinkNotExists(t, fs, "/test/target/.bashrc")
}

func TestClient_PlanUnstow(t *testing.T) {
    fs := memfs.New()
    setupInstalledPackage(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    plan, err := client.PlanUnstow(ctx, "package1")
    require.NoError(t, err)
    require.NotEmpty(t, plan.Operations())
}
```

**Implementation**:
```go
// Unstow removes the specified packages by deleting symlinks.
func (c *Client) Unstow(ctx context.Context, packages ...string) error {
    plan, err := c.PlanUnstow(ctx, packages...)
    if err != nil {
        return err
    }
    
    if c.config.DryRun {
        return c.renderPlan(plan)
    }
    
    result := c.executor.Execute(ctx, plan)
    _, err = result.Unwrap()
    return err
}

// PlanUnstow computes the execution plan for unstowing packages.
func (c *Client) PlanUnstow(ctx context.Context, packages ...string) (Plan, error) {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    result := c.pipeline.Unstow(ctx, input)
    return result.Unwrap()
}
```

**Tasks**:
- [ ] Implement Unstow() method
- [ ] Implement PlanUnstow() method
- [ ] Write tests for successful unstow
- [ ] Write tests for unstow of non-installed package
- [ ] Write tests for dry-run mode
- [ ] Write tests for partial unstow
- [ ] Test context cancellation

**Commit**: `feat(api): implement Unstow and PlanUnstow methods`

---

#### 12.1.4: Restow Operations

**Test-First Approach**:
```go
func TestClient_Restow(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Initial stow
    err = client.Stow(ctx, "package1")
    require.NoError(t, err)
    
    // Modify package
    modifyPackage(fs, "package1")
    
    // Restow
    err = client.Restow(ctx, "package1")
    require.NoError(t, err)
    
    // Verify updated links
    assertLinkPointsTo(t, fs, "/test/target/.bashrc", "/test/stow/package1/dot-bashrc")
}

func TestClient_Restow_Incremental(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1", "package2")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Initial stow
    err = client.Stow(ctx, "package1", "package2")
    require.NoError(t, err)
    
    // Modify only package1
    modifyPackage(fs, "package1")
    
    // Restow - should only process package1
    err = client.Restow(ctx, "package1", "package2")
    require.NoError(t, err)
    
    // Verify incremental behavior via logs/metrics
}
```

**Implementation**:
```go
// Restow reinstalls packages by unstowing then stowing.
// Uses incremental planning to skip unchanged packages.
func (c *Client) Restow(ctx context.Context, packages ...string) error {
    plan, err := c.PlanRestow(ctx, packages...)
    if err != nil {
        return err
    }
    
    if c.config.DryRun {
        return c.renderPlan(plan)
    }
    
    result := c.executor.Execute(ctx, plan)
    _, err = result.Unwrap()
    return err
}

// PlanRestow computes an incremental restow plan.
// Only processes packages that changed since last operation.
func (c *Client) PlanRestow(ctx context.Context, packages ...string) (Plan, error) {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    result := c.pipeline.Restow(ctx, input)
    return result.Unwrap()
}
```

**Tasks**:
- [ ] Implement Restow() method
- [ ] Implement PlanRestow() method
- [ ] Write tests for successful restow
- [ ] Write tests for incremental restow
- [ ] Write tests for restow with no changes
- [ ] Write tests for dry-run mode
- [ ] Test context cancellation

**Commit**: `feat(api): implement Restow and PlanRestow methods`

---

#### 12.1.5: Adopt Operations

**Test-First Approach**:
```go
func TestClient_Adopt(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    // Create existing file in target
    writeFile(t, fs, "/test/target/.bashrc", "content")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    err = client.Adopt(ctx, []string{".bashrc"}, "package1")
    require.NoError(t, err)
    
    // Verify file moved and link created
    assertFileExists(t, fs, "/test/stow/package1/dot-bashrc")
    assertLinkExists(t, fs, "/test/target/.bashrc")
    assertFileContent(t, fs, "/test/stow/package1/dot-bashrc", "content")
}

func TestClient_PlanAdopt(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    writeFile(t, fs, "/test/target/.bashrc", "content")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    plan, err := client.PlanAdopt(ctx, []string{".bashrc"}, "package1")
    require.NoError(t, err)
    require.NotEmpty(t, plan.Operations())
}
```

**Implementation**:
```go
// Adopt moves existing files from target into package then creates symlinks.
func (c *Client) Adopt(ctx context.Context, files []string, pkg string) error {
    plan, err := c.PlanAdopt(ctx, files, pkg)
    if err != nil {
        return err
    }
    
    if c.config.DryRun {
        return c.renderPlan(plan)
    }
    
    result := c.executor.Execute(ctx, plan)
    _, err = result.Unwrap()
    return err
}

// PlanAdopt computes the execution plan for adopting files.
func (c *Client) PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error) {
    input := adopt.AdoptInput{
        Package:   pkg,
        Files:     files,
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
    }
    
    result := c.pipeline.Adopt(ctx, input)
    return result.Unwrap()
}
```

**Tasks**:
- [ ] Implement Adopt() method with files-first signature
- [ ] Implement PlanAdopt() method with files-first signature
- [ ] Write tests for successful adoption
- [ ] Write tests for adopting non-existent file
- [ ] Write tests for adopting multiple files
- [ ] Write tests for dry-run mode
- [ ] Test content preservation
- [ ] Test context cancellation

**Commit**: `feat(api): implement Adopt and PlanAdopt methods`

---

#### 12.1.6: Query Methods

**Test-First Approach**:
```go
func TestClient_Status(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Stow package
    err = client.Stow(ctx, "package1")
    require.NoError(t, err)
    
    // Get status
    status, err := client.Status(ctx, "package1")
    require.NoError(t, err)
    require.Len(t, status.Packages, 1)
    require.Equal(t, "package1", status.Packages[0].Name)
}

func TestClient_Doctor(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Create broken link
    createBrokenLink(fs, "/test/target/.vimrc")
    
    // Run doctor
    report, err := client.Doctor(ctx)
    require.NoError(t, err)
    require.NotEmpty(t, report.BrokenLinks)
}

func TestClient_List(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1", "package2")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Stow packages
    err = client.Stow(ctx, "package1", "package2")
    require.NoError(t, err)
    
    // List installed
    packages, err := client.List(ctx)
    require.NoError(t, err)
    require.Len(t, packages, 2)
}
```

**Implementation**:
```go
// Status reports the current installation state for packages.
func (c *Client) Status(ctx context.Context, packages ...string) (Status, error) {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    result := c.pipeline.Status(ctx, input)
    return result.Unwrap()
}

// Doctor validates installation consistency and detects issues.
func (c *Client) Doctor(ctx context.Context) (DiagnosticReport, error) {
    input := doctor.DoctorInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
    }
    
    result := c.pipeline.Doctor(ctx, input)
    return result.Unwrap()
}

// List returns all installed packages from the manifest.
func (c *Client) List(ctx context.Context) ([]PackageInfo, error) {
    manifest, err := c.loadManifest(ctx)
    if err != nil {
        return nil, err
    }
    
    packages := make([]PackageInfo, 0, len(manifest.Packages))
    for _, pkg := range manifest.Packages {
        packages = append(packages, pkg)
    }
    
    return packages, nil
}
```

**Tasks**:
- [ ] Implement Status() method
- [ ] Implement Doctor() method
- [ ] Implement List() method
- [ ] Write tests for Status with installed packages
- [ ] Write tests for Status with no packages
- [ ] Write tests for Doctor with issues
- [ ] Write tests for Doctor with healthy state
- [ ] Write tests for List
- [ ] Test context cancellation

**Commit**: `feat(api): implement Status, Doctor, and List query methods`

---

### 12.2: Configuration (Priority: High)

#### 12.2.1: Config Structure

**Test-First Approach**:
```go
// pkg/dot/config_test.go
func TestConfig_Validate_Valid(t *testing.T) {
    cfg := Config{
        StowDir:   "/test/stow",
        TargetDir: "/test/target",
        FS:        memfs.New(),
        Logger:    noop.NewLogger(),
    }
    
    err := cfg.Validate()
    require.NoError(t, err)
}

func TestConfig_Validate_RelativeStowDir(t *testing.T) {
    cfg := Config{
        StowDir:   "relative/path",
        TargetDir: "/test/target",
    }
    
    err := cfg.Validate()
    require.Error(t, err)
    require.Contains(t, err.Error(), "stowDir")
}

func TestConfig_Validate_RelativeTargetDir(t *testing.T) {
    cfg := Config{
        StowDir:   "/test/stow",
        TargetDir: "relative/path",
    }
    
    err := cfg.Validate()
    require.Error(t, err)
    require.Contains(t, err.Error(), "targetDir")
}
```

**Implementation**:
```go
// pkg/dot/config.go
package dot

import (
    "fmt"
    "path/filepath"
    
    "github.com/yourorg/dot/internal/ignore"
    "github.com/yourorg/dot/pkg/dot/ports"
)

// Config holds configuration for the dot Client.
type Config struct {
    // StowDir is the source directory containing packages.
    // Must be an absolute path.
    StowDir string
    
    // TargetDir is the destination directory for symlinks.
    // Must be an absolute path.
    TargetDir string
    
    // LinkMode specifies whether to create relative or absolute symlinks.
    LinkMode LinkMode
    
    // Folding enables directory-level linking when all contents
    // belong to a single package.
    Folding bool
    
    // DryRun enables preview mode without applying changes.
    DryRun bool
    
    // Verbosity controls logging detail (0=quiet, 1=info, 2=debug, 3=trace).
    Verbosity int
    
    // Ignore contains patterns for excluding files from operations.
    Ignore ignore.IgnoreSet
    
    // BackupDir specifies where to store backup files.
    // If empty, backups go to <TargetDir>/.dot-backup/
    BackupDir string
    
    // Concurrency limits parallel operation execution.
    // If zero, defaults to runtime.NumCPU().
    Concurrency int
    
    // Infrastructure dependencies (required)
    FS      ports.FS
    Logger  ports.Logger
    Tracer  ports.Tracer
    Metrics ports.Metrics
}

// LinkMode specifies symlink creation strategy.
type LinkMode int

const (
    // LinkRelative creates relative symlinks (default).
    LinkRelative LinkMode = iota
    // LinkAbsolute creates absolute symlinks.
    LinkAbsolute
)

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
    if c.StowDir == "" {
        return fmt.Errorf("stowDir is required")
    }
    if !filepath.IsAbs(c.StowDir) {
        return fmt.Errorf("stowDir must be absolute path: %s", c.StowDir)
    }
    
    if c.TargetDir == "" {
        return fmt.Errorf("targetDir is required")
    }
    if !filepath.IsAbs(c.TargetDir) {
        return fmt.Errorf("targetDir must be absolute path: %s", c.TargetDir)
    }
    
    if c.FS == nil {
        return fmt.Errorf("FS is required")
    }
    
    if c.Logger == nil {
        return fmt.Errorf("Logger is required")
    }
    
    if c.Verbosity < 0 {
        return fmt.Errorf("verbosity cannot be negative")
    }
    
    if c.Concurrency < 0 {
        return fmt.Errorf("concurrency cannot be negative")
    }
    
    return nil
}

// withDefaults returns a copy of the config with defaults applied.
func (c Config) withDefaults() Config {
    cfg := c
    
    if cfg.Tracer == nil {
        cfg.Tracer = noop.NewTracer()
    }
    
    if cfg.Metrics == nil {
        cfg.Metrics = noop.NewMetrics()
    }
    
    if cfg.BackupDir == "" {
        cfg.BackupDir = filepath.Join(cfg.TargetDir, ".dot-backup")
    }
    
    if cfg.Concurrency == 0 {
        cfg.Concurrency = runtime.NumCPU()
    }
    
    // Folding enabled by default
    // (caller can explicitly set to false)
    
    return cfg
}
```

**Tasks**:
- [ ] Define Config struct with all fields
- [ ] Define LinkMode enum
- [ ] Implement Validate() method
- [ ] Implement withDefaults() method
- [ ] Write tests for valid configuration
- [ ] Write tests for each validation error
- [ ] Write tests for default application
- [ ] Document all configuration fields

**Commit**: `feat(config): define Config struct with validation`

---

#### 12.2.2: Configuration Builder

**Test-First Approach**:
```go
func TestConfigBuilder(t *testing.T) {
    cfg := NewConfig().
        WithStowDir("/test/stow").
        WithTargetDir("/test/target").
        WithFS(memfs.New()).
        WithLogger(noop.NewLogger()).
        WithLinkMode(LinkAbsolute).
        WithFolding(false).
        Build()
    
    require.NoError(t, cfg.Validate())
    require.Equal(t, "/test/stow", cfg.StowDir)
    require.Equal(t, "/test/target", cfg.TargetDir)
    require.Equal(t, LinkAbsolute, cfg.LinkMode)
    require.False(t, cfg.Folding)
}

func TestConfigBuilder_Defaults(t *testing.T) {
    cfg := NewConfig().
        WithStowDir("/test/stow").
        WithTargetDir("/test/target").
        WithFS(memfs.New()).
        WithLogger(noop.NewLogger()).
        Build()
    
    require.NoError(t, cfg.Validate())
    require.Equal(t, LinkRelative, cfg.LinkMode)
    require.True(t, cfg.Folding)
    require.NotNil(t, cfg.Tracer)
    require.NotNil(t, cfg.Metrics)
}
```

**Implementation**:
```go
// ConfigBuilder provides a fluent interface for building Config.
type ConfigBuilder struct {
    config Config
}

// NewConfig creates a new ConfigBuilder with defaults.
func NewConfig() *ConfigBuilder {
    return &ConfigBuilder{
        config: Config{
            LinkMode:    LinkRelative,
            Folding:     true,
            Verbosity:   0,
            Concurrency: 0, // Will default to NumCPU
        },
    }
}

// WithStowDir sets the stow directory.
func (b *ConfigBuilder) WithStowDir(dir string) *ConfigBuilder {
    b.config.StowDir = dir
    return b
}

// WithTargetDir sets the target directory.
func (b *ConfigBuilder) WithTargetDir(dir string) *ConfigBuilder {
    b.config.TargetDir = dir
    return b
}

// WithFS sets the filesystem implementation.
func (b *ConfigBuilder) WithFS(fs ports.FS) *ConfigBuilder {
    b.config.FS = fs
    return b
}

// WithLogger sets the logger implementation.
func (b *ConfigBuilder) WithLogger(logger ports.Logger) *ConfigBuilder {
    b.config.Logger = logger
    return b
}

// WithTracer sets the tracer implementation.
func (b *ConfigBuilder) WithTracer(tracer ports.Tracer) *ConfigBuilder {
    b.config.Tracer = tracer
    return b
}

// WithMetrics sets the metrics implementation.
func (b *ConfigBuilder) WithMetrics(metrics ports.Metrics) *ConfigBuilder {
    b.config.Metrics = metrics
    return b
}

// WithLinkMode sets the symlink creation mode.
func (b *ConfigBuilder) WithLinkMode(mode LinkMode) *ConfigBuilder {
    b.config.LinkMode = mode
    return b
}

// WithFolding enables or disables directory folding.
func (b *ConfigBuilder) WithFolding(enabled bool) *ConfigBuilder {
    b.config.Folding = enabled
    return b
}

// WithDryRun enables or disables dry-run mode.
func (b *ConfigBuilder) WithDryRun(enabled bool) *ConfigBuilder {
    b.config.DryRun = enabled
    return b
}

// WithVerbosity sets the logging verbosity level.
func (b *ConfigBuilder) WithVerbosity(level int) *ConfigBuilder {
    b.config.Verbosity = level
    return b
}

// WithIgnore sets the ignore pattern set.
func (b *ConfigBuilder) WithIgnore(ignore ignore.IgnoreSet) *ConfigBuilder {
    b.config.Ignore = ignore
    return b
}

// WithBackupDir sets the backup directory.
func (b *ConfigBuilder) WithBackupDir(dir string) *ConfigBuilder {
    b.config.BackupDir = dir
    return b
}

// WithConcurrency sets the concurrency limit.
func (b *ConfigBuilder) WithConcurrency(limit int) *ConfigBuilder {
    b.config.Concurrency = limit
    return b
}

// Build returns the built Config.
func (b *ConfigBuilder) Build() Config {
    return b.config
}
```

**Tasks**:
- [ ] Implement ConfigBuilder struct
- [ ] Implement NewConfig() constructor
- [ ] Implement all With*() methods
- [ ] Implement Build() method
- [ ] Write tests for builder pattern
- [ ] Write tests for method chaining
- [ ] Write tests for default values
- [ ] Document builder usage

**Commit**: `feat(config): add ConfigBuilder for fluent API`

---

### 12.3: Streaming API (Priority: Medium)

#### 12.3.1: Streaming Operations

**Test-First Approach**:
```go
// pkg/dot/streaming_test.go
func TestClient_StowStream(t *testing.T) {
    fs := memfs.New()
    setupStowFixtures(fs, "package1")
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx := context.Background()
    stream := client.StowStream(ctx, "package1")
    
    var ops []Operation
    for result := range stream {
        require.True(t, result.IsOk())
        ops = append(ops, result.Value())
    }
    
    require.NotEmpty(t, ops)
}

func TestClient_StowStream_Cancellation(t *testing.T) {
    fs := memfs.New()
    setupLargeFixtures(fs, "package1") // Many files
    
    cfg := testConfig(fs)
    client, err := New(cfg)
    require.NoError(t, err)
    
    ctx, cancel := context.WithCancel(context.Background())
    stream := client.StowStream(ctx, "package1")
    
    // Read a few operations then cancel
    count := 0
    for result := range stream {
        count++
        if count >= 5 {
            cancel()
            break
        }
    }
    
    // Drain remaining (should complete quickly)
    for range stream {
    }
    
    require.GreaterOrEqual(t, count, 5)
}
```

**Implementation**:
```go
// pkg/dot/streaming.go
package dot

import (
    "context"
    
    "github.com/yourorg/dot/internal/domain"
    "github.com/yourorg/dot/internal/scanner"
)

// StowStream returns a stream of operations for stowing packages.
// Operations are emitted as they are computed, enabling memory-efficient
// processing of large package sets.
func (c *Client) StowStream(ctx context.Context, packages ...string) <-chan Result[Operation] {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    return c.pipeline.StowStream(ctx, input)
}

// UnstowStream returns a stream of operations for unstowing packages.
func (c *Client) UnstowStream(ctx context.Context, packages ...string) <-chan Result[Operation] {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    return c.pipeline.UnstowStream(ctx, input)
}

// RestowStream returns a stream of operations for restowing packages.
func (c *Client) RestowStream(ctx context.Context, packages ...string) <-chan Result[Operation] {
    input := scanner.ScanInput{
        StowDir:   c.mustParsePath(c.config.StowDir),
        TargetDir: c.mustParsePath(c.config.TargetDir),
        Packages:  packages,
        Ignore:    c.config.Ignore,
    }
    
    return c.pipeline.RestowStream(ctx, input)
}
```

**Tasks**:
- [ ] Implement StowStream() method
- [ ] Implement UnstowStream() method
- [ ] Implement RestowStream() method
- [ ] Write tests for successful streaming
- [ ] Write tests for stream cancellation
- [ ] Write tests for error propagation
- [ ] Write tests for backpressure
- [ ] Document memory characteristics

**Commit**: `feat(streaming): add streaming operation methods`

---

#### 12.3.2: Stream Operators

**Test-First Approach**:
```go
func TestStreamMap(t *testing.T) {
    ctx := context.Background()
    
    input := make(chan Result[int])
    go func() {
        defer close(input)
        input <- Ok(1)
        input <- Ok(2)
        input <- Ok(3)
    }()
    
    output := StreamMap(ctx, input, func(x int) int {
        return x * 2
    })
    
    var results []int
    for result := range output {
        require.True(t, result.IsOk())
        results = append(results, result.Value())
    }
    
    require.Equal(t, []int{2, 4, 6}, results)
}

func TestStreamFilter(t *testing.T) {
    ctx := context.Background()
    
    input := make(chan Result[int])
    go func() {
        defer close(input)
        input <- Ok(1)
        input <- Ok(2)
        input <- Ok(3)
        input <- Ok(4)
    }()
    
    output := StreamFilter(ctx, input, func(x int) bool {
        return x%2 == 0
    })
    
    var results []int
    for result := range output {
        require.True(t, result.IsOk())
        results = append(results, result.Value())
    }
    
    require.Equal(t, []int{2, 4}, results)
}

func TestCollectStream(t *testing.T) {
    ctx := context.Background()
    
    input := make(chan Result[int])
    go func() {
        defer close(input)
        input <- Ok(1)
        input <- Ok(2)
        input <- Ok(3)
    }()
    
    result := CollectStream(ctx, input)
    require.True(t, result.IsOk())
    require.Equal(t, []int{1, 2, 3}, result.Value())
}

func TestCollectStream_WithError(t *testing.T) {
    ctx := context.Background()
    
    input := make(chan Result[int])
    go func() {
        defer close(input)
        input <- Ok(1)
        input <- Err[int](errors.New("test error"))
        input <- Ok(3)
    }()
    
    result := CollectStream(ctx, input)
    require.False(t, result.IsOk())
}
```

**Implementation**:
```go
// StreamMap transforms values in a stream using the provided function.
func StreamMap[A, B any](ctx context.Context, stream <-chan Result[A], f func(A) B) <-chan Result[B] {
    out := make(chan Result[B])
    
    go func() {
        defer close(out)
        
        for result := range stream {
            select {
            case out <- Map(result, f):
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}

// StreamFilter filters values in a stream using the provided predicate.
func StreamFilter[T any](ctx context.Context, stream <-chan Result[T], pred func(T) bool) <-chan Result[T] {
    out := make(chan Result[T])
    
    go func() {
        defer close(out)
        
        for result := range stream {
            if result.IsOk() && !pred(result.Value()) {
                continue
            }
            
            select {
            case out <- result:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}

// CollectStream collects all values from a stream into a slice.
// Returns an error if any stream value is an error.
func CollectStream[T any](ctx context.Context, stream <-chan Result[T]) Result[[]T] {
    var values []T
    var errs []error
    
    for result := range stream {
        select {
        case <-ctx.Done():
            return Err[[]T](ctx.Err())
        default:
            if result.IsOk() {
                values = append(values, result.Value())
            } else {
                errs = append(errs, result.Error())
            }
        }
    }
    
    if len(errs) > 0 {
        return Err[[]T](domain.ErrMultiple{Errors: errs})
    }
    
    return Ok(values)
}

// StreamTake takes the first n values from a stream.
func StreamTake[T any](ctx context.Context, stream <-chan Result[T], n int) <-chan Result[T] {
    out := make(chan Result[T])
    
    go func() {
        defer close(out)
        
        count := 0
        for result := range stream {
            if count >= n {
                return
            }
            
            select {
            case out <- result:
                count++
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}

// StreamReduce reduces a stream to a single value using the provided function.
func StreamReduce[T, A any](ctx context.Context, stream <-chan Result[T], init A, f func(A, T) A) Result[A] {
    acc := init
    
    for result := range stream {
        select {
        case <-ctx.Done():
            return Err[A](ctx.Err())
        default:
            if !result.IsOk() {
                return Err[A](result.Error())
            }
            acc = f(acc, result.Value())
        }
    }
    
    return Ok(acc)
}
```

**Tasks**:
- [ ] Implement StreamMap() operator
- [ ] Implement StreamFilter() operator
- [ ] Implement CollectStream() collector
- [ ] Implement StreamTake() operator
- [ ] Implement StreamReduce() operator
- [ ] Write tests for each operator
- [ ] Write tests for context cancellation
- [ ] Write tests for error propagation
- [ ] Document operator semantics

**Commit**: `feat(streaming): add stream operator functions`

---

### 12.4: Type Exports (Priority: Medium)

#### 12.4.1: Domain Type Re-exports

**Test-First Approach**:
```go
// pkg/dot/types_test.go
func TestTypeExports(t *testing.T) {
    // Verify all required types are exported
    var _ Operation
    var _ Plan
    var _ Package
    var _ LinkCreate
    var _ LinkDelete
    var _ DirCreate
    var _ DirDelete
    var _ FileMove
    var _ FileBackup
    var _ Result[string]
    var _ Status
    var _ DiagnosticReport
    var _ PackageInfo
    var _ Conflict
    var _ Warning
}

func TestOperation_Interface(t *testing.T) {
    // Verify Operation interface is usable
    var op Operation = &LinkCreate{
        ID:     "test-op",
        Source: "/test/source",
        Target: "/test/target",
    }
    
    require.Equal(t, OperationKindLinkCreate, op.Kind())
    require.NoError(t, op.Validate())
}
```

**Implementation**:
```go
// pkg/dot/types.go
package dot

import (
    "github.com/yourorg/dot/internal/domain"
    "github.com/yourorg/dot/internal/planner"
    "github.com/yourorg/dot/internal/resolver"
)

// Operation represents a single operation in an execution plan.
type Operation = domain.Operation

// Operation types
type (
    LinkCreate = domain.LinkCreate
    LinkDelete = domain.LinkDelete
    DirCreate  = domain.DirCreate
    DirDelete  = domain.DirDelete
    FileMove   = domain.FileMove
    FileBackup = domain.FileBackup
)

// OperationKind identifies the type of operation.
type OperationKind = domain.OperationKind

// Operation kinds
const (
    OperationKindLinkCreate = domain.OperationKindLinkCreate
    OperationKindLinkDelete = domain.OperationKindLinkDelete
    OperationKindDirCreate  = domain.OperationKindDirCreate
    OperationKindDirDelete  = domain.OperationKindDirDelete
    OperationKindFileMove   = domain.OperationKindFileMove
    OperationKindFileBackup = domain.OperationKindFileBackup
)

// Plan represents a validated, executable plan.
type Plan = domain.Plan

// Package represents a package with its files and metadata.
type Package = domain.Package

// Result represents a value or an error.
type Result[T any] = domain.Result[T]

// Result constructors
var (
    Ok  = domain.Ok
    Err = domain.Err
)

// Map transforms a Result value.
var Map = domain.Map

// FlatMap chains Result operations.
var FlatMap = domain.FlatMap

// Status represents installation status information.
type Status = planner.Status

// DiagnosticReport contains health check results.
type DiagnosticReport = resolver.DiagnosticReport

// PackageInfo contains metadata about an installed package.
type PackageInfo = domain.PackageInfo

// Conflict represents a detected conflict.
type Conflict = resolver.Conflict

// ConflictType identifies the type of conflict.
type ConflictType = resolver.ConflictType

// Conflict types
const (
    ConflictFileExists   = resolver.ConflictFileExists
    ConflictWrongLink    = resolver.ConflictWrongLink
    ConflictPermission   = resolver.ConflictPermission
    ConflictCircular     = resolver.ConflictCircular
)

// Warning represents a non-fatal issue.
type Warning = resolver.Warning
```

**Tasks**:
- [ ] Re-export Operation interface and types
- [ ] Re-export Plan type
- [ ] Re-export Package type
- [ ] Re-export Result type and functions
- [ ] Re-export Status and diagnostic types
- [ ] Re-export Conflict and Warning types
- [ ] Write tests verifying exports
- [ ] Document all exported types

**Commit**: `feat(types): re-export domain types for public API`

---

#### 12.4.2: Package Documentation

**Implementation**:
```go
// pkg/dot/doc.go
/*
Package dot provides a modern, type-safe symlink manager for dotfiles.

dot is a feature-complete GNU Stow replacement written in Go 1.25.1.
It follows strict constitutional principles: test-driven development,
atomic operations, functional programming, and comprehensive error handling.

# Basic Usage

Create a client and stow packages:

    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        Build()

    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Stow(ctx, "vim", "zsh", "git"); err != nil {
        log.Fatal(err)
    }

# Dry Run Mode

Preview operations without applying changes:

    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        WithDryRun(true).
        Build()

    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Shows plan without executing
    if err := client.Stow(ctx, "vim"); err != nil {
        log.Fatal(err)
    }

# Streaming API

For memory-efficient processing of large package sets:

    stream := client.StowStream(ctx, "large-package")

    for result := range stream {
        if !result.IsOk() {
            log.Printf("error: %v", result.Error())
            continue
        }
        
        op := result.Value()
        log.Printf("operation: %v", op)
    }

# Query Operations

Check installation status:

    status, err := client.Status(ctx, "vim")
    if err != nil {
        log.Fatal(err)
    }

    for _, pkg := range status.Packages {
        fmt.Printf("%s: %d links\n", pkg.Name, pkg.LinkCount)
    }

Validate installation health:

    report, err := client.Doctor(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if len(report.BrokenLinks) > 0 {
        fmt.Printf("Found %d broken links\n", len(report.BrokenLinks))
    }

List installed packages:

    packages, err := client.List(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, pkg := range packages {
        fmt.Printf("%s (installed %s)\n", pkg.Name, pkg.InstalledAt)
    }

# Architecture

The library follows a functional core, imperative shell architecture:

  - Pure planning functions with no side effects
  - Explicit dependency injection for all infrastructure
  - Result monad for composable error handling
  - Phantom types for compile-time path safety
  - Transaction safety with two-phase commit and rollback

# Observability

The library provides first-class observability through injected ports:

  - Structured logging via ports.Logger interface
  - Distributed tracing via ports.Tracer interface (OpenTelemetry)
  - Metrics collection via ports.Metrics interface (Prometheus)

# Testing

The library is designed for testability:

  - All operations accept context.Context for cancellation
  - Filesystem abstraction enables testing without disk I/O
  - Pure functional core enables property-based testing
  - Comprehensive test coverage (>80%)

For more examples and detailed documentation, see the examples/ directory
and https://github.com/yourorg/dot/docs
*/
package dot
```

**Tasks**:
- [ ] Write comprehensive package documentation
- [ ] Add usage examples in doc.go
- [ ] Document architecture principles
- [ ] Document observability features
- [ ] Document testing approach
- [ ] Add links to detailed documentation

**Commit**: `docs(api): add comprehensive package documentation`

---

#### 12.4.3: Example Tests

**Implementation**:
```go
// pkg/dot/examples_test.go
package dot_test

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yourorg/dot/internal/adapters/osfs"
    "github.com/yourorg/dot/internal/adapters/slogger"
    "github.com/yourorg/dot/pkg/dot"
)

func ExampleClient_Stow() {
    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    if err := client.Stow(ctx, "vim", "zsh"); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Packages installed successfully")
    // Output: Packages installed successfully
}

func ExampleClient_PlanStow() {
    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    plan, err := client.PlanStow(ctx, "vim")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Plan contains %d operations\n", len(plan.Operations()))
    // Output: Plan contains 3 operations
}

func ExampleClient_StowStream() {
    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    stream := client.StowStream(ctx, "vim")
    
    count := 0
    for result := range stream {
        if !result.IsOk() {
            log.Printf("error: %v", result.Error())
            continue
        }
        count++
    }
    
    fmt.Printf("Processed %d operations\n", count)
    // Output: Processed 3 operations
}

func ExampleClient_Status() {
    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    status, err := client.Status(ctx, "vim")
    if err != nil {
        log.Fatal(err)
    }
    
    for _, pkg := range status.Packages {
        fmt.Printf("%s: %d links\n", pkg.Name, pkg.LinkCount)
    }
    // Output: vim: 5 links
}

func ExampleConfigBuilder() {
    cfg := dot.NewConfig().
        WithStowDir("/home/user/dotfiles").
        WithTargetDir("/home/user").
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        WithLinkMode(dot.LinkAbsolute).
        WithFolding(false).
        WithDryRun(true).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Client configured with dry-run: %v\n", client.Config().DryRun)
    // Output: Client configured with dry-run: true
}

func ExampleStreamMap() {
    ctx := context.Background()
    
    // Create input stream
    input := make(chan dot.Result[int])
    go func() {
        defer close(input)
        for i := 1; i <= 3; i++ {
            input <- dot.Ok(i)
        }
    }()
    
    // Map values
    output := dot.StreamMap(ctx, input, func(x int) int {
        return x * 2
    })
    
    // Collect results
    for result := range output {
        if result.IsOk() {
            fmt.Printf("%d ", result.Value())
        }
    }
    // Output: 2 4 6
}

func ExampleCollectStream() {
    ctx := context.Background()
    
    // Create input stream
    input := make(chan dot.Result[int])
    go func() {
        defer close(input)
        for i := 1; i <= 3; i++ {
            input <- dot.Ok(i)
        }
    }()
    
    // Collect all values
    result := dot.CollectStream(ctx, input)
    if result.IsOk() {
        fmt.Printf("Collected: %v\n", result.Value())
    }
    // Output: Collected: [1 2 3]
}
```

**Tasks**:
- [ ] Write Example_Stow test
- [ ] Write Example_PlanStow test
- [ ] Write Example_StowStream test
- [ ] Write Example_Status test
- [ ] Write Example_ConfigBuilder test
- [ ] Write Example_StreamMap test
- [ ] Write Example_CollectStream test
- [ ] Verify examples appear in godoc

**Commit**: `docs(api): add example tests for godoc`

---

### 12.5: Standalone Examples (Priority: Low)

#### 12.5.1: Basic Usage Example

**File**: `examples/basic/main.go`

**Implementation**:
```go
// examples/basic/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yourorg/dot/internal/adapters/osfs"
    "github.com/yourorg/dot/internal/adapters/slogger"
    "github.com/yourorg/dot/pkg/dot"
)

func main() {
    // Get directories from environment or use defaults
    stowDir := getEnv("STOW_DIR", "./dotfiles")
    targetDir := getEnv("TARGET_DIR", os.Getenv("HOME"))
    
    // Create configuration
    cfg := dot.NewConfig().
        WithStowDir(stowDir).
        WithTargetDir(targetDir).
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        WithFolding(true).
        Build()
    
    // Create client
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    ctx := context.Background()
    
    // Get packages from command line args
    packages := os.Args[1:]
    if len(packages) == 0 {
        packages = []string{"vim", "zsh", "git"}
    }
    
    // Stow packages
    fmt.Printf("Installing packages: %v\n", packages)
    if err := client.Stow(ctx, packages...); err != nil {
        log.Fatalf("Failed to stow packages: %v", err)
    }
    
    fmt.Println("Packages installed successfully")
    
    // Show status
    status, err := client.Status(ctx, packages...)
    if err != nil {
        log.Fatalf("Failed to get status: %v", err)
    }
    
    fmt.Println("\nInstalled packages:")
    for _, pkg := range status.Packages {
        fmt.Printf("  %s: %d links\n", pkg.Name, pkg.LinkCount)
    }
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**Tasks**:
- [ ] Create examples/basic/main.go
- [ ] Add README explaining usage
- [ ] Test example compiles
- [ ] Test example runs
- [ ] Add to CI verification

**Commit**: `docs(examples): add basic usage example`

---

#### 12.5.2: Streaming Example

**File**: `examples/streaming/main.go`

**Implementation**:
```go
// examples/streaming/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yourorg/dot/internal/adapters/osfs"
    "github.com/yourorg/dot/internal/adapters/slogger"
    "github.com/yourorg/dot/pkg/dot"
)

func main() {
    stowDir := getEnv("STOW_DIR", "./dotfiles")
    targetDir := getEnv("TARGET_DIR", os.Getenv("HOME"))
    
    cfg := dot.NewConfig().
        WithStowDir(stowDir).
        WithTargetDir(targetDir).
        WithFS(osfs.New()).
        WithLogger(slogger.New()).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    packages := os.Args[1:]
    if len(packages) == 0 {
        log.Fatal("Usage: streaming PACKAGE...")
    }
    
    ctx := context.Background()
    
    fmt.Printf("Streaming operations for packages: %v\n", packages)
    
    // Get operation stream
    stream := client.StowStream(ctx, packages...)
    
    // Process operations as they arrive
    var (
        totalOps   int
        linkOps    int
        dirOps     int
        errors     int
    )
    
    for result := range stream {
        totalOps++
        
        if !result.IsOk() {
            errors++
            fmt.Printf("ERROR: %v\n", result.Error())
            continue
        }
        
        op := result.Value()
        switch op.Kind() {
        case dot.OperationKindLinkCreate, dot.OperationKindLinkDelete:
            linkOps++
        case dot.OperationKindDirCreate, dot.OperationKindDirDelete:
            dirOps++
        }
        
        if totalOps%10 == 0 {
            fmt.Printf("  Processed %d operations...\n", totalOps)
        }
    }
    
    fmt.Printf("\nCompleted: %d operations (%d links, %d dirs, %d errors)\n",
        totalOps, linkOps, dirOps, errors)
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**Tasks**:
- [ ] Create examples/streaming/main.go
- [ ] Add README explaining streaming benefits
- [ ] Test example compiles
- [ ] Test example runs
- [ ] Add to CI verification

**Commit**: `docs(examples): add streaming API example`

---

#### 12.5.3: Custom Filesystem Example

**File**: `examples/custom-fs/main.go`

**Implementation**:
```go
// examples/custom-fs/main.go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yourorg/dot/internal/adapters/memfs"
    "github.com/yourorg/dot/internal/adapters/slogger"
    "github.com/yourorg/dot/pkg/dot"
)

func main() {
    // Use in-memory filesystem for testing
    fs := memfs.New()
    
    // Set up test fixtures
    setupFixtures(fs)
    
    // Create configuration with custom filesystem
    cfg := dot.NewConfig().
        WithStowDir("/test/stow").
        WithTargetDir("/test/target").
        WithFS(fs). // Custom filesystem
        WithLogger(slogger.New()).
        Build()
    
    client, err := dot.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    ctx := context.Background()
    
    // Stow packages
    if err := client.Stow(ctx, "vim", "git"); err != nil {
        log.Fatalf("Failed to stow: %v", err)
    }
    
    fmt.Println("Stowed packages successfully")
    
    // Verify links created
    verifyLinks(fs)
}

func setupFixtures(fs dot.FS) {
    // Create stow directory structure
    fs.MkdirAll(context.Background(), "/test/stow/vim", 0755)
    fs.MkdirAll(context.Background(), "/test/stow/git", 0755)
    fs.MkdirAll(context.Background(), "/test/target", 0755)
    
    // Create package files
    fs.WriteFile(context.Background(), "/test/stow/vim/dot-vimrc", []byte("vim config"), 0644)
    fs.WriteFile(context.Background(), "/test/stow/git/dot-gitconfig", []byte("git config"), 0644)
    
    fmt.Println("Created test fixtures")
}

func verifyLinks(fs dot.FS) {
    ctx := context.Background()
    
    links := []string{
        "/test/target/.vimrc",
        "/test/target/.gitconfig",
    }
    
    for _, link := range links {
        if fs.IsSymlink(ctx, link) {
            target, _ := fs.ReadLink(ctx, link)
            fmt.Printf("✓ %s -> %s\n", link, target)
        } else {
            fmt.Printf("✗ %s (not a symlink)\n", link)
        }
    }
}
```

**Tasks**:
- [ ] Create examples/custom-fs/main.go
- [ ] Add README explaining custom filesystem usage
- [ ] Test example compiles
- [ ] Test example runs
- [ ] Add to CI verification

**Commit**: `docs(examples): add custom filesystem example`

---

## Testing Strategy

### Unit Tests
- Test each public method with valid inputs
- Test each public method with invalid inputs
- Test configuration validation
- Test default value application
- Test context cancellation
- Test error propagation

### Integration Tests
- Test complete stow workflow
- Test complete unstow workflow
- Test complete restow workflow
- Test complete adopt workflow
- Test query operations
- Test streaming API end-to-end

### Property Tests
- Verify configuration validation laws
- Verify Result monad laws
- Verify stream operator laws

### Example Tests
- Verify all examples compile
- Verify all examples appear in godoc
- Verify examples in CI

---

## Quality Gates

### Definition of Done

Each task is complete when:
- [ ] Implementation follows test-first approach
- [ ] All tests pass
- [ ] Test coverage ≥ 80% for new code
- [ ] All linters pass (golangci-lint)
- [ ] Code reviewed against constitution
- [ ] Documentation updated
- [ ] Examples updated if needed
- [ ] Atomic commit created

### Phase Completion Criteria

Phase 12 is complete when:
- [ ] All 12.1-12.5 tasks completed
- [ ] Client facade fully functional
- [ ] Configuration system complete
- [ ] Streaming API operational
- [ ] Types properly exported
- [ ] All examples working
- [ ] Documentation comprehensive
- [ ] Test coverage ≥ 80%
- [ ] All linters pass
- [ ] Integration tests pass
- [ ] API suitable for library embedding

---

## Development Workflow

### For Each Task

1. **Read Test**: Review test specifications
2. **Write Test**: Implement failing tests (red)
3. **Implement**: Write minimum code to pass (green)
4. **Refactor**: Improve while maintaining green
5. **Lint**: Run `make check`
6. **Commit**: Create atomic commit

### Commit Message Format

```
<type>(scope): <description>

[optional body]

[optional footer]
```

Examples:
```
feat(api): add Client facade with constructor
feat(api): implement Stow and PlanStow methods
feat(config): define Config struct with validation
feat(streaming): add streaming operation methods
docs(api): add comprehensive package documentation
```

---

## Dependencies

### Internal Dependencies (Must Exist)
- `internal/domain/*` - Domain model (Phase 1)
- `internal/ports/*` - Infrastructure ports (Phase 2)
- `internal/adapters/*` - Port implementations (Phase 3)
- `internal/scanner/*` - Scanner (Phase 4)
- `internal/ignore/*` - Ignore system (Phase 5)
- `internal/planner/*` - Planner (Phase 6)
- `internal/resolver/*` - Resolver (Phase 7)
- `internal/planner/graph.go` - Topological sorter (Phase 8)
- `internal/pipeline/*` - Pipeline orchestration (Phase 9)
- `internal/executor/*` - Executor (Phase 10)
- `internal/manifest/*` - Manifest management (Phase 11)

### External Dependencies
- Standard library only for core
- `github.com/stretchr/testify` for tests

---

## Risk Mitigation

### Technical Risks

1. **API Surface Too Large**
   - Mitigation: Start minimal, add based on need
   - Mitigation: Keep internal complexity hidden

2. **Backward Compatibility**
   - Mitigation: Semantic versioning
   - Mitigation: Deprecation warnings before removal

3. **Performance of Streaming API**
   - Mitigation: Benchmark early
   - Mitigation: Profile memory usage
   - Mitigation: Add backpressure handling

### Process Risks

1. **Documentation Drift**
   - Mitigation: Example tests in godoc
   - Mitigation: CI verification of examples
   - Mitigation: Regular documentation review

2. **API Complexity**
   - Mitigation: Simple primary API (Stow, Unstow, etc.)
   - Mitigation: Advanced features optional
   - Mitigation: Builder pattern for configuration

---

## Success Metrics

- [ ] All core operations accessible via public API
- [ ] Configuration system flexible and type-safe
- [ ] Streaming API memory-efficient for large operations
- [ ] Documentation comprehensive and accurate
- [ ] Examples demonstrate common use cases
- [ ] Test coverage ≥ 80%
- [ ] No linter warnings
- [ ] Library embeddable in other tools
- [ ] Zero CLI dependencies in pkg/dot/

---

## Timeline Estimate

**Total Effort**: 8-12 hours

- 12.1 Client Facade: 4-5 hours
- 12.2 Configuration: 2-3 hours
- 12.3 Streaming API: 2-3 hours
- 12.4 Type Exports: 1-2 hours
- 12.5 Examples: 1-2 hours

**Assumptions**:
- Phases 1-11 complete and tested
- All internal dependencies available
- No major design changes required
- Single developer, focused work

---

## Next Steps After Phase 12

After completing the public library API:

1. **Phase 13**: CLI Layer - Core Commands
   - Cobra integration
   - Flag binding to Config
   - Command implementations using Client API

2. **Phase 14**: CLI Layer - Query Commands
   - Output formatting
   - Table rendering
   - JSON/YAML output

3. **Integration Testing**: End-to-end testing with public API

---

## References

- [Implementation Plan](./Implementation-Plan.md)
- [Architecture Documentation](./Architecture.md)
- [Features Specification](./Features.md)
- [Project Constitution](../.cursor/rules/*.mdc)
- [Go Module Documentation](https://go.dev/blog/v2-go-modules)

