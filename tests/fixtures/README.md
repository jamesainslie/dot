# Test Fixtures

Test data and sample configurations for integration tests.

## Structure

### bootstrap-configs/
Bootstrap configuration YAML samples:
- **minimal.yaml**: Smallest valid config
- **with-profiles.yaml**: Profile selection
- **platform-specific.yaml**: Platform-conditional packages
- **invalid-syntax.yaml**: Malformed YAML, for parser error tests
- **invalid-missing-version.yaml**: Schema violation, for validation tests

## Fixtures Elsewhere

- `cmd/dot/testdata/golden/{adopt,errors,manage}/`: Golden outputs for CLI
  commands. `internal/cli/golden` resolves golden paths relative to the test
  package, so these live beside the tests that use them.
- `internal/adapters/testdata/test-repo/`: Git adapter fixtures.

## Creating Fixtures

Use the FixtureBuilder API in test code rather than pre-creating static fixtures where possible. This ensures tests are self-contained and maintainable.

Static fixtures should be used for:
- Golden test comparison files
- Complex scenarios that are reused across multiple tests
- Migration testing from other tools

## Naming Conventions

- Use kebab-case for directory names
- Prefix dotfile sources with `dot-` (e.g., `dot-vimrc`)
- Use descriptive names indicating purpose

## Navigation

**[↑ Back to Main README](../../README.md)** | [Integration Tests](../integration/README.md)

