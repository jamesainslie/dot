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
	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/pipeline"
	"github.com/yaklabco/dot/internal/planner"
)

// machinesBootstrapYAML maps work hosts and personal hosts to different
// profiles, and names a third profile as the default.
const machinesBootstrapYAML = `version: "1.0"
packages:
  - name: dot-vim
  - name: dot-zsh
profiles:
  work:
    description: Work
    packages:
      - dot-vim
  personal:
    description: Personal
    packages:
      - dot-zsh
machines:
  - host: "*.geicoinf.com"
    profile: work
  - host: zeus*
    profile: personal
defaults:
  profile: personal
`

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
	return newTestCloneServiceWithStore(fs, manifest.NewFSManifestStore(fs))
}

func newTestCloneServiceWithStore(fs FS, store *manifest.FSManifestStore) *CloneService {
	logger := adapters.NewNoopLogger()
	sel := selector.NewInteractiveSelector(strings.NewReader(""), &strings.Builder{})
	return newCloneService(fs, logger, nil, nil, sel, store, "/packages", "/home", false)
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
			t.Parallel()

			svc := newTestCloneService(adapters.NewMemFS())
			svc.hostname = func() (string, error) { return tt.hostname, nil }
			profile := svc.profileFromMachines(context.Background(), tt.config, tt.opts)

			assert.Equal(t, tt.expected, profile)
		})
	}
}

func TestCloneService_ProfileFromMachines_HostnameError(t *testing.T) {
	t.Parallel()

	svc := newTestCloneService(adapters.NewMemFS())
	svc.hostname = func() (string, error) { return "", assert.AnError }
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

func TestCloneService_UpdateManifestRepository_UsesConfiguredStore(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	svc := newTestCloneServiceWithStore(fs, manifest.NewFSManifestStoreWithDir(fs, "/state/dot"))

	err := svc.updateManifestRepository(ctx, manifest.RepositoryInfo{
		URL:     "https://github.com/user/dotfiles",
		Branch:  "main",
		Profile: "work",
	})
	require.NoError(t, err)

	loaded := manifest.NewFSManifestStoreWithDir(fs, "/state/dot").Load(ctx, NewTargetPath("/home").Unwrap())
	require.True(t, loaded.IsOk())

	stored := loaded.Unwrap()
	repo, exists := stored.GetRepository()
	require.True(t, exists)
	assert.Equal(t, "work", repo.Profile)
}

// TestNewClient_CloneServiceWritesToConfiguredStore checks that the clone
// service a client builds records repository info where every other command
// reads it, rather than in the target directory.
func TestNewClient_CloneServiceWritesToConfiguredStore(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	require.NoError(t, fs.MkdirAll(ctx, "/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/state", 0755))

	client, err := NewClient(Config{
		PackageDir:  "/packages",
		TargetDir:   "/home",
		ManifestDir: "/state",
		FS:          fs,
		Logger:      adapters.NewNoopLogger(),
	})
	require.NoError(t, err)

	err = client.cloneSvc.updateManifestRepository(ctx, manifest.RepositoryInfo{
		URL:     "https://github.com/user/dotfiles",
		Branch:  "main",
		Profile: "work",
	})
	require.NoError(t, err)

	info, exists, err := client.RepositoryInfo(ctx)
	require.NoError(t, err)
	require.True(t, exists, "repository info should be readable through the client")
	assert.Equal(t, "work", info.Profile)
	assert.False(t, fs.Exists(ctx, "/home/.dot-manifest.json"), "manifest should not land in the target directory")
}

// newMachinesCloneService wires a clone service with a real manage pipeline so
// package installation and manifest writes happen for real.
func newMachinesCloneService(t *testing.T, fs FS, store *manifest.FSManifestStore, hostname string) *CloneService {
	t.Helper()

	logger := adapters.NewNoopLogger()
	managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
		FS:                 fs,
		IgnoreSet:          ignore.NewDefaultIgnoreSet(),
		Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicySkip},
		PackageNameMapping: false,
	})
	exec := executor.New(executor.Opts{FS: fs, Logger: logger, Tracer: adapters.NewNoopTracer()})
	manifestSvc := newManifestService(fs, logger, store)
	unmanageSvc := newUnmanageService(fs, logger, exec, manifestSvc, "/packages", "/home", false)
	manageSvc := newManageService(fs, logger, managePipe, exec, manifestSvc, unmanageSvc, "/packages", "/home", false)

	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			if err := fs.MkdirAll(ctx, dest+"/dot-vim", 0755); err != nil {
				return err
			}
			if err := fs.WriteFile(ctx, dest+"/dot-vim/dot-vimrc", []byte("set nocompatible"), 0644); err != nil {
				return err
			}
			if err := fs.MkdirAll(ctx, dest+"/dot-zsh", 0755); err != nil {
				return err
			}
			if err := fs.WriteFile(ctx, dest+"/dot-zsh/dot-zshrc", []byte("setopt no_beep"), 0644); err != nil {
				return err
			}
			return fs.WriteFile(ctx, dest+"/.dotbootstrap.yaml", []byte(machinesBootstrapYAML), 0644)
		},
	}

	svc := newCloneService(fs, logger, manageSvc, cloner, &mockPackageSelector{}, store, "/packages", "/home", false)
	svc.hostname = func() (string, error) { return hostname, nil }

	return svc
}

// TestCloneService_Clone_MachinesProfile covers the wiring from a machines
// match through package selection to the manifest, end to end.
func TestCloneService_Clone_MachinesProfile(t *testing.T) {
	tests := []struct {
		name            string
		hostname        string
		opts            CloneOptions
		wantInstalled   string
		wantUninstalled string
		wantProfile     string
	}{
		{
			name:            "machines entry selects the work profile",
			hostname:        "hephaestus.o11y.geicoinf.com",
			wantInstalled:   "dot-vim",
			wantUninstalled: "dot-zsh",
			wantProfile:     "work",
		},
		{
			name:            "explicit profile overrides the machines entry",
			hostname:        "hephaestus.o11y.geicoinf.com",
			opts:            CloneOptions{Profile: "personal"},
			wantInstalled:   "dot-zsh",
			wantUninstalled: "dot-vim",
			wantProfile:     "personal",
		},
		{
			name:            "unmatched host falls back to the default profile",
			hostname:        "unknown-box.example.org",
			wantInstalled:   "dot-zsh",
			wantUninstalled: "dot-vim",
			wantProfile:     "personal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fs := adapters.NewMemFS()
			require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

			store := manifest.NewFSManifestStoreWithDir(fs, "/state")
			svc := newMachinesCloneService(t, fs, store, tt.hostname)

			require.NoError(t, svc.Clone(ctx, "https://github.com/user/dotfiles", tt.opts))

			loaded := store.Load(ctx, NewTargetPath("/home").Unwrap())
			require.True(t, loaded.IsOk())
			m := loaded.Unwrap()

			_, installed := m.GetPackage(tt.wantInstalled)
			assert.True(t, installed, "expected %s to be installed", tt.wantInstalled)

			_, unwanted := m.GetPackage(tt.wantUninstalled)
			assert.False(t, unwanted, "expected %s to stay uninstalled", tt.wantUninstalled)

			repo, exists := m.GetRepository()
			require.True(t, exists, "repository info should be recorded")
			assert.Equal(t, tt.wantProfile, repo.Profile)
		})
	}
}
