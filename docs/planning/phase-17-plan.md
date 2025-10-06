# Phase 17: Integration Testing - Detailed Implementation Plan

## Overview

Comprehensive end-to-end integration testing to verify complete workflows, concurrent operations, error recovery, and cross-platform compatibility. This phase ensures all components work correctly together under realistic conditions.

## Goals

- Verify complete workflows from CLI to filesystem
- Test concurrent operation safety and correctness
- Validate error recovery and rollback mechanisms
- Ensure cross-platform compatibility
- Detect performance regressions
- Establish baseline for regression testing

## Architecture Context

Integration tests verify the complete system:
- CLI commands → Client API → Pipeline → Executor → Filesystem
- Multi-package operations with dependencies
- State management and manifest persistence
- Conflict resolution with various policies
- Concurrent execution with parallelization
- Transaction safety and rollback

## Implementation Tasks

### 17.1: Test Infrastructure Setup

**Goal**: Establish foundation for integration testing with fixtures, utilities, and test harnesses.

#### 17.1.1: Test Package Structure
- [ ] Create `tests/integration/` package structure
- [ ] Create `tests/fixtures/` for test data
- [ ] Create `tests/fixtures/scenarios/` for realistic setups
- [ ] Create `tests/fixtures/packages/` for sample packages
- [ ] Add README documenting test organization
- [ ] Write package documentation

**Acceptance**: Clean test structure with documentation

#### 17.1.2: Fixture Builder Framework
- [ ] Define `FixtureBuilder` interface for test setup
- [ ] Implement `PackageBuilder` for creating test packages
- [ ] Implement `FileTreeBuilder` for constructing directory trees
- [ ] Add `SymlinkBuilder` for pre-existing symlinks
- [ ] Implement `StateBuilder` for manifest setup
- [ ] Write builder composition utilities
- [ ] Add builder tests

**Acceptance**: Builders enable easy fixture creation

#### 17.1.3: Test Utilities
- [ ] Implement `TestFS` wrapper with assertion helpers
- [ ] Add `AssertLink()` for verifying symlinks
- [ ] Add `AssertFile()` for verifying file content
- [ ] Add `AssertDir()` for verifying directories
- [ ] Add `AssertManifest()` for verifying state
- [ ] Implement `CaptureState()` for filesystem snapshots
- [ ] Add `CompareStates()` for state diffing
- [ ] Write utility tests

**Acceptance**: Utilities simplify test assertions

#### 17.1.4: Golden Test Framework
- [ ] Create `tests/fixtures/golden/` for expected outputs
- [ ] Implement `GoldenTest` harness
- [ ] Add update mode for regenerating golden files
- [ ] Implement comparison with diff output
- [ ] Add support for multiple output formats (text, JSON, YAML)
- [ ] Create golden file naming conventions
- [ ] Write framework tests

**Acceptance**: Golden tests verify output consistency

#### 17.1.5: Test Harness
- [ ] Implement `TestEnvironment` for isolated testing
- [ ] Add temporary directory management
- [ ] Implement cleanup with defer safety
- [ ] Add test timeout configuration
- [ ] Implement test context with cancellation
- [ ] Add parallel test coordination
- [ ] Write harness tests

**Acceptance**: Harness provides clean test environment

### 17.2: End-to-End Workflow Tests

**Goal**: Verify complete workflows for all core operations work correctly.

#### 17.2.1: Manage Workflow Tests
- [ ] Test single package installation
- [ ] Test multiple package installation
- [ ] Test nested directory structure installation
- [ ] Test installation with ignore patterns
- [ ] Test installation with absolute links
- [ ] Test installation with directory folding
- [ ] Test installation with --no-folding
- [ ] Test idempotent re-installation
- [ ] Write comprehensive workflow tests

**Acceptance**: All manage scenarios pass

#### 17.2.2: Unmanage Workflow Tests
- [ ] Test single package removal
- [ ] Test multiple package removal
- [ ] Test removal with empty directory cleanup
- [ ] Test removal preserves unmanaged files
- [ ] Test removal validates link ownership
- [ ] Test removal of folded directories
- [ ] Test removal with partial installation
- [ ] Write comprehensive workflow tests

**Acceptance**: All unmanage scenarios pass

#### 17.2.3: Remanage Workflow Tests
- [ ] Test remanage with unchanged packages (no-op)
- [ ] Test remanage with modified packages
- [ ] Test remanage with new files added
- [ ] Test remanage with files deleted
- [ ] Test remanage with structure changes
- [ ] Test incremental detection via hashing
- [ ] Test remanage maintains other packages
- [ ] Write comprehensive workflow tests

**Acceptance**: All remanage scenarios pass

#### 17.2.4: Adopt Workflow Tests
- [ ] Test adopting single file
- [ ] Test adopting multiple files
- [ ] Test adopting nested files
- [ ] Test adopt preserves content
- [ ] Test adopt preserves permissions
- [ ] Test adopt creates correct symlinks
- [ ] Test adopt with non-existent package
- [ ] Test adopt with backup policy
- [ ] Write comprehensive workflow tests

**Acceptance**: All adopt scenarios pass

#### 17.2.5: Combined Workflow Tests
- [ ] Test manage + unmanage identity
- [ ] Test manage + adopt + unmanage
- [ ] Test manage + remanage workflow
- [ ] Test multiple operations in sequence
- [ ] Test state consistency across operations
- [ ] Test manifest updates across operations
- [ ] Write comprehensive combined tests

**Acceptance**: Combined workflows maintain consistency

### 17.3: Concurrent Testing

**Goal**: Verify thread safety and correctness of parallel operations.

#### 17.3.1: Parallel Package Processing
- [ ] Test concurrent package scanning
- [ ] Test parallel operation execution
- [ ] Test batch parallelization correctness
- [ ] Test dependency ordering with parallelism
- [ ] Test worker pool behavior
- [ ] Test cancellation during parallel execution
- [ ] Write concurrency tests

**Acceptance**: Parallel operations are safe and correct

#### 17.3.2: Race Condition Detection
- [ ] Enable race detector for all tests
- [ ] Test concurrent access to manifest
- [ ] Test concurrent filesystem operations
- [ ] Test concurrent checkpoint updates
- [ ] Test concurrent cache access
- [ ] Identify and fix any races
- [ ] Write race-focused tests

**Acceptance**: No race conditions detected

#### 17.3.3: Concurrent Operation Tests
- [ ] Test multiple manage operations concurrently
- [ ] Test concurrent manage + unmanage
- [ ] Test concurrent status queries
- [ ] Test concurrent manifest reads
- [ ] Test lock contention scenarios
- [ ] Test operation isolation
- [ ] Write concurrent operation tests

**Acceptance**: Concurrent operations are properly isolated

#### 17.3.4: Stress Testing
- [ ] Test with high concurrency (100+ goroutines)
- [ ] Test with large package sets (1000+ packages)
- [ ] Test with deep directory trees (100+ levels)
- [ ] Test with many files per package (10000+ files)
- [ ] Test memory usage under load
- [ ] Identify performance bottlenecks
- [ ] Write stress tests

**Acceptance**: System handles stress without failure

### 17.4: Error Recovery Testing

**Goal**: Verify error handling, rollback, and recovery mechanisms work correctly.

#### 17.4.1: Transaction Rollback Tests
- [ ] Test rollback on filesystem error
- [ ] Test rollback on permission error
- [ ] Test rollback on disk full error
- [ ] Test rollback maintains original state
- [ ] Test rollback reverses partial operations
- [ ] Test rollback in correct dependency order
- [ ] Test rollback logging and reporting
- [ ] Write comprehensive rollback tests

**Acceptance**: Rollback correctly restores state

#### 17.4.2: Checkpoint Recovery Tests
- [ ] Test checkpoint creation before modifications
- [ ] Test checkpoint restoration on failure
- [ ] Test checkpoint cleanup on success
- [ ] Test checkpoint with parallel execution
- [ ] Test checkpoint with large operations
- [ ] Test recovery from interrupted operations
- [ ] Write checkpoint tests

**Acceptance**: Checkpoints enable reliable recovery

#### 17.4.3: Error Propagation Tests
- [ ] Test error collection without fail-fast
- [ ] Test multiple error aggregation
- [ ] Test error context preservation
- [ ] Test error formatting for users
- [ ] Test error types and categorization
- [ ] Test error recovery suggestions
- [ ] Write error propagation tests

**Acceptance**: Errors are properly collected and reported

#### 17.4.4: Partial Failure Tests
- [ ] Test failure in middle of operation sequence
- [ ] Test failure in parallel batch
- [ ] Test failure during rollback
- [ ] Test recovery from partial manifest update
- [ ] Test state consistency after partial failure
- [ ] Test user notification of partial results
- [ ] Write partial failure tests

**Acceptance**: Partial failures are handled gracefully

#### 17.4.5: Validation Tests
- [ ] Test pre-execution validation prevents errors
- [ ] Test permission validation
- [ ] Test disk space validation
- [ ] Test path validation
- [ ] Test dependency cycle detection
- [ ] Test validation error reporting
- [ ] Write validation tests

**Acceptance**: Validation catches errors early

### 17.5: Conflict Resolution Scenarios

**Goal**: Verify conflict detection and resolution with all policies.

#### 17.5.1: Conflict Detection Tests
- [ ] Test file exists conflict
- [ ] Test wrong link target conflict
- [ ] Test directory vs file conflict
- [ ] Test permission conflict
- [ ] Test circular dependency conflict
- [ ] Test multi-package overlap conflict
- [ ] Write conflict detection tests

**Acceptance**: All conflict types are detected

#### 17.5.2: Resolution Policy Tests
- [ ] Test fail policy stops and reports
- [ ] Test backup policy preserves files
- [ ] Test overwrite policy replaces files
- [ ] Test skip policy continues with warnings
- [ ] Test per-conflict policy application
- [ ] Test policy configuration via flags
- [ ] Write policy tests

**Acceptance**: All resolution policies work correctly

#### 17.5.3: Conflict Resolution Integration
- [ ] Test manage with conflicts and fail policy
- [ ] Test manage with conflicts and backup policy
- [ ] Test manage with conflicts and overwrite policy
- [ ] Test conflict reporting format
- [ ] Test conflict suggestions
- [ ] Test conflict resolution workflow
- [ ] Write integration tests

**Acceptance**: Conflict resolution integrates properly

### 17.6: State Management Integration

**Goal**: Verify manifest tracking and incremental operations.

#### 17.6.1: Manifest Persistence Tests
- [ ] Test manifest creation on first manage
- [ ] Test manifest updates on operations
- [ ] Test manifest preservation on errors
- [ ] Test manifest format and versioning
- [ ] Test manifest with multiple packages
- [ ] Test manifest with complex structures
- [ ] Write persistence tests

**Acceptance**: Manifest is reliably persisted

#### 17.6.2: Incremental Detection Tests
- [ ] Test unchanged package detection
- [ ] Test modified package detection
- [ ] Test added file detection
- [ ] Test deleted file detection
- [ ] Test hash computation correctness
- [ ] Test incremental remanage performance
- [ ] Write incremental tests

**Acceptance**: Incremental detection is accurate

#### 17.6.3: State Validation Tests
- [ ] Test manifest consistency validation
- [ ] Test drift detection
- [ ] Test orphaned link detection
- [ ] Test manifest repair from filesystem
- [ ] Test validation error reporting
- [ ] Write validation tests

**Acceptance**: State validation catches inconsistencies

### 17.7: Query Command Integration

**Goal**: Verify status, doctor, and list commands work correctly.

#### 17.7.1: Status Command Tests
- [ ] Test status with no packages
- [ ] Test status with installed packages
- [ ] Test status with specific packages
- [ ] Test status with conflicts
- [ ] Test status output formats (text, JSON, YAML, table)
- [ ] Test status performance
- [ ] Write status tests

**Acceptance**: Status command reports accurately

#### 17.7.2: Doctor Command Tests
- [ ] Test doctor detects broken links
- [ ] Test doctor detects orphaned links
- [ ] Test doctor detects permission issues
- [ ] Test doctor detects manifest issues
- [ ] Test doctor suggestions
- [ ] Test doctor exit codes
- [ ] Write doctor tests

**Acceptance**: Doctor command diagnoses issues

#### 17.7.3: List Command Tests
- [ ] Test list with no packages
- [ ] Test list with installed packages
- [ ] Test list sorting by various fields
- [ ] Test list output formats
- [ ] Test list performance
- [ ] Write list tests

**Acceptance**: List command displays packages correctly

### 17.8: Cross-Platform Testing

**Goal**: Ensure functionality works across operating systems and filesystems.

#### 17.8.1: Platform-Specific Tests
- [ ] Create Linux-specific test suite
- [ ] Create macOS-specific test suite
- [ ] Create Windows-specific test suite (with limitations)
- [ ] Create BSD-specific test suite
- [ ] Test symlink behavior per platform
- [ ] Test path handling per platform
- [ ] Write platform tests

**Acceptance**: Tests pass on all supported platforms

#### 17.8.2: Filesystem Compatibility Tests
- [ ] Test on ext4 filesystem
- [ ] Test on btrfs filesystem
- [ ] Test on xfs filesystem
- [ ] Test on apfs filesystem (macOS)
- [ ] Test on zfs filesystem
- [ ] Test symlink limits per filesystem
- [ ] Document filesystem-specific issues
- [ ] Write filesystem tests

**Acceptance**: Functionality verified across filesystems

#### 17.8.3: Path Convention Tests
- [ ] Test with Unix path separators
- [ ] Test with Windows path separators
- [ ] Test path absolutization per platform
- [ ] Test path normalization
- [ ] Test case sensitivity differences
- [ ] Write path convention tests

**Acceptance**: Path handling works cross-platform

### 17.9: Performance Regression Testing

**Goal**: Establish baselines and detect performance degradation.

#### 17.9.1: Benchmark Suite
- [ ] Create benchmark for single package install
- [ ] Create benchmark for 100 package install
- [ ] Create benchmark for large file tree (10k files)
- [ ] Create benchmark for deep nesting (100 levels)
- [ ] Create benchmark for remanage unchanged
- [ ] Create benchmark for status query
- [ ] Write benchmark tests

**Acceptance**: Benchmarks establish baselines

#### 17.9.2: Performance Monitoring
- [ ] Implement benchmark result tracking
- [ ] Add regression detection thresholds (10% degradation)
- [ ] Create performance comparison reports
- [ ] Integrate benchmarks into CI
- [ ] Document performance characteristics
- [ ] Write monitoring utilities

**Acceptance**: Performance regressions are detected

#### 17.9.3: Memory Profiling
- [ ] Profile memory usage during operations
- [ ] Test memory usage with large operations
- [ ] Verify streaming prevents unbounded growth
- [ ] Test memory leaks with repeated operations
- [ ] Document memory characteristics
- [ ] Write memory tests

**Acceptance**: Memory usage is reasonable and bounded

### 17.10: CLI Integration Testing

**Goal**: Verify CLI commands work end-to-end with all options.

#### 17.10.1: Command Invocation Tests
- [ ] Test manage command with all flag combinations
- [ ] Test unmanage command with all flag combinations
- [ ] Test remanage command with all flag combinations
- [ ] Test adopt command with all flag combinations
- [ ] Test status command with all flag combinations
- [ ] Test doctor command invocation
- [ ] Test list command with all flag combinations
- [ ] Write invocation tests

**Acceptance**: All command combinations work

#### 17.10.2: Flag Interaction Tests
- [ ] Test --dry-run prevents modifications
- [ ] Test --verbose output levels
- [ ] Test --quiet suppresses output
- [ ] Test --log-json format
- [ ] Test --no-folding behavior
- [ ] Test --absolute link mode
- [ ] Test --ignore patterns
- [ ] Test global + command flags
- [ ] Write flag interaction tests

**Acceptance**: Flags interact correctly

#### 17.10.3: Exit Code Tests
- [ ] Test exit code 0 on success
- [ ] Test exit code 1 on general error
- [ ] Test exit code 2 on invalid arguments
- [ ] Test exit code 3 on conflicts
- [ ] Test exit code 4 on permission denied
- [ ] Test exit code 5 on package not found
- [ ] Write exit code tests

**Acceptance**: Exit codes match specification

#### 17.10.4: Output Format Tests
- [ ] Test text output format
- [ ] Test JSON output format
- [ ] Test YAML output format
- [ ] Test table output format
- [ ] Test colorized vs plain output
- [ ] Test output consistency across formats
- [ ] Use golden tests for output verification
- [ ] Write output tests

**Acceptance**: All output formats work correctly

### 17.11: Scenario-Based Testing

**Goal**: Test realistic user scenarios and workflows.

#### 17.11.1: New User Scenario
- [ ] Create fresh installation scenario
- [ ] Test first-time manage of dotfiles
- [ ] Test discovery and status checking
- [ ] Test adoption of existing files
- [ ] Test unmanage and cleanup
- [ ] Write new user scenario tests

**Acceptance**: New user workflow is smooth

#### 17.11.2: Migration Scenario
- [ ] Create GNU Stow migration scenario
- [ ] Test migrating existing Stow setup
- [ ] Test compatibility with Stow structure
- [ ] Test side-by-side operation
- [ ] Write migration scenario tests

**Acceptance**: Migration from Stow works

#### 17.11.3: Multi-Machine Scenario
- [ ] Create multi-machine sync scenario
- [ ] Test installing same packages on different machines
- [ ] Test machine-specific overrides
- [ ] Test portable vs absolute links
- [ ] Write multi-machine scenario tests

**Acceptance**: Multi-machine use case works

#### 17.11.4: Development Workflow Scenario
- [ ] Create development workflow scenario
- [ ] Test manage, modify, remanage cycle
- [ ] Test incremental updates
- [ ] Test rapid iteration
- [ ] Write development scenario tests

**Acceptance**: Development workflow is efficient

#### 17.11.5: Large Repository Scenario
- [ ] Create large repository scenario (100+ packages)
- [ ] Test performance with many packages
- [ ] Test selective installation
- [ ] Test incremental updates
- [ ] Write large repository tests

**Acceptance**: Large repositories perform well

### 17.12: Test Organization and Maintenance

**Goal**: Organize tests for maintainability and discoverability.

#### 17.12.1: Test Documentation
- [ ] Write integration testing guide
- [ ] Document test organization
- [ ] Document fixture creation
- [ ] Document test utilities
- [ ] Create test naming conventions
- [ ] Add examples for common patterns
- [ ] Write maintenance guide

**Acceptance**: Tests are well-documented

#### 17.12.2: Test Categorization
- [ ] Tag tests by category (e2e, concurrent, recovery, etc.)
- [ ] Implement test filtering by tag
- [ ] Create test suites for CI stages
- [ ] Add quick smoke test suite
- [ ] Add comprehensive full test suite
- [ ] Document test categories

**Acceptance**: Tests are organized and filterable

#### 17.12.3: CI Integration
- [ ] Add integration tests to CI pipeline
- [ ] Configure test parallelization in CI
- [ ] Add test result reporting
- [ ] Add coverage reporting for integration tests
- [ ] Set up test artifact collection
- [ ] Configure test timeouts
- [ ] Write CI documentation

**Acceptance**: Integration tests run in CI

#### 17.12.4: Test Maintenance
- [ ] Implement test flake detection
- [ ] Add retry logic for flaky tests
- [ ] Create test cleanup verification
- [ ] Add test isolation verification
- [ ] Implement test dependency checking
- [ ] Write maintenance utilities

**Acceptance**: Tests are maintainable and reliable

## Testing Strategy

### Test Pyramid

Integration tests sit at the middle layer:
- **Unit Tests (base)**: Test individual functions and types
- **Integration Tests (middle)**: Test component interaction and workflows
- **Property Tests (top)**: Verify algebraic laws and invariants

### Test Coverage Goals

- All core workflows covered by end-to-end tests
- All error paths tested with recovery verification
- All command combinations tested
- Cross-platform compatibility verified
- Performance baselines established

### Test Execution

#### Local Development
```bash
# Run all integration tests
make test-integration

# Run specific category
go test ./tests/integration/... -tags=e2e

# Run with race detector
go test -race ./tests/integration/...

# Run with verbose output
go test -v ./tests/integration/...
```

#### Continuous Integration
- Run on every pull request
- Run on multiple platforms (Linux, macOS, Windows)
- Run with race detector enabled
- Collect and report coverage
- Detect and report flaky tests

### Test Fixtures

Organized by scenario type:
```
tests/
├── fixtures/
│   ├── scenarios/
│   │   ├── simple/          # Basic single package
│   │   ├── complex/         # Multiple packages with dependencies
│   │   ├── conflicts/       # Pre-existing conflicts
│   │   └── migration/       # GNU Stow migration
│   ├── packages/
│   │   ├── dotfiles/        # Sample dotfiles package
│   │   ├── nvim/            # Sample neovim config
│   │   └── shell/           # Sample shell config
│   └── golden/
│       ├── status/          # Expected status outputs
│       ├── doctor/          # Expected doctor outputs
│       └── list/            # Expected list outputs
└── integration/
    ├── e2e_test.go          # End-to-end tests
    ├── concurrent_test.go   # Concurrency tests
    ├── recovery_test.go     # Error recovery tests
    └── cli_test.go          # CLI integration tests
```

## Dependencies

### Test Frameworks
- Standard library `testing` package
- `testify/assert` for assertions
- `testify/require` for critical assertions
- `testify/suite` for test organization

### Test Utilities
- Memory filesystem from adapters
- Fixture builders (custom)
- Golden test framework (custom)
- State comparison utilities (custom)

## Success Criteria

Phase 17 is complete when:
- [ ] All integration test categories implemented
- [ ] Test coverage for all core workflows ≥ 95%
- [ ] All tests pass on Linux, macOS, Windows
- [ ] No race conditions detected
- [ ] Performance benchmarks established
- [ ] Golden tests verify output consistency
- [ ] Error recovery verified for all failure modes
- [ ] Concurrent execution safety verified
- [ ] Test documentation complete
- [ ] CI integration working

## Deliverables

1. **Comprehensive Integration Test Suite**
   - End-to-end workflow tests for all operations
   - Concurrent execution tests
   - Error recovery tests
   - Cross-platform tests

2. **Test Infrastructure**
   - Fixture builders and utilities
   - Golden test framework
   - Test harness and environment

3. **Performance Baselines**
   - Benchmark suite
   - Regression detection
   - Memory profiling

4. **Documentation**
   - Integration testing guide
   - Test organization documentation
   - Scenario documentation

5. **CI Integration**
   - Automated test execution
   - Multi-platform testing
   - Test reporting

## Timeline Estimate

- **17.1-17.2**: 8-10 hours (infrastructure and core workflows)
- **17.3-17.4**: 6-8 hours (concurrency and recovery)
- **17.5-17.6**: 4-6 hours (conflicts and state)
- **17.7-17.8**: 4-6 hours (queries and cross-platform)
- **17.9-17.10**: 4-6 hours (performance and CLI)
- **17.11-17.12**: 4-6 hours (scenarios and organization)

**Total**: 30-42 hours

## Risks and Mitigations

### Risks
1. **Test Flakiness**: Timing-dependent tests may fail intermittently
2. **Platform Differences**: Behavior may vary across OS/filesystem
3. **Test Maintenance**: Large test suite requires ongoing maintenance
4. **CI Performance**: Long test execution in CI

### Mitigations
1. Use deterministic test fixtures, retry flaky tests, add timeouts
2. Document platform-specific behavior, skip unsupported features
3. Good organization, clear naming, comprehensive documentation
4. Parallelize tests, cache dependencies, optimize slow tests

## Post-Phase 17

After completing integration testing:
- **Phase 18**: Performance optimization based on benchmarks
- **Phase 19**: Documentation refinement
- **Phase 20**: Release preparation and polish

Integration tests provide confidence for release and serve as regression suite for ongoing development.

