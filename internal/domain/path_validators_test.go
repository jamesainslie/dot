package domain

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbsolutePathValidator(t *testing.T) {
	validator := &AbsolutePathValidator{}

	t.Run("accepts absolute paths", func(t *testing.T) {
		absolutePaths := []string{
			"/absolute/path",
			"/home/user/config",
			"/etc/config.yaml",
		}

		for _, path := range absolutePaths {
			err := validator.Validate(path)
			assert.NoError(t, err, "should accept absolute path: %s", path)
		}
	})

	t.Run("rejects relative paths", func(t *testing.T) {
		relativePaths := []string{
			"relative/path",
			"./current/dir",
			"../parent/dir",
			"file.txt",
		}

		for _, path := range relativePaths {
			err := validator.Validate(path)
			require.Error(t, err, "should reject relative path: %s", path)
			assert.Contains(t, err.Error(), "must be absolute")
		}
	})

	t.Run("handles empty path", func(t *testing.T) {
		err := validator.Validate("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be absolute")
	})
}

func TestRelativePathValidator(t *testing.T) {
	validator := &RelativePathValidator{}

	t.Run("accepts relative paths", func(t *testing.T) {
		relativePaths := []string{
			"relative/path",
			"./current/dir",
			"../parent/dir",
			"file.txt",
			"subdir/file.txt",
		}

		for _, path := range relativePaths {
			err := validator.Validate(path)
			assert.NoError(t, err, "should accept relative path: %s", path)
		}
	})

	t.Run("rejects absolute paths", func(t *testing.T) {
		absolutePaths := []string{
			"/absolute/path",
			"/home/user/config",
		}

		for _, path := range absolutePaths {
			err := validator.Validate(path)
			require.Error(t, err, "should reject absolute path: %s", path)
			assert.Contains(t, err.Error(), "must be relative")
		}
	})

	t.Run("handles empty path", func(t *testing.T) {
		err := validator.Validate("")
		// Empty path is considered relative
		assert.NoError(t, err)
	})
}

func TestTraversalFreeValidator(t *testing.T) {
	validator := &TraversalFreeValidator{}

	t.Run("accepts clean paths", func(t *testing.T) {
		cleanPaths := []string{
			"/clean/path",
			"relative/path",
			"/home/user/file.txt",
			"subdir/file",
		}

		for _, path := range cleanPaths {
			err := validator.Validate(path)
			assert.NoError(t, err, "should accept clean path: %s", path)
		}
	})

	t.Run("rejects paths with parent directory references", func(t *testing.T) {
		traversalPaths := []string{
			"../parent",
			"path/../other",
			"/home/../etc/passwd",
			"../../escape",
		}

		for _, path := range traversalPaths {
			err := validator.Validate(path)
			require.Error(t, err, "should reject traversal path: %s", path)
			assert.Contains(t, err.Error(), "traversal")
		}
	})

	t.Run("rejects paths that change when cleaned", func(t *testing.T) {
		dirtyPaths := []string{
			"/path//double//slash",
			"/path/./current",
			"path/to/./file",
		}

		for _, path := range dirtyPaths {
			cleaned := filepath.Clean(path)
			if path != cleaned {
				err := validator.Validate(path)
				require.Error(t, err, "should reject dirty path: %s", path)
			}
		}
	})

	t.Run("handles empty path", func(t *testing.T) {
		err := validator.Validate("")
		// Empty path cleans to "." which is different
		require.Error(t, err)
	})
}

func TestNonEmptyPathValidator(t *testing.T) {
	validator := &NonEmptyPathValidator{}

	t.Run("accepts non-empty paths", func(t *testing.T) {
		paths := []string{
			"/absolute",
			"relative",
			".",
			"..",
		}

		for _, path := range paths {
			err := validator.Validate(path)
			assert.NoError(t, err, "should accept non-empty path: %s", path)
		}
	})

	t.Run("rejects empty path", func(t *testing.T) {
		err := validator.Validate("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})
}

func TestValidatorChaining(t *testing.T) {
	t.Run("chain multiple validators", func(t *testing.T) {
		validators := []PathValidator{
			&NonEmptyPathValidator{},
			&AbsolutePathValidator{},
			&TraversalFreeValidator{},
		}

		// Valid path passes all validators
		validPath := "/clean/absolute/path"
		for _, v := range validators {
			err := v.Validate(validPath)
			assert.NoError(t, err)
		}

		// Invalid path fails appropriate validator
		invalidPath := "relative/path"
		passedFirst := false
		for _, v := range validators {
			err := v.Validate(invalidPath)
			if err != nil {
				assert.True(t, passedFirst, "should fail on second validator (absolute check)")
				break
			}
			passedFirst = true
		}
	})
}

func TestValidateWithValidators(t *testing.T) {
	t.Run("runs all validators in order", func(t *testing.T) {
		validators := []PathValidator{
			&NonEmptyPathValidator{},
			&AbsolutePathValidator{},
		}

		err := ValidateWithValidators("/valid/path", validators)
		assert.NoError(t, err)
	})

	t.Run("stops at first validation error", func(t *testing.T) {
		validators := []PathValidator{
			&NonEmptyPathValidator{},
			&AbsolutePathValidator{}, // This will fail
			&TraversalFreeValidator{},
		}

		err := ValidateWithValidators("relative/path", validators)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("handles empty validator list", func(t *testing.T) {
		err := ValidateWithValidators("/any/path", []PathValidator{})
		assert.NoError(t, err, "empty validator list should pass")
	})

	t.Run("validates empty path with appropriate validator", func(t *testing.T) {
		validators := []PathValidator{
			&NonEmptyPathValidator{},
		}

		err := ValidateWithValidators("", validators)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})
}

func TestPathValidatorErrorTypes(t *testing.T) {
	t.Run("AbsolutePathValidator returns ErrInvalidPath", func(t *testing.T) {
		validator := &AbsolutePathValidator{}
		err := validator.Validate("relative")

		require.Error(t, err)
		var pathErr ErrInvalidPath
		assert.True(t, As(err, &pathErr), "should be ErrInvalidPath type")
	})

	t.Run("TraversalFreeValidator returns ErrInvalidPath", func(t *testing.T) {
		validator := &TraversalFreeValidator{}
		err := validator.Validate("../traversal")

		require.Error(t, err)
		var pathErr ErrInvalidPath
		assert.True(t, As(err, &pathErr), "should be ErrInvalidPath type")
	})

	t.Run("NonEmptyPathValidator returns ErrInvalidPath", func(t *testing.T) {
		validator := &NonEmptyPathValidator{}
		err := validator.Validate("")

		require.Error(t, err)
		var pathErr ErrInvalidPath
		assert.True(t, As(err, &pathErr), "should be ErrInvalidPath type")
	})
}

// Helper to check error types (like errors.As but returns bool)
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	// Simple type assertion for our use case
	if t, ok := target.(*ErrInvalidPath); ok {
		if e, ok := err.(ErrInvalidPath); ok {
			*t = e
			return true
		}
	}
	return false
}
