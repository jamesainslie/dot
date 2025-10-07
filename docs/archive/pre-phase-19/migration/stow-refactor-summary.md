# Stow Terminology Refactor - Quick Reference

## The Problem in One Sentence

The CLI uses "manage/unmanage/remanage" commands but the configuration and codebase still use legacy "stow" terminology, creating semantic inconsistency and user confusion.

## Key Findings

### What We Found
- **779 total occurrences** of "stow" terminology
- **170 occurrences** of `PackageDir` field/variable
- **50+ files** require updates
- **Inconsistency**: Commands say "manage", config says "stow"

### What Should Change

| Current | Target | Type |
|---------|--------|------|
| `PackageDir` | `PackageDir` | Field name |
| `packageDir` | `packageDir` | Variable name |
| `directories.stow` | `directories.packages` | Config key |
| `DOT_DIRECTORIES_STOW` | `DOT_DIRECTORIES_PACKAGES` | Env var |
| `StowPipeline` | `ManagePipeline` | Type name |
| `StowInput` | `ManageInput` | Type name |
| `stow.go` | `manage.go` | File name |

## Why "PackageDir"?

1. **Consistent**: Matches existing `PackagePath` type
2. **Semantic**: Directory contains packages
3. **Clear**: Self-documenting what it holds
4. **Aligned**: Matches "manage packages" command language

## Three-Phase Plan

### Phase 1: Internal (Non-Breaking) - 8-12 hours
- Add `PackageDir` alongside `PackageDir` (deprecated)
- Support both `directories.stow` and `directories.packages`
- Update all internal code to use new names
- Emit deprecation warnings
- **Result**: Backward compatible, warnings active

### Phase 2: Documentation - 3-4 hours
- Update all docs to new terminology
- Create migration guide
- Update examples
- **Result**: Complete, consistent documentation

### Phase 3: Cleanup (Breaking) - 2-3 hours
- Remove deprecated fields
- Remove old config keys
- Major version bump
- **Result**: Clean API, no legacy terminology

## Migration Strategy

### For Users

**v0.1.0** (Backward Compatible):
```yaml
# Old way still works with warning
directories:
  stow: ~/dotfiles    # Deprecated warning shown
  
# New way preferred
directories:
  packages: ~/dotfiles
```

**v0.2.0** (Breaking Change):
```yaml
# Old way removed
directories:
  packages: ~/dotfiles  # Only this works
```

**Migration Tool**:
```bash
dot config migrate           # Auto-update config file
dot config migrate --dry-run # Preview changes
```

### For Developers

**Public API Change**:
```go
// v0.1.0 - Both fields available
type Config struct {
    PackageDir    string  // Deprecated
    PackageDir string  // Preferred
}

// v0.2.0 - Only new field
type Config struct {
    PackageDir string
}
```

## Files to Update (Organized by Layer)

### Domain Layer (6 files)
- `pkg/dot/config.go`
- `pkg/dot/config_test.go`
- `pkg/dot/config_validation_test.go`
- `pkg/dot/config_missing_validation_test.go`
- `pkg/dot/client.go`
- `pkg/dot/doc.go`

### Configuration System (8 files)
- `internal/config/extended.go`
- `internal/config/extended_test.go`
- `internal/config/loader.go`
- `internal/config/loader_test.go`
- `internal/config/writer.go`
- `internal/config/writer_test.go`

### Pipeline System (4 files)
- `internal/pipeline/stow.go` → rename to `manage.go`
- `internal/pipeline/stow_test.go` → rename to `manage_test.go`
- `internal/pipeline/stages.go`
- `internal/pipeline/stages_test.go`

### API Layer (13 files)
- All files in `internal/api/` with config references

### CLI Layer (6 files)
- All files in `cmd/dot/` with flag/config references

### Documentation (13+ files)
- `README.md`
- `docs/Configuration.md`
- `docs/Architecture.md`
- Plus various phase plans

## Success Criteria

✓ Zero "stow" occurrences in active code
✓ All tests pass with 100% coverage maintained
✓ Backward compatibility during Phase 1-2
✓ Clear migration path documented
✓ Deprecation warnings guide users

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Breaking user configs | Support both keys with warnings |
| Third-party breakage | Deprecation period + semantic versioning |
| Incomplete refactoring | Comprehensive checklist + code review |
| Test failures | Existing tests define behavior |

## Timeline

- **Phase 1**: 8-12 hours (internal refactor)
- **Phase 2**: 3-4 hours (documentation)
- **Phase 3**: 2-3 hours (cleanup)
- **Total**: 13-19 hours

## Next Steps

1. [ ] Review analysis and plan with maintainers
2. [ ] Approve terminology choice (PackageDir)
3. [ ] Create feature branch: `refactor-stow-terminology`
4. [ ] Execute Phase 1 (see detailed checklist in full plan)
5. [ ] Execute Phase 2 (documentation updates)
6. [ ] Release v0.1.0 with deprecation warnings
7. [ ] Gather feedback (1-2 release cycles)
8. [ ] Execute Phase 3 (breaking changes)
9. [ ] Release v0.2.0

## References

- **Full Analysis**: `docs/Stow-Terminology-Analysis.md`
- **Detailed Plan**: `docs/Phase-21-Stow-Terminology-Refactor-Plan.md`
- **Project Terminology**: `docs/TERMINOLOGY.md`

## Quick Decision Points

**Q: Why not keep "stow" since it's established?**
A: Commands already use "manage", creating inconsistency. TERMINOLOGY.md documents the rationale.

**Q: Why PackageDir instead of SourceDir?**
A: Consistency with existing PackagePath type and semantic accuracy.

**Q: Can we do this without breaking changes?**
A: Not completely, but we minimize impact with phased rollout and compatibility layer.

**Q: What about GNU Stow users?**
A: Clear documentation that dot is independent, with documented rationale for different terminology.

**Q: How long until breaking change?**
A: At least one minor release (v0.1.0) with warnings before v0.2.0 breaking change.

