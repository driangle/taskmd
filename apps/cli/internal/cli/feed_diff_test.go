package cli

import (
	"testing"

	"github.com/driangle/taskmd/sdk/go/feed"
)

// These tests exercise pure helper functions (no CLI command, no global state),
// so every test and subtest runs with t.Parallel().

func TestExtractFrontmatterFields(t *testing.T) {
	t.Parallel()
	content := "---\nid: 042\ntitle: \"Add Auth\"\nstatus: pending\npriority: medium\n---\n# Body"
	fields := feed.ExtractFrontmatterFields(content)

	if fields["id"] != "042" {
		t.Errorf("expected id=042, got %q", fields["id"])
	}
	if fields["title"] != "Add Auth" {
		t.Errorf("expected title=Add Auth, got %q", fields["title"])
	}
	if fields["status"] != "pending" {
		t.Errorf("expected status=pending, got %q", fields["status"])
	}
	if fields["priority"] != "medium" {
		t.Errorf("expected priority=medium, got %q", fields["priority"])
	}
}

func TestExtractFrontmatterFields_NoFrontmatter(t *testing.T) {
	t.Parallel()
	fields := feed.ExtractFrontmatterFields("# Just markdown")
	if len(fields) != 0 {
		t.Errorf("expected no fields, got %d", len(fields))
	}
}

func TestExtractSubtasks(t *testing.T) {
	t.Parallel()
	content := "---\nid: 042\n---\n# Task\n\n- [ ] Add tests\n- [x] Write docs\n- [ ] Deploy\n"
	subtasks := feed.ExtractSubtasks(content)

	if len(subtasks) != 3 {
		t.Fatalf("expected 3 subtasks, got %d", len(subtasks))
	}
	if subtasks["Add tests"] != false {
		t.Error("expected 'Add tests' unchecked")
	}
	if subtasks["Write docs"] != true {
		t.Error("expected 'Write docs' checked")
	}
	if subtasks["Deploy"] != false {
		t.Error("expected 'Deploy' unchecked")
	}
}

func TestExtractSubtasks_NoFrontmatter(t *testing.T) {
	t.Parallel()
	content := "# Task\n\n- [x] Done\n- [ ] Not done\n"
	subtasks := feed.ExtractSubtasks(content)
	if len(subtasks) != 2 {
		t.Fatalf("expected 2 subtasks, got %d", len(subtasks))
	}
}

// wantFieldChange / wantSubtaskChange describe the expected AnalyzeDiff output
// for one table row, in the order AnalyzeDiff emits them.
type wantFieldChange struct {
	field, old, new string
}
type wantSubtaskChange struct {
	text string
	done bool
}

// assertFieldChanges verifies got matches want element-for-element (order included).
func assertFieldChanges(t *testing.T, got []feed.FieldChange, want []wantFieldChange) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("field changes: got %d, want %d (%v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Field != w.field || got[i].OldValue != w.old || got[i].NewValue != w.new {
			t.Errorf("field change %d: got {%q %q->%q}, want {%q %q->%q}",
				i, got[i].Field, got[i].OldValue, got[i].NewValue, w.field, w.old, w.new)
		}
	}
}

// assertSubtaskChanges verifies got matches want element-for-element (order included).
func assertSubtaskChanges(t *testing.T, got []feed.SubtaskChange, want []wantSubtaskChange) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("subtask changes: got %d, want %d (%v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Text != w.text || got[i].Done != w.done {
			t.Errorf("subtask change %d: got {%q done=%v}, want {%q done=%v}",
				i, got[i].Text, got[i].Done, w.text, w.done)
		}
	}
}

// TestAnalyzeDiff collapses the many single-scenario AnalyzeDiff_* tests into one
// table. Each row supplies old/new content and the expected field and subtask
// changes in the order AnalyzeDiff emits them (both sorted by key/text), so the
// table also guards that ordering.
func TestAnalyzeDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		old          string
		new          string
		wantFields   []wantFieldChange
		wantSubtasks []wantSubtaskChange
	}{
		{
			name:       "status change",
			old:        "---\nid: 042\nstatus: pending\npriority: medium\n---\n# Task",
			new:        "---\nid: 042\nstatus: in-progress\npriority: medium\n---\n# Task",
			wantFields: []wantFieldChange{{"status", "pending", "in-progress"}},
		},
		{
			name:       "priority change",
			old:        "---\nid: 042\nstatus: pending\npriority: medium\n---\n# Task",
			new:        "---\nid: 042\nstatus: pending\npriority: high\n---\n# Task",
			wantFields: []wantFieldChange{{"priority", "medium", "high"}},
		},
		{
			name:         "subtask completion",
			old:          "---\nid: 042\nstatus: in-progress\n---\n# Task\n\n- [ ] Add tests\n- [ ] Write docs\n",
			new:          "---\nid: 042\nstatus: in-progress\n---\n# Task\n\n- [x] Add tests\n- [ ] Write docs\n",
			wantSubtasks: []wantSubtaskChange{{"Add tests", true}},
		},
		{
			name:         "multiple subtask completions sorted by text",
			old:          "---\nid: 042\n---\n# Task\n\n- [ ] Add tests\n- [ ] Write docs\n- [ ] Deploy\n",
			new:          "---\nid: 042\n---\n# Task\n\n- [x] Add tests\n- [ ] Write docs\n- [x] Deploy\n",
			wantSubtasks: []wantSubtaskChange{{"Add tests", true}, {"Deploy", true}},
		},
		{
			name:         "mixed field and subtask changes sorted by key",
			old:          "---\nid: 042\nstatus: pending\npriority: medium\n---\n# Task\n\n- [ ] Add tests\n",
			new:          "---\nid: 042\nstatus: in-progress\npriority: high\n---\n# Task\n\n- [x] Add tests\n",
			wantFields:   []wantFieldChange{{"priority", "medium", "high"}, {"status", "pending", "in-progress"}},
			wantSubtasks: []wantSubtaskChange{{"Add tests", true}},
		},
		{
			name: "no changes",
			old:  "---\nid: 042\nstatus: pending\n---\n# Task\n\n- [ ] Add tests\n",
			new:  "---\nid: 042\nstatus: pending\n---\n# Task\n\n- [ ] Add tests\n",
		},
		{
			name: "no frontmatter",
			old:  "# Just markdown",
			new:  "# Updated markdown",
		},
		{
			name:         "subtask unchecked",
			old:          "---\nid: 042\n---\n# Task\n\n- [x] Add tests\n",
			new:          "---\nid: 042\n---\n# Task\n\n- [ ] Add tests\n",
			wantSubtasks: []wantSubtaskChange{{"Add tests", false}},
		},
		{
			name:       "new field added has empty old value",
			old:        "---\nid: 042\nstatus: pending\n---\n# Task",
			new:        "---\nid: 042\nstatus: pending\npriority: high\n---\n# Task",
			wantFields: []wantFieldChange{{"priority", "", "high"}},
		},
		{
			name:       "status to completed",
			old:        "---\nid: 042\nstatus: in-progress\n---\n# Task",
			new:        "---\nid: 042\nstatus: completed\n---\n# Task",
			wantFields: []wantFieldChange{{"status", "in-progress", "completed"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fieldChanges, subtaskChanges := feed.AnalyzeDiff(tt.old, tt.new)
			assertFieldChanges(t, fieldChanges, tt.wantFields)
			assertSubtaskChanges(t, subtaskChanges, tt.wantSubtasks)
		})
	}
}
