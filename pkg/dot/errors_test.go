package dot_test

import (
	"errors"
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

func TestErrInvalidPath(t *testing.T) {
	err := dot.ErrInvalidPath{
		Path:   "/some/path",
		Reason: "must be absolute",
	}
	
	assert.Contains(t, err.Error(), "/some/path")
	assert.Contains(t, err.Error(), "must be absolute")
}

func TestErrPackageNotFound(t *testing.T) {
	err := dot.ErrPackageNotFound{
		Package: "vim",
	}
	
	assert.Contains(t, err.Error(), "vim")
	assert.Contains(t, err.Error(), "not found")
}

func TestErrConflict(t *testing.T) {
	err := dot.ErrConflict{
		Path:   "/home/user/.vimrc",
		Reason: "file already exists",
	}
	
	assert.Contains(t, err.Error(), "/home/user/.vimrc")
	assert.Contains(t, err.Error(), "file already exists")
}

func TestErrCyclicDependency(t *testing.T) {
	err := dot.ErrCyclicDependency{
		Cycle: []string{"a", "b", "c", "a"},
	}
	
	msg := err.Error()
	assert.Contains(t, msg, "a")
	assert.Contains(t, msg, "b")
	assert.Contains(t, msg, "c")
	assert.Contains(t, msg, "cyclic")
}

func TestErrFilesystemOperation(t *testing.T) {
	inner := errors.New("permission denied")
	err := dot.ErrFilesystemOperation{
		Operation: "create symlink",
		Path:      "/home/user/.vimrc",
		Err:       inner,
	}
	
	assert.Contains(t, err.Error(), "create symlink")
	assert.Contains(t, err.Error(), "/home/user/.vimrc")
	assert.ErrorIs(t, err, inner)
}

func TestErrPermissionDenied(t *testing.T) {
	err := dot.ErrPermissionDenied{
		Path:      "/root/.vimrc",
		Operation: "write",
	}
	
	assert.Contains(t, err.Error(), "/root/.vimrc")
	assert.Contains(t, err.Error(), "write")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestErrMultiple(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	
	multi := dot.ErrMultiple{
		Errors: []error{err1, err2, err3},
	}
	
	msg := multi.Error()
	assert.Contains(t, msg, "3 errors")
	assert.Contains(t, msg, "error 1")
	assert.Contains(t, msg, "error 2")
	assert.Contains(t, msg, "error 3")
}

func TestErrMultipleUnwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	
	multi := dot.ErrMultiple{
		Errors: []error{err1, err2},
	}
	
	unwrapped := multi.Unwrap()
	assert.Len(t, unwrapped, 2)
	assert.Equal(t, err1, unwrapped[0])
	assert.Equal(t, err2, unwrapped[1])
}

func TestUserFacingErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name: "ErrPackageNotFound",
			err: dot.ErrPackageNotFound{
				Package: "vim",
			},
			contains: []string{"vim", "not found"},
		},
		{
			name: "ErrInvalidPath",
			err: dot.ErrInvalidPath{
				Path:   "relative/path",
				Reason: "must be absolute",
			},
			contains: []string{"relative/path", "absolute"},
		},
		{
			name: "ErrConflict",
			err: dot.ErrConflict{
				Path:   "/home/user/.vimrc",
				Reason: "file exists",
			},
			contains: []string{".vimrc", "file exists"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := dot.UserFacingError(tt.err)
			for _, expected := range tt.contains {
				assert.Contains(t, msg, expected)
			}
		})
	}
}

