package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/next"
)

var earlyLatePhases = []PhaseInfo{{ID: "early", Name: "Early"}, {ID: "late", Name: "Late"}}

// createPhasedTaskDir writes two otherwise identical tasks where the
// later-phase task has the lower ID, so only phase ranking can put the
// earlier-phase task first.
func createPhasedTaskDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"001-late.md":  "---\nid: \"001\"\ntitle: \"Late\"\nstatus: pending\npriority: medium\nphase: late\n---\n# Late\n",
		"002-early.md": "---\nid: \"002\"\ntitle: \"Early\"\nstatus: pending\npriority: medium\nphase: early\n---\n# Early\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

func serveNext(t *testing.T, handler http.HandlerFunc, req *http.Request) []next.Recommendation {
	t.Helper()
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var recs []next.Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &recs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return recs
}

func firstID(t *testing.T, recs []next.Recommendation) string {
	t.Helper()
	if len(recs) == 0 {
		t.Fatal("expected at least one recommendation")
	}
	return recs[0].ID
}

func TestHandleNext_PhaseOrder_RanksEarlierPhaseFirst(t *testing.T) {
	dp := NewDataProvider(createPhasedTaskDir(t), false)
	req := httptest.NewRequest(http.MethodGet, "/api/next", nil)

	recs := serveNext(t, handleNext(dp, earlyLatePhases, effort.Default()), req)

	if id := firstID(t, recs); id != "002" {
		t.Errorf("expected earlier-phase task 002 first, got %s", id)
	}
}

func TestHandleNext_NoPhases_IgnoresPhases(t *testing.T) {
	dp := NewDataProvider(createPhasedTaskDir(t), false)
	req := httptest.NewRequest(http.MethodGet, "/api/next", nil)

	recs := serveNext(t, handleNext(dp, nil, effort.Default()), req)

	if id := firstID(t, recs); id != "001" {
		t.Errorf("without phases, expected ID tiebreak to put 001 first, got %s", id)
	}
}

func TestHandleNext_ProjectScopedPhasesOverrideDefault(t *testing.T) {
	dp := NewDataProvider(createPhasedTaskDir(t), false)
	req := httptest.NewRequest(http.MethodGet, "/api/next", nil)
	// The resolved project orders the phases opposite to the server default.
	projectPhases := []PhaseInfo{{ID: "late"}, {ID: "early"}}
	req = req.WithContext(context.WithValue(req.Context(), projectPhasesKey, projectPhases))

	recs := serveNext(t, handleNext(dp, earlyLatePhases, effort.Default()), req)

	if id := firstID(t, recs); id != "001" {
		t.Errorf("expected project phase order to put 001 first, got %s", id)
	}
}

func TestExport_NextJSON_UsesPhaseOrder(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "export")
	err := exportWithMockFS(t, ExportConfig{
		OutputDir:  outDir,
		ScanDir:    createPhasedTaskDir(t),
		BasePath:   "/",
		PhaseOrder: []string{"early", "late"},
	})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "next.json"))
	if err != nil {
		t.Fatalf("failed to read next.json: %v", err)
	}
	var recs []next.Recommendation
	if err := json.Unmarshal(data, &recs); err != nil {
		t.Fatalf("invalid next.json: %v", err)
	}

	if id := firstID(t, recs); id != "002" {
		t.Errorf("expected earlier-phase task 002 first in next.json, got %s", id)
	}
}
