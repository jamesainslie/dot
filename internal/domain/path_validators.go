package domain

import (
	"path/filepath"
	"strings"
)

// PathValidator validates path strings according to specific rules.
// Multiple validators can be composed to enforce complex validation requirements.
type PathValidator interface {
	Validate(path string) error
}

// AbsolutePathValidator ensures paths are absolute.
type AbsolutePathValidator struct{}

// Validate checks if the path is absolute.
func (v *AbsolutePathValidator) Validate(path string) error {
	if !filepath.IsAbs(path) {
		return ErrInvalidPath{Path: path, Reason: "path must be absolute"}
	}
	return nil
}

// RelativePathValidator ensures paths are relative.
type RelativePathValidator struct{}

// Validate checks if the path is relative.
func (v *RelativePathValidator) Validate(path string) error {
	if filepath.IsAbs(path) {
		return ErrInvalidPath{Path: path, Reason: "path must be relative"}
	}
	return nil
}

// TraversalFreeValidator ensures paths do not contain traversal sequences.
// Rejects paths containing ".." or paths that change when cleaned.
type TraversalFreeValidator struct{}

// Validate checks if the path is free of traversal sequences.
func (v *TraversalFreeValidator) Validate(path string) error {
	// Check for explicit ".." references
	if strings.Contains(path, "..") {
		return ErrInvalidPath{Path: path, Reason: "path contains traversal sequences"}
	}

	// Verify path doesn't change when cleaned
	cleaned := filepath.Clean(path)
	if cleaned != path {
		return ErrInvalidPath{Path: path, Reason: "path contains traversal sequences"}
	}

	return nil
}

// NonEmptyPathValidator ensures paths are not empty.
type NonEmptyPathValidator struct{}

// Validate checks if the path is non-empty.
func (v *NonEmptyPathValidator) Validate(path string) error {
	if path == "" {
		return ErrInvalidPath{Path: path, Reason: "path cannot be empty"}
	}
	return nil
}

// ValidateWithValidators runs multiple validators in sequence.
// Returns the first error encountered, or nil if all validators pass.
func ValidateWithValidators(path string, validators []PathValidator) error {
	for _, validator := range validators {
		if err := validator.Validate(path); err != nil {
			return err
		}
	}
	return nil
}
