package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update-golden", false, "update golden test files")

// GoldenTest provides golden file testing capabilities.
type GoldenTest struct {
	t         *testing.T
	goldenDir string
	testName  string
	extension string
}

// NewGoldenTest creates a new golden test.
func NewGoldenTest(t *testing.T, goldenDir, testName, extension string) *GoldenTest {
	t.Helper()
	return &GoldenTest{
		t:         t,
		goldenDir: goldenDir,
		testName:  testName,
		extension: extension,
	}
}

// Path returns the path to the golden file.
func (gt *GoldenTest) Path() string {
	filename := gt.testName + "." + gt.extension
	return filepath.Join(gt.goldenDir, filename)
}

// AssertMatch compares actual content with golden file.
func (gt *GoldenTest) AssertMatch(actual string) {
	gt.t.Helper()

	goldenPath := gt.Path()

	if *updateGolden {
		gt.t.Logf("updating golden file: %s", goldenPath)
		err := os.MkdirAll(gt.goldenDir, 0755)
		require.NoError(gt.t, err)
		err = os.WriteFile(goldenPath, []byte(actual), 0644) //nolint:gosec // Golden test files
		require.NoError(gt.t, err)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(gt.t, err, "failed to read golden file: %s", goldenPath)

	assert.Equal(gt.t, string(expected), actual, "content does not match golden file")
}

// AssertMatchBytes compares actual bytes with golden file.
func (gt *GoldenTest) AssertMatchBytes(actual []byte) {
	gt.t.Helper()
	gt.AssertMatch(string(actual))
}

// Update forces update of the golden file with given content.
func (gt *GoldenTest) Update(content string) {
	gt.t.Helper()

	goldenPath := gt.Path()
	gt.t.Logf("updating golden file: %s", goldenPath)
	err := os.MkdirAll(gt.goldenDir, 0755)
	require.NoError(gt.t, err)
	err = os.WriteFile(goldenPath, []byte(content), 0644) //nolint:gosec // Golden test files
	require.NoError(gt.t, err)
}

// Exists checks if the golden file exists.
func (gt *GoldenTest) Exists() bool {
	_, err := os.Stat(gt.Path())
	return err == nil
}

// GoldenTestSuite manages multiple golden tests for a test suite.
type GoldenTestSuite struct {
	t         *testing.T
	goldenDir string
}

// NewGoldenTestSuite creates a new golden test suite.
func NewGoldenTestSuite(t *testing.T, goldenDir string) *GoldenTestSuite {
	t.Helper()
	return &GoldenTestSuite{
		t:         t,
		goldenDir: goldenDir,
	}
}

// Test creates a golden test for a specific test case.
func (gts *GoldenTestSuite) Test(testName, extension string) *GoldenTest {
	gts.t.Helper()
	return NewGoldenTest(gts.t, gts.goldenDir, testName, extension)
}

// TextTest creates a golden test for text output.
func (gts *GoldenTestSuite) TextTest(testName string) *GoldenTest {
	gts.t.Helper()
	return gts.Test(testName, "txt")
}

// JSONTest creates a golden test for JSON output.
func (gts *GoldenTestSuite) JSONTest(testName string) *GoldenTest {
	gts.t.Helper()
	return gts.Test(testName, "json")
}

// YAMLTest creates a golden test for YAML output.
func (gts *GoldenTestSuite) YAMLTest(testName string) *GoldenTest {
	gts.t.Helper()
	return gts.Test(testName, "yaml")
}
