# Glossary

Technical terms and concepts used in dot documentation.

## A

### Absolute Link
A symbolic link using an absolute path to its target. Example: `/home/user/.vimrc -> /home/user/dotfiles/vim/dot-vimrc`. Less portable but robust to directory moves. See also: [Relative Link](#relative-link).

### Adopt
Operation that moves existing files into a package and replaces them with symlinks. Brings unmanaged files under dot management. Command: `dot adopt PACKAGE FILE...`

## C

### Conflict
Situation where dot cannot create a symlink because a file or directory already exists at the target location. Resolved via policies: fail, backup, overwrite, or skip.

### Conflict Resolution Policy
Strategy for handling conflicts during installation. Options: `fail` (stop), `backup` (preserve existing), `overwrite` (replace), `skip` (continue).

### Concurrency
Number of parallel operations dot executes simultaneously. Configurable via `concurrency` option or `DOT_CONCURRENCY` environment variable.

### Content Hash
Cryptographic hash of package contents used for change detection in incremental operations. Stored in manifest.

## D

### Directory Folding
Optimization that creates directory-level symlinks instead of per-file symlinks when all directory contents belong to single package. Reduces symlink count and improves performance. Disable with `--no-folding`.

### Dotfile
Configuration file with name starting with period (`.`), typically hidden by default in Unix systems. Examples: `.vimrc`, `.bashrc`, `.gitconfig`.

### Dotfile Translation
Automatic renaming of files with `dot-` prefix to dotfiles in target directory. Example: `dot-vimrc` becomes `.vimrc`. Enables version control friendly storage.

### Dry-Run Mode
Execution mode that previews operations without applying changes. Enabled with `--dry-run` or `-n` flag. Shows plan and detects conflicts without filesystem modification.

## F

### Folding
See [Directory Folding](#directory-folding).

## I

### Ignore Pattern
Pattern specifying files to exclude from management. Supports glob and regex syntax. Configured in ignore lists or per-package.

### Incremental Operation
Operation that processes only changed packages by comparing content hashes. Significantly faster for large package collections. Used by `remanage` command.

## L

### Link Mode
Type of symlink to create: `relative` (default, portable) or `absolute` (robust to moves). Configured via `linkMode` option.

## M

### Manage
Primary operation that installs packages by creating symlinks from target directory to package files. Command: `dot manage PACKAGE...`

### Manifest
State file (`.dot-manifest.json`) tracking installed packages, symlinks, and content hashes. Enables fast status queries and incremental operations. Stored in target directory.

## O

### Operation
Atomic action performed during installation: create symlink, create directory, delete symlink, etc. Operations have dependencies and execute in topological order.

### Override Pattern
Pattern forcing inclusion of files that would otherwise be ignored. Has higher priority than ignore patterns.

## P

### Package
Directory within package directory containing related configuration files. Structure mirrors target directory layout. Each subdirectory in package directory represents a package.

### Package Directory
Source directory containing packages. Each subdirectory is a package. Also called package directory. Default: current directory. Specified with `--dir` flag.

### Parallel Execution
Concurrent execution of independent operations. dot automatically computes safe parallelization based on operation dependencies.

### Phantom Type
Type-level marker providing compile-time safety without runtime overhead. dot uses phantom types for paths to prevent mixing incompatible path types.

## R

### Relative Link
Symbolic link using relative path to its target. Example: `.vimrc -> ../../dotfiles/vim/dot-vimrc`. Portable across different mount points. Default link mode.

### Remanage
Operation that updates packages efficiently using incremental detection. Processes only changed packages. Command: `dot remanage PACKAGE...`

### Resolution Policy
See [Conflict Resolution Policy](#conflict-resolution-policy).

### Rollback
Automatic reversal of operations when errors occur. dot uses two-phase commit with rollback to guarantee consistent state.

## S

### State
Current installation status including which packages are installed and where symlinks point. Tracked in manifest file.

### Stow Directory
See [Package Directory](#package-directory).

### Symbolic Link (Symlink)
Filesystem reference pointing to another file or directory. dot creates symlinks from target directory to package files. Also called soft link.

## T

### Target Directory
Destination directory where symlinks are created. Typically home directory (`$HOME`). Specified with `--target` flag.

### Two-Phase Commit
Transaction protocol ensuring atomic operations. Phase 1 validates all operations, phase 2 executes them. Enables rollback on failure.

## U

### Unmanage
Operation that removes package by deleting its symlinks. Only removes links pointing to package directory, preserves other files. Command: `dot unmanage PACKAGE...`

## V

### Verbosity
Level of logging detail. Levels: 0 (errors), 1 (info), 2 (debug), 3 (trace). Controlled with `-v` flags (repeatable) or `verbosity` configuration.

## GNU Stow Terminology Mapping

Mapping between GNU Stow and dot terminology:

| GNU Stow | dot | Notes |
|----------|-----|-------|
| stow | manage | Primary installation command |
| unstow | unmanage | Removal command |
| restow | remanage | Update command (dot has incremental detection) |
| - | adopt | New command for importing files |
| package directory | package directory | Source directory |
| target directory | target directory | Same |
| package | package | Same |
| folding | directory folding | Same concept |

## Command Terminology

### Core Commands

- **manage**: Install package(s) by creating symlinks
- **unmanage**: Remove package(s) by deleting symlinks
- **remanage**: Update package(s) efficiently
- **adopt**: Import existing files into package

### Query Commands

- **status**: Show installation status
- **doctor**: Validate installation health
- **list**: Show installed packages

### Utility Commands

- **version**: Display version information
- **help**: Show help information
- **completion**: Generate shell completion
- **config**: Manage configuration

## Acronyms and Abbreviations

- **ADR**: Architecture Decision Record
- **API**: Application Programming Interface
- **CLI**: Command-Line Interface
- **CI/CD**: Continuous Integration/Continuous Deployment
- **JSON**: JavaScript Object Notation
- **LRU**: Least Recently Used (cache eviction strategy)
- **TDD**: Test-Driven Development
- **TOML**: Tom's Obvious, Minimal Language
- **TUI**: Terminal User Interface
- **XDG**: Cross-Desktop Group (specification)
- **YAML**: YAML Ain't Markup Language

## Configuration Terms

- **Precedence**: Order in which configuration sources override each other
- **Merge Strategy**: How array values combine across configuration sources
- **Config File**: YAML/JSON/TOML file containing dot configuration
- **Environment Variable**: Shell variable prefixed with `DOT_` controlling behavior
- **Global Option**: Flag affecting all commands

## Technical Terms

- **Phantom Type**: Type parameter used only for compile-time checks
- **Result Monad**: Functional programming pattern for error handling
- **Pipeline**: Composition of processing stages
- **Transactional**: All-or-nothing operation semantics
- **Idempotent**: Operation produces same result when repeated
- **Atomic Commit**: Indivisible unit of work
- **Topological Sort**: Ordering by dependencies

## See Also

- [Introduction](01-introduction.md): Core concepts explained
- [Command Reference](05-commands.md): Complete command documentation
- [Configuration Reference](04-configuration.md): Configuration options

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

