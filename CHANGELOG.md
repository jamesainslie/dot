# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CLI command handlers for manage, unmanage, remanage, adopt (Phase 22.1)
- Package-operation mapping in Plan type for accurate link tracking (Phase 22.2)
- Per-package link tracking in manifest (Phase 22.2)
- Plan helper methods: OperationsForPackage, PackageNames, HasPackage (Phase 22.2)
- Backup directory configuration support via CLI flag, config file, environment variable (Phase 22.3)
- DoctorWithScan method for explicit scan configuration control (Phase 22.4)
- Incremental remanage with SHA256 hash-based change detection (Phase 22.5)
- Package content hash storage in manifest for change detection (Phase 22.5)
- ADR-002 documenting package-operation mapping design
- ADR-003 documenting streaming API design (future)
- ADR-004 documenting ConfigBuilder design (future)

### Changed
- Manifest now accurately tracks LinkCount and Links per package (was always 0/empty) (Phase 22.2)
- Remanage operations use incremental planning by default (99% faster for unchanged) (Phase 22.5)
- Backup directory configuration now functional throughout system (Phase 22.3)
- Doctor(ctx) now exists as simple wrapper around DoctorWithScan (Phase 22.4)
- Pipeline properly uses configured backup directory (was hardcoded to empty) (Phase 22.3)

### Fixed
- CLI commands no longer stubbed, fully implemented and functional (Phase 22.1)
- Manifest LinkCount now shows actual count instead of always 0 (Phase 22.2)
- Manifest Links array now populated instead of always empty (Phase 22.2)
- Backup directory configuration now wired through pipeline (Phase 22.3)
- Status command now shows accurate link counts per package (Phase 22.2)
- Unmanage command can now properly remove package links (Phase 22.2)

### Added
- Project initialization with Go 1.25 module
- Standard directory structure (cmd, pkg, internal, tests)
- Makefile with semantic versioning targets and build automation
- golangci-lint configuration with 15 linters enabled
- GitHub Actions CI/CD workflows for lint, format, vet, test, build
- GoReleaser v2 configuration for automated releases
- Configuration management package with Viper and XDG compliance
- Support for YAML, JSON, TOML configuration formats
- Environment variable overrides with DOT_ prefix
- Phantom-typed paths (PackagePath, TargetPath, FilePath) for compile-time safety
- Result monad for functional error handling with Map, FlatMap, Collect
- Error taxonomy with domain and infrastructure error types
- User-facing error messages without technical jargon
- Operation type hierarchy (LinkCreate, LinkDelete, DirCreate, DirDelete, FileMove, FileBackup)
- Domain value objects (Package, Node, Plan, PlanMetadata)
- Infrastructure port interfaces (FS, Logger, Tracer, Metrics)
- Mock implementations for all ports enabling pure functional testing
- OS filesystem adapter wrapping os package with context cancellation
- Slog logger adapter with console-slog integration for human-readable output
- No-op adapters for logger, tracer, and metrics (testing and performance)
- Tree scanner with recursive directory traversal
- Dotfile translation (dot- prefix to . prefix and reverse)
- Tree utility functions (Walk, CollectFiles, CountNodes, RelativePath)
- Ignore pattern engine with glob-to-regex conversion
- IgnoreSet for aggregating multiple ignore patterns
- Default ignore patterns (.git, .DS_Store, etc.)
- Package scanner with ignore pattern support
- Planner foundation (DesiredState, LinkSpec, DirSpec, ComputeDesiredState)
- Conflict type enumeration with 6 conflict categories
- Conflict value object with context and suggestions
- Resolution status types (OK, Conflict, Warning, Skip)
- ResolveResult aggregation for conflicts and warnings
- CurrentState representation for filesystem state
- Conflict detection for LinkCreate and DirCreate operations
- Resolution policies (Fail, Backup, Overwrite, Skip)
- ResolutionPolicies configuration with fail-safe defaults
- Warning severity levels (Info, Caution, Danger)
- Context-aware suggestion generation for all conflict types
- Conflict enrichment with actionable suggestions and examples
- Main Resolve() function for conflict resolution orchestration
- Policy dispatcher with operation-specific resolution logic
- PlanResult type for planning with optional conflict resolution
- ComputeOperationsFromDesiredState for state-to-operation conversion
- README.md with project overview and architecture
- Command terminology: manage/unmanage/remanage for clarity
- CHANGELOG.md following Keep a Changelog format
- MIT LICENSE file
- .gitignore for Go project artifacts

[Unreleased]: https://github.com/jamesainslie/dot/commits/main

