# Git Hooks

This directory contains Git hook templates for the dot project.

## Pre-Commit Hook

The pre-commit hook enforces project quality standards before allowing commits:

### Checks Performed

1. **Test Coverage** (pkg/dot): Must be ≥ 80%
2. **Linting**: All linters must pass (golangci-lint)
3. **Tests**: Full test suite must pass

### Installation

```bash
# Copy the hook to .git/hooks
cp .githooks/pre-commit .git/hooks/pre-commit

# Make it executable
chmod +x .git/hooks/pre-commit
```

Or use git config to set the hooks path:

```bash
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

### Usage

Once installed, the hook runs automatically on `git commit`. If any check fails, the commit is rejected with a clear error message.

**Example output (passing):**
```
Running pre-commit checks...
→ Checking test coverage...
→ pkg/dot coverage: 80.4%
✅ Coverage check passed: 80.4% >= 80.0%
→ Running linters...
✅ Linting passed
→ Running full test suite...
✅ All tests passed

✅ All pre-commit checks passed
```

**Example output (failing):**
```
Running pre-commit checks...
→ Checking test coverage...
→ pkg/dot coverage: 75.3%

❌ COMMIT REJECTED: Test coverage below threshold
   Current:   75.3%
   Required:  80.0%
   Shortfall: 4.7%

Please add tests to pkg/dot to reach 80% coverage before committing.
```

### Bypassing the Hook

In exceptional circumstances, you can bypass the hook with:

```bash
git commit --no-verify
```

**Warning:** This is prohibited by project rules except in emergencies. See `.cursorrules` for details.

### Configuration

The coverage threshold is set in the hook script:

```bash
THRESHOLD=80.0
```

This matches the project constitution's 80% minimum coverage requirement.
