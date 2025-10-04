package planner

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

// Task 7.2.1: Test Resolution Policy Types
func TestResolutionPolicyTypes(t *testing.T) {
	tests := []struct {
		name   string
		policy ResolutionPolicy
		want   string
	}{
		{"fail", PolicyFail, "fail"},
		{"backup", PolicyBackup, "backup"},
		{"overwrite", PolicyOverwrite, "overwrite"},
		{"skip", PolicySkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Task 7.2.2: Test Resolution Policies Configuration
func TestResolutionPoliciesConfiguration(t *testing.T) {
	policies := ResolutionPolicies{
		OnFileExists:    PolicyBackup,
		OnWrongLink:     PolicyOverwrite,
		OnPermissionErr: PolicyFail,
		OnCircular:      PolicyFail,
		OnTypeMismatch:  PolicyFail,
	}

	assert.Equal(t, PolicyBackup, policies.OnFileExists)
	assert.Equal(t, PolicyOverwrite, policies.OnWrongLink)
	assert.Equal(t, PolicyFail, policies.OnPermissionErr)
}

func TestDefaultPolicies(t *testing.T) {
	policies := DefaultPolicies()

	// All policies should default to fail for safety
	assert.Equal(t, PolicyFail, policies.OnFileExists)
	assert.Equal(t, PolicyFail, policies.OnWrongLink)
	assert.Equal(t, PolicyFail, policies.OnPermissionErr)
	assert.Equal(t, PolicyFail, policies.OnCircular)
	assert.Equal(t, PolicyFail, policies.OnTypeMismatch)
}

// Task 7.2.3: Test PolicyFail
func TestPolicyFail(t *testing.T) {
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(
		ConflictFileExists,
		targetPath,
		"File exists",
	)

	outcome := applyFailPolicy(conflict)

	assert.Equal(t, ResolveConflict, outcome.Status)
	assert.NotNil(t, outcome.Conflict)
	assert.Equal(t, conflict, *outcome.Conflict)
	assert.Empty(t, outcome.Operations)
}

// Task 7.2.4-5: Backup and Overwrite policies require additional operations
// These will be implemented in a future task
// For now, focusing on Fail and Skip policies

// Task 7.2.6: Test PolicySkip
func TestPolicySkip(t *testing.T) {
	sourcePath := dot.NewFilePath("/stow/bash/dot-bashrc").Unwrap()
	targetPath := dot.NewFilePath("/home/user/.bashrc").Unwrap()

	op := dot.NewLinkCreate(sourcePath, targetPath)

	conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

	outcome := applySkipPolicy(op, conflict)

	assert.Equal(t, ResolveSkip, outcome.Status)
	assert.Empty(t, outcome.Operations)
	assert.NotNil(t, outcome.Warning)
	assert.Contains(t, outcome.Warning.Message, "Skipping")
}

