package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildConfig_ValidPaths(t *testing.T) {
	// Set global config
	globalCfg = globalConfig{
		stowDir:   ".",
		targetDir: ".",
		dryRun:    false,
		verbose:   0,
		quiet:     false,
		logJSON:   false,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Paths should be absolute
	require.True(t, filepath.IsAbs(cfg.StowDir))
	require.True(t, filepath.IsAbs(cfg.TargetDir))
	require.False(t, cfg.DryRun)
	require.Equal(t, 0, cfg.Verbosity)
	require.NotNil(t, cfg.FS)
	require.NotNil(t, cfg.Logger)
}

func TestBuildConfig_DryRunEnabled(t *testing.T) {
	globalCfg = globalConfig{
		stowDir:   ".",
		targetDir: ".",
		dryRun:    true,
		verbose:   0,
		quiet:     false,
		logJSON:   false,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)
	require.True(t, cfg.DryRun)
}

func TestBuildConfig_VerbositySet(t *testing.T) {
	globalCfg = globalConfig{
		stowDir:   ".",
		targetDir: ".",
		dryRun:    false,
		verbose:   2,
		quiet:     false,
		logJSON:   false,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)
	require.Equal(t, 2, cfg.Verbosity)
}

func TestBuildConfig_MultipleCombinations(t *testing.T) {
	tests := []struct {
		name    string
		dryRun  bool
		verbose int
		quiet   bool
		logJSON bool
	}{
		{"default", false, 0, false, false},
		{"dry-run", true, 0, false, false},
		{"verbose", false, 1, false, false},
		{"quiet", false, 0, true, false},
		{"json", false, 0, false, true},
		{"dry-run verbose", true, 2, false, false},
		{"verbose json", false, 3, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalCfg = globalConfig{
				stowDir:   ".",
				targetDir: ".",
				dryRun:    tt.dryRun,
				verbose:   tt.verbose,
				quiet:     tt.quiet,
				logJSON:   tt.logJSON,
			}

			cfg, err := buildConfig()
			require.NoError(t, err)
			require.Equal(t, tt.dryRun, cfg.DryRun)
			require.Equal(t, tt.verbose, cfg.Verbosity)
			require.NotNil(t, cfg.FS)
			require.NotNil(t, cfg.Logger)
		})
	}
}

func TestCreateLogger_Quiet(t *testing.T) {
	globalCfg = globalConfig{
		quiet: true,
	}

	logger := createLogger()
	require.NotNil(t, logger)
}

func TestCreateLogger_JSONFormat(t *testing.T) {
	globalCfg = globalConfig{
		quiet:   false,
		logJSON: true,
		verbose: 0,
	}

	logger := createLogger()
	require.NotNil(t, logger)
}

func TestCreateLogger_TextFormat(t *testing.T) {
	globalCfg = globalConfig{
		quiet:   false,
		logJSON: false,
		verbose: 0,
	}

	logger := createLogger()
	require.NotNil(t, logger)
}

func TestCreateLogger_VerboseLevels(t *testing.T) {
	tests := []struct {
		name    string
		verbose int
	}{
		{"no verbosity", 0},
		{"level 1", 1},
		{"level 2", 2},
		{"level 3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalCfg = globalConfig{
				quiet:   false,
				logJSON: false,
				verbose: tt.verbose,
			}

			logger := createLogger()
			require.NotNil(t, logger)
		})
	}
}

func TestVerbosityToLevel(t *testing.T) {
	tests := []struct {
		name      string
		verbose   int
		wantInfo  bool
		wantDebug bool
	}{
		{
			name:      "level 0 is Info",
			verbose:   0,
			wantInfo:  true,
			wantDebug: false,
		},
		{
			name:      "level 1 is Debug",
			verbose:   1,
			wantInfo:  true,
			wantDebug: true,
		},
		{
			name:      "level 2 is more verbose",
			verbose:   2,
			wantInfo:  true,
			wantDebug: true,
		},
		{
			name:      "level 3 is even more verbose",
			verbose:   3,
			wantInfo:  true,
			wantDebug: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := verbosityToLevel(tt.verbose)
			require.NotNil(t, level)
		})
	}
}

func TestBuildConfig_AbsolutePaths(t *testing.T) {
	tmpDir := t.TempDir()
	stowDir := filepath.Join(tmpDir, "stow")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(stowDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	globalCfg = globalConfig{
		stowDir:   stowDir,
		targetDir: targetDir,
		dryRun:    false,
		verbose:   0,
		quiet:     false,
		logJSON:   false,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)
	require.Equal(t, stowDir, cfg.StowDir)
	require.Equal(t, targetDir, cfg.TargetDir)
}

func TestBuildConfig_RelativePaths(t *testing.T) {
	// Start with relative paths
	globalCfg = globalConfig{
		stowDir:   "./test/stow",
		targetDir: "./test/target",
		dryRun:    false,
		verbose:   0,
		quiet:     false,
		logJSON:   false,
	}

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should be converted to absolute
	require.True(t, filepath.IsAbs(cfg.StowDir))
	require.True(t, filepath.IsAbs(cfg.TargetDir))
	require.Contains(t, cfg.StowDir, "test/stow")
	require.Contains(t, cfg.TargetDir, "test/target")
}
