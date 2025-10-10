# Test Dotfiles Repository

This is a test repository containing sample dotfiles for testing the `dot` dotfile manager.

## Packages

### dot-zsh
Shell configuration for Zsh with basic settings, aliases, and prompt customization.

**Files:**
- `zshrc` → `~/.zshrc`
- `zshenv` → `~/.zshenv`

### dot-git
Git configuration with user settings, aliases, and global ignore patterns.

**Files:**
- `config` → `~/.gitconfig`
- `ignore` → `~/.gitignore_global`

### dot-vim
Vim editor configuration with sensible defaults.

**Files:**
- `vimrc` → `~/.vimrc`
- `colors/` → `~/.vim/colors/`

### dot-ssh
SSH client configuration with host-specific settings.

**Files:**
- `config` → `~/.ssh/config`

## Installation Profiles

See `.dotbootstrap.yaml` for available installation profiles:

- **minimal**: Basic shell and git configuration
- **default**: Standard development setup
- **full**: Complete environment including SSH config

## Usage

This repository is designed to be used with the `dot` tool:

```bash
dot clone file:///path/to/test-repo
dot manage dot-zsh
dot status
```

