package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/next"
)

// writeUnphasedTaskFiles writes a critical unphased task 001 that outscores the
// low-priority early-phase task 002, so only UnphasedLast puts 002 first.
func writeUnphasedTaskFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	files := map[string]string{
		"001-bug.md":   "---\nid: \"001\"\ntitle: \"Bug\"\nstatus: pending\npriority: critical\n---\n# Bug\n",
		"002-early.md": "---\nid: \"002\"\ntitle: \"Early\"\nstatus: pending\npriority: low\nphase: early\n---\n# Early\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return tmpDir
}

func TestNextTool_UnphasedPlacement(t *testing.T) {
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
			tmpDir := writeUnphasedTaskFiles(t)
			session := setupTestServerWithConfig(t, Config{
				Efforts:           effort.Default(),
				PhaseOrder:        []string{"early", "late"},
				UnphasedPlacement: tc.placement,
			})

			recs := callNext(t, session, map[string]any{"task_dir": tmpDir})

			if len(recs) != 2 || recs[0].ID != tc.wantFirst {
				t.Errorf("unphased=%q: expected %s first, got %+v", tc.placement, tc.wantFirst, recs)
			}
		})
	}
}
