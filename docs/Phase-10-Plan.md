# Phase 10: Imperative Shell - Executor

## Overview

Phase 10 implements the imperative shell that executes plans with side effects. This phase bridges the pure functional core (Phases 4-8) with the filesystem, providing transaction safety, rollback capabilities, and observability.

## Dependencies

**Prerequisites:**
- Phase 1: Domain Model (Operation interface, Result type, error types)
- Phase 2: Infrastructure Ports (FS, Logger, Tracer, Metrics interfaces)
- Phase 3: Adapters (OSFilesystem, MemoryFilesystem, SlogLogger, etc.)
- Phase 7: Resolver (Plan type with validated operations)
- Phase 8: Topological Sorter (DependencyGraph, parallelization plan)

**Phase 10 Deliverable:** Robust executor with two-phase commit, rollback, parallel execution, and full observability integration.

## Design Principles

- **Functional Core, Imperative Shell**: Core logic remains pure; executor handles all side effects
- **Transaction Safety**: Two-phase commit ensures atomicity
- **Fail-Safe**: Automatic rollback on any failure
- **Observable**: Full tracing, metrics, and structured logging
- **Testable**: Memory filesystem enables thorough testing without real I/O
- **Performance**: Parallel execution based on dependency analysis

## Architecture Context

The Executor sits at the boundary between pure planning and side-effecting execution:

```text
Pure Functional Core → Plan (validated, sorted) → Executor → Filesystem Changes
                                                   ↓
                                              Checkpoint
                                                   ↓
                                            Rollback (on failure)
```

## Phase Structure

### 10.1: Basic Execution
### 10.2: Two-Phase Commit  
### 10.3: Rollback Mechanism
### 10.4: Parallel Execution
### 10.5: Instrumentation

---

## 10.1: Basic Execution

**Goal:** Implement core executor with sequential operation execution and error handling.

### 10.1.1: Executor Structure

**Test:** `internal/executor/executor_test.go`

```go
func TestNewExecutor(t *testing.T) {
    fs := memfs.New()
    logger := noop.NewLogger()
    tracer := noop.NewTracer()
    
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: logger,
        Tracer: tracer,
    })
    
    require.NotNil(t, exec)
}
```

**Implementation:** `internal/executor/executor.go`

```go
package executor

import (
    "context"
    "github.com/yourorg/dot/internal/domain"
    "github.com/yourorg/dot/internal/ports"
)

// Executor executes validated plans with transaction safety
type Executor struct {
    fs         ports.FS
    log        ports.Logger
    tracer     ports.Tracer
    checkpoint CheckpointStore
}

type Opts struct {
    FS         ports.FS
    Logger     ports.Logger
    Tracer     ports.Tracer
    Checkpoint CheckpointStore
}

func New(opts Opts) *Executor {
    if opts.Checkpoint == nil {
        opts.Checkpoint = NewMemoryCheckpointStore()
    }
    
    return &Executor{
        fs:         opts.FS,
        log:        opts.Logger,
        tracer:     opts.Tracer,
        checkpoint: opts.Checkpoint,
    }
}
```

**Tasks:**
- [ ] Create `internal/executor/` package
- [ ] Define `Executor` struct with dependencies
- [ ] Define `Opts` for constructor dependency injection
- [ ] Implement `New()` constructor with validation
- [ ] Add default checkpoint store when not provided
- [ ] Write constructor tests

**Commit:** `feat(executor): implement Executor struct and constructor`

### 10.1.2: Execution Result Type

**Test:** `internal/executor/result_test.go`

```go
func TestExecutionResult(t *testing.T) {
    result := ExecutionResult{
        Executed:   []OperationID{"op1", "op2"},
        Failed:     []OperationID{"op3"},
        RolledBack: nil,
        Errors:     []error{errors.New("op3 failed")},
    }
    
    require.Len(t, result.Executed, 2)
    require.Len(t, result.Failed, 1)
    require.NotEmpty(t, result.Errors)
}
```

**Implementation:** `internal/executor/result.go`

```go
package executor

import "github.com/yourorg/dot/internal/domain"

type ExecutionResult struct {
    Executed   []domain.OperationID
    Failed     []domain.OperationID
    RolledBack []domain.OperationID
    Errors     []error
}

func (r ExecutionResult) Success() bool {
    return len(r.Failed) == 0 && len(r.Errors) == 0
}

func (r ExecutionResult) PartialFailure() bool {
    return len(r.Executed) > 0 && len(r.Failed) > 0
}
```

**Tasks:**
- [ ] Define `ExecutionResult` struct
- [ ] Add `Success()` method
- [ ] Add `PartialFailure()` method for detecting partial execution
- [ ] Write result type tests

**Commit:** `feat(executor): add ExecutionResult type`

### 10.1.3: Sequential Execution

**Test:** `internal/executor/executor_test.go`

```go
func TestExecuteSequential_Success(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    logger := noop.NewLogger()
    tracer := noop.NewTracer()
    
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: logger,
        Tracer: tracer,
    })
    
    // Create simple plan with LinkCreate operation
    source := "/stow/pkg/file"
    target := "/home/file"
    fs.WriteFile(ctx, source, []byte("content"), 0644)
    
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath(source),
        Target: mustPath(target),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    result := exec.Execute(ctx, plan)
    require.NoError(t, result.Unwrap())
    
    // Verify symlink created
    exists := fs.IsSymlink(ctx, target)
    require.True(t, exists)
}

func TestExecuteSequential_OperationFailure(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Create operation that will fail (source doesn't exist)
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/nonexistent"),
        Target: mustPath("/home/file"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    result := exec.Execute(ctx, plan)
    execResult, err := result.Unwrap()
    
    require.Error(t, err)
    require.Len(t, execResult.Failed, 1)
    require.Contains(t, execResult.Failed, domain.OperationID("link1"))
}
```

**Implementation:** `internal/executor/executor.go`

```go
func (e *Executor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
    ctx, span := e.tracer.Start(ctx, "executor.Execute")
    defer span.End()
    
    e.log.Info(ctx, "executing_plan", 
        "operation_count", len(plan.Operations()),
        "plan_checksum", plan.Checksum())
    
    // Validate plan
    if err := e.validatePlan(ctx, plan); err != nil {
        return domain.Err[ExecutionResult](err)
    }
    
    // Execute sequentially
    result := e.executeSequential(ctx, plan)
    
    if len(result.Failed) > 0 {
        e.log.Error(ctx, "execution_failed", 
            "executed", len(result.Executed),
            "failed", len(result.Failed))
        return domain.Err[ExecutionResult](domain.ErrExecutionFailed{
            Executed: result.Executed,
            Failed:   result.Failed,
            Errors:   result.Errors,
        })
    }
    
    e.log.Info(ctx, "execution_complete", "operations", len(result.Executed))
    return domain.Ok(result)
}

func (e *Executor) validatePlan(ctx context.Context, plan domain.Plan) error {
    if len(plan.Operations()) == 0 {
        return domain.ErrEmptyPlan{}
    }
    return nil
}

func (e *Executor) executeSequential(ctx context.Context, plan domain.Plan) ExecutionResult {
    result := ExecutionResult{
        Executed: make([]domain.OperationID, 0),
        Failed:   make([]domain.OperationID, 0),
        Errors:   make([]error, 0),
    }
    
    for _, op := range plan.Operations() {
        opID := op.ID()
        
        e.log.Debug(ctx, "executing_operation", 
            "op_id", opID, 
            "op_kind", op.Kind())
        
        if err := op.Execute(ctx, e.fs); err != nil {
            e.log.Error(ctx, "operation_failed", 
                "op_id", opID, 
                "error", err)
            result.Failed = append(result.Failed, opID)
            result.Errors = append(result.Errors, err)
            break // Stop on first failure
        }
        
        result.Executed = append(result.Executed, opID)
    }
    
    return result
}
```

**Tasks:**
- [ ] Implement `Execute()` method accepting context and plan
- [ ] Add `validatePlan()` for basic plan validation
- [ ] Implement `executeSequential()` for operation-by-operation execution
- [ ] Call `op.Execute(ctx, fs)` for each operation
- [ ] Collect executed operations in result
- [ ] Stop on first failure (fail-fast for now)
- [ ] Return Result[ExecutionResult] with proper error wrapping
- [ ] Write success tests
- [ ] Write failure tests
- [ ] Write tests for empty plan

**Commit:** `feat(executor): implement sequential operation execution`

---

## 10.2: Two-Phase Commit

**Goal:** Add prepare-then-commit transaction safety with validation before execution.

### 10.2.1: Prepare Phase

**Test:** `internal/executor/prepare_test.go`

```go
func TestPrepare_Success(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Set up valid preconditions
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    err := exec.Prepare(ctx, plan)
    require.NoError(t, err)
}

func TestPrepare_MissingSource(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/nonexistent"),
        Target: mustPath("/home/file"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    err := exec.Prepare(ctx, plan)
    require.Error(t, err)
    require.Contains(t, err.Error(), "source does not exist")
}

func TestPrepare_PermissionDenied(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Create read-only target directory
    fs.MkdirAll(ctx, "/home", 0444)
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    err := exec.Prepare(ctx, plan)
    require.Error(t, err)
    require.IsType(t, domain.ErrPermissionDenied{}, err)
}
```

**Implementation:** `internal/executor/prepare.go`

```go
package executor

import (
    "context"
    "fmt"
    "github.com/yourorg/dot/internal/domain"
)

func (e *Executor) prepare(ctx context.Context, plan domain.Plan) error {
    ctx, span := e.tracer.Start(ctx, "executor.Prepare")
    defer span.End()
    
    e.log.Debug(ctx, "preparing_plan", "operations", len(plan.Operations()))
    
    for _, op := range plan.Operations() {
        if err := op.Validate(); err != nil {
            return fmt.Errorf("validation failed for %v: %w", op.ID(), err)
        }
        
        if err := e.checkPreconditions(ctx, op); err != nil {
            return fmt.Errorf("precondition check failed for %v: %w", op.ID(), err)
        }
    }
    
    e.log.Debug(ctx, "prepare_complete")
    return nil
}

func (e *Executor) checkPreconditions(ctx context.Context, op domain.Operation) error {
    switch operation := op.(type) {
    case *domain.LinkCreate:
        return e.checkLinkCreatePreconditions(ctx, operation)
    case *domain.DirCreate:
        return e.checkDirCreatePreconditions(ctx, operation)
    case *domain.FileMove:
        return e.checkFileMovePreconditions(ctx, operation)
    default:
        return nil
    }
}

func (e *Executor) checkLinkCreatePreconditions(ctx context.Context, op *domain.LinkCreate) error {
    // Verify source exists
    if !e.fs.Exists(ctx, op.Source.String()) {
        return domain.ErrSourceNotFound{Path: op.Source.String()}
    }
    
    // Verify target parent directory is writable
    parent := op.Target.Parent()
    if !e.fs.Exists(ctx, parent.String()) {
        return domain.ErrParentNotFound{Path: parent.String()}
    }
    
    // Check write permission on parent
    info, err := e.fs.Stat(ctx, parent.String())
    if err != nil {
        return err
    }
    
    if info.Mode().Perm()&0200 == 0 {
        return domain.ErrPermissionDenied{
            Path: parent.String(),
            Op:   "write",
        }
    }
    
    return nil
}

func (e *Executor) checkDirCreatePreconditions(ctx context.Context, op *domain.DirCreate) error {
    // Check parent directory exists and is writable
    parent := op.Path.Parent()
    if !e.fs.Exists(ctx, parent.String()) {
        return domain.ErrParentNotFound{Path: parent.String()}
    }
    
    return nil
}

func (e *Executor) checkFileMovePreconditions(ctx context.Context, op *domain.FileMove) error {
    // Verify source exists
    if !e.fs.Exists(ctx, op.From.String()) {
        return domain.ErrSourceNotFound{Path: op.From.String()}
    }
    
    // Verify destination parent is writable
    parent := op.To.Parent()
    if !e.fs.Exists(ctx, parent.String()) {
        return domain.ErrParentNotFound{Path: parent.String()}
    }
    
    return nil
}
```

**Tasks:**
- [ ] Implement `prepare()` method
- [ ] Call `op.Validate()` for each operation
- [ ] Implement `checkPreconditions()` dispatcher
- [ ] Implement `checkLinkCreatePreconditions()`
  - [ ] Verify source exists
  - [ ] Verify target parent exists
  - [ ] Check write permissions on parent
- [ ] Implement `checkDirCreatePreconditions()`
  - [ ] Verify parent directory exists
  - [ ] Check permissions
- [ ] Implement `checkFileMovePreconditions()`
  - [ ] Verify source exists
  - [ ] Verify destination is writable
- [ ] Write tests for each precondition check
- [ ] Write tests for permission errors

**Commit:** `feat(executor): implement prepare phase with precondition checks`

### 10.2.2: Checkpoint Creation

**Test:** `internal/executor/checkpoint_test.go`

```go
func TestCheckpoint_Create(t *testing.T) {
    ctx := context.Background()
    store := executor.NewMemoryCheckpointStore()
    
    checkpoint := store.Create(ctx)
    
    require.NotEmpty(t, checkpoint.ID)
    require.NotZero(t, checkpoint.CreatedAt)
    require.NotNil(t, checkpoint.Operations)
}

func TestCheckpoint_Record(t *testing.T) {
    ctx := context.Background()
    store := executor.NewMemoryCheckpointStore()
    checkpoint := store.Create(ctx)
    
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/source"),
        Target: mustPath("/target"),
    }
    
    checkpoint.Record("link1", op)
    
    retrieved := checkpoint.Lookup("link1")
    require.NotNil(t, retrieved)
    require.Equal(t, op.ID, retrieved.ID())
}
```

**Implementation:** `internal/executor/checkpoint.go`

```go
package executor

import (
    "context"
    "time"
    "github.com/google/uuid"
    "github.com/yourorg/dot/internal/domain"
)

type CheckpointID string

type Checkpoint struct {
    ID         CheckpointID
    CreatedAt  time.Time
    Operations map[domain.OperationID]domain.Operation
}

func (c *Checkpoint) Record(id domain.OperationID, op domain.Operation) {
    c.Operations[id] = op
}

func (c *Checkpoint) Lookup(id domain.OperationID) domain.Operation {
    return c.Operations[id]
}

type CheckpointStore interface {
    Create(ctx context.Context) *Checkpoint
    Delete(ctx context.Context, id CheckpointID) error
    Restore(ctx context.Context, id CheckpointID) (*Checkpoint, error)
}

// MemoryCheckpointStore keeps checkpoints in memory (for testing and simple cases)
type MemoryCheckpointStore struct {
    checkpoints map[CheckpointID]*Checkpoint
}

func NewMemoryCheckpointStore() *MemoryCheckpointStore {
    return &MemoryCheckpointStore{
        checkpoints: make(map[CheckpointID]*Checkpoint),
    }
}

func (s *MemoryCheckpointStore) Create(ctx context.Context) *Checkpoint {
    id := CheckpointID(uuid.New().String())
    checkpoint := &Checkpoint{
        ID:         id,
        CreatedAt:  time.Now(),
        Operations: make(map[domain.OperationID]domain.Operation),
    }
    s.checkpoints[id] = checkpoint
    return checkpoint
}

func (s *MemoryCheckpointStore) Delete(ctx context.Context, id CheckpointID) error {
    delete(s.checkpoints, id)
    return nil
}

func (s *MemoryCheckpointStore) Restore(ctx context.Context, id CheckpointID) (*Checkpoint, error) {
    checkpoint, exists := s.checkpoints[id]
    if !exists {
        return nil, domain.ErrCheckpointNotFound{ID: string(id)}
    }
    return checkpoint, nil
}
```

**Tasks:**
- [ ] Define `Checkpoint` struct with ID, timestamp, operations map
- [ ] Implement `Record()` to store executed operations
- [ ] Implement `Lookup()` to retrieve operations
- [ ] Define `CheckpointStore` interface
- [ ] Implement `MemoryCheckpointStore` for testing
- [ ] Implement `Create()` with UUID generation
- [ ] Implement `Delete()` for cleanup
- [ ] Implement `Restore()` for recovery
- [ ] Write checkpoint tests

**Commit:** `feat(executor): implement checkpoint system`

### 10.2.3: Integrate Two-Phase Commit

**Test:** `internal/executor/executor_test.go`

```go
func TestExecute_TwoPhaseCommit(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Set up filesystem
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    // Execute should succeed after prepare
    result := exec.Execute(ctx, plan)
    require.NoError(t, result.Unwrap())
}

func TestExecute_PrepareFails(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Source doesn't exist - prepare should fail
    op := domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/nonexistent"),
        Target: mustPath("/home/file"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    result := exec.Execute(ctx, plan)
    require.Error(t, result.Unwrap())
    
    // No operations should have been executed
    execResult, _ := result.Unwrap()
    require.Empty(t, execResult.Executed)
}
```

**Implementation:** Update `internal/executor/executor.go`

```go
func (e *Executor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
    ctx, span := e.tracer.Start(ctx, "executor.Execute")
    defer span.End()
    
    // Phase 1: Prepare - validate all operations
    if err := e.prepare(ctx, plan); err != nil {
        e.log.Error(ctx, "prepare_failed", "error", err)
        span.RecordError(err)
        return domain.Err[ExecutionResult](err)
    }
    
    // Create checkpoint before execution
    checkpoint := e.checkpoint.Create(ctx)
    e.log.Info(ctx, "checkpoint_created", "checkpoint_id", checkpoint.ID)
    
    // Phase 2: Commit - execute operations
    result := e.commit(ctx, plan, checkpoint)
    
    if len(result.Failed) > 0 {
        e.log.Warn(ctx, "execution_failed_rolling_back", "failed_count", len(result.Failed))
        // Rollback will be implemented in 10.3
        return domain.Err[ExecutionResult](domain.ErrExecutionFailed{
            Executed: result.Executed,
            Failed:   result.Failed,
            Errors:   result.Errors,
        })
    }
    
    // Success - delete checkpoint
    e.checkpoint.Delete(ctx, checkpoint.ID)
    e.log.Info(ctx, "execution_complete", "operations", len(result.Executed))
    
    return domain.Ok(result)
}

func (e *Executor) commit(ctx context.Context, plan domain.Plan, checkpoint *Checkpoint) ExecutionResult {
    result := e.executeSequential(ctx, plan)
    
    // Record executed operations to checkpoint
    for _, opID := range result.Executed {
        op := findOperation(plan, opID)
        checkpoint.Record(opID, op)
    }
    
    return result
}

func findOperation(plan domain.Plan, id domain.OperationID) domain.Operation {
    for _, op := range plan.Operations() {
        if op.ID() == id {
            return op
        }
    }
    return nil
}
```

**Tasks:**
- [ ] Update `Execute()` to call `prepare()` first
- [ ] Create checkpoint before commit
- [ ] Implement `commit()` method wrapping execution
- [ ] Record executed operations to checkpoint
- [ ] Delete checkpoint on success
- [ ] Add helper `findOperation()` to locate operations by ID
- [ ] Write tests for two-phase flow
- [ ] Write tests for prepare failure (no execution)
- [ ] Write tests for commit success (checkpoint deleted)

**Commit:** `feat(executor): integrate two-phase commit with checkpointing`

---

## 10.3: Rollback Mechanism

**Goal:** Implement automatic rollback on failure using checkpoints.

### 10.3.1: Rollback Logic

**Test:** `internal/executor/rollback_test.go`

```go
func TestRollback_Success(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Set up filesystem and create link
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    fs.Symlink(ctx, "/stow/pkg/file", "/home/file")
    
    // Create checkpoint with the operation
    checkpoint := exec.CheckpointStore().Create(ctx)
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
    }
    checkpoint.Record("link1", op)
    
    // Rollback
    rolledBack := exec.Rollback(ctx, []domain.OperationID{"link1"}, checkpoint)
    
    require.Len(t, rolledBack, 1)
    require.Contains(t, rolledBack, domain.OperationID("link1"))
    
    // Verify link was removed
    exists := fs.Exists(ctx, "/home/file")
    require.False(t, exists)
}

func TestRollback_ReverseOrder(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Create operations in order: DirCreate, then LinkCreate
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    fs.MkdirAll(ctx, "/home/subdir", 0755)
    fs.Symlink(ctx, "/stow/pkg/file", "/home/subdir/file")
    
    checkpoint := exec.CheckpointStore().Create(ctx)
    
    dirOp := &domain.DirCreate{
        ID:   "dir1",
        Path: mustPath("/home/subdir"),
    }
    linkOp := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/subdir/file"),
    }
    
    checkpoint.Record("dir1", dirOp)
    checkpoint.Record("link1", linkOp)
    
    // Rollback should happen in reverse order: link first, then dir
    executed := []domain.OperationID{"dir1", "link1"}
    rolledBack := exec.Rollback(ctx, executed, checkpoint)
    
    require.Len(t, rolledBack, 2)
    
    // Verify both were removed
    require.False(t, fs.Exists(ctx, "/home/subdir/file"))
    require.False(t, fs.Exists(ctx, "/home/subdir"))
}
```

**Implementation:** `internal/executor/rollback.go`

```go
package executor

import (
    "context"
    "github.com/yourorg/dot/internal/domain"
)

func (e *Executor) rollback(ctx context.Context, executed []domain.OperationID, checkpoint *Checkpoint) []domain.OperationID {
    ctx, span := e.tracer.Start(ctx, "executor.Rollback")
    defer span.End()
    
    e.log.Warn(ctx, "starting_rollback", "operations", len(executed))
    
    var rolledBack []domain.OperationID
    
    // Rollback in reverse order
    for i := len(executed) - 1; i >= 0; i-- {
        opID := executed[i]
        op := checkpoint.Lookup(opID)
        
        if op == nil {
            e.log.Error(ctx, "operation_not_in_checkpoint", "op_id", opID)
            continue
        }
        
        e.log.Debug(ctx, "rolling_back_operation", "op_id", opID, "op_kind", op.Kind())
        
        if err := op.Rollback(ctx, e.fs); err != nil {
            e.log.Error(ctx, "rollback_failed", "op_id", opID, "error", err)
            // Continue rolling back other operations
        } else {
            rolledBack = append(rolledBack, opID)
        }
    }
    
    e.log.Info(ctx, "rollback_complete", 
        "attempted", len(executed),
        "succeeded", len(rolledBack))
    
    return rolledBack
}
```

**Tasks:**
- [ ] Implement `rollback()` method
- [ ] Iterate executed operations in reverse order
- [ ] Lookup each operation from checkpoint
- [ ] Call `op.Rollback(ctx, fs)` for each
- [ ] Handle rollback errors gracefully (log but continue)
- [ ] Collect successfully rolled back operations
- [ ] Add tracing and logging
- [ ] Write tests for successful rollback
- [ ] Write tests for reverse order
- [ ] Write tests for partial rollback (some fail)

**Commit:** `feat(executor): implement rollback mechanism`

### 10.3.2: Integrate Rollback into Execute

**Test:** `internal/executor/executor_test.go`

```go
func TestExecute_AutomaticRollback(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Set up for success on first operation, failure on second
    fs.WriteFile(ctx, "/stow/pkg/file1", []byte("content1"), 0644)
    
    op1 := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file1"),
        Target: mustPath("/home/file1"),
        Mode:   domain.LinkRelative,
    }
    
    op2 := &domain.LinkCreate{
        ID:     "link2",
        Source: mustPath("/nonexistent"), // Will fail
        Target: mustPath("/home/file2"),
        Mode:   domain.LinkRelative,
    }
    
    plan := mustPlan([]domain.Operation{op1, op2})
    
    result := exec.Execute(ctx, plan)
    execResult, err := result.Unwrap()
    
    require.Error(t, err)
    require.Len(t, execResult.Executed, 1)
    require.Len(t, execResult.Failed, 1)
    require.Len(t, execResult.RolledBack, 1)
    
    // Verify first operation was rolled back
    exists := fs.Exists(ctx, "/home/file1")
    require.False(t, exists, "rolled back operation should be undone")
}
```

**Implementation:** Update `internal/executor/executor.go`

```go
func (e *Executor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
    ctx, span := e.tracer.Start(ctx, "executor.Execute")
    defer span.End()
    
    // Phase 1: Prepare
    if err := e.prepare(ctx, plan); err != nil {
        e.log.Error(ctx, "prepare_failed", "error", err)
        span.RecordError(err)
        return domain.Err[ExecutionResult](err)
    }
    
    // Create checkpoint
    checkpoint := e.checkpoint.Create(ctx)
    e.log.Info(ctx, "checkpoint_created", "checkpoint_id", checkpoint.ID)
    
    // Phase 2: Commit
    result := e.commit(ctx, plan, checkpoint)
    
    if len(result.Failed) > 0 {
        // Automatic rollback
        e.log.Warn(ctx, "execution_failed_rolling_back", "failed_count", len(result.Failed))
        rolledBack := e.rollback(ctx, result.Executed, checkpoint)
        result.RolledBack = rolledBack
        
        return domain.Err[ExecutionResult](domain.ErrExecutionFailed{
            Executed:   result.Executed,
            Failed:     result.Failed,
            RolledBack: result.RolledBack,
            Errors:     result.Errors,
        })
    }
    
    // Success - delete checkpoint
    e.checkpoint.Delete(ctx, checkpoint.ID)
    e.log.Info(ctx, "execution_complete", "operations", len(result.Executed))
    
    return domain.Ok(result)
}
```

**Tasks:**
- [ ] Call `rollback()` when execution fails
- [ ] Store rolled back operations in result
- [ ] Include rollback information in error
- [ ] Write tests for automatic rollback
- [ ] Write tests verifying filesystem state after rollback
- [ ] Verify checkpoint is not deleted on failure

**Commit:** `feat(executor): integrate automatic rollback on failure`

---

## 10.4: Parallel Execution

**Goal:** Execute independent operations concurrently for performance.

### 10.4.1: Batch Execution

**Test:** `internal/executor/parallel_test.go`

```go
func TestExecuteBatch_Concurrent(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Create independent operations (no dependencies)
    fs.WriteFile(ctx, "/stow/pkg/file1", []byte("content1"), 0644)
    fs.WriteFile(ctx, "/stow/pkg/file2", []byte("content2"), 0644)
    fs.WriteFile(ctx, "/stow/pkg/file3", []byte("content3"), 0644)
    
    ops := []domain.Operation{
        &domain.LinkCreate{
            ID:     "link1",
            Source: mustPath("/stow/pkg/file1"),
            Target: mustPath("/home/file1"),
        },
        &domain.LinkCreate{
            ID:     "link2",
            Source: mustPath("/stow/pkg/file2"),
            Target: mustPath("/home/file2"),
        },
        &domain.LinkCreate{
            ID:     "link3",
            Source: mustPath("/stow/pkg/file3"),
            Target: mustPath("/home/file3"),
        },
    }
    
    checkpoint := exec.CheckpointStore().Create(ctx)
    result := exec.ExecuteBatch(ctx, ops, checkpoint)
    
    require.Len(t, result.Executed, 3)
    require.Empty(t, result.Failed)
    
    // Verify all links created
    require.True(t, fs.Exists(ctx, "/home/file1"))
    require.True(t, fs.Exists(ctx, "/home/file2"))
    require.True(t, fs.Exists(ctx, "/home/file3"))
}

func TestExecuteBatch_PartialFailure(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Mix of success and failure
    fs.WriteFile(ctx, "/stow/pkg/file1", []byte("content1"), 0644)
    // file2 doesn't exist - will fail
    fs.WriteFile(ctx, "/stow/pkg/file3", []byte("content3"), 0644)
    
    ops := []domain.Operation{
        &domain.LinkCreate{ID: "link1", Source: mustPath("/stow/pkg/file1"), Target: mustPath("/home/file1")},
        &domain.LinkCreate{ID: "link2", Source: mustPath("/stow/pkg/file2"), Target: mustPath("/home/file2")},
        &domain.LinkCreate{ID: "link3", Source: mustPath("/stow/pkg/file3"), Target: mustPath("/home/file3")},
    }
    
    checkpoint := exec.CheckpointStore().Create(ctx)
    result := exec.ExecuteBatch(ctx, ops, checkpoint)
    
    require.Len(t, result.Executed, 2)
    require.Len(t, result.Failed, 1)
    require.Contains(t, result.Failed, domain.OperationID("link2"))
}
```

**Implementation:** `internal/executor/parallel.go`

```go
package executor

import (
    "context"
    "sync"
    "github.com/yourorg/dot/internal/domain"
)

func (e *Executor) executeBatch(ctx context.Context, batch []domain.Operation, checkpoint *Checkpoint) ExecutionResult {
    result := ExecutionResult{
        Executed: make([]domain.OperationID, 0),
        Failed:   make([]domain.OperationID, 0),
        Errors:   make([]error, 0),
    }
    var mu sync.Mutex
    
    var wg sync.WaitGroup
    for _, op := range batch {
        wg.Add(1)
        go func(operation domain.Operation) {
            defer wg.Done()
            
            opID := operation.ID()
            
            e.log.Debug(ctx, "executing_operation_parallel", 
                "op_id", opID, 
                "op_kind", operation.Kind())
            
            if err := operation.Execute(ctx, e.fs); err != nil {
                e.log.Error(ctx, "operation_failed", "op_id", opID, "error", err)
                mu.Lock()
                result.Failed = append(result.Failed, opID)
                result.Errors = append(result.Errors, err)
                mu.Unlock()
                return
            }
            
            mu.Lock()
            result.Executed = append(result.Executed, opID)
            checkpoint.Record(opID, operation)
            mu.Unlock()
        }(op)
    }
    
    wg.Wait()
    return result
}
```

**Tasks:**
- [ ] Implement `executeBatch()` for concurrent execution
- [ ] Use `sync.WaitGroup` for goroutine synchronization
- [ ] Protect shared result with `sync.Mutex`
- [ ] Execute each operation in goroutine
- [ ] Collect executed operations thread-safely
- [ ] Collect failed operations thread-safely
- [ ] Record to checkpoint under lock
- [ ] Write tests for concurrent execution
- [ ] Write tests for partial failure in batch
- [ ] Test with race detector (`go test -race`)

**Commit:** `feat(executor): implement parallel batch execution`

### 10.4.2: Integrate Parallel Execution

**Test:** `internal/executor/executor_test.go`

```go
func TestExecute_ParallelBatches(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: noop.NewTracer(),
    })
    
    // Create plan with parallelizable operations
    fs.MkdirAll(ctx, "/home/dir1", 0755)
    fs.MkdirAll(ctx, "/home/dir2", 0755)
    fs.WriteFile(ctx, "/stow/pkg/file1", []byte("c1"), 0644)
    fs.WriteFile(ctx, "/stow/pkg/file2", []byte("c2"), 0644)
    
    ops := []domain.Operation{
        // Batch 1: Independent file links
        &domain.LinkCreate{
            ID: "link1", 
            Source: mustPath("/stow/pkg/file1"), 
            Target: mustPath("/home/dir1/file1"),
        },
        &domain.LinkCreate{
            ID: "link2", 
            Source: mustPath("/stow/pkg/file2"), 
            Target: mustPath("/home/dir2/file2"),
        },
    }
    
    plan := mustPlanWithParallelization(ops)
    
    result := exec.Execute(ctx, plan)
    require.NoError(t, result.Unwrap())
    
    execResult, _ := result.Unwrap()
    require.Len(t, execResult.Executed, 2)
}
```

**Implementation:** Update `internal/executor/executor.go`

```go
func (e *Executor) commit(ctx context.Context, plan domain.Plan, checkpoint *Checkpoint) ExecutionResult {
    // Check if plan supports parallelization
    if plan.CanParallelize() {
        return e.executeParallel(ctx, plan, checkpoint)
    }
    return e.executeSequential(ctx, plan, checkpoint)
}

func (e *Executor) executeParallel(ctx context.Context, plan domain.Plan, checkpoint *Checkpoint) ExecutionResult {
    batches := plan.ParallelBatches()
    
    e.log.Info(ctx, "executing_parallel", 
        "batch_count", len(batches),
        "total_operations", len(plan.Operations()))
    
    result := ExecutionResult{
        Executed: make([]domain.OperationID, 0),
        Failed:   make([]domain.OperationID, 0),
        Errors:   make([]error, 0),
    }
    
    for i, batch := range batches {
        e.log.Debug(ctx, "executing_batch", "batch", i, "size", len(batch))
        
        batchResult := e.executeBatch(ctx, batch, checkpoint)
        
        result.Executed = append(result.Executed, batchResult.Executed...)
        result.Failed = append(result.Failed, batchResult.Failed...)
        result.Errors = append(result.Errors, batchResult.Errors...)
        
        if len(batchResult.Failed) > 0 {
            // Stop on first batch failure
            e.log.Error(ctx, "batch_failed", "batch", i, "failures", len(batchResult.Failed))
            break
        }
    }
    
    return result
}

func (e *Executor) executeSequential(ctx context.Context, plan domain.Plan, checkpoint *Checkpoint) ExecutionResult {
    // Existing sequential implementation with checkpoint recording
    result := ExecutionResult{
        Executed: make([]domain.OperationID, 0),
        Failed:   make([]domain.OperationID, 0),
        Errors:   make([]error, 0),
    }
    
    for _, op := range plan.Operations() {
        opID := op.ID()
        
        e.log.Debug(ctx, "executing_operation", "op_id", opID, "op_kind", op.Kind())
        
        if err := op.Execute(ctx, e.fs); err != nil {
            e.log.Error(ctx, "operation_failed", "op_id", opID, "error", err)
            result.Failed = append(result.Failed, opID)
            result.Errors = append(result.Errors, err)
            break
        }
        
        result.Executed = append(result.Executed, opID)
        checkpoint.Record(opID, op)
    }
    
    return result
}
```

**Tasks:**
- [ ] Update `commit()` to check `plan.CanParallelize()`
- [ ] Implement `executeParallel()` method
- [ ] Get parallel batches from plan
- [ ] Execute each batch using `executeBatch()`
- [ ] Aggregate results from all batches
- [ ] Stop on first batch failure
- [ ] Update `executeSequential()` to record to checkpoint
- [ ] Write tests for parallel execution
- [ ] Write tests for sequential fallback
- [ ] Test batch-level failure handling

**Commit:** `feat(executor): integrate parallel execution with batching`

---

## 10.5: Instrumentation

**Goal:** Add comprehensive observability through tracing, metrics, and logging.

### 10.5.1: Detailed Tracing

**Test:** `internal/executor/tracing_test.go`

```go
func TestExecute_Tracing(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    
    // Use mock tracer that captures spans
    tracer := testutil.NewMockTracer()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: tracer,
    })
    
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
    }
    
    plan := mustPlan([]domain.Operation{op})
    
    exec.Execute(ctx, plan)
    
    // Verify spans were created
    spans := tracer.CapturedSpans()
    require.NotEmpty(t, spans)
    
    // Check for expected span names
    spanNames := extractSpanNames(spans)
    require.Contains(t, spanNames, "executor.Execute")
    require.Contains(t, spanNames, "executor.Prepare")
    require.Contains(t, spanNames, "operation.Execute")
}

func TestExecute_TracingAttributes(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    tracer := testutil.NewMockTracer()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: noop.NewLogger(),
        Tracer: tracer,
    })
    
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
    }
    plan := mustPlan([]domain.Operation{op})
    
    exec.Execute(ctx, plan)
    
    // Find executor.Execute span
    executeSpan := tracer.FindSpan("executor.Execute")
    require.NotNil(t, executeSpan)
    
    // Verify attributes
    attrs := executeSpan.Attributes()
    require.Equal(t, 1, attrs["operation_count"])
    require.NotEmpty(t, attrs["plan_checksum"])
}
```

**Implementation:** Update `internal/executor/executor.go` and `parallel.go`

```go
func (e *Executor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
    ctx, span := e.tracer.Start(ctx, "executor.Execute",
        trace.WithAttributes(
            attribute.Int("operation_count", len(plan.Operations())),
            attribute.String("plan_checksum", plan.Checksum()),
        ),
    )
    defer span.End()
    
    // ... existing implementation ...
    
    if err := e.prepare(ctx, plan); err != nil {
        span.RecordError(err)
        span.SetAttributes(attribute.String("failure_phase", "prepare"))
        return domain.Err[ExecutionResult](err)
    }
    
    // ... rest of implementation ...
    
    if len(result.Failed) > 0 {
        span.SetAttributes(
            attribute.Int("executed", len(result.Executed)),
            attribute.Int("failed", len(result.Failed)),
            attribute.Int("rolled_back", len(result.RolledBack)),
        )
        span.RecordError(domain.ErrExecutionFailed{...})
        // ...
    }
    
    span.SetAttributes(attribute.Int("operations_executed", len(result.Executed)))
    return domain.Ok(result)
}

func (e *Executor) prepare(ctx context.Context, plan domain.Plan) error {
    ctx, span := e.tracer.Start(ctx, "executor.Prepare",
        trace.WithAttributes(
            attribute.Int("operations", len(plan.Operations())),
        ),
    )
    defer span.End()
    
    // ... existing implementation ...
}

func (e *Executor) rollback(ctx context.Context, executed []domain.OperationID, checkpoint *Checkpoint) []domain.OperationID {
    ctx, span := e.tracer.Start(ctx, "executor.Rollback",
        trace.WithAttributes(
            attribute.Int("operations_to_rollback", len(executed)),
        ),
    )
    defer span.End()
    
    // ... existing implementation ...
    
    span.SetAttributes(
        attribute.Int("rolled_back", len(rolledBack)),
        attribute.Int("failed_rollback", len(executed)-len(rolledBack)),
    )
    
    return rolledBack
}

func (e *Executor) executeSequential(ctx context.Context, plan domain.Plan, checkpoint *Checkpoint) ExecutionResult {
    // Add span for each operation
    result := ExecutionResult{...}
    
    for _, op := range plan.Operations() {
        opID := op.ID()
        
        ctx, span := e.tracer.Start(ctx, "operation.Execute",
            trace.WithAttributes(
                attribute.String("op.id", string(opID)),
                attribute.String("op.kind", string(op.Kind())),
            ),
        )
        
        if err := op.Execute(ctx, e.fs); err != nil {
            span.RecordError(err)
            span.End()
            // ... handle error ...
        }
        
        span.End()
        // ...
    }
    
    return result
}
```

**Tasks:**
- [ ] Add span attributes to `Execute()` (operation count, checksum)
- [ ] Record errors in spans
- [ ] Add span for `prepare()` phase
- [ ] Add span for `rollback()` with rollback stats
- [ ] Add per-operation spans in `executeSequential()`
- [ ] Add per-operation spans in `executeBatch()`
- [ ] Include operation ID and kind in attributes
- [ ] Write tests with mock tracer
- [ ] Verify span hierarchy
- [ ] Verify attributes are set correctly

**Commit:** `feat(executor): add comprehensive tracing instrumentation`

### 10.5.2: Metrics Collection

**Test:** `internal/executor/metrics_test.go`

```go
func TestExecute_Metrics(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    
    metrics := testutil.NewMockMetrics()
    exec := executor.New(executor.Opts{
        FS:      fs,
        Logger:  noop.NewLogger(),
        Tracer:  noop.NewTracer(),
        Metrics: metrics,
    })
    
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
    }
    plan := mustPlan([]domain.Operation{op})
    
    exec.Execute(ctx, plan)
    
    // Verify metrics were recorded
    require.Equal(t, 1, metrics.CounterValue("executor.executions.total"))
    require.Equal(t, 1, metrics.CounterValue("executor.executions.success"))
    require.Equal(t, 1, metrics.CounterValue("executor.operations.executed"))
    require.NotZero(t, metrics.HistogramCount("executor.duration.seconds"))
}

func TestExecute_MetricsOnFailure(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    metrics := testutil.NewMockMetrics()
    exec := executor.New(executor.Opts{
        FS:      fs,
        Logger:  noop.NewLogger(),
        Tracer:  noop.NewTracer(),
        Metrics: metrics,
    })
    
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/nonexistent"),
        Target: mustPath("/home/file"),
    }
    plan := mustPlan([]domain.Operation{op})
    
    exec.Execute(ctx, plan)
    
    require.Equal(t, 1, metrics.CounterValue("executor.executions.total"))
    require.Equal(t, 1, metrics.CounterValue("executor.executions.failed"))
    require.Equal(t, 1, metrics.CounterValue("executor.operations.failed"))
}
```

**Implementation:** Create `internal/executor/instrumented.go`

```go
package executor

import (
    "context"
    "time"
    "github.com/yourorg/dot/internal/domain"
    "github.com/yourorg/dot/internal/ports"
)

// InstrumentedExecutor wraps Executor with metrics collection
type InstrumentedExecutor struct {
    inner   *Executor
    metrics ports.Metrics
}

func NewInstrumented(inner *Executor, metrics ports.Metrics) *InstrumentedExecutor {
    return &InstrumentedExecutor{
        inner:   inner,
        metrics: metrics,
    }
}

func (e *InstrumentedExecutor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
    start := time.Now()
    
    e.metrics.Counter("executor.executions.total").Inc()
    e.metrics.Gauge("executor.operations.queued").Set(float64(len(plan.Operations())))
    
    result := e.inner.Execute(ctx, plan)
    
    duration := time.Since(start)
    e.metrics.Histogram("executor.duration.seconds").Observe(duration.Seconds())
    
    if result.IsOk() {
        execResult := result.Value()
        e.metrics.Counter("executor.executions.success").Inc()
        e.metrics.Counter("executor.operations.executed").Add(float64(len(execResult.Executed)))
    } else {
        execResult, _ := result.Unwrap()
        e.metrics.Counter("executor.executions.failed").Inc()
        e.metrics.Counter("executor.operations.failed").Add(float64(len(execResult.Failed)))
        
        if len(execResult.RolledBack) > 0 {
            e.metrics.Counter("executor.operations.rolled_back").Add(float64(len(execResult.RolledBack)))
        }
    }
    
    return result
}
```

Update `internal/executor/executor.go`:

```go
type Opts struct {
    FS         ports.FS
    Logger     ports.Logger
    Tracer     ports.Tracer
    Metrics    ports.Metrics
    Checkpoint CheckpointStore
}

func New(opts Opts) *Executor {
    if opts.Checkpoint == nil {
        opts.Checkpoint = NewMemoryCheckpointStore()
    }
    
    executor := &Executor{
        fs:         opts.FS,
        log:        opts.Logger,
        tracer:     opts.Tracer,
        checkpoint: opts.Checkpoint,
    }
    
    // Wrap with metrics if provided
    if opts.Metrics != nil {
        return &InstrumentedExecutor{
            inner:   executor,
            metrics: opts.Metrics,
        }
    }
    
    return executor
}
```

**Tasks:**
- [ ] Create `InstrumentedExecutor` wrapper
- [ ] Add `executor.executions.total` counter
- [ ] Add `executor.executions.success` counter
- [ ] Add `executor.executions.failed` counter
- [ ] Add `executor.operations.executed` counter
- [ ] Add `executor.operations.failed` counter
- [ ] Add `executor.operations.rolled_back` counter
- [ ] Add `executor.operations.queued` gauge
- [ ] Add `executor.duration.seconds` histogram
- [ ] Update `New()` to wrap with metrics when provided
- [ ] Write metrics tests
- [ ] Verify counters increment correctly
- [ ] Verify histogram records durations

**Commit:** `feat(executor): add metrics collection`

### 10.5.3: Enhanced Logging

**Test:** `internal/executor/logging_test.go`

```go
func TestExecute_Logging(t *testing.T) {
    ctx := context.Background()
    fs := memfs.New()
    
    logger := testutil.NewMockLogger()
    exec := executor.New(executor.Opts{
        FS:     fs,
        Logger: logger,
        Tracer: noop.NewTracer(),
    })
    
    fs.WriteFile(ctx, "/stow/pkg/file", []byte("content"), 0644)
    op := &domain.LinkCreate{
        ID:     "link1",
        Source: mustPath("/stow/pkg/file"),
        Target: mustPath("/home/file"),
    }
    plan := mustPlan([]domain.Operation{op})
    
    exec.Execute(ctx, plan)
    
    // Verify log messages
    logs := logger.CapturedLogs()
    require.NotEmpty(t, logs)
    
    // Check for expected log messages
    require.Contains(t, logs, testutil.LogEntry{
        Level:   "info",
        Message: "executing_plan",
    })
    require.Contains(t, logs, testutil.LogEntry{
        Level:   "info",
        Message: "execution_complete",
    })
}
```

**Implementation:** Logging is already integrated in previous implementations. Add additional context:

```go
func (e *Executor) Execute(ctx context.Context, plan domain.Plan) domain.Result[ExecutionResult] {
    ctx, span := e.tracer.Start(ctx, "executor.Execute", ...)
    defer span.End()
    
    e.log.Info(ctx, "executing_plan",
        "operation_count", len(plan.Operations()),
        "plan_checksum", plan.Checksum(),
        "can_parallelize", plan.CanParallelize())
    
    // ... implementation ...
}

func (e *Executor) executeParallel(ctx context.Context, plan domain.Plan, checkpoint *Checkpoint) ExecutionResult {
    batches := plan.ParallelBatches()
    
    e.log.Info(ctx, "executing_parallel",
        "batch_count", len(batches),
        "total_operations", len(plan.Operations()))
    
    for i, batch := range batches {
        e.log.Debug(ctx, "executing_batch",
            "batch_index", i,
            "batch_size", len(batch))
        
        // ... execute batch ...
        
        if len(batchResult.Failed) > 0 {
            e.log.Error(ctx, "batch_failed",
                "batch_index", i,
                "failures", len(batchResult.Failed),
                "succeeded", len(batchResult.Executed))
        }
    }
    
    // ...
}
```

**Tasks:**
- [ ] Review all log statements for completeness
- [ ] Ensure structured fields in all log calls
- [ ] Add timing information where useful
- [ ] Include relevant IDs (operation, checkpoint, batch)
- [ ] Use appropriate log levels (Debug, Info, Warn, Error)
- [ ] Write tests with mock logger
- [ ] Verify log messages are emitted at correct points
- [ ] Verify structured fields are present

**Commit:** `feat(executor): enhance structured logging`

---

## Testing Strategy

### Unit Tests
- Test each component in isolation
- Use memory filesystem for all filesystem operations
- Mock logger, tracer, metrics for verification
- Test success and failure paths
- Test edge cases (empty plan, single operation, many operations)

### Integration Tests
- Test complete execute flow with real implementations
- Test two-phase commit
- Test rollback scenarios
- Test parallel execution
- Test observability integration

### Race Detection
- Run all tests with `-race` flag
- Verify parallel execution is thread-safe
- Check checkpoint recording concurrency

### Property-Based Tests
- Verify rollback restores original state
- Verify parallel and sequential produce same result
- Verify checkpoint completeness

## Success Criteria

Phase 10 is complete when:

- [ ] All subtasks implemented and tested
- [ ] Executor executes plans with operations
- [ ] Two-phase commit implemented (prepare + commit)
- [ ] Checkpoint system working
- [ ] Automatic rollback on failure
- [ ] Parallel execution based on dependency graph
- [ ] Full tracing instrumentation
- [ ] Metrics collection
- [ ] Structured logging throughout
- [ ] Test coverage ≥ 80%
- [ ] All tests pass including race detector
- [ ] golangci-lint passes without warnings
- [ ] CHANGELOG.md updated
- [ ] Documentation updated

## Estimated Effort

- 10.1: Basic Execution - 1 day
- 10.2: Two-Phase Commit - 2 days
- 10.3: Rollback Mechanism - 1.5 days
- 10.4: Parallel Execution - 1.5 days
- 10.5: Instrumentation - 1 day

**Total: ~7 days**

## Next Phase

After Phase 10 completion, proceed to **Phase 11: Manifest and State Management** for incremental operation support.

