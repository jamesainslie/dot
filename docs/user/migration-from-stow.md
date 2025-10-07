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
| `stow -d DIR` | `dot --dir DIR` | Specify package directory |
| `stow -t DIR` | `dot --target DIR` | Specify target directory |
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

### Step 2: Test with One Package

```bash
cd ~/dotfiles

# Unstow with Stow
stow -D vim

# Stow with dot
dot manage vim

# Verify identical result
ls -la ~/.vimrc
```

### Step 3: Migrate All Packages

```bash
cd ~/dotfiles

# Unstow all packages
for pkg in */; do
    stow -D "$pkg"
done

# Stow with dot
dot manage $(ls -d */ | tr -d '/')

# Verify
dot status
```

### Step 4: Remove GNU Stow (Optional)

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
7. **Performance**: Parallel execution, directory folding

### Behavioral Differences

1. **Dotfile Translation**: dot uses `dot-` prefix, Stow uses `.` in package
2. **Manifest**: dot maintains `.dot-manifest.json` for state tracking
3. **Conflict Resolution**: dot has configurable policies (fail, backup, overwrite, skip)
4. **Error Handling**: dot collects all errors, Stow stops at first

## Configuration Migration

### GNU Stow Configuration

Stow uses command-line options or `.stowrc` file.

### dot Configuration

dot uses YAML/JSON/TOML configuration:


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

Note: dot also works with Stow's structure (files starting with `.`).

## Common Migration Issues

### Issue: Different Link Paths

**Cause**: Stow and dot may create links differently

**Solution**: Unmanage with Stow before managing with dot

### Issue: Dotfile Translation

**Cause**: dot uses `dot-` prefix, Stow uses `.`

**Solution**: Rename files or use both structures

### Issue: Manifest File

**Cause**: dot creates `.dot-manifest.json` in target directory

**Solution**: Add to `.gitignore` if target is version controlled

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

