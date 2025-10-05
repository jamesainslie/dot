# Phase 15b: Configuration Management - Detailed Implementation Plan

## Overview

Phase 15b introduces comprehensive configuration management for the dot CLI through XDG-compliant configuration files and a `config` command for managing settings. This phase transforms the current flag-based configuration into a persistent, layered system that supports defaults, configuration files, environment variables, and command-line flags.

## Prerequisites

- Phase 13 complete: CLI core commands functional
- Phase 14 complete: Query commands implemented
- Phase 15 complete: Error handling and UX system implemented
- Current config package exists: `internal/config/config.go`

## Design Principles

- **XDG Compliance**: Configuration files follow XDG Base Directory specification
- **Layered Configuration**: Clear precedence: CLI flags > Environment > Config file > Defaults
- **Validation First**: All configuration values validated before application
- **Security Conscious**: Proper file permissions, path validation, no sensitive data exposure
- **User-Friendly**: Interactive configuration, clear error messages, helpful examples
- **Backward Compatible**: Existing flags continue to work without configuration file

## Architecture

### Configuration Precedence

```
┌─────────────────┐
│  CLI Flags      │ ← Highest precedence
├─────────────────┤
│  Environment    │
├─────────────────┤
│  Config File    │
├─────────────────┤
│  Defaults       │ ← Lowest precedence
└─────────────────┘
```

### Configuration Flow

```
User Command → Parse Flags → Load Environment → Load Config File → Merge with Defaults → Validate → Execute
```

### Component Structure

```
internal/
├── config/
│   ├── config.go              # Extended Config struct (existing)
│   ├── config_test.go         # Config tests (existing)
│   ├── loader.go              # Configuration loading logic
│   ├── loader_test.go         # Loader tests
│   ├── validator.go           # Configuration validation
│   ├── validator_test.go      # Validator tests
│   ├── writer.go              # Configuration file writing
│   ├── writer_test.go         # Writer tests
│   └── schema.go              # Configuration schema and documentation

cmd/dot/
├── config.go                  # Config command implementation
├── config_test.go             # Config command tests
├── config_init.go             # Init subcommand
├── config_get.go              # Get subcommand
├── config_set.go              # Set subcommand
├── config_list.go             # List subcommand
├── config_edit.go             # Edit subcommand
└── config_validate.go         # Validate subcommand
```

## Configuration Schema

### Complete Configuration Structure

```yaml
# Dot Configuration File
# Location: $XDG_CONFIG_HOME/dot/config.yaml (default: ~/.config/dot/config.yaml)
# Documentation: https://github.com/user/dot/docs/configuration.md

# Core Directories
directories:
  # Stow directory containing packages
  # Default: current directory
  stow: ~/dotfiles
  
  # Target directory for symlinks
  # Default: $HOME
  target: ~
  
  # Manifest directory for tracking
  # Default: $XDG_DATA_HOME/dot/manifest (or ~/.local/share/dot/manifest)
  manifest: ~/.local/share/dot/manifest

# Logging Configuration
logging:
  # Log level: DEBUG, INFO, WARN, ERROR
  # Default: INFO
  level: INFO
  
  # Log format: text, json
  # Default: text
  format: text
  
  # Log destination: stderr, stdout, file
  # Default: stderr
  destination: stderr
  
  # Log file path (only used if destination is "file")
  # Default: $XDG_STATE_HOME/dot/dot.log (or ~/.local/state/dot/dot.log)
  file: ~/.local/state/dot/dot.log

# Symlink Behavior
symlinks:
  # Link mode: relative, absolute
  # Default: relative
  mode: relative
  
  # Enable directory folding optimization
  # Default: true
  folding: true
  
  # Overwrite existing files when conflicts occur
  # Default: false (fail on conflicts)
  overwrite: false
  
  # Create backup of overwritten files
  # Default: false
  backup: false
  
  # Backup suffix when backups enabled
  # Default: .bak
  backup_suffix: .bak

# Ignore Patterns
ignore:
  # Use default ignore patterns
  # Default: true
  # Defaults: .git, .svn, .hg, .DS_Store, Thumbs.db, desktop.ini, .Trash, .Spotlight-V100, .TemporaryItems
  use_defaults: true
  
  # Additional patterns to ignore (glob format)
  # Default: [] (empty list)
  patterns:
    - "*.swp"
    - "*.tmp"
    - ".*.un~"
    - "*~"
    - "#*#"
  
  # Patterns to override (force include even if ignored)
  # Default: [] (empty list)
  overrides: []

# Dotfile Translation
dotfile:
  # Enable dot- to . translation (e.g., dot-vimrc → .vimrc)
  # Default: true
  translate: true
  
  # Prefix for dotfile translation
  # Default: dot-
  prefix: dot-

# Output Configuration
output:
  # Default output format: text, json, yaml, table
  # Default: text
  format: text
  
  # Enable colored output: auto, always, never
  # Default: auto
  color: auto
  
  # Show progress indicators
  # Default: true
  progress: true
  
  # Verbosity level: 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)
  # Default: 1
  verbosity: 1
  
  # Terminal width for text wrapping (0 = auto-detect)
  # Default: 0
  width: 0

# Operation Defaults
operations:
  # Enable dry-run mode by default
  # Default: false
  dry_run: false
  
  # Enable atomic operations with rollback
  # Default: true
  atomic: true
  
  # Maximum number of parallel operations
  # Default: 0 (auto-detect CPU count)
  max_parallel: 0

# Package Management
packages:
  # Default sort order for list command: name, links, date
  # Default: name
  sort_by: name
  
  # Automatically scan for new packages
  # Default: true
  auto_discover: true
  
  # Package naming convention validation
  # Default: true
  validate_names: true

# Doctor Configuration
doctor:
  # Auto-fix issues when possible
  # Default: false
  auto_fix: false
  
  # Check manifest integrity
  # Default: true
  check_manifest: true
  
  # Check for broken symlinks
  # Default: true
  check_broken_links: true
  
  # Check for orphaned links
  # Default: true
  check_orphaned: true
  
  # Check file permissions
  # Default: true
  check_permissions: true

# Experimental Features
experimental:
  # Enable parallel operations
  # Default: false
  parallel: false
  
  # Enable performance profiling
  # Default: false
  profiling: false
```

### Environment Variable Mapping

All configuration values can be set via environment variables with `DOT_` prefix and underscores for hierarchy:

```bash
# Examples
export DOT_DIRECTORIES_STOW=/path/to/dotfiles
export DOT_DIRECTORIES_TARGET=/home/user
export DOT_LOGGING_LEVEL=DEBUG
export DOT_LOGGING_FORMAT=json
export DOT_SYMLINKS_MODE=absolute
export DOT_SYMLINKS_FOLDING=false
export DOT_OUTPUT_COLOR=never
export DOT_OUTPUT_VERBOSITY=2
export DOT_OPERATIONS_DRY_RUN=true
export DOT_IGNORE_PATTERNS="*.swp,*.tmp"
```

### Configuration File Formats

Support three formats based on file extension:

```yaml
# YAML: config.yaml (recommended)
directories:
  stow: ~/dotfiles
  target: ~
```

```json
// JSON: config.json
{
  "directories": {
    "stow": "~/dotfiles",
    "target": "~"
  }
}
```

```toml
# TOML: config.toml
[directories]
stow = "~/dotfiles"
target = "~"
```

## Config Command Design

### Command Structure

```
dot config                    # Show current configuration
dot config init               # Create initial config file
dot config get <key>          # Get configuration value
dot config set <key> <value>  # Set configuration value
dot config unset <key>        # Remove configuration value
dot config list               # List all settings (alias for base command)
dot config edit               # Open config file in $EDITOR
dot config path               # Show config file path
dot config validate           # Validate configuration file
dot config reset              # Reset to defaults
```

### Subcommand: init

Create initial configuration file with sensible defaults and interactive prompts.

```bash
# Create config with defaults
dot config init

# Create config with prompts
dot config init --interactive

# Create config without comments
dot config init --no-comments

# Force overwrite existing config
dot config init --force

# Use specific format
dot config init --format yaml
dot config init --format json
dot config init --format toml
```

**Behavior**:
- Detects existing config file and prompts before overwriting (unless `--force`)
- Creates XDG config directory if it doesn't exist
- Sets secure file permissions (0600)
- Includes helpful comments explaining each option (unless `--no-comments`)
- In interactive mode, prompts for key settings with current values as defaults

**Example Interactive Session**:
```
$ dot config init --interactive

Initializing dot configuration file...
Config will be created at: /home/user/.config/dot/config.yaml

Directories:
  Stow directory [.]: /home/user/dotfiles
  Target directory [/home/user]: 
  
Symlinks:
  Use relative symlinks? [Y/n]: 
  Enable directory folding? [Y/n]: 
  
Logging:
  Log level (DEBUG, INFO, WARN, ERROR) [INFO]: 
  Log format (text, json) [text]: 

Configuration file created successfully.
Edit with: dot config edit
```

### Subcommand: get

Retrieve configuration value(s) by key path.

```bash
# Get single value
dot config get directories.stow
# Output: /home/user/dotfiles

# Get section (outputs in format specified by --format)
dot config get directories
# Output:
# stow: /home/user/dotfiles
# target: /home/user
# manifest: /home/user/.local/share/dot/manifest

# Get with specific format
dot config get logging --format json
# Output: {"level":"INFO","format":"text","destination":"stderr"}

# Get from specific source
dot config get directories.stow --source file
dot config get directories.stow --source env
dot config get directories.stow --source merged  # default: show final value

# Show all sources for a key
dot config get directories.stow --show-source
# Output:
# directories.stow
#   default: .
#   file: /home/user/dotfiles
#   environment: not set
#   flag: not set
#   final: /home/user/dotfiles
```

**Output Formats**:
- Plain text (default): Just the value
- JSON: Machine-readable format
- YAML: Structured format
- Table: Tabular format for multiple values

### Subcommand: set

Set configuration value by key path.

```bash
# Set string value
dot config set directories.stow ~/dotfiles

# Set boolean value
dot config set symlinks.folding true
dot config set symlinks.folding false

# Set numeric value
dot config set output.verbosity 2

# Set list value
dot config set ignore.patterns '*.swp,*.tmp,*~'

# Set nested value with dotted key path
dot config set logging.level DEBUG

# Validate before setting
dot config set directories.stow ~/invalid --validate

# Set without creating file (dry-run)
dot config set directories.stow ~/dotfiles --dry-run
```

**Behavior**:
- Creates config file if it doesn't exist (prompts first)
- Validates value before writing
- Preserves comments and formatting when possible
- Shows old and new values
- Returns error if value is invalid

**Example Output**:
```
$ dot config set directories.stow ~/dotfiles
Updated configuration: /home/user/.config/dot/config.yaml
  directories.stow: . → /home/user/dotfiles
```

### Subcommand: unset

Remove configuration value, reverting to default.

```bash
# Unset specific value
dot config unset directories.stow

# Unset entire section
dot config unset logging

# Unset list element
dot config unset ignore.patterns[0]
```

**Example Output**:
```
$ dot config unset directories.stow
Updated configuration: /home/user/.config/dot/config.yaml
  directories.stow: /home/user/dotfiles → . (default)
```

### Subcommand: list

Display all configuration settings with their sources.

```bash
# List all settings
dot config list

# List specific section
dot config list logging

# Show sources
dot config list --show-source

# Show only non-default values
dot config list --non-default

# List in specific format
dot config list --format json
dot config list --format yaml
dot config list --format table
```

**Example Output (table format with sources)**:
```
$ dot config list --show-source --format table

┌───────────────────────────┬─────────────────────┬──────────┬─────────────┐
│ Key                       │ Value               │ Source   │ Default     │
├───────────────────────────┼─────────────────────┼──────────┼─────────────┤
│ directories.stow          │ /home/user/dotfiles │ file     │ .           │
│ directories.target        │ /home/user          │ default  │ /home/user  │
│ directories.manifest      │ ~/.local/share/...  │ default  │ ~/.local... │
│ logging.level             │ DEBUG               │ env      │ INFO        │
│ logging.format            │ json                │ flag     │ text        │
│ symlinks.mode             │ relative            │ default  │ relative    │
│ symlinks.folding          │ true                │ default  │ true        │
│ output.color              │ always              │ file     │ auto        │
│ output.verbosity          │ 2                   │ flag     │ 1           │
└───────────────────────────┴─────────────────────┴──────────┴─────────────┘

Legend:
  default = built-in default value
  file    = configuration file
  env     = environment variable
  flag    = command-line flag
```

### Subcommand: edit

Open configuration file in user's preferred editor.

```bash
# Open in $EDITOR (or fallback to vim/nano/vi)
dot config edit

# Open in specific editor
dot config edit --editor nano
dot config edit --editor code

# Validate after editing
dot config edit --validate
```

**Behavior**:
- Detects editor from: `--editor` flag > `$VISUAL` > `$EDITOR` > fallback sequence
- Fallback sequence: `vim`, `nano`, `vi`, `notepad` (Windows)
- Creates config file if it doesn't exist
- Validates after editing if `--validate` flag set
- Shows validation errors and prompts to re-edit

**Example Session**:
```
$ dot config edit --validate
Opening /home/user/.config/dot/config.yaml in vim...
[Editor opens]
[User saves and exits]
Validating configuration...
Error: Invalid value for logging.level: "TRACE" (must be DEBUG, INFO, WARN, or ERROR)

Open editor again? [y/N]: y
[Editor opens again]
```

### Subcommand: path

Display configuration file paths and status.

```bash
# Show config file path
dot config path

# Show all config-related paths
dot config path --all

# Check if config file exists
dot config path --check
```

**Example Output**:
```
$ dot config path --all

Configuration file: /home/user/.config/dot/config.yaml ✓
Manifest directory: /home/user/.local/share/dot/manifest ✓
Log file: /home/user/.local/state/dot/dot.log ✗ (not created yet)

XDG directories:
  XDG_CONFIG_HOME: /home/user/.config
  XDG_DATA_HOME: /home/user/.local/share
  XDG_STATE_HOME: /home/user/.local/state
```

### Subcommand: validate

Validate configuration file without applying changes.

```bash
# Validate current config
dot config validate

# Validate specific file
dot config validate --file ~/test-config.yaml

# Verbose validation
dot config validate --verbose
```

**Example Output (valid config)**:
```
$ dot config validate
Validating /home/user/.config/dot/config.yaml...
✓ Configuration is valid

Validation summary:
  Total settings: 28
  Using defaults: 20
  Custom values: 8
  Warnings: 0
  Errors: 0
```

**Example Output (invalid config)**:
```
$ dot config validate
Validating /home/user/.config/dot/config.yaml...
✗ Configuration is invalid

Errors:
  Line 8: logging.level: invalid value "TRACE" (must be DEBUG, INFO, WARN, or ERROR)
  Line 15: directories.stow: path does not exist: /invalid/path
  Line 22: symlinks.mode: invalid value "hardlink" (must be relative or absolute)

Warnings:
  Line 30: ignore.patterns: pattern "**/*.swp" may be overly broad

Fix these errors and run 'dot config validate' again.
```

### Subcommand: reset

Reset configuration to defaults, optionally preserving some values.

```bash
# Reset all configuration
dot config reset

# Reset specific section
dot config reset logging

# Reset interactively (confirms each section)
dot config reset --interactive

# Create backup before reset
dot config reset --backup
```

**Example Interactive Session**:
```
$ dot config reset --interactive --backup

This will reset configuration to defaults.
A backup will be created at: /home/user/.config/dot/config.yaml.bak

Reset directories? [y/N]: n
Reset logging? [y/N]: y
Reset symlinks? [y/N]: n
Reset ignore? [y/N]: y
Reset output? [y/N]: n

Configuration reset complete.
Reset settings: logging, ignore
Preserved settings: directories, symlinks, output
```

## Integration with Existing CLI

### Command Flag Mapping

Map existing global flags to configuration keys:

| Flag | Config Key | Environment Variable |
|------|-----------|---------------------|
| `-d, --dir` | `directories.stow` | `DOT_DIRECTORIES_STOW` |
| `-t, --target` | `directories.target` | `DOT_DIRECTORIES_TARGET` |
| `-n, --dry-run` | `operations.dry_run` | `DOT_OPERATIONS_DRY_RUN` |
| `-v, --verbose` | `output.verbosity` | `DOT_OUTPUT_VERBOSITY` |
| `-q, --quiet` | `output.verbosity` (0) | `DOT_OUTPUT_VERBOSITY=0` |
| `--log-json` | `logging.format` (json) | `DOT_LOGGING_FORMAT=json` |
| `--color` | `output.color` | `DOT_OUTPUT_COLOR` |
| `--format` | `output.format` | `DOT_OUTPUT_FORMAT` |

### Loading Order in Commands

Update `buildConfig()` in `cmd/dot/root.go`:

```go
func buildConfig() (dot.Config, error) {
    // 1. Load defaults
    cfg := config.Default()
    
    // 2. Load config file
    configPath := filepath.Join(config.GetConfigPath("dot"), "config.yaml")
    if fileExists(configPath) {
        fileCfg, err := config.LoadFromFile(configPath)
        if err != nil {
            return dot.Config{}, fmt.Errorf("load config file: %w", err)
        }
        cfg = config.Merge(cfg, fileCfg)
    }
    
    // 3. Apply environment variables
    envCfg := config.LoadFromEnv()
    cfg = config.Merge(cfg, envCfg)
    
    // 4. Apply command-line flags (highest precedence)
    flagCfg := configFromFlags()
    cfg = config.Merge(cfg, flagCfg)
    
    // 5. Validate merged configuration
    if err := cfg.Validate(); err != nil {
        return dot.Config{}, fmt.Errorf("invalid configuration: %w", err)
    }
    
    // 6. Convert to dot.Config and return
    return cfg.ToDotConfig(), nil
}
```

## Implementation Tasks

### Task 15b.1: Extended Configuration Schema

**Goal**: Expand configuration struct to support all configurable options

#### 15b.1.1: Extend Config Struct

**File**: `internal/config/config.go`

**Implementation**:
```go
// Config contains all application configuration.
type Config struct {
    Directories DirectoriesConfig `mapstructure:"directories"`
    Logging     LoggingConfig     `mapstructure:"logging"`
    Symlinks    SymlinksConfig    `mapstructure:"symlinks"`
    Ignore      IgnoreConfig      `mapstructure:"ignore"`
    Dotfile     DotfileConfig     `mapstructure:"dotfile"`
    Output      OutputConfig      `mapstructure:"output"`
    Operations  OperationsConfig  `mapstructure:"operations"`
    Packages    PackagesConfig    `mapstructure:"packages"`
    Doctor      DoctorConfig      `mapstructure:"doctor"`
    Experimental ExperimentalConfig `mapstructure:"experimental"`
}

type DirectoriesConfig struct {
    Stow     string `mapstructure:"stow"`
    Target   string `mapstructure:"target"`
    Manifest string `mapstructure:"manifest"`
}

type LoggingConfig struct {
    Level       string `mapstructure:"level"`
    Format      string `mapstructure:"format"`
    Destination string `mapstructure:"destination"`
    File        string `mapstructure:"file"`
}

type SymlinksConfig struct {
    Mode         string `mapstructure:"mode"`
    Folding      bool   `mapstructure:"folding"`
    Overwrite    bool   `mapstructure:"overwrite"`
    Backup       bool   `mapstructure:"backup"`
    BackupSuffix string `mapstructure:"backup_suffix"`
}

type IgnoreConfig struct {
    UseDefaults bool     `mapstructure:"use_defaults"`
    Patterns    []string `mapstructure:"patterns"`
    Overrides   []string `mapstructure:"overrides"`
}

type DotfileConfig struct {
    Translate bool   `mapstructure:"translate"`
    Prefix    string `mapstructure:"prefix"`
}

type OutputConfig struct {
    Format    string `mapstructure:"format"`
    Color     string `mapstructure:"color"`
    Progress  bool   `mapstructure:"progress"`
    Verbosity int    `mapstructure:"verbosity"`
    Width     int    `mapstructure:"width"`
}

type OperationsConfig struct {
    DryRun      bool `mapstructure:"dry_run"`
    Atomic      bool `mapstructure:"atomic"`
    MaxParallel int  `mapstructure:"max_parallel"`
}

type PackagesConfig struct {
    SortBy        string `mapstructure:"sort_by"`
    AutoDiscover  bool   `mapstructure:"auto_discover"`
    ValidateNames bool   `mapstructure:"validate_names"`
}

type DoctorConfig struct {
    AutoFix          bool `mapstructure:"auto_fix"`
    CheckManifest    bool `mapstructure:"check_manifest"`
    CheckBrokenLinks bool `mapstructure:"check_broken_links"`
    CheckOrphaned    bool `mapstructure:"check_orphaned"`
    CheckPermissions bool `mapstructure:"check_permissions"`
}

type ExperimentalConfig struct {
    Parallel  bool `mapstructure:"parallel"`
    Profiling bool `mapstructure:"profiling"`
}
```

**Tasks**:
- [ ] Define all configuration structs
- [ ] Add mapstructure tags for Viper unmarshaling
- [ ] Add JSON/YAML/TOML tags for serialization
- [ ] Write struct tests
- [ ] Document each field with comments

**Test Cases**:
- Unmarshal YAML to Config
- Unmarshal JSON to Config
- Unmarshal TOML to Config
- Marshal Config to YAML
- Handle missing fields (use defaults)
- Handle invalid types

#### 15b.1.2: Default Configuration

**File**: `internal/config/config.go`

**Implementation**:
```go
// Default returns configuration with sensible defaults.
func Default() *Config {
    homeDir, _ := os.UserHomeDir()
    if homeDir == "" {
        homeDir = "."
    }
    
    return &Config{
        Directories: DirectoriesConfig{
            Stow:     ".",
            Target:   homeDir,
            Manifest: getXDGDataPath("dot/manifest"),
        },
        Logging: LoggingConfig{
            Level:       "INFO",
            Format:      "text",
            Destination: "stderr",
            File:        getXDGStatePath("dot/dot.log"),
        },
        Symlinks: SymlinksConfig{
            Mode:         "relative",
            Folding:      true,
            Overwrite:    false,
            Backup:       false,
            BackupSuffix: ".bak",
        },
        Ignore: IgnoreConfig{
            UseDefaults: true,
            Patterns:    []string{},
            Overrides:   []string{},
        },
        Dotfile: DotfileConfig{
            Translate: true,
            Prefix:    "dot-",
        },
        Output: OutputConfig{
            Format:    "text",
            Color:     "auto",
            Progress:  true,
            Verbosity: 1,
            Width:     0,
        },
        Operations: OperationsConfig{
            DryRun:      false,
            Atomic:      true,
            MaxParallel: 0,
        },
        Packages: PackagesConfig{
            SortBy:        "name",
            AutoDiscover:  true,
            ValidateNames: true,
        },
        Doctor: DoctorConfig{
            AutoFix:          false,
            CheckManifest:    true,
            CheckBrokenLinks: true,
            CheckOrphaned:    true,
            CheckPermissions: true,
        },
        Experimental: ExperimentalConfig{
            Parallel:  false,
            Profiling: false,
        },
    }
}

// getXDGDataPath returns XDG data directory path
func getXDGDataPath(suffix string) string {
    if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
        return filepath.Join(dataHome, suffix)
    }
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".local", "share", suffix)
}

// getXDGStatePath returns XDG state directory path
func getXDGStatePath(suffix string) string {
    if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
        return filepath.Join(stateHome, suffix)
    }
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".local", "state", suffix)
}
```

**Tasks**:
- [ ] Implement Default() function
- [ ] Add XDG directory helpers
- [ ] Handle missing HOME directory gracefully
- [ ] Write default config tests

**Test Cases**:
- Default config has all required fields
- XDG paths are correct
- Handles missing environment variables
- Cross-platform compatibility

### Task 15b.2: Configuration Validation

**Goal**: Comprehensive validation of all configuration values

#### 15b.2.1: Validator Implementation

**File**: `internal/config/validator.go`

**Implementation**:
```go
// ValidationError represents a configuration validation error.
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult contains validation outcome.
type ValidationResult struct {
    Errors   []ValidationError
    Warnings []ValidationError
}

// IsValid returns true if no errors exist.
func (r *ValidationResult) IsValid() bool {
    return len(r.Errors) == 0
}

// Validate checks configuration for errors.
func (c *Config) Validate() error {
    result := &ValidationResult{
        Errors:   []ValidationError{},
        Warnings: []ValidationError{},
    }
    
    c.validateDirectories(result)
    c.validateLogging(result)
    c.validateSymlinks(result)
    c.validateIgnore(result)
    c.validateDotfile(result)
    c.validateOutput(result)
    c.validateOperations(result)
    c.validatePackages(result)
    c.validateDoctor(result)
    
    if !result.IsValid() {
        return &ValidationResult{
            Errors:   result.Errors,
            Warnings: result.Warnings,
        }
    }
    
    return nil
}

func (c *Config) validateDirectories(result *ValidationResult) {
    // Validate stow directory
    if c.Directories.Stow == "" {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "directories.stow",
            Value:   c.Directories.Stow,
            Message: "stow directory cannot be empty",
        })
    }
    
    // Validate target directory
    if c.Directories.Target == "" {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "directories.target",
            Value:   c.Directories.Target,
            Message: "target directory cannot be empty",
        })
    }
    
    // Check for path traversal in paths
    if strings.Contains(c.Directories.Stow, "..") {
        result.Warnings = append(result.Warnings, ValidationError{
            Field:   "directories.stow",
            Value:   c.Directories.Stow,
            Message: "path contains '..' which may be unexpected",
        })
    }
}

func (c *Config) validateLogging(result *ValidationResult) {
    validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
    if !contains(validLevels, c.Logging.Level) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "logging.level",
            Value:   c.Logging.Level,
            Message: fmt.Sprintf("invalid log level (must be one of: %s)", strings.Join(validLevels, ", ")),
        })
    }
    
    validFormats := []string{"text", "json"}
    if !contains(validFormats, c.Logging.Format) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "logging.format",
            Value:   c.Logging.Format,
            Message: fmt.Sprintf("invalid log format (must be one of: %s)", strings.Join(validFormats, ", ")),
        })
    }
    
    validDestinations := []string{"stderr", "stdout", "file"}
    if !contains(validDestinations, c.Logging.Destination) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "logging.destination",
            Value:   c.Logging.Destination,
            Message: fmt.Sprintf("invalid log destination (must be one of: %s)", strings.Join(validDestinations, ", ")),
        })
    }
    
    if c.Logging.Destination == "file" && c.Logging.File == "" {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "logging.file",
            Value:   c.Logging.File,
            Message: "log file must be specified when destination is 'file'",
        })
    }
}

func (c *Config) validateSymlinks(result *ValidationResult) {
    validModes := []string{"relative", "absolute"}
    if !contains(validModes, c.Symlinks.Mode) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "symlinks.mode",
            Value:   c.Symlinks.Mode,
            Message: fmt.Sprintf("invalid symlink mode (must be one of: %s)", strings.Join(validModes, ", ")),
        })
    }
    
    if c.Symlinks.Backup && c.Symlinks.BackupSuffix == "" {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "symlinks.backup_suffix",
            Value:   c.Symlinks.BackupSuffix,
            Message: "backup suffix cannot be empty when backup is enabled",
        })
    }
}

func (c *Config) validateIgnore(result *ValidationResult) {
    // Validate ignore patterns are valid globs
    for i, pattern := range c.Ignore.Patterns {
        if _, err := filepath.Match(pattern, "test"); err != nil {
            result.Errors = append(result.Errors, ValidationError{
                Field:   fmt.Sprintf("ignore.patterns[%d]", i),
                Value:   pattern,
                Message: fmt.Sprintf("invalid glob pattern: %v", err),
            })
        }
    }
    
    // Validate override patterns
    for i, pattern := range c.Ignore.Overrides {
        if _, err := filepath.Match(pattern, "test"); err != nil {
            result.Errors = append(result.Errors, ValidationError{
                Field:   fmt.Sprintf("ignore.overrides[%d]", i),
                Value:   pattern,
                Message: fmt.Sprintf("invalid glob pattern: %v", err),
            })
        }
    }
}

func (c *Config) validateDotfile(result *ValidationResult) {
    if c.Dotfile.Prefix == "" {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "dotfile.prefix",
            Value:   c.Dotfile.Prefix,
            Message: "dotfile prefix cannot be empty when translate is enabled",
        })
    }
}

func (c *Config) validateOutput(result *ValidationResult) {
    validFormats := []string{"text", "json", "yaml", "table"}
    if !contains(validFormats, c.Output.Format) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "output.format",
            Value:   c.Output.Format,
            Message: fmt.Sprintf("invalid output format (must be one of: %s)", strings.Join(validFormats, ", ")),
        })
    }
    
    validColors := []string{"auto", "always", "never"}
    if !contains(validColors, c.Output.Color) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "output.color",
            Value:   c.Output.Color,
            Message: fmt.Sprintf("invalid color mode (must be one of: %s)", strings.Join(validColors, ", ")),
        })
    }
    
    if c.Output.Verbosity < 0 || c.Output.Verbosity > 3 {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "output.verbosity",
            Value:   c.Output.Verbosity,
            Message: "verbosity must be between 0 and 3",
        })
    }
    
    if c.Output.Width < 0 {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "output.width",
            Value:   c.Output.Width,
            Message: "width cannot be negative (use 0 for auto-detect)",
        })
    }
}

func (c *Config) validateOperations(result *ValidationResult) {
    if c.Operations.MaxParallel < 0 {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "operations.max_parallel",
            Value:   c.Operations.MaxParallel,
            Message: "max_parallel cannot be negative (use 0 for auto-detect)",
        })
    }
}

func (c *Config) validatePackages(result *ValidationResult) {
    validSortBy := []string{"name", "links", "date"}
    if !contains(validSortBy, c.Packages.SortBy) {
        result.Errors = append(result.Errors, ValidationError{
            Field:   "packages.sort_by",
            Value:   c.Packages.SortBy,
            Message: fmt.Sprintf("invalid sort field (must be one of: %s)", strings.Join(validSortBy, ", ")),
        })
    }
}

func (c *Config) validateDoctor(result *ValidationResult) {
    // Doctor configuration is all booleans, no validation needed
}

func contains(slice []string, value string) bool {
    for _, item := range slice {
        if item == value {
            return true
        }
    }
    return false
}
```

**Tasks**:
- [ ] Define ValidationError and ValidationResult types
- [ ] Implement Validate() method
- [ ] Add validation for each configuration section
- [ ] Include field-specific validation
- [ ] Support warnings in addition to errors
- [ ] Write comprehensive validation tests

**Test Cases**:
- Valid configuration passes
- Invalid log level fails
- Invalid symlink mode fails
- Invalid glob patterns fail
- Empty required fields fail
- Out-of-range numeric values fail
- Warnings for suspicious values

### Task 15b.3: Configuration Loader

**Goal**: Load and merge configuration from multiple sources

#### 15b.3.1: Loader Implementation

**File**: `internal/config/loader.go`

**Implementation**:
```go
// Loader handles loading configuration from multiple sources.
type Loader struct {
    appName string
}

// NewLoader creates a configuration loader.
func NewLoader(appName string) *Loader {
    return &Loader{
        appName: appName,
    }
}

// Load loads configuration from all sources with proper precedence.
func (l *Loader) Load() (*Config, error) {
    // Start with defaults
    cfg := Default()
    
    // Load from config file
    configPath := l.getConfigFilePath()
    if fileExists(configPath) {
        fileCfg, err := l.loadFromFile(configPath)
        if err != nil {
            return nil, fmt.Errorf("load config file: %w", err)
        }
        cfg = mergeConfigs(cfg, fileCfg)
    }
    
    // Load from environment
    envCfg := l.loadFromEnv()
    cfg = mergeConfigs(cfg, envCfg)
    
    // Validate merged configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return cfg, nil
}

// LoadWithFlags loads configuration and applies flag overrides.
func (l *Loader) LoadWithFlags(flags map[string]interface{}) (*Config, error) {
    cfg, err := l.Load()
    if err != nil {
        return nil, err
    }
    
    // Apply flag overrides
    flagCfg := l.configFromFlags(flags)
    cfg = mergeConfigs(cfg, flagCfg)
    
    // Validate again after flag overrides
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return cfg, nil
}

// loadFromFile loads configuration from a file.
func (l *Loader) loadFromFile(path string) (*Config, error) {
    v := viper.New()
    v.SetConfigFile(path)
    
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("read config file: %w", err)
    }
    
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }
    
    return &cfg, nil
}

// loadFromEnv loads configuration from environment variables.
func (l *Loader) loadFromEnv() *Config {
    v := viper.New()
    
    // Set up environment variable handling
    v.SetEnvPrefix(strings.ToUpper(l.appName))
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()
    
    // Bind all configuration keys
    l.bindEnvKeys(v)
    
    var cfg Config
    v.Unmarshal(&cfg)
    
    return &cfg
}

// bindEnvKeys binds all configuration keys to environment variables.
func (l *Loader) bindEnvKeys(v *viper.Viper) {
    // Bind all keys from Config struct
    v.BindEnv("directories.stow")
    v.BindEnv("directories.target")
    v.BindEnv("directories.manifest")
    
    v.BindEnv("logging.level")
    v.BindEnv("logging.format")
    v.BindEnv("logging.destination")
    v.BindEnv("logging.file")
    
    v.BindEnv("symlinks.mode")
    v.BindEnv("symlinks.folding")
    v.BindEnv("symlinks.overwrite")
    v.BindEnv("symlinks.backup")
    v.BindEnv("symlinks.backup_suffix")
    
    v.BindEnv("ignore.use_defaults")
    v.BindEnv("ignore.patterns")
    v.BindEnv("ignore.overrides")
    
    v.BindEnv("dotfile.translate")
    v.BindEnv("dotfile.prefix")
    
    v.BindEnv("output.format")
    v.BindEnv("output.color")
    v.BindEnv("output.progress")
    v.BindEnv("output.verbosity")
    v.BindEnv("output.width")
    
    v.BindEnv("operations.dry_run")
    v.BindEnv("operations.atomic")
    v.BindEnv("operations.max_parallel")
    
    v.BindEnv("packages.sort_by")
    v.BindEnv("packages.auto_discover")
    v.BindEnv("packages.validate_names")
    
    v.BindEnv("doctor.auto_fix")
    v.BindEnv("doctor.check_manifest")
    v.BindEnv("doctor.check_broken_links")
    v.BindEnv("doctor.check_orphaned")
    v.BindEnv("doctor.check_permissions")
    
    v.BindEnv("experimental.parallel")
    v.BindEnv("experimental.profiling")
}

// configFromFlags creates partial config from flag map.
func (l *Loader) configFromFlags(flags map[string]interface{}) *Config {
    cfg := &Config{}
    
    // Map flags to config fields
    if val, ok := flags["dir"].(string); ok && val != "" {
        cfg.Directories.Stow = val
    }
    if val, ok := flags["target"].(string); ok && val != "" {
        cfg.Directories.Target = val
    }
    if val, ok := flags["dry-run"].(bool); ok {
        cfg.Operations.DryRun = val
    }
    if val, ok := flags["verbose"].(int); ok {
        cfg.Output.Verbosity = val
    }
    if val, ok := flags["quiet"].(bool); ok && val {
        cfg.Output.Verbosity = 0
    }
    if val, ok := flags["log-json"].(bool); ok && val {
        cfg.Logging.Format = "json"
    }
    if val, ok := flags["color"].(string); ok && val != "" {
        cfg.Output.Color = val
    }
    if val, ok := flags["format"].(string); ok && val != "" {
        cfg.Output.Format = val
    }
    
    return cfg
}

// getConfigFilePath returns the configuration file path.
func (l *Loader) getConfigFilePath() string {
    // Check for explicit config file path
    if path := os.Getenv("DOT_CONFIG"); path != "" {
        return path
    }
    
    // Use XDG config directory
    configDir := GetConfigPath(l.appName)
    
    // Try each supported format
    for _, ext := range []string{"yaml", "yml", "json", "toml"} {
        path := filepath.Join(configDir, "config."+ext)
        if fileExists(path) {
            return path
        }
    }
    
    // Default to YAML if none exist
    return filepath.Join(configDir, "config.yaml")
}

// mergeConfigs merges two configs, with override taking precedence.
func mergeConfigs(base, override *Config) *Config {
    merged := *base
    
    // Merge directories
    if override.Directories.Stow != "" {
        merged.Directories.Stow = override.Directories.Stow
    }
    if override.Directories.Target != "" {
        merged.Directories.Target = override.Directories.Target
    }
    if override.Directories.Manifest != "" {
        merged.Directories.Manifest = override.Directories.Manifest
    }
    
    // Merge logging
    if override.Logging.Level != "" {
        merged.Logging.Level = override.Logging.Level
    }
    if override.Logging.Format != "" {
        merged.Logging.Format = override.Logging.Format
    }
    if override.Logging.Destination != "" {
        merged.Logging.Destination = override.Logging.Destination
    }
    if override.Logging.File != "" {
        merged.Logging.File = override.Logging.File
    }
    
    // Merge symlinks
    if override.Symlinks.Mode != "" {
        merged.Symlinks.Mode = override.Symlinks.Mode
    }
    merged.Symlinks.Folding = override.Symlinks.Folding || base.Symlinks.Folding
    merged.Symlinks.Overwrite = override.Symlinks.Overwrite || base.Symlinks.Overwrite
    merged.Symlinks.Backup = override.Symlinks.Backup || base.Symlinks.Backup
    if override.Symlinks.BackupSuffix != "" {
        merged.Symlinks.BackupSuffix = override.Symlinks.BackupSuffix
    }
    
    // Merge ignore
    merged.Ignore.UseDefaults = override.Ignore.UseDefaults || base.Ignore.UseDefaults
    if len(override.Ignore.Patterns) > 0 {
        merged.Ignore.Patterns = append(base.Ignore.Patterns, override.Ignore.Patterns...)
    }
    if len(override.Ignore.Overrides) > 0 {
        merged.Ignore.Overrides = append(base.Ignore.Overrides, override.Ignore.Overrides...)
    }
    
    // Merge dotfile
    merged.Dotfile.Translate = override.Dotfile.Translate || base.Dotfile.Translate
    if override.Dotfile.Prefix != "" {
        merged.Dotfile.Prefix = override.Dotfile.Prefix
    }
    
    // Merge output
    if override.Output.Format != "" {
        merged.Output.Format = override.Output.Format
    }
    if override.Output.Color != "" {
        merged.Output.Color = override.Output.Color
    }
    merged.Output.Progress = override.Output.Progress || base.Output.Progress
    if override.Output.Verbosity != 0 {
        merged.Output.Verbosity = override.Output.Verbosity
    }
    if override.Output.Width != 0 {
        merged.Output.Width = override.Output.Width
    }
    
    // Merge operations
    merged.Operations.DryRun = override.Operations.DryRun || base.Operations.DryRun
    merged.Operations.Atomic = override.Operations.Atomic || base.Operations.Atomic
    if override.Operations.MaxParallel != 0 {
        merged.Operations.MaxParallel = override.Operations.MaxParallel
    }
    
    // Merge packages
    if override.Packages.SortBy != "" {
        merged.Packages.SortBy = override.Packages.SortBy
    }
    merged.Packages.AutoDiscover = override.Packages.AutoDiscover || base.Packages.AutoDiscover
    merged.Packages.ValidateNames = override.Packages.ValidateNames || base.Packages.ValidateNames
    
    // Merge doctor
    merged.Doctor.AutoFix = override.Doctor.AutoFix || base.Doctor.AutoFix
    merged.Doctor.CheckManifest = override.Doctor.CheckManifest || base.Doctor.CheckManifest
    merged.Doctor.CheckBrokenLinks = override.Doctor.CheckBrokenLinks || base.Doctor.CheckBrokenLinks
    merged.Doctor.CheckOrphaned = override.Doctor.CheckOrphaned || base.Doctor.CheckOrphaned
    merged.Doctor.CheckPermissions = override.Doctor.CheckPermissions || base.Doctor.CheckPermissions
    
    // Merge experimental
    merged.Experimental.Parallel = override.Experimental.Parallel || base.Experimental.Parallel
    merged.Experimental.Profiling = override.Experimental.Profiling || base.Experimental.Profiling
    
    return &merged
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}
```

**Tasks**:
- [ ] Implement Loader struct
- [ ] Implement Load() method
- [ ] Implement LoadWithFlags() method
- [ ] Implement file loading with format detection
- [ ] Implement environment variable loading
- [ ] Implement flag-to-config mapping
- [ ] Implement config merging logic
- [ ] Write comprehensive loader tests

**Test Cases**:
- Load from YAML file
- Load from JSON file
- Load from TOML file
- Load from environment variables
- Load with flag overrides
- Merge multiple sources correctly
- Precedence order is correct
- Handle missing config file gracefully
- Handle invalid config file

### Task 15b.4: Configuration Writer

**Goal**: Write configuration to file with proper formatting

#### 15b.4.1: Writer Implementation

**File**: `internal/config/writer.go`

**Implementation**:
```go
// Writer handles writing configuration to files.
type Writer struct {
    path string
}

// NewWriter creates a configuration writer.
func NewWriter(path string) *Writer {
    return &Writer{
        path: path,
    }
}

// Write writes configuration to file.
func (w *Writer) Write(cfg *Config, opts WriteOptions) error {
    // Ensure directory exists
    dir := filepath.Dir(w.path)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return fmt.Errorf("create config directory: %w", err)
    }
    
    // Marshal config based on format
    data, err := w.marshal(cfg, opts)
    if err != nil {
        return fmt.Errorf("marshal config: %w", err)
    }
    
    // Write to file with secure permissions
    if err := os.WriteFile(w.path, data, 0600); err != nil {
        return fmt.Errorf("write config file: %w", err)
    }
    
    return nil
}

// WriteDefault writes default configuration with comments.
func (w *Writer) WriteDefault(opts WriteOptions) error {
    cfg := Default()
    opts.IncludeComments = true
    return w.Write(cfg, opts)
}

// Update updates specific value in configuration file.
func (w *Writer) Update(key string, value interface{}) error {
    // Load existing config
    cfg, err := LoadFromFile(w.path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            // File doesn't exist, create with default
            cfg = Default()
        } else {
            return fmt.Errorf("load existing config: %w", err)
        }
    }
    
    // Update value
    if err := w.setValue(cfg, key, value); err != nil {
        return fmt.Errorf("set value: %w", err)
    }
    
    // Validate
    if err := cfg.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Write back
    opts := WriteOptions{
        Format:          w.detectFormat(),
        IncludeComments: false,
    }
    return w.Write(cfg, opts)
}

// WriteOptions controls configuration file output.
type WriteOptions struct {
    Format          string // yaml, json, toml
    IncludeComments bool
    Indent          int
}

// marshal converts config to bytes in specified format.
func (w *Writer) marshal(cfg *Config, opts WriteOptions) ([]byte, error) {
    format := opts.Format
    if format == "" {
        format = w.detectFormat()
    }
    
    switch format {
    case "yaml", "yml":
        return w.marshalYAML(cfg, opts)
    case "json":
        return w.marshalJSON(cfg, opts)
    case "toml":
        return w.marshalTOML(cfg, opts)
    default:
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
}

// marshalYAML marshals config to YAML.
func (w *Writer) marshalYAML(cfg *Config, opts WriteOptions) ([]byte, error) {
    var buf bytes.Buffer
    
    if opts.IncludeComments {
        w.writeYAMLWithComments(&buf, cfg)
    } else {
        data, err := yaml.Marshal(cfg)
        if err != nil {
            return nil, err
        }
        buf.Write(data)
    }
    
    return buf.Bytes(), nil
}

// marshalJSON marshals config to JSON.
func (w *Writer) marshalJSON(cfg *Config, opts WriteOptions) ([]byte, error) {
    indent := opts.Indent
    if indent == 0 {
        indent = 2
    }
    
    var buf bytes.Buffer
    encoder := json.NewEncoder(&buf)
    encoder.SetIndent("", strings.Repeat(" ", indent))
    
    if err := encoder.Encode(cfg); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

// marshalTOML marshals config to TOML.
func (w *Writer) marshalTOML(cfg *Config, opts WriteOptions) ([]byte, error) {
    var buf bytes.Buffer
    encoder := toml.NewEncoder(&buf)
    
    if err := encoder.Encode(cfg); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

// writeYAMLWithComments writes YAML with helpful comments.
func (w *Writer) writeYAMLWithComments(buf *bytes.Buffer, cfg *Config) {
    buf.WriteString("# Dot Configuration File\n")
    buf.WriteString("# Location: " + w.path + "\n")
    buf.WriteString("# Documentation: https://github.com/user/dot/docs/configuration.md\n\n")
    
    buf.WriteString("# Core Directories\n")
    buf.WriteString("directories:\n")
    buf.WriteString("  # Stow directory containing packages\n")
    buf.WriteString(fmt.Sprintf("  stow: %s\n", cfg.Directories.Stow))
    buf.WriteString("  # Target directory for symlinks\n")
    buf.WriteString(fmt.Sprintf("  target: %s\n", cfg.Directories.Target))
    buf.WriteString("  # Manifest directory for tracking\n")
    buf.WriteString(fmt.Sprintf("  manifest: %s\n\n", cfg.Directories.Manifest))
    
    buf.WriteString("# Logging Configuration\n")
    buf.WriteString("logging:\n")
    buf.WriteString("  # Log level: DEBUG, INFO, WARN, ERROR\n")
    buf.WriteString(fmt.Sprintf("  level: %s\n", cfg.Logging.Level))
    buf.WriteString("  # Log format: text, json\n")
    buf.WriteString(fmt.Sprintf("  format: %s\n", cfg.Logging.Format))
    buf.WriteString("  # Log destination: stderr, stdout, file\n")
    buf.WriteString(fmt.Sprintf("  destination: %s\n", cfg.Logging.Destination))
    buf.WriteString("  # Log file path (only used if destination is file)\n")
    buf.WriteString(fmt.Sprintf("  file: %s\n\n", cfg.Logging.File))
    
    buf.WriteString("# Symlink Behavior\n")
    buf.WriteString("symlinks:\n")
    buf.WriteString("  # Link mode: relative, absolute\n")
    buf.WriteString(fmt.Sprintf("  mode: %s\n", cfg.Symlinks.Mode))
    buf.WriteString("  # Enable directory folding optimization\n")
    buf.WriteString(fmt.Sprintf("  folding: %t\n", cfg.Symlinks.Folding))
    buf.WriteString("  # Overwrite existing files when conflicts occur\n")
    buf.WriteString(fmt.Sprintf("  overwrite: %t\n", cfg.Symlinks.Overwrite))
    buf.WriteString("  # Create backup of overwritten files\n")
    buf.WriteString(fmt.Sprintf("  backup: %t\n", cfg.Symlinks.Backup))
    buf.WriteString("  # Backup suffix when backups enabled\n")
    buf.WriteString(fmt.Sprintf("  backup_suffix: %s\n\n", cfg.Symlinks.BackupSuffix))
    
    buf.WriteString("# Ignore Patterns\n")
    buf.WriteString("ignore:\n")
    buf.WriteString("  # Use default ignore patterns\n")
    buf.WriteString(fmt.Sprintf("  use_defaults: %t\n", cfg.Ignore.UseDefaults))
    buf.WriteString("  # Additional patterns to ignore (glob format)\n")
    w.writeYAMLList(buf, "patterns", cfg.Ignore.Patterns, 2)
    buf.WriteString("  # Patterns to override (force include even if ignored)\n")
    w.writeYAMLList(buf, "overrides", cfg.Ignore.Overrides, 2)
    buf.WriteString("\n")
    
    buf.WriteString("# Dotfile Translation\n")
    buf.WriteString("dotfile:\n")
    buf.WriteString("  # Enable dot- to . translation\n")
    buf.WriteString(fmt.Sprintf("  translate: %t\n", cfg.Dotfile.Translate))
    buf.WriteString("  # Prefix for dotfile translation\n")
    buf.WriteString(fmt.Sprintf("  prefix: %s\n\n", cfg.Dotfile.Prefix))
    
    buf.WriteString("# Output Configuration\n")
    buf.WriteString("output:\n")
    buf.WriteString("  # Default output format: text, json, yaml, table\n")
    buf.WriteString(fmt.Sprintf("  format: %s\n", cfg.Output.Format))
    buf.WriteString("  # Enable colored output: auto, always, never\n")
    buf.WriteString(fmt.Sprintf("  color: %s\n", cfg.Output.Color))
    buf.WriteString("  # Show progress indicators\n")
    buf.WriteString(fmt.Sprintf("  progress: %t\n", cfg.Output.Progress))
    buf.WriteString("  # Verbosity level: 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)\n")
    buf.WriteString(fmt.Sprintf("  verbosity: %d\n", cfg.Output.Verbosity))
    buf.WriteString("  # Terminal width for text wrapping (0 = auto-detect)\n")
    buf.WriteString(fmt.Sprintf("  width: %d\n\n", cfg.Output.Width))
    
    buf.WriteString("# Operation Defaults\n")
    buf.WriteString("operations:\n")
    buf.WriteString("  # Enable dry-run mode by default\n")
    buf.WriteString(fmt.Sprintf("  dry_run: %t\n", cfg.Operations.DryRun))
    buf.WriteString("  # Enable atomic operations with rollback\n")
    buf.WriteString(fmt.Sprintf("  atomic: %t\n", cfg.Operations.Atomic))
    buf.WriteString("  # Maximum number of parallel operations (0 = auto)\n")
    buf.WriteString(fmt.Sprintf("  max_parallel: %d\n\n", cfg.Operations.MaxParallel))
    
    buf.WriteString("# Package Management\n")
    buf.WriteString("packages:\n")
    buf.WriteString("  # Default sort order: name, links, date\n")
    buf.WriteString(fmt.Sprintf("  sort_by: %s\n", cfg.Packages.SortBy))
    buf.WriteString("  # Automatically scan for new packages\n")
    buf.WriteString(fmt.Sprintf("  auto_discover: %t\n", cfg.Packages.AutoDiscover))
    buf.WriteString("  # Package naming convention validation\n")
    buf.WriteString(fmt.Sprintf("  validate_names: %t\n\n", cfg.Packages.ValidateNames))
    
    buf.WriteString("# Doctor Configuration\n")
    buf.WriteString("doctor:\n")
    buf.WriteString("  # Auto-fix issues when possible\n")
    buf.WriteString(fmt.Sprintf("  auto_fix: %t\n", cfg.Doctor.AutoFix))
    buf.WriteString("  # Check manifest integrity\n")
    buf.WriteString(fmt.Sprintf("  check_manifest: %t\n", cfg.Doctor.CheckManifest))
    buf.WriteString("  # Check for broken symlinks\n")
    buf.WriteString(fmt.Sprintf("  check_broken_links: %t\n", cfg.Doctor.CheckBrokenLinks))
    buf.WriteString("  # Check for orphaned links\n")
    buf.WriteString(fmt.Sprintf("  check_orphaned: %t\n", cfg.Doctor.CheckOrphaned))
    buf.WriteString("  # Check file permissions\n")
    buf.WriteString(fmt.Sprintf("  check_permissions: %t\n\n", cfg.Doctor.CheckPermissions))
    
    buf.WriteString("# Experimental Features\n")
    buf.WriteString("experimental:\n")
    buf.WriteString("  # Enable parallel operations\n")
    buf.WriteString(fmt.Sprintf("  parallel: %t\n", cfg.Experimental.Parallel))
    buf.WriteString("  # Enable performance profiling\n")
    buf.WriteString(fmt.Sprintf("  profiling: %t\n", cfg.Experimental.Profiling))
}

// writeYAMLList writes a YAML list with proper indentation.
func (w *Writer) writeYAMLList(buf *bytes.Buffer, key string, items []string, indent int) {
    prefix := strings.Repeat(" ", indent)
    
    if len(items) == 0 {
        buf.WriteString(fmt.Sprintf("%s%s: []\n", prefix, key))
        return
    }
    
    buf.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
    for _, item := range items {
        buf.WriteString(fmt.Sprintf("%s  - \"%s\"\n", prefix, item))
    }
}

// detectFormat detects format from file extension.
func (w *Writer) detectFormat() string {
    ext := filepath.Ext(w.path)
    switch ext {
    case ".yaml", ".yml":
        return "yaml"
    case ".json":
        return "json"
    case ".toml":
        return "toml"
    default:
        return "yaml"
    }
}

// setValue sets a configuration value by dotted key path.
func (w *Writer) setValue(cfg *Config, key string, value interface{}) error {
    parts := strings.Split(key, ".")
    if len(parts) < 2 {
        return fmt.Errorf("invalid key: %s (must be section.field)", key)
    }
    
    section := parts[0]
    field := parts[1]
    
    switch section {
    case "directories":
        return w.setDirectoriesValue(cfg, field, value)
    case "logging":
        return w.setLoggingValue(cfg, field, value)
    case "symlinks":
        return w.setSymlinksValue(cfg, field, value)
    case "ignore":
        return w.setIgnoreValue(cfg, field, value)
    case "dotfile":
        return w.setDotfileValue(cfg, field, value)
    case "output":
        return w.setOutputValue(cfg, field, value)
    case "operations":
        return w.setOperationsValue(cfg, field, value)
    case "packages":
        return w.setPackagesValue(cfg, field, value)
    case "doctor":
        return w.setDoctorValue(cfg, field, value)
    case "experimental":
        return w.setExperimentalValue(cfg, field, value)
    default:
        return fmt.Errorf("unknown section: %s", section)
    }
}

// Section-specific setValue methods omitted for brevity
// Each would handle type conversion and field assignment
```

**Tasks**:
- [ ] Implement Writer struct
- [ ] Implement Write() method
- [ ] Implement WriteDefault() method
- [ ] Implement Update() method
- [ ] Implement marshaling for YAML/JSON/TOML
- [ ] Implement commented YAML generation
- [ ] Implement setValue() for updating specific keys
- [ ] Ensure secure file permissions (0600)
- [ ] Write comprehensive writer tests

**Test Cases**:
- Write to YAML file
- Write to JSON file
- Write to TOML file
- Write with comments
- Write without comments
- Update existing value
- Create new value
- File permissions are correct
- Invalid key path fails
- Invalid value type fails

### Task 15b.5: Config Command Implementation

**Goal**: Implement `dot config` command with all subcommands

#### 15b.5.1: Base Config Command

**File**: `cmd/dot/config.go`

**Implementation**:
```go
// newConfigCommand creates the config command.
func newConfigCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage dot configuration",
        Long: `View and modify dot configuration settings.

Configuration is loaded from multiple sources in order of precedence:
  1. Command-line flags (highest)
  2. Environment variables (DOT_* prefix)
  3. Configuration file (~/.config/dot/config.yaml)
  4. Built-in defaults (lowest)

The config command allows viewing current settings, modifying configuration
files, and managing configuration across sources.`,
        Example: `  # Show current configuration
  dot config

  # Initialize configuration file
  dot config init

  # Get specific value
  dot config get directories.stow

  # Set configuration value
  dot config set directories.stow ~/dotfiles

  # Edit configuration in editor
  dot config edit

  # Validate configuration
  dot config validate`,
        RunE: runConfigList,
    }

    cmd.AddCommand(
        newConfigInitCommand(),
        newConfigGetCommand(),
        newConfigSetCommand(),
        newConfigUnsetCommand(),
        newConfigListCommand(),
        newConfigEditCommand(),
        newConfigPathCommand(),
        newConfigValidateCommand(),
        newConfigResetCommand(),
    )

    return cmd
}

// runConfigList is the default action (list config).
func runConfigList(cmd *cobra.Command, args []string) error {
    // Delegate to list command
    return runConfigListCmd(cmd, args)
}
```

**Tasks**:
- [ ] Create newConfigCommand() function
- [ ] Add command metadata and examples
- [ ] Add subcommands
- [ ] Implement default behavior (list)
- [ ] Write base command tests

#### 15b.5.2: Init Subcommand

**File**: `cmd/dot/config_init.go`

**Implementation** (partial - see full spec for complete implementation):
```go
func newConfigInitCommand() *cobra.Command {
    var (
        interactive bool
        force       bool
        noComments  bool
        format      string
    )

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Create initial configuration file",
        Long: `Create a new configuration file with default values.

The configuration file is created in the XDG config directory:
  ~/.config/dot/config.yaml (default)

Use --interactive to be prompted for key settings.
Use --force to overwrite existing configuration.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runConfigInit(interactive, force, noComments, format)
        },
    }

    cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Prompt for settings")
    cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing config")
    cmd.Flags().BoolVar(&noComments, "no-comments", false, "Omit comments in config file")
    cmd.Flags().StringVar(&format, "format", "yaml", "Config format (yaml, json, toml)")

    return cmd
}

func runConfigInit(interactive, force, noComments bool, format string) error {
    // Implementation
    configPath := getConfigFilePath(format)
    
    // Check if exists
    if fileExists(configPath) && !force {
        return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
    }
    
    writer := config.NewWriter(configPath)
    
    if interactive {
        cfg, err := promptForConfig()
        if err != nil {
            return err
        }
        return writer.Write(cfg, config.WriteOptions{
            Format:          format,
            IncludeComments: !noComments,
        })
    }
    
    return writer.WriteDefault(config.WriteOptions{
        Format:          format,
        IncludeComments: !noComments,
    })
}
```

**Tasks**:
- [ ] Implement newConfigInitCommand()
- [ ] Implement runConfigInit()
- [ ] Implement interactive prompting
- [ ] Handle existing config file
- [ ] Create config directory if needed
- [ ] Set secure file permissions
- [ ] Write init command tests

**Test Cases**:
- Create new config file
- Create with interactive prompts
- Create without comments
- Create in JSON format
- Create in TOML format
- Fail if exists (without --force)
- Overwrite with --force
- Create config directory

#### 15b.5.3-15b.5.9: Other Subcommands

Similar implementation for:
- `get`: Retrieve config values
- `set`: Set config values
- `unset`: Remove config values
- `list`: List all config
- `edit`: Open in editor
- `path`: Show config paths
- `validate`: Validate config
- `reset`: Reset to defaults

(Full implementation details in specification above)

### Task 15b.6: Integration and Testing

**Goal**: Integrate configuration system with existing CLI and comprehensive testing

#### 15b.6.1: Update Root Command

**File**: `cmd/dot/root.go`

**Tasks**:
- [ ] Update buildConfig() to use configuration loader
- [ ] Integrate config file loading
- [ ] Maintain backward compatibility with flags
- [ ] Update createLogger() to use config
- [ ] Update globalConfig mapping

#### 15b.6.2: Update Commands

Update all commands to use configuration:

**Files**: `cmd/dot/manage.go`, `cmd/dot/unmanage.go`, etc.

**Tasks**:
- [ ] Use config for default values
- [ ] Respect configuration settings
- [ ] Allow flag overrides
- [ ] Update command tests

#### 15b.6.3: Integration Tests

**File**: `cmd/dot/config_integration_test.go`

**Test Scenarios**:
- [ ] Config precedence: flags > env > file > defaults
- [ ] Create and load config file
- [ ] Modify config with set command
- [ ] Validate config file
- [ ] Edit config in editor (mock)
- [ ] Reset config to defaults
- [ ] Environment variable overrides
- [ ] Flag overrides

#### 15b.6.4: Unit Tests

**Files**: `internal/config/*_test.go`

**Test Coverage**:
- [ ] Config struct validation
- [ ] Configuration loading from each format
- [ ] Environment variable parsing
- [ ] Flag mapping
- [ ] Config merging logic
- [ ] Config writing with comments
- [ ] Value update logic
- [ ] Validation error messages

**Coverage Target**: ≥ 80% for all config package code

## Documentation

### User Documentation

**File**: `docs/configuration.md`

**Contents**:
- Configuration file structure
- All configuration options with descriptions
- Environment variable mapping
- Precedence rules
- Examples for common scenarios
- Migration guide for existing users

### CLI Help

**Files**: `cmd/dot/config_*.go`

**Requirements**:
- Comprehensive help text for each subcommand
- Examples for common operations
- Cross-references to related commands
- Links to full documentation

### Developer Documentation

**File**: `docs/dev/configuration.md`

**Contents**:
- Configuration architecture
- Adding new configuration options
- Testing configuration changes
- Configuration schema evolution

## Acceptance Criteria

Phase 15b is complete when:

### Functionality
- [ ] Configuration file can be created with `dot config init`
- [ ] Configuration values can be retrieved with `dot config get`
- [ ] Configuration values can be set with `dot config set`
- [ ] Configuration can be edited interactively with `dot config edit`
- [ ] Configuration can be validated with `dot config validate`
- [ ] Configuration paths displayed with `dot config path`
- [ ] Configuration can be reset with `dot config reset`
- [ ] All configuration options are exposed
- [ ] Environment variables work for all options
- [ ] Flags override configuration correctly
- [ ] Multiple file formats supported (YAML, JSON, TOML)

### Quality
- [ ] Unit test coverage ≥ 80%
- [ ] Integration tests cover all subcommands
- [ ] All validation rules have tests
- [ ] Configuration precedence tested
- [ ] No linter warnings
- [ ] All tests pass

### Documentation
- [ ] Configuration reference complete
- [ ] User guide with examples
- [ ] Migration guide for existing users
- [ ] Developer documentation complete
- [ ] All commands have help text

### Security
- [ ] Config files created with 0600 permissions
- [ ] Path validation prevents traversal
- [ ] No sensitive data logged
- [ ] Input validation on all set operations
- [ ] XDG directories used correctly

### Compatibility
- [ ] Existing flags continue to work
- [ ] No breaking changes to CLI
- [ ] Backward compatible with no config file
- [ ] Cross-platform support (Linux, macOS, Windows)

## Rollout Plan

### Phase 1: Configuration Schema (Tasks 15b.1-15b.2)
1. Extend Config struct
2. Implement validation
3. Test with existing code

### Phase 2: Configuration Loading (Task 15b.3)
1. Implement loader
2. Integrate with root command
3. Test precedence rules

### Phase 3: Configuration Writing (Task 15b.4)
1. Implement writer
2. Support all formats
3. Test file operations

### Phase 4: Config Command (Task 15b.5)
1. Implement init subcommand
2. Implement get/set/unset subcommands
3. Implement edit/path/validate subcommands
4. Test all subcommands

### Phase 5: Integration (Task 15b.6)
1. Update all commands
2. Run integration tests
3. Fix any issues
4. Verify backward compatibility

### Phase 6: Documentation
1. Write configuration reference
2. Write user guide
3. Write developer documentation
4. Update README

## Risks and Mitigations

### Risk: Configuration Complexity
**Impact**: Users find configuration confusing
**Mitigation**: 
- Provide sensible defaults
- Interactive init wizard
- Clear documentation
- Examples for common scenarios

### Risk: Breaking Changes
**Impact**: Existing users' workflows break
**Mitigation**:
- Maintain flag compatibility
- Work without config file
- Migration guide
- Clear changelog

### Risk: File Format Issues
**Impact**: Config file parsing errors
**Mitigation**:
- Validate on write
- Clear error messages
- Support multiple formats
- Include format examples

### Risk: Security Issues
**Impact**: Config files expose sensitive data
**Mitigation**:
- Secure file permissions
- Path validation
- No secrets in config
- Audit logging

## Success Metrics

### Quantitative
- Configuration file adoption: ≥ 30% of users
- Time to configure: < 2 minutes for common scenarios
- Configuration errors: < 5% of config operations
- Test coverage: ≥ 80%

### Qualitative
- Users find configuration intuitive
- Documentation is comprehensive
- Reduces need for flag repetition
- Enables team configuration sharing

## Timeline Estimate

Total: 30-36 hours

Breakdown:
- 15b.1 Configuration Schema: 4-5 hours
- 15b.2 Validation: 3-4 hours
- 15b.3 Loader: 5-6 hours
- 15b.4 Writer: 5-6 hours
- 15b.5 Config Command: 8-10 hours
- 15b.6 Integration & Testing: 3-4 hours
- Documentation: 2-3 hours

## Deliverables

1. Extended configuration schema with all options
2. Configuration validation system
3. Configuration loader with precedence
4. Configuration writer with multiple formats
5. Complete `config` command with subcommands
6. Integration with all existing commands
7. Comprehensive test suite (≥ 80% coverage)
8. User documentation and reference
9. Developer documentation
10. Migration guide

## Next Steps

After Phase 15b completion:
- Consider configuration profiles for different environments
- Add configuration import/export functionality
- Support configuration inheritance
- Add configuration templates for common setups
