# Release Workflow

## Overview

This document describes the release process for the `dot` project, including automated changelog generation, versioning, and deployment.

## Prerequisites

- Git commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) specification
- All tests pass (`make test`)
- All linters pass (`make lint`)
- Working branch is clean and up to date with `main`

## Automated Changelog

The project uses `git-chglog` to automatically generate `CHANGELOG.md` from git commit history. The changelog includes only commits that impact users:

**Included commit types:**
- `feat`: New features
- `fix`: Bug fixes
- `perf`: Performance improvements
- `refactor`: Code refactoring
- `build`: Build system changes
- `revert`: Reverted changes

**Excluded commit types:**
- `docs`: Documentation updates
- `test`: Test changes
- `chore`: Maintenance tasks
- `ci`: CI/CD changes
- `style`: Code formatting

### Configuration

Changelog configuration is stored in `.chglog/`:
- `config.yml`: Defines filters, commit groups, and repository metadata
- `CHANGELOG.tpl.md`: Template for changelog output

### Makefile Targets

#### Generate Changelog
```bash
make changelog
```
Generates `CHANGELOG.md` from all git commits.

#### Preview Next Version
```bash
make changelog-next
```
Shows what will be included in the next release without modifying files.

## Release Process

### Standard Release (Patch Version)

1. **Update changelog:**
   ```bash
   make changelog
   ```

2. **Review changes:**
   ```bash
   git diff CHANGELOG.md
   ```

3. **Commit changelog:**
   ```bash
   git add CHANGELOG.md .chglog/
   git commit -m "docs(changelog): update for v0.1.1 release"
   ```

4. **Create release:**
   ```bash
   make version-patch
   ```
   This runs tests and linters, then suggests the tag command.

5. **Tag release:**
   ```bash
   git tag -a v0.1.1 -m "Release v0.1.1"
   ```

6. **Push changes:**
   ```bash
   git push origin main
   git push origin v0.1.1
   ```

7. **GitHub Actions builds and publishes:**
   - GoReleaser creates GitHub Release with notes
   - Binaries compiled for all platforms
   - Homebrew tap updated automatically
   - Release artifacts uploaded

### Minor or Major Release

For minor version bumps:
```bash
make changelog
git add CHANGELOG.md
git commit -m "docs(changelog): update for v0.2.0 release"
make version-minor
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin main v0.2.0
```

For major version bumps:
```bash
make changelog
git add CHANGELOG.md
git commit -m "docs(changelog): update for v1.0.0 release"
make version-major
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin main v1.0.0
```

## Version Bump Examples

### Current version: v0.1.0

```bash
# Patch (v0.1.0 → v0.1.1): Bug fixes, small improvements
make version-patch

# Minor (v0.1.0 → v0.2.0): New features, backward compatible
make version-minor

# Major (v0.1.0 → v1.0.0): Breaking changes
make version-major
```

## Breaking Changes

For releases with breaking changes:

1. Ensure commit messages include `BREAKING CHANGE:` footer
2. Update migration documentation if needed
3. Use major version bump
4. Document migration path in release notes

Example commit:
```
feat(api): restructure configuration interface

The configuration API now uses a builder pattern for improved
type safety and flexibility.

BREAKING CHANGE: Configuration loading API changed.

Replace config.Load() with config.NewLoader().Load().
See docs/migration/v2.0.0.md for complete guide.
```

## Hotfix Process

For emergency fixes to released versions:

1. Create hotfix branch from tag:
   ```bash
   git checkout -b hotfix-0.1.1 v0.1.0
   ```

2. Make fix and commit:
   ```bash
   git commit -m "fix(critical): resolve security vulnerability"
   ```

3. Update changelog:
   ```bash
   make changelog
   git add CHANGELOG.md
   git commit -m "docs(changelog): update for v0.1.1 hotfix"
   ```

4. Tag and push:
   ```bash
   git tag -a v0.1.1 -m "Hotfix v0.1.1"
   git push origin v0.1.1
   ```

5. Merge back to main:
   ```bash
   git checkout main
   git merge hotfix-0.1.1
   git push origin main
   ```

## Release Checklist

- [ ] All tests pass (`make test`)
- [ ] All linters pass (`make lint`)
- [ ] Dependencies up to date (`make deps-verify`)
- [ ] Changelog updated (`make changelog`)
- [ ] Changelog reviewed and committed
- [ ] Version bumped and tagged
- [ ] Tag pushed to GitHub
- [ ] GitHub Release created by CI
- [ ] Homebrew formula updated
- [ ] Installation instructions tested
- [ ] Documentation updated if needed

## Troubleshooting

### Changelog Missing Commits

Check commit message format:
```bash
git log --oneline --format="%s" v0.1.0..HEAD
```

Ensure commits match pattern: `type(scope): description`

### Tag Already Exists

Delete local and remote tag:
```bash
git tag -d v0.1.1
git push origin :refs/tags/v0.1.1
```

### Release Failed in CI

Check GitHub Actions logs:
1. Navigate to repository on GitHub
2. Click "Actions" tab
3. Select failed workflow
4. Review error logs
5. Fix issue and re-tag

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [git-chglog Documentation](https://github.com/git-chglog/git-chglog)
- [GoReleaser Documentation](https://goreleaser.com/)
