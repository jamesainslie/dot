# Simple Package Example

Minimal example demonstrating basic dot usage with a single package.

## Structure

```
vim/
└── dot-vimrc
```

## Setup

```bash
# View package structure
ls -la vim/

# Preview installation
dot --dry-run manage vim

# Install package
dot manage vim
```

## Expected Result

```bash
# Symlink created
ls -la ~/.vimrc
# Output: .vimrc -> /path/to/examples/basic/simple-package/vim/dot-vimrc

# Verify content
cat ~/.vimrc
```

## Cleanup

```bash
# Remove package
dot unmanage vim

# Verify removal
ls ~/.vimrc
# Output: ls: ~/.vimrc: No such file or directory
```

## Key Concepts

- **Package**: Directory containing configuration files (vim/)
- **Dotfile Translation**: `dot-vimrc` becomes `.vimrc` in target
- **Symlink**: Link from target directory to package file
- **Manage/Unmanage**: Install and remove operations

## Navigation

**[↑ Back to Main README](../../../README.md)** | [Examples Index](../../README.md)

