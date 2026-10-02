package next

import (
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

// recommendUnphased ranks tasks with strict phase ordering over v0.2, v0.3,
// placing unphased tasks as given, and returns the ranked IDs.
func recommendUnphased(t *testing.T, placement UnphasedPlacement, tasks []*model.Task) []string {
	t.Helper()
	recs, err := Recommend(tasks, Options{
		Limit:             10,
		PhaseOrder:        []string{"v0.2", "v0.3"},
		StrictPhases:      true,
		UnphasedPlacement: placement,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ids := make([]string, len(recs))
	for i, r := range recs {
		ids[i] = r.ID
	}
	return ids
}

func TestRecommend_UnphasedLast_RanksAfterEveryPhase(t *testing.T) {
	// The critical unphased bug outscores both phased tasks but still trails
	// them, while staying ahead of the unknown-phase task.
	ids := recommendUnphased(t, UnphasedLast, []*model.Task{
		makeTaskWithPhase("chore", model.StatusPending, model.PriorityLow, "v0.2"),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityLow, "v0.3"),
		makeTaskWithPhase("typo", model.StatusPending, model.PriorityCritical, "v9.9"),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
	})
	assertOrder(t, ids, "chore", "late", "bug", "typo")
}

func TestRecommend_UnphasedLast_CompetesOnScoreAmongUnphased(t *testing.T) {
	ids := recommendUnphased(t, UnphasedLast, []*model.Task{
		makeTask("misc", model.StatusPending, model.PriorityLow, nil),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityLow, "v0.3"),
	})
	assertOrder(t, ids, "late", "bug", "misc")
}

func TestRecommend_UnphasedCurrent_MatchesZeroValue(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithPhase("chore", model.StatusPending, model.PriorityLow, "v0.2"),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityCritical, "v0.3"),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
	}
	assertOrder(t, recommendUnphased(t, UnphasedCurrent, tasks), "bug", "chore", "late")
	assertOrder(t, recommendUnphased(t, "", tasks), "bug", "chore", "late")
}

func TestRecommend_UnphasedLast_NoEffectWithoutStrictPhases(t *testing.T) {
	recs, err := Recommend([]*model.Task{
		makeTaskWithPhase("chore", model.StatusPending, model.PriorityLow, "v0.2"),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
	}, Options{Limit: 10, PhaseOrder: []string{"v0.2", "v0.3"}, UnphasedPlacement: UnphasedLast})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs[0].ID != "bug" {
		t.Errorf("without StrictPhases the higher-scoring unphased task should lead, got %s", recs[0].ID)
	}
}

func TestRecommend_UnphasedLast_NoEffectWithoutPhases(t *testing.T) {
	recs, err := Recommend([]*model.Task{
		makeTask("misc", model.StatusPending, model.PriorityLow, nil),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
	}, Options{Limit: 10, StrictPhases: true, UnphasedPlacement: UnphasedLast})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs[0].ID != "bug" {
		t.Errorf("without phases ranking is by score, got %s first", recs[0].ID)
	}
}
