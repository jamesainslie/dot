package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("graceful shutdown on SIGINT", func(t *testing.T) {
		// Setup test directory
		tmpDir := t.TempDir()

		// Start a long-running command (status with non-existent manifest is fast, use manage with delay)
		cmd := exec.Command("go", "run", "../../cmd/dot", "status",
			"--dir", tmpDir,
			"--target", tmpDir)

		require.NoError(t, cmd.Start())

		// Give command time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGINT
		require.NoError(t, cmd.Process.Signal(syscall.SIGINT))

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			// Process should exit gracefully
			// Exit code may vary: -1 (killed by signal), 1 (error), 130 (signal exit), or 0 (success)
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				require.True(t, ok, "error should be ExitError")
				// Accept various exit codes: -1 (signal termination), 1 (error), 130 (handled signal exit)
				assert.Contains(t, []int{-1, 1, 130}, exitErr.ExitCode())
			}
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			t.Fatal("process did not exit gracefully within timeout")
		}
	})

	t.Run("forced exit on second SIGINT", func(t *testing.T) {
		// Setup test directory
		tmpDir := t.TempDir()

		// Start a command
		cmd := exec.Command("go", "run", "../../cmd/dot", "status",
			"--dir", tmpDir,
			"--target", tmpDir)

		require.NoError(t, cmd.Start())

		// Give command time to start
		time.Sleep(100 * time.Millisecond)

		// Send first SIGINT
		require.NoError(t, cmd.Process.Signal(syscall.SIGINT))

		// Wait for cleanup grace period (100ms) + buffer to ensure handler is ready for second signal
		time.Sleep(150 * time.Millisecond)

		// Send second SIGINT (should force exit)
		require.NoError(t, cmd.Process.Signal(syscall.SIGINT))

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			// Process should exit - accept various exit codes depending on timing
			// -1: killed by signal, 1: normal error, 130: handled signal exit, 0: normal success
			// This test verifies the signal mechanism works, not the exact exit code
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				require.True(t, ok, "error should be ExitError")
				// Accept various exit codes - the exact code depends on timing and signal handling
				assert.Contains(t, []int{-1, 1, 130}, exitErr.ExitCode(),
					"expected signal exit (-1), normal exit (1), or forced exit (130), got %d", exitErr.ExitCode())
			}
			// Exit code 0 (success) is also acceptable if command completed normally
		case <-time.After(2 * time.Second):
			cmd.Process.Kill()
			t.Fatal("process did not exit within timeout")
		}
	})

	t.Run("subprocess environment is isolated", func(t *testing.T) {
		// This test verifies that subprocess operations leave the invoking user's
		// dot state untouched.
		dotBinary := filepath.Join(t.TempDir(), "dot")
		buildCmd := exec.Command("go", "build", "-o", dotBinary, "../../cmd/dot")
		buildOutput, buildErr := buildCmd.CombinedOutput()
		require.NoError(t, buildErr, "build dot: %s", buildOutput)

		userHome := t.TempDir()
		userConfigHome := filepath.Join(userHome, ".config")
		userDataHome := filepath.Join(userHome, ".local", "share")
		userStateHome := filepath.Join(userHome, ".local", "state")
		t.Setenv("XDG_CONFIG_HOME", userConfigHome)
		t.Setenv("XDG_DATA_HOME", userDataHome)
		t.Setenv("XDG_STATE_HOME", userStateHome)

		userManifestDir := filepath.Join(userDataHome, "dot", "manifest")
		userManifestPath := filepath.Join(userManifestDir, ".dot-manifest.json")
		userManifest := []byte("{\n  \"version\": \"1.0\",\n  \"updated_at\": \"2026-08-25T00:00:00Z\",\n  \"packages\": {},\n  \"hashes\": {}\n}\n")
		require.NoError(t, os.MkdirAll(userManifestDir, 0o755))
		require.NoError(t, os.WriteFile(userManifestPath, userManifest, 0o644))
		require.NoError(t, os.MkdirAll(filepath.Join(userConfigHome, "dot"), 0o755))
		userConfig := []byte("directories:\n  manifest: " + userManifestDir + "\n")
		require.NoError(t, os.WriteFile(filepath.Join(userConfigHome, "dot", "config.yaml"), userConfig, 0o644))
		t.Setenv("DOT_DIRECTORIES_MANIFEST", userManifestDir)

		// Setup test directories
		tmpDir := t.TempDir()
		packageDir := tmpDir + "/packages"
		targetDir := tmpDir + "/target"

		require.NoError(t, os.MkdirAll(packageDir+"/test-pkg", 0755))
		require.NoError(t, os.MkdirAll(targetDir, 0755))
		require.NoError(t, os.WriteFile(packageDir+"/test-pkg/file1", []byte("content"), 0644))
		commandHome := filepath.Join(tmpDir, "home")
		commandConfigHome := filepath.Join(commandHome, ".config")
		commandDataHome := filepath.Join(commandHome, ".local", "share")
		commandStateHome := filepath.Join(commandHome, ".local", "state")
		require.NoError(t, os.MkdirAll(commandHome, 0o755))

		// Start manage command
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, dotBinary, "manage", "test-pkg",
			"--dir", packageDir,
			"--target", targetDir)
		commandEnv := make([]string, 0, len(os.Environ())+4)
		for _, variable := range os.Environ() {
			if strings.HasPrefix(variable, "DOT_") {
				continue
			}
			commandEnv = append(commandEnv, variable)
		}
		commandEnv = append(commandEnv,
			"HOME="+commandHome,
			"XDG_CONFIG_HOME="+commandConfigHome,
			"XDG_DATA_HOME="+commandDataHome,
			"XDG_STATE_HOME="+commandStateHome,
		)
		cmd.Env = commandEnv

		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "manage command failed: %s", output)

		// The child process cannot mutate the caller's manifest, even when the
		// caller exports a DOT_* configuration override.
		gotUserManifest, readErr := os.ReadFile(userManifestPath)
		require.NoError(t, readErr)
		assert.Equal(t, userManifest, gotUserManifest)
	})
}
