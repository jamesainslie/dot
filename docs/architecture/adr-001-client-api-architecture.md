# ADR-001: Client API Architecture Pattern

## Status
**Accepted** - Phase 12 implementation approach

## Context

Phase 12 requires implementing a public Client API in `pkg/dot/` that wraps internal pipeline and executor components. However, the current architecture creates an import cycle:

- Domain types (Operation, Plan, Result, etc.) live in `pkg/dot/`
- Internal packages (`internal/executor`, `internal/pipeline`, etc.) import these domain types
- A Client in `pkg/dot/` would need to import internal packages
- This creates a circular dependency: `pkg/dot` → `internal/*` → `pkg/dot`

## Decision

Implement Phase 12 using **Option 4: Interface-based Client with implementation in internal/api**.

### Implementation Pattern

```go
// pkg/dot/client.go - Public interface
package dot

type Client interface {
    Manage(ctx context.Context, packages ...string) error
    // ... other methods
}

func NewClient(cfg Config) (Client, error) {
    return newClientImpl(cfg)
}

// Registration for implementation
var newClientImpl func(Config) (Client, error)

func RegisterClientImpl(fn func(Config) (Client, error)) {
    newClientImpl = fn
}
```

```go
// internal/api/client.go - Private implementation
package api

import (
    "github.com/jamesainslie/dot/internal/executor"
    "github.com/jamesainslie/dot/internal/pipeline"
    "github.com/jamesainslie/dot/pkg/dot"
)

func init() {
    dot.RegisterClientImpl(newClient)
}

type client struct {
    config   dot.Config
    pipeline *pipeline.StowPipeline
    executor *executor.Executor
}

func newClient(cfg dot.Config) (dot.Client, error) {
    // Implementation
}

func (c *client) Manage(ctx context.Context, packages ...string) error {
    // Implementation
}
```

### Import Flow
```
pkg/dot/client.go
    ↓ defines interface (no import of internal/*)
    
internal/api/client.go
    ↓ imports pkg/dot (for interface)
    ↓ imports internal/* (for implementation)
    
internal/executor, internal/pipeline, etc.
    ↓ import pkg/dot (for domain types)
    ✅ NO CYCLE
```

## Alternatives Considered

### Option 1: Move Domain Types to internal/domain (Future: Phase 12b)
**Pros**: Clean architecture, no workarounds, direct Client struct
**Cons**: Requires refactoring all of Phases 1-10 (12-16 hours), high risk
**Decision**: Defer to Phase 12b after Phase 12 is stable

### Option 2: Separate pkg/client Module
**Pros**: No refactoring needed, clean separation
**Cons**: Two packages to import (`pkg/dot` and `pkg/client`), less intuitive
**Decision**: Rejected - less idiomatic than interface pattern

### Option 3: Functional API (No Client Struct)
**Pros**: Simple, no state
**Cons**: Can't maintain state, inefficient (recreates components), not idiomatic
**Decision**: Rejected - doesn't meet Phase 12 requirements

### Option 5: Accept Current Architecture (No Public Client)
**Pros**: Zero work
**Cons**: Poor library ergonomics, exposes internal packages, no semver guarantees
**Decision**: Rejected - violates "library first" constitutional principle

## Consequences

### Positive

1. **Achieves Phase 12 Goals**: Working public Client API
2. **Zero Risk to Existing Code**: No changes to completed Phases 1-11
3. **Idiomatic Go**: Interface-based clients are standard pattern
4. **Mockable**: Interface enables easy testing
5. **Fast Implementation**: 6-8 hours vs 12-16 for Option 1
6. **Future-Proof**: Can refactor to Option 1 later (Phase 12b)

### Negative

1. **Indirection**: One extra level of interface dispatch
2. **Registration Complexity**: init() mechanism less transparent
3. **Extra Package**: `internal/api/` exists only for import cycle workaround
4. **Not Ideal**: Architecturally Option 1 is cleaner

### Trade-offs

**Short-term**:
- ✅ Get working Client API quickly
- ✅ No risk to existing code
- ❌ Slightly more complex than direct struct

**Long-term**:
- ✅ Can refactor to Option 1 via Phase 12b
- ✅ Public API surface unchanged by refactoring
- ❌ Carries technical debt until Phase 12b

## Implementation

### Phase 12: Interface Pattern (Now)
- Effort: 6-8 hours
- Risk: Low
- Status: **In Progress**

### Phase 12b: Refactor to Struct (Later)
- Effort: 12-16 hours
- Risk: Medium-High
- Status: **Optional Future Work**
- Prerequisites: Phase 12 stable, team capacity available

## Validation

Phase 12 successful when:
- [ ] Client interface accessible via `import "pkg/dot"`
- [ ] All operations work (Manage, Unmanage, Remanage, Adopt, Status, List)
- [ ] Test coverage ≥ 80%
- [ ] No import cycles
- [ ] Documentation complete
- [ ] Examples demonstrate usage

## References

- [Phase 12 Implementation Plan](./Phase-12-Plan.md) - Detailed implementation steps
- [Phase 12b Refactoring Plan](./Phase-12b-Refactor-Plan.md) - Future architecture improvement
- [Architecture Documentation](./Architecture.md) - Overall system architecture

## Review

- **Proposed**: 2024-10-05
- **Discussed**: Examined 5 options, selected Option 4 as pragmatic choice
- **Accepted**: 2024-10-05
- **Supersedes**: Original Phase 12 plan (struct-based Client)
- **Superseded by**: Potentially Phase 12b (if executed)

