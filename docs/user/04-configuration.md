# Configuration Reference

Complete reference for configuring dot.

## Configuration Sources

### Precedence Order (Highest to Lowest)

1. Command-line flags: `--dir`, `--target`, `--ignore`, and so on.
2. Environment variables with the `DOT_` prefix (only the keys listed under Environment Variables).
3. Repository config: `<packageDir>/.config/dot/config.yaml`.
4. User config: `$DOT_CONFIG`, else `$XDG_CONFIG_HOME/dot/config.yaml`, else `~/.config/dot/config.yaml`.
5. Built-in defaults.

Levels 3 and 4 are not layered. If a repository config is found it is used in full and the user
config is not read at all. See [Repository Configuration](repository-config.md).

## Configuration File Locations

On Linux and macOS the file is `$XDG_CONFIG_HOME/dot/config.yaml`, defaulting to
`~/.config/dot/config.yaml`. On Windows it is `%AppData%\dot\config.yaml`, resolved through
`os.UserConfigDir`. Set `DOT_CONFIG` to point at any other path.

There is no `.dotrc` file and no system-wide `/etc/dot/config.yaml`.

## Supported Formats

dot accepts YAML, JSON, and TOML, detected by file extension. The default lookup path is always
`config.yaml`; to use JSON or TOML, set `DOT_CONFIG` to the file explicitly. `dot config init
--format json` writes `config.json` next to the default path, which is not read unless `DOT_CONFIG`
points at it.

All examples below use YAML format.

## Configuration Management

### Creating Configuration

Create a new configuration file with defaults:

```bash
dot config init
```

This creates `~/.config/dot/config.yaml` with default values and documentation comments.

Options:
- `--force`, `-f`: Overwrite existing configuration
- `--yes`, `-y`: Alias for `--force`
- `--format`: Specify format (yaml, json, toml; default yaml)

The generated file covers the commonly edited keys. It omits `ignore.per_package_ignore`,
`ignore.max_file_size`, `ignore.interactive_large_files`, `dotfile.package_name_mapping`,
`output.table_style`, and the `update` and `network` sections; add them by hand if needed.

### Upgrading Configuration

When upgrading dot to a new version, configuration files may need updating to include new fields or
migrate deprecated options.

#### Automatic Upgrade

Upgrade your configuration safely:

```bash
dot config upgrade
```

The upgrade process:

1. **Creates backup**: Timestamped backup saved to `~/.config/dot/backups/`
2. **Merges configuration**: Preserves your customizations while adding new fields
3. **Migrates deprecated fields**: Converts old patterns to new syntax
4. **Validates result**: Ensures upgraded config is valid before saving
5. **Cleans old backups**: Keeps last 5 backups automatically

#### Skipping the Prompt

```bash
dot config upgrade --force
dot config upgrade --yes
```

#### What Gets Preserved

During upgrade, your customizations are preserved:

- Directory paths
- Logging settings
- Symlink modes and options
- Custom ignore patterns
- All user-modified values

#### What Gets Added

New configuration fields from the latest version are added with their defaults.

#### Deprecated Field Migration

**Example**: `ignore.overrides` becomes a negation entry in `ignore.patterns`.

```yaml
# Before upgrade
ignore:
  patterns:
    - "*.log"
  overrides:
    - "important.conf"

# After upgrade
ignore:
  patterns:
    - "*.log"
    - "!important.conf"  # Migrated from overrides
  overrides:  # Left in place, but inert
    - "important.conf"
```

The original `overrides` list is not removed by the migration. It has no effect on matching; only
the negation entry in `patterns` does.

#### Backup Management

Backups are stored in `~/.config/dot/backups/` with format:

```
YYYYMMDD-HHMMSS-config.bak
```

Example: `20241110-153045-config.bak`

The last 5 backups are automatically retained. Older backups are cleaned up after each successful
upgrade.

#### Manual Rollback

If needed, restore from a backup:

```bash
cp ~/.config/dot/backups/20241110-153045-config.bak ~/.config/dot/config.yaml
```

#### Upgrade Header

Upgraded configuration files include a header with upgrade information:

```yaml
# Dot Configuration
# Upgraded on 2024-11-10 15:30:45
# Backup saved to: /home/user/.config/dot/backups/20241110-153045-config.bak
#

directories:
  package: "."
  # ... rest of configuration
```

### Viewing Configuration

View current configuration:

```bash
dot config list
```

Get a specific value:

```bash
dot config get directories.package
```

Show configuration file path:

```bash
dot config path
```

### Modifying Configuration

Set a specific value:

```bash
dot config set directories.package ~/dotfiles
```

`dot config set` and `dot config upgrade` rewrite the whole file, but they read it verbatim first,
so a `~` or `$VAR` you wrote by hand is written back unchanged. A repository config shared across
machines stays portable through both commands.

`dot config list` and `dot config get` report expanded values, which is the quickest way to confirm
where a path actually resolves.

## Configuration Options

Configuration is organised into sections. Keys are given as `section.field`. There is no flat
top-level key namespace: a key written outside its section is parsed, discarded without warning,
and the default silently remains in effect. Confirm every edit with `dot config list`.

### directories

| Key | Type | Default |
|-----|------|---------|
| `directories.package` | string | `.` |
| `directories.target` | string | the user's home directory |
| `directories.manifest` | string | `<XDG_DATA_HOME>/dot/manifest`, usually `$HOME/.local/share/dot/manifest` |

**Path values are expanded when the configuration is loaded.** A leading `~` resolves to your home
directory and `$VAR` or `${VAR}` references resolve from the environment, so `package: ~/.dotfiles`
and `manifest: $XDG_DATA_HOME/dot/manifest` both work and stay portable across machines. Expansion
applies to path-typed values only: `directories.package`, `directories.target`,
`directories.manifest`, `symlinks.backup_dir`, and `logging.file`. Pattern and prefix values such
as `ignore.patterns` and `dotfile.prefix` are matched verbatim, tilde and all.

Three details are worth knowing:

- A variable that is not set is an error naming the variable, unlike a shell, where
  `package: $UNSET_VAR/dotfiles` would quietly become `/dotfiles`.
- Only a leading `~` or `~/` is a home reference. `~alice/.dotfiles` is left alone, and so is a
  tilde anywhere but the start.
- In YAML, an unquoted `target: ~` is null rather than a string; it leaves the key unset and the
  default applies. Write `target: "~"` when you mean the home directory.

Relative paths in `directories.package` are resolved from the working directory. The manifest is a
single JSON file, `.dot-manifest.json`, inside `directories.manifest`; by default that is
`$HOME/.local/share/dot/manifest/.dot-manifest.json`. It tracks installed packages, their links,
and content hashes for incremental updates.

### logging

| Key | Type | Default | Valid values |
|-----|------|---------|--------------|
| `logging.level` | string | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `logging.format` | string | `text` | `text`, `json` |
| `logging.destination` | string | `stderr` | `stderr`, `stdout`, `file` |
| `logging.file` | string | `~/.local/state/dot/dot.log` | any path |

`logging.file` applies only when `logging.destination` is `file`.

### symlinks

| Key | Type | Default | Valid values |
|-----|------|---------|--------------|
| `symlinks.mode` | string | `relative` | `relative`, `absolute` |
| `symlinks.folding` | bool | `true` | |
| `symlinks.overwrite` | bool | `false` | |
| `symlinks.backup` | bool | `false` | |
| `symlinks.backup_suffix` | string | `.bak` | |
| `symlinks.backup_dir` | string | unset | any path |

**`symlinks.mode` currently has no effect.** It is validated, stored, and reported by
`dot config list`, but the CLI never passes it to the planner; created symlinks always point at an
absolute path regardless of this setting.

**`symlinks.folding` currently has no effect.** It is validated and reported, but nothing in the
planner or the pipeline reads it; directory folding behaviour does not change when it is set to
`false`.

Conflict handling is expressed by two booleans, not by a policy name. `symlinks.overwrite: true`
replaces conflicting files. `symlinks.backup: true` moves them aside first. With both false, a
conflict is reported and the operation stops. `symlinks.overwrite` wins if both are set. There is
no `skip` policy.

When backups are taken, files are written to `symlinks.backup_dir`, defaulting to
`<target>/.dot-backup`, which is created on demand. Backup filenames have the form
`<filename>.<8-hex-digest-of-full-source-path>.<YYYYMMDD-HHMMSS>`, which makes them collision-free
across packages. `symlinks.backup_suffix` is not applied to conflict backups.

### ignore

| Key | Type | Default |
|-----|------|---------|
| `ignore.use_defaults` | bool | `true` |
| `ignore.patterns` | []string | `[]` |
| `ignore.overrides` | []string | `[]` (deprecated and inert) |
| `ignore.per_package_ignore` | bool | `true` |
| `ignore.max_file_size` | int (bytes) | `0` (no limit) |
| `ignore.interactive_large_files` | bool | `true` |

Example:

```yaml
ignore:
  use_defaults: true
  patterns:
    - "*.log"
    - "*.tmp"
    - "!important.log"
  per_package_ignore: true
  max_file_size: 104857600  # 100 MB
  interactive_large_files: true
```

**Pattern Types**:
- Glob patterns only: `*` (any sequence of characters, including `/`) and `?` (any single
  character). Bracket expressions and regex metacharacters are escaped and matched literally, so a
  regex such as `/^test.*\.go$/` is treated as a literal string and never matches.
- Negation: a leading `!` un-ignores a previously matched path.

Patterns are anchored and are tested against the full absolute path and, separately, against the
basename. **Patterns containing `/` never match anything.** A pattern such as `.cache/`,
`logs/*.log`, or `.ssh/id_*` matches neither the absolute path nor the basename, so it is silently
inert. Use bare directory names (`.cache`, `node_modules`) and basename patterns (`*.log`).

`ignore.overrides` is deprecated and inert: entries in it are never consulted during matching. Use
negation entries in `ignore.patterns` instead.

```yaml
ignore:
  patterns:
    - ".*"           # Ignore all dotfiles
    - "!.gitignore"  # Except .gitignore
    - "!.gitconfig"  # And .gitconfig
```

`ignore.per_package_ignore` enables reading a `.dotignore` file from the root of each package
directory. Only that one file is read; there is no inheritance from parent or child directories.

`ignore.max_file_size` filters by size in bytes; `0` disables size filtering. Files exceeding the
limit are prompted for in interactive mode and skipped in batch mode. Command-line flags accept
human-readable sizes:

```bash
dot manage mypackage --max-file-size 100MB
```

`ignore.interactive_large_files` controls prompting. When `true` and running in a TTY, dot asks
whether to include each oversized file. When `false`, or under `--batch`, or when not attached to a
TTY, oversized files are skipped silently.

**Built-in Default Patterns** (`ignore.use_defaults: true`):

```
.git
.svn
.hg
.DS_Store
Thumbs.db
desktop.ini
.Trash
.Spotlight-V100
.TemporaryItems
.dotignore
.dotbootstrap.yaml
.gnupg
.ssh/*.pem
.ssh/id_*
.ssh/*_rsa
.ssh/*_ecdsa
.ssh/*_ed25519
.password-store
```

Of these, `.gnupg` and `.password-store` are effective. **The five `.ssh/*` entries contain a path
separator and therefore do not match; SSH keys inside a package are not excluded and will be
symlinked.** Add basename patterns to exclude them:

```yaml
ignore:
  patterns:
    - "id_*"
    - "*_rsa"
    - "*_ecdsa"
    - "*_ed25519"
    - "*.pem"
```

Editor swap and backup files (`*.swp`, `*~`, `.#*`) are not ignored by default; add them to
`ignore.patterns` if you want them excluded.

**See also**: [Ignore System Guide](ignore-system.md) for complete documentation on ignore patterns,
negation, and size filtering.

### dotfile

| Key | Type | Default |
|-----|------|---------|
| `dotfile.translate` | bool | `true` |
| `dotfile.prefix` | string | `dot-` |
| `dotfile.package_name_mapping` | bool | `true` |

With `dotfile.translate` enabled, a leading `dot-` in a source name becomes a leading `.` in the
target name.

`dotfile.package_name_mapping` selects between the two supported repository layouts. Both are
fully supported; pick the one that matches how your repository is organised.

**Name-mapped layout** (`package_name_mapping: true`, the default). The package name determines the
target directory:

- Package `dot-vim` installs into `~/.vim/`
- Package `dot-gnupg` installs into `~/.gnupg/`
- Package `config` installs into `~/config/`

This removes a level of nesting inside the repository: `dot-gnupg/gpg.conf` links to
`~/.gnupg/gpg.conf`, and the package directory holds only the files themselves.

```yaml
dotfile:
  translate: true
  prefix: "dot-"
  package_name_mapping: true
```

**Full-tree layout** (`package_name_mapping: false`). The package name identifies the package only,
and the directory tree inside the package mirrors the target directory, which is how GNU Stow
works:

- `gnupg/dot-gnupg/gpg.conf` links to `~/.gnupg/gpg.conf`
- `vim/dot-vimrc` links to `~/.vimrc`

Choose this layout when you are migrating an existing Stow repository or when a single package
needs to place files in several different directories under the target.

```yaml
dotfile:
  translate: true
  prefix: "dot-"
  package_name_mapping: false
```

### output

| Key | Type | Default | Valid values |
|-----|------|---------|--------------|
| `output.format` | string | `text` | `text`, `json`, `yaml`, `table` |
| `output.color` | string | `auto` | `auto`, `always`, `never` |
| `output.table_style` | string | `default` | |
| `output.progress` | bool | `true` | |
| `output.verbosity` | int | `1` | `0` to `3` |
| `output.width` | int | `0` (auto-detect) | |

`output.verbosity` is validated and reported by `dot config list`, but it does not affect runtime
verbosity: commands take the verbosity level from the `-v` flag count alone, and `-v` does not
increment the configured value. The same applies to `output.format`, `output.progress`, and
`output.width`, which are superseded by the `--format`, `--quiet`, and terminal-detection paths.
`output.table_style` is read when rendering plan tables.

### operations

| Key | Type | Default |
|-----|------|---------|
| `operations.dry_run` | bool | `false` |
| `operations.atomic` | bool | `true` |
| `operations.max_parallel` | int | `0` |

`operations.max_parallel` is accepted and validated but is not consumed by the execution engine; it
does not change concurrency.

### packages

| Key | Type | Default | Valid values |
|-----|------|---------|--------------|
| `packages.sort_by` | string | `name` | `name`, `links`, `date` |
| `packages.auto_discover` | bool | `true` | |
| `packages.validate_names` | bool | `true` | |

This section configures package handling globally. It is not a map of per-package overrides; there
is no per-package configuration mechanism and no `.dotmeta` file.

### doctor

| Key | Type | Default |
|-----|------|---------|
| `doctor.auto_fix` | bool | `false` |
| `doctor.check_manifest` | bool | `true` |
| `doctor.check_broken_links` | bool | `true` |
| `doctor.check_orphaned` | bool | `true` |
| `doctor.check_permissions` | bool | `true` |

Orphan scan breadth is a CLI concern only: `dot doctor --scan-mode` and `dot doctor --max-depth`.
There is no `doctor.orphan_scan_mode` or `doctor.orphan_scan_depth` configuration key.

### update

| Key | Type | Default | Valid values |
|-----|------|---------|--------------|
| `update.check_on_startup` | bool | `true` | |
| `update.check_frequency` | int (hours) | `24` | |
| `update.package_manager` | string | `auto` | `auto`, `brew`, `apt`, `yum`, `pacman`, `dnf`, `zypper`, `manual` |
| `update.repository` | string | `yaklabco/dot` | must be `owner/repo` |
| `update.include_prerelease` | bool | `false` | |

### network

| Key | Type | Default |
|-----|------|---------|
| `network.http_proxy` | string | unset |
| `network.https_proxy` | string | unset |
| `network.no_proxy` | string | unset |
| `network.timeout` | int (seconds) | `10` |
| `network.connect_timeout` | int (seconds) | `5` |
| `network.tls_timeout` | int (seconds) | `5` |

### experimental

| Key | Type | Default |
|-----|------|---------|
| `experimental.parallel` | bool | `false` |
| `experimental.profiling` | bool | `false` |

## Environment Variables

Environment variables use the prefix `DOT_` followed by the section-qualified key with dots replaced
by single underscores, uppercased. For example `symlinks.mode` becomes `DOT_SYMLINKS_MODE`. There is
no double-underscore convention.

Only the following keys are bound:

```
DOT_DIRECTORIES_PACKAGE   DOT_DIRECTORIES_TARGET   DOT_DIRECTORIES_MANIFEST
DOT_LOGGING_LEVEL         DOT_LOGGING_FORMAT       DOT_LOGGING_DESTINATION   DOT_LOGGING_FILE
DOT_SYMLINKS_MODE         DOT_SYMLINKS_FOLDING     DOT_SYMLINKS_OVERWRITE
DOT_SYMLINKS_BACKUP       DOT_SYMLINKS_BACKUP_SUFFIX
DOT_IGNORE_USE_DEFAULTS   DOT_IGNORE_PATTERNS      DOT_IGNORE_OVERRIDES
DOT_IGNORE_PER_PACKAGE_IGNORE   DOT_IGNORE_MAX_FILE_SIZE   DOT_IGNORE_INTERACTIVE_LARGE_FILES
DOT_DOTFILE_TRANSLATE     DOT_DOTFILE_PREFIX
DOT_OUTPUT_FORMAT         DOT_OUTPUT_COLOR         DOT_OUTPUT_PROGRESS
DOT_OUTPUT_VERBOSITY      DOT_OUTPUT_WIDTH
DOT_OPERATIONS_DRY_RUN    DOT_OPERATIONS_ATOMIC    DOT_OPERATIONS_MAX_PARALLEL
DOT_PACKAGES_SORT_BY      DOT_PACKAGES_AUTO_DISCOVER   DOT_PACKAGES_VALIDATE_NAMES
DOT_DOCTOR_AUTO_FIX       DOT_DOCTOR_CHECK_MANIFEST    DOT_DOCTOR_CHECK_BROKEN_LINKS
DOT_DOCTOR_CHECK_ORPHANED DOT_DOCTOR_CHECK_PERMISSIONS
DOT_EXPERIMENTAL_PARALLEL DOT_EXPERIMENTAL_PROFILING
```

`symlinks.backup_dir`, `dotfile.package_name_mapping`, `output.table_style`, and every key under
`update` and `network` have no environment binding; set them in the configuration file.

`DOT_CONFIG` is handled separately and overrides the configuration file path.

Booleans accept `true`/`false` and `1`/`0`. List values are comma-separated.

### Examples

```bash
# Directories
export DOT_DIRECTORIES_PACKAGE=/path/to/dotfiles
export DOT_DIRECTORIES_TARGET=$HOME

# Conflict handling
export DOT_SYMLINKS_OVERWRITE=false
export DOT_SYMLINKS_BACKUP=true

# Ignore patterns
export DOT_IGNORE_PATTERNS="*.log,*.tmp"

# Logging
export DOT_LOGGING_LEVEL=DEBUG
export DOT_LOGGING_FORMAT=json
```

## Complete Configuration Example

`~/.config/dot/config.yaml`:

Paths are written home-relative so the same file works on every machine; absolute paths are equally
valid where a location really is machine-specific.

```yaml
directories:
  package: ~/dotfiles
  target: "~"
  manifest: ~/.local/share/dot/manifest

logging:
  level: INFO
  format: text
  destination: stderr
  file: ~/.local/state/dot/dot.log

symlinks:
  mode: relative
  folding: true
  overwrite: false
  backup: true
  backup_suffix: ".bak"
  backup_dir: ~/.dot-backups

ignore:
  use_defaults: true
  patterns:
    - "*.log"
    - "*.tmp"
    - "*.swp"
    - "node_modules"
    - "id_*"
    - "*_ed25519"
    - "!.gitignore"
  per_package_ignore: true
  max_file_size: 0
  interactive_large_files: true

dotfile:
  translate: true
  prefix: "dot-"
  package_name_mapping: true

output:
  format: text
  color: auto
  table_style: default
  progress: true
  verbosity: 1
  width: 0

operations:
  dry_run: false
  atomic: true
  max_parallel: 0

packages:
  sort_by: name
  auto_discover: true
  validate_names: true

doctor:
  auto_fix: false
  check_manifest: true
  check_broken_links: true
  check_orphaned: true
  check_permissions: true

update:
  check_on_startup: true
  check_frequency: 24
  package_manager: auto
  repository: yaklabco/dot
  include_prerelease: false

network:
  timeout: 10
  connect_timeout: 5
  tls_timeout: 5

experimental:
  parallel: false
  profiling: false
```

### JSON Example

`~/.config/dot/config.json`, used only when `DOT_CONFIG` points at it:

```json
{
  "directories": {
    "package": "~/dotfiles",
    "target": "~"
  },
  "symlinks": {
    "mode": "relative",
    "backup": true
  },
  "ignore": {
    "use_defaults": true,
    "patterns": ["*.log", "*.tmp"]
  }
}
```

### TOML Example

`~/.config/dot/config.toml`, used only when `DOT_CONFIG` points at it:

```toml
[directories]
package = "~/dotfiles"
target = "~"

[symlinks]
mode = "relative"
backup = true

[ignore]
use_defaults = true
patterns = ["*.log", "*.tmp"]
```

## Configuration Management Commands

```bash
# Create the configuration file with defaults
dot config init
dot config init --force            # overwrite an existing file (--yes is an alias)
dot config init --format json      # yaml (default), json, or toml

# Display configuration
dot config list                    # aliases: show, ls
dot config list --format json      # text (default), json, yaml

# Read or write a single key
dot config get directories.package
dot config set directories.package ~/dotfiles

# Show the resolved configuration file path
dot config path

# Migrate an older file to the current format
dot config upgrade                 # --force / --yes skip the confirmation prompt
```

`dot config get` and `dot config set` accept only these keys:
`directories.package`, `directories.target`, `directories.manifest`, `logging.level`,
`logging.format`, `logging.destination`, `symlinks.mode`, `symlinks.backup_suffix`,
`symlinks.backup_dir`, `dotfile.prefix`, `dotfile.translate`, `dotfile.package_name_mapping`,
`output.format`, `output.color`, `packages.sort_by`. Any other key returns `unknown config key`.
Edit the file directly for the remaining settings.

`dot config list` renders the Directories, Logging, Symlinks, Ignore, Dotfile, Output, Operations,
Packages, Doctor, and Experimental sections. The Update and Network sections are not rendered; use
`dot config list --format yaml` to see them.

There is no `dot config validate` and no `dot config unset`. Note that `dot config validate` does
not fail: the argument is swallowed by the parent command, which prints the configuration listing
instead, so the output resembles a successful validation. Do not rely on it. Invalid values are
rejected on load, so running any ordinary command exercises validation.

### Configuration File Location

```bash
dot config path
```

Output:

```
Configuration file: /home/user/.config/dot/config.yaml ✗ (not created)

XDG directories:
  XDG_CONFIG_HOME: /home/user/.config
```

The XDG lines are printed only for variables that are set. Note that this reports the user config
path; it does not account for a repository config, which takes priority when present.

## Configuration Scenarios

### Scenario 1: Multiple Machine Setup

**Laptop** (`~/.config/dot/config.yaml`):

```yaml
directories:
  package: /home/alice/dotfiles
  target: /home/alice
logging:
  level: INFO
```

**Server** (`~/.config/dot/config.yaml`):

```yaml
directories:
  package: /opt/dotfiles
  target: /home/admin
logging:
  level: WARN
```

### Scenario 2: Repository-Local Configuration

A dotfiles repository can carry its own configuration at `<packageDir>/.config/dot/config.yaml`.
It replaces the user configuration entirely rather than merging with it. See
[Repository Configuration](repository-config.md).

### Scenario 3: CI/CD Environment

Non-interactive, scripted usage:

```yaml
directories:
  package: /build/configs
  target: /app
logging:
  level: WARN
  format: json
symlinks:
  overwrite: true
```

Invoke with `--batch`, which implies `--quiet` and disables interactive prompts:

```bash
dot manage --batch mypackage
```

Environment variables:

```bash
export DOT_DIRECTORIES_PACKAGE=/build/configs
export DOT_DIRECTORIES_TARGET=/app
export DOT_SYMLINKS_OVERWRITE=true
export DOT_LOGGING_FORMAT=json
```

## Configuration Best Practices

### 1. Use Version Control

Store configuration in your dotfiles repository at `.config/dot/config.yaml`, where dot will find
it automatically. See [Repository Configuration](repository-config.md).

### 2. Minimal Configuration

Only set values that differ from the defaults:

```yaml
directories:
  package: /home/alice/dotfiles
```

### 3. Prefer Home-Relative Paths

Path values are expanded on load, so `~/dotfiles` and `$HOME/dotfiles` both resolve correctly and
keep the file portable across machines and user accounts. Reach for an absolute path such as
`/home/alice/dotfiles` only when the location genuinely is machine-specific.

### 4. Document Custom Settings

Add comments explaining non-obvious choices:

```yaml
symlinks:
  # Back up rather than fail, because this host has pre-existing configs
  backup: true
  backup_dir: ~/.dot-backups
```

### 5. Verify After Editing

Unknown keys are discarded silently, so confirm that an edit took effect:

```bash
dot config list

# Then preview the effect on a real package
dot manage --dry-run test-package
```

## Troubleshooting Configuration

### Configuration Not Loading

Check the file location:

```bash
dot config path
dot config list
```

If a repository config exists at `<packageDir>/.config/dot/config.yaml`, it is used instead of the
user config, and neither `dot config path` nor `dot config list` reflects that. Inspect the
repository file directly.

### Configuration Errors

Invalid values are rejected when the file is loaded, so any command surfaces them. Common errors:

- Invalid YAML, JSON, or TOML syntax
- Values outside the permitted set, for example `logging.level: TRACE`
- `output.verbosity` outside `0` to `3`
- `update.repository` not in `owner/repo` form
- Negative `operations.max_parallel` or `output.width`

Unknown keys are not an error. A misspelled or wrongly nested key is discarded without a message and
the default remains in effect.

### Unexpected Values

There is no source-attribution output. To narrow down where a value comes from, compare
`dot config list` with the file contents at `dot config path`, then unset any `DOT_*` variables and
re-run. Remember that a repository config at `<packageDir>/.config/dot/config.yaml` replaces the
user config entirely, and that several keys are inert (`symlinks.mode`, `symlinks.folding`,
`operations.max_parallel`, `output.verbosity`, `ignore.overrides`).

## Next Steps

- [Command Reference](05-commands.md): Learn all commands with configuration options
- [Common Workflows](06-workflows.md): See configuration in real-world scenarios
- [Advanced Features](07-advanced.md): Deep dive into ignore patterns and policies

## Navigation

**[Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)
