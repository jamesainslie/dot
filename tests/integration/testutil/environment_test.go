package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestEnvironment(t *testing.T) {
	env := NewTestEnvironment(t)

	// Verify directories created
	assert.DirExists(t, env.PackageDir)
	assert.DirExists(t, env.TargetDir)

	// Verify context
	assert.NotNil(t, env.Context())

	// Verify fixture builder
	fb := env.FixtureBuilder()
	assert.NotNil(t, fb)
}

func TestTestEnvironment_CreatePackage(t *testing.T) {
	env := NewTestEnvironment(t)

	files := map[string]string{
		"dot-vimrc":          "set nocompatible",
		"dot-vim/colors.vim": "colorscheme default",
	}

	pkgPath := env.CreatePackage("vim", files)

	// Verify package created
	assert.DirExists(t, pkgPath)
	assert.Equal(t, filepath.Join(env.PackageDir, "vim"), pkgPath)

	// Verify files
	vimrc := filepath.Join(pkgPath, "dot-vimrc")
	assert.FileExists(t, vimrc)

	colors := filepath.Join(pkgPath, "dot-vim/colors.vim")
	assert.FileExists(t, colors)
}

func TestTestEnvironment_CreateSimplePackage(t *testing.T) {
	env := NewTestEnvironment(t)

	pkgPath := env.CreateSimplePackage("zsh", "dot-zshrc", "export EDITOR=vim")

	// Verify package created
	assert.DirExists(t, pkgPath)

	// Verify file
	zshrc := filepath.Join(pkgPath, "dot-zshrc")
	assert.FileExists(t, zshrc)

	content, err := os.ReadFile(zshrc)
	require.NoError(t, err)
	assert.Equal(t, "export EDITOR=vim", string(content))
}

func TestTestEnvironment_AddCleanup(t *testing.T) {
	env := NewTestEnvironment(t)

	cleaned := false
	env.AddCleanup(func() {
		cleaned = true
	})

	// Trigger cleanup
	env.Cleanup()

	// Verify cleanup ran
	assert.True(t, cleaned)
}

func TestTestEnvironment_ManifestPath(t *testing.T) {
	env := NewTestEnvironment(t)

	manifestPath := env.ManifestPath()
	assert.Contains(t, manifestPath, "manifest.json")
}
