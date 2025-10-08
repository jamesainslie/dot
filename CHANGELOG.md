# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Bug Fixes
- **client:** properly propagate manifest errors in Doctor
- **client:** populate packages from plan when empty in updateManifest
- **config:** add missing KeyDoctorCheckPermissions constant
- **config:** honor MarshalOptions.Indent in TOML strategy
- **doctor:** detect and report permission errors on link targets
- **domain:** make path validator tests OS-aware for Windows
- **hooks:** show linting output in pre-commit hook
- **hooks:** show test output in pre-commit hook
- **hooks:** check overall project coverage to match CI
- **manage:** implement proper unmanage in planFullRemanage
- **manifest:** propagate non-not-found errors in Update
- **path:** add method forwarding to Path wrapper type
- **status:** propagate non-not-found manifest errors
- **test:** rename ExecutionFailure test to match actual behavior
- **test:** strengthen PackageOperations assertion in exhaustive test
- **test:** skip file mode test on Windows
- **test:** correct comment in PlanOperationsEmpty test
- **test:** add proper error handling to CLI integration tests
- **test:** add Windows build constraints to Unix-specific tests
- **test:** add Windows compatibility to testutil symlink tests
- **test:** correct mock variadic parameter handling in ports_test

### Code Refactoring
- **api:** replace Client interface with concrete struct
- **config:** migrate writer to use strategy pattern
- **config:** use permission constants
- **domain:** clean up temporary migration scripts
- **domain:** complete internal package migration and simplify pkg/dot
- **domain:** create internal/domain package structure
- **domain:** move Result monad to internal/domain
- **domain:** move Path and errors types to internal/domain
- **domain:** use validators in path constructors
- **domain:** move MustParsePath to testing.go
- **domain:** improve TraversalFreeValidator implementation
- **domain:** move all remaining domain types to internal/domain
- **domain:** update all internal package imports to use internal/domain
- **domain:** use TargetPath for operation targets
- **domain:** format code and fix linter issues
- **hooks:** eliminate duplicate test run in pre-commit
- **path:** remove Path generic wrapper to eliminate code smell
- **pkg:** simplify scanForOrphanedLinks method
- **pkg:** convert Client to facade pattern
- **pkg:** extract helper methods in DoctorService
- **pkg:** simplify DoctorWithScan method
- **test:** improve benchmark tests with proper error handling

### Features
- **config:** implement TOML marshal strategy
- **config:** implement JSON marshal strategy
- **config:** implement YAML marshal strategy
- **domain:** add chainable Result methods
- **pkg:** extract DoctorService from Client
- **pkg:** extract AdoptService from Client
- **pkg:** extract StatusService from Client
- **pkg:** extract UnmanageService from Client
- **pkg:** extract ManageService from Client
- **pkg:** extract ManifestService from Client


## v0.1.0 - 2025-10-07
### Bug Fixes
- **api:** address CodeRabbit feedback on Phase 12
- **api:** use configured skip patterns in recursive orphan scanning
- **api:** improve error handling and test robustness
- **api:** use package-operation mapping for accurate manifest tracking
- **api:** normalize paths for cross-platform link lookup
- **api:** enforce depth and context limits in recursive orphan scanning
- **cli:** resolve critical bugs in progress, config, and rendering
- **cli:** handle both pointer and value operation types in renderers
- **cli:** improve config format detection and help text indentation
- **cli:** correct scan flag variable scope in NewDoctorCommand
- **cli:** add error templates for checkpoint and not implemented errors
- **cli:** respect NO_COLOR environment variable in shouldColorize
- **cli:** improve JSON/YAML output and doctor performance
- **cli:** render execution plan in dry-run mode
- **cli:** improve TTY detection portability and path truncation
- **config:** enable CodeRabbit auto-review for all pull requests
- **executor:** make Checkpoint operations map thread-safe
- **executor:** address code review feedback for concurrent safety and error handling
- **manifest:** add security guards and prevent hash collisions
- **pipeline:** prevent shared mutation of context maps in metadata conversion
- **release:** separate archive configs for Homebrew compatibility
- **scanner:** implement real package tree scanning with ignore filtering
- **test:** improve test isolation and cross-platform compatibility
- **test:** make Adopt execution error test deterministic

### Build System
- **make:** add buildvcs flag for reproducible builds
- **makefile:** add build infrastructure with semantic versioning
- **release:** add Homebrew tap integration

### Code Refactoring
- **adopt:** update Adopt and PlanAdopt methods to use files-first signature
- **api:** reduce cyclomatic complexity in PlanRemanage
- **api:** extract orphan scan logic to reduce complexity
- **cli:** address code review nitpicks for improved code quality
- **cli:** reduce cyclomatic complexity in table renderer
- **cli:** add default case and eliminate type assertion duplication
- **pipeline:** use safe unwrap pattern in path construction tests
- **pipeline:** improve test quality and organization
- **quality:** improve error handling documentation and panic messages
- **terminology:** update suggestion text from unstow to unmanage
- **terminology:** replace stow with package directory terminology
- **terminology:** complete stow removal from test fixtures
- **terminology:** rename stow-prefixed variables to package/manage

### Features
- **adapters:** implement slog logger and no-op adapters
- **adapters:** implement OS filesystem adapter
- **api:** implement Unmanage, Remanage, and Adopt operations
- **api:** add foundational types for Phase 12 Client API
- **api:** define Client interface for public API
- **api:** implement Client with Manage operation
- **api:** add comprehensive tests and documentation
- **api:** implement directory extraction and link set optimization
- **api:** update Doctor API to accept ScanConfig parameter
- **api:** update Doctor API to accept ScanConfig parameter
- **api:** implement incremental remanage with hash-based change detection
- **api:** add depth calculation and directory skip logic
- **api:** implement link count extraction from plan
- **api:** add DoctorWithScan for explicit scan configuration
- **api:** wire up orphaned link detection with safety limits
- **cli:** implement list command for package inventory
- **cli:** add scan control flags to doctor command
- **cli:** implement help system with examples and completion
- **cli:** implement progress indicators for operation feedback
- **cli:** implement terminal styling and layout system
- **cli:** implement output renderer infrastructure
- **cli:** add minimal CLI entry point for build validation
- **cli:** add config command for XDG configuration management
- **cli:** implement error formatting foundation for Phase 15
- **cli:** implement status command for installation state inspection
- **cli:** implement doctor command for health checks
- **cli:** implement Phase 13 CLI infrastructure with core commands
- **cli:** implement UX polish with output formatting
- **cli:** implement command handlers for manage, unmanage, remanage, adopt
- **cli:** show complete operation breakdown in table summary
- **config:** wire backup directory through system
- **config:** implement extended configuration infrastructure
- **config:** implement configuration management with Viper and XDG compliance
- **config:** add Config struct with validation
- **domain:** implement operation type hierarchy
- **domain:** implement error taxonomy with user-facing messages
- **domain:** implement Result monad for functional error handling
- **domain:** implement phantom-typed paths for compile-time safety
- **domain:** implement domain value objects
- **domain:** add package-operation mapping to Plan
- **dot:** add ScanConfig types for orphaned link detection
- **executor:** add metrics instrumentation wrapper
- **executor:** implement parallel batch execution
- **executor:** implement Phase 10 executor with two-phase commit
- **ignore:** implement pattern matching engine and ignore sets
- **manifest:** implement FSManifestStore persistence
- **manifest:** add core manifest domain types
- **manifest:** implement content hashing for packages
- **manifest:** define ManifestStore interface
- **manifest:** implement manifest validation
- **operation:** add Execute and Rollback methods to operations
- **pipeline:** track package ownership in operation plans
- **pipeline:** surface conflicts and warnings in plan metadata
- **pipeline:** enhance context cancellation handling in pipeline stages
- **pipeline:** implement stow pipeline with scanning, planning, resolution, and sorting stages
- **planner:** implement suggestion generation and conflict enrichment
- **planner:** implement conflict detection for links and directories
- **planner:** define conflict type enumeration
- **planner:** implement real desired state computation
- **planner:** implement desired state computation foundation
- **planner:** define resolution status types
- **planner:** implement resolve result type
- **planner:** implement conflict value object
- **planner:** implement resolution policy types and basic policies
- **planner:** implement main resolver function and policy dispatcher
- **planner:** integrate resolver with planning pipeline
- **planner:** implement dependency graph construction
- **planner:** implement parallelization analysis
- **planner:** implement topological sort with cycle detection
- **ports:** define infrastructure port interfaces
- **scanner:** implement tree scanning with recursive traversal
- **scanner:** implement dotfile translation logic
- **scanner:** implement package scanner with ignore support
- **types:** add Status and PackageInfo types


[Unreleased]: https://github.com/jamesainslie/dot/compare/v0.1.0...HEAD
