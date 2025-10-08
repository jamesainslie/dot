package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamesainslie/dot/internal/domain"
)

const manifestFileName = ".dot-manifest.json"

// FSManifestStore implements ManifestStore using filesystem
type FSManifestStore struct {
	fs domain.FS
}

// NewFSManifestStore creates filesystem-based manifest store
func NewFSManifestStore(fs domain.FS) *FSManifestStore {
	return &FSManifestStore{fs: fs}
}

// Load retrieves manifest from target directory
func (s *FSManifestStore) Load(ctx context.Context, targetDir domain.TargetPath) domain.Result[Manifest] {
	if ctx.Err() != nil {
		return domain.Err[Manifest](ctx.Err())
	}

	manifestPath := filepath.Join(targetDir.String(), manifestFileName)

	data, err := s.fs.ReadFile(ctx, manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Missing manifest is not an error - return empty manifest
			return domain.Ok(New())
		}
		return domain.Err[Manifest](fmt.Errorf("failed to read manifest: %w", err))
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return domain.Err[Manifest](fmt.Errorf("failed to parse manifest: %w", err))
	}

	return domain.Ok(m)
}

// Save persists manifest to target directory
func (s *FSManifestStore) Save(ctx context.Context, targetDir domain.TargetPath, manifest Manifest) error {
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
		// Ignore cleanup error: best-effort during error recovery.
		// Temp file (.dot-manifest.json.tmp) is harmless and will be
		// overwritten on next successful write operation.
		_ = s.fs.Remove(ctx, tempPath)
		return fmt.Errorf("failed to rename manifest: %w", err)
	}

	return nil
}
