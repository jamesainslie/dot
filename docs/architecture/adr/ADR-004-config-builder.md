# ADR-004: Fluent Configuration Builder

**Status**: Proposed  
**Date**: 2025-10-07  
**Deciders**: Development Team  
**Context**: Phase 22.6 Future Enhancements Planning

## Context

The current `Config` struct requires all fields to be set at initialization, leading to verbose code especially when using non-default values.

**Current API**:
```go
cfg := dot.Config{
    PackageDir: "/home/user/dotfiles",
    TargetDir:  "/home/user",
    BackupDir:  "/custom/backup",
    DryRun:     true,
    Verbosity:  2,
    FS:         adapters.NewOSFilesystem(),
    Logger:     adapters.NewSlogLogger(slog.Default()),
    Tracer:     adapters.NewNoopTracer(),
    Metrics:    adapters.NewNoopMetrics(),
}
cfg = cfg.WithDefaults() // Must remember to call this
```

**Problems**:
1. Verbose initialization even for simple cases
2. Easy to forget `WithDefaults()`
3. No progressive disclosure of options
4. Unclear which fields are required vs optional

## Decision

Add **optional ConfigBuilder** for fluent configuration while maintaining struct-based configuration for direct initialization.

### Proposed API

```go
// Simple case
cfg := dot.NewConfigBuilder().
    WithPackageDir("~/dotfiles").
    WithTargetDir("~").
    Build()

// Advanced case
cfg := dot.NewConfigBuilder().
    WithPackageDir("~/dotfiles").
    WithTargetDir("~").
    WithBackupDir("/custom/backup").
    WithDryRun(true).
    WithVerbosity(2).
    WithFS(customFS).
    WithLogger(customLogger).
    Build()

// Minimal case (all defaults)
cfg := dot.NewConfigBuilder().Build()
```

### Implementation

```go
// ConfigBuilder provides fluent API for building Config
type ConfigBuilder struct {
    cfg Config
}

// NewConfigBuilder creates a new builder with defaults
func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{
        cfg: Config{
            FS:     adapters.NewOSFilesystem(),
            Logger: adapters.NewNoopLogger(),
            Tracer: adapters.NewNoopTracer(),
            Metrics: adapters.NewNoopMetrics(),
        },
    }
}

// WithPackageDir sets the package directory
func (b *ConfigBuilder) WithPackageDir(dir string) *ConfigBuilder {
    b.cfg.PackageDir = dir
    return b
}

// WithTargetDir sets the target directory
func (b *ConfigBuilder) WithTargetDir(dir string) *ConfigBuilder {
    b.cfg.TargetDir = dir
    return b
}

// ... more With methods ...

// Build returns the configured Config
func (b *ConfigBuilder) Build() (Config, error) {
    // Validate
    if err := b.cfg.Validate(); err != nil {
        return Config{}, err
    }
    
    // Apply defaults
    return b.cfg.WithDefaults(), nil
}

// MustBuild returns the configured Config or panics
func (b *ConfigBuilder) MustBuild() Config {
    cfg, err := b.Build()
    if err != nil {
        panic(err)
    }
    return cfg
}
```

## Alternatives Considered

### Option A: Functional Options
```go
func NewClient(opts ...ConfigOption) (Client, error) {
    cfg := DefaultConfig()
    for _, opt := range opts {
        opt(&cfg)
    }
    return newClient(cfg)
}
```

**Rejected**: Makes Config opaque, harder to inspect, less discoverable.

### Option B: Required + Optional Pattern
```go
func NewConfig(required RequiredConfig, opts ...ConfigOption) Config
```

**Rejected**: Still verbose, unclear what's required, not fluent.

### Option C: Builder Pattern ✅
**Selected**: Fluent, discoverable, validates at build time, coexists with struct pattern.

## Benefits

### User Experience
- ✅ Fluent, readable API
- ✅ Progressive disclosure of options
- ✅ Defaults applied automatically
- ✅ Validation at build time
- ✅ Clear error messages

### Developer Experience
- ✅ Easy to add new options
- ✅ Backward compatible (new methods don't break old code)
- ✅ Self-documenting (method names describe options)
- ✅ Works well with IDE autocomplete

### Flexibility
- ✅ Both builder and struct patterns available
- ✅ Can switch between patterns easily
- ✅ Builder can be extended without breaking changes

## Trade-offs

### Pros
- Clear, fluent API
- Progressive disclosure
- Automatic defaults
- Validation built-in
- Backward compatible

### Cons
- Additional API surface area
- Two ways to do same thing
- Slightly more code to maintain
- Learning curve for new users (which pattern to use?)

## Implementation Plan

**Phase A: Core Builder** (2 hours)
- Create ConfigBuilder struct
- Implement With methods for all Config fields
- Add Build() and MustBuild()
- Unit tests

**Phase B: Validation** (1 hour)
- Validate during Build()
- Helpful error messages
- Test error scenarios

**Phase C: Documentation** (1 hour)
- Godoc with examples
- Update user guide
- Migration examples

**Total Effort**: 4 hours

## Usage Examples

### Basic Usage
```go
// Simple setup with defaults
cfg := dot.NewConfigBuilder().
    WithPackageDir("~/dotfiles").
    Build()
```

### Advanced Usage
```go
// Custom adapters and options
cfg := dot.NewConfigBuilder().
    WithPackageDir("~/dotfiles").
    WithTargetDir("~").
    WithBackupDir("/var/backup").
    WithDryRun(true).
    WithFS(customFS).
    WithLogger(slog.Default()).
    Build()
```

### Testing
```go
// Easy to create test configs
cfg := dot.NewConfigBuilder().
    WithPackageDir(t.TempDir()).
    WithTargetDir(t.TempDir()).
    WithFS(adapters.NewMemFS()).
    WithLogger(adapters.NewNoopLogger()).
    MustBuild()
```

## Migration Path

### For Existing Code

Old code continues to work:
```go
// Still valid - no changes needed
cfg := dot.Config{
    PackageDir: "/path",
    TargetDir:  "/target",
    FS:         fs,
    Logger:     logger,
}
cfg = cfg.WithDefaults()
```

New code can use builder:
```go
// Cleaner alternative
cfg := dot.NewConfigBuilder().
    WithPackageDir("/path").
    WithTargetDir("/target").
    WithFS(fs).
    WithLogger(logger).
    Build()
```

### Recommendation

Documentation should show builder pattern in primary examples while noting that struct pattern is also available for direct initialization.

## Success Criteria

- [ ] ConfigBuilder implemented with all Config fields
- [ ] Build() validates and applies defaults
- [ ] MustBuild() panics on invalid config
- [ ] All With methods return *ConfigBuilder for chaining
- [ ] Comprehensive unit tests
- [ ] Documentation with examples
- [ ] No breaking changes to existing API

## When to Implement

**Defer to**: v0.3.0 or later  
**Trigger**: User requests for simpler configuration API  
**Priority**: Low (nice-to-have)  
**Effort**: 4 hours

This is a quality-of-life improvement, not a critical feature. The current struct-based API is sufficient and works well. ConfigBuilder would improve ergonomics but isn't necessary for core functionality.

## References

- **Builder Pattern**: https://refactoring.guru/design-patterns/builder
- **Fluent Interface**: https://en.wikipedia.org/wiki/Fluent_interface
- **Go Builder Patterns**: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis

## Review

**Status**: Design complete, implementation deferred  
**Next Steps**: Implement if user feedback indicates configuration ergonomics are a pain point  
**Alternative**: Improve documentation for struct-based pattern instead

