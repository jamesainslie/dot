# ADR-002: Package-Operation Mapping for Manifest Tracking

**Status**: Accepted  
**Date**: 2025-10-07  
**Deciders**: Development Team  
**Context**: Phase 22.2 Implementation

## Context

The manifest currently records package installation but does not track which specific links belong to which package. This creates several problems:

1. **Inaccurate Manifest**: `LinkCount` always shows 0, `Links` array is always empty
2. **Broken Unmanage**: Cannot properly unmanage packages because we don't know which links to remove
3. **Incomplete Doctor**: Cannot accurately report package health
4. **Poor Status**: Status command cannot show which files belong to which package

**Current State**:
```go
// internal/api/manage.go:109
m.AddPackage(manifest.PackageInfo{
    Name:        pkg,
    InstalledAt: time.Now(),
    LinkCount:   0,     // Always 0 - WRONG
    Links:       []string{},  // Always empty - WRONG
})
```

**Root Cause**: The pipeline returns a unified `dot.Plan` with all operations, but operations don't carry package ownership information.

## Decision

We will use **per-package operation mapping** in the Plan type without modifying the Operation interface. This approach adds a `PackageOperations` field to `Plan` that maps package names to lists of operation IDs.

```go
type Plan struct {
    Operations []Operation
    Metadata   PlanMetadata
    
    // PackageOperations maps package names to operation IDs
    // that belong to that package
    PackageOperations map[string][]OperationID
}
```

## Alternatives Considered

### Option A: Add PackageName to Operation Interface
```go
type Operation interface {
    Kind() OperationKind
    Package() string  // NEW - would require modifying interface
    // ...
}
```

**Rejected**: Breaking change to Operation interface, requires updating all operation types, violates interface stability principle.

### Option B: Return Per-Package Plans from Pipeline
```go
type ManageResult struct {
    Plans map[string]dot.Plan  // package name → plan
}
```

**Rejected**: More complex API, harder to execute unified plan, requires changes to executor.

### Option C: Add Metadata to Operations
```go
type LinkCreate struct {
    // ...
    Metadata map[string]string  // {"package": "vim"}
}
```

**Rejected**: Adds mutable state to operations, less type-safe than dedicated field.

### Option D: Chosen - Package Mapping in Plan ✅
```go
type Plan struct {
    Operations        []Operation
    PackageOperations map[string][]OperationID
}
```

**Selected**: Clean separation, no breaking changes, maintains operation immutability, enables future features.

## Implementation Strategy

### Phase 1: Extend Plan Type (T22.2-002)
Add `PackageOperations` field to Plan with helper methods:
```go
func (p Plan) OperationsForPackage(pkg string) []Operation
```

### Phase 2: Pipeline Tracking (T22.2-003)
Update `ManagePipeline.Execute()` to build package-operation mapping:
```go
func (p *ManagePipeline) Execute(ctx context.Context, input ManageInput) dot.Result[dot.Plan] {
    // ... existing stages ...
    
    packageOps := make(map[string][]dot.OperationID)
    for _, pkg := range input.Packages {
        ops := collectOpsForPackage(desired, pkg)
        packageOps[pkg] = extractOperationIDs(ops)
    }
    
    plan := dot.Plan{
        Operations:        operations,
        PackageOperations: packageOps,
    }
    return dot.Ok(plan)
}
```

### Phase 3: Manifest Integration (T22.2-004)
Update manifest writer to extract links from package operations:
```go
func (c *client) updateManifest(...) error {
    for _, pkg := range packages {
        ops := plan.OperationsForPackage(pkg)
        links := extractLinksFromOperations(ops)
        
        m.AddPackage(manifest.PackageInfo{
            Name:      pkg,
            LinkCount: len(links),
            Links:     links,
        })
    }
}
```

## Rationale

### Benefits

1. **No Breaking Changes**: Existing Operation interface unchanged
2. **Type-Safe**: Explicit mapping with compile-time checks
3. **Flexible**: Supports one-to-many package-operation relationships
4. **Future-Proof**: Enables parallel package installation, selective remanage
5. **Maintainable**: Clear separation of concerns
6. **Testable**: Easy to verify mapping correctness

### Trade-offs

1. **Additional Memory**: Stores operation ID list per package (~few hundred bytes)
2. **Lookup Complexity**: O(n*m) worst case for OperationsForPackage (n=operations, m=IDs)
3. **Pipeline Complexity**: Requires tracking which operations come from which package

### Mitigation

- Memory overhead is negligible (typical plan has <100 operations)
- Lookup can be optimized with caching if needed
- Pipeline tracking logic is straightforward with clear data flow

## Consequences

### Positive

- ✅ Accurate manifest tracking
- ✅ Unmanage works correctly
- ✅ Doctor reports accurate statistics  
- ✅ Status shows correct file counts
- ✅ Enables incremental remanage (Phase 22.5)
- ✅ Foundation for parallel operations

### Negative

- ❌ Slightly more complex Plan type
- ❌ Pipeline must track operation provenance
- ❌ Existing plans need migration (handled via omitempty)

### Neutral

- Plan serialization unchanged (PackageOperations can be omitted)
- Backward compatible (old code ignores new field)
- Forward compatible (new code handles missing field)

## Migration Path

### For Existing Code

Old code without PackageOperations:
```go
// Still works - PackageOperations is optional
plan := dot.Plan{
    Operations: ops,
    Metadata:   meta,
}
```

New code with PackageOperations:
```go
plan := dot.Plan{
    Operations:        ops,
    Metadata:          meta,
    PackageOperations: pkgOps,  // NEW
}
```

### For Tests

Tests can verify package mapping:
```go
func TestPlan_PackageOperations(t *testing.T) {
    plan := createTestPlan()
    vimOps := plan.OperationsForPackage("vim")
    assert.Len(t, vimOps, 3)
}
```

## Success Criteria

- [ ] Plan type extended with PackageOperations
- [ ] Pipeline populates PackageOperations correctly
- [ ] Manifest shows accurate LinkCount and Links
- [ ] All existing tests pass
- [ ] New integration tests verify accuracy
- [ ] No breaking changes to public API
- [ ] Test coverage ≥ 80%

## References

- **Phase 22 Plan**: `docs/planning/phase-22-complete-stubs-plan.md`
- **Manifest Types**: `internal/manifest/types.go`
- **Plan Type**: `pkg/dot/plan.go`
- **Pipeline**: `internal/pipeline/packages.go`

## Notes

This decision enables Phase 22.5 (Incremental Remanage) by providing the foundation for tracking which files belong to which packages. Future enhancements could include:

- Parallel package installation (operations grouped by package)
- Selective remanage (only changed packages)
- Package dependency tracking
- Cross-package conflict detection

## Review

**Reviewed by**: [Development Team]  
**Approved by**: [Tech Lead]  
**Implementation**: Phase 22.2 (Tasks T22.2-001 through T22.2-005)

