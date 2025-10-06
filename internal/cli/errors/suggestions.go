package errors

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/pkg/dot"
)

// SuggestionEngine generates actionable resolution steps.
type SuggestionEngine struct {
	context ErrorContext
}

// Generate creates suggestions for an error.
func (e *SuggestionEngine) Generate(err error) []string {
	if err == nil {
		return nil
	}

	// Domain Errors
	var invalidPath dot.ErrInvalidPath
	if errors.As(err, &invalidPath) {
		return e.suggestForInvalidPath(invalidPath)
	}

	var pkgNotFound dot.ErrPackageNotFound
	if errors.As(err, &pkgNotFound) {
		return e.suggestForPackageNotFound(pkgNotFound)
	}

	var conflict dot.ErrConflict
	if errors.As(err, &conflict) {
		return e.suggestForConflict(conflict)
	}

	var cyclicDep dot.ErrCyclicDependency
	if errors.As(err, &cyclicDep) {
		return e.suggestForCyclicDependency(cyclicDep)
	}

	// Infrastructure Errors
	var permDenied dot.ErrPermissionDenied
	if errors.As(err, &permDenied) {
		return e.suggestForPermissionDenied(permDenied)
	}

	var fsOp dot.ErrFilesystemOperation
	if errors.As(err, &fsOp) {
		return e.suggestForFilesystemOperation(fsOp)
	}

	// Executor Errors
	var execFailed dot.ErrExecutionFailed
	if errors.As(err, &execFailed) {
		return e.suggestForExecutionFailed(execFailed)
	}

	var srcNotFound dot.ErrSourceNotFound
	if errors.As(err, &srcNotFound) {
		return e.suggestForSourceNotFound(srcNotFound)
	}

	return nil
}

func (e *SuggestionEngine) suggestForInvalidPath(err dot.ErrInvalidPath) []string {
	suggestions := []string{
		"Use absolute paths starting with /",
		"Check for typos in the path",
	}

	if strings.Contains(err.Path, "~") {
		suggestions = append(suggestions, "Expand ~ to your home directory path")
	}

	if strings.Contains(err.Path, "..") {
		suggestions = append(suggestions, "Avoid relative path components like ..")
	}

	return suggestions
}

func (e *SuggestionEngine) suggestForPackageNotFound(err dot.ErrPackageNotFound) []string {
	suggestions := []string{
		"Check available packages with: dot list",
	}

	if e.context.Config.StowDir != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Verify stow directory: %s", e.context.Config.StowDir))
	}

	suggestions = append(suggestions,
		"Check for typos in package name",
		"Ensure the package directory exists in your stow directory")

	return suggestions
}

func (e *SuggestionEngine) suggestForConflict(err dot.ErrConflict) []string {
	suggestions := []string{}

	if strings.Contains(err.Reason, "exists") || strings.Contains(err.Reason, "file") {
		suggestions = append(suggestions,
			"Use 'dot adopt' to move the existing file into the package",
			"Remove the conflicting file manually if it's not needed",
			"Use --backup to preserve the existing file")
	}

	suggestions = append(suggestions,
		"Use --dry-run to preview operations before applying them")

	return suggestions
}

func (e *SuggestionEngine) suggestForCyclicDependency(err dot.ErrCyclicDependency) []string {
	return []string{
		"This indicates a bug in the operation planning logic",
		"Please report this issue with the command you ran",
		"Try running operations on individual packages separately",
	}
}

func (e *SuggestionEngine) suggestForPermissionDenied(err dot.ErrPermissionDenied) []string {
	suggestions := []string{}

	if err.Path != "" {
		dir := filepath.Dir(err.Path)
		suggestions = append(suggestions,
			fmt.Sprintf("Check file permissions: ls -la %s", dir),
			fmt.Sprintf("Verify you have write access to: %s", dir))
	}

	suggestions = append(suggestions,
		"Check directory ownership and permissions",
		"You may need appropriate permissions to modify this location")

	return suggestions
}

func (e *SuggestionEngine) suggestForFilesystemOperation(err dot.ErrFilesystemOperation) []string {
	suggestions := []string{
		"Verify the file system is writable",
		"Check available disk space",
	}

	if err.Path != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Verify path exists: %s", err.Path))
	}

	return suggestions
}

func (e *SuggestionEngine) suggestForExecutionFailed(err dot.ErrExecutionFailed) []string {
	suggestions := []string{
		"Review the individual error messages above for specific issues",
	}

	if err.RolledBack > 0 {
		suggestions = append(suggestions,
			"Some operations were rolled back automatically",
			"The system should be in a consistent state")
	} else {
		suggestions = append(suggestions,
			"Some operations may have been partially completed",
			"You may need to manually clean up")
	}

	if e.context.Config.DryRun {
		suggestions = append(suggestions,
			"Remove --dry-run to apply the operations")
	}

	return suggestions
}

func (e *SuggestionEngine) suggestForSourceNotFound(err dot.ErrSourceNotFound) []string {
	suggestions := []string{
		"Verify the source file exists in the package",
	}

	if e.context.Config.StowDir != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Check package directory: %s", e.context.Config.StowDir))
	}

	suggestions = append(suggestions,
		"Ensure the package structure is correct")

	return suggestions
}

// Prioritize orders suggestions by likely usefulness.
func (e *SuggestionEngine) Prioritize(suggestions []string) []string {
	// For now, return as-is. In the future, this could use heuristics
	// to reorder based on error context, user history, etc.
	return suggestions
}
