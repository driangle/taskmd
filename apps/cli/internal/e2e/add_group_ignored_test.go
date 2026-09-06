//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Reproduces the reported defect end-to-end: with `ignore: [content]` loaded
// from a real .taskmd.yaml, `add --group content` used to succeed and produce a
// task no read command could see. It must now refuse.
func TestAddGroup_IgnoredDirectory_Refused(t *testing.T) {
	root := t.TempDir()
	writeConfigFile(t, root, "dir: ./tasks\nignore:\n  - content\n")

	result := run(t, root, "add", "Probe task", "--group", "content")

	if result.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for an ignored group\nstdout: %s", result.Stdout)
	}
	for _, want := range []string{`--group "content"`, "ignore", ".taskmd.yaml"} {
		if !strings.Contains(result.Stderr, want) {
			t.Errorf("expected stderr to mention %q, got: %s", want, result.Stderr)
		}
	}

	matches, _ := filepath.Glob(filepath.Join(root, "tasks", "content", "*.md"))
	if len(matches) != 0 {
		t.Errorf("expected no task file written, found %v", matches)
	}

	// The non-ignored path is unaffected, and the task it creates is visible.
	mustRun(t, root, "add", "Real task", "--group", "cli")
	list := mustRun(t, root, "list")
	if !strings.Contains(list.Stdout, "Real task") {
		t.Errorf("expected the non-ignored task to be listed, got: %s", list.Stdout)
	}
}

func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".taskmd.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write .taskmd.yaml: %v", err)
	}
}
