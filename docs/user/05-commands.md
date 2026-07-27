# Command Reference

Complete reference for all dot commands and options.

## Command Structure

```bash
dot [global-options] <command> [command-options] [arguments]
```

**Components**:
- `global-options`: Flags affecting all commands
- `command`: Operation to perform (manage, status, etc.)
- `command-options`: Flags specific to command
- `arguments`: Command-specific arguments (package names, files, etc.)

## Global Options

Available for all commands.

### Directory Options

#### `-d, --dir PATH`

Specify package directory (source directory containing packages).

**Default**: Current directory  
**Example**:
```bash
dot --dir ~/dotfiles manage dot-vim
dot -d /opt/configs status
```

#### `-t, --target PATH`

Specify target directory (destination for symlinks).

**Default**: `$HOME`  
**Example**:
```bash
dot --target ~ manage dot-vim
dot -t /home/user unmanage dot-zsh
```

### Execution Mode Options

#### `-n, --dry-run`

Preview operations without applying changes.

**Example**:
```bash
dot --dry-run manage dot-vim
dot -n unmanage dot-zsh
```

Shows planned operations with no filesystem modifications.

#### `-q, --quiet`

Suppress non-error output.

**Example**:
```bash
dot --quiet manage dot-vim
```

Only errors printed. Useful for scripting.

### Verbosity Options

#### `-v, --verbose`

Increase verbosity (repeatable).

**Levels**:
- No flag: Errors and warnings
- `-v`: Info messages
- `-vv`: Debug messages
- `-vvv`: Trace messages

**Example**:
```bash
dot -v manage dot-vim  # Info level
dot -vv status         # Debug level
dot -vvv remanage dot-zsh  # Trace level
```

### Output Format Options

#### `--log-json`

Output logs in JSON format.

**Example**:
```bash
dot --log-json manage dot-vim
```

JSON output for log aggregation and parsing.

#### `--no-color`

Disable color output globally.

**Example**:
```bash
dot --no-color manage dot-vim
```

`status`, `list`, and `doctor` additionally accept a per-command
`--color WHEN` flag (`auto`, `always`, `never`; default `auto`). `--color` is
not accepted by other commands.

### Ignore Options

#### `--ignore PATTERN`

Add ignore pattern (repeatable). Patterns are globs. Prefix a pattern with `!`
to re-include a file that an earlier pattern excluded.

**Example**:
```bash
dot --ignore "*.log" manage dot-vim
dot --ignore "*.log" --ignore "!important.log" manage dot-zsh
```

### Other Global Options

- `--backup-dir PATH`: Directory for conflict backups; default `<target>/.dot-backup`
- `--batch`: Non-interactive mode for scripting; implies `--quiet`
- `--max-file-size SIZE`: Skip files larger than this (e.g. `100MB`); `0` or empty means no limit
- `--no-defaults`: Disable the built-in ignore patterns
- `--no-dotignore`: Ignore per-package `.dotignore` files
- `--cpu-profile FILE`, `--mem-profile FILE`, `--pprof ADDR`: Diagnostics

### Conflict Resolution

There is no conflict-resolution flag. The policy is read from the
configuration file:

```yaml
symlinks:
  overwrite: false  # replace conflicting files
  backup: false     # move conflicting files to the backup directory
```

Precedence is `overwrite`, then `backup`, then fail (the default). The `skip`
policy exists in the planner but is not reachable from the CLI or the
configuration file.

When the backup policy applies, the conflicting file is moved to
`<target>/.dot-backup` (override with `--backup-dir`) under the name
`<filename>.<hash>.<timestamp>`, for example
`.vimrc.3f9a1c07.20260727-121603`. The hash is derived from the full source
path, so files sharing a basename across directories never collide. The backup
directory is created on demand. The `symlinks.backup_suffix` setting is not
used by the backup policy.

## Package Management Commands

### clone

Clone a dotfiles repository and install packages.

**Synopsis**:
```bash
dot clone [options] REPOSITORY_URL
```

**Arguments**:
- `REPOSITORY_URL`: Git repository URL (HTTPS or SSH format)

**Options**:
- `--profile NAME`: Installation profile from bootstrap config
- `--interactive`: Interactively select packages to install
- `--force`: Overwrite package directory if exists
- `--branch NAME`: Branch to clone (defaults to repository default)

All global options also apply.

**Description**:

The clone command provides single-command setup for new machines. It clones a dotfiles repository and installs packages based on optional bootstrap configuration.

Like `git clone`, the repository is cloned into a subdirectory named after the repository. For example, `dot clone https://github.com/user/my-dotfiles` creates a `my-dotfiles` directory in the current location. Use `--dir` to specify a different target directory.

**Workflow**:
1. Determines target directory (from repository name or `--dir` flag)
2. Validates target directory is empty (unless `--force`)
3. Clones repository to target directory
4. Loads optional `.dotbootstrap.yaml` configuration
5. Selects packages (via profile, interactively, or all)
6. Filters packages by current platform
7. Installs selected packages via `manage` command
8. Updates manifest with repository tracking information

**Authentication**:

Authentication is automatically resolved in priority order:
1. `GITHUB_TOKEN` environment variable (GitHub repositories)
2. `GIT_TOKEN` environment variable (general git repositories)
3. SSH keys in `~/.ssh/` directory (for SSH URLs like `git@github.com:user/repo.git`)
4. GitHub CLI (`gh`) authenticated session (for HTTPS GitHub repositories)
5. No authentication (public repositories only)

If you've authenticated with `gh auth login`, dot will automatically use your GitHub CLI credentials when cloning private GitHub repositories via HTTPS. For SSH URLs, SSH keys are preferred as expected.

**Bootstrap Configuration**:

If `.dotbootstrap.yaml` exists in repository root, it defines:
- Available packages with platform requirements
- Named installation profiles
- Default profile and conflict resolution policies

Without bootstrap configuration, all discovered packages are offered for installation.

See [Bootstrap Configuration Specification](bootstrap-config-spec.md) for complete documentation.

**Examples**:

```bash
# Clone and install all packages (creates ./dotfiles directory)
dot clone https://github.com/user/dotfiles

# Clone creates ./my-dotfiles directory based on repo name
dot clone https://github.com/user/my-dotfiles

# Clone specific branch
dot clone https://github.com/user/dotfiles --branch develop

# Use named profile from bootstrap config
dot clone https://github.com/user/dotfiles --profile minimal

# Force interactive selection
dot clone https://github.com/user/dotfiles --interactive

# Clone to specific directory (overrides default behavior)
dot clone --dir ~/my-packages https://github.com/user/dotfiles

# Overwrite existing package directory
dot clone --force https://github.com/user/dotfiles

# Clone via SSH
dot clone git@github.com:user/dotfiles.git

# Preview what would be installed
dot --dry-run clone https://github.com/user/dotfiles
```

**Error Handling**:

Common errors and solutions:

- **Package directory not empty**: Use `--force` to overwrite
- **Authentication failed**: Set `GITHUB_TOKEN` or configure SSH keys
- **Clone failed**: Verify URL, network connection, and repository access
- **Bootstrap invalid**: Check `.dotbootstrap.yaml` syntax
- **Profile not found**: Verify profile exists in bootstrap config

**Platform Filtering**:

Packages in bootstrap configuration can specify target platforms:

```yaml
packages:
  - name: dot-vim           # All platforms
  - name: dot-linux-config  # Linux only
    platform: [linux]
  - name: dot-macos-config  # macOS only
    platform: [darwin]
```

Platform filtering is automatic based on current system.

**Related Commands**:
- `manage`: Manually install additional packages after cloning
- `status`: Check installation status and repository information
- `unmanage`: Remove installed packages
- `clone bootstrap`: Generate bootstrap configuration from installation

### clone bootstrap

Generate bootstrap configuration from existing dotfiles installation.

**Synopsis**:
```bash
dot clone bootstrap [options]
```

**Options**:
- `-o, --output PATH`: Output file path (default: `.dotbootstrap.yaml` in package directory)
- `--dry-run`: Print configuration to stdout instead of writing file
- `--from-manifest`: Only include packages currently in manifest
- `--conflict-policy POLICY`: Default conflict policy (backup, fail, overwrite, skip)
- `--force`: Overwrite existing bootstrap file

All global options also apply.

**Description**:

The clone bootstrap subcommand generates a `.dotbootstrap.yaml` configuration file from your current dotfiles installation. This allows you to create a bootstrap configuration for an existing repository, enabling others to clone your dotfiles with predefined package selections and profiles.

The command discovers all packages in the package directory and creates a bootstrap configuration with sensible defaults. The generated file includes helpful comments and example structures that you should review and customize before committing.

**Generated Configuration**:

The output includes:
- All discovered packages marked as `required: false`
- Default conflict resolution policy
- Example profile structures with comments
- Helpful guidance for customization
- Timestamps and documentation links

**Examples**:

```bash
# Generate bootstrap config in package directory
dot clone bootstrap

# Specify custom output location
dot clone bootstrap --output ~/dotfiles/.dotbootstrap.yaml

# Preview without writing file
dot clone bootstrap --dry-run

# Only include packages from manifest
dot clone bootstrap --from-manifest

# Set default conflict policy
dot clone bootstrap --conflict-policy backup

# Overwrite existing file
dot clone bootstrap --force
```

**Workflow**:

1. Run command in dotfiles repository
2. Review generated `.dotbootstrap.yaml`
3. Customize package requirements and platform restrictions
4. Define installation profiles for different use cases
5. Commit configuration to repository
6. Others can clone with `--profile` flag

**Error Handling**:

Common errors and solutions:

- **No packages found**: Ensure package directory contains subdirectories
- **Bootstrap file exists**: Use `--force` to overwrite
- **Invalid conflict policy**: Use backup, fail, overwrite, or skip
- **Package directory not found**: Check global `--dir` flag

**Customization After Generation**:

After generating the configuration, customize it by:

1. **Mark required packages**: Set `required: true` for essential packages
2. **Add platform restrictions**: Specify `platform: [linux, darwin]` as needed
3. **Create profiles**: Define named sets like `minimal`, `work`, `full`
4. **Set conflict policies**: Override default per-package if needed
5. **Add descriptions**: Document profiles for other users

Example customization:

```yaml
version: "1.0"

defaults:
  on_conflict: backup
  profile: minimal

packages:
  - name: vim
    required: true  # Essential package
  - name: zsh
    required: true
  - name: macos-config
    required: false
    platform: [darwin]  # macOS only

profiles:
  minimal:
    description: Basic shell and editor
    packages:
      - vim
      - zsh
  full:
    description: Complete development environment
    packages:
      - vim
      - zsh
      - git
      - tmux
```

**Related Commands**:
- `clone`: Clone repository using bootstrap configuration
- `status`: View currently installed packages
- `list`: List all installed packages

See [Bootstrap Configuration Specification](bootstrap-config-spec.md) for complete configuration reference.

### manage

Install packages by creating symlinks.

**Synopsis**:
```bash
dot manage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to install

**Options**: All global options

**Path Mapping**:

The package name becomes a subdirectory of the target. A name beginning with
`dot-` is translated to a leading dot; a plain name is kept verbatim.

| Package | Target |
|---------|--------|
| `dot-ssh/config` | `~/.ssh/config` |
| `vim/dot-vimrc` | `~/vim/.vimrc` |
| `dot-vim/dot-vimrc` | `~/.vim/.vimrc` |
| `scripts/hello.sh` | `~/scripts/hello.sh` |

To place files directly in the target root (the pre-0.4 layout), set
`dotfile.package_name_mapping: false` in the configuration file. Package name
mapping is enabled by default.

Symlinks are always absolute. The `symlinks.mode` configuration key is
accepted but is not connected to the planner, so it has no effect.

**Examples**:
```bash
# Single package
dot manage dot-vim

# Multiple packages
dot manage dot-vim dot-zsh dot-tmux dot-git

# Preview
dot --dry-run manage test-package

# Different directories
dot --dir ~/dotfiles --target ~ manage dot-vim
```

**Behavior**:
1. Scans package directories
2. Computes desired symlink state
3. Detects conflicts
4. Resolves conflicts per policy
5. Creates symlinks with dependency ordering
6. Updates manifest

**Example Output**:
```
✓ Managed 2 packages
```

In `--dry-run` mode the plan is rendered instead of executed:

```
Dry run mode - no changes will be applied

Plan:
  + Create directory: /home/user/.vim
  + Create directory: /home/user/.vim/colors
  + Create symlink: /home/user/.vim/colors/theme.vim -> /home/user/dotfiles/dot-vim/colors/theme.vim
  + Create symlink: /home/user/.vim/.vimrc -> /home/user/dotfiles/dot-vim/dot-vimrc

Summary:
  Directories: 2
  Symlinks: 2
  Conflicts: 0
```

When a package is already installed and unchanged, the command reports that
there is nothing to do and exits successfully.

**Exit Codes**:
- `0`: Success
- `1`: Error during operation

### unmanage

Remove packages by deleting symlinks, with optional restoration or cleanup.

**Synopsis**:
```bash
dot unmanage [options] PACKAGE [PACKAGE...]
dot unmanage --all [options]
```

**Arguments**:
- `PACKAGE`: One or more package names to remove

**Options**:
- All global options
- `--all`: Remove all managed packages
- `--yes, --force`: Skip confirmation prompt (for use with --all)
- `--purge`: Delete package directory after removing links
- `--no-restore`: Skip restoring adopted packages to target
- `--cleanup`: Remove orphaned packages from manifest only

**Examples**:
```bash
# Remove managed package (removes links only)
dot unmanage dot-vim

# Remove adopted package (restores files to target, keeps in package)
dot unmanage dot-ssh

# Remove with purge (deletes package directory)
dot unmanage --purge dot-vim

# Remove without restoring (for adopted packages)
dot unmanage --no-restore dot-ssh

# Clean up orphaned packages
dot unmanage --cleanup dot-old-package

# Remove all packages (with confirmation prompt)
dot unmanage --all

# Remove all packages without confirmation
dot unmanage --all --yes

# Remove all packages and delete directories
dot unmanage --all --purge --force

# Preview removing all packages
dot --dry-run unmanage --all
```

**Behavior**:

For **managed packages** (created with `dot manage`):
1. Removes symlinks
2. Cleans up empty directories
3. Removes from manifest
4. Package directory preserved (unless `--purge`)

For **adopted packages** (created with `dot adopt`):
1. Removes symlinks
2. **Copies files back to target** (unless `--no-restore`)
3. Removes from manifest  
4. Package directory preserved (unless `--purge`)

**Restoration for Adopted Packages**:

By default, `unmanage` **restores** adopted files to their original locations:

```bash
# Before unmanage:
~/.ssh -> ~/dotfiles/dot-ssh  # Symlink
~/dotfiles/dot-ssh/config     # Files in package

# After: dot unmanage dot-ssh
~/.ssh/config                 # Files restored (copied back)
~/dotfiles/dot-ssh/config     # Package preserved as backup
```

Files are **copied** (not moved), so they remain in the package as a backup.

**Remove All Packages**:

Use `--all` to remove all managed packages at once:

```bash
# With confirmation prompt
dot unmanage --all

# Skip confirmation
dot unmanage --all --yes
```

When using `--all`:
1. Shows summary of all packages to be removed
2. Displays operation type for each (remove, restore, purge)
3. Requires confirmation unless `--yes`, `--force`, or `--dry-run` specified
4. Applies same behavior as individual unmanage (restore adopted by default)

This is useful for completely resetting your system to pre-dot state.

**Cleanup Mode**:

Use `--cleanup` to remove orphaned packages (missing links or directories):

```bash
dot unmanage --cleanup old-package
```

Only updates manifest, no filesystem operations.

**Example Output**:
```
✓ Unmanaged and restored 1 package
```

**Safety Guarantees**:
- Only removes links pointing to package directory
- Preserves non-managed files
- Deletion is verified before it happens: dot refuses to remove a regular file
  or a symlink that now points elsewhere, reporting
  `refusing to delete <path>: <reason>`
- If a rollback cannot restore a path, the failure is reported rather than
  silently ignored
- Adopted packages restored by default (preserves your data)
- Confirmation required for `--all` operations

**Exit Codes**:
- `0`: Success
- `1`: Error during operation

### remanage

Update packages efficiently using incremental detection and restore missing symlinks.

**Synopsis**:
```bash
dot remanage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to update

**Options**: All global options

**Examples**:
```bash
# Single package
dot remanage dot-vim

# Multiple packages
dot remanage dot-vim dot-zsh dot-tmux

# Preview changes
dot --dry-run remanage dot-vim

# Verbose output to see detection details
dot -vv remanage dot-zsh
```

**Behavior**:
1. Loads manifest with previous state
2. Computes content hashes for package directories
3. Verifies all symlinks still exist
4. Compares with stored hashes and link states
5. Processes changed or broken packages
6. Updates manifest while preserving package source type

**Incremental Detection**:
- **Unchanged packages with valid links**: Skipped entirely (no-op)
- **Changed packages**: Unmanaged then managed (full update)
- **Packages with missing links**: Recreates missing symlinks
- **New packages**: Managed
- **Adopted packages**: Preserves adoption structure (single directory symlink)

**Missing Link Detection**:

If symlinks were accidentally deleted, `remanage` automatically recreates them:

```bash
# Symlink accidentally deleted
rm ~/.vim/.vimrc

# Check status
dot doctor
# ✗ Errors detected

# Recreate missing link
dot remanage dot-vim
# ✓ Remanaged 1 package

# Link restored
ls -la ~/.vim/.vimrc
# /home/user/.vim/.vimrc -> /home/user/dotfiles/dot-vim/dot-vimrc
```

The command reports only a total; there are no per-package changed or skipped
lines.

**Package Source Preservation**:

`remanage` preserves the original package type:
- **Adopted packages**: Maintains single directory symlink structure
- **Managed packages**: Maintains individual file symlinks

This ensures adopted directories aren't converted to managed packages.

**Exit Codes**:
- `0`: Success, changes applied or no changes needed
- `1`: Error during operation

### adopt

Move existing files or directories into a package and create symlinks.

**Synopsis**:
```bash
# Interactive mode (no arguments; requires a terminal)
dot adopt [options]

# Auto-naming mode (exactly one file or directory)
dot adopt [options] FILE|DIRECTORY

# Explicit package mode (two or more arguments)
dot adopt [options] PACKAGE FILE|DIRECTORY [FILE|DIRECTORY...]
```

**Arguments**:
- `FILE|DIRECTORY`: Path to file or directory to adopt
- `PACKAGE`: Package name, required whenever more than one path is given

**Options**:
- `--scan-dirs DIR,...`: Additional directories to scan (interactive mode only)
- `--exclude-dirs DIR,...`: Directories to skip during discovery (interactive mode only)
- `--max-size SIZE`: Largest file to offer for adoption (interactive mode only); default `10M`
- All global options

**Interactive Mode**: run `dot adopt` with no arguments to discover and select
dotfiles interactively. This requires a terminal; in a pipe or script it exits
with an error directing you to the argument forms.

**Modes**:

#### Auto-Naming Mode
Exactly one argument; the package name is derived from the path:
```bash
dot adopt .vimrc      # Creates package: dot-vimrc
dot adopt .ssh        # Creates package: dot-ssh
dot adopt .config     # Creates package: dot-config
```

#### Multiple Files

Two or more arguments means the first is the package name and the rest are
files. Shell globs expand before dot runs, so a glob must be preceded by an
explicit package name:

```bash
dot adopt dot-git .git*     # Package dot-git, all .git* files
dot adopt dot-zsh .zsh*     # Package dot-zsh, all .zsh* files
```

Without a leading package name, `dot adopt .git*` treats the first expanded
file as the package name. There is no common-prefix derivation.

#### Explicit Package Mode
Specify package name explicitly:
```bash
dot adopt dot-vim .vimrc .vim/      # Package: dot-vim
dot adopt configs .config/ .local/  # Package: configs
```

#### Path Resolution

File paths are resolved based on the following rules:

1. **Absolute paths** (`/etc/config`, `~/file`): Used as-is
2. **Explicit relative paths** (`./file`, `../dir`): Resolved from current working directory
3. **Bare paths** (`file`, `.config/nvim`): Resolved from target directory (default: `$HOME`)

**Examples:**

```bash
# From ~/.config directory:
cd ~/.config
dot adopt ado-cli ./ado-cli        # Adopts ~/.config/ado-cli (from pwd)
dot adopt fish .config/fish        # Adopts $HOME/.config/fish (from target)

# Using explicit pwd paths:
cd ~/.config
dot adopt nvim ./nvim              # Adopts ~/.config/nvim
dot adopt configs ./fish ./nvim    # Adopts multiple from pwd

# Backward compatible - bare paths from target:
cd /tmp
dot adopt .vimrc                   # Adopts $HOME/.vimrc (not /tmp/.vimrc)
```

**Note:** The `./` prefix explicitly means "from current directory", while bare paths maintain backward compatibility by resolving from the target directory.

**Directory Adoption**:

When adopting a directory, `dot` creates a **flat structure** in the package with the directory contents at the package root:

```bash
# Before: ~/.ssh/ with files
~/.ssh/
├── config
├── id_rsa
└── known_hosts

# After: dot adopt .ssh
~/dotfiles/dot-ssh/       # Package root contains directory contents
├── config
├── id_rsa
└── known_hosts

~/.ssh -> ~/dotfiles/dot-ssh  # Single symlink to package root
```

**File Adoption**:

Single files are placed in a package directory with dotfile translation:

```bash
# Before: ~/.vimrc

# After: dot adopt .vimrc
~/dotfiles/dot-vimrc/
└── dot-vimrc

~/.vimrc -> ~/dotfiles/dot-vimrc/dot-vimrc
```

**Dotfile Translation**:

Dotfiles (starting with `.`) have the dot replaced with `dot-` prefix:
- `.vimrc` → `dot-vimrc`
- `.ssh` → `dot-ssh`
- `.config` → `dot-config`
- Nested: `.config/nvim/init.vim` → `dot-config/nvim/init.vim`

**Behavior**:
1. Determines adoption mode (interactive, auto-naming, or explicit)
2. Derives or uses provided package name
3. Creates package directory structure
4. Moves files/directories to package (applying dotfile translation)
5. Creates symlinks in original locations
6. Records package as "adopted" in manifest

**Exit Codes**:
- `0`: Success
- `1`: Error during operation

## Query Commands

### status

Display installation status for packages.

**Synopsis**:
```bash
dot status [options] [PACKAGE...]
```

**Arguments**:
- `PACKAGE` (optional): Specific packages to query (default: all)

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`); default `text`
- `--color WHEN`: Colorize output (`auto`, `always`, `never`); default `auto`
- All global options

**Examples**:
```bash
# All packages
dot status

# Specific packages
dot status dot-vim dot-zsh

# JSON output
dot status --format json

# YAML output
dot status --format yaml

# Table format
dot status --format table

# Combine with verbosity
dot -v status dot-vim
```

**Output Fields**:
- Package name
- Link count
- Installation time, expressed as an interval
- List of link paths relative to the target directory

**Example Output (text)**:
```
dot-vim
  Links: 2
  Installed: just now
  Files:
    .vim/.vimrc
    .vim/colors/theme.vim
```

Text output lists at most five files per package, followed by
`... and N more`.

**Example Output (JSON)**:
```json
{
  "packages": [
    {
      "name": "dot-vim",
      "source": "managed",
      "installed_at": "2026-07-27T12:25:00Z",
      "link_count": 2,
      "links": [
        ".vim/.vimrc",
        ".vim/colors/theme.vim"
      ],
      "target_dir": "/home/user",
      "package_dir": "/home/user/dotfiles/dot-vim",
      "is_healthy": false,
      "issue_type": "missing links"
    }
  ]
}
```

JSON output is an object with a `packages` array, not a bare array. Scripts
must address it as `.packages[]`.

Naming a package that is not installed produces an error.

**Exit Codes**:
- `0`: Success
- `1`: Error querying status, or a named package was not found

### doctor

Validate installation health and detect issues.

**Synopsis**:
```bash
dot doctor [options]
```

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`); default `text`
- `--color WHEN`: Color output (`auto`, `always`, `never`); default `auto`
- `--scan-mode MODE`: Orphaned link detection (`off`, `scoped`, `deep`); default `scoped`
- `--max-depth N`: Maximum recursion depth for deep scan; default `10`
- `--mode MODE`: Diagnostic mode (`fast`, `deep`); default `fast`
- `--detailed`: Expand each issue with type, path, and suggestion
- `--triage`: Interactively categorize orphaned symlinks
- `--auto-ignore`: In triage, auto-ignore high-confidence categories
- All global options

`--format table` is rendered identically to `--format text`.

**Subcommands**:
```bash
dot doctor ignore [PATH] --reason <text>     # suppress one link
dot doctor ignore --pattern <glob> --reason <text>
dot doctor unignore [PATH]
dot doctor unignore --pattern <glob>
dot doctor ignores                           # list ignored links and patterns
```

`ignore` and `unignore` accept exactly one of `PATH` or `--pattern`. An invalid
`--scan-mode` fails with `invalid scan-mode: <value> (must be off, scoped, or deep)`.

**Scan Modes**:

- **off**: Skip orphaned link detection (fastest, ~50ms)
  - Only checks managed links from manifest
  - Use for quick health checks or in automated scripts

- **scoped** (default): Scan directories containing managed links (fast, ~600ms)
  - Limited to depth 3 to avoid deep recursion
  - Skips common large directories (Library, node_modules, .docker, etc.)
  - Parallel scanning using multiple CPU cores
  - Recommended for regular health checks

- **deep**: Full recursive scan of target directory (thorough, ~3-5s)
  - Scans entire home directory up to depth 10
  - Still skips large cache/build directories
  - Use when investigating orphaned links from other tools
  - Significantly slower but more comprehensive

**Performance Notes**:

The doctor command has been optimized for speed:
- Parallel directory scanning using worker pools
- DirEntry type checking (no extra syscalls for regular files)
- Intelligent skip patterns for common large directories
- Depth limits to prevent excessive recursion

For systems with many symlinks (10,000+), use `scoped` mode for regular checks
and `deep` mode only when investigating specific issues.

**Examples**:
```bash
# Basic health check (scoped scan - default, fast)
dot doctor

# Quick check without orphan detection (fastest)
dot doctor --scan-mode=off

# Deep scan for comprehensive orphan detection
dot doctor --scan-mode=deep

# Expanded issue detail
dot doctor --detailed

# JSON output for scripting
dot doctor --format json

# Interactive triage of orphaned symlinks
dot doctor --triage

# Force color output even when piped
dot doctor --color=always | less -R

# Suppress a known orphan permanently
dot doctor ignore ~/.bashrc --reason "managed by nix"
dot doctor ignores
```

**Checks Performed**:
1. **Broken symlinks**: Links pointing to non-existent targets
2. **Orphaned links**: Links not in manifest but pointing to package directory
3. **Wrong links**: Links in manifest but pointing elsewhere
4. **Manifest consistency**: Manifest matches filesystem state
5. **Permission issues**: Files with incorrect permissions
6. **Circular dependencies**: Circular symlink chains

**Example Output (healthy)**:
```
✓ Healthy
  • 4 total links (4 managed, 0 broken, 0 orphaned)
  No issues found
```

**Example Output (issues, `--detailed`)**:
```
✗ Errors detected
============================================================

Statistics:
  Total links: 2
  Managed links: 2
  Broken links: 1
  Orphaned links: 0

✗ Errors (1):

  ✗ [broken_link] Link does not exist
     Path: .vim/.vimrc
     Suggestion: Run 'dot remanage dot-vim' to restore link

------------------------------------------------------------
Summary: 1 total issues
```

There is no repair flag. Broken and missing links are fixed by remanaging the
affected packages; orphaned links are resolved with `dot adopt` or suppressed
with `dot doctor ignore`.

**Exit Codes**:
- `0`: Healthy (no issues found)
- `1`: Warnings detected (e.g., orphaned links)
- `2`: Errors detected (e.g., broken links)

### list

Show installed package inventory with health status indicators.

**Synopsis**:
```bash
dot list [options]
```

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`); default `text`
- `--sort FIELD`: Sort by field (`name`, `links`, `date`); default `name`
- `--color WHEN`: Colorize output (`auto`, `always`, `never`); default `auto`
- `--show-target`: Include the target directory column
- All global options

`--sort` has no single-letter shorthand.

**Health Status**:

Each package is automatically checked for health when listing. A package is considered healthy if all its managed symlinks exist and point to their correct targets. Health indicators:

- `✓` (green checkmark): All symlinks are valid
- `✗` (red X): Package has issues with specific type indicated (e.g., "broken links", "wrong target", "missing links")

The health check is fast and only validates symlink existence and targets, without the full diagnostic scan that `doctor` performs.

**Examples**:
```bash
# List all packages with health status
dot list

# Sort by link count
dot list --sort links

# Sort by installation date
dot list --sort date

# JSON output (includes health status)
dot list --format json

# Table format with health column
dot list --format table

# Combine sorting and format
dot list --sort links --format table
```

**Example Output (text)**:
```
Packages: 2 packages in /home/user/dotfiles

✓  dot-vim  (2 links)  installed just now
✓  dot-zsh  (2 links)  installed just now

2 healthy
```

**Example Output (table)**:
```
Package directory: /home/user/dotfiles
Target directory:  /home/user
Manifest:          /home/user/.local/share/dot/manifest

╭────────────────┬─────────┬───────┬───────────╮
│     HEALTH     │ PACKAGE │ LINKS │ INSTALLED │
├────────────────┼─────────┼───────┼───────────┤
│ ✗ broken links │ dot-vim │ 2     │ just now  │
│ ✓              │ dot-zsh │ 2     │ just now  │
╰────────────────┴─────────┴───────┴───────────╯
1 healthy, 1 unhealthy
```

Set `output.table_style: simple` in the configuration file for a borderless
table.

**Example Output (JSON)**:
```json
{
  "packages": [
    {
      "name": "dot-vim",
      "source": "managed",
      "installed_at": "2026-07-27T12:25:00Z",
      "link_count": 2,
      "links": [
        ".vim/.vimrc",
        ".vim/colors/theme.vim"
      ],
      "target_dir": "/home/user",
      "package_dir": "/home/user/dotfiles/dot-vim",
      "is_healthy": false,
      "issue_type": "missing links"
    }
  ]
}
```

JSON output is an object with a `packages` array, not a bare array. Scripts
must address it as `.packages[]`, for example
`dot list --format json | jq -r '.packages[].name'`.

**Issue Types**:
- `broken links`: Symlinks point to non-existent targets
- `wrong target`: Symlinks point to unexpected locations outside the package directory
- `missing links`: Expected symlinks do not exist

**Exit Codes**:
- `0`: Success
- `1`: Error listing packages

## Utility Commands

### version

Version information is available through the root flag only; there is no
`version` subcommand.

**Synopsis**:
```bash
dot --version
```

**Example Output**:
```
dot version 0.6.5 (commit: a1b2c3d4, built: 2026-07-27)
```

Source builds without release metadata report
`dot version unknown (built from source)`.

### help

Display help information.

**Synopsis**:
```bash
dot help [COMMAND]
```

**Arguments**:
- `COMMAND` (optional): Show help for specific command

**Examples**:
```bash
# General help
dot help

# Command-specific help
dot help manage
dot help status

# Alternative using flag
dot --help
dot manage --help
```

### completion

Generate shell completion script.

**Synopsis**:
```bash
dot completion SHELL
```

**Arguments**:
- `SHELL`: Shell type (`bash`, `zsh`, `fish`, `powershell`)

**Examples**:
```bash
# Bash
dot completion bash > /etc/bash_completion.d/dot

# Zsh
dot completion zsh > "${fpath[1]}/_dot"

# Fish
dot completion fish > ~/.config/fish/completions/dot.fish

# PowerShell
dot completion powershell > dot.ps1
```

## Exit Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Operation completed successfully |
| 1 | Error | Any failure: invalid arguments, conflicts, permissions, missing package |
| 2 | Doctor errors | `dot doctor` only: errors detected |
| 130 | Interrupted | A second SIGINT forced exit |

`dot doctor` is the only command with graded exit codes: 0 healthy, 1 warnings,
2 errors. Every other command returns 0 or 1; the specific failure is reported
on stderr, not encoded in the exit status.

**Usage in Scripts**:
```bash
#!/bin/bash

if ! dot manage dot-vim; then
    echo "manage failed; see stderr for the reason" >&2
    exit 1
fi

dot doctor
case $? in
    0) echo "healthy" ;;
    1) echo "warnings" ;;
    2) echo "errors" ;;
esac
```

## Command Patterns

### Dry Run Pattern

Preview before applying:

```bash
# Always preview first
dot --dry-run manage dot-vim

# Review output, then apply
dot manage dot-vim
```

### Verbose Debugging Pattern

Debug issues with verbose output:

```bash
# Increase verbosity to see details
dot -vvv manage dot-vim

# Or use with doctor
dot -vv doctor
```

### Scripting Pattern

Quiet mode with JSON logs:

```bash
#!/bin/bash

# Run command quietly
output=$(dot --quiet --log-json manage dot-vim 2>&1)

# Parse JSON output
if [ $? -eq 0 ]; then
    echo "Success"
else
    echo "$output" | jq '.error'
fi
```

### Batch Operations Pattern

Manage multiple packages from list:

```bash
# From file
cat packages.txt | xargs dot manage

# From array
packages=(dot-vim dot-zsh dot-tmux dot-git)
dot manage "${packages[@]}"

# With error checking
for pkg in dot-vim dot-zsh dot-tmux; do
    if ! dot manage "$pkg"; then
        echo "Failed to manage: $pkg"
    fi
done
```

## Command Aliases

The only built-in aliases are `cfg` for `config` and `show`/`ls` for
`config list`. Shell aliases are otherwise recommended:

```bash
# Common aliases
alias dm='dot manage'
alias du='dot unmanage'
alias dr='dot remanage'
alias ds='dot status'
alias dl='dot list'
alias dd='dot doctor'

# With default options
alias dot-dry='dot --dry-run'
alias dot-verbose='dot -vv'
```

Add to `~/.bashrc`, `~/.zshrc`, or equivalent.

## Next Steps

- [Common Workflows](06-workflows.md): See commands in real-world scenarios
- [Advanced Features](07-advanced.md): Deep dive into options and features
- [Troubleshooting Guide](08-troubleshooting.md): Solve common issues

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

