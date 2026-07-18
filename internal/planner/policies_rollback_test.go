package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

// TestApplyBackupPolicy_FileDeleteRecordsBackupPath verifies the backup triple
// wires the backup path into the FileDelete operation so its rollback can
// restore the original file.
func TestApplyBackupPolicy_FileDeleteRecordsBackupPath(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	conflictPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()

	op := domain.NewLinkCreate("link1", sourcePath, targetPath)
	conflict := NewConflict(ConflictFileExists, conflictPath, "File exists")

	outcome := applyBackupPolicy(op, conflict, "/home/user/.dot-backup")

	require.Equal(t, ResolveOK, outcome.Status)
	require.Len(t, outcome.Operations, 3)

	backupOp, ok := outcome.Operations[0].(domain.FileBackup)
	require.True(t, ok, "first operation must be the FileBackup")

	deleteOp, ok := outcome.Operations[1].(domain.FileDelete)
	require.True(t, ok, "second operation must be the FileDelete")

	assert.Equal(t, backupOp.Backup.String(), deleteOp.Backup.String(),
		"FileDelete must record the same backup path as FileBackup so rollback can restore")
	assert.NotEmpty(t, deleteOp.Backup.String())
}
