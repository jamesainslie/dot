package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAuth_WithGitHubToken(t *testing.T) {
	ctx := context.Background()

	// Set GITHUB_TOKEN environment variable
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	os.Setenv("GITHUB_TOKEN", "ghp_test123")

	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	tokenAuth, ok := auth.(TokenAuth)
	assert.True(t, ok)
	assert.Equal(t, "ghp_test123", tokenAuth.Token)
}

func TestResolveAuth_WithGitToken(t *testing.T) {
	ctx := context.Background()

	// Clear GITHUB_TOKEN and set GIT_TOKEN
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		} else {
			os.Unsetenv("GIT_TOKEN")
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Setenv("GIT_TOKEN", "token123")

	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	tokenAuth, ok := auth.(TokenAuth)
	assert.True(t, ok)
	assert.Equal(t, "token123", tokenAuth.Token)
}

func TestResolveAuth_WithSSHKey(t *testing.T) {
	ctx := context.Background()

	// Clear token environment variables
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")

	// Create a temporary SSH key file
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "id_rsa")
	err := os.WriteFile(keyPath, []byte("fake-ssh-key"), 0600)
	require.NoError(t, err)

	// Create mock home directory structure
	sshDir := filepath.Join(tempDir, ".ssh")
	err = os.MkdirAll(sshDir, 0700)
	require.NoError(t, err)

	mockKeyPath := filepath.Join(sshDir, "id_rsa")
	err = os.WriteFile(mockKeyPath, []byte("fake-ssh-key"), 0600)
	require.NoError(t, err)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()
	os.Setenv("HOME", tempDir)

	auth, err := ResolveAuth(ctx, "git@github.com:user/repo.git")
	require.NoError(t, err)

	sshAuth, ok := auth.(SSHAuth)
	assert.True(t, ok)
	assert.Equal(t, mockKeyPath, sshAuth.PrivateKeyPath)
}

func TestResolveAuth_NoAuth(t *testing.T) {
	ctx := context.Background()

	// Clear all auth environment variables
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")
	os.Setenv("HOME", "/nonexistent")

	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	_, ok := auth.(NoAuth)
	assert.True(t, ok)
}

func TestResolveAuth_SSHURLWithoutKeys(t *testing.T) {
	ctx := context.Background()

	// Clear all auth
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")
	os.Setenv("HOME", "/nonexistent")

	auth, err := ResolveAuth(ctx, "git@github.com:user/repo.git")
	require.NoError(t, err)

	// Should fall back to NoAuth if no SSH keys found
	_, ok := auth.(NoAuth)
	assert.True(t, ok)
}

func TestFindSSHKey(t *testing.T) {
	t.Run("finds id_rsa", func(t *testing.T) {
		tempDir := t.TempDir()
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		keyPath := filepath.Join(sshDir, "id_rsa")
		err = os.WriteFile(keyPath, []byte("fake-key"), 0600)
		require.NoError(t, err)

		found := findSSHKey(tempDir)
		assert.Equal(t, keyPath, found)
	})

	t.Run("finds id_ed25519", func(t *testing.T) {
		tempDir := t.TempDir()
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		keyPath := filepath.Join(sshDir, "id_ed25519")
		err = os.WriteFile(keyPath, []byte("fake-key"), 0600)
		require.NoError(t, err)

		found := findSSHKey(tempDir)
		assert.Equal(t, keyPath, found)
	})

	t.Run("prefers id_ed25519 over id_rsa", func(t *testing.T) {
		tempDir := t.TempDir()
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		rsaPath := filepath.Join(sshDir, "id_rsa")
		err = os.WriteFile(rsaPath, []byte("fake-key"), 0600)
		require.NoError(t, err)

		ed25519Path := filepath.Join(sshDir, "id_ed25519")
		err = os.WriteFile(ed25519Path, []byte("fake-key"), 0600)
		require.NoError(t, err)

		found := findSSHKey(tempDir)
		assert.Equal(t, ed25519Path, found)
	})

	t.Run("returns empty when no keys found", func(t *testing.T) {
		tempDir := t.TempDir()
		found := findSSHKey(tempDir)
		assert.Empty(t, found)
	})
}

func TestIsSSHURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"git@github.com:user/repo.git", true},
		{"git@gitlab.com:user/repo.git", true},
		{"ssh://git@github.com/user/repo.git", true},
		{"https://github.com/user/repo", false},
		{"http://github.com/user/repo", false},
		{"file:///path/to/repo", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isSSHURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}
