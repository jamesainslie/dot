package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Remanage reinstalls packages by unmanaging then managing.
func (c *client) Remanage(ctx context.Context, packages ...string) error {
	return dot.ErrNotImplemented{Feature: "Remanage"}
}

// PlanRemanage computes the execution plan for remanaging packages.
func (c *client) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	return dot.Plan{}, dot.ErrNotImplemented{Feature: "PlanRemanage"}
}
