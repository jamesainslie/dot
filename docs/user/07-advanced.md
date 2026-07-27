# Advanced Features

Deep dive into advanced dot features and capabilities.

## Ignore Pattern System

### Pattern Types

#### Glob Patterns

Patterns are globs. Regular expressions are not supported.

```yaml
ignore:
  patterns:
    - "*.log"          # All .log files
    - "*.tmp"          # All .tmp files
    - "test_*"         # Files starting with test_
    - "cache/"         # Directory named cache
    - "**/*.bak"       # .bak files in any directory
```

#### Negation

Prefix a pattern with `!` to re-include a file that an earlier pattern
excluded. Patterns are evaluated in order and the last match wins.

```yaml
ignore:
  patterns:
    - "*.log"
    - "!important.log"
```

The `ignore.overrides` key is deprecated and has no effect; use `!pattern`.
The equivalent command-line form is repeated `--ignore` flags.

### Path Matching

Each pattern is fully anchored and tested twice: against the whole path, and
against the basename alone. A bare name such as `.DS_Store` therefore matches
at any depth.

Patterns containing a path separator behave differently depending on the
caller. Package scanning tests absolute paths, so a relative pattern such as
`.ssh/id_*` matches neither form and does not exclude the file from `manage` or
`adopt`. Several shipped defaults take this form (`.ssh/*.pem`, `.ssh/id_*`,
`.ssh/*_rsa`, `.ssh/*_ecdsa`, `.ssh/*_ed25519`) and are consequently inert
during scanning. Use a basename pattern such as `id_*` to exclude a file
reliably.

Doctor orphan detection maintains its own ignore set, built only from patterns
recorded by `dot doctor ignore --pattern` and matched against target-relative
paths. Separator patterns work there, and the default ignore patterns are not
consulted at all.

### Performance

Each pattern is converted to a regular expression and compiled once when the
ignore set is built, then reused for every path evaluation.

## Directory Folding

Directory folding is not implemented. The planner emits one symlink per file
in the package; there is no case in which a directory-level symlink is created
for a managed package.

```
~/.vim/colors/theme.vim -> ~/dotfiles/dot-vim/colors/theme.vim
~/.vim/autoload/plugin.vim -> ~/dotfiles/dot-vim/autoload/plugin.vim
~/.vim/ftplugin/go.vim -> ~/dotfiles/dot-vim/ftplugin/go.vim
```

The `symlinks.folding` configuration key is accepted and appears in
`dot config list`, but nothing reads it, so setting it has no effect. There is
no `--no-folding` flag and no per-package metadata file that controls folding.

Adopted directories are the one place a single directory-level symlink appears,
because `dot adopt` on a directory moves its contents to the package root and
links the original path to that root. That structure is a property of adoption,
not of folding.

## Link Mode

All symlinks are absolute:

```
/home/user/.vim/.vimrc -> /home/user/dotfiles/dot-vim/dot-vimrc
```

The `symlinks.mode` configuration key accepts `relative` and `absolute` but is
not connected to the planner, so it has no effect. Moving the package directory
breaks every link; recreate them with `dot remanage`.

## Dry-Run Mode

### Usage

Preview operations before applying:

```bash
# Preview any command
dot --dry-run manage dot-vim
dot --dry-run unmanage dot-zsh
dot --dry-run remanage dot-tmux
```

### Output

Detailed plan showing operations:

```
Dry run mode - no changes will be applied

Plan:
  + Create directory: /home/user/.vim
  + Create directory: /home/user/.vim/colors
  + Create symlink: /home/user/.vim/colors/theme.vim -> /home/user/dotfiles/dot-vim/colors/theme.vim
  + Create symlink: /home/user/.vim/.vimrc -> /home/user/dotfiles/dot-vim/dot-vimrc

Summary:
  Directories: 2
  Symlinks: 2
  Conflicts: 0
```

The only operation verbs are `Create directory`, `Create symlink`,
`Move file`, `Backup file`, `Delete directory`, and `Delete symlink`. There is
no update verb; a changed link is expressed as a delete followed by a create.
Zero-valued summary rows are omitted.

### Conflict Detection

Dry-run detects conflicts without modification:

```bash
dot --dry-run manage dot-vim
# Output shows conflicts without creating any symlinks
```

## Resolution Policies

Conflict policy is read from the configuration file. There is no
command-line flag for it.

```yaml
symlinks:
  overwrite: false  # replace conflicting files
  backup: false     # move conflicting files to the backup directory
```

Precedence is `overwrite`, then `backup`, then fail.

#### Fail Policy (Default)

Stops on the first conflict and reports it. Safest option; requires manual
resolution.

#### Backup Policy

Moves the conflicting file into the backup directory, then creates the symlink.
Enable with `symlinks.backup: true`.

Backups are written to `<target>/.dot-backup` (override with `--backup-dir`) as
`<filename>.<hash>.<timestamp>`, for example
`.vimrc.3f9a1c07.20260727-121603`. The hash is derived from the full source
path, so files sharing a basename across directories never collide. The backup
directory is created on demand. The `symlinks.backup_suffix` setting is not
used by this policy.

#### Overwrite Policy

Deletes the conflicting file and creates the symlink. Enable with
`symlinks.overwrite: true`. This takes precedence over `symlinks.backup`.

#### Skip Policy

The planner implements a skip policy, but it is not reachable from the CLI or
the configuration file.

Policies cannot be configured per package. The configuration file has no
per-package section.

## State Management

### Manifest Structure

The manifest is `.dot-manifest.json` in the manifest directory, by default
`~/.local/share/dot/manifest/`. A `.dot-manifest.lock` file sits beside it. The
manifest falls back to `<target>/.dot-manifest.json` only when
`directories.manifest` is empty.

```json
{
  "version": "1.0",
  "updated_at": "2026-07-27T12:25:00Z",
  "packages": {
    "dot-vim": {
      "name": "dot-vim",
      "installed_at": "2026-07-27T12:25:00Z",
      "link_count": 2,
      "links": [".vim/.vimrc", ".vim/colors/theme.vim"],
      "source": "managed",
      "target_dir": "/home/user",
      "package_dir": "/home/user/dotfiles/dot-vim"
    }
  },
  "hashes": {
    "dot-vim": "a3f2c8b4d9e1f0..."
  }
}
```

### Fast Status Queries

The manifest allows status queries without walking the filesystem:

```bash
dot status
```

Each listed package is still stat-checked for link health, which is why
`status` reports an `is_healthy` field. The full orphan scan belongs to
`doctor`.

### State Validation

Check manifest consistency:

```bash
dot doctor
```

There is no repair flag. To rebuild state, remanage the affected packages. If
the manifest itself is unreadable, remove
`~/.local/share/dot/manifest/.dot-manifest.json` and reinstall the packages
with `dot manage`.

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
dot remanage dot-vim dot-zsh dot-tmux dot-git
```

The command reports a total only:

```
✓ Remanaged 4 packages
```

There are no per-package changed or skipped lines. Use `-vv` to see the
detection decisions in the debug log.

### Forcing Full Processing

There is no flag to disable incremental detection. To force a full reinstall,
unmanage and manage the packages:

```bash
dot unmanage dot-vim dot-zsh dot-tmux
dot manage dot-vim dot-zsh dot-tmux
```

## Parallel Execution

### Concurrent Operations

dot executes independent operations concurrently within each dependency batch:

```bash
# Operations across these packages are batched and run in parallel
dot manage dot-vim dot-zsh dot-tmux dot-git dot-nvim
```

Parallelism is fixed at the CPU count. It is not configurable from the command
line, the configuration file, or the environment.

### Dependency-Safe Batching

Operations grouped by dependencies:

- Batch 1: Independent operations (parallel)
- Batch 2: Operations depending on batch 1 (parallel)
- Batch 3: Operations depending on batch 2 (parallel)

### Rollback and Interruption

When an operation in a plan fails, dot rolls back the operations that already
succeeded. Rollback runs on a context detached from cancellation, so a single
interrupt still restores state; a second interrupt forces exit 130 and can
leave rollback incomplete.

Some operations cannot be reversed. Removing a directory tree is one; these
report `cannot roll back operation on <path>` and the failure summary counts
them as "operations could not be rolled back". Deletion is also verified
before it happens: dot refuses to remove a regular file or a symlink that now
points elsewhere, reporting `refusing to delete <path>: <reason>`.

## Performance Tuning

### Optimization Strategies

#### 1. Use Incremental Updates

Remanage skips unchanged packages:

```bash
dot remanage dot-vim dot-zsh dot-tmux  # Fast
```

#### 2. Limit Doctor Scanning

`dot doctor --scan-mode=off` skips orphan detection entirely; `scoped` (the
default) restricts the scan to directories containing managed links.

#### 3. Optimize Ignore Patterns

Fewer patterns mean faster scanning:

```yaml
ignore:
  patterns:
    - ".git"       # Essential only
    - "node_modules"
```

### Performance Monitoring

Profile operations:

```bash
# Time operations
time dot manage dot-vim

# Verbose timing
dot -vv manage dot-vim

# CPU and heap profiles
dot --cpu-profile cpu.pprof manage dot-vim
dot --mem-profile mem.pprof manage dot-vim
```

## Logging and Output

### Verbosity Levels

Control detail level:

```bash
# Level 0: Errors only
dot manage dot-vim

# Level 1: Info
dot -v manage dot-vim

# Level 2: Debug
dot -vv manage dot-vim

# Level 3: Trace
dot -vvv manage dot-vim
```

### Structured Logging

JSON output for automation:

```bash
# JSON logs
dot --log-json manage dot-vim

# Parse with jq
dot --log-json manage dot-vim 2>&1 | jq '.level'
```

### Quiet Mode

Suppress all output except errors:

```bash
# Script-friendly
dot --quiet manage dot-vim
result=$?

# Batch mode additionally disables interactive prompts
dot --batch manage dot-vim
```

## Output Formats

### Multiple Format Support

`status`, `list`, and `doctor` support several output formats:

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

For `doctor`, `--format table` renders identically to `--format text`.

`status` and `list` emit a JSON object with a `packages` array, not a bare
array. Address it as `.packages[]`:

```bash
dot list --format json | jq -r '.packages[].name'
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

