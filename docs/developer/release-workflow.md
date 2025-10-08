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

Both files must be committed to the repository for consistent changelog generation across all environments.

### Makefile Targets

#### Automated Release Targets

- `make version-patch`: Complete patch release workflow
- `make version-minor`: Complete minor release workflow
- `make version-major`: Complete major release workflow
- `make release-tag VERSION=v1.2.3`: Create release with specific version

#### Changelog Targets

- `make changelog`: Generate CHANGELOG.md from all git commits
- `make changelog-next`: Preview next version changelog without modifying files
- `make changelog-update`: Update and commit changelog (internal use)

#### Verification Targets

- `make release VERSION=v1.2.3`: Verify release readiness without creating tags
- `make check`: Run tests and linting

## Release Process

### Automated Release (Recommended)

The project includes automated release targets that handle the complete workflow, including changelog generation, quality checks, and git tagging.

#### Patch Release

```bash
make version-patch
```

This single command:
1. Runs all quality checks (tests, linting)
2. Generates and commits changelog
3. Creates git tag
4. Displays push instructions

#### Minor Release

```bash
make version-minor
```

#### Major Release

```bash
make version-major
```

#### Push Release

After running any version command above:
```bash
git push origin main
git push origin v0.1.1  # Replace with your version
```

GitHub Actions will automatically:
- Build binaries for all platforms
- Create GitHub Release with notes
- Update Homebrew tap
- Upload release artifacts

### Manual Release (Advanced)

For more control over the release process, use individual steps:

1. **Preview changelog:**
   ```bash
   make changelog-next
   ```

2. **Run quality checks:**
   ```bash
   make check
   ```

3. **Generate changelog:**
   ```bash
   make changelog
   ```

4. **Review changes:**
   ```bash
   git diff CHANGELOG.md
   ```

5. **Commit changelog:**
   ```bash
   git add CHANGELOG.md .chglog/
   git commit -m "docs(changelog): update for v0.1.1 release"
   ```

6. **Tag release:**
   ```bash
   git tag -a v0.1.1 -m "Release v0.1.1"
   ```

7. **Regenerate changelog with tag:**
   ```bash
   make changelog
   git add CHANGELOG.md
   git commit --amend --no-edit
   ```

8. **Push changes:**
   ```bash
   git push origin main
   git push origin v0.1.1
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

1. **Create hotfix branch from tag:**
   ```bash
   git checkout -b hotfix-v0.1.1 v0.1.0
   ```

2. **Make fix and commit:**
   ```bash
   git commit -m "fix(critical): resolve security vulnerability"
   ```

3. **Create hotfix release:**
   ```bash
   make release-tag VERSION=v0.1.1
   ```

4. **Push hotfix:**
   ```bash
   git push origin hotfix-v0.1.1
   git push origin v0.1.1
   ```

5. **Merge back to main:**
   ```bash
   git checkout main
   git merge hotfix-v0.1.1
   git push origin main
   git branch -d hotfix-v0.1.1
   ```

## Release Checklist

### Pre-Release
- [ ] All feature branches merged to main
- [ ] Working tree clean (`git status`)
- [ ] Local main up to date with origin (`git pull`)
- [ ] All tests pass locally (`make test`)
- [ ] All linters pass locally (`make lint`)
- [ ] Dependencies verified (`make deps-verify`)
- [ ] Breaking changes documented in commit messages

### Release Execution
- [ ] Run version bump command (`make version-patch/minor/major`)
- [ ] Verify changelog accuracy
- [ ] Push main branch (`git push origin main`)
- [ ] Push release tag (`git push origin vX.Y.Z`)

### Post-Release Verification
- [ ] GitHub Actions workflow completed successfully
- [ ] GitHub Release created with correct notes
- [ ] Binaries available for all platforms
- [ ] Homebrew formula updated (if applicable)
- [ ] Test installation from release artifacts
- [ ] Documentation reflects new version

## Troubleshooting

### Changelog Missing Commits

**Problem:** Expected commits not appearing in changelog.

**Solution:** Check commit message format:
```bash
git log --oneline --format="%s" v0.1.0..HEAD
```

Ensure commits match pattern: `type(scope): description`

Commits must use valid types: `feat`, `fix`, `refactor`, `perf`, `build`, `revert`.

### Changelog Shows "Unreleased"

**Problem:** CHANGELOG.md shows all recent changes under "Unreleased".

**Solution:** Regenerate changelog after creating tags:
```bash
make changelog
git add CHANGELOG.md .chglog/
git commit -m "docs(changelog): regenerate with current tags"
```

### Tag Already Exists

**Problem:** Cannot create tag because version already exists.

**Solution:** Delete local and remote tag:
```bash
git tag -d v0.1.1
git push origin :refs/tags/v0.1.1
```

Then recreate release with corrected version.

### Release Failed Quality Checks

**Problem:** `make version-patch` fails during quality checks.

**Solution:**
1. Fix failing tests or linter errors
2. Commit fixes
3. Re-run version command

### Release Failed in CI

**Problem:** GitHub Actions workflow failed after pushing tag.

**Solution:**
1. Navigate to repository Actions tab on GitHub
2. Select failed workflow run
3. Review error logs
4. Fix issue locally
5. Delete failed tag (local and remote)
6. Recreate release

### git-chglog Not Found

**Problem:** `make changelog` fails with command not found.

**Solution:** Install git-chglog:
```bash
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
```

Or let Makefile auto-install on first run.

### Commits in Wrong Order

**Problem:** Changelog entries appear in unexpected order.

**Solution:** This is normal. git-chglog orders by commit type and scope, not chronological order. Each type section shows commits grouped by scope.

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [git-chglog Documentation](https://github.com/git-chglog/git-chglog)
- [GoReleaser Documentation](https://goreleaser.com/)

## Navigation

**[↑ Back to Main README](../../README.md)** | [Documentation Index](../README.md)

