# Glossary

Technical terms and concepts used in dot documentation.

## A

### Absolute Link
A symbolic link using an absolute path to its target. Example: `/home/user/.vimrc -> /home/user/dotfiles/dot-vim/dot-vimrc`. dot always creates absolute links. See also: [Link Mode](#link-mode).

### Adopt
Operation that moves existing files into a package and replaces them with symlinks. Brings unmanaged files under dot management. Command: `dot adopt PACKAGE FILE...`

## C

### Conflict
Situation where dot cannot create a symlink because a file or directory already exists at the target location.

### Conflict Resolution Policy
Strategy for handling conflicts during installation. The policy is derived from two configuration keys, not from a flag: `symlinks.overwrite` replaces the conflicting file, `symlinks.backup` moves it into the backup directory first. Precedence is overwrite, then backup, then fail. Fail is the default. A `skip` policy exists in the planner but is not reachable from the CLI or the configuration file.

### Concurrency
Number of parallel operations dot executes within a batch. Fixed at the CPU count. It is not configurable from the CLI, the configuration file, or the environment.

### Content Hash
Cryptographic hash of package contents used for change detection in incremental operations. Stored in manifest.

## D

### Directory Folding
Optimization that would create directory-level symlinks instead of per-file symlinks. Not implemented; the planner always creates one symlink per file. The `symlinks.folding` configuration key is accepted but has no effect, and there is no `--no-folding` flag.

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
Pattern specifying files to exclude from management. Glob syntax only; regular expressions are not supported. Configured in `ignore.patterns`, via repeated `--ignore` flags, or in a package's `.dotignore`. Each pattern is fully anchored and tested against the whole path and, separately, against the basename, so `.DS_Store` matches at any depth.

Package scanning tests absolute paths, so a pattern containing a path separator, such as `.ssh/id_*`, never matches and does not exclude the file from `manage` or `adopt`. The shipped defaults that name a directory (`.ssh/*.pem`, `.ssh/id_*`, `.ssh/*_rsa`, `.ssh/*_ecdsa`, `.ssh/*_ed25519`) are therefore inert. Use a basename pattern such as `id_*` instead.

Doctor orphan detection uses a separate ignore set built only from entries added with `dot doctor ignore --pattern`, matched against target-relative paths. Patterns containing a separator do work there. The default ignore patterns are not consulted by orphan detection at all.

### Incremental Operation
Operation that processes only changed packages by comparing content hashes. Significantly faster for large package collections. Used by `remanage` command.

## L

### Link Mode
Intended selector between relative and absolute symlinks, expressed as `symlinks.mode`. The key is validated and reported but is not wired to the planner; all links dot creates are absolute.

## M

### Manage
Primary operation that installs packages by creating symlinks from target directory to package files. Command: `dot manage PACKAGE...`

### Manifest
State file (`.dot-manifest.json`) tracking installed packages, symlinks, and content hashes. Enables fast status queries and incremental operations. Stored in `~/.local/share/dot/manifest/` by default (`directories.manifest`), falling back to the target directory only when that setting is empty. A `.dot-manifest.lock` file sits alongside it.

## N

### Negation Pattern
Ignore pattern prefixed with `!`, re-including a file that an earlier pattern excluded. Patterns are evaluated in order and the last match wins. The deprecated `ignore.overrides` key is not read by any command.

## O

### Operation
Atomic action performed during installation: create symlink, create directory, delete symlink, etc. Operations have dependencies and execute in topological order.

## P

### Package
Directory within the package directory containing related configuration files. The package name becomes a subdirectory of the target: a name beginning with `dot-` is translated to a leading dot (`dot-ssh` targets `~/.ssh/`) and a plain name is kept verbatim (`vim` targets `~/vim/`). This mapping is controlled by `dotfile.package_name_mapping` and is enabled by default.

### Package Name Mapping
Translation of a package name into a target subdirectory. Enabled by default via `dotfile.package_name_mapping`. Disable it to place package contents directly in the target root.

### Package Directory
Source directory containing packages. Each subdirectory is a package. Also called package directory. Default: current directory. Specified with `--dir` flag.

### Parallel Execution
Batched concurrent execution of independent operations. The executor supports it, but plans generated by the CLI never populate execution batches, so operations currently run sequentially.

### Phantom Type
Type-level marker providing compile-time safety without runtime overhead. dot uses phantom types for paths to prevent mixing incompatible path types.

## R

### Relative Link
Symbolic link using a relative path to its target. dot does not currently create relative links; see [Link Mode](#link-mode).

### Remanage
Operation that updates packages efficiently using incremental detection. Processes only changed packages. Command: `dot remanage PACKAGE...`

### Resolution Policy
See [Conflict Resolution Policy](#conflict-resolution-policy).

### Rollback
Reversal of already-executed operations when a plan fails. Rollback runs on a context detached from cancellation, so a single interrupt still restores state; a second interrupt forces exit 130 and can leave rollback incomplete. Some operations cannot be reversed (removing a directory tree, for example); these report `cannot roll back operation on <path>` and the failure summary counts them as "operations could not be rolled back". Deletion is refused outright when the filesystem no longer matches the plan, reporting `refusing to delete <path>: <reason>`.

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
Transaction protocol used by the executor. Phase 1 validates all operations, phase 2 executes them. Enables rollback on failure, subject to the limits described under [Rollback](#rollback).

## U

### Unmanage
Operation that removes package by deleting its symlinks. Only removes links pointing to package directory, preserves other files. Command: `dot unmanage PACKAGE...`

## V

### Verbosity
Level of logging detail: no flag (errors only), `-v` (info), `-vv` (debug), `-vvv` (trace). Controlled by the repeatable `-v` flag alone. The `output.verbosity` configuration key is validated and reported by `dot config list` but does not affect runtime verbosity.

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
| package | package | dot additionally maps the package name to a target subdirectory |
| folding | directory folding | Stow only; dot does not implement folding |

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

- **help**: Show help information
- **completion**: Generate shell completion
- **config**: Manage configuration
- **clone**: Clone a dotfiles repository and install packages
- **upgrade**: Upgrade dot to the latest version

Version information is available through `dot --version`; there is no `version`
subcommand.

## Acronyms and Abbreviations

- **ADR**: Architecture Decision Record
- **API**: Application Programming Interface
- **CLI**: Command-Line Interface
- **CI/CD**: Continuous Integration/Continuous Deployment
- **JSON**: JavaScript Object Notation
- **TDD**: Test-Driven Development
- **TOML**: Tom's Obvious, Minimal Language
- **TUI**: Terminal User Interface
- **XDG**: Cross-Desktop Group (specification)
- **YAML**: YAML Ain't Markup Language

## Configuration Terms

- **Precedence**: Order in which configuration sources override each other
- **Config File**: YAML, JSON, or TOML file containing dot configuration. The default lookup path is always `config.yaml`; JSON and TOML require `DOT_CONFIG`
- **Repository Config**: `.config/dot/config.yaml` inside the package directory. It replaces the user config entirely rather than merging with it
- **Environment Variable**: Shell variable named `DOT_` plus the section-qualified key with dots replaced by underscores, for example `DOT_SYMLINKS_MODE`. Only the keys bound by the loader are read
- **Global Option**: Persistent flag accepted by all commands

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

