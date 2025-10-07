# Phase 9: Pipeline Orchestration

## Overview

Phase 9 implements the pipeline orchestration layer that composes all functional core stages (scanner, planner, resolver, sorter) into cohesive, type-safe pipelines. This phase transforms individual pure functions into complete workflows for stow, unstow, restow, and adopt operations.

**Status**: Not Started  
**Prerequisites**: Phases 1-8 complete (Domain Model, Ports, Adapters, Scanner, Ignore System, Planner, Resolver, Sorter)  
**Deliverable**: Working pipeline engine composing all functional stages

## Architecture Context

From Architecture.md (lines 494-558), the pipeline system provides:
- Generic `Pipeline[A, B]` function type for composability
- `Compose()` for sequential pipeline composition
- `Parallel()` for concurrent pipeline execution
- Type-safe stage composition with Result[T] propagation

## Phase Breakdown

### 9.1: Pipeline Types and Composition

Implement generic pipeline abstractions and composition functions.

#### Tasks

**9.1.1: Core Pipeline Type**
- Define `Pipeline[A, B]` as `func(context.Context, A) Result[B]`
- Add package documentation explaining pipeline concept
- Create `pkg/dot/pipeline.go` or `internal/pipeline/types.go`
- Write examples of pipeline usage in comments

**9.1.2: Sequential Composition**
- Implement `Compose[A, B, C](p1 Pipeline[A, B], p2 Pipeline[B, C]) Pipeline[A, C]`
- Ensure proper Result[T] error propagation
- Add context cancellation checking
- Verify short-circuit on first error
- Write unit tests for composition

**9.1.3: Parallel Composition**
- Implement `Parallel[A, B](pipelines []Pipeline[A, B]) Pipeline[A, []B]`
- Use sync.WaitGroup for goroutine coordination
- Collect results from all parallel pipelines
- Aggregate errors using ErrMultiple
- Add context cancellation support
- Write tests for parallel execution

**9.1.4: Pipeline Utilities**
- Implement `Map[A, B](f func(A) B) Pipeline[A, B]` for lifting functions
- Implement `FlatMap[A, B](f func(A) Result[B]) Pipeline[A, B]`
- Add `Filter[A](pred func(A) bool) Pipeline[A, A]`
- Write tests for utility functions

**Tests**
```go
// Test sequential composition
func TestCompose(t *testing.T) {
    p1 := func(ctx context.Context, n int) Result[int] {
        return Ok(n + 1)
    }
    p2 := func(ctx context.Context, n int) Result[int] {
        return Ok(n * 2)
    }
    
    composed := Compose(p1, p2)
    result := composed(context.Background(), 5)
    
    require.True(t, result.IsOk())
    val, _ := result.Unwrap()
    assert.Equal(t, 12, val) // (5 + 1) * 2
}

// Test parallel composition
func TestParallel(t *testing.T) {
    pipelines := []Pipeline[int, int]{
        func(ctx context.Context, n int) Result[int] {
            return Ok(n + 1)
        },
        func(ctx context.Context, n int) Result[int] {
            return Ok(n * 2)
        },
    }
    
    parallel := Parallel(pipelines)
    result := parallel(context.Background(), 5)
    
    require.True(t, result.IsOk())
    vals, _ := result.Unwrap()
    assert.ElementsMatch(t, []int{6, 10}, vals)
}

// Test error propagation
func TestCompose_ErrorPropagation(t *testing.T) {
    p1 := func(ctx context.Context, n int) Result[int] {
        return Err[int](errors.New("first error"))
    }
    p2 := func(ctx context.Context, n int) Result[int] {
        t.Fatal("should not be called")
        return Ok(n)
    }
    
    composed := Compose(p1, p2)
    result := composed(context.Background(), 5)
    
    require.False(t, result.IsOk())
}

// Test context cancellation
func TestCompose_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    p1 := func(ctx context.Context, n int) Result[int] {
        select {
        case <-ctx.Done():
            return Err[int](ctx.Err())
        default:
            return Ok(n + 1)
        }
    }
    
    result := p1(ctx, 5)
    require.False(t, result.IsOk())
}
```

**Files Created**
- `internal/pipeline/types.go`
- `internal/pipeline/compose.go`
- `internal/pipeline/types_test.go`
- `internal/pipeline/compose_test.go`

---

### 9.2: Core Pipelines

Implement concrete pipelines for each operation type.

#### Tasks

**9.2.1: Stow Pipeline**
- Create `internal/pipeline/stow.go`
- Define `StowPipeline` struct with scan, plan, resolve, order stages
- Implement `NewStowPipeline(opts PipelineOpts) *StowPipeline`
- Define `PipelineOpts` struct with FS, Ignore, LinkMode, Folding, Policies
- Implement `Execute(ctx context.Context, input ScanInput) Result[Plan]`
- Compose scan -> plan -> resolve -> order stages
- Write unit tests with mock stages

**9.2.2: Unstow Pipeline**
- Create `internal/pipeline/unstow.go`
- Define `UnstowPipeline` struct
- Implement unstow-specific planning logic
- Compose scan -> plan (for deletion) -> resolve -> order
- Write unit tests

**9.2.3: Restow Pipeline**
- Create `internal/pipeline/restow.go`
- Define `RestowPipeline` struct
- Integrate IncrementalPlanner for changed package detection
- Implement incremental restow logic
- Combine unstow and stow operations
- Write unit tests verifying incremental behavior

**9.2.4: Adopt Pipeline**
- Create `internal/pipeline/adopt.go`
- Define `AdoptPipeline` struct
- Implement adoption-specific logic
- Generate FileMove + LinkCreate operations
- Compose adoption stages
- Write unit tests

**9.2.5: Pipeline Options**
- Define `PipelineOpts` struct with all configuration
- Add validation for pipeline options
- Implement builder pattern for options
- Write tests for option validation

**Tests**
```go
// Test StowPipeline execution
func TestStowPipeline_Execute(t *testing.T) {
    fs := memfs.New()
    setupFixtures(fs)
    
    opts := PipelineOpts{
        FS:       fs,
        Ignore:   DefaultIgnoreSet(),
        LinkMode: LinkRelative,
        Folding:  true,
        Policies: DefaultResolutionPolicies(),
    }
    
    pipeline := NewStowPipeline(opts)
    
    input := ScanInput{
        PackageDir:   mustParsePath("/stow"),
        TargetDir: mustParsePath("/home/user"),
        Packages:  []string{"vim", "git"},
    }
    
    result := pipeline.Execute(context.Background(), input)
    
    require.True(t, result.IsOk())
    plan, _ := result.Unwrap()
    assert.NotEmpty(t, plan.Operations())
}

// Test pipeline stage integration
func TestStowPipeline_StageIntegration(t *testing.T) {
    // Test that scan -> plan -> resolve -> order stages work together
    // Verify data flows correctly through stages
    // Check that each stage receives correct input type
}

// Test RestowPipeline incremental behavior
func TestRestowPipeline_IncrementalDetection(t *testing.T) {
    fs := memfs.New()
    
    // First stow
    pipeline := NewRestowPipeline(opts)
    input := ScanInput{...}
    result1 := pipeline.Execute(context.Background(), input)
    require.True(t, result1.IsOk())
    
    // Restow with no changes
    result2 := pipeline.Execute(context.Background(), input)
    require.True(t, result2.IsOk())
    plan2, _ := result2.Unwrap()
    
    // Should detect no changes and generate empty plan
    assert.Empty(t, plan2.Operations())
}

// Test AdoptPipeline generates correct operations
func TestAdoptPipeline_Operations(t *testing.T) {
    fs := memfs.New()
    // Write files to target
    writeFile(fs, "/home/user/.bashrc", []byte("content"))
    
    pipeline := NewAdoptPipeline(opts)
    input := AdoptInput{
        Package:   "bash",
        Files:     []string{".bashrc"},
        PackageDir:   mustParsePath("/stow"),
        TargetDir: mustParsePath("/home/user"),
    }
    
    result := pipeline.Execute(context.Background(), input)
    
    require.True(t, result.IsOk())
    plan, _ := result.Unwrap()
    
    // Verify FileMove + LinkCreate operations generated
    assert.Len(t, plan.Operations(), 2)
}
```

**Files Created**
- `internal/pipeline/stow.go`
- `internal/pipeline/unstow.go`
- `internal/pipeline/restow.go`
- `internal/pipeline/adopt.go`
- `internal/pipeline/opts.go`
- `internal/pipeline/stow_test.go`
- `internal/pipeline/unstow_test.go`
- `internal/pipeline/restow_test.go`
- `internal/pipeline/adopt_test.go`

---

### 9.3: Pipeline Engine

Implement orchestration layer with observability integration.

#### Tasks

**9.3.1: Engine Structure**
- Create `internal/pipeline/engine.go`
- Define `Engine` struct with dependencies (FS, Logger, Tracer, Metrics)
- Implement `New(deps Dependencies) *Engine`
- Add pipeline registration map
- Write tests for engine construction

**9.3.2: Pipeline Execution**
- Implement `Execute(ctx context.Context, pipeline Pipeline[A, B], input A) Result[B]`
- Add context propagation throughout execution
- Implement timeout handling
- Add panic recovery with error conversion
- Write tests for execution

**9.3.3: Error Handling**
- Implement error wrapping with context
- Add error aggregation for multiple failures
- Implement user-friendly error formatting
- Add error logging with structured fields
- Write tests for error scenarios

**9.3.4: Observability Integration**
- Add tracing spans for pipeline execution
- Implement metrics collection (duration, success/failure counts)
- Add structured logging at pipeline boundaries
- Record pipeline input/output metadata
- Write tests for observability

**9.3.5: Status and Progress**
- Define `PipelineStatus` type for progress tracking
- Implement progress reporting channel
- Add status callbacks for long-running operations
- Implement cancellation support
- Write tests for progress reporting

**Tests**
```go
// Test engine execution
func TestEngine_Execute(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    tracer := &NoOpTracer{}
    metrics := &NoOpMetrics{}
    
    engine := New(Dependencies{
        Logger:  logger,
        Tracer:  tracer,
        Metrics: metrics,
    })
    
    pipeline := func(ctx context.Context, n int) Result[int] {
        return Ok(n + 1)
    }
    
    result := engine.Execute(context.Background(), pipeline, 5)
    
    require.True(t, result.IsOk())
    val, _ := result.Unwrap()
    assert.Equal(t, 6, val)
}

// Test error handling
func TestEngine_ErrorHandling(t *testing.T) {
    engine := New(deps)
    
    pipeline := func(ctx context.Context, n int) Result[int] {
        return Err[int](errors.New("test error"))
    }
    
    result := engine.Execute(context.Background(), pipeline, 5)
    
    require.False(t, result.IsOk())
    _, err := result.Unwrap()
    assert.Error(t, err)
}

// Test panic recovery
func TestEngine_PanicRecovery(t *testing.T) {
    engine := New(deps)
    
    pipeline := func(ctx context.Context, n int) Result[int] {
        panic("unexpected panic")
    }
    
    result := engine.Execute(context.Background(), pipeline, 5)
    
    require.False(t, result.IsOk())
    _, err := result.Unwrap()
    assert.Contains(t, err.Error(), "panic")
}

// Test context cancellation
func TestEngine_ContextCancellation(t *testing.T) {
    engine := New(deps)
    
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    pipeline := func(ctx context.Context, n int) Result[int] {
        time.Sleep(100 * time.Millisecond)
        return Ok(n)
    }
    
    result := engine.Execute(ctx, pipeline, 5)
    
    require.False(t, result.IsOk())
}

// Test tracing integration
func TestEngine_Tracing(t *testing.T) {
    tracer := &MockTracer{}
    engine := New(Dependencies{Tracer: tracer, ...})
    
    pipeline := func(ctx context.Context, n int) Result[int] {
        return Ok(n + 1)
    }
    
    result := engine.Execute(context.Background(), pipeline, 5)
    
    require.True(t, result.IsOk())
    assert.True(t, tracer.SpanStarted("pipeline.Execute"))
    assert.True(t, tracer.SpanEnded())
}

// Test metrics collection
func TestEngine_Metrics(t *testing.T) {
    metrics := &MockMetrics{}
    engine := New(Dependencies{Metrics: metrics, ...})
    
    pipeline := func(ctx context.Context, n int) Result[int] {
        return Ok(n + 1)
    }
    
    result := engine.Execute(context.Background(), pipeline, 5)
    
    require.True(t, result.IsOk())
    assert.Equal(t, 1, metrics.CounterValue("pipeline.executions.total"))
    assert.Equal(t, 1, metrics.CounterValue("pipeline.executions.success"))
}
```

**Files Created**
- `internal/pipeline/engine.go`
- `internal/pipeline/engine_test.go`
- `internal/pipeline/status.go`
- `internal/pipeline/status_test.go`

---

### 9.4: Integration Testing

Comprehensive tests for complete pipeline flows.

#### Tasks

**9.4.1: End-to-End Pipeline Tests**
- Create `internal/pipeline/integration_test.go`
- Test complete stow pipeline with real filesystem
- Test unstow pipeline removes correct links
- Test restow pipeline with incremental changes
- Test adopt pipeline moves and links files
- Write tests for each pipeline

**9.4.2: Error Propagation Tests**
- Test scanner errors propagate through pipeline
- Test planner errors propagate correctly
- Test resolver errors propagate correctly
- Test sorter errors propagate correctly
- Verify error messages are user-friendly

**9.4.3: Concurrent Pipeline Tests**
- Test multiple pipelines executing concurrently
- Verify thread safety of shared components
- Test context cancellation across pipelines
- Verify no race conditions (run with -race flag)

**9.4.4: Performance Tests**
- Benchmark pipeline execution time
- Measure memory usage during pipeline execution
- Test with large package sets (100+ packages)
- Profile pipeline performance
- Verify acceptable performance targets

**Tests**
```go
// Integration test for complete stow flow
func TestIntegration_StowPipeline(t *testing.T) {
    fs := memfs.New()
    
    // Setup package structure
    createPackage(fs, "/stow/vim", map[string]string{
        "vimrc": "set number",
        "ftplugin/go.vim": "setlocal noexpandtab",
    })
    
    opts := PipelineOpts{
        FS:       fs,
        Ignore:   DefaultIgnoreSet(),
        LinkMode: LinkRelative,
        Folding:  true,
        Policies: DefaultResolutionPolicies(),
    }
    
    pipeline := NewStowPipeline(opts)
    executor := executor.New(executor.Opts{FS: fs, ...})
    
    // Execute pipeline
    input := ScanInput{
        PackageDir:   mustParsePath("/stow"),
        TargetDir: mustParsePath("/home/user"),
        Packages:  []string{"vim"},
    }
    
    planResult := pipeline.Execute(context.Background(), input)
    require.True(t, planResult.IsOk())
    
    plan, _ := planResult.Unwrap()
    
    // Execute plan
    execResult := executor.Execute(context.Background(), plan)
    require.True(t, execResult.IsOk())
    
    // Verify links created
    assert.True(t, fs.IsSymlink(context.Background(), "/home/user/.vimrc"))
    assert.True(t, fs.IsSymlink(context.Background(), "/home/user/.vim/ftplugin/go.vim"))
}

// Test error propagation through pipeline
func TestIntegration_ErrorPropagation(t *testing.T) {
    fs := memfs.New()
    
    // Create conflicting file
    writeFile(fs, "/home/user/.vimrc", []byte("existing"))
    
    opts := PipelineOpts{
        FS:       fs,
        Policies: ResolutionPolicies{OnFileExists: PolicyFail},
        ...
    }
    
    pipeline := NewStowPipeline(opts)
    input := ScanInput{...}
    
    result := pipeline.Execute(context.Background(), input)
    
    // Should fail with conflict error
    require.False(t, result.IsOk())
    _, err := result.Unwrap()
    
    var conflictErr ErrConflict
    assert.ErrorAs(t, err, &conflictErr)
}

// Test concurrent pipeline execution
func TestIntegration_ConcurrentPipelines(t *testing.T) {
    fs := memfs.New()
    setupFixtures(fs)
    
    pipeline := NewStowPipeline(opts)
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            
            input := ScanInput{
                Packages: []string{fmt.Sprintf("pkg%d", idx)},
                ...
            }
            
            result := pipeline.Execute(context.Background(), input)
            require.True(t, result.IsOk())
        }(i)
    }
    
    wg.Wait()
}

// Performance benchmark
func BenchmarkStowPipeline(b *testing.B) {
    fs := memfs.New()
    setupLargeFixtures(fs, 100) // 100 packages
    
    pipeline := NewStowPipeline(opts)
    input := ScanInput{...}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := pipeline.Execute(context.Background(), input)
        if !result.IsOk() {
            b.Fatal("pipeline failed")
        }
    }
}
```

**Files Created**
- `internal/pipeline/integration_test.go`
- `internal/pipeline/benchmark_test.go`

---

## Implementation Order

1. **Pipeline Types** (9.1)
   - Define generic types
   - Implement Compose and Parallel
   - Write utility functions
   - Test composition thoroughly

2. **Core Pipelines** (9.2)
   - Implement StowPipeline first (most common use case)
   - Implement UnstowPipeline
   - Implement RestowPipeline with incremental support
   - Implement AdoptPipeline last

3. **Pipeline Engine** (9.3)
   - Build engine structure
   - Add execution logic
   - Integrate observability
   - Add error handling

4. **Integration Testing** (9.4)
   - Write end-to-end tests
   - Add concurrent tests
   - Performance benchmarks
   - Race condition testing

## Testing Strategy

### Unit Tests
- Test each pipeline composition function independently
- Mock dependencies using interfaces
- Verify error propagation
- Test context cancellation
- Target 80%+ coverage

### Integration Tests
- Test complete pipeline flows
- Use in-memory filesystem for speed
- Test error scenarios end-to-end
- Verify observability integration

### Property Tests
- Verify pipeline composition laws:
  - Associativity: `Compose(Compose(p1, p2), p3) = Compose(p1, Compose(p2, p3))`
  - Identity: `Compose(identity, p) = Compose(p, identity) = p`
- Test parallel execution produces same results as sequential

### Performance Tests
- Benchmark pipeline execution time
- Profile memory allocations
- Test with large package sets
- Verify scalability

## Acceptance Criteria

Phase 9 is complete when:

- [ ] All pipeline types and composition functions implemented
- [ ] Stow, Unstow, Restow, Adopt pipelines working
- [ ] Pipeline engine orchestrates execution
- [ ] Context propagation works throughout
- [ ] Error handling with Result[T] working
- [ ] Observability integrated (logging, tracing, metrics)
- [ ] Unit tests pass with 80%+ coverage
- [ ] Integration tests verify end-to-end flows
- [ ] Property tests verify composition laws
- [ ] Performance benchmarks show acceptable performance
- [ ] No race conditions detected with -race flag
- [ ] Documentation complete for all public APIs
- [ ] All linters pass
- [ ] Code reviewed and approved

## Dependencies

### Prerequisites (Must Be Complete)
- Phase 1: Domain Model (Result[T], Operation types)
- Phase 2: Infrastructure Ports (FS, Logger, Tracer, Metrics)
- Phase 3: Adapters (implementations of ports)
- Phase 4: Scanner (pure scanning logic)
- Phase 5: Ignore System (pattern matching)
- Phase 6: Planner (desired state computation)
- Phase 7: Resolver (conflict resolution)
- Phase 8: Sorter (topological ordering)

### Blocks
- Phase 10: Executor (needs Plan from pipelines)
- Phase 12: Public API (needs pipeline execution)
- Phase 13: CLI (needs complete pipeline system)

## Success Metrics

- Pipeline execution completes in <100ms for typical use case (5 packages)
- Memory usage stays bounded with streaming
- Zero race conditions detected
- All tests pass consistently
- Code coverage ≥ 80%
- Linter warnings = 0
- Integration tests cover all major workflows

## Risks and Mitigations

### Risk: Generic Type Complexity
**Impact**: High  
**Likelihood**: Medium  
**Mitigation**: 
- Start with simple examples
- Comprehensive documentation
- Clear naming conventions
- Extensive testing

### Risk: Error Propagation Issues
**Impact**: High  
**Likelihood**: Medium  
**Mitigation**:
- Thorough testing of error paths
- Verify all Result[T] handling
- Test error aggregation
- User-friendly error messages

### Risk: Context Cancellation Bugs
**Impact**: Medium  
**Likelihood**: Medium  
**Mitigation**:
- Test all context cancellation scenarios
- Verify cleanup on cancellation
- Use race detector
- Review all goroutine usage

### Risk: Performance Issues
**Impact**: Medium  
**Likelihood**: Low  
**Mitigation**:
- Early benchmarking
- Profile regularly
- Optimize hot paths
- Test with large datasets

## Notes

- Keep pipeline logic pure - only Engine should have side effects
- Maintain type safety throughout composition
- Result[T] monad ensures error handling
- Context threading enables cancellation and timeouts
- Observability built in from the start
- Test with -race flag throughout development
- Follow constitutional TDD principles
- Atomic commits with conventional messages

