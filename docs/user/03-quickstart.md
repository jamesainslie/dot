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
# Create vim package directory
mkdir -p vim

# Create vim configuration
cat > vim/dot-vimrc << 'EOF'
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
mkdir -p vim/dot-vim/{colors,autoload}

# Add a color scheme
cat > vim/dot-vim/colors/custom.vim << 'EOF'
" Custom color scheme
hi Normal guibg=black guifg=white
EOF
```

Package structure:
```
~/dotfiles/vim/
├── dot-vimrc              → ~/.vimrc
└── dot-vim/               → ~/.vim/
    ├── colors/
    │   └── custom.vim
    └── autoload/
```

## Step 3: Create Second Package (zsh)

Create a zsh package:

```bash
# Create zsh package directory
mkdir -p zsh

# Create zsh configuration
cat > zsh/dot-zshrc << 'EOF'
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
cat > zsh/dot-zshenv << 'EOF'
# Zsh environment variables (loaded for all shells)
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8
EOF
```

Current structure:
```
~/dotfiles/
├── vim/
│   ├── dot-vimrc
│   └── dot-vim/
│       ├── colors/
│       │   └── custom.vim
│       └── autoload/
└── zsh/
    ├── dot-zshrc
    └── dot-zshenv
```

## Step 4: Preview Installation (Dry Run)

Preview what dot will do before applying changes:

```bash
dot --dry-run manage vim
```

Expected output:
```
Dry run mode - no changes will be applied

Plan:
  + Create directory: ~/.vim/colors
  + Create directory: ~/.vim/autoload
  + Create symlink: ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
  + Create symlink: ~/.vim/colors/custom.vim -> ~/dotfiles/vim/dot-vim/colors/custom.vim

Summary:
  Directories: 2
  Symlinks: 2
  Conflicts: 0
```

## Step 5: Install First Package

Install vim package:

```bash
dot manage vim
```

Expected output:
```
Managing package: vim
Created symlink: ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
Created directory-level symlink: ~/.vim/ -> ~/dotfiles/vim/dot-vim/
Successfully managed: vim
```

Verify installation:

```bash
# Check symlink
ls -la ~/.vimrc
# Output: lrwxr-xr-x ... .vimrc -> /home/user/dotfiles/vim/dot-vimrc

# Verify vim configuration works
vim --version | head -1
```

## Step 6: Install Multiple Packages

Install vim and zsh together:

```bash
dot manage vim zsh
```

Expected output:
```
Managing packages: vim, zsh
Package vim: already installed
Created symlink: ~/.zshrc -> ~/dotfiles/zsh/dot-zshrc
Created symlink: ~/.zshenv -> ~/dotfiles/zsh/dot-zshenv
Successfully managed: zsh
```

## Step 7: Check Status

View installed packages:

```bash
dot status
```

Expected output:
```
Package: vim
  Status: installed
  Links: 2
  Installed: 2025-10-07 10:30:00

  Links:
    ~/.vimrc -> ~/dotfiles/vim/dot-vimrc
    ~/.vim/ -> ~/dotfiles/vim/dot-vim/ (folded)

Package: zsh
  Status: installed
  Links: 2
  Installed: 2025-10-07 10:31:00

  Links:
    ~/.zshrc -> ~/dotfiles/zsh/dot-zshrc
    ~/.zshenv -> ~/dotfiles/zsh/dot-zshenv
```

Status for specific package:

```bash
dot status vim
```

## Step 8: Adopt Existing File

Suppose you have an existing `.gitconfig` file to bring under management:

```bash
# Create sample existing file (if needed)
cat > ~/.gitconfig << 'EOF'
[user]
    name = Your Name
    email = your.email@example.com
[core]
    editor = vim
EOF

# Adopt it into a new git package
dot adopt git ~/.gitconfig
```

Expected output:
```
Adopting files into package: git
Moved: ~/.gitconfig -> ~/dotfiles/git/dot-gitconfig
Created symlink: ~/.gitconfig -> ~/dotfiles/git/dot-gitconfig
Successfully adopted 1 file into: git
```

Verification:

```bash
# Check git package was created
ls ~/dotfiles/git/
# Output: dot-gitconfig

# Verify symlink
ls -la ~/.gitconfig
# Output: lrwxr-xr-x ... .gitconfig -> /home/user/dotfiles/git/dot-gitconfig

# Content preserved
cat ~/.gitconfig
```

## Step 9: Modify and Update Package

Modify vim configuration and update:

```bash
# Edit vim configuration
cat >> ~/dotfiles/vim/dot-vimrc << 'EOF'

" Additional settings
set cursorline
set hlsearch
EOF

# Update vim package
dot remanage vim
```

Expected output:
```
Remanaging package: vim
Detected changes in: vim
Relinked: ~/.vimrc
Successfully remanaged: vim
```

Changes are immediately reflected:

```bash
# Verify changes
tail -3 ~/.vimrc
# Output shows new lines added
```

## Step 10: List Installed Packages

View package inventory:

```bash
dot list
```

Expected output:
```
NAME  LINKS  INSTALLED
vim   2      2025-10-07 10:30:00
zsh   2      2025-10-07 10:31:00
git   1      2025-10-07 10:35:00
```

Sort options:

```bash
# Sort by link count
dot list --sort links

# Sort by installation date
dot list --sort date
```

## Step 11: Verify Installation Health

Check for issues:

```bash
dot doctor
```

Expected output (healthy):
```
Running health checks...

✓ All symlinks valid
✓ No broken links
✓ No orphaned links
✓ Manifest consistent
✓ No permission issues

Health check passed: 0 issues found
```

## Step 12: Unmanage Package

Remove a package:

```bash
# Preview removal
dot --dry-run unmanage git

# Actual removal
dot unmanage git
```

Expected output:
```
Unmanaging package: git
Removed symlink: ~/.gitconfig
Removed empty directory: ~/dotfiles/git/
Successfully unmanaged: git
```

Verification:

```bash
# File is gone
ls ~/.gitconfig
# Output: ls: ~/.gitconfig: No such file or directory

# Package directory removed
ls ~/dotfiles/git
# Output: ls: ~/dotfiles/git: No such file or directory
```

## Step 13: Clean Up (Tutorial Completion)

Remove tutorial files:

```bash
# Unmanage all packages
dot unmanage vim zsh

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
dot manage vim

# Multiple packages
dot manage vim zsh tmux

# With absolute links
dot --absolute manage vim

# Without directory folding
dot --no-folding manage vim
```

### Query

```bash
# Status of all packages
dot status

# Status of specific packages
dot status vim zsh

# List packages
dot list

# Health check
dot doctor
```

### Modification

```bash
# Update packages
dot remanage vim

# Adopt existing files
dot adopt package file1 file2

# Remove packages
dot unmanage vim
```

### Dry Run

```bash
# Preview any operation
dot --dry-run manage vim
dot --dry-run unmanage zsh
dot --dry-run adopt git ~/.gitconfig
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
git add vim/ zsh/ git/

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
dot manage vim zsh git
```

## Troubleshooting

### Conflicts During Installation

If dot reports conflicts:

```bash
# Check what conflicts
dot manage vim
# Output: Error: conflict at ~/.vimrc: file exists

# Options:
# 1. Backup conflicting file
dot --on-conflict backup manage vim

# 2. Adopt conflicting file
dot adopt vim ~/.vimrc

# 3. Manually resolve
rm ~/.vimrc
dot manage vim
```

### Broken Symlinks

If symlinks break after moving dotfiles:

```bash
# Unmanage old links
dot unmanage vim

# Remanage with new location
cd /new/path/to/dotfiles
dot manage vim
```

### Permission Errors

If operations fail with permission errors:

```bash
# Check file permissions
ls -la ~/dotfiles/vim/

# Fix permissions if needed
chmod -R u+rw ~/dotfiles/vim/
```

See [Troubleshooting Guide](08-troubleshooting.md) for more issues and solutions.

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

