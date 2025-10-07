# Phase 23: Homebrew Tap Distribution

## Overview

Establish automated Homebrew distribution for the `dot` CLI tool through a dedicated tap repository with GoReleaser integration. This enables users to install and update `dot` using standard Homebrew commands.

## Objectives

1. Create dedicated tap repository at `jamesainslie/homebrew-dot`
2. Configure GoReleaser in main `dot` repository for automated releases
3. Implement automated formula updates on version releases
4. Document installation and maintenance procedures
5. Validate installation workflow end-to-end

## Repository Architecture

### Main Repository (`jamesainslie/dot`)
- Contains GoReleaser configuration (`.goreleaser.yml`)
- Manages release workflow
- Pushes updates to tap repository automatically

### Tap Repository (`jamesainslie/homebrew-dot`)
- Contains Homebrew formula (`Formula/dot.rb`)
- Receives automated updates from GoReleaser
- Follows Homebrew tap naming conventions

## Implementation Tasks

### Task 1: Create Tap Repository

**Location:** `/Users/jamesainslie/Development/homebrew-dot`  
**Remote:** `https://github.com/jamesainslie/homebrew-dot.git`

#### Subtasks
1. Initialize repository structure
   ```
   homebrew-dot/
   ├── Formula/
   │   └── dot.rb
   ├── README.md
   └── LICENSE
   ```

2. Create initial Formula template
   - Define basic formula structure
   - Set package metadata (description, homepage, license)
   - Configure installation paths and completion files
   - Add test block for basic functionality validation

3. Document tap usage in README
   - Installation instructions: `brew tap jamesainslie/dot`
   - Install command: `brew install dot`
   - Update procedures
   - Uninstallation steps

4. Copy LICENSE from main repository

5. Create repository on GitHub
   - Set repository description
   - Add relevant topics (homebrew, tap, dotfiles, golang)
   - Configure branch protection if needed

### Task 2: Configure GoReleaser in Main Repository

**Location:** `/Volumes/Development/dot/.goreleaser.yml`

#### Subtasks
1. Install/verify GoReleaser
   ```bash
   brew install goreleaser/tap/goreleaser
   ```

2. Create `.goreleaser.yml` configuration with:
   - Build configuration for multiple platforms (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64)
   - Binary naming and paths
   - Homebrew tap integration pointing to `jamesainslie/homebrew-dot`
   - Archive configuration
   - Checksum generation
   - Changelog generation from commit messages
   - Release notes configuration

3. Configure Homebrew-specific settings:
   - Package name: `dot`
   - Description from project README
   - Homepage URL
   - License identifier
   - Installation configuration
   - Shell completion files (bash, zsh, fish)
   - Dependencies (if any)
   - Caveats for post-installation instructions

4. Add GitHub token configuration
   - Document required GitHub token scopes (repo, write:packages)
   - Environment variable: `GITHUB_TOKEN` or `GORELEASER_GITHUB_TOKEN`

5. Test configuration locally
   ```bash
   goreleaser check
   goreleaser build --snapshot --clean
   ```

### Task 3: GitHub Actions Release Workflow

**Location:** `/Volumes/Development/dot/.github/workflows/release.yml`

#### Subtasks
1. Create release workflow file
   - Trigger on pushed tags matching `v*.*.*`
   - Set up Go environment (1.25.1)
   - Run tests before release
   - Execute GoReleaser with GitHub token
   - Upload artifacts

2. Configure required secrets
   - `GITHUB_TOKEN` (automatically provided) or custom token
   - Tap repository write access verification

3. Add release checklist to CONTRIBUTING.md
   - Pre-release testing requirements
   - Version bump procedures
   - Tag creation standards
   - Post-release verification steps

### Task 4: Formula Configuration Details

**File:** `homebrew-dot/Formula/dot.rb`

#### Required Components
1. Formula class definition
   - Inherits from `Formula`
   - Descriptive comment block

2. Metadata
   - `desc`: Short description (max 80 chars)
   - `homepage`: Project URL
   - `url`: Source tarball URL (populated by GoReleaser)
   - `sha256`: Archive checksum (populated by GoReleaser)
   - `license`: "MIT"
   - `version`: Semantic version (populated by GoReleaser)

3. Dependencies
   - Runtime dependencies (if any)
   - Build dependencies (Go, if building from source)

4. Installation block
   - Binary installation: `bin.install "dot"`
   - Completion installation:
     - `bash_completion.install "completions/dot.bash" => "dot"`
     - `zsh_completion.install "completions/dot.zsh" => "_dot"`
     - `fish_completion.install "completions/dot.fish"`
   - Man page installation (if applicable)

5. Test block
   - Version check: `assert_match version.to_s, shell_output("#{bin}/dot --version")`
   - Basic command validation
   - Ensure binary is executable

6. Caveats (optional)
   - Configuration file location notice
   - XDG specification compliance note
   - Shell completion activation instructions

### Task 5: Version Management

#### Subtasks
1. Document versioning strategy
   - Follow Semantic Versioning 2.0.0
   - Version format: `v{major}.{minor}.{patch}`
   - Pre-release suffix handling (alpha, beta, rc)

2. Create version bump procedure
   - Update `CHANGELOG.md`
   - Create annotated tag: `git tag -a v0.x.x -m "Release v0.x.x"`
   - Push tag: `git push origin v0.x.x`
   - Verify GitHub Actions workflow execution

3. Add version constant in code (if not present)
   - Location: `cmd/dot/root.go` or dedicated version file
   - Injected at build time via ldflags

### Task 6: Documentation Updates

#### Main Repository
1. Update `README.md`
   - Add Homebrew installation section
   - Include tap installation instructions
   - Document installation verification
   - Add badge: `![Homebrew](https://img.shields.io/badge/homebrew-available-orange)`

2. Update `CONTRIBUTING.md`
   - Release process documentation
   - GoReleaser workflow explanation
   - Formula maintenance procedures

3. Create `docs/user/installation-homebrew.md`
   - Detailed Homebrew installation guide
   - Troubleshooting common issues
   - Platform-specific notes
   - Upgrade and uninstall procedures

#### Tap Repository
1. Comprehensive `README.md`
   - Tap purpose and scope
   - Installation instructions
   - Formula maintenance information
   - Link to main repository

2. Formula documentation
   - Inline comments in `dot.rb`
   - Explanation of custom configurations
   - Testing procedures

### Task 7: Testing and Validation

#### Local Testing
1. Test GoReleaser configuration
   ```bash
   goreleaser check
   goreleaser release --snapshot --clean
   ```

2. Validate generated artifacts
   - Verify binary builds for all platforms
   - Check archive contents
   - Validate checksums

3. Test formula locally
   ```bash
   brew install --build-from-source Formula/dot.rb
   dot --version
   brew test dot
   brew uninstall dot
   ```

#### CI/CD Testing
1. Create test release (pre-release or draft)
   - Tag with `-rc1` suffix
   - Verify GoReleaser execution
   - Check tap repository updates

2. Validate formula generation
   - Inspect generated `dot.rb` in tap repo
   - Verify URLs and checksums
   - Test installation from tap

3. End-to-end installation test
   ```bash
   brew tap jamesainslie/dot
   brew install dot
   dot --version
   dot --help
   ```

#### Platform Testing
1. Test on macOS (Intel and Apple Silicon)
2. Test on Linux (if applicable)
3. Verify shell completions activation

### Task 8: Release Checklist

Create checklist for each release:

1. **Pre-Release**
   - [ ] All tests passing (`make test`)
   - [ ] Linters passing (`make lint`)
   - [ ] `CHANGELOG.md` updated
   - [ ] Version bumped in relevant files
   - [ ] Documentation reflects new version

2. **Release**
   - [ ] Create annotated tag
   - [ ] Push tag to trigger workflow
   - [ ] Monitor GitHub Actions execution
   - [ ] Verify artifacts generated

3. **Post-Release**
   - [ ] Verify formula updated in tap repository
   - [ ] Test installation: `brew install jamesainslie/dot/dot`
   - [ ] Verify version: `dot --version`
   - [ ] Create GitHub release notes
   - [ ] Announce release (if applicable)

4. **Validation**
   - [ ] Fresh installation works
   - [ ] Upgrade from previous version works
   - [ ] Shell completions installed correctly
   - [ ] No broken links in documentation

## Configuration Examples

### .goreleaser.yml Template
```yaml
project_name: dot

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: dot
    binary: dot
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: dot
    format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}

checksum:
  name_template: 'checksums.txt'

changelog:
  use: github
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

brews:
  - name: dot
    repository:
      owner: jamesainslie
      name: homebrew-dot
      token: "{{ .Env.GITHUB_TOKEN }}"
    
    directory: Formula
    
    homepage: "https://github.com/jamesainslie/dot"
    description: "Dotfile management tool with XDG Base Directory specification support"
    license: "MIT"
    
    install: |
      bin.install "dot"
    
    test: |
      system "#{bin}/dot", "--version"
```

### GitHub Actions Workflow Template
```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25.1'
      
      - name: Run tests
        run: make test
      
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Dependencies

### Tools Required
- GoReleaser v2.x
- GitHub CLI (optional, for automation)
- Homebrew (for testing)

### GitHub Permissions
- Write access to `jamesainslie/dot`
- Write access to `jamesainslie/homebrew-dot`
- GitHub token with `repo` and `write:packages` scopes

## Success Criteria

1. Tap repository created and accessible
2. GoReleaser successfully builds and releases
3. Formula automatically updated on new releases
4. Users can install via `brew install jamesainslie/dot/dot`
5. Installation includes binary and shell completions
6. `brew test dot` passes
7. Documentation complete and accurate
8. Release workflow documented and repeatable

## Risks and Mitigations

### Risk: GoReleaser Configuration Errors
**Mitigation:** Test with `--snapshot` flag before actual release

### Risk: Formula Update Failures
**Mitigation:** Verify GitHub token permissions; test with pre-release

### Risk: Platform-Specific Build Issues
**Mitigation:** Test snapshot builds on target platforms

### Risk: Version String Mismatches
**Mitigation:** Use ldflags for version injection; validate in tests

## Timeline Estimate

- Task 1 (Tap Repository): 1 hour
- Task 2 (GoReleaser Config): 2 hours
- Task 3 (GitHub Actions): 1 hour
- Task 4 (Formula Details): 1 hour
- Task 5 (Version Management): 30 minutes
- Task 6 (Documentation): 2 hours
- Task 7 (Testing): 2 hours
- Task 8 (Release Checklist): 30 minutes

**Total:** ~10 hours

## References

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)

## Next Steps

After completion of Phase 23:
1. Monitor initial user installations and feedback
2. Consider submitting to Homebrew core when adoption grows
3. Evaluate additional distribution channels (apt, yum, snap, etc.)
4. Set up automated security scanning for releases

