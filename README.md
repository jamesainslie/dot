# dot

[![CI](https://github.com/jamesainslie/dot/actions/workflows/ci.yml/badge.svg)](https://github.com/jamesainslie/dot/actions/workflows/ci.yml)
[![Release](https://github.com/jamesainslie/dot/actions/workflows/release.yml/badge.svg)](https://github.com/jamesainslie/dot/actions/workflows/release.yml)
![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/jamesainslie/dot?utm_source=oss&utm_medium=github&utm_campaign=jamesainslie%2Fdot&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)

A type-safe dotfile and configuration management tool with modern Go practices.

## Overview

dot manages symbolic links for configuration files and dotfiles through a clean, intuitive interface. The tool uses phantom types for compile-time path safety and follows functional programming principles with a pure core and imperative shell architecture.

## Status

**Current Version**: v0.0.0 (in development)

**Completed Phases**:
- Phase 0-13: Core infrastructure and commands
- Phase 14: Query commands (status, doctor, list)

**Current Phase**: Phase 15 (Error Handling and User Experience)

## Features

Planned features for v0.1.0:
- Manage packages by creating symbolic links to target directory
- Unmanage packages by removing symbolic links from target directory
- Remanage packages with incremental updates
- Adopt existing files into packages
- Conflict detection and resolution
- Transactional operations with rollback
- Cross-platform support (Linux, macOS, Windows)

## Requirements

- Go 1.25 or later

## Installation

```bash
# From source (once available)
go install github.com/user/dot/cmd/dot@latest
```

## Usage

### Package Management

```bash
# Manage a package (create symbolic links)
dot manage vim

# Manage multiple packages
dot manage vim tmux zsh

# Unmanage a package (remove symbolic links)
dot unmanage vim

# Remanage a package (update symbolic links)
dot remanage vim

# Adopt existing files into a package
dot adopt .vimrc vim
```

### Query Commands

```bash
# Show status of all installed packages
dot status

# Show status for specific packages
dot status vim tmux

# Show status in different formats
dot status --format=json
dot status --format=yaml
dot status --format=table

# List all installed packages
dot list

# List with sorting
dot list --sort=links    # Sort by link count
dot list --sort=date     # Sort by installation date
dot list --sort=name     # Sort by name (default)

# Verify installation health
dot doctor

# Doctor with different output formats
dot doctor --format=json
dot doctor --format=table
```

### Global Options

```bash
# Specify directories
dot --dir=/path/to/dotfiles --target=$HOME manage vim

# Dry run (preview changes)
dot --dry-run manage vim

# Verbose output
dot -v manage vim
dot -vv manage vim  # More verbose
dot -vvv manage vim # Maximum verbosity

# Quiet mode (errors only)
dot --quiet manage vim

# JSON logging
dot --log-json manage vim
```

Detailed documentation will be provided as features are implemented.

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## Architecture

The project follows a layered architecture:
- **Domain Layer**: Pure domain model with phantom-typed paths
- **Port Layer**: Infrastructure interfaces
- **Adapter Layer**: Concrete implementations (filesystem, logging)
- **Core Layer**: Pure functional planning and resolution logic
- **Shell Layer**: Side-effecting execution with transactions
- **API Layer**: Public Go library interface
- **CLI Layer**: Cobra-based command-line interface

## Contributing

Contributions must follow project standards:
- Test-driven development (write tests first)
- Minimum 80% test coverage
- All linters must pass
- Conventional Commits specification
- No emojis in code, documentation, or output

See CONTRIBUTING.md for detailed guidelines.

## License

MIT License - see LICENSE file for details.

## Project Standards

This project adheres to strict development standards:
- Go 1.25 target version
- Test-first development (TDD)
- Atomic commits with conventional commit messages
- Functional programming where applicable
- Academic documentation style (no hyperbole)
- golangci-lint v2 with comprehensive linter set
- Semantic versioning
- Keep a Changelog format

---
Made with ❤️ by James Ainslie
