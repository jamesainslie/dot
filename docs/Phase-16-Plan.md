# Phase 16: Property-Based Testing

## Overview

Implement comprehensive property-based testing using gopter to verify algebraic laws, invariants, and mathematical properties of the dot system. This phase ensures correctness through automated generation and testing of thousands of input combinations, catching edge cases that unit tests might miss.

## Objectives

- Verify core algebraic laws (idempotence, commutativity, reversibility)
- Validate domain invariants across all operations
- Generate realistic test data automatically
- Detect edge cases and boundary conditions
- Integrate property tests into CI pipeline
- Document mathematical properties of the system

## Prerequisites

- Phase 13 complete (CLI with core commands)
- All core domain types implemented
- Working Client API
- Comprehensive unit test coverage
- Memory filesystem adapter for fast testing

## Architecture Context

Property-based testing sits at the highest level of the testing pyramid. It verifies that the system satisfies mathematical laws and invariants across a wide input space, providing confidence that unit and integration tests haven't missed critical edge cases.

```
Property Tests (verify laws & invariants)
           ↓
    Integration Tests
           ↓
       Unit Tests
```

---

## Phase 16.1: Test Infrastructure Setup

### 16.1.1: Gopter Framework Integration

**Goal**: Set up gopter framework with project conventions

Tasks:
- [ ] Add gopter dependency to go.mod
- [ ] Create test/properties/ package structure
- [ ] Add properties_test.go with basic test harness
- [ ] Configure gopter parameters (iteration count, seed)
- [ ] Set up test reporter with verbose output
- [ ] Add property test runner helper
- [ ] Create Makefile target for property tests (`make test-properties`)
- [ ] Configure CI to run property tests separately
- [ ] Add timeout configuration for long-running tests
- [ ] Write documentation for adding new property tests

**Validation**:
- Property test suite runs successfully
- CI pipeline includes property test stage
- Tests complete within reasonable time (under 5 minutes)
- Failed property tests show shrunk counterexamples

**Files**:
```
test/properties/
├── properties_test.go       # Test harness
├── generators.go            # Data generators
├── laws.go                  # Algebraic law tests
├── invariants.go            # Invariant tests
└── helpers.go               # Test utilities
```

### 16.1.2: Test Configuration

**Goal**: Configure gopter behavior for project needs

Tasks:
- [ ] Set default iteration count (1000 for CI, 10000 for thorough)
- [ ] Configure shrinking strategy for failure minimization
- [ ] Set up reproducible seeding for deterministic tests
- [ ] Add test timeout configuration
- [ ] Configure parallel test execution where safe
- [ ] Add verbose mode flag for debugging
- [ ] Set up test result caching
- [ ] Configure example generation count
- [ ] Add performance budget assertions
- [ ] Document configuration options

**Validation**:
- Tests are reproducible with same seed
- Shrinking produces minimal counterexamples
- Performance stays within budget
- Configuration is documented

---

## Phase 16.2: Data Generators

### 16.2.1: Path Generators

**Goal**: Generate valid and invalid paths for testing

Tasks:
- [ ] Implement genAbsolutePath() for absolute paths
- [ ] Implement genRelativePath() for relative paths
- [ ] Implement genStowPath() constrained to stow directory
- [ ] Implement genTargetPath() constrained to target directory
- [ ] Implement genPackagePath() for package paths
- [ ] Add genInvalidPath() for malformed paths
- [ ] Add genPathWithTraversal() for security testing
- [ ] Add genSymlinkPath() for existing symlinks
- [ ] Write generator validation tests
- [ ] Document generator contracts

**Example**:
```go
func genAbsolutePath() gopter.Gen {
    return gen.SliceOf(gen.Identifier()).
        Map(func(parts []string) string {
            return "/" + filepath.Join(parts...)
        })
}
```

**Validation**:
- Generators produce valid path formats
- Generated paths satisfy type constraints
- Invalid path generator catches edge cases
- Generators cover diverse path structures

### 16.2.2: Package Generators

**Goal**: Generate realistic package structures

Tasks:
- [ ] Implement genPackageName() for valid package names
- [ ] Implement genPackageList() for multiple packages
- [ ] Implement genPackageTree() for directory structures
- [ ] Implement genFileContent() for file data
- [ ] Add genPackageMetadata() for package configuration
- [ ] Add genNestedPackage() for deep hierarchies
- [ ] Add genEmptyPackage() for edge case
- [ ] Add genLargePackage() for performance testing
- [ ] Add genPackageWithIgnores() for ignore patterns
- [ ] Write generator tests

**Example**:
```go
func genPackageTree() gopter.Gen {
    return gopter.CombineGens(
        gen.Identifier(),
        gen.SliceOf(genFileNode()),
    ).Map(func(vals []interface{}) Package {
        return buildPackage(vals[0].(string), vals[1].([]FileNode))
    })
}
```

**Validation**:
- Packages have realistic structures
- Generated packages are installable
- Edge cases (empty, large) are handled
- Metadata generation is valid

### 16.2.3: Operation Generators

**Goal**: Generate valid operations for testing

Tasks:
- [ ] Implement genLinkCreate() for link creation ops
- [ ] Implement genLinkDelete() for link deletion ops
- [ ] Implement genDirCreate() for directory creation ops
- [ ] Implement genDirDelete() for directory deletion ops
- [ ] Implement genFileMove() for file move ops
- [ ] Implement genOperationList() for operation sequences
- [ ] Add genDependentOps() with valid dependencies
- [ ] Add genConflictingOps() for conflict scenarios
- [ ] Write operation generator tests
- [ ] Document operation generation strategy

**Validation**:
- Operations are individually valid
- Dependencies are correctly generated
- Conflict scenarios are realistic
- Operation sequences are executable

### 16.2.4: Filesystem State Generators

**Goal**: Generate valid and interesting filesystem states

Tasks:
- [ ] Implement genFilesystemState() for complete states
- [ ] Implement genWithSymlinks() for existing links
- [ ] Implement genWithFiles() for existing files
- [ ] Implement genWithDirectories() for directory trees
- [ ] Add genConflictState() for conflict scenarios
- [ ] Add genPartialInstall() for incomplete states
- [ ] Add genBrokenState() for error conditions
- [ ] Add genLargeState() for performance testing
- [ ] Write state generator tests
- [ ] Document state generation patterns

**Validation**:
- Generated states are internally consistent
- States cover diverse scenarios
- Conflict states are realistic
- Large states test performance boundaries

### 16.2.5: Generator Composition

**Goal**: Compose generators for complex scenarios

Tasks:
- [ ] Implement genScenario() combining multiple generators
- [ ] Add genWorkflow() for operation sequences
- [ ] Add genWithContext() for contextual generation
- [ ] Implement generator filtering for constraints
- [ ] Add generator weighting for bias control
- [ ] Add generator debugging utilities
- [ ] Write composition tests
- [ ] Document composition patterns
- [ ] Create generator cookbook with examples
- [ ] Add generator performance benchmarks

**Validation**:
- Composed generators produce valid inputs
- Scenarios are realistic and diverse
- Debugging utilities aid troubleshooting
- Documentation enables reuse

---

## Phase 16.3: Algebraic Law Verification

### 16.3.1: Idempotence Properties

**Goal**: Verify operations can be safely repeated

Tasks:
- [ ] Test manage idempotence (manage twice = manage once)
- [ ] Test unmanage idempotence
- [ ] Test remanage idempotence
- [ ] Test adopt idempotence
- [ ] Test status query idempotence
- [ ] Add edge case testing (empty packages, large packages)
- [ ] Test idempotence under failures
- [ ] Test idempotence with conflicts
- [ ] Document idempotence guarantees
- [ ] Add idempotence examples to docs

**Example**:
```go
func TestManageIdempotent(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("manage is idempotent", prop.ForAll(
        func(packages []string) bool {
            ctx := context.Background()
            fs := setupTestFS()
            client := newTestClient(fs)
            
            // First manage
            if err := client.Manage(ctx, packages...); err != nil {
                return false
            }
            state1 := captureState(fs)
            
            // Second manage
            if err := client.Manage(ctx, packages...); err != nil {
                return false
            }
            state2 := captureState(fs)
            
            return statesEqual(state1, state2)
        },
        genPackageList(),
    ))
    
    properties.TestingRun(t)
}
```

**Validation**:
- All operations pass idempotence tests
- Edge cases don't break idempotence
- Failures preserve idempotence
- Documentation is accurate

### 16.3.2: Reversibility Properties

**Goal**: Verify operations can be cleanly undone

Tasks:
- [ ] Test manage-unmanage identity (manage + unmanage = noop)
- [ ] Test adopt reversibility (adopt + restore = original)
- [ ] Test operation rollback completeness
- [ ] Test partial reversal scenarios
- [ ] Add edge case testing (conflicts, errors)
- [ ] Test reversibility with permissions
- [ ] Test reversibility with symlinks
- [ ] Document reversibility guarantees
- [ ] Add reversibility examples
- [ ] Test checkpoint-restore identity

**Example**:
```go
func TestManageUnmanageIdentity(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("manage then unmanage is identity", prop.ForAll(
        func(packages []string) bool {
            ctx := context.Background()
            fs := setupTestFS()
            client := newTestClient(fs)
            
            initial := captureState(fs)
            
            if err := client.Manage(ctx, packages...); err != nil {
                return false
            }
            
            if err := client.Unmanage(ctx, packages...); err != nil {
                return false
            }
            
            final := captureState(fs)
            
            return statesEqual(initial, final)
        },
        genPackageList(),
    ))
    
    properties.TestingRun(t)
}
```

**Validation**:
- Reversibility holds for all operations
- Edge cases preserve reversibility
- Partial operations can be reversed
- Documentation is comprehensive

### 16.3.3: Commutativity Properties

**Goal**: Verify operation order independence where expected

Tasks:
- [ ] Test package installation order independence
- [ ] Test unmanage order independence
- [ ] Test parallel operation equivalence
- [ ] Test batch operation ordering
- [ ] Add conflict scenario testing
- [ ] Test with shared directories
- [ ] Test with dependencies
- [ ] Document commutativity guarantees
- [ ] Add commutativity examples
- [ ] Test operation batching equivalence

**Example**:
```go
func TestManageCommutative(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("package order doesn't matter", prop.ForAll(
        func(packages []string) bool {
            if len(packages) < 2 {
                return true
            }
            
            ctx := context.Background()
            
            // Forward order
            fs1 := setupTestFS()
            client1 := newTestClient(fs1)
            if err := client1.Manage(ctx, packages...); err != nil {
                return false
            }
            state1 := captureState(fs1)
            
            // Reverse order
            reversed := reverse(packages)
            fs2 := setupTestFS()
            client2 := newTestClient(fs2)
            if err := client2.Manage(ctx, reversed...); err != nil {
                return false
            }
            state2 := captureState(fs2)
            
            return statesEqual(state1, state2)
        },
        genPackageList(),
    ))
    
    properties.TestingRun(t)
}
```

**Validation**:
- Commutativity holds where documented
- Non-commutative operations are identified
- Dependencies don't break commutativity
- Documentation is clear

### 16.3.4: Associativity Properties

**Goal**: Verify operation grouping independence

Tasks:
- [ ] Test batch operation associativity
- [ ] Test package group installation
- [ ] Test operation composition
- [ ] Test pipeline stage grouping
- [ ] Add complex grouping scenarios
- [ ] Test with errors and conflicts
- [ ] Document associativity guarantees
- [ ] Add associativity examples
- [ ] Test nested operation groups
- [ ] Test parallel batch equivalence

**Validation**:
- Associativity holds for batch operations
- Grouping doesn't affect outcomes
- Errors preserve associativity
- Documentation is accurate

### 16.3.5: Conservation Properties

**Goal**: Verify data and state preservation

Tasks:
- [ ] Test file content preservation during adopt
- [ ] Test permission preservation
- [ ] Test timestamp preservation
- [ ] Test symlink target preservation
- [ ] Test metadata preservation
- [ ] Add large file testing
- [ ] Test with special file types
- [ ] Document conservation guarantees
- [ ] Add conservation examples
- [ ] Test across operations

**Example**:
```go
func TestAdoptPreservesContent(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("adopt preserves file content", prop.ForAll(
        func(files []FileContent) bool {
            ctx := context.Background()
            fs := setupTestFS()
            client := newTestClient(fs)
            
            // Write files and capture content
            original := make(map[string][]byte)
            for _, f := range files {
                writeFile(fs, f.Path, f.Content)
                original[f.Path] = f.Content
            }
            
            // Adopt files
            paths := extractPaths(files)
            if err := client.Adopt(ctx, "pkg", paths...); err != nil {
                return false
            }
            
            // Verify content unchanged
            for path, content := range original {
                actual := readFile(fs, path)
                if !bytes.Equal(content, actual) {
                    return false
                }
            }
            
            return true
        },
        genFileList(),
    ))
    
    properties.TestingRun(t)
}
```

**Validation**:
- Content is preserved across operations
- Metadata is maintained
- Special cases are handled
- Documentation is complete

---

## Phase 16.4: Domain Invariant Verification

### 16.4.1: Path Invariants

**Goal**: Verify path type safety and correctness

Tasks:
- [ ] Test absolute path invariant (all paths are absolute)
- [ ] Test path normalization invariant
- [ ] Test path containment (target in target dir, etc.)
- [ ] Test no directory traversal
- [ ] Test symlink resolution correctness
- [ ] Add malformed path testing
- [ ] Test path type preservation
- [ ] Document path invariants
- [ ] Add path invariant examples
- [ ] Test cross-platform path handling

**Validation**:
- Path invariants hold universally
- Type system prevents violations
- Malformed paths are rejected
- Documentation is thorough

### 16.4.2: Graph Invariants

**Goal**: Verify dependency graph properties

Tasks:
- [ ] Test acyclicity (no circular dependencies)
- [ ] Test reachability (all nodes reachable)
- [ ] Test topological order validity
- [ ] Test dependency completeness
- [ ] Test parallel batch safety
- [ ] Add complex graph scenarios
- [ ] Test graph modifications
- [ ] Document graph invariants
- [ ] Add graph invariant examples
- [ ] Test edge cases (empty, single node)

**Validation**:
- Graphs maintain acyclicity
- Topological sorts are valid
- Parallel batches are safe
- Documentation is accurate

### 16.4.3: Manifest Invariants

**Goal**: Verify state tracking consistency

Tasks:
- [ ] Test manifest matches filesystem (consistency)
- [ ] Test manifest completeness (all links tracked)
- [ ] Test manifest accuracy (no phantom entries)
- [ ] Test hash consistency with content
- [ ] Test timestamp ordering
- [ ] Add concurrent modification testing
- [ ] Test manifest repair correctness
- [ ] Document manifest invariants
- [ ] Add manifest examples
- [ ] Test edge cases (empty, corrupted)

**Validation**:
- Manifest accurately reflects state
- Invariants hold after operations
- Repair restores consistency
- Documentation is complete

### 16.4.4: Operation Invariants

**Goal**: Verify operation execution properties

Tasks:
- [ ] Test operation validity post-creation
- [ ] Test dependency satisfaction
- [ ] Test precondition verification
- [ ] Test postcondition guarantee
- [ ] Test atomicity (all-or-nothing)
- [ ] Add failure scenario testing
- [ ] Test rollback completeness
- [ ] Document operation invariants
- [ ] Add operation examples
- [ ] Test operation composition

**Validation**:
- Operations maintain validity
- Dependencies are satisfied
- Atomicity is preserved
- Documentation is thorough

### 16.4.5: Conflict Invariants

**Goal**: Verify conflict detection completeness

Tasks:
- [ ] Test all conflicts detected
- [ ] Test no false positives
- [ ] Test conflict resolution correctness
- [ ] Test policy application consistency
- [ ] Test suggestion validity
- [ ] Add complex conflict scenarios
- [ ] Test multi-package conflicts
- [ ] Document conflict invariants
- [ ] Add conflict examples
- [ ] Test edge cases

**Validation**:
- All conflicts are detected
- False positives are eliminated
- Resolutions are valid
- Documentation is complete

---

## Phase 16.5: Performance Properties

### 16.5.1: Complexity Bounds

**Goal**: Verify algorithmic complexity guarantees

Tasks:
- [ ] Test scanner scales linearly with file count
- [ ] Test planner scales with package count
- [ ] Test topological sort is O(V+E)
- [ ] Test executor parallelization effectiveness
- [ ] Test incremental planner performance
- [ ] Add large input testing (10K+ files)
- [ ] Test memory usage bounds
- [ ] Document complexity guarantees
- [ ] Add complexity examples
- [ ] Test worst-case scenarios

**Validation**:
- Complexity bounds are maintained
- Performance degrades gracefully
- Memory usage is bounded
- Documentation is accurate

### 16.5.2: Incremental Operation Properties

**Goal**: Verify incremental operations are faster

Tasks:
- [ ] Test remanage faster than manage+unmanage
- [ ] Test unchanged packages are skipped
- [ ] Test hash computation is cached
- [ ] Test manifest lookup is O(1)
- [ ] Test incremental vs full comparison
- [ ] Add benchmark property tests
- [ ] Test cache effectiveness
- [ ] Document performance properties
- [ ] Add performance examples
- [ ] Test scaling behavior

**Validation**:
- Incremental operations are faster
- Caching provides benefit
- Scaling is predictable
- Documentation is complete

### 16.5.3: Parallelization Properties

**Goal**: Verify parallel execution correctness and performance

Tasks:
- [ ] Test parallel execution produces same result as sequential
- [ ] Test speedup with increasing parallelism
- [ ] Test no race conditions under load
- [ ] Test batch execution correctness
- [ ] Test resource utilization
- [ ] Add stress testing
- [ ] Test deadlock freedom
- [ ] Document parallelization guarantees
- [ ] Add parallelization examples
- [ ] Test edge cases (single operation, many operations)

**Validation**:
- Parallelization maintains correctness
- Performance improves with cores
- No races or deadlocks occur
- Documentation is thorough

---

## Phase 16.6: Error Handling Properties

### 16.6.1: Error Propagation

**Goal**: Verify errors are correctly propagated

Tasks:
- [ ] Test all errors are captured
- [ ] Test error context is preserved
- [ ] Test error aggregation correctness
- [ ] Test no errors are silently dropped
- [ ] Test error wrapping maintains chain
- [ ] Add error injection testing
- [ ] Test error recovery paths
- [ ] Document error properties
- [ ] Add error handling examples
- [ ] Test edge cases

**Validation**:
- All errors are captured and reported
- Context is maintained
- No silent failures occur
- Documentation is complete

### 16.6.2: Rollback Properties

**Goal**: Verify rollback completeness and correctness

Tasks:
- [ ] Test rollback restores original state
- [ ] Test partial rollback correctness
- [ ] Test rollback idempotence
- [ ] Test rollback error handling
- [ ] Test checkpoint accuracy
- [ ] Add complex rollback scenarios
- [ ] Test cascade rollback
- [ ] Document rollback properties
- [ ] Add rollback examples
- [ ] Test failure during rollback

**Validation**:
- Rollback fully restores state
- Errors during rollback are handled
- Properties hold under failure
- Documentation is thorough

### 16.6.3: Validation Properties

**Goal**: Verify validation catches all errors

Tasks:
- [ ] Test invalid inputs are rejected
- [ ] Test validation is exhaustive
- [ ] Test validation precedes execution
- [ ] Test validation error messages are clear
- [ ] Test no false negatives in validation
- [ ] Add edge case validation testing
- [ ] Test validation performance
- [ ] Document validation properties
- [ ] Add validation examples
- [ ] Test validation composition

**Validation**:
- Invalid inputs are always rejected
- Validation is complete
- Performance is acceptable
- Documentation is complete

---

## Phase 16.7: Integration and Documentation

### 16.7.1: CI/CD Integration

**Goal**: Integrate property tests into automation

Tasks:
- [ ] Add property test stage to CI pipeline
- [ ] Configure appropriate iteration counts for CI
- [ ] Set up failure reproduction instructions
- [ ] Add property test coverage reporting
- [ ] Configure test result archiving
- [ ] Add performance regression detection
- [ ] Set up scheduled thorough test runs
- [ ] Document CI configuration
- [ ] Add troubleshooting guide
- [ ] Configure notifications for failures

**Validation**:
- Property tests run on every commit
- Failures are easy to reproduce
- Performance regressions are detected
- Documentation is clear

### 16.7.2: Test Documentation

**Goal**: Document property test suite comprehensively

Tasks:
- [ ] Write property testing guide
- [ ] Document all verified properties
- [ ] Add generator usage examples
- [ ] Create troubleshooting guide for failures
- [ ] Document how to add new properties
- [ ] Add mathematical foundations explanation
- [ ] Create property catalog
- [ ] Document testing strategy
- [ ] Add FAQ section
- [ ] Create video walkthrough (optional)

**Files**:
```
docs/
├── Property-Testing-Guide.md
├── Verified-Properties.md
├── Generator-Cookbook.md
└── Property-Test-Troubleshooting.md
```

**Validation**:
- Documentation is comprehensive
- Examples are runnable
- Troubleshooting guide is helpful
- Mathematical basis is explained

### 16.7.3: Example Properties

**Goal**: Provide reference implementations

Tasks:
- [ ] Create example property test suite
- [ ] Add commented walkthroughs
- [ ] Provide generator examples
- [ ] Show failure debugging process
- [ ] Add performance property examples
- [ ] Create video demonstrations (optional)
- [ ] Add to documentation site
- [ ] Include in repository examples/
- [ ] Add to test suite as documentation tests
- [ ] Create interactive tutorial (optional)

**Validation**:
- Examples are clear and runnable
- Walkthroughs are instructive
- Examples cover common patterns
- Tutorial is helpful

### 16.7.4: Test Coverage Analysis

**Goal**: Measure property test effectiveness

Tasks:
- [ ] Measure input space coverage
- [ ] Analyze edge case detection
- [ ] Compare to unit test coverage
- [ ] Identify coverage gaps
- [ ] Add missing property tests
- [ ] Document coverage strategy
- [ ] Set coverage targets
- [ ] Add coverage reporting
- [ ] Track coverage over time
- [ ] Integrate into CI

**Validation**:
- Coverage metrics are meaningful
- Gaps are identified and filled
- Coverage trends upward
- Documentation explains metrics

### 16.7.5: Maintenance Guide

**Goal**: Enable ongoing property test maintenance

Tasks:
- [ ] Document when to add property tests
- [ ] Create decision tree for test type selection
- [ ] Document generator maintenance
- [ ] Add property evolution guidance
- [ ] Create review checklist
- [ ] Document common pitfalls
- [ ] Add debugging strategies
- [ ] Create maintenance schedule
- [ ] Document tooling
- [ ] Add team training materials

**Validation**:
- Maintenance process is clear
- Team can maintain tests
- Pitfalls are avoided
- Documentation is actionable

---

## Success Criteria

### Phase Completion

Phase 16 is complete when:
- [ ] All algebraic laws are verified with property tests
- [ ] All domain invariants have property test coverage
- [ ] Generators cover all major domain types
- [ ] Property tests run in CI pipeline
- [ ] Documentation is comprehensive
- [ ] Examples demonstrate usage
- [ ] Team is trained on property testing
- [ ] Test suite finds and shrinks failures effectively
- [ ] Property test execution time is reasonable
- [ ] Coverage gaps are identified and documented

### Quality Gates

- [ ] Property tests catch edge cases missed by unit tests
- [ ] Shrinking produces minimal counterexamples
- [ ] All tests pass with 10,000 iterations
- [ ] No flaky property tests
- [ ] Documentation enables team autonomy
- [ ] CI integration is reliable
- [ ] Performance properties verify scaling behavior
- [ ] Error handling properties verify reliability

---

## Testing Strategy

### Property Test Focus

Property tests verify:
- **Laws**: Mathematical properties (idempotence, commutativity, etc.)
- **Invariants**: Conditions that always hold
- **Bounds**: Performance and resource constraints
- **Relationships**: Cross-component consistency

### Unit Test Complementarity

Property tests complement but don't replace:
- Specific edge case unit tests
- Example-based tests for known issues
- Regression tests for bugs
- Integration tests for workflows

### Test Pyramid Position

```
Property Tests: Verify mathematical properties across input space
        ↑
Integration Tests: Verify component interaction
        ↑
Unit Tests: Verify individual functions
```

---

## Risk Mitigation

### Technical Risks

1. **Slow Property Tests**
   - Mitigation: Tune iteration counts, parallelize where safe
   - Fallback: Run thorough tests nightly, quick tests on PR

2. **Flaky Tests**
   - Mitigation: Use fixed seeds, avoid timing dependencies
   - Fallback: Investigate and fix flakiness immediately

3. **Complex Generators**
   - Mitigation: Test generators themselves, start simple
   - Fallback: Use simpler generators initially

4. **Shrinking Ineffective**
   - Mitigation: Tune shrinking strategy, add custom shrinkers
   - Fallback: Manual analysis of failures

### Process Risks

1. **Learning Curve**
   - Mitigation: Provide training, pair programming
   - Fallback: Focus on critical properties first

2. **Maintenance Burden**
   - Mitigation: Document well, review regularly
   - Fallback: Reduce property test scope to critical paths

---

## Deliverable

Complete property-based test suite verifying:
- Core algebraic laws (idempotence, reversibility, commutativity)
- Domain invariants (paths, graphs, manifests, operations)
- Performance properties (complexity bounds, scaling)
- Error handling properties (propagation, rollback, validation)

With:
- Comprehensive generators for all domain types
- CI integration with appropriate iteration counts
- Complete documentation and examples
- Team training and maintenance guide

**Estimated Effort**: 40-50 hours

**Dependencies**: Phase 13 (CLI), all domain types implemented

**Milestone**: Mathematical correctness verification complete

