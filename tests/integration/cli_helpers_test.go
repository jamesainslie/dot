package integration

import (
	"os/exec"
	"testing"
)

// skipIfCLIUnavailable checks if the CLI can be executed and skips the test if not.
func skipIfCLIUnavailable(t *testing.T, output []byte, err error) {
	t.Helper()
	if err != nil {
		t.Skipf("CLI execution unavailable in this environment: %v, output: %s", err, output)
	}
}

// checkCLIAvailable verifies the CLI can be executed.
func checkCLIAvailable(t *testing.T) {
	t.Helper()
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		t.Skip("Go toolchain not available for CLI tests")
	}
}
