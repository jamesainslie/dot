# dot

[![CI](https://github.com/jamesainslie/dot/actions/workflows/ci.yml/badge.svg)](https://github.com/jamesainslie/dot/actions/workflows/ci.yml)
[![Release](https://github.com/jamesainslie/dot/actions/workflows/release.yml/badge.svg)](https://github.com/jamesainslie/dot/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamesainslie/dot)](https://goreportcard.com/report/github.com/jamesainslie/dot)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A type-safe symbolic link manager for configuration files and dotfiles, written in Go.

## Overview

dot manages configuration files through symbolic links, providing a centralized approach to dotfile management with strong safety guarantees. The tool creates and maintains symlinks from a directory (package repository) to a target directory (typically home directory), enabling version control and synchronization of configuration files across multiple machines.

### Key Capabilities

- **Package Management**: Install, remove, and update packages containing configuration files
- **Conflict Resolution**: Detect and resolve conflicts with configurable resolution policies
- **Incremental Operations**: Content-based change detection for efficient updates
- **Transactional Safety**: Two-phase commit with automatic rollback on failure
- **State Tracking**: Manifest-based state management for fast status queries
- **Parallel Execution**: Concurrent operation processing for improved performance
- **Cross-Platform**: Supports Linux, macOS, BSD, and Windows (with limitations)

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap jamesainslie/dot
brew install dot
```

### From Binary Releases

Download the latest release for your platform from [GitHub Releases](https://github.com/jamesainslie/dot/releases).

```bash
# Linux/macOS
curl -L https://github.com/jamesainslie/dot/releases/latest/download/dot-$(uname -s)-$(uname -m).tar.gz | tar xz
sudo mv dot /usr/local/bin/
```

### From Source

Requires Go 1.25 or later:

```bash
go install github.com/jamesainslie/dot/cmd/dot@latest
```

### Verification

```bash
dot --version
```

## Quick Start

### Initial Setup

Create a package directory to store your packages:

```bash
mkdir -p ~/dotfiles/vim
echo "set number" > ~/dotfiles/vim/dot-vimrc
```

### Manage a Package

Install the package by creating symlinks:

```bash
cd ~/dotfiles
dot manage vim
```

This creates `~/.vimrc` pointing to `~/dotfiles/vim/dot-vimrc`.

### Check Status

View installed packages and their status:

```bash
dot status
```

### Unmanage a Package

Remove symlinks for a package:

```bash
dot unmanage vim
```

## Core Concepts

### Package Directory

The source directory containing packages. Each subdirectory represents a package. Default: current directory.

### Target Directory

The destination directory where symlinks are created. Default: `$HOME`.

### Package

A directory within the package directory containing configuration files. Package structure mirrors the target directory structure.

### Dotfile Translation

Files prefixed with `dot-` in packages are linked as dotfiles (leading `.`). For example:
- `dot-bashrc` → `.bashrc`
- `dot-config/nvim/init.vim` → `.config/nvim/init.vim`

### Directory Folding

When all files in a directory belong to a single package, dot creates a single directory-level symlink instead of individual file symlinks, reducing symlink count and improving performance.

## Usage

### Package Management Commands

#### Manage (Install)

Create symlinks for packages:

```bash
# Single package
dot manage vim

# Multiple packages
dot manage vim tmux zsh

# With options
dot --dir ~/dotfiles --target $HOME manage vim
dot --dry-run manage vim        # Preview changes
dot --no-folding manage vim     # Disable directory folding
dot --absolute manage vim       # Use absolute symlinks
```

#### Unmanage (Remove)

Remove symlinks for packages:

```bash
# Single package
dot unmanage vim

# Multiple packages
dot unmanage vim tmux

# Preview removal
dot --dry-run unmanage vim
```

#### Remanage (Update)

Update packages (remove and reinstall with incremental detection):

```bash
# Update packages efficiently
dot remanage vim

# Updates only changed packages
dot remanage vim tmux zsh
```

#### Adopt (Import)

Move existing files into a package and replace with symlinks:

```bash
# Adopt single file
dot adopt vim ~/.vimrc

# Adopt multiple files
dot adopt zsh ~/.zshrc ~/.zprofile ~/.zshenv
```

### Query Commands

#### Status

Display installation status:

```bash
# All packages
dot status

# Specific packages
dot status vim tmux

# Different formats
dot status --format json
dot status --format yaml
dot status --format table
```

#### Doctor

Validate installation consistency and detect issues:

```bash
# Health check
dot doctor

# With detailed output
dot doctor -v

# JSON output for scripting
dot doctor --format json
```

Exit codes:
- 0: No issues detected
- 1: Issues found

#### List

Show installed packages:

```bash
# List all packages
dot list

# Sort by various fields
dot list --sort name      # Alphabetical (default)
dot list --sort links     # By link count
dot list --sort date      # By installation date

# Different formats
dot list --format table
dot list --format json
```

### Global Options

```bash
-d, --dir PATH       Package directory (default: current directory)
-t, --target PATH    Target directory (default: $HOME)
-n, --dry-run        Preview changes without applying
-v, --verbose        Increase verbosity (repeatable: -v, -vv, -vvv)
    --quiet          Suppress non-error output
    --log-json       Output logs in JSON format
    --no-folding     Disable directory folding optimization
    --absolute       Use absolute symlinks instead of relative
    --ignore PATTERN Add ignore pattern (repeatable)
```

## Configuration

dot supports configuration files in YAML, JSON, or TOML formats.

### Configuration Locations

Searched in order (later sources override earlier):

1. System-wide: `/etc/dot/config.yaml`
2. User global: `~/.config/dot/config.yaml` (XDG) or `~/.dotrc`
3. Project local: `./.dotrc`
4. Environment variables: `DOT_*` prefix
5. Command-line flags (highest priority)

### Configuration Example

```yaml
# ~/.config/dot/config.yaml
packageDir: ~/dotfiles
targetDir: ~
linkMode: relative
folding: true
verbosity: 0

ignore:
  - "*.log"
  - ".git"
  - ".DS_Store"
  - "*.swp"

override: []

backupDir: ~/.dot-backups
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `packageDir` | string | `.` | Source directory containing packages |
| `targetDir` | string | `$HOME` | Destination for symlinks |
| `linkMode` | string | `relative` | Link mode: `relative` or `absolute` |
| `folding` | boolean | `true` | Enable directory folding optimization |
| `verbosity` | integer | `0` | Logging verbosity (0-3) |
| `ignore` | array | (defaults) | File patterns to exclude |
| `override` | array | `[]` | Patterns to force include |
| `backupDir` | string | (none) | Directory for conflict backups |

See [Configuration Reference](docs/user/04-configuration.md) for complete documentation.

## Conflict Resolution

dot detects conflicts when existing files or symlinks prevent package installation.

### Resolution Policies

Configure via `--on-conflict` flag or configuration file:

- **fail** (default): Stop and report conflicts
- **backup**: Move conflicting files to backup location
- **overwrite**: Replace conflicting files with symlinks
- **skip**: Skip conflicting files and continue

Example:

```bash
# Backup existing files
dot --on-conflict backup manage vim

# Skip conflicts and continue
dot --on-conflict skip manage vim tmux
```

See [User Guide - Workflows](docs/user/06-workflows.md) for conflict resolution strategies.

## System Requirements

### Operating Systems

- Linux (all distributions)
- macOS 10.15 or later
- FreeBSD, OpenBSD, NetBSD
- Windows 10 or later (with symlink support enabled)

### Filesystems

Full support:
- ext4, btrfs, xfs (Linux)
- APFS, HFS+ (macOS)
- ZFS (all platforms)

Limited support:
- FAT32, exFAT (no symlink support)
- Network filesystems (NFS, SMB) with caveats

### Architectures

- amd64 (x86-64)
- arm64 (aarch64)
- 386 (x86)
- arm (32-bit ARM)

## Documentation

### User Documentation

- [Introduction and Core Concepts](docs/user/01-introduction.md)
- [Installation Guide](docs/user/02-installation.md)
- [Quick Start Tutorial](docs/user/03-quickstart.md)
- [Configuration Reference](docs/user/04-configuration.md)
- [Command Reference](docs/user/05-commands.md)
- [Common Workflows](docs/user/06-workflows.md)
- [Advanced Features](docs/user/07-advanced.md)
- [Troubleshooting Guide](docs/user/08-troubleshooting.md)
- [Glossary](docs/user/09-glossary.md)

### Developer Documentation

- [Architecture Overview](docs/architecture/architecture.md)
- [Architecture Decision Records](docs/architecture/adr/)
- [Contributing Guide](CONTRIBUTING.md)
- [Release Workflow](docs/developer/release-workflow.md)
- [Testing Strategy](docs/developer/testing.md)
- [API Reference](docs/developer/api-reference.md)
- [Internal Architecture](docs/developer/internal-architecture.md)
- [Code Style Guide](docs/developer/style-guide.md)
- [Performance Guide](docs/developer/performance.md)

### Examples

- [Basic Usage Examples](examples/basic/)
- [Configuration Examples](examples/configuration/)
- [Workflow Examples](examples/workflows/)
- [Library Embedding Examples](examples/library/)

## Development

### Building

```bash
make build
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Linting

```bash
# Run all linters
make lint

# Run specific checks
make lint-go
make lint-docs
```

### Quality Checks

```bash
# Run complete quality suite
make check
```

This runs tests, linting, and builds in sequence.

## Architecture

dot follows a layered architecture with functional programming principles:

### Layers

1. **Domain Layer**: Pure domain model with phantom-typed paths for compile-time safety
2. **Core Layer**: Pure functional logic for scanning, planning, and resolution
3. **Pipeline Layer**: Composable pipeline stages with generic type parameters
4. **Executor Layer**: Side-effecting operations with two-phase commit and rollback
5. **API Layer**: Clean public Go library interface for embedding
6. **CLI Layer**: Cobra-based command-line interface

### Design Principles

- **Functional Core, Imperative Shell**: Pure planning with isolated side effects
- **Type Safety**: Phantom types prevent path-related bugs at compile time
- **Explicit Errors**: Result types and error aggregation, never silent failures
- **Transactional**: Two-phase commit with automatic rollback on errors
- **Observable**: Structured logging, distributed tracing, metrics collection
- **Testable**: Pure core enables property-based testing of algebraic laws

See [Architecture Documentation](docs/architecture/architecture.md) for details.

## Library Usage

dot can be embedded as a Go library:

```go
package main

import (
    "context"
    "github.com/jamesainslie/dot/pkg/dot"
)

func main() {
    cfg := dot.Config{
        PackageDir: "/home/user/dotfiles",
        TargetDir:  "/home/user",
        LinkMode:   dot.LinkRelative,
        Folding:    true,
    }
    
    client, err := dot.New(cfg)
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    if err := client.Manage(ctx, "vim", "tmux"); err != nil {
        panic(err)
    }
}
```

See [API Reference](docs/developer/api-reference.md) and [Library Examples](examples/library/).

## Contributing

Contributions are welcome. All contributions must follow project standards:

### Requirements

- Test-driven development: write tests before implementation
- Minimum 80% test coverage for new code
- All linters must pass without warnings
- Conventional Commits specification for commit messages
- Atomic commits: one logical change per commit
- Academic documentation style: factual, precise, no hyperbole

### Process

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Implement the feature
5. Ensure all tests and linters pass
6. Submit a pull request

See [Contributing Guide](CONTRIBUTING.md) for detailed guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Project Standards

This project adheres to strict development standards:

- **Language**: Go 1.25
- **Development**: Test-Driven Development (TDD) mandatory
- **Testing**: Minimum 80% coverage, property-based tests for core logic
- **Commits**: Atomic commits with Conventional Commits format
- **Code Style**: golangci-lint v2 with comprehensive linter set
- **Documentation**: Academic style, factual, technically precise
- **Versioning**: Semantic Versioning 2.0.0
- **Changelog**: Keep a Changelog format

## Comparison with GNU Stow

dot was inspired by the awesome [GNU Stow](https://www.gnu.org/software/stow/) tool. Dot provides feature parity with GNU Stow plus additional capabilities:

| Feature | dot | GNU Stow |
|---------|-----|----------|
| Basic stow/unstow | Yes | Yes |
| Conflict detection | Yes | Yes |
| Directory folding | Yes | Yes |
| Incremental updates | Yes | No |
| Transactional operations | Yes | No |
| Parallel execution | Yes | No |
| Adopt existing files | Yes | No |
| Status/health checks | Yes | No |
| Multiple output formats | Yes | No |
| Type safety | Yes | No |
| Cross-platform | Yes | Limited |

See [Migration Guide](docs/user/migration-from-stow.md) for transitioning from GNU Stow.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/jamesainslie/dot/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jamesainslie/dot/discussions)

## Project Status

**Current Version**: v0.1.0-rc (Release Candidate)

**Stability**: Release Candidate - API may change before v0.1.0 final release

See [Implementation Plan](docs/planning/implementation-plan.md) for project roadmap.

## Acknowledgments

Inspired by GNU Stow, reimagined with modern language features and safety guarantees.
