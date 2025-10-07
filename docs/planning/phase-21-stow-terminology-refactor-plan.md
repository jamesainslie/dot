# Phase 21: Stow Terminology Refactoring Plan

## Executive Summary

This phase eliminates all remnants of GNU Stow terminology from the codebase, aligning with the project's established naming conventions documented in `docs/TERMINOLOGY.md`. The project uses semantic, implementation-agnostic terms: `manage`, `unmanage`, `remanage` rather than `stow`, `unstow`, `restow`.

## Problem Statement

Despite having established terminology documented in `TERMINOLOGY.md`, the codebase contains approximately 779 occurrences of "stow" terminology across 37+ files:

- **170 occurrences** of `StowDir`
- **29 occurrences** of `StowPath`
- **42 occurrences** of `stowDir`
- **Hundreds more** in comments, documentation, configuration keys, test names, and pipeline names

This creates semantic inconsistency and confuses users about whether the tool implements GNU Stow or uses its own approach.

## Rationale for Terminology Change

From `docs/TERMINOLOGY.md`:

1. **Semantic Clarity**: "Package directory" describes what it contains (packages)
2. **Implementation Agnostic**: Doesn't reveal internal mechanism (linking vs copying)
3. **User-Focused**: Users manage packages from a source, not "stow" them
4. **Professional**: Matches terminology from modern infrastructure tools
5. **Scalable**: Works with future features (remote packages, templates, encryption)

## Terminology Mapping

### Primary Replacements

| Old Term | New Term | Rationale |
|----------|----------|-----------|
| `stow directory` | `package directory` | Describes content (packages) |
| `StowDir` | `PackageDir` | Consistent with PackagePath |
| `stowDir` | `packageDir` | Lowercase variant |
| `directories.stow` | `directories.packages` | Configuration key alignment |
| `DOT_DIRECTORIES_STOW` | `DOT_DIRECTORIES_PACKAGES` | Environment variable alignment |
| `StowPipeline` | `ManagePipeline` | Aligns with `manage` command |
| `StowInput` | `ManageInput` | Aligns with pipeline name |
| `StowPath` | `SourcePath` | Generic source location |
| `genStowPath()` | `genSourcePath()` | Test generator alignment |

### Alternative Considerations

**Option A: PackageDir** (Recommended)
- Pro: Matches `PackagePath` type already in use
- Pro: Describes what the directory contains
- Pro: Consistent with `packages` terminology throughout codebase
- Con: Slightly verbose

**Option B: SourceDir**
- Pro: Shorter, more generic
- Pro: Clear opposition to "target"
- Con: Less specific about containing packages
- Con: Doesn't match existing `PackagePath` type

**Decision**: Use `PackageDir` for consistency with existing `PackagePath` type.

## Scope of Changes

### 1. Core Domain Types (`pkg/dot/`)

**Files to Update**:
- `pkg/dot/config.go` - Config struct field
- `pkg/dot/config_test.go` - Test configurations
- `pkg/dot/config_validation_test.go` - Validation tests
- `pkg/dot/config_missing_validation_test.go` - Missing field tests
- `pkg/dot/client.go` - Documentation comments
- `pkg/dot/doc.go` - Package documentation

**Changes**:
```go
// Before
type Config struct {
    StowDir   string
    TargetDir string
    // ...
}

// After
type Config struct {
    PackageDir string
    TargetDir  string
    // ...
}
```

**Impact**: BREAKING CHANGE - Public API modification

### 2. Configuration System (`internal/config/`)

**Files to Update**:
- `internal/config/extended.go` - DirectoriesConfig struct
- `internal/config/extended_test.go` - Configuration tests
- `internal/config/loader.go` - Viper key bindings
- `internal/config/loader_test.go` - Loader tests
- `internal/config/writer.go` - Configuration serialization
- `internal/config/writer_test.go` - Writer tests

**Changes**:
```go
// Before
type DirectoriesConfig struct {
    Stow     string `mapstructure:"stow" json:"stow" yaml:"stow" toml:"stow"`
    Target   string `mapstructure:"target"`
    Manifest string `mapstructure:"manifest"`
}

// After
type DirectoriesConfig struct {
    Packages string `mapstructure:"packages" json:"packages" yaml:"packages" toml:"packages"`
    Target   string `mapstructure:"target"`
    Manifest string `mapstructure:"manifest"`
}
```

**Migration Strategy**:
- Support both `stow` and `packages` keys during transition
- Deprecation warning when `stow` is used
- Migration tool to update configuration files
- Documentation of breaking change

### 3. Pipeline System (`internal/pipeline/`)

**Files to Update**:
- `internal/pipeline/stow.go` → rename to `manage.go`
- `internal/pipeline/stow_test.go` → rename to `manage_test.go`
- `internal/pipeline/stages.go` - Function references
- `internal/pipeline/stages_test.go` - Test references

**Changes**:
```go
// Before
type StowPipelineOpts struct {
    FS        dot.FS
    IgnoreSet *ignore.IgnoreSet
    Policies  planner.ResolutionPolicies
}

type StowInput struct {
    StowDir   dot.PackagePath
    TargetDir dot.TargetPath
    Packages  []string
}

type StowPipeline struct {
    opts StowPipelineOpts
}

func NewStowPipeline(opts StowPipelineOpts) *StowPipeline
func (p *StowPipeline) Execute(ctx context.Context, input StowInput) dot.Result[dot.Plan]

// After
type ManagePipelineOpts struct {
    FS        dot.FS
    IgnoreSet *ignore.IgnoreSet
    Policies  planner.ResolutionPolicies
}

type ManageInput struct {
    PackageDir dot.PackagePath
    TargetDir  dot.TargetPath
    Packages   []string
}

type ManagePipeline struct {
    opts ManagePipelineOpts
}

func NewManagePipeline(opts ManagePipelineOpts) *ManagePipeline
func (p *ManagePipeline) Execute(ctx context.Context, input ManageInput) dot.Result[dot.Plan]
```

**Rationale**: Pipeline implements the `manage` command logic

### 4. API Layer (`internal/api/`)

**Files to Update**:
- `internal/api/client_test.go` - Client tests
- `internal/api/manage.go` - Manage implementation
- `internal/api/adopt.go` - Adopt implementation
- `internal/api/adopt_test.go` - Adopt tests
- `internal/api/adopt_edge_test.go` - Edge case tests
- `internal/api/coverage_boost_test.go` - Coverage tests
- `internal/api/doctor_comprehensive_test.go` - Doctor tests
- `internal/api/doctor_edge_test.go` - Doctor edge tests
- `internal/api/doctor_test.go` - Doctor tests
- `internal/api/error_paths_test.go` - Error tests
- `internal/api/execution_test.go` - Execution tests
- `internal/api/final_coverage_test.go` - Coverage tests

**Changes**: Update all references to `StowDir` in Config structs and assertions

### 5. CLI Layer (`cmd/dot/`)

**Files to Update**:
- `cmd/dot/root.go` - Global config and flag definition
- `cmd/dot/commands_test.go` - Command tests
- `cmd/dot/helpers_test.go` - Helper tests
- `cmd/dot/config.go` - Config command
- `cmd/dot/config_test.go` - Config command tests

**Changes**:
```go
// Before
type globalConfig struct {
    stowDir   string
    targetDir string
    // ...
}

rootCmd.PersistentFlags().StringVarP(&globalCfg.stowDir, "dir", "d", ".",
    "Stow directory containing packages")

// After
type globalConfig struct {
    packageDir string
    targetDir  string
    // ...
}

rootCmd.PersistentFlags().StringVarP(&globalCfg.packageDir, "dir", "d", ".",
    "Package directory containing packages")
```

**Flag Considerations**:
- Keep `-d, --dir` flag for backward compatibility
- Update help text to say "Package directory"
- Consider adding `--package-dir` alias for clarity

### 6. CLI Error Handling (`internal/cli/errors/`)

**Files to Update**:
- `internal/cli/errors/context.go` - ErrorContext struct
- `internal/cli/errors/context_test.go` - Context tests
- `internal/cli/errors/formatter_test.go` - Formatter tests
- `internal/cli/errors/suggestions.go` - Error suggestions
- `internal/cli/errors/suggestions_test.go` - Suggestion tests

**Changes**: Update error context and suggestion messages

### 7. CLI Help System (`internal/cli/help/`)

**Files to Update**:
- `internal/cli/help/examples.go` - Example commands
- `internal/cli/help/completion_test.go` - Completion tests

**Changes**: Update example descriptions and flag references

### 8. Documentation (`docs/`)

**Files to Update**:
- `docs/Configuration.md` - Configuration reference
- `docs/Features.md` - Feature descriptions
- `docs/Architecture.md` - Architecture diagrams and descriptions
- `docs/Implementation-Plan.md` - Historical implementation notes
- `docs/Phase-9-Plan.md` - Phase 9 historical notes
- `docs/Phase-11-Plan.md` - Phase 11 historical notes
- `docs/Phase-12-Plan.md` - Phase 12 historical notes
- `docs/Phase-15b-Plan.md` - Configuration system notes
- `docs/Phase-16-Plan.md` - Property testing notes
- `docs/Phase-18-Plan.md` - Historical notes
- `README.md` - Main project documentation

**Changes**: Update all references to "stow" terminology

### 9. Phase Completion Documents

**Files to Update**:
- `PHASE_9_COMPLETE.md` - Historical record
- `PHASE_12_COMPLETE.md` - Historical record

**Changes**: Note: These are historical records and may be left as-is or updated with clarifying notes

### 10. Test Fixtures and Data

**Files to Search**:
- All test files with hardcoded paths containing "stow"
- Property test generators with "stow" in names

**Changes**: Update test data to use "packages" terminology

## Migration Strategy

### Phase 1: Internal Implementation (Non-Breaking)

**Goal**: Update all internal code without breaking public API

1. **Add New Fields with Deprecation**:
   ```go
   type Config struct {
       // Deprecated: Use PackageDir instead
       StowDir    string
       PackageDir string
       // ...
   }
   ```

2. **Configuration Compatibility Layer**:
   ```go
   // Loader supports both keys, prefers "packages"
   if v.IsSet("directories.packages") {
       cfg.Packages = v.GetString("directories.packages")
   } else if v.IsSet("directories.stow") {
       cfg.Packages = v.GetString("directories.stow")
       logger.Warn("directories.stow is deprecated, use directories.packages")
   }
   ```

3. **Environment Variable Support**:
   ```go
   v.BindEnv("directories.packages", "DOT_DIRECTORIES_PACKAGES")
   // Legacy support
   v.BindEnv("directories.stow", "DOT_DIRECTORIES_STOW")
   ```

4. **Update Internal Code**:
   - Pipeline system rename
   - Internal API references
   - Test code

**Deliverable**: All internal code uses new terminology, backward compatibility maintained

### Phase 2: Documentation Update

**Goal**: Update all documentation to reflect new terminology

1. Update `docs/TERMINOLOGY.md` with terminology mappings
2. Update `README.md` with new configuration examples
3. Update `docs/Configuration.md` with migration guide
4. Update all phase plan documents
5. Add deprecation notices to relevant sections

**Deliverable**: Complete documentation using new terminology

### Phase 3: Public API Update (Breaking Change)

**Goal**: Remove deprecated fields and complete migration

**Prerequisites**:
- Phase 1 and 2 complete
- At least one release with deprecation warnings
- Migration tool available

**Changes**:
1. Remove `StowDir` from `pkg/dot/Config`
2. Keep only `PackageDir`
3. Remove `directories.stow` support from config loader
4. Remove legacy environment variable support
5. Update version to indicate breaking change (v0.2.0 or v1.0.0)

**Migration Tool**:
```bash
# Automatic config migration
dot config migrate

# Dry-run to preview changes
dot config migrate --dry-run
```

**Deliverable**: Clean public API with no deprecated terminology

## Implementation Checklist

### Pre-Implementation

- [ ] Review this plan with maintainers
- [ ] Confirm terminology choice (`PackageDir` vs `SourceDir`)
- [ ] Decide on version number strategy (major/minor bump)
- [ ] Create feature branch: `refactor-stow-terminology`

### Phase 1: Internal Implementation

#### Core Domain
- [ ] Update `pkg/dot/config.go` - Add `PackageDir` field
- [ ] Update `pkg/dot/config.go` - Deprecate `StowDir` field
- [ ] Update `pkg/dot/config.go` - Add compatibility logic
- [ ] Update `pkg/dot/config_test.go`
- [ ] Update `pkg/dot/config_validation_test.go`
- [ ] Update `pkg/dot/config_missing_validation_test.go`
- [ ] Update `pkg/dot/client.go` documentation
- [ ] Update `pkg/dot/doc.go` package docs

#### Configuration System
- [ ] Update `internal/config/extended.go` - Add `Packages` field
- [ ] Update `internal/config/extended.go` - Deprecate `Stow` field
- [ ] Update `internal/config/loader.go` - Support both keys
- [ ] Update `internal/config/loader.go` - Add deprecation warning
- [ ] Update `internal/config/writer.go` - Write new key
- [ ] Update `internal/config/writer_test.go`
- [ ] Update `internal/config/loader_test.go`
- [ ] Update `internal/config/extended_test.go`

#### Pipeline System
- [ ] Rename `internal/pipeline/stow.go` to `manage.go`
- [ ] Rename `internal/pipeline/stow_test.go` to `manage_test.go`
- [ ] Update `StowPipeline` → `ManagePipeline`
- [ ] Update `StowPipelineOpts` → `ManagePipelineOpts`
- [ ] Update `StowInput` → `ManageInput`
- [ ] Update `StowInput.StowDir` → `ManageInput.PackageDir`
- [ ] Update `NewStowPipeline` → `NewManagePipeline`
- [ ] Update `internal/pipeline/stages.go` references
- [ ] Update `internal/pipeline/stages_test.go` references

#### API Layer
- [ ] Update `internal/api/manage.go` references
- [ ] Update `internal/api/adopt.go` references
- [ ] Update `internal/api/client_test.go` test configs
- [ ] Update `internal/api/adopt_test.go` test configs
- [ ] Update `internal/api/adopt_edge_test.go` test configs
- [ ] Update `internal/api/coverage_boost_test.go` test configs
- [ ] Update `internal/api/doctor_comprehensive_test.go` test configs
- [ ] Update `internal/api/doctor_edge_test.go` test configs
- [ ] Update `internal/api/doctor_test.go` test configs
- [ ] Update `internal/api/error_paths_test.go` test configs
- [ ] Update `internal/api/execution_test.go` test configs
- [ ] Update `internal/api/final_coverage_test.go` test configs

#### CLI Layer
- [ ] Update `cmd/dot/root.go` - Rename `stowDir` → `packageDir`
- [ ] Update `cmd/dot/root.go` - Update flag help text
- [ ] Update `cmd/dot/root.go` - Update `buildConfig()` function
- [ ] Update `cmd/dot/commands_test.go` test configurations
- [ ] Update `cmd/dot/helpers_test.go` test helpers
- [ ] Update `cmd/dot/config.go` key references
- [ ] Update `cmd/dot/config_test.go` test cases

#### CLI Subsystems
- [ ] Update `internal/cli/errors/context.go` - ErrorConfig struct
- [ ] Update `internal/cli/errors/context_test.go` tests
- [ ] Update `internal/cli/errors/formatter_test.go` tests
- [ ] Update `internal/cli/errors/suggestions.go` messages
- [ ] Update `internal/cli/errors/suggestions_test.go` tests
- [ ] Update `internal/cli/help/examples.go` descriptions
- [ ] Update `internal/cli/help/completion_test.go` tests

#### Testing
- [ ] Run `make test` - Ensure all tests pass
- [ ] Run `make lint` - Ensure no linting errors
- [ ] Run `make check` - Full quality check
- [ ] Verify backward compatibility with old config
- [ ] Verify deprecation warnings appear

### Phase 2: Documentation

- [ ] Update `README.md` - Replace "stow" terminology
- [ ] Update `docs/TERMINOLOGY.md` - Add migration section
- [ ] Update `docs/Configuration.md` - Configuration examples
- [ ] Update `docs/Configuration.md` - Add migration guide
- [ ] Update `docs/Architecture.md` - Architectural descriptions
- [ ] Update `docs/Features.md` - Feature descriptions
- [ ] Update `docs/Implementation-Plan.md` - Historical notes
- [ ] Update `docs/Phase-9-Plan.md` - Historical notes
- [ ] Update `docs/Phase-11-Plan.md` - Historical notes
- [ ] Update `docs/Phase-12-Plan.md` - Historical notes
- [ ] Update `docs/Phase-15b-Plan.md` - Configuration notes
- [ ] Update `docs/Phase-16-Plan.md` - Property test notes
- [ ] Update `docs/Phase-18-Plan.md` - Historical notes
- [ ] Add `docs/Migration-Stow-to-Packages.md` - Migration guide

### Phase 3: Public API Cleanup (Future Major Version)

- [ ] Remove `StowDir` field from `pkg/dot/Config`
- [ ] Remove `directories.stow` support from loader
- [ ] Remove `DOT_DIRECTORIES_STOW` environment variable
- [ ] Remove compatibility layer
- [ ] Remove deprecation warnings
- [ ] Update CHANGELOG.md with breaking changes
- [ ] Tag release with major version bump

## Testing Strategy

### Unit Tests

All existing tests must pass after refactoring:
- Configuration loading with both old and new keys
- Pipeline execution with new types
- Error messages with new terminology
- CLI flag parsing

### Integration Tests

- [ ] Test configuration file with `directories.stow` (backward compat)
- [ ] Test configuration file with `directories.packages` (new)
- [ ] Test environment variable `DOT_DIRECTORIES_STOW` (backward compat)
- [ ] Test environment variable `DOT_DIRECTORIES_PACKAGES` (new)
- [ ] Test deprecation warnings appear correctly
- [ ] Test CLI commands work with new config

### Property Tests

- [ ] Update `docs/Phase-16-Plan.md` property generators
- [ ] Update `genStowPath()` → `genSourcePath()`
- [ ] Ensure property tests pass with new types

## Risk Assessment

### High Risk

**Breaking Public API** (`pkg/dot/Config`)
- **Mitigation**: Phased approach with deprecation period
- **Impact**: Requires major version bump, user migration

**Configuration File Changes**
- **Mitigation**: Support both old and new keys with warnings
- **Impact**: Users need to update config files

### Medium Risk

**Pipeline Rename** (Internal API)
- **Mitigation**: Internal change only, no public impact
- **Impact**: Requires careful refactoring of internal references

**Test Data Updates**
- **Mitigation**: Comprehensive test suite ensures correctness
- **Impact**: Time-consuming but straightforward

### Low Risk

**Documentation Updates**
- **Mitigation**: Clear before/after examples
- **Impact**: Improves clarity and consistency

**Comment Updates**
- **Mitigation**: Search-and-replace with review
- **Impact**: Minimal functional impact

## Success Criteria

1. **Zero occurrences of "stow" terminology** in active codebase (excluding historical documents)
2. **All tests pass** with 100% of existing coverage maintained
3. **Backward compatibility maintained** during Phase 1 and 2
4. **Documentation consistency** - all docs use new terminology
5. **Clear migration path** documented for users
6. **Deprecation warnings** guide users to new configuration

## Timeline Estimate

- **Phase 1 (Internal Implementation)**: 8-12 hours
  - Core domain updates: 2 hours
  - Configuration system: 2 hours
  - Pipeline refactoring: 2 hours
  - API and CLI updates: 3 hours
  - Testing and validation: 2-3 hours

- **Phase 2 (Documentation)**: 3-4 hours
  - Documentation updates: 2 hours
  - Migration guide creation: 1-2 hours

- **Phase 3 (Public API Cleanup)**: 2-3 hours
  - Remove deprecated code: 1 hour
  - Final testing: 1 hour
  - Release preparation: 1 hour

**Total**: 13-19 hours of focused development time

## Rollout Plan

### Release v0.1.0 (Current - No Breaking Changes)

- Ship Phase 1 and Phase 2
- Deprecation warnings active
- Both old and new terminology supported
- Documentation updated

### Release v0.2.0 (Breaking Changes)

- Ship Phase 3
- Remove deprecated terminology
- Breaking change: `StowDir` removed
- Migration tool included

### Communication

1. **Changelog Entry**:
   ```markdown
   ## [v0.1.0] - YYYY-MM-DD
   
   ### Changed
   - Configuration key `directories.stow` deprecated in favor of `directories.packages`
   - Environment variable `DOT_DIRECTORIES_STOW` deprecated in favor of `DOT_DIRECTORIES_PACKAGES`
   - All internal terminology updated to use "package directory" instead of "stow directory"
   
   ### Added
   - Backward compatibility support for old configuration keys
   - Deprecation warnings when using old terminology
   - Migration guide in docs/Migration-Stow-to-Packages.md
   ```

2. **Release Notes**:
   - Highlight terminology change
   - Link to migration guide
   - Emphasize backward compatibility
   - Note future breaking change in v0.2.0

3. **Documentation Updates**:
   - Prominent notice in README
   - Migration guide with examples
   - Updated all configuration examples

## References

- `docs/TERMINOLOGY.md` - Official terminology rationale
- `docs/Configuration.md` - Configuration documentation
- Project Constitution - Atomic commits and TDD requirements
- Conventional Commits specification

## Conclusion

This refactoring eliminates the last remnants of GNU Stow terminology, bringing the codebase into full alignment with the project's documented terminology standards. The phased approach ensures backward compatibility during migration while providing a clear path to a clean, semantically consistent API.

The terminology change improves:
- **Clarity**: "Package directory" is self-explanatory
- **Consistency**: Aligns with `manage`, `unmanage`, `remanage` commands
- **Professionalism**: Matches modern infrastructure tooling conventions
- **Maintainability**: Removes confusion about GNU Stow relationship

