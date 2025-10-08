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

Remove packages by deleting symlinks.

**Synopsis**:
```bash
dot unmanage [options] PACKAGE [PACKAGE...]
```

**Arguments**:
- `PACKAGE`: One or more package names to remove

**Options**: All global options

**Examples**:
```bash
# Single package
dot unmanage vim

# Multiple packages
dot unmanage vim zsh tmux

# Preview removal
dot --dry-run unmanage vim

# Verbose output
dot -v unmanage zsh
```

**Behavior**:
1. Loads manifest for package state
2. Validates link ownership
3. Removes symlinks
4. Cleans up empty directories
5. Updates manifest

**Safety Guarantees**:
- Only removes links pointing to package directory
- Preserves non-managed files
- Validates link targets before deletion

**Exit Codes**:
- `0`: Success
- `1`: Error during operation
- `5`: Package not found/not installed

### remanage

Update packages efficiently using incremental detection.

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

# Force full remanage (skip incremental detection)
dot remanage --no-incremental vim

# Preview changes
dot --dry-run remanage vim
```

**Behavior**:
1. Loads manifest with previous state
2. Computes content hashes
3. Compares with stored hashes
4. Processes only changed packages
5. Updates manifest

**Incremental Detection**:
- Unchanged packages: Skipped entirely
- Changed packages: Unmanaged then managed
- New packages: Managed
- Missing packages: Unmanaged

**Exit Codes**:
- `0`: Success, changes applied or no changes needed
- `1`: Error during operation

### adopt

Move existing files into package and create symlinks.

**Synopsis**:
```bash
dot adopt [options] PACKAGE FILE [FILE...]
```

**Arguments**:
- `PACKAGE`: Target package name
- `FILE`: One or more files to adopt

**Options**: All global options

**Examples**:
```bash
# Adopt single file
dot adopt vim ~/.vimrc

# Adopt multiple files
dot adopt zsh ~/.zshrc ~/.zshenv ~/.zprofile

# Adopt to new package
dot adopt git ~/.gitconfig ~/.gitignore_global

# Preview adoption
dot --dry-run adopt vim ~/.vimrc

# With backup
dot --on-conflict backup adopt vim ~/.vimrc
```

**Behavior**:
1. Validates source files exist
2. Determines target paths in package (respecting dotfile translation)
3. Creates package directory if needed
4. Moves files to package
5. Creates symlinks in original locations
6. Updates manifest

**Dotfile Translation**:
- `.vimrc` → `dot-vimrc` in package
- `.config/nvim/init.vim` → `dot-config/nvim/init.vim`

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

