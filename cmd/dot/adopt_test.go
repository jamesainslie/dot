package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/internal/adapters"
)

func TestCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name:     "equal strings",
			a:        "hello",
			b:        "hello",
			expected: "hello",
		},
		{
			name:     "common prefix",
			a:        ".gitconfig",
			b:        ".gitignore",
			expected: ".git",
		},
		{
			name:     "no common prefix",
			a:        "foo",
			b:        "bar",
			expected: "",
		},
		{
			name:     "one is prefix of other",
			a:        "test",
			b:        "testing",
			expected: "test",
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: "",
		},
		{
			name:     "one empty",
			a:        "hello",
			b:        "",
			expected: "",
		},
		{
			name:     "single char match",
			a:        "a",
			b:        "a",
			expected: "a",
		},
		{
			name:     "single char no match",
			a:        "a",
			b:        "b",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commonPrefix(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeriveCommonPackageName(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "empty slice",
			paths:    []string{},
			expected: "",
		},
		{
			name:     "single file",
			paths:    []string{".vimrc"},
			expected: ".vimrc",
		},
		{
			name:     "multiple files with common prefix",
			paths:    []string{".gitconfig", ".gitignore", ".git_credentials"},
			expected: ".git",
		},
		{
			name:     "multiple files no common prefix",
			paths:    []string{".vimrc", ".bashrc", ".zshrc"},
			expected: ".vimrc", // Falls back to first file
		},
		{
			name:     "common prefix too short",
			paths:    []string{".a", ".b"},
			expected: ".a", // Falls back to first file when prefix < 2 chars
		},
		{
			name:     "with directory paths",
			paths:    []string{"/home/user/.gitconfig", "/home/user/.gitignore"},
			expected: ".git",
		},
		{
			name:     "prefix with trailing special chars",
			paths:    []string{"config-test", "config-prod"},
			expected: "config",
		},
		{
			name:     "exact match",
			paths:    []string{".ssh", ".ssh"},
			expected: ".ssh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveCommonPackageName(tt.paths)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileExists(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewOSFilesystem()

	t.Run("existing file", func(t *testing.T) {
		// Create temp file
		tmpfile, err := os.CreateTemp("", "test")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())
		tmpfile.Close()

		assert.True(t, fileExists(ctx, fs, tmpfile.Name()))
	})

	t.Run("existing directory", func(t *testing.T) {
		// Create temp directory
		tmpdir, err := os.MkdirTemp("", "test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		assert.True(t, fileExists(ctx, fs, tmpdir))
	})

	t.Run("non-existing file", func(t *testing.T) {
		nonExistent := filepath.Join(os.TempDir(), "non-existent-file-12345")
		assert.False(t, fileExists(ctx, fs, nonExistent))
	})

	t.Run("empty path", func(t *testing.T) {
		assert.False(t, fileExists(ctx, fs, ""))
	})
}
