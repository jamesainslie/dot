package planner

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

// Task 7.1.1: Test ConflictType enumeration
func TestConflictTypeString(t *testing.T) {
	tests := []struct {
		name string
		ct   ConflictType
		want string
	}{
		{"file exists", ConflictFileExists, "file_exists"},
		{"wrong link", ConflictWrongLink, "wrong_link"},
		{"permission", ConflictPermission, "permission"},
		{"circular", ConflictCircular, "circular"},
		{"dir expected", ConflictDirExpected, "dir_expected"},
		{"file expected", ConflictFileExpected, "file_expected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Task 7.1.2: Test Conflict value object
func TestConflictCreation(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists at target location",
	)

	assert.Equal(t, ConflictFileExists, conflict.Type)
	assert.Equal(t, targetPath, conflict.Path)
	assert.Equal(t, "File exists at target location", conflict.Details)
	assert.NotNil(t, conflict.Context)
	assert.Empty(t, conflict.Suggestions)
}

func TestConflictWithContext(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists",
	)

	conflict = conflict.WithContext("size", "1024")
	conflict = conflict.WithContext("mode", "0644")

	assert.Equal(t, "1024", conflict.Context["size"])
	assert.Equal(t, "0644", conflict.Context["mode"])
}

func TestConflictWithSuggestion(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists",
	)

	suggestion := Suggestion{
		Action:      "Use --backup flag",
		Explanation: "Preserves existing file",
	}

	conflict = conflict.WithSuggestion(suggestion)

	assert.Len(t, conflict.Suggestions, 1)
	assert.Equal(t, "Use --backup flag", conflict.Suggestions[0].Action)
}

// Task 7.1.3: Test Resolution Status Types
func TestResolutionStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status ResolutionStatus
		want   string
	}{
		{"ok", ResolveOK, "ok"},
		{"conflict", ResolveConflict, "conflict"},
		{"warning", ResolveWarning, "warning"},
		{"skip", ResolveSkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolutionOutcomeCreation(t *testing.T) {
	t.Run("ok status", func(t *testing.T) {
		outcome := ResolutionOutcome{
			Status:     ResolveOK,
			Operations: []dot.Operation{},
		}
		assert.Equal(t, ResolveOK, outcome.Status)
		assert.NotNil(t, outcome.Operations)
		assert.Nil(t, outcome.Conflict)
		assert.Nil(t, outcome.Warning)
	})

	t.Run("conflict status", func(t *testing.T) {
		targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

		outcome := ResolutionOutcome{
			Status:   ResolveConflict,
			Conflict: &conflict,
		}
		assert.Equal(t, ResolveConflict, outcome.Status)
		assert.NotNil(t, outcome.Conflict)
		assert.Equal(t, ConflictFileExists, outcome.Conflict.Type)
	})
}

// Task 7.1.4: Test ResolveResult Type
func TestResolveResultConstruction(t *testing.T) {
	t.Run("with operations", func(t *testing.T) {
		ops := []dot.Operation{}
		result := NewResolveResult(ops)
		assert.Len(t, result.Operations, 0)
		assert.Empty(t, result.Conflicts)
		assert.Empty(t, result.Warnings)
	})

	t.Run("with conflicts", func(t *testing.T) {
		targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

		result := NewResolveResult(nil)
		result = result.WithConflict(conflict)

		assert.Len(t, result.Conflicts, 1)
		assert.Equal(t, ConflictFileExists, result.Conflicts[0].Type)
	})

	t.Run("with warnings", func(t *testing.T) {
		warning := Warning{
			Message:  "File backed up",
			Severity: WarnInfo,
		}

		result := NewResolveResult(nil)
		result = result.WithWarning(warning)

		assert.Len(t, result.Warnings, 1)
		assert.Equal(t, "File backed up", result.Warnings[0].Message)
	})
}

func TestResolveResultQueries(t *testing.T) {
	t.Run("HasConflicts", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.False(t, result.HasConflicts())

		targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := NewConflict(ConflictFileExists, targetPath, "File exists")
		result = result.WithConflict(conflict)

		assert.True(t, result.HasConflicts())
	})

	t.Run("ConflictCount", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.Equal(t, 0, result.ConflictCount())

		targetPath1 := dot.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict1 := NewConflict(ConflictFileExists, targetPath1, "File exists")
		result = result.WithConflict(conflict1)

		targetPath2 := dot.NewFilePath("/home/user/.vimrc").Unwrap()
		conflict2 := NewConflict(ConflictWrongLink, targetPath2, "Wrong link")
		result = result.WithConflict(conflict2)

		assert.Equal(t, 2, result.ConflictCount())
	})

	t.Run("WarningCount", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.Equal(t, 0, result.WarningCount())

		warning := Warning{Message: "Test warning"}
		result = result.WithWarning(warning)

		assert.Equal(t, 1, result.WarningCount())
	})
}

// Task 7.1.5-7: Test Conflict Detection
func TestDetectFileExistsConflict(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()

	op := dot.NewLinkCreate(sourcePath, targetPath)

	current := CurrentState{
		Files: map[string]FileInfo{
			targetPath.String(): {Size: 100},
		},
		Links: make(map[string]LinkTarget),
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveConflict, outcome.Status)
	assert.NotNil(t, outcome.Conflict)
	assert.Equal(t, ConflictFileExists, outcome.Conflict.Type)
	assert.Contains(t, outcome.Conflict.Details, "File exists")
}

func TestDetectWrongLinkConflict(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
	wrongPath := dot.NewFilePath("/stow/other/dot-bashrc").Unwrap()

	op := dot.NewLinkCreate(sourcePath, targetPath)

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: map[string]LinkTarget{
			targetPath.String(): {Target: wrongPath.String()},
		},
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveConflict, outcome.Status)
	assert.NotNil(t, outcome.Conflict)
	assert.Equal(t, ConflictWrongLink, outcome.Conflict.Type)
}

func TestDetectNoConflict(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()

	op := dot.NewLinkCreate(sourcePath, targetPath)

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: make(map[string]LinkTarget),
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveOK, outcome.Status)
	assert.Nil(t, outcome.Conflict)
	assert.Len(t, outcome.Operations, 1)
}

func TestDetectLinkAlreadyCorrect(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()

	op := dot.NewLinkCreate(sourcePath, targetPath)

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: map[string]LinkTarget{
			targetPath.String(): {Target: sourcePath.String()},
		},
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveSkip, outcome.Status)
	assert.Nil(t, outcome.Conflict)
}

func TestDetectDirCreateConflicts(t *testing.T) {
	t.Run("file exists where directory expected", func(t *testing.T) {
		dirPath := dot.NewFilePath("/home/user/.config").Unwrap()

		op := dot.NewDirCreate(dirPath)

		current := CurrentState{
			Files: map[string]FileInfo{
				dirPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
		}

		outcome := detectDirCreateConflicts(op, current)

		assert.Equal(t, ResolveConflict, outcome.Status)
		assert.NotNil(t, outcome.Conflict)
		assert.Equal(t, ConflictFileExpected, outcome.Conflict.Type)
	})

	t.Run("directory already exists", func(t *testing.T) {
		dirPath := dot.NewFilePath("/home/user/.config").Unwrap()

		op := dot.NewDirCreate(dirPath)

		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: make(map[string]LinkTarget),
			Dirs:  map[string]bool{dirPath.String(): true},
		}

		outcome := detectDirCreateConflicts(op, current)

		assert.Equal(t, ResolveSkip, outcome.Status)
	})

	t.Run("no conflict", func(t *testing.T) {
		dirPath := dot.NewFilePath("/home/user/.config").Unwrap()

		op := dot.NewDirCreate(dirPath)

		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		outcome := detectDirCreateConflicts(op, current)

		assert.Equal(t, ResolveOK, outcome.Status)
	})
}

// Task 7.4.1: Test Main Resolve Function
func TestResolveFunction(t *testing.T) {
	t.Run("no conflicts", func(t *testing.T) {
		sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
		targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

		ops := []dot.Operation{
			dot.NewLinkCreate(sourcePath, targetPath),
		}

		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		policies := DefaultPolicies()

		result := Resolve(ops, current, policies, "/backup")

		assert.False(t, result.HasConflicts())
		assert.Len(t, result.Operations, 1)
		assert.Empty(t, result.Warnings)
	})

	t.Run("with conflict using fail policy", func(t *testing.T) {
		sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
		targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

		ops := []dot.Operation{
			dot.NewLinkCreate(sourcePath, targetPath),
		}

		current := CurrentState{
			Files: map[string]FileInfo{
				targetPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		policies := DefaultPolicies() // Defaults to PolicyFail

		result := Resolve(ops, current, policies, "/backup")

		assert.True(t, result.HasConflicts())
		assert.Len(t, result.Conflicts, 1)
		assert.Equal(t, ConflictFileExists, result.Conflicts[0].Type)

		// Should have suggestions
		assert.NotEmpty(t, result.Conflicts[0].Suggestions)
	})

	t.Run("with conflict using skip policy", func(t *testing.T) {
		sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
		targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

		ops := []dot.Operation{
			dot.NewLinkCreate(sourcePath, targetPath),
		}

		current := CurrentState{
			Files: map[string]FileInfo{
				targetPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		policies := DefaultPolicies()
		policies.OnFileExists = PolicySkip

		result := Resolve(ops, current, policies, "/backup")

		assert.False(t, result.HasConflicts())
		assert.Empty(t, result.Operations) // Operation was skipped
		assert.Len(t, result.Warnings, 1)
	})
}

// Task 7.4.2: Test Conflict Aggregation
func TestConflictAggregation(t *testing.T) {
	source1 := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
	target1 := dot.NewFilePath("/home/user/.bashrc").Unwrap()

	source2 := dot.NewFilePath("/stow/vim/dot-vimrc").Unwrap()
	target2 := dot.NewFilePath("/home/user/.vimrc").Unwrap()

	ops := []dot.Operation{
		dot.NewLinkCreate(source1, target1),
		dot.NewLinkCreate(source2, target2),
	}

	current := CurrentState{
		Files: map[string]FileInfo{
			target1.String(): {Size: 100},
			target2.String(): {Size: 200},
		},
		Links: make(map[string]LinkTarget),
		Dirs:  make(map[string]bool),
	}

	policies := DefaultPolicies()

	result := Resolve(ops, current, policies, "/backup")

	// Both operations should have conflicts
	assert.True(t, result.HasConflicts())
	assert.Equal(t, 2, result.ConflictCount())

	// Both conflicts should have suggestions
	for _, c := range result.Conflicts {
		assert.NotEmpty(t, c.Suggestions)
	}
}

func TestMixedOperations(t *testing.T) {
	sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	dirPath := dot.NewFilePath("/home/user/.config").Unwrap()

	ops := []dot.Operation{
		dot.NewLinkCreate(sourcePath, targetPath),
		dot.NewDirCreate(dirPath),
	}

	current := CurrentState{
		Files: map[string]FileInfo{
			targetPath.String(): {Size: 100}, // Conflict for link
		},
		Links: make(map[string]LinkTarget),
		Dirs:  make(map[string]bool), // No conflict for dir
	}

	policies := DefaultPolicies()
	policies.OnFileExists = PolicySkip

	result := Resolve(ops, current, policies, "/backup")

	// One operation skipped (link), one succeeded (dir)
	assert.False(t, result.HasConflicts())
	assert.Len(t, result.Operations, 1) // Only dir create
	assert.Len(t, result.Warnings, 1)   // Warning for skipped link
}
