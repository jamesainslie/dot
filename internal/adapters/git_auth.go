package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// ResolveAuth determines the appropriate authentication method for a repository URL.
//
// Resolution priority:
//  1. GITHUB_TOKEN environment variable → TokenAuth
//  2. GIT_TOKEN environment variable → TokenAuth
//  3. SSH keys in ~/.ssh/ → SSHAuth (for SSH URLs)
//  4. NoAuth (public repositories)
//
// The function inspects the URL to determine if SSH auth is needed.
// For HTTPS URLs, token auth is preferred if available.
// For SSH URLs (git@... or ssh://...), SSH key auth is used if keys exist.
func ResolveAuth(ctx context.Context, repoURL string) (AuthMethod, error) {
	// Check for token in environment variables
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return TokenAuth{Token: token}, nil
	}

	if token := os.Getenv("GIT_TOKEN"); token != "" {
		return TokenAuth{Token: token}, nil
	}

	// For SSH URLs, try to find SSH keys
	if isSSHURL(repoURL) {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			if keyPath := findSSHKey(homeDir); keyPath != "" {
				return SSHAuth{PrivateKeyPath: keyPath}, nil
			}
		}
	}

	// Fall back to no authentication (public repos)
	return NoAuth{}, nil
}

// isSSHURL checks if a URL uses SSH protocol.
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://")
}

// findSSHKey searches for common SSH private keys in the user's home directory.
//
// Checks for keys in this order:
//  1. ~/.ssh/id_ed25519 (modern, preferred)
//  2. ~/.ssh/id_rsa (older, common)
//
// Returns the path to the first key found, or empty string if none exist.
func findSSHKey(homeDir string) string {
	sshDir := filepath.Join(homeDir, ".ssh")

	// Check for Ed25519 key (preferred)
	ed25519Key := filepath.Join(sshDir, "id_ed25519")
	if _, err := os.Stat(ed25519Key); err == nil {
		return ed25519Key
	}

	// Check for RSA key (fallback)
	rsaKey := filepath.Join(sshDir, "id_rsa")
	if _, err := os.Stat(rsaKey); err == nil {
		return rsaKey
	}

	return ""
}
