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

func TestAssign_NoOverlap_SingleTrack(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-b"}),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-c"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 1 {
		t.Errorf("expected 1 track (no overlaps), got %d", len(result.Tracks))
	}
	if len(result.Tracks) > 0 && len(result.Tracks[0].Tasks) != 3 {
		t.Errorf("expected 3 tasks in single track, got %d", len(result.Tracks[0].Tasks))
	}
}

func TestAssign_FullOverlap_SeparateTracks(t *testing.T) {
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil, []string{"scope-a"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil, []string{"scope-a"}),
		makeTask("003", model.StatusPending, model.PriorityLow, nil, []string{"scope-a"}),
	}

	result, err := Assign(tasks, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tracks) != 3 {
		t.Errorf("expected 3 tracks (all overlap), got %d", len(result.Tracks))
	}
	for _, track := range result.Tracks {
		if len(track.Tasks) != 1 {
			t.Errorf("expected 1 task per track, got %d in track %d", len(track.Tasks), track.ID)
		}
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

	// 001 touches a,b -> track 1
	// 002 touches b,c -> overlaps with 001 (b) -> track 2
	// 003 touches c,d -> overlaps with 002 (c) but not 001 -> track 1
	if len(result.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(result.Tracks))
	}
	if len(result.Tracks[0].Tasks) != 2 {
		t.Errorf("expected 2 tasks in track 1, got %d", len(result.Tracks[0].Tasks))
	}
	if len(result.Tracks[1].Tasks) != 1 {
		t.Errorf("expected 1 task in track 2, got %d", len(result.Tracks[1].Tasks))
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

	if len(result.Tracks) != 1 {
		t.Errorf("expected 1 track (no overlap), got %d", len(result.Tracks))
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

	if len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(result.Tracks))
	}
	scopes := result.Tracks[0].Scopes
	if len(scopes) != 3 {
		t.Errorf("expected 3 scopes in track, got %d: %v", len(scopes), scopes)
	}
	scopeSet := make(map[string]bool)
	for _, s := range scopes {
		scopeSet[s] = true
	}
	for _, want := range []string{"scope-a", "scope-b", "scope-c"} {
		if !scopeSet[want] {
			t.Errorf("expected scope %q in track scopes %v", want, scopes)
		}
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

	// All same score, so sorted by ID ascending.
	// Each overlaps, so 3 separate tracks.
	if len(result.Tracks) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(result.Tracks))
	}
	ids := []string{
		result.Tracks[0].Tasks[0].ID,
		result.Tracks[1].Tasks[0].ID,
		result.Tracks[2].Tasks[0].ID,
	}
	if ids[0] != "001" || ids[1] != "002" || ids[2] != "003" {
		t.Errorf("expected tracks ordered by ID [001, 002, 003], got %v", ids)
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

	if len(result.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(result.Tracks))
	}
	if result.Tracks[0].ID != 1 {
		t.Errorf("expected track 1 ID=1, got %d", result.Tracks[0].ID)
	}
	if result.Tracks[1].ID != 2 {
		t.Errorf("expected track 2 ID=2, got %d", result.Tracks[1].ID)
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

	if len(result.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(result.Tracks))
	}
	// Higher priority task (002) should be in track 1
	if result.Tracks[0].Tasks[0].ID != "002" {
		t.Errorf("expected task 002 (critical) in track 1, got %s", result.Tracks[0].Tasks[0].ID)
	}
	if result.Tracks[1].Tasks[0].ID != "001" {
		t.Errorf("expected task 001 (low) in track 2, got %s", result.Tracks[1].Tasks[0].ID)
	}
}
