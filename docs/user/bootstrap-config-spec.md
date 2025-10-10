# Bootstrap Configuration Specification

## Overview

The `.dotbootstrap.yaml` file provides declarative configuration for package installation during repository cloning. It enables:

- Defining available packages with platform requirements
- Creating named installation profiles
- Specifying default behaviors
- Managing conflict resolution policies

## File Location

The bootstrap configuration file must be located at the root of the dotfiles repository:

```
dotfiles/
├── .dotbootstrap.yaml    # Bootstrap configuration
├── dot-vim/              # Package directory
├── dot-zsh/              # Package directory
└── ...
```

## Configuration Schema

### Root Structure

```yaml
version: "1.0"           # Required: Configuration version
packages: []             # Required: List of package specifications
profiles: {}             # Optional: Named installation profiles
defaults: {}             # Optional: Default settings
```

### Version

**Type:** String  
**Required:** Yes  
**Values:** `"1.0"`

Specifies the bootstrap configuration schema version. Currently, only version `1.0` is supported.

```yaml
version: "1.0"
```

### Packages

**Type:** Array of PackageSpec  
**Required:** Yes  
**Minimum:** 1 package

Defines all packages available in the repository.

#### PackageSpec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Package directory name (must exist in repository) |
| `required` | boolean | No | Whether package is mandatory (default: false) |
| `platform` | string[] | No | Target platforms (empty = all platforms) |
| `depends` | string[] | No | Package dependencies (not yet implemented) |
| `on_conflict` | string | No | Conflict resolution policy for this package |

#### Platform Values

Supported platform identifiers:

- `linux` - Linux systems
- `darwin` - macOS systems
- `windows` - Windows systems
- `freebsd` - FreeBSD systems

Packages without `platform` specified are available on all platforms.

#### Conflict Policy Values

- `fail` - Abort if conflicts detected (safest, default)
- `backup` - Backup existing files before linking
- `overwrite` - Replace existing files
- `skip` - Skip conflicting files

### Profiles

**Type:** Map of string to Profile  
**Required:** No

Named collections of packages for specific installation scenarios.

#### Profile Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | Yes | Human-readable profile description |
| `packages` | string[] | Yes | List of package names to install |

Profile package names must reference packages defined in the `packages` section.

### Defaults

**Type:** Object  
**Required:** No

Global default settings applied when not overridden.

#### Defaults Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `on_conflict` | string | No | Default conflict resolution policy |
| `profile` | string | No | Default profile name to use |

## Complete Example

```yaml
version: "1.0"

packages:
  # Core packages - all platforms
  - name: dot-vim
    required: true

  - name: dot-zsh
    required: false

  - name: dot-tmux
    required: false
    on_conflict: backup

  # Platform-specific packages
  - name: dot-linux-config
    required: false
    platform:
      - linux

  - name: dot-macos-config
    required: false
    platform:
      - darwin

  # Optional packages
  - name: dot-git
    required: false

  - name: dot-ssh
    required: false
    on_conflict: fail

profiles:
  minimal:
    description: "Minimal configuration with essential tools only"
    packages:
      - dot-vim
      - dot-zsh

  full:
    description: "Complete configuration with all packages"
    packages:
      - dot-vim
      - dot-zsh
      - dot-tmux
      - dot-git
      - dot-ssh

  development:
    description: "Development environment setup"
    packages:
      - dot-vim
      - dot-zsh
      - dot-tmux
      - dot-git

defaults:
  on_conflict: backup
  profile: minimal
```

## Validation Rules

### Package Validation

1. **Unique Names:** Package names must be unique within the `packages` array
2. **Directory Existence:** Package names must correspond to directories in the repository
3. **Platform Values:** Platform identifiers must be from the supported list
4. **Conflict Policies:** Must be one of: `fail`, `backup`, `overwrite`, `skip`

### Profile Validation

1. **Package References:** Profile packages must reference defined package names
2. **Non-Empty:** Profiles must contain at least one package
3. **Description Required:** Each profile must have a description

### Defaults Validation

1. **Profile Existence:** Default profile must exist in `profiles` map
2. **Valid Conflict Policy:** Default conflict policy must be valid value

## Usage Examples

### Clone with Profile

```bash
# Use specific profile
dot clone https://github.com/user/dotfiles --profile minimal

# Use default profile from bootstrap config
dot clone https://github.com/user/dotfiles
```

### Interactive Selection

```bash
# Force interactive selection (ignores profiles)
dot clone https://github.com/user/dotfiles --interactive
```

### Platform Filtering

Platform filtering is automatic. On macOS:

```yaml
packages:
  - name: dot-linux-config
    platform: [linux]          # Not offered

  - name: dot-macos-config
    platform: [darwin]         # Offered

  - name: dot-vim              # Offered (all platforms)
```

### Without Bootstrap Config

If `.dotbootstrap.yaml` is not present:

- All package directories are discovered
- Interactive terminal: User selects packages
- Non-interactive mode: All packages installed

## Error Messages

### Invalid YAML Syntax

```
Error: invalid bootstrap config: failed to parse bootstrap configuration
Check the .dotbootstrap.yaml syntax and validation rules
```

### Missing Required Fields

```
Error: invalid bootstrap config: version field is required
```

### Invalid Package Reference

```
Error: invalid bootstrap config: profile "development" references unknown package "dot-invalid"
```

### Platform Not Supported

```
Error: invalid bootstrap config: package "dot-custom" has invalid platform "solaris"
```

## Migration Guide

### From No Bootstrap Config

If you have an existing dotfiles repository without bootstrap configuration:

1. Identify your package directories
2. Create `.dotbootstrap.yaml` at repository root
3. Define packages with appropriate platforms
4. Create profiles for common scenarios
5. Set sensible defaults

### Example Migration

Before (no bootstrap):

```
dotfiles/
├── dot-vim/
├── dot-zsh/
└── dot-tmux/
```

After (with bootstrap):

```
dotfiles/
├── .dotbootstrap.yaml    # New file
├── dot-vim/
├── dot-zsh/
└── dot-tmux/
```

```yaml
version: "1.0"

packages:
  - name: dot-vim
  - name: dot-zsh
  - name: dot-tmux

profiles:
  default:
    description: "Standard configuration"
    packages:
      - dot-vim
      - dot-zsh
      - dot-tmux

defaults:
  profile: default
```

## Best Practices

### Package Organization

- Use consistent naming: `dot-<tool>` format
- Group related configuration in single packages
- Keep platform-specific packages separate
- Mark essential packages as `required: true`

### Profile Design

- Create `minimal` profile for new machines
- Provide `full` profile for complete setup
- Define role-specific profiles (development, server, etc.)
- Document profile purposes in descriptions

### Conflict Management

- Use `fail` for sensitive files (SSH, GPG keys)
- Use `backup` for user configuration
- Use `skip` for optional enhancements
- Set sensible defaults based on use case

### Platform Support

- Test configurations on target platforms
- Document platform-specific requirements
- Avoid platform-specific package dependencies
- Consider cross-platform alternatives

## Troubleshooting

### Bootstrap Not Found

**Symptom:** `ErrBootstrapNotFound` after cloning

**Causes:**
- File not at repository root
- Incorrect filename (must be `.dotbootstrap.yaml`)
- Clone operation incomplete

**Solution:**
```bash
# Verify file exists
ls -la .dotbootstrap.yaml

# Check file location
pwd  # Should be repository root
```

### Profile Not Found

**Symptom:** `ErrProfileNotFound: profile "name" does not exist`

**Causes:**
- Typo in profile name
- Profile not defined in bootstrap config

**Solution:**
```bash
# List available profiles
cat .dotbootstrap.yaml | grep -A 2 "profiles:"

# Use correct profile name
dot clone <url> --profile minimal
```

### Package Not Available on Platform

**Symptom:** Package not offered during selection

**Cause:** Package filtered by platform specification

**Expected Behavior:** Platform filtering is automatic and correct

