//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

// listGroupIDs runs `list --format json --sort id <args...>` and returns the task IDs.
func listGroupIDs(t *testing.T, dir string, args ...string) []string {
	t.Helper()
	result := mustRun(t, dir, append([]string{"list", "--format", "json", "--sort", "id"}, args...)...)

	var tasks []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &tasks); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, result.Stdout)
	}
	ids := make([]string, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}
	return ids
}

func TestListGroup(t *testing.T) {
	dir := setupTaskDir(t)

	writeTask(t, dir, "cli/001-cli.md", "001", "CLI feature", "pending", nil)
	writeTask(t, dir, "cli/002-cli-done.md", "002", "CLI done", "completed", nil)
	writeTask(t, dir, "web/003-web.md", "003", "Web feature", "pending", nil)
	writeTask(t, dir, "web-admin/004-admin.md", "004", "Admin page", "pending", nil)
	writeTask(t, dir, "005-root.md", "005", "Root task", "pending", nil)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"exact group", []string{"--group", "cli"}, []string{"001", "002"}},
		{"wildcard", []string{"--group", "web*"}, []string{"003", "004"}},
		{"none matches root-level tasks", []string{"--group", "none"}, []string{"005"}},
		{"combined with status", []string{"--group", "cli", "--status", "pending"}, []string{"001"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listGroupIDs(t, dir, tt.args...)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("list %v = %v, want %v", tt.args, got, tt.want)
			}
		})
	}

	t.Run("matches --filter group=", func(t *testing.T) {
		viaFlag := listGroupIDs(t, dir, "--group", "cli")
		viaFilter := listGroupIDs(t, dir, "--filter", "group=cli")
		if strings.Join(viaFlag, ",") != strings.Join(viaFilter, ",") {
			t.Errorf("--group cli = %v, --filter group=cli = %v", viaFlag, viaFilter)
		}
	})
}
