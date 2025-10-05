# Phase 14: CLI Layer - Query Commands

## Overview

Implement read-only query commands (status, doctor, list) with multiple output formats. These commands provide observability into the current installation state, health diagnostics, and package inventory.

**Target**: Complete query command suite with rich output formatting

**Estimated Effort**: 16-24 hours

**Dependencies**: Phase 13 (Core Commands), Phase 11 (Manifest)

## Objectives

- Implement status command for installation state inspection
- Implement doctor command for health checks and diagnostics
- Implement list command for package inventory
- Create flexible output rendering system supporting text, JSON, YAML, and tables
- Provide actionable diagnostics with suggestions
- Enable machine-readable output for automation
- Maintain consistency with core command patterns

## Prerequisites

**Required**:
- Phase 13 complete (core commands working)
- Phase 11 complete (manifest system operational)
- Client API implements Status() and List() methods
- Domain types for Status, DiagnosticReport, PackageInfo defined

**Optional**:
- lipgloss for table rendering (alternative: tablewriter)
- yaml.v3 for YAML output
- color library for terminal colorization

## Architecture

### Command Structure

```
cmd/dot/
├── status.go      # Status command implementation
├── doctor.go      # Doctor command implementation
├── list.go        # List command implementation
└── render.go      # Output rendering helpers
```

### Renderer System

```
internal/cli/renderer/
├── renderer.go    # Renderer interface and factory
├── text.go        # Human-readable text output
├── json.go        # JSON output
├── yaml.go        # YAML output
├── table.go       # Tabular output with lipgloss
└── color.go       # Color scheme definitions
```

### Data Flow

```
Command → Client API → Domain Logic → Result
         ↓
    Select Renderer → Format Output → Write to stdout
```

## Detailed Implementation Plan

### 14.1: Output Renderer Infrastructure

**Goal**: Create flexible rendering system for multiple output formats

**Steps**:

1. **Define Renderer Interface**
   ```go
   type Renderer interface {
       RenderStatus(w io.Writer, status Status) error
       RenderDiagnostics(w io.Writer, report DiagnosticReport) error
       RenderPackageList(w io.Writer, packages []PackageInfo) error
   }
   ```

2. **Implement Renderer Factory**
   ```go
   func NewRenderer(format string, colorize bool) (Renderer, error)
   ```
   - Support formats: text, json, yaml, table
   - Validate format string
   - Configure colorization based on TTY detection and flag
   - Return appropriate renderer implementation

3. **Create Color Scheme**
   ```go
   type ColorScheme struct {
       Success   lipgloss.Color
       Warning   lipgloss.Color
       Error     lipgloss.Color
       Info      lipgloss.Color
       Muted     lipgloss.Color
   }
   ```
   - Define semantic colors
   - Support NO_COLOR environment variable
   - Provide both light and dark themes
   - Auto-detect terminal capabilities

4. **Implement Base Renderer Utilities**
   - Helper for indentation
   - Helper for line wrapping
   - Helper for truncation
   - Helper for pluralization
   - Writer abstraction for testing

**Tests**:
- Renderer factory creates correct type
- Color scheme respects NO_COLOR
- TTY detection works correctly
- Utilities handle edge cases (empty strings, long lines)

**Acceptance Criteria**:
- [ ] Renderer interface defined
- [ ] Factory creates all renderer types
- [ ] Color scheme configurable
- [ ] Base utilities tested
- [ ] Tests pass with 80%+ coverage

---

### 14.2: Text Renderer

**Goal**: Human-readable plain text output with optional colorization

**Steps**:

1. **Implement Text Renderer Structure**
   ```go
   type TextRenderer struct {
       colorize bool
       scheme   ColorScheme
       width    int
   }
   ```
   - Store colorization preference
   - Store color scheme
   - Store terminal width for wrapping

2. **Implement Status Rendering**
   ```go
   func (r *TextRenderer) RenderStatus(w io.Writer, status Status) error
   ```
   - Show package name and state (installed, not installed, conflict)
   - Display link count and last updated timestamp
   - List sample links (first 5, with "and N more...")
   - Show conflicts with details
   - Colorize based on state (green=ok, yellow=warning, red=error)
   - Add visual separators between packages

3. **Implement Diagnostics Rendering**
   ```go
   func (r *TextRenderer) RenderDiagnostics(w io.Writer, report DiagnosticReport) error
   ```
   - Show overall health status (healthy, warnings, errors)
   - List issues by severity
   - Display broken symlinks with paths
   - Show orphaned links
   - List permission issues
   - Provide actionable suggestions per issue
   - Add summary statistics
   - Use icons/symbols for visual appeal (✓, ⚠, ✗)

4. **Implement Package List Rendering**
   ```go
   func (r *TextRenderer) RenderPackageList(w io.Writer, packages []PackageInfo) error
   ```
   - Show package name, link count, size, installation date
   - Format dates relative (2 days ago) or absolute
   - Align columns for readability
   - Add header row
   - Support sorting indication
   - Show totals at bottom

5. **Add Text Formatting Helpers**
   - formatTimestamp() for human-readable dates
   - formatBytes() for human-readable sizes
   - formatDuration() for relative times
   - formatPath() for truncated paths with ellipsis
   - formatList() for comma-separated items with "and N more"

**Tests**:
- Status renders correctly for all states
- Diagnostics shows all issue types
- Package list handles empty list
- Colorization can be disabled
- Long paths are truncated
- Large numbers are formatted
- Unicode symbols handled gracefully

**Acceptance Criteria**:
- [ ] Status output is clear and readable
- [ ] Diagnostics provide actionable information
- [ ] Package list is well-formatted
- [ ] Colors enhance readability
- [ ] Text wraps appropriately
- [ ] Tests cover all rendering paths

---

### 14.3: JSON Renderer

**Goal**: Machine-readable JSON output for automation

**Steps**:

1. **Implement JSON Renderer Structure**
   ```go
   type JSONRenderer struct {
       pretty bool
   }
   ```
   - Support both compact and pretty-printed output
   - Use json.MarshalIndent for pretty mode

2. **Implement Status Rendering**
   ```go
   func (r *JSONRenderer) RenderStatus(w io.Writer, status Status) error
   ```
   - Marshal Status to JSON
   - Use camelCase for field names (via struct tags)
   - Include all fields without truncation
   - Maintain type information (timestamps as RFC3339)
   - Handle errors gracefully

3. **Implement Diagnostics Rendering**
   ```go
   func (r *JSONRenderer) RenderDiagnostics(w io.Writer, report DiagnosticReport) error
   ```
   - Marshal DiagnosticReport to JSON
   - Preserve issue categorization
   - Include all diagnostic details
   - Maintain suggestion ordering

4. **Implement Package List Rendering**
   ```go
   func (r *JSONRenderer) RenderPackageList(w io.Writer, packages []PackageInfo) error
   ```
   - Marshal package slice to JSON array
   - Include all package metadata
   - Preserve sorting order

5. **Add JSON Utilities**
   - Ensure all domain types have json tags
   - Add custom MarshalJSON for special types if needed
   - Handle null vs empty slice correctly
   - Validate JSON output is well-formed

**Tests**:
- JSON is valid and parseable
- All fields are present
- Timestamps in correct format
- Pretty and compact modes both work
- Special characters are escaped
- Empty collections handled correctly

**Acceptance Criteria**:
- [ ] Valid JSON output for all commands
- [ ] All fields serialized correctly
- [ ] Timestamps in ISO 8601 format
- [ ] Pretty printing works
- [ ] Parseable by jq and other tools
- [ ] Tests validate JSON structure

---

### 14.4: YAML Renderer

**Goal**: YAML output for configuration-like readability

**Steps**:

1. **Implement YAML Renderer Structure**
   ```go
   type YAMLRenderer struct {
       indent int
   }
   ```
   - Use gopkg.in/yaml.v3 for marshaling
   - Configure indentation (default 2 spaces)

2. **Implement Status Rendering**
   ```go
   func (r *YAMLRenderer) RenderStatus(w io.Writer, status Status) error
   ```
   - Marshal Status to YAML
   - Use snake_case for field names (via struct tags)
   - Maintain hierarchical structure
   - Format timestamps appropriately

3. **Implement Diagnostics Rendering**
   ```go
   func (r *YAMLRenderer) RenderDiagnostics(w io.Writer, report DiagnosticReport) error
   ```
   - Marshal DiagnosticReport to YAML
   - Group issues by type
   - Preserve nested structures

4. **Implement Package List Rendering**
   ```go
   func (r *YAMLRenderer) RenderPackageList(w io.Writer, packages []PackageInfo) error
   ```
   - Marshal package slice to YAML array
   - Use flow style for compact lists where appropriate
   - Maintain readability

5. **Add YAML Tags**
   - Add yaml struct tags to all relevant types
   - Ensure consistent naming convention
   - Handle omitempty for optional fields

**Tests**:
- YAML is valid and parseable
- All fields present in output
- Indentation consistent
- Can be round-tripped (parse and re-serialize)
- Comments preserved if added

**Acceptance Criteria**:
- [ ] Valid YAML output for all commands
- [ ] Human-readable structure
- [ ] Consistent formatting
- [ ] Parseable by yaml tools
- [ ] Tests validate YAML structure

---

### 14.5: Table Renderer

**Goal**: Rich tabular output using lipgloss

**Steps**:

1. **Implement Table Renderer Structure**
   ```go
   type TableRenderer struct {
       colorize bool
       scheme   ColorScheme
       width    int
   }
   ```
   - Use lipgloss for table styling
   - Support borders, headers, alignment
   - Handle terminal width constraints

2. **Create Table Builder**
   ```go
   func (r *TableRenderer) buildTable(headers []string, rows [][]string) string
   ```
   - Define table style (borders, padding, alignment)
   - Calculate column widths dynamically
   - Handle truncation for narrow terminals
   - Add header styling
   - Apply row coloring based on content

3. **Implement Status Rendering**
   ```go
   func (r *TableRenderer) RenderStatus(w io.Writer, status Status) error
   ```
   - Create table with columns: Package, State, Links, Last Updated
   - Color-code state column
   - Right-align numeric columns
   - Add totals row if multiple packages

4. **Implement Diagnostics Rendering**
   ```go
   func (r *TableRenderer) RenderDiagnostics(w io.Writer, report DiagnosticReport) error
   ```
   - Create issues table: Severity, Type, Path, Suggestion
   - Color-code severity column
   - Truncate long paths intelligently
   - Group by severity
   - Add summary section

5. **Implement Package List Rendering**
   ```go
   func (r *TableRenderer) RenderPackageList(w io.Writer, packages []PackageInfo) error
   ```
   - Create table with columns: Name, Links, Size, Installed
   - Sort by specified field
   - Right-align numeric columns
   - Format dates and sizes
   - Add totals row

6. **Add Table Utilities**
   - Column width calculation
   - Cell truncation with ellipsis
   - Multi-line cell handling
   - Border style definitions
   - Header/footer generation

**Tests**:
- Tables render correctly
- Columns align properly
- Long content truncates
- Colors apply correctly
- Empty tables handled
- Terminal width respected

**Acceptance Criteria**:
- [ ] Tables are visually appealing
- [ ] Columns align correctly
- [ ] Content fits terminal width
- [ ] Colors enhance readability
- [ ] Tests cover edge cases

---

### 14.6: Status Command

**Goal**: Show installation status for packages

**Steps**:

1. **Create Status Command Structure**
   ```go
   func NewStatusCommand(cfg *dot.Config) *cobra.Command
   ```
   - Accept optional package names as arguments
   - Add --format flag (text, json, yaml, table)
   - Add --color flag (auto, always, never)
   - Set default format to text

2. **Implement Command Handler**
   ```go
   RunE: func(cmd *cobra.Command, args []string) error
   ```
   - Create Client from config
   - Call client.Status(ctx, args...)
   - Handle empty args (show all packages)
   - Select renderer based on format flag
   - Render status to stdout
   - Return appropriate error

3. **Add Status Query Logic** (in pkg/dot or internal/api)
   ```go
   func (c *client) Status(ctx context.Context, packages ...string) (Status, error)
   ```
   - Load manifest from target directory
   - For each specified package (or all if none specified):
     - Check if package exists in stow dir
     - Check if package is in manifest (installed)
     - Verify links still exist and point correctly
     - Detect broken or wrong links
     - Identify conflicts
   - Return Status structure with findings

4. **Define Status Types**
   ```go
   type Status struct {
       Packages []PackageStatus
       Summary  StatusSummary
   }
   
   type PackageStatus struct {
       Name         string
       State        PackageState
       LinkCount    int
       LastUpdated  time.Time
       Links        []LinkInfo
       Conflicts    []Conflict
   }
   
   type PackageState int
   const (
       StateInstalled PackageState = iota
       StateNotInstalled
       StatePartiallyInstalled
       StateConflict
       StateNotFound
   )
   ```

5. **Add Help Text**
   - Command description
   - Usage examples:
     - `dot status` - show all packages
     - `dot status vim tmux` - show specific packages
     - `dot status --format=json` - machine-readable output
   - Flag descriptions

**Tests**:
- Status with no args shows all packages
- Status with specific packages filters correctly
- Status detects installed packages
- Status detects not installed packages
- Status detects partially installed (some links broken)
- Status detects conflicts
- Output format selection works
- Empty package list handled
- Package not found handled
- Manifest missing handled gracefully

**Acceptance Criteria**:
- [ ] Command executes successfully
- [ ] All package states detected correctly
- [ ] Links verified accurately
- [ ] Conflicts identified
- [ ] All output formats work
- [ ] Help text is clear
- [ ] Tests pass with 80%+ coverage

---

### 14.7: Doctor Command

**Goal**: Comprehensive health check and diagnostics

**Steps**:

1. **Create Doctor Command Structure**
   ```go
   func NewDoctorCommand(cfg *dot.Config) *cobra.Command
   ```
   - No positional arguments (always checks entire installation)
   - Add --format flag (text, json, yaml)
   - Add --fix flag for auto-repair (future enhancement)
   - Set default format to text

2. **Implement Command Handler**
   ```go
   RunE: func(cmd *cobra.Command, args []string) error
   ```
   - Create Client from config
   - Call client.Doctor(ctx)
   - Select renderer based on format flag
   - Render diagnostic report to stdout
   - Return exit code based on health status:
     - 0: healthy
     - 1: warnings
     - 2: errors

3. **Add Doctor Logic** (in pkg/dot or internal/api)
   ```go
   func (c *client) Doctor(ctx context.Context) (DiagnosticReport, error)
   ```
   - Check manifest exists and is valid
   - Scan target directory for all symlinks
   - Categorize each symlink:
     - Managed by dot and correct
     - Managed by dot but broken
     - Managed by dot but wrong target
     - Not managed by dot (orphaned)
   - Check for permission issues
   - Validate stow directory exists and is readable
   - Check for circular symlink dependencies
   - Verify manifest consistency with filesystem
   - Generate suggestions for each issue

4. **Define Diagnostic Types**
   ```go
   type DiagnosticReport struct {
       OverallHealth HealthStatus
       Issues        []Issue
       Statistics    DiagnosticStats
       Suggestions   []string
   }
   
   type Issue struct {
       Severity   Severity
       Type       IssueType
       Path       string
       Message    string
       Suggestion string
   }
   
   type Severity int
   const (
       SeverityInfo Severity = iota
       SeverityWarning
       SeverityError
   )
   
   type IssueType int
   const (
       IssueBrokenLink IssueType = iota
       IssueOrphanedLink
       IssueWrongTarget
       IssuePermission
       IssueCircular
       IssueManifestInconsistency
   )
   ```

5. **Implement Issue Detection**
   - detectBrokenLinks(): Find symlinks pointing to non-existent targets
   - detectOrphanedLinks(): Find links not in manifest
   - detectWrongTargets(): Find links in manifest pointing to wrong locations
   - detectPermissionIssues(): Check read/write access
   - detectCircularLinks(): Find recursive symlink chains
   - validateManifest(): Check manifest consistency

6. **Generate Suggestions**
   - For broken links: "Remove broken link or reinstall package"
   - For orphaned links: "Add package to dot management or remove link"
   - For wrong targets: "Run 'dot remanage' to fix"
   - For permissions: "Check filesystem permissions"
   - For circular: "Break circular dependency by removing one link"
   - For manifest: "Run 'dot remanage' to rebuild manifest"

7. **Add Help Text**
   - Command description
   - Usage examples:
     - `dot doctor` - full health check
     - `dot doctor --format=json` - machine-readable report
   - Flag descriptions
   - Exit code documentation

**Tests**:
- Doctor detects broken links
- Doctor detects orphaned links
- Doctor detects wrong targets
- Doctor detects permission issues
- Doctor detects circular dependencies
- Doctor validates manifest correctly
- Suggestions are actionable
- Exit codes are correct
- Empty target directory handled
- Missing manifest handled
- All output formats work

**Acceptance Criteria**:
- [ ] All issue types detected
- [ ] Suggestions are helpful
- [ ] Health status accurate
- [ ] Exit codes correct
- [ ] All output formats work
- [ ] Help text is clear
- [ ] Tests pass with 80%+ coverage

---

### 14.8: List Command

**Goal**: Display installed package inventory

**Steps**:

1. **Create List Command Structure**
   ```go
   func NewListCommand(cfg *dot.Config) *cobra.Command
   ```
   - No positional arguments (always lists all packages)
   - Add --format flag (text, json, yaml, table)
   - Add --sort flag (name, links, size, date)
   - Add --filter flag (future enhancement)
   - Set default format to table
   - Set default sort to name

2. **Implement Command Handler**
   ```go
   RunE: func(cmd *cobra.Command, args []string) error
   ```
   - Create Client from config
   - Call client.List(ctx)
   - Sort packages according to --sort flag
   - Select renderer based on format flag
   - Render package list to stdout
   - Return appropriate error

3. **Add List Logic** (in pkg/dot or internal/api)
   ```go
   func (c *client) List(ctx context.Context) ([]PackageInfo, error)
   ```
   - Load manifest from target directory
   - Extract package list with metadata
   - For each package:
     - Get link count from manifest
     - Get installation timestamp
     - Calculate total size (optional, may be expensive)
   - Return package slice

4. **Define PackageInfo Type** (if not already in pkg/dot)
   ```go
   type PackageInfo struct {
       Name        string
       LinkCount   int
       Size        int64
       InstalledAt time.Time
       Links       []string
   }
   ```

5. **Implement Sorting**
   ```go
   func sortPackages(packages []PackageInfo, sortBy string) []PackageInfo
   ```
   - Support sort by: name, links, size, date
   - Use sort.Slice with appropriate comparator
   - Handle case-insensitive name sorting
   - Handle reverse sort (future enhancement)

6. **Add Help Text**
   - Command description
   - Usage examples:
     - `dot list` - show all packages
     - `dot list --sort=links` - sort by link count
     - `dot list --format=json` - machine-readable list
   - Flag descriptions
   - Sort field documentation

**Tests**:
- List shows all installed packages
- List handles empty manifest
- List handles missing manifest
- Sorting works for all fields
- All output formats work
- Link count is accurate
- Timestamps are correct
- Empty package list handled

**Acceptance Criteria**:
- [ ] All packages listed
- [ ] Sorting works correctly
- [ ] All output formats work
- [ ] Metadata is accurate
- [ ] Help text is clear
- [ ] Tests pass with 80%+ coverage

---

### 14.9: Integration and Polish

**Goal**: Ensure consistency and quality across all query commands

**Steps**:

1. **Command Consistency**
   - Verify all commands follow same flag patterns
   - Ensure --format flag works identically
   - Standardize help text format
   - Consistent error messages
   - Same exit code conventions

2. **Error Handling**
   - Handle missing manifest gracefully (suggest running manage first)
   - Handle missing stow directory
   - Handle missing target directory
   - Handle filesystem permission errors
   - Provide clear error messages

3. **Output Consistency**
   - Ensure all renderers produce equivalent information
   - Verify JSON and YAML are parseable
   - Check table rendering in narrow terminals
   - Test colorization on/off
   - Validate NO_COLOR environment variable support

4. **Performance**
   - Profile status command with many packages
   - Optimize manifest loading (cache if needed)
   - Minimize filesystem operations
   - Benchmark rendering performance

5. **Documentation**
   - Add examples to README
   - Document output formats
   - Add troubleshooting section
   - Document exit codes
   - Add screenshots/examples

**Tests**:
- Integration test: manage → status → doctor → list workflow
- Test all commands with empty manifest
- Test all commands with large package sets
- Test all output formats for all commands
- Test error conditions
- Test edge cases (empty dirs, broken links, etc.)

**Acceptance Criteria**:
- [ ] All commands follow consistent patterns
- [ ] Error handling is robust
- [ ] Output is consistent across formats
- [ ] Performance is acceptable
- [ ] Documentation is complete

---

## Testing Strategy

### Unit Tests

**Renderer Tests**:
- Test each renderer independently
- Mock domain types
- Verify output format
- Test colorization on/off
- Test truncation and wrapping

**Command Tests**:
- Test flag parsing
- Test argument validation
- Mock Client interface
- Verify renderer selection
- Test exit codes

### Integration Tests

**End-to-End Tests**:
```go
func TestQueryCommandsE2E(t *testing.T) {
    // Setup: create packages, run manage
    // Test: run status, verify output
    // Test: run doctor, verify health
    // Test: run list, verify inventory
    // Cleanup
}
```

**Format Tests**:
- Test each command with each format
- Verify output is parseable
- Validate JSON schema
- Validate YAML structure
- Check text readability

**State Tests**:
- Test status with various installation states
- Test doctor with various health issues
- Test list with various package counts
- Test commands with manifest missing
- Test commands with manifest corrupted

### Property Tests

- Status always reports consistent state
- Doctor finds all issues (no false negatives)
- List returns all installed packages
- Output format preserves information (no data loss)

### Performance Tests

**Benchmarks**:
```go
func BenchmarkStatusCommand(b *testing.B)
func BenchmarkDoctorCommand(b *testing.B)
func BenchmarkListCommand(b *testing.B)
func BenchmarkTextRenderer(b *testing.B)
func BenchmarkTableRenderer(b *testing.B)
```

**Load Tests**:
- Status with 1000 packages
- Doctor with 10000 links
- List with large package sets
- Renderer with long paths

---

## Success Criteria

### Functional Requirements

- [ ] Status command shows installation state accurately
- [ ] Doctor command detects all issue types
- [ ] List command displays all installed packages
- [ ] All output formats work for all commands
- [ ] Color output enhances readability
- [ ] Machine-readable formats are valid and parseable

### Non-Functional Requirements

- [ ] Commands execute in < 100ms for typical installations
- [ ] Commands scale to 1000+ packages
- [ ] Output fits standard terminal widths (80, 120 columns)
- [ ] Memory usage is bounded (< 100MB for large installations)
- [ ] Exit codes follow conventions

### Quality Requirements

- [ ] Test coverage ≥ 80% for all new code
- [ ] All linters pass without warnings
- [ ] No golangci-lint errors
- [ ] Documentation complete
- [ ] Examples in help text
- [ ] Integration tests pass

### User Experience Requirements

- [ ] Help text is clear and includes examples
- [ ] Error messages are actionable
- [ ] Output is readable and well-formatted
- [ ] Tables render correctly in terminals
- [ ] Colors used meaningfully (not decorative)
- [ ] NO_COLOR environment variable respected

---

## Timeline

### Week 1: Renderer Infrastructure (6-8 hours)
- Day 1-2: Implement renderer interface and factory (14.1)
- Day 2-3: Implement text renderer (14.2)
- Day 3-4: Implement JSON and YAML renderers (14.3, 14.4)

### Week 2: Commands (8-10 hours)
- Day 1-2: Implement table renderer (14.5)
- Day 3: Implement status command (14.6)
- Day 4: Implement doctor command (14.7)
- Day 5: Implement list command (14.8)

### Week 3: Testing and Polish (4-6 hours)
- Day 1-2: Integration testing (14.9)
- Day 2: Performance testing and optimization
- Day 3: Documentation and examples

**Total Estimated Effort**: 18-24 hours

---

## Dependencies and Integration

### Internal Dependencies

**Required**:
- `pkg/dot.Client` interface with Status(), List() methods
- `pkg/dot.Status`, `pkg/dot.PackageInfo` types defined
- `internal/manifest` package for state loading
- Phase 13 core commands for consistency patterns

**New**:
- `pkg/dot.DiagnosticReport`, `pkg/dot.Issue` types
- `internal/api.Doctor()` implementation

### External Dependencies

**Required**:
- `github.com/spf13/cobra` for command structure
- `encoding/json` for JSON rendering
- `gopkg.in/yaml.v3` for YAML rendering

**Optional**:
- `github.com/charmbracelet/lipgloss` for tables
- `github.com/fatih/color` for terminal colors
- `golang.org/x/term` for TTY detection

---

## Implementation Notes

### Renderer Design Pattern

Use strategy pattern for output formatting:
- Define common Renderer interface
- Implement separate renderer for each format
- Factory creates appropriate renderer based on flag
- Commands depend on Renderer interface, not concrete types

### Terminal Width Detection

```go
func getTerminalWidth() int {
    width, _, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil || width == 0 {
        return 80 // default fallback
    }
    return width
}
```

### Color Scheme Configuration

Support three modes:
- `auto`: colorize if stdout is TTY and NO_COLOR not set
- `always`: always colorize (for pagers that support color)
- `never`: never colorize (for logging, piping)

### Truncation Strategy

For narrow terminals:
- Truncate paths from middle: `/very/long/.../path`
- Truncate descriptions from end: `Description text...`
- Stack columns vertically if too narrow
- Minimum width: 40 columns

### JSON Schema Validation

Consider adding JSON schema output:
```bash
dot status --format=json-schema
```
Useful for tool integration and validation.

---

## Risk Mitigation

### Risks

1. **Renderer Complexity**: Multiple output formats increase maintenance burden
   - Mitigation: Share logic via helpers, comprehensive tests

2. **Performance**: Large package sets may be slow to query
   - Mitigation: Optimize manifest loading, cache where appropriate

3. **Terminal Compatibility**: Different terminals may render differently
   - Mitigation: Test on multiple terminals, fallback to safe defaults

4. **Breaking Changes**: New status format may break scripts
   - Mitigation: Version output format, maintain backward compatibility

### Testing Challenges

- Difficult to test terminal output visually
- Color rendering depends on terminal
- Table rendering depends on width

**Solutions**:
- Snapshot testing for text output
- JSON validation for structured output
- Mock terminal for width testing

---

## Future Enhancements

Post-Phase 14 improvements:

### Interactive Selection
- Use bubbletea for interactive package selection
- Visual status browsing
- Interactive issue resolution

### Watch Mode
```bash
dot status --watch
```
- Continuously monitor installation state
- Alert on changes or issues
- Real-time health monitoring

### Filtering and Querying
```bash
dot list --filter='links>10'
dot status --since=7d
dot doctor --severity=error
```
- Complex filtering expressions
- Date range queries
- Severity filtering

### Export Formats
- HTML report generation
- Markdown summary
- CSV export for spreadsheets

### Diff Command
```bash
dot diff vim
```
- Show difference between installed and package
- Show changes since last install
- Highlight configuration drift

---

## Deliverables

### Code
- [ ] cmd/dot/status.go
- [ ] cmd/dot/doctor.go
- [ ] cmd/dot/list.go
- [ ] internal/cli/renderer/renderer.go
- [ ] internal/cli/renderer/text.go
- [ ] internal/cli/renderer/json.go
- [ ] internal/cli/renderer/yaml.go
- [ ] internal/cli/renderer/table.go

### Tests
- [ ] Renderer unit tests (internal/cli/renderer/*_test.go)
- [ ] Command unit tests (cmd/dot/*_test.go)
- [ ] Integration tests (tests/integration/query_test.go)
- [ ] Benchmark tests

### Documentation
- [ ] Updated README with query command examples
- [ ] Command help text
- [ ] Output format documentation
- [ ] Troubleshooting guide additions

---

## Validation Checklist

Before marking Phase 14 complete:

- [ ] All commands execute successfully
- [ ] All output formats produce valid output
- [ ] Status accurately reflects installation state
- [ ] Doctor finds all issue types
- [ ] List shows all packages
- [ ] Colors enhance readability
- [ ] Tables render correctly
- [ ] JSON/YAML are parseable
- [ ] Exit codes are correct
- [ ] Error messages are clear
- [ ] Help text is complete
- [ ] Tests pass with ≥80% coverage
- [ ] All linters pass
- [ ] Integration tests pass
- [ ] Performance is acceptable
- [ ] Documentation is complete
- [ ] Examples work as documented

---

## References

- Architecture: [Architecture.md](./Architecture.md) - Client API, domain types
- Features: [Features.md](./Features.md) - Status, Doctor, List user stories
- Phase 13: [Phase-13-Plan.md](./Phase-13-Plan.md) - Command patterns
- Implementation Plan: [Implementation-Plan.md](./Implementation-Plan.md) - Overall context

