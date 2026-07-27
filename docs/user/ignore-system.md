# Ignore System

The dot ignore system provides flexible control over which files are included when managing packages. This system supports global configuration, per-package rules, and size-based filtering.

## Overview

Ignore patterns work similarly to `.gitignore`, allowing you to exclude files from being symlinked. The system supports:

- Default ignore patterns for common system files
- Custom global patterns via configuration or flags
- Per-package `.dotignore` files
- Negation patterns to un-ignore files
- Size-based filtering for large files

## Pattern Syntax

### Glob Patterns

Patterns use glob syntax and are matched against the full absolute path and, separately, against
the basename:

- `*` matches any sequence of characters, including `/`
- `?` matches any single character
- `*.ext` matches all files with that extension, at any depth, via the basename match

Patterns are anchored: the whole path or the whole basename must match. Bracket expressions and
regex metacharacters are escaped and matched literally.

**Patterns containing `/` do not work.** Scanned paths are absolute, so a pattern such as
`.cache/`, `logs/*.log`, or `.ssh/id_*` matches neither the absolute path nor the basename and is
silently ignored. To exclude a directory, use its bare name (`.cache`, `node_modules`); to exclude
files by extension anywhere in the tree, use `*.log`. A trailing slash has no special meaning and
prevents the pattern from matching.

### Negation Patterns

Patterns starting with `!` un-ignore previously ignored files:

```
# Ignore all .log files
*.log
# But include important.log
!important.log
```

**Important**: Order matters. Patterns are processed sequentially, and the last matching pattern wins.

In a `.dotignore` file, only whole-line comments are recognised. A trailing comment on a pattern
line becomes part of the pattern and stops it matching. In the configuration file this does not
apply, because YAML strips comments before dot sees the value.

### Examples

```
# Ignore all temporary files
*.tmp
*.swp
*~

# But keep backup files
!*.bak

# Ignore cache directories (bare names, no trailing slash)
.cache
node_modules

# Ignore large files
*.qcow2
*.vmdk
*.iso
```

## Configuration

### Global Configuration

Configure ignore settings in your `config.yaml`:

```yaml
ignore:
  # Use default patterns (.git, .DS_Store, etc.)
  use_defaults: true
  
  # Additional patterns to ignore
  patterns:
    - "*.log"
    - "*.tmp"
    - "!important.log"  # Negation
  
  # Enable per-package .dotignore files
  per_package_ignore: true
  
  # Maximum file size in bytes (0 = no limit)
  max_file_size: 104857600  # 100MB
  
  # Prompt for large files in interactive mode
  interactive_large_files: true
```

### Command-Line Flags

Override configuration with flags:

```bash
# Add ignore patterns
dot manage mypackage --ignore "*.qcow2" --ignore "*.vmdk"

# Set size limit
dot manage mypackage --max-file-size 100MB

# Disable default patterns
dot manage mypackage --no-defaults

# Disable .dotignore files
dot manage mypackage --no-dotignore

# Batch mode (non-interactive, auto-skip large files)
dot manage mypackage --batch --max-file-size 50MB
```

### Size Format

File sizes support human-readable formats:

- `B` or `b` - Bytes
- `KB`, `K`, `k` - Kilobytes (1024 bytes)
- `MB`, `M`, `m` - Megabytes (1024 KB)
- `GB`, `G`, `g` - Gigabytes (1024 MB)
- `TB`, `T`, `t` - Terabytes (1024 GB)

Examples: `100MB`, `1.5GB`, `500M`, `1024KB`. A space between the number and the unit is tolerated.

## Per-Package .dotignore Files

Create a `.dotignore` file at the root of a package directory:

```
# .dotignore in colima package
# Ignore VM disk images
*.qcow2
*.vmdk

# But keep configuration backups
!*.qcow2.backup

# Ignore logs by extension, not by directory path
*.log

# Comments are supported
# Empty lines are ignored
```

A line beginning with `!!` is rejected as an invalid pattern.

### Scope

Only the `.dotignore` file at the root of the package directory is read. Files in parent
directories and in subdirectories of the package are not consulted; there is no inheritance.

```
packages/
├── .dotignore           # not read
├── vim/
│   ├── .dotignore       # read for package vim
│   └── colors/
│       └── .dotignore   # not read
└── colima/
    └── .dotignore       # read for package colima
```

Patterns shared across packages belong in `ignore.patterns` in the configuration file.

## Default Ignore Patterns

When `use_defaults: true`, these patterns are applied:

- `.git`, `.svn`, `.hg` - version control metadata
- `.DS_Store`, `.Trash`, `.Spotlight-V100`, `.TemporaryItems` - macOS metadata
- `Thumbs.db`, `desktop.ini` - Windows metadata
- `.dotignore`, `.dotbootstrap.yaml` - dot's own metadata files
- `.gnupg`, `.password-store` - key and password stores
- `.ssh/*.pem`, `.ssh/id_*`, `.ssh/*_rsa`, `.ssh/*_ecdsa`, `.ssh/*_ed25519` - SSH keys

The five `.ssh/*` entries contain a path separator and, per the pattern rules above, do not match
anything during a scan. **SSH keys inside a package are not excluded and will be symlinked.** Do
not rely on these defaults; add basename patterns instead:

```yaml
ignore:
  patterns:
    - "id_*"
    - "*_rsa"
    - "*_ecdsa"
    - "*_ed25519"
    - "*.pem"
```

Editor swap and backup files (`*.swp`, `*~`, `.#*`) are not ignored by default.

## Size-Based Filtering

### Interactive Mode

When a file exceeds the size limit in TTY mode with `interactive_large_files: true`:

```
Large file detected:
  Path: .colima/default/diffdisk.qcow2
  Size: 2.5 GB (limit: 100.0 MB)

Options:
  i) Include this file
  s) Skip this file
  a) Skip all large files
Choice [s]:
```

### Batch Mode

In batch mode (`--batch` flag or non-TTY), large files are automatically skipped silently.

## Common Use Cases

### Ignore Virtual Machine Images

For virtualization tools like Colima:

```yaml
# In config.yaml
ignore:
  patterns:
    - "*.qcow2"
    - "*.vmdk"
    - "*.vdi"
  max_file_size: 104857600  # 100MB
```

Or in package `.dotignore`:

```
# .dotfiles/colima/.dotignore
*.qcow2
*.vmdk
diffdisk.*
```

### Ignore Cache and Build Artifacts

```yaml
ignore:
  patterns:
    - ".cache"
    - "node_modules"
    - "*.pyc"
    - "__pycache__"
    - "*.o"
    - "*.so"
```

### Development with Selective Inclusion

```
# Ignore all logs
*.log

# But keep important logs
!error.log
!access.log

# Ignore compiled files
*.o
*.so
*.a

# But keep specific libraries
!libimportant.so
```

## Precedence Order

Patterns are concatenated in this order and the last matching pattern decides:

1. Default ignore patterns, unless `ignore.use_defaults: false` or `--no-defaults`
2. `ignore.patterns` from the configuration file
3. Command-line `--ignore` flags
4. The package's own `.dotignore`, unless `ignore.per_package_ignore: false` or `--no-dotignore`

A package `.dotignore` therefore has the highest priority and can un-ignore, with `!`, anything set
by the configuration file or by `--ignore`.

## Best Practices

1. **Use Defaults**: Keep `use_defaults: true` to automatically ignore system files
2. **Size Limits**: Set reasonable limits for package types (e.g., 100MB for configs, 1GB for application data)
3. **Per-Package Rules**: Use `.dotignore` for package-specific exclusions
4. **Negation Sparingly**: Use negation patterns when you need exceptions, but keep rules simple
5. **Document Patterns**: Add comments to explain non-obvious patterns
6. **Test Patterns**: Use `dot manage --dry-run` to preview what will be managed

## Troubleshooting

### Files Not Being Ignored

1. Check that the pattern contains no `/`. Patterns with a path separator never match.
2. Remove any trailing slash: `.cache/` does not match, `.cache` does.
3. Verify pattern order. Later patterns override earlier ones, and the package `.dotignore` is
   applied last.
4. Check whether a package `.dotignore` un-ignores the file with `!`.

### Files Unexpectedly Ignored

1. Check for overly broad patterns
2. Look for default patterns that might match
3. Check the package's own `.dotignore` (parent directories are not read)
4. Use negation to explicitly include files

### Size Filtering Issues

1. Verify the unit is one of B, KB, MB, GB, TB (case-insensitive; `K`, `M`, `G`, `T` also
   accepted). A space between the number and the unit is tolerated.
2. Check if `max_file_size` is set to 0 (disabled)
3. Ensure `interactive_large_files` matches your use case
4. Use `--batch` for non-interactive environments

## Environment Variables

All ignore configuration can be set via environment variables:

```bash
export DOT_IGNORE_USE_DEFAULTS=true
export DOT_IGNORE_PATTERNS="*.log,*.tmp"
export DOT_IGNORE_MAX_FILE_SIZE=104857600
export DOT_IGNORE_PER_PACKAGE_IGNORE=true
export DOT_IGNORE_INTERACTIVE_LARGE_FILES=true
```

## See Also

- [Configuration Guide](04-configuration.md)
- [Commands Reference](05-commands.md)
- [Advanced Usage](07-advanced.md)



