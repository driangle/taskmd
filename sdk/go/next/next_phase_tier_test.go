package next

import (
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

// recommendStrictPhases ranks tasks with strict phase ordering over v0.2, v0.3
// and returns the ranked IDs.
func recommendStrictPhases(t *testing.T, tasks []*model.Task) []string {
	t.Helper()
	recs, err := Recommend(tasks, Options{
		Limit:        10,
		PhaseOrder:   []string{"v0.2", "v0.3"},
		StrictPhases: true,
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

func assertOrder(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestRecommend_StrictPhases_UnphasedJoinsCurrentTier(t *testing.T) {
	// The unphased critical bug outscores the v0.2 chore, so it ranks first
	// within the current (v0.2) tier — and both still precede v0.3 work.
	ids := recommendStrictPhases(t, []*model.Task{
		makeTaskWithPhase("chore", model.StatusPending, model.PriorityLow, "v0.2"),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityCritical, "v0.3"),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
	})
	assertOrder(t, ids, "bug", "chore", "late")
}

func TestRecommend_StrictPhases_UnphasedCompetesOnScore(t *testing.T) {
	// A low-priority unphased task does not jump the current-phase task.
	ids := recommendStrictPhases(t, []*model.Task{
		makeTask("misc", model.StatusPending, model.PriorityLow, nil),
		makeTaskWithPhase("early", model.StatusPending, model.PriorityHigh, "v0.2"),
	})
	assertOrder(t, ids, "early", "misc")
}

func TestRecommend_StrictPhases_CurrentTierIsEarliestActionablePhase(t *testing.T) {
	// v0.2 has no actionable work (its only task is completed), so v0.3 is the
	// current tier and the unphased task competes there.
	ids := recommendStrictPhases(t, []*model.Task{
		makeTaskWithPhase("done", model.StatusCompleted, model.PriorityHigh, "v0.2"),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityLow, "v0.3"),
		makeTask("bug", model.StatusPending, model.PriorityCritical, nil),
	})
	assertOrder(t, ids, "bug", "late")
}

func TestRecommend_StrictPhases_UnknownPhaseSortsLast(t *testing.T) {
	ids := recommendStrictPhases(t, []*model.Task{
		makeTaskWithPhase("typo", model.StatusPending, model.PriorityCritical, "v9.9"),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityLow, "v0.3"),
		makeTask("misc", model.StatusPending, model.PriorityLow, nil),
	})
	assertOrder(t, ids, "late", "misc", "typo")
}

func TestRecommend_StrictPhases_UnknownPhaseAfterUnphasedWithoutKnownPhases(t *testing.T) {
	// With no task in a configured phase, unphased tasks still rank ahead of
	// tasks whose phase is not in the configured order.
	ids := recommendStrictPhases(t, []*model.Task{
		makeTaskWithPhase("typo", model.StatusPending, model.PriorityCritical, "v9.9"),
		makeTask("misc", model.StatusPending, model.PriorityLow, nil),
	})
	assertOrder(t, ids, "misc", "typo")
}

func TestRecommend_StrictPhasesOff_ScoreOnly(t *testing.T) {
	recs, err := Recommend([]*model.Task{
		makeTaskWithPhase("chore", model.StatusPending, model.PriorityLow, "v0.2"),
		makeTaskWithPhase("late", model.StatusPending, model.PriorityCritical, "v0.3"),
	}, Options{Limit: 10, PhaseOrder: []string{"v0.2", "v0.3"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs[0].ID != "late" {
		t.Errorf("without StrictPhases the higher-scoring later-phase task should lead, got %s", recs[0].ID)
	}
}
