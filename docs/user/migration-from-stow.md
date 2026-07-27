# Migration from GNU Stow

Guide for transitioning from GNU Stow to dot.

## Overview

dot provides feature parity with GNU Stow plus modern enhancements. Migration is straightforward as both tools use the same basic concepts.

## Command Mapping

| GNU Stow | dot | Notes |
|----------|-----|-------|
| `stow PACKAGE` | `dot manage PACKAGE` | Install package |
| `stow -D PACKAGE` | `dot unmanage PACKAGE` | Remove package |
| `stow -R PACKAGE` | `dot remanage PACKAGE` | Update package (dot has incremental detection) |
| `stow -n PACKAGE` | `dot --dry-run manage PACKAGE` | Simulate operation |
| `stow -v PACKAGE` | `dot -v manage PACKAGE` | Verbose output |
| `stow -d DIR` | `dot --dir DIR manage PACKAGE` | Specify package directory (also `-d`) |
| `stow -t DIR` | `dot --target DIR manage PACKAGE` | Specify target directory (also `-t`) |
| - | `dot adopt PACKAGE FILE` | New: import existing files |
| - | `dot status` | New: show installation status |
| - | `dot doctor` | New: validate health |
| - | `dot list` | New: list packages |

## Migration Process

### Step 1: Install dot

```bash
brew install dot
# or download binary from releases
```

### Step 2: Understand the Layout Difference

By default dot maps the package name to a target subdirectory, controlled by the
`dotfile.package_name_mapping` key, which is enabled by default. Package `vim`
targets `~/vim/`, not `~/`. A Stow package carried across unchanged will
therefore link into the wrong location.

Choose one of the following before migrating:

1. Keep Stow semantics by setting `dotfile.package_name_mapping: false` in
   `~/.config/dot/config.yaml`. Package contents are then placed directly under
   the target directory, as Stow does. This is the closest equivalent and the
   only option that reproduces top-level dotfiles such as `~/.vimrc` from a
   package named `vim`.
2. Restructure each package so that its name is the directory it owns. A package
   named `dot-ssh` targets `~/.ssh/`, so `dot-ssh/config` becomes `~/.ssh/config`.
   This suits packages that own a single dotted directory, but it cannot place a
   file directly at `~/.vimrc`, because every package maps to a subdirectory.

### Step 3: Test with One Package

```bash
cd ~/dotfiles

# Unstow with Stow
stow -D vim

# Inspect the plan before applying it
dot --dry-run manage vim

# Install with dot
dot manage vim

# Verify the links landed where expected
dot status vim
```

### Step 4: Migrate All Packages

```bash
cd ~/dotfiles

# Unstow all packages
for pkg in */; do
    stow -D "$pkg"
done

# Install with dot
dot manage $(ls -d */ | tr -d '/')

# Verify
dot status
```

### Step 5: Remove GNU Stow (Optional)

```bash
# Once confident
brew uninstall stow
```

## Feature Differences

### dot Enhancements

1. **Incremental Updates**: `remanage` only processes changed packages
2. **Adoption**: Import existing files with `adopt`
3. **Status Queries**: Check installation with `status`, `list`, `doctor`
4. **Multiple Formats**: JSON, YAML, table output
5. **Transactional**: Automatic rollback on failure
6. **Type Safety**: Compile-time path safety
7. **Performance**: Parallel execution

Note that directory folding, which Stow performs, is not implemented in dot.
`manage` always creates one symlink per file.

### Behavioral Differences

1. **Dotfile Translation**: dot uses `dot-` prefix, Stow uses `.` in package
2. **Package Name Mapping**: the package name selects the target subdirectory
3. **Manifest**: dot maintains `.dot-manifest.json` under
   `~/.local/share/dot/manifest/` for state tracking
4. **Conflict Resolution**: dot fails on conflict by default; the
   `symlinks.overwrite` and `symlinks.backup` configuration keys change this
5. **Error Handling**: dot collects all errors, Stow stops at first

## Configuration Migration

### GNU Stow Configuration

Stow uses command-line options or `.stowrc` file.

### dot Configuration

dot reads YAML, JSON, or TOML from `~/.config/dot/config.yaml`, overridable with
the `DOT_CONFIG` environment variable. Generate a commented default with
`dot config init`. The nearest Stow equivalents:

```yaml
directories:
  package: ~/dotfiles    # stow -d
  target: ~              # stow -t

symlinks:
  mode: relative         # Stow's default link style
```

## Package Structure

### GNU Stow

```
dotfiles/
└── vim/
    ├── .vimrc
    └── .vim/
        └── colors/
```

### dot (Recommended)

```
dotfiles/
└── vim/
    ├── dot-vimrc        # Translates to .vimrc
    └── dot-vim/         # Translates to .vim/
        └── colors/
```

Note: dot reads files whose names already begin with `.`, but the package name
still determines the target subdirectory unless `dotfile.package_name_mapping`
is disabled.

## Common Migration Issues

### Issue: Different Link Paths

**Cause**: Stow and dot may create links differently

**Solution**: Unmanage with Stow before managing with dot

### Issue: Dotfile Translation

**Cause**: dot uses `dot-` prefix, Stow uses `.`

**Solution**: Rename files or use both structures

### Issue: Manifest File

**Cause**: dot writes `.dot-manifest.json` under
`~/.local/share/dot/manifest/` by default, not into the target directory.

**Solution**: No `.gitignore` entry is needed unless `directories.manifest` has
been pointed at a version-controlled directory.

## Compatibility Mode

Use both tools simultaneously (not recommended but possible):

```bash
# Different package sets
stow legacy-packages
dot manage new-packages
```

## Next Steps

- [Quick Start Tutorial](03-quickstart.md): Learn dot workflow
- [Command Reference](05-commands.md): Complete command documentation
- [Configuration Reference](04-configuration.md): Configure dot

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

