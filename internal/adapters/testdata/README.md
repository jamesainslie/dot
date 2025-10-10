# Test Fixtures for Adapters Package

## test-repo

A hermetic git repository fixture containing realistic dotfile packages for testing git clone operations without network dependencies.

### Structure

The test repository mirrors a real-world dotfiles repository with multiple packages:

**Packages:**
- `dot-zsh`: Shell configuration (zshrc, zshenv)
- `dot-git`: Git configuration (config, global ignore)
- `dot-vim`: Vim editor configuration (vimrc, colors)
- `dot-ssh`: SSH client configuration (config)
- `dot-tmux`: Terminal multiplexer configuration (feature-branch only)

**Files:**
- `README.md`: Repository documentation
- `.dotbootstrap.yaml`: Installation profiles (minimal, default, full)

**Branches:**
- `main`: Base configuration with zsh, git, vim, and ssh packages
- `feature-branch`: Adds tmux package to full profile

### Usage

Tests use the `getTestRepoURL(t)` helper function which returns a `file://` URL to this repository:

```go
url := getTestRepoURL(t)
err := cloner.Clone(ctx, url, targetPath, opts)

// Verify dotfile packages were cloned
assert.DirExists(t, filepath.Join(targetPath, "dot-zsh"))
assert.FileExists(t, filepath.Join(targetPath, ".dotbootstrap.yaml"))
```

### Benefits

- **Offline**: Tests run without network access
- **Realistic**: Contains actual dotfile configurations
- **Deterministic**: Same test results every time
- **Fast**: Local clones complete in milliseconds
- **Reliable**: No network flakiness or external dependencies

### Maintenance

To modify the test repository:

1. Navigate to `internal/adapters/testdata/test-repo`
2. Switch to the desired branch: `git checkout main` or `git checkout feature-branch`
3. Make changes using standard git commands
4. Commit changes with descriptive messages
5. Verify tests still pass: `go test ./internal/adapters -run TestGoGitCloner`

Example of adding a new package:

```bash
cd internal/adapters/testdata/test-repo
mkdir dot-bash
echo '# Bash config' > dot-bash/bashrc
git add dot-bash
git commit -m "feat: add bash package"
```

### TODO

**Windows Support**: The test repository currently contains Unix-focused dotfile packages (zsh, vim, ssh). Consider adding Windows-specific packages to test cross-platform scenarios:

- `dot-powershell`: PowerShell profile configuration
- `dot-wsl`: Windows Subsystem for Linux settings
- `dot-windows-terminal`: Windows Terminal configuration
- `dot-git`: Already cross-platform, but could add Windows-specific settings

This would ensure the cloner and package management work correctly on Windows machines with appropriate file paths, line endings, and configuration locations.
