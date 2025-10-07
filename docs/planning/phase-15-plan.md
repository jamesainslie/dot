# Phase 15: Error Handling and User Experience - Detailed Implementation Plan

## Overview

Phase 15 focuses on creating a polished, user-friendly experience through enhanced error messages, comprehensive help system, and progress indicators. This phase transforms technical errors into actionable guidance and provides users with clear feedback during operations.

## Prerequisites

- Phase 13 complete: CLI core commands functional
- Phase 14 complete: Query commands implemented
- All domain errors defined and used throughout codebase

## Design Principles

- **User-Centric Errors**: Messages explain problems in plain language, not technical jargon
- **Actionable Guidance**: Every error includes suggestions for resolution
- **Progressive Disclosure**: Show appropriate detail based on verbosity level
- **Accessibility**: Support color-blind users, screen readers, and terminal constraints
- **Consistency**: Uniform formatting across all commands and error types

## Architecture

### Error Rendering Pipeline

```
Domain Error → Error Formatter → Styled Renderer → Terminal Output
     ↓              ↓                  ↓                 ↓
  Technical    User Message       Color/Style      Final Display
  Context      + Suggestions      Application
```

### Components

```
internal/cli/
├── errors/
│   ├── formatter.go          # Error message formatting
│   ├── context.go            # Error context extraction
│   ├── suggestions.go        # Resolution suggestion engine
│   └── templates.go          # Error message templates
├── help/
│   ├── generator.go          # Help text generation
│   ├── examples.go           # Command examples
│   ├── templates.go          # Help templates
│   └── completion.go         # Shell completion generation
├── progress/
│   ├── indicator.go          # Progress indicator interface
│   ├── bar.go                # Progress bar implementation
│   ├── spinner.go            # Spinner implementation
│   └── tracker.go            # Operation progress tracking
└── render/
    ├── style.go              # Terminal styling utilities
    ├── color.go              # Color scheme definitions
    └── layout.go             # Output layout helpers
```

## Implementation Tasks

### 15.1: Error Formatting Foundation

**Goal**: Create infrastructure for user-friendly error messages

#### 15.1.1: Error Formatter Core

**File**: `internal/cli/errors/formatter.go`

```go
// Formatter converts domain errors to user-friendly messages
type Formatter struct {
    colorEnabled bool
    verbosity    int
    width        int
}

// Format converts an error to a formatted message
func (f *Formatter) Format(err error) string

// FormatWithContext adds contextual information
func (f *Formatter) FormatWithContext(err error, ctx ErrorContext) string
```

**Implementation**:
- [ ] Define Formatter struct with configuration options
- [ ] Implement Format() for basic error conversion
- [ ] Add FormatWithContext() for rich error messages
- [ ] Support multi-error aggregation formatting
- [ ] Add terminal width detection for text wrapping
- [ ] Handle color enabling/disabling based on terminal capabilities
- [ ] Write comprehensive formatter tests

**Test Cases**:
- Format simple domain errors
- Format errors with context
- Format multiple errors (ErrMultiple)
- Format with and without color
- Format for different terminal widths
- Format at different verbosity levels

#### 15.1.2: Error Context Extraction

**File**: `internal/cli/errors/context.go`

```go
// ErrorContext provides additional information for error rendering
type ErrorContext struct {
    Command   string
    Arguments []string
    Config    ConfigSummary
    Timestamp time.Time
}

// Extract pulls context from various sources
func Extract(cmd *cobra.Command, cfg *Config) ErrorContext
```

**Implementation**:
- [ ] Define ErrorContext struct
- [ ] Implement Extract() to gather context from command state
- [ ] Add ConfigSummary type for relevant config info
- [ ] Extract package names, paths from arguments
- [ ] Add timestamp and command chain tracking
- [ ] Write context extraction tests

**Test Cases**:
- Extract from manage command
- Extract from status command
- Extract with various flag combinations
- Extract with missing context gracefully

#### 15.1.3: Error Message Templates

**File**: `internal/cli/errors/templates.go`

```go
// Template for structured error messages
type Template struct {
    Title       string
    Description string
    Details     []string
    Suggestions []string
    Footer      string
}

// Render applies template to produce final message
func (t *Template) Render(style *Style) string
```

**Implementation**:
- [ ] Define Template struct for error structure
- [ ] Create templates for each domain error type
- [ ] Implement Render() with styling support
- [ ] Add template for ErrInvalidPath with path validation guidance
- [ ] Add template for ErrPackageNotFound with discovery help
- [ ] Add template for ErrConflict with resolution options
- [ ] Add template for ErrPermissionDenied with permission fix guidance
- [ ] Add template for ErrCyclicDependency with cycle visualization
- [ ] Add template for ErrMultiple with grouped error display
- [ ] Write template rendering tests

**Templates to Create**:
```
ErrInvalidPath:
  Title: "Invalid Path"
  Description: "The path '%s' is not valid"
  Details: [reason for invalidity]
  Suggestions:
    - "Use absolute paths starting with /"
    - "Check for typos in the path"
    - "Verify the directory exists"

ErrPackageNotFound:
  Title: "Package Not Found"
  Description: "Package '%s' does not exist in package directory"
  Suggestions:
    - "Check available packages with: dot list"
    - "Verify package directory: %s"
    - "Check for typos in package name"

ErrConflict (FileExists):
  Title: "File Exists"
  Description: "Cannot create symlink at '%s' - file already exists"
  Suggestions:
    - "Use --backup to preserve existing file"
    - "Use 'dot adopt' to move file into package"
    - "Remove the conflicting file manually"
    - "Use --skip to continue with other operations"

ErrPermissionDenied:
  Title: "Permission Denied"
  Description: "Cannot access '%s' - permission denied"
  Suggestions:
    - "Check file permissions with: ls -la %s"
    - "Verify you have write access to target directory"
    - "Run with appropriate permissions if needed"

ErrCyclicDependency:
  Title: "Circular Dependency Detected"
  Description: "Operations form a dependency cycle"
  Details: [visual cycle display]
  Suggestions:
    - "This is likely a bug - please report it"
    - "Try running operations separately"
```

**Test Cases**:
- Render each template type
- Render with different styling options
- Render with variable substitution
- Verify template completeness

#### 15.1.4: Suggestion Engine

**File**: `internal/cli/errors/suggestions.go`

```go
// SuggestionEngine generates actionable resolution steps
type SuggestionEngine struct {
    context ErrorContext
}

// Generate creates suggestions for an error
func (e *SuggestionEngine) Generate(err error) []string

// Prioritize orders suggestions by likely usefulness
func (e *SuggestionEngine) Prioritize(suggestions []string) []string
```

**Implementation**:
- [ ] Define SuggestionEngine struct
- [ ] Implement Generate() dispatcher for error types
- [ ] Add suggestion generation for path errors
- [ ] Add suggestion generation for conflict errors
- [ ] Add suggestion generation for permission errors
- [ ] Add suggestion generation for package errors
- [ ] Implement Prioritize() based on error context
- [ ] Add command suggestion based on similar names (fuzzy match)
- [ ] Write suggestion engine tests

**Suggestion Strategies**:
- Path errors: Check common path issues, suggest corrections
- Conflict errors: Suggest resolution flags, adopt command
- Permission errors: Suggest permission checks, ownership fixes
- Package errors: Fuzzy match package names, suggest list command
- Config errors: Show example configurations, point to documentation

**Test Cases**:
- Generate suggestions for each error type
- Prioritize based on different contexts
- Handle errors with no applicable suggestions
- Fuzzy match package names
- Suggest related commands

### 15.2: Terminal Styling and Rendering

**Goal**: Create consistent, accessible styled output

#### 15.2.1: Style System

**File**: `internal/cli/render/style.go`

```go
// Style defines terminal styling options
type Style struct {
    color       ColorScheme
    bold        bool
    underline   bool
    italic      bool
}

// StyleFunc applies styling to text
type StyleFunc func(string) string

// Predefined styles
var (
    ErrorStyle      StyleFunc
    WarningStyle    StyleFunc
    SuccessStyle    StyleFunc
    InfoStyle       StyleFunc
    EmphasisStyle   StyleFunc
    DimStyle        StyleFunc
    CodeStyle       StyleFunc
    PathStyle       StyleFunc
)
```

**Implementation**:
- [ ] Define Style struct with styling attributes
- [ ] Implement StyleFunc type for style application
- [ ] Create predefined styles for common use cases
- [ ] Add terminal capability detection (TERM, COLORTERM)
- [ ] Implement NO_COLOR environment variable support
- [ ] Add --color flag support (auto, always, never)
- [ ] Detect terminal width for layout
- [ ] Write styling tests

**Styles to Define**:
- ErrorStyle: Red, bold for error messages
- WarningStyle: Yellow for warnings
- SuccessStyle: Green for success messages
- InfoStyle: Blue for informational messages
- EmphasisStyle: Bold for emphasis
- DimStyle: Dim/gray for secondary text
- CodeStyle: Monospace styling for code/paths
- PathStyle: Distinct styling for file paths

**Test Cases**:
- Apply each predefined style
- Detect terminal capabilities correctly
- Respect NO_COLOR environment variable
- Handle --color flag values
- Work on terminals without color support
- Nested style application

#### 15.2.2: Color Schemes

**File**: `internal/cli/render/color.go`

```go
// ColorScheme defines a color palette
type ColorScheme struct {
    Error   Color
    Warning Color
    Success Color
    Info    Color
    Dim     Color
}

// Predefined schemes
var (
    DefaultScheme ColorScheme
    NoColorScheme ColorScheme
)

// Color represents a terminal color
type Color struct {
    ANSI   string
    RGB    [3]uint8  // For 24-bit color support
}
```

**Implementation**:
- [ ] Define ColorScheme struct
- [ ] Create DefaultScheme with accessible colors
- [ ] Create NoColorScheme for plain text
- [ ] Define Color type with ANSI and RGB values
- [ ] Implement color application functions
- [ ] Add support for 16-color, 256-color, and 24-bit color
- [ ] Test colors for accessibility (contrast ratios)
- [ ] Write color tests

**Color Palette**:
- Error: Red (#E06C75)
- Warning: Yellow (#E5C07B)
- Success: Green (#98C379)
- Info: Blue (#61AFEF)
- Dim: Gray (#5C6370)

**Test Cases**:
- Apply colors from default scheme
- Verify no-color scheme produces no ANSI codes
- Test color detection for terminal types
- Verify accessibility contrast ratios

#### 15.2.3: Layout Helpers

**File**: `internal/cli/render/layout.go`

```go
// Layout provides text layout utilities
type Layout struct {
    width int
}

// Wrap wraps text to terminal width
func (l *Layout) Wrap(text string, indent int) string

// Indent adds indentation to text
func (l *Layout) Indent(text string, level int) string

// Box draws a box around text
func (l *Layout) Box(text string, title string) string

// Table formats data as a table
func (l *Layout) Table(headers []string, rows [][]string) string
```

**Implementation**:
- [ ] Define Layout struct with width configuration
- [ ] Implement Wrap() for text wrapping at terminal width
- [ ] Implement Indent() for consistent indentation
- [ ] Implement Box() for boxed content
- [ ] Implement Table() for tabular data
- [ ] Add column alignment support (left, right, center)
- [ ] Handle multi-line cells in tables
- [ ] Handle ANSI codes in width calculations
- [ ] Write layout tests

**Test Cases**:
- Wrap text at various widths
- Wrap text with ANSI codes
- Indent single and multi-line text
- Draw boxes with various content
- Format tables with different column counts
- Handle empty tables gracefully

### 15.3: Progress Indicators

**Goal**: Provide visual feedback for long-running operations

#### 15.3.1: Progress Indicator Interface

**File**: `internal/cli/progress/indicator.go`

```go
// Indicator provides progress feedback
type Indicator interface {
    Start(message string)
    Update(current, total int, message string)
    Stop(message string)
    Fail(message string)
}

// Config for progress indicators
type Config struct {
    Enabled     bool
    Interactive bool  // Terminal supports cursor control
    Width       int
}

// New creates appropriate indicator for terminal
func New(cfg Config) Indicator
```

**Implementation**:
- [ ] Define Indicator interface
- [ ] Define Config struct with terminal capabilities
- [ ] Implement New() to select appropriate indicator type
- [ ] Add NoOpIndicator for non-interactive terminals
- [ ] Detect terminal interactivity (isatty)
- [ ] Respect --quiet flag to disable indicators
- [ ] Write indicator interface tests

**Test Cases**:
- Create indicators for interactive terminals
- Create indicators for non-interactive terminals
- Respect quiet mode
- Handle missing terminal info gracefully

#### 15.3.2: Progress Bar

**File**: `internal/cli/progress/bar.go`

```go
// Bar displays a progress bar
type Bar struct {
    width   int
    current int
    total   int
    message string
}

// Render generates progress bar string
func (b *Bar) Render() string
```

**Implementation**:
- [ ] Define Bar struct with state
- [ ] Implement Indicator interface
- [ ] Implement Render() for progress bar visualization
- [ ] Add percentage calculation and display
- [ ] Add estimated time remaining (ETA)
- [ ] Add elapsed time display
- [ ] Support customizable bar characters (█▓▒░)
- [ ] Handle terminal width constraints
- [ ] Update in place using ANSI cursor control
- [ ] Write progress bar tests

**Progress Bar Format**:
```
Installing packages [████████████░░░░░░░░] 60% (6/10) ETA 5s
```

**Test Cases**:
- Render at 0%, 50%, 100%
- Update progress incrementally
- Display messages correctly
- Calculate ETA accurately
- Handle terminal resize

#### 15.3.3: Spinner

**File**: `internal/cli/progress/spinner.go`

```go
// Spinner displays an indeterminate progress spinner
type Spinner struct {
    frames   []string
    current  int
    message  string
    ticker   *time.Ticker
}

// Spin advances to next frame
func (s *Spinner) Spin()
```

**Implementation**:
- [ ] Define Spinner struct with animation state
- [ ] Implement Indicator interface
- [ ] Define spinner frame sequences
- [ ] Implement Spin() to advance animation
- [ ] Add auto-spin with ticker in goroutine
- [ ] Support multiple spinner styles
- [ ] Update in place using ANSI cursor control
- [ ] Write spinner tests

**Spinner Styles**:
- Dots: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
- Line: - \ | /
- Arrows: ← ↖ ↑ ↗ → ↘ ↓ ↙

**Spinner Format**:
```
⠋ Scanning packages...
```

**Test Cases**:
- Spin through all frames
- Display messages correctly
- Start and stop cleanly
- Don't interfere with output

#### 15.3.4: Operation Progress Tracker

**File**: `internal/cli/progress/tracker.go`

```go
// Tracker manages progress for multi-stage operations
type Tracker struct {
    stages     []Stage
    current    int
    indicator  Indicator
}

// Stage represents one stage of a multi-stage operation
type Stage struct {
    Name     string
    Total    int
    Current  int
}

// Advance moves to next stage
func (t *Tracker) Advance()

// UpdateCurrent updates current stage progress
func (t *Tracker) UpdateCurrent(current int, message string)
```

**Implementation**:
- [ ] Define Tracker struct for multi-stage tracking
- [ ] Define Stage struct for stage information
- [ ] Implement Advance() for stage progression
- [ ] Implement UpdateCurrent() for within-stage updates
- [ ] Integrate with Indicator interface
- [ ] Support nested progress (overall + stage)
- [ ] Write tracker tests

**Multi-Stage Display**:
```
[1/3] Scanning packages  [████████████████████] 100% (10/10)
[2/3] Planning operations [████████░░░░░░░░░░░░]  45% (9/20)
[3/3] Executing plan      [░░░░░░░░░░░░░░░░░░░░]   0% (0/20)
```

**Test Cases**:
- Track single-stage operations
- Track multi-stage operations
- Update stage progress
- Advance between stages
- Display overall progress correctly

### 15.4: Help System

**Goal**: Comprehensive, discoverable help for all commands

#### 15.4.1: Help Generator

**File**: `internal/cli/help/generator.go`

```go
// Generator creates help text for commands
type Generator struct {
    width int
}

// Generate creates complete help text
func (g *Generator) Generate(cmd *cobra.Command) string

// GenerateUsage creates usage string
func (g *Generator) GenerateUsage(cmd *cobra.Command) string

// GenerateExamples creates examples section
func (g *Generator) GenerateExamples(cmd *cobra.Command) string
```

**Implementation**:
- [ ] Define Generator struct
- [ ] Implement Generate() for complete help
- [ ] Implement GenerateUsage() for usage line
- [ ] Implement GenerateExamples() for examples section
- [ ] Add flag descriptions with formatting
- [ ] Add command description with wrapping
- [ ] Group related flags together
- [ ] Write help generator tests

**Test Cases**:
- Generate help for each command
- Generate usage strings
- Format flag descriptions correctly
- Wrap long text appropriately

#### 15.4.2: Command Examples

**File**: `internal/cli/help/examples.go`

```go
// Example represents a command usage example
type Example struct {
    Description string
    Command     string
    Output      string  // Optional expected output
}

// Examples for each command
var (
    ManageExamples   []Example
    UnmanageExamples []Example
    RemanageExamples []Example
    AdoptExamples    []Example
    StatusExamples   []Example
    DoctorExamples   []Example
    ListExamples     []Example
)
```

**Implementation**:
- [ ] Define Example struct
- [ ] Create examples for manage command
- [ ] Create examples for unmanage command
- [ ] Create examples for remanage command
- [ ] Create examples for adopt command
- [ ] Create examples for status command
- [ ] Create examples for doctor command
- [ ] Create examples for list command
- [ ] Add examples with common flag combinations
- [ ] Format examples with syntax highlighting
- [ ] Write example tests

**Example Format**:
```
Examples:
  # Install a single package
  $ dot manage vim

  # Install multiple packages
  $ dot manage vim tmux zsh

  # Preview installation without applying
  $ dot manage --dry-run vim

  # Install with absolute symlinks
  $ dot manage --absolute vim
```

**Examples to Create for Each Command**:
- Basic usage (1-2 examples)
- Common flag combinations (2-3 examples)
- Complex scenarios (1-2 examples)
- Integration with other commands (1 example)

**Test Cases**:
- Verify example commands are valid
- Check example descriptions are clear
- Ensure examples cover common use cases

#### 15.4.3: Help Templates

**File**: `internal/cli/help/templates.go`

```go
// Template defines help text structure
type Template struct {
    Sections []Section
}

// Section represents a help section
type Section struct {
    Title   string
    Content string
}

// Predefined templates
var (
    CommandTemplate Template
    FlagTemplate    Template
)
```

**Implementation**:
- [ ] Define Template struct for help structure
- [ ] Create CommandTemplate for command help
- [ ] Create FlagTemplate for flag descriptions
- [ ] Add section templates (Usage, Description, Examples, Flags, See Also)
- [ ] Implement template rendering with styling
- [ ] Add cross-references between related commands
- [ ] Write template tests

**Help Template Structure**:
```
NAME
    dot manage - Install packages

SYNOPSIS
    dot manage [OPTIONS] PACKAGE...

DESCRIPTION
    Install packages by creating symlinks from the package directory
    to the target directory. Files in each package are linked to
    corresponding locations in the target.

OPTIONS
    -d, --dir PATH
        Stow directory containing packages (default: current directory)
    
    -t, --target PATH
        Target directory for symlinks (default: $HOME)
    
    -n, --dry-run
        Preview operations without applying changes

EXAMPLES
    [examples section]

SEE ALSO
    dot unmanage, dot remanage, dot status
```

**Test Cases**:
- Render complete help for each command
- Verify section ordering
- Check cross-references work
- Ensure consistent formatting

#### 15.4.4: Shell Completion

**File**: `internal/cli/help/completion.go`

```go
// CompletionGenerator creates shell completion scripts
type CompletionGenerator struct {
    rootCmd *cobra.Command
}

// GenerateBash creates bash completion
func (g *CompletionGenerator) GenerateBash() (string, error)

// GenerateZsh creates zsh completion
func (g *CompletionGenerator) GenerateZsh() (string, error)

// GenerateFish creates fish completion
func (g *CompletionGenerator) GenerateFish() (string, error)
```

**Implementation**:
- [ ] Define CompletionGenerator struct
- [ ] Implement GenerateBash() using Cobra's built-in support
- [ ] Implement GenerateZsh() using Cobra's built-in support
- [ ] Implement GenerateFish() using Cobra's built-in support
- [ ] Add dynamic completion for package names
- [ ] Add dynamic completion for paths
- [ ] Add completion for flag values
- [ ] Create `dot completion` command
- [ ] Write completion tests

**Completion Features**:
- Command name completion
- Flag name completion
- Package name completion (from package directory)
- Path completion for --dir and --target
- Flag value completion where applicable

**Test Cases**:
- Generate bash completion script
- Generate zsh completion script
- Generate fish completion script
- Verify dynamic package completion works
- Test flag value completion

### 15.5: User Experience Polish

**Goal**: Final touches for professional user experience

#### 15.5.1: Enhanced Error Reporting

**File**: `cmd/dot/errors.go`

```go
// formatError converts any error to user-friendly output
func formatError(err error) string

// printError outputs formatted error to stderr
func printError(err error)

// exitWithError prints error and exits with appropriate code
func exitWithError(err error) int
```

**Implementation**:
- [ ] Implement formatError() using error formatter
- [ ] Implement printError() with styling
- [ ] Implement exitWithError() with proper exit codes
- [ ] Integrate error formatting into all commands
- [ ] Add error context extraction from command state
- [ ] Ensure consistent error handling across CLI
- [ ] Write error reporting tests

**Exit Codes**:
- 0: Success
- 1: General error
- 2: Invalid arguments
- 3: Conflicts detected
- 4: Permission denied
- 5: Package not found

**Test Cases**:
- Format each error type correctly
- Print errors to stderr
- Exit with correct codes
- Include appropriate context

#### 15.5.2: Success Messages

**File**: `cmd/dot/messages.go`

```go
// printSuccess outputs success message
func printSuccess(message string)

// printSummary outputs operation summary
func printSummary(result ExecutionResult)

// printDryRunSummary outputs dry-run summary
func printDryRunSummary(plan Plan)
```

**Implementation**:
- [ ] Implement printSuccess() with success styling
- [ ] Implement printSummary() for execution results
- [ ] Implement printDryRunSummary() for dry-run mode
- [ ] Add operation statistics display
- [ ] Add duration display for operations
- [ ] Format summaries consistently across commands
- [ ] Write message tests

**Summary Format**:
```
✓ Successfully installed 3 packages

Summary:
  Packages: vim, tmux, zsh
  Links created: 42
  Directories created: 7
  Duration: 1.2s
```

**Dry-Run Summary Format**:
```
Dry-run: No changes will be applied

Planned Operations:
  Links to create: 42
  Directories to create: 7
  No conflicts detected

Run without --dry-run to apply changes.
```

**Test Cases**:
- Print success messages with correct styling
- Display execution summaries correctly
- Display dry-run summaries correctly
- Include relevant statistics

#### 15.5.3: Verbose Output

**File**: `cmd/dot/verbose.go`

```go
// VerboseLogger provides verbosity-aware logging
type VerboseLogger struct {
    level int
}

// Debug logs debug information (level 3+)
func (l *VerboseLogger) Debug(format string, args ...interface{})

// Info logs informational messages (level 2+)
func (l *VerboseLogger) Info(format string, args ...interface{})

// Summary logs summary information (level 1+)
func (l *VerboseLogger) Summary(format string, args ...interface{})
```

**Implementation**:
- [ ] Define VerboseLogger struct
- [ ] Implement Debug() for level 3+ messages
- [ ] Implement Info() for level 2+ messages
- [ ] Implement Summary() for level 1+ messages
- [ ] Integrate with existing logger
- [ ] Add operation progress logging at appropriate levels
- [ ] Write verbose logging tests

**Verbosity Levels**:
- Level 0 (default): Errors and final summary only
- Level 1 (-v): High-level operation summary
- Level 2 (-vv): Per-operation progress
- Level 3 (-vvv): Detailed internal state

**Output Examples**:

Level 1:
```
Scanning 3 packages...
Planning operations...
Executing 42 operations...
✓ Successfully installed 3 packages
```

Level 2:
```
Scanning package: vim (10 files)
Scanning package: tmux (5 files)
Scanning package: zsh (8 files)
Planning: 42 operations
Creating link: ~/.vimrc -> stow/vim/vimrc
Creating link: ~/.tmux.conf -> stow/tmux/tmux.conf
...
✓ Successfully installed 3 packages
```

Level 3:
```
[DEBUG] Loading configuration from ~/.dotrc
[DEBUG] Stow directory: /home/user/dotfiles
[DEBUG] Target directory: /home/user
[DEBUG] Scanning package: vim
[DEBUG] Found file: vimrc (1024 bytes)
[DEBUG] Mapping: vimrc -> .vimrc
[DEBUG] Computing desired state...
[DEBUG] Desired links: 42
[DEBUG] Current links: 0
[DEBUG] Diff: 42 create operations
...
```

**Test Cases**:
- Log at each verbosity level
- Verify output matches level
- Test with progress indicators
- Ensure structured logging for --log-json

#### 15.5.4: Quiet Mode

**File**: `cmd/dot/quiet.go`

```go
// QuietOutput suppresses all non-error output
type QuietOutput struct {
    enabled bool
}

// IsEnabled checks if quiet mode is active
func (q *QuietOutput) IsEnabled() bool

// ShouldPrint determines if message should print
func (q *QuietOutput) ShouldPrint(level MessageLevel) bool
```

**Implementation**:
- [ ] Define QuietOutput struct
- [ ] Implement IsEnabled() check
- [ ] Implement ShouldPrint() with level filtering
- [ ] Suppress all indicators in quiet mode
- [ ] Suppress all success messages in quiet mode
- [ ] Allow only errors in quiet mode
- [ ] Ensure exit codes still indicate success/failure
- [ ] Write quiet mode tests

**Quiet Mode Behavior**:
- No progress indicators
- No success messages
- No summaries
- Only error output to stderr
- Exit codes indicate status

**Test Cases**:
- Verify no output on success
- Verify errors still print
- Verify exit codes correct
- Test with various commands

### 15.6: Integration and Testing

**Goal**: Comprehensive testing of error handling and UX features

#### 15.6.1: Unit Tests

**Implementation**:
- [ ] Test error formatter with all error types
- [ ] Test suggestion engine with various errors
- [ ] Test style system on different terminals
- [ ] Test progress indicators (bar, spinner, tracker)
- [ ] Test help generator for all commands
- [ ] Test completion generation
- [ ] Test verbose and quiet modes
- [ ] Achieve 80%+ coverage for new code

**Test Coverage Areas**:
- Error formatting: All error types, all contexts
- Styling: All styles, color modes, terminal types
- Progress: All indicator types, all states
- Help: All commands, all sections
- Completion: All shells, dynamic completion

#### 15.6.2: Integration Tests

**File**: `tests/integration/ux_test.go`

**Implementation**:
- [ ] Test error formatting in actual command execution
- [ ] Test progress indicators with real operations
- [ ] Test help display for all commands
- [ ] Test completion installation and usage
- [ ] Test verbose output at all levels
- [ ] Test quiet mode end-to-end
- [ ] Write comprehensive integration tests

**Test Scenarios**:
- Execute manage command with errors, verify error output
- Execute long operation, verify progress display
- Request help for each command, verify completeness
- Generate and test shell completion
- Run with -v, -vv, -vvv, verify output levels
- Run with --quiet, verify silence on success

#### 15.6.3: Visual Regression Tests

**File**: `tests/integration/visual_test.go`

**Implementation**:
- [ ] Capture output snapshots for error messages
- [ ] Capture output snapshots for help text
- [ ] Capture output snapshots for progress indicators
- [ ] Compare against golden files
- [ ] Detect unintended output changes
- [ ] Write visual regression tests

**Golden Files**:
- `testdata/golden/error_conflict.txt`
- `testdata/golden/error_permission.txt`
- `testdata/golden/help_manage.txt`
- `testdata/golden/help_status.txt`
- `testdata/golden/progress_bar.txt`
- `testdata/golden/summary_success.txt`

**Test Cases**:
- Error message formatting matches golden files
- Help text matches golden files
- Progress indicators match golden files
- Summaries match golden files

#### 15.6.4: Accessibility Testing

**Implementation**:
- [ ] Test with NO_COLOR environment variable
- [ ] Test with --color=never flag
- [ ] Test on terminals without color support
- [ ] Test with screen readers (via text output)
- [ ] Verify contrast ratios for color choices
- [ ] Test with various terminal widths
- [ ] Write accessibility tests

**Accessibility Checks**:
- All information conveyed with and without color
- Contrast ratios meet WCAG AA standards (4.5:1)
- Text wraps correctly at various widths
- Plain text output is screen reader friendly
- Progress indicators degrade gracefully

### 15.7: Documentation

**Goal**: Document error handling and UX features

#### 15.7.1: User Documentation

**File**: `docs/errors.md`

**Implementation**:
- [ ] Document common errors and solutions
- [ ] Create error reference guide
- [ ] Add troubleshooting flowcharts
- [ ] Document exit codes
- [ ] Write user error documentation

**Contents**:
- Error types and meanings
- Common causes and solutions
- Exit code reference
- Troubleshooting guide
- FAQ for error scenarios

#### 15.7.2: Help System Documentation

**File**: `docs/help.md`

**Implementation**:
- [ ] Document help system usage
- [ ] Explain shell completion setup
- [ ] Add examples of help usage
- [ ] Write help system documentation

**Contents**:
- Getting help for commands
- Shell completion installation
- Man page installation
- Examples reference

#### 15.7.3: Developer Documentation

**File**: `docs/dev/error-handling.md`

**Implementation**:
- [ ] Document error handling architecture
- [ ] Explain error formatting pipeline
- [ ] Document adding new error types
- [ ] Document styling system
- [ ] Document progress indicators
- [ ] Write developer documentation

**Contents**:
- Error handling architecture
- Adding new error types
- Error template creation
- Styling system usage
- Progress indicator integration

## Testing Strategy

### Unit Testing

**Coverage Target**: 80% minimum for all new code

**Test Focus**:
- Error formatter logic
- Suggestion generation
- Style application
- Progress indicator state management
- Help text generation
- Completion script generation

**Test Approach**:
- Table-driven tests for error formatting
- Property tests for suggestion relevance
- Snapshot tests for help text
- Mock terminal for style testing

### Integration Testing

**Test Focus**:
- Error display in real commands
- Progress indicators during operations
- Help system completeness
- Shell completion functionality

**Test Approach**:
- Execute commands with various error conditions
- Verify output matches expectations
- Test interactive and non-interactive modes
- Validate golden file outputs

### Manual Testing

**Test Focus**:
- Visual appearance in real terminals
- Color scheme accessibility
- Progress indicator smoothness
- Help text readability

**Test Checklist**:
- [ ] Test in iTerm2, Terminal.app, Windows Terminal
- [ ] Test in Alacritty, Kitty, Gnome Terminal
- [ ] Test in tmux and screen
- [ ] Test with various terminal widths (80, 120, 160 columns)
- [ ] Test with color and without color
- [ ] Test on dark and light terminal backgrounds
- [ ] Test with high contrast themes
- [ ] Verify with accessibility tools

## Acceptance Criteria

Phase 15 is complete when:

### Functionality
- [ ] All domain errors have user-friendly formatting
- [ ] All errors include actionable suggestions
- [ ] Progress indicators work for all long operations
- [ ] Help system covers all commands comprehensively
- [ ] Shell completion works for bash, zsh, and fish
- [ ] Verbose mode provides appropriate detail at each level
- [ ] Quiet mode suppresses all non-error output
- [ ] Exit codes correctly indicate operation status

### Quality
- [ ] Unit test coverage ≥ 80% for new code
- [ ] Integration tests pass for all scenarios
- [ ] Visual regression tests detect output changes
- [ ] Accessibility tests pass for all modes
- [ ] Manual testing checklist complete
- [ ] No linter warnings
- [ ] All tests pass

### Documentation
- [ ] Error reference guide complete
- [ ] Help system documentation complete
- [ ] Developer documentation complete
- [ ] Examples demonstrate all features
- [ ] Troubleshooting guide addresses common issues

### User Experience
- [ ] Error messages are clear and actionable
- [ ] Progress indicators provide meaningful feedback
- [ ] Help text is comprehensive and discoverable
- [ ] Output is visually consistent across commands
- [ ] Styling works on all supported terminals
- [ ] Accessibility requirements met

## Rollout Plan

### Step 1: Error Formatting Foundation (15.1)
- Implement error formatter, templates, suggestions
- Test with existing domain errors
- Integrate into one command (manage) as proof of concept

### Step 2: Terminal Styling (15.2)
- Implement styling system
- Test on various terminals
- Apply to error messages from Step 1

### Step 3: Progress Indicators (15.3)
- Implement bar, spinner, tracker
- Integrate into long-running operations
- Test interactive and non-interactive modes

### Step 4: Help System (15.4)
- Generate help text for all commands
- Create examples for all commands
- Implement shell completion

### Step 5: UX Polish (15.5)
- Integrate error formatting into all commands
- Add success messages and summaries
- Implement verbose and quiet modes

### Step 6: Testing and Documentation (15.6, 15.7)
- Complete test suite
- Run accessibility tests
- Write all documentation
- Perform manual testing

## Dependencies

### External Libraries

Consider using:
- `github.com/charmbracelet/lipgloss` for styling (optional)
- `github.com/mattn/go-isatty` for terminal detection
- `github.com/muesli/termenv` for terminal capability detection
- Standard library `golang.org/x/term` for terminal width

**Decision**: Evaluate libraries vs custom implementation
- **Pro libraries**: Save time, battle-tested
- **Con libraries**: External dependencies
- **Recommendation**: Use isatty and termenv, implement custom styling

### Internal Dependencies

- Phase 13: CLI commands for integration
- Phase 14: Query commands for error scenarios
- Domain errors from earlier phases

## Risks and Mitigations

### Risk: Terminal Compatibility Issues
**Mitigation**: Extensive testing on various terminals, graceful degradation

### Risk: Performance Impact of Styling
**Mitigation**: Profile styling overhead, cache styled strings

### Risk: Accessibility Concerns
**Mitigation**: Follow WCAG guidelines, test with tools, provide plain text mode

### Risk: Complex Error Formatting Logic
**Mitigation**: Keep templates simple, test thoroughly, use table-driven tests

### Risk: Progress Indicators Interfering with Output
**Mitigation**: Proper ANSI cursor control, test non-interactive mode

## Success Metrics

### Quantitative
- Error message clarity score from user testing: ≥ 4.0/5.0
- Time to resolve errors reduced by 50% (compared to generic errors)
- Help system usage: ≥ 80% of users find answers without external docs
- Test coverage: ≥ 80%
- Zero accessibility violations

### Qualitative
- Users report errors are understandable and actionable
- Help text is comprehensive and easy to navigate
- Progress indicators provide confidence during operations
- Output is visually appealing and professional
- Accessibility needs are met

## Timeline Estimate

Total: 40-48 hours

Breakdown:
- 15.1 Error Formatting: 8-10 hours
- 15.2 Terminal Styling: 6-8 hours
- 15.3 Progress Indicators: 8-10 hours
- 15.4 Help System: 8-10 hours
- 15.5 UX Polish: 4-6 hours
- 15.6 Integration Testing: 4-5 hours
- 15.7 Documentation: 2-3 hours

## Deliverables

1. Complete error formatting system with templates
2. Comprehensive help system with examples
3. Working progress indicators for all operation types
4. Shell completion for bash, zsh, fish
5. Verbose and quiet modes
6. Test suite with ≥ 80% coverage
7. Complete user and developer documentation
8. Accessibility-compliant output

## Next Phase

After Phase 15 completion, proceed to:
- **Phase 16**: Property-Based Testing - Verify algebraic laws and invariants

