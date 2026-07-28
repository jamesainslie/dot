package dot

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/bootstrap"
	"github.com/yaklabco/dot/internal/cli/selector"
)

// machineTestConfig returns a bootstrap config with a machines section.
func machineTestConfig() bootstrap.Config {
	return bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
		},
		Profiles: map[string]bootstrap.Profile{
			"work":     {Description: "Work", Packages: []string{"dot-vim"}},
			"personal": {Description: "Personal", Packages: []string{"dot-zsh"}},
		},
		Machines: []bootstrap.MachineRule{
			{Host: "*.geicoinf.com", Profile: "work"},
			{Host: "zeus*", Profile: "personal"},
		},
		Defaults: bootstrap.Defaults{Profile: "personal"},
	}
}

func newTestCloneService(fs FS) *CloneService {
	logger := adapters.NewNoopLogger()
	sel := selector.NewInteractiveSelector(strings.NewReader(""), &strings.Builder{})
	return newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)
}

func TestCloneService_ProfileFromMachines(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		config   bootstrap.Config
		opts     CloneOptions
		expected string
	}{
		{
			name:     "hostname matches first entry",
			hostname: "hephaestus.o11y.geicoinf.com",
			config:   machineTestConfig(),
			expected: "work",
		},
		{
			name:     "hostname matches later entry",
			hostname: "zeus.local",
			config:   machineTestConfig(),
			expected: "personal",
		},
		{
			name:     "explicit profile flag wins",
			hostname: "zeus.local",
			config:   machineTestConfig(),
			opts:     CloneOptions{Profile: "work"},
			expected: "",
		},
		{
			name:     "interactive suppresses machine mapping",
			hostname: "zeus.local",
			config:   machineTestConfig(),
			opts:     CloneOptions{Interactive: true},
			expected: "",
		},
		{
			name:     "no matching entry",
			hostname: "unknown-box.example.org",
			config:   machineTestConfig(),
			expected: "",
		},
		{
			name:     "no machines section",
			hostname: "zeus.local",
			config:   bootstrap.Config{Version: "1.0"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := osHostname
			osHostname = func() (string, error) { return tt.hostname, nil }
			t.Cleanup(func() { osHostname = restore })

			svc := newTestCloneService(adapters.NewMemFS())
			profile := svc.profileFromMachines(context.Background(), tt.config, tt.opts)

			assert.Equal(t, tt.expected, profile)
		})
	}
}

func TestCloneService_ProfileFromMachines_HostnameError(t *testing.T) {
	restore := osHostname
	osHostname = func() (string, error) { return "", assert.AnError }
	t.Cleanup(func() { osHostname = restore })

	svc := newTestCloneService(adapters.NewMemFS())
	profile := svc.profileFromMachines(context.Background(), machineTestConfig(), CloneOptions{})

	assert.Empty(t, profile)
}

func TestSelectedProfileName(t *testing.T) {
	tests := []struct {
		name         string
		config       bootstrap.Config
		opts         CloneOptions
		hasBootstrap bool
		expected     string
	}{
		{
			name:         "explicit profile",
			config:       machineTestConfig(),
			opts:         CloneOptions{Profile: "work"},
			hasBootstrap: true,
			expected:     "work",
		},
		{
			name:         "falls back to default profile",
			config:       machineTestConfig(),
			hasBootstrap: true,
			expected:     "personal",
		},
		{
			name:         "interactive selection records no profile",
			config:       machineTestConfig(),
			opts:         CloneOptions{Interactive: true},
			hasBootstrap: true,
			expected:     "",
		},
		{
			name:         "no bootstrap config",
			config:       bootstrap.Config{},
			opts:         CloneOptions{Profile: "work"},
			hasBootstrap: false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, selectedProfileName(tt.config, tt.opts, tt.hasBootstrap))
		})
	}
}

func TestBuildRepositoryInfo_Profile(t *testing.T) {
	info := buildRepositoryInfo("https://github.com/user/dotfiles", "main", "abc123", "work")

	require.Equal(t, "https://github.com/user/dotfiles", info.URL)
	assert.Equal(t, "main", info.Branch)
	assert.Equal(t, "abc123", info.CommitSHA)
	assert.Equal(t, "work", info.Profile)
}
