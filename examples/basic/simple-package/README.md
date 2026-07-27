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

The package name `vim` becomes a target subdirectory, and the `dot-` file prefix
becomes a leading dot.

```bash
# Symlink created
ls -la ~/vim/.vimrc
# Output: .vimrc -> /path/to/examples/basic/simple-package/vim/dot-vimrc

# Verify content
cat ~/vim/.vimrc
```

To install to `~/.vim/` instead, rename the package directory to `dot-vim`.

## Cleanup

```bash
# Remove package
dot unmanage vim

# Verify removal
ls ~/vim/.vimrc
# Output: ls: /home/user/vim/.vimrc: No such file or directory
```

## Key Concepts

- **Package**: Directory containing configuration files (`vim/`)
- **Package Name Mapping**: package `vim` installs under `~/vim/`; package
  `dot-vim` installs under `~/.vim/`
- **Dotfile Translation**: `dot-vimrc` becomes `.vimrc` in the target
- **Symlink**: Link from target directory to package file
- **Manage/Unmanage**: Install and remove operations

## Navigation

**[Back to Main README](../../../README.md)** | [Examples Index](../../README.md)

