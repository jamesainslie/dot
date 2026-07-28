package bootstrap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestResolveMachineProfile(t *testing.T) {
	machines := []MachineRule{
		{Host: "zeus.local", Profile: "personal"},
		{Host: "*.geicoinf.com", Profile: "work"},
		{Host: "hephaestus*", Profile: "devhost"},
		{Host: "*", Profile: "minimal"},
	}

	tests := []struct {
		name        string
		hostname    string
		wantProfile string
		wantHost    string
		wantMatch   bool
	}{
		{
			name:        "exact full hostname",
			hostname:    "zeus.local",
			wantProfile: "personal",
			wantHost:    "zeus.local",
			wantMatch:   true,
		},
		{
			name:        "glob on domain",
			hostname:    "titan-01.geicoinf.com",
			wantProfile: "work",
			wantHost:    "*.geicoinf.com",
			wantMatch:   true,
		},
		{
			name:        "short hostname matches prefix glob",
			hostname:    "hephaestus.o11y.geicoinf.com",
			wantProfile: "work",
			wantHost:    "*.geicoinf.com",
			wantMatch:   true,
		},
		{
			name:        "first match wins over later entries",
			hostname:    "zeus.local",
			wantProfile: "personal",
			wantHost:    "zeus.local",
			wantMatch:   true,
		},
		{
			name:        "case insensitive hostname",
			hostname:    "ZEUS.LOCAL",
			wantProfile: "personal",
			wantHost:    "zeus.local",
			wantMatch:   true,
		},
		{
			name:        "catch-all matches anything else",
			hostname:    "some-random-box",
			wantProfile: "minimal",
			wantHost:    "*",
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, matched := ResolveMachineProfile(machines, tt.hostname)

			assert.Equal(t, tt.wantMatch, matched)
			if tt.wantMatch {
				assert.Equal(t, tt.wantProfile, rule.Profile)
				assert.Equal(t, tt.wantHost, rule.Host)
			}
		})
	}
}

func TestResolveMachineProfile_ShortHostname(t *testing.T) {
	machines := []MachineRule{
		{Host: "hephaestus", Profile: "devhost"},
	}

	t.Run("short pattern matches first label of fqdn", func(t *testing.T) {
		rule, matched := ResolveMachineProfile(machines, "hephaestus.o11y.geicoinf.com")
		require.True(t, matched)
		assert.Equal(t, "devhost", rule.Profile)
	})

	t.Run("short pattern matches bare hostname", func(t *testing.T) {
		rule, matched := ResolveMachineProfile(machines, "hephaestus")
		require.True(t, matched)
		assert.Equal(t, "devhost", rule.Profile)
	})

	t.Run("no match returns false", func(t *testing.T) {
		_, matched := ResolveMachineProfile(machines, "zeus.local")
		assert.False(t, matched)
	})
}

func TestResolveMachineProfile_Empty(t *testing.T) {
	t.Run("no machines", func(t *testing.T) {
		_, matched := ResolveMachineProfile(nil, "zeus.local")
		assert.False(t, matched)
	})

	t.Run("empty hostname", func(t *testing.T) {
		machines := []MachineRule{{Host: "zeus", Profile: "personal"}}
		_, matched := ResolveMachineProfile(machines, "")
		assert.False(t, matched)
	})
}

func TestConfig_Validate_Machines(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid machines section",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"work": {Description: "Work", Packages: []string{"dot-vim"}},
				},
				Machines: []MachineRule{
					{Host: "*.geicoinf.com", Profile: "work"},
				},
			},
		},
		{
			name: "empty host pattern",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"work": {Description: "Work", Packages: []string{"dot-vim"}},
				},
				Machines: []MachineRule{
					{Host: "", Profile: "work"},
				},
			},
			wantErr: true,
			errMsg:  "machine host pattern cannot be empty",
		},
		{
			name: "empty profile name",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Machines: []MachineRule{
					{Host: "zeus", Profile: ""},
				},
			},
			wantErr: true,
			errMsg:  "machine entry for host \"zeus\" has no profile",
		},
		{
			name: "unknown profile",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"work": {Description: "Work", Packages: []string{"dot-vim"}},
				},
				Machines: []MachineRule{
					{Host: "zeus", Profile: "ghost"},
				},
			},
			wantErr: true,
			errMsg:  "machine entry for host \"zeus\" references unknown profile: ghost",
		},
		{
			name: "malformed glob pattern",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"work": {Description: "Work", Packages: []string{"dot-vim"}},
				},
				Machines: []MachineRule{
					{Host: "[unclosed", Profile: "work"},
				},
			},
			wantErr: true,
			errMsg:  "invalid machine host pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestLoad_Machines(t *testing.T) {
	content := `version: "1.0"
packages:
  - name: dot-vim
  - name: dot-git
profiles:
  work:
    description: Work
    packages:
      - dot-vim
  personal:
    description: Personal
    packages:
      - dot-git
machines:
  - host: "*.geicoinf.com"
    profile: work
  - host: zeus*
    profile: personal
`

	ctx := context.Background()
	fs := adapters.NewMemFS()
	require.NoError(t, fs.WriteFile(ctx, "/.dotbootstrap.yaml", []byte(content), 0644))

	cfg, err := Load(ctx, fs, "/.dotbootstrap.yaml")
	require.NoError(t, err)

	// Ordering from the file is preserved: first match wins.
	require.Len(t, cfg.Machines, 2)
	assert.Equal(t, "*.geicoinf.com", cfg.Machines[0].Host)
	assert.Equal(t, "work", cfg.Machines[0].Profile)
	assert.Equal(t, "zeus*", cfg.Machines[1].Host)
	assert.Equal(t, "personal", cfg.Machines[1].Profile)

	rule, matched := ResolveMachineProfile(cfg.Machines, "zeus.local")
	require.True(t, matched)
	assert.Equal(t, "personal", rule.Profile)
}
