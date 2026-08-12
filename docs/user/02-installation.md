# Installation Guide

This guide covers all methods for installing dot on supported platforms.

## System Requirements

### Operating Systems

- **Linux**: All major distributions (Ubuntu, Debian, Fedora, Arch, Alpine, etc.)
- **macOS**: 10.15 (Catalina) or later
- **BSD**: FreeBSD, OpenBSD, NetBSD
- **Windows**: 10 or later (with symlink support)

### Filesystem Requirements

dot requires filesystem support for symbolic links:

**Fully Supported**:
- ext4, btrfs, xfs (Linux)
- APFS, HFS+ (macOS)
- ZFS (all platforms)
- UFS (BSD)

**Limited Support**:
- FAT32, exFAT: No symlink support
- NFS, SMB: Symlinks work with caveats (ensure symlink support enabled)

### Hardware Requirements

- Architecture: amd64 (x86-64) and arm64 (aarch64) release binaries; Windows
  release binaries are amd64 only. Other architectures can be built from source.
- Memory: 50 MB minimum
- Disk: 10 MB for binary, additional space for packages

## Installation Methods

### Binary Releases (Recommended)

Pre-compiled binaries are available for all supported platforms.

#### Linux and macOS

Release archives are named `dot_<version>_<os>_<arch>.tar.gz`, so the version must
be resolved before downloading.

```bash
# Resolve the latest version, then download the matching archive
VERSION=$(curl -fsSL https://api.github.com/repos/yaklabco/dot/releases/latest | grep -m1 '"tag_name"' | cut -d'"' -f4 | tr -d v)
OS=$(uname -s)
ARCH=$(uname -m)
curl -fsSL "https://github.com/yaklabco/dot/releases/latest/download/dot_${VERSION}_${OS}_${ARCH}.tar.gz" | tar xz

# Move to system path
sudo mv dot /usr/local/bin/

# Verify installation
dot --version
```

#### Windows

Download the Windows binary from [releases page](https://github.com/yaklabco/dot/releases):

1. Download `dot_<version>_Windows_x86_64.zip`
2. Extract to desired location
3. Add directory to PATH environment variable
4. Open new terminal and run `dot --version`

**Windows Note**: Administrator privileges required for creating symlinks. Run terminal as administrator or enable Developer Mode.

### Package Managers

#### Homebrew (macOS and Linux)

```bash
# Add tap
brew tap yaklabco/tap

# Install
brew install dot

# Verify
dot --version
```

Homebrew is the only package manager currently supported. On Windows, Arch Linux,
and BSD, use the binary release or `go install` as described below.

### From Source

Requires Go 1.26.1 or later.

#### Quick Install

```bash
go install github.com/yaklabco/dot/cmd/dot@latest
```

Binary installed to `$GOPATH/bin` or `~/go/bin`.

#### Build from Repository

```bash
# Clone repository
git clone https://github.com/yaklabco/dot.git
cd dot

# Build
make build

# Install to system
sudo make install

# Verify
dot --version
```

#### Build Options

```bash
# Build for specific platform
GOOS=linux GOARCH=amd64 make build

# Build for all supported platforms
make build-all

# Run tests, coverage, lint, vet and vulnerability scan
make check
```

## Post-Installation Setup

### Shell Completion

Enable command completion for your shell.

#### Bash

```bash
# Generate completion script
dot completion bash > /etc/bash_completion.d/dot

# Or for user only
dot completion bash > ~/.local/share/bash-completion/completions/dot

# Reload shell
source ~/.bashrc
```

#### Zsh

```bash
# Generate completion script
dot completion zsh > "${fpath[1]}/_dot"

# Or add to .zshrc
echo 'source <(dot completion zsh)' >> ~/.zshrc

# Reload shell
source ~/.zshrc
```

#### Fish

```bash
# Generate completion script
dot completion fish > ~/.config/fish/completions/dot.fish

# Reload shell
source ~/.config/fish/config.fish
```

### Built-in Help

dot does not install man pages. Use the built-in help:

```bash
dot --help
dot manage --help
dot status --help
```

### Configuration

Create initial configuration file:

```bash
# Generate default configuration
dot config init

# Edit configuration
$EDITOR ~/.config/dot/config.yaml

# Show the effective configuration
dot config list
```

## Platform-Specific Notes

### Linux

#### Permissions

Standard users can create symlinks on all Linux distributions. No special setup required.

#### SELinux (Fedora, RHEL, CentOS)

If SELinux is enabled and blocking operations:

```bash
# Allow dot to manage symlinks
sudo setsebool -P allow_user_symlink_target 1

# Or add SELinux policy (consult SELinux documentation)
```

#### AppArmor (Ubuntu, Debian)

AppArmor typically does not restrict symlink creation.

### macOS

#### Symlink Permissions

Modern macOS versions allow symlink creation without special permissions.

#### Gatekeeper

Binary may be quarantined on first run:

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/dot

# Or allow via System Preferences: Security & Privacy
```

#### M1/M2 (Apple Silicon)

Use arm64 binary or universal binary. Intel binary (amd64) works via Rosetta 2 but native arm64 is recommended.

### Windows

#### Symlink Requirements

Windows 10/11 requires either:

1. **Administrator Privileges**: Run terminal as administrator
2. **Developer Mode**: Enable via Settings → Update & Security → For Developers

To enable Developer Mode:

1. Open Settings
2. Navigate to Update & Security → For Developers
3. Enable Developer Mode
4. Restart if prompted

#### Path Configuration

Add dot to PATH:

```powershell
# PowerShell (as administrator)
$path = [Environment]::GetEnvironmentVariable("Path", "Machine")
$path += ";C:\path\to\dot"
[Environment]::SetEnvironmentVariable("Path", $path, "Machine")
```

Or add via GUI:
1. Search "Environment Variables"
2. Edit System Variables → Path
3. Add dot directory
4. Restart terminal

#### WSL Consideration

For Windows Subsystem for Linux, install Linux version inside WSL environment rather than Windows version.

### BSD

No BSD packages are published. Build from source using the method above.

## Verification

Verify installation succeeded:

```bash
# Check version
dot --version

# Check help
dot --help

# Run basic command
dot status
```

Expected output from a release build:
```
dot version 0.6.5 (commit: a1b2c3d, built: 2026-03-18T00:00:00Z)
```

A binary built from a source checkout without release metadata prints:
```
dot version unknown (built from source)
```

## Upgrading

### Binary Releases

```bash
# Download latest release
VERSION=$(curl -fsSL https://api.github.com/repos/yaklabco/dot/releases/latest | grep -m1 '"tag_name"' | cut -d'"' -f4 | tr -d v)
curl -fsSL "https://github.com/yaklabco/dot/releases/latest/download/dot_${VERSION}_$(uname -s)_$(uname -m).tar.gz" | tar xz

# Replace existing binary
sudo mv dot /usr/local/bin/

# Verify new version
dot --version
```

### Package Managers

```bash
# Homebrew
brew upgrade dot
```

dot can also update itself; see [Updates and Version Management](10-updates.md).

```bash
dot upgrade --check-only
dot upgrade --yes
```

### From Source

```bash
cd dot
git pull origin main
make build
sudo make install
```

## Uninstallation

### Binary Installation

```bash
# Remove binary
sudo rm /usr/local/bin/dot

# Remove configuration (optional)
rm -rf ~/.config/dot

# Remove manifest and other state (optional, default XDG location)
rm -rf ~/.local/share/dot
```

### Package Managers

```bash
# Homebrew
brew uninstall dot
```

### Clean Removal

To remove all dot data:

```bash
# Remove installed symlinks first
dot unmanage --all --yes

# Remove binary
sudo rm /usr/local/bin/dot

# Remove user configuration
rm -rf ~/.config/dot

# Remove manifest and other state (default XDG location)
rm -rf ~/.local/share/dot

# Remove shell completion
rm /etc/bash_completion.d/dot  # bash
rm ~/.local/share/bash-completion/completions/dot
rm "${fpath[1]}/_dot"  # zsh
rm ~/.config/fish/completions/dot.fish  # fish
```

## Troubleshooting

### Command Not Found

Binary not in PATH:

```bash
# Check installation location
which dot

# Add to PATH if needed
export PATH="$PATH:/usr/local/bin"

# Make permanent by adding to shell rc file
echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
```

### Permission Denied (Linux/macOS)

Binary lacks execute permission:

```bash
chmod +x /usr/local/bin/dot
```

### Permission Denied (Windows)

Run terminal as administrator or enable Developer Mode (see Windows section above).

### Version Mismatch

Multiple installations exist:

```bash
# Find all installations
which -a dot

# Remove unwanted versions
sudo rm /path/to/old/dot
```

### Symlink Creation Fails

Filesystem does not support symlinks:

```bash
# Check filesystem type
df -T .

# Verify symlink support
ln -s /tmp/test /tmp/testlink
```

If symlinks are unsupported, dot cannot function. Use alternative filesystem or tool.

## Next Steps

- [Quick Start Tutorial](03-quickstart.md): Get started with basic operations
- [Configuration Reference](04-configuration.md): Configure dot for your environment
- [Command Reference](05-commands.md): Learn all available commands

## Navigation

**[↑ Back to Main README](../../README.md)** | [User Guide Index](index.md) | [Documentation Index](../README.md)

