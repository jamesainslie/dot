package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestErrDuplicateTarget(t *testing.T) {
	err := domain.ErrDuplicateTarget{
		TargetPath:    "/home/user/.vimrc",
		FirstPackage:  "base",
		SecondPackage: "overlay",
	}

	msg := err.Error()
	assert.Contains(t, msg, "/home/user/.vimrc")
	assert.Contains(t, msg, "base")
	assert.Contains(t, msg, "overlay")
}

func TestErrTargetKindConflict(t *testing.T) {
	err := domain.ErrTargetKindConflict{
		TargetPath:  "/home/user/.config",
		FilePackage: "base",
		DirPackage:  "overlay",
	}

	msg := err.Error()
	assert.Contains(t, msg, "/home/user/.config")
	assert.Contains(t, msg, "base")
	assert.Contains(t, msg, "overlay")
}

func TestUserFacingErrorForTargetCollisions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name: "duplicate target names both packages and the path",
			err: domain.ErrDuplicateTarget{
				TargetPath:    "/home/user/.vimrc",
				FirstPackage:  "base",
				SecondPackage: "overlay",
			},
			contains: []string{"/home/user/.vimrc", "base", "overlay"},
		},
		{
			name: "target kind conflict names both packages and the path",
			err: domain.ErrTargetKindConflict{
				TargetPath:  "/home/user/.config",
				FilePackage: "base",
				DirPackage:  "overlay",
			},
			contains: []string{"/home/user/.config", "base", "overlay"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := domain.UserFacingError(tt.err)
			for _, want := range tt.contains {
				assert.Contains(t, msg, want)
			}
		})
	}
}
