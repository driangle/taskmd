package sync

import (
	"testing"
	"time"
)

func TestLoadState_NonexistentFile(t *testing.T) {
	dir := t.TempDir()

	state, err := LoadState(dir, "github")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.Source != "github" {
		t.Errorf("expected source=github, got %q", state.Source)
	}
	if len(state.Tasks) != 0 {
		t.Errorf("expected empty tasks map, got %d entries", len(state.Tasks))
	}
}

func TestState_SaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	original := &SyncState{
		Source:   "github",
		LastSync: now,
		Tasks: map[string]TaskState{
			"GH-1": {
				ExternalID:   "GH-1",
				LocalID:      "001",
				FilePath:     "tasks/github/001-fix-bug.md",
				ExternalHash: "abc123",
				LocalHash:    "def456",
				LastSynced:   now,
			},
			"GH-2": {
				ExternalID:   "GH-2",
				LocalID:      "002",
				FilePath:     "tasks/github/002-add-feature.md",
				ExternalHash: "ghi789",
				LocalHash:    "jkl012",
				LastSynced:   now,
			},
		},
	}

	if err := SaveState(dir, "github", original); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	loaded, err := LoadState(dir, "github")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.Source != original.Source {
		t.Errorf("source mismatch: got %q, want %q", loaded.Source, original.Source)
	}
	if len(loaded.Tasks) != len(original.Tasks) {
		t.Fatalf("task count mismatch: got %d, want %d", len(loaded.Tasks), len(original.Tasks))
	}

	for id, orig := range original.Tasks {
		got, ok := loaded.Tasks[id]
		if !ok {
			t.Errorf("missing task %q", id)
			continue
		}
		if got.LocalID != orig.LocalID {
			t.Errorf("task %s: local_id mismatch: got %q, want %q", id, got.LocalID, orig.LocalID)
		}
		if got.FilePath != orig.FilePath {
			t.Errorf("task %s: file_path mismatch: got %q, want %q", id, got.FilePath, orig.FilePath)
		}
		if got.ExternalHash != orig.ExternalHash {
			t.Errorf("task %s: external_hash mismatch", id)
		}
		if got.LocalHash != orig.LocalHash {
			t.Errorf("task %s: local_hash mismatch", id)
		}
	}
}

func TestState_SaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()

	state := &SyncState{
		Source: "jira",
		Tasks:  map[string]TaskState{},
	}

	if err := SaveState(dir, "jira", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Load should succeed
	loaded, err := LoadState(dir, "jira")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if loaded.Source != "jira" {
		t.Errorf("expected source=jira, got %q", loaded.Source)
	}
}
