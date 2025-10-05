package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Adopt moves existing files from target into package then creates symlinks.
func (c *client) Adopt(ctx context.Context, files []string, pkg string) error {
	// TODO: Implement in next commit
	return nil
}

// PlanAdopt computes the execution plan for adopting files.
func (c *client) PlanAdopt(ctx context.Context, files []string, pkg string) (dot.Plan, error) {
	// TODO: Implement in next commit
	return dot.Plan{}, nil
}
