package api

import (
	"testing"

	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/stretchr/testify/assert"
)

func TestExtractManagedDirectories(t *testing.T) {
	tests := []struct {
		name     string
		manifest manifest.Manifest
		want     []string
	}{
		{
			name:     "empty manifest",
			manifest: manifest.New(),
			want:     []string{},
		},
		{
			name: "single package single link in root",
			manifest: func() manifest.Manifest {
				m := manifest.New()
				m.AddPackage(manifest.PackageInfo{
					Name:  "vim",
					Links: []string{".vimrc"},
				})
				return m
			}(),
			want: []string{"."},
		},
		{
			name: "nested links",
			manifest: func() manifest.Manifest {
				m := manifest.New()
				m.AddPackage(manifest.PackageInfo{
					Name:  "nvim",
					Links: []string{".config/nvim/init.vim", ".config/nvim/lua/init.lua"},
				})
				return m
			}(),
			want: []string{".config", ".config/nvim", ".config/nvim/lua"},
		},
		{
			name: "multiple packages various depths",
			manifest: func() manifest.Manifest {
				m := manifest.New()
				m.AddPackage(manifest.PackageInfo{
					Name:  "vim",
					Links: []string{".vimrc"},
				})
				m.AddPackage(manifest.PackageInfo{
					Name:  "zsh",
					Links: []string{".zshrc", ".zsh/aliases.zsh"},
				})
				return m
			}(),
			want: []string{".", ".zsh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractManagedDirectories(&tt.manifest)
			
			// Verify all expected directories are present
			for _, expected := range tt.want {
				assert.Contains(t, got, expected, "expected directory %s not found", expected)
			}
		})
	}
}

func TestBuildManagedLinkSet(t *testing.T) {
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:  "vim",
		Links: []string{".vimrc", ".vim/vimrc"},
	})
	m.AddPackage(manifest.PackageInfo{
		Name:  "zsh",
		Links: []string{".zshrc"},
	})

	linkSet := buildManagedLinkSet(&m)

	assert.True(t, linkSet[".vimrc"])
	assert.True(t, linkSet[".vim/vimrc"])
	assert.True(t, linkSet[".zshrc"])
	assert.False(t, linkSet[".bashrc"])
	assert.Len(t, linkSet, 3)
}

