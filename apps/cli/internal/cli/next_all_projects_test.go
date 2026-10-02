package cli

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func phasedTaskMD(id, title, priority, phase string) string {
	return "---\nid: \"" + id + "\"\ntitle: \"" + title + "\"\nstatus: pending\npriority: " +
		priority + "\nphase: " + phase + "\n---\n# " + title + "\n"
}

// createPhasedProject writes a registered-project layout whose own
// .taskmd.yaml orders phases early → late. Task 001 is a critical late-phase
// task, so on score alone it outranks the low-priority early-phase task 002.
func createPhasedProject(t *testing.T) string {
	t.Helper()
	repo := newTaskRepo(t, nil)
	repo.Write(configFilename, "task-dir: tasks\nphases:\n  - id: early\n  - id: late\n")
	repo.Write(filepath.Join("tasks", "001-late.md"), phasedTaskMD("001", "Late", "critical", "late"))
	repo.Write(filepath.Join("tasks", "002-early.md"), phasedTaskMD("002", "Early", "low", "early"))
	return repo.Dir
}

type allProjectsRec struct {
	Project        string `json:"project"`
	ID             string `json:"id"`
	ScoreBreakdown []struct {
		Label string `json:"label"`
	} `json:"score_breakdown"`
}

func runNextAllProjectsJSON(t *testing.T, args ...string) []allProjectsRec {
	t.Helper()
	res := newTaskRepo(t, nil).Run(append([]string{"next", "--all-projects", "--format", "json"}, args...)...)
	if res.Err != nil {
		t.Fatalf("next --all-projects: %v", res.Err)
	}
	var recs []allProjectsRec
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("parse next output: %v\n%s", err, res.Stdout)
	}
	return recs
}

func indexOfRec(recs []allProjectsRec, project, id string) int {
	for i, r := range recs {
		if r.Project == project && r.ID == id {
			return i
		}
	}
	return -1
}

func TestNextAllProjects_UsesEachProjectsPhaseOrder(t *testing.T) {
	resetCLIState()
	defer resetCLIState()
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+createPhasedProject(t)+"\n")

	recs := runNextAllProjectsJSON(t)

	i := indexOfRec(recs, "alpha", "002")
	if i == -1 {
		t.Fatalf("expected alpha/002 in output, got %+v", recs)
	}
	found := false
	for _, c := range recs[i].ScoreBreakdown {
		if c.Label == "phase early" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected alpha/002 to carry its project's phase bonus, got %+v", recs[i].ScoreBreakdown)
	}
}

func TestNextAllProjects_StrictPhases_OrdersWithinProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()
	other := createProjectWithTasks(t, "tasks", map[string]string{
		"001-other.md": phasedTaskMD("001", "Other", "high", "unrelated"),
	})
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+createPhasedProject(t)+
		"\n  - id: beta\n    path: "+other+"\n")

	recs := runNextAllProjectsJSON(t, "--strict-phases")

	early, late := indexOfRec(recs, "alpha", "002"), indexOfRec(recs, "alpha", "001")
	if early == -1 || late == -1 {
		t.Fatalf("expected both alpha tasks in output, got %+v", recs)
	}
	if early > late {
		t.Errorf("with --strict-phases, alpha's early-phase 002 must precede late-phase 001; got %+v", recs)
	}
	if indexOfRec(recs, "beta", "001") == -1 {
		t.Errorf("expected beta/001 in merged output, got %+v", recs)
	}
}

func projectRec(project, id string, score int) ProjectRecommendation {
	return ProjectRecommendation{ProjectID: project, Recommendation: Recommendation{ID: id, Score: score}}
}

func TestMergeProjectRecs_PreservesPerProjectOrder(t *testing.T) {
	// Project a is strictly tiered: its first task scores lower than its second.
	merged := mergeProjectRecs([][]ProjectRecommendation{
		{projectRec("a", "a1", 10), projectRec("a", "a2", 50)},
		{projectRec("b", "b1", 30), projectRec("b", "b2", 20)},
	}, false)

	var got []string
	for _, r := range merged {
		got = append(got, r.ID)
	}
	want := []string{"b1", "b2", "a1", "a2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestMergeProjectRecs_ScoreSortedListsInterleaveByScore(t *testing.T) {
	merged := mergeProjectRecs([][]ProjectRecommendation{
		{projectRec("a", "a1", 50), projectRec("a", "a2", 10)},
		{projectRec("b", "b1", 40), projectRec("b", "b2", 20)},
		nil,
	}, false)

	want := []string{"a1", "b1", "b2", "a2"}
	for i, r := range merged {
		if r.ID != want[i] {
			t.Fatalf("position %d: got %s, want %s (full: %+v)", i, r.ID, want[i], merged)
		}
	}
	if len(merged) != len(want) {
		t.Fatalf("got %d recs, want %d", len(merged), len(want))
	}
}

func priorityRec(project, id, priority string, score int) ProjectRecommendation {
	r := projectRec(project, id, score)
	r.Priority = priority
	return r
}

func TestMergeProjectRecs_StrictPriority_TiersAcrossProjects(t *testing.T) {
	// b's high-priority task scores lower than a's medium one; strict priority
	// must still put it first, while each project keeps its own order.
	perProject := [][]ProjectRecommendation{
		{priorityRec("a", "a1", "medium", 60), priorityRec("a", "a2", "low", 70)},
		{priorityRec("b", "b1", "high", 20)},
	}

	got := mergeProjectRecs(perProject, true)

	want := []string{"b1", "a1", "a2"}
	for i, r := range got {
		if r.ID != want[i] {
			t.Fatalf("position %d: got %s, want %s (full: %+v)", i, r.ID, want[i], got)
		}
	}
	if loose := mergeProjectRecs(perProject, false); loose[0].ID != "a1" {
		t.Errorf("without strict priority, expected score to put a1 first, got %s", loose[0].ID)
	}
}

func TestNextAllProjects_StrictPriority_TiersAcrossProjects(t *testing.T) {
	resetCLIState()
	defer resetCLIState()
	// alpha's low task 002 carries a phase bonus; beta's high task carries a
	// large-effort penalty. The precondition below pins that, on score alone,
	// alpha/002 outranks beta/001.
	beta := createProjectWithTasks(t, "tasks", map[string]string{
		"001-beta.md": "---\nid: \"001\"\ntitle: \"Beta\"\nstatus: pending\npriority: high\neffort: large\n---\n# Beta\n",
	})
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+createPhasedProject(t)+
		"\n  - id: beta\n    path: "+beta+"\n")

	loose := runNextAllProjectsJSON(t)
	if indexOfRec(loose, "alpha", "002") > indexOfRec(loose, "beta", "001") {
		t.Skipf("fixture precondition not met (scoring changed?): %+v", loose)
	}

	recs := runNextAllProjectsJSON(t, "--strict-priority")

	if indexOfRec(recs, "beta", "001") > indexOfRec(recs, "alpha", "002") {
		t.Errorf("with --strict-priority, beta's high task must outrank alpha's low task; got %+v", recs)
	}
}

func TestNextAllProjects_UsesEachProjectsEffortScale(t *testing.T) {
	resetCLIState()
	defer resetCLIState()
	repo := newTaskRepo(t, nil)
	repo.Write(configFilename, "task-dir: tasks\neffort: [tiny, huge]\n")
	repo.Write(filepath.Join("tasks", "001-tiny.md"),
		"---\nid: \"001\"\ntitle: \"Tiny\"\nstatus: pending\npriority: medium\neffort: tiny\n---\n# Tiny\n")
	setupProjectFlagRegistry(t, "  - id: gamma\n    path: "+repo.Dir+"\n")

	// --quick-wins keeps only the lowest configured effort: "tiny" in gamma's
	// own vocabulary, which the default small/medium/large scale lacks.
	recs := runNextAllProjectsJSON(t, "--quick-wins")

	if indexOfRec(recs, "gamma", "001") == -1 {
		t.Errorf("expected gamma/001 as a quick win under its own effort scale, got %+v", recs)
	}
}
