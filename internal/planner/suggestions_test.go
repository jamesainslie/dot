package planner

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

// Task 7.3.3: Test Suggestion Generation
func TestGenerateSuggestionsForFileExists(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)
	assert.GreaterOrEqual(t, len(suggestions), 2)

	// Should suggest backup
	hasBackup := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "backup") {
			hasBackup = true
			assert.NotEmpty(t, s.Explanation)
		}
	}
	assert.True(t, hasBackup, "Should suggest backup option")

	// Should suggest adopt
	hasAdopt := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "adopt") {
			hasAdopt = true
			assert.NotEmpty(t, s.Explanation)
		}
	}
	assert.True(t, hasAdopt, "Should suggest adopt option")
}

func TestGenerateSuggestionsForWrongLink(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictWrongLink, targetPath, "Symlink points elsewhere")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should suggest unstowing other package
	hasUnstow := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "unstow") {
			hasUnstow = true
			assert.NotEmpty(t, s.Explanation)
		}
	}
	assert.True(t, hasUnstow, "Should suggest unstow option")
}

func TestGenerateSuggestionsForPermission(t *testing.T) {
	targetPath := dot.NewFilePath("/etc/config").Unwrap()
	conflict := NewConflict(ConflictPermission, targetPath, "Permission denied")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should mention checking permissions
	hasPermCheck := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "permission") || containsIgnoreCase(s.Action, "access") {
			hasPermCheck = true
		}
	}
	assert.True(t, hasPermCheck, "Should suggest checking permissions")
}

func TestGenerateSuggestionsForCircular(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.config").Unwrap()
	conflict := NewConflict(ConflictCircular, targetPath, "Circular dependency")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should have actionable suggestions
	for _, s := range suggestions {
		assert.NotEmpty(t, s.Action)
		assert.NotEmpty(t, s.Explanation)
	}
}

func TestGenerateSuggestionsForTypeMismatch(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.config").Unwrap()
	conflict := NewConflict(ConflictFileExpected, targetPath, "File exists where directory expected")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)
}

// Task 7.3.4: Test Conflict Enrichment
func TestEnrichConflictWithSuggestions(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

	// Initially no suggestions
	assert.Empty(t, conflict.Suggestions)

	// Enrich with suggestions
	enriched := enrichConflictWithSuggestions(conflict)

	// Should now have suggestions
	assert.NotEmpty(t, enriched.Suggestions)
	assert.GreaterOrEqual(t, len(enriched.Suggestions), 2)

	// All suggestions should have required fields
	for _, s := range enriched.Suggestions {
		assert.NotEmpty(t, s.Action, "Suggestion should have action")
		assert.NotEmpty(t, s.Explanation, "Suggestion should have explanation")
	}
}

func TestEnrichMultipleConflicts(t *testing.T) {
	path1 := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict1 := NewConflict(ConflictFileExists, path1, "File exists")

	path2 := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	conflict2 := NewConflict(ConflictWrongLink, path2, "Wrong link")

	enriched1 := enrichConflictWithSuggestions(conflict1)
	enriched2 := enrichConflictWithSuggestions(conflict2)

	// Both should have suggestions
	assert.NotEmpty(t, enriched1.Suggestions)
	assert.NotEmpty(t, enriched2.Suggestions)

	// Suggestions should be different for different conflict types
	assert.NotEqual(t, enriched1.Suggestions, enriched2.Suggestions)
}

// Helper function
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	// Simple ASCII lowercase
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
