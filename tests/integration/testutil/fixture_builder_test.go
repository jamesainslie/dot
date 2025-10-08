package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureBuilder_Package(t *testing.T) {
	tmpDir := t.TempDir()
	fb := NewFixtureBuilder(t, tmpDir)

	pkgPath := fb.Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		WithFile("dot-vim/colors/theme.vim", "colorscheme default").
		WithDir("dot-vim/syntax").
		Create()

	// Verify package directory created
	assert.DirExists(t, pkgPath)
	assert.Equal(t, filepath.Join(tmpDir, "vim"), pkgPath)

	// Verify files created
	vimrc := filepath.Join(pkgPath, "dot-vimrc")
	assert.FileExists(t, vimrc)
	content, err := os.ReadFile(vimrc)
	require.NoError(t, err)
	assert.Equal(t, "set nocompatible", string(content))

	// Verify nested files
	theme := filepath.Join(pkgPath, "dot-vim/colors/theme.vim")
	assert.FileExists(t, theme)

	// Verify directories
	syntax := filepath.Join(pkgPath, "dot-vim/syntax")
	assert.DirExists(t, syntax)
}

func TestFixtureBuilder_FileTree(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	tmpDir := t.TempDir()
	fb := NewFixtureBuilder(t, tmpDir)

	base := filepath.Join(tmpDir, "target")
	require.NoError(t, os.MkdirAll(base, 0755))

	fb.FileTree(base).
		File(".bashrc", "export PATH=/usr/local/bin:$PATH").
		Dir(".config").
		File(".config/nvim/init.vim", "syntax on").
		Symlink("/etc/hosts", ".hosts")

	// Verify file
	bashrc := filepath.Join(base, ".bashrc")
	assert.FileExists(t, bashrc)

	// Verify directory
	config := filepath.Join(base, ".config")
	assert.DirExists(t, config)

	// Verify nested file
	initVim := filepath.Join(base, ".config/nvim/init.vim")
	assert.FileExists(t, initVim)

	// Verify symlink
	hosts := filepath.Join(base, ".hosts")
	info, err := os.Lstat(hosts)
	require.NoError(t, err)
	assert.True(t, info.Mode()&os.ModeSymlink != 0)
}

func TestFixtureBuilder_FileWithMode(t *testing.T) {
	tmpDir := t.TempDir()
	fb := NewFixtureBuilder(t, tmpDir)

	base := filepath.Join(tmpDir, "target")
	require.NoError(t, os.MkdirAll(base, 0755))

	fb.FileTree(base).
		FileWithMode("script.sh", "#!/bin/bash", 0755)

	script := filepath.Join(base, "script.sh")
	info, err := os.Stat(script)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}
