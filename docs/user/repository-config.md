# Repository Configuration

## Overview

`dot` automatically looks for configuration in your dotfiles repository first, before checking your local XDG configuration directory. This solves the circular dependency problem where your dot configuration is itself managed by dot.

## How It Works

When you run any `dot` command, the first of these that exists is used:

1. **Repository config**: `<packageDir>/.config/dot/config.yaml`
2. **User config**: `~/.config/dot/config.yaml`
3. **Built-in defaults**

The repository config takes precedence, and it replaces the user config rather than merging with
it. This means your dotfiles repository can define its own configuration for how it should be
managed.

`<packageDir>` is the value of `--dir` when that flag is given; otherwise the lookup checks
`~/.dotfiles`. A repository in any other location is only found when `--dir` names it.

## Setup

### 1. Add Config to Your Repository

Create a configuration file in your dotfiles repository:

```bash
mkdir -p ~/.dotfiles/.config/dot
$EDITOR ~/.dotfiles/.config/dot/config.yaml
```

Example configuration. Path values are not expanded, so write them in full and substitute your own
home directory for `/home/alice`:

```yaml
directories:
  package: /home/alice/.dotfiles
  target: /home/alice
  manifest: /home/alice/.local/share/dot/manifest

symlinks:
  backup: true
  backup_dir: /home/alice/.dotfiles.backup

dotfile:
  translate: true
  package_name_mapping: true
```

A leading `~` or a `$VAR` reference is stored literally and resolved against the working directory,
which produces a directory named `~` or `$HOME` rather than the intended location. This makes a
committed repository config machine-specific: see Machine-Specific Settings below.

### 2. Commit and Share

Commit the configuration to your repository:

```bash
cd ~/.dotfiles
git add .config/dot/config.yaml
git commit -m "feat(config): add dot configuration"
git push
```

### 3. Clone on New Machines

`dot clone` behaves like `git clone`: without `--dir` it creates a directory named after the
repository, under the current working directory. Pass `--dir` explicitly so the repository lands
where the automatic config lookup will find it:

```bash
dot clone --dir ~/.dotfiles https://github.com/yourname/dotfiles
```

`dot` will:
1. Clone into `~/.dotfiles`
2. Find `~/.dotfiles/.config/dot/config.yaml`
3. Use that configuration automatically
4. Install packages according to the repository's settings

If you clone into some other location, subsequent commands will not find the repository config
unless you pass `--dir` to each of them, because the automatic lookup only checks `~/.dotfiles`.

## Benefits

### No Circular Dependency

The config file lives in the repository root at `.config/dot/config.yaml`. It is a regular file, not a managed symlink. This means:

- Config is available immediately after clone
- No chicken-and-egg problem

### Single Source of Truth

Your repository defines how it should be managed:

- Package name mapping and dotfile translation
- Ignore patterns
- Backup behaviour
- Output and logging preferences

Everyone who clones your repository gets the same configuration.

### Machine-Specific Settings

The two files are not layered. If the repository config exists it is loaded in full and
`~/.config/dot/config.yaml` is not read at all, so a partial machine-local override is not possible
while a repository config is present.

This matters for paths in particular. Because `directories.package` and `directories.target` must
be absolute and are not expanded, a committed repository config pins them to one machine's layout.
Either omit those two keys from the repository config and supply them per machine, or accept that
every machine must use the same paths.

For per-machine differences, use command-line flags or the bound `DOT_*` environment variables,
both of which take precedence over either file:

```bash
export DOT_DIRECTORIES_TARGET=/custom/path
dot --dir ~/.dotfiles --target ~ manage vim
```

Shell expansion applies on the command line, so `~` works there even though it does not work inside
the configuration file.

## Configuration Precedence

1. **Command-line flags** (highest priority)
2. **Environment variables** (`DOT_*`)
3. **Repository config** (`<packageDir>/.config/dot/config.yaml`)
4. **User config** (`~/.config/dot/config.yaml`)
5. **Built-in defaults** (lowest priority)

Levels 3 and 4 are mutually exclusive. When the repository config exists, the user config is not
read.

## Examples

### Basic Repository Config

```yaml
# .config/dot/config.yaml in your repository
directories:
  package: /home/alice/.dotfiles
  target: /home/alice

dotfile:
  package_name_mapping: true
```

### Advanced Configuration

```yaml
# .config/dot/config.yaml
directories:
  package: /home/alice/my-dotfiles
  target: /home/alice
  manifest: /home/alice/.local/share/dot/manifest

logging:
  level: INFO
  format: text

symlinks:
  mode: relative
  folding: true
  backup_dir: /home/alice/.dotfiles.backup

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

Several of these keys are accepted and validated but currently inert: `symlinks.mode`,
`symlinks.folding`, and `operations.max_parallel`. See
[Configuration Reference](04-configuration.md) for the full list.

## Troubleshooting

### Config Not Being Used

**Symptom**: Changes to repository config don't take effect.

**Check**:
1. Is the config at `<packageDir>/.config/dot/config.yaml`?
2. Is `~/.dotfiles` your actual package directory, or are you passing `--dir`?
3. Did you misspell or misnest a key? Unknown keys are discarded without a warning.

**Solution**:
```bash
# Verify repository config exists
ls -la ~/.dotfiles/.config/dot/config.yaml
```

Note that `dot config list` always reads the user config at `dot config path` and does not follow
the repository-first lookup. It cannot be used to confirm which file the other commands are using;
inspect the repository file directly.

### Different Package Directory

If you use a different package directory (not `~/.dotfiles`), update the path:

```yaml
# In your repository's .config/dot/config.yaml
directories:
  package: /home/alice/my-dotfiles  # Absolute, must match actual location
```

Or use the `--dir` flag (short form `-d`):

```bash
dot --dir ~/my-dotfiles list
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

If you're using defaults, generate the file at the destination by pointing `DOT_CONFIG` at it
(`dot config init` has no `--output` flag; it always writes to the resolved config path):

```bash
# Generate default config in repository
mkdir -p ~/.dotfiles/.config/dot
DOT_CONFIG=~/.dotfiles/.config/dot/config.yaml dot config init

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
- Ignore patterns
- Package name mapping preferences

### 2. Use Flags or Environment Variables for Machine-Specific Settings

Because the repository config replaces the user config rather than merging with it, express
per-machine differences through flags (`--target`, `--dir`) or bound `DOT_*` variables, not through
a second file:

- Custom target directories: `--target` or `DOT_DIRECTORIES_TARGET`
- Machine-specific ignore patterns: `--ignore`
- Logging preferences: `-v` or `DOT_LOGGING_LEVEL`

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
└── dot-config/          # Do not do this
    └── .config/dot/config.yaml
```

**Right**:
```
~/.dotfiles/
└── .config/dot/config.yaml  # Regular file in repo root
```

## See Also

- [Configuration Management](04-configuration.md) - Complete configuration reference
- [Installation Guide](02-installation.md) - Setting up dot
- [Quickstart](03-quickstart.md) - Getting started with dot

