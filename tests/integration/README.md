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
```

### Benchmarks
```bash
go test -bench=. ./tests/integration/
```

### With Verbose Output
```bash
go test -v ./tests/integration/...
```

## Test Utilities

The `testutil` package provides:
- **FixtureBuilder**: Create test packages and directory structures
- **TestEnvironment**: Isolated test execution environment
- **Assertions**: Specialized assertions for symlinks, manifests, etc.
- **GoldenTest**: Compare outputs against golden files
- **StateSnapshot**: Capture and compare filesystem states

## Fixtures

Located in `tests/fixtures/`:
- **scenarios/**: Pre-built test scenarios (simple, complex, conflicts, migration)
- **packages/**: Sample packages (dotfiles, nvim, shell)
- **golden/**: Expected outputs for golden tests

## Test Principles

1. **Isolation**: Each test runs in isolated temporary directory
2. **Cleanup**: Automatic cleanup with defer safety
3. **Determinism**: Tests produce consistent results
4. **Parallelization**: Tests run concurrently where safe
5. **Coverage**: All workflows and error paths tested

