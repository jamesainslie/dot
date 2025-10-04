# Phase 8: Functional Core - Topological Sorter

## Overview

Implement pure dependency graph analysis and topological sorting for operation ordering. This phase enables safe sequential and parallel execution by computing operation dependencies and detecting cycles.

## Objectives

- Build dependency graphs from operation lists
- Detect cyclic dependencies before execution
- Compute topologically sorted operation order
- Analyze parallelization opportunities
- Enable safe concurrent execution planning

## Architecture Context

Located in `internal/planner/` alongside desired state and resolver components. Pure functions with no I/O dependencies, operating on domain types only.

## Dependencies

**Required (Must be complete)**:
- Phase 1: Domain model (Operation interface, OperationID type, Result monad)
- Phase 6: Planner (produces operations that need ordering)

**Consumed By**:
- Phase 9: Pipeline orchestration (uses sorted plans)
- Phase 10: Executor (uses parallelization analysis)

## Implementation Tasks

### 8.1: Dependency Graph Construction

**File**: `internal/planner/graph.go`

#### 8.1.1: Core Types

```go
// DependencyGraph represents operation dependencies
type DependencyGraph struct {
    nodes map[dot.OperationID]dot.Operation
    edges map[dot.OperationID][]dot.OperationID
}
```

**Test**: `graph_test.go`
- Test empty graph construction
- Test single node graph
- Test graph with no edges (independent operations)
- Test graph with simple linear dependencies
- Test graph with complex dependency patterns

#### 8.1.2: Graph Builder

```go
func BuildGraph(ops []dot.Operation) *DependencyGraph
```

**Algorithm**:
1. Initialize graph with empty maps
2. For each operation:
   - Extract operation ID via type assertion
   - Add to nodes map
   - Extract dependencies from operation
   - Add to edges map

**Tests**: `graph_test.go::TestBuildGraph`
- Empty operation list returns empty graph
- Single operation with no dependencies
- Multiple independent operations
- Linear dependency chain (A→B→C)
- Tree structure dependencies
- Diamond dependency pattern (A→B,C; B,C→D)
- Operation with multiple dependencies
- Verify all operations present in nodes map
- Verify all edges correctly recorded

#### 8.1.3: Operation ID Extraction

```go
// Internal helper for type assertion
type hasID interface {
    ID() dot.OperationID
}

func getOperationID(op dot.Operation) dot.OperationID
```

**Tests**: `graph_test.go::TestGetOperationID`
- Extract ID from LinkCreate operation
- Extract ID from LinkDelete operation
- Extract ID from DirCreate operation
- Extract ID from all operation types
- Verify no panics on valid operations

**Note**: This may require augmenting the Operation interface or using type assertions. Consider adding ID() method to Operation interface if not already present.

### 8.2: Topological Sort with Cycle Detection

**File**: `internal/planner/topo.go`

#### 8.2.1: Topological Sort Algorithm

```go
func (g *DependencyGraph) TopologicalSort() ([]dot.Operation, error)
```

**Algorithm**: Depth-first search with visited tracking
1. Initialize visited and temp (recursion stack) sets
2. Initialize result slice
3. For each unvisited node:
   - Call recursive visit function
   - If cycle detected, return error
4. Return reversed result (DFS produces reverse order)

**Implementation Details**:
- Use DFS post-order traversal
- Track visited nodes to avoid reprocessing
- Track temporary marks to detect cycles
- Collect operations in post-order
- Reverse at end for correct dependency order

**Tests**: `topo_test.go::TestTopologicalSort`
- Empty graph returns empty slice
- Single node returns single element
- Two independent nodes (any order valid)
- Linear chain (A→B→C) returns [A, B, C]
- Tree dependencies sorted correctly
- Diamond pattern (A→B,C; B,C→D) returns valid order
- Multiple valid orders (verify one is returned)
- Large graph (100+ operations) performance test

#### 8.2.2: Cycle Detection

```go
func (g *DependencyGraph) FindCycle() []dot.OperationID
```

**Algorithm**: DFS with recursion stack tracking
1. Initialize visited set and recursion stack
2. Track parent pointers for path reconstruction
3. For each unvisited node:
   - Perform DFS maintaining recursion stack
   - If back edge detected (node in recursion stack):
     - Reconstruct cycle path using parent pointers
     - Return cycle
4. Return nil if no cycle found

**Tests**: `topo_test.go::TestFindCycle`
- Acyclic graph returns nil
- Self-loop detected (A→A)
- Simple cycle (A→B→A)
- Longer cycle (A→B→C→A)
- Cycle in larger graph with acyclic parts
- Multiple cycles (returns one)
- Verify returned cycle path is correct
- Verify cycle path starts and ends with same node

#### 8.2.3: Cycle Error Handling

**File**: `pkg/dot/errors.go` (if not already present)

```go
type ErrCyclicDependency struct {
    Cycle []OperationID
}

func (e ErrCyclicDependency) Error() string {
    return fmt.Sprintf("cyclic dependency detected: %v", e.Cycle)
}
```

**Integration**: TopologicalSort should call FindCycle and return ErrCyclicDependency if found.

**Tests**: `topo_test.go::TestTopologicalSortWithCycle`
- Cyclic graph returns error
- Error type is ErrCyclicDependency
- Error contains cycle information
- Error message is user-readable

### 8.3: Parallelization Analysis

**File**: `internal/planner/parallel.go`

#### 8.3.1: Level-Based Grouping

```go
func (g *DependencyGraph) ParallelizationPlan() [][]dot.Operation
```

**Algorithm**: Compute operation levels
1. For each operation, compute its level:
   - Level = max(dependency levels) + 1
   - Operations with no dependencies have level 0
2. Group operations by level
3. Operations at same level can execute in parallel
4. Return ordered batches (batch i must complete before batch i+1)

**Implementation Details**:
```go
func (g *DependencyGraph) computeLevel(id OperationID) int {
    // Memoization: check if already computed
    if level, exists := opLevels[id]; exists {
        return level
    }
    
    // Base case: no dependencies = level 0
    deps := g.edges[id]
    if len(deps) == 0 {
        opLevels[id] = 0
        return 0
    }
    
    // Recursive case: level = max(dep levels) + 1
    maxDepLevel := -1
    for _, dep := range deps {
        depLevel := g.computeLevel(dep)
        if depLevel > maxDepLevel {
            maxDepLevel = depLevel
        }
    }
    
    level := maxDepLevel + 1
    opLevels[id] = level
    return level
}
```

**Tests**: `parallel_test.go::TestParallelizationPlan`
- Empty graph returns empty batches
- Single operation returns single batch with one op
- Two independent ops return single batch with two ops
- Linear chain (A→B→C) returns three batches
- Diamond pattern batching:
  - Batch 0: [A]
  - Batch 1: [B, C]  (parallel)
  - Batch 2: [D]
- Complex graph produces correct batches
- All operations in a batch have no dependencies between them
- All dependencies of batch N are in batches 0..N-1

#### 8.3.2: Parallelization Validation

**Tests**: `parallel_test.go::TestParallelizationSafety`
- Verify no operation depends on operation in same batch
- Verify all dependencies satisfied by previous batches
- Property test: randomly generated graphs have safe parallelization
- Stress test: large graphs (1000+ operations)

#### 8.3.3: Batch Ordering

**Tests**: `parallel_test.go::TestBatchOrdering`
- Batches returned in dependency order
- Executing batches sequentially satisfies all dependencies
- No batch N+1 operation depends on batch N+2 operation

### 8.4: Integration with Operation Types

**File**: Update operation implementations as needed

Ensure all operation types correctly implement Dependencies() method:

#### LinkCreate Dependencies
- Depends on parent DirCreate if parent directory doesn't exist
- No dependency if target parent exists

#### LinkDelete Dependencies
- No dependencies (leaf operations)

#### DirCreate Dependencies
- Depends on parent DirCreate if parent doesn't exist
- No dependency if parent exists

#### DirDelete Dependencies
- Depends on all LinkDelete and DirDelete operations for children
- Must be last operations for a directory tree

#### FileMove Dependencies
- No dependencies (atomic operation)

#### FileBackup Dependencies
- Must occur before LinkCreate that would overwrite the file

**Tests**: `internal/planner/operation_deps_test.go`
- Test Dependencies() for each operation type
- Verify dependency computation is correct
- Test with mock filesystem states

## Testing Strategy

### Unit Tests

**Coverage Target**: 100% (this is pure logic, fully testable)

**Test Files**:
- `internal/planner/graph_test.go`: Graph construction
- `internal/planner/topo_test.go`: Sorting and cycle detection
- `internal/planner/parallel_test.go`: Parallelization analysis

**Test Patterns**:
- Table-driven tests for algorithm correctness
- Property-based tests for invariants
- Benchmark tests for performance
- Edge case coverage (empty, single, large)

### Property-Based Tests

**File**: `test/properties/graph_laws_test.go`

Properties to verify:
1. **Topological sort produces valid order**: Every operation appears after its dependencies
2. **Cycle detection is sound**: If FindCycle returns nil, TopologicalSort succeeds
3. **Cycle detection is complete**: If TopologicalSort fails, FindCycle returns a cycle
4. **Parallelization is safe**: No operation in a batch depends on another in same batch
5. **Parallelization is maximal**: Cannot move any operation to an earlier batch
6. **Idempotence**: Sorting twice produces same result

### Integration Tests

**File**: `internal/planner/integration_test.go`

Scenarios:
1. Build graph from planner output
2. Sort operations and verify execution order
3. Compute parallelization plan and verify safety
4. Handle cyclic dependencies from malformed plans
5. Large-scale graph (1000+ operations) performance

## Performance Requirements

- BuildGraph: O(n) where n = number of operations
- TopologicalSort: O(n + e) where e = number of edges
- FindCycle: O(n + e)
- ParallelizationPlan: O(n + e)
- Memory: O(n + e)

**Benchmarks**: `internal/planner/graph_bench_test.go`
- Benchmark graph construction
- Benchmark topological sort
- Benchmark parallelization plan
- Test with 100, 1000, 10000 operations

## Error Handling

### Error Types

**File**: `pkg/dot/errors.go`

```go
type ErrCyclicDependency struct {
    Cycle []OperationID
}

type ErrInvalidOperation struct {
    Op     Operation
    Reason string
}
```

### Error Cases

1. **Cyclic dependencies**: Return ErrCyclicDependency with cycle path
2. **Missing operation ID**: Panic (programming error)
3. **Invalid graph structure**: Return error

**Tests**: `internal/planner/errors_test.go`
- Test error types implement error interface
- Test error messages are descriptive
- Test error unwrapping works correctly

## Code Quality Requirements

### Linting
- All linters must pass
- Cyclomatic complexity ≤ 15
- No naked returns in functions > 10 lines
- All exported functions have documentation

### Documentation
- Package-level documentation explaining graph algorithms
- Function documentation with complexity analysis
- Example code for common usage patterns
- Document algorithm choices and trade-offs

### Code Style
- Pure functions (no side effects)
- Immutable data structures where possible
- Clear variable names
- Short functions (< 50 lines preferred)
- Use early returns for error cases

## Dependencies

### Standard Library
- None required (pure data structure manipulation)

### Internal Packages
- `pkg/dot`: Domain types (Operation, OperationID, Result)
- `internal/planner`: Shared planner types

### Testing Dependencies
- `github.com/stretchr/testify/assert`
- `github.com/stretchr/testify/require`
- `github.com/leanovate/gopter` (property-based testing)

## Implementation Order

### Week 1: Core Graph Structure
1. Create `internal/planner/graph.go` with types
2. Implement BuildGraph function
3. Write comprehensive tests for graph construction
4. Commit: `feat(planner): implement dependency graph construction`

### Week 1: Topological Sort
5. Create `internal/planner/topo.go`
6. Implement TopologicalSort algorithm
7. Implement FindCycle algorithm
8. Add ErrCyclicDependency type
9. Write comprehensive tests
10. Commit: `feat(planner): implement topological sort with cycle detection`

### Week 2: Parallelization
11. Create `internal/planner/parallel.go`
12. Implement ParallelizationPlan algorithm
13. Write comprehensive tests
14. Commit: `feat(planner): implement parallelization analysis`

### Week 2: Integration and Testing
15. Write property-based tests
16. Write integration tests
17. Write benchmarks
18. Profile and optimize hot paths
19. Commit: `test(planner): add property-based and integration tests`

### Week 2: Documentation
20. Add package documentation
21. Add function documentation
22. Add usage examples
23. Update Architecture.md if needed
24. Commit: `docs(planner): document graph algorithms`

## Acceptance Criteria

Phase 8 is complete when:

- [ ] BuildGraph correctly constructs dependency graphs
- [ ] TopologicalSort produces valid ordering for acyclic graphs
- [ ] TopologicalSort returns error for cyclic graphs
- [ ] FindCycle correctly identifies cycles
- [ ] ParallelizationPlan produces safe concurrent batches
- [ ] All unit tests pass with 100% coverage
- [ ] Property-based tests verify invariants
- [ ] Integration tests verify end-to-end workflows
- [ ] Benchmarks show acceptable performance
- [ ] All linters pass without warnings
- [ ] Documentation is complete and accurate
- [ ] Code follows project conventions
- [ ] Atomic commits with conventional commit messages

## Verification Commands

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./internal/planner/...

# Verify 100% coverage for this phase
go test -coverprofile=coverage.out ./internal/planner/
go tool cover -func=coverage.out | grep total

# Run property-based tests with high iterations
go test -v -run TestGraphProperties -gopter.iterations=10000

# Run benchmarks
go test -bench=. -benchmem ./internal/planner/

# Run linters
make lint

# Check cyclomatic complexity
gocyclo -over 15 internal/planner/
```

## Example Usage

```go
// Build graph from operations
ops := []dot.Operation{
    LinkCreate{ID: "link1", Dependencies: []OperationID{"dir1"}},
    DirCreate{ID: "dir1", Dependencies: nil},
    LinkCreate{ID: "link2", Dependencies: []OperationID{"dir1"}},
}

graph := planner.BuildGraph(ops)

// Check for cycles
if cycle := graph.FindCycle(); cycle != nil {
    return ErrCyclicDependency{Cycle: cycle}
}

// Get topologically sorted operations
sorted, err := graph.TopologicalSort()
if err != nil {
    return err
}
// sorted = [dir1, link1, link2] or [dir1, link2, link1]

// Get parallelization plan
batches := graph.ParallelizationPlan()
// batches = [[dir1], [link1, link2]]
// Batch 0 must complete before Batch 1
// Operations in Batch 1 can run in parallel
```

## Risks and Mitigations

### Risk: Incorrect cycle detection
**Impact**: Execution of cyclic dependencies causes infinite loops
**Mitigation**: Comprehensive testing with known cyclic graphs, property-based tests

### Risk: Invalid parallelization
**Impact**: Race conditions during execution
**Mitigation**: Thorough validation in tests, property-based invariant checking

### Risk: Performance issues with large graphs
**Impact**: Slow planning for large package sets
**Mitigation**: Benchmark-driven development, algorithmic complexity awareness

### Risk: Memory usage with large graphs
**Impact**: Excessive memory for large operations
**Mitigation**: Benchmark memory usage, optimize data structures if needed

## Future Enhancements (Post-Phase 8)

- Incremental graph updates for restow operations
- Graph visualization for debugging
- Advanced parallelization heuristics (considering I/O patterns)
- Graph simplification (removing redundant edges)
- Caching topological sort results

## References

- Architecture.md: DependencyGraph implementation (lines 933-1080)
- Implementation-Plan.md: Phase 8 overview (lines 300-323)
- [Topological Sorting](https://en.wikipedia.org/wiki/Topological_sorting)
- [Cycle Detection Algorithms](https://en.wikipedia.org/wiki/Cycle_detection)

