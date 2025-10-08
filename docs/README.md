# Documentation Index

This directory contains all technical documentation for the dot CLI project, organized into logical categories for easy navigation.

## Directory Structure

```
docs/
├── developer/            # Developer documentation and workflows
├── planning/            # Active development plans and completed phases
└── user/               # End-user documentation and guides
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

### Additional User Resources
- [`user/installation-homebrew.md`](user/installation-homebrew.md) - Homebrew installation guide
- [`user/migration-from-stow.md`](user/migration-from-stow.md) - Migration guide from GNU Stow

## Developer Documentation

Documentation for developers contributing to the dot project.

- [`developer/architecture.md`](developer/architecture.md) - System architecture and design
- [`developer/release-workflow.md`](developer/release-workflow.md) - Release process and workflow

## Planning Documentation

Development phases, implementation plans, and project roadmaps.

### Active Phase Plans
Current and recent development phases located in repository root:

- [`../phase-16-plan.md`](../phase-16-plan.md) - Phase 16 plan
- [`../phase-17-plan.md`](../phase-17-plan.md) - Phase 17 plan
- [`../phase-18-plan.md`](../phase-18-plan.md) - Phase 18 plan
- [`../phase-20-plan.md`](../phase-20-plan.md) - Phase 20 plan

### Completed Phases
Phase completion status and milestones achieved:

- [`planning/PHASE-19-COMPLETE.md`](planning/PHASE-19-COMPLETE.md) - Phase 19 completion
- [`planning/phase-24-complete.md`](planning/phase-24-complete.md) - Phase 24 completion
- [`planning/phase-24-progress-checkpoint.md`](planning/phase-24-progress-checkpoint.md) - Phase 24 progress checkpoint
- [`planning/phase-24-code-smell-remediation-plan.md`](planning/phase-24-code-smell-remediation-plan.md) - Phase 24 code smell remediation plan

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
3. Review [`../CONTRIBUTING.md`](../CONTRIBUTING.md) for contribution guidelines
4. Check [`user/09-glossary.md`](user/09-glossary.md) for project terminology

### For Feature Development
1. Review active phase plans in repository root (phase-16 through phase-20)
2. Check [`planning/`](planning/) for completed phases and status
3. Consult [`developer/release-workflow.md`](developer/release-workflow.md) for release process

### For Users Migrating from GNU Stow
1. Start with [`user/migration-from-stow.md`](user/migration-from-stow.md)
2. Review [`user/01-introduction.md`](user/01-introduction.md) for terminology differences
3. Follow [`user/06-workflows.md`](user/06-workflows.md) for common use cases

## Document Maintenance

### Adding New Documentation
- User guides go in `user/`
- Developer documentation goes in `developer/`
- Planning documents go in `planning/` for completion records, or repository root for active plans
- Update this index when adding new documents

### Phase Completion
When completing a development phase:
1. Create completion document in `planning/`
2. Update this index to list the completion document
3. Consider moving active plan from repository root to archive if no longer relevant

### Organization Guidelines
- Keep this index synchronized with actual file structure
- User documentation should be comprehensive and accessible to non-technical users
- Developer documentation should focus on contribution and development workflows
- Planning documents track project progress and decisions

