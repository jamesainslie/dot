package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoldenTest_Path(t *testing.T) {
	goldenDir := t.TempDir()
	gt := NewGoldenTest(t, goldenDir, "test-case", "txt")

	expectedPath := filepath.Join(goldenDir, "test-case.txt")
	assert.Equal(t, expectedPath, gt.Path())
}

func TestGoldenTest_Exists(t *testing.T) {
	goldenDir := t.TempDir()
	gt := NewGoldenTest(t, goldenDir, "test-case", "txt")

	// Initially does not exist
	assert.False(t, gt.Exists())

	// Create file
	require.NoError(t, os.WriteFile(gt.Path(), []byte("content"), 0644))

	// Now exists
	assert.True(t, gt.Exists())
}

func TestGoldenTest_Update(t *testing.T) {
	goldenDir := t.TempDir()
	gt := NewGoldenTest(t, goldenDir, "test-case", "txt")

	content := "golden content"
	gt.Update(content)

	// Verify file created
	assert.FileExists(t, gt.Path())

	// Verify content
	actual, err := os.ReadFile(gt.Path())
	require.NoError(t, err)
	assert.Equal(t, content, string(actual))
}

func TestGoldenTestSuite_Test(t *testing.T) {
	goldenDir := t.TempDir()
	suite := NewGoldenTestSuite(t, goldenDir)

	gt := suite.Test("test-case", "json")
	assert.Contains(t, gt.Path(), "test-case.json")
}

func TestGoldenTestSuite_TextTest(t *testing.T) {
	goldenDir := t.TempDir()
	suite := NewGoldenTestSuite(t, goldenDir)

	gt := suite.TextTest("test-case")
	assert.Contains(t, gt.Path(), ".txt")
}

func TestGoldenTestSuite_JSONTest(t *testing.T) {
	goldenDir := t.TempDir()
	suite := NewGoldenTestSuite(t, goldenDir)

	gt := suite.JSONTest("test-case")
	assert.Contains(t, gt.Path(), ".json")
}

func TestGoldenTestSuite_YAMLTest(t *testing.T) {
	goldenDir := t.TempDir()
	suite := NewGoldenTestSuite(t, goldenDir)

	gt := suite.YAMLTest("test-case")
	assert.Contains(t, gt.Path(), ".yaml")
}
