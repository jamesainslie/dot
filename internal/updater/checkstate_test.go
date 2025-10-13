package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager("/test/config")
	require.NotNil(t, sm)
	assert.Equal(t, filepath.Join("/test/config", "update-check.json"), sm.statePath)
}

func TestStateManager_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	// Load from non-existent file should return empty state
	state, err := sm.Load()
	require.NoError(t, err)
	assert.True(t, state.LastCheck.IsZero())
	assert.True(t, state.LastSkip.IsZero())

	// Save state
	now := time.Now()
	state.LastCheck = now
	state.LastSkip = now.Add(-1 * time.Hour)

	err = sm.Save(state)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(sm.statePath)
	require.NoError(t, err)

	// Load saved state
	loaded, err := sm.Load()
	require.NoError(t, err)
	assert.WithinDuration(t, now, loaded.LastCheck, time.Second)
	assert.WithinDuration(t, now.Add(-1*time.Hour), loaded.LastSkip, time.Second)
}

func TestStateManager_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	// Write invalid JSON
	err := os.WriteFile(sm.statePath, []byte("invalid json"), 0o644)
	require.NoError(t, err)

	// Load should return error
	_, err = sm.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse state file")
}

func TestStateManager_Save_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "nested", "path")
	sm := NewStateManager(subDir)

	state := &CheckState{
		LastCheck: time.Now(),
	}

	err := sm.Save(state)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(subDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestStateManager_ShouldCheck(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	t.Run("first check", func(t *testing.T) {
		// No state file exists
		should, err := sm.ShouldCheck(24 * time.Hour)
		require.NoError(t, err)
		assert.True(t, should, "should check on first run")
	})

	t.Run("check too soon", func(t *testing.T) {
		// Save recent check
		state := &CheckState{
			LastCheck: time.Now().Add(-1 * time.Hour),
		}
		err := sm.Save(state)
		require.NoError(t, err)

		should, err := sm.ShouldCheck(24 * time.Hour)
		require.NoError(t, err)
		assert.False(t, should, "should not check if less than frequency has passed")
	})

	t.Run("check after frequency", func(t *testing.T) {
		// Save old check
		state := &CheckState{
			LastCheck: time.Now().Add(-25 * time.Hour),
		}
		err := sm.Save(state)
		require.NoError(t, err)

		should, err := sm.ShouldCheck(24 * time.Hour)
		require.NoError(t, err)
		assert.True(t, should, "should check if more than frequency has passed")
	})

	t.Run("exact frequency boundary", func(t *testing.T) {
		// Save check exactly at frequency
		state := &CheckState{
			LastCheck: time.Now().Add(-24 * time.Hour),
		}
		err := sm.Save(state)
		require.NoError(t, err)

		should, err := sm.ShouldCheck(24 * time.Hour)
		require.NoError(t, err)
		assert.True(t, should, "should check at exact frequency boundary")
	})
}

func TestStateManager_RecordCheck(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	before := time.Now()
	err := sm.RecordCheck()
	require.NoError(t, err)
	after := time.Now()

	// Verify check was recorded
	state, err := sm.Load()
	require.NoError(t, err)
	assert.False(t, state.LastCheck.IsZero())
	assert.True(t, state.LastCheck.After(before) || state.LastCheck.Equal(before))
	assert.True(t, state.LastCheck.Before(after) || state.LastCheck.Equal(after))
}

func TestStateManager_RecordSkip(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	before := time.Now()
	err := sm.RecordSkip()
	require.NoError(t, err)
	after := time.Now()

	// Verify skip was recorded
	state, err := sm.Load()
	require.NoError(t, err)
	assert.False(t, state.LastSkip.IsZero())
	assert.True(t, state.LastSkip.After(before) || state.LastSkip.Equal(before))
	assert.True(t, state.LastSkip.Before(after) || state.LastSkip.Equal(after))
}

func TestStateManager_RecordCheck_PreservesSkip(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	// Record a skip
	skipTime := time.Now().Add(-1 * time.Hour)
	state := &CheckState{
		LastSkip: skipTime,
	}
	err := sm.Save(state)
	require.NoError(t, err)

	// Record a check
	err = sm.RecordCheck()
	require.NoError(t, err)

	// Verify skip is preserved
	state, err = sm.Load()
	require.NoError(t, err)
	assert.WithinDuration(t, skipTime, state.LastSkip, time.Second)
	assert.False(t, state.LastCheck.IsZero())
}

func TestCheckState_JSON(t *testing.T) {
	state := &CheckState{
		LastCheck: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		LastSkip:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// Marshal
	data, err := json.MarshalIndent(state, "", "  ")
	require.NoError(t, err)

	// Unmarshal
	var loaded CheckState
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, state.LastCheck.Unix(), loaded.LastCheck.Unix())
	assert.Equal(t, state.LastSkip.Unix(), loaded.LastSkip.Unix())
}
