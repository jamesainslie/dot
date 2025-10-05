package manifest

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
)

func mustTargetPath(t *testing.T, path string) dot.TargetPath {
	t.Helper()
	result := dot.NewTargetPath(path)
	if result.IsErr() {
		t.Fatalf("failed to create target path: %v", result.UnwrapErr())
	}
	return result.Unwrap()
}

func mustPackagePath(t *testing.T, path string) dot.PackagePath {
	t.Helper()
	result := dot.NewPackagePath(path)
	if result.IsErr() {
		t.Fatalf("failed to create package path: %v", result.UnwrapErr())
	}
	return result.Unwrap()
}

