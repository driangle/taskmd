package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/next"
)

// createPhasedTaskFiles writes two otherwise identical tasks where the
// later-phase task has the lower ID, so only phase ranking can put the
// earlier-phase task first.
func createPhasedTaskFiles(t *testing.T) string {
	t.Helper()
	return writePhasedTaskFiles(t, "medium", "medium")
}

// writePhasedTaskFiles writes task 001 in phase "late" and task 002 in phase
// "early" with the given priorities.
func writePhasedTaskFiles(t *testing.T, latePriority, earlyPriority string) string {
	t.Helper()
	tmpDir := t.TempDir()
	files := map[string]string{
		"001-late.md":  "---\nid: \"001\"\ntitle: \"Late\"\nstatus: pending\npriority: " + latePriority + "\nphase: late\n---\n# Late\n",
		"002-early.md": "---\nid: \"002\"\ntitle: \"Early\"\nstatus: pending\npriority: " + earlyPriority + "\nphase: early\n---\n# Early\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return tmpDir
}

func TestNextTool_PhaseOrder_RanksEarlierPhaseFirst(t *testing.T) {
	tmpDir := createPhasedTaskFiles(t)
	session := setupTestServerWithConfig(t, Config{
		Efforts:    effort.Default(),
		PhaseOrder: []string{"early", "late"},
	})

	recs := callNext(t, session, map[string]any{"task_dir": tmpDir})

	if len(recs) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(recs))
	}
	if recs[0].ID != "002" {
		t.Errorf("expected earlier-phase task 002 first, got %s", recs[0].ID)
	}
	if !hasScoreComponent(recs[0], "phase early") {
		t.Errorf("expected a phase score component, got %+v", recs[0].ScoreBreakdown)
	}
}

func TestNextTool_NoPhaseOrder_IgnoresPhases(t *testing.T) {
	tmpDir := createPhasedTaskFiles(t)
	session := setupTestServer(t)

	recs := callNext(t, session, map[string]any{"task_dir": tmpDir})

	if len(recs) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(recs))
	}
	if recs[0].ID != "001" {
		t.Errorf("without phase order, expected ID tiebreak to put 001 first, got %s", recs[0].ID)
	}
}

func TestNextTool_PhaseOrder_StrictByDefault(t *testing.T) {
	// The critical later-phase task outscores the low earlier-phase one, so
	// only strict phase tiering can put the earlier-phase task first.
	tmpDir := writePhasedTaskFiles(t, "critical", "low")
	session := setupTestServerWithConfig(t, Config{
		Efforts:    effort.Default(),
		PhaseOrder: []string{"early", "late"},
	})

	recs := callNext(t, session, map[string]any{"task_dir": tmpDir})

	if len(recs) != 2 || recs[0].ID != "002" {
		t.Errorf("expected strict phase order to put low-priority early task 002 first, got %+v", recs)
	}
}

func hasScoreComponent(rec next.Recommendation, label string) bool {
	for _, c := range rec.ScoreBreakdown {
		if c.Label == label {
			return true
		}
	}
	return false
}
