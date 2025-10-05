package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Remanage reinstalls packages by unmanaging then managing.
// Currently implements as sequential unmanage + manage.
// TODO: Use incremental planning with hash-based change detection.
func (c *client) Remanage(ctx context.Context, packages ...string) error {
	// Unmanage first (will succeed even if nothing installed)
	err := c.Unmanage(ctx, packages...)
	if err != nil {
		c.config.Logger.Warn(ctx, "unmanage_failed_during_remanage", "error", err)
		// Continue to manage anyway
	}

	// Then manage
	return c.Manage(ctx, packages...)
}

// PlanRemanage computes the execution plan for remanaging packages.
// TODO: Implement incremental planning using manifest hashes.
func (c *client) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	// For now, combine unmanage + manage plans
	// In the future, this should be incremental
	unmanagePlan, err := c.PlanUnmanage(ctx, packages...)
	if err != nil {
		// If no manifest, just plan manage
		return c.PlanManage(ctx, packages...)
	}

	managePlan, err := c.PlanManage(ctx, packages...)
	if err != nil {
		return dot.Plan{}, err
	}

	// Combine operations
	combined := make([]dot.Operation, 0, len(unmanagePlan.Operations)+len(managePlan.Operations))
	combined = append(combined, unmanagePlan.Operations...)
	combined = append(combined, managePlan.Operations...)

	return dot.Plan{
		Operations: combined,
		Metadata: dot.PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(combined),
		},
	}, nil
}
