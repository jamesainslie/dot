package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Remanage reinstalls packages by unmanaging then managing.
func (c *client) Remanage(ctx context.Context, packages ...string) error {
	// TODO: Implement in next commit
	return nil
}

// PlanRemanage computes the execution plan for remanaging packages.
func (c *client) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	// TODO: Implement in next commit
	return dot.Plan{}, nil
}
