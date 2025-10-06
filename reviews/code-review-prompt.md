# Code Review Prompt for dot CLI Project

## Overview
This prompt guides comprehensive code review for the `dot` CLI project, a type-safe dotfile and configuration management tool built with Go 1.25.1. The review must verify constitutional compliance, architectural alignment, and adherence to functional programming principles.

## Project Context

**Project**: dot - Dotfile and Configuration Management Tool  
**Language**: Go 1.25.1  
**Architecture**: Functional Core, Imperative Shell  
**Current Phase**: Phase 15 (Error Handling and User Experience)  
**Test Coverage Requirement**: 80% minimum  
**Commit Standard**: Conventional Commits v1.0.0  

### Technology Stack
- **CLI Framework**: Cobra
- **Configuration**: Viper with XDG compliance
- **Logging**: log/slog with console-slog adapter
- **Testing**: testify assertions
- **Linting**: golangci-lint v2

---

## Constitutional Principles Verification

### I. Test-First Development (TDD)
- [ ] All new features have tests written before implementation
- [ ] Tests follow red-green-refactor cycle
- [ ] No code exists without corresponding tests
- [ ] Test coverage meets 80% minimum threshold
- [ ] All tests pass without failures
- [ ] Test files follow naming convention `*_test.go`
- [ ] Table-driven tests used for multiple scenarios
- [ ] Tests use testify assertions (not gotest.tools/v3)

**Review Questions**:
1. Can you identify test files for each implementation file?
2. Do tests cover edge cases and error paths?
3. Are integration tests separated from unit tests?
4. Do tests verify both positive and negative cases?

### II. Atomic Commits
- [ ] Each commit represents complete, working state
- [ ] One logical change per commit
- [ ] Conventional Commits format enforced
- [ ] Breaking changes documented in footers
- [ ] Commit messages are descriptive and precise
- [ ] Subject line under 50 characters
- [ ] Body wraps at 72 characters
- [ ] Scope is mandatory and specific

**Review Questions**:
1. Can each commit be reverted independently?
2. Do commit messages explain why, not just what?
3. Are breaking changes properly documented?
4. Does the commit history serve as documentation?

### III. Functional Programming Preference
- [ ] Functions are pure where practical (no side effects)
- [ ] Mutable state is minimized
- [ ] Higher-order functions used for composition
- [ ] Errors returned explicitly, never panic
- [ ] Functions are primary unit of composition
- [ ] Closures used for encapsulation
- [ ] No global state where avoidable

**Review Questions**:
1. Are functions deterministic and testable?
2. Is state mutation isolated to imperative shell?
3. Are side effects clearly separated from business logic?
4. Can functions be composed and reused easily?

### IV. Standard Technology Stack
- [ ] Uses specified dependencies (Cobra, Viper, slog)
- [ ] No unauthorized dependencies added
- [ ] Dependency versions pinned in go.mod
- [ ] Standard library preferred over third-party when possible
- [ ] Uses `errors` package, not `github.com/pkg/errors`
- [ ] Deviations are documented with justification

### V. Academic Documentation Standard
- [ ] Factual, objective writing style
- [ ] No hyperbole or subjective qualifiers
- [ ] No emojis in code, docs, comments, or output
- [ ] Technical precision in terminology
- [ ] Documentation serves as reference, not marketing
- [ ] Code comments explain why, not what

**Prohibited Language**:
- "awesome", "quick", "simple", "easy"
- Emojis of any kind
- Marketing language
- Hyperbolic statements

### VI. Code Quality Gates
- [ ] All linters pass (golangci-lint v2)
- [ ] Cyclomatic complexity ≤ 15
- [ ] Security checks pass (gosec)
- [ ] Import organization enforced (goimports)
- [ ] No code duplication detected
- [ ] All formatters run before commit
- [ ] No warnings tolerated

---

## Architectural Compliance

### Layered Architecture Verification

#### Domain Layer (`pkg/dot`)
- [ ] Pure domain model with phantom-typed paths
- [ ] Value objects are immutable
- [ ] No external dependencies
- [ ] Type safety enforced at compile time
- [ ] Business logic isolated from infrastructure

**Review Questions**:
1. Are domain types self-contained and testable?
2. Does the domain layer depend on infrastructure?
3. Are phantom types preventing path type confusion?

#### Port Layer (Interfaces in `internal/`)
- [ ] Infrastructure interfaces defined
- [ ] Ports are abstract and implementation-agnostic
- [ ] No concrete implementations in port definitions
- [ ] Interface segregation principle followed

#### Adapter Layer (`internal/adapters/`)
- [ ] Concrete implementations of ports
- [ ] Filesystem adapters (osfs, memfs) implement same interface
- [ ] Logger adapters wrap underlying implementations
- [ ] Adapters are testable in isolation

**Review Questions**:
1. Can adapters be swapped without changing business logic?
2. Are adapters properly mocked in tests?
3. Do all adapters implement their ports correctly?

#### Core Layer (`internal/api`, `internal/planner`, etc.)
- [ ] Pure functional planning logic
- [ ] No direct I/O operations
- [ ] Resolution logic separated from execution
- [ ] Topological sorting for dependency management
- [ ] Error collection without failing fast silently

**Review Questions**:
1. Can core logic be tested without I/O?
2. Is planning deterministic given same inputs?
3. Are all errors collected and returned?

#### Shell Layer (`internal/executor`)
- [ ] Side-effecting execution isolated
- [ ] Transactional operations with rollback
- [ ] Checkpoint-based recovery
- [ ] Parallel execution where safe
- [ ] Error handling with cleanup

**Review Questions**:
1. Are transactions properly isolated?
2. Can operations be rolled back on failure?
3. Is error cleanup comprehensive?

#### API Layer (`internal/api`)
- [ ] Public Go library interface
- [ ] Zero CLI dependencies
- [ ] Client interface for operations
- [ ] Operation results properly typed
- [ ] Embeddable in other applications

#### CLI Layer (`cmd/dot`)
- [ ] Cobra-based commands
- [ ] Thin wrapper around API layer
- [ ] Flag parsing and validation
- [ ] Output rendering separated
- [ ] No business logic in commands

**Review Questions**:
1. Do CLI commands only orchestrate API calls?
2. Is output rendering consistent across commands?
3. Are flags properly validated?

---

## Code Quality Review

### Error Handling
- [ ] All errors handled explicitly
- [ ] Errors never ignored without documented justification
- [ ] Context added with `fmt.Errorf` and `%w` verb
- [ ] Errors returned to caller, not just logged
- [ ] Panic only for unrecoverable programming errors
- [ ] Error messages are actionable and clear
- [ ] Custom error types used where appropriate

**Anti-patterns to Flag**:
```go
// Bad: Ignored error
os.Remove(filename)

// Bad: Panic for recoverable error
if err != nil { panic(err) }

// Bad: Lost error context
return errors.New("failed to read file")

// Good: Explicit handling with context
if err := os.Remove(filename); err != nil {
    return fmt.Errorf("remove temp file: %w", err)
}
```

### Function Design
- [ ] Functions are small and focused (< 50 lines ideal)
- [ ] Single responsibility principle
- [ ] No naked returns in functions > 10 lines
- [ ] Parameter count reasonable (< 5 parameters)
- [ ] Return values clearly documented
- [ ] Side effects documented in comments

### Type Safety
- [ ] Phantom types prevent path confusion
- [ ] Type parameters used for compile-time safety
- [ ] No unnecessary type assertions
- [ ] Interfaces used for abstraction
- [ ] Value objects enforce invariants

### Concurrency
- [ ] Goroutines used appropriately
- [ ] No data races (verified with `go test -race`)
- [ ] Channels used correctly
- [ ] Context propagation for cancellation
- [ ] WaitGroups or sync primitives properly used
- [ ] Mutex usage justified and minimal

---

## Testing Quality Review

### Unit Tests
- [ ] Each function has corresponding unit tests
- [ ] Tests are isolated and independent
- [ ] No shared state between tests
- [ ] Tests use t.Parallel() where safe
- [ ] Mock dependencies properly
- [ ] Test names describe scenario: `TestFunctionName_Condition_ExpectedResult`

### Table-Driven Tests
- [ ] Multiple scenarios in single test function
- [ ] Clear test case structure with name, input, expected
- [ ] Edge cases included
- [ ] Error cases tested
- [ ] Easy to add new cases

**Example Pattern**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {name: "valid input", input: "test", expected: "result", wantErr: false},
        {name: "empty input", input: "", expected: "", wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Integration Tests
- [ ] Located in `tests/integration/`
- [ ] Test complete workflows
- [ ] Use test fixtures appropriately
- [ ] Clean up resources after tests
- [ ] Test error scenarios
- [ ] Verify transactional behavior

### Test Coverage
- [ ] Overall coverage ≥ 80%
- [ ] Critical paths have > 90% coverage
- [ ] Error paths are tested
- [ ] Edge cases covered
- [ ] No unreachable code

**Verification**:
```bash
go test -cover -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Test Quality
- [ ] Tests are readable and maintainable
- [ ] Assertions use testify require/assert appropriately
- [ ] Test failures provide clear messages
- [ ] Tests run quickly (< 1 second per unit test)
- [ ] No flaky tests (random failures)

---

## Security Review

### Input Validation
- [ ] All user input validated
- [ ] Command-line arguments sanitized
- [ ] Configuration values validated
- [ ] File paths checked for traversal
- [ ] Whitelist approach for allowed values
- [ ] Length limits enforced
- [ ] Format validation using regex

**Critical Areas**:
1. Path handling in symlink operations
2. Package name validation
3. Configuration file parsing
4. Command execution (if any)

### File System Security
- [ ] File permissions set restrictively (0600 for files, 0700 for dirs)
- [ ] Temporary files created securely
- [ ] Path traversal prevented
- [ ] Symlink following controlled
- [ ] No world-readable sensitive files

### Credential Handling
- [ ] No hardcoded credentials
- [ ] Credentials from environment variables
- [ ] Sensitive data not logged
- [ ] Configuration files have secure permissions
- [ ] API keys properly protected

### Command Execution
- [ ] No shell execution with user input
- [ ] Parameterized commands only
- [ ] Whitelist of allowed operations
- [ ] Arguments validated before execution

### Dependency Security
- [ ] Dependencies pinned to specific versions
- [ ] No known vulnerabilities (run `govulncheck`)
- [ ] Minimal dependency count
- [ ] License compliance verified

---

## Documentation Review

### Code Documentation
- [ ] Public functions have godoc comments
- [ ] Package documentation present
- [ ] Complex logic explained with comments
- [ ] Examples provided for public APIs
- [ ] No commented-out code

**Godoc Format**:
```go
// FunctionName describes what the function does.
// It continues with more detail if needed.
//
// Parameter explanation if complex.
// Returns explanation.
func FunctionName(param string) (string, error) { }
```

### README and Documentation Files
- [ ] README.md accurate and current
- [ ] Installation instructions clear
- [ ] Usage examples provided
- [ ] Architecture documented
- [ ] Contributing guidelines present
- [ ] No outdated information

### Inline Comments
- [ ] Comments explain why, not what
- [ ] Complex algorithms documented
- [ ] Assumptions stated
- [ ] Trade-offs explained
- [ ] TODO/FIXME items have issue references

---

## Performance Review

### Memory Efficiency
- [ ] Slices pre-allocated when size known
- [ ] Unnecessary allocations avoided
- [ ] Large data structures passed by pointer
- [ ] Memory leaks checked (profiling)
- [ ] Defer usage appropriate (not in loops)

### Algorithm Efficiency
- [ ] Appropriate data structures chosen
- [ ] Time complexity reasonable
- [ ] No unnecessary iterations
- [ ] Caching used where beneficial
- [ ] Streaming used for large datasets

### Benchmarks
- [ ] Performance-critical code has benchmarks
- [ ] Benchmark results documented
- [ ] No performance regressions
- [ ] Profiling done for optimizations

**Benchmark Pattern**:
```go
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Function(input)
    }
}
```

---

## Configuration Review

### Viper Integration
- [ ] Configuration in `internal/config` package
- [ ] Direct Viper usage isolated
- [ ] XDG specification compliance
- [ ] Multiple format support (YAML, JSON, TOML)
- [ ] Precedence: Flags > Environment > Config Files > Defaults

### Configuration Files
- [ ] Stored in `$XDG_CONFIG_HOME/<app>/config.yaml`
- [ ] Sensible defaults provided
- [ ] Validation on load
- [ ] Error messages for invalid config
- [ ] Migration path for config changes

---

## CLI/UX Review

### Command Design
- [ ] Commands are intuitive and consistent
- [ ] Help text is clear and comprehensive
- [ ] Examples provided for complex commands
- [ ] Flags follow conventions
- [ ] Subcommands logically grouped

### Output Quality
- [ ] Consistent formatting across commands
- [ ] Multiple output formats supported (table, JSON, YAML)
- [ ] Progress indication for long operations
- [ ] Error messages are actionable
- [ ] Verbose mode provides detailed information
- [ ] Quiet mode respects user preference

### User Experience
- [ ] Dry-run mode available for destructive operations
- [ ] Confirmation prompts for dangerous actions
- [ ] Clear success/failure feedback
- [ ] Helpful error messages with suggestions
- [ ] Exit codes follow conventions (0 = success)

---

## Linting and Formatting

### Required Checks
- [ ] `golangci-lint run` passes with zero warnings
- [ ] `go vet ./...` passes
- [ ] `gofmt -s` applied to all files
- [ ] `goimports` applied for import organization
- [ ] No cyclomatic complexity > 15
- [ ] No cognitive complexity issues

### Linter Configuration
Verify `.golangci.yml` includes:
- errcheck: unhandled errors
- gosec: security issues
- staticcheck: static analysis
- unused: unused code detection
- ineffassign: ineffective assignments
- gocyclo: cyclomatic complexity
- dupl: code duplication
- misspell: spelling errors

---

## Git and Version Control

### Branch Management
- [ ] Branch naming: kebab-case only (no path separators)
- [ ] Feature branches from main
- [ ] Branch up to date with main
- [ ] No merge conflicts

**Correct**: `feature-error-handling`, `fix-123-path-issue`  
**Incorrect**: `feature/error-handling`, `fix/123-path-issue`

### Commit Quality
- [ ] Each commit passes tests and lints
- [ ] Commits are atomic
- [ ] Commit messages follow Conventional Commits
- [ ] No `--no-verify` flag used
- [ ] No fixup commits in final PR

### Pull Request Checklist
- [ ] PR description explains changes
- [ ] Related issues referenced
- [ ] Screenshots for UI changes
- [ ] Breaking changes documented
- [ ] Migration guide provided if needed

---

## Prohibited Practices Verification

### Absolute Prohibitions
- [ ] No emojis in code, docs, comments, or output
- [ ] No `github.com/pkg/errors` package used
- [ ] No `gotest.tools/v3` for testing
- [ ] No ignored errors without justification
- [ ] No naked returns in functions > 10 lines
- [ ] No committed code with test failures
- [ ] No committed code with linting errors
- [ ] No global state where avoidable
- [ ] No panic for recoverable errors
- [ ] No `--no-verify` in git commands
- [ ] No force push to main/master

### Shell-Specific Issues
- [ ] No exclamation marks in git commit `-m` messages with double quotes
- [ ] Proper escaping of special characters

---

## Review Checklist Summary

### Critical Issues (Must Fix)
- [ ] Test failures
- [ ] Linting errors
- [ ] Security vulnerabilities
- [ ] Breaking changes without documentation
- [ ] Constitutional violations
- [ ] Architectural violations

### Important Issues (Should Fix)
- [ ] Missing tests for new code
- [ ] Incomplete error handling
- [ ] Performance concerns
- [ ] Documentation gaps
- [ ] UX inconsistencies

### Minor Issues (Nice to Have)
- [ ] Code style improvements
- [ ] Additional test coverage
- [ ] Refactoring opportunities
- [ ] Documentation enhancements

---

## Review Process

1. **Initial Scan**: Check for prohibited practices and obvious issues
2. **Constitutional Compliance**: Verify adherence to project principles
3. **Architectural Review**: Ensure proper layer separation
4. **Code Quality**: Examine implementation details
5. **Testing**: Verify test coverage and quality
6. **Security**: Check for vulnerabilities
7. **Documentation**: Review clarity and completeness
8. **Integration**: Run `make check` to verify all quality gates
9. **Final Assessment**: Summarize findings and recommendations

---

## Commands to Run During Review

```bash
# Run all quality checks
make check

# Run tests with coverage
make test

# Run linters
make lint

# Check formatting
make fmt

# Run tests with pretty output
make qa

# Check for vulnerabilities
govulncheck ./...

# Check for race conditions
go test -race ./...

# Generate coverage report
make coverage

# Verify dependencies
make deps-verify
```

---

## Review Output Template

```markdown
## Code Review Summary

**Reviewer**: [Name]
**Date**: [Date]
**Branch**: [branch-name]
**Commit**: [commit-hash]

### Constitutional Compliance
- [x] TDD principles followed
- [x] Atomic commits
- [x] Functional programming principles
- [x] Academic documentation standard
- [x] Code quality gates passed

### Architecture
- [x] Layer separation maintained
- [x] Pure core, imperative shell
- [x] Type safety enforced
- [x] Proper abstraction

### Quality Metrics
- Test Coverage: [%]
- Linting: [PASS/FAIL]
- Security Scan: [PASS/FAIL]
- Cyclomatic Complexity: [max value]

### Critical Issues
[List critical issues that must be fixed]

### Important Issues
[List important issues that should be addressed]

### Minor Suggestions
[List optional improvements]

### Overall Assessment
[APPROVE / REQUEST CHANGES / NEEDS DISCUSSION]

### Additional Comments
[Any other relevant observations]
```

---

## Conclusion

This code review prompt ensures comprehensive evaluation of changes against project standards. Every pull request must pass all critical checks before merging. The goal is maintaining high code quality, architectural integrity, and constitutional compliance throughout the project lifecycle.

