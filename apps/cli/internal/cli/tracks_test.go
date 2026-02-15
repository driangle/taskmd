package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/tracks"
)

// createTracksTestTaskFiles creates task files with touches fields for testing.
//
// Task graph:
//
//	001 (completed)                      - root, completed
//	002 (pending, high, scope-a)         - depends on 001 (completed) -> actionable
//	003 (pending, critical, scope-a,b)   - depends on 001 (completed) -> actionable
//	004 (pending, medium, scope-c)       - no deps -> actionable
//	005 (pending, low, no touches)       - no deps -> actionable, flexible
//	006 (pending, medium)                - depends on 007 (pending) -> blocked
//	007 (pending, low, scope-b)          - no deps -> actionable
func createTracksTestTaskFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	tasks := map[string]string{
		"001.md": `---
id: "001"
title: "Setup infrastructure"
status: completed
priority: high
dependencies: []
tags: ["infra"]
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "Build graph colors"
status: pending
priority: high
dependencies: ["001"]
tags: ["cli"]
touches: ["scope-a"]
created: 2026-02-02
---`,
		"003.md": `---
id: "003"
title: "Refactor output"
status: pending
priority: critical
dependencies: ["001"]
tags: ["cli"]
touches: ["scope-a", "scope-b"]
created: 2026-02-03
---`,
		"004.md": `---
id: "004"
title: "Scanner improvements"
status: pending
priority: medium
dependencies: []
tags: ["core"]
touches: ["scope-c"]
created: 2026-02-04
---`,
		"005.md": `---
id: "005"
title: "Write README"
status: pending
priority: low
dependencies: []
tags: ["docs"]
created: 2026-02-05
---`,
		"006.md": `---
id: "006"
title: "Add auth"
status: pending
priority: medium
dependencies: ["007"]
tags: ["api"]
touches: ["scope-d"]
created: 2026-02-06
---`,
		"007.md": `---
id: "007"
title: "Create user model"
status: pending
priority: low
dependencies: []
tags: ["api"]
touches: ["scope-b"]
created: 2026-02-07
---`,
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
	return tmpDir
}

func captureTracksOutput(t *testing.T, args []string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runTracks(tracksCmd, args)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func resetTracksFlags() {
	tracksFormat = "table"
	tracksFilters = []string{}
	tracksLimit = 0
}

func TestTracks_JSONFormat(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if len(result.Tracks) == 0 {
		t.Error("Expected at least one track")
	}

	// Verify structure
	for _, track := range result.Tracks {
		if track.ID == 0 {
			t.Error("Expected non-zero track ID")
		}
		if len(track.Tasks) == 0 {
			t.Errorf("Track %d has no tasks", track.ID)
		}
		if len(track.Scopes) == 0 {
			t.Errorf("Track %d has no scopes", track.ID)
		}
	}
}

func TestTracks_YAMLFormat(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "yaml"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	if !strings.Contains(output, "tracks:") {
		t.Error("Expected YAML output to contain 'tracks:'")
	}
	if !strings.Contains(output, "flexible:") {
		t.Error("Expected YAML output to contain 'flexible:'")
	}
	if !strings.Contains(output, "scopes:") {
		t.Error("Expected YAML output to contain 'scopes:'")
	}
}

func TestTracks_TableFormat(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "table"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	if !strings.Contains(output, "Track") {
		t.Error("Expected table output to contain 'Track'")
	}
	if !strings.Contains(output, "Flexible") {
		t.Error("Expected table output to contain 'Flexible'")
	}
}

func TestTracks_OverlapSameTrack(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Tasks 002 and 003 both touch scope-a, so they must be in the same track (sequential).
	task002Track := -1
	task003Track := -1
	for _, track := range result.Tracks {
		for _, task := range track.Tasks {
			if task.ID == "002" {
				task002Track = track.ID
			}
			if task.ID == "003" {
				task003Track = track.ID
			}
		}
	}

	if task002Track == -1 {
		t.Error("Task 002 not found in any track")
	}
	if task003Track == -1 {
		t.Error("Task 003 not found in any track")
	}
	if task002Track != task003Track {
		t.Errorf("Tasks 002 and 003 share scope-a but are in different tracks (%d vs %d)", task002Track, task003Track)
	}
}

func TestTracks_NonOverlappingShareTrack(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Task 004 (scope-c) doesn't overlap with 002 (scope-a) or 003 (scope-a,b),
	// or 007 (scope-b), so it can potentially share a track.
	// Verify 004 exists in some track.
	found := false
	for _, track := range result.Tracks {
		for _, task := range track.Tasks {
			if task.ID == "004" {
				found = true
			}
		}
	}
	if !found {
		t.Error("Task 004 not found in any track")
	}
}

func TestTracks_BlockedExcluded(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Task 006 depends on 007 (pending) -> blocked.
	for _, track := range result.Tracks {
		for _, task := range track.Tasks {
			if task.ID == "006" {
				t.Error("Blocked task 006 should not appear in tracks")
			}
		}
	}
	for _, task := range result.Flexible {
		if task.ID == "006" {
			t.Error("Blocked task 006 should not appear in flexible")
		}
	}
}

func TestTracks_FlexibleTasks(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Task 005 has no touches -> flexible.
	found := false
	for _, task := range result.Flexible {
		if task.ID == "005" {
			found = true
		}
	}
	if !found {
		t.Error("Task 005 (no touches) should be in flexible section")
	}
}

func TestTracks_NoActionableTasks(t *testing.T) {
	tmpDir := t.TempDir()

	task := `---
id: "001"
title: "Done task"
status: completed
priority: high
dependencies: []
touches: ["scope-a"]
created: 2026-02-01
---`
	if err := os.WriteFile(filepath.Join(tmpDir, "001.md"), []byte(task), 0644); err != nil {
		t.Fatal(err)
	}

	resetTracksFlags()
	tracksFormat = "table"

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	if !strings.Contains(output, "No actionable tasks found") {
		t.Errorf("Expected 'No actionable tasks found', got: %s", output)
	}
}

func TestTracks_UnsupportedFormat(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "csv"

	_, err := captureTracksOutput(t, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestTracks_FilterFlag(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"
	tracksFilters = []string{"tag=cli"}

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Only CLI-tagged actionable tasks: 002, 003
	allIDs := make(map[string]bool)
	for _, track := range result.Tracks {
		for _, task := range track.Tasks {
			allIDs[task.ID] = true
		}
	}
	for _, task := range result.Flexible {
		allIDs[task.ID] = true
	}

	if allIDs["004"] || allIDs["005"] || allIDs["007"] {
		t.Error("Expected only CLI-tagged tasks after filter")
	}
	if !allIDs["002"] || !allIDs["003"] {
		t.Errorf("Expected tasks 002 and 003, got %v", allIDs)
	}
}

func TestTracks_LimitFlag(t *testing.T) {
	tmpDir := createTracksTestTaskFiles(t)
	resetTracksFlags()
	tracksFormat = "json"
	tracksLimit = 1

	output, err := captureTracksOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runTracks failed: %v", err)
	}

	var result tracks.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(result.Tracks) > 1 {
		t.Errorf("Expected at most 1 track with --limit 1, got %d", len(result.Tracks))
	}

	// Flexible tasks should still be present regardless of limit.
	if len(result.Flexible) == 0 {
		t.Error("Expected flexible tasks to still be present with --limit")
	}
}
