package ignore_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/stretchr/testify/assert"
)

func TestNewIgnoreSet(t *testing.T) {
	set := ignore.NewIgnoreSet()
	assert.NotNil(t, set)
}

func TestIgnoreSet_Add(t *testing.T) {
	set := ignore.NewIgnoreSet()

	err := set.Add("*.txt")
	assert.NoError(t, err)

	err = set.Add(".git")
	assert.NoError(t, err)
}

func TestIgnoreSet_ShouldIgnore(t *testing.T) {
	set := ignore.NewIgnoreSet()
	set.Add("*.txt")
	set.Add(".git")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "matches txt pattern",
			path:     "file.txt",
			expected: true,
		},
		{
			name:     "matches git pattern",
			path:     ".git",
			expected: true,
		},
		{
			name:     "no match",
			path:     "README.md",
			expected: false,
		},
		{
			name:     "matches git in subdir",
			path:     "project/.git",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := set.ShouldIgnore(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := ignore.DefaultIgnorePatterns()

	assert.NotEmpty(t, patterns)
	assert.Contains(t, patterns, ".git")
	assert.Contains(t, patterns, ".DS_Store")
}

func TestNewDefaultIgnoreSet(t *testing.T) {
	set := ignore.NewDefaultIgnoreSet()

	// Should ignore .git
	assert.True(t, set.ShouldIgnore(".git"))
	assert.True(t, set.ShouldIgnore("path/.git"))

	// Should ignore .DS_Store
	assert.True(t, set.ShouldIgnore(".DS_Store"))
	assert.True(t, set.ShouldIgnore("path/.DS_Store"))

	// Should not ignore regular files
	assert.False(t, set.ShouldIgnore("README.md"))
}

func TestIgnoreSet_AddPattern(t *testing.T) {
	set := ignore.NewIgnoreSet()

	pattern := ignore.NewPattern("*.log").Unwrap()
	set.AddPattern(pattern)

	assert.True(t, set.ShouldIgnore("error.log"))
	assert.False(t, set.ShouldIgnore("error.txt"))
}

func TestIgnoreSet_Empty(t *testing.T) {
	set := ignore.NewIgnoreSet()

	// Empty set should not ignore anything
	assert.False(t, set.ShouldIgnore("any/file"))
	assert.False(t, set.ShouldIgnore(".git"))
}

func TestIgnoreSet_Size(t *testing.T) {
	set := ignore.NewIgnoreSet()

	assert.Equal(t, 0, set.Size())

	set.Add("*.txt")
	assert.Equal(t, 1, set.Size())

	set.Add(".git")
	assert.Equal(t, 2, set.Size())
}
