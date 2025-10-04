package pipeline

import (
	"context"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/pkg/dot"
)

// ScanInput contains the input for scanning packages
type ScanInput struct {
	StowDir   dot.PackagePath
	TargetDir dot.TargetPath
	Packages  []string
	IgnoreSet *ignore.IgnoreSet
	FS        dot.FS
}

// ScanStage creates a pipeline stage that scans packages.
// Returns a slice of scanned packages with their file trees.
func ScanStage() Pipeline[ScanInput, []dot.Package] {
	return func(ctx context.Context, input ScanInput) dot.Result[[]dot.Package] {
		packages := make([]dot.Package, 0, len(input.Packages))

		for _, pkgName := range input.Packages {
			// Create package path by joining stow dir with package name
			pkgPathStr := filepath.Join(input.StowDir.String(), pkgName)
			pkgPathResult := dot.NewPackagePath(pkgPathStr)
			if pkgPathResult.IsErr() {
				return dot.Err[[]dot.Package](pkgPathResult.UnwrapErr())
			}
			pkgPath := pkgPathResult.Unwrap()

			pkgResult := scanner.ScanPackage(ctx, input.FS, pkgPath, pkgName, input.IgnoreSet)

			if pkgResult.IsErr() {
				return dot.Err[[]dot.Package](pkgResult.UnwrapErr())
			}

			packages = append(packages, pkgResult.Unwrap())
		}

		return dot.Ok(packages)
	}
}

// PlanInput contains the input for planning operations
type PlanInput struct {
	Packages  []dot.Package
	TargetDir dot.TargetPath
}

// PlanStage creates a pipeline stage that computes desired state.
// Takes scanned packages and computes what links should exist.
func PlanStage() Pipeline[PlanInput, planner.DesiredState] {
	return func(ctx context.Context, input PlanInput) dot.Result[planner.DesiredState] {
		return planner.ComputeDesiredState(input.Packages, input.TargetDir)
	}
}

// ResolveInput contains the input for conflict resolution
type ResolveInput struct {
	Desired   planner.DesiredState
	FS        dot.FS
	Policies  planner.ResolutionPolicies
	BackupDir string
}

// ResolveStage creates a pipeline stage that resolves conflicts.
// Takes desired state and current filesystem state to generate operations.
func ResolveStage() Pipeline[ResolveInput, planner.ResolveResult] {
	return func(ctx context.Context, input ResolveInput) dot.Result[planner.ResolveResult] {
		// Convert desired state to operations
		operations := planner.ComputeOperationsFromDesiredState(input.Desired)

		// For now, use empty current state (will scan target in future phases)
		// This will be enhanced when we add current state scanning
		current := planner.CurrentState{
			Files: make(map[string]planner.FileInfo),
			Links: make(map[string]planner.LinkTarget),
			Dirs:  make(map[string]bool),
		}

		// Resolve conflicts
		result := planner.Resolve(operations, current, input.Policies, input.BackupDir)
		return dot.Ok(result)
	}
}

// SortInput contains the input for topological sorting
type SortInput struct {
	Operations []dot.Operation
}

// SortStage creates a pipeline stage that sorts operations.
// Takes operations and returns them in dependency order.
func SortStage() Pipeline[SortInput, []dot.Operation] {
	return func(ctx context.Context, input SortInput) dot.Result[[]dot.Operation] {
		if len(input.Operations) == 0 {
			return dot.Ok([]dot.Operation{})
		}

		graph := planner.BuildGraph(input.Operations)
		sorted, err := graph.TopologicalSort()
		if err != nil {
			return dot.Err[[]dot.Operation](err)
		}
		return dot.Ok(sorted)
	}
}
