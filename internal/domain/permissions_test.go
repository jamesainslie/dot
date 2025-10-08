package domain

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant os.FileMode
		expected os.FileMode
		desc     string
	}{
		{
			name:     "PermUserRW",
			constant: PermUserRW,
			expected: 0600,
			desc:     "user read/write only",
		},
		{
			name:     "PermUserRWX",
			constant: PermUserRWX,
			expected: 0700,
			desc:     "user read/write/execute only",
		},
		{
			name:     "PermUserW",
			constant: PermUserW,
			expected: 0200,
			desc:     "user write bit",
		},
		{
			name:     "PermGroupWorld",
			constant: PermGroupWorld,
			expected: 0044,
			desc:     "group/world readable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant,
				"permission constant %s should equal %o (%s)", tt.name, tt.expected, tt.desc)
		})
	}
}

func TestPermissionBitPatterns(t *testing.T) {
	t.Run("PermUserRW excludes execute", func(t *testing.T) {
		assert.Equal(t, os.FileMode(0), PermUserRW&0111, "should have no execute bits")
	})

	t.Run("PermUserRW has read and write", func(t *testing.T) {
		assert.NotEqual(t, os.FileMode(0), PermUserRW&0400, "should have user read bit")
		assert.NotEqual(t, os.FileMode(0), PermUserRW&0200, "should have user write bit")
	})

	t.Run("PermUserRWX has all user permissions", func(t *testing.T) {
		assert.NotEqual(t, os.FileMode(0), PermUserRWX&0400, "should have user read bit")
		assert.NotEqual(t, os.FileMode(0), PermUserRWX&0200, "should have user write bit")
		assert.NotEqual(t, os.FileMode(0), PermUserRWX&0100, "should have user execute bit")
	})

	t.Run("PermUserRWX excludes group and world", func(t *testing.T) {
		assert.Equal(t, os.FileMode(0), PermUserRWX&0077, "should have no group/world bits")
	})

	t.Run("PermGroupWorld excludes user", func(t *testing.T) {
		assert.Equal(t, os.FileMode(0), PermGroupWorld&0700, "should have no user bits")
	})

	t.Run("PermGroupWorld has group and world read", func(t *testing.T) {
		assert.NotEqual(t, os.FileMode(0), PermGroupWorld&0040, "should have group read bit")
		assert.NotEqual(t, os.FileMode(0), PermGroupWorld&0004, "should have world read bit")
	})
}

func TestPermissionSecurityInvariants(t *testing.T) {
	t.Run("secure file permissions exclude group/world", func(t *testing.T) {
		// Secure files (like config files) should only be user-accessible
		assert.Equal(t, os.FileMode(0), PermUserRW&PermGroupWorld,
			"secure file permissions should not overlap with group/world permissions")
	})

	t.Run("secure directory permissions exclude group/world", func(t *testing.T) {
		assert.Equal(t, os.FileMode(0), PermUserRWX&PermGroupWorld,
			"secure directory permissions should not overlap with group/world permissions")
	})

	t.Run("user write bit is contained in user read/write", func(t *testing.T) {
		assert.Equal(t, PermUserW, PermUserRW&PermUserW,
			"user write bit should be set in user read/write permissions")
	})
}

func TestPermissionDocumentation(t *testing.T) {
	// This test documents the intended use of each permission constant
	t.Run("usage documentation", func(t *testing.T) {
		useCases := map[string]struct {
			perm os.FileMode
			use  string
		}{
			"PermUserRW": {
				perm: PermUserRW,
				use:  "configuration files, sensitive data, manifest files",
			},
			"PermUserRWX": {
				perm: PermUserRWX,
				use:  "configuration directories, cache directories, data directories",
			},
			"PermUserW": {
				perm: PermUserW,
				use:  "write permission checks, permission validation",
			},
			"PermGroupWorld": {
				perm: PermGroupWorld,
				use:  "detecting insecure permissions, security validation",
			},
		}

		for name, uc := range useCases {
			t.Logf("%s (0%o): %s", name, uc.perm, uc.use)
		}
	})
}
