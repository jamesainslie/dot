# Quick Start Tutorial

This tutorial provides a hands-on introduction to dot through a practical example.

## Prerequisites

- dot installed (see [Installation Guide](02-installation.md))
- Terminal access
- Basic familiarity with command line

## Tutorial Scenario

You will:
1. Create a dotfiles repository with vim and zsh configurations
2. Install packages using `manage`
3. Check status with `status`
4. Adopt an existing file with `adopt`
5. Update packages with `remanage`
6. Remove packages with `unmanage`

## Step 1: Create Dotfiles Repository

Create a directory to store your packages:

```bash
# Create dotfiles directory
mkdir -p ~/dotfiles
cd ~/dotfiles
```

This directory will be your **package directory** containing packages.

## Step 2: Create First Package (vim)

Create a vim package with configuration:

```bash
# Create vim package directory (package name determines target)
mkdir -p dot-vim

# Create vim configuration
cat > dot-vim/vimrc << 'EOF'
" Basic vim configuration
set number
set relativenumber
set expandtab
set tabstop=4
set shiftwidth=4
set autoindent
syntax on
colorscheme desert
EOF

# Create vim plugin directory structure
mkdir -p dot-vim/colors
mkdir -p dot-vim/autoload

# Add a color scheme
cat > dot-vim/colors/custom.vim << 'EOF'
" Custom color scheme
hi Normal guibg=black guifg=white
EOF
```

Package structure (with package name mapping):
```
~/dotfiles/dot-vim/          Package name "dot-vim" → target ~/.vim/
├── vimrc                    → ~/.vim/vimrc
└── colors/                  → ~/.vim/colors/
    └── custom.vim           → ~/.vim/colors/custom.vim
```

## Step 3: Create Second Package (zsh)

Create a zsh package:

```bash
# Create zsh package directory (package name "dot-zsh" → target ~/.zsh/)
mkdir -p dot-zsh

# Create zsh configuration
cat > dot-zsh/zshrc << 'EOF'
# Zsh configuration
export EDITOR=vim
export VISUAL=vim

# Path configuration
export PATH="$HOME/.local/bin:$PATH"

# Aliases
alias ll='ls -lah'
alias la='ls -A'
alias l='ls -CF'

# History configuration
HISTSIZE=10000
SAVEHIST=10000
HISTFILE=~/.zsh_history
EOF

# Create zsh environment file
cat > dot-zsh/zshenv << 'EOF'
# Zsh environment variables (loaded for all shells)
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8
EOF
```

Current structure:
```
~/dotfiles/
├── dot-vim/               → targets ~/.vim/
│   ├── vimrc
│   └── colors/
│       └── custom.vim
└── dot-zsh/               → targets ~/.zsh/
    ├── zshrc
    └── zshenv
```

## Step 4: Preview Installation (Dry Run)

Preview what dot will do before applying changes:

```bash
dot --dry-run manage dot-vim
```

Expected output:
```
Dry run mode - no changes will be applied

Plan:
  + Create directory: ~/.vim
  + Create directory: ~/.vim/colors
  + Create symlink: ~/.vim/vimrc -> ~/dotfiles/dot-vim/vimrc
  + Create symlink: ~/.vim/colors/custom.vim -> ~/dotfiles/dot-vim/colors/custom.vim

Summary:
  Directories: 2
  Symlinks: 2
  Conflicts: 0
```

## Step 5: Install First Package

Install vim package:

```bash
dot manage dot-vim
```

Expected output:
```
✓ Managed 1 package
```

Verify installation:

```bash
# Check symlinks created in ~/.vim/
ls -la ~/.vim/
# Output shows vimrc and colors/ directory

# Check the vimrc link
ls -la ~/.vim/vimrc
# Output: lrwxr-xr-x ... vimrc -> /home/user/dotfiles/dot-vim/vimrc

# Verify vim configuration works
cat ~/.vim/vimrc | head -3
```

## Step 6: Install Multiple Packages

Install vim and zsh together:

```bash
dot manage dot-vim dot-zsh
```

Expected output:
```
✓ Managed 2 packages
```

## Step 7: Check Status

View installed packages:

```bash
dot status
```

Expected output:
```
dot-vim
  Links: 2
  Installed: just now
  Files:
    .vim/colors/custom.vim
    .vim/vimrc

dot-zsh
  Links: 2
  Installed: just now
  Files:
    .zsh/zshenv
    .zsh/zshrc
```

Installation times are rendered as intervals relative to now, not absolute
timestamps.

Status for specific package:

```bash
dot status dot-vim
```

## Step 8: Adopt Existing Files

Suppose you have existing configuration files to bring under management. The `adopt` command supports several modes:

### Auto-Naming: Single File

Adopt a single file (package name derived automatically):

```bash
# Create sample existing file (if needed)
cat > ~/.gitconfig << 'EOF'
[user]
    name = Your Name
    email = your.email@example.com
[core]
    editor = vim
EOF

# Adopt with auto-naming (creates package "dot-gitconfig")
dot adopt ~/.gitconfig
```

Expected output:
```
✓ Adopted /home/user/.gitconfig into dot-gitconfig
```

### Glob Expansion: Multiple Related Files

Shell globs are expanded by the shell before dot sees them. With two or more
arguments, the first argument is the package name and the remaining arguments are
the files to adopt. The package name must therefore be given explicitly whenever
a glob may match more than one file:

```bash
# Create sample git-related files
cat > ~/.gitconfig << 'EOF'
[user]
    name = Your Name
EOF

cat > ~/.gitignore << 'EOF'
*.swp
.DS_Store
EOF

# Adopt all .git* files into a package named dot-git
dot adopt dot-git .git*
```

Expected output:
```
✓ Adopted 2 files into dot-git
```

**Warning**: Omitting the package name causes the first expanded filename to be
used as the package name, and that file is then not adopted. `dot adopt .git*`
creates a package literally named `.git-credentials` and adopts only the
remaining matches.

### Explicit Package Name

The same rule applies when adopting an explicit list of files:

```bash
# Adopt with explicit package name
dot adopt my-git ~/.gitconfig ~/.gitignore
```

Verification:

```bash
# Check the package was created; names are translated inside the package
ls ~/dotfiles/dot-git/
# Output: dot-gitconfig  dot-gitignore

# Verify symlink
ls -la ~/.gitconfig
# Output: lrwxrwxrwx ... .gitconfig -> /home/user/dotfiles/dot-git/dot-gitconfig

# Content preserved
cat ~/.gitconfig
```

## Step 9: Modify and Update Package

Modify vim configuration and update:

```bash
# Edit vim configuration
cat >> ~/dotfiles/dot-vim/vimrc << 'EOF'

" Additional settings
set cursorline
set hlsearch
EOF

# Update vim package
dot remanage dot-vim
```

Expected output:
```
✓ Remanaged 1 package
```

If no package content changed, dot reports that no changes were detected and
does nothing:

```
ℹ No changes detected for 1 package
```

Changes are immediately reflected:

```bash
# Verify changes
tail -3 ~/.vim/vimrc
# Output shows new lines added
```

## Step 10: Ignoring Files

Sometimes packages contain files you do not want to symlink (temporary files, cache, etc.). Use the ignore system:

### Using .dotignore Files

Create a `.dotignore` file in your package:

```bash
# Add some cache files to vim package
touch ~/dotfiles/dot-vim/.netrwhist
mkdir -p ~/dotfiles/dot-vim/swap
mkdir -p ~/dotfiles/dot-vim/undo

# Create .dotignore to exclude them
cat > ~/dotfiles/dot-vim/.dotignore << 'EOF'
# Ignore vim cache files
.netrwhist
swap/
undo/

# Ignore backup files
*.swp
*~
EOF

# Remanage to apply ignore rules
dot remanage dot-vim
```

The ignored files stay in your repository but are not symlinked.

### Using Command-Line Flags

Temporarily ignore patterns:

```bash
# Ignore specific files for this operation
dot manage dot-vim --ignore "*.log" --ignore "cache/"
```

### Handling Large Files

For packages with large files (like VM images), use size filtering:

```bash
# Skip files larger than 100MB
dot manage dot-vim --max-file-size 100MB

# In batch mode (auto-skip without prompting)
dot manage dot-vim --batch --max-file-size 50MB
```

### Negation Patterns

Un-ignore specific files:

```bash
# Create .dotignore with negation
cat > ~/dotfiles/dot-vim/.dotignore << 'EOF'
# Ignore all backup files
*.bak

# But keep important backups
!vimrc.bak
!important.bak
EOF
```

For more details, see the [Ignore System Guide](ignore-system.md).

## Step 11: List Installed Packages

View package inventory:

```bash
dot list
```

Expected output:
```
Packages: 3 packages in /home/user/dotfiles

✓  dot-git  (2 links)  installed just now
✓  dot-vim  (2 links)  installed 5 minutes ago
✓  dot-zsh  (2 links)  installed 5 minutes ago

3 healthy
```

A column-oriented table is available with `dot list --format table`.

Sort options:

```bash
# Sort by link count
dot list --sort links

# Sort by installation date
dot list --sort date
```

## Step 12: Verify Installation Health

Check for issues:

```bash
dot doctor
```

Expected output (healthy):
```
✓ Healthy
  • 4 total links (4 managed, 0 broken, 0 orphaned)
  No issues found
```

When problems are found, `dot doctor --detailed` reports each one with its path
and a suggested remedy:

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

`dot doctor` exits 0 when healthy, 1 when only warnings were found, and 2 when
errors were found. All other commands exit 0 on success and 1 on failure, or 130
when interrupted a second time.

## Step 13: Unmanage Package

Remove a package:

```bash
# Preview removal
dot --dry-run unmanage dot-git

# Actual removal
dot unmanage dot-git
```

Expected output:
```
✓ Unmanaged and restored 1 package
```

Verification:

```bash
# The symlink has been replaced by the restored regular file
ls -la ~/.gitconfig
# Output: -rw-r--r-- ... .gitconfig

# The package directory is retained
ls ~/dotfiles/dot-git
# Output: dot-gitconfig  dot-gitignore
```

Adopted packages are restored to their original locations by default. Use
`--no-restore` to leave the files in the package directory, or `--purge` to
delete the package directory instead of restoring.

## Step 14: Clean Up (Tutorial Completion)

Remove tutorial files:

```bash
# Unmanage all packages
dot unmanage dot-vim dot-zsh

# Verify removal
dot status
# Output: No packages installed

# Remove dotfiles directory (optional)
rm -rf ~/dotfiles
```

## Common Operations Reference

### Installation

```bash
# Single package
dot manage dot-vim

# Multiple packages
dot manage dot-vim dot-zsh dot-tmux
```

### Query

```bash
# Status of all packages
dot status

# Status of specific packages
dot status dot-vim dot-zsh

# List packages
dot list

# Health check
dot doctor
```

### Modification

```bash
# Update packages
dot remanage dot-vim

# Adopt existing files
dot adopt dot-package file1 file2

# Remove packages
dot unmanage dot-vim
```

### Dry Run

```bash
# Preview any operation
dot --dry-run manage dot-vim
dot --dry-run unmanage dot-zsh
dot --dry-run adopt dot-git ~/.gitconfig
```

## Tutorial Summary

You learned:
- Creating a dotfiles repository with packages
- Installing packages with `manage`
- Checking status with `status` and `list`
- Adopting existing files with `adopt`
- Updating packages with `remanage`
- Removing packages with `unmanage`
- Using dry-run mode for preview
- Verifying health with `doctor`

## Next Steps

### Organize Your Dotfiles

1. Create packages for each application
2. Structure packages to mirror target directory
3. Use dotfile translation (`dot-` prefix)
4. Commit packages to version control

### Advanced Usage

- [Configuration Reference](04-configuration.md): Customize dot behavior
- [Command Reference](05-commands.md): Detailed command documentation
- [Common Workflows](06-workflows.md): Real-world usage patterns
- [Advanced Features](07-advanced.md): Ignore patterns, policies, performance

### Version Control Integration

```bash
cd ~/dotfiles

# Initialize git repository
git init

# Add packages
git add dot-vim/ dot-zsh/ dot-git/

# Commit
git commit -m "feat(dotfiles): add initial configurations"

# Add remote and push
git remote add origin https://github.com/username/dotfiles.git
git push -u origin main
```

### Multi-Machine Setup

On other machines:

```bash
# Clone dotfiles
git clone https://github.com/username/dotfiles.git ~/dotfiles

# Install packages
cd ~/dotfiles
dot manage dot-vim dot-zsh dot-git
```

## Troubleshooting

### Conflicts During Installation

If dot reports conflicts:

```bash
# Check what conflicts
dot manage dot-vim
# Output: Error: conflict at ~/.vim/.vimrc: file exists

# Options:
# 1. Enable backups in the configuration, then retry. Conflicting files are
#    moved to <target>/.dot-backup before the symlink is created.
dot config set symlinks.backup true
dot manage dot-vim

# 2. Adopt conflicting file
dot adopt dot-vim ~/.vim/.vimrc

# 3. Manually resolve
rm ~/.vim/.vimrc
dot manage dot-vim
```

There is no `--on-conflict` flag. Conflict behaviour is controlled by the
`symlinks.overwrite` and `symlinks.backup` configuration keys, and the backup
directory by `--backup-dir` or `symlinks.backup_dir`.

### Broken Symlinks

If symlinks break after moving dotfiles:

```bash
# Unmanage old links
dot unmanage dot-vim

# Remanage with new location
cd /new/path/to/dotfiles
dot manage dot-vim
```

### Permission Errors

If operations fail with permission errors:

```bash
# Check file permissions
ls -la ~/dotfiles/dot-vim/

# Fix permissions if needed
chmod -R u+rw ~/dotfiles/dot-vim/
```

See [Troubleshooting Guide](08-troubleshooting.md) for more issues and solutions.

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

