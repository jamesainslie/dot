# Breaking Changes in v0.2.0

## Client.Doctor() Signature Change

**Affected**: Library consumers using `pkg/dot.Client` interface  
**Impact**: API breaking change  
**Introduced In**: Phase-15c (PR #13)  

### What Changed

The `Doctor()` method signature was updated to accept a `ScanConfig` parameter for controlling orphaned link detection behavior.

**Before** (v0.1.x):
```go
Doctor(ctx context.Context) (DiagnosticReport, error)
```

**After** (v0.2.0):
```go
Doctor(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)
```

### Migration Guide

**For library consumers**, update all `Doctor()` calls to include `ScanConfig`:

**Minimal Change** (preserve existing behavior):
```go
// Before
report, err := client.Doctor(ctx)

// After  
report, err := client.Doctor(ctx, dot.DefaultScanConfig())
```

**Enable Orphan Detection** (new feature):
```go
// Scoped scanning (recommended)
report, err := client.Doctor(ctx, dot.ScopedScanConfig())

// Deep scanning with custom depth
report, err := client.Doctor(ctx, dot.DeepScanConfig(5))
```

### Rationale

Adding the `ScanConfig` parameter enables:
- Control over orphaned link detection
- Performance tuning (depth limits, skip patterns)
- Opt-in scanning modes without global state
- Configuration per-call instead of per-client

### CLI Impact

**None** - The CLI (dot doctor command) is not affected. The `--scan-mode` flag is optional and defaults to "off", preserving current behavior.

```bash
# Existing usage still works
dot doctor

# New functionality available
dot doctor --scan-mode=scoped
```

### Future Improvements

**TODO for v0.3.0**: Consider transitional API pattern:

1. Add new method:
   ```go
   DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)
   ```

2. Update existing method to call new one:
   ```go
   // Deprecated: Use DoctorWithScan instead. Will be removed in v1.0.0.
   func (c *Client) Doctor(ctx context.Context) (DiagnosticReport, error) {
       return c.DoctorWithScan(ctx, DefaultScanConfig())
   }
   ```

3. Provide deprecation notice and migration period

4. Remove deprecated method in v1.0.0

This would allow gradual migration for library consumers.

### Detection

Library consumers will encounter compilation errors:

```
not enough arguments in call to client.Doctor
    have (context.Context)
    want (context.Context, ScanConfig)
```

The error clearly indicates the required change.

### Alternatives Considered

**Option 1**: Functional options pattern
```go
Doctor(ctx context.Context, opts ...DoctorOption) (DiagnosticReport, error)
```
- Pro: Backward compatible
- Con: More complex, requires option type infrastructure

**Option 2**: Separate method
```go
Doctor(ctx context.Context) (DiagnosticReport, error)  // existing
DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)  // new
```
- Pro: No breaking change
- Con: API duplication, both methods must be maintained

**Option 3**: Required parameter (chosen)
```go
Doctor(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)
```
- Pro: Clean API, explicit configuration
- Con: Breaking change for existing consumers

**Decision**: Chose Option 3 as project is pre-v1.0.0 and breaking changes are acceptable. Future versions can add transitional paths if needed.

### Version Policy

Per semantic versioning for pre-1.0.0 releases:
- Minor version bump (0.1.x → 0.2.0) acceptable for breaking changes
- Major version (1.0.0) not required until API is stable
- Document all breaking changes in CHANGELOG

---

**Document Version**: 1.0  
**Created**: 2025-10-06  
**Phase**: 15c

