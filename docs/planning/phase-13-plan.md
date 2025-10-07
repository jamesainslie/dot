# Phase 13: CLI Layer - Core Commands - Implementation Plan

## Overview

Phase 13 delivers the command-line interface for dot, providing user-facing commands that wrap the Client API from Phase 12. This phase implements a thin CLI layer using Cobra, translating user input into Client operations.

**Architecture Strategy**: CLI commands are thin adapters over `dot.Client`, focusing on argument parsing, flag handling, and output rendering.

**Dependencies**: Phase 12 must be complete (Client API with Manage, Unmanage, Remanage, Adopt operations).

**Deliverable**: Working CLI with core commands (manage, unmanage, remanage, adopt) following constitutional standards.

---

## New Verb Terminology

**Standard Verbs** (replacing GNU Stow terminology):

- **manage**: Install packages by creating symlinks (was: stow)
- **unmanage**: Remove packages by deleting symlinks (was: unstow)
- **remanage**: Reinstall packages with incremental updates (was: restow)
- **adopt**: Move existing files into package then link (new: adopt)

**Rationale**: Clear, professional terminology that describes operations without Unix-specific jargon.

---

## Package Structure

```
cmd/dot/
├── main.go              # Entry point
├── main_test.go         # Entry point tests
├── root.go              # Root command
├── root_test.go         # Root command tests
├── manage.go            # Manage command
├── manage_test.go       # Manage tests
├── unmanage.go          # Unmanage command
├── unmanage_test.go     # Unmanage tests
├── remanage.go          # Remanage command
├── remanage_test.go     # Remanage tests
├── adopt.go             # Adopt command
├── adopt_test.go        # Adopt tests
├── errors.go            # Error formatting
├── errors_test.go       # Error tests
└── integration_test.go  # Integration tests
```

---

## Design Principles

- **Thin CLI Layer**: Commands delegate to `dot.Client`, minimal logic
- **Test-Driven**: Write command tests before implementation
- **User-Focused**: Clear help text, examples, error messages
- **Consistent UX**: Uniform flag naming and behavior across commands
- **Atomic Commits**: One discrete change per commit
- **Academic Tone**: Professional documentation without hyperbole

---

## Global Flags

All commands support these persistent flags:

- `-d, --dir`: Stow directory containing packages (default: ".")
- `-t, --target`: Target directory for symlinks (default: $HOME)
- `-n, --dry-run`: Show what would be done without applying changes
- `-v, --verbose`: Increase verbosity (repeatable: -v, -vv, -vvv)
- `-q, --quiet`: Suppress all non-error output
- `--log-json`: Output logs in JSON format

---

## Command-Specific Flags

### manage
- `--no-folding`: Disable directory folding optimization
- `--absolute`: Use absolute symlinks instead of relative

### unmanage
(No command-specific flags)

### remanage
(No command-specific flags)

### adopt
(No command-specific flags)

---

## Implementation Tasks

### Task 13.1: CLI Infrastructure

#### 13.1.1: Main Entry Point
- Create cmd/dot/main.go
- Add version variables for ldflags
- Write basic test

#### 13.1.2: Root Command
- Implement NewRootCommand with version info
- Add global persistent flags
- Set up command metadata
- Write version and help tests

#### 13.1.3: Configuration Builder
- Implement buildConfig() from flags
- Add path absolutization
- Create adapter instances
- Write validation tests

**Deliverable**: Working root command with global flags

---

### Task 13.2: Manage Command

#### Implementation
- Create cmd/dot/manage.go
- Add command metadata and flags
- Implement runManage function
- Integrate with dot.Client.Manage()

#### Tests
- Single package management
- Multiple packages
- Dry-run mode
- No packages (error case)
- --no-folding flag
- --absolute flag

**Deliverable**: Functional manage command

---

### Task 13.3: Unmanage Command

#### Implementation
- Create cmd/dot/unmanage.go
- Add command metadata
- Implement runUnmanage function
- Integrate with dot.Client.Unmanage()

#### Tests
- Single package unmanagement
- Multiple packages
- Dry-run mode
- Non-existent package

**Deliverable**: Functional unmanage command

---

### Task 13.4: Remanage Command

#### Implementation
- Create cmd/dot/remanage.go
- Add command metadata
- Implement runRemanage function
- Integrate with dot.Client.Remanage()

#### Tests
- Single package remanagement
- Multiple packages
- Dry-run mode
- Incremental behavior

**Deliverable**: Functional remanage command

---

### Task 13.5: Adopt Command

#### Implementation
- Create cmd/dot/adopt.go
- Add command metadata
- Implement runAdopt function
- Integrate with dot.Client.Adopt()

#### Tests
- Single file adoption
- Multiple files
- Dry-run mode
- Non-existent file

**Deliverable**: Functional adopt command

---

### Task 13.6: Error Handling

#### Implementation
- Create cmd/dot/errors.go
- Implement formatError function
- Add formatConflict helper
- Add formatMultipleErrors helper

#### Tests
- Each error type formatting
- Nested errors
- User-friendly output

**Deliverable**: User-friendly error messages

---

### Task 13.7: Integration Testing

#### Tests
- Manage + unmanage workflow
- Multiple package operations
- Remanage workflow
- Adopt workflow
- End-to-end scenarios

**Deliverable**: Comprehensive integration test suite

---

## Quality Gates

### Definition of Done

Each task is complete when:
- [ ] Implementation follows test-first approach
- [ ] All tests pass
- [ ] Test coverage ≥ 80% for command code
- [ ] All linters pass (golangci-lint)
- [ ] Help text comprehensive and accurate
- [ ] Examples demonstrate common usage
- [ ] Error messages user-friendly
- [ ] Documentation updated
- [ ] Atomic commit created

### Phase Completion Criteria

Phase 13 is complete when:
- [ ] All core commands implemented
- [ ] Root command with global flags working
- [ ] Configuration builder functional
- [ ] User-friendly error formatting in place
- [ ] Comprehensive help text for all commands
- [ ] Integration tests verify workflows
- [ ] Test coverage ≥ 80%
- [ ] All linters pass
- [ ] Commands use Client API correctly
- [ ] Dry-run mode works for all commands

---

## Development Workflow

### For Each Command

1. **Write Tests**: Create comprehensive command tests
2. **Run Tests**: Verify tests fail (red)
3. **Implement**: Write command implementation (green)
4. **Refactor**: Improve while maintaining tests
5. **Lint**: Run `make check`
6. **Document**: Add help text and examples
7. **Commit**: Create atomic commit

### Testing Commands

```bash
# Test specific command
go test ./cmd/dot -v -run TestManageCommand

# Test all CLI
go test ./cmd/dot -v

# Test with race detector
go test ./cmd/dot -race

# Build and test manually
make build
./dot --help
./dot manage --help
```

---

## Timeline Estimate

**Total Effort**: 8-12 hours

- 13.1 CLI Infrastructure: 2-3 hours
- 13.2 Manage Command: 1.5-2 hours
- 13.3 Unmanage Command: 1 hour
- 13.4 Remanage Command: 1 hour
- 13.5 Adopt Command: 1 hour
- 13.6 Error Handling: 1-1.5 hours
- 13.7 Integration Tests: 1.5-2 hours

---

## References

- [Implementation Plan](./Implementation-Plan.md)
- [Architecture Documentation](./Architecture.md)
- [Phase 12 Completion](./PHASE_12_COMPLETE.md)
- [Conventional Commits](https://www.conventionalcommits.org/)

