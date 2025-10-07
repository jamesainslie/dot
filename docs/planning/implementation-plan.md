# dot CLI Implementation Plan

## Project Overview

Implementation plan for dot v2: a modern, type-safe GNU Stow replacement written in Go 1.25.1. This plan follows constitutional principles: test-driven development, atomic commits, functional programming, and incremental delivery.

## Design Principles

- **Test-First Development**: Write tests before implementation
- **Functional Core, Imperative Shell**: Pure planning with isolated side effects
- **Type-Driven Development**: Leverage phantom types for compile-time safety
- **Incremental Delivery**: Each phase delivers working, tested functionality
- **Atomic Commits**: One discrete change per commit
- **Library First**: Core has zero CLI dependencies

## Implementation Phases

### Phase 0: Project Initialization

Foundation setup following constitutional standards.

#### 0.1: Repository Structure
- [ ] Initialize Go module with go.mod
- [ ] Create standard directory structure (cmd/, pkg/, internal/, test/)
- [ ] Set up .gitignore
- [ ] Create initial README.md
- [ ] Create CHANGELOG.md following Keep a Changelog format
- [ ] Add LICENSE file (MIT)

#### 0.2: Build Infrastructure
- [ ] Create Makefile with semantic versioning targets
- [ ] Add golangci-lint configuration (.golangci.yml)
- [ ] Configure linters: contextcheck, copyloopvar, depguard, dupl, gocritic, gocyclo, gosec, importas, misspell, nakedret, nolintlint, prealloc, revive, unconvert, whitespace
- [ ] Set cyclomatic complexity threshold to 15
- [ ] Configure depguard to prohibit github.com/pkg/errors and gotest.tools/v3

#### 0.3: CI/CD Pipeline
- [ ] Create .github/workflows/ci.yml with lint, format, vet, test, build jobs
- [ ] Set up coverage reporting with Codecov integration
- [ ] Verify 80% coverage threshold enforcement
- [ ] Create .github/workflows/release.yml with goreleaser
- [ ] Configure .goreleaser.yml for multi-platform builds

#### 0.4: Configuration Management
- [ ] Create internal/config package structure
- [ ] Implement Viper-based configuration loader
- [ ] Add XDG Base Directory Specification compliance
- [ ] Support YAML, JSON, TOML formats
- [ ] Implement configuration precedence: flags > env > file > defaults
- [ ] Write configuration tests

**Deliverable**: Working build system with CI/CD, ready for development

---

### Phase 1: Domain Model and Core Types

Pure domain model with no I/O dependencies. Enables property-based testing.

#### 1.1: Phantom-Typed Paths
- [ ] Define PathKind interface and type markers (PackageDirKind, TargetDirKind, FileDirKind)
- [ ] Implement Path[T PathKind] generic type
- [ ] Create smart constructors with validation (NewStowPath, NewTargetPath, etc.)
- [ ] Add type-safe path operations (Join, Parent, String)
- [ ] Write property tests for path operations

#### 1.2: Result Monad
- [ ] Implement Result[T] type for error handling
- [ ] Add Ok() and Err() constructors
- [ ] Implement Map() for functorial operations
- [ ] Implement FlatMap() for monadic composition
- [ ] Add Collect() for aggregating results
- [ ] Write property tests verifying monad laws

#### 1.3: Operation Type Hierarchy
- [ ] Define Operation interface with Kind(), Validate(), Dependencies(), Execute(), Rollback()
- [ ] Implement LinkCreate operation
- [ ] Implement LinkDelete operation
- [ ] Implement DirCreate operation
- [ ] Implement DirDelete operation
- [ ] Implement FileMove operation
- [ ] Implement FileBackup operation
- [ ] Write tests for each operation type

#### 1.4: Domain Value Objects
- [ ] Define Package type with metadata
- [ ] Implement FileTree with Node structure
- [ ] Add NodeType enum (File, Dir, Symlink)
- [ ] Implement functional tree operations (Map, Filter, Fold)
- [ ] Define Plan type with validation
- [ ] Create PlanMetadata structure
- [ ] Write tests for all value objects

#### 1.5: Error Taxonomy
- [ ] Define domain error types (ErrInvalidPath, ErrPackageNotFound, ErrCyclicDependency, ErrConflict)
- [ ] Define infrastructure error types (ErrFilesystemOperation, ErrPermissionDenied)
- [ ] Implement ErrMultiple for error aggregation
- [ ] Add user-facing error rendering
- [ ] Write tests for error formatting

**Deliverable**: Complete, tested domain model with zero dependencies

---

### Phase 2: Infrastructure Ports

Define interfaces for external dependencies. Enables testing with mocks.

#### 2.1: Filesystem Port
- [ ] Define FS interface with read operations (Stat, ReadDir, ReadLink, ReadFile)
- [ ] Add write operations (WriteFile, Mkdir, MkdirAll, Remove, RemoveAll, Symlink, Rename)
- [ ] Add query operations (Exists, IsDir, IsSymlink)
- [ ] Define Transaction interface for transactional operations
- [ ] Define FileInfo and DirEntry interfaces
- [ ] Write interface documentation

#### 2.2: Logger Port
- [ ] Define Logger interface with Debug, Info, Warn, Error methods
- [ ] Add With() for contextual logging
- [ ] Accept context.Context for correlation
- [ ] Write interface documentation

#### 2.3: Tracer Port
- [ ] Define Tracer interface for distributed tracing
- [ ] Define Span interface with End(), RecordError(), SetAttributes()
- [ ] Add SpanOption and Attribute types
- [ ] Write interface documentation

#### 2.4: Metrics Port
- [ ] Define Metrics interface
- [ ] Define Counter, Histogram, Gauge interfaces
- [ ] Write interface documentation

**Deliverable**: Complete port definitions for all infrastructure dependencies

---

### Phase 3: Adapters

Concrete implementations of infrastructure ports.

#### 3.1: OS Filesystem Adapter
- [ ] Implement OSFilesystem implementing FS interface
- [ ] Wrap os and filepath functions
- [ ] Add context awareness for cancellation
- [ ] Write adapter tests with real filesystem

#### 3.2: Memory Filesystem Adapter
- [ ] Integrate afero for in-memory filesystem
- [ ] Implement MemoryFilesystem implementing FS interface
- [ ] Add thread-safe operations
- [ ] Write comprehensive test suite

#### 3.3: Slog Logger Adapter
- [ ] Implement SlogLogger wrapping log/slog
- [ ] Add console-slog integration for human-readable output
- [ ] Implement With() for field accumulation
- [ ] Configure log levels and formatting
- [ ] Write adapter tests

#### 3.4: OpenTelemetry Tracer Adapter
- [ ] Implement OTelTracer wrapping trace.Tracer
- [ ] Implement OTelSpan wrapping trace.Span
- [ ] Add attribute mapping
- [ ] Write adapter tests

#### 3.5: Prometheus Metrics Adapter
- [ ] Implement PrometheusMetrics with counter, histogram, gauge maps
- [ ] Add metric registration and collection
- [ ] Implement label handling
- [ ] Write adapter tests

**Deliverable**: Working adapters for all infrastructure dependencies

---

### Phase 4: Functional Core - Scanner

Pure scanning logic with no side effects.

#### 4.1: Tree Scanning
- [ ] Implement scanTree() function for recursive directory traversal
- [ ] Add buildNode() for creating Node structures
- [ ] Implement Walk() for tree traversal
- [ ] Write property tests for tree operations

#### 4.2: Package Scanner
- [ ] Implement scanPackage() for single package
- [ ] Add parallel package scanning with goroutines
- [ ] Implement ScanInput and ScanResult types
- [ ] Create Scanner pipeline stage
- [ ] Write comprehensive scanner tests

#### 4.3: Target Directory Scanner
- [ ] Implement target directory state capture
- [ ] Add existing symlink detection
- [ ] Implement CurrentState computation
- [ ] Write tests for target scanning

#### 4.4: Dotfile Translation
- [ ] Implement dot- prefix translation logic
- [ ] Add bidirectional mapping
- [ ] Handle nested dotfile paths
- [ ] Write property tests for translation

**Deliverable**: Pure, tested scanning functions

---

### Phase 5: Ignore Pattern System

Pattern matching engine for file exclusion.

#### 5.1: Pattern Engine
- [ ] Implement Pattern type
- [ ] Add regex pattern compilation
- [ ] Implement glob-to-regex conversion
- [ ] Add case-sensitive and case-insensitive modes
- [ ] Write pattern matching tests

#### 5.2: Pattern Sources
- [ ] Define default ignore patterns (.git, .DS_Store, etc.)
- [ ] Implement IgnoreSet aggregation
- [ ] Add pattern precedence handling
- [ ] Implement negation patterns
- [ ] Write tests for pattern sources

#### 5.3: Performance Optimization
- [ ] Implement compiled pattern cache with LRU eviction
- [ ] Add early rejection optimization
- [ ] Profile pattern matching performance
- [ ] Write performance benchmarks

**Deliverable**: Fast, tested ignore pattern system

---

### Phase 6: Functional Core - Planner

Pure planning logic for computing desired state.

#### 6.1: Desired State Computation
- [ ] Implement computeDesiredState() from packages
- [ ] Add DesiredState type with Links and Dirs maps
- [ ] Implement LinkSpec and DirSpec types
- [ ] Handle directory folding logic
- [ ] Write tests for desired state computation

#### 6.2: Current State Computation
- [ ] Implement computeCurrentState() from target tree
- [ ] Add CurrentState type with Links and Files maps
- [ ] Detect existing symlinks and their targets
- [ ] Write tests for current state computation

#### 6.3: State Diffing
- [ ] Implement diffStates() to generate operations
- [ ] Add logic for link creation, deletion, updates
- [ ] Handle directory operations
- [ ] Implement PlanResult type
- [ ] Write comprehensive diff tests

#### 6.4: Incremental Planner
- [ ] Implement IncrementalPlanner with manifest integration
- [ ] Add content-based change detection using hashing
- [ ] Implement PlanRestow() for efficient updates
- [ ] Write tests for incremental planning

**Deliverable**: Pure, tested planning logic

---

### Phase 7: Functional Core - Resolver

Pure conflict detection and resolution.

#### 7.1: Conflict Detection
- [ ] Implement Conflict type with ConflictType enum
- [ ] Add conflict detection in resolveLinkCreate()
- [ ] Detect file exists, wrong link, permission, circular conflicts
- [ ] Implement ResolveResult type
- [ ] Write tests for conflict detection

#### 7.2: Resolution Policies
- [ ] Define ResolutionPolicies configuration
- [ ] Implement PolicyFail, PolicyBackup, PolicyOverwrite, PolicySkip
- [ ] Add per-conflict resolution strategies
- [ ] Implement resolveOperation() dispatcher
- [ ] Write tests for each policy

#### 7.3: Warning and Suggestion System
- [ ] Define Warning type
- [ ] Generate actionable suggestions for conflicts
- [ ] Implement suggestion templates
- [ ] Write tests for suggestions

**Deliverable**: Pure, tested conflict resolution logic

---

### Phase 8: Functional Core - Topological Sorter

Dependency graph and operation ordering.

#### 8.1: Dependency Graph
- [ ] Implement DependencyGraph with nodes and edges
- [ ] Add BuildGraph() from operations
- [ ] Implement operation dependency detection
- [ ] Write tests for graph construction

#### 8.2: Topological Sort
- [ ] Implement TopologicalSort() with DFS
- [ ] Add cycle detection with FindCycle()
- [ ] Handle cyclic dependency errors
- [ ] Write tests for sorting

#### 8.3: Parallelization Analysis
- [ ] Implement ParallelizationPlan() for batch computation
- [ ] Add level-based grouping
- [ ] Compute independent operation batches
- [ ] Write tests for parallelization

**Deliverable**: Pure, tested sorting and parallelization logic

---

### Phase 9: Pipeline Orchestration

Compose functional core stages into pipelines.

#### 9.1: Pipeline Types
- [ ] Define Pipeline[A, B] generic function type
- [ ] Implement Compose() for pipeline composition
- [ ] Add Parallel() for concurrent pipeline execution
- [ ] Write tests for pipeline composition

#### 9.2: Core Pipeline
- [ ] Implement StowPipeline composing scan, plan, resolve, order
- [ ] Add UnstowPipeline
- [ ] Implement RestowPipeline with incremental planner
- [ ] Add AdoptPipeline
- [ ] Write tests for each pipeline

#### 9.3: Pipeline Engine
- [ ] Implement pipeline.Engine for orchestration
- [ ] Add pipeline execution with context propagation
- [ ] Implement error handling and propagation
- [ ] Write integration tests

**Deliverable**: Working pipeline engine composing all functional stages

---

### Phase 10: Imperative Shell - Executor

Side-effecting execution with transactions and rollback.

#### 10.1: Basic Execution
- [ ] Implement Executor with FS, Logger, Tracer dependencies
- [ ] Add sequential operation execution
- [ ] Implement operation.Execute() calls with error handling
- [ ] Write tests with memory filesystem

#### 10.2: Two-Phase Commit
- [ ] Implement prepare() phase for validation
- [ ] Add precondition checking (permissions, space, etc.)
- [ ] Implement commit() phase for execution
- [ ] Add checkpoint creation before modifications
- [ ] Write tests for two-phase commit

#### 10.3: Rollback Mechanism
- [ ] Implement CheckpointStore interface
- [ ] Add FSCheckpointStore implementation
- [ ] Implement rollback() with reverse operation execution
- [ ] Add checkpoint cleanup on success
- [ ] Write comprehensive rollback tests

#### 10.4: Parallel Execution
- [ ] Implement executeParallel() using parallelization plan
- [ ] Add executeBatch() with goroutines and sync
- [ ] Handle concurrent errors safely with mutex
- [ ] Write tests for parallel execution

#### 10.5: Instrumentation
- [ ] Add tracing spans for operations
- [ ] Implement metrics collection (counters, histograms, gauges)
- [ ] Add structured logging throughout execution
- [ ] Write observability tests

**Deliverable**: Robust, tested executor with transactions and observability

---

### Phase 11: Manifest and State Management

Persistent state tracking for incremental operations.

#### 11.1: Manifest Types
- [ ] Define Manifest type with version, packages, hashes
- [ ] Implement PackageInfo with installation metadata
- [ ] Add ManifestStore interface
- [ ] Write manifest tests

#### 11.2: Manifest Persistence
- [ ] Implement FSManifestStore with JSON serialization
- [ ] Add Load() and Save() operations
- [ ] Store manifest in target directory (.dot-manifest.json)
- [ ] Handle missing manifest gracefully
- [ ] Write persistence tests

#### 11.3: Content Hashing
- [ ] Implement ContentHasher for package hashing
- [ ] Add fast hash computation (xxhash or similar)
- [ ] Integrate with IncrementalPlanner
- [ ] Write hashing tests

#### 11.4: State Validation
- [ ] Implement manifest consistency validation
- [ ] Add drift detection between manifest and filesystem
- [ ] Implement repair from filesystem
- [ ] Write validation tests

**Deliverable**: Working state management system

---

### Phase 12: Public Library API (Interface Pattern)

Clean Go API for embedding in other tools using interface-based pattern to avoid import cycles.

**Architecture**: Client interface in `pkg/dot/`, implementation in `internal/api/`

#### 12.1: Client Interface
- [ ] Define Client interface with all operations in pkg/dot/
- [ ] Add registration mechanism for implementation
- [ ] Write interface tests

#### 12.2: Client Implementation
- [ ] Implement client struct in internal/api/
- [ ] Add Manage(), Unmanage(), Remanage(), Adopt() methods
- [ ] Implement PlanManage(), PlanUnmanage(), etc. for dry-run
- [ ] Add Status(), List() query methods
- [ ] Write implementation tests

#### 12.3: Supporting Types
- [ ] Define Status and PackageInfo types
- [ ] Move ExecutionResult and Checkpoint to pkg/dot
- [ ] Write type tests

#### 12.4: Documentation
- [ ] Add comprehensive package documentation
- [ ] Create example tests for godoc
- [ ] Document interface pattern rationale

**Deliverable**: Clean, tested public library API (interface-based)

**See Also**: [Phase 12 Detailed Plan](./Phase-12-Plan.md)

---

### Phase 12b: Domain Architecture Refactoring (Optional Future Work)

Refactor from interface pattern (Phase 12) to direct Client struct by moving domain types to `internal/domain/`.

**Architecture**: Domain in `internal/domain/`, Client struct in `pkg/dot/`

**Prerequisite**: Phase 12 stable and in production use

#### 12b.1: Domain Type Migration
- [ ] Create internal/domain/ package
- [ ] Move all domain types from pkg/dot/ to internal/domain/
- [ ] Create re-exports in pkg/dot/
- [ ] Atomic migration per type with testing

#### 12b.2: Internal Package Updates
- [ ] Update all internal/* imports from pkg/dot to internal/domain
- [ ] Test each package after update
- [ ] Verify no breaking changes

#### 12b.3: Client Simplification
- [ ] Replace Client interface with concrete struct
- [ ] Move implementation from internal/api to pkg/dot
- [ ] Remove registration mechanism
- [ ] Delete internal/api/ package

#### 12b.4: Validation
- [ ] Verify API compatibility
- [ ] Run full test suite
- [ ] Benchmark comparison
- [ ] Update documentation

**Deliverable**: Cleaner architecture with direct Client struct (no interface indirection)

**Effort**: 12-16 hours of careful refactoring

**See Also**: [Phase 12b Refactoring Plan](./Phase-12b-Refactor-Plan.md)

---

### Phase 13: CLI Layer - Core Commands

Cobra-based CLI with core operations using new verb terminology.

**New Verbs**: manage (was stow), unmanage (was unstow), remanage (was restow), adopt (new)

#### 13.1: CLI Infrastructure
- [ ] Implement cmd/dot/main.go entry point with version support
- [ ] Create root command with global flags (dir, target, dry-run, verbose, quiet, log-json)
- [ ] Add configuration builder from flags
- [ ] Implement logger creation with format selection
- [ ] Add path absolutization and validation
- [ ] Write infrastructure tests

#### 13.2: Manage Command
- [ ] Implement manage command (replaces stow)
- [ ] Add flags: --no-folding, --absolute
- [ ] Integrate with dot.Client.Manage()
- [ ] Add dry-run support
- [ ] Write comprehensive command tests
- [ ] Add help text with examples

#### 13.3: Unmanage Command
- [ ] Implement unmanage command (replaces unstow)
- [ ] Add package argument parsing
- [ ] Integrate with dot.Client.Unmanage()
- [ ] Add dry-run support
- [ ] Write command tests
- [ ] Add help text with examples

#### 13.4: Remanage Command
- [ ] Implement remanage command (replaces restow)
- [ ] Use incremental planner via Client
- [ ] Integrate with dot.Client.Remanage()
- [ ] Add dry-run support
- [ ] Write command tests
- [ ] Add help text with examples

#### 13.5: Adopt Command
- [ ] Implement adopt command with package and file arguments
- [ ] Integrate with dot.Client.Adopt()
- [ ] Add dry-run support
- [ ] Write command tests
- [ ] Add help text with examples

#### 13.6: Error Handling and UX
- [ ] Implement user-friendly error formatting
- [ ] Add formatError helper for domain errors
- [ ] Add formatConflict for conflict errors
- [ ] Add formatMultipleErrors for aggregated errors
- [ ] Write error formatting tests

#### 13.7: Integration Testing
- [ ] Write manage + unmanage workflow test
- [ ] Write multiple package test
- [ ] Write remanage workflow test
- [ ] Write adopt workflow test
- [ ] Verify end-to-end scenarios

**Deliverable**: Working CLI with core commands using Client API

**See Also**: [Phase 13 Detailed Plan](./Phase-13-Plan.md)

---

### Phase 14: CLI Layer - Query Commands

Status, diagnostic, and listing commands with multiple output formats.

#### 14.1: Output Renderer Infrastructure
- [ ] Define Renderer interface for multiple formats
- [ ] Implement renderer factory
- [ ] Create color scheme definitions
- [ ] Implement base renderer utilities
- [ ] Write renderer infrastructure tests

#### 14.2: Text Renderer
- [ ] Implement TextRenderer structure
- [ ] Implement status rendering with colorization
- [ ] Implement diagnostics rendering with suggestions
- [ ] Implement package list rendering
- [ ] Add text formatting helpers
- [ ] Write text renderer tests

#### 14.3: JSON Renderer
- [ ] Implement JSONRenderer structure
- [ ] Implement status rendering to JSON
- [ ] Implement diagnostics rendering to JSON
- [ ] Implement package list rendering to JSON
- [ ] Add JSON utilities and validation
- [ ] Write JSON renderer tests

#### 14.4: YAML Renderer
- [ ] Implement YAMLRenderer structure
- [ ] Implement status rendering to YAML
- [ ] Implement diagnostics rendering to YAML
- [ ] Implement package list rendering to YAML
- [ ] Add YAML struct tags
- [ ] Write YAML renderer tests

#### 14.5: Table Renderer
- [ ] Implement TableRenderer structure with lipgloss
- [ ] Create table builder with dynamic column widths
- [ ] Implement status rendering as table
- [ ] Implement diagnostics rendering as table
- [ ] Implement package list rendering as table
- [ ] Add table utilities
- [ ] Write table renderer tests

#### 14.6: Status Command
- [ ] Create status command structure
- [ ] Implement command handler with format selection
- [ ] Add status query logic in Client API
- [ ] Define Status and PackageStatus types
- [ ] Add help text with examples
- [ ] Write status command tests

#### 14.7: Doctor Command
- [ ] Create doctor command structure
- [ ] Implement command handler with exit codes
- [ ] Add doctor logic in Client API
- [ ] Define DiagnosticReport and Issue types
- [ ] Implement issue detection (broken links, orphaned, permissions)
- [ ] Generate actionable suggestions
- [ ] Add help text with examples
- [ ] Write doctor command tests

#### 14.8: List Command
- [ ] Create list command structure
- [ ] Implement command handler with sorting
- [ ] Add list logic in Client API
- [ ] Implement package sorting by multiple fields
- [ ] Add help text with examples
- [ ] Write list command tests

#### 14.9: Integration and Polish
- [ ] Ensure command consistency across all query commands
- [ ] Standardize error handling
- [ ] Verify output consistency across all formats
- [ ] Performance profiling and optimization
- [ ] Add documentation and examples
- [ ] Write integration tests

**Deliverable**: Complete query command suite with rich output formatting

**See Also**: [Phase 14 Detailed Plan](./Phase-14-Plan.md)

---

### Phase 15: Error Handling and User Experience

User-friendly error messages and help system.

#### 15.1: Error Formatting
- [ ] Implement RenderUserError() for friendly messages
- [ ] Add conflict formatting with suggestions
- [ ] Create error templates
- [ ] Remove technical jargon from user errors
- [ ] Write error formatting tests

#### 15.2: Help and Documentation
- [ ] Add detailed help text to all commands
- [ ] Include usage examples in help
- [ ] Generate man pages
- [ ] Add shell completion (bash, zsh, fish)
- [ ] Write documentation

#### 15.3: Progress Indicators
- [ ] Add progress bars for long operations
- [ ] Implement spinner for indeterminate operations
- [ ] Add percentage completion where deterministic
- [ ] Respect quiet mode
- [ ] Write progress tests

**Deliverable**: Polished user experience

---

### Phase 16: Property-Based Testing

Verify algebraic laws, invariants, and mathematical properties using gopter.

#### 16.1: Test Infrastructure Setup
- [ ] Integrate gopter framework with project conventions
- [ ] Configure test parameters and CI integration
- [ ] Create test/properties/ package structure
- [ ] Set up test harness and runners

#### 16.2: Data Generators
- [ ] Implement path generators (absolute, relative, typed)
- [ ] Implement package structure generators
- [ ] Implement operation generators with dependencies
- [ ] Implement filesystem state generators
- [ ] Create generator composition utilities

#### 16.3: Algebraic Law Verification
- [ ] Test idempotence (operations can repeat safely)
- [ ] Test reversibility (operations can be undone)
- [ ] Test commutativity (order independence)
- [ ] Test associativity (grouping independence)
- [ ] Test conservation (data preservation)

#### 16.4: Domain Invariant Verification
- [ ] Verify path invariants (type safety, containment)
- [ ] Verify graph invariants (acyclicity, reachability)
- [ ] Verify manifest invariants (consistency, completeness)
- [ ] Verify operation invariants (validity, atomicity)
- [ ] Verify conflict invariants (detection completeness)

#### 16.5: Performance Properties
- [ ] Test algorithmic complexity bounds
- [ ] Verify incremental operation performance
- [ ] Test parallelization correctness and speedup

#### 16.6: Error Handling Properties
- [ ] Test error propagation completeness
- [ ] Verify rollback correctness
- [ ] Test validation exhaustiveness

#### 16.7: Integration and Documentation
- [ ] Integrate property tests into CI/CD
- [ ] Write comprehensive property testing guide
- [ ] Create example property tests
- [ ] Document maintenance procedures

**Deliverable**: Comprehensive property-based test suite verifying mathematical correctness

**See Also**: [Phase 16 Detailed Plan](./Phase-16-Plan.md)

---

### Phase 17: Integration Testing

End-to-end scenario testing to verify complete workflows, concurrent operations, error recovery, and cross-platform compatibility.

#### 17.1: Test Infrastructure Setup
- [ ] Create test package structure (tests/integration/, tests/fixtures/)
- [ ] Implement fixture builder framework (PackageBuilder, FileTreeBuilder)
- [ ] Create test utilities (TestFS, assertion helpers)
- [ ] Implement golden test framework
- [ ] Build test harness with environment isolation
- [ ] Write infrastructure tests

#### 17.2: End-to-End Workflow Tests
- [ ] Test complete manage workflow with all options
- [ ] Test complete unmanage workflow with cleanup
- [ ] Test remanage with incremental detection
- [ ] Test adopt workflow with file preservation
- [ ] Test combined workflows (manage + unmanage identity)
- [ ] Write comprehensive workflow tests

#### 17.3: Concurrent Testing
- [ ] Test parallel package processing and batches
- [ ] Enable race detector and fix races
- [ ] Test concurrent operations with isolation
- [ ] Perform stress testing with high load
- [ ] Write concurrency tests

#### 17.4: Error Recovery Testing
- [ ] Test transaction rollback on failures
- [ ] Test checkpoint recovery mechanisms
- [ ] Test error propagation and aggregation
- [ ] Test partial failure handling
- [ ] Test pre-execution validation
- [ ] Write recovery tests

#### 17.5: Conflict Resolution Scenarios
- [ ] Test all conflict type detection
- [ ] Test all resolution policies (fail, backup, overwrite, skip)
- [ ] Test conflict resolution integration
- [ ] Write conflict scenario tests

#### 17.6: State Management Integration
- [ ] Test manifest persistence and updates
- [ ] Test incremental detection with hashing
- [ ] Test state validation and drift detection
- [ ] Write state management tests

#### 17.7: Query Command Integration
- [ ] Test status command with all formats
- [ ] Test doctor command diagnostics
- [ ] Test list command sorting and filtering
- [ ] Write query command tests

#### 17.8: Cross-Platform Testing
- [ ] Create platform-specific test suites
- [ ] Test filesystem compatibility
- [ ] Test path convention handling
- [ ] Write cross-platform tests

#### 17.9: Performance Regression Testing
- [ ] Create benchmark suite for key operations
- [ ] Implement performance monitoring with thresholds
- [ ] Profile memory usage and leaks
- [ ] Write performance tests

#### 17.10: CLI Integration Testing
- [ ] Test all command invocations with flags
- [ ] Test flag interactions and precedence
- [ ] Test exit codes match specification
- [ ] Test output formats with golden tests
- [ ] Write CLI integration tests

#### 17.11: Scenario-Based Testing
- [ ] Test new user scenario
- [ ] Test GNU Stow migration scenario
- [ ] Test multi-machine scenario
- [ ] Test development workflow scenario
- [ ] Test large repository scenario
- [ ] Write scenario tests

#### 17.12: Test Organization and Maintenance
- [ ] Write integration testing guide and documentation
- [ ] Implement test categorization and filtering
- [ ] Integrate tests into CI pipeline
- [ ] Create test maintenance utilities
- [ ] Write maintenance documentation

**Deliverable**: Comprehensive integration test suite with cross-platform verification, performance baselines, and CI integration

**See Also**: [Phase 17 Detailed Plan](./Phase-17-Plan.md)

---

### Phase 18: Performance Optimization

Profile and optimize critical paths.

#### 18.1: Profiling
- [ ] Add CPU profiling support
- [ ] Add memory profiling support
- [ ] Profile scanner performance
- [ ] Profile planner performance
- [ ] Profile executor performance
- [ ] Write benchmarks

#### 18.2: Optimization
- [ ] Optimize hot paths identified in profiling
- [ ] Add caching where beneficial
- [ ] Tune concurrency parameters
- [ ] Optimize memory allocations
- [ ] Write performance tests

#### 18.3: Streaming Optimization
- [ ] Implement streaming scanner for large trees
- [ ] Add backpressure handling
- [ ] Optimize channel buffer sizes
- [ ] Write streaming benchmarks

**Deliverable**: Optimized performance with benchmarks

---

### Phase 19: Documentation

Comprehensive documentation for users and developers.

#### 19.1: User Documentation
- [ ] Write comprehensive README.md
- [ ] Create user guide in docs/
- [ ] Write quickstart tutorial
- [ ] Add GNU Stow migration guide
- [ ] Write troubleshooting guide
- [ ] Add configuration reference

#### 19.2: Developer Documentation
- [ ] Write architecture decision records (ADRs)
- [ ] Document design patterns used
- [ ] Add API reference documentation
- [ ] Write contributing guide
- [ ] Document testing strategy

#### 19.3: Examples
- [ ] Create examples/ directory
- [ ] Add basic usage examples
- [ ] Add advanced usage examples
- [ ] Add library embedding examples
- [ ] Add configuration examples

**Deliverable**: Complete documentation suite

---

### Phase 20: Polish and Release Preparation

Final polish and release preparation for production v0.1.0 release.

#### 20.1: Code Quality Assurance
- [ ] Run complete linter suite and fix all warnings
- [ ] Verify 80% test coverage across all packages
- [ ] Run property tests with high iteration count (10,000+)
- [ ] Perform comprehensive security audit
- [ ] Document code quality metrics

#### 20.2: Release Artifacts and Automation
- [ ] Test cross-compilation for all platforms (12+ targets)
- [ ] Verify and configure goreleaser for automated releases
- [ ] Test release process locally with snapshots
- [ ] Prepare comprehensive CHANGELOG.md for v0.1.0
- [ ] Tag and create pre-release (v0.1.0-rc.1) for validation

#### 20.3: Distribution Infrastructure
- [ ] Create Homebrew formula with multi-platform support
- [ ] Create Scoop manifest for Windows installation
- [ ] Test all installation methods on target platforms
- [ ] Write comprehensive installation documentation
- [ ] Validate distribution channels

#### 20.4: Final Cross-Platform Validation
- [ ] Test on Linux distributions (Ubuntu, Debian, Fedora, Arch, Alpine)
- [ ] Test on macOS versions (13, 14, 15) for Intel and Apple Silicon
- [ ] Test on Windows (10, 11, Server 2022)
- [ ] Test on BSD systems (FreeBSD, OpenBSD, NetBSD)
- [ ] Execute comprehensive integration test matrix
- [ ] Document platform-specific behaviors and limitations

#### 20.5: Documentation Finalization
- [ ] Polish README.md with badges, examples, and complete installation guide
- [ ] Complete user guide with all features documented
- [ ] Finalize developer documentation and architecture docs
- [ ] Create working examples and tutorials
- [ ] Verify all documentation links and accuracy

#### 20.6: Release Execution
- [ ] Complete pre-release checklist
- [ ] Create and push v0.1.0 tag with detailed release notes
- [ ] Publish release on GitHub with all artifacts
- [ ] Update all distribution channels
- [ ] Post release announcements
- [ ] Establish post-release monitoring

**Deliverable**: Production-ready v0.1.0 release with complete distribution infrastructure

**See Also**: [Phase 20 Detailed Plan](./Phase-20-Plan.md)

---

## Development Workflow

### Test-Driven Development Cycle

For each feature:

1. **Write Test**: Create failing test(s) describing desired behavior
2. **Run Test**: Verify test fails for the right reason (red)
3. **Implement**: Write minimum code to pass test
4. **Run Test**: Verify test passes (green)
5. **Refactor**: Improve code while maintaining passing tests
6. **Commit**: Make atomic commit with conventional commit message

### Commit Message Format

```
<type>(scope): <description>

[optional body]

[optional footer]
```

Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build

### Release Process

1. Update CHANGELOG.md with version and changes
2. Run `make version-{major|minor|patch}`
3. Verify all tests pass
4. Verify all linters pass
5. Create annotated git tag
6. Push tag to trigger release automation
7. Verify goreleaser artifacts
8. Update documentation

## Success Criteria

### Phase Completion Criteria

Each phase is complete when:
- [ ] All functionality implemented and tested
- [ ] Test coverage ≥ 80% for new code
- [ ] All linters pass without warnings
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Changes committed atomically
- [ ] Integration tests pass

### Project Completion Criteria

Project v0.1.0 is complete when:
- [ ] All core features implemented (stow, unstow, restow, adopt)
- [ ] All query commands implemented (status, doctor, list)
- [ ] Property-based tests verify algebraic laws
- [ ] Integration tests cover major scenarios
- [ ] Test coverage ≥ 80% overall
- [ ] All linters pass
- [ ] Documentation complete
- [ ] Cross-platform builds successful
- [ ] Release automation working
- [ ] GNU Stow feature parity achieved

## Risk Mitigation

### Technical Risks

1. **Phantom Types Complexity**: Mitigate by thorough testing and documentation
2. **Performance**: Profile early and often, optimize incrementally
3. **Cross-Platform**: Test on all platforms throughout development
4. **Concurrent Bugs**: Use race detector, extensive concurrent testing

### Process Risks

1. **Scope Creep**: Strict adherence to phases, defer advanced features to v0.2.0
2. **Quality Debt**: No skipping tests, maintain 80% coverage, fix linter issues immediately
3. **Integration Issues**: Continuous integration testing, frequent integration

## Post-v0.1.0 Roadmap

Future enhancements for v0.2.0 and beyond:

- Interactive TUI mode using bubbletea
- Remote package support (Git repositories)
- Package registries and discovery
- Diff and merge capabilities
- Configuration profiles
- Monitoring dashboard
- Webhooks and event system
- Template support with variable substitution
- Multi-target support
- Package groups and dependencies
