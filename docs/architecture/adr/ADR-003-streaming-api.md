# ADR-003: Streaming API for Large Operations

**Status**: Proposed  
**Date**: 2025-10-07  
**Deciders**: Development Team  
**Context**: Phase 22.6 Future Enhancements Planning

## Context

The current API returns all operations in memory, which works well for typical dotfile setups with hundreds of files. However, this approach has limitations for very large package sets:

**Memory Constraints**:
- Package with 10,000 files → ~1-2 MB Plan size in memory
- Multiple packages → Memory usage scales linearly
- Large operations block until complete

**Current API**:
```go
func (c Client) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
    // Returns complete Plan with all operations loaded
}
```

**Problem Scenarios**:
1. Managing 100+ packages with 1000+ files each
2. Limited memory environments (containers, embedded systems)
3. Users who want progress updates during planning
4. Long-running operations that could be interrupted

## Decision

Implement **streaming API using Go channels** for memory-efficient operation processing while maintaining the existing batch API for simplicity.

### Proposed API

```go
// StreamingPlan represents a stream of operations
type StreamingPlan struct {
    Operations <-chan Operation // Stream of operations
    Errors     <-chan error     // Stream of errors
    Metadata   PlanMetadata     // Available immediately
    Done       <-chan struct{}  // Signals completion
}

// PlanManageStreaming plans package installation with streaming results
func (c Client) PlanManageStreaming(ctx context.Context, packages ...string) StreamingPlan

// Example usage:
plan := client.PlanManageStreaming(ctx, "large-package")
for op := range plan.Operations {
    // Process operation incrementally
    result := executor.Execute(ctx, op)
    // Can show progress, cancel early, etc.
}
```

### Architecture

```
Scanner → Channel → Planner → Channel → Resolver → Channel → Consumer
   ↓                   ↓                    ↓                    ↓
Discover           Compute             Resolve              Execute
files              state               conflicts            operations
```

### Implementation Phases

**Phase A: Streaming Scanner** (2-3 hours)
```go
func (s *Scanner) StreamPackage(ctx context.Context, pkg PackagePath) <-chan File {
    files := make(chan File)
    go func() {
        defer close(files)
        // Walk package directory, emit files incrementally
    }()
    return files
}
```

**Phase B: Streaming Planner** (3-4 hours)
```go
func StreamOperations(files <-chan File, targetDir TargetPath) <-chan Operation {
    ops := make(chan Operation)
    go func() {
        defer close(ops)
        // Convert files to operations incrementally
    }()
    return ops
}
```

**Phase C: Streaming Executor** (4-5 hours)
```go
func (e *Executor) ExecuteStream(ctx context.Context, ops <-chan Operation) <-chan Result {
    results := make(chan Result)
    go func() {
        defer close(results)
        // Execute operations as they arrive
    }()
    return results
}
```

**Phase D: Streaming Manifest Updates** (2-3 hours)
- Incremental manifest updates as operations complete
- Atomic save at end
- Rollback on failure

**Total Estimated Effort**: 11-15 hours

## Alternatives Considered

### Option A: Iterator Pattern
```go
type OperationIterator interface {
    Next() (Operation, bool)
    Error() error
}

func (c Client) PlanManageIter(ctx context.Context, packages ...string) OperationIterator
```

**Rejected**: More complex API, less idiomatic in Go, harder to use with goroutines.

### Option B: Batch Processing
```go
func (c Client) PlanManageBatches(ctx context.Context, packages ...string) ([][]Operation, error)
```

**Rejected**: Still loads all operations in memory, just organized differently.

### Option C: Callback Pattern
```go
func (c Client) PlanManageWithCallback(ctx context.Context, packages []string, fn func(Operation)) error
```

**Rejected**: Harder to test, less flexible, doesn't compose well.

### Option D: Streaming API with Channels ✅
**Selected**: Idiomatic Go, memory efficient, composable, cancellable.

## Implementation Strategy

### Backward Compatibility

Keep existing batch API:
```go
// Existing - stays unchanged
func (c Client) PlanManage(ctx context.Context, packages ...string) (Plan, error)

// New - streaming variant
func (c Client) PlanManageStreaming(ctx context.Context, packages ...string) StreamingPlan
```

Users choose based on their needs:
- Small packages: Use batch API (simpler)
- Large packages: Use streaming API (memory efficient)

### Error Handling

```go
plan := client.PlanManageStreaming(ctx, "large-package")

for {
    select {
    case op, ok := <-plan.Operations:
        if !ok {
            goto done
        }
        // Process operation
    case err := <-plan.Errors:
        // Handle error (but continue processing)
        log.Warn("operation error", err)
    case <-plan.Done:
        goto done
    }
}
done:
```

### Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

plan := client.PlanManageStreaming(ctx, packages...)

// User presses Ctrl+C
signal.Notify(sigChan, os.Interrupt)
go func() {
    <-sigChan
    cancel() // Stops streaming
}()

for op := range plan.Operations {
    // Stream stops when context cancelled
}
```

## Benefits

### Performance
- ✅ Constant memory usage regardless of package size
- ✅ Can start processing before scanning completes
- ✅ Better CPU utilization with pipelined processing
- ✅ Enables progress bars during long operations

### User Experience
- ✅ Progress updates for large operations
- ✅ Can cancel long-running operations
- ✅ Faster time-to-first-result

### Developer Experience
- ✅ Composable with other stream processing
- ✅ Testable with buffered channels
- ✅ Familiar Go idioms

## Trade-offs

### Complexity
- ❌ More complex implementation (goroutines, channels)
- ❌ Harder to debug (concurrent execution)
- ❌ Requires careful resource management (channel closure)

### Testing
- ❌ Harder to test (need to drain channels)
- ❌ Race conditions possible if not careful
- ❌ More edge cases (channel full, slow consumer)

### API Surface
- ❌ Two APIs for same functionality
- ❌ Users must choose which to use
- ❌ Documentation complexity

## Mitigation

### Complexity
- Provide well-tested streaming primitives
- Document goroutine lifecycle clearly
- Use defer for cleanup
- Comprehensive examples

### Testing
- Unit test each stage independently
- Integration test full pipeline
- Benchmark memory usage
- Test cancellation scenarios

### API Surface
- Clear documentation on when to use each API
- Default to batch API in examples
- Only recommend streaming for known large cases

## Decision Criteria

Implement streaming API when:
- ✅ User has 1000+ files in a package
- ✅ User needs progress updates
- ✅ Memory constraints exist
- ✅ User wants early cancellation

Use batch API when:
- ✅ Typical dotfile setup (< 1000 files)
- ✅ Simplicity preferred
- ✅ All-or-nothing semantics desired

## Migration Path

### Adding Streaming Support

```go
// Phase 1: Add StreamingPlan type
type StreamingPlan struct {
    Operations <-chan Operation
    Errors     <-chan error
    Metadata   PlanMetadata
    Done       <-chan struct{}
}

// Phase 2: Add streaming methods to Client interface
type Client interface {
    // ... existing methods ...
    
    PlanManageStreaming(ctx context.Context, packages ...string) StreamingPlan
    PlanUnmanageStreaming(ctx context.Context, packages ...string) StreamingPlan
}

// Phase 3: Implement in internal/api
// Phase 4: Add CLI support (--stream flag)
// Phase 5: Documentation and examples
```

### Testing Strategy

```go
func TestStreamingPlan(t *testing.T) {
    plan := client.PlanManageStreaming(ctx, "large-package")
    
    ops := collectFromChannel(plan.Operations)
    assert.NotEmpty(t, ops)
    
    <-plan.Done // Wait for completion
}

func collectFromChannel(ch <-chan Operation) []Operation {
    var ops []Operation
    for op := range ch {
        ops = append(ops, op)
    }
    return ops
}
```

## Success Criteria

- [ ] Streaming API defined and documented
- [ ] Memory usage constant for large packages
- [ ] Can cancel during execution
- [ ] Progress updates available
- [ ] All tests passing
- [ ] Benchmarks show memory improvement
- [ ] Documentation with examples

## References

- **Go Concurrency Patterns**: https://go.dev/blog/pipelines
- **Channel Best Practices**: https://go.dev/ref/mem
- **Context Cancellation**: https://go.dev/blog/context

## Timeline

**Deferred to Future Phase**: Not critical for v0.2.0 release  
**Estimated Implementation**: 11-15 hours  
**Target**: v0.3.0 or later

## Review

**Status**: Design complete, implementation deferred  
**Next Steps**: Implement if user demand for large package support  
**Metrics to Track**: Memory usage reports, package sizes in the wild

