# Testing Strategy and Paradigms

This document describes the comprehensive testing strategy employed by the dot project, including Test-Driven Development practices, testing patterns, infrastructure, and execution flows.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Categories](#test-categories)
- [Testing Infrastructure](#testing-infrastructure)
- [Layer-Specific Testing](#layer-specific-testing)
- [Testing Patterns](#testing-patterns)
- [Test Execution Flows](#test-execution-flows)
- [Coverage Requirements](#coverage-requirements)
- [Running Tests](#running-tests)

## Testing Philosophy

The dot project follows strict Test-Driven Development (TDD) principles with a focus on deterministic, isolated, and maintainable tests.

### Core Principles

1. **Test-First Development**: Tests are written before implementation code
2. **Red-Green-Refactor**: Follow the TDD cycle strictly
3. **Deterministic Tests**: Tests produce consistent results across environments
4. **Test Isolation**: Each test runs independently without shared state
5. **Fast Feedback**: Tests execute quickly to enable rapid iteration
6. **Comprehensive Coverage**: Enforced gate is mean per-function coverage, 60% locally and 75% in CI

### Testing Pyramid

```mermaid
graph TB
    subgraph "Testing Pyramid"
        E2E[End-to-End Tests<br/>Scenario-Based<br/>Full System Integration]:::e2eLayer
        Integration[Integration Tests<br/>Multi-Component<br/>Workflow Verification]:::integrationLayer
        Unit[Unit Tests<br/>Pure Functions<br/>Component Isolation]:::unitLayer
    end
    
    Unit --> Integration
    Integration --> E2E
    
    style E2E fill:#E67E22,stroke:#A84E0F,color:#fff
    style Integration fill:#3498DB,stroke:#1F618D,color:#fff
    style Unit fill:#2ECC71,stroke:#1E8449,color:#fff
    
    classDef e2eLayer stroke-width:2px
    classDef integrationLayer stroke-width:3px
    classDef unitLayer stroke-width:4px
```

## Test Categories

### Unit Tests

**Purpose**: Test individual functions and components in isolation.

**Characteristics**:
- No external dependencies (filesystem, network, database)
- Use memory-based adapters and mocks
- Fast execution (milliseconds)
- High coverage of edge cases
- Table-driven test patterns

**Location**: Colocated with implementation files (`*_test.go`)

**Example**:
```go
func TestScanPackage_ValidStructure(t *testing.T) {
    fs := adapters.NewMemFS()
    // Setup in-memory filesystem
    // Test scanning logic
}
```

### Integration Tests

**Purpose**: Test multiple components working together.

**Characteristics**:
- Test complete workflows
- Use real component interactions
- Isolated temporary directories
- Verify state transitions
- Test error propagation

**Location**: `tests/integration/`

**Categories**:
- `e2e_test.go`: Complete workflows (manage, unmanage, remanage)
- `concurrent_test.go`: Parallel operations and race conditions
- `recovery_test.go`: Error recovery and rollback mechanisms
- `conflict_test.go`: Conflict detection and resolution
- `state_test.go`: Manifest persistence and state management
- `query_test.go`: Status, doctor, and list commands
- `cli_test.go`: CLI integration with flags and options
- `cli_helpers_test.go`: Shared helpers for CLI invocation
- `clone_client_test.go`: Repository clone workflows
- `signal_test.go`: Signal handling and interrupt behavior
- `platform_test.go`: Cross-platform compatibility
- `scenario_test.go`: Realistic user workflows

`conflict_test.go` and `recovery_test.go` carry a `//go:build !windows`
constraint and are skipped on Windows.

### End-to-End Tests

**Purpose**: Validate complete system behavior from user perspective.

**Characteristics**:
- Test full CLI commands
- Verify user-facing outputs
- Test configuration precedence
- Validate error messages
- Cross-platform compatibility

### Benchmark Tests

**Purpose**: Detect performance regressions and measure operation costs.

**Location**: `tests/integration/benchmark_test.go`, plus package benchmarks under `internal/planner/`, `internal/scanner/`, and `internal/executor/`. `make bench` and `make bench-compare` run the three internal packages only.

**Benchmarks**:
- Single package operations
- Multiple package operations (10, 100, 1000 packages)
- Large file tree scenarios
- Query operation performance

### Fuzz Tests

**Purpose**: Exercise parsers and path validators against adversarial input.

**Location**: `internal/config/fuzz_test.go`, `internal/ignore/fuzz_test.go`, `internal/domain/fuzz_test.go`

**Run**: `make fuzz` (30s per target), or `go test -fuzz=FuzzGlobToRegex -run=^$ ./internal/ignore/`

**Targets**: config loading and validation, glob-to-regex translation and pattern matching, and package/target/file path construction.

## Testing Infrastructure

### Test Utilities Package

**Location**: `tests/integration/testutil/`

The testutil package provides comprehensive testing infrastructure:

```mermaid
graph LR
    TestEnvironment[TestEnvironment<br/>Isolated Execution]:::infraNode
    FixtureBuilder[FixtureBuilder<br/>Test Data Creation]:::infraNode
    Assertions[Assertions<br/>Specialized Checks]:::infraNode
    StateSnapshot[StateSnapshot<br/>State Capture]:::infraNode
    GoldenTest[GoldenTest<br/>Output Comparison]:::infraNode
    
    TestEnvironment --> FixtureBuilder
    TestEnvironment --> Assertions
    TestEnvironment --> StateSnapshot
    TestEnvironment --> GoldenTest
    
    style TestEnvironment fill:#4A90E2,stroke:#2C5F8D,color:#fff
    style FixtureBuilder fill:#50C878,stroke:#2D7A4A,color:#fff
    style Assertions fill:#3498DB,stroke:#1F618D,color:#fff
    style StateSnapshot fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style GoldenTest fill:#E67E22,stroke:#A84E0F,color:#fff
    
    classDef infraNode stroke-width:2px
```

#### TestEnvironment

Provides isolated test execution with automatic cleanup:

**Features**:
- Temporary directory creation
- Automatic cleanup on test completion
- Package and target directory setup
- Client instance creation
- Environment variable isolation

**Usage**:
```go
func TestExample(t *testing.T) {
    env := testutil.NewTestEnvironment(t)
    // Test code here
    // Automatic cleanup via defer
}
```

#### FixtureBuilder

Declarative test data creation:

**Features**:
- Package structure creation
- File and directory generation
- Symlink creation
- Content specification
- Nested structure support

**Usage**:
```go
env.FixtureBuilder().Package("vim").
    WithFile("dot-vimrc", "set nocompatible").
    WithDir("dot-vim").
    WithFile("dot-vim/colors.vim", "colorscheme desert").
    Create()
```

#### Assertions

Specialized assertion functions for filesystem verification:

**Functions**:
- `AssertLink`: Verify a symlink exists and points to the expected target
- `AssertLinkContains`: Verify a symlink target contains a substring
- `AssertFile`: Verify a file exists with exact content
- `AssertFileContains`: Verify file content contains a substring
- `AssertDir`: Verify a directory exists
- `AssertNotExists`: Verify a path is absent
- `AssertFileMode`: Verify file permission bits
- `AssertDirEmpty` / `AssertDirHasEntries`: Verify directory entry counts
- `AssertSymlinkChain`: Verify a multi-hop symlink resolution chain

#### StateSnapshot

Filesystem state capture and comparison:

**Features**:
- Capture directory tree state
- Compare states before/after operations
- Detect new files and directories
- Detect deletions
- Generate diff reports

#### GoldenTest (testutil, currently unused)

Compare test outputs against golden files:

**Features**:
- Save expected outputs as golden files
- Automatic comparison
- Update mode for golden file regeneration
- Diff display on mismatches

Note: the golden tests that actually ship live in `cmd/dot` and use
`internal/cli/golden` (`Golden.New` / `Assert` / `AssertString`), reading and
writing `cmd/dot/testdata/golden/<fixture>/<name>.golden` and updating via the
`-update` flag. `testutil.GoldenTest` is exercised only by its own unit tests.

### Test Fixtures

**Location**: `tests/fixtures/`, `cmd/dot/testdata/`, `internal/adapters/testdata/`

**Structure**:
```
tests/fixtures/
└── bootstrap-configs/          # Bootstrap config YAML samples
    ├── minimal.yaml
    ├── with-profiles.yaml
    ├── platform-specific.yaml
    ├── invalid-syntax.yaml
    └── invalid-missing-version.yaml

cmd/dot/testdata/golden/        # Golden outputs for CLI commands
├── adopt/
├── errors/
└── manage/

internal/adapters/testdata/     # Git adapter fixtures
└── test-repo/
```

## Layer-Specific Testing

### Domain Layer Testing

**Focus**: Pure function testing without side effects.

**Test Strategy**:
- Unit tests for all domain types
- Property-based testing for algebraic laws
- Result type behavior verification
- Error type construction
- Phantom type safety

**No Dependencies Required**:
- No filesystem access
- No external services
- Pure computation only

**Example Test Flow**:

```mermaid
sequenceDiagram
    participant Test as Test Function
    participant Domain as Domain Logic
    participant Result as Result[T]
    
    Test->>Domain: Call pure function
    Note over Domain: No side effects<br/>Deterministic computation
    Domain->>Result: Return Result[T]
    
    alt Success Case
        Test->>Result: IsOk()
        Result-->>Test: true
        Test->>Result: Unwrap()
        Result-->>Test: value
        Test->>Test: Assert value correct
    else Error Case
        Test->>Result: IsErr()
        Result-->>Test: true
        Test->>Result: UnwrapErr()
        Result-->>Test: error
        Test->>Test: Assert error correct
    end
```

### Core Layer Testing

**Focus**: Scanning, planning, and conflict resolution logic.

**Test Strategy**:
- Table-driven tests for multiple scenarios
- In-memory filesystem adapter
- Edge case coverage
- Error path verification

**Components Tested**:
- Scanner: Package structure traversal
- Planner: Operation dependency graphs
- Ignore: Pattern matching logic

**Example Test Flow**:

```mermaid
sequenceDiagram
    participant Test as Test Function
    participant MemFS as MemFilesystem
    participant Scanner as Scanner
    participant Planner as Planner
    
    Test->>MemFS: Setup test structure
    MemFS-->>Test: Filesystem ready
    
    Test->>Scanner: ScanPackage(fs, path)
    Scanner->>MemFS: ReadDir, Stat
    MemFS-->>Scanner: File metadata
    Scanner->>Scanner: Build package tree
    Scanner-->>Test: Package structure
    
    Test->>Test: Assert package correct
    
    Test->>Planner: ComputeDesiredState(packages)
    Planner->>Planner: Build dependency graph
    Planner->>Planner: Topological sort
    Planner-->>Test: Operations plan
    
    Test->>Test: Assert operations correct
    Test->>Test: Assert dependencies valid
```

### Pipeline Layer Testing

**Focus**: Stage composition and data flow.

**Test Strategy**:
- Integration tests with memory filesystem
- Error propagation through stages
- Context cancellation handling
- Type safety verification

**Test Flow**:

```mermaid
sequenceDiagram
    participant Test as Test Function
    participant Pipeline as Pipeline
    participant ScanStage as Scan Stage
    participant PlanStage as Plan Stage
    participant ResolveStage as Resolve Stage
    
    Test->>Test: Create test context
    Test->>Pipeline: Execute(ctx, input)
    
    Pipeline->>ScanStage: Execute
    ScanStage->>ScanStage: Process
    alt Scan Success
        ScanStage-->>Pipeline: []Package
        Pipeline->>PlanStage: Execute
        PlanStage->>PlanStage: Process
        alt Plan Success
            PlanStage-->>Pipeline: DesiredState
            Pipeline->>ResolveStage: Execute
            ResolveStage->>ResolveStage: Process
            alt Resolve Success
                ResolveStage-->>Pipeline: Plan
                Pipeline-->>Test: Success result
            else Resolve Error
                ResolveStage-->>Pipeline: Error
                Pipeline-->>Test: Error result
            end
        else Plan Error
            PlanStage-->>Pipeline: Error
            Pipeline-->>Test: Error result
        end
    else Scan Error
        ScanStage-->>Pipeline: Error
        Pipeline-->>Test: Error result
    end
    
    Test->>Test: Assert result correct
```

### Executor Layer Testing

**Focus**: Transaction execution, rollback, and checkpointing.

**Test Strategy**:
- Verify precondition validation
- Test checkpoint creation and restoration
- Verify rollback on failure
- Test parallel execution
- Verify atomic guarantees

**Two-Phase Commit Test Flow**:

```mermaid
sequenceDiagram
    participant Test as Test Function
    participant Executor as Executor
    participant Checkpoint as Checkpoint Store
    participant FS as Filesystem
    
    Test->>Executor: Execute(plan)
    
    rect rgb(40, 70, 100)
        note right of Executor: Prepare Phase
        Executor->>Executor: Validate operations
        Executor->>Checkpoint: Create checkpoint
        Checkpoint->>FS: Save state
        FS-->>Checkpoint: Checkpoint saved
        Checkpoint-->>Executor: Ready
    end
    
    rect rgb(60, 50, 100)
        note right of Executor: Commit Phase
        loop For each operation
            Executor->>FS: Execute operation
            alt Success
                FS-->>Executor: OK
            else Failure (Injected)
                FS-->>Executor: Error
                Executor->>Executor: Abort remaining
            end
        end
    end
    
    rect rgb(100, 60, 60)
        note right of Executor: Rollback Phase
        Executor->>Checkpoint: Load checkpoint
        Checkpoint->>FS: Read saved state
        FS-->>Checkpoint: State data
        Checkpoint-->>Executor: Checkpoint loaded
        
        loop For each executed operation (reverse)
            Executor->>FS: Undo operation
            FS-->>Executor: Undone
        end
        
        Executor->>Checkpoint: Delete checkpoint
        Checkpoint->>FS: Remove checkpoint file
    end
    
    Executor-->>Test: ExecutionError
    Test->>FS: Verify rollback complete
    FS-->>Test: Original state restored
    Test->>Test: Assert rollback successful
```

### API Layer Testing

**Focus**: Service integration and client operations.

**Test Strategy**:
- End-to-end service testing
- Manifest persistence verification
- Service interaction validation
- Configuration handling

### CLI Layer Testing

**Focus**: Command execution and user interaction.

**Test Strategy**:
- Command parsing verification
- Flag and option handling
- Output format testing
- Exit code validation
- Error message verification

## Testing Patterns

### Table-Driven Testing

Table-driven tests enable comprehensive scenario coverage with minimal code duplication.

**Pattern**:
```go
func TestOperation(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        want     Output
        wantErr  bool
    }{
        {
            name: "valid input",
            input: Input{...},
            want: Output{...},
            wantErr: false,
        },
        {
            name: "invalid input",
            input: Input{...},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Operation(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Table-Driven Test Flow**:

```mermaid
graph TB
    Start([Start Test])
    DefineTable[Define Test Cases Table]
    LoopStart{More Cases?}
    GetCase[Get Next Test Case]
    RunSubtest[Run Subtest t.Run]
    Setup[Setup Test Input]
    Execute[Execute Function]
    Verify[Verify Output]
    Cleanup[Cleanup Resources]
    LoopEnd[Next Case]
    End([End Test])
    
    Start --> DefineTable
    DefineTable --> LoopStart
    LoopStart -->|Yes| GetCase
    GetCase --> RunSubtest
    RunSubtest --> Setup
    Setup --> Execute
    Execute --> Verify
    Verify --> Cleanup
    Cleanup --> LoopEnd
    LoopEnd --> LoopStart
    LoopStart -->|No| End
    
    style Start fill:#2ECC71,stroke:#1E8449,color:#fff
    style DefineTable fill:#3498DB,stroke:#1F618D,color:#fff
    style Execute fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style Verify fill:#E67E22,stroke:#A84E0F,color:#fff
    style End fill:#2ECC71,stroke:#1E8449,color:#fff
```

### Property-Based Testing

Property-based testing verifies algebraic laws and invariants.

**Pattern**:
```go
func TestResultMonad_IdentityLaw(t *testing.T) {
    // For any value v:
    // Return(v).Bind(f) == f(v)
    
    v := 42
    f := func(x int) Result[int] {
        return Ok(x * 2)
    }
    
    left := Ok(v).Bind(f)
    right := f(v)
    
    assert.Equal(t, left, right)
}
```

### Golden File Testing

Golden file tests compare outputs against reference files.

**Test Flow**:

```mermaid
sequenceDiagram
    participant Test as Test Function
    participant System as System Under Test
    participant Golden as Golden File Manager
    participant FS as Filesystem
    
    Test->>System: Execute operation
    System-->>Test: Output
    
    Test->>Golden: CompareWithGolden(output)
    
    alt Update Mode (-update flag)
        Golden->>FS: Write output to golden file
        FS-->>Golden: Written
        Golden-->>Test: Updated golden file
    else Compare Mode
        Golden->>FS: Read golden file
        FS-->>Golden: Expected output
        Golden->>Golden: Compare actual vs expected
        
        alt Match
            Golden-->>Test: Pass
        else Mismatch
            Golden->>Golden: Generate diff
            Golden-->>Test: Fail with diff
        end
    end
```

### Concurrent Testing

Concurrent tests verify thread safety and race condition freedom.

**Test Strategy**:
- Use `t.Parallel()` for concurrent test execution
- Run with `-race` flag to detect data races
- Test concurrent client operations
- Verify manifest locking
- Test parallel operation execution

**Concurrent Test Flow**:

```mermaid
sequenceDiagram
    participant Test as Test Function
    participant Goroutine1 as Goroutine 1
    participant Goroutine2 as Goroutine 2
    participant Goroutine3 as Goroutine 3
    participant Client as Client (Shared)
    participant Manifest as Manifest Store
    
    Test->>Test: Setup shared client
    
    par Goroutine 1
        Test->>Goroutine1: Start
    and Goroutine 2
        Test->>Goroutine2: Start
    and Goroutine 3
        Test->>Goroutine3: Start
    end
    
    Goroutine1->>Client: Manage("package1")
    Goroutine2->>Client: Manage("package2")
    Goroutine3->>Client: Status()
    
    Client->>Manifest: Lock
    Client->>Manifest: Update
    Client->>Manifest: Unlock
    
    Client-->>Goroutine1: Result 1
    Client-->>Goroutine2: Result 2
    Client-->>Goroutine3: Result 3
    
    Goroutine1-->>Test: Complete
    Goroutine2-->>Test: Complete
    Goroutine3-->>Test: Complete
    
    Test->>Test: Wait for all goroutines
    Test->>Test: Assert no data races
    Test->>Test: Verify all operations succeeded
```

## Test Execution Flows

### TDD Cycle

```mermaid
stateDiagram-v2
    [*] --> WriteTest: Write Failing Test
    
    WriteTest --> RunTestRed: Run Test
    RunTestRed --> VerifyFailure: Verify Correct Failure
    VerifyFailure --> ImplementCode: Implement Code
    
    ImplementCode --> RunTestGreen: Run Test
    RunTestGreen --> CheckPass: Test Passes?
    
    CheckPass --> Refactor: Yes
    CheckPass --> DebugCode: No
    
    DebugCode --> RunTestGreen
    
    Refactor --> RunTestStillGreen: Run Tests
    RunTestStillGreen --> CheckStillPass: Tests Pass?
    
    CheckStillPass --> Commit: Yes
    CheckStillPass --> FixRefactor: No
    
    FixRefactor --> RunTestStillGreen
    
    Commit --> [*]
    
    note right of WriteTest
        RED: Write test that fails
        for the right reason
    end note
    
    note right of ImplementCode
        GREEN: Write minimum code
        to make test pass
    end note
    
    note right of Refactor
        REFACTOR: Improve code
        while keeping tests green
    end note
```

### Complete Test Execution Flow

```mermaid
graph TB
    Start([make check])
    
    UnitTests[Run Unit Tests<br/>Pure Functions]
    CoreTests[Run Core Layer Tests<br/>Scanner, Planner, Ignore]
    PipelineTests[Run Pipeline Tests<br/>Stage Composition]
    ExecutorTests[Run Executor Tests<br/>Transactions, Rollback]
    APITests[Run API Tests<br/>Service Integration]
    CLITests[Run CLI Tests<br/>Command Integration]
    IntegrationTests[Run Integration Tests<br/>E2E Workflows]
    
    RaceDetector{Race Detector<br/>Enabled?}
    CoverageCalc[Calculate Coverage]
    CoverageCheck{Mean func coverage >= 60%?}
    
    AllPass{All Tests<br/>Pass?}
    
    Success([Success])
    Failure([Failure])
    
    Start --> UnitTests
    UnitTests --> CoreTests
    CoreTests --> PipelineTests
    PipelineTests --> ExecutorTests
    ExecutorTests --> APITests
    APITests --> CLITests
    CLITests --> IntegrationTests
    
    IntegrationTests --> RaceDetector
    RaceDetector -->|Yes| CoverageCalc
    RaceDetector -->|No| AllPass
    
    CoverageCalc --> CoverageCheck
    CoverageCheck -->|Yes| AllPass
    CoverageCheck -->|No| Failure
    
    AllPass -->|Yes| Success
    AllPass -->|No| Failure
    
    style Start fill:#3498DB,stroke:#1F618D,color:#fff
    style Success fill:#2ECC71,stroke:#1E8449,color:#fff
    style Failure fill:#E74C3C,stroke:#C0392B,color:#fff
```

### Integration Test Execution

```mermaid
sequenceDiagram
    participant Runner as Test Runner
    participant Env as TestEnvironment
    participant Fixture as FixtureBuilder
    participant Client as Client
    participant Snapshot as StateSnapshot
    participant Assert as Assertions
    
    Runner->>Env: NewTestEnvironment(t)
    Env->>Env: Create temp directories
    Env->>Env: Setup cleanup hooks
    Env-->>Runner: Environment ready
    
    Runner->>Fixture: Create test packages
    Fixture->>Env: Write files to package dir
    Env-->>Fixture: Files created
    
    Runner->>Snapshot: Capture initial state
    Snapshot->>Env: Scan target directory
    Env-->>Snapshot: State data
    Snapshot-->>Runner: Initial snapshot
    
    Runner->>Client: Execute operation
    Client->>Client: Perform workflow
    Client-->>Runner: Operation result
    
    Runner->>Snapshot: Capture final state
    Snapshot->>Env: Scan target directory
    Env-->>Snapshot: State data
    Snapshot-->>Runner: Final snapshot
    
    Runner->>Assert: Verify symlinks created
    Assert->>Env: Check filesystem
    Env-->>Assert: Verification result
    Assert-->>Runner: Assertions passed
    
    Runner->>Env: Cleanup
    Env->>Env: Remove temp directories
    Env-->>Runner: Cleanup complete
```

## Coverage Requirements

### Enforced Coverage Gate

The gate is not total statement coverage. Both `make check-coverage` and CI
compute the unweighted mean of the per-function percentages emitted by
`go tool cover -func=coverage.out`, after excluding the Bubble Tea UI and
interactive adoption files:

- `internal/cli/adopt/selector.go`
- `internal/cli/adopt/scanner.go`
- `internal/cli/adopt/interactive.go`
- `internal/cli/adopt/discovery.go`

Thresholds:

- Local (`make check-coverage`, `make cs`, and therefore `make check`): 60.0%
- CI (`.github/workflows/ci.yml`, `test` job): 75.0%

### Per-Layer Targets

The per-layer targets below are aspirational; no tooling enforces them.

- **Domain Layer**: 100% (critical path)
- **Core Layer**: 95% minimum
- **Pipeline Layer**: 90% minimum
- **Executor Layer**: 100% (critical path)
- **API Layer**: 85% minimum
- **CLI Layer**: 75% minimum

### Coverage Verification

```bash
# Generate coverage report
make coverage

# View coverage in browser
go tool cover -html=coverage.out

# Enforce the local coverage gate (requires coverage.out from `make test`)
make check-coverage
```

### Critical Paths Requiring 100% Coverage

1. **Operation Execution**: All operation types
2. **Rollback Mechanism**: Complete rollback path
3. **Manifest Persistence**: Read, write, validation
4. **Error Handling**: All error types and wrapping
5. **Path Validation**: Security-critical path operations

## Running Tests

### Basic Test Execution

```bash
# All tests
make test

# With race detection
go test -race ./...

# With coverage profile and HTML report
make coverage

# Enforce the local coverage gate (requires coverage.out from `make test`)
make check-coverage

# Verbose output
go test -v ./...
```

### Test Categories

```bash
# Unit tests only
go test ./internal/...

# Integration tests only
go test ./tests/integration/...

# Specific test function
go test -run TestE2E_Manage_SinglePackage ./tests/integration/

# Specific package with a run filter
go test -run TestScanPackage ./internal/scanner/
```

### Test Options

```bash
# Short mode (skip slow tests)
go test -short ./...

# Parallel execution
go test -parallel 4 ./...

# With timeout
go test -timeout 30s ./...

# Benchmarks (`make bench` targets the same packages)
go test -bench=. -benchmem -run=^$ ./internal/planner/ ./internal/scanner/ ./internal/executor/

# CPU profiling
go test -cpuprofile=cpu.prof ./...

# Memory profiling
go test -memprofile=mem.prof ./...
```

### Continuous Integration

```bash
# CI-friendly test execution
make check

# Human-friendly test output
make qa
```

### Test Debugging

```bash
# Run single test with verbose output
go test -v -run TestSpecificTest ./internal/package/

# Print test logs
go test -v ./... 2>&1 | tee test.log

# Use delve debugger
dlv test ./internal/package/ -- -test.run TestSpecificTest
```

## Test Maintenance

### Adding New Tests

1. **Identify Test Category**: Unit, integration, or E2E
2. **Create Test File**: Follow naming convention `*_test.go`
3. **Write Test First**: Follow TDD red-green-refactor
4. **Use Table-Driven Pattern**: For multiple scenarios
5. **Add Documentation**: Document test purpose and setup
6. **Verify Coverage**: Ensure coverage thresholds met

### Test Naming Conventions

```
Test<Layer>_<Component>_<Scenario>
Benchmark<Operation>_<Scenario>
Example<Function>_<UseCase>
```

**Examples**:
- `TestScanner_ScanPackage_ValidStructure`
- `TestExecutor_Rollback_OnFailure`
- `TestClient_Manage_MultiplePackages`
- `BenchmarkManage_100Packages`

### Updating Golden Files

```bash
# Update golden files for the CLI command tests
go test ./cmd/dot/ -update

# Update a single command's golden files
go test ./cmd/dot/ -run TestManageGolden -update

# The tests/integration/testutil helper uses a separate flag
go test ./tests/integration/ -update-golden
```

## Test Architecture Best Practices

1. **Arrange-Act-Assert**: Structure tests with clear phases
2. **Single Assertion Focus**: Each test verifies one behavior
3. **Descriptive Names**: Test names document expected behavior
4. **Minimal Setup**: Keep test setup simple and focused
5. **Isolated State**: No shared state between tests
6. **Fast Execution**: Optimize for quick feedback
7. **Deterministic Results**: Tests pass consistently
8. **Comprehensive Error Testing**: Test all error paths

## References

### Related Documentation

- [Architecture Documentation](architecture.md) - System design and layers
- [Contributing Guide](../../CONTRIBUTING.md) - Development workflow
- [Integration Tests README](../../tests/integration/README.md) - Integration test details

### External Resources

- [Test-Driven Development](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Go Testing Package](https://pkg.go.dev/testing)
- [testify](https://github.com/stretchr/testify) - Assertion library

## Navigation

**[↑ Back to Documentation Index](../README.md)** | [Architecture](architecture.md) | [Release Workflow](release-workflow.md)

