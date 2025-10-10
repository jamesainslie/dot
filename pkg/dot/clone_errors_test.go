package dot

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrPackageDirNotEmpty(t *testing.T) {
	err := ErrPackageDirNotEmpty{Path: "/home/user/dotfiles"}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "/home/user/dotfiles")
	assert.Contains(t, err.Error(), "not empty")
}

func TestErrPackageDirNotEmpty_Unwrap(t *testing.T) {
	baseErr := errors.New("directory contains files")
	err := ErrPackageDirNotEmpty{Path: "/test", Cause: baseErr}

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, baseErr, unwrapped)
}

func TestErrBootstrapNotFound(t *testing.T) {
	err := ErrBootstrapNotFound{Path: "/repo/.dotbootstrap.yaml"}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ".dotbootstrap.yaml")
	assert.Contains(t, err.Error(), "not found")
}

func TestErrInvalidBootstrap(t *testing.T) {
	err := ErrInvalidBootstrap{Reason: "missing version field"}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	assert.Contains(t, err.Error(), "missing version field")
}

func TestErrInvalidBootstrap_Unwrap(t *testing.T) {
	baseErr := errors.New("YAML parse error")
	err := ErrInvalidBootstrap{Reason: "syntax error", Cause: baseErr}

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, baseErr, unwrapped)
}

func TestErrAuthFailed(t *testing.T) {
	baseErr := errors.New("invalid token")
	err := ErrAuthFailed{Cause: baseErr}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, baseErr, unwrapped)
}

func TestErrCloneFailed(t *testing.T) {
	baseErr := errors.New("connection timeout")
	err := ErrCloneFailed{
		URL:   "https://github.com/user/repo",
		Cause: baseErr,
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "clone failed")
	assert.Contains(t, err.Error(), "github.com/user/repo")

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, baseErr, unwrapped)
}

func TestErrProfileNotFound(t *testing.T) {
	err := ErrProfileNotFound{Profile: "minimal"}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimal")
	assert.Contains(t, err.Error(), "not found")
}

func TestErrorUnwrapping(t *testing.T) {
	// Test error chain unwrapping
	innerErr := errors.New("root cause")
	wrappedErr := fmt.Errorf("wrapped: %w", innerErr)
	cloneErr := ErrCloneFailed{
		URL:   "https://github.com/user/repo",
		Cause: wrappedErr,
	}

	// Test errors.Is
	assert.True(t, errors.Is(cloneErr, innerErr))

	// Test errors.Unwrap
	unwrapped := errors.Unwrap(cloneErr)
	assert.Equal(t, wrappedErr, unwrapped)

	// Unwrap again
	unwrapped2 := errors.Unwrap(unwrapped)
	assert.Equal(t, innerErr, unwrapped2)
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "package dir not empty",
			err:      ErrPackageDirNotEmpty{Path: "/test"},
			contains: []string{"package directory", "not empty", "/test"},
		},
		{
			name:     "bootstrap not found",
			err:      ErrBootstrapNotFound{Path: "/test/.dotbootstrap.yaml"},
			contains: []string{"bootstrap", "not found", ".dotbootstrap.yaml"},
		},
		{
			name:     "invalid bootstrap",
			err:      ErrInvalidBootstrap{Reason: "missing field"},
			contains: []string{"invalid", "bootstrap", "missing field"},
		},
		{
			name:     "auth failed",
			err:      ErrAuthFailed{Cause: errors.New("token expired")},
			contains: []string{"authentication failed", "token expired"},
		},
		{
			name:     "clone failed",
			err:      ErrCloneFailed{URL: "https://example.com", Cause: errors.New("timeout")},
			contains: []string{"clone failed", "example.com", "timeout"},
		},
		{
			name:     "profile not found",
			err:      ErrProfileNotFound{Profile: "full"},
			contains: []string{"profile", "full", "not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				assert.Contains(t, errMsg, substr)
			}
		})
	}
}
