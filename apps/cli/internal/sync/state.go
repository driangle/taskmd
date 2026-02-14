package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const stateSubDir = ".taskmd/sync-state"

// SyncState tracks the last-synced state for a single source.
type SyncState struct {
	Source   string               `yaml:"source"`
	Tasks    map[string]TaskState `yaml:"tasks"`
	LastSync time.Time            `yaml:"last_sync"`
}

// TaskState tracks the sync state of a single task.
type TaskState struct {
	ExternalID   string    `yaml:"external_id"`
	LocalID      string    `yaml:"local_id"`
	FilePath     string    `yaml:"file_path"`
	ExternalHash string    `yaml:"external_hash"`
	LocalHash    string    `yaml:"local_hash"`
	LastSynced   time.Time `yaml:"last_synced"`
}

// LoadState reads the state file for a source. Returns empty state if file doesn't exist.
func LoadState(dir, sourceName string) (*SyncState, error) {
	path := stateFilePath(dir, sourceName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &SyncState{
			Source: sourceName,
			Tasks:  make(map[string]TaskState),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state SyncState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Tasks == nil {
		state.Tasks = make(map[string]TaskState)
	}

	return &state, nil
}

// SaveState writes the state file for a source.
func SaveState(dir, sourceName string, state *SyncState) error {
	stateDir := filepath.Join(dir, stateSubDir)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path := stateFilePath(dir, sourceName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func stateFilePath(dir, sourceName string) string {
	return filepath.Join(dir, stateSubDir, sourceName+".yaml")
}
