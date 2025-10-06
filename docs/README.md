# Documentation Index

This directory contains all technical documentation for the dot CLI project, organized into logical categories for easy navigation.

## Directory Structure

```
docs/
├── architecture/          # System architecture and design decisions
├── migration/            # Breaking changes and refactoring documentation
├── planning/             # Development plans and roadmaps
│   └── completed/        # Completed phase markers and status reports
├── reference/            # Reference documentation
└── reviews/              # Code review history and templates
```

## Architecture Documentation

Technical architecture and architectural decision records.

### Core Architecture
- [`architecture/architecture.md`](architecture/architecture.md) - Complete system architecture documentation
- [`architecture/adr-001-client-api-architecture.md`](architecture/adr-001-client-api-architecture.md) - Client API architectural decision record

## Planning Documentation

Development phases, implementation plans, and project roadmaps.

### Master Plan
- [`planning/implementation-plan.md`](planning/implementation-plan.md) - Master implementation plan and strategy

### Phase Plans
Development organized into discrete phases, each with specific goals and deliverables.

- [`planning/phase-7-plan.md`](planning/phase-7-plan.md) - Phase 7
- [`planning/phase-8-plan.md`](planning/phase-8-plan.md) - Phase 8
- [`planning/phase-9-plan.md`](planning/phase-9-plan.md) - Phase 9
- [`planning/phase-10-plan.md`](planning/phase-10-plan.md) - Phase 10
- [`planning/phase-11-plan.md`](planning/phase-11-plan.md) - Phase 11
- [`planning/phase-12-plan.md`](planning/phase-12-plan.md) - Phase 12
- [`planning/phase-12b-refactor-plan.md`](planning/phase-12b-refactor-plan.md) - Phase 12b Refactoring
- [`planning/phase-13-plan.md`](planning/phase-13-plan.md) - Phase 13
- [`planning/phase-14-plan.md`](planning/phase-14-plan.md) - Phase 14
- [`planning/phase-15-plan.md`](planning/phase-15-plan.md) - Phase 15
- [`planning/phase-15b-plan.md`](planning/phase-15b-plan.md) - Phase 15b
- [`planning/phase-15c-plan.md`](planning/phase-15c-plan.md) - Phase 15c
- [`planning/phase-16-plan.md`](planning/phase-16-plan.md) - Phase 16
- [`planning/phase-17-plan.md`](planning/phase-17-plan.md) - Phase 17
- [`planning/phase-18-plan.md`](planning/phase-18-plan.md) - Phase 18
- [`planning/phase-19-plan.md`](planning/phase-19-plan.md) - Phase 19
- [`planning/phase-20-plan.md`](planning/phase-20-plan.md) - Phase 20
- [`planning/phase-21-stow-terminology-refactor-plan.md`](planning/phase-21-stow-terminology-refactor-plan.md) - Phase 21: Stow Terminology Refactor

### Completed Phases
Phase completion status and milestones achieved.

- [`planning/completed/phase-0-complete.md`](planning/completed/phase-0-complete.md) - Phase 0 completion
- [`planning/completed/phase-1-complete.md`](planning/completed/phase-1-complete.md) - Phase 1 completion
- [`planning/completed/phase-2-complete.md`](planning/completed/phase-2-complete.md) - Phase 2 completion
- [`planning/completed/phase-3-complete.md`](planning/completed/phase-3-complete.md) - Phase 3 completion
- [`planning/completed/phase-4-complete.md`](planning/completed/phase-4-complete.md) - Phase 4 completion
- [`planning/completed/phase-5-complete.md`](planning/completed/phase-5-complete.md) - Phase 5 completion
- [`planning/completed/phase-6-complete.md`](planning/completed/phase-6-complete.md) - Phase 6 completion
- [`planning/completed/phase-7-complete.md`](planning/completed/phase-7-complete.md) - Phase 7 completion
- [`planning/completed/phase-8-complete.md`](planning/completed/phase-8-complete.md) - Phase 8 completion
- [`planning/completed/phase-9-complete.md`](planning/completed/phase-9-complete.md) - Phase 9 completion
- [`planning/completed/phase-10-complete.md`](planning/completed/phase-10-complete.md) - Phase 10 completion
- [`planning/completed/phase-12-complete.md`](planning/completed/phase-12-complete.md) - Phase 12 completion
- [`planning/completed/phase-14-complete.md`](planning/completed/phase-14-complete.md) - Phase 14 completion
- [`planning/completed/phase-15-complete.md`](planning/completed/phase-15-complete.md) - Phase 15 completion
- [`planning/completed/phase-15c-complete.md`](planning/completed/phase-15c-complete.md) - Phase 15c completion
- [`planning/completed/remediation-complete.md`](planning/completed/remediation-complete.md) - Code remediation completion
- [`planning/completed/final-coverage-status.md`](planning/completed/final-coverage-status.md) - Final test coverage status

## Reference Documentation

Reference materials for configuration, features, and terminology.

- [`reference/configuration.md`](reference/configuration.md) - Configuration system documentation
- [`reference/features.md`](reference/features.md) - Feature documentation and capabilities
- [`reference/terminology.md`](reference/terminology.md) - Project terminology and glossary

## Migration Documentation

Breaking changes, refactoring summaries, and migration guides.

- [`migration/breaking-changes-v0.2.0.md`](migration/breaking-changes-v0.2.0.md) - Breaking changes in v0.2.0
- [`migration/stow-refactor-summary.md`](migration/stow-refactor-summary.md) - Stow terminology refactor summary
- [`migration/stow-terminology-analysis.md`](migration/stow-terminology-analysis.md) - Analysis of stow terminology migration

## Code Reviews

Historical code reviews and review templates.

- [`../reviews/readme.md`](../reviews/readme.md) - Code review process documentation
- [`../reviews/template-code-review.md`](../reviews/template-code-review.md) - Code review template
- [`../reviews/code-review-prompt.md`](../reviews/code-review-prompt.md) - Code review prompt and guidelines
- [`../reviews/code-review-remediation-progress.md`](../reviews/code-review-remediation-progress.md) - Remediation progress tracking
- [`../reviews/code-review-remediation-summary.md`](../reviews/code-review-remediation-summary.md) - Remediation summary
- Historical reviews available in [`../reviews/`](../reviews/) directory

## Additional Documentation

Other project documentation is located in the repository root:

- [`../README.md`](../README.md) - Project README and getting started guide
- [`../CHANGELOG.md`](../CHANGELOG.md) - Version history and release notes
- [`../LICENSE`](../LICENSE) - License information
- [`../REVIEW_SYSTEM.md`](../REVIEW_SYSTEM.md) - Code review system documentation

## Navigation Tips

### For New Contributors
1. Start with [`../README.md`](../README.md) for project overview
2. Review [`architecture/architecture.md`](architecture/architecture.md) for system design
3. Check [`reference/terminology.md`](reference/terminology.md) for project vocabulary

### For Feature Development
1. Review [`planning/implementation-plan.md`](planning/implementation-plan.md) for overall strategy
2. Check relevant phase plans in [`planning/`](planning/)
3. Review [`reference/features.md`](reference/features.md) for existing features

### For Migration Work
1. Check [`migration/`](migration/) directory for breaking changes
2. Review refactoring documentation for context
3. Consult [`reference/terminology.md`](reference/terminology.md) for updated terms

### For Code Review
1. Use [`../reviews/template-code-review.md`](../reviews/template-code-review.md) as a template
2. Review past code reviews in [`../reviews/`](../reviews/) for examples
3. Follow [`../REVIEW_SYSTEM.md`](../REVIEW_SYSTEM.md) guidelines

## Document Maintenance

### Adding New Documentation
- Architecture documents go in `architecture/`
- Planning documents go in `planning/`
- Reference materials go in `reference/`
- Migration guides go in `migration/`
- Update this index when adding new documents

### Phase Completion
When completing a development phase:
1. Create completion document in `planning/completed/`
2. Update this index if necessary
3. Reference completion document in relevant phase plan

### Deprecation
When deprecating documentation:
1. Move to appropriate `deprecated/` subdirectory
2. Add deprecation notice to document header
3. Update this index to reflect status

