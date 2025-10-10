package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMethodTypes(t *testing.T) {
	t.Run("NoAuth", func(t *testing.T) {
		auth := NoAuth{}
		assert.NotNil(t, auth)
		assert.Implements(t, (*AuthMethod)(nil), auth)
	})

	t.Run("TokenAuth", func(t *testing.T) {
		auth := TokenAuth{Token: "ghp_test123"}
		assert.Equal(t, "ghp_test123", auth.Token)
		assert.Implements(t, (*AuthMethod)(nil), auth)
	})

	t.Run("SSHAuth", func(t *testing.T) {
		auth := SSHAuth{PrivateKeyPath: "/home/user/.ssh/id_rsa"}
		assert.Equal(t, "/home/user/.ssh/id_rsa", auth.PrivateKeyPath)
		assert.Implements(t, (*AuthMethod)(nil), auth)
	})
}

func TestCloneOptions(t *testing.T) {
	t.Run("with all options", func(t *testing.T) {
		auth := TokenAuth{Token: "test"}
		opts := CloneOptions{
			Auth:   auth,
			Branch: "main",
			Depth:  1,
		}

		assert.NotNil(t, opts.Auth)
		assert.Equal(t, "main", opts.Branch)
		assert.Equal(t, 1, opts.Depth)
	})

	t.Run("with defaults", func(t *testing.T) {
		opts := CloneOptions{}
		assert.Nil(t, opts.Auth)
		assert.Empty(t, opts.Branch)
		assert.Zero(t, opts.Depth)
	})
}
