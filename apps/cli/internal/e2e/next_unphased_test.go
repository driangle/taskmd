//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newUnphasedProject writes a project with one phase and the given extra
// config. Unphased 001 is critical and outscores the low early-phase 002, so
// only next.unphased: last puts 002 first.
func newUnphasedProject(t *testing.T, extraConfig string) string {
	t.Helper()
	root := t.TempDir()
	writeConfig(t, root, "phases:\n  - id: early\n"+extraConfig)
	files := map[string]string{
		"001-bug.md":   "---\nid: \"001\"\ntitle: \"Bug\"\nstatus: pending\npriority: critical\n---\n# Bug\n",
		"002-early.md": "---\nid: \"002\"\ntitle: \"Early\"\nstatus: pending\npriority: low\nphase: early\n---\n# Early\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return root
}

func nextFirstID(t *testing.T, root string) string {
	t.Helper()
	result := mustRun(t, root, "next", "--format", "json")
	var recs []nextRec
	if err := json.Unmarshal([]byte(result.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, result.Stdout)
	}
	if len(recs) == 0 {
		t.Fatal("expected recommendations")
	}
	return recs[0].ID
}

func TestNext_UnphasedConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		wantFirst string
	}{
		{"absent", "", "001"},
		{"current", "next:\n  unphased: current\n", "001"},
		{"last", "next:\n  unphased: last\n", "002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextFirstID(t, newUnphasedProject(t, tt.config)); got != tt.wantFirst {
				t.Errorf("expected %s first, got %s", tt.wantFirst, got)
			}
		})
	}
}

func TestNext_UnphasedConfig_InvalidValue(t *testing.T) {
	root := newUnphasedProject(t, "next:\n  unphased: lst\n")

	for _, args := range [][]string{{"next"}, {"validate"}} {
		result := run(t, root, args...)
		if result.ExitCode == 0 {
			t.Fatalf("expected %v to fail, got exit 0:\n%s", args, result.Stdout)
		}
		if combined := result.Stdout + result.Stderr; !strings.Contains(combined, "invalid next.unphased") {
			t.Errorf("%v: expected a next.unphased config error, got:\n%s", args, combined)
		}
	}
}

func TestNext_UnphasedConfig_ValidKeyNotUnknown(t *testing.T) {
	root := newUnphasedProject(t, "next:\n  unphased: last\n")

	result := run(t, root, "validate")
	if combined := result.Stdout + result.Stderr; strings.Contains(combined, "unknown config key") {
		t.Errorf("next must be a known config key, got:\n%s", combined)
	}
}
