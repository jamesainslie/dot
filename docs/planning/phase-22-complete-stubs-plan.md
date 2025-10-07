# Phase 22: Complete Stubbed Features — Implementation Plan

## Overview

**Objective**: Complete all stubbed and incomplete features identified in code review, establishing full functionality for all documented commands and features.

**Prerequisites**: Phase 19 complete, Phase 21 (terminology refactor) complete or in progress

**Estimated Effort**: 32-40 hours

**Success Criteria**:
- All CLI commands fully functional
- Per-package link tracking operational
- Backup directory configuration wired throughout system
- Incremental remanage implemented with hash-based detection
- Test coverage maintained at ≥ 80%
- All linters passing with zero warnings
- Documentation updated for all new functionality

---

## Problem Statement

Code review identified six categories of incomplete functionality:

1. **CLI Commands Stubbed**: Four commands registered but not implemented
2. **Incomplete API Features**: Missing link tracking and incremental planning
3. **Backup Directory Disconnected**: Configuration exists but not wired
4. **API Breaking Change**: No migration path for Doctor() signature change
5. **Deferred Features**: Four tasks deferred from Phase 15c
6. **Future Enhancements**: Streaming API, ConfigBuilder, architectural improvements

**Impact**: Users cannot execute core commands despite API implementations existing. Manifest tracking incomplete. Performance suboptimal.

---

## Phase Structure

This phase is divided into 6 sub-phases with clear dependencies:

```
Phase 22.1: CLI Command Implementation (High Priority)
    ├─ Blocking: Core functionality
    └─ Effort: 8-10 hours

Phase 22.2: Per-Package Link Tracking (High Priority)
    ├─ Blocking: Manifest accuracy
    └─ Effort: 6-8 hours

Phase 22.3: Backup Directory Wiring (Medium Priority)
    ├─ Depends: Phase 22.1
    └─ Effort: 4-6 hours

Phase 22.4: API Migration Path (Medium Priority)
    ├─ Independent
    └─ Effort: 3-4 hours

Phase 22.5: Incremental Remanage (Low Priority)
    ├─ Depends: Phase 22.2
    └─ Effort: 6-8 hours

Phase 22.6: Future Enhancements (Low Priority)
    ├─ Depends: All previous phases
    └─ Effort: 5-6 hours (scoping only)
```

---

## Phase 22.1: CLI Command Implementation

**Priority**: CRITICAL — Blocks all user-facing functionality  
**Estimated Effort**: 8-10 hours  
**Files Affected**: `cmd/dot/{manage,unmanage,remanage,adopt}.go`, tests

### Current State

All four commands have stub implementations:
```go
// cmd/dot/manage.go:16
RunE: func(cmd *cobra.Command, args []string) error {
    // TODO: Implement
    return nil
},
```

API implementations exist and are tested:
- `internal/api/manage.go` - ✅ Complete, tested
- `internal/api/unmanage.go` - ✅ Complete, tested
- `internal/api/adopt.go` - ✅ Complete, tested
- `internal/api/remanage.go` - ✅ Partial (needs enhancement in 22.5)

### Implementation Tasks

#### T22.1-001: Implement Manage Command Handler
**Estimate**: 2 hours

**Test (Write First)**:
```go
// cmd/dot/manage_test.go
func TestManageCommand_Execute(t *testing.T) {
    // Setup test package directory with fixtures
    // Execute: dot manage vim
    // Verify: Links created in target directory
    // Verify: Manifest updated
}

func TestManageCommand_DryRun(t *testing.T) {
    // Execute: dot manage vim --dry-run
    // Verify: No links created
    // Verify: Appropriate output shown
}

func TestManageCommand_MultiplePackages(t *testing.T) {
    // Execute: dot manage vim zsh git
    // Verify: All packages installed
}

func TestManageCommand_Errors(t *testing.T) {
    // Test: Invalid package directory
    // Test: Package not found
    // Test: Permission errors
}
```

**Implementation**:
```go
// cmd/dot/manage.go
func newManageCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "manage PACKAGE [PACKAGE...]",
        Short: "Install packages by creating symlinks",
        Long: `Install one or more packages by creating symlinks from the package 
directory to the target directory.`,
        Args: cobra.MinimumNArgs(1),
        RunE: runManage,
    }
    return cmd
}

func runManage(cmd *cobra.Command, args []string) error {
    cfg, err := buildConfig()
    if err != nil {
        return formatError(err)
    }

    client, err := dot.NewClient(cfg)
    if err != nil {
        return formatError(err)
    }

    ctx := cmd.Context()
    if ctx == nil {
        ctx = context.Background()
    }

    packages := args

    if err := client.Manage(ctx, packages...); err != nil {
        return formatError(err)
    }

    if !cfg.DryRun {
        fmt.Printf("Successfully managed %d package(s)\n", len(packages))
    }

    return nil
}
```

**Commit Message**:
```
feat(cli): implement manage command handler

Wire manage command to client.Manage() API method with proper error
handling, context propagation, and dry-run support.

Users can now install packages using the manage command. The handler
builds configuration from flags, creates a client, and executes the
manage operation with appropriate output formatting.

- Add runManage handler calling client.Manage()
- Add comprehensive tests for success and error paths
- Support dry-run mode with appropriate output
- Format errors for user-friendly display

Closes #[issue-number]
```

#### T22.1-002: Implement Unmanage Command Handler
**Estimate**: 2 hours

**Test Structure**: Similar to T22.1-001 but for unmanage operation

**Implementation**: Parallel to manage with `client.Unmanage(ctx, packages...)`

**Commit Message**:
```
feat(cli): implement unmanage command handler

Wire unmanage command to client.Unmanage() API method enabling package
removal functionality.

- Add runUnmanage handler calling client.Unmanage()
- Add tests for package removal scenarios
- Handle "package not installed" gracefully
- Support dry-run mode
```

#### T22.1-003: Implement Adopt Command Handler
**Estimate**: 2 hours

**Test (Write First)**:
```go
func TestAdoptCommand_Execute(t *testing.T) {
    // Setup: Create file in target directory
    // Execute: dot adopt vim .vimrc
    // Verify: File moved to package directory
    // Verify: Symlink created at original location
}

func TestAdoptCommand_MultipleFiles(t *testing.T) {
    // Execute: dot adopt vim .vimrc .vim/
    // Verify: All files adopted
}
```

**Implementation**: Note that adopt has different signature `(ctx, files, package)`

```go
func runAdopt(cmd *cobra.Command, args []string) error {
    if len(args) < 2 {
        return fmt.Errorf("adopt requires package name and at least one file")
    }

    pkg := args[0]
    files := args[1:]

    // ... build config, create client ...

    if err := client.Adopt(ctx, files, pkg); err != nil {
        return formatError(err)
    }

    fmt.Printf("Successfully adopted %d file(s) into %s\n", len(files), pkg)
    return nil
}
```

**Commit Message**:
```
feat(cli): implement adopt command handler

Enable file adoption workflow where existing files in target directory
are moved into a package and symlinked back to original location.

- Add runAdopt handler with package and files arguments
- Add tests for single and multiple file adoption
- Verify file movement and symlink creation
- Support dry-run mode
```

#### T22.1-004: Implement Remanage Command Handler
**Estimate**: 2 hours

**Implementation**: Straightforward delegation to `client.Remanage(ctx, packages...)`

**Commit Message**:
```
feat(cli): implement remanage command handler

Wire remanage command to client.Remanage() API for package reinstallation
workflow.

- Add runRemanage handler calling client.Remanage()
- Add tests for remanage scenarios
- Support dry-run mode
- Handle packages not currently installed
```

#### T22.1-005: Integration Testing
**Estimate**: 2 hours

**Test Suite**:
```go
// cmd/dot/integration_commands_test.go
func TestCommandsIntegration_CompleteWorkflow(t *testing.T) {
    // 1. Manage packages
    // 2. Verify status shows managed
    // 3. Remanage packages
    // 4. Unmanage packages
    // 5. Verify status shows empty
}

func TestCommandsIntegration_AdoptWorkflow(t *testing.T) {
    // 1. Create files in target
    // 2. Adopt files into package
    // 3. Verify files moved and linked
    // 4. Unmanage package
    // 5. Verify symlinks removed
}
```

**Commit Message**:
```
test(cli): add integration tests for command workflows

Verify end-to-end functionality of manage, unmanage, remanage, and adopt
commands through complete user workflows.

- Add complete workflow integration test
- Add adopt workflow integration test
- Verify state consistency across operations
- Test error recovery scenarios
```

### Deliverables

- ✅ All four CLI commands fully functional
- ✅ Comprehensive test coverage (≥ 80%)
- ✅ Error handling with user-friendly messages
- ✅ Dry-run support verified
- ✅ Integration tests passing

---

## Phase 22.2: Per-Package Link Tracking

**Priority**: HIGH — Required for manifest accuracy  
**Estimated Effort**: 6-8 hours  
**Files Affected**: `internal/api/manage.go`, `internal/pipeline/`, `internal/planner/`, tests

### Current State

```go
// internal/api/manage.go:109
// TODO: Track links per package in planner/pipeline
m.AddPackage(manifest.PackageInfo{
    Name:        pkg,
    InstalledAt: time.Now(),
    LinkCount:   0,     // Wrong: Always 0
    Links:       []string{},  // Wrong: Always empty
})
```

**Problem**: Manifest records package installation but doesn't track which links belong to which package. This breaks unmanage, doctor, and status commands.

### Root Cause Analysis

The pipeline returns a unified `dot.Plan` with all operations, but operations don't carry package ownership information.

**Current Flow**:
```
Scanner → Planner → Pipeline → Executor
  ↓         ↓          ↓           ↓
Package  Links      Plan        Execute
Names    Created   (unified)    (all ops)
```

**Missing**: Mapping from operation → package

### Design Decision

**Option A**: Add PackageName to Operation interface
```go
type Operation interface {
    Kind() OperationKind
    Package() string  // NEW
    // ...
}
```

**Option B**: Return per-package plans from pipeline
```go
type ManageResult struct {
    Plans map[string]dot.Plan  // package name → plan
}
```

**Option C**: Add metadata to operations
```go
type LinkCreate struct {
    // ...
    Metadata map[string]string  // {"package": "vim"}
}
```

**Recommendation**: **Option B** — Cleaner separation, no operation interface changes, easier to implement.

### Implementation Tasks

#### T22.2-001: Design Package-Operation Mapping
**Estimate**: 1 hour

**Create ADR**:
```markdown
# ADR-002: Package-Operation Mapping for Manifest Tracking

## Decision
Use per-package plan aggregation in pipeline with operation tagging.

## Rationale
- No breaking changes to Operation interface
- Maintains operation immutability
- Enables future features (parallel package installation)
- Clean separation of concerns

## Implementation
Pipeline returns PackageResults with per-package operation lists.
```

**Commit Message**:
```
docs(adr): add ADR-002 for package-operation mapping

Document decision to use per-package plan aggregation for manifest
tracking without modifying operation interface.
```

#### T22.2-002: Extend Plan with Package Mapping
**Estimate**: 2 hours

**Test (Write First)**:
```go
// pkg/dot/plan_test.go
func TestPlan_OperationsByPackage(t *testing.T) {
    plan := dot.Plan{
        Operations: []dot.Operation{...},
        PackageOperations: map[string][]dot.OperationID{
            "vim": []dot.OperationID{...},
            "zsh": []dot.OperationID{...},
        },
    }
    
    vimOps := plan.OperationsForPackage("vim")
    assert.Len(t, vimOps, 3)
}
```

**Implementation**:
```go
// pkg/dot/plan.go
type Plan struct {
    Operations []Operation
    Metadata   PlanMetadata
    
    // PackageOperations maps package names to operation IDs
    // that belong to that package
    PackageOperations map[string][]OperationID
}

// OperationsForPackage returns all operations for a package
func (p Plan) OperationsForPackage(pkg string) []Operation {
    ids := p.PackageOperations[pkg]
    ops := make([]Operation, 0, len(ids))
    
    for _, op := range p.Operations {
        for _, id := range ids {
            if op.ID() == id {
                ops = append(ops, op)
                break
            }
        }
    }
    
    return ops
}
```

**Commit Message**:
```
feat(domain): add package-operation mapping to Plan

Extend Plan type with PackageOperations field mapping package names to
operation IDs. Enables manifest tracking of which links belong to which
packages without modifying Operation interface.

- Add PackageOperations map to Plan struct
- Add OperationsForPackage helper method
- Add tests for package filtering
- Maintain backward compatibility

Refs: ADR-002
```

#### T22.2-003: Update Pipeline to Track Package Operations
**Estimate**: 2 hours

**Implementation**:
```go
// internal/pipeline/packages.go
func (p *ManagePipeline) Execute(ctx context.Context, input ManageInput) dot.Result[dot.Plan] {
    // ... existing stages ...
    
    // Build package-operation mapping
    packageOps := make(map[string][]dot.OperationID)
    
    // Track which operations came from which package
    currentPackage := ""
    for _, pkg := range input.Packages {
        pkgOps := collectOpsForPackage(desired, pkg)
        packageOps[pkg] = extractOperationIDs(pkgOps)
    }
    
    plan := dot.Plan{
        Operations:        operations,
        Metadata:          metadata,
        PackageOperations: packageOps,
    }
    
    return dot.Ok(plan)
}
```

**Commit Message**:
```
feat(pipeline): track package ownership in operation plans

Build package-operation mapping during pipeline execution by tracking
which operations originate from each package's desired state.

- Add package tracking to ManagePipeline
- Populate PackageOperations map in Plan
- Add tests verifying correct mapping
- Handle edge cases (shared directories)
```

#### T22.2-004: Update Manifest Writer to Use Package Operations
**Estimate**: 1.5 hours

**Implementation**:
```go
// internal/api/manage.go
func (c *client) updateManifest(ctx context.Context, packages []string, plan dot.Plan) error {
    // ... load manifest ...
    
    for _, pkg := range packages {
        ops := plan.OperationsForPackage(pkg)
        
        // Extract link paths from LinkCreate operations
        links := make([]string, 0)
        for _, op := range ops {
            if linkOp, ok := op.(dot.LinkCreate); ok {
                // Get relative path from target directory
                relPath, _ := filepath.Rel(c.config.TargetDir, linkOp.Target().String())
                links = append(links, relPath)
            }
        }
        
        m.AddPackage(manifest.PackageInfo{
            Name:        pkg,
            InstalledAt: time.Now(),
            LinkCount:   len(links),
            Links:       links,
        })
    }
    
    return c.manifest.Save(ctx, targetPath, m)
}
```

**Commit Message**:
```
fix(api): track actual links in manifest per package

Replace placeholder link tracking with actual link extraction from
package operations. Manifest now accurately records which links belong
to which package.

- Extract links from plan.OperationsForPackage()
- Compute relative paths for manifest storage
- Update LinkCount to actual count
- Add tests verifying accurate tracking

Fixes: #[issue] (Manifest shows LinkCount: 0)
```

#### T22.2-005: Integration Testing
**Estimate**: 1.5 hours

**Test Suite**:
```go
func TestManifestTracking_AccurateLinks(t *testing.T) {
    // Manage package with 3 files
    // Load manifest
    // Verify LinkCount == 3
    // Verify Links contains all 3 paths
}

func TestManifestTracking_MultiplePackages(t *testing.T) {
    // Manage vim (3 files) and zsh (2 files)
    // Verify vim shows 3 links
    // Verify zsh shows 2 links
    // Verify no cross-contamination
}
```

**Commit Message**:
```
test(api): verify accurate manifest link tracking

Add integration tests confirming manifest accurately tracks links per
package across various scenarios.

- Test single package with multiple links
- Test multiple packages with separate links
- Test shared directories
- Verify unmanage updates manifest correctly
```

### Deliverables

- ✅ Package-operation mapping implemented
- ✅ Manifest accurately tracks links per package
- ✅ LinkCount reflects actual link count
- ✅ Links array populated correctly
- ✅ Test coverage maintained ≥ 80%

---

## Phase 22.3: Backup Directory Wiring

**Priority**: MEDIUM — Feature exists but non-functional  
**Estimated Effort**: 4-6 hours  
**Dependencies**: Phase 22.1 (CLI commands functional)  
**Files Affected**: `internal/pipeline/packages.go`, `pkg/dot/config.go`, `cmd/dot/root.go`

### Current State

```go
// internal/pipeline/packages.go:73
BackupDir: "", // TODO: Add backup dir to options
```

Configuration exists:
```go
// pkg/dot/config.go
type Config struct {
    BackupDir string  // Field exists
}

func (cfg Config) WithDefaults() Config {
    if cfg.BackupDir == "" {
        cfg.BackupDir = filepath.Join(cfg.TargetDir, ".dot-backup")
    }
    return cfg
}
```

But pipeline ignores it and hardcodes empty string.

### Implementation Tasks

#### T22.3-001: Add BackupDir to PipelineOptions
**Estimate**: 1 hour

**Test (Write First)**:
```go
// internal/pipeline/packages_test.go
func TestManagePipeline_UsesBackupDir(t *testing.T) {
    opts := NewPipelineOptions(
        WithFS(fs),
        WithBackupDir("/custom/backup"),
    )
    
    pipe := NewManagePipeline(opts)
    // ... execute pipeline ...
    // Verify backup dir passed to resolver
}
```

**Implementation**:
```go
// internal/pipeline/types.go
type PipelineOptions struct {
    FS        dot.FS
    Policies  planner.ResolutionPolicies
    BackupDir string  // NEW
}

func WithBackupDir(dir string) PipelineOption {
    return func(opts *PipelineOptions) {
        opts.BackupDir = dir
    }
}
```

**Commit Message**:
```
feat(pipeline): add backup directory to pipeline options

Add BackupDir field to PipelineOptions with functional option setter
for configuration propagation.

- Add BackupDir field to PipelineOptions
- Add WithBackupDir functional option
- Add tests for option setting
```

#### T22.3-002: Wire BackupDir Through Pipeline
**Estimate**: 1 hour

**Implementation**:
```go
// internal/pipeline/packages.go
func (p *ManagePipeline) Execute(ctx context.Context, input ManageInput) dot.Result[dot.Plan] {
    // ... existing stages ...
    
    resolveInput := ResolveInput{
        Desired:   desired,
        FS:        p.opts.FS,
        Policies:  p.opts.Policies,
        BackupDir: p.opts.BackupDir,  // Use from options instead of ""
    }
    
    // ...
}
```

**Commit Message**:
```
fix(pipeline): use configured backup directory

Replace hardcoded empty backup directory with configured value from
pipeline options.

- Use p.opts.BackupDir in resolve stage
- Remove TODO comment
- Verify backup operations use correct directory
```

#### T22.3-003: Add CLI Flag for Backup Directory
**Estimate**: 1.5 hours

**Implementation**:
```go
// cmd/dot/root.go
type globalConfig struct {
    packageDir string
    targetDir  string
    backupDir  string  // NEW
    dryRun     bool
    // ...
}

func NewRootCommand(version, commit, date string) *cobra.Command {
    // ...
    rootCmd.PersistentFlags().StringVar(&globalCfg.backupDir, "backup-dir", "",
        "Directory for backup files (default: <target>/.dot-backup)")
    // ...
}

func buildConfig() (dot.Config, error) {
    // ...
    
    cfg := dot.Config{
        PackageDir: packageDir,
        TargetDir:  targetDir,
        BackupDir:  globalCfg.backupDir,  // NEW
        DryRun:     globalCfg.dryRun,
        // ...
    }
    
    return cfg.WithDefaults(), nil  // Will set default if empty
}
```

**Commit Message**:
```
feat(cli): add backup-dir flag for custom backup location

Enable users to specify custom backup directory via --backup-dir flag
with sensible default.

- Add backupDir to globalConfig
- Add --backup-dir persistent flag
- Wire through to dot.Config
- Default to <target>/.dot-backup if not specified
```

#### T22.3-004: Add Configuration File Support
**Estimate**: 1 hour

**Implementation**:
```go
// internal/config/extended.go
type SymlinksConfig struct {
    Mode         string `mapstructure:"mode"`
    Folding      bool   `mapstructure:"folding"`
    Backup       bool   `mapstructure:"backup"`
    BackupDir    string `mapstructure:"backup_dir"`    // NEW
    BackupSuffix string `mapstructure:"backup_suffix"`
}
```

**Update config loader**:
```go
// cmd/dot/config.go
func getConfigValue(cfg *config.ExtendedConfig, key string) (string, error) {
    switch key {
    // ...
    case "symlinks.backup_dir":
        return cfg.Symlinks.BackupDir, nil
    // ...
    }
}
```

**Commit Message**:
```
feat(config): add backup_dir configuration option

Support backup directory configuration in config files with environment
variable override.

- Add backup_dir to SymlinksConfig
- Add DOT_SYMLINKS_BACKUP_DIR environment variable
- Add config get/set support
- Update default config template
```

#### T22.3-005: Integration Testing and Documentation
**Estimate**: 1.5 hours

**Test Suite**:
```go
func TestBackupDir_CustomLocation(t *testing.T) {
    // Set custom backup dir
    // Trigger conflict requiring backup
    // Verify backup created in custom location
}

func TestBackupDir_DefaultLocation(t *testing.T) {
    // Don't specify backup dir
    // Trigger conflict requiring backup
    // Verify backup in <target>/.dot-backup
}
```

**Documentation Update**:
- Add to `docs/user/04-configuration.md`
- Add to `docs/user/07-advanced.md` (backup strategies)
- Update `docs/reference/configuration.md`

**Commit Message**:
```
docs(config): document backup directory configuration

Add comprehensive documentation for backup directory configuration
including flag, environment variable, and config file methods.

- Add backup_dir to configuration reference
- Add backup strategies to advanced guide
- Add examples with custom backup locations
```

### Deliverables

- ✅ Backup directory configuration functional end-to-end
- ✅ CLI flag support
- ✅ Configuration file support
- ✅ Environment variable support
- ✅ Tests verifying backup to correct location
- ✅ Documentation updated

---

## Phase 22.4: API Migration Path

**Priority**: MEDIUM — Developer experience  
**Estimated Effort**: 3-4 hours  
**Files Affected**: `pkg/dot/client.go`, tests, documentation

### Current State

```go
// pkg/dot/client.go:94
// TODO: BREAKING CHANGE in v0.2.0 - Added scanCfg parameter.
```

The `Doctor()` method signature changed from:
```go
Doctor(ctx context.Context) (DiagnosticReport, error)
```

To:
```go
Doctor(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)
```

**Problem**: Library consumers updating to v0.2.0 face compile errors with no transitional path.

### Design Decision

Provide transitional API following Go best practices:

**Option A**: Add DoctorWithScan, deprecate Doctor
```go
// Deprecated: Use DoctorWithScan for explicit scan configuration
func (c Client) Doctor(ctx context.Context) (DiagnosticReport, error)

func (c Client) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)
```

**Option B**: Make scanCfg optional with variadic parameters
```go
func (c Client) Doctor(ctx context.Context, scanCfg ...ScanConfig) (DiagnosticReport, error)
```

**Recommendation**: **Option A** — Explicit, follows Go deprecation patterns, clear migration path.

### Implementation Tasks

#### T22.4-001: Add DoctorWithScan Method
**Estimate**: 1 hour

**Test (Write First)**:
```go
func TestClient_DoctorWithScan(t *testing.T) {
    client := setupTestClient(t)
    
    // Test with scoped scanning
    report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())
    require.NoError(t, err)
    assert.NotEmpty(t, report.Statistics)
    
    // Test with deep scanning
    report, err = client.DoctorWithScan(ctx, dot.DeepScanConfig(3))
    require.NoError(t, err)
}
```

**Implementation**:
```go
// pkg/dot/client.go

// Doctor performs health check with default scan configuration.
//
// Deprecated: Use DoctorWithScan to explicitly specify scan configuration.
// Doctor uses DefaultScanConfig() which performs no orphan scanning.
// This method will be removed in v1.0.0.
func (c Client) Doctor(ctx context.Context) (DiagnosticReport, error) {
    return c.DoctorWithScan(ctx, DefaultScanConfig())
}

// DoctorWithScan performs health check with specified scan configuration.
//
// The scanCfg parameter controls orphaned link detection:
//   - DefaultScanConfig(): No orphan scanning (fastest)
//   - ScopedScanConfig(): Smart scanning of managed directories
//   - DeepScanConfig(depth): Full scan with depth limit
//
// Returns DiagnosticReport with health status, issues, and statistics.
func (c Client) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
    impl := registry.GetImplementation()
    if impl == nil {
        return DiagnosticReport{}, ErrNoImplementation{Operation: "doctor"}
    }
    return impl.Doctor(ctx, scanCfg)
}
```

**Commit Message**:
```
feat(api): add DoctorWithScan with migration path

Add DoctorWithScan method for explicit scan configuration while
deprecating original Doctor method. Provides smooth migration path for
library consumers.

- Add DoctorWithScan(ctx, scanCfg) method
- Deprecate Doctor(ctx) with clear migration guidance
- Doctor delegates to DoctorWithScan with DefaultScanConfig
- Add tests for both methods
- Maintain backward compatibility

BREAKING CHANGE: Doctor() is deprecated in favor of DoctorWithScan().
Existing code continues to work but should migrate:

  Before:
    report, err := client.Doctor(ctx)

  After:
    report, err := client.DoctorWithScan(ctx, dot.DefaultScanConfig())

Doctor() will be removed in v1.0.0.
```

#### T22.4-002: Update Internal API Implementation
**Estimate**: 0.5 hours

**Implementation**:
```go
// internal/api/client.go
type clientImpl interface {
    // ...
    Doctor(ctx context.Context, scanCfg dot.ScanConfig) (dot.DiagnosticReport, error)
    DoctorWithScan(ctx context.Context, scanCfg dot.ScanConfig) (dot.DiagnosticReport, error)
}

// Implement DoctorWithScan as alias to Doctor
func (c *client) DoctorWithScan(ctx context.Context, scanCfg dot.ScanConfig) (dot.DiagnosticReport, error) {
    return c.Doctor(ctx, scanCfg)
}
```

**Commit Message**:
```
refactor(api): add DoctorWithScan to internal implementation

Add DoctorWithScan method to internal client implementation maintaining
consistency with public interface.
```

#### T22.4-003: Update CLI to Use New Method
**Estimate**: 0.5 hours

**Implementation**:
```go
// cmd/dot/doctor.go
func runDoctor(cmd *cobra.Command, args []string, scanMode string) error {
    // ... setup ...
    
    // Use new DoctorWithScan method
    report, err := client.DoctorWithScan(ctx, scanCfg)
    if err != nil {
        return formatError(err)
    }
    
    // ... format output ...
}
```

**Commit Message**:
```
refactor(cli): use DoctorWithScan in doctor command

Update doctor command to use new DoctorWithScan method for explicit
scan configuration.
```

#### T22.4-004: Documentation and Migration Guide
**Estimate**: 1.5 hours

**Create Migration Guide**:
```markdown
# Migration Guide: v0.1.x → v0.2.x

## API Changes

### Doctor Method Signature

**What Changed**:
The `Doctor()` method now requires explicit scan configuration.

**Old API (v0.1.x)**:
```go
report, err := client.Doctor(ctx)
```

**New API (v0.2.x)**:
```go
// Option 1: No orphan scanning (fastest, default)
report, err := client.DoctorWithScan(ctx, dot.DefaultScanConfig())

// Option 2: Smart scanning (recommended)
report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())

// Option 3: Deep scanning with depth limit
report, err := client.DoctorWithScan(ctx, dot.DeepScanConfig(5))
```

**Backward Compatibility**:
The old `Doctor(ctx)` method still works but is deprecated:
```go
// Works but shows deprecation warning
report, err := client.Doctor(ctx)  
```

**Timeline**:
- v0.2.x: `Doctor(ctx)` deprecated, `DoctorWithScan()` recommended
- v1.0.0: `Doctor(ctx)` removed, must use `DoctorWithScan()`

**Automated Migration**:
```bash
# Find all Doctor calls
grep -r "\.Doctor(ctx)" .

# Replace with new method
sed -i 's/\.Doctor(ctx)/\.DoctorWithScan(ctx, dot.DefaultScanConfig())/g' **/*.go
```
```

**Update Documentation**:
- Add to `docs/migration/v0.2.0-migration.md`
- Update `docs/user/05-commands.md` (doctor command)
- Update `pkg/dot/doc.go` examples
- Add to CHANGELOG.md

**Commit Message**:
```
docs(migration): add v0.2.0 migration guide

Document Doctor method signature change and provide migration paths for
library consumers.

- Create v0.2.0 migration guide
- Add automated migration examples
- Update API documentation
- Add to CHANGELOG
```

### Deliverables

- ✅ DoctorWithScan method implemented
- ✅ Backward compatibility maintained
- ✅ Deprecation warnings in place
- ✅ Migration guide complete
- ✅ Examples updated
- ✅ Tests passing

---

## Phase 22.5: Incremental Remanage

**Priority**: LOW — Performance optimization  
**Estimated Effort**: 6-8 hours  
**Dependencies**: Phase 22.2 (link tracking)  
**Files Affected**: `internal/api/remanage.go`, `internal/manifest/`, tests

### Current State

```go
// internal/api/remanage.go:11
// TODO: Use incremental planning with hash-based change detection.

func (c *client) Remanage(ctx context.Context, packages ...string) error {
    // Currently does full unmanage + manage
    err := c.Unmanage(ctx, packages...)
    // ...
    return c.Manage(ctx, packages...)
}
```

**Problem**: Remanage removes all links then recreates them, even for unchanged files. Inefficient and causes unnecessary filesystem churn.

### Design: Hash-Based Change Detection

**Approach**: Store content hashes in manifest, compare on remanage

**Manifest Schema Extension**:
```go
type PackageInfo struct {
    Name        string
    InstalledAt time.Time
    LinkCount   int
    Links       []string
    FileHashes  map[string]string  // NEW: file path → SHA256 hash
}
```

**Algorithm**:
```
1. Load current manifest with hashes
2. Scan package directory, compute current hashes
3. Compare:
   - Unchanged: Keep existing link
   - Modified: Remove old link, create new link
   - Added: Create new link
   - Deleted: Remove link
4. Generate minimal operation plan
5. Update manifest with new hashes
```

### Implementation Tasks

#### T22.5-001: Add Hash Storage to Manifest
**Estimate**: 1.5 hours

**Test (Write First)**:
```go
// internal/manifest/manifest_test.go
func TestManifest_StoreFileHashes(t *testing.T) {
    m := manifest.New()
    
    hashes := map[string]string{
        ".vimrc": "abc123...",
        ".vim/colors/theme.vim": "def456...",
    }
    
    m.AddPackage(manifest.PackageInfo{
        Name:       "vim",
        FileHashes: hashes,
    })
    
    pkg, _ := m.GetPackage("vim")
    assert.Equal(t, hashes, pkg.FileHashes)
}
```

**Implementation**:
```go
// internal/manifest/types.go
type PackageInfo struct {
    Name        string            `yaml:"name"`
    InstalledAt time.Time         `yaml:"installed_at"`
    LinkCount   int               `yaml:"link_count"`
    Links       []string          `yaml:"links"`
    FileHashes  map[string]string `yaml:"file_hashes,omitempty"`  // NEW
}
```

**Commit Message**:
```
feat(manifest): add file hash tracking for change detection

Extend manifest schema to store SHA256 hashes of package files enabling
incremental remanage operations.

- Add FileHashes field to PackageInfo
- Add tests for hash storage and retrieval
- Maintain backward compatibility (omitempty)
- Update manifest version if needed
```

#### T22.5-002: Implement Hash Computation Utility
**Estimate**: 1 hour

**Test (Write First)**:
```go
// internal/api/hashing_test.go
func TestComputeFileHash(t *testing.T) {
    fs := setupTestFS(t)
    hash, err := computeFileHash(ctx, fs, "/path/to/file")
    require.NoError(t, err)
    assert.Len(t, hash, 64)  // SHA256 hex string
}

func TestComputePackageHashes(t *testing.T) {
    fs := setupTestFS(t)
    hashes, err := computePackageHashes(ctx, fs, "/pkg/vim")
    require.NoError(t, err)
    assert.Contains(t, hashes, ".vimrc")
}
```

**Implementation**:
```go
// internal/api/hashing.go
package api

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
)

// computeFileHash computes SHA256 hash of file contents
func computeFileHash(ctx context.Context, fs dot.FS, path string) (string, error) {
    data, err := fs.ReadFile(ctx, path)
    if err != nil {
        return "", err
    }
    
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:]), nil
}

// computePackageHashes computes hashes for all files in package
func computePackageHashes(ctx context.Context, fs dot.FS, pkgPath string) (map[string]string, error) {
    hashes := make(map[string]string)
    
    // Walk package directory
    entries, err := fs.ReadDir(ctx, pkgPath)
    if err != nil {
        return nil, err
    }
    
    for _, entry := range entries {
        if entry.IsDir() {
            continue  // TODO: Handle nested directories
        }
        
        fullPath := filepath.Join(pkgPath, entry.Name())
        hash, err := computeFileHash(ctx, fs, fullPath)
        if err != nil {
            return nil, err
        }
        
        hashes[entry.Name()] = hash
    }
    
    return hashes, nil
}

// compareHashes identifies changed, added, and deleted files
func compareHashes(old, new map[string]string) (changed, added, deleted []string) {
    // Files in new but not old or with different hash
    for file, newHash := range new {
        if oldHash, exists := old[file]; !exists {
            added = append(added, file)
        } else if oldHash != newHash {
            changed = append(changed, file)
        }
    }
    
    // Files in old but not new
    for file := range old {
        if _, exists := new[file]; !exists {
            deleted = append(deleted, file)
        }
    }
    
    return changed, added, deleted
}
```

**Commit Message**:
```
feat(api): add file hashing utilities for change detection

Implement SHA256-based file hashing for detecting changes during
remanage operations.

- Add computeFileHash for single file hashing
- Add computePackageHashes for package-wide hashing
- Add compareHashes for change detection
- Add comprehensive tests including edge cases
```

#### T22.5-003: Implement Incremental PlanRemanage
**Estimate**: 2.5 hours

**Test (Write First)**:
```go
// internal/api/remanage_test.go
func TestPlanRemanage_Incremental_UnchangedFiles(t *testing.T) {
    // Setup: Package installed with hashes in manifest
    // Change: Nothing changed
    // Verify: Empty operation plan
}

func TestPlanRemanage_Incremental_ChangedFile(t *testing.T) {
    // Setup: Package installed
    // Change: Modify one file
    // Verify: Plan contains delete+create for changed file only
}

func TestPlanRemanage_Incremental_AddedFile(t *testing.T) {
    // Setup: Package with 2 files installed
    // Change: Add 3rd file to package
    // Verify: Plan contains create for new file only
}

func TestPlanRemanage_Incremental_DeletedFile(t *testing.T) {
    // Setup: Package with 3 files installed
    // Change: Remove 1 file from package
    // Verify: Plan contains delete for removed link only
}
```

**Implementation**:
```go
// internal/api/remanage.go
func (c *client) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
    targetPathResult := dot.NewTargetPath(c.config.TargetDir)
    if !targetPathResult.IsOk() {
        return dot.Plan{}, targetPathResult.UnwrapErr()
    }
    targetPath := targetPathResult.Unwrap()
    
    // Load manifest
    manifestResult := c.manifest.Load(ctx, targetPath)
    if !manifestResult.IsOk() {
        // No manifest - fall back to full manage
        return c.PlanManage(ctx, packages...)
    }
    
    m := manifestResult.Unwrap()
    
    operations := make([]dot.Operation, 0)
    
    for _, pkg := range packages {
        pkgInfo, exists := m.GetPackage(pkg)
        if !exists {
            // Package not installed - plan as new install
            pkgPlan, err := c.PlanManage(ctx, pkg)
            if err != nil {
                return dot.Plan{}, err
            }
            operations = append(operations, pkgPlan.Operations...)
            continue
        }
        
        // Compute current hashes
        pkgPath := filepath.Join(c.config.PackageDir, pkg)
        currentHashes, err := computePackageHashes(ctx, c.config.FS, pkgPath)
        if err != nil {
            return dot.Plan{}, err
        }
        
        // Compare with stored hashes
        oldHashes := pkgInfo.FileHashes
        if oldHashes == nil {
            // No hashes stored - fall back to full unmanage + manage
            pkgPlan, err := c.planFullRemanage(ctx, pkg)
            if err != nil {
                return dot.Plan{}, err
            }
            operations = append(operations, pkgPlan.Operations...)
            continue
        }
        
        changed, added, deleted := compareHashes(oldHashes, currentHashes)
        
        // Generate operations for changes
        
        // Changed files: Delete old link, create new link
        for _, file := range changed {
            targetFile := filepath.Join(c.config.TargetDir, file)
            filePathResult := dot.NewFilePath(targetFile)
            if !filePathResult.IsOk() {
                continue
            }
            
            // Delete old link
            delID := dot.OperationID(fmt.Sprintf("remanage-del-%s-%s", pkg, file))
            operations = append(operations, dot.NewLinkDelete(delID, filePathResult.Unwrap()))
            
            // Create new link
            sourceFile := filepath.Join(pkgPath, file)
            sourcePathResult := dot.NewFilePath(sourceFile)
            if !sourcePathResult.IsOk() {
                continue
            }
            
            createID := dot.OperationID(fmt.Sprintf("remanage-create-%s-%s", pkg, file))
            operations = append(operations, dot.NewLinkCreate(createID, sourcePathResult.Unwrap(), filePathResult.Unwrap()))
        }
        
        // Added files: Create link
        for _, file := range added {
            sourceFile := filepath.Join(pkgPath, file)
            targetFile := filepath.Join(c.config.TargetDir, file)
            
            sourcePathResult := dot.NewFilePath(sourceFile)
            targetPathResult := dot.NewFilePath(targetFile)
            if !sourcePathResult.IsOk() || !targetPathResult.IsOk() {
                continue
            }
            
            id := dot.OperationID(fmt.Sprintf("remanage-add-%s-%s", pkg, file))
            operations = append(operations, dot.NewLinkCreate(id, sourcePathResult.Unwrap(), targetPathResult.Unwrap()))
        }
        
        // Deleted files: Remove link
        for _, file := range deleted {
            targetFile := filepath.Join(c.config.TargetDir, file)
            filePathResult := dot.NewFilePath(targetFile)
            if !filePathResult.IsOk() {
                continue
            }
            
            id := dot.OperationID(fmt.Sprintf("remanage-rm-%s-%s", pkg, file))
            operations = append(operations, dot.NewLinkDelete(id, filePathResult.Unwrap()))
        }
    }
    
    return dot.Plan{
        Operations: operations,
        Metadata: dot.PlanMetadata{
            PackageCount:   len(packages),
            OperationCount: len(operations),
        },
    }, nil
}

// planFullRemanage does full unmanage + manage for fallback cases
func (c *client) planFullRemanage(ctx context.Context, pkg string) (dot.Plan, error) {
    unmanagePlan, err := c.PlanUnmanage(ctx, pkg)
    if err != nil {
        return dot.Plan{}, err
    }
    
    managePlan, err := c.PlanManage(ctx, pkg)
    if err != nil {
        return dot.Plan{}, err
    }
    
    combined := make([]dot.Operation, 0, len(unmanagePlan.Operations)+len(managePlan.Operations))
    combined = append(combined, unmanagePlan.Operations...)
    combined = append(combined, managePlan.Operations...)
    
    return dot.Plan{
        Operations: combined,
        Metadata: dot.PlanMetadata{
            PackageCount:   1,
            OperationCount: len(combined),
        },
    }, nil
}
```

**Commit Message**:
```
feat(api): implement incremental remanage with hash comparison

Replace naive unmanage+manage with intelligent incremental remanage
detecting only changed files using SHA256 hashes.

Algorithm compares stored hashes with current hashes to identify:
- Changed files: Remove old link, create new link
- Added files: Create new link
- Deleted files: Remove link
- Unchanged files: No operation (keep existing link)

Significantly reduces filesystem operations and improves performance
for large packages with minimal changes.

- Implement hash-based change detection
- Generate minimal operation plan
- Fall back to full remanage if no hashes stored
- Add comprehensive tests for all change scenarios

Performance: Remanaging 100-file package with 1 change goes from
200 operations (unmanage all + manage all) to 2 operations
(delete changed + create changed).
```

#### T22.5-004: Update Manifest Writer to Store Hashes
**Estimate**: 1 hour

**Implementation**:
```go
// internal/api/manage.go (in updateManifest)
func (c *client) updateManifest(ctx context.Context, packages []string, plan dot.Plan) error {
    // ... existing code ...
    
    for _, pkg := range packages {
        // ... extract links ...
        
        // Compute and store hashes
        pkgPath := filepath.Join(c.config.PackageDir, pkg)
        hashes, err := computePackageHashes(ctx, c.config.FS, pkgPath)
        if err != nil {
            c.config.Logger.Warn(ctx, "failed_to_compute_hashes", "package", pkg, "error", err)
            hashes = nil  // Continue without hashes (fallback to full remanage)
        }
        
        m.AddPackage(manifest.PackageInfo{
            Name:        pkg,
            InstalledAt: time.Now(),
            LinkCount:   len(links),
            Links:       links,
            FileHashes:  hashes,  // NEW
        })
    }
    
    return c.manifest.Save(ctx, targetPath, m)
}
```

**Commit Message**:
```
feat(api): store file hashes in manifest during manage

Compute and store SHA256 hashes of package files in manifest enabling
future incremental remanage operations.

- Compute hashes during package installation
- Store in manifest FileHashes field
- Handle errors gracefully (continue without hashes)
- Add tests verifying hash storage
```

#### T22.5-005: Update Remanage to Use Incremental Planning
**Estimate**: 0.5 hours

**Implementation**:
```go
// internal/api/remanage.go
func (c *client) Remanage(ctx context.Context, packages ...string) error {
    // Use new incremental planning (no changes needed - already calls PlanRemanage)
    plan, err := c.PlanRemanage(ctx, packages...)
    if err != nil {
        return err
    }
    
    if len(plan.Operations) == 0 {
        c.config.Logger.Info(ctx, "no_changes_detected", "packages", packages)
        return nil
    }
    
    // ... execute plan as before ...
}
```

**Commit Message**:
```
refactor(api): use incremental planning in remanage

Update Remanage to use new incremental PlanRemanage, eliminating
unnecessary filesystem operations for unchanged files.
```

#### T22.5-006: Performance Testing and Documentation
**Estimate**: 1.5 hours

**Benchmark**:
```go
// internal/api/remanage_benchmark_test.go
func BenchmarkRemanage_FullPlan(b *testing.B) {
    // 100 files, all unchanged
    // Measure time for old approach
}

func BenchmarkRemanage_IncrementalPlan(b *testing.B) {
    // 100 files, all unchanged
    // Measure time for new approach
}

func BenchmarkRemanage_OneChanged(b *testing.B) {
    // 100 files, 1 changed
    // Measure operation count
}
```

**Documentation**:
```markdown
# Incremental Remanage

## Overview
Remanage operations use hash-based change detection to minimize filesystem
operations.

## Algorithm
1. Load manifest with stored file hashes
2. Compute current hashes for package files
3. Compare to identify changes:
   - Changed: Delete old link + create new link (2 ops)
   - Added: Create link (1 op)
   - Deleted: Remove link (1 op)
   - Unchanged: No operation (0 ops)

## Performance
For package with N files and C changes:
- Old approach: 2N operations (unmanage all + manage all)
- New approach: ~2C operations (only changed files)

Example: 100 files, 1 change
- Old: 200 operations
- New: 2 operations
- Improvement: 99% reduction

## Fallback Behavior
Falls back to full remanage if:
- No manifest exists (first install)
- Package not in manifest
- No hashes stored (old manifest format)
- Hash computation fails

## Future Enhancements
- Nested directory support
- Parallel hash computation
- Content-addressable storage
```

**Commit Message**:
```
docs(api): document incremental remanage algorithm

Add comprehensive documentation for hash-based incremental remanage
including algorithm, performance characteristics, and fallback behavior.

- Add algorithm description
- Add performance benchmarks
- Document fallback cases
- Add user-facing documentation
```

### Deliverables

- ✅ Hash storage in manifest
- ✅ Hash computation utilities
- ✅ Incremental PlanRemanage implementation
- ✅ Manifest updated to store hashes
- ✅ Performance significantly improved
- ✅ Tests covering all scenarios
- ✅ Documentation complete

---

## Phase 22.6: Future Enhancements Scoping

**Priority**: LOW — Planning and documentation  
**Estimated Effort**: 5-6 hours  
**Files Affected**: Documentation, ADRs, phase plans

### Scope

This phase doesn't implement features but creates detailed plans for:
1. Streaming API for large operations
2. ConfigBuilder for fluent configuration
3. Client struct refactoring (Phase 12b)

### Implementation Tasks

#### T22.6-001: Streaming API Design Document
**Estimate**: 2 hours

**Create**:
```markdown
# ADR-003: Streaming API for Large Operations

## Context
Current API returns all operations in memory, limiting scalability for
very large packages (1000+ files).

## Decision
Implement streaming API using Go channels for memory-efficient operation
processing.

## Proposed API
```go
type StreamingPlan struct {
    Operations <-chan Operation
    Errors     <-chan error
    Metadata   PlanMetadata
}

func (c Client) PlanManageStreaming(ctx context.Context, packages ...string) StreamingPlan
```

## Implementation Plan
- Phase A: Streaming scanner (2-3 hours)
- Phase B: Streaming planner (3-4 hours)
- Phase C: Streaming executor (4-5 hours)
- Phase D: Streaming manifest updates (2-3 hours)

Total: 11-15 hours

## Alternatives Considered
- Iterator pattern: More complex API
- Batch processing: Less memory efficient

## Decision Date
[Future phase]
```

**Commit Message**:
```
docs(adr): add ADR-003 for streaming API design

Document design decision for streaming API to handle large-scale
operations with minimal memory footprint.
```

#### T22.6-002: ConfigBuilder Design Document
**Estimate**: 1.5 hours

**Create**:
```markdown
# ADR-004: Fluent Configuration Builder

## Context
Current Config struct requires all fields set at once, leading to
verbose initialization code.

## Decision
Add optional ConfigBuilder for fluent configuration API while maintaining
struct-based configuration for simplicity.

## Proposed API
```go
cfg := dot.NewConfigBuilder().
    WithPackageDir("~/dotfiles").
    WithTargetDir("~").
    WithDryRun(true).
    WithVerbosity(2).
    Build()
```

## Implementation Plan
- Phase A: Builder struct and methods (2 hours)
- Phase B: Validation and defaults (1 hour)
- Phase C: Tests and documentation (1 hour)

Total: 4 hours

## Trade-offs
- Pro: Fluent, readable API
- Pro: Progressive disclosure of options
- Con: Additional API surface
- Con: Two ways to do same thing

## Decision Date
[Future phase]
```

**Commit Message**:
```
docs(adr): add ADR-004 for fluent configuration builder

Document design for optional ConfigBuilder providing fluent API while
maintaining simple struct-based configuration.
```

#### T22.6-003: Phase 12b Refactoring Plan Update
**Estimate**: 1.5 hours

**Update**: `docs/planning/phase-12b-refactor-plan.md`

Add sections:
- Dependencies on Phase 22
- Updated effort estimates
- Migration strategy refinements
- Risk assessment

**Commit Message**:
```
docs(planning): update Phase 12b plan with Phase 22 learnings

Refine Phase 12b refactoring plan based on completed work in Phase 22,
updated effort estimates, and identified risks.
```

### Deliverables

- ✅ Streaming API design documented
- ✅ ConfigBuilder design documented
- ✅ Phase 12b plan updated
- ✅ Future phases scoped with effort estimates
- ✅ Architecture decision records complete

---

## Phase 22 Timeline

### Week 1: High Priority Features
- Day 1-2: Phase 22.1 (CLI Commands) - Tasks 001-003
- Day 3: Phase 22.1 (CLI Commands) - Tasks 004-005
- Day 4-5: Phase 22.2 (Link Tracking) - Tasks 001-003

### Week 2: High Priority Completion + Medium Priority
- Day 1: Phase 22.2 (Link Tracking) - Tasks 004-005
- Day 2-3: Phase 22.3 (Backup Directory) - All tasks
- Day 4: Phase 22.4 (API Migration) - All tasks

### Week 3: Low Priority Features
- Day 1-3: Phase 22.5 (Incremental Remanage) - All tasks
- Day 4: Phase 22.6 (Future Scoping) - All tasks
- Day 5: Integration testing and documentation finalization

---

## Testing Strategy

### Unit Tests
- Every function has dedicated test(s)
- Table-driven tests for multiple scenarios
- Mock filesystem for deterministic testing
- Error path coverage ≥ 80%

### Integration Tests
- End-to-end CLI command execution
- Multi-package workflows
- Dry-run verification
- Error recovery scenarios

### Benchmark Tests
- Performance comparison (old vs new remanage)
- Memory usage profiling
- Large package handling (100+ files)

### Manual Testing Checklist
```bash
# After Phase 22.1
dot manage vim
dot manage vim zsh git
dot manage nonexistent  # Should error gracefully
dot unmanage vim
dot adopt vim .vimrc
dot remanage vim

# After Phase 22.2
dot status  # Should show accurate link counts
dot doctor  # Should show correct statistics

# After Phase 22.3
dot manage vim --backup-dir /custom/path
# Trigger conflict, verify backup location

# After Phase 22.5
dot remanage vim  # With no changes, should be fast
# Modify one file in vim package
dot remanage vim  # Should only update changed file
```

---

## Documentation Updates

### User Documentation
- `docs/user/05-commands.md`: Update all command examples
- `docs/user/04-configuration.md`: Add backup_dir configuration
- `docs/user/07-advanced.md`: Add incremental remanage section
- `docs/user/08-troubleshooting.md`: Add new error scenarios

### Developer Documentation
- `docs/architecture/adr/`: Add ADR-002, ADR-003, ADR-004
- `docs/developer/`: Update contribution guide if API changes
- `CONTRIBUTING.md`: Update testing requirements

### API Documentation
- `pkg/dot/doc.go`: Update examples
- Add godoc comments for all new public methods
- Update migration guide

### Changelog
Add to CHANGELOG.md under `[Unreleased]`:
```markdown
### Added
- CLI commands now fully functional (manage, unmanage, remanage, adopt)
- Per-package link tracking in manifest
- Backup directory configuration (flag, config file, environment variable)
- DoctorWithScan method for explicit scan configuration
- Incremental remanage with hash-based change detection
- File hash storage in manifest for change detection

### Changed
- Manifest now accurately tracks links per package
- Remanage operations use incremental planning by default
- Backup directory configuration now functional throughout system

### Deprecated
- Doctor(ctx) method (use DoctorWithScan instead)

### Fixed
- CLI commands no longer stubbed, fully implemented
- Manifest link count now shows actual count instead of 0
- Backup directory configuration now properly wired through pipeline
```

---

## Risk Mitigation

### Risk: Breaking Changes to Manifest Format
**Mitigation**:
- Use optional fields (omitempty) for backward compatibility
- Implement manifest migration if needed
- Test with old manifest formats

### Risk: Performance Regression in Remanage
**Mitigation**:
- Benchmark before and after
- Ensure fallback to full remanage works
- Add performance tests to CI

### Risk: CLI Incompatibility
**Mitigation**:
- Maintain all existing flags
- Add deprecation warnings before removal
- Document migration paths

### Risk: Hash Computation Failures
**Mitigation**:
- Graceful degradation (fall back to full remanage)
- Log warnings but don't fail operations
- Test with permission errors, missing files

---

## Success Criteria

### Functional
- ✅ All CLI commands execute successfully
- ✅ Manifest tracks links accurately
- ✅ Backup directory configurable and functional
- ✅ Remanage only processes changed files
- ✅ API migration path documented

### Technical
- ✅ Test coverage ≥ 80% across all packages
- ✅ All linters pass with zero warnings
- ✅ No performance regressions
- ✅ Backward compatibility maintained
- ✅ All integration tests pass

### Documentation
- ✅ User documentation complete and accurate
- ✅ Developer documentation updated
- ✅ API documentation comprehensive
- ✅ Migration guides provided
- ✅ ADRs document key decisions

### Quality
- ✅ Atomic commits following Conventional Commits
- ✅ TDD approach throughout
- ✅ No TODO comments remaining
- ✅ Code review ready

---

## Post-Phase Validation

### Validation Checklist
```bash
# Build and install
make build
make install

# Run test suite
make test
make test-integration
make coverage

# Lint and format
make lint
make fmt

# Manual validation
dot manage vim zsh git
dot status
dot list
dot doctor --scan-mode scoped
dot remanage vim
dot unmanage vim zsh git

# Configuration
dot config init
dot config set symlinks.backup_dir /custom/backup
dot manage vim --backup-dir /tmp/backup

# Verify all TODO comments resolved
grep -r "TODO" internal/ cmd/ pkg/ --exclude="*_test.go" | grep -v "docs/"
```

### Pre-Merge Requirements
- [ ] All tasks complete
- [ ] All tests passing
- [ ] Coverage ≥ 80%
- [ ] All linters passing
- [ ] Documentation complete
- [ ] CHANGELOG updated
- [ ] No TODO comments in production code
- [ ] Manual validation passed
- [ ] Code review approved

---

## Future Phases

After Phase 22 completion:

**Phase 23**: Cross-Platform Testing and Validation
- Windows compatibility testing
- macOS testing
- Linux distribution testing
- CI/CD pipeline enhancements

**Phase 24**: Performance Optimization
- Parallel operation execution
- Improved caching
- Memory usage optimization
- Large-scale performance testing

**Phase 25**: Streaming API Implementation (ADR-003)
- Streaming scanner
- Streaming planner
- Streaming executor
- Memory-efficient large package handling

**Phase 26**: ConfigBuilder Implementation (ADR-004)
- Fluent configuration API
- Progressive option disclosure
- Validation and defaults

**Phase 27**: Release Preparation (Phase 20 revisited)
- Final quality assurance
- Release automation
- Distribution testing
- v0.2.0 release

---

## Conclusion

Phase 22 completes all identified stubbed and incomplete features, establishing
full functionality for the dot CLI. The phase prioritizes user-facing features
first (CLI commands), then correctness (link tracking), then configuration
(backup directory), then developer experience (API migration), then performance
(incremental remanage), and finally future planning.

Estimated total effort: 32-40 hours over 3 weeks.

Upon completion, dot will have:
- Fully functional CLI matching documented commands
- Accurate manifest tracking
- Complete configuration system
- Optimized remanage operations
- Clear migration paths for API consumers
- Documented future enhancement plans

This phase establishes dot as a production-ready dotfile manager.

