# Bootstrap Configuration Specification

## Overview

The `.dotbootstrap.yaml` file provides declarative configuration for package installation during repository cloning. It enables:

- Defining available packages with platform requirements
- Creating named installation profiles, optionally inheriting from another profile
- Mapping hostnames to profiles
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
machines: []             # Optional: Ordered host pattern to profile mappings
defaults: {}             # Optional: Default settings
```

### Version

**Type:** String  
**Required:** Yes  
**Values:** `"1.0"`

Specifies the bootstrap configuration schema version. The generated configuration uses `"1.0"`.
Validation only requires the field to be non-empty; other values are accepted but not interpreted.

```yaml
version: "1.0"
```

### Packages

**Type:** Array of PackageSpec  
**Required:** The key is expected, but an empty list passes validation.

Defines all packages available in the repository.

#### PackageSpec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Package directory name |
| `required` | boolean | No | Whether package is mandatory (default: false) |
| `platform` | string[] | No | Target platforms (omit for all platforms) |
| `on_conflict` | string | No | Conflict resolution policy for this package |

Unknown keys are ignored without error. There is no dependency field.

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

**`on_conflict` is currently inert.** The value is validated on load and is written into the file
by `dot clone bootstrap`, but the clone service does not apply it at install time; it reads only
`defaults.profile` from this file. Conflict behaviour comes from `symlinks.overwrite` and
`symlinks.backup` in the dot configuration instead. See
[Configuration Reference](04-configuration.md).

### Profiles

**Type:** Map of string to Profile  
**Required:** No

Named collections of packages for specific installation scenarios.

#### Profile Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | Yes | Human-readable profile description |
| `extends` | string | No | Name of a single parent profile to inherit packages from |
| `packages` | string[] | Yes | List of package names to install |

Profile package names must reference packages defined in the `packages` section.

#### Profile Inheritance

A profile may name one parent with `extends`. The resolved package list is the union of
the whole parent chain and the profile's own packages, ordered from the root ancestor
down to the profile itself, with duplicates removed. A package listed by both a parent
and a child keeps its earlier, more ancestral position.

```yaml
profiles:
  base:
    description: "Everything every machine gets"
    packages:
      - dot-git
      - dot-zsh

  dev:
    description: "Base plus editors"
    extends: base
    packages:
      - dot-vim
      - dot-tmux

  work:
    description: "Dev plus work-only configuration"
    extends: dev
    packages:
      - dot-ssh
```

Here `work` resolves to `dot-git`, `dot-zsh`, `dot-vim`, `dot-tmux`, `dot-ssh`.

Chains may be any depth but each profile has at most one parent. A parent that does not
exist, or a chain that loops back on itself, is a validation error.

### Machines

**Type:** Array of MachineRule  
**Required:** No

Maps hostnames to profiles so a single repository can install a different package set on
each machine without passing `--profile`.

#### MachineRule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `host` | string | Yes | Glob pattern matched against the hostname |
| `profile` | string | Yes | Name of the profile to use on matching hosts |

#### Ordering and Matching

`machines` is an ordered **list**, not a map, because evaluation order is part of the
semantics: patterns are allowed to overlap and **the first matching entry wins**. YAML
mappings do not preserve order once loaded, so a list is the only way to express that.
Put the most specific patterns first and any catch-all last.

Patterns use `path.Match` glob syntax (`*`, `?`, `[class]`), with no special treatment of
dots. Each pattern is matched against both the full hostname and its first label, so an
entry for `hephaestus` also matches `hephaestus.example.com`. Matching is
case-insensitive.

```yaml
machines:
  - host: "hephaestus*"          # Checked first
    profile: devhost

  - host: "*.geicoinf.com"       # Any other work host
    profile: work

  - host: "zeus.local"
    profile: personal

  - host: "*"                    # Catch-all, checked last
    profile: minimal
```

#### Resolution During Clone

When `--profile` is not given, `dot clone` resolves the profile for the current host from
`machines` and logs the entry that matched. Precedence:

1. `--profile`, if given, always wins
2. `--interactive` skips the machines section entirely
3. The first `machines` entry matching the hostname
4. `defaults.profile`, when no entry matches
5. An interactive prompt or all packages, as before

The profile that was used is recorded in the manifest repository section. `dot doctor`
prints an advisory note when the `machines` section maps the current host to a different
profile than the one recorded at clone time. The note is informational only and does not
change doctor's health status or exit code.

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

  development:
    description: "Development environment setup"
    extends: minimal
    packages:
      - dot-tmux
      - dot-git

  full:
    description: "Complete configuration with all packages"
    extends: development
    packages:
      - dot-ssh

machines:
  - host: "*.corp.example.com"
    profile: development

  - host: "laptop"
    profile: full

  - host: "*"
    profile: minimal

defaults:
  on_conflict: backup
  profile: minimal
```

## Validation Rules

Validation is performed on load and covers:

1. **Version present:** `version` must be non-empty. Any non-empty string is accepted.
2. **Package names:** must be non-empty and unique within `packages`.
3. **Platform values:** each entry must be one of `linux`, `darwin`, `windows`, `freebsd`.
4. **Conflict policies:** each `on_conflict`, per package or in `defaults`, must be one of `fail`,
   `backup`, `overwrite`, `skip`.
5. **Profile package references:** every name in a profile's resolved package list, its own
   packages plus everything inherited through `extends`, must be a defined package.
6. **Profile inheritance:** if `extends` is set, the named parent must exist in `profiles`, and
   the chain must not loop back on itself.
7. **Default profile:** if `defaults.profile` is set, it must exist in `profiles`.
8. **Machine entries:** each entry needs a non-empty, well-formed `host` glob and a `profile`
   that exists in `profiles`.

Not validated: whether a package name corresponds to a directory in the repository, whether a
profile is non-empty, and whether a profile has a description. A profile with no description or no
packages loads without complaint.

## Usage Examples

### Clone with Profile

```bash
# Use specific profile
dot clone https://github.com/user/dotfiles --profile minimal

# Use default profile from bootstrap config
dot clone https://github.com/user/dotfiles
```

Like `git clone`, `dot clone` creates a directory named after the repository under the current
working directory unless `--dir` is given.

### Interactive Selection

```bash
# Force interactive selection (overrides defaults.profile)
dot clone https://github.com/user/dotfiles --interactive
```

Package selection resolves in this order:

1. `--profile`, if given
2. `--interactive`, if given
3. The first `machines` entry matching the hostname, if the section is present
4. `defaults.profile` from the bootstrap config, if set
5. An interactive prompt, if attached to a terminal
6. All platform-compatible packages

`--interactive` does not override an explicit `--profile`. Packages whose names are reserved by dot
are skipped with a warning at every step.

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
Error: invalid bootstrap configuration: parse YAML: ...

Check the .dotbootstrap.yaml syntax and validation rules
```

### Missing Required Fields

```
Error: invalid bootstrap configuration: version is required
```

### Invalid Package Reference

```
Error: invalid bootstrap configuration: profile "development" references unknown package: dot-invalid
```

### Platform Not Supported

```
Error: invalid bootstrap configuration: invalid platform "solaris" for package dot-custom
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

### Generating the File

Alternatively, generate the file from an existing installation:

```bash
# Write .dotbootstrap.yaml into the package directory
dot clone bootstrap

# Preview without writing
dot clone bootstrap --dry-run

# Write elsewhere
dot clone bootstrap --output ~/dotfiles/.dotbootstrap.yaml

# Restrict to packages recorded in the manifest
dot clone bootstrap --from-manifest

# Set the default conflict policy in the generated file
dot clone bootstrap --conflict-policy backup

# Overwrite an existing file
dot clone bootstrap --force
```

All discovered packages are emitted with `required: false`, alongside a default conflict policy and
example profile structures. Profiles, `extends`, and `machines` are never invented for you; add
them by hand. Review and customise the result before committing it.

#### Repository Layout Detection

`dot clone bootstrap` also classifies the repository layout by reading one directory level inside
each package:

- A package name or a top-level entry carrying the `dot-` prefix means the prefixed layout, the
  historical default, and nothing extra is written.
- Otherwise, a top-level entry that is itself a dotfile, such as `.config` or `.zshrc`, means the
  full-tree layout: package contents are already real dotfile paths.

For a full-tree repository the command writes `.config/dot/config.yaml` in the package directory
with `dotfile.package_name_mapping: false`, so later dot commands in that repository do not
translate package names. An existing file at that path is left untouched.

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
- Put shared packages in a base profile and `extends` it, rather than repeating lists
- Order `machines` from most specific host pattern to least, with any catch-all last

### Conflict Management

`on_conflict` is not applied at install time (see Conflict Policy Values above), so record intent
here but control actual behaviour through `symlinks.overwrite` and `symlinks.backup` in the dot
configuration.

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

