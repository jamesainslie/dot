# Configuration Guide

## Overview

dot supports configuration through XDG-compliant configuration files, environment variables, and command-line flags with clear precedence rules.

## Configuration Precedence

Configuration is loaded from multiple sources in the following order (highest to lowest priority):

1. **Command-line flags** (highest priority)
2. **Environment variables** (`DOT_*` prefix)
3. **Configuration file** (`~/.config/dot/config.yaml`)
4. **Built-in defaults** (lowest priority)

## Configuration File Location

By default, dot stores configuration in:

```
~/.config/dot/config.yaml
```

This follows the XDG Base Directory Specification:
- Uses `$XDG_CONFIG_HOME/dot/config.yaml` if `XDG_CONFIG_HOME` is set
- Falls back to `~/.config/dot/config.yaml` otherwise

You can override the configuration file path with the `DOT_CONFIG` environment variable:

```bash
export DOT_CONFIG=/custom/path/dot-config.yaml
```

## Creating Configuration

### Initialize Default Configuration

```bash
# Create configuration file with defaults and comments
dot config init

# Force overwrite existing configuration
dot config init --force

# Create in different format
dot config init --format json
dot config init --format toml
```

## Managing Configuration

### View Configuration

```bash
# List all settings
dot config list

# Get specific value
dot config get directories.stow
dot config get logging.level
```

### Modify Configuration

```bash
# Set configuration value
dot config set directories.stow ~/dotfiles
dot config set logging.level DEBUG
dot config set symlinks.mode absolute
dot config set output.verbosity 2
```

### Configuration File Path

```bash
# Show config file location
dot config path
```

## Configuration Schema

### Directories

Controls directory paths used by dot:

```yaml
directories:
  stow: ~/dotfiles              # Package source directory
  target: ~                     # Symlink target directory  
  manifest: ~/.local/share/dot/manifest  # State tracking directory
```

### Logging

Controls logging behavior:

```yaml
logging:
  level: INFO                   # DEBUG, INFO, WARN, ERROR
  format: text                  # text, json
  destination: stderr           # stderr, stdout, file
  file: ~/.local/state/dot/dot.log  # Log file (if destination=file)
```

### Symlinks

Controls symlink creation behavior:

```yaml
symlinks:
  mode: relative                # relative, absolute
  folding: true                 # Enable directory folding
  overwrite: false              # Overwrite existing files
  backup: false                 # Backup before overwrite
  backup_suffix: .bak           # Backup file suffix
```

### Ignore Patterns

Controls which files are ignored:

```yaml
ignore:
  use_defaults: true            # Use built-in ignore patterns
  patterns:                     # Additional patterns (glob format)
    - "*.swp"
    - "*.tmp"
    - "*~"
  overrides: []                 # Force include patterns
```

Default ignore patterns:
- `.git`, `.svn`, `.hg` (version control)
- `.DS_Store`, `Thumbs.db`, `desktop.ini` (system files)
- `.Trash`, `.Spotlight-V100`, `.TemporaryItems` (macOS system)

### Dotfile Translation

Controls dot- prefix translation:

```yaml
dotfile:
  translate: true               # Enable translation
  prefix: dot-                  # Prefix to translate
```

When enabled, `dot-vimrc` becomes `.vimrc` in the target directory.

### Output

Controls output formatting and display:

```yaml
output:
  format: text                  # text, json, yaml, table
  color: auto                   # auto, always, never
  progress: true                # Show progress indicators
  verbosity: 1                  # 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)
  width: 0                      # Terminal width (0=auto)
```

### Operations

Controls operation behavior:

```yaml
operations:
  dry_run: false                # Default dry-run mode
  atomic: true                  # Atomic operations with rollback
  max_parallel: 0               # Max parallel ops (0=auto)
```

### Packages

Controls package management:

```yaml
packages:
  sort_by: name                 # name, links, date
  auto_discover: true           # Auto-scan for packages
  validate_names: true          # Validate package names
```

### Doctor

Controls health check behavior:

```yaml
doctor:
  auto_fix: false               # Auto-fix issues
  check_manifest: true          # Check manifest integrity
  check_broken_links: true      # Check for broken symlinks
  check_orphaned: true          # Check for orphaned links
  check_permissions: true       # Check file permissions
```

### Experimental

Experimental feature flags:

```yaml
experimental:
  parallel: false               # Enable parallel operations
  profiling: false              # Enable performance profiling
```

## Environment Variables

All configuration values can be set via environment variables using the `DOT_` prefix and underscores for hierarchy:

```bash
# Examples
export DOT_DIRECTORIES_STOW=/path/to/dotfiles
export DOT_DIRECTORIES_TARGET=/home/user
export DOT_LOGGING_LEVEL=DEBUG
export DOT_LOGGING_FORMAT=json
export DOT_SYMLINKS_MODE=absolute
export DOT_OUTPUT_COLOR=never
export DOT_OUTPUT_VERBOSITY=2
export DOT_OPERATIONS_DRY_RUN=true
```

## File Formats

dot supports three configuration formats:

### YAML (Recommended)

```yaml
directories:
  stow: ~/dotfiles
  target: ~

logging:
  level: INFO
  format: text
```

### JSON

```json
{
  "directories": {
    "stow": "~/dotfiles",
    "target": "~"
  },
  "logging": {
    "level": "INFO",
    "format": "text"
  }
}
```

### TOML

```toml
[directories]
stow = "~/dotfiles"
target = "~"

[logging]
level = "INFO"
format = "text"
```

## Common Scenarios

### Set Custom Dotfiles Location

```bash
dot config set directories.stow ~/my-dotfiles
```

### Enable Debug Logging

```bash
dot config set logging.level DEBUG
dot config set logging.format text
```

### Use Absolute Symlinks

```bash
dot config set symlinks.mode absolute
```

### Disable Directory Folding

```bash
dot config set symlinks.folding false
```

### Add Custom Ignore Patterns

Edit the configuration file and add patterns:

```yaml
ignore:
  use_defaults: true
  patterns:
    - "*.swp"
    - "*.tmp"
    - ".*.un~"
    - "*~"
    - "#*#"
```

## Troubleshooting

### Configuration File Not Found

If `dot config list` reports the file doesn't exist, create it:

```bash
dot config init
```

### Invalid Configuration

Validate your configuration:

```bash
# Check config file syntax
dot config get directories.stow

# If there are errors, reinitialize
dot config init --force
```

### Permission Errors

Ensure the configuration file has correct permissions:

```bash
chmod 600 ~/.config/dot/config.yaml
```

### Finding Configuration Path

```bash
dot config path
```

## See Also

- [Phase-15b Plan](./Phase-15b-Plan.md) - Detailed implementation plan
- [README](../README.md) - General usage guide

