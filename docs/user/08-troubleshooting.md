# Troubleshooting Guide

Solutions for common issues and problems.

## Installation Issues

### Command Not Found

**Problem**: `dot: command not found`

**Diagnosis**:
```bash
which dot
echo $PATH
```

**Solutions**:

1. **Binary not in PATH**:
```bash
# Add to PATH
export PATH="$PATH:/usr/local/bin"

# Make permanent
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
source ~/.bashrc
```

2. **Binary not installed**:
```bash
# Verify installation location
ls -la /usr/local/bin/dot

# Reinstall if missing. Release archives are named
# dot_<version>_<os>_<arch>.tar.gz, so the version must be resolved first.
VERSION=$(curl -fsSL https://api.github.com/repos/yaklabco/dot/releases/latest | jq -r .tag_name)
curl -fsSL "https://github.com/yaklabco/dot/releases/download/${VERSION}/dot_${VERSION#v}_$(uname -s)_$(uname -m).tar.gz" | tar xz
sudo mv dot /usr/local/bin/
```

### Permission Denied

**Problem**: `permission denied: dot`

**Solution**:
```bash
# Add execute permission
chmod +x /usr/local/bin/dot
```

### Version Mismatch

**Problem**: Multiple dot installations with different versions

**Diagnosis**:
```bash
which -a dot
```

**Solution**:
```bash
# Remove old versions
sudo rm /path/to/old/dot

# Verify
dot --version
```

## Operation Failures

### Symlink Creation Failed

**Problem**: Cannot create symlinks

**Common Causes**:

1. **No symlink support** (FAT32, exFAT):
```bash
# Check filesystem
df -T .
```

**Solution**: Use filesystem with symlink support (ext4, APFS, etc.)

2. **Permission denied**:
```bash
# Check permissions
ls -la ~/

# Fix permissions
chmod u+w ~/
```

3. **Windows without privileges**:

**Solution**: Enable Developer Mode or run as administrator

### File Conflicts

**Problem**: `Error: conflict at ~/.vim/.vimrc`

**Diagnosis**:
```bash
# Check what exists
ls -la ~/.vim/.vimrc

# If symlink, check target
readlink ~/.vim/.vimrc
```

**Solutions**:

1. **Backup existing file**. There is no conflict flag; the policy comes from
   the configuration file.
```bash
dot config set symlinks.backup true
dot manage dot-vim

# Backups are <filename>.<hash>.<timestamp> under <target>/.dot-backup
ls ~/.dot-backup
diff ~/.dot-backup/.vimrc.3f9a1c07.20260727-121603 ~/dotfiles/dot-vim/dot-vimrc
```

2. **Adopt existing file**:
```bash
dot adopt dot-vim ~/.vimrc
```

3. **Remove conflicting file**:
```bash
rm ~/.vim/.vimrc
dot manage dot-vim
```

4. **Overwrite conflicts**:
```bash
dot config set symlinks.overwrite true
dot manage dot-vim
```

There is no skip policy reachable from the CLI or the configuration file.

### Package Not Found

**Problem**: `Error: package not found: dot-vim`

**Diagnosis**:
```bash
# Check package directory exists
ls -la ~/dotfiles/dot-vim

# Check current directory
pwd
```

**Solutions**:

1. **Wrong directory**:
```bash
cd ~/dotfiles
dot manage dot-vim
```

2. **Specify directory**:
```bash
dot --dir ~/dotfiles manage dot-vim
```

3. **Package doesn't exist**:
```bash
# Create package
mkdir ~/dotfiles/dot-vim
```

### Broken Symlinks

**Problem**: Symlinks point to non-existent targets or were accidentally deleted

**Diagnosis**:
```bash
# Check for broken links with doctor
dot doctor

# Find broken links manually
find ~ -xtype l

# Check specific link
ls -la ~/.vim/.vimrc
readlink ~/.vim/.vimrc
```

**Solutions**:

1. **Recreate missing links** (Recommended):
```bash
# Check what's broken. Use --detailed for the path and suggested fix.
dot doctor --detailed

# Remanage automatically detects and recreates missing links
dot remanage dot-vim

# For multiple packages
dot remanage dot-vim dot-zsh dot-tmux
```

The `remanage` command detects missing symlinks and recreates them even if the
package content has not changed.

2. **Fix target path if package moved**. All symlinks are absolute, so moving
   the package directory breaks every link:
```bash
# Update package location in config
dot config set directories.package /new/path

# Recreate links
dot remanage dot-vim
```

3. **Remove and recreate** (if remanage doesn't fix it):
```bash
dot unmanage dot-vim
dot manage dot-vim
```

## Manifest Issues

### Manifest Corrupted

**Problem**: `Error: cannot parse manifest`

The manifest lives in the manifest directory, by default
`~/.local/share/dot/manifest/.dot-manifest.json`, alongside a
`.dot-manifest.lock` file. Confirm the location with `dot config get
directories.manifest`.

**Diagnosis**:
```bash
# Check manifest
cat ~/.local/share/dot/manifest/.dot-manifest.json

# Validate JSON
jq . ~/.local/share/dot/manifest/.dot-manifest.json
```

**Solution**:

There is no repair command. Delete the manifest and reinstall:

```bash
# Backup first
cp ~/.local/share/dot/manifest/.dot-manifest.json /tmp/dot-manifest.bak

# Remove
rm ~/.local/share/dot/manifest/.dot-manifest.json

# Reinstall packages
dot manage dot-vim dot-zsh dot-tmux
```

### Manifest Out of Sync

**Problem**: Manifest doesn't match filesystem

**Diagnosis**:
```bash
dot doctor --detailed
```

**Solution**:
```bash
# Remanage all packages. list emits an object, so address .packages[].
dot remanage $(dot list --format json | jq -r '.packages[].name')
```

## Configuration Issues

### Configuration Not Loading

**Problem**: Configuration settings not applied

**Diagnosis**:
```bash
# Show active configuration
dot config show

# Show configuration file path
dot config path

# Verify file exists
ls -la $(dot config path)
```

**Solutions**:

There is no validation subcommand. `dot config list` fails to load an invalid
file and reports the parse error, so it doubles as a syntax check.

1. **Check syntax**:
```bash
dot config list
```

2. **Check precedence**:
```bash
# Environment variables are DOT_ plus the key path with dots as underscores
unset DOT_OUTPUT_VERBOSITY

# Command-line flags override everything
dot manage dot-vim     # Uses config file
dot -v manage dot-vim  # -v overrides config
```

Every configuration key is bound to an environment variable: `DOT_` plus the
uppercased key with dots replaced by underscores, including `symlinks.backup_dir`,
`dotfile.package_name_mapping`, `output.table_style`, and the `update` and
`network` sections.

### Invalid Configuration

**Problem**: `Error: invalid configuration`

**Diagnosis**:
```bash
dot config list
```

**Common Errors**:

1. **Wrong key nesting**. Settings live under a section:
```yaml
# Wrong
packageDir: ~/dotfiles

# Correct
directories:
  package: ~/dotfiles
```

2. **Invalid paths**:
```yaml
directories:
  package: ~/dotfiles   # absolute or tilde-expanded, not "dotfiles"
```

3. **Invalid values**:
```yaml
symlinks:
  mode: relative        # relative or absolute
```

Run `dot config list` to see the merged result and `dot config path` to see
which file was read.

## Performance Issues

### Slow Operations

**Problem**: Operations take too long

**Diagnosis**:
```bash
# Profile operation
time dot -vv manage dot-vim

# Or capture a CPU profile
dot --cpu-profile cpu.pprof manage dot-vim
```

**Solutions**:

1. **Use remanage instead of manage**. Remanage skips packages whose content
   hash and links are unchanged:
```bash
dot remanage dot-vim dot-zsh dot-tmux
```

2. **Limit doctor scanning**:
```bash
dot doctor --scan-mode=off
```

3. **Optimize ignore patterns**. Patterns are globs; regular expressions are
   not supported and are matched literally:
```yaml
ignore:
  patterns:
    - ".git"
    - "node_modules"
```

4. **Disable the startup update check**, which adds up to three seconds per
   invocation and is enabled by default:
```yaml
update:
  check_on_startup: false
```

Parallelism is fixed at the CPU count and is not configurable.

### High Memory Usage

**Problem**: dot consumes excessive memory

**Diagnosis**:
```bash
# Monitor memory during operation
/usr/bin/time -v dot manage large-package

# Or capture a heap profile
dot --mem-profile mem.pprof manage large-package
```

**Solutions**:

1. **Exclude large files**:
```bash
dot --max-file-size 10MB manage large-package
```

2. **Process packages in batches**:
```bash
# Instead of all at once
dot manage pkg1 pkg2 pkg3
dot manage pkg4 pkg5 pkg6
```

## Platform-Specific Issues

### macOS Issues

**Problem**: Gatekeeper blocks execution

**Solution**:
```bash
# Remove quarantine
xattr -d com.apple.quarantine /usr/local/bin/dot

# Or allow via System Preferences
```

**Problem**: Symlinks broken after upgrade

**Solution**:
```bash
# Remanage all packages
cd ~/dotfiles
dot remanage $(dot list --format json | jq -r '.packages[].name')
```

### Windows Issues

**Problem**: Symlink creation fails

**Solutions**:

1. **Enable Developer Mode**:
   - Settings → Update & Security → For Developers
   - Enable Developer Mode
   - Restart terminal

2. **Run as administrator**:
   - Right-click terminal
   - "Run as administrator"

3. **Check symlink support**:
```powershell
# Test symlink creation
New-Item -ItemType SymbolicLink -Path test -Target C:\Windows\System32
```

### Linux Issues

**Problem**: SELinux blocking operations

**Solution**:
```bash
# Allow symlinks
sudo setsebool -P allow_user_symlink_target 1
```

**Problem**: Permission denied on NFS

**Solution**:
```bash
# Check NFS mount options
mount | grep nfs

# Ensure no_root_squash option if needed
```

## Error Messages

### Common Errors and Solutions

#### "conflict at PATH: file exists"

**Cause**: File exists at target location

**Solutions**:
- Set `symlinks.backup: true` to preserve the existing file in the backup directory
- Use `dot adopt` to move file into package
- Remove conflicting file manually

#### "package not found: NAME"

**Cause**: Package directory doesn't exist

**Solutions**:
- Check package directory exists: `ls ~/dotfiles/NAME`
- Specify correct package directory: `--dir ~/dotfiles`
- Create package: `mkdir ~/dotfiles/NAME`

#### "permission denied"

**Cause**: Insufficient permissions

**Solutions**:
- Check file/directory permissions
- Ensure write access to target directory
- Run with appropriate privileges (not sudo unless necessary)

#### "manifest corrupted"

**Cause**: Invalid JSON in manifest file

**Solutions**:
- There is no repair command. Delete and recreate:
  `rm ~/.local/share/dot/manifest/.dot-manifest.json && dot manage ...`

#### "broken symlink"

**Cause**: Symlink target doesn't exist

**Solutions**:
- Remanage package: `dot remanage PACKAGE`
- Fix target location
- Remove and reinstall: `dot unmanage PACKAGE && dot manage PACKAGE`

#### "refusing to delete PATH: REASON"

**Cause**: Deletion is verified against the plan before it happens. dot refuses
to remove a regular file, or a symlink that now points somewhere other than
where the plan recorded.

**Solutions**:
- Inspect the path with `ls -la` and `readlink`
- Remove it manually if it is genuinely unwanted, then rerun the command

#### "cannot roll back operation on PATH"

**Cause**: A plan failed and one of the already-executed operations cannot be
reversed. Removing a directory tree is the common case.

**Solutions**:
- Read the accompanying summary, which reports how many operations could not be
  rolled back
- Run `dot doctor --detailed` to see the resulting state, then remanage the
  affected packages

## Diagnostic Procedures

### Health Check Procedure

```bash
# 1. Run doctor. Exits 0 healthy, 1 warnings, 2 errors.
dot doctor --detailed

# 2. Check configuration
dot config list
dot config path

# 3. Verify installation
dot --version

# 4. Check packages
dot status

# 5. List installed
dot list

# 6. Test with dry-run
dot --dry-run manage test-package
```

### Debug Procedure

```bash
# 1. Enable maximum verbosity
dot -vvv manage dot-vim 2>&1 | tee debug.log

# 2. Check system state
ls -la ~/dotfiles/
ls -la ~/

# 3. Check manifest
jq . ~/.local/share/dot/manifest/.dot-manifest.json

# 4. Verify symlinks
find ~ -type l -ls

# 5. Check configuration and which file it came from
dot config list
dot config path
```

### Recovery Procedure

```bash
# 1. Backup current state
tar -czf ~/dot-backup-$(date +%Y%m%d).tar.gz \
    ~/dotfiles ~/.local/share/dot/manifest

# 2. Unmanage all packages
dot unmanage --all --yes

# 3. Clean manifest
rm -f ~/.local/share/dot/manifest/.dot-manifest.json

# 4. Reinstall
cd ~/dotfiles
dot manage $(ls -d */ | tr -d '/')

# 5. Verify
dot doctor
```

## Getting Help

### Information to Provide

When reporting issues, include:

1. **Version information**:
```bash
dot --version
```

2. **Configuration**:
```bash
dot config list
```

3. **Error output**:
```bash
dot -vvv <command> 2>&1
```

4. **System information**:
```bash
uname -a
echo $SHELL
```

5. **Directory structure** (sanitized):
```bash
tree -L 2 ~/dotfiles
```

### Where to Get Help

- **Documentation**: Search this guide
- **GitHub Issues**: Report bugs
- **GitHub Discussions**: Ask questions
- **Doctor command**: `dot doctor` for automated diagnostics

## FAQ

**Q: Can I use dot with GNU Stow simultaneously?**

A: Yes, but not recommended. They may conflict. Migrate completely or keep separate package sets.

**Q: Does dot work on Windows?**

A: Limited support. Requires Developer Mode or administrator privileges. Symlink support varies by filesystem.

**Q: Can I use absolute and relative links together?**

A: No. All symlinks are absolute. There is no per-package link-mode setting and
no `.dotmeta` file; the only per-package file dot reads is `.dotignore`. The
`symlinks.mode` configuration key is accepted but has no effect.

**Q: What happens if I move my dotfiles directory?**

A: Every link breaks, because links are absolute. Update the configured
location and remanage all packages:

```bash
dot config set directories.package /new/path
dot remanage $(dot list --format json | jq -r '.packages[].name')
```

**Q: Can I nest packages?**

A: No, packages must be top-level directories in package directory.

**Q: Does dot follow symlinks in packages?**

A: No. Symlinks inside a package are skipped entirely; no corresponding link is
created in the target.

**Q: Why did `dot manage vim` put files in `~/vim/` instead of `~/`?**

A: The package name becomes a subdirectory of the target. Name the package
`dot-vim` to target `~/.vim/`, or set
`dotfile.package_name_mapping: false` to place package contents directly in the
target root. See [Command Reference](05-commands.md) for the mapping table.

## Next Steps

- [Glossary](09-glossary.md): Reference for technical terms
- [Command Reference](05-commands.md): Complete command documentation
- [Configuration Reference](04-configuration.md): Configuration options

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

