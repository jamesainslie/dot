// Package testutil provides utilities for integration testing.
//
// This package includes:
//   - FixtureBuilder: Create test packages and directory structures
//   - TestEnvironment: Isolated test execution environment
//   - Assertions: Specialized assertions for symlinks, files, and directories
//   - StateSnapshot: Capture and compare filesystem states
//   - GoldenTest: Compare outputs against golden files
//
// Example usage:
//
//	func TestManageWorkflow(t *testing.T) {
//	    env := testutil.NewTestEnvironment(t)
//
//	    // Create test package
//	    env.FixtureBuilder().Package("vim").
//	        WithFile("dot-vimrc", "set nocompatible").
//	        Create()
//
//	    // Capture state before operation
//	    before := testutil.CaptureState(t, env.TargetDir)
//
//	    // Perform operation
//	    // ...
//
//	    // Capture state after operation
//	    after := testutil.CaptureState(t, env.TargetDir)
//
//	    // Verify changes
//	    testutil.AssertLink(t, filepath.Join(env.TargetDir, ".vimrc"), "...")
//	}
package testutil
