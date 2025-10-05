package api

import (
	"errors"
	"os"
)

// isManifestNotFoundError checks if an error represents a missing manifest file.
func isManifestNotFoundError(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
