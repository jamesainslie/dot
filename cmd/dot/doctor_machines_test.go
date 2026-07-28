package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/pkg/dot"
)

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
`

// machineDriftFixture builds a sandbox config with an optional bootstrap file
// and an optional recorded profile in the manifest.
func machineDriftFixture(t *testing.T, bootstrapYAML, recordedProfile string) dot.Config {
	t.Helper()

	ctx := context.Background()
	fs := adapters.NewMemFS()

	if bootstrapYAML != "" {
		require.NoError(t, fs.MkdirAll(ctx, "/packages", 0755))
		require.NoError(t, fs.WriteFile(ctx, "/packages/.dotbootstrap.yaml", []byte(bootstrapYAML), 0644))
	}

	store := manifest.NewFSManifestStoreWithDir(fs, "/state")
	targetPath := dot.NewTargetPath("/home")
	require.True(t, targetPath.IsOk())

	m := manifest.New()
	if recordedProfile != "" {
		m.SetRepository(manifest.RepositoryInfo{
			URL:     "https://github.com/user/dotfiles",
			Branch:  "main",
			Profile: recordedProfile,
		})
	}
	require.NoError(t, store.Save(ctx, targetPath.Unwrap(), m))

	return dot.Config{
		PackageDir:  "/packages",
		TargetDir:   "/home",
		ManifestDir: "/state",
		FS:          fs,
		Logger:      adapters.NewNoopLogger(),
	}
}

// machineDriftClient builds the client the drift check reads the manifest through.
func machineDriftClient(t *testing.T, cfg dot.Config) *dot.Client {
	t.Helper()

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	return client
}

func TestMachineProfileDrift(t *testing.T) {
	tests := []struct {
		name            string
		hostname        string
		bootstrapYAML   string
		recordedProfile string
		wantContains    []string
		wantEmpty       bool
	}{
		{
			name:            "profile differs from recorded",
			hostname:        "zeus.local",
			bootstrapYAML:   machinesBootstrapYAML,
			recordedProfile: "work",
			wantContains:    []string{"zeus.local", "zeus*", "personal", "work"},
		},
		{
			name:            "profile matches recorded",
			hostname:        "zeus.local",
			bootstrapYAML:   machinesBootstrapYAML,
			recordedProfile: "personal",
			wantEmpty:       true,
		},
		{
			name:            "no machines entry matches host",
			hostname:        "unknown-box",
			bootstrapYAML:   machinesBootstrapYAML,
			recordedProfile: "work",
			wantEmpty:       true,
		},
		{
			name:            "no recorded profile",
			hostname:        "zeus.local",
			bootstrapYAML:   machinesBootstrapYAML,
			recordedProfile: "",
			wantEmpty:       true,
		},
		{
			name:            "no bootstrap config",
			hostname:        "zeus.local",
			bootstrapYAML:   "",
			recordedProfile: "work",
			wantEmpty:       true,
		},
		{
			name:            "bootstrap without machines section",
			hostname:        "zeus.local",
			bootstrapYAML:   "version: \"1.0\"\npackages:\n  - name: dot-vim\n",
			recordedProfile: "work",
			wantEmpty:       true,
		},
		{
			name:            "unparsable bootstrap config is ignored",
			hostname:        "zeus.local",
			bootstrapYAML:   "version: [unclosed\n",
			recordedProfile: "work",
			wantEmpty:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := machineDriftFixture(t, tt.bootstrapYAML, tt.recordedProfile)
			client := machineDriftClient(t, cfg)
			message := machineProfileDrift(context.Background(), cfg, client, tt.hostname)

			if tt.wantEmpty {
				assert.Empty(t, message)
				return
			}

			for _, want := range tt.wantContains {
				assert.Contains(t, message, want)
			}
		})
	}
}

func TestMachineProfileDrift_NoHostname(t *testing.T) {
	cfg := machineDriftFixture(t, machinesBootstrapYAML, "work")
	client := machineDriftClient(t, cfg)

	assert.Empty(t, machineProfileDrift(context.Background(), cfg, client, ""))
}

func TestMachineProfileDrift_NoClient(t *testing.T) {
	cfg := machineDriftFixture(t, machinesBootstrapYAML, "work")

	assert.Empty(t, machineProfileDrift(context.Background(), cfg, nil, "zeus.local"))
}

// catchAllMachinesYAML maps every host to the work profile, so the drift check
// fires whatever the machine running the test is called.
const catchAllMachinesYAML = `version: "1.0"
packages:
  - name: vim
profiles:
  work:
    description: Work
    packages:
      - vim
  personal:
    description: Personal
    packages:
      - vim
machines:
  - host: "*"
    profile: work
`

// runDoctorForDrift runs the real doctor command against sandbox directories
// with the given recorded profile, returning stderr and the exit code.
func runDoctorForDrift(t *testing.T, recordedProfile string) (string, int) {
	t.Helper()

	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")
	stateDir := filepath.Join(tmpDir, "state")
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "vim"), 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	require.NoError(t, os.MkdirAll(stateDir, 0755))

	require.NoError(t, os.WriteFile(filepath.Join(packageDir, ".dotbootstrap.yaml"), []byte(catchAllMachinesYAML), 0644))

	// Keep the manifest inside the sandbox rather than the real home directory.
	repoConfig := fmt.Sprintf("directories:\n  manifest: %s\n", stateDir)
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, ".config", "dot"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(packageDir, ".config", "dot", "config.yaml"), []byte(repoConfig), 0644))

	ctx := context.Background()
	store := manifest.NewFSManifestStoreWithDir(dot.NewOSFilesystem(), stateDir)
	targetPath := dot.NewTargetPath(targetDir)
	require.True(t, targetPath.IsOk())

	m := manifest.New()
	m.SetRepository(manifest.RepositoryInfo{
		URL:     "https://github.com/user/dotfiles",
		Branch:  "main",
		Profile: recordedProfile,
	})
	require.NoError(t, store.Save(ctx, targetPath.Unwrap(), m))

	holder := &DoctorResultHolder{}
	rootCmd := NewRootCommand("test", "none", "unknown")
	rootCmd.SetContext(WithDoctorResultHolder(ctx, holder))

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{
		"--dir", packageDir,
		"--target", targetDir,
		"doctor", "--format", "json",
	})

	require.NoError(t, rootCmd.Execute())
	require.True(t, holder.Executed, "doctor should have reported a status")

	return stderr.String(), DoctorExitCode(holder.Status)
}

func TestDoctorCommand_PrintsMachineProfileDrift(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	driftOutput, driftExit := runDoctorForDrift(t, "personal")
	assert.Contains(t, driftOutput, "Note: machines entry")
	assert.Contains(t, driftOutput, "profile \"work\"")

	matchOutput, matchExit := runDoctorForDrift(t, "work")
	assert.NotContains(t, matchOutput, "Note: machines entry")

	assert.Equal(t, matchExit, driftExit, "the advisory note must not change the exit code")
}
