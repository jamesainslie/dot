package updater

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrewManager(t *testing.T) {
	mgr := &BrewManager{}
	assert.Equal(t, "brew", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Equal(t, []string{"brew", "upgrade", "dot"}, mgr.UpgradeCommand())
}

func TestAptManager(t *testing.T) {
	mgr := &AptManager{}
	assert.Equal(t, "apt", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "apt")
}

func TestYumManager(t *testing.T) {
	mgr := &YumManager{}
	assert.Equal(t, "yum", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "yum")
}

func TestPacmanManager(t *testing.T) {
	mgr := &PacmanManager{}
	assert.Equal(t, "pacman", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "pacman")
}

func TestDnfManager(t *testing.T) {
	mgr := &DnfManager{}
	assert.Equal(t, "dnf", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "dnf")
}

func TestZypperManager(t *testing.T) {
	mgr := &ZypperManager{}
	assert.Equal(t, "zypper", mgr.Name())
	assert.NotEmpty(t, mgr.UpgradeCommand())
	assert.Contains(t, mgr.UpgradeCommand(), "zypper")
}

func TestManualManager(t *testing.T) {
	mgr := &ManualManager{}
	assert.Equal(t, "manual", mgr.Name())
	assert.True(t, mgr.IsAvailable()) // Manual is always available
	assert.Empty(t, mgr.UpgradeCommand())
}

func TestGetPackageManager(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"brew", "brew", "brew", false},
		{"apt", "apt", "apt", false},
		{"yum", "yum", "yum", false},
		{"pacman", "pacman", "pacman", false},
		{"dnf", "dnf", "dnf", false},
		{"zypper", "zypper", "zypper", false},
		{"manual", "manual", "manual", false},
		{"unknown", "unknown-manager", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := GetPackageManager(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, mgr.Name())
		})
	}
}

func TestDetectPackageManager(t *testing.T) {
	// This test is platform-dependent, so we just verify it returns something
	mgr := DetectPackageManager()
	require.NotNil(t, mgr)
	assert.NotEmpty(t, mgr.Name())

	// On macOS, we expect brew to be preferred if available
	if runtime.GOOS == "darwin" {
		brew := &BrewManager{}
		if brew.IsAvailable() {
			assert.Equal(t, "brew", mgr.Name())
		}
	}
}

func TestResolvePackageManager(t *testing.T) {
	t.Run("auto detection", func(t *testing.T) {
		mgr, err := ResolvePackageManager("auto")
		require.NoError(t, err)
		require.NotNil(t, mgr)
		assert.NotEmpty(t, mgr.Name())
	})

	t.Run("manual always works", func(t *testing.T) {
		mgr, err := ResolvePackageManager("manual")
		require.NoError(t, err)
		assert.Equal(t, "manual", mgr.Name())
		assert.True(t, mgr.IsAvailable())
	})

	t.Run("unknown manager", func(t *testing.T) {
		_, err := ResolvePackageManager("nonexistent")
		assert.Error(t, err)
	})

	t.Run("specified manager", func(t *testing.T) {
		// Test with brew (should work on macOS if brew is installed)
		mgr, err := ResolvePackageManager("brew")
		brew := &BrewManager{}
		if brew.IsAvailable() {
			require.NoError(t, err)
			assert.Equal(t, "brew", mgr.Name())
		} else if runtime.GOOS == "darwin" {
			// brew not available on macOS
			assert.Error(t, err)
		}
		// On other platforms, brew availability varies - no assertion
	})
}

func TestPackageManager_UpgradeCommands(t *testing.T) {
	tests := []struct {
		name    string
		manager PackageManager
	}{
		{"brew", &BrewManager{}},
		{"apt", &AptManager{}},
		{"yum", &YumManager{}},
		{"pacman", &PacmanManager{}},
		{"dnf", &DnfManager{}},
		{"zypper", &ZypperManager{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.manager.UpgradeCommand()
			assert.NotEmpty(t, cmd, "upgrade command should not be empty")
			assert.Contains(t, cmd, "dot", "upgrade command should include 'dot'")
		})
	}
}

func TestBrewManager_IsAvailable(t *testing.T) {
	mgr := &BrewManager{}

	// Just verify it doesn't panic - availability depends on environment
	available := mgr.IsAvailable()

	// On macOS, brew is more commonly available
	// On other platforms, brew is less common
	// We don't assert specific value as it depends on system setup
	_ = available
}

func TestManualManager_AlwaysAvailable(t *testing.T) {
	mgr := &ManualManager{}

	assert.True(t, mgr.IsAvailable())
	assert.Empty(t, mgr.UpgradeCommand())
}

func TestPackageManagerInterface(t *testing.T) {
	// Verify all managers implement the interface
	var _ PackageManager = &BrewManager{}
	var _ PackageManager = &AptManager{}
	var _ PackageManager = &YumManager{}
	var _ PackageManager = &PacmanManager{}
	var _ PackageManager = &DnfManager{}
	var _ PackageManager = &ZypperManager{}
	var _ PackageManager = &ManualManager{}
}
