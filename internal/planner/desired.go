// Package planner provides pure planning logic for computing operations.
package planner

import (
	"github.com/jamesainslie/dot/pkg/dot"
)

// LinkSpec specifies a desired symbolic link.
type LinkSpec struct {
	Source dot.FilePath // Source file in package
	Target dot.FilePath // Target location
}

// DirSpec specifies a desired directory.
type DirSpec struct {
	Path dot.FilePath
}

// DesiredState represents the desired filesystem state.
type DesiredState struct {
	Links map[string]LinkSpec // Key: target path
	Dirs  map[string]DirSpec  // Key: directory path
}

// ComputeDesiredState computes desired state from packages.
// This is a pure function that determines what links and directories
// should exist based on the package contents.
func ComputeDesiredState(packages []dot.Package, target dot.TargetPath) dot.Result[DesiredState] {
	state := DesiredState{
		Links: make(map[string]LinkSpec),
		Dirs:  make(map[string]DirSpec),
	}

	// For now, return empty state
	// Full implementation will process packages and build link/dir specs

	return dot.Ok(state)
}
