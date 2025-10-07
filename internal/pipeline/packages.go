package pipeline

import (
	"context"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// ManagePipelineOpts contains options for the Manage pipeline
type ManagePipelineOpts struct {
	FS        dot.FS
	IgnoreSet *ignore.IgnoreSet
	Policies  planner.ResolutionPolicies
	BackupDir string
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
		BackupDir: p.opts.BackupDir,
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

	// Build package-operation mapping by matching operations to package source paths
	packageOps := buildPackageOperationMapping(packages, sorted)

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
		PackageOperations: packageOps,
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

// buildPackageOperationMapping creates a mapping from package names to operation IDs
// by matching operation source paths to package paths.
func buildPackageOperationMapping(packages []dot.Package, operations []dot.Operation) map[string][]dot.OperationID {
	packageOps := make(map[string][]dot.OperationID)

	// For each package, find operations that reference files from that package
	for _, pkg := range packages {
		pkgPath := pkg.Path.String()
		ops := make([]dot.OperationID, 0)

		for _, op := range operations {
			// Check if this operation's source is from this package
			if operationBelongsToPackage(op, pkgPath) {
				ops = append(ops, op.ID())
			}
		}

		if len(ops) > 0 {
			packageOps[pkg.Name] = ops
		}
	}

	return packageOps
}

// operationBelongsToPackage checks if an operation's source is from the given package path.
func operationBelongsToPackage(op dot.Operation, pkgPath string) bool {
	switch o := op.(type) {
	case dot.LinkCreate:
		// LinkCreate source is the file in the package
		return isUnderPath(o.Source.String(), pkgPath)
	case dot.FileMove:
		// FileMove destination is the file in the package
		return isUnderPath(o.Dest.String(), pkgPath)
	default:
		// Other operations (DirCreate, LinkDelete, etc.) don't belong to a specific package
		return false
	}
}

// isUnderPath checks if path is under basePath.
func isUnderPath(path, basePath string) bool {
	// Clean both paths for consistent comparison
	cleanPath := filepath.Clean(path)
	cleanBase := filepath.Clean(basePath)

	// Check if path starts with basePath
	rel, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil {
		return false
	}

	// If relative path doesn't go up (..), it's under basePath
	return rel != "." && !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.'
}
