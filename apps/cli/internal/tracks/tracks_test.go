package tracks

import (
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func makeTask(id string, status model.Status, priority model.Priority, deps []string, touches []string) *model.Task {
	return &model.Task{
		ID:           id,
		Title:        "Task " + id,
		Status:       status,
		Priority:     priority,
		Dependencies: deps,
		Touches:      touches,
	}
}

func TestAssign_NoTasks(t *testing.T) {
	result, err := Assign(nil, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(result.Tracks))
	}
	if len(result.Flexible) != 0 {
		t.Errorf("expected 0 flexible, got %d", len(result.Flexible))
	}
}

func TestAssign_AllFlexible(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, nil),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, nil),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, nil),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 0 {
		t.Errorf("expected 0 tracks when all tasks are flexible, got %d", len(result.Tracks))
	}
	if len(result.Flexible) != 3 {
		t.Errorf("expected 3 flexible tasks, got %d", len(result.Flexible))
	}
}

func TestAssign_NoOverlap_SeparateTracks(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-b"}),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-c"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No scope overlap -> each task gets its own track (all parallelizable)
	if len(result.Tracks) != 3 {
		t.Errorf("expected 3 tracks (no overlaps, all parallel), got %d", len(result.Tracks))
	}
	for _, track := range result.Tracks {
		if len(track.Tasks) != 1 {
			t.Errorf("expected 1 task per track, got %d in track %d", len(track.Tasks), track.ID)
		}
	}
}

func TestAssign_FullOverlap_SingleTrack(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-a"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All share scope-a -> 1 track (must be sequential)
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track (all overlap), got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 3 {
		t.Errorf("expected 3 tasks in single track, got %d", len(result.Tracks[0].Tasks))
	}
}

func TestAssign_PartialOverlap(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a", "scope-b"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-b", "scope-c"}),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-c", "scope-d"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 001↔002 share scope-b, 002↔003 share scope-c → transitive → all 1 track
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track (transitive overlap), got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 3 {
		t.Errorf("expected 3 tasks in track, got %d", len(result.Tracks[0].Tasks))
	}
}

func TestAssign_CompletedAndBlockedExcluded(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusCompleted, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
		makeTask("003", model.StatusPending, model.PriorityLow, []string{"004"}, []string{"scope-b"}),
		makeTask("004", model.StatusPending, model.PriorityLow, nil, nil),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 001 completed -> excluded
	// 002 actionable with touches -> track
	// 003 blocked (dep 004 pending) -> excluded
	// 004 actionable, no touches -> flexible
	if len(result.Tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(result.Tracks))
	}
	if len(result.Tracks) > 0 {
		if len(result.Tracks[0].Tasks) != 1 || result.Tracks[0].Tasks[0].ID != "002" {
			t.Errorf("expected only task 002 in track, got %v", result.Tracks[0].Tasks)
		}
	}
	if len(result.Flexible) != 1 || result.Flexible[0].ID != "004" {
		t.Errorf("expected task 004 as flexible, got %v", result.Flexible)
	}
}

func TestAssign_UnknownScopeWarnings(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"known", "unknown-x"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"unknown-y"}),
	}

	known := map[string]bool{"known": true}
	result, err := Assign(tasks, Options{KnownScopes: known})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestAssign_NoWarningsWhenKnownScopesNil(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"anything"}),
	}

	result, err := Assign(tasks, Options{KnownScopes: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings when KnownScopes is nil, got %v", result.Warnings)
	}
}

func TestAssign_MixedTouchesAndFlexible(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, nil),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-b"}),
		makeTask("004", model.StatusInProgress, model.PriorityLow, nil, nil),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// scope-a and scope-b don't overlap -> 2 tracks
	if len(result.Tracks) != 2 {
		t.Errorf("expected 2 tracks (no overlap between a and b), got %d", len(result.Tracks))
	}
	if len(result.Flexible) != 2 {
		t.Errorf("expected 2 flexible tasks, got %d", len(result.Flexible))
	}
}

func TestAssign_TrackScopesUnion(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a", "scope-b"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-c"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No overlap between tasks -> 2 separate tracks
	if len(result.Tracks) != 2 {
		t.Fatalf("expected 2 tracks (no overlap), got %d", len(result.Tracks))
	}
	// Track 1 should have scopes a,b from task 001
	scopes1 := make(map[string]bool)
	for _, s := range result.Tracks[0].Scopes {
		scopes1[s] = true
	}
	if !scopes1["scope-a"] || !scopes1["scope-b"] {
		t.Errorf("expected track 1 scopes [scope-a, scope-b], got %v", result.Tracks[0].Scopes)
	}
	// Track 2 should have scope-c from task 002
	if len(result.Tracks[1].Scopes) != 1 || result.Tracks[1].Scopes[0] != "scope-c" {
		t.Errorf("expected track 2 scopes [scope-c], got %v", result.Tracks[1].Scopes)
	}
}

func TestAssign_DeterministicOrdering(t *testing.T) {
	tasks := []*model.Task{
		makeTask("003", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
		makeTask("001", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All share scope-a -> 1 track, tasks sorted by ID ascending
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 3 {
		t.Fatalf("expected 3 tasks in track, got %d", len(result.Tracks[0].Tasks))
	}
	ids := []string{
		result.Tracks[0].Tasks[0].ID,
		result.Tracks[0].Tasks[1].ID,
		result.Tracks[0].Tasks[2].ID,
	}
	if ids[0] != "001" || ids[1] != "002" || ids[2] != "003" {
		t.Errorf("expected tasks ordered by ID [001, 002, 003], got %v", ids)
	}
}

func TestAssign_TrackIDs(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both share scope-a -> 1 track
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(result.Tracks))
	}
	if result.Tracks[0].ID != 1 {
		t.Errorf("expected track ID=1, got %d", result.Tracks[0].ID)
	}
	if len(result.Tracks[0].Tasks) != 2 {
		t.Errorf("expected 2 tasks in track, got %d", len(result.Tracks[0].Tasks))
	}
}

func TestAssign_HigherScoreTaskFirst(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityLow, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityCritical, nil, []string{"scope-a"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both share scope-a -> 1 track, higher score (002) first
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 2 {
		t.Fatalf("expected 2 tasks in track, got %d", len(result.Tracks[0].Tasks))
	}
	if result.Tracks[0].Tasks[0].ID != "002" {
		t.Errorf("expected task 002 (critical) first, got %s", result.Tracks[0].Tasks[0].ID)
	}
	if result.Tracks[0].Tasks[1].ID != "001" {
		t.Errorf("expected task 001 (low) second, got %s", result.Tracks[0].Tasks[1].ID)
	}
}

func TestAssign_IdenticalTouches_SameTrack(t *testing.T) {
	// Bug report scenario: two tasks with identical touches must land in the same track
	tasks := []*model.Task{
		makeTask("112", model.StatusPending, model.PriorityHigh, nil, []string{"cli", "cli/mcp"}),
		makeTask("113", model.StatusPending, model.PriorityHigh, nil, []string{"cli", "cli/mcp"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track (identical touches), got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 2 {
		t.Errorf("expected 2 tasks in track, got %d", len(result.Tracks[0].Tasks))
	}
}

func TestAssign_TransitiveOverlap(t *testing.T) {
	// A shares scope-x with B, B shares scope-y with C.
	// A and C don't directly overlap but are transitively connected -> 1 track.
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-x"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-x", "scope-y"}),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-y"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track (transitive overlap), got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 3 {
		t.Errorf("expected 3 tasks in track, got %d", len(result.Tracks[0].Tasks))
	}
}

func TestAssign_IndependentGroups(t *testing.T) {
	// Two disjoint clusters: {001, 002} share scope-a, {003, 004} share scope-b
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
		makeTask("003", model.StatusPending, model.PriorityHigh, nil, []string{"scope-b"}),
		makeTask("004", model.StatusPending, model.PriorityMedium, nil, []string{"scope-b"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 2 {
		t.Fatalf("expected 2 tracks (two independent groups), got %d", len(result.Tracks))
	}
	// Each track should have 2 tasks
	for _, track := range result.Tracks {
		if len(track.Tasks) != 2 {
			t.Errorf("expected 2 tasks in track %d, got %d", track.ID, len(track.Tasks))
		}
	}
}
