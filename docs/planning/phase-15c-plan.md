# Phase 15c: API Enhancement - Orphaned Link Detection and Incremental Remanage

**Status**: Planning  
**Priority**: Medium  
**Estimated Effort**: 2-3 days  
**Dependencies**: Phase 15 (Error Handling and UX) complete  

---

## Overview

This phase addresses three TODO items in the internal/api package that represent partially implemented or planned features. The primary focus is implementing orphaned link detection with performance constraints and enhancing remanage operations with incremental planning.

### Current State

Three TODO items exist in internal/api:

1. **Orphaned Link Detection** (doctor.go:55)
   - Complete implementation exists but is not called
   - Functions: `scanForOrphanedLinks()`, `isLinkManaged()`
   - Impact: 52 lines of code at 0% coverage
   - Reason: Performance concern about scanning entire home directory

2. **Incremental Remanage** (remanage.go:11, 25)
   - Currently performs full unmanage + manage cycle
   - TODO: Use hash-based change detection
   - Impact: Inefficient for large packages with few changes

3. **Link Count Tracking** (manage.go:109)
   - Manifest stores link count but not populated from planner
   - Minor enhancement for consistency

### Goals

- Enable orphaned link detection with performance safeguards
- Implement hash-based incremental remanage
- Improve link count tracking accuracy
- Increase internal/api test coverage from 73.3% to 80%+

---

## Problem Statement

### 1. Orphaned Link Detection

**Problem**: Users may have symlinks in their target directory that are no longer managed by dot (orphaned links). The `doctor` command should detect these, but the implementation is disabled due to performance concerns.

**Current Implementation**: 
- `scanForOrphanedLinks()` recursively scans a directory tree
- `isLinkManaged()` checks if a link is in the manifest
- Both functions complete but never called

**Performance Concern**:
- Scanning entire $HOME directory could check thousands of files
- O(n*m) complexity: n files × m manifest entries
- Potential for very slow doctor command

**User Impact**:
- Orphaned links go undetected
- Manual cleanup required
- Reduced confidence in system state

### 2. Inefficient Remanage

**Problem**: `Remanage` operation currently uninstalls all links then reinstalls them, even if most files haven't changed.

**Current Behavior**:
```go
func Remanage(packages) {
    Unmanage(packages)  // Remove ALL links
    Manage(packages)    // Recreate ALL links
}
```

**Inefficiency**:
- Unnecessary I/O for unchanged files
- Potential for temporary unavailability
- Slower than necessary for large packages

**User Impact**:
- Slow remanage operations
- Brief window where config files unavailable
- Reduced user experience

### 3. Inaccurate Link Counts

**Problem**: Manifest stores `link_count` per package, but the value is not populated from actual planning results.

**Current Behavior**:
```go
m.AddPackage(PackageInfo{
    LinkCount: 0,  // TODO: Track from planner
})
```

**User Impact**:
- Inaccurate package statistics
- Misleading output in `status` and `list` commands

---

## Requirements

### Functional Requirements

#### FR-1: Orphaned Link Detection
- Doctor command SHOULD detect symlinks not managed by dot
- Scanning MUST be limited to prevent performance issues
- Detection MUST be opt-in or use intelligent scope limiting
- False positives MUST be minimized

#### FR-2: Scoped Scanning
- By default, ONLY scan directories containing managed links
- Provide `--deep` flag for full home directory scan with warning
- Respect depth limits to prevent infinite recursion
- Skip known large directories (.git, node_modules, etc.)

#### FR-3: Incremental Remanage
- Remanage SHOULD skip unchanged files
- Use manifest hash comparison to detect changes
- Only unmanage/remanage files that changed
- Maintain transactional safety

#### FR-4: Link Count Accuracy
- Manifest MUST store accurate link counts
- Counts SHOULD be updated after each operation
- Counts MUST match actual filesystem state

### Non-Functional Requirements

#### NFR-1: Performance
- Orphaned link scan MUST complete in < 5 seconds for typical home directory
- Scoped scan SHOULD complete in < 1 second
- Incremental remanage MUST be faster than full remanage for unchanged files

#### NFR-2: Test Coverage
- All new code MUST have 80%+ test coverage
- Orphaned link detection MUST have comprehensive tests
- Incremental logic MUST be thoroughly tested
- Performance characteristics MUST be validated

#### NFR-3: Backward Compatibility
- Existing doctor behavior MUST not change by default
- Manifest format remains compatible
- No breaking changes to public API

---

## Technical Design

### Architecture

```
┌─────────────────────────────────────────┐
│ cmd/dot/doctor.go                       │
│ - Add --deep flag for full scan          │
│ - Add --scope flag for custom dirs      │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│ internal/api/doctor.go                  │
│ - Enable scanForOrphanedLinks()         │
│ - Add ScanConfig parameter              │
│ - Implement scope limiting               │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│ pkg/dot/diagnostics.go                  │
│ - Add ScanConfig type                   │
│ - Add ScanMode enum (scoped/deep/off)  │
└─────────────────────────────────────────┘
```

### Data Model

#### ScanConfig

```go
// ScanConfig controls orphaned link detection behavior.
type ScanConfig struct {
    Mode       ScanMode
    MaxDepth   int
    ScopeToDirs []string  // Only scan these dirs (empty = auto-detect)
    SkipPatterns []string  // Patterns to skip (.git, node_modules, etc.)
}

type ScanMode int

const (
    ScanOff     ScanMode = iota  // No orphan scanning (current behavior)
    ScanScoped                   // Only scan dirs with managed links (default)
    ScanDeep                     // Full recursive scan with limits
)
```

#### HashComparison

```go
// PackageChanges represents what changed in a package.
type PackageChanges struct {
    PackageName string
    Changed     []string  // Files that changed
    Added       []string  // New files
    Removed     []string  // Deleted files
    Unchanged   []string  // Files with same hash
}

// DetectChanges compares manifest hash with current package state.
func DetectChanges(ctx context.Context, pkg string, manifest Manifest, fs FS) (PackageChanges, error)
```

### Implementation Approach

#### 1. Orphaned Link Detection

**Phase 1**: Enable with Scope Limiting (Default Safe)

```go
func (c *client) Doctor(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
    // ... existing checks ...

    if scanCfg.Mode != ScanOff {
        // Determine scan directories
        scanDirs := scanCfg.ScopeToDirs
        if len(scanDirs) == 0 && scanCfg.Mode == ScanScoped {
            // Auto-detect: extract unique directories from managed links
            scanDirs = extractManagedDirectories(m)
        } else if len(scanDirs) == 0 && scanCfg.Mode == ScanDeep {
            // Full scan: use target directory
            scanDirs = []string{c.config.TargetDir}
        }

        // Scan with limits
        for _, dir := range scanDirs {
            err := c.scanForOrphanedLinksWithLimits(ctx, dir, m, scanCfg, &issues, &stats)
            if err != nil {
                // Log but continue - orphan detection is best-effort
            }
        }
    }

    // ... rest of function ...
}
```

**Phase 2**: Add Safety Guards

```go
func (c *client) scanForOrphanedLinksWithLimits(
    ctx context.Context,
    dir string,
    m *Manifest,
    scanCfg ScanConfig,
    issues *[]Issue,
    stats *DiagnosticStats,
) error {
    // Check context cancellation
    if ctx.Err() != nil {
        return ctx.Err()
    }

    // Check depth limit
    depth := calculateDepth(dir, c.config.TargetDir)
    if depth > scanCfg.MaxDepth {
        return nil  // Skip too-deep directories
    }

    // Skip known large directories
    if shouldSkipDirectory(dir, scanCfg.SkipPatterns) {
        return nil
    }

    // Call existing scanForOrphanedLinks with limits
    return c.scanForOrphanedLinks(ctx, dir, m, issues, stats)
}
```

**Phase 3**: Optimize Link Lookup

```go
// Pre-build set for O(1) lookup instead of O(n)
func buildManagedLinkSet(m *Manifest) map[string]bool {
    linkSet := make(map[string]bool)
    for _, pkgInfo := range m.Packages {
        for _, link := range pkgInfo.Links {
            linkSet[link] = true
        }
    }
    return linkSet
}

// Updated isLinkManaged using set
func (c *client) isLinkManagedFast(relPath, fullPath string, linkSet map[string]bool) bool {
    return linkSet[relPath] || linkSet[fullPath]
}
```

#### 2. Incremental Remanage

**Phase 1**: Hash Comparison Infrastructure

```go
func (c *client) detectPackageChanges(
    ctx context.Context,
    pkgName string,
    m *Manifest,
) (PackageChanges, error) {
    // Get current hash from manifest
    oldHash, exists := m.GetHash(pkgName)
    if !exists {
        // Not previously installed - all files are new
        return PackageChanges{PackageName: pkgName, AllNew: true}, nil
    }

    // Compute current hash
    hasher := manifest.NewContentHasher(c.config.FS)
    newHash, err := hasher.HashPackage(ctx, c.config.StowDir, pkgName)
    if err != nil {
        return PackageChanges{}, err
    }

    // Compare hashes
    if oldHash == newHash {
        // No changes - skip remanage entirely
        return PackageChanges{PackageName: pkgName, NoChanges: true}, nil
    }

    // TODO: Implement file-level change detection
    // For now, mark all as changed if package hash differs
    return PackageChanges{PackageName: pkgName, Changed: true}, nil
}
```

**Phase 2**: Selective Remanage

```go
func (c *client) Remanage(ctx context.Context, packages ...string) error {
    targetPath := ... // validate target path

    // Load manifest
    manifestResult := c.manifest.Load(ctx, targetPath)
    m := manifestResult.UnwrapOr(manifest.New())

    for _, pkg := range packages {
        // Detect changes
        changes, err := c.detectPackageChanges(ctx, pkg, m)
        if err != nil {
            return err
        }

        if changes.NoChanges {
            // Skip - nothing to do
            c.config.Logger.Info("package unchanged, skipping", "package", pkg)
            continue
        }

        // Full remanage for changed packages
        if err := c.unmanagePackage(ctx, pkg); err != nil {
            return err
        }
        if err := c.managePackage(ctx, pkg); err != nil {
            return err
        }
    }

    return nil
}
```

#### 3. Link Count Tracking

**Simple Fix**: Extract from plan results

```go
func (c *client) updateManifest(ctx context.Context, targetPath TargetPath, packages []string, plan Plan) error {
    // ... existing code ...

    for _, pkg := range packages {
        // Count links for this package in the plan
        linkCount := countLinksForPackage(plan, pkg)

        m.AddPackage(manifest.PackageInfo{
            Name:        pkg,
            InstalledAt: time.Now(),
            LinkCount:   linkCount,  // Actual count from plan
        })
    }

    // ... save manifest ...
}

func countLinksForPackage(plan Plan, pkgName string) int {
    count := 0
    for _, op := range plan.Operations {
        if link, ok := op.(LinkCreateOperation); ok {
            // Check if this link belongs to the package
            // This requires tracking package ownership in operations
            if linkBelongsToPackage(link, pkgName) {
                count++
            }
        }
    }
    return count
}
```

---

## Implementation Tasks

### Task 1: Design and Infrastructure (Setup)

**T15c-001**: Define ScanConfig and ScanMode types
- File: `pkg/dot/diagnostics.go`
- Add ScanConfig struct with mode, depth, scope
- Add ScanMode enum (Off, Scoped, Deep)
- Add default configurations
- **Tests**: Validate ScanConfig construction and defaults
- **Estimate**: 1 hour

**T15c-002**: Add scan configuration to Doctor API
- File: `pkg/dot/client.go` (interface)
- Update Doctor method signature to accept optional ScanConfig
- Maintain backward compatibility with default ScanOff
- **Tests**: Verify interface changes
- **Estimate**: 0.5 hours

### Task 2: Orphaned Link Detection (Core)

**T15c-003**: Implement directory scope extraction
- File: `internal/api/doctor.go`
- Add `extractManagedDirectories()` function
- Extract unique parent directories from manifest links
- Handle edge cases (root directory, nested paths)
- **Tests**: Table-driven tests with various manifest structures
- **Estimate**: 1.5 hours

**T15c-004**: Add depth calculation and limiting
- File: `internal/api/doctor.go`
- Add `calculateDepth()` function
- Add `shouldSkipDirectory()` for pattern matching
- Implement max depth enforcement
- **Tests**: Test depth limits and skip patterns
- **Estimate**: 1 hour

**T15c-005**: Optimize link lookup with set data structure
- File: `internal/api/doctor.go`
- Add `buildManagedLinkSet()` function
- Replace O(n) `isLinkManaged()` with O(1) set lookup
- **Tests**: Performance comparison tests
- **Estimate**: 1 hour

**T15c-006**: Wire up orphaned link detection
- File: `internal/api/doctor.go`
- Enable `scanForOrphanedLinks()` call in `Doctor()`
- Use scoped scanning by default
- Add error handling for scan failures
- **Tests**: Integration tests with orphaned links
- **Estimate**: 2 hours

**T15c-007**: Add CLI flags for scan control
- File: `cmd/dot/doctor.go`
- Add `--scan-mode` flag (off, scoped, deep)
- Add `--max-depth` flag with default 10
- Add `--scan-dir` flag for custom scope
- **Tests**: Flag parsing and config building
- **Estimate**: 1 hour

### Task 3: Incremental Remanage (Core)

**T15c-008**: Implement package change detection
- File: `internal/api/remanage.go`
- Add `detectPackageChanges()` method
- Compare manifest hash with current package hash
- Return change status (unchanged, modified, new)
- **Tests**: Test with unchanged, modified, and new packages
- **Estimate**: 2 hours

**T15c-009**: Implement selective remanage logic
- File: `internal/api/remanage.go`
- Skip unchanged packages
- Full cycle for changed packages
- Log decisions for transparency
- **Tests**: Integration tests with various change scenarios
- **Estimate**: 2 hours

**T15c-010**: Add dry-run support for incremental detection
- File: `internal/api/remanage.go`
- Show what would be skipped/processed
- Display hash comparison results
- **Tests**: Dry-run output verification
- **Estimate**: 1 hour

### Task 4: Link Count Tracking (Polish)

**T15c-011**: Extract link count from plan
- File: `internal/api/manage.go`
- Add `countLinksInPlan()` helper
- Update manifest with actual link count
- **Tests**: Verify counts match plan operations
- **Estimate**: 1.5 hours

**T15c-012**: Add package ownership to operations
- File: `pkg/dot/operation.go`
- Add `Package` field to link operations
- Track which package each link belongs to
- **Tests**: Operation construction with package field
- **Estimate**: 1 hour

### Task 5: Testing and Documentation

**T15c-013**: Comprehensive orphaned link tests
- File: `internal/api/doctor_orphan_test.go` (new)
- Test scoped vs deep scanning
- Test depth limits
- Test skip patterns
- Test performance characteristics
- **Tests**: 15+ test cases
- **Estimate**: 2 hours

**T15c-014**: Incremental remanage tests
- File: `internal/api/remanage_incremental_test.go` (new)
- Test unchanged package skip
- Test modified package remanage
- Test new package installation
- Test hash comparison logic
- **Tests**: 10+ test cases
- **Estimate**: 2 hours

**T15c-015**: Update documentation
- File: `README.md`, `docs/Features.md`
- Document orphaned link detection
- Document --scan-mode flag
- Document incremental remanage behavior
- Add usage examples
- **Estimate**: 1 hour

**T15c-016**: Performance benchmarks
- File: `internal/api/doctor_bench_test.go` (new)
- Benchmark orphaned link scanning
- Benchmark with various directory sizes
- Validate < 5 second requirement
- **Estimate**: 1.5 hours

---

## Implementation Plan

### Phase 1: Foundation (4 hours)
- T15c-001: ScanConfig types
- T15c-002: API updates
- T15c-003: Directory scope extraction
- T15c-004: Depth and skip logic

**Deliverable**: Infrastructure ready for orphan detection

### Phase 2: Orphaned Link Detection (5 hours)
- T15c-005: Optimize link lookup
- T15c-006: Wire up scanning
- T15c-007: CLI flags
- T15c-013: Comprehensive tests

**Deliverable**: Working orphaned link detection with tests

### Phase 3: Incremental Remanage (5 hours)
- T15c-008: Change detection
- T15c-009: Selective remanage
- T15c-010: Dry-run support
- T15c-014: Incremental tests

**Deliverable**: Hash-based incremental remanage

### Phase 4: Polish and Performance (4 hours)
- T15c-011: Link count tracking
- T15c-012: Package ownership
- T15c-015: Documentation
- T15c-016: Benchmarks

**Deliverable**: Complete, documented, performant features

**Total Estimated Effort**: 18 hours (2-3 days)

---

## Testing Strategy

### Unit Tests

**Orphaned Link Detection**:
- `extractManagedDirectories()` with various manifests
- `calculateDepth()` with different path structures
- `shouldSkipDirectory()` with various patterns
- `buildManagedLinkSet()` for correctness
- Link lookup performance (O(1) vs O(n))

**Incremental Remanage**:
- `detectPackageChanges()` with same/different hashes
- Skip logic for unchanged packages
- Full cycle for changed packages
- Hash comparison accuracy

**Link Count Tracking**:
- Count extraction from plans
- Manifest update accuracy
- Various package sizes

### Integration Tests

**Orphaned Link Scenarios**:
- No orphaned links (healthy)
- Orphaned links in scanned directories
- Orphaned links in unscanned directories
- Mixed managed and orphaned links
- Deeply nested orphaned links
- Large directory trees (performance)

**Incremental Remanage Scenarios**:
- Package unchanged (skip)
- Single file modified (full remanage)
- Package deleted and recreated (detect change)
- New package (full install)
- Multiple packages with mixed changes

### Performance Tests

**Benchmarks Required**:
- Scoped scan vs deep scan timing
- Set-based lookup vs linear search
- Hash computation overhead
- Change detection performance

**Performance Targets**:
- Scoped scan: < 1 second (typical)
- Deep scan: < 5 seconds (with limits)
- Set lookup: O(1) - < 1µs per lookup
- Change detection: < 100ms per package

---

## Migration Strategy

### Backward Compatibility

**Doctor Command**:
- Default behavior: `ScanOff` (current behavior - no change)
- Opt-in: `--scan-mode=scoped` for smart scanning
- Advanced: `--scan-mode=deep --max-depth=5` for full scan

**Remanage Command**:
- Transparent optimization - no API changes
- Users see faster operations automatically
- Dry-run shows skip decisions

**Manifest Format**:
- No schema changes required
- Existing manifests work as-is
- Link counts updated on next operation

### Feature Flags

```go
// In ExtendedConfig
type ExperimentalConfig struct {
    Enabled bool
    OrphanDetection bool  // Enable orphan scanning
    IncrementalRemanage bool  // Enable hash-based skipping
}
```

Enable gradually:
1. Phase 15c: Implement behind feature flags (default off)
2. Phase 16: Enable by default after validation
3. v0.2.0: Remove flags, features stable

---

## Risk Assessment

### High Risk

**Orphaned Link Detection**:
- **Risk**: Performance degradation on large directories
- **Mitigation**: Scope limiting, depth limits, skip patterns, timeouts
- **Validation**: Benchmarks must pass before merge

**Incremental Remanage**:
- **Risk**: Hash collision causes incorrect skipping
- **Mitigation**: SHA-256 has negligible collision probability
- **Validation**: Test with intentionally similar content

### Medium Risk

**Complexity**:
- **Risk**: Added complexity in doctor and remanage logic
- **Mitigation**: Comprehensive tests, clear documentation
- **Validation**: Code review for maintainability

**False Positives**:
- **Risk**: Legitimate links flagged as orphaned
- **Mitigation**: Careful link matching logic, whitelist support
- **Validation**: Integration tests with edge cases

### Low Risk

**Link Count Tracking**:
- **Risk**: Minimal - straightforward counting
- **Mitigation**: Unit tests verify accuracy
- **Validation**: Compare counts with actual filesystem

---

## Success Criteria

### Functional

- [ ] Doctor detects orphaned links in scoped directories
- [ ] Scoped scan completes in < 1 second
- [ ] Deep scan completes in < 5 seconds with limits
- [ ] Remanage skips unchanged packages
- [ ] Remanage correctly processes changed packages
- [ ] Link counts accurate in manifest

### Technical

- [ ] internal/api coverage ≥ 80% (from 73.3%)
- [ ] All new code has 80%+ test coverage
- [ ] No performance regression in existing commands
- [ ] All linters pass (0 issues)
- [ ] All tests pass including new ones

### Quality

- [ ] Documentation complete and accurate
- [ ] Examples provided for new flags
- [ ] Benchmarks validate performance claims
- [ ] Code review approved
- [ ] Constitutional compliance maintained

---

## Dependencies

### Required

- Phase 15 complete (error handling and UX)
- Manifest hash infrastructure (already exists)
- Content hasher implementation (already exists)

### Optional

- Performance profiling tools
- Large test fixture generation
- Benchmark comparison infrastructure

---

## Deliverables

### Code

1. Enhanced `Doctor()` with orphaned link detection
2. Optimized `Remanage()` with incremental planning
3. Accurate link count tracking
4. CLI flags for scan control
5. ScanConfig infrastructure

### Tests

1. Unit tests for all new functions (80%+ coverage)
2. Integration tests for orphaned link scenarios
3. Integration tests for incremental remanage
4. Performance benchmarks
5. Edge case coverage

### Documentation

1. Updated README with new doctor flags
2. Features documentation for orphan detection
3. Architecture documentation for incremental planning
4. Usage examples and best practices
5. Performance characteristics

---

## Open Questions

1. **Default Scan Mode**: Should orphan detection be:
   - Off by default (safest, current behavior)
   - Scoped by default (better UX, minor performance cost)
   - Configurable in config file

2. **Skip Patterns**: Which directories should be skipped by default?
   - `.git`, `node_modules`, `.cache`
   - User-configurable or hardcoded?
   - Respect .gitignore?

3. **File-Level Change Detection**: For Phase 2, should we:
   - Detect which specific files changed in package
   - Only unmanage/remanage changed files
   - Track per-file hashes in manifest

4. **Orphan Actions**: Should doctor:
   - Only detect and report orphans
   - Offer to remove orphans
   - Offer to adopt orphans into packages

---

## Timeline

**Week 1** (8 hours):
- Phase 1: Foundation
- Phase 2: Orphaned link detection (partial)

**Week 2** (10 hours):
- Phase 2: Complete orphan detection
- Phase 3: Incremental remanage
- Phase 4: Polish (partial)

**Total**: 2-3 days of focused development

---

## Future Enhancements (Post-Phase 15c)

### Phase 16 Candidates

1. **File-Level Incremental Remanage**
   - Track per-file hashes in manifest
   - Only touch changed files
   - Maximum efficiency

2. **Intelligent Orphan Actions**
   - `doctor --fix` to remove orphans
   - `doctor --adopt` to interactively adopt orphans
   - Batch operations

3. **Performance Monitoring**
   - Track scan durations
   - Report performance metrics
   - Adaptive depth limiting

4. **Advanced Skip Logic**
   - Respect .gitignore patterns
   - User-defined skip patterns in config
   - Size-based skipping (skip dirs > 1GB)

---

## References

- **Current Code**: `internal/api/doctor.go:79-143` (orphan detection functions)
- **Current Code**: `internal/api/remanage.go:11-45` (remanage implementation)
- **Related**: `internal/manifest/hash.go` (hash infrastructure)
- **Related**: Phase 15 error handling (templates, user experience)

---

## Constitutional Compliance

### Test-First Development
- All tasks require tests before implementation
- 80% minimum coverage enforced
- TDD red-green-refactor cycle

### Functional Programming
- Pure change detection logic
- Side effects isolated to executor
- Immutable data structures where possible

### Academic Documentation
- Factual descriptions
- No hyperbole about performance
- Precise technical terminology
- No emojis

### Code Quality
- All linters must pass
- Cyclomatic complexity ≤ 15
- Clear error handling
- Proper context propagation

---

**Phase Owner**: TBD  
**Reviewer**: TBD  
**Target Release**: v0.2.0  
**Created**: 2025-10-06  

