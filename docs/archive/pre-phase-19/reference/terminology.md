# Command Terminology

## Core Commands

### manage
**Action**: Bring a package under management by creating symbolic links

```bash
dot manage vim
```

Creates symbolic links from the package directory to the target directory (typically `$HOME`). Files in the package become "managed" - tracked and linked by dot.

**Implementation**: Creates symbolic links for all files in the specified package.

### unmanage
**Action**: Remove a package from management by removing symbolic links

```bash
dot unmanage vim
```

Removes symbolic links that were created by `manage`. Only removes links pointing to the package directory - never touches actual configuration files.

**Implementation**: Removes symbolic links, restores directories if empty.

### remanage
**Action**: Update a managed package with changes

```bash
dot remanage vim
```

Efficiently updates symbolic links for a package that has changed. Uses incremental planning to only process modifications, additions, and deletions.

**Implementation**: Compares current state with desired state, applies minimal changes.

### adopt
**Action**: Import existing files into a package

```bash
dot adopt ~/.vimrc --package vim
```

Moves an existing configuration file into a package directory and creates a symbolic link in its place. Allows bringing existing configurations under management.

**Implementation**: Moves file to package, creates symbolic link.

## Query Commands

### status
**Action**: Display management status of packages

```bash
dot status
dot status vim
```

Shows which packages are managed, their health status, and any issues detected.

### doctor
**Action**: Diagnose configuration health

```bash
dot doctor
```

Performs health checks on managed configurations:
- Detects broken symbolic links
- Finds orphaned links (pointing to non-existent packages)
- Identifies permission issues
- Suggests fixes for problems

### list
**Action**: List available or managed packages

```bash
dot list
dot list --managed
```

Displays packages available in the package directory or currently under management.

## Design Rationale

### Why "manage" instead of "stow"?

1. **Semantic Clarity**: "Manage" directly describes the tool's purpose - managing dotfiles
2. **Implementation Agnostic**: Doesn't reveal whether we're linking, copying, or using another mechanism
3. **User-Focused**: Users want to "manage configurations" not "create symbolic links"
4. **Professional**: Matches terminology from infrastructure tools (Chef, Puppet, Ansible)
5. **Scalable**: Works if we add features like templates, encryption, or remote packages

### Comparison with GNU Stow

| GNU Stow | dot | Rationale |
|----------|-----|-----------|
| stow | manage | More intuitive for newcomers |
| unstow | unmanage | Clear reversal of manage |
| restow | remanage | Consistent prefix pattern |
| adopt | adopt | Perfect semantic fit, keep as-is |

### Implementation Independence

The term "manage" allows flexibility in implementation:
- Currently: Creates symbolic links
- Future: Could add hard links, copies, or other mechanisms
- Users care about: "My config is managed"
- Users don't care about: "It's a symbolic link vs hard link"

This abstraction keeps the interface stable even if implementation details evolve.

## Usage Patterns

### Initial Setup
```bash
# Manage your dotfiles
dot manage vim
dot manage zsh
dot manage git
```

### Daily Operations
```bash
# After editing package files
dot remanage vim

# Check everything is healthy
dot doctor
```

### Adopting Existing Configs
```bash
# Import existing configurations
dot adopt ~/.vimrc --package vim
dot adopt ~/.zshrc --package zsh
```

### Cleanup
```bash
# Stop managing a package
dot unmanage vim
```

## Extensibility

The manage/unmanage/remanage terminology supports future features:

- **Remote Packages**: `dot manage user/dotfiles/vim@github`
- **Templates**: `dot manage vim --template hostname={{hostname}}`
- **Encryption**: `dot manage secrets --encrypted`
- **Profiles**: `dot manage vim --profile work`

All extensions fit naturally with the "management" metaphor.

