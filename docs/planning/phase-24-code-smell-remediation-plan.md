# Phase 24: Code Smell Remediation and Architectural Improvements

## Overview

Systematic refactoring to address identified code smells while maintaining backward compatibility, test coverage, and adherence to project constitutional principles. This phase focuses on improving maintainability, reducing complexity, and enhancing code quality without introducing breaking changes.

## Objectives

1. Decompose god objects into focused, single-responsibility components
2. Reduce configuration complexity and improve testability
3. Simplify Result type usage patterns
4. Eliminate code duplication and extract reusable abstractions
5. Reduce cyclomatic complexity in long methods
6. Establish constants for magic values
7. Improve error handling patterns
8. Clean up technical debt (TODOs, unused code)

## Constitutional Compliance

All work must adhere to:
- Test-First Development (TDD): Write tests before implementation
- Atomic Commits: One logical change per commit with conventional commit messages
- 80% Coverage Minimum: Maintain or improve test coverage
- No Breaking Changes: Internal refactoring only, preserve public API
- All Linters Pass: Zero tolerance for linting errors
- Academic Documentation: Factual, technical documentation style

## Phase Structure

This phase is divided into 8 sub-phases to enable incremental progress with validation at each step:

- **Phase 24.1**: Constants and Magic Values
- **Phase 24.2**: Error Handling Patterns
- **Phase 24.3**: Configuration Simplification
- **Phase 24.4**: Path Validation Consolidation
- **Phase 24.5**: Client Decomposition
- **Phase 24.6**: Long Method Refactoring
- **Phase 24.7**: Result Type Helpers
- **Phase 24.8**: Technical Debt Cleanup

---

## Phase 24.1: Constants and Magic Values

### Problem
Magic numbers (file permissions) and strings (config keys) scattered throughout codebase reduce maintainability and increase error risk.

### Objectives
1. Extract all magic values to named constants
2. Create constant packages for different domains
3. Update all usages to reference constants
4. Document constant meanings and rationale

### Tasks

#### Task 1.1: File Permission Constants
**File:** `internal/domain/permissions.go`

Create constants for file permissions:
```go
package domain

const (
    // PermUserRW is user read/write only (0600)
    PermUserRW = 0600
    
    // PermUserRWX is user read/write/execute only (0700)
    PermUserRWX = 0700
    
    // PermUserW is user write bit (0200)
    PermUserW = 0200
    
    // PermGroupWorld is group/world readable (0044)
    PermGroupWorld = 0044
)
```

**Test File:** `internal/domain/permissions_test.go`
- Verify constants have expected values
- Document permission bit patterns

**Files to Update:**
- `internal/config/writer.go` (lines 31, 42)
- `internal/executor/executor.go` (line 164)
- `pkg/dot/client.go` (security checks)

**Commits:**
1. `test(domain): add permission constant tests`
2. `feat(domain): add file permission constants`
3. `refactor(config): use permission constants`
4. `refactor(executor): use permission constants`

#### Task 1.2: Configuration Key Constants
**File:** `internal/config/keys.go`

Create constants for configuration keys:
```go
package config

const (
    // Directory keys
    KeyDirPackage  = "directories.package"
    KeyDirTarget   = "directories.target"
    KeyDirManifest = "directories.manifest"
    
    // Logging keys
    KeyLogLevel       = "logging.level"
    KeyLogFormat      = "logging.format"
    KeyLogDestination = "logging.destination"
    KeyLogFile        = "logging.file"
    
    // ... additional keys
)
```

**Test File:** `internal/config/keys_test.go`
- Verify key format consistency
- Test key parsing and validation

**Files to Update:**
- `internal/config/loader.go`
- `internal/config/writer.go`
- `cmd/dot/config.go`

**Commits:**
1. `test(config): add configuration key constant tests`
2. `feat(config): add configuration key constants`
3. `refactor(config): use key constants throughout`

#### Task 1.3: Default Value Constants
**File:** `internal/config/defaults.go`

Extract default values:
```go
package config

const (
    DefaultLogLevel         = "INFO"
    DefaultOutputFormat     = "text"
    DefaultBackupSuffix     = ".bak"
    DefaultDotfilePrefix    = "dot-"
    DefaultMaxDepth         = 0  // Unlimited
    DefaultVerbosity        = 1  // Normal
)
```

**Commits:**
1. `test(config): add default value constant tests`
2. `feat(config): extract default value constants`
3. `refactor(config): use default constants`

### Success Criteria
- [ ] All magic numbers replaced with named constants
- [ ] All magic strings replaced with named constants
- [ ] Constants have comprehensive documentation
- [ ] Test coverage maintained at 80%+
- [ ] All linters pass

### Estimated Effort
**3-4 hours** (TDD adds overhead but ensures correctness)

---

## Phase 24.2: Error Handling Patterns

### Problem
Repetitive error handling patterns (88 instances of `if err != nil`) add boilerplate and reduce readability.

### Objectives
1. Create error handling helper functions
2. Establish consistent error wrapping patterns
3. Reduce boilerplate while maintaining explicit error handling

### Tasks

#### Task 2.1: Error Context Helpers
**File:** `internal/domain/errors_helpers.go`

Create error wrapping helpers:
```go
package domain

// WrapError wraps an error with contextual message
func WrapError(err error, context string) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorf wraps an error with formatted context
func WrapErrorf(err error, format string, args ...interface{}) error {
    if err == nil {
        return nil
    }
    msg := fmt.Sprintf(format, args...)
    return fmt.Errorf("%s: %w", msg, err)
}
```

**Test File:** `internal/domain/errors_helpers_test.go`
- Test nil error handling
- Test context wrapping
- Test error chain preservation

**Commits:**
1. `test(domain): add error helper tests`
2. `feat(domain): add error wrapping helpers`
3. `refactor: use error helpers in high-frequency areas`

#### Task 2.2: Result Type Validation Helper
**File:** `internal/domain/result_helpers.go`

Add helper for common Result validation pattern:
```go
package domain

// UnwrapResult extracts value or returns error with context
func UnwrapResult[T any](r Result[T], context string) (T, error) {
    if !r.IsOk() {
        return *new(T), WrapError(r.UnwrapErr(), context)
    }
    return r.Unwrap(), nil
}
```

**Test File:** `internal/domain/result_helpers_test.go`

**Commits:**
1. `test(domain): add result helper tests`
2. `feat(domain): add result unwrap helper`
3. `refactor: simplify result unwrapping patterns`

### Success Criteria
- [ ] Error helpers created and tested
- [ ] Boilerplate reduced by ~30% in target files
- [ ] Error chains preserved correctly
- [ ] Test coverage maintained

### Estimated Effort
**2-3 hours**

---

## Phase 24.3: Configuration Simplification

### Problem
Configuration complexity: `writer.go` (638 lines), `loader.go` (528 lines), `extended.go` (424 lines) with 10 nested config types.

### Objectives
1. Extract configuration marshaling logic into separate strategies
2. Simplify loader with functional options
3. Reduce ExtendedConfig surface area
4. Improve testability through dependency injection

### Tasks

#### Task 3.1: Configuration Marshal Strategy Pattern
**File:** `internal/config/marshal/strategy.go`

Extract marshaling to strategy pattern:
```go
package marshal

type Strategy interface {
    Marshal(cfg *config.ExtendedConfig, opts MarshalOptions) ([]byte, error)
    Unmarshal(data []byte) (*config.ExtendedConfig, error)
}

type MarshalOptions struct {
    IncludeComments bool
    Indent          int
}
```

**Implementation Files:**
- `internal/config/marshal/yaml_strategy.go`
- `internal/config/marshal/json_strategy.go`
- `internal/config/marshal/toml_strategy.go`

**Test Files:**
- `internal/config/marshal/strategy_test.go`
- `internal/config/marshal/yaml_strategy_test.go`
- `internal/config/marshal/json_strategy_test.go`
- `internal/config/marshal/toml_strategy_test.go`

**Benefits:**
- Reduces `writer.go` from 638 to ~200 lines
- Each strategy is independently testable
- Clear separation of concerns

**Commits:**
1. `test(config): add marshal strategy interface tests`
2. `feat(config): add marshal strategy interface`
3. `test(config): add YAML strategy tests`
4. `feat(config): implement YAML marshal strategy`
5. `test(config): add JSON strategy tests`
6. `feat(config): implement JSON marshal strategy`
7. `test(config): add TOML strategy tests`
8. `feat(config): implement TOML marshal strategy`
9. `refactor(config): migrate writer to use strategies`

#### Task 3.2: Configuration Group Consolidation
**File:** `internal/config/groups.go`

Group related configurations:
```go
package config

// CoreConfig contains essential runtime configuration
type CoreConfig struct {
    Directories DirectoriesConfig
    Logging     LoggingConfig
}

// BehaviorConfig contains operational behavior settings
type BehaviorConfig struct {
    Symlinks   SymlinksConfig
    Operations OperationsConfig
    Dotfile    DotfileConfig
}

// UIConfig contains user interface settings
type UIConfig struct {
    Output   OutputConfig
    Packages PackagesConfig
}

// DiagnosticConfig contains health check settings
type DiagnosticConfig struct {
    Doctor DoctorConfig
}
```

**Migration Path:**
1. Create new grouped structures alongside existing
2. Add conversion functions
3. Update internal usage incrementally
4. Deprecate old structure in future phase (no breaking changes yet)

**Commits:**
1. `test(config): add config group tests`
2. `feat(config): add config grouping structures`
3. `feat(config): add group conversion functions`
4. `refactor(config): use config groups internally`

#### Task 3.3: Loader Functional Options
**File:** `internal/config/loader_options.go`

Simplify loader with functional options:
```go
package config

type LoaderOption func(*Loader)

func WithEnvOverrides() LoaderOption {
    return func(l *Loader) {
        l.enableEnv = true
    }
}

func WithFlagOverrides(flags map[string]interface{}) LoaderOption {
    return func(l *Loader) {
        l.flags = flags
    }
}

// Usage:
loader := NewLoader(appName, configPath, 
    WithEnvOverrides(),
    WithFlagOverrides(flags))
cfg, err := loader.Load()
```

**Benefits:**
- Reduces loader complexity
- Clearer intent
- Easier testing of individual options

**Commits:**
1. `test(config): add loader option tests`
2. `feat(config): add functional options for loader`
3. `refactor(config): migrate loader to functional options`

### Success Criteria
- [ ] `writer.go` reduced to <300 lines
- [ ] `loader.go` reduced to <300 lines
- [ ] Each strategy independently tested
- [ ] Configuration grouping improves clarity
- [ ] No breaking changes to public API
- [ ] Test coverage maintained

### Estimated Effort
**8-10 hours** (most complex refactoring)

---

## Phase 24.4: Path Validation Consolidation

### Problem
Path validation logic duplicated across `internal/domain/path.go`, `pkg/dot/path.go`, `internal/manifest/validate.go`.

### Objectives
1. Consolidate path validation in domain layer
2. Remove duplication from pkg and manifest layers
3. Establish single source of truth for path operations

### Tasks

#### Task 4.1: Domain Path Validators
**File:** `internal/domain/path_validators.go`

Extract common validation logic:
```go
package domain

type PathValidator interface {
    Validate(path string) error
}

type AbsolutePathValidator struct{}

func (v *AbsolutePathValidator) Validate(path string) error {
    if !filepath.IsAbs(path) {
        return ErrInvalidPath{Path: path, Reason: "path must be absolute"}
    }
    return nil
}

type RelativePathValidator struct{}

func (v *RelativePathValidator) Validate(path string) error {
    if filepath.IsAbs(path) {
        return ErrInvalidPath{Path: path, Reason: "path must be relative"}
    }
    return nil
}

type TraversalFreeValidator struct{}

func (v *TraversalFreeValidator) Validate(path string) error {
    cleaned := filepath.Clean(path)
    if cleaned != path || strings.Contains(path, "..") {
        return ErrInvalidPath{Path: path, Reason: "path contains traversal sequences"}
    }
    return nil
}
```

**Test File:** `internal/domain/path_validators_test.go`

**Commits:**
1. `test(domain): add path validator tests`
2. `feat(domain): add path validator interfaces`
3. `feat(domain): implement path validators`

#### Task 4.2: Update Path Constructors
**File:** `internal/domain/path.go`

Use validators in path constructors:
```go
func NewPackagePath(s string) Result[PackagePath] {
    validators := []PathValidator{
        &AbsolutePathValidator{},
        &TraversalFreeValidator{},
    }
    
    for _, v := range validators {
        if err := v.Validate(s); err != nil {
            return Err[PackagePath](err)
        }
    }
    
    return Ok(Path[PackageDirKind]{path: clean(s)})
}
```

**Commits:**
1. `refactor(domain): use validators in path constructors`

#### Task 4.3: Remove Duplication
**Files to Update:**
- `pkg/dot/path.go` - Use domain validators
- `internal/manifest/validate.go` - Use domain validators

**Commits:**
1. `refactor(pkg): use domain path validators`
2. `refactor(manifest): use domain path validators`

### Success Criteria
- [ ] Path validation in single location
- [ ] Duplication eliminated
- [ ] All path validation tests pass
- [ ] Test coverage maintained

### Estimated Effort
**3-4 hours**

---

## Phase 24.5: Client Decomposition

### Problem
`pkg/dot/client.go` is a god object at 1,020 lines with 19 methods combining all operations.

### Objectives
1. Extract operations into separate service types
2. Maintain backward compatibility through Client facade
3. Improve testability of individual operations
4. Reduce cognitive load per file

### Tasks

#### Task 5.1: Extract Manage Service
**File:** `pkg/dot/manage_service.go`

Extract manage-related operations:
```go
package dot

type ManageService struct {
    fs           FS
    logger       Logger
    tracer       Tracer
    managePipe   *pipeline.ManagePipeline
    executor     *executor.Executor
    manifestSvc  *ManifestService
    packageDir   string
    targetDir    string
    backupDir    string
    dryRun       bool
}

func newManageService(cfg Config, exec *executor.Executor, manifestSvc *ManifestService, pipe *pipeline.ManagePipeline) *ManageService {
    return &ManageService{
        fs:          cfg.FS,
        logger:      cfg.Logger,
        tracer:      cfg.Tracer,
        managePipe:  pipe,
        executor:    exec,
        manifestSvc: manifestSvc,
        packageDir:  cfg.PackageDir,
        targetDir:   cfg.TargetDir,
        backupDir:   cfg.BackupDir,
        dryRun:      cfg.DryRun,
    }
}

func (s *ManageService) Manage(ctx context.Context, packages ...string) error {
    // Current Client.Manage logic
}

func (s *ManageService) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
    // Current Client.PlanManage logic
}

func (s *ManageService) Remanage(ctx context.Context, packages ...string) error {
    // Current Client.Remanage logic
}

func (s *ManageService) PlanRemanage(ctx context.Context, packages ...string) (Plan, error) {
    // Current Client.PlanRemanage logic
}
```

**Test File:** `pkg/dot/manage_service_test.go`
- Move manage-related tests from `client_test.go`
- Add service-specific tests

**Commits:**
1. `test(pkg): add ManageService tests`
2. `feat(pkg): extract ManageService from Client`
3. `refactor(pkg): Client delegates to ManageService`

#### Task 5.2: Extract Unmanage Service
**File:** `pkg/dot/unmanage_service.go`

Extract unmanage operations:
```go
package dot

type UnmanageService struct {
    fs          FS
    logger      Logger
    executor    *executor.Executor
    manifestSvc *ManifestService
    targetDir   string
    dryRun      bool
}

func (s *UnmanageService) Unmanage(ctx context.Context, packages ...string) error {
    // Current Client.Unmanage logic
}

func (s *UnmanageService) PlanUnmanage(ctx context.Context, packages ...string) (Plan, error) {
    // Current Client.PlanUnmanage logic
}
```

**Test File:** `pkg/dot/unmanage_service_test.go`

**Commits:**
1. `test(pkg): add UnmanageService tests`
2. `feat(pkg): extract UnmanageService from Client`
3. `refactor(pkg): Client delegates to UnmanageService`

#### Task 5.3: Extract Status Service
**File:** `pkg/dot/status_service.go`

Extract status operations:
```go
package dot

type StatusService struct {
    manifestSvc *ManifestService
    targetDir   string
}

func (s *StatusService) Status(ctx context.Context, packages ...string) (Status, error) {
    // Current Client.Status logic
}

func (s *StatusService) List(ctx context.Context) ([]PackageInfo, error) {
    // Current Client.List logic
}
```

**Test File:** `pkg/dot/status_service_test.go`

**Commits:**
1. `test(pkg): add StatusService tests`
2. `feat(pkg): extract StatusService from Client`
3. `refactor(pkg): Client delegates to StatusService`

#### Task 5.4: Extract Doctor Service
**File:** `pkg/dot/doctor_service.go`

Extract doctor operations:
```go
package dot

type DoctorService struct {
    fs          FS
    logger      Logger
    manifestSvc *ManifestService
    targetDir   string
}

func (s *DoctorService) Doctor(ctx context.Context) (DiagnosticReport, error) {
    return s.DoctorWithScan(ctx, DefaultScanConfig())
}

func (s *DoctorService) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
    // Current Client.DoctorWithScan logic
}
```

**Test File:** `pkg/dot/doctor_service_test.go`

**Commits:**
1. `test(pkg): add DoctorService tests`
2. `feat(pkg): extract DoctorService from Client`
3. `refactor(pkg): Client delegates to DoctorService`

#### Task 5.5: Extract Adopt Service
**File:** `pkg/dot/adopt_service.go`

Extract adopt operations:
```go
package dot

type AdoptService struct {
    fs          FS
    logger      Logger
    executor    *executor.Executor
    manifestSvc *ManifestService
    packageDir  string
    targetDir   string
    dryRun      bool
}

func (s *AdoptService) Adopt(ctx context.Context, files []string, pkg string) error {
    // Current Client.Adopt logic
}

func (s *AdoptService) PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error) {
    // Current Client.PlanAdopt logic
}
```

**Test File:** `pkg/dot/adopt_service_test.go`

**Commits:**
1. `test(pkg): add AdoptService tests`
2. `feat(pkg): extract AdoptService from Client`
3. `refactor(pkg): Client delegates to AdoptService`

#### Task 5.6: Extract Manifest Service
**File:** `pkg/dot/manifest_service.go`

Centralize manifest operations:
```go
package dot

type ManifestService struct {
    fs       FS
    logger   Logger
    store    manifest.ManifestStore
}

func (s *ManifestService) Load(ctx context.Context, targetPath TargetPath) (manifest.Manifest, error) {
    // Manifest loading logic
}

func (s *ManifestService) Save(ctx context.Context, targetPath TargetPath, m manifest.Manifest) error {
    // Manifest saving logic
}

func (s *ManifestService) Update(ctx context.Context, targetPath TargetPath, packages []string, plan Plan) error {
    // Current updateManifest logic
}
```

**Test File:** `pkg/dot/manifest_service_test.go`

**Commits:**
1. `test(pkg): add ManifestService tests`
2. `feat(pkg): extract ManifestService`
3. `refactor(pkg): services use ManifestService`

#### Task 5.7: Refactor Client to Facade
**File:** `pkg/dot/client.go` (reduced to ~150 lines)

Client becomes a facade delegating to services:
```go
package dot

type Client struct {
    config      Config
    manageSvc   *ManageService
    unmanageSvc *UnmanageService
    statusSvc   *StatusService
    doctorSvc   *DoctorService
    adoptSvc    *AdoptService
}

func NewClient(cfg Config) (*Client, error) {
    // Validation
    // Create services
    
    return &Client{
        config:      cfg,
        manageSvc:   manageSvc,
        unmanageSvc: unmanageSvc,
        statusSvc:   statusSvc,
        doctorSvc:   doctorSvc,
        adoptSvc:    adoptSvc,
    }, nil
}

// Delegation methods
func (c *Client) Manage(ctx context.Context, packages ...string) error {
    return c.manageSvc.Manage(ctx, packages...)
}

func (c *Client) PlanManage(ctx context.Context, packages ...string) (Plan, error) {
    return c.manageSvc.PlanManage(ctx, packages...)
}

// ... other delegation methods
```

**Test File:** `pkg/dot/client_test.go` (simplified)

**Commits:**
1. `refactor(pkg): convert Client to facade pattern`

### Benefits
- Client reduced from 1,020 to ~150 lines
- Each service is ~150-200 lines, focused on single responsibility
- Services independently testable
- Clear boundaries between operations
- Easier to extend with new operations

### Success Criteria
- [ ] Client file <200 lines
- [ ] Each service file <250 lines
- [ ] All existing tests pass
- [ ] No breaking changes to public API
- [ ] Test coverage maintained
- [ ] Each service has comprehensive tests

### Estimated Effort
**12-15 hours** (most time-consuming refactoring)

---

## Phase 24.6: Long Method Refactoring

### Problem
Several methods exceed 50 lines with high cyclomatic complexity.

### Objectives
1. Break long methods into smaller, focused functions
2. Reduce cyclomatic complexity to <15
3. Improve readability and testability

### Tasks

#### Task 6.1: Refactor performOrphanScan
**File:** `pkg/dot/doctor_service.go` (after Phase 24.5)

Current method: 43 lines, multiple responsibilities

Extract to:
```go
func (s *DoctorService) performOrphanScan(ctx context.Context, m *manifest.Manifest, scanCfg ScanConfig, issues *[]Issue, stats *DiagnosticStats) {
    scanDirs := s.determineScanDirectories(m, scanCfg)
    rootDirs := s.normalizeAndDeduplicateDirs(scanDirs, scanCfg.Mode)
    linkSet := buildManagedLinkSet(m)
    
    for _, dir := range rootDirs {
        s.scanDirectory(ctx, dir, m, linkSet, scanCfg, issues, stats)
    }
}

func (s *DoctorService) determineScanDirectories(m *manifest.Manifest, scanCfg ScanConfig) []string {
    // Lines 702-712
}

func (s *DoctorService) normalizeAndDeduplicateDirs(dirs []string, mode ScanMode) []string {
    // Lines 713-724
}
```

**Test Files:**
- Test each extracted method independently
- Table-driven tests for edge cases

**Commits:**
1. `test(pkg): add tests for scan directory helpers`
2. `refactor(pkg): extract determineScanDirectories`
3. `refactor(pkg): extract normalizeAndDeduplicateDirs`
4. `refactor(pkg): simplify performOrphanScan`

#### Task 6.2: Refactor DoctorWithScan
**File:** `pkg/dot/doctor_service.go`

Current method: 60+ lines

Extract to:
```go
func (s *DoctorService) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
    m, issues, err := s.loadManifestOrDefault(ctx)
    if err != nil {
        return DiagnosticReport{}, err
    }
    
    stats := s.checkManagedPackages(ctx, m, &issues)
    
    if scanCfg.Mode != ScanOff {
        s.performOrphanScan(ctx, m, scanCfg, &issues, &stats)
    }
    
    health := s.determineOverallHealth(issues)
    
    return DiagnosticReport{
        OverallHealth: health,
        Issues:        issues,
        Statistics:    stats,
    }, nil
}

func (s *DoctorService) loadManifestOrDefault(ctx context.Context) (*manifest.Manifest, []Issue, error) {
    // Lines 637-660
}

func (s *DoctorService) checkManagedPackages(ctx context.Context, m *manifest.Manifest, issues *[]Issue) DiagnosticStats {
    // Lines 662-670
}

func (s *DoctorService) determineOverallHealth(issues []Issue) Health {
    // Lines 676-685
}
```

**Commits:**
1. `test(pkg): add tests for doctor helpers`
2. `refactor(pkg): extract loadManifestOrDefault`
3. `refactor(pkg): extract checkManagedPackages`
4. `refactor(pkg): extract determineOverallHealth`
5. `refactor(pkg): simplify DoctorWithScan`

#### Task 6.3: Refactor scanForOrphanedLinks
**File:** `pkg/dot/doctor_service.go`

Current method: 48 lines with deep nesting

Extract to:
```go
func (s *DoctorService) scanForOrphanedLinks(ctx context.Context, dir string, m *manifest.Manifest, linkSet map[string]bool, scanCfg ScanConfig, issues *[]Issue, stats *DiagnosticStats) error {
    entries, err := s.fs.ReadDir(ctx, dir)
    if err != nil {
        return err
    }
    
    for _, entry := range entries {
        if ctx.Err() != nil {
            return ctx.Err()
        }
        
        fullPath := filepath.Join(dir, entry.Name())
        
        if s.shouldSkipEntry(entry, fullPath) {
            continue
        }
        
        if entry.IsDir() {
            s.scanDirectoryRecursive(ctx, fullPath, m, linkSet, scanCfg, issues, stats)
        } else {
            s.checkForOrphanedLink(ctx, fullPath, linkSet, issues, stats)
        }
    }
    
    return nil
}

func (s *DoctorService) shouldSkipEntry(entry fs.DirEntry, fullPath string) bool {
    // Skip manifest file
}

func (s *DoctorService) checkForOrphanedLink(ctx context.Context, fullPath string, linkSet map[string]bool, issues *[]Issue, stats *DiagnosticStats) {
    // Lines 905-927
}
```

**Commits:**
1. `test(pkg): add tests for scan helpers`
2. `refactor(pkg): extract shouldSkipEntry`
3. `refactor(pkg): extract checkForOrphanedLink`
4. `refactor(pkg): simplify scanForOrphanedLinks`

#### Task 6.4: Refactor extractManagedDirectories
**File:** `pkg/dot/doctor_service.go`

Current method: nested loops

Simplify with set operations:
```go
func extractManagedDirectories(m *manifest.Manifest) []string {
    dirSet := make(map[string]bool)
    
    for _, pkgInfo := range m.Packages {
        for _, link := range pkgInfo.Links {
            s.addParentDirectories(link, dirSet)
        }
    }
    
    return setToSlice(dirSet)
}

func (s *DoctorService) addParentDirectories(link string, dirSet map[string]bool) {
    // Extracted parent directory logic
}
```

**Commits:**
1. `test(pkg): add tests for directory extraction`
2. `refactor(pkg): extract addParentDirectories`
3. `refactor(pkg): simplify extractManagedDirectories`

### Success Criteria
- [ ] No methods exceed 50 lines
- [ ] Cyclomatic complexity <15 for all methods
- [ ] Extracted helpers have dedicated tests
- [ ] All existing tests pass
- [ ] Test coverage maintained

### Estimated Effort
**6-8 hours**

---

## Phase 24.7: Result Type Helpers

### Problem
411 instances of Result unwrapping with repetitive `IsOk()` checks add boilerplate.

### Objectives
1. Add convenience methods to Result type
2. Simplify common unwrapping patterns
3. Maintain explicit error handling

### Tasks

#### Task 7.1: Add Result Methods
**File:** `internal/domain/result.go`

Add helper methods:
```go
// OrElse returns the contained value or executes fallback function
func (r Result[T]) OrElse(fn func() T) T {
    if r.IsOk() {
        return r.Unwrap()
    }
    return fn()
}

// OrDefault returns the contained value or zero value
func (r Result[T]) OrDefault() T {
    if r.IsOk() {
        return r.Unwrap()
    }
    return *new(T)
}

// AndThen chains operations on Ok values
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
    if !r.IsOk() {
        return Err[U](r.UnwrapErr())
    }
    return fn(r.Unwrap())
}

// Map transforms Ok values
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
    if !r.IsOk() {
        return Err[U](r.UnwrapErr())
    }
    return Ok(fn(r.Unwrap()))
}
```

**Test File:** `internal/domain/result_test.go`

**Commits:**
1. `test(domain): add tests for Result helper methods`
2. `feat(domain): add Result helper methods`

#### Task 7.2: Apply Result Helpers
**Target Files:**
- `pkg/dot/client.go` (after Phase 24.5 decomposition)
- `internal/pipeline/stages.go`
- `internal/planner/desired.go`

Example transformation:
```go
// Before
packagePathResult := NewPackagePath(c.config.PackageDir)
if !packagePathResult.IsOk() {
    return Plan{}, fmt.Errorf("invalid package directory: %w", packagePathResult.UnwrapErr())
}
packagePath := packagePathResult.Unwrap()

// After
packagePath, err := UnwrapResult(NewPackagePath(c.config.PackageDir), "invalid package directory")
if err != nil {
    return Plan{}, err
}
```

**Commits:**
1. `refactor(pkg): use Result helpers in services`
2. `refactor(pipeline): use Result helpers`
3. `refactor(planner): use Result helpers`

### Success Criteria
- [ ] Result type has chainable helpers
- [ ] Boilerplate reduced by ~20-30%
- [ ] Error handling remains explicit
- [ ] All tests pass
- [ ] Test coverage maintained

### Estimated Effort
**3-4 hours**

---

## Phase 24.8: Technical Debt Cleanup

### Problem
TODOs in documentation, unused panic in production code paths, incomplete cleanup.

### Objectives
1. Remove or resolve all TODO comments
2. Move test-only panics to test files
3. Clean up obsolete documentation
4. Update architecture documentation

### Tasks

#### Task 8.1: TODO Audit and Resolution
**Files to Audit:**
- `docs/planning/phase-22-complete-stubs-plan.md`
- `docs/planning/completed/phase-22-complete.md`
- `docs/planning/phase-15c-plan.md`

**Actions:**
1. Archive planning documents that are complete
2. Convert remaining TODOs to GitHub issues
3. Remove obsolete TODOs

**Commits:**
1. `docs(planning): archive completed phase documents`
2. `docs: remove obsolete TODO comments`
3. `chore: create issues for remaining TODOs`

#### Task 8.2: Move Test-Only Code
**File:** `internal/domain/path.go`

Move `MustParsePath` to test helper:
```go
// Remove from path.go

// Add to internal/domain/testhelpers.go or internal/domain/path_testing.go
package domain

// MustParsePath creates a FilePath from a string, panicking on error.
// This function is intended for use in tests only.
func MustParsePath(s string) FilePath {
    result := NewFilePath(s)
    if result.IsErr() {
        panic(result.UnwrapErr())
    }
    return result.Unwrap()
}
```

**Files to Update:**
- All test files using `MustParsePath`

**Commits:**
1. `test(domain): extract test helpers`
2. `refactor(domain): move MustParsePath to test helpers`
3. `refactor(tests): update MustParsePath imports`

#### Task 8.3: Documentation Updates
**Files to Update:**
- `docs/architecture/architecture.md` - Reflect service decomposition
- `docs/reference/features.md` - Update technical details
- `README.md` - Ensure accuracy

**Commits:**
1. `docs(architecture): update for service decomposition`
2. `docs(reference): update technical implementation details`
3. `docs: update README with current architecture`

#### Task 8.4: Code Coverage Report
Generate and document current coverage:

```bash
make test-coverage
go tool cover -html=coverage.out -o coverage.html
```

**Commits:**
1. `docs(testing): add coverage report generation docs`

### Success Criteria
- [ ] No TODOs in production code
- [ ] All planning TODOs resolved or tracked as issues
- [ ] Test-only code in test files
- [ ] Documentation reflects current architecture
- [ ] Coverage report generated and documented

### Estimated Effort
**2-3 hours**

---

## Overall Phase Summary

### Total Estimated Effort
- Phase 24.1: 3-4 hours
- Phase 24.2: 2-3 hours
- Phase 24.3: 8-10 hours
- Phase 24.4: 3-4 hours
- Phase 24.5: 12-15 hours
- Phase 24.6: 6-8 hours
- Phase 24.7: 3-4 hours
- Phase 24.8: 2-3 hours

**Total: 39-51 hours** (approximately 5-7 working days)

### Success Metrics

#### Code Quality
- [ ] Client reduced from 1,020 to <200 lines
- [ ] No files exceed 400 lines
- [ ] No methods exceed 50 lines
- [ ] Cyclomatic complexity <15 for all functions
- [ ] Zero magic numbers or strings
- [ ] All linters pass with zero warnings

#### Test Coverage
- [ ] Maintain 80%+ overall coverage
- [ ] Each new service has 85%+ coverage
- [ ] All edge cases tested
- [ ] Table-driven tests for complex logic

#### Documentation
- [ ] All public APIs documented
- [ ] Architecture documentation updated
- [ ] Migration guide for internal changes
- [ ] Code examples updated

#### Performance
- [ ] No performance regression (benchmark tests)
- [ ] Memory usage unchanged
- [ ] Build time unchanged

### Risk Mitigation

#### Risk: Breaking Public API
**Mitigation:** All changes are internal refactoring; Client facade maintains exact same public interface

#### Risk: Test Coverage Decrease
**Mitigation:** TDD approach ensures tests written before implementation; coverage checked after each commit

#### Risk: Performance Regression
**Mitigation:** Benchmark tests before and after; service delegation adds negligible overhead

#### Risk: Merge Conflicts
**Mitigation:** Atomic commits allow easy rebase; each sub-phase independently mergeable

### Execution Strategy

1. **Sequential Execution:** Execute phases in order (24.1 → 24.8)
2. **Validation Points:** After each sub-phase, validate:
   - All tests pass
   - Coverage maintained
   - Linters pass
   - Documentation updated
3. **Atomic Commits:** Each task broken into test + implementation commits
4. **Review Points:** After phases 24.3, 24.5, and 24.8 (major milestones)

### Follow-Up Work

After Phase 24 completion:
1. **Phase 25:** Implement additional service features (e.g., batch operations)
2. **Phase 26:** Performance optimization based on benchmarks
3. **Phase 27:** API stability review and v1.0 preparation

### Governance

All commits follow constitutional requirements:
- Conventional Commits specification
- Test-first development
- Atomic commits with complete working state
- Academic documentation style
- No breaking changes

---

## Quick Reference

### File Impact Summary
- **New files:** ~25 (services, helpers, tests)
- **Modified files:** ~30 (refactored implementations)
- **Deleted files:** 0 (maintain backward compatibility)
- **Net line change:** ~-500 lines (reduced complexity)

### Commit Estimate
- **Total commits:** ~80-100 (following atomic commit principle)
- **PR strategy:** One PR per sub-phase or combine 24.1-24.2, 24.3-24.4, 24.5, 24.6-24.8

### Testing Requirements
- **New test files:** ~25
- **Updated test files:** ~15
- **Test cases added:** ~200+
- **Coverage target:** Maintain 80%+ overall

---

## Appendix A: Conventional Commit Examples

### Phase 24.1 - Constants
```
test(domain): add permission constant tests
feat(domain): add file permission constants
refactor(config): use permission constants
```

### Phase 24.5 - Client Decomposition
```
test(pkg): add ManageService tests
feat(pkg): extract ManageService from Client
refactor(pkg): Client delegates to ManageService
```

### Phase 24.6 - Long Method Refactoring
```
test(pkg): add tests for scan directory helpers
refactor(pkg): extract determineScanDirectories
refactor(pkg): simplify performOrphanScan
```

---

## Appendix B: Service Dependency Graph

```
Client (Facade)
├── ManageService
│   ├── Executor
│   ├── ManagePipeline
│   └── ManifestService
├── UnmanageService
│   ├── Executor
│   └── ManifestService
├── StatusService
│   └── ManifestService
├── DoctorService
│   └── ManifestService
└── AdoptService
    ├── Executor
    └── ManifestService
```

---

## Appendix C: Pre-Refactoring Checklist

Before starting Phase 24:
- [ ] All existing tests pass
- [ ] Coverage at baseline (document current %)
- [ ] All linters pass
- [ ] Create feature branch: `refactor/phase-24-code-smell-remediation`
- [ ] Tag current state: `v0.x.x-pre-phase-24`
- [ ] Backup current coverage report
- [ ] Document current file sizes for comparison
