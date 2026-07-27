# Architecture Documentation

This document describes the technical architecture of dot, a type-safe symbolic link manager for configuration files.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Architectural Layers](#architectural-layers)
- [Design Principles](#design-principles)
- [Component Structure](#component-structure)
- [Data Flow](#data-flow)
- [Type System](#type-system)
- [Error Handling](#error-handling)
- [Concurrency Model](#concurrency-model)
- [Testing Strategy](#testing-strategy)
- [Dependency Rules](#dependency-rules)

## Architecture Overview

dot follows a layered architecture inspired by hexagonal architecture (ports and adapters) and functional programming principles. The system separates pure functional logic from side-effecting operations, enabling deterministic testing and safe execution.

### Core Architecture Pattern

The architecture implements the "Functional Core, Imperative Shell" pattern:

- **Functional Core**: Pure domain logic with no side effects (scanning, planning, resolution)
- **Imperative Shell**: Side-effecting operations isolated to executor layer (filesystem modifications)

This separation enables:
- Deterministic testing of core logic without filesystem access
- Safe rollback of failed operations
- Property-based testing of algebraic laws
- Parallelization of independent operations

## Architectural Layers

The system comprises six distinct layers, each with specific responsibilities and dependency constraints.

```mermaid
graph TB
    CLI[CLI Layer<br/>cmd/dot/]:::cliLayer
    API[API Layer<br/>pkg/dot/]:::apiLayer
    Pipeline[Pipeline Layer<br/>internal/pipeline/]:::pipelineLayer
    Executor[Executor Layer<br/>internal/executor/]:::executorLayer
    Core[Core Layer<br/>internal/scanner/<br/>internal/planner/<br/>internal/ignore/]:::coreLayer
    Domain[Domain Layer<br/>internal/domain/]:::domainLayer
    Adapters[Adapters<br/>internal/adapters/]:::adaptersLayer
    
    CLI --> API
    API --> Pipeline
    API --> Executor
    Pipeline --> Core
    Executor --> Domain
    Core --> Domain
    Adapters --> Domain
    
    style CLI fill:#4A90E2,stroke:#2C5F8D,color:#fff
    style API fill:#50C878,stroke:#2D7A4A,color:#fff
    style Pipeline fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style Executor fill:#E67E22,stroke:#A84E0F,color:#fff
    style Core fill:#3498DB,stroke:#1F618D,color:#fff
    style Domain fill:#2ECC71,stroke:#1E8449,color:#fff
    style Adapters fill:#95A5A6,stroke:#5D6D7E,color:#fff
    
    classDef cliLayer stroke-width:3px
    classDef apiLayer stroke-width:3px
    classDef pipelineLayer stroke-width:3px
    classDef executorLayer stroke-width:3px
    classDef coreLayer stroke-width:3px
    classDef domainLayer stroke-width:3px
    classDef adaptersLayer stroke-width:2px,stroke-dasharray: 5 5
```

### 1. Domain Layer

**Location**: `internal/domain/`

**Purpose**: Pure domain model defining core types, operations, and port interfaces.

**Key Components**:
- Domain entities: `Package`, `Node`, `Plan`, `Operation`
- Phantom-typed paths: `PackagePath`, `TargetPath`, `FilePath`
- Port interfaces: `FS`, `Logger`, `Tracer`, `Metrics`
- Result types: `Result[T]` for monadic error handling
- Conflict representations
- Error types

**Characteristics**:
- No external dependencies except standard library
- All functions are pure (no I/O operations)
- Phantom types provide compile-time path safety
- Defines contracts (interfaces) for infrastructure

**Dependencies**: None (depends only on Go standard library)

### 2. Core Layer

**Location**: `internal/scanner/`, `internal/planner/`, `internal/ignore/`

**Purpose**: Pure functional logic for scanning packages, computing desired state, and planning operations.

**Key Components**:

**Scanner** (`internal/scanner/`):
- Package scanning with ignore pattern support
- Filesystem tree construction
- Dotfile name translation (e.g., `dot-bashrc` to `.bashrc`)

**Planner** (`internal/planner/`):
- Desired state computation
- Conflict detection and resolution
- Dependency graph construction
- Topological sorting for operation ordering
- Parallel execution batch computation

**Ignore** (`internal/ignore/`):
- Pattern matching for file exclusion
- Default ignore patterns
- Custom pattern support

**Characteristics**:
- Pure functions with no side effects
- Deterministic outputs for given inputs
- Testable without filesystem access
- Uses Result types for error handling

**Dependencies**: Domain layer only

### 3. Pipeline Layer

**Location**: `internal/pipeline/`

**Purpose**: Composable pipeline stages with generic type parameters for operation orchestration.

**Key Components**:
- `Pipeline[TIn, TOut]`: Generic pipeline type
- `ScanStage()`: Package scanning stage
- `PlanStage()`: Desired state computation stage
- `ResolveStage()`: Conflict resolution stage
- `ManagePipeline`: Composition of stages for manage operations

**Characteristics**:
- Generic type parameters for type safety
- Composable stages using function composition
- Context-aware for cancellation support
- Monadic error propagation through stages

**Pipeline Composition Example**:
```
ScanInput -> ScanStage -> []Package -> PlanStage -> DesiredState -> ResolveStage -> SortStage -> Plan
```

**Dependencies**: Domain and Core layers

### 4. Executor Layer

**Location**: `internal/executor/`

**Purpose**: Transactional execution of plans with two-phase commit and automatic rollback.

**Key Components**:
- `Executor`: Main execution engine
- `CheckpointStore`: State checkpoint for rollback
- Precondition validation
- Operation execution
- Automatic rollback on failure
- Parallel execution support

**Execution Phases**:

1. **Prepare Phase**: Validate all operations before execution
2. **Checkpoint Creation**: Save state for potential rollback
3. **Commit Phase**: Execute operations (sequential or parallel)
4. **Rollback Phase**: Undo operations if failures occur
5. **Checkpoint Cleanup**: Remove checkpoint on success

```mermaid
stateDiagram-v2
    [*] --> Prepare: Receive Plan
    
    Prepare --> ValidateOps: Validate Operations
    ValidateOps --> CheckPreconditions: Check Preconditions
    CheckPreconditions --> CreateCheckpoint: All Valid
    CheckPreconditions --> Failed: Validation Failed
    
    CreateCheckpoint --> CommitPhase: Checkpoint Saved
    CreateCheckpoint --> Failed: Checkpoint Failed
    
    CommitPhase --> ExecuteBatch1: Batch 1 (Parallel)
    ExecuteBatch1 --> ExecuteBatch2: Success
    ExecuteBatch1 --> Rollback: Operation Failed
    
    ExecuteBatch2 --> ExecuteBatch3: Success
    ExecuteBatch2 --> Rollback: Operation Failed
    
    ExecuteBatch3 --> UpdateManifest: All Batches Complete
    ExecuteBatch3 --> Rollback: Operation Failed
    
    UpdateManifest --> CleanupCheckpoint: Manifest Updated
    UpdateManifest --> Rollback: Manifest Update Failed
    
    CleanupCheckpoint --> Success: Checkpoint Removed
    
    Rollback --> RestoreState: Undo Operations
    RestoreState --> RemoveCheckpoint: State Restored
    RemoveCheckpoint --> Failed: Rollback Complete
    
    Success --> [*]
    Failed --> [*]
    
    note right of Prepare
        Validate all operations
        can be executed
    end note
    
    note right of CommitPhase
        Execute in topologically
        sorted batches
    end note
    
    note right of Rollback
        Rollback reverts executed
        operations; irreversible ones
        are reported, not silently lost
    end note
```

**Characteristics**:
- Best-effort transactional semantics: executed operations are reverted on failure
- Irreversible operations (notably `DirRemoveAll`) are counted, not silently dropped
- Support for parallel execution of independent operations
- Comprehensive error tracking

**Dependencies**: Domain layer for types and ports

### 5. API Layer

**Location**: `pkg/dot/`

**Purpose**: Clean public Go library interface for embedding dot in other applications.

**Key Components**:
- `Client`: Facade delegating to specialized services
- `Config`: Configuration structure with validation
- Service implementations:
  - `ManageService`: Package installation
  - `UnmanageService`: Package removal
  - `StatusService`: Status queries
  - `DoctorService`: Health checks
  - `AdoptService`: File adoption
  - `ManifestService`: State persistence
  - `BootstrapService`: Bootstrap from a config manifest
  - `CloneService`: Clone a dotfiles repository and manage selected packages

**Service Pattern**:

The Client uses a service-based architecture where each major operation is implemented by a dedicated service. This provides:
- Single Responsibility Principle adherence
- Independent testing of each service
- Clear boundaries between concerns
- Maintainable codebase

```mermaid
graph LR
    Client[Client<br/>Facade]:::clientNode
    
    ManageService[ManageService<br/>Package Installation]:::serviceNode
    UnmanageService[UnmanageService<br/>Package Removal]:::serviceNode
    StatusService[StatusService<br/>Status Queries]:::serviceNode
    DoctorService[DoctorService<br/>Health Checks]:::serviceNode
    AdoptService[AdoptService<br/>File Adoption]:::serviceNode
    ManifestService[ManifestService<br/>State Persistence]:::serviceNode
    BootstrapService[BootstrapService<br/>Config Bootstrap]:::serviceNode
    CloneService[CloneService<br/>Repository Clone]:::serviceNode
    
    Pipeline[Pipeline Layer]:::layerNode
    Executor[Executor Layer]:::layerNode
    Manifest[Manifest Store]:::layerNode
    
    Client --> ManageService
    Client --> UnmanageService
    Client --> StatusService
    Client --> DoctorService
    Client --> AdoptService
    Client --> ManifestService
    Client --> BootstrapService
    Client --> CloneService
    
    CloneService --> ManageService
    
    ManageService --> Pipeline
    ManageService --> Executor
    ManageService --> Manifest
    
    UnmanageService --> Executor
    UnmanageService --> Manifest
    
    StatusService --> Manifest
    DoctorService --> Manifest
    AdoptService --> Executor
    AdoptService --> Manifest
    
    ManifestService --> Manifest
    
    style Client fill:#4A90E2,stroke:#2C5F8D,color:#fff,stroke-width:4px
    style ManageService fill:#50C878,stroke:#2D7A4A,color:#fff
    style UnmanageService fill:#E67E22,stroke:#A84E0F,color:#fff
    style StatusService fill:#3498DB,stroke:#1F618D,color:#fff
    style DoctorService fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style AdoptService fill:#1ABC9C,stroke:#148F77,color:#fff
    style ManifestService fill:#F39C12,stroke:#B97A0F,color:#fff
    style BootstrapService fill:#7F8C8D,stroke:#5D6D7E,color:#fff
    style CloneService fill:#C0392B,stroke:#7B241C,color:#fff
    style Pipeline fill:#34495E,stroke:#1C2833,color:#fff
    style Executor fill:#34495E,stroke:#1C2833,color:#fff
    style Manifest fill:#34495E,stroke:#1C2833,color:#fff
    
    classDef clientNode stroke-width:4px
    classDef serviceNode stroke-width:2px
    classDef layerNode stroke-width:2px,stroke-dasharray: 5 5
```

**Characteristics**:
- Stable public API
- Thread-safe operations
- Service-based delegation
- Comprehensive validation

**Dependencies**: All internal layers

### 6. CLI Layer

**Location**: `cmd/dot/`

**Purpose**: Cobra-based command-line interface providing user interaction.

**Key Components**:
- Command definitions (manage, unmanage, status, doctor, list, adopt)
- Flag parsing and validation
- Configuration loading from files and environment
- Output formatting (table, JSON, YAML)
- Progress indicators
- Error rendering with suggestions

**Characteristics**:
- Cobra command structure
- Viper configuration management
- Multiple output formats
- Rich error messages with context

**Dependencies**: API layer, plus presentation-only helpers under `internal/cli/`. No other `internal/*` package is imported by non-test files.

## Design Principles

### Functional Core, Imperative Shell

Pure functional logic (scanning, planning, resolution) is separated from side-effecting operations (filesystem modifications). This enables:

- Deterministic testing without filesystem access
- Property-based testing of algebraic laws
- Safe parallelization
- Reliable rollback mechanisms

### Type Safety

Phantom types encode path semantics at compile time:

```go
// Path[K] is a single generic type parameterised by a phantom kind marker.
type Path[K PathKind] struct {
    path string
}

// The three concrete path types are aliases of distinct instantiations.
type PackagePath = Path[PackageDirKind]
type TargetPath  = Path[TargetDirKind]
type FilePath    = Path[FileDirKind]
```

This prevents path-related bugs:
- Cannot pass target path where package path expected
- Cannot mix relative and absolute paths incorrectly
- Compile-time validation of path operations

### Explicit Error Handling

The system uses `Result[T]` types for monadic error handling:

```go
type Result[T any] struct {
    value T
    err   error
    isOk  bool
}
```

The `isOk` discriminant distinguishes `Ok(zeroValue)` from `Err(nil)`.

This provides:
- No silent failures
- Explicit error propagation
- Composable error handling
- Type-safe success values

### Transactional Operations

All operations use two-phase commit:

1. **Validate**: Check preconditions
2. **Execute**: Apply changes
3. **Rollback**: Undo on failure

This provides:
- Precondition failures detected before any change is applied
- Automatic cleanup on failure, limited to reversible operations
- Explicit accounting of operations that could not be reverted
- Safe concurrent execution

### Dependency Inversion

Infrastructure dependencies are abstracted through port interfaces:

```go
type FS interface {
    FSReader
    FSWriter
}

type FSReader interface {
    Stat(ctx context.Context, path string) (FileInfo, error)
    Lstat(ctx context.Context, path string) (FileInfo, error)
    ReadDir(ctx context.Context, path string) ([]DirEntry, error)
    ReadLink(ctx context.Context, path string) (string, error)
    ReadFile(ctx context.Context, path string) ([]byte, error)
    Exists(ctx context.Context, path string) bool
    IsDir(ctx context.Context, path string) (bool, error)
    IsSymlink(ctx context.Context, path string) (bool, error)
}

type FSWriter interface {
    WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error
    Mkdir(ctx context.Context, path string, perm os.FileMode) error
    MkdirAll(ctx context.Context, path string, perm os.FileMode) error
    Remove(ctx context.Context, path string) error
    RemoveAll(ctx context.Context, path string) error
    Symlink(ctx context.Context, oldname, newname string) error
    Rename(ctx context.Context, oldpath, newpath string) error
}
```

This enables:
- Testing with memory-based implementations
- Platform-specific adapters
- Mock implementations for testing
- Isolation of domain logic from infrastructure

## Component Structure

### Adapter Pattern

The system uses adapters to implement port interfaces:

**Filesystem Adapters** (`internal/adapters/`):
- `OSFilesystem` (`NewOSFilesystem`): Production filesystem using the `os` package
- `MemFS` (`NewMemFS`): In-memory filesystem for testing

**Logging Adapters** (`internal/adapters/`):
- `SlogLogger` (`NewSlogLogger`, `NewConsoleLogger`): Production logger using `log/slog`
- `NoopLogger` (`NewNoopLogger`): Silent logger for testing

**Observability Adapters** (`internal/adapters/`):
- `NoopTracer`, `NoopMetrics`: Default no-op implementations

Dry-run mode is not a filesystem adapter. `Config.DryRun` is threaded into each
service (pkg/dot/client.go:122-131), which plans operations but skips execution.

This pattern provides:
- Swappable implementations
- Testability without real filesystem
- Dry-run mode support
- Platform-specific optimizations

### Manifest Persistence

**Location**: `internal/manifest/`

**Purpose**: State persistence for tracking installed packages.

**Components**:
- `Manifest`: Package installation record
- `ManifestStore`: Interface for persistence
- `FSManifestStore`: File-based implementation

**Manifest Structure**:
```go
type Manifest struct {
    Version    string                 `json:"version"`
    UpdatedAt  time.Time              `json:"updated_at"`
    Packages   map[string]PackageInfo `json:"packages"`
    Hashes     map[string]string      `json:"hashes"`
    Repository *RepositoryInfo        `json:"repository,omitempty"`
    Doctor     *DoctorState           `json:"doctor,omitempty"`
}

type PackageInfo struct {
    Name        string            `json:"name"`
    InstalledAt time.Time         `json:"installed_at"`
    LinkCount   int               `json:"link_count"`
    Links       []string          `json:"links"`
    Backups     map[string]string `json:"backups,omitempty"`
    Source      PackageSource     `json:"source,omitempty"`
    TargetDir   string            `json:"target_dir,omitempty"`
    PackageDir  string            `json:"package_dir,omitempty"`
}
```

**Persistence Location**: `<ManifestDir>/.dot-manifest.json`, defaulting to `<TargetDir>/.dot-manifest.json` when `Config.ManifestDir` is empty. Writes are guarded by a `.dot-manifest.lock` sibling file.

**Purpose**:
- Track installed packages
- Enable incremental updates (detect changed packages)
- Support status queries without filesystem scanning
- Facilitate safe uninstall operations

### Backup Handling

When `Config.Backup` is set, `PolicyBackup` (`applyBackupPolicy`, internal/planner/policies.go:84)
expands a conflict into three operations: `FileBackup`, `FileDelete`, and the
original link operation.

- Backup names are `<backupDir>/<filename>.<pathTag>.<timestamp>`, where
  `pathTag` is a short hash of the full conflict path. Files sharing a basename
  in different directories therefore never overwrite each other's backups
  (internal/planner/policies.go:95-101).
- `FileBackup.Execute` creates the backup directory on demand via `MkdirAll`
  before writing, so a missing `<TargetDir>/.dot-backup` is not an error
  (internal/domain/operation.go:678-683).
- The paired `FileDelete` records the backup path
  (`NewFileDeleteWithBackup`, internal/domain/operation.go:744), so rollback
  restores the original file from the backup.
- Backup paths are recorded per package in `PackageInfo.Backups`
  (internal/manifest/manifest.go:36) for later restore.

### Configuration System

**Location**: `internal/config/`

**Purpose**: Configuration loading, validation, and marshaling.

**Features**:
- Multiple format support (YAML, JSON, TOML)
- Precedence handling (flags > environment > files > defaults)
- XDG Base Directory Specification compliance
- Schema validation
- Default value application

**Configuration Sources** (in precedence order):
1. Command-line flags
2. Environment variables (`DOT_*` prefix)
3. Project-local config (`./.dotrc`)
4. User config (`~/.config/dot/config.yaml`)
5. System config (`/etc/dot/config.yaml`)
6. Default values

## Data Flow

### Manage Operation Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Layer
    participant API as API Layer
    participant Pipeline as Pipeline Layer
    participant Scanner as Scanner
    participant Planner as Planner
    participant Executor as Executor Layer
    participant FS as Filesystem
    participant Manifest as Manifest Store
    
    User->>CLI: dot manage vim tmux
    CLI->>CLI: Parse flags & config
    CLI->>API: Client.Manage(ctx, packages)
    
    API->>Pipeline: ManagePipeline.Execute()
    
    rect rgb(40, 70, 100)
        note right of Pipeline: Scan Stage
        Pipeline->>Scanner: Scan packages
        Scanner->>FS: Read package directories
        FS-->>Scanner: File tree
        Scanner-->>Pipeline: []Package
    end
    
    rect rgb(60, 50, 100)
        note right of Pipeline: Plan Stage
        Pipeline->>Planner: Compute desired state
        Planner->>Planner: Build dependency graph
        Planner->>Planner: Topological sort
        Planner-->>Pipeline: Plan with operations
    end
    
    rect rgb(80, 100, 50)
        note right of Pipeline: Execute Stage
        Pipeline->>Executor: Execute plan
        Executor->>Executor: Validate preconditions
        Executor->>FS: Create checkpoint
        
        loop For each operation batch
            Executor->>FS: Execute operations (parallel)
            alt Success
                FS-->>Executor: OK
            else Failure
                FS-->>Executor: Error
                Executor->>FS: Rollback changes
                Executor-->>API: ExecutionError
            end
        end
        
        Executor->>Manifest: Update package records
        Manifest->>FS: Write .dot-manifest.json
        Executor-->>Pipeline: Success
    end
    
    Pipeline-->>API: Result
    API-->>CLI: ManageResult
    CLI->>User: Display results
```

### Unmanage Operation Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Layer
    participant API as API Layer
    participant Manifest as Manifest Store
    participant Planner as Planner
    participant Executor as Executor Layer
    participant FS as Filesystem
    
    User->>CLI: dot unmanage vim
    CLI->>CLI: Parse flags & config
    CLI->>API: Client.Unmanage(ctx, packages)
    
    API->>Manifest: Load package manifests
    Manifest->>FS: Read .dot-manifest.json
    FS-->>Manifest: Manifest data
    Manifest-->>API: Package records
    
    rect rgb(100, 60, 60)
        note right of API: Generate Delete Operations
        API->>API: Create delete operations
        API->>Planner: Build dependency graph (reverse)
        Planner->>Planner: Topological sort (deletion order)
        Planner-->>API: Deletion plan
    end
    
    rect rgb(80, 100, 50)
        note right of API: Execute Deletions
        API->>Executor: Execute deletion plan
        Executor->>Executor: Validate preconditions
        Executor->>FS: Create checkpoint
        
        loop For each operation batch (reverse order)
            Executor->>FS: Delete links/dirs
            alt Success
                FS-->>Executor: OK
            else Failure
                FS-->>Executor: Error
                Executor->>FS: Rollback deletions
                Executor-->>API: ExecutionError
            end
        end
        
        Executor-->>API: Success
    end
    
    API->>Manifest: Remove package records
    Manifest->>FS: Update .dot-manifest.json
    API-->>CLI: UnmanageResult
    CLI->>User: Display results
```

### Status Query Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Layer
    participant API as API Layer
    participant StatusService as Status Service
    participant Manifest as Manifest Store
    participant FS as Filesystem
    participant Renderer as Output Renderer
    
    User->>CLI: dot status [packages]
    CLI->>CLI: Parse flags & config
    CLI->>API: Client.Status(ctx, packages)
    
    API->>StatusService: Query status
    
    rect rgb(50, 80, 120)
        note right of StatusService: Load Manifests
        StatusService->>Manifest: Load package manifests
        Manifest->>FS: Read .dot-manifest.json
        FS-->>Manifest: Manifest data
        Manifest-->>StatusService: Package records
    end
    
    rect rgb(70, 90, 70)
        note right of StatusService: Build Status
        StatusService->>StatusService: Build status structures
        StatusService->>StatusService: Compute installation state
        StatusService->>StatusService: Calculate timestamps
        StatusService->>StatusService: Count links per package
    end
    
    StatusService-->>API: StatusResult
    API-->>CLI: Status data
    
    rect rgb(90, 70, 100)
        note right of CLI: Render Output
        CLI->>Renderer: Format output (table/JSON/YAML)
        
        alt Table Format
            Renderer->>Renderer: Build table structure
            Renderer->>Renderer: Format columns
            Renderer-->>CLI: Table output
        else JSON Format
            Renderer->>Renderer: Marshal to JSON
            Renderer-->>CLI: JSON output
        else YAML Format
            Renderer->>Renderer: Marshal to YAML
            Renderer-->>CLI: YAML output
        end
    end
    
    CLI->>User: Display status
```

### Doctor Health Check Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Layer
    participant API as API Layer
    participant DoctorService as Doctor Service
    participant Manifest as Manifest Store
    participant FS as Filesystem
    participant Renderer as Diagnostic Renderer
    
    User->>CLI: dot doctor
    CLI->>CLI: Parse flags & config
    CLI->>API: Client.Doctor(ctx)
    
    API->>DoctorService: Run health checks
    
    rect rgb(50, 80, 120)
        note right of DoctorService: Load State
        DoctorService->>Manifest: Load all package manifests
        Manifest->>FS: Read .dot-manifest.json
        FS-->>Manifest: All package records
        Manifest-->>DoctorService: Installed packages
    end
    
    rect rgb(70, 90, 70)
        note right of DoctorService: Verify Links
        loop For each package
            DoctorService->>FS: Check symlink exists
            DoctorService->>FS: Read symlink target
            
            alt Link Valid
                FS-->>DoctorService: Target matches expected
                DoctorService->>DoctorService: Mark as healthy
            else Link Broken
                FS-->>DoctorService: Target missing/wrong
                DoctorService->>DoctorService: Record issue
            else Link Missing
                FS-->>DoctorService: Symlink not found
                DoctorService->>DoctorService: Record issue
            end
        end
    end
    
    rect rgb(100, 70, 80)
        note right of DoctorService: Detect Issues
        DoctorService->>FS: Scan target directory
        FS-->>DoctorService: All files/links
        
        DoctorService->>DoctorService: Detect broken links
        DoctorService->>DoctorService: Detect orphaned files
        DoctorService->>DoctorService: Detect conflicts
        DoctorService->>DoctorService: Detect wrong targets
        DoctorService->>DoctorService: Check manifest integrity
    end
    
    rect rgb(90, 70, 100)
        note right of DoctorService: Generate Report
        DoctorService->>DoctorService: Categorize issues by severity
        DoctorService->>DoctorService: Generate suggestions
        DoctorService->>DoctorService: Build diagnostic report
    end
    
    DoctorService-->>API: DiagnosticReport
    API-->>CLI: Report data
    
    CLI->>Renderer: Render diagnostics
    
    alt All Healthy
        Renderer->>Renderer: Format success message
        Renderer-->>CLI: Health report
    else Issues Found
        Renderer->>Renderer: Format issue list
        Renderer->>Renderer: Add suggestions
        Renderer->>Renderer: Add fix commands
        Renderer-->>CLI: Diagnostic report with fixes
    end
    
    CLI->>User: Display diagnostics
```

## Type System

### Phantom Types for Path Safety

Phantom types encode path semantics at the type level:

```go
// PackagePath represents a path within the package directory
type PackagePath struct {
    path string
}

// TargetPath represents a path in the target directory
type TargetPath struct {
    path string
}

// FilePath represents a generic file path
type FilePath struct {
    path string
}
```

**Benefits**:
- Compile-time prevention of path mix-ups
- Self-documenting function signatures
- Type-guided refactoring
- Elimination of path-related bugs

**Usage Example**:
```go
// Function signature clearly indicates path expectations
func scanPackage(path PackagePath) Result[Package]

// Compiler prevents incorrect usage
scanPackage(targetPath)  // Compile error: type mismatch
```

### Result Type for Error Handling

The `Result[T]` type provides monadic error handling:

```go
type Result[T any] struct {
    value T
    err   error
}

func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool
func (r Result[T]) Unwrap() T
func (r Result[T]) UnwrapErr() error
func (r Result[T]) UnwrapOr(defaultValue T) T
```

**Benefits**:
- Explicit success or failure states
- Type-safe value extraction
- Composable error handling
- No nil pointer dereferencing

### Operation Types

Operations are represented as an interface with concrete implementations:

```go
type Operation interface {
    ID() OperationID
    Kind() OperationKind
    Validate() error
    Dependencies() []Operation
    Execute(ctx context.Context, fs FS) error
    Rollback(ctx context.Context, fs FS) error
    String() string
    Equals(other Operation) bool
}

// Concrete operation types:
type LinkCreate struct { ... }
type LinkDelete struct { ... }
type DirCreate struct { ... }
type DirDelete struct { ... }
type DirRemoveAll struct { ... }
type DirCopy struct { ... }
type FileMove struct { ... }
type FileBackup struct { ... }
type FileDelete struct { ... }
```

**Operation Kinds**:
- `OpKindLinkCreate`: Create symbolic link
- `OpKindLinkDelete`: Remove symbolic link
- `OpKindDirCreate`: Create directory
- `OpKindDirDelete`: Remove an empty directory
- `OpKindDirRemoveAll`: Recursively remove a directory and its contents (not reversible)
- `OpKindDirCopy`: Recursively copy a directory
- `OpKindFileMove`: Move a file
- `OpKindFileBackup`: Back up an existing file
- `OpKindFileDelete`: Delete a file

## Error Handling

### Error Type Hierarchy

Domain-specific errors with rich context:

```go
// Core errors
type ErrInvalidPath struct { Path string }
type ErrPackageNotFound struct { Package string }
type ErrConflict struct { Path string, Reason string }

// Execution errors
type ErrExecutionFailed struct {
    Executed       int
    Failed         int
    RolledBack     int
    RollbackFailed int   // executed ops whose rollback failed or was impossible
    Errors         []error
}

// Planning errors
type ErrCyclicDependency struct { Cycle []Operation }
type ErrEmptyPlan struct {}
```

### Error Wrapping

Errors are wrapped with context using `fmt.Errorf` and `%w`:

```go
if err := operation.Execute(); err != nil {
    return fmt.Errorf("failed to execute %s: %w", operation.Kind(), err)
}
```

### Error Aggregation

Multiple errors are collected and reported together:

```go
type ExecutionResult struct {
    Executed       []domain.OperationID
    Failed         []domain.OperationID
    RolledBack     []domain.OperationID
    RollbackFailed []domain.OperationID
    Errors         []error
}
```

## Concurrency Model

### Thread Safety

All public API operations are safe for concurrent use:

```go
client, _ := dot.NewClient(config)

// Safe to call from multiple goroutines
go client.Manage(ctx, "vim")
go client.Status(ctx)
```

### Parallel Execution

The planner computes parallel execution batches:

1. **Dependency Analysis**: Build dependency graph
2. **Topological Sort**: Order operations respecting dependencies
3. **Batch Computation**: Group independent operations
4. **Parallel Execution**: Execute batches concurrently

**Example**:
```
Batch 1 (parallel):
  - CreateDir ~/.config
  - CreateDir ~/.local

Batch 2 (parallel, depends on Batch 1):
  - CreateLink ~/.config/nvim
  - CreateLink ~/.local/bin/script

Batch 3 (depends on Batch 2):
  - CreateLink ~/.config/nvim/init.vim
```

```mermaid
graph TB
    subgraph "Batch 1 - Parallel Execution"
        A[CreateDir<br/>~/.config]:::batch1
        B[CreateDir<br/>~/.local]:::batch1
        C[CreateDir<br/>~/.cache]:::batch1
    end
    
    subgraph "Batch 2 - Parallel Execution"
        D[CreateLink<br/>~/.config/nvim]:::batch2
        E[CreateLink<br/>~/.local/bin]:::batch2
        F[CreateLink<br/>~/.cache/app]:::batch2
    end
    
    subgraph "Batch 3 - Parallel Execution"
        G[CreateLink<br/>~/.config/nvim/init.vim]:::batch3
        H[CreateLink<br/>~/.config/nvim/lua]:::batch3
        I[CreateLink<br/>~/.local/bin/script]:::batch3
    end
    
    subgraph "Batch 4 - Sequential"
        J[CreateLink<br/>~/.config/nvim/lua/config.lua]:::batch4
    end
    
    A --> D
    A --> G
    A --> H
    B --> E
    B --> I
    C --> F
    
    D --> G
    D --> H
    E --> I
    
    G --> J
    H --> J
    
    style A fill:#3498DB,stroke:#1F618D,color:#fff
    style B fill:#3498DB,stroke:#1F618D,color:#fff
    style C fill:#3498DB,stroke:#1F618D,color:#fff
    style D fill:#50C878,stroke:#2D7A4A,color:#fff
    style E fill:#50C878,stroke:#2D7A4A,color:#fff
    style F fill:#50C878,stroke:#2D7A4A,color:#fff
    style G fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style H fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style I fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style J fill:#E67E22,stroke:#A84E0F,color:#fff
    
    classDef batch1 stroke-width:3px
    classDef batch2 stroke-width:3px
    classDef batch3 stroke-width:3px
    classDef batch4 stroke-width:3px
```

### Context Support

All operations support `context.Context` for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := client.Manage(ctx, packages...)
// Respects context cancellation and timeout
```

## Testing Strategy

### Layer-Specific Testing

**Domain Layer**:
- Pure function testing
- Property-based testing of algebraic laws
- No filesystem access required

**Core Layer**:
- Table-driven tests
- Edge case coverage
- In-memory filesystem for deterministic tests

**Pipeline Layer**:
- Integration tests with memory filesystem
- Error propagation verification
- Context cancellation testing

**Executor Layer**:
- Rollback mechanism verification
- Checkpoint functionality
- Failure scenario coverage

**API Layer**:
- End-to-end integration tests
- Service interaction testing
- Manifest persistence verification

**CLI Layer**:
- Command parsing tests
- Output format verification
- Error message validation

### Test Coverage Requirements

The enforced gate is the unweighted mean of per-function coverage percentages
reported by `go tool cover -func`, not total statement coverage. Bubble Tea UI
and interactive adoption files under `internal/cli/adopt/` are excluded.

- Local (`make check-coverage`, `make cs`, and therefore `make check`): 60%
- CI (`.github/workflows/ci.yml`, `test` job): 75%
- All error paths must be tested
- Edge cases must have explicit tests

### Testing Tools

- Standard library `testing` package
- `testify/assert` for assertions
- Table-driven test pattern
- Memory-based filesystem adapter
- Golden file testing for outputs

## Architectural Change Log

Entries are dated and describe the state at the time of the change. The CHANGELOG
is authoritative for released versions.

### Rollback Accounting (2026-07)

**Problem**: Rollback failures were invisible. An operation whose `Rollback` returned an error, or which could not be reverted at all, left the system in a partial state that callers could not detect.

**Solution**:
- `Executor.rollback` returns `(rolledBack, rollbackFailed)` (internal/executor/executor.go:442)
- `ExecutionResult.RollbackFailed` records the affected operation IDs (internal/executor/result.go:11)
- `ErrExecutionFailed.RollbackFailed` reports the count in the error message (internal/domain/errors.go:100)
- `ErrRollbackImpossible` (internal/domain/errors.go:142) is returned by `DirRemoveAll.Rollback`, since recursive deletion is not reversible

**Impact**: Callers can distinguish clean rollback from partial restoration. The former all-or-nothing claim no longer holds and is documented as best-effort.

### Rollback Detachment from Cancellation (2026-07)

**Problem**: Ctrl-C mid-plan cancelled the context that rollback itself ran under, so filesystem adapters refused every restore call and the plan's partial effects persisted.

**Solution**: `rollback` calls `context.WithoutCancel(ctx)` (internal/executor/executor.go:443), preserving tracing and logging values while ignoring the parent's cancellation.

**Impact**: Interrupting a plan restores prior state instead of leaving it half-applied.

### Deletion Safety and Backup Naming (2026-07)

**Problem**: `LinkDelete` deleted whatever occupied the recorded path, and backup filenames keyed only on basename plus timestamp, so two conflicting files with the same basename in different directories overwrote each other's backups. `FileBackup` also failed outright when the backup directory did not exist.

**Solution**:
- `NewLinkDeleteWithDestination` (internal/domain/operation.go:183) records the planned symlink destination; `LinkDelete` re-verifies with `Lstat` and refuses mismatches with `ErrUnsafeDelete`
- Backup names include a short hash of the full conflict path (internal/planner/policies.go:95-101)
- `FileBackup.Execute` creates the backup directory on demand (internal/domain/operation.go:678-683)

**Impact**: Stale plans cannot destroy changed user data, and backups no longer collide or fail on a missing directory.

### CLI State Management (2025-01)

**Problem**: The CLI layer used a global `globalCfg` variable to store parsed command-line flags, creating implicit dependencies and making testing difficult.

**Solution**: Refactored to use explicit `CLIFlags` struct passed as parameters:
- Eliminated global state from `cmd/dot/root.go`
- All functions accept `*CLIFlags` parameter explicitly
- Configuration flows explicitly through call chains
- Improved testability through dependency injection

**Impact**:
- Clearer data flow and dependencies
- Easier to test individual functions
- No hidden global state
- Aligns with functional programming principles

### Result[T] Usage Guidelines (2025-01)

**Problem**: The `Result[T]` monad's `Unwrap()` and `UnwrapErr()` methods panic if called on the wrong variant, creating potential runtime failures.

**Solution**: Established comprehensive usage guidelines in `internal/domain/doc.go`:
- Use `Result[T]` for functional core composition
- Prefer `(T, error)` for leaf functions and public APIs
- Always guard `Unwrap()` calls with `IsOk()`/`IsErr()` checks
- Use `UnwrapOr()` for safe access with defaults
- Avoid redundant path validation reconstruction

**Audit Findings**:
- All production `Unwrap()` calls are properly guarded
- Error propagation paths check before unwrapping
- No unsafe patterns found in codebase
- A few redundant but safe reconstructions identified

**Impact**:
- Safer error handling patterns
- Clear guidelines for contributors
- Reduced panic risk in production
- Better alignment with Go idioms at boundaries

### Executor Context Cancellation (2025-01)

**Problem**: Executor loops did not check for context cancellation, preventing graceful shutdown and proper resource cleanup.

**Solution**: Added explicit `ctx.Err()` checks at key points and detached rollback from cancellation:
- Prepare phase: Check before validating each operation
- Sequential execution: Check before executing each operation
- Parallel execution: Check before executing each batch
- Rollback: runs under `context.WithoutCancel(ctx)` (internal/executor/executor.go:443), preserving trace and logging values while ignoring the cancelled deadline, because filesystem adapters return `ctx.Err()` on every call

**New Error Type**:
```go
type ErrExecutionCancelled struct {
    Executed int
    Skipped  int
}
```

**Cancellation Behavior**:
- Operations already executed are rolled back even after Ctrl-C, because rollback uses a detached context
- Remaining operations are skipped and counted
- Cancellation error returned with accurate metrics
- Operations that cannot be reverted are counted in `ExecutionResult.RollbackFailed` rather than silently dropped

**Impact**:
- Graceful shutdown support
- Proper resource cleanup
- Better user experience for long operations
- System consistency maintained during cancellation

### Set Type Optimization (2025-01)

**Problem**: Many `map[string]bool` types were used as sets, wasting memory (1 byte per entry) and obscuring intent.

**Solution**: Replaced with `map[string]struct{}` where only key presence matters:
- `struct{}` has zero size vs `bool` (1 byte)
- Changed membership checks to comma-ok idiom
- Updated 15 files across internal packages

**Before**:
```go
set := make(map[string]bool)
set[key] = true
if set[key] { ... }
```

**After**:
```go
set := make(map[string]struct{})
set[key] = struct{}{}
if _, exists := set[key]; exists { ... }
```

**Affected Areas**:
- Executor: pending operations tracking
- Planner: current state directory tracking
- Pipeline: path checking during state scan
- Ignore system: visited path tracking
- Doctor: managed link and directory sets
- CLI: managed paths and excluded directories

**Impact**:
- Reduced memory overhead for large sets
- Clearer intent (set vs boolean map)
- More idiomatic Go code
- No behavior changes

## Dependency Rules

### Inward Dependencies

Dependencies flow inward toward the domain:

```mermaid
graph TD
    CLI[CLI Layer<br/>cmd/dot/]:::cliLayer
    API[API Layer<br/>pkg/dot/]:::apiLayer
    
    Pipeline[Pipeline Layer<br/>internal/pipeline/]:::middlewareLayer
    Executor[Executor Layer<br/>internal/executor/]:::middlewareLayer
    
    Scanner[Scanner<br/>internal/scanner/]:::coreLayer
    Planner[Planner<br/>internal/planner/]:::coreLayer
    Ignore[Ignore<br/>internal/ignore/]:::coreLayer
    
    Domain[Domain Layer<br/>internal/domain/<br/><br/>No Dependencies<br/>Standard Library Only]:::domainLayer
    
    Adapters[Adapters<br/>internal/adapters/]:::adapterLayer
    DomainPorts[Domain Ports<br/>Interfaces: FS, Logger,<br/>Tracer, Metrics]:::portsLayer
    
    CLI -->|depends on| API
    API -->|depends on| Pipeline
    API -->|depends on| Executor
    API -->|depends on| Scanner
    API -->|depends on| Planner
    API -->|depends on| Ignore
    
    Pipeline -->|depends on| Scanner
    Pipeline -->|depends on| Planner
    Pipeline -->|depends on| Ignore
    Pipeline -->|depends on| Domain
    
    Executor -->|depends on| Domain
    
    Scanner -->|depends on| Domain
    Planner -->|depends on| Domain
    Ignore -->|depends on| Domain
    
    Adapters -->|implements| DomainPorts
    DomainPorts -.defined in.-> Domain
    
    style CLI fill:#4A90E2,stroke:#2C5F8D,color:#fff,stroke-width:3px
    style API fill:#50C878,stroke:#2D7A4A,color:#fff,stroke-width:3px
    style Pipeline fill:#9B59B6,stroke:#6C3A7C,color:#fff,stroke-width:2px
    style Executor fill:#E67E22,stroke:#A84E0F,color:#fff,stroke-width:2px
    style Scanner fill:#3498DB,stroke:#1F618D,color:#fff,stroke-width:2px
    style Planner fill:#3498DB,stroke:#1F618D,color:#fff,stroke-width:2px
    style Ignore fill:#3498DB,stroke:#1F618D,color:#fff,stroke-width:2px
    style Domain fill:#2ECC71,stroke:#1E8449,color:#fff,stroke-width:4px
    style Adapters fill:#95A5A6,stroke:#5D6D7E,color:#fff,stroke-width:2px
    style DomainPorts fill:#7F8C8D,stroke:#5D6D7E,color:#fff,stroke-width:2px,stroke-dasharray: 5 5
    
    classDef cliLayer stroke-width:3px
    classDef apiLayer stroke-width:3px
    classDef middlewareLayer stroke-width:2px
    classDef coreLayer stroke-width:2px
    classDef domainLayer stroke-width:4px
    classDef adapterLayer stroke-width:2px,stroke-dasharray: 5 5
    classDef portsLayer stroke-width:2px,stroke-dasharray: 5 5
```

**Rules**:
1. Domain layer has no dependencies (except standard library)
2. Core layer depends only on domain
3. Pipeline and Executor depend on domain and core
4. API layer depends on all internal layers
5. CLI layer depends on the API layer plus presentation helpers under `internal/cli/`; it imports no other internal package
6. Adapters depend only on domain ports

### Import Restrictions

**Prohibited**:
- Non-presentation internal packages (`domain`, `scanner`, `planner`, `pipeline`, `executor`, `manifest`, `config`, `adapters`) importing from `pkg/dot`, which would create a cycle
- Domain importing from infrastructure packages

**Permitted exception**: presentation-only packages under `internal/cli/` may import `pkg/dot` to render its public types. `pkg/dot` never imports `internal/cli`, so no cycle exists.

**Required**:
- All internal packages import from `internal/domain` for types
- API layer re-exports domain types for public consumption
- Type aliases in `pkg/dot` for stable public API

### Adapter Independence

Adapters are swappable implementations:

```go
// Production
cfg := dot.Config{
    FS:     adapters.NewOSFilesystem(),
    Logger: adapters.NewSlogLogger(slog.Default()),
}

// Testing
cfg := dot.Config{
    FS:     adapters.NewMemFS(),
    Logger: adapters.NewNoopLogger(),
}

// Dry-run (no special filesystem; the flag is threaded into each service)
cfg := dot.Config{
    FS:     adapters.NewOSFilesystem(),
    Logger: adapters.NewSlogLogger(slog.Default()),
    DryRun: true,
}
```

## Performance Characteristics

### Time Complexity

**Scanning**: O(n) where n is number of files in packages
**Planning**: O(m + e) where m is operations and e is dependency edges
**Topological Sort**: O(m + e) using depth-first search
**Execution (sequential)**: O(m) where m is number of operations
**Execution (parallel)**: O(b) where b is number of batches

### Space Complexity

**Manifest Storage**: O(p × l) where p is packages and l is links per package
**Dependency Graph**: O(m + e) where m is operations and e is edges
**Checkpoint**: O(m) to store operation state

### Optimizations

1. **Directory Folding**: Reduce symlink count when entire directory owned by package
2. **Incremental Updates**: Use content hashing to detect changed packages
3. **Parallel Execution**: Execute independent operations concurrently
4. **Lazy Loading**: Load manifests on demand
5. **Efficient Scanning**: Skip ignored directories early in traversal

## Extensibility Points

### Custom Filesystem Implementations

Implement `dot.FS` (aliased from `internal/domain` in pkg/dot/config.go) for
cloud storage backends, VCS integration, or virtual filesystems.

### Conflict Resolution

`planner.ResolutionPolicy` is a closed enumeration (`PolicyFail`, `PolicyBackup`,
`PolicyOverwrite`, `PolicySkip`) in `internal/planner/policies.go:12`, selected
through `Config.Backup` and `Config.Overwrite`. It is not an extension point:
adding a policy requires a change to `internal/planner`.

### Metrics and Observability

Implement `dot.Logger`, `dot.Tracer`, and `dot.Metrics` for Prometheus, StatsD,
OpenTelemetry, or custom telemetry. All four infrastructure ports are injected
through `Config` (pkg/dot/config.go:93-96).

Output rendering is not an extension point: `internal/cli/renderer` is internal
to the CLI. External consumers format `pkg/dot` result types themselves.

## Security Considerations

### Path Traversal Prevention

- All paths validated before use
- Phantom types prevent path confusion
- Relative paths resolved before operations
- Symlink targets validated

### Deletion Safety

`LinkDelete` records the symlink destination observed at plan time
(`NewLinkDeleteWithDestination`, internal/domain/operation.go:183) and re-checks
it with `Lstat` before deleting. A path that has become a regular file, or a
symlink now pointing elsewhere, is refused with `ErrUnsafeDelete`
(internal/domain/errors.go:130). This prevents a stale plan from destroying
user data that changed between planning and execution.

### Rollback

- Checkpoint created before operations
- On failure, executed operations are reverted in reverse order
- Rollback runs on a context detached from cancellation, so Ctrl-C mid-plan still restores state
- Operations that cannot be reverted (notably `DirRemoveAll`) return `ErrRollbackImpossible`; the executor counts them in `ExecutionResult.RollbackFailed` and `ErrExecutionFailed` reports "N operations could not be rolled back"
- Rollback is therefore best-effort, not a hard atomicity guarantee: callers must read the failure count rather than assume clean state

### Manifest Integrity

- Manifest stored in target directory (user-controlled)
- Content hashing for change detection
- Validation before loading

### Error Information Disclosure

- Error messages avoid exposing sensitive paths
- Detailed errors logged but sanitized for display
- Security-relevant errors handled specially

## Future Architecture Considerations

### Potential Enhancements

1. **Distributed Locking**: Support for network filesystem coordination
2. **Incremental Manifest Updates**: Avoid full manifest rewrite
3. **Plugin System**: External conflict resolution strategies
4. **Remote Package Sources**: Support for fetching packages from URLs
5. **Advanced Caching**: Cache scanning results for large repositories

### Backward Compatibility

The architecture supports evolution while maintaining compatibility:

- Public API in `pkg/dot` with type aliases
- Internal implementation can change freely
- Manifest versioning for format changes
- Deprecation warnings for API changes

## References

### Related Documentation

- [User Guide](../user/index.md) - End-user documentation
- [Contributing Guide](../../CONTRIBUTING.md) - Development guidelines
- [Release Workflow](release-workflow.md) - Release process

### External Resources

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Functional Core, Imperative Shell](https://www.destroyallsoftware.com/screencasts/catalog/functional-core-imperative-shell)
- [Go Module Documentation](https://golang.org/ref/mod)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

