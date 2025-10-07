package pipeline

import (
	"context"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// ManagePipelineOpts contains options for the Manage pipeline
type ManagePipelineOpts struct {
	FS        dot.FS
	IgnoreSet *ignore.IgnoreSet
	Policies  planner.ResolutionPolicies
}

// ManageInput contains the input for manage operations
type ManageInput struct {
	PackageDir dot.PackagePath
	TargetDir  dot.TargetPath
	Packages   []string
}

// ManagePipeline implements the complete manage workflow.
// It composes scanning, planning, resolution, and sorting stages.
type ManagePipeline struct {
	opts ManagePipelineOpts
}

// NewManagePipeline creates a new Manage pipeline with the given options.
func NewManagePipeline(opts ManagePipelineOpts) *ManagePipeline {
	return &ManagePipeline{
		opts: opts,
	}
}

// Execute runs the complete manage pipeline.
// It performs: scan packages -> compute desired state -> resolve conflicts -> sort operations
func (p *ManagePipeline) Execute(ctx context.Context, input ManageInput) dot.Result[dot.Plan] {
	// Stage 1: Scan packages
	scanInput := ScanInput{
		PackageDir: input.PackageDir,
		TargetDir:  input.TargetDir,
		Packages:   input.Packages,
		IgnoreSet:  p.opts.IgnoreSet,
		FS:         p.opts.FS,
	}

	scanResult := ScanStage()(ctx, scanInput)
	if scanResult.IsErr() {
		return dot.Err[dot.Plan](scanResult.UnwrapErr())
	}
	packages := scanResult.Unwrap()

	// Stage 2: Compute desired state
	planInput := PlanInput{
		Packages:  packages,
		TargetDir: input.TargetDir,
	}

	planResult := PlanStage()(ctx, planInput)
	if planResult.IsErr() {
		return dot.Err[dot.Plan](planResult.UnwrapErr())
	}
	desired := planResult.Unwrap()

	// Stage 3: Resolve conflicts and generate operations
	resolveInput := ResolveInput{
		Desired:   desired,
		FS:        p.opts.FS,
		Policies:  p.opts.Policies,
		BackupDir: "", // TODO: Add backup dir to options
	}

	resolveResult := ResolveStage()(ctx, resolveInput)
	if resolveResult.IsErr() {
		return dot.Err[dot.Plan](resolveResult.UnwrapErr())
	}
	resolved := resolveResult.Unwrap()

	// Check for unresolved conflicts
	if resolved.HasConflicts() {
		// Return plan with conflicts for user to handle
		// The caller can inspect the conflicts in the metadata
		return dot.Ok(dot.Plan{
			Operations: resolved.Operations,
			Metadata: dot.PlanMetadata{
				PackageCount:   len(packages),
				OperationCount: len(resolved.Operations),
				LinkCount:      countOperationsByKind(resolved.Operations, dot.OpKindLinkCreate),
				DirCount:       countOperationsByKind(resolved.Operations, dot.OpKindDirCreate),
				Conflicts:      convertConflicts(resolved.Conflicts),
				Warnings:       convertWarnings(resolved.Warnings),
			},
		})
	}

	// Stage 4: Sort operations topologically
	sortInput := SortInput{
		Operations: resolved.Operations,
	}

	sortResult := SortStage()(ctx, sortInput)
	if sortResult.IsErr() {
		return dot.Err[dot.Plan](sortResult.UnwrapErr())
	}
	sorted := sortResult.Unwrap()

	// Build final plan with metadata including any warnings
	plan := dot.Plan{
		Operations: sorted,
		Metadata: dot.PlanMetadata{
			PackageCount:   len(packages),
			OperationCount: len(sorted),
			LinkCount:      countOperationsByKind(sorted, dot.OpKindLinkCreate),
			DirCount:       countOperationsByKind(sorted, dot.OpKindDirCreate),
			Conflicts:      nil, // No conflicts in success path
			Warnings:       convertWarnings(resolved.Warnings),
		},
	}

	return dot.Ok(plan)
}

// countOperationsByKind counts operations of a specific kind
func countOperationsByKind(ops []dot.Operation, kind dot.OperationKind) int {
	count := 0
	for _, op := range ops {
		if op.Kind() == kind {
			count++
		}
	}
	return count
}
