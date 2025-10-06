# Phase 7 Implementation Plan: Functional Core - Resolver

## Overview

Phase 7 implements pure conflict detection and resolution logic. The resolver analyzes planned operations, detects conflicts with current state, applies resolution policies, and generates actionable warnings and suggestions.

## Design Principles

- **Pure Functions**: All resolver logic is side-effect-free
- **Test-First Development**: Write tests before implementation
- **Atomic Commits**: One discrete change per commit
- **Type-Driven**: Leverage type safety for correctness
- **Comprehensive Error Handling**: All conflicts detected and reported

## Dependencies

**Required Phases**:
- Phase 1: Domain Model (Path types, Result monad, Operation types)
- Phase 2: Infrastructure Ports (interfaces defined)
- Phase 6: Functional Core - Planner (PlanResult, DesiredState, CurrentState)

**Phase Location**: `internal/planner/` (resolver is part of planning stage)

## Architecture

### Package Structure

```
internal/planner/
├── desired.go          # Existing: desired state computation
├── desired_test.go
├── resolver.go         # New: conflict detection and resolution
├── resolver_test.go
├── policies.go         # New: resolution policy implementations
├── policies_test.go
├── suggestions.go      # New: warning and suggestion generation
└── suggestions_test.go
```

### Type Hierarchy

```go
// Core types
type Conflict struct
type ConflictType int
type ResolutionPolicies struct
type ResolutionPolicy int
type ResolveResult struct
type Warning struct
type Suggestion struct

// Resolution status
type ResolutionStatus int
type ResolutionOutcome struct
```

## Implementation Tasks

### Task 7.1: Core Conflict Types and Detection

#### 7.1.1: Define Conflict Type Hierarchy

**Test**: `TestConflictTypeString`
```go
func TestConflictTypeString(t *testing.T) {
    tests := []struct {
        ct   ConflictType
        want string
    }{
        {ConflictFileExists, "file_exists"},
        {ConflictWrongLink, "wrong_link"},
        {ConflictPermission, "permission"},
        {ConflictCircular, "circular"},
    }
    // Verify string representation
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// ConflictType categorizes conflicts by their nature
type ConflictType int

const (
    ConflictFileExists ConflictType = iota // File exists at link target
    ConflictWrongLink                       // Symlink points to wrong source
    ConflictPermission                      // Permission denied
    ConflictCircular                        // Circular symlink dependency
    ConflictDirExpected                     // Directory expected, file found
    ConflictFileExpected                    // File expected, directory found
)

func (ct ConflictType) String() string
```

**Commit**: `feat(planner): define conflict type enumeration`

#### 7.1.2: Implement Conflict Value Object

**Test**: `TestConflictCreation`
```go
func TestConflictCreation(t *testing.T) {
    targetPath := mustParseTargetPath("/home/user/.bashrc")
    
    conflict := NewConflict(
        ConflictFileExists,
        targetPath,
        "File exists at target location",
    )
    
    assert.Equal(t, ConflictFileExists, conflict.Type)
    assert.Equal(t, targetPath, conflict.Path)
    assert.NotEmpty(t, conflict.Details)
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// Conflict represents a detected conflict during planning
type Conflict struct {
    Type        ConflictType
    Path        dot.TargetPath
    Details     string
    Context     map[string]string // Additional context
    Suggestions []Suggestion
}

func NewConflict(ct ConflictType, path dot.TargetPath, details string) Conflict
func (c Conflict) WithContext(key, value string) Conflict
func (c Conflict) WithSuggestion(s Suggestion) Conflict
```

**Commit**: `feat(planner): implement conflict value object`

#### 7.1.3: Define Resolution Status Types

**Test**: `TestResolutionStatus`
```go
func TestResolutionStatus(t *testing.T) {
    tests := []struct {
        status ResolutionStatus
        want   string
    }{
        {ResolveOK, "ok"},
        {ResolveConflict, "conflict"},
        {ResolveWarning, "warning"},
        {ResolveSkip, "skip"},
    }
    // Verify status values
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// ResolutionStatus indicates the outcome of conflict resolution
type ResolutionStatus int

const (
    ResolveOK       ResolutionStatus = iota // No conflict, proceed
    ResolveConflict                         // Unresolved conflict, fail
    ResolveWarning                          // Resolved with warning
    ResolveSkip                             // Operation skipped
)

// ResolutionOutcome captures resolution results
type ResolutionOutcome struct {
    Status      ResolutionStatus
    Operations  []dot.Operation  // Modified operations
    Conflict    *Conflict        // If status is ResolveConflict
    Warning     *Warning         // If status is ResolveWarning
}
```

**Commit**: `feat(planner): define resolution status types`

#### 7.1.4: Implement ResolveResult Type

**Test**: `TestResolveResultConstruction`
```go
func TestResolveResultConstruction(t *testing.T) {
    t.Run("with operations", func(t *testing.T) {
        ops := []dot.Operation{/* test operations */}
        result := NewResolveResult(ops)
        assert.Len(t, result.Operations, len(ops))
        assert.Empty(t, result.Conflicts)
        assert.Empty(t, result.Warnings)
    })
    
    t.Run("with conflicts", func(t *testing.T) {
        result := NewResolveResult(nil)
        result = result.WithConflict(conflict)
        assert.Len(t, result.Conflicts, 1)
    })
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// ResolveResult contains all resolved operations, conflicts, and warnings
type ResolveResult struct {
    Operations []dot.Operation
    Conflicts  []Conflict
    Warnings   []Warning
}

func NewResolveResult(ops []dot.Operation) ResolveResult
func (r ResolveResult) WithConflict(c Conflict) ResolveResult
func (r ResolveResult) WithWarning(w Warning) ResolveResult
func (r ResolveResult) HasConflicts() bool
func (r ResolveResult) ConflictCount() int
func (r ResolveResult) WarningCount() int
```

**Commit**: `feat(planner): implement resolve result type`

#### 7.1.5: Detect File Exists Conflicts

**Test**: `TestDetectFileExistsConflict`
```go
func TestDetectFileExistsConflict(t *testing.T) {
    targetPath := mustParseTargetPath("/home/user/.bashrc")
    sourcePath := mustParseFilePath("/stow/bash/dot-bashrc")
    
    op := dot.LinkCreate{
        ID:     "link-1",
        Source: sourcePath,
        Target: targetPath,
    }
    
    current := CurrentState{
        Files: map[dot.TargetPath]FileInfo{
            targetPath: {Size: 100},
        },
    }
    
    outcome := detectLinkCreateConflicts(op, current)
    
    assert.Equal(t, ResolveConflict, outcome.Status)
    assert.NotNil(t, outcome.Conflict)
    assert.Equal(t, ConflictFileExists, outcome.Conflict.Type)
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// detectLinkCreateConflicts checks for conflicts with LinkCreate operation
func detectLinkCreateConflicts(
    op dot.LinkCreate,
    current CurrentState,
) ResolutionOutcome {
    // Check if file exists at target
    if fileInfo, exists := current.Files[op.Target]; exists {
        conflict := NewConflict(
            ConflictFileExists,
            op.Target,
            fmt.Sprintf("File exists at target: size=%d", fileInfo.Size),
        )
        return ResolutionOutcome{
            Status:   ResolveConflict,
            Conflict: &conflict,
        }
    }
    
    return ResolutionOutcome{
        Status:     ResolveOK,
        Operations: []dot.Operation{op},
    }
}
```

**Commit**: `feat(planner): detect file exists conflicts`

#### 7.1.6: Detect Wrong Link Conflicts

**Test**: `TestDetectWrongLinkConflict`
```go
func TestDetectWrongLinkConflict(t *testing.T) {
    targetPath := mustParseTargetPath("/home/user/.bashrc")
    sourcePath := mustParseFilePath("/stow/bash/dot-bashrc")
    wrongPath := mustParseFilePath("/stow/other/dot-bashrc")
    
    op := dot.LinkCreate{
        ID:     "link-1",
        Source: sourcePath,
        Target: targetPath,
    }
    
    current := CurrentState{
        Links: map[dot.TargetPath]LinkTarget{
            targetPath: {Target: wrongPath},
        },
    }
    
    outcome := detectLinkCreateConflicts(op, current)
    
    assert.Equal(t, ResolveConflict, outcome.Status)
    assert.NotNil(t, outcome.Conflict)
    assert.Equal(t, ConflictWrongLink, outcome.Conflict.Type)
}
```

**Implementation**: Update `detectLinkCreateConflicts`
```go
func detectLinkCreateConflicts(
    op dot.LinkCreate,
    current CurrentState,
) ResolutionOutcome {
    // Check if symlink exists pointing elsewhere
    if linkTarget, exists := current.Links[op.Target]; exists {
        if linkTarget.Target != op.Source {
            conflict := NewConflict(
                ConflictWrongLink,
                op.Target,
                fmt.Sprintf("Symlink points to %s, expected %s",
                    linkTarget.Target, op.Source),
            )
            return ResolutionOutcome{
                Status:   ResolveConflict,
                Conflict: &conflict,
            }
        }
        // Link already correct, no-op
        return ResolutionOutcome{Status: ResolveSkip}
    }
    
    // ... rest of checks
}
```

**Commit**: `feat(planner): detect wrong link conflicts`

#### 7.1.7: Detect Directory/File Mismatches

**Test**: `TestDetectTypeMismatchConflicts`
```go
func TestDetectTypeMismatchConflicts(t *testing.T) {
    t.Run("directory expected file found", func(t *testing.T) {
        // Test directory creation when file exists
    })
    
    t.Run("file expected directory found", func(t *testing.T) {
        // Test link creation when directory exists
    })
}
```

**Implementation**: `internal/planner/resolver.go`
```go
func detectDirCreateConflicts(
    op dot.DirCreate,
    current CurrentState,
) ResolutionOutcome {
    // Check if file exists where directory expected
    if _, exists := current.Files[op.Path]; exists {
        conflict := NewConflict(
            ConflictFileExpected,
            op.Path,
            "File exists where directory expected",
        )
        return ResolutionOutcome{
            Status:   ResolveConflict,
            Conflict: &conflict,
        }
    }
    
    return ResolutionOutcome{
        Status:     ResolveOK,
        Operations: []dot.Operation{op},
    }
}
```

**Commit**: `feat(planner): detect directory/file mismatch conflicts`

### Task 7.2: Resolution Policies

#### 7.2.1: Define Resolution Policy Types

**Test**: `TestResolutionPolicyTypes`
```go
func TestResolutionPolicyTypes(t *testing.T) {
    tests := []struct {
        policy ResolutionPolicy
        want   string
    }{
        {PolicyFail, "fail"},
        {PolicyBackup, "backup"},
        {PolicyOverwrite, "overwrite"},
        {PolicySkip, "skip"},
    }
    // Verify policy names
}
```

**Implementation**: `internal/planner/policies.go`
```go
// ResolutionPolicy defines how to handle conflicts
type ResolutionPolicy int

const (
    PolicyFail      ResolutionPolicy = iota // Stop and report (default)
    PolicyBackup                             // Backup conflicting file
    PolicyOverwrite                          // Replace conflicting file
    PolicySkip                               // Skip conflicting operation
)

func (rp ResolutionPolicy) String() string
```

**Commit**: `feat(planner): define resolution policy types`

#### 7.2.2: Implement Resolution Policies Configuration

**Test**: `TestResolutionPoliciesConfiguration`
```go
func TestResolutionPoliciesConfiguration(t *testing.T) {
    policies := ResolutionPolicies{
        OnFileExists:    PolicyBackup,
        OnWrongLink:     PolicyOverwrite,
        OnPermissionErr: PolicyFail,
    }
    
    assert.Equal(t, PolicyBackup, policies.OnFileExists)
    assert.Equal(t, PolicyOverwrite, policies.OnWrongLink)
}
```

**Implementation**: `internal/planner/policies.go`
```go
// ResolutionPolicies configures conflict resolution behavior
type ResolutionPolicies struct {
    OnFileExists    ResolutionPolicy
    OnWrongLink     ResolutionPolicy
    OnPermissionErr ResolutionPolicy
    OnCircular      ResolutionPolicy
    OnTypeMismatch  ResolutionPolicy
}

// DefaultPolicies returns safe defaults (all fail)
func DefaultPolicies() ResolutionPolicies {
    return ResolutionPolicies{
        OnFileExists:    PolicyFail,
        OnWrongLink:     PolicyFail,
        OnPermissionErr: PolicyFail,
        OnCircular:      PolicyFail,
        OnTypeMismatch:  PolicyFail,
    }
}
```

**Commit**: `feat(planner): implement resolution policies configuration`

#### 7.2.3: Implement PolicyFail

**Test**: `TestPolicyFail`
```go
func TestPolicyFail(t *testing.T) {
    conflict := NewConflict(
        ConflictFileExists,
        targetPath,
        "File exists",
    )
    
    outcome := applyFailPolicy(conflict)
    
    assert.Equal(t, ResolveConflict, outcome.Status)
    assert.NotNil(t, outcome.Conflict)
    assert.Empty(t, outcome.Operations)
}
```

**Implementation**: `internal/planner/policies.go`
```go
// applyFailPolicy returns unresolved conflict
func applyFailPolicy(c Conflict) ResolutionOutcome {
    return ResolutionOutcome{
        Status:   ResolveConflict,
        Conflict: &c,
    }
}
```

**Commit**: `feat(planner): implement fail resolution policy`

#### 7.2.4: Implement PolicyBackup

**Test**: `TestPolicyBackup`
```go
func TestPolicyBackup(t *testing.T) {
    op := dot.LinkCreate{
        ID:     "link-1",
        Source: sourcePath,
        Target: targetPath,
    }
    
    conflict := NewConflict(ConflictFileExists, targetPath, "File exists")
    
    outcome := applyBackupPolicy(op, conflict, "/backup")
    
    assert.Equal(t, ResolveWarning, outcome.Status)
    assert.Len(t, outcome.Operations, 2) // backup + link
    
    backupOp := outcome.Operations[0].(dot.FileBackup)
    assert.Equal(t, targetPath, backupOp.Path)
    
    linkOp := outcome.Operations[1].(dot.LinkCreate)
    assert.Equal(t, op, linkOp)
}
```

**Implementation**: `internal/planner/policies.go`
```go
// applyBackupPolicy creates backup operation before link
func applyBackupPolicy(
    op dot.LinkCreate,
    c Conflict,
    backupDir string,
) ResolutionOutcome {
    // Generate backup path
    backupPath := generateBackupPath(op.Target, backupDir)
    
    // Create backup operation
    backupOp := dot.FileBackup{
        ID:   dot.OperationID(fmt.Sprintf("backup-%s", op.ID)),
        Path: op.Target,
        Dest: backupPath,
    }
    
    warning := Warning{
        Message: fmt.Sprintf(
            "Backing up existing file: %s -> %s",
            op.Target, backupPath,
        ),
        Severity: WarnInfo,
    }
    
    return ResolutionOutcome{
        Status:     ResolveWarning,
        Operations: []dot.Operation{backupOp, op},
        Warning:    &warning,
    }
}
```

**Commit**: `feat(planner): implement backup resolution policy`

#### 7.2.5: Implement PolicyOverwrite

**Test**: `TestPolicyOverwrite`
```go
func TestPolicyOverwrite(t *testing.T) {
    op := dot.LinkCreate{
        ID:     "link-1",
        Source: sourcePath,
        Target: targetPath,
    }
    
    conflict := NewConflict(ConflictFileExists, targetPath, "File exists")
    
    outcome := applyOverwritePolicy(op, conflict)
    
    assert.Equal(t, ResolveWarning, outcome.Status)
    assert.Len(t, outcome.Operations, 2) // delete + link
    
    deleteOp := outcome.Operations[0].(dot.FileDelete)
    assert.Equal(t, targetPath, deleteOp.Path)
}
```

**Implementation**: `internal/planner/policies.go`
```go
// applyOverwritePolicy creates delete then link operations
func applyOverwritePolicy(
    op dot.LinkCreate,
    c Conflict,
) ResolutionOutcome {
    // Create delete operation
    deleteOp := dot.FileDelete{
        ID:   dot.OperationID(fmt.Sprintf("delete-%s", op.ID)),
        Path: op.Target,
    }
    
    warning := Warning{
        Message: fmt.Sprintf(
            "Overwriting existing file: %s",
            op.Target,
        ),
        Severity: WarnCaution,
    }
    
    return ResolutionOutcome{
        Status:     ResolveWarning,
        Operations: []dot.Operation{deleteOp, op},
        Warning:    &warning,
    }
}
```

**Commit**: `feat(planner): implement overwrite resolution policy`

#### 7.2.6: Implement PolicySkip

**Test**: `TestPolicySkip`
```go
func TestPolicySkip(t *testing.T) {
    op := dot.LinkCreate{
        ID:     "link-1",
        Source: sourcePath,
        Target: targetPath,
    }
    
    conflict := NewConflict(ConflictFileExists, targetPath, "File exists")
    
    outcome := applySkipPolicy(op, conflict)
    
    assert.Equal(t, ResolveSkip, outcome.Status)
    assert.Empty(t, outcome.Operations)
    assert.NotNil(t, outcome.Warning)
}
```

**Implementation**: `internal/planner/policies.go`
```go
// applySkipPolicy skips operation with warning
func applySkipPolicy(
    op dot.LinkCreate,
    c Conflict,
) ResolutionOutcome {
    warning := Warning{
        Message: fmt.Sprintf(
            "Skipping due to conflict: %s",
            op.Target,
        ),
        Severity: WarnInfo,
    }
    
    return ResolutionOutcome{
        Status:  ResolveSkip,
        Warning: &warning,
    }
}
```

**Commit**: `feat(planner): implement skip resolution policy`

#### 7.2.7: Implement Policy Dispatcher

**Test**: `TestResolveOperationDispatch`
```go
func TestResolveOperationDispatch(t *testing.T) {
    policies := DefaultPolicies()
    policies.OnFileExists = PolicyBackup
    
    op := dot.LinkCreate{/* ... */}
    current := CurrentState{/* file exists */}
    
    outcome := resolveOperation(op, current, policies, "/backup")
    
    assert.Equal(t, ResolveWarning, outcome.Status)
    assert.Len(t, outcome.Operations, 2) // backup + link
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// resolveOperation applies policies to detected conflicts
func resolveOperation(
    op dot.Operation,
    current CurrentState,
    policies ResolutionPolicies,
    backupDir string,
) ResolutionOutcome {
    switch op := op.(type) {
    case dot.LinkCreate:
        return resolveLinkCreate(op, current, policies, backupDir)
    case dot.DirCreate:
        return resolveDirCreate(op, current, policies)
    case dot.LinkDelete:
        return resolveLinkDelete(op, current, policies)
    default:
        // Operations without conflict potential pass through
        return ResolutionOutcome{
            Status:     ResolveOK,
            Operations: []dot.Operation{op},
        }
    }
}

func resolveLinkCreate(
    op dot.LinkCreate,
    current CurrentState,
    policies ResolutionPolicies,
    backupDir string,
) ResolutionOutcome {
    // Detect conflicts
    outcome := detectLinkCreateConflicts(op, current)
    if outcome.Status == ResolveOK {
        return outcome
    }
    
    // Apply policy based on conflict type
    conflict := *outcome.Conflict
    switch conflict.Type {
    case ConflictFileExists:
        return applyPolicyForConflict(
            op, conflict, policies.OnFileExists, backupDir,
        )
    case ConflictWrongLink:
        return applyPolicyForConflict(
            op, conflict, policies.OnWrongLink, backupDir,
        )
    default:
        return outcome
    }
}
```

**Commit**: `feat(planner): implement resolution policy dispatcher`

### Task 7.3: Warning and Suggestion System

#### 7.3.1: Define Warning Type

**Test**: `TestWarningCreation`
```go
func TestWarningCreation(t *testing.T) {
    warning := Warning{
        Message:  "File will be backed up",
        Severity: WarnInfo,
        Context: map[string]string{
            "path": "/home/user/.bashrc",
        },
    }
    
    assert.NotEmpty(t, warning.Message)
    assert.Equal(t, WarnInfo, warning.Severity)
}
```

**Implementation**: `internal/planner/suggestions.go`
```go
// Warning represents a non-fatal issue
type Warning struct {
    Message  string
    Severity WarningSeverity
    Context  map[string]string
}

type WarningSeverity int

const (
    WarnInfo    WarningSeverity = iota // Informational
    WarnCaution                         // Requires attention
    WarnDanger                          // Potentially destructive
)

func (ws WarningSeverity) String() string
```

**Commit**: `feat(planner): define warning type`

#### 7.3.2: Define Suggestion Type

**Test**: `TestSuggestionCreation`
```go
func TestSuggestionCreation(t *testing.T) {
    suggestion := Suggestion{
        Action:      "Use --backup flag",
        Explanation: "This will preserve the existing file",
        Example:     "dot stow --backup bash",
    }
    
    assert.NotEmpty(t, suggestion.Action)
    assert.NotEmpty(t, suggestion.Explanation)
}
```

**Implementation**: `internal/planner/suggestions.go`
```go
// Suggestion provides actionable resolution guidance
type Suggestion struct {
    Action      string // What to do
    Explanation string // Why this helps
    Example     string // Example command (optional)
}
```

**Commit**: `feat(planner): define suggestion type`

#### 7.3.3: Implement Suggestion Templates

**Test**: `TestSuggestionGeneration`
```go
func TestSuggestionGeneration(t *testing.T) {
    t.Run("file exists conflict", func(t *testing.T) {
        conflict := NewConflict(
            ConflictFileExists,
            targetPath,
            "File exists",
        )
        
        suggestions := generateSuggestions(conflict)
        
        assert.NotEmpty(t, suggestions)
        assert.Contains(t, suggestions[0].Action, "backup")
    })
}
```

**Implementation**: `internal/planner/suggestions.go`
```go
// generateSuggestions creates actionable suggestions for conflicts
func generateSuggestions(c Conflict) []Suggestion {
    switch c.Type {
    case ConflictFileExists:
        return []Suggestion{
            {
                Action:      "Use --backup flag to preserve existing file",
                Explanation: "Moves conflicting file to backup location before linking",
                Example:     fmt.Sprintf("dot stow --backup <package>"),
            },
            {
                Action:      "Use dot adopt to move file into package",
                Explanation: "Incorporates existing file into package management",
                Example:     fmt.Sprintf("dot adopt <package> %s", c.Path),
            },
            {
                Action:      "Remove conflicting file manually",
                Explanation: "Delete file if no longer needed",
                Example:     fmt.Sprintf("rm %s", c.Path),
            },
        }
    
    case ConflictWrongLink:
        return []Suggestion{
            {
                Action:      "Unstow the other package first",
                Explanation: "Removes conflicting symlink from different package",
                Example:     "dot unstow <other-package>",
            },
            {
                Action:      "Use --overwrite to replace the link",
                Explanation: "Forces link to point to new package",
                Example:     "dot stow --overwrite <package>",
            },
        }
    
    case ConflictPermission:
        return []Suggestion{
            {
                Action:      "Check file permissions on target directory",
                Explanation: "Ensure you have write access",
                Example:     fmt.Sprintf("ls -ld %s", c.Path.Parent()),
            },
            {
                Action:      "Run with appropriate permissions",
                Explanation: "May need sudo for system directories",
            },
        }
    
    default:
        return nil
    }
}
```

**Commit**: `feat(planner): implement suggestion generation`

#### 7.3.4: Attach Suggestions to Conflicts

**Test**: `TestConflictWithSuggestions`
```go
func TestConflictWithSuggestions(t *testing.T) {
    conflict := NewConflict(
        ConflictFileExists,
        targetPath,
        "File exists",
    )
    
    conflict = enrichConflictWithSuggestions(conflict)
    
    assert.NotEmpty(t, conflict.Suggestions)
    assert.GreaterOrEqual(t, len(conflict.Suggestions), 2)
}
```

**Implementation**: `internal/planner/suggestions.go`
```go
// enrichConflictWithSuggestions adds suggestions to conflict
func enrichConflictWithSuggestions(c Conflict) Conflict {
    c.Suggestions = generateSuggestions(c)
    return c
}
```

**Commit**: `feat(planner): attach suggestions to conflicts`

### Task 7.4: Main Resolver Function

#### 7.4.1: Implement Resolve Entry Point

**Test**: `TestResolveFunction`
```go
func TestResolveFunction(t *testing.T) {
    ops := []dot.Operation{
        dot.LinkCreate{/* ... */},
        dot.DirCreate{/* ... */},
    }
    
    current := CurrentState{/* ... */}
    policies := DefaultPolicies()
    
    result := Resolve(ops, current, policies, "/backup")
    
    assert.NotNil(t, result)
    assert.GreaterOrEqual(t, len(result.Operations), 0)
}
```

**Implementation**: `internal/planner/resolver.go`
```go
// Resolve applies conflict resolution to operations
func Resolve(
    operations []dot.Operation,
    current CurrentState,
    policies ResolutionPolicies,
    backupDir string,
) ResolveResult {
    result := NewResolveResult(nil)
    
    for _, op := range operations {
        outcome := resolveOperation(op, current, policies, backupDir)
        
        switch outcome.Status {
        case ResolveOK:
            result.Operations = append(result.Operations, outcome.Operations...)
        
        case ResolveWarning:
            result.Operations = append(result.Operations, outcome.Operations...)
            if outcome.Warning != nil {
                result = result.WithWarning(*outcome.Warning)
            }
        
        case ResolveConflict:
            if outcome.Conflict != nil {
                enriched := enrichConflictWithSuggestions(*outcome.Conflict)
                result = result.WithConflict(enriched)
            }
        
        case ResolveSkip:
            if outcome.Warning != nil {
                result = result.WithWarning(*outcome.Warning)
            }
        }
    }
    
    return result
}
```

**Commit**: `feat(planner): implement main resolve function`

#### 7.4.2: Add Conflict Aggregation

**Test**: `TestConflictAggregation`
```go
func TestConflictAggregation(t *testing.T) {
    ops := []dot.Operation{
        // Multiple operations with conflicts
    }
    
    result := Resolve(ops, current, policies, "/backup")
    
    // All conflicts should be collected
    assert.GreaterOrEqual(t, result.ConflictCount(), 2)
}
```

**Implementation**: Already handled by `ResolveResult.WithConflict`

**Commit**: `test(planner): verify conflict aggregation`

### Task 7.5: Integration with Planner

#### 7.5.1: Update PlanResult to Include ResolveResult

**Test**: `TestPlanResultWithResolution`
```go
func TestPlanResultWithResolution(t *testing.T) {
    planResult := PlanResult{
        Desired:   desired,
        Current:   current,
        Diff:      ops,
        Resolved:  &resolveResult,
    }
    
    assert.NotNil(t, planResult.Resolved)
    assert.True(t, planResult.HasConflicts())
}
```

**Implementation**: Update `internal/planner/desired.go`
```go
// Update PlanResult to include resolution
type PlanResult struct {
    Desired  DesiredState
    Current  CurrentState
    Diff     []dot.Operation
    Resolved *ResolveResult // Optional resolution results
}

func (pr PlanResult) HasConflicts() bool {
    return pr.Resolved != nil && pr.Resolved.HasConflicts()
}
```

**Commit**: `feat(planner): integrate resolver with plan result`

#### 7.5.2: Add Resolve Step to Planning Pipeline

**Test**: `TestPlanningWithResolution`
```go
func TestPlanningWithResolution(t *testing.T) {
    packages := []Package{/* ... */}
    policies := DefaultPolicies()
    policies.OnFileExists = PolicyBackup
    
    result := PlanWithResolution(packages, target, policies)
    
    assert.NotNil(t, result.Resolved)
    if result.HasConflicts() {
        t.Log("Conflicts:", result.Resolved.Conflicts)
    }
}
```

**Implementation**: `internal/planner/desired.go`
```go
// PlanWithResolution plans and resolves conflicts
func PlanWithResolution(
    packages []Package,
    targetTree FileTree,
    policies ResolutionPolicies,
    backupDir string,
) PlanResult {
    // Compute desired and current state
    desired := computeDesiredState(packages)
    current := computeCurrentState(targetTree)
    
    // Generate diff
    diff := diffStates(desired, current)
    
    // Resolve conflicts
    resolved := Resolve(diff, current, policies, backupDir)
    
    return PlanResult{
        Desired:  desired,
        Current:  current,
        Diff:     diff,
        Resolved: &resolved,
    }
}
```

**Commit**: `feat(planner): add resolution to planning pipeline`

## Testing Strategy

### Unit Tests
- Test each conflict type detection independently
- Test each resolution policy in isolation
- Test suggestion generation for all conflict types
- Test conflict enrichment with suggestions
- Verify warning severity levels

### Property Tests
- Resolution is deterministic (same input → same output)
- All conflicts are detected (no silent failures)
- Resolved operations maintain validity
- Suggestions are always actionable

### Integration Tests
- Test complete resolution workflow
- Test multiple conflicts in single operation set
- Test policy combinations
- Test conflict aggregation across packages

### Edge Cases
- Empty operation list
- All operations conflict
- Mixed conflict types
- Recursive resolution scenarios

## Validation Criteria

### Phase 7.1 Complete When:
- [ ] All conflict types defined and tested
- [ ] Conflict detection works for all operation types
- [ ] Resolution status types implemented
- [ ] ResolveResult type fully functional
- [ ] All detection tests pass
- [ ] Coverage ≥ 80%

### Phase 7.2 Complete When:
- [ ] All resolution policies defined
- [ ] Policy configuration implemented
- [ ] All policies tested independently
- [ ] Policy dispatcher works correctly
- [ ] Integration tests pass
- [ ] Coverage ≥ 80%

### Phase 7.3 Complete When:
- [ ] Warning type implemented
- [ ] Suggestion type implemented
- [ ] Suggestion generation works for all conflicts
- [ ] Conflict enrichment functional
- [ ] All suggestion tests pass
- [ ] Coverage ≥ 80%

### Phase 7 Complete When:
- [ ] All subsections complete
- [ ] Main Resolve function tested
- [ ] Integration with planner complete
- [ ] All linters pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated

## Commit Sequence Summary

1. `feat(planner): define conflict type enumeration`
2. `feat(planner): implement conflict value object`
3. `feat(planner): define resolution status types`
4. `feat(planner): implement resolve result type`
5. `feat(planner): detect file exists conflicts`
6. `feat(planner): detect wrong link conflicts`
7. `feat(planner): detect directory/file mismatch conflicts`
8. `feat(planner): define resolution policy types`
9. `feat(planner): implement resolution policies configuration`
10. `feat(planner): implement fail resolution policy`
11. `feat(planner): implement backup resolution policy`
12. `feat(planner): implement overwrite resolution policy`
13. `feat(planner): implement skip resolution policy`
14. `feat(planner): implement resolution policy dispatcher`
15. `feat(planner): define warning type`
16. `feat(planner): define suggestion type`
17. `feat(planner): implement suggestion generation`
18. `feat(planner): attach suggestions to conflicts`
19. `feat(planner): implement main resolve function`
20. `test(planner): verify conflict aggregation`
21. `feat(planner): integrate resolver with plan result`
22. `feat(planner): add resolution to planning pipeline`
23. `docs(planner): document resolver architecture`
24. `chore(changelog): add Phase 7 completion entry`

## Dependencies on Domain Model

Ensure these types exist in `pkg/dot/`:
- `Operation` interface
- `LinkCreate`, `LinkDelete`, `DirCreate`, `FileDelete`, `FileBackup` operations
- `OperationID` type
- `TargetPath`, `FilePath` phantom types

## Notes

- All resolver logic is pure (no I/O)
- Resolver does not modify filesystem
- Policies are configuration, not behavior
- Suggestions are templates, can be customized
- Warning severity guides user attention
- Multiple conflicts collected, not fail-fast
- Resolution preserves operation dependencies

