package manifest

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// ManifestStore provides persistence for manifests
type ManifestStore interface {
	// Load retrieves manifest from target directory
	// Returns empty manifest if file doesn't exist
	Load(ctx context.Context, targetDir dot.TargetPath) dot.Result[Manifest]

	// Save persists manifest to target directory
	// Write is atomic via temp file and rename
	Save(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) error
}

