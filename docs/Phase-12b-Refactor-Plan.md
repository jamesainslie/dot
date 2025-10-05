# Phase 12b: Domain Architecture Refactoring - Detailed Plan

## Overview

Phase 12b refactors the codebase from the current architecture (domain types in `pkg/dot/`) to the ideal architecture (domain types in `internal/domain/`). This eliminates the need for the interface-based workaround from Phase 12 and allows a proper Client struct in the public API.

**Status**: Optional future work, to be executed after Phase 12 is stable and in production use.

**Effort**: 12-16 hours of careful, methodical refactoring

**Risk**: High (touches all completed phases) - requires comprehensive testing

**Benefit**: Cleaner architecture, simpler maintenance, better separation of concerns

## Current Architecture (Post-Phase 12)

```
pkg/dot/
├── Domain types:       Operation, Plan, Result, Path, etc.
├── Port interfaces:    FS, Logger, Tracer, Metrics
├── Config types:       Config, LinkMode
├── Client interface:   Client interface definition
└── Support types:      Status, PackageInfo, ExecutionResult

internal/api/
└── Client impl:        Concrete Client implementation

internal/executor/      → imports pkg/dot (domain types)
internal/pipeline/      → imports pkg/dot (domain types)
internal/scanner/       → imports pkg/dot (domain types)
internal/planner/       → imports pkg/dot (domain types)
internal/manifest/      → imports pkg/dot (domain types)
```

**Problem**: Domain types mixed with public API in `pkg/dot/`

## Target Architecture (Post-Phase 12b)

```
internal/domain/
├── Domain types:       Operation, Plan, Result, Path, etc.
├── Port interfaces:    FS, Logger, Tracer, Metrics  
└── Support types:      Status, PackageInfo, ExecutionResult

pkg/dot/
├── Re-exports:         All domain types (type aliases)
├── Config types:       Config, LinkMode (public API concern)
├── Client struct:      Concrete Client implementation (no longer interface)
└── Documentation:      Package docs and examples

internal/executor/      → imports internal/domain (no cycle)
internal/pipeline/      → imports internal/domain (no cycle)
internal/scanner/       → imports internal/domain (no cycle)
internal/planner/       → imports internal/domain (no cycle)
internal/manifest/      → imports internal/domain (no cycle)
```

**Benefit**: `pkg/dot/client.go` can directly import `internal/*` packages without cycles

---

## Detailed Refactoring Steps

### Step 1: Create Migration Branch

**Actions**:
```bash
git checkout -b feature-domain-refactor
git checkout -b feature-domain-refactor-backup  # Safety backup
git checkout feature-domain-refactor
```

**Why**: Isolate potentially disruptive changes from main development branch.

**Time**: 5 minutes

---

### Step 2: Create `internal/domain` Package Structure

**Actions**:
```bash
mkdir -p internal/domain
```

**Create initial package structure**:
```
internal/domain/
├── operation.go       # Operation interface and types
├── operation_test.go
├── plan.go            # Plan type
├── plan_test.go
├── result.go          # Result monad
├── result_test.go
├── path.go            # Phantom-typed paths
├── path_test.go
├── domain.go          # Package, Node, FileTree
├── domain_test.go
├── errors.go          # Error types
├── errors_test.go
├── conflict.go        # Conflict types
├── conflict_test.go
├── ports.go           # Port interfaces
├── ports_test.go
├── execution.go       # ExecutionResult
├── checkpoint.go      # Checkpoint types
└── doc.go             # Package documentation
```

**Tasks**:
- [ ] Create directory structure
- [ ] Add package doc.go with description

**Time**: 10 minutes

**Commit**: `refactor(domain): create internal/domain package structure`

---

### Step 3: Move Domain Types (Atomic Per-Type Migration)

This is the most critical step. Must be done **atomically per type** to maintain buildability.

#### 3.1: Move Result Monad First (Foundation)

**Rationale**: Many types depend on Result, so move it first.

**Actions**:
1. Copy `pkg/dot/result.go` → `internal/domain/result.go`
2. Copy `pkg/dot/result_test.go` → `internal/domain/result_test.go`
3. Update package declaration to `package domain`
4. Run tests: `go test ./internal/domain -v`
5. Create re-export in `pkg/dot/result.go`:
   ```go
   package dot
   
   import "github.com/jamesainslie/dot/internal/domain"
   
   // Result represents a value or an error.
   type Result[T any] = domain.Result[T]
   
   // Result constructors
   var (
       Ok      = domain.Ok
       Err     = domain.Err
       Map     = domain.Map
       FlatMap = domain.FlatMap
       Collect = domain.Collect
       UnwrapOr = domain.UnwrapOr
   )
   ```
6. Run all tests: `go test ./...`
7. Verify no breaking changes

**Verification**:
```bash
# Should still compile and pass
go test ./pkg/dot/...
go test ./internal/executor/...
go test ./internal/pipeline/...
```

**Tasks**:
- [ ] Copy result.go and result_test.go
- [ ] Update package declaration
- [ ] Test in isolation
- [ ] Create re-export
- [ ] Run full test suite
- [ ] Verify zero breaking changes

**Time**: 30 minutes

**Commit**: `refactor(domain): move Result monad to internal/domain`

---

#### 3.2: Move Path Types

**Actions**:
1. Copy `pkg/dot/path.go` → `internal/domain/path.go`
2. Copy `pkg/dot/path_test.go` → `internal/domain/path_test.go`
3. Update package to `package domain`
4. Run tests: `go test ./internal/domain -v`
5. Create re-export in `pkg/dot/path.go`:
   ```go
   package dot
   
   import "github.com/jamesainslie/dot/internal/domain"
   
   // Path types
   type Path[K PathKind] = domain.Path[K]
   type PathKind = domain.PathKind
   type PackagePath = domain.PackagePath
   type TargetPath = domain.TargetPath
   type FilePath = domain.FilePath
   
   // Path constructors
   var (
       NewPackagePath = domain.NewPackagePath
       NewTargetPath  = domain.NewTargetPath
       NewFilePath    = domain.NewFilePath
   )
   ```
6. Run all tests
7. Verify compatibility

**Tasks**:
- [ ] Copy path files
- [ ] Test in isolation
- [ ] Create re-exports
- [ ] Run full test suite
- [ ] Verify zero breaking changes

**Time**: 30 minutes

**Commit**: `refactor(domain): move Path types to internal/domain`

---

#### 3.3: Move Operation Types

**Actions**:
1. Copy `pkg/dot/operation.go` → `internal/domain/operation.go`
2. Copy `pkg/dot/operation_test.go` → `internal/domain/operation_test.go`
3. Update imports within operation.go (Path types now in same package)
4. Test in isolation
5. Create re-export:
   ```go
   package dot
   
   import "github.com/jamesainslie/dot/internal/domain"
   
   // Operation types
   type Operation = domain.Operation
   type OperationID = domain.OperationID
   type OperationKind = domain.OperationKind
   
   type LinkCreate = domain.LinkCreate
   type LinkDelete = domain.LinkDelete
   type DirCreate = domain.DirCreate
   type DirDelete = domain.DirDelete
   type FileMove = domain.FileMove
   type FileBackup = domain.FileBackup
   
   // Operation constructors
   var NewOperationID = domain.NewOperationID
   
   // Operation kinds
   const (
       OperationKindLinkCreate = domain.OperationKindLinkCreate
       OperationKindLinkDelete = domain.OperationKindLinkDelete
       OperationKindDirCreate  = domain.OperationKindDirCreate
       OperationKindDirDelete  = domain.OperationKindDirDelete
       OperationKindFileMove   = domain.OperationKindFileMove
       OperationKindFileBackup = domain.OperationKindFileBackup
   )
   ```
6. Run tests
7. Verify operations work

**Tasks**:
- [ ] Copy operation files
- [ ] Update internal imports
- [ ] Test in isolation
- [ ] Create re-exports with all operation types
- [ ] Run full test suite
- [ ] Verify operation execution still works

**Time**: 45 minutes

**Commit**: `refactor(domain): move Operation types to internal/domain`

---

#### 3.4: Move Remaining Domain Types

**Order of migration** (respecting dependencies):
1. **errors.go** (independent)
2. **domain.go** (Package, Node, FileTree - depends on Path)
3. **plan.go** (depends on Operation, Result)
4. **conflict.go** (depends on Path)
5. **execution.go** (ExecutionResult - depends on OperationID)
6. **checkpoint.go** (depends on Operation, OperationID)

**For each file**:
1. Copy to `internal/domain/`
2. Update package declaration
3. Update imports (other domain types now in same package)
4. Test in isolation
5. Create re-export in `pkg/dot/`
6. Run full test suite
7. Commit atomically

**Example for errors.go**:
```go
// pkg/dot/errors.go (re-export)
package dot

import "github.com/jamesainslie/dot/internal/domain"

type ErrInvalidPath = domain.ErrInvalidPath
type ErrPackageNotFound = domain.ErrPackageNotFound
type ErrConflict = domain.ErrConflict
type ErrCyclicDependency = domain.ErrCyclicDependency
type ErrFilesystemOperation = domain.ErrFilesystemOperation
type ErrPermissionDenied = domain.ErrPermissionDenied
type ErrMultiple = domain.ErrMultiple
type ErrEmptyPlan = domain.ErrEmptyPlan

var UserFacingErrorMessage = domain.UserFacingErrorMessage
```

**Tasks per file**:
- [ ] Copy errors.go and create re-export
- [ ] Copy domain.go and create re-export
- [ ] Copy plan.go and create re-export
- [ ] Copy conflict.go and create re-export
- [ ] Copy execution.go and create re-export
- [ ] Copy checkpoint.go and create re-export

**Time**: 2-3 hours (6 files × 20-30 minutes each)

**Commits**: One atomic commit per type moved

---

### Step 4: Move Port Interfaces

**Actions**:
1. Copy `pkg/dot/ports.go` → `internal/domain/ports.go`
2. Copy `pkg/dot/ports_test.go` → `internal/domain/ports_test.go`
3. Test in isolation
4. Create re-export:
   ```go
   package dot
   
   import "github.com/jamesainslie/dot/internal/domain"
   
   // Port interfaces
   type FS = domain.FS
   type Logger = domain.Logger
   type Tracer = domain.Tracer
   type Metrics = domain.Metrics
   type Span = domain.Span
   type FileInfo = domain.FileInfo
   type DirEntry = domain.DirEntry
   
   // Noop implementations
   var (
       NewNoopTracer  = domain.NewNoopTracer
       NewNoopMetrics = domain.NewNoopMetrics
   )
   ```
5. Run all tests
6. Verify adapters still work

**Tasks**:
- [ ] Copy port files
- [ ] Test in isolation
- [ ] Create re-exports
- [ ] Run adapter tests
- [ ] Run full test suite

**Time**: 30 minutes

**Commit**: `refactor(domain): move port interfaces to internal/domain`

---

### Step 5: Update Internal Package Imports

Now all internal packages need to change imports from `pkg/dot` to `internal/domain`.

#### 5.1: Update import statements

**Files to update** (estimate: 80-100 files):
- `internal/executor/*.go` (8 files)
- `internal/pipeline/*.go` (6 files)
- `internal/scanner/*.go` (6 files)
- `internal/planner/*.go` (12 files)
- `internal/manifest/*.go` (8 files)
- `internal/config/*.go` (2 files)
- `internal/adapters/*.go` (8 files)
- `internal/ignore/*.go` (4 files)

**Approach**: Use automated tool for safety

**Script**:
```bash
#!/bin/bash
# scripts/update-domain-imports.sh

# Find all .go files in internal/ and replace import
find internal -name "*.go" -type f -exec sed -i '' \
    's|"github.com/jamesainslie/dot/pkg/dot"|"github.com/jamesainslie/dot/internal/domain"|g' {} \;

# Verify changes
git diff internal/

# Run tests to verify
go test ./internal/...
```

**Manual verification required for**:
- Files that import both `pkg/dot` and domain types (need to keep pkg/dot for Config, etc.)
- Test files that might need both imports
- Edge cases with qualified vs unqualified imports

**Tasks**:
- [ ] Create import update script
- [ ] Run on internal/executor
- [ ] Test: `go test ./internal/executor/...`
- [ ] Run on internal/pipeline
- [ ] Test: `go test ./internal/pipeline/...`
- [ ] Run on internal/scanner
- [ ] Test: `go test ./internal/scanner/...`
- [ ] Run on internal/planner
- [ ] Test: `go test ./internal/planner/...`
- [ ] Run on internal/manifest
- [ ] Test: `go test ./internal/manifest/...`
- [ ] Run on internal/config
- [ ] Test: `go test ./internal/config/...`
- [ ] Run on internal/adapters
- [ ] Test: `go test ./internal/adapters/...`
- [ ] Run on internal/ignore
- [ ] Test: `go test ./internal/ignore/...`
- [ ] Run full test suite
- [ ] Fix any import issues manually

**Time**: 3-4 hours

**Commits**: One commit per internal package
- `refactor(executor): update imports to use internal/domain`
- `refactor(pipeline): update imports to use internal/domain`
- `refactor(scanner): update imports to use internal/domain`
- `refactor(planner): update imports to use internal/domain`
- `refactor(manifest): update imports to use internal/domain`
- `refactor(config): update imports to use internal/domain`
- `refactor(adapters): update imports to use internal/domain`
- `refactor(ignore): update imports to use internal/domain`

---

### Step 6: Replace Client Interface with Struct

Now that there's no import cycle, we can replace the interface pattern with a direct struct.

#### 6.1: Remove Interface Pattern

**Actions**:
1. Delete `internal/api/` directory (no longer needed)
2. Move implementation directly to `pkg/dot/client.go`
3. Change from interface to concrete struct

**Before** (pkg/dot/client.go):
```go
package dot

type Client interface {
    Manage(ctx context.Context, packages ...string) error
    // ...
}

var newClientImpl func(Config) (Client, error)

func NewClient(cfg Config) (Client, error) {
    return newClientImpl(cfg)
}
```

**After** (pkg/dot/client.go):
```go
package dot

import (
    "context"
    "fmt"
    
    "github.com/jamesainslie/dot/internal/executor"
    "github.com/jamesainslie/dot/internal/pipeline"
)

// Client provides the main API for dot operations.
type Client struct {
    config   Config
    stowPipe *pipeline.StowPipeline
    executor *executor.Executor
    manifest manifest.Store
}

// NewClient creates a new Client with the given configuration.
func NewClient(cfg Config) (*Client, error) {
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    cfg = cfg.WithDefaults()
    
    stowPipe := pipeline.NewStowPipeline(pipeline.StowPipelineOpts{
        FS:     cfg.FS,
        Logger: cfg.Logger,
        Tracer: cfg.Tracer,
    })
    
    exec := executor.New(executor.Opts{
        FS:     cfg.FS,
        Logger: cfg.Logger,
        Tracer: cfg.Tracer,
    })
    
    manifestStore := manifest.NewFSStore(cfg.FS)
    
    return &Client{
        config:   cfg,
        stowPipe: stowPipe,
        executor: exec,
        manifest: manifestStore,
    }, nil
}

// Manage installs packages by creating symlinks.
func (c *Client) Manage(ctx context.Context, packages ...string) error {
    // Implementation from internal/api/manage.go
}

// ... rest of methods
```

**Tasks**:
- [ ] Remove RegisterClientImpl mechanism from pkg/dot/client.go
- [ ] Move all methods from internal/api/*.go to pkg/dot/client.go
- [ ] Change Client from interface to struct
- [ ] Update NewClient to return *Client instead of Client interface
- [ ] Delete internal/api/ directory
- [ ] Run tests: `go test ./pkg/dot/...`
- [ ] Verify no breaking changes to public API surface

**Time**: 1 hour

**Commit**: `refactor(api): replace Client interface with concrete struct`

---

### Step 7: Clean Up Old pkg/dot Domain Files

Now that domain types are re-exported from internal/domain, we can clean up the original files in pkg/dot.

**Actions**:
1. Replace each domain file's contents with just re-exports
2. Keep tests in pkg/dot (they test the public API surface)
3. Ensure re-exports are complete

**Example - pkg/dot/operation.go**:
```go
// Before: Full implementation (300+ lines)
package dot

type Operation interface { ... }
type LinkCreate struct { ... }
// ... all implementation

// After: Re-export only (20 lines)
package dot

import "github.com/jamesainslie/dot/internal/domain"

type Operation = domain.Operation
type OperationID = domain.OperationID
// ... all re-exports
```

**Files to update**:
- [x] result.go (done in Step 3.1)
- [ ] path.go
- [ ] operation.go
- [ ] plan.go
- [ ] domain.go
- [ ] errors.go
- [ ] conflict.go
- [ ] execution.go
- [ ] checkpoint.go
- [ ] ports.go

**Verification per file**:
```bash
# After updating each file
go test ./pkg/dot -v
go test ./internal/... -v
```

**Tasks**:
- [ ] Simplify path.go to re-exports only
- [ ] Simplify operation.go to re-exports only
- [ ] Simplify plan.go to re-exports only
- [ ] Simplify domain.go to re-exports only
- [ ] Simplify errors.go to re-exports only
- [ ] Simplify conflict.go to re-exports only
- [ ] Simplify execution.go to re-exports only
- [ ] Simplify checkpoint.go to re-exports only
- [ ] Simplify ports.go to re-exports only
- [ ] Run full test suite after each change

**Time**: 2-3 hours

**Commits**: One per file
- `refactor(types): simplify pkg/dot/path.go to re-exports`
- `refactor(types): simplify pkg/dot/operation.go to re-exports`
- etc.

---

### Step 8: Update Test Imports

Test files that directly imported types from pkg/dot may need adjustment if they're testing internal behavior.

**Actions**:
1. Scan all `*_test.go` files for imports of `pkg/dot`
2. Determine if they're testing public API (keep `pkg/dot`) or internal behavior (change to `internal/domain`)
3. Update as needed

**Script**:
```bash
# Find all test files importing pkg/dot
grep -r "github.com/jamesainslie/dot/pkg/dot" --include="*_test.go" .

# Review each manually
```

**Decision tree**:
- Testing public API surface → Keep `pkg/dot` import
- Testing internal domain logic → Change to `internal/domain`
- Testing adapters/implementations → Likely need `internal/domain`

**Tasks**:
- [ ] Identify all test files needing updates
- [ ] Update internal package tests
- [ ] Keep pkg/dot tests as-is (test public API)
- [ ] Run full test suite
- [ ] Verify coverage maintained

**Time**: 1-2 hours

**Commit**: `refactor(test): update test imports for domain separation`

---

### Step 9: Verify No Breaking Changes

**Critical Verification Step**: Ensure public API is unchanged.

#### 9.1: API Surface Verification

**Test that consumers see no difference**:

Create `tests/api-compatibility/main.go`:
```go
// This file tests that the public API surface is unchanged
package main

import (
	"context"
	
	"github.com/jamesainslie/dot/pkg/dot"
)

func main() {
	// All these should compile exactly as before refactoring
	
	var _ dot.Result[string]
	var _ dot.Operation
	var _ dot.Plan
	var _ dot.Path[dot.PackagePath]
	var _ dot.Client
	
	cfg := dot.Config{}
	_ = cfg.Validate()
	
	client, _ := dot.NewClient(cfg)
	_ = client
	
	ctx := context.Background()
	_ = client.Manage(ctx, "test")
	
	// If this compiles, public API is compatible
}
```

**Run**:
```bash
go build ./tests/api-compatibility
```

**Tasks**:
- [ ] Create API compatibility test
- [ ] Verify it compiles
- [ ] Test with example from old docs
- [ ] Verify godoc output unchanged

**Time**: 30 minutes

---

#### 9.2: Full Test Suite

**Run comprehensive testing**:
```bash
# All tests with race detector
go test ./... -race

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total

# Linting
make lint

# Full check
make check
```

**Success criteria**:
- [ ] All tests pass
- [ ] Coverage ≥ 80% (same as before)
- [ ] Zero linter errors
- [ ] Zero race conditions
- [ ] Build succeeds

**Time**: 30 minutes

**Commit**: `refactor(verify): confirm zero breaking changes to public API`

---

### Step 10: Update Documentation

**Actions**:
1. Update Architecture.md with new package structure
2. Update README.md if needed
3. Update Phase completion markers
4. Add migration notes

**Files to update**:
- [ ] docs/Architecture.md - Update package structure diagram
- [ ] README.md - Update architecture section
- [ ] PHASE_12_COMPLETE.md - Mark Phase 12b complete
- [ ] CHANGELOG.md - Add refactoring notes

**Documentation updates**:
```markdown
## Architecture (Updated)

The project follows a layered architecture:

- **Domain Layer** (`internal/domain/`): Pure domain model with phantom-typed paths
- **Port Layer** (`internal/domain/ports.go`): Infrastructure interfaces  
- **Public API** (`pkg/dot/`): Re-exports domain types + Client struct
- **Adapter Layer** (`internal/adapters/`): Concrete implementations
- **Core Layer** (`internal/scanner/`, `internal/planner/`, etc.): Pure functional logic
- **Shell Layer** (`internal/executor/`): Side-effecting execution
- **CLI Layer** (`cmd/dot/`): Cobra-based command-line interface
```

**Tasks**:
- [ ] Update Architecture.md package structure
- [ ] Update Architecture.md import diagram
- [ ] Update README architecture section
- [ ] Document refactoring in CHANGELOG
- [ ] Create PHASE_12B_COMPLETE.md

**Time**: 1 hour

**Commit**: `docs(arch): update documentation for domain separation`

---

### Step 11: Clean Up and Finalize

#### 11.1: Remove Dead Code

**Check for orphaned code**:
```bash
# Find files in pkg/dot that are now just re-exports
ls -la pkg/dot/*.go

# Verify each is necessary
# Remove any that are purely internal and shouldn't be public
```

**Tasks**:
- [ ] Review all pkg/dot files
- [ ] Ensure only public API remains
- [ ] Remove any accidental internal exports
- [ ] Verify minimal public surface

**Time**: 30 minutes

---

#### 11.2: Performance Verification

**Benchmark key operations**:
```go
func BenchmarkClientManage(b *testing.B) {
    // Before and after should be identical
}
```

**Tasks**:
- [ ] Run benchmarks before refactoring (baseline)
- [ ] Run benchmarks after refactoring
- [ ] Compare results
- [ ] Ensure no performance regression

**Time**: 30 minutes

---

#### 11.3: Final Validation

**Comprehensive validation**:
```bash
# Clean build
go clean -cache
go build ./...

# Full test suite
go test ./... -v -race -coverprofile=coverage.out

# Check coverage
go tool cover -func=coverage.out | grep total

# Lint everything
make lint

# Try building example programs
cd examples/basic && go build
cd examples/streaming && go build

# Integration tests
go test ./tests/integration/... -v
```

**Success Criteria**:
- [ ] All tests pass (same count as before)
- [ ] Coverage ≥ 80% (maintained or improved)
- [ ] Zero linter errors
- [ ] All examples build
- [ ] Integration tests pass
- [ ] No performance regression

**Time**: 1 hour

**Commit**: `refactor(verify): validate Phase 12b refactoring complete`

---

## Risk Mitigation

### Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking public API | Medium | Critical | Compatibility test file, comprehensive testing |
| Import errors | High | High | Atomic migration per type, test after each |
| Test failures | Medium | High | Run tests after every change, maintain coverage |
| Performance regression | Low | Medium | Benchmark before/after, profile if needed |
| Git merge conflicts | Medium | Medium | Work on isolated branch, merge frequently from main |
| Incomplete re-exports | Medium | High | API compatibility test, godoc review |

### Safety Measures

1. **Backup Branch**: Create safety backup before starting
2. **Atomic Commits**: One logical change per commit, always buildable
3. **Continuous Testing**: Run tests after every single file change
4. **Automated Import Updates**: Use script for mechanical changes
5. **Manual Review**: Manually review all automated changes
6. **Rollback Plan**: Keep backup branch for quick rollback

### Rollback Procedure

If refactoring fails or introduces bugs:
```bash
# Abandon refactoring
git checkout main

# Or cherry-pick good commits
git checkout main
git cherry-pick <commit-hash>
```

---

## Pre-Refactoring Checklist

Before starting Phase 12b:

### Prerequisites
- [ ] Phase 12 (Option 4) complete and stable
- [ ] Phase 12 in production use for at least 2 weeks
- [ ] No known bugs in Phase 12 Client
- [ ] Full test suite passing
- [ ] Coverage ≥ 80%
- [ ] All linters passing

### Preparation
- [ ] Create feature branch: `feature-domain-refactor`
- [ ] Create backup branch: `feature-domain-refactor-backup`
- [ ] Document current API surface (godoc snapshot)
- [ ] Run and save benchmark results (baseline)
- [ ] Ensure clean git state (no uncommitted changes)
- [ ] Review this plan thoroughly
- [ ] Allocate dedicated time block (12-16 hours)

### Team Communication
- [ ] Notify team of refactoring effort
- [ ] Coordinate to avoid conflicts
- [ ] Plan for code review time
- [ ] Schedule integration testing

---

## Step-by-Step Execution Checklist

### Day 1: Foundation (4-5 hours)

**Morning**:
- [ ] Step 1: Create migration branch (5 min)
- [ ] Step 2: Create internal/domain structure (10 min)
- [ ] Step 3.1: Move Result monad (30 min)
  - [ ] Verify tests pass
- [ ] Step 3.2: Move Path types (30 min)
  - [ ] Verify tests pass
- [ ] Step 3.3: Move Operation types (45 min)
  - [ ] Verify tests pass

**Afternoon**:
- [ ] Step 3.4: Move remaining types (2-3 hours)
  - [ ] errors.go
  - [ ] domain.go
  - [ ] plan.go
  - [ ] conflict.go
  - [ ] execution.go
  - [ ] checkpoint.go
  - [ ] Verify after each

**End of Day**: All domain types moved, all tests passing

---

### Day 2: Internal Updates (4-5 hours)

**Morning**:
- [ ] Step 4: Move port interfaces (30 min)
- [ ] Step 5.1: Update internal/executor imports (45 min)
  - [ ] Run tests
- [ ] Step 5.1: Update internal/pipeline imports (45 min)
  - [ ] Run tests
- [ ] Step 5.1: Update internal/scanner imports (30 min)
  - [ ] Run tests

**Afternoon**:
- [ ] Step 5.1: Update internal/planner imports (45 min)
- [ ] Step 5.1: Update internal/manifest imports (30 min)
- [ ] Step 5.1: Update internal/config imports (15 min)
- [ ] Step 5.1: Update internal/adapters imports (30 min)
- [ ] Step 5.1: Update internal/ignore imports (15 min)
- [ ] Full test suite run

**End of Day**: All internal packages updated, all tests passing

---

### Day 3: API Finalization (3-4 hours)

**Morning**:
- [ ] Step 6: Replace Client interface with struct (1 hour)
  - [ ] Move implementations from internal/api
  - [ ] Update Client definition
  - [ ] Delete internal/api
  - [ ] Run tests
- [ ] Step 7: Clean up pkg/dot files (30 min)
  - [ ] Simplify to re-exports
  - [ ] Verify each file

**Afternoon**:
- [ ] Step 8: Update test imports (1-2 hours)
- [ ] Step 9: Verify no breaking changes (1 hour)
  - [ ] API compatibility test
  - [ ] Full test suite
  - [ ] Benchmark comparison
- [ ] Step 10: Update documentation (1 hour)

**End of Day**: Refactoring complete, ready for review

---

## Validation Checklist

### Code Quality
- [ ] All tests pass: `go test ./...`
- [ ] Race detector clean: `go test ./... -race`
- [ ] Coverage maintained: ≥ 80%
- [ ] All linters pass: `make lint`
- [ ] Builds successfully: `make build`
- [ ] Examples compile: `cd examples/* && go build`

### API Compatibility
- [ ] Client struct has same methods as old interface
- [ ] NewClient signature unchanged
- [ ] All domain types still exported from pkg/dot
- [ ] Godoc output similar to pre-refactoring
- [ ] Example code from old docs still works

### Performance
- [ ] Benchmarks show no regression
- [ ] Memory usage unchanged
- [ ] No new allocations in hot paths

### Documentation
- [ ] Architecture docs updated
- [ ] Package documentation accurate
- [ ] Examples still work
- [ ] CHANGELOG updated
- [ ] Migration guide written (if needed)

---

## Benefits of Refactoring

### Architectural Benefits

**Before** (Option 4 - Interface Pattern):
```
pkg/dot/client.go              # Interface definition
internal/api/client.go         # Implementation (separate package)
                               # Indirection via registration
```

**After** (Option 1 - Direct Implementation):
```
pkg/dot/client.go              # Direct struct implementation
                               # No indirection needed
```

### Concrete Improvements

1. **Simpler Code**:
   - No init() registration
   - No var newClientImpl indirection
   - Direct method calls (no interface dispatch)
   - Fewer packages to navigate

2. **Better Performance**:
   - No interface indirection (minor but measurable)
   - Compiler can inline more aggressively
   - Smaller binary size (dead code elimination)

3. **Easier Debugging**:
   - Stack traces show concrete type
   - No interface method calls in traces
   - Clearer call graphs

4. **Cleaner Architecture**:
   - Domain in `internal/domain/` (private implementation)
   - API in `pkg/dot/` (public interface)
   - Clear separation of concerns
   - Standard Go project layout

5. **Maintainability**:
   - Internal packages can refactor freely
   - Public API surface is just re-exports + Client
   - Less coupling between layers
   - Easier to understand for new contributors

---

## Post-Refactoring

### Update Phase Markers

Create `PHASE_12B_COMPLETE.md`:
```markdown
# Phase 12b: Domain Architecture Refactoring - COMPLETE

## Overview

Successfully refactored from interface-based Client (Option 4) to direct
Client struct (Option 1) by moving all domain types from pkg/dot/ to
internal/domain/.

## Changes

- Moved ~15 domain type files to internal/domain/
- Updated ~80-100 internal package imports
- Replaced Client interface with concrete Client struct
- Removed internal/api package (no longer needed)
- Simplified pkg/dot to re-exports + Client + Config

## Verification

- All tests pass (X,XXX tests)
- Coverage maintained at X%
- Zero breaking changes to public API
- Performance unchanged (benchmarks within 2%)
- All linters pass

## Benefits

- Cleaner architecture
- Simpler Client implementation (no interface indirection)
- Better separation of concerns
- Standard Go project layout
```

### Documentation Updates

- [ ] Update all diagrams in Architecture.md
- [ ] Update package structure in README.md
- [ ] Add refactoring notes to CHANGELOG.md
- [ ] Update examples if needed

---

## Rollback Plan

If critical issues discovered post-merge:

### Immediate Rollback
```bash
# Revert the merge commit
git revert <merge-commit-hash>
git push origin main
```

### Selective Rollback
```bash
# Cherry-pick good commits, abandon problematic ones
git checkout -b feature-selective-rollback main~20
git cherry-pick <good-commit-1>
git cherry-pick <good-commit-2>
# ... etc
```

### Full Rollback
```bash
# Reset to before refactoring
git reset --hard <commit-before-refactor>
git push --force-with-lease origin feature-domain-refactor
```

---

## Timeline

### Estimated Duration
**Total**: 12-16 hours over 2-3 days

**Breakdown**:
- Day 1 (4-5 hours): Move domain types to internal/domain
- Day 2 (4-5 hours): Update all internal package imports
- Day 3 (3-4 hours): Replace interface with struct, finalize

### Prerequisites
- Dedicated time blocks (avoid context switching)
- Clean git state
- No other major changes in progress
- Team awareness and code review availability

---

## Success Metrics

Phase 12b is successful when:

- [ ] All domain types in `internal/domain/`
- [ ] All re-exported from `pkg/dot/`
- [ ] Client is concrete struct, not interface
- [ ] No `internal/api/` package
- [ ] No interface registration mechanism
- [ ] Zero breaking changes to public API
- [ ] All tests pass
- [ ] Coverage ≥ 80%
- [ ] Performance maintained
- [ ] Documentation updated
- [ ] Cleaner, more maintainable codebase

---

## Why This Refactoring Matters

### Current State (Phase 12 with Option 4)
**Works but has technical debt**:
- Interface indirection adds complexity
- Registration mechanism is clever but opaque
- `internal/api` exists only to avoid import cycle
- Not the "natural" Go way

### Target State (Phase 12b with Option 1)
**Clean, idiomatic architecture**:
- Direct Client struct (like sql.DB, http.Client)
- No workarounds or clever tricks
- Standard Go project layout
- Easy for contributors to understand
- Maintainable long-term

### The Investment

**Cost**: 12-16 hours of careful refactoring
**Benefit**: Years of easier maintenance and cleaner architecture
**Risk**: Mitigated through atomic commits and comprehensive testing

---

## When to Execute Phase 12b

### Recommended Timing

**Not immediately after Phase 12**:
- Let Phase 12 (Option 4) stabilize in production
- Gather feedback on API usability
- Identify any issues with current approach
- Build confidence in test suite

**Execute Phase 12b when**:
- [ ] Phase 12 has been stable for 2+ weeks
- [ ] No critical bugs in current implementation
- [ ] Team has capacity for careful refactoring
- [ ] Code review bandwidth available
- [ ] No other major features in progress

### Indicators It's Time

1. **Pain Points Emerge**:
   - Interface indirection causing confusion
   - Debug traces harder to follow
   - New contributors struggling with architecture

2. **Feature Development Needs**:
   - Adding advanced Client features awkward with interface
   - Want to store state in Client (cache, connection pooling)
   - Need better performance (eliminate interface indirection)

3. **Maintenance Burden**:
   - Registration mechanism causing issues
   - Want to refactor internal packages but worried about API
   - Contributing guide has to explain workarounds

### Indicators to Defer

1. **Instability**:
   - Bugs still being found in Phase 12
   - API changes being considered
   - Test suite not comprehensive enough

2. **Resource Constraints**:
   - No dedicated time for refactoring
   - Team focused on features
   - Can't afford potential regressions

3. **Not Worth It**:
   - Current approach working fine
   - No pain points in practice
   - Other priorities more important

---

## Alternative: Never Execute Phase 12b

### Option: Keep Interface Pattern Permanently

**Argument**: If Option 4 works well in practice, there's no requirement to refactor.

**Pros**:
- Avoid refactoring risk
- Current code is tested and working
- Interface is arguably more flexible
- Can mock Client in tests

**Cons**:
- Carries technical debt permanently
- Registration mechanism remains opaque
- Not standard Go pattern
- Extra package (`internal/api`) to maintain

**Decision Framework**:

Keep Option 4 if:
- No maintenance pain after 6 months of use
- Interface flexibility proves valuable
- Team comfortable with the pattern

Execute Phase 12b if:
- Pain points emerge in practice
- Want simpler architecture
- Contributing complexity is an issue
- Performance optimization needed

---

## Summary

Phase 12b is an **optional architectural improvement** that transforms the codebase from a working-but-complex interface pattern to a clean, idiomatic struct-based design.

**Execute when**: Stable, have time, value long-term maintainability
**Skip if**: Current approach works well, other priorities higher

The plan provides a detailed, step-by-step roadmap with:
- 11 major steps
- Atomic commits throughout
- Comprehensive testing at each stage
- Risk mitigation strategies
- Clear success criteria
- Rollback procedures

**Estimated effort**: 12-16 hours over 2-3 days
**Risk level**: Medium-High (mitigated through careful process)
**Value**: Long-term architectural cleanliness and maintainability

