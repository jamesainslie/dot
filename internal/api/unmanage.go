package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Unmanage removes the specified packages by deleting symlinks.
func (c *client) Unmanage(ctx context.Context, packages ...string) error {
	// TODO: Implement in next commit
	return nil
}

// PlanUnmanage computes the execution plan for unmanaging packages.
func (c *client) PlanUnmanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	// TODO: Implement in next commit
	return dot.Plan{}, nil
}
