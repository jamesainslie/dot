package dot_test

import (
	"testing"
	"time"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	now := time.Now()
	status := dot.Status{
		Packages: []dot.PackageInfo{
			{
				Name:        "vim",
				InstalledAt: now,
				LinkCount:   3,
				Links:       []string{".vimrc", ".vim/colors/"},
			},
		},
	}

	require.Len(t, status.Packages, 1)
	require.Equal(t, "vim", status.Packages[0].Name)
	require.Equal(t, 3, status.Packages[0].LinkCount)
	require.Equal(t, now, status.Packages[0].InstalledAt)
}

func TestPackageInfo(t *testing.T) {
	now := time.Now()
	info := dot.PackageInfo{
		Name:        "zsh",
		InstalledAt: now,
		LinkCount:   5,
		Links:       []string{".zshrc", ".zshenv", ".zsh/"},
	}

	require.Equal(t, "zsh", info.Name)
	require.Equal(t, 5, info.LinkCount)
	require.Len(t, info.Links, 3)
}

func TestStatusEmpty(t *testing.T) {
	status := dot.Status{
		Packages: []dot.PackageInfo{},
	}

	require.Empty(t, status.Packages)
}

