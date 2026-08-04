package cli

import (
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

// commitMsgFiles is the recurring commit-msg fixture: tasks with subtasks,
// groups, and mixed statuses. It is command-specific (not one of the canonical
// shared sets), so it stays inline.
func commitMsgFiles() map[string]string {
	return map[string]string{
		"cli/042-feature.md": `---
id: "042"
title: "Add commit-msg command"
status: completed
priority: medium
effort: medium
tags: [cli, git]
group: cli
created: 2026-02-16
---

# Add commit-msg command

- [x] Create command scaffolding
- [x] Implement message generation
- [ ] Add documentation
`,
		"cli/043-bugfix.md": `---
id: "043"
title: "Fix output formatting"
status: completed
priority: high
effort: small
tags: [cli, bug]
group: cli
created: 2026-02-16
---

# Fix output formatting

- [x] Fix newline handling
- [ ] Update tests
`,
		"044-no-group.md": `---
id: "044"
title: "Update README"
status: pending
priority: low
effort: small
tags: [docs]
created: 2026-02-16
---

# Update README

- [x] Add installation instructions
- [x] Add usage examples
`,
		"cli/045-new-pending.md": `---
id: "045"
title: "Add search feature"
status: pending
priority: medium
effort: medium
tags: [cli]
group: cli
created: 2026-02-17
---

# Add search feature
`,
		"cli/046-new-pending.md": `---
id: "046"
title: "Add filter feature"
status: pending
priority: medium
effort: small
tags: [cli]
group: cli
created: 2026-02-17
---

# Add filter feature
`,
		"cli/047-in-progress.md": `---
id: "047"
title: "Implement export"
status: in-progress
priority: medium
effort: medium
tags: [cli]
group: cli
created: 2026-02-17
---

# Implement export
`,
		"cli/048-blocked.md": `---
id: "048"
title: "Add import feature"
status: blocked
priority: low
effort: medium
tags: [cli]
group: cli
created: 2026-02-17
---

# Add import feature
`,
	}
}

// newCommitMsgRepo seeds a repo with the shared commit-msg fixture set.
func newCommitMsgRepo(t *testing.T) *taskRepo {
	t.Helper()
	return newTaskRepo(t, commitMsgFiles())
}

// commitMsgStdout runs `commit-msg <args...>`, fails on error, and returns stdout.
func commitMsgStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"commit-msg"}, args...)...)
	if res.Err != nil {
		t.Fatalf("commit-msg %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

// mockGitDiffAndRoot sets up gitDiffFunc and gitRootFunc for tests.
// diffOutput is the fake diff content. gitRoot is the fake git root directory.
// Registers a cleanup to restore the originals.
func mockGitDiffAndRoot(t *testing.T, diffOutput string, gitRoot string) {
	t.Helper()
	oldDiff := gitDiffFunc
	oldRoot := gitRootFunc
	gitDiffFunc = func(_ string) (string, error) { return diffOutput, nil }
	gitRootFunc = func(_ string) (string, error) { return gitRoot, nil }
	t.Cleanup(func() {
		gitDiffFunc = oldDiff
		gitRootFunc = oldRoot
	})
}

func TestCommitMsg_SingleTask(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "042")

	if !strings.Contains(output, "chore(cli): add commit-msg command (task 042)") {
		t.Errorf("expected subject line with task ID, got:\n%s", output)
	}
}

func TestCommitMsg_TypeFlag(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "042", "--type", "feat")

	if !strings.HasPrefix(output, "feat(cli): add commit-msg command (task 042)") {
		t.Errorf("expected 'feat' prefix with task ID, got:\n%s", output)
	}
}

func TestCommitMsg_InvalidType(t *testing.T) {
	repo := newCommitMsgRepo(t)

	res := repo.Run("commit-msg", "--task-id", "042", "--type", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(res.Err.Error(), "invalid commit type") {
		t.Errorf("expected 'invalid commit type' error, got: %v", res.Err)
	}
}

func TestCommitMsg_BodyFlag(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "042", "--body")

	if !strings.Contains(output, "- Create command scaffolding") {
		t.Errorf("expected completed subtask in body, got:\n%s", output)
	}
	if !strings.Contains(output, "- Implement message generation") {
		t.Errorf("expected second completed subtask in body, got:\n%s", output)
	}
	// Unchecked subtask should NOT be included
	if strings.Contains(output, "Add documentation") {
		t.Errorf("unchecked subtask should not appear in body, got:\n%s", output)
	}
}

func TestCommitMsg_BodyFlagNoSubtasks(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"099-empty.md": `---
id: "099"
title: "Empty task"
status: pending
created: 2026-02-16
---

# Empty task

No subtasks here.
`,
	})

	output := commitMsgStdout(t, repo, "--task-id", "099", "--body")

	// Should still produce a valid message with subject line only (no body section)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least a subject line")
	}
	if !strings.Contains(lines[0], "(task 099)") {
		t.Errorf("expected task ID in subject line, got: %s", lines[0])
	}
}

func TestCommitMsg_ShortFlag(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "042", "--short")

	trimmed := strings.TrimSpace(output)
	if strings.Contains(trimmed, "\n") {
		t.Errorf("--short should produce single line, got:\n%s", output)
	}
	if !strings.HasPrefix(trimmed, "chore(cli):") || !strings.Contains(trimmed, "(task 042)") {
		t.Errorf("expected subject line with task ID, got: %s", trimmed)
	}
}

func TestCommitMsg_ShortWithBody(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "042", "--short", "--body")

	// --short should override --body: single line only
	trimmed := strings.TrimSpace(output)
	if strings.Contains(trimmed, "\n") {
		t.Errorf("--short should suppress body even when --body is set, got:\n%s", output)
	}
}

func TestCommitMsg_NoGroupScope(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "044")

	if !strings.HasPrefix(output, "chore: update README (task 044)") {
		t.Errorf("expected no scope for task without group, got:\n%s", output)
	}
}

func TestCommitMsg_TaskNotFound(t *testing.T) {
	repo := newCommitMsgRepo(t)

	res := repo.Run("commit-msg", "--task-id", "999")
	if res.Err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("expected 'task not found' error, got: %v", res.Err)
	}
}

func TestCommitMsg_MultiTask(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"diff --git a/cli/042-feature.md b/cli/042-feature.md\n"+
			"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"@@ -4 +4 @@\n"+
			"-status: pending\n"+
			"+status: completed\n"+
			"diff --git a/cli/043-bugfix.md b/cli/043-bugfix.md\n"+
			"--- a/cli/043-bugfix.md\n"+
			"+++ b/cli/043-bugfix.md\n"+
			"@@ -4 +4 @@\n"+
			"-status: pending\n"+
			"+status: completed\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	if !strings.Contains(output, "chore: complete tasks 042, 043") {
		t.Errorf("expected multi-task subject, got:\n%s", output)
	}
}

func TestCommitMsg_MultiTaskWithBody(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"diff --git a/cli/042-feature.md b/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"diff --git a/cli/043-bugfix.md b/cli/043-bugfix.md\n"+
			"+++ b/cli/043-bugfix.md\n"+
			"+status: completed\n",
		repo.Dir)

	output := commitMsgStdout(t, repo, "--body")

	if !strings.Contains(output, "Add commit-msg command:") {
		t.Errorf("expected task title section in body, got:\n%s", output)
	}
	if !strings.Contains(output, "- Create command scaffolding") {
		t.Errorf("expected subtask from first task, got:\n%s", output)
	}
	if !strings.Contains(output, "Fix output formatting:") {
		t.Errorf("expected second task title section, got:\n%s", output)
	}
}

func TestCommitMsg_MultiTaskShort(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"+++ b/cli/042-feature.md\n+status: completed\n"+
			"+++ b/cli/043-bugfix.md\n+status: completed\n",
		repo.Dir)

	output := commitMsgStdout(t, repo, "--short")

	trimmed := strings.TrimSpace(output)
	if strings.Contains(trimmed, "\n") {
		t.Errorf("--short should produce single line, got:\n%s", output)
	}
}

func TestCommitMsg_AutoInferFromDiff(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"diff --git a/cli/042-feature.md b/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"@@ -4 +4 @@\n"+
			"-status: pending\n"+
			"+status: completed\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	if !strings.Contains(output, "chore(cli): add commit-msg command (task 042)") {
		t.Errorf("expected auto-detected task message with ID in header, got:\n%s", output)
	}
}

func TestCommitMsg_EmptyDiff(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t, "", repo.Dir)

	res := repo.Run("commit-msg")
	if res.Err == nil {
		t.Fatal("expected error for empty diff")
	}
	if !strings.Contains(res.Err.Error(), "no task changes found") {
		t.Errorf("expected 'no task changes found' error, got: %v", res.Err)
	}
}

func TestCommitMsg_DiffOnlyInProgress(t *testing.T) {
	repo := newCommitMsgRepo(t)

	// in-progress status change should now be detected as a "started" task
	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: in-progress\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	if !strings.Contains(output, "start task 042") {
		t.Errorf("expected 'start task' message, got:\n%s", output)
	}
}

// Unit tests for helper functions

func TestExtractCompletedSubtasks(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "mixed checked and unchecked",
			body: "- [x] Done one\n- [ ] Not done\n- [x] Done two\n",
			want: []string{"Done one", "Done two"},
		},
		{
			name: "uppercase X",
			body: "- [X] Done with uppercase\n",
			want: []string{"Done with uppercase"},
		},
		{
			name: "no subtasks",
			body: "Just some text\n",
			want: nil,
		},
		{
			name: "empty body",
			body: "",
			want: nil,
		},
		{
			name: "indented checkboxes",
			body: "  - [x] Indented item\n",
			want: []string{"Indented item"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCompletedSubtasks(tt.body)
			if len(got) != len(tt.want) {
				t.Fatalf("extractCompletedSubtasks() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractCompletedSubtasks()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseDiffResult(t *testing.T) {
	tests := []struct {
		name      string
		diff      string
		completed []string
		added     []string
		started   []string
		blocked   []string
		cancelled []string
	}{
		{
			name:      "single completed file",
			diff:      "+++ b/tasks/cli/042.md\n@@ -4 +4 @@\n-status: pending\n+status: completed\n",
			completed: []string{"tasks/cli/042.md"},
		},
		{
			name:      "multiple completed files",
			diff:      "+++ b/a.md\n+status: completed\n+++ b/b.md\n+status: completed\n",
			completed: []string{"a.md", "b.md"},
		},
		{
			name:    "in-progress status change",
			diff:    "--- a/a.md\n+++ b/a.md\n+status: in-progress\n",
			started: []string{"a.md"},
		},
		{
			name:  "new pending file",
			diff:  "--- /dev/null\n+++ b/tasks/cli/045.md\n+status: pending\n",
			added: []string{"tasks/cli/045.md"},
		},
		{
			name: "modified pending file ignored",
			diff: "--- a/tasks/cli/045.md\n+++ b/tasks/cli/045.md\n+status: pending\n",
		},
		{
			name:    "blocked status change",
			diff:    "--- a/a.md\n+++ b/a.md\n+status: blocked\n",
			blocked: []string{"a.md"},
		},
		{
			name:      "cancelled status change",
			diff:      "--- a/a.md\n+++ b/a.md\n+status: cancelled\n",
			cancelled: []string{"a.md"},
		},
		{
			name: "empty diff",
			diff: "",
		},
		{
			name: "no status line",
			diff: "+++ b/a.md\n+title: something\n",
		},
		{
			name:  "mixed new and modified files",
			diff:  "--- /dev/null\n+++ b/new.md\n+status: pending\n--- a/old.md\n+++ b/old.md\n+status: pending\n",
			added: []string{"new.md"},
		},
		{
			name:      "multiple categories in one diff",
			diff:      "--- a/a.md\n+++ b/a.md\n+status: completed\n--- /dev/null\n+++ b/b.md\n+status: pending\n--- a/c.md\n+++ b/c.md\n+status: in-progress\n",
			completed: []string{"a.md"},
			added:     []string{"b.md"},
			started:   []string{"c.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDiffResult(tt.diff)
			assertStringSlice(t, "Completed", got.Completed, tt.completed)
			assertStringSlice(t, "Added", got.Added, tt.added)
			assertStringSlice(t, "Started", got.Started, tt.started)
			assertStringSlice(t, "Blocked", got.Blocked, tt.blocked)
			assertStringSlice(t, "Cancelled", got.Cancelled, tt.cancelled)
		})
	}
}

func assertStringSlice(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %v, want %v", label, got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %q, want %q", label, i, got[i], want[i])
		}
	}
}

func TestLowerFirst(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Add feature", "add feature"},
		{"fix bug", "fix bug"},
		{"", ""},
		{"A", "a"},
		{"already lower", "already lower"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := lowerFirst(tt.input)
			if got != tt.want {
				t.Errorf("lowerFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCommitMsg_AllowedTypes(t *testing.T) {
	types := []string{"feat", "fix", "chore", "docs", "test", "refactor"}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			repo := newCommitMsgRepo(t)

			output := commitMsgStdout(t, repo, "--task-id", "042", "--type", typ)

			if !strings.HasPrefix(output, typ+"(cli):") {
				t.Errorf("expected prefix %q, got:\n%s", typ+"(cli):", output)
			}
			if !strings.Contains(output, "(task 042)") {
				t.Errorf("expected task ID in header, got:\n%s", output)
			}
		})
	}
}

func TestCommitMsg_MessageFormat(t *testing.T) {
	repo := newCommitMsgRepo(t)

	output := commitMsgStdout(t, repo, "--task-id", "042", "--type", "feat", "--body")

	// Verify format: subject, blank line, body
	parts := strings.Split(strings.TrimSpace(output), "\n\n")
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts (subject, body), got %d: %v", len(parts), parts)
	}

	subject := parts[0]
	if subject != "feat(cli): add commit-msg command (task 042)" {
		t.Errorf("unexpected subject: %s", subject)
	}

	body := parts[1]
	if !strings.Contains(body, "- Create command scaffolding") {
		t.Errorf("unexpected body: %s", body)
	}
}

// Tests for new pending task detection

func TestCommitMsg_SingleNewPendingTask(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"diff --git a/cli/045-new-pending.md b/cli/045-new-pending.md\n"+
			"new file mode 100644\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"@@ -0,0 +1,10 @@\n"+
			"+---\n"+
			"+id: \"045\"\n"+
			"+title: \"Add search feature\"\n"+
			"+status: pending\n"+
			"+---\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	expected := "chore: added task 045\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestCommitMsg_MultipleNewPendingTasks(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"diff --git a/cli/045-new-pending.md b/cli/045-new-pending.md\n"+
			"new file mode 100644\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"@@ -0,0 +1,10 @@\n"+
			"+status: pending\n"+
			"diff --git a/cli/046-new-pending.md b/cli/046-new-pending.md\n"+
			"new file mode 100644\n"+
			"--- /dev/null\n"+
			"+++ b/cli/046-new-pending.md\n"+
			"@@ -0,0 +1,10 @@\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	expected := "chore: added tasks 045, 046\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestCommitMsg_NewPendingWithTypeFlag(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"diff --git a/cli/045-new-pending.md b/cli/045-new-pending.md\n"+
			"new file mode 100644\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo, "--type", "feat")

	expected := "feat: added task 045\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestCommitMsg_MixedCompletedAndNewPending(t *testing.T) {
	repo := newCommitMsgRepo(t)

	// Diff contains both a completed task and a new pending task
	mockGitDiffAndRoot(t,
		"diff --git a/cli/042-feature.md b/cli/042-feature.md\n"+
			"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"@@ -4 +4 @@\n"+
			"-status: pending\n"+
			"+status: completed\n"+
			"diff --git a/cli/045-new-pending.md b/cli/045-new-pending.md\n"+
			"new file mode 100644\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	// Should produce a combined message with both change types
	if !strings.Contains(output, "complete task 042") {
		t.Errorf("expected 'complete task 042' in mixed message, got:\n%s", output)
	}
	if !strings.Contains(output, "add task 045") {
		t.Errorf("expected 'add task 045' in mixed message, got:\n%s", output)
	}
	if !strings.Contains(output, ";") {
		t.Errorf("expected semicolon separator in mixed message, got:\n%s", output)
	}
}

func TestCommitMsg_ModifiedPendingNotDetected(t *testing.T) {
	repo := newCommitMsgRepo(t)

	// File was modified (not new) with status: pending — should NOT be detected
	mockGitDiffAndRoot(t,
		"diff --git a/cli/045-new-pending.md b/cli/045-new-pending.md\n"+
			"--- a/cli/045-new-pending.md\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"@@ -4 +4 @@\n"+
			"+status: pending\n",
		repo.Dir)

	res := repo.Run("commit-msg")
	if res.Err == nil {
		t.Fatal("expected error when only modified (not new) pending file staged")
	}
	if !strings.Contains(res.Err.Error(), "no task changes found") {
		t.Errorf("expected 'no task changes found' error, got: %v", res.Err)
	}
}

func TestBuildAddedTaskMessage(t *testing.T) {
	tests := []struct {
		name       string
		taskIDs    []string
		commitType string
		want       string
	}{
		{
			name:       "single task",
			taskIDs:    []string{"045"},
			commitType: "chore",
			want:       "chore: added task 045\n",
		},
		{
			name:       "multiple tasks",
			taskIDs:    []string{"045", "046"},
			commitType: "chore",
			want:       "chore: added tasks 045, 046\n",
		},
		{
			name:       "custom type",
			taskIDs:    []string{"045"},
			commitType: "feat",
			want:       "feat: added task 045\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tasks []*model.Task
			for _, id := range tt.taskIDs {
				tasks = append(tasks, &model.Task{ID: id})
			}
			got := buildAddedTaskMessage(tasks, tt.commitType)
			if got != tt.want {
				t.Errorf("buildAddedTaskMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Tests for mixed change type detection

func TestCommitMsg_MixedCompletedAndStarted(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"--- a/cli/047-in-progress.md\n"+
			"+++ b/cli/047-in-progress.md\n"+
			"+status: in-progress\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	if !strings.Contains(output, "complete task 042") {
		t.Errorf("expected 'complete task 042', got:\n%s", output)
	}
	if !strings.Contains(output, "start task 047") {
		t.Errorf("expected 'start task 047', got:\n%s", output)
	}
	if !strings.Contains(output, ";") {
		t.Errorf("expected semicolon separator, got:\n%s", output)
	}
}

func TestCommitMsg_ThreeChangeTypes(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n"+
			"--- a/cli/047-in-progress.md\n"+
			"+++ b/cli/047-in-progress.md\n"+
			"+status: in-progress\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	expected := "chore: complete task 042; add task 045; start task 047\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestCommitMsg_MixedWithShort(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo, "--short")

	trimmed := strings.TrimSpace(output)
	if strings.Contains(trimmed, "\n") {
		t.Errorf("--short should produce single line, got:\n%s", output)
	}
	if !strings.Contains(trimmed, "complete task 042") || !strings.Contains(trimmed, "add task 045") {
		t.Errorf("expected both change types in short output, got: %s", trimmed)
	}
}

func TestCommitMsg_MixedWithBody(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo, "--body")

	// Subject line should contain both change types
	lines := strings.Split(output, "\n")
	if !strings.Contains(lines[0], "complete task 042") || !strings.Contains(lines[0], "add task 045") {
		t.Errorf("expected both change types in subject, got: %s", lines[0])
	}

	// Body should contain category sections
	if !strings.Contains(output, "Completed:") {
		t.Errorf("expected 'Completed:' section in body, got:\n%s", output)
	}
	if !strings.Contains(output, "Added:") {
		t.Errorf("expected 'Added:' section in body, got:\n%s", output)
	}
	// Completed task has subtasks, so they should appear
	if !strings.Contains(output, "Create command scaffolding") {
		t.Errorf("expected completed subtasks in body, got:\n%s", output)
	}
}

func TestCommitMsg_MixedWithTypeFlag(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo, "--type", "feat")

	if !strings.HasPrefix(output, "feat:") {
		t.Errorf("expected 'feat:' prefix, got:\n%s", output)
	}
}

func TestCommitMsg_MixedMultipleTasksPerCategory(t *testing.T) {
	repo := newCommitMsgRepo(t)

	mockGitDiffAndRoot(t,
		"--- a/cli/042-feature.md\n"+
			"+++ b/cli/042-feature.md\n"+
			"+status: completed\n"+
			"--- a/cli/043-bugfix.md\n"+
			"+++ b/cli/043-bugfix.md\n"+
			"+status: completed\n"+
			"--- /dev/null\n"+
			"+++ b/cli/045-new-pending.md\n"+
			"+status: pending\n"+
			"--- /dev/null\n"+
			"+++ b/cli/046-new-pending.md\n"+
			"+status: pending\n",
		repo.Dir)

	output := commitMsgStdout(t, repo)

	expected := "chore: complete tasks 042, 043; add tasks 045, 046\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestBuildMixedCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		changes TaskChanges
		short   bool
		body    bool
		want    string
	}{
		{
			name: "completed and added",
			changes: TaskChanges{
				Completed: []*model.Task{{ID: "042"}},
				Added:     []*model.Task{{ID: "045"}},
			},
			want: "chore: complete task 042; add task 045\n",
		},
		{
			name: "all five categories",
			changes: TaskChanges{
				Completed: []*model.Task{{ID: "001"}},
				Added:     []*model.Task{{ID: "002"}},
				Started:   []*model.Task{{ID: "003"}},
				Blocked:   []*model.Task{{ID: "004"}},
				Cancelled: []*model.Task{{ID: "005"}},
			},
			want: "chore: complete task 001; add task 002; start task 003; block task 004; cancel task 005\n",
		},
		{
			name: "multiple per category",
			changes: TaskChanges{
				Completed: []*model.Task{{ID: "001"}, {ID: "002"}},
				Added:     []*model.Task{{ID: "003"}},
			},
			want: "chore: complete tasks 001, 002; add task 003\n",
		},
		{
			name: "short flag",
			changes: TaskChanges{
				Completed: []*model.Task{{ID: "001"}},
				Added:     []*model.Task{{ID: "002"}},
			},
			short: true,
			want:  "chore: complete task 001; add task 002\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMixedCommitMessage(tt.changes, "chore", tt.body, tt.short)
			if got != tt.want {
				t.Errorf("buildMixedCommitMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChangeSegment(t *testing.T) {
	tests := []struct {
		verb  string
		tasks []*model.Task
		want  string
	}{
		{"complete", []*model.Task{{ID: "042"}}, "complete task 042"},
		{"complete", []*model.Task{{ID: "042"}, {ID: "043"}}, "complete tasks 042, 043"},
		{"add", []*model.Task{{ID: "045"}}, "add task 045"},
		{"start", []*model.Task{{ID: "047"}, {ID: "048"}}, "start tasks 047, 048"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := changeSegment(tt.verb, tt.tasks)
			if got != tt.want {
				t.Errorf("changeSegment() = %q, want %q", got, tt.want)
			}
		})
	}
}
