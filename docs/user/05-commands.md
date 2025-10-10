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
dot --dir ~/dotfiles manage vim
dot -d /opt/configs status
```

#### `-t, --target PATH`

Specify target directory (destination for symlinks).

**Default**: `$HOME`  
**Example**:
```bash
dot --target ~ manage vim
dot -t /home/user unmanage zsh
```

### Execution Mode Options

#### `-n, --dry-run`

Preview operations without applying changes.

**Example**:
```bash
dot --dry-run manage vim
dot -n unmanage zsh
```

Shows planned operations with no filesystem modifications.

#### `--quiet`

Suppress non-error output.

**Example**:
```bash
dot --quiet manage vim
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
dot -v manage vim      # Info level
dot -vv status         # Debug level
dot -vvv remanage zsh  # Trace level
```

### Output Format Options

#### `--log-json`

Output logs in JSON format.

**Example**:
```bash
dot --log-json manage vim
```

JSON output for log aggregation and parsing.

#### `--color WHEN`

Control color output.

**Values**: `auto`, `always`, `never`  
**Default**: `auto`  
**Example**:
```bash
dot --color always status
dot --color never list
```

### Link Options

#### `--absolute`

Create absolute symlinks instead of relative.

**Example**:
```bash
dot --absolute manage vim
```

#### `--no-folding`

Disable directory folding optimization.

**Example**:
```bash
dot --no-folding manage vim
```

Creates per-file links instead of directory links.

### Ignore Options

#### `--ignore PATTERN`

Add ignore pattern (repeatable).

**Example**:
```bash
dot --ignore "*.log" manage vim
dot --ignore "*.log" --ignore "*.tmp" manage zsh
```

#### `--override PATTERN`

Force include pattern despite ignore rules (repeatable).

**Example**:
```bash
dot --override ".gitignore" manage git
```

### Conflict Resolution Options

#### `--on-conflict POLICY`

Set conflict resolution policy.

**Values**: `fail`, `backup`, `overwrite`, `skip`  
**Default**: `fail`  
**Example**:
```bash
dot --on-conflict backup manage vim
dot --on-conflict skip manage zsh
```

## Package Management Commands

### manage

Install packages by creating symlinks.

**Synopsis**:
```bash
dot manage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to install

**Options**: All global options

**Examples**:
```bash
# Single package
dot manage vim

# Multiple packages
dot manage vim zsh tmux git

# With options
dot --no-folding manage vim
dot --absolute manage configs
dot --dry-run manage test-package

# Different directories
dot --dir ~/dotfiles --target ~ manage vim
```

**Behavior**:
1. Scans package directories
2. Computes desired symlink state
3. Detects conflicts
4. Resolves conflicts per policy
5. Creates symlinks with dependency ordering
6. Updates manifest

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `2`: Invalid arguments
- `3`: Conflicts detected (with fail policy)
- `4`: Permission denied

### unmanage

Remove packages by deleting symlinks, with optional restoration or cleanup.

**Synopsis**:
```bash
dot unmanage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to remove

**Options**:
- All global options
- `--purge`: Delete package directory after removing links
- `--no-restore`: Skip restoring adopted packages to target
- `--cleanup`: Remove orphaned packages from manifest only

**Examples**:
```bash
# Remove managed package (removes links only)
dot unmanage vim

# Remove adopted package (restores files to target, keeps in package)
dot unmanage dot-ssh

# Remove with purge (deletes package directory)
dot unmanage --purge vim

# Remove without restoring (for adopted packages)
dot unmanage --no-restore dot-ssh

# Clean up orphaned packages
dot unmanage --cleanup dot-old-package

# Preview removal
dot --dry-run unmanage vim
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

**Cleanup Mode**:

Use `--cleanup` to remove orphaned packages (missing links or directories):

```bash
dot unmanage --cleanup old-package
```

Only updates manifest, no filesystem operations.

**Safety Guarantees**:
- Only removes links pointing to package directory
- Preserves non-managed files
- Validates link targets before deletion
- Adopted packages restored by default (preserves your data)

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `5`: Package not found/not installed

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
dot remanage vim

# Multiple packages
dot remanage vim zsh tmux

# Preview changes
dot --dry-run remanage vim

# Verbose output to see detection details
dot -vv remanage zsh
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
rm ~/.vimrc

# Check status
dot doctor
# ✗ error: .vimrc link does not exist

# Recreate missing link
dot remanage vim
# Successfully remanaged 1 package(s)

# Link restored
ls -la ~/.vimrc
# ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
```

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
# Auto-naming mode (single file/directory)
dot adopt [options] FILE|DIRECTORY

# Glob expansion mode (multiple files with common prefix)
dot adopt [options] PATTERN...

# Explicit package mode
dot adopt [options] PACKAGE FILE|DIRECTORY [FILE|DIRECTORY...]
```

**Arguments**:
- `FILE|DIRECTORY`: Path to file or directory to adopt
- `PACKAGE`: Explicit package name (optional)
- `PATTERN`: Shell glob pattern (e.g., `.git*`)

**Options**: All global options

**Modes**:

#### Auto-Naming Mode
Single file or directory - package name derived automatically:
```bash
dot adopt .vimrc      # Creates package: dot-vimrc
dot adopt .ssh        # Creates package: dot-ssh
dot adopt .config     # Creates package: dot-config
```

#### Glob Expansion Mode
Multiple files with common prefix - package name derived from prefix:
```bash
dot adopt .git*       # Expands to .gitconfig, .gitignore, etc.
                      # Creates package: dot-git
                      # All files adopted into single package

dot adopt .vim*       # Expands to .vimrc, .viminfo, etc.
                      # Creates package: dot-vim
```

#### Explicit Package Mode
Specify package name explicitly:
```bash
dot adopt vim .vimrc .vim/          # Package: vim
dot adopt configs .config/ .local/  # Package: configs
```

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
1. Determines adoption mode (auto-naming, glob, or explicit)
2. Derives or uses provided package name
3. Creates package directory structure
4. Moves files/directories to package (applying dotfile translation)
5. Creates symlinks in original locations
6. Records package as "adopted" in manifest

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `2`: Invalid arguments
- `4`: Permission denied

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
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`)
- All global options

**Examples**:
```bash
# All packages
dot status

# Specific packages
dot status vim zsh

# JSON output
dot status --format json

# YAML output
dot status --format yaml

# Table format
dot status --format table

# Combine with verbosity
dot -v status vim
```

**Output Fields**:
- Package name
- Installation status
- Link count
- Installation date
- List of symlinks
- Conflicts or issues

**Example Output (text)**:
```
Package: vim
  Status: installed
  Links: 3
  Installed: 2025-10-07 10:30:00
  
  Links:
    ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
    ~/.vim/colors/ -> ~/dotfiles/vim/dot-vim/colors/
    ~/.vim/autoload/ -> ~/dotfiles/vim/dot-vim/autoload/
```

**Example Output (JSON)**:
```json
{
  "packages": [
    {
      "name": "vim",
      "status": "installed",
      "link_count": 3,
      "installed_at": "2025-10-07T10:30:00Z",
      "links": [
        {
          "target": "~/.vimrc",
          "source": "~/dotfiles/vim/dot-vimrc",
          "type": "file"
        }
      ]
    }
  ]
}
```

**Exit Codes**:
- `0`: Success
- `1`: Error querying status

### doctor

Validate installation health and detect issues.

**Synopsis**:
```bash
dot doctor [options]
```

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`)
- All global options

**Examples**:
```bash
# Basic health check
dot doctor

# Detailed output
dot -v doctor

# JSON output for scripting
dot doctor --format json

# Table format
dot doctor --format table
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
Running health checks...

✓ All symlinks valid
✓ No broken links
✓ No orphaned links
✓ Manifest consistent
✓ No permission issues

Health check passed: 0 issues found
```

**Example Output (issues)**:
```
Running health checks...

✗ Broken links found: 2
  ~/.vimrc -> ~/dotfiles/vim/dot-vimrc (target missing)
  ~/.zshrc -> ~/dotfiles/zsh/dot-zshrc (target missing)

✗ Orphaned links found: 1
  ~/.bashrc -> ~/old-dotfiles/bash/bashrc

Suggestions:
  - Remove broken links: dot doctor --fix-broken
  - Adopt orphaned links: dot adopt bash ~/.bashrc
  - Reinstall packages: dot remanage vim zsh

Health check failed: 3 issues found
```

**Exit Codes**:
- `0`: No issues found
- `1`: Issues detected
- `2`: Invalid arguments

### list

Show installed package inventory.

**Synopsis**:
```bash
dot list [options]
```

**Options**:
- `-f, --format FORMAT`: Output format (`text`, `json`, `yaml`, `table`)
- `-s, --sort FIELD`: Sort by field (`name`, `links`, `date`)
- All global options

**Examples**:
```bash
# List all packages
dot list

# Sort by link count
dot list --sort links

# Sort by installation date
dot list --sort date

# JSON output
dot list --format json

# Table format
dot list --format table

# Combine sorting and format
dot list --sort links --format table
```

**Example Output (text)**:
```
vim    (3 links) installed 2025-10-07 10:30:00
zsh    (2 links) installed 2025-10-07 10:31:00
tmux   (1 link)  installed 2025-10-07 10:32:00
```

**Example Output (table)**:
```
NAME   LINKS  INSTALLED
vim    3      2025-10-07 10:30:00
zsh    2      2025-10-07 10:31:00
tmux   1      2025-10-07 10:32:00
```

**Example Output (JSON)**:
```json
[
  {
    "name": "vim",
    "link_count": 3,
    "installed_at": "2025-10-07T10:30:00Z"
  },
  {
    "name": "zsh",
    "link_count": 2,
    "installed_at": "2025-10-07T10:31:00Z"
  }
]
```

**Exit Codes**:
- `0`: Success
- `1`: Error listing packages

## Utility Commands

### version

Display version information.

**Synopsis**:
```bash
dot version [options]
```

**Options**:
- `--short`: Show version number only
- All global options

**Examples**:
```bash
# Full version info
dot version

# Short version
dot version --short

# Alternative using flag
dot --version
```

**Example Output**:
```
dot version v0.1.0
Built with Go 1.25
Commit: abc1234
Build date: 2025-10-07
Platform: linux/amd64
```

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

Standard exit codes across all commands:

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Operation completed successfully |
| 1 | General error | Unspecified error occurred |
| 2 | Invalid arguments | Command-line arguments invalid |
| 3 | Conflicts detected | Conflicts found (with fail policy) |
| 4 | Permission denied | Insufficient permissions |
| 5 | Package not found | Specified package does not exist |

**Usage in Scripts**:
```bash
#!/bin/bash

# Check if operation succeeded
if dot manage vim; then
    echo "vim installed successfully"
else
    exit_code=$?
    case $exit_code in
        3) echo "Conflicts detected" ;;
        4) echo "Permission denied" ;;
        5) echo "Package not found" ;;
        *) echo "Error: $exit_code" ;;
    esac
    exit 1
fi
```

## Command Patterns

### Dry Run Pattern

Preview before applying:

```bash
# Always preview first
dot --dry-run manage vim

# Review output, then apply
dot manage vim
```

### Verbose Debugging Pattern

Debug issues with verbose output:

```bash
# Increase verbosity to see details
dot -vvv manage vim

# Or use with doctor
dot -vv doctor
```

### Scripting Pattern

Quiet mode with JSON output:

```bash
#!/bin/bash

# Run command quietly
output=$(dot --quiet --log-json manage vim 2>&1)

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
packages=(vim zsh tmux git)
dot manage "${packages[@]}"

# With error checking
for pkg in vim zsh tmux; do
    if ! dot manage "$pkg"; then
        echo "Failed to manage: $pkg"
    fi
done
```

## Command Aliases

No built-in aliases, but shell aliases recommended:

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

