# Documentation Index

This directory contains all technical documentation for the dot CLI project, organized into logical categories for easy navigation.

## Directory Structure

```
docs/
├── developer/            # Developer documentation and workflows
├── plans/                # Historical design and implementation plans
└── user/                 # End-user documentation and guides
```

## User Documentation

End-user guides, tutorials, and reference materials.

### User Guide Index
- [`user/index.md`](user/index.md) - User guide table of contents and navigation

### Core Guides
- [`user/01-introduction.md`](user/01-introduction.md) - Introduction and core concepts
- [`user/02-installation.md`](user/02-installation.md) - Installation guide
- [`user/03-quickstart.md`](user/03-quickstart.md) - Quick start tutorial
- [`user/04-configuration.md`](user/04-configuration.md) - Configuration reference
- [`user/05-commands.md`](user/05-commands.md) - Command reference
- [`user/06-workflows.md`](user/06-workflows.md) - Common workflows
- [`user/07-advanced.md`](user/07-advanced.md) - Advanced features
- [`user/08-troubleshooting.md`](user/08-troubleshooting.md) - Troubleshooting guide
- [`user/09-glossary.md`](user/09-glossary.md) - Glossary of terms
- [`user/10-updates.md`](user/10-updates.md) - Update and upgrade behavior

### Additional User Resources
- [`user/installation-homebrew.md`](user/installation-homebrew.md) - Homebrew installation guide
- [`user/migration-from-stow.md`](user/migration-from-stow.md) - Migration guide from GNU Stow
- [`user/migration-v0.3.md`](user/migration-v0.3.md) - Migration guide for the v0.3 format change
- [`user/bootstrap-config-spec.md`](user/bootstrap-config-spec.md) - Bootstrap configuration file specification
- [`user/repository-config.md`](user/repository-config.md) - Repository-level configuration reference
- [`user/ignore-system.md`](user/ignore-system.md) - Ignore pattern system reference
- [`user/secrets-management.md`](user/secrets-management.md) - Secret detection and handling

## Developer Documentation

Documentation for developers contributing to the dot project.

- [`developer/architecture.md`](developer/architecture.md) - System architecture and design
- [`developer/cli-architecture.md`](developer/cli-architecture.md) - CLI layering contract and dependency rules
- [`developer/testing.md`](developer/testing.md) - Testing strategy and paradigms
- [`developer/error-handling.md`](developer/error-handling.md) - Error and Result[T] conventions
- [`developer/release-workflow.md`](developer/release-workflow.md) - Release process and workflow
- [`developer/doctor-system-cujs.md`](developer/doctor-system-cujs.md) - Doctor customer usage journeys (analysis, partly superseded)
- [`developer/doctor-system-ux-design.md`](developer/doctor-system-ux-design.md) - Doctor UX design proposal (largely unimplemented)
- [`developer/triage-improvements.md`](developer/triage-improvements.md) - Triage improvement plan (historical record)

## Plans

Dated design and implementation plans, retained for historical context. They
describe intent at the time of writing, not current behavior.

- [`plans/2026-02-27-stress-test-skill-design.md`](plans/2026-02-27-stress-test-skill-design.md) - Stress-test skill design

## Additional Documentation

Other project documentation is located in the repository root:

- [`../README.md`](../README.md) - Project README and getting started guide
- [`../CHANGELOG.md`](../CHANGELOG.md) - Version history and release notes
- [`../LICENSE`](../LICENSE) - License information
- [`../CONTRIBUTING.md`](../CONTRIBUTING.md) - Contributing guidelines

## Navigation Tips

### For New Users
1. Start with [`user/index.md`](user/index.md) for the complete user guide navigation
2. Read [`user/01-introduction.md`](user/01-introduction.md) for core concepts
3. Follow [`user/03-quickstart.md`](user/03-quickstart.md) to get started quickly

### For New Contributors
1. Read [`../README.md`](../README.md) for project overview
2. Review [`developer/architecture.md`](developer/architecture.md) for system architecture
3. Review [`developer/testing.md`](developer/testing.md) for testing strategy and TDD practices
4. Review [`../CONTRIBUTING.md`](../CONTRIBUTING.md) for contribution guidelines
5. Check [`user/09-glossary.md`](user/09-glossary.md) for project terminology

### For Users Migrating from GNU Stow
1. Start with [`user/migration-from-stow.md`](user/migration-from-stow.md)
2. Review [`user/01-introduction.md`](user/01-introduction.md) for terminology differences
3. Follow [`user/06-workflows.md`](user/06-workflows.md) for common use cases

## Document Maintenance

### Adding New Documentation
- User guides go in `user/`
- Developer documentation goes in `developer/`
- Update this index when adding new documents

### Organization Guidelines
- Keep this index synchronized with actual file structure
- User documentation should be comprehensive and accessible to non-technical users
- Developer documentation should focus on contribution and development workflows

