# Test Fixtures

Test data and sample configurations for integration tests.

## Structure

### scenarios/
Pre-built test scenarios representing common use cases:
- **simple/**: Basic single package setup
- **complex/**: Multiple packages with dependencies
- **conflicts/**: Scenarios with pre-existing conflicts
- **migration/**: GNU Stow migration scenarios

### packages/
Sample packages for testing:
- **dotfiles/**: Basic dotfiles package
- **nvim/**: Neovim configuration package
- **shell/**: Shell configuration (zsh, bash)

### golden/
Expected outputs for golden tests:
- **status/**: Expected status command outputs
- **doctor/**: Expected doctor command outputs
- **list/**: Expected list command outputs

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

