# Common Workflows

Real-world usage patterns and workflows for dot.

## Initial Dotfiles Setup

### Scenario: First-Time Setup

Starting from scratch with no existing dotfiles repository.

**Steps**:

1. Create dotfiles repository
2. Create first package
3. Install package
4. Commit to version control

```bash
# Create repository
mkdir ~/dotfiles
cd ~/dotfiles
git init

# Create vim package. The package name maps to a target subdirectory, so
# dot-vim places files under ~/.vim/.
mkdir dot-vim
cat > dot-vim/dot-vimrc << 'EOF'
set number
syntax on
EOF

# Install: creates ~/.vim/.vimrc -> ~/dotfiles/dot-vim/dot-vimrc
dot manage dot-vim

# Commit
git add dot-vim/
git commit -m "feat(vim): add initial configuration"
git remote add origin https://github.com/username/dotfiles.git
git push -u origin main
```

**Package naming**: the package name becomes a subdirectory of the target. A
name beginning with `dot-` is translated to a leading dot, so `dot-vim`
targets `~/.vim/` and a plain `vim` targets `~/vim/`. See
[Command Reference](05-commands.md) for the full mapping table.

## Multi-Machine Synchronization

### New Machine Setup

```bash
# Clone dotfiles
git clone https://github.com/username/dotfiles.git ~/dotfiles
cd ~/dotfiles

# Install packages
dot manage dot-vim dot-zsh dot-tmux

# Verify
dot status
```

### Keeping Machines in Sync

On machine with changes:
```bash
cd ~/dotfiles
vim dot-vim/dot-vimrc
dot remanage dot-vim
git add dot-vim/
git commit -m "feat(vim): update configuration"
git push
```

On other machines:
```bash
cd ~/dotfiles
git pull
dot remanage dot-vim dot-zsh dot-tmux
```

## Conflict Resolution

### Backup and Compare

Conflict policy is set in the configuration file, not on the command line.

```bash
# Conflicts detected
dot manage dot-vim
# Error: conflict at ~/.vim/.vimrc

# Enable the backup policy
dot config set symlinks.backup true

# Retry: the conflicting file is moved into the backup directory
dot manage dot-vim

# Compare versions. Backups are named
# <filename>.<hash>.<timestamp> under <target>/.dot-backup.
ls ~/.dot-backup
diff ~/.dot-backup/.vimrc.3f9a1c07.20260727-121603 ~/dotfiles/dot-vim/dot-vimrc

# Merge if needed
vim ~/dotfiles/dot-vim/dot-vimrc
dot remanage dot-vim
```

### Adopt Existing Files

**Single File Adoption**:

```bash
# Conflict exists
dot manage dot-vim
# Error: conflict at ~/.vim/.vimrc

# Adopt instead (explicit package name)
dot adopt dot-vim ~/.vimrc

# Edit in package
vim ~/dotfiles/dot-vim/dot-vimrc
git add dot-vim/
git commit -m "feat(vim): adopt existing configuration"
```

**Auto-Naming (Single File)**:

```bash
# Let dot derive package name from filename
dot adopt ~/.vimrc
# Creates package: dot-vimrc

# Or for a directory
dot adopt ~/.ssh
# Creates package: dot-ssh
```

**Multiple Related Files**:

Shell globs expand before dot runs. With two or more arguments the first is
taken as the package name, so a glob must be preceded by an explicit package
name.

```bash
# Adopt all git-related config files into a single package
dot adopt dot-git .git*
# Shell expands to: dot adopt dot-git .gitconfig .gitignore .git-credentials

# Or with zsh configs
dot adopt dot-zsh .zsh*
# Adopts .zshrc .zshenv .zprofile into package dot-zsh

# Commit
git add dot-git/ dot-zsh/
git commit -m "feat(git,zsh): adopt existing configurations"
```

Omitting the package name is a mistake: `dot adopt .git*` treats the first
expanded filename as the package name and adopts the remaining files into it.
There is no common-prefix derivation.

## Testing New Packages

### Dry-Run Testing

```bash
# Create package
mkdir ~/dotfiles/test-package
echo "test" > ~/dotfiles/test-package/dot-testrc

# Preview
dot --dry-run manage test-package

# Install if satisfied
dot manage test-package

# Remove if not needed
dot unmanage test-package
rm -rf ~/dotfiles/test-package
```

## Package Updates

### Update Single Package

```bash
# Edit configuration
vim ~/dotfiles/dot-vim/dot-vimrc

# Preview changes
dot --dry-run remanage dot-vim

# Apply
dot remanage dot-vim

# Verify
dot status dot-vim

# Commit
git add dot-vim/
git commit -m "feat(vim): add plugins"
git push
```

### Bulk Updates

```bash
# Pull changes
cd ~/dotfiles
git pull

# Update all packages (incremental)
dot remanage dot-vim dot-zsh dot-tmux dot-git

# Verify
dot doctor
```

## Backup and Recovery

### Create Backup

The manifest lives in the manifest directory, by default
`~/.local/share/dot/manifest/`, alongside a `.dot-manifest.lock` file.

```bash
# Backup before changes
tar -czf ~/dotfiles-backup-$(date +%Y%m%d).tar.gz \
    ~/dotfiles ~/.local/share/dot/manifest

# Make changes
# ...

# Restore if needed
tar -xzf ~/dotfiles-backup-20251007.tar.gz
dot remanage dot-vim dot-zsh
```

### Disaster Recovery

```bash
# Reinstall dot
VERSION=$(curl -fsSL https://api.github.com/repos/yaklabco/dot/releases/latest | jq -r .tag_name)
curl -fsSL "https://github.com/yaklabco/dot/releases/download/${VERSION}/dot_${VERSION#v}_$(uname -s)_$(uname -m).tar.gz" | tar xz
sudo mv dot /usr/local/bin/

# Clone dotfiles
git clone https://github.com/username/dotfiles.git ~/dotfiles

# Reinstall all
cd ~/dotfiles
dot manage $(ls -d */ | tr -d '/')

# Verify
dot status
dot doctor
```

## Migration from GNU Stow

Package names map to target subdirectories in dot but not in Stow. A Stow
package `vim` whose contents land in `~/` corresponds to a dot package named
`dot-vim` only if the contents belong under `~/.vim/`. Review the mapping
table in [Command Reference](05-commands.md) before migrating, or set
`dotfile.package_name_mapping: false` to preserve the Stow layout.

### Gradual Migration

```bash
# Phase 1: Test with one package
stow -D test-package
dot manage test-package

# Phase 2: Migrate all
cd ~/dotfiles
for pkg in */; do stow -D "$pkg"; done
dot manage $(ls -d */ | tr -d '/')

# Phase 3: Remove Stow
brew uninstall stow
```

## CI/CD Integration

### GitHub Actions Example

`.github/workflows/deploy.yml`:
```yaml
name: Deploy Dotfiles
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install dot
        run: |
          VERSION=$(curl -fsSL https://api.github.com/repos/yaklabco/dot/releases/latest | jq -r .tag_name)
          curl -fsSL "https://github.com/yaklabco/dot/releases/download/${VERSION}/dot_${VERSION#v}_Linux_x86_64.tar.gz" | tar xz
          sudo mv dot /usr/local/bin/
      - name: Deploy
        run: dot --batch manage dot-vim dot-zsh dot-tmux
      - name: Verify
        run: dot doctor
```

`dot doctor` exits 1 on warnings and 2 on errors, so the verify step fails the
job when the installation is unhealthy. Disable the startup update check in CI
with `check_on_startup: false`; it is enabled by default and adds up to three
seconds per invocation.

## Next Steps

- [Advanced Features](07-advanced.md): Deep dive into features
- [Troubleshooting Guide](08-troubleshooting.md): Solve common issues
- [Configuration Reference](04-configuration.md): Customize behavior

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

