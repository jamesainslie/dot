package planner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

// backupPathFor applies the backup policy to a single conflict and returns the
// backup destination path the policy chose.
func backupPathFor(t *testing.T, sourcePath, targetPath, backupDir string) string {
	t.Helper()

	source := domain.MustParsePath(sourcePath)
	target := domain.MustParseTargetPath(targetPath)
	op := domain.NewLinkCreate(domain.OperationID("link-"+targetPath), source, target)

	conflict := NewConflict(ConflictFileExists, domain.MustParsePath(targetPath), "File exists")

	outcome := applyBackupPolicy(op, conflict, backupDir)
	require.Equal(t, ResolveOK, outcome.Status)
	require.Len(t, outcome.Operations, 3)

	backupOp, ok := outcome.Operations[0].(domain.FileBackup)
	require.True(t, ok, "first operation must be the FileBackup")

	return backupOp.Backup.String()
}

// TestApplyBackupPolicy_SameBasenameDifferentDirs_DistinctBackupPaths verifies
// that two conflicting files sharing a basename but living in different
// directories never receive the same backup filename, even when backed up
// within the same timestamp granularity.
func TestApplyBackupPolicy_SameBasenameDifferentDirs_DistinctBackupPaths(t *testing.T) {
	backupDir := "/backup"

	first := backupPathFor(t,
		"/packages/git/config", "/home/user/.config/git/config", backupDir)
	second := backupPathFor(t,
		"/packages/hg/config", "/home/user/.config/hg/config", backupDir)

	assert.NotEqual(t, first, second,
		"backup paths for same-basename files in different directories must not collide")
}

// TestApplyBackupPolicy_SameBasenameDifferentDirs_BothBackupsSurvive executes
// the backup operations for two same-basename conflicts and verifies neither
// backup overwrites the other.
func TestApplyBackupPolicy_SameBasenameDifferentDirs_BothBackupsSurvive(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, fs.MkdirAll(ctx, "/home/user/a", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/user/b", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/backup", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/home/user/a/config", []byte("content-a"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/home/user/b/config", []byte("content-b"), 0644))

	conflicts := []struct {
		targetPath string
		content    []byte
	}{
		{"/home/user/a/config", []byte("content-a")},
		{"/home/user/b/config", []byte("content-b")},
	}

	backupPaths := make([]string, 0, len(conflicts))
	for _, c := range conflicts {
		source := domain.MustParsePath("/packages/pkg" + c.targetPath)
		target := domain.MustParseTargetPath(c.targetPath)
		op := domain.NewLinkCreate(domain.OperationID("link-"+c.targetPath), source, target)

		conflict := NewConflict(ConflictFileExists, domain.MustParsePath(c.targetPath), "File exists")

		outcome := applyBackupPolicy(op, conflict, "/backup")
		require.Equal(t, ResolveOK, outcome.Status)

		backupOp, ok := outcome.Operations[0].(domain.FileBackup)
		require.True(t, ok, "first operation must be the FileBackup")

		require.NoError(t, backupOp.Execute(ctx, fs))
		backupPaths = append(backupPaths, backupOp.Backup.String())
	}

	for i, c := range conflicts {
		data, err := fs.ReadFile(ctx, backupPaths[i])
		require.NoError(t, err, "backup for %s must exist", c.targetPath)
		assert.Equal(t, c.content, data,
			"backup for %s must retain its own content, not be overwritten by a same-basename backup",
			c.targetPath)
	}
}
