# Triage Improvements

> **Status: historical plan record.** This document is a proposal written before
> the triage work landed, not a description of current behavior. Items below are
> annotated with their implementing `file:line` where they have since shipped.
> Last verified against code 2026-07-27.

## Original Issues (as of the plan date)

### Critical Bugs
1. **Line 365 bug**: Apply-to-all logic checks `choice == "a" && strings.Contains(choice, "A")` which is always false. FIXED at `pkg/dot/doctor_triage.go:534-541`, matching the snippet in section 1.1 below
2. **Pattern duplication**: No deduplication when adding patterns. FIXED: `addIgnorePatternIfNew` at `pkg/dot/doctor_triage.go:132`

### UX Issues
1. Uses `fmt.Scanln()` instead of existing `prompt` package. STILL OUTSTANDING: `fmt.Scanln` remains at `pkg/dot/doctor_triage.go` lines 170, 323, 403, 448, 530, 605, 696
2. No color in triage output
3. No confirmation summary before saving changes. FIXED: `confirmTriageChanges` at `pkg/dot/doctor_triage.go:153`, gated at `:117`
4. Already-ignored items shown again in triage. FIXED: `filterAlreadyIgnored` at `pkg/dot/doctor_triage.go:193`, called at `:69`
5. No way to review/manage existing ignored items. FIXED: `dot doctor ignore`, `dot doctor unignore`, `dot doctor ignores`
6. Adoption marked but never executed (orphaned feature)

### Missing Features
1. Dry-run: implemented at the service layer (`TriageOptions.DryRun`, honoured at
   pkg/dot/doctor_triage.go:112) but not reachable from the CLI. cmd/dot/doctor.go:357
   never sets it and no `--dry-run` flag is registered at cmd/dot/doctor.go:487.
2. Batch mode: same status. `TriageOptions.AutoConfirm` exists and is honoured at
   pkg/dot/doctor_triage.go:117, but no `--yes`/`--batch` flag is wired in.
3. Doesn't respect `--quiet` or `--verbose`
4. No undo functionality
5. No progress indication for large operations
6. Can't export/import ignore lists

## Recommended Improvements

### Phase 1: Critical Fixes

#### 1.1 Fix Apply-to-All Logic
```go
// Current (broken):
if choice == "a" && strings.Contains(choice, "A") {
    return choice, true
}

// Fixed:
if choice == "A" {
    return strings.ToLower(choice), true
}
```

#### 1.2 Use Prompt Package
```go
// Instead of fmt.Scanln, use:
prompter := prompt.New(os.Stdin, os.Stdout)
choice, err := prompter.Input("Choice")
```

#### 1.3 Filter Already-Ignored
```go
func (s *DoctorService) filterAlreadyIgnored(issues []Issue, m *manifest.Manifest) []Issue {
    if m.Doctor == nil {
        return issues
    }
    
    filtered := []Issue{}
    for _, issue := range issues {
        if _, ignored := m.Doctor.IgnoredLinks[issue.Path]; !ignored {
            // Check patterns too
            if !s.matchesIgnorePattern(issue.Path, m.Doctor.IgnoredPatterns) {
                filtered = append(filtered, issue)
            }
        }
    }
    return filtered
}
```

#### 1.4 Add Confirmation Summary
```go
func (s *DoctorService) confirmTriageChanges(result TriageResult) (bool, error) {
    fmt.Printf("\nSummary of changes:\n")
    fmt.Printf("  • %d links to ignore\n", len(result.Ignored))
    fmt.Printf("  • %d patterns to add\n", len(result.Patterns))
    fmt.Printf("  • %d links to adopt\n", len(result.Adopted))
    
    prompter := prompt.New(os.Stdin, os.Stdout)
    return prompter.ConfirmWithDefault("Save these changes?", true)
}
```

### Phase 2: Enhanced Features

#### 2.1 Triage Options Enhancement

Partially shipped. The current struct (`pkg/dot/doctor_triage.go:15-20`) is:

```go
type TriageOptions struct {
    AutoIgnoreHighConfidence bool // Automatically ignore high confidence categories
    DryRun                   bool // Show what would change without modifying
    AutoConfirm              bool // Skip confirmation prompts (--yes flag)
}
```

`ShowIgnored` and `RemoveIgnored` remain unimplemented, and only
`AutoIgnoreHighConfidence` is reachable from the CLI (`cmd/dot/doctor.go:357`).

#### 2.2 Manage Existing Ignores
```go
func (s *DoctorService) ManageIgnored(ctx context.Context) error {
    // Show current ignored links and patterns
    // Allow removing them
    // Allow editing reasons
}
```

#### 2.3 Smart Pattern Suggestions
```go
func (s *DoctorService) suggestPatterns(issues []Issue, cat *doctor.PatternCategory) []string {
    // Analyze common prefixes
    // Suggest most specific pattern that covers all
    // Show coverage count for each pattern
}
```

#### 2.4 Adoption Workflow
```go
// Either:
// A) Actually execute adoptions during triage
// B) Generate a script to run later
// C) Add to a "pending adoptions" queue

func (s *DoctorService) executeAdoptions(ctx context.Context, adoptions map[string]string) error {
    for linkPath, pkgName := range adoptions {
        // Call AdoptService to actually adopt
    }
}
```

### Phase 3: Advanced Features

#### 3.1 Interactive TUI
Consider using `bubbletea` or `tview` for:
- Multi-select with arrow keys
- Real-time filtering
- Better visualization
- Keyboard shortcuts

#### 3.2 Export/Import Ignore Lists
```go
func (s *DoctorService) ExportIgnoreList(format string) ([]byte, error)
func (s *DoctorService) ImportIgnoreList(data []byte) error
```

#### 3.3 Undo/Redo
```go
type TriageHistory struct {
    Actions []TriageAction
}

func (s *DoctorService) UndoLastTriage(ctx context.Context) error
```

#### 3.4 Diff Mode
Show what changed in manifest:
```bash
dot doctor --triage --show-diff
```

## Implementation Priority

1. **DONE**: Fix apply-to-all bug (`pkg/dot/doctor_triage.go:534`)
2. **DONE**: Filter already-ignored items (`pkg/dot/doctor_triage.go:193`)
3. **DONE**: Add confirmation summary (`pkg/dot/doctor_triage.go:153`)
4. **MEDIUM**: Use prompt package consistently
5. **MEDIUM**: Wire the existing `TriageOptions.DryRun` and `AutoConfirm` to CLI flags
6. **DONE**: Deduplicate patterns (`pkg/dot/doctor_triage.go:132`)
7. **LOW**: Color output
8. **DONE**: Manage existing ignores command (shipped as `dot doctor ignore` / `unignore` / `ignores`)
9. **FUTURE**: TUI interface
10. **FUTURE**: Export/Import

## Testing Requirements

For each improvement:
1. Unit tests for logic
2. Integration tests for persistence
3. Golden tests for output
4. Manual testing for UX

## API Stability

Changes should be backward compatible:
- Existing TriageOptions can add fields with defaults
- Manifest format already supports doctor state
- CLI flags are additive

## Related Issues

- Triage doesn't handle circular symlinks
- No support for triaging by age (acknowledge old, flag new)
- Can't triage based on target existence (broken vs valid orphans)

