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

func TestCalculateDepth(t *testing.T) {
	tests := []struct {
		name      string
		targetDir string
		path      string
		want      int
	}{
		{"same directory", "/home/user", "/home/user", 0},
		{"one level deep", "/home/user", "/home/user/.config", 1},
		{"two levels deep", "/home/user", "/home/user/.config/nvim", 2},
		{"three levels deep", "/home/user", "/home/user/.config/nvim/lua", 3},
		{"root", "/home/user", "/home/user/.", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDepth(tt.path, tt.targetDir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldSkipDirectory(t *testing.T) {
	skipPatterns := []string{".git", "node_modules", ".cache"}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"skip .git", "/home/user/.git", true},
		{"skip node_modules", "/home/user/project/node_modules", true},
		{"skip .cache", "/home/user/.cache", true},
		{"do not skip .config", "/home/user/.config", false},
		{"do not skip regular dir", "/home/user/documents", false},
		{"skip nested .git", "/home/user/project/.git/objects", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipDirectory(tt.path, skipPatterns)
			assert.Equal(t, tt.want, got)
		})
	}
}

