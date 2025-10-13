package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CheckState stores information about the last update check.
type CheckState struct {
	LastCheck time.Time `json:"last_check"`
	LastSkip  time.Time `json:"last_skip"`
}

// StateManager manages the update check state file.
type StateManager struct {
	statePath string
}

// NewStateManager creates a new state manager.
func NewStateManager(configDir string) *StateManager {
	return &StateManager{
		statePath: filepath.Join(configDir, "update-check.json"),
	}
}

// Load loads the check state from disk.
func (sm *StateManager) Load() (*CheckState, error) {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default state if file doesn't exist
			return &CheckState{}, nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var state CheckState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}

	return &state, nil
}

// Save saves the check state to disk.
func (sm *StateManager) Save(state *CheckState) error {
	// Ensure directory exists
	dir := filepath.Dir(sm.statePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(sm.statePath, data, 0o600); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// ShouldCheck determines if enough time has passed since the last check.
func (sm *StateManager) ShouldCheck(frequency time.Duration) (bool, error) {
	state, err := sm.Load()
	if err != nil {
		return false, err
	}

	// If we've never checked, we should check
	if state.LastCheck.IsZero() {
		return true, nil
	}

	// Check if enough time has passed
	return time.Since(state.LastCheck) >= frequency, nil
}

// RecordCheck records that a check was performed.
func (sm *StateManager) RecordCheck() error {
	state, err := sm.Load()
	if err != nil {
		// If we can't load, create new state
		state = &CheckState{}
	}

	state.LastCheck = time.Now()
	return sm.Save(state)
}

// RecordSkip records that the user skipped an upgrade prompt.
func (sm *StateManager) RecordSkip() error {
	state, err := sm.Load()
	if err != nil {
		state = &CheckState{}
	}

	state.LastSkip = time.Now()
	return sm.Save(state)
}
