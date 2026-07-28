package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectLayout(t *testing.T) {
	tests := []struct {
		name     string
		packages map[string][]string
		expected Layout
	}{
		{
			name: "dot prefixed package names",
			packages: map[string][]string{
				"dot-vim": {"dot-vimrc"},
				"dot-zsh": {"dot-zshrc"},
			},
			expected: LayoutPrefixed,
		},
		{
			name: "dot prefixed contents under plain package names",
			packages: map[string][]string{
				"vim": {"dot-vimrc"},
				"zsh": {"dot-zshrc"},
			},
			expected: LayoutPrefixed,
		},
		{
			name: "real dotfile contents",
			packages: map[string][]string{
				"nvim": {".config"},
				"zsh":  {".zshrc"},
			},
			expected: LayoutFullTree,
		},
		{
			name: "mixed layout prefers prefixed",
			packages: map[string][]string{
				"dot-vim": {"dot-vimrc"},
				"nvim":    {".config"},
			},
			expected: LayoutPrefixed,
		},
		{
			name: "one full tree package is enough",
			packages: map[string][]string{
				"bin":  {"backup.sh"},
				"nvim": {".config"},
			},
			expected: LayoutFullTree,
		},
		{
			name: "no dotfiles and no prefixes stays prefixed",
			packages: map[string][]string{
				"bin": {"backup.sh"},
			},
			expected: LayoutPrefixed,
		},
		{
			name:     "empty repository",
			packages: map[string][]string{},
			expected: LayoutPrefixed,
		},
		{
			name: "package with no contents",
			packages: map[string][]string{
				"empty": {},
			},
			expected: LayoutPrefixed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectLayout(tt.packages))
		})
	}
}

func TestRepoConfigYAML(t *testing.T) {
	data := RepoConfigYAML()

	assert.Contains(t, string(data), "dotfile:")
	assert.Contains(t, string(data), "package_name_mapping: false")
}
