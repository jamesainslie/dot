# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- README.md with project overview and architecture
- Command terminology: manage/unmanage/remanage for clarity
- CHANGELOG.md following Keep a Changelog format
- MIT LICENSE file
- .gitignore for Go project artifacts

[Unreleased]: https://github.com/jamesainslie/dot/commits/main

