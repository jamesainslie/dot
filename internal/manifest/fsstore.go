package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamesainslie/dot/pkg/dot"
)

const manifestFileName = ".dot-manifest.json"

// FSManifestStore implements ManifestStore using filesystem
type FSManifestStore struct {
	fs dot.FS
}

// NewFSManifestStore creates filesystem-based manifest store
func NewFSManifestStore(fs dot.FS) *FSManifestStore {
	return &FSManifestStore{fs: fs}
}

// Load retrieves manifest from target directory
func (s *FSManifestStore) Load(ctx context.Context, targetDir dot.TargetPath) dot.Result[Manifest] {
	if ctx.Err() != nil {
		return dot.Err[Manifest](ctx.Err())
	}

	manifestPath := filepath.Join(targetDir.String(), manifestFileName)

	data, err := s.fs.ReadFile(ctx, manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Missing manifest is not an error - return empty manifest
			return dot.Ok(New())
		}
		return dot.Err[Manifest](fmt.Errorf("failed to read manifest: %w", err))
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return dot.Err[Manifest](fmt.Errorf("failed to parse manifest: %w", err))
	}

	return dot.Ok(m)
}

// Save persists manifest to target directory
func (s *FSManifestStore) Save(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Update timestamp
	manifest.UpdatedAt = time.Now()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestPath := filepath.Join(targetDir.String(), manifestFileName)

	// Atomic write via temp file and rename
	tempPath := manifestPath + ".tmp"

	// Write to temp file
	if err := s.fs.WriteFile(ctx, tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp manifest: %w", err)
	}

	// Atomic rename
	if err := s.fs.Rename(ctx, tempPath, manifestPath); err != nil {
		// Clean up temp file on failure
		_ = s.fs.Remove(ctx, tempPath)
		return fmt.Errorf("failed to rename manifest: %w", err)
	}

	return nil
}
