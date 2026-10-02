package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/sdk/go/next"
)

// createUnphasedTaskDir writes a critical unphased task 001 that outscores the
// low-priority early-phase task 002, so only UnphasedLast puts 002 first.
func createUnphasedTaskDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"001-bug.md":   "---\nid: \"001\"\ntitle: \"Bug\"\nstatus: pending\npriority: critical\n---\n# Bug\n",
		"002-early.md": "---\nid: \"002\"\ntitle: \"Early\"\nstatus: pending\npriority: low\nphase: early\n---\n# Early\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

func TestHandleNext_UnphasedPlacement(t *testing.T) {
	cases := []struct {
		placement next.UnphasedPlacement
		wantFirst string
	}{
		{"", "001"},
		{next.UnphasedCurrent, "001"},
		{next.UnphasedLast, "002"},
	}
	for _, tc := range cases {
		t.Run(string(tc.placement), func(t *testing.T) {
			dp := NewDataProvider(createUnphasedTaskDir(t), false)
			req := httptest.NewRequest(http.MethodGet, "/api/next", nil)

			recs := serveNext(t, handleNext(dp, Config{Phases: earlyLatePhases, UnphasedPlacement: tc.placement}), req)

			if id := firstID(t, recs); id != tc.wantFirst {
				t.Errorf("unphased=%q: expected %s first, got %s", tc.placement, tc.wantFirst, id)
			}
		})
	}
}

func TestExport_NextJSON_UnphasedLast(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "export")
	err := exportWithMockFS(t, ExportConfig{
		OutputDir:         outDir,
		ScanDir:           createUnphasedTaskDir(t),
		BasePath:          "/",
		PhaseOrder:        []string{"early", "late"},
		UnphasedPlacement: next.UnphasedLast,
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
		t.Errorf("expected unphased task to rank after early task 002 in next.json, got %s first", id)
	}
}
