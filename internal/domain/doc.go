// Package domain contains the core domain types and port interfaces for dot.
//
// This package defines the essential types used throughout the system:
//   - Domain entities (Package, Node, FileTree)
//   - Operations (LinkCreate, LinkDelete, DirCreate, etc.)
//   - Plans (collections of operations)
//   - Results (monadic error handling)
//   - Paths (phantom-typed file paths)
//   - Conflicts (installation conflicts)
//   - Errors (domain-specific errors)
//   - Execution results and checkpoints
//   - Diagnostic reports and health checks
//
// Port Interfaces:
//   - FS: Filesystem abstraction
//   - Logger: Structured logging
//   - Tracer: Distributed tracing
//   - Metrics: Metrics collection
//
// This package is internal and not exposed to library consumers.
// Public API consumers should use pkg/dot which re-exports these types.
//
// Architecture:
//
// The domain package is the foundation of the dot architecture. All other
// internal packages (executor, pipeline, scanner, planner, manifest) import
// from this package rather than from pkg/dot to avoid import cycles.
//
// The pkg/dot package re-exports all domain types using type aliases,
// providing a stable public API while allowing internal implementation
// to evolve independently.
package domain
