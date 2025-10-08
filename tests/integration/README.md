# Integration Tests

Comprehensive end-to-end integration testing suite for verifying complete workflows, concurrent operations, error recovery, and cross-platform compatibility.

## Organization

### Test Categories

- **e2e_test.go**: End-to-end workflow tests (manage, unmanage, remanage, adopt)
- **concurrent_test.go**: Concurrent operation and race condition tests
- **recovery_test.go**: Error recovery, rollback, and checkpoint tests
- **conflict_test.go**: Conflict detection and resolution tests
- **state_test.go**: Manifest persistence and incremental detection tests
- **query_test.go**: Status, doctor, and list command tests
- **cli_test.go**: CLI integration with flags and options
- **platform_test.go**: Cross-platform compatibility tests
- **scenario_test.go**: Realistic user scenario tests
- **benchmark_test.go**: Performance regression tests

## Running Tests

### All Integration Tests
```bash
go test ./tests/integration/...
```

### With Race Detector
```bash
go test -race ./tests/integration/...
```

### Specific Category
```bash
go test ./tests/integration/ -run TestE2E
go test ./tests/integration/ -run TestConcurrent
go test ./tests/integration/ -run TestRecovery
go test ./tests/integration/ -run TestConflict
go test ./tests/integration/ -run TestQuery
go test ./tests/integration/ -run TestState
go test ./tests/integration/ -run TestCLI
go test ./tests/integration/ -run TestPlatform
go test ./tests/integration/ -run TestScenario
```

### Benchmarks
```bash
go test -bench=. ./tests/integration/
go test -bench=BenchmarkManage ./tests/integration/
go test -bench=BenchmarkStatus ./tests/integration/
```

### With Verbose Output
```bash
go test -v ./tests/integration/...
```

### With Coverage
```bash
go test -cover ./tests/integration/...
go test -coverprofile=coverage.out ./tests/integration/...
```

### Short Mode (Skip Slow Tests)
```bash
go test -short ./tests/integration/...
```

### Parallel Execution
```bash
go test -parallel 4 ./tests/integration/...
```

## Test Categories by Phase

### Phase 17.1: Test Infrastructure
- `testutil/` package with fixtures, builders, assertions
- Golden test framework
- State snapshot utilities

### Phase 17.2: End-to-End Workflows
- Single and multiple package management
- Manage, unmanage, remanage operations
- Combined workflow tests

### Phase 17.3: Concurrent Testing
- Parallel package scanning and execution
- Concurrent operation isolation
- Race condition detection
- Stress testing with many packages

### Phase 17.4: Error Recovery
- Non-existent package handling
- Conflicting file detection
- Permission error handling
- Manifest corruption recovery
- Dry run verification

### Phase 17.5: Conflict Resolution
- File exists conflicts
- Wrong link target detection
- Directory vs file conflicts
- Multiple package overlap
- Broken symlink handling

### Phase 17.6: State Management
- Manifest creation and updates
- Incremental change detection
- Hash-based remanage
- State consistency

### Phase 17.7: Query Commands
- Status queries (all, specific packages)
- List operations
- Doctor health checks
- Performance testing

### Phase 17.8: Cross-Platform
- Path separator handling
- Symlink support verification
- Case sensitivity handling
- Platform-specific behaviors

### Phase 17.9: Performance
- Single package benchmarks
- Multiple package benchmarks (10, 100)
- Large file tree benchmarks
- Query operation benchmarks

### Phase 17.10: CLI Integration
- Command execution tests
- Flag and option handling
- Output format tests
- Exit code verification
- Environment variable support

### Phase 17.11: Scenario-Based
- New user setup workflow
- Development iteration workflow
- Multi-machine synchronization
- Selective installation
- Large repository management
- Backup and restore workflows

### Phase 17.12: Organization
- Documentation (this file)
- Test categorization
- CI integration guidelines
- Maintenance procedures

## Test Utilities

The `testutil` package provides:
- **FixtureBuilder**: Create test packages and directory structures
- **TestEnvironment**: Isolated test execution environment with cleanup
- **Assertions**: Specialized assertions for symlinks, files, directories
- **GoldenTest**: Compare outputs against golden files
- **StateSnapshot**: Capture and compare filesystem states
- **ClientHelper**: Easy client creation with test options

## Fixtures

Located in `tests/fixtures/`:
- **scenarios/**: Pre-built test scenarios (simple, complex, conflicts, migration)
- **packages/**: Sample packages (dotfiles, nvim, shell)
- **golden/**: Expected outputs for golden tests

Create fixtures dynamically using FixtureBuilder:
```go
env.FixtureBuilder().Package("vim").
    WithFile("dot-vimrc", "set nocompatible").
    WithFile("dot-vim/colors.vim", "colorscheme desert").
    Create()
```

## Test Principles

1. **Isolation**: Each test runs in isolated temporary directory
2. **Cleanup**: Automatic cleanup with defer safety
3. **Determinism**: Tests produce consistent results
4. **Parallelization**: Tests run concurrently where safe
5. **Coverage**: All workflows and error paths tested
6. **Speed**: Use `-short` flag to skip slow tests during development
7. **Race Detection**: Run with `-race` before committing

## Writing New Tests

### Test Naming Convention
- Test functions: `Test<Category>_<Feature>_<Scenario>`
- Benchmark functions: `Benchmark<Operation>_<Scenario>`
- Examples:
  - `TestE2E_Manage_SinglePackage`
  - `TestConcurrent_ParallelExecution`
  - `BenchmarkManage_100Packages`

### Test Structure Template
```go
func TestCategory_Feature(t *testing.T) {
    // Setup
    env := testutil.NewTestEnvironment(t)
    client := testutil.NewTestClient(t, env)
    
    // Create fixtures
    env.FixtureBuilder().Package("test").
        WithFile("dot-file", "content").
        Create()
    
    // Capture state if needed
    before := testutil.CaptureState(t, env.TargetDir)
    
    // Perform operation
    err := client.Manage(env.Context(), "test")
    require.NoError(t, err)
    
    // Verify results
    testutil.AssertLink(t, filepath.Join(env.TargetDir, ".file"), "...")
    
    // Verify state changes if needed
    after := testutil.CaptureState(t, env.TargetDir)
    // assertions...
}
```

### Benchmark Template
```go
func BenchmarkOperation(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        // Setup
        env := testutil.NewTestEnvironment(b)
        client := testutil.NewTestClient(b, env)
        // fixtures...
        
        b.StartTimer()
        // Operation to benchmark
        if err := client.Manage(context.Background(), "test"); err != nil {
            b.Fatal(err)
        }
    }
}
```

## CI Integration

### GitHub Actions Example
```yaml
- name: Run integration tests
  run: make test-integration

- name: Run with race detector
  run: go test -race ./tests/integration/...

- name: Upload coverage
  run: |
    go test -coverprofile=coverage.out ./tests/integration/...
    go tool cover -html=coverage.out -o coverage.html
```

### Test Selection for CI Stages
```bash
# Quick smoke tests
go test -short ./tests/integration/...

# Full test suite
go test ./tests/integration/...

# Race detection (slower)
go test -race ./tests/integration/...

# Benchmarks
go test -bench=. -benchtime=3x ./tests/integration/...
```

## Troubleshooting

### Flaky Tests
- Check for timing dependencies
- Verify proper cleanup
- Use deterministic test data
- Check for race conditions with `-race`

### Slow Tests
- Mark expensive tests with `if testing.Short() { t.Skip() }`
- Use benchmarks for performance testing
- Profile with `-cpuprofile` and `-memprofile`

### Permission Issues
- Tests running as root may behave differently
- Permission tests skip when running as root
- Ensure temp directories have proper permissions

### Platform Issues
- Windows symlink support requires admin/developer mode
- Use `filepath.Join` for path construction
- Test on target platforms before release

## Coverage Goals

- Overall integration test coverage: >75%
- Critical paths (manage, unmanage): 100%
- Error paths: >80%
- Concurrent code: 100% with race detector

## Maintenance

### Adding New Tests
1. Choose appropriate test category file
2. Follow naming conventions
3. Use testutil helpers
4. Add documentation if adding new patterns
5. Run with `-race` before committing

### Updating Tests
1. Keep tests isolated and independent
2. Update golden files with `-update-golden` flag
3. Verify changes don't break other tests
4. Update documentation if behavior changes

### Deprecating Tests
1. Mark test as deprecated with comment
2. Plan removal for next major version
3. Update documentation
4. Remove after transition period

## Performance Baselines

Current benchmarks on Apple M4 Pro:
- Single package manage: ~300μs
- 10 packages manage: ~3ms
- 100 packages manage: ~30ms
- Status query (10 packages): <1ms
- List query (10 packages): <1ms

Regression threshold: >10% performance degradation

## Contact

For questions about integration tests:
- See `testutil/doc.go` for utility documentation
- Check test examples in each category file
- Refer to phase-17-plan.md for detailed specifications

