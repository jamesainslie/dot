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
)

// duplicateTargetSandbox builds a self-contained package/target/manifest tree
// where two packages claim the same target path. Nothing outside the sandbox is
// read or written: the package directory carries its own config, and every
// directory is passed explicitly.
func duplicateTargetSandbox(t *testing.T) (packageDir, targetDir string) {
	t.Helper()

	tmpDir := t.TempDir()
	packageDir = filepath.Join(tmpDir, "packages")
	targetDir = filepath.Join(tmpDir, "target")
	manifestDir := filepath.Join(tmpDir, "manifest")

	for _, dir := range []string{packageDir, targetDir, manifestDir} {
		require.NoError(t, os.MkdirAll(dir, 0o755))
	}

	// Two packages, same relative file: with the stow-style full-tree layout
	// both resolve to <target>/.vimrc.
	for _, pkg := range []string{"base", "overlay"} {
		pkgDir := filepath.Join(packageDir, pkg)
		require.NoError(t, os.MkdirAll(pkgDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "dot-vimrc"), []byte(pkg), 0o644))
	}

	// Repository-local config keeps the run away from the real home directory
	// and disables package name mapping so the two packages overlap.
	configDir := filepath.Join(packageDir, ".config", "dot")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	config := fmt.Sprintf(`directories:
  package: %s
  target: %s
  manifest: %s
dotfile:
  translate: true
  package_name_mapping: false
`, packageDir, targetDir, manifestDir)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(config), 0o644))

	return packageDir, targetDir
}

// TestDuplicateTargetAcrossPackages_CLISurface checks the message a user sees
// when two packages fight over the same target path.
//
// clone routes through ManageService.Manage with the full package list, so it
// hits the same guard as manage; remanage plans each package on its own and is
// guarded separately in pkg/dot.
func TestDuplicateTargetAcrossPackages_CLISurface(t *testing.T) {
	packageDir, targetDir := duplicateTargetSandbox(t)

	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
	})

	var out bytes.Buffer
	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"base", "overlay"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.Error(t, err, "overlapping packages must not be planned silently")

	for _, want := range []string{filepath.Join(targetDir, ".vimrc"), "base", "overlay"} {
		assert.Contains(t, err.Error(), want)
		assert.Contains(t, out.String(), want)
	}

	// The printed message must tell the user what to do about it, not just
	// state that the collision happened.
	assert.Contains(t, out.String(), "Remove the file from one package")
	assert.Contains(t, out.String(), "package_name_mapping")

	// Planning failed, so nothing was linked.
	_, statErr := os.Lstat(filepath.Join(targetDir, ".vimrc"))
	assert.True(t, os.IsNotExist(statErr), "no link should be created when planning fails")
}

// TestDuplicateTargetAcrossPackages_RemanageCLISurface covers the remanage
// path, which plans each package separately.
func TestDuplicateTargetAcrossPackages_RemanageCLISurface(t *testing.T) {
	packageDir, targetDir := duplicateTargetSandbox(t)

	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
	})

	var out bytes.Buffer
	cmd := newRemanageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"base", "overlay"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	require.Error(t, err, "overlapping packages must not be remanaged silently")

	for _, want := range []string{filepath.Join(targetDir, ".vimrc"), "base", "overlay"} {
		assert.Contains(t, err.Error(), want)
	}
	assert.Contains(t, out.String(), "Remove the file from one package")
}

func TestDuplicateTargetAcrossPackages_DryRunAlsoFails(t *testing.T) {
	packageDir, targetDir := duplicateTargetSandbox(t)

	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     true,
	})

	var stderr bytes.Buffer
	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"base", "overlay"})
	cmd.SetOut(&stderr)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	require.Error(t, err, "dry run must report the collision rather than a plan")
	assert.Contains(t, err.Error(), "base")
	assert.Contains(t, err.Error(), "overlay")
}

func TestSinglePackageStillManages(t *testing.T) {
	packageDir, targetDir := duplicateTargetSandbox(t)

	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
	})

	cmd := newManageCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"base"})

	require.NoError(t, cmd.Execute())
	assert.FileExists(t, filepath.Join(targetDir, ".vimrc"))
}
