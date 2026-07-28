package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveProfilePackages(t *testing.T) {
	profiles := map[string]Profile{
		"base": {
			Description: "Shared base",
			Packages:    []string{"dot-git", "dot-zsh"},
		},
		"dev": {
			Description: "Development",
			Extends:     "base",
			Packages:    []string{"dot-vim", "dot-tmux"},
		},
		"work": {
			Description: "Work laptop",
			Extends:     "dev",
			Packages:    []string{"dot-ssh", "dot-git"},
		},
		"standalone": {
			Description: "No inheritance",
			Packages:    []string{"dot-vim"},
		},
		"empty-parent": {
			Description: "Parent with no packages",
			Extends:     "no-packages",
			Packages:    []string{"dot-vim"},
		},
		"no-packages": {
			Description: "Nothing here",
		},
	}

	tests := []struct {
		name     string
		profile  string
		expected []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "profile without extends",
			profile:  "standalone",
			expected: []string{"dot-vim"},
		},
		{
			name:     "single parent, parent packages first",
			profile:  "dev",
			expected: []string{"dot-git", "dot-zsh", "dot-vim", "dot-tmux"},
		},
		{
			name:     "grandparent chain preserves root-first order and dedupes",
			profile:  "work",
			expected: []string{"dot-git", "dot-zsh", "dot-vim", "dot-tmux", "dot-ssh"},
		},
		{
			name:     "parent with no packages",
			profile:  "empty-parent",
			expected: []string{"dot-vim"},
		},
		{
			name:     "parent alone",
			profile:  "base",
			expected: []string{"dot-git", "dot-zsh"},
		},
		{
			name:    "unknown profile",
			profile: "missing",
			wantErr: true,
			errMsg:  "profile not found: missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages, err := ResolveProfilePackages(profiles, tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, packages)
		})
	}
}

func TestResolveProfilePackages_Errors(t *testing.T) {
	tests := []struct {
		name     string
		profiles map[string]Profile
		profile  string
		errMsg   string
	}{
		{
			name: "unknown parent",
			profiles: map[string]Profile{
				"dev": {Extends: "ghost", Packages: []string{"dot-vim"}},
			},
			profile: "dev",
			errMsg:  `profile "dev" extends unknown profile: ghost`,
		},
		{
			name: "self reference",
			profiles: map[string]Profile{
				"loop": {Extends: "loop", Packages: []string{"dot-vim"}},
			},
			profile: "loop",
			errMsg:  "circular extends chain",
		},
		{
			name: "two profile cycle",
			profiles: map[string]Profile{
				"a": {Extends: "b"},
				"b": {Extends: "a"},
			},
			profile: "a",
			errMsg:  "circular extends chain",
		},
		{
			name: "three profile cycle",
			profiles: map[string]Profile{
				"a": {Extends: "b"},
				"b": {Extends: "c"},
				"c": {Extends: "a"},
			},
			profile: "b",
			errMsg:  "circular extends chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages, err := ResolveProfilePackages(tt.profiles, tt.profile)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
			assert.Nil(t, packages)
		})
	}
}

func TestConfig_Validate_Extends(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid extends chain",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-git"},
					{Name: "dot-vim"},
				},
				Profiles: map[string]Profile{
					"base": {Description: "Base", Packages: []string{"dot-git"}},
					"dev":  {Description: "Dev", Extends: "base", Packages: []string{"dot-vim"}},
				},
			},
		},
		{
			name: "inherited package must exist",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"base": {Description: "Base", Packages: []string{"dot-ghost"}},
					"dev":  {Description: "Dev", Extends: "base", Packages: []string{"dot-vim"}},
				},
			},
			wantErr: true,
			errMsg:  "unknown package",
		},
		{
			name: "unknown parent profile",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"dev": {Description: "Dev", Extends: "ghost", Packages: []string{"dot-vim"}},
				},
			},
			wantErr: true,
			errMsg:  "extends unknown profile",
		},
		{
			name: "circular extends chain",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{{Name: "dot-vim"}},
				Profiles: map[string]Profile{
					"a": {Description: "A", Extends: "b", Packages: []string{"dot-vim"}},
					"b": {Description: "B", Extends: "a"},
				},
			},
			wantErr: true,
			errMsg:  "circular extends chain",
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

func TestGetProfile_ResolvesExtends(t *testing.T) {
	cfg := Config{
		Version: "1.0",
		Packages: []PackageSpec{
			{Name: "dot-git"},
			{Name: "dot-vim"},
		},
		Profiles: map[string]Profile{
			"base": {Description: "Base", Packages: []string{"dot-git"}},
			"dev":  {Description: "Dev", Extends: "base", Packages: []string{"dot-vim"}},
		},
	}

	packages, err := GetProfile(cfg, "dev")
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-git", "dot-vim"}, packages)
}
