# Advanced Features

Deep dive into advanced dot features and capabilities.

## Ignore Pattern System

### Pattern Types

dot supports multiple pattern types for flexible file exclusion.

#### Glob Patterns

Standard glob syntax:

```yaml
ignore:
  - "*.log"          # All .log files
  - "*.tmp"          # All .tmp files
  - "test_*"         # Files starting with test_
  - "cache/"         # Directory named cache
  - "**/*.bak"       # .bak files in any directory
```

#### Regex Patterns

Regular expressions for complex matching:

```yaml
ignore:
  - "/^test.*\\.go$/"    # Go test files
  - "/.*\\.swp$/"         # Vim swap files
  - "/\\.#.*/"            # Emacs lock files
```

### Pattern Precedence

Patterns evaluated in order with override capability:

```yaml
ignore:
  - "*.log"           # Ignore all logs

override:
  - "important.log"   # But include this one
```

### Performance Optimization

Pattern compilation and caching:

- Patterns compiled once at startup
- Cached for repeated evaluations
- LRU eviction for memory efficiency

## Directory Folding

### Folding Algorithm

Directory folding creates directory-level symlinks when all contents belong to single package.

**Without folding**:
```
~/.vim/colors/theme.vim -> ~/dotfiles/vim/dot-vim/colors/theme.vim
~/.vim/autoload/plugin.vim -> ~/dotfiles/vim/dot-vim/autoload/plugin.vim
~/.vim/ftplugin/go.vim -> ~/dotfiles/vim/dot-vim/ftplugin/go.vim
```

**With folding**:
```
~/.vim/ -> ~/dotfiles/vim/dot-vim/
```

### Folding Rules

1. **Exclusive ownership**: Directory only folded if all files from single package
2. **Mixed ownership**: Falls back to per-file links if multiple packages
3. **Automatic unfolding**: Unfolds when new package adds files to directory
4. **Manual control**: Disable with `--no-folding` flag

### Controlling Folding

```bash
# Disable folding globally
dot --no-folding manage vim

# Disable for specific package
# In package/.dotmeta:
folding: false

# Force per-file granularity
dot --no-folding manage all-packages
```

## Dry-Run Mode

### Usage

Preview operations before applying:

```bash
# Preview any command
dot --dry-run manage vim
dot --dry-run unmanage zsh
dot --dry-run remanage tmux
```

### Output

Detailed plan showing operations:

```
Dry run mode - no changes will be applied

Plan:
  + Create directory: ~/.vim
  + Create symlink: ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
  - Remove symlink: ~/.old-config
  ~ Update symlink: ~/.zshrc
  
Summary:
  Directories: 1
  Symlinks created: 1
  Symlinks removed: 1
  Symlinks updated: 1
  Conflicts: 0
```

### Conflict Detection

Dry-run detects conflicts without modification:

```bash
dot --dry-run manage vim
# Output shows conflicts without creating any symlinks
```

## Resolution Policies

### Available Policies

#### Fail Policy (Default)

Stop on first conflict:

```bash
dot manage vim
# Stops and reports conflict
```

Safest option, requires manual resolution.

#### Backup Policy

Move conflicting files to backup:

```bash
dot --on-conflict backup manage vim
# ~/.vimrc moved to ~/.vimrc.bak
```

Preserves existing files for comparison.

#### Overwrite Policy

Replace conflicting files:

```bash
dot --on-conflict overwrite manage vim
# ~/.vimrc deleted, symlink created
```

Aggressive, use with caution.

#### Skip Policy

Continue past conflicts:

```bash
dot --on-conflict skip manage vim
# Skips ~/.vimrc, creates other links
```

Useful for partial installation.

### Per-Package Policies

Configure different policies per package:

```yaml
# In config.yaml
packages:
  vim:
    onConflict: backup
  
  zsh:
    onConflict: overwrite
```

## State Management

### Manifest Structure

`.dot-manifest.json` in target directory:

```json
{
  "version": "1.0",
  "updated_at": "2025-10-07T10:30:00Z",
  "packages": {
    "vim": {
      "name": "vim",
      "installed_at": "2025-10-07T10:30:00Z",
      "link_count": 3,
      "links": ["~/.vimrc", "~/.vim/"]
    }
  },
  "hashes": {
    "vim": "a3f2c8b4d9e1f0..."
  }
}
```

### Fast Status Queries

Manifest enables instant status without filesystem scanning:

```bash
# Fast - reads manifest only
dot status

# Compare to full scan
dot status --full-scan
```

### State Validation

Check manifest consistency:

```bash
# Validate manifest
dot doctor

# Repair if corrupted
dot doctor --repair
```

## Incremental Operations

### Change Detection

Content-based detection via hashing:

1. Compute hash for each package
2. Compare with stored hash in manifest
3. Skip packages with unchanged hash
4. Process only changed packages

### Efficiency

Incremental operations skip unchanged packages:

```bash
# Only processes changed packages
dot remanage vim zsh tmux git

# Output shows what was skipped:
# vim: changed (processed)
# zsh: unchanged (skipped)
# tmux: unchanged (skipped)
# git: changed (processed)
```

### Forcing Full Processing

Disable incremental detection:

```bash
# Process all packages regardless of changes
dot remanage --no-incremental vim zsh tmux
```

## Parallel Execution

### Concurrent Operations

dot executes independent operations concurrently:

```bash
# Processes multiple packages in parallel
dot manage vim zsh tmux git neovim
```

### Concurrency Control

Configure parallelism:

```yaml
# Limit concurrent operations
concurrency: 4

# Auto-detect (uses CPU cores)
concurrency: 0
```

Or via environment:
```bash
export DOT_CONCURRENCY=4
```

### Dependency-Safe Batching

Operations grouped by dependencies:

- Batch 1: Independent operations (parallel)
- Batch 2: Operations depending on batch 1 (parallel)
- Batch 3: Operations depending on batch 2 (parallel)

## Performance Tuning

### Optimization Strategies

#### 1. Enable Folding

Directory folding reduces symlink count:

```yaml
folding: true
```

#### 2. Use Incremental Updates

Remanage skips unchanged packages:

```bash
dot remanage vim zsh tmux  # Fast
```

#### 3. Tune Concurrency

Match hardware capabilities:

```yaml
concurrency: 8  # For 8-core CPU
```

#### 4. Optimize Ignore Patterns

Fewer patterns = faster scanning:

```yaml
ignore:
  - ".git"       # Essential only
  - "node_modules"
```

### Performance Monitoring

Profile operations:

```bash
# Time operations
time dot manage vim

# Verbose timing
dot -vv manage vim
# Shows timing for each stage
```

## Logging and Output

### Verbosity Levels

Control detail level:

```bash
# Level 0: Errors only
dot manage vim

# Level 1: Info
dot -v manage vim

# Level 2: Debug
dot -vv manage vim

# Level 3: Trace
dot -vvv manage vim
```

### Structured Logging

JSON output for automation:

```bash
# JSON logs
dot --log-json manage vim

# Parse with jq
dot --log-json manage vim 2>&1 | jq '.level'
```

### Quiet Mode

Suppress all output except errors:

```bash
# Script-friendly
dot --quiet manage vim
result=$?
```

## Output Formats

### Multiple Format Support

Commands support various output formats:

```bash
# Human-readable text
dot status

# JSON for scripting
dot status --format json

# YAML for configuration
dot status --format yaml

# Table for structured data
dot status --format table
```

### Format Selection

Based on use case:

- **Text**: Interactive use, human readers
- **JSON**: Scripts, automation, parsing
- **YAML**: Configuration files
- **Table**: Structured comparison

## Next Steps

- [Troubleshooting Guide](08-troubleshooting.md): Solve common issues
- [Glossary](09-glossary.md): Reference for terms
- [Configuration Reference](04-configuration.md): Complete configuration options

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

