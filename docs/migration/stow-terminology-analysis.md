# Stow Terminology Analysis

## Overview

This document summarizes the analysis of "stow" terminology usage in the dot project codebase and provides the rationale for the comprehensive refactoring plan.

## Current State

### Terminology Discovery

Search for "stow" (case-insensitive) revealed **779 total occurrences** across the codebase:

**By Type**:
- `StowDir` (field/variable): 170 occurrences in 37 files
- `StowPath` (type): 29 occurrences in 4 files
- `stowDir` (local variable): 42 occurrences in 9 files
- `StowPipeline`, `StowInput`, etc.: Multiple occurrences in pipeline layer
- Comments and documentation: Hundreds of additional references

**By Layer**:

| Layer | Files Affected | Key Items |
|-------|----------------|-----------|
| Domain (`pkg/dot/`) | 6 files | Config.StowDir field, documentation |
| Configuration (`internal/config/`) | 8 files | DirectoriesConfig.Stow, loaders, writers |
| Pipeline (`internal/pipeline/`) | 4 files | StowPipeline, StowInput, stow.go |
| API (`internal/api/`) | 13 files | Client implementations, tests |
| CLI (`cmd/dot/`) | 6 files | Global config, flags, help text |
| CLI Subsystems (`internal/cli/`) | 6 files | Error contexts, suggestions, examples |
| Documentation (`docs/`) | 13 files | Architecture, configuration, plans |

## Established Project Terminology

### From `docs/TERMINOLOGY.md`

The project has **explicitly documented** its terminology choices:

**Commands**:
- `manage` (not "stow") - Bring package under management
- `unmanage` (not "unstow") - Remove package from management  
- `remanage` (not "restow") - Update managed package
- `adopt` - Import existing files

**Rationale Documented**:
1. **Semantic Clarity**: "Manage" describes purpose (managing dotfiles)
2. **Implementation Agnostic**: Doesn't reveal linking mechanism
3. **User-Focused**: Users "manage configurations" not "stow files"
4. **Professional**: Matches modern infrastructure tools (Ansible, Chef)
5. **Scalable**: Supports future features (templates, encryption, remote packages)

### Comparison with GNU Stow

| GNU Stow Term | dot Term | Status |
|---------------|----------|--------|
| stow | manage | ✓ Implemented |
| unstow | unmanage | ✓ Implemented |
| restow | remanage | ✓ Implemented |
| stow directory | **package directory** | ✗ Not fully implemented |
| $STOW_DIR | ??? | ✗ Not addressed |

**Gap Identified**: Commands use new terminology, but **directory naming remains old**.

## Problem Statement

### Inconsistency Between Commands and Configuration

**Commands use new terminology**:
```bash
dot manage vim      # ✓ Correct terminology
dot unmanage vim    # ✓ Correct terminology
dot remanage vim    # ✓ Correct terminology
```

**But configuration still uses old terminology**:
```yaml
directories:
  stow: ~/dotfiles    # ✗ Inconsistent with "manage"
  target: ~
```

**And code uses old terminology**:
```go
type Config struct {
    StowDir   string    // ✗ Should be PackageDir
    TargetDir string
}
```

### User Confusion

This inconsistency confuses users:

1. **Command implies one thing**: "I'm managing packages"
2. **Config says another**: "This is a stow directory"
3. **Mental model conflict**: "Is this a GNU Stow wrapper or independent tool?"

### Maintenance Issues

Developers face similar confusion:

- Which terminology to use in new code?
- What does "stow" mean in this context?
- Is there a dependency on GNU Stow?
- Why do commands not match the configuration?

## Terminology Decision

### Chosen Replacement: `PackageDir`

**Decision**: Use `PackageDir` (not `SourceDir`)

**Rationale**:

1. **Consistency with existing types**: The codebase already uses `PackagePath`
   ```go
   type PackagePath struct { ... }  // Already exists
   ```

2. **Semantic accuracy**: The directory contains **packages**, each of which contains dotfiles
   ```
   ~/dotfiles/           # Package directory
   ├── vim/             # Package
   │   └── .vimrc
   ├── zsh/             # Package
   │   └── .zshrc
   └── git/             # Package
       └── .gitconfig
   ```

3. **Clear relationship**: `PackageDir` → contains → `Package` → contains → files

4. **Descriptive**: Immediately clear what the directory contains

5. **Matches project terminology**: Aligns with "manage packages" command language

### Alternative Considered: `SourceDir`

**Pros**:
- Shorter, more generic
- Clear opposition to "target"
- Common in other tools

**Cons**:
- Less specific about contents
- Doesn't match existing `PackagePath` type
- Could be confused with source code directory

**Verdict**: Rejected in favor of `PackageDir` for consistency

## Scope of Refactoring

### Breaking Changes Required

**Public API** (`pkg/dot/Config`):
```go
// Before
type Config struct {
    StowDir   string    // Public field
    TargetDir string
    // ...
}

// After
type Config struct {
    PackageDir string   // Breaking change
    TargetDir  string
    // ...
}
```

**Configuration Keys**:
```yaml
# Before
directories:
  stow: ~/dotfiles     # Old key

# After
directories:
  packages: ~/dotfiles # New key
```

**Environment Variables**:
```bash
# Before
export DOT_DIRECTORIES_STOW=/path/to/dotfiles

# After
export DOT_DIRECTORIES_PACKAGES=/path/to/dotfiles
```

### Internal Changes (Non-Breaking)

**Pipeline Types**:
- `StowPipeline` → `ManagePipeline`
- `StowInput` → `ManageInput`
- `StowPipelineOpts` → `ManagePipelineOpts`

**File Names**:
- `internal/pipeline/stow.go` → `manage.go`
- `internal/pipeline/stow_test.go` → `manage_test.go`

**Test Generators**:
- `genStowPath()` → `genSourcePath()` or `genPackagePath()`

## Migration Strategy

### Three-Phase Approach

**Phase 1: Internal Implementation (Non-Breaking)**
- Add new fields alongside old ones
- Mark old fields as deprecated
- Support both configuration keys
- Emit deprecation warnings
- Update all internal code

**Phase 2: Documentation Update**
- Update all documentation
- Create migration guide
- Update examples
- Clarify terminology

**Phase 3: Public API Cleanup (Breaking)**
- Remove deprecated fields
- Remove old configuration keys
- Remove compatibility layer
- Major version bump (v0.2.0 or v1.0.0)

### Backward Compatibility During Transition

**Configuration Loader**:
```go
// Support both keys during transition
if v.IsSet("directories.packages") {
    cfg.Packages = v.GetString("directories.packages")
} else if v.IsSet("directories.stow") {
    cfg.Packages = v.GetString("directories.stow")
    logger.Warn("directories.stow is deprecated, use directories.packages")
}
```

**Environment Variables**:
```go
v.BindEnv("directories.packages", "DOT_DIRECTORIES_PACKAGES")
v.BindEnv("directories.stow", "DOT_DIRECTORIES_STOW")  // Legacy
```

**Config Struct**:
```go
type Config struct {
    // Deprecated: Use PackageDir instead. Will be removed in v0.2.0
    StowDir    string
    
    PackageDir string
    TargetDir  string
}
```

## Impact Analysis

### Files Requiring Changes

**Critical Path** (37 files with `StowDir`):
1. Domain layer: 6 files
2. Configuration system: 8 files
3. Pipeline system: 4 files
4. API layer: 13 files
5. CLI layer: 6 files

**Documentation** (13+ files):
1. Main docs: README.md, Configuration.md, Architecture.md
2. Plan docs: 8 phase plan documents
3. Completion docs: 2 phase completion documents

**Total Estimated**: 50+ files requiring updates

### Test Coverage Impact

**Existing Tests Must Pass**:
- All 100+ test files must continue to pass
- Configuration loading tests (both old and new keys)
- Pipeline execution tests
- API integration tests
- CLI command tests

**New Tests Required**:
- Deprecation warning tests
- Backward compatibility tests
- Migration tests

## Risks and Mitigations

### Risk: Breaking User Configurations

**Impact**: Users with existing `~/.config/dot/config.yaml` files

**Mitigation**:
1. Support both keys during transition (Phase 1-2)
2. Emit clear deprecation warnings
3. Provide migration tool: `dot config migrate`
4. Document migration in release notes
5. Wait for at least one minor release before breaking change

### Risk: Third-Party Code Breakage

**Impact**: Any code importing `pkg/dot` and using `Config.StowDir`

**Mitigation**:
1. Deprecation warnings in Go docs
2. Compatibility period with both fields
3. Clear communication in CHANGELOG
4. Semantic versioning (major bump for breaking change)

### Risk: Incomplete Refactoring

**Impact**: Missing some occurrences leaves inconsistency

**Mitigation**:
1. Comprehensive search for all variants (StowDir, stowDir, stow, STOW)
2. Detailed checklist in refactoring plan
3. Automated checks in CI for "stow" occurrences
4. Code review before merge

## Alignment with Project Standards

### Constitutional Compliance

**Test-Driven Development**: 
- ✓ Existing tests define behavior
- ✓ Refactoring must not break tests
- ✓ New deprecation behavior requires tests

**Atomic Commits**:
- Each layer can be committed separately
- Each commit leaves codebase in working state
- Clear commit messages using Conventional Commits

**Functional Programming Preference**:
- Refactoring preserves functional structure
- Pure functions remain pure
- Side effects remain in shell layer

**Academic Documentation**:
- ✓ Documentation updated to reflect changes
- ✓ No hyperbole, factual descriptions
- ✓ Clear rationale provided

### Code Quality Gates

All gates must pass after refactoring:
- ✓ `make test` - All tests pass
- ✓ `make lint` - No linting errors
- ✓ `make coverage` - 80% coverage maintained
- ✓ `make check` - Complete quality validation

## Recommendation

**Proceed with refactoring** using the comprehensive plan in `Phase-21-Stow-Terminology-Refactor-Plan.md`:

1. **Strong Rationale**: Terminology documented in TERMINOLOGY.md but not implemented
2. **User Benefit**: Eliminates confusion between commands and configuration
3. **Maintainability**: Improves code clarity and consistency
4. **Professional Image**: Aligns with modern tooling standards
5. **Phased Approach**: Minimizes risk with backward compatibility
6. **Clear Migration Path**: Users have smooth upgrade experience

**Timeline**: 13-19 hours of focused development across three phases

**Version Strategy**:
- v0.1.0: Ship Phase 1-2 (backward compatible, deprecation warnings)
- v0.2.0: Ship Phase 3 (breaking changes, remove deprecated code)

## Conclusion

The analysis reveals significant terminology inconsistency: commands use modern "manage" terminology while configuration and code retain legacy "stow" terminology. The documented rationale in TERMINOLOGY.md supports complete migration to "package directory" terminology for semantic clarity and professional presentation.

The recommended three-phase approach balances completeness with risk mitigation, providing users a smooth migration path while achieving full terminology alignment.

---

**Next Steps**:
1. Review this analysis and refactoring plan
2. Approve terminology choice (PackageDir)
3. Create feature branch
4. Execute Phase 1 (Internal Implementation)
5. Execute Phase 2 (Documentation)
6. Release v0.1.0 with deprecation warnings
7. After feedback period, execute Phase 3
8. Release v0.2.0 with breaking changes

