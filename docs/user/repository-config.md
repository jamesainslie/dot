# Repository Configuration

## Overview

`dot` automatically looks for configuration in your dotfiles repository first, before checking your local XDG configuration directory. This solves the circular dependency problem where your dot configuration is itself managed by dot.

## How It Works

When you run any `dot` command, configuration is loaded in this order:

1. **Repository config**: `~/.dotfiles/.config/dot/config.yaml`
2. **XDG config**: `~/.config/dot/config.yaml`
3. **Built-in defaults**

The repository config takes precedence. This means your dotfiles repository can define its own configuration for how it should be managed.

## Setup

### 1. Add Config to Your Repository

Create a configuration file in your dotfiles repository:

```bash
mkdir -p ~/.dotfiles/.config/dot
$EDITOR ~/.dotfiles/.config/dot/config.yaml
```

Example configuration:

```yaml
directories:
  package: ~/.dotfiles
  target: $HOME
  manifest: ~/.local/share/dot/manifest

symlinks:
  mode: relative
  folding: true
  backup_dir: ~/.dotfiles.backup

dotfile:
  translate: true
  package_name_mapping: true
```

### 2. Commit and Share

Commit the configuration to your repository:

```bash
cd ~/.dotfiles
git add .config/dot/config.yaml
git commit -m "feat(config): add dot configuration"
git push
```

### 3. Clone on New Machines

When someone clones your repository:

```bash
dot clone https://github.com/yourname/dotfiles
```

`dot` will:
1. Clone to `~/.dotfiles`
2. Find `~/.dotfiles/.config/dot/config.yaml`
3. Use that configuration automatically
4. Install packages according to the repository's settings

## Benefits

### No Circular Dependency

The config file lives in the repository root at `.config/dot/config.yaml` — it's a regular file, not a managed symlink. This means:

- Config is available immediately after clone
- No chicken-and-egg problem
- Works automatically on all machines

### Single Source of Truth

Your repository defines how it should be managed:

- Target directory locations
- Symlink preferences
- Package name mapping
- Backup behavior

Everyone who clones your repository gets the same configuration.

### Machine-Specific Overrides

You can still have machine-specific settings in `~/.config/dot/config.yaml` if needed:

```bash
# Create local override (not committed to repo)
mkdir -p ~/.config/dot
echo 'directories:
  target: /custom/path' > ~/.config/dot/config.yaml
```

The local config will take precedence for this machine only.

## Configuration Precedence

### Complete Order

1. **Command-line flags** (highest priority)
2. **Environment variables** (`DOT_*`)
3. **XDG config** (`~/.config/dot/config.yaml`) 
4. **Repository config** (`~/.dotfiles/.config/dot/config.yaml`)
5. **Built-in defaults** (lowest priority)

Wait - that's backwards! Actually:

1. **Command-line flags** (highest priority)
2. **Environment variables** (`DOT_*`)
3. **Repository config** (`~/.dotfiles/.config/dot/config.yaml`) ← checked first
4. **XDG config** (`~/.config/dot/config.yaml`) ← fallback
5. **Built-in defaults** (lowest priority)

The repository config is checked *before* XDG config, so it takes precedence when both exist.

## Examples

### Basic Repository Config

```yaml
# .config/dot/config.yaml in your repository
directories:
  package: ~/.dotfiles
  target: $HOME

symlinks:
  mode: relative
  folding: true

dotfile:
  package_name_mapping: true
```

### Advanced Configuration

```yaml
# .config/dot/config.yaml
directories:
  package: ~/my-dotfiles
  target: $HOME
  manifest: ~/.local/share/dot/manifest

logging:
  level: INFO
  format: text

symlinks:
  mode: relative
  folding: true
  backup_dir: ~/.dotfiles.backup

ignore:
  use_defaults: true
  patterns:
    - "*.local"
    - "*.secret"

dotfile:
  translate: true
  prefix: "dot-"
  package_name_mapping: true

output:
  format: text
  color: auto
  progress: true

operations:
  dry_run: false
  atomic: true
  max_parallel: 4
```

## Troubleshooting

### Config Not Being Used

**Symptom**: Changes to repository config don't take effect.

**Check**:
1. Is the config at `~/.dotfiles/.config/dot/config.yaml`?
2. Is `~/.dotfiles` your actual package directory?
3. Do you have a local override at `~/.config/dot/config.yaml`?

**Solution**:
```bash
# Verify repository config exists
ls -la ~/.dotfiles/.config/dot/config.yaml

# Check which config is being used
dot config list
```

### Different Package Directory

If you use a different package directory (not `~/.dotfiles`), update the path:

```yaml
# In your repository's .config/dot/config.yaml
directories:
  package: ~/my-dotfiles  # Must match actual location
```

Or use the `--package-dir` flag:

```bash
dot --package-dir ~/my-dotfiles list
```

### No Repository Config

If you don't have a repository config, `dot` falls back to:

1. XDG config (`~/.config/dot/config.yaml`)
2. Built-in defaults

This is fine for personal use, but sharing your repository works better with a repository config.

## Migration

### From XDG-Only Config

If you currently have config at `~/.config/dot/config.yaml`:

```bash
# Copy to repository
cp ~/.config/dot/config.yaml ~/.dotfiles/.config/dot/config.yaml

# Add to repository
cd ~/.dotfiles
git add .config/dot/config.yaml
git commit -m "feat(config): add repository configuration"

# Optional: Remove local config to use only repository config
rm ~/.config/dot/config.yaml
```

### From No Config

If you're using defaults:

```bash
# Generate default config in repository
mkdir -p ~/.dotfiles/.config/dot
dot config init --output ~/.dotfiles/.config/dot/config.yaml

# Edit as needed
$EDITOR ~/.dotfiles/.config/dot/config.yaml

# Commit
cd ~/.dotfiles
git add .config/dot/config.yaml
git commit -m "feat(config): add dot configuration"
```

## Best Practices

### 1. Use Repository Config for Shared Settings

Put configuration that should be the same across all machines in the repository:

- Package directory location
- Target directory structure  
- Symlink mode (relative vs absolute)
- Package name mapping preferences

### 2. Use Local Config for Machine-Specific Settings

Put machine-specific overrides in `~/.config/dot/config.yaml`:

- Custom target directories
- Machine-specific ignore patterns
- Logging preferences

### 3. Commit Repository Config

Always commit `.config/dot/config.yaml` to your repository so others benefit:

```bash
git add .config/dot/config.yaml
git commit -m "feat(config): add dot configuration"
```

### 4. Don't Manage the Config File

Don't add `.config/dot/config.yaml` as a managed dotfile package. It should live in the repository root, not be symlinked.

**Wrong**:
```
~/.dotfiles/
└── dot-config/          # ❌ Don't do this
    └── .config/dot/config.yaml
```

**Right**:
```
~/.dotfiles/
└── .config/dot/config.yaml  # ✅ Regular file in repo root
```

## See Also

- [Configuration Management](04-configuration.md) - Complete configuration reference
- [Installation Guide](02-installation.md) - Setting up dot
- [Quickstart](03-quickstart.md) - Getting started with dot

