package pipeline

import (
	"context"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/jamesainslie/dot/internal/scanner"
)

// ScanInput contains the input for scanning packages
type ScanInput struct {
	PackageDir domain.PackagePath
	TargetDir  domain.TargetPath
	Packages   []string
	IgnoreSet  *ignore.IgnoreSet
	FS         domain.FS
}

// ScanStage creates a pipeline stage that scans packages.
// Returns a slice of scanned packages with their file trees.
func ScanStage() Pipeline[ScanInput, []domain.Package] {
	return func(ctx context.Context, input ScanInput) domain.Result[[]domain.Package] {
		// Early cancellation check
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Package](ctx.Err())
		default:
		}

		packages := make([]domain.Package, 0, len(input.Packages))

		for _, pkgName := range input.Packages {
			// Check for cancellation before processing each package
			select {
			case <-ctx.Done():
				return domain.Err[[]domain.Package](ctx.Err())
			default:
			}

			// Create package path by joining package dir with package name
			pkgPathStr := filepath.Join(input.PackageDir.String(), pkgName)
			pkgPathResult := domain.NewPackagePath(pkgPathStr)
			if pkgPathResult.IsErr() {
				return domain.Err[[]domain.Package](pkgPathResult.UnwrapErr())
			}
			pkgPath := pkgPathResult.Unwrap()

			// scanner.ScanPackage already accepts context and should handle cancellation
			pkgResult := scanner.ScanPackage(ctx, input.FS, pkgPath, pkgName, input.IgnoreSet)

			if pkgResult.IsErr() {
				return domain.Err[[]domain.Package](pkgResult.UnwrapErr())
			}

			packages = append(packages, pkgResult.Unwrap())
		}

		return domain.Ok(packages)
	}
}

// PlanInput contains the input for planning operations
type PlanInput struct {
	Packages  []domain.Package
	TargetDir domain.TargetPath
}

// PlanStage creates a pipeline stage that computes desired state.
// Takes scanned packages and computes what links should exist.
func PlanStage() Pipeline[PlanInput, planner.DesiredState] {
	return func(ctx context.Context, input PlanInput) domain.Result[planner.DesiredState] {
		// Early cancellation check before potentially long-running planning
		select {
		case <-ctx.Done():
			return domain.Err[planner.DesiredState](ctx.Err())
		default:
		}

		return planner.ComputeDesiredState(input.Packages, input.TargetDir)
	}
}

// ResolveInput contains the input for conflict resolution
type ResolveInput struct {
	Desired   planner.DesiredState
	FS        domain.FS
	Policies  planner.ResolutionPolicies
	BackupDir string
}

// ResolveStage creates a pipeline stage that resolves conflicts.
// Takes desired state and current filesystem state to generate operations.
func ResolveStage() Pipeline[ResolveInput, planner.ResolveResult] {
	return func(ctx context.Context, input ResolveInput) domain.Result[planner.ResolveResult] {
		// Early cancellation check
		select {
		case <-ctx.Done():
			return domain.Err[planner.ResolveResult](ctx.Err())
		default:
		}

		// Convert desired state to operations
		operations := planner.ComputeOperationsFromDesiredState(input.Desired)

		// Check for cancellation before building current state
		select {
		case <-ctx.Done():
			return domain.Err[planner.ResolveResult](ctx.Err())
		default:
		}

		// For now, use empty current state (will scan target in future phases)
		// This will be enhanced when we add current state scanning
		current := planner.CurrentState{
			Files: make(map[string]planner.FileInfo),
			Links: make(map[string]planner.LinkTarget),
			Dirs:  make(map[string]bool),
		}

		// Check for cancellation before potentially long-running conflict resolution
		select {
		case <-ctx.Done():
			return domain.Err[planner.ResolveResult](ctx.Err())
		default:
		}

		// Resolve conflicts
		result := planner.Resolve(operations, current, input.Policies, input.BackupDir)
		return domain.Ok(result)
	}
}

// SortInput contains the input for topological sorting
type SortInput struct {
	Operations []domain.Operation
}

// SortStage creates a pipeline stage that sorts operations.
// Takes operations and returns them in dependency order.
func SortStage() Pipeline[SortInput, []domain.Operation] {
	return func(ctx context.Context, input SortInput) domain.Result[[]domain.Operation] {
		// Early cancellation check
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Operation](ctx.Err())
		default:
		}

		if len(input.Operations) == 0 {
			return domain.Ok([]domain.Operation{})
		}

		// Check for cancellation before building dependency graph
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Operation](ctx.Err())
		default:
		}

		graph := planner.BuildGraph(input.Operations)

		// Check for cancellation before potentially long-running topological sort
		select {
		case <-ctx.Done():
			return domain.Err[[]domain.Operation](ctx.Err())
		default:
		}

		sorted, err := graph.TopologicalSort()
		if err != nil {
			return domain.Err[[]domain.Operation](err)
		}
		return domain.Ok(sorted)
	}
}
