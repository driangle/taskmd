package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/taskfile"
)

// setTestFiles returns the canonical 3-task shape used across the set tests.
// It is kept inline rather than reusing the shared "dependency-chain" fixture
// because these tests assert on task 001 being `pending` with tags `[infra]`
// and task 003 being `blocked` — both differ from the shared fixture, and the
// assertions depend on those exact starting values.
func setTestFiles() map[string]string {
	return map[string]string{
		"001-setup.md": `---
id: "001"
title: "Setup project"
status: pending
priority: high
effort: small
dependencies: []
tags: ["infra"]
created: 2026-02-08
---

# Setup project

Initial project setup with build tooling.
`,
		"002-auth.md": `---
id: "002"
title: "Implement authentication"
status: in-progress
priority: critical
effort: large
dependencies: ["001"]
tags: ["backend", "security"]
created: 2026-02-08
---

# Implement authentication

Add JWT-based auth with refresh tokens.
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: blocked
priority: medium
effort: medium
dependencies: ["002"]
tags: ["frontend"]
created: 2026-02-08
---

# Build UI components

Create reusable component library.
`,
	}
}

// multilineTagFile returns a one-off task whose tags use the multiline YAML
// sequence form, so tests can assert that the multiline format is preserved.
func multilineTagFile() map[string]string {
	return map[string]string{
		"010-multiline.md": `---
id: "010"
title: "Multiline tags task"
status: pending
priority: high
effort: small
dependencies: []
tags:
  - backend
  - api
created: 2026-02-08
---

# Multiline tags task

Task with multiline YAML tags.
`,
	}
}

// verifySetFile returns a one-off task with an optional verify block, used by
// the --verify gating tests.
func verifySetFile(id, verifyYAML string) map[string]string {
	return map[string]string{
		id + "-verify.md": fmt.Sprintf(`---
id: "%s"
title: "Task with verify"
status: pending
created: 2026-02-14
%s---

# Task with verify
`, id, verifyYAML),
	}
}

// setStdout runs `set <args...>` against repo, fails on error, and returns stdout.
func setStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"set"}, args...)...)
	if res.Err != nil {
		t.Fatalf("set %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestSet_Status(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	// Identify the task via the --task-id flag (rather than a positional arg).
	output := setStdout(t, repo, "--task-id", "001", "--status", "completed")

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain updated status")
	}
}

func TestSet_Priority(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--priority", "low")

	if !strings.Contains(output, "priority: high -> low") {
		t.Errorf("Expected priority change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "priority: low") {
		t.Error("Expected file to contain updated priority")
	}
}

func TestSet_Effort(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "002", "--effort", "small")

	if !strings.Contains(output, "effort: large -> small") {
		t.Errorf("Expected effort change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	if !strings.Contains(string(content), "effort: small") {
		t.Error("Expected file to contain updated effort")
	}
}

func TestSet_Owner(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--owner", "alice")

	if !strings.Contains(output, "owner: (unset) -> alice") {
		t.Errorf("Expected owner change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "owner: alice") {
		t.Errorf("Expected file to contain owner: alice, got:\n%s", string(content))
	}
}

func TestSet_OwnerUpdateExisting(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"020-owned.md": `---
id: "020"
title: "Task with owner"
status: pending
owner: alice
created: 2026-02-08
---

# Task with owner
`,
	})

	output := setStdout(t, repo, "020", "--owner", "bob")

	if !strings.Contains(output, "owner: alice -> bob") {
		t.Errorf("Expected owner change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("020-owned.md"))
	if !strings.Contains(string(updated), "owner: bob") {
		t.Errorf("Expected file to contain owner: bob, got:\n%s", string(updated))
	}
}

func TestSet_DoneFlag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--done")

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected --done to set status to completed, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain completed status")
	}
}

func TestSet_MultipleFields(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "003", "--status", "in-progress", "--priority", "critical", "--effort", "large")

	if !strings.Contains(output, "status: blocked -> in-progress") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "priority: medium -> critical") {
		t.Error("Expected priority change in output")
	}
	if !strings.Contains(output, "effort: medium -> large") {
		t.Error("Expected effort change in output")
	}

	content, _ := os.ReadFile(repo.Path("003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: in-progress") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("Expected file to contain updated priority")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("Expected file to contain updated effort")
	}
}

func TestSet_AllValidStatuses(t *testing.T) {
	statuses := []string{"pending", "in-progress", "completed", "in-review", "blocked", "cancelled"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			repo := newTaskRepo(t, setTestFiles())

			res := repo.Run("set", "001", "--status", status)
			if res.Err != nil {
				t.Fatalf("unexpected error for status %q: %v", status, res.Err)
			}

			content, _ := os.ReadFile(repo.Path("001-setup.md"))
			if !strings.Contains(string(content), "status: "+status) {
				t.Errorf("Expected file to contain status: %s", status)
			}
		})
	}
}

func TestSet_CancelledStatus(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "002", "--status", "cancelled")

	if !strings.Contains(output, "status: in-progress -> cancelled") {
		t.Errorf("Expected status change to cancelled in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	if !strings.Contains(string(content), "status: cancelled") {
		t.Error("Expected file to contain status: cancelled")
	}
}

func TestSet_AllValidPriorities(t *testing.T) {
	priorities := []string{"low", "medium", "high", "critical"}
	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			repo := newTaskRepo(t, setTestFiles())

			res := repo.Run("set", "001", "--priority", priority)
			if res.Err != nil {
				t.Fatalf("unexpected error for priority %q: %v", priority, res.Err)
			}

			content, _ := os.ReadFile(repo.Path("001-setup.md"))
			if !strings.Contains(string(content), "priority: "+priority) {
				t.Errorf("Expected file to contain priority: %s", priority)
			}
		})
	}
}

func TestSet_AllValidEfforts(t *testing.T) {
	efforts := []string{"small", "medium", "large"}
	for _, effort := range efforts {
		t.Run(effort, func(t *testing.T) {
			repo := newTaskRepo(t, setTestFiles())

			res := repo.Run("set", "002", "--effort", effort)
			if res.Err != nil {
				t.Fatalf("unexpected error for effort %q: %v", effort, res.Err)
			}

			content, _ := os.ReadFile(repo.Path("002-auth.md"))
			if !strings.Contains(string(content), "effort: "+effort) {
				t.Errorf("Expected file to contain effort: %s", effort)
			}
		})
	}
}

func TestSet_InvalidStatus(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--status", "invalid")
	if res.Err == nil {
		t.Fatal("Expected error for invalid status")
	}
	if !strings.Contains(res.Err.Error(), "invalid status") {
		t.Errorf("Expected 'invalid status' error, got: %v", res.Err)
	}
}

func TestSet_InvalidPriority(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--priority", "urgent")
	if res.Err == nil {
		t.Fatal("Expected error for invalid priority")
	}
	if !strings.Contains(res.Err.Error(), "invalid priority") {
		t.Errorf("Expected 'invalid priority' error, got: %v", res.Err)
	}
}

func TestSet_InvalidEffort(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--effort", "huge")
	if res.Err == nil {
		t.Fatal("Expected error for invalid effort")
	}
	if !strings.Contains(res.Err.Error(), "invalid effort") {
		t.Errorf("Expected 'invalid effort' error, got: %v", res.Err)
	}
}

func TestSet_Type(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--type", "bug")

	if !strings.Contains(output, "type: (unset) -> bug") {
		t.Errorf("Expected type change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "type: bug") {
		t.Errorf("Expected file to contain type: bug, got:\n%s", string(content))
	}
}

func TestSet_InvalidType(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--type", "task")
	if res.Err == nil {
		t.Fatal("Expected error for invalid type")
	}
	if !strings.Contains(res.Err.Error(), "invalid type") {
		t.Errorf("Expected 'invalid type' error, got: %v", res.Err)
	}
}

func TestSet_TaskNotFound(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "nonexistent", "--status", "completed")
	if res.Err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestSet_NoFlagsProvided(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001")
	if res.Err == nil {
		t.Fatal("Expected error when no update flags provided")
	}
	if !strings.Contains(res.Err.Error(), "nothing to update") {
		t.Errorf("Expected 'nothing to update' error, got: %v", res.Err)
	}
}

func TestSet_DoneWithStatusMutuallyExclusive(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--done", "--status", "blocked")
	if res.Err == nil {
		t.Fatal("Expected error when --done and --status are both set")
	}
	if !strings.Contains(res.Err.Error(), "mutually exclusive") {
		t.Errorf("Expected 'mutually exclusive' error, got: %v", res.Err)
	}
}

func TestSet_BodyPreserved(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	setStdout(t, repo, "002", "--status", "completed")

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	fileStr := string(content)

	if !strings.Contains(fileStr, "# Implement authentication") {
		t.Error("Expected body heading to be preserved")
	}
	if !strings.Contains(fileStr, "Add JWT-based auth with refresh tokens.") {
		t.Error("Expected body content to be preserved")
	}
}

func TestSet_OtherFieldsPreserved(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	setStdout(t, repo, "002", "--status", "completed")

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	fileStr := string(content)

	// Verify non-updated fields are preserved
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("Expected priority to be preserved")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("Expected effort to be preserved")
	}
	if !strings.Contains(fileStr, `dependencies: ["001"]`) {
		t.Error("Expected dependencies to be preserved")
	}
	if !strings.Contains(fileStr, `tags: ["backend", "security"]`) {
		t.Error("Expected tags to be preserved")
	}
	if !strings.Contains(fileStr, "created: 2026-02-08") {
		t.Error("Expected created date to be preserved")
	}
}

func TestSet_FrontmatterBounds(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		wantOpen  int
		wantClose int
	}{
		{
			name:      "standard frontmatter",
			lines:     []string{"---", "id: foo", "---", "body"},
			wantOpen:  0,
			wantClose: 2,
		},
		{
			name:      "no frontmatter",
			lines:     []string{"# Just a heading", "body"},
			wantOpen:  -1,
			wantClose: -1,
		},
		{
			name:      "unclosed frontmatter",
			lines:     []string{"---", "id: foo"},
			wantOpen:  -1,
			wantClose: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, closeIdx := taskfile.FindFrontmatterBounds(tt.lines)
			if open != tt.wantOpen || closeIdx != tt.wantClose {
				t.Errorf("findFrontmatterBounds() = (%d, %d), want (%d, %d)",
					open, closeIdx, tt.wantOpen, tt.wantClose)
			}
		})
	}
}

func TestSet_MatchByTitle(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "Setup project", "--status", "completed")

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation for task found by title match")
	}
}

func TestSet_AddSingleTag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--add-tag", "new-tag")

	if !strings.Contains(output, "tags: [infra] -> [infra, new-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra", "new-tag"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_AddMultipleTags(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--add-tag", "tag-a", "--add-tag", "tag-b")

	if !strings.Contains(output, "tags: [infra] -> [infra, tag-a, tag-b]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra", "tag-a", "tag-b"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_RemoveTag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "002", "--remove-tag", "security")

	if !strings.Contains(output, "tags: [backend, security] -> [backend]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	if !strings.Contains(string(content), `tags: ["backend"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_AddAndRemoveTag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "002", "--add-tag", "new-feature", "--remove-tag", "security")

	if !strings.Contains(output, "tags: [backend, security] -> [backend, new-feature]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	if !strings.Contains(string(content), `tags: ["backend", "new-feature"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestSet_AddDuplicateTag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--add-tag", "infra")

	// Tags should remain unchanged since "infra" already exists.
	if !strings.Contains(output, "tags: [infra] -> [infra]") {
		t.Errorf("Expected no-op tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra"]`) {
		t.Errorf("Expected tags to remain unchanged, got:\n%s", string(content))
	}
}

func TestSet_RemoveNonexistentTag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--remove-tag", "nonexistent")

	// Tags should remain unchanged since "nonexistent" isn't present.
	if !strings.Contains(output, "tags: [infra] -> [infra]") {
		t.Errorf("Expected no-op tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra"]`) {
		t.Errorf("Expected tags to remain unchanged, got:\n%s", string(content))
	}
}

func TestSet_TagOnlyUpdate(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	// Should NOT produce "nothing to update" error.
	res := repo.Run("set", "001", "--add-tag", "new-tag")
	if res.Err != nil {
		t.Fatalf("tag-only update should succeed, got error: %v", res.Err)
	}
}

func TestSet_TagsWithOtherFlags(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--status", "completed", "--add-tag", "done-tag")

	if !strings.Contains(output, "status: pending -> completed") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "tags: [infra] -> [infra, done-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: completed") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, `tags: ["infra", "done-tag"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", fileStr)
	}
}

func TestSet_TagsPreservedFormat(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	setStdout(t, repo, "001", "--add-tag", "extra")

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	fileStr := string(content)

	// Inline format should stay inline.
	if !strings.Contains(fileStr, `tags: ["infra", "extra"]`) {
		t.Errorf("Expected inline tag format to be preserved, got:\n%s", fileStr)
	}

	// Other fields should be preserved.
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be preserved")
	}
	if !strings.Contains(fileStr, "# Setup project") {
		t.Error("Expected body to be preserved")
	}
}

func TestSet_MultilineTagFormat(t *testing.T) {
	repo := newTaskRepo(t, multilineTagFile())

	output := setStdout(t, repo, "010", "--add-tag", "new-tag")

	if !strings.Contains(output, "tags: [backend, api] -> [backend, api, new-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("010-multiline.md"))
	fileStr := string(content)

	// Multiline format should stay multiline.
	if !strings.Contains(fileStr, "tags:\n  - backend\n  - api\n  - new-tag") {
		t.Errorf("Expected multiline tag format to be preserved, got:\n%s", fileStr)
	}

	// Other fields should be preserved.
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be preserved")
	}
	if !strings.Contains(fileStr, "# Multiline tags task") {
		t.Error("Expected body to be preserved")
	}
}

func TestSet_MultilineTagRemove(t *testing.T) {
	repo := newTaskRepo(t, multilineTagFile())

	setStdout(t, repo, "010", "--remove-tag", "api")

	content, _ := os.ReadFile(repo.Path("010-multiline.md"))
	fileStr := string(content)

	if !strings.Contains(fileStr, "tags:\n  - backend\ncreated:") {
		t.Errorf("Expected multiline format with 'api' removed, got:\n%s", fileStr)
	}
}

func TestSet_TagConfirmationOutput(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "002", "--add-tag", "feature", "--remove-tag", "security")

	if !strings.Contains(output, "Updated task 002") {
		t.Error("Expected confirmation message with task ID")
	}
	if !strings.Contains(output, "tags: [backend, security] -> [backend, feature]") {
		t.Errorf("Expected formatted tag change, got: %s", output)
	}
}

func TestSet_Parent(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--parent", "002")

	if !strings.Contains(output, "parent: (unset) -> 002") {
		t.Errorf("Expected parent change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "parent: 002") {
		t.Errorf("Expected file to contain parent: 002, got:\n%s", string(content))
	}
}

func TestSet_ParentClear(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"030-child.md": `---
id: "030"
title: "Task with parent"
status: pending
parent: "001"
created: 2026-02-08
---

# Task with parent
`,
	})

	output := setStdout(t, repo, "030", "--parent", "")

	if !strings.Contains(output, "parent: 001 ->") {
		t.Errorf("Expected parent change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("030-child.md"))
	if strings.Contains(string(updated), "parent: 001") {
		t.Error("Expected parent to be cleared, but still found 'parent: 001'")
	}
}

func TestSet_VerifyPassThenComplete(t *testing.T) {
	repo := newTaskRepo(t, verifySetFile("050", `verify:
  - type: bash
    run: "echo pass"
`))

	output := setStdout(t, repo, "050", "--status", "completed", "--verify")

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("050-verify.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain completed status")
	}
}

func TestSet_VerifyFailAborts(t *testing.T) {
	repo := newTaskRepo(t, verifySetFile("051", `verify:
  - type: bash
    run: "exit 1"
`))

	res := repo.Run("set", "051", "--status", "completed", "--verify")
	if res.Err == nil {
		t.Fatal("expected error when verify fails")
	}
	if !strings.Contains(res.Err.Error(), "verification failed") {
		t.Errorf("expected 'verification failed' error, got: %v", res.Err)
	}

	// Status should NOT be changed
	content, _ := os.ReadFile(repo.Path("051-verify.md"))
	if strings.Contains(string(content), "status: completed") {
		t.Error("Status should not be changed when verification fails")
	}
}

func TestSet_VerifyNoFieldProceeds(t *testing.T) {
	repo := newTaskRepo(t, verifySetFile("052", ""))

	output := setStdout(t, repo, "052", "--status", "completed", "--verify")

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("expected status change, got: %s", output)
	}
}

func TestSet_VerifyNonCompletedSkips(t *testing.T) {
	repo := newTaskRepo(t, verifySetFile("053", `verify:
  - type: bash
    run: "exit 1"
`))

	output := setStdout(t, repo, "053", "--status", "in-progress", "--verify")

	// Should succeed because --verify only gates completion
	if !strings.Contains(output, "status: pending -> in-progress") {
		t.Errorf("expected status change, got: %s", output)
	}
}

func TestSet_PositionalArg(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--status", "completed")

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain updated status")
	}
}

func TestSet_PositionalArgAndFlagSameValue(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--task-id", "001", "--status", "completed")

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
}

func TestSet_PositionalArgAndFlagConflict(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--task-id", "002", "--status", "completed")
	if res.Err == nil {
		t.Fatal("Expected error when positional arg and --task-id conflict")
	}
	if !strings.Contains(res.Err.Error(), "conflicting task ID") {
		t.Errorf("Expected 'conflicting task ID' error, got: %v", res.Err)
	}
}

func TestSet_NeitherPositionalNorFlag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "--status", "completed")
	if res.Err == nil {
		t.Fatal("Expected error when neither positional arg nor --task-id provided")
	}
	if !strings.Contains(res.Err.Error(), "task ID required") {
		t.Errorf("Expected 'task ID required' error, got: %v", res.Err)
	}
}

func TestSet_InReviewStatus(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--status", "in-review")

	if !strings.Contains(output, "status: pending -> in-review") {
		t.Errorf("Expected status change to in-review, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "status: in-review") {
		t.Error("Expected file to contain status: in-review")
	}
}

func TestSet_AddPR(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--add-pr", "https://github.com/example/repo/pull/1")

	if !strings.Contains(output, "pr:") {
		t.Errorf("Expected PR change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "https://github.com/example/repo/pull/1") {
		t.Errorf("Expected file to contain PR URL, got:\n%s", string(content))
	}
}

func TestSet_RemovePR(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"040-pr.md": `---
id: "040"
title: "Task with PR"
status: in-review
pr: ["https://github.com/example/repo/pull/1", "https://github.com/example/repo/pull/2"]
created: 2026-02-08
---

# Task with PR
`,
	})

	output := setStdout(t, repo, "040", "--remove-pr", "https://github.com/example/repo/pull/1")

	if !strings.Contains(output, "pr:") {
		t.Errorf("Expected PR change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("040-pr.md"))
	fileStr := string(updated)
	if strings.Contains(fileStr, "pull/1") {
		t.Error("Expected PR 1 to be removed")
	}
	if !strings.Contains(fileStr, "pull/2") {
		t.Error("Expected PR 2 to be preserved")
	}
}

func TestSet_AddAndRemovePR(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"041-pr.md": `---
id: "041"
title: "Task with PR"
status: in-review
pr: ["https://github.com/example/repo/pull/1"]
created: 2026-02-08
---

# Task with PR
`,
	})

	output := setStdout(t, repo, "041", "--add-pr", "https://github.com/example/repo/pull/2", "--remove-pr", "https://github.com/example/repo/pull/1")

	if !strings.Contains(output, "pr:") {
		t.Errorf("Expected PR change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("041-pr.md"))
	fileStr := string(updated)
	if strings.Contains(fileStr, "pull/1") {
		t.Error("Expected PR 1 to be removed")
	}
	if !strings.Contains(fileStr, "pull/2") {
		t.Error("Expected PR 2 to be added")
	}
}

func TestSet_DoneFlag_PRReviewWorkflow(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	// Simulate pr-review workflow via viper, seeded after the harness reset.
	// Use positional arg to avoid flag state leakage.
	res := repo.RunWith(func() { viper.Set("workflow", "pr-review") }, "set", "001", "--done")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "status: pending -> in-review") {
		t.Errorf("Expected --done to set status to in-review in pr-review workflow, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	if !strings.Contains(string(content), "status: in-review") {
		t.Error("Expected file to contain in-review status")
	}
}

func TestSet_DependsOn(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "003", "--depends-on", "001,002")

	if !strings.Contains(output, "dependencies: [002] -> [001, 002]") {
		t.Errorf("Expected dependencies change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, `dependencies: ["001", "002"]`) {
		t.Errorf("Expected file to contain updated dependencies, got:\n%s", fileStr)
	}
}

func TestSet_DependsOn_InvalidID(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--depends-on", "999")
	if res.Err == nil {
		t.Fatal("Expected error for non-existent dependency ID")
	}
	if !strings.Contains(res.Err.Error(), `dependency "999" not found`) {
		t.Errorf("Expected 'not found' error, got: %v", res.Err)
	}
}

func TestSet_DependsOn_CircularDep(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	// 001 has no deps, 002 depends on 001, 003 depends on 002.
	// Setting 001 to depend on 003 creates: 001->003->002->001 (cycle).
	res := repo.Run("set", "001", "--depends-on", "003")
	if res.Err == nil {
		t.Fatal("Expected error for circular dependency")
	}
	if !strings.Contains(res.Err.Error(), "circular dependency detected") {
		t.Errorf("Expected 'circular dependency' error, got: %v", res.Err)
	}
}

func TestSet_DependsOn_SelfDep(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--depends-on", "001")
	if res.Err == nil {
		t.Fatal("Expected error for self-dependency")
	}
	if !strings.Contains(res.Err.Error(), "cannot depend on itself") {
		t.Errorf("Expected 'cannot depend on itself' error, got: %v", res.Err)
	}
}

func TestSet_DependsOn_WithOtherFlags(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "003", "--status", "in-progress", "--depends-on", "001")

	if !strings.Contains(output, "status: blocked -> in-progress") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "dependencies: [002] -> [001]") {
		t.Errorf("Expected dependencies change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: in-progress") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, `dependencies: ["001"]`) {
		t.Errorf("Expected file to contain updated dependencies, got:\n%s", fileStr)
	}
}

func TestSet_DependsOn_Clear(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "002", "--depends-on", "")

	if !strings.Contains(output, "dependencies: [001] -> []") {
		t.Errorf("Expected dependencies cleared in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("002-auth.md"))
	fileStr := string(content)
	// When clearing, the dependencies line should be removed
	if strings.Contains(fileStr, "dependencies:") {
		t.Errorf("Expected dependencies line to be removed, got:\n%s", fileStr)
	}
}

func TestComputeNewTags(t *testing.T) {
	tests := []struct {
		name       string
		current    []string
		addTags    []string
		removeTags []string
		want       []string
	}{
		{
			name:    "add to empty",
			current: nil,
			addTags: []string{"a", "b"},
			want:    []string{"a", "b"},
		},
		{
			name:       "remove from list",
			current:    []string{"a", "b", "c"},
			removeTags: []string{"b"},
			want:       []string{"a", "c"},
		},
		{
			name:       "add and remove",
			current:    []string{"a", "b"},
			addTags:    []string{"c"},
			removeTags: []string{"a"},
			want:       []string{"b", "c"},
		},
		{
			name:    "add duplicate is no-op",
			current: []string{"a", "b"},
			addTags: []string{"a"},
			want:    []string{"a", "b"},
		},
		{
			name:       "remove nonexistent is no-op",
			current:    []string{"a"},
			removeTags: []string{"z"},
			want:       []string{"a"},
		},
		{
			name:       "remove all tags",
			current:    []string{"a"},
			removeTags: []string{"a"},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := taskfile.ComputeNewTags(tt.current, tt.addTags, tt.removeTags)
			if len(got) != len(tt.want) {
				t.Fatalf("computeNewTags() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("computeNewTags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSet_AddTouches(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--add-touches", "cli/graph", "--add-touches", "cli/output")

	if !strings.Contains(output, "touches:") {
		t.Errorf("Expected touches change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "cli/graph") {
		t.Errorf("Expected file to contain cli/graph, got:\n%s", fileStr)
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Errorf("Expected file to contain cli/output, got:\n%s", fileStr)
	}
}

func TestSet_RemoveTouches(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"050-touches.md": `---
id: "050"
title: "Task with touches"
status: pending
touches: ["cli/graph", "cli/output", "web/board"]
created: 2026-02-08
---

# Task with touches
`,
	})

	output := setStdout(t, repo, "050", "--remove-touches", "cli/graph")

	if !strings.Contains(output, "touches:") {
		t.Errorf("Expected touches change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("050-touches.md"))
	fileStr := string(updated)
	if strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be removed")
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Error("Expected cli/output to be preserved")
	}
	if !strings.Contains(fileStr, "web/board") {
		t.Error("Expected web/board to be preserved")
	}
}

func TestSet_AddTouches_Deduplication(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"051-touches-dedup.md": `---
id: "051"
title: "Task with existing touches"
status: pending
touches: ["cli/graph"]
created: 2026-02-08
---

# Task with existing touches
`,
	})

	res := repo.Run("set", "051", "--add-touches", "cli/graph", "--add-touches", "cli/output")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	updated, _ := os.ReadFile(repo.Path("051-touches-dedup.md"))
	fileStr := string(updated)
	// cli/graph should appear exactly once (not duplicated)
	count := strings.Count(fileStr, "cli/graph")
	if count != 1 {
		t.Errorf("Expected cli/graph to appear once, appeared %d times in:\n%s", count, fileStr)
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Error("Expected cli/output to be added")
	}
}

func TestSet_RemoveTouches_NonExistent(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"052-touches.md": `---
id: "052"
title: "Task with touches"
status: pending
touches: ["cli/graph"]
created: 2026-02-08
---

# Task with touches
`,
	})

	res := repo.Run("set", "052", "--remove-touches", "nonexistent/scope")
	if res.Err != nil {
		t.Fatalf("unexpected error removing non-existent touches value: %v", res.Err)
	}

	updated, _ := os.ReadFile(repo.Path("052-touches.md"))
	fileStr := string(updated)
	if !strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be preserved")
	}
}

func TestSet_Touches_DryRun(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"053-touches-dry.md": `---
id: "053"
title: "Task for dry run"
status: pending
touches: ["cli/graph"]
created: 2026-02-08
---

# Task for dry run
`,
	})

	output := setStdout(t, repo, "053", "--add-touches", "cli/output", "--dry-run")

	if !strings.Contains(output, "Dry run") {
		t.Errorf("Expected dry run message, got: %s", output)
	}

	// File should be unchanged
	updated, _ := os.ReadFile(repo.Path("053-touches-dry.md"))
	fileStr := string(updated)
	if strings.Contains(fileStr, "cli/output") {
		t.Error("Expected file to be unchanged in dry run mode")
	}
}

func TestSet_AddAndRemoveTouches(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"054-touches.md": `---
id: "054"
title: "Task with touches"
status: pending
touches: ["cli/graph", "cli/set"]
created: 2026-02-08
---

# Task with touches
`,
	})

	res := repo.Run("set", "054", "--add-touches", "cli/output", "--remove-touches", "cli/graph")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	updated, _ := os.ReadFile(repo.Path("054-touches.md"))
	fileStr := string(updated)
	if strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be removed")
	}
	if !strings.Contains(fileStr, "cli/set") {
		t.Error("Expected cli/set to be preserved")
	}
	if !strings.Contains(fileStr, "cli/output") {
		t.Error("Expected cli/output to be added")
	}
}

func TestSet_AddTouches_EmptyArray(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	res := repo.Run("set", "001", "--add-touches", "cli/graph")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "touches:") {
		t.Error("Expected touches field to be added")
	}
	if !strings.Contains(fileStr, "cli/graph") {
		t.Error("Expected cli/graph to be added")
	}
}

func TestSet_Phase_Add(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	output := setStdout(t, repo, "001", "--phase", "v0.2")

	if !strings.Contains(output, "phase:") {
		t.Errorf("Expected phase change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "phase: v0.2") {
		t.Errorf("Expected file to contain phase: v0.2, got:\n%s", fileStr)
	}
}

func TestSet_Phase_Change(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"060-phase.md": `---
id: "060"
title: "Task with phase"
status: pending
phase: v0.1
created: 2026-02-08
---

# Task with phase
`,
	})

	output := setStdout(t, repo, "060", "--phase", "v0.2")

	if !strings.Contains(output, "phase: v0.1 -> v0.2") {
		t.Errorf("Expected phase change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("060-phase.md"))
	fileStr := string(updated)
	if !strings.Contains(fileStr, "phase: v0.2") {
		t.Errorf("Expected phase to be updated, got:\n%s", fileStr)
	}
}

func TestSet_Phase_Clear(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"061-phase.md": `---
id: "061"
title: "Task with phase"
status: pending
phase: v0.1
created: 2026-02-08
---

# Task with phase
`,
	})

	output := setStdout(t, repo, "061", "--phase", "")

	if !strings.Contains(output, "phase:") {
		t.Errorf("Expected phase change in output, got: %s", output)
	}

	updated, _ := os.ReadFile(repo.Path("061-phase.md"))
	fileStr := string(updated)
	if !strings.Contains(fileStr, "phase: ") || strings.Contains(fileStr, "phase: v0.1") {
		t.Errorf("Expected phase to be cleared, got:\n%s", fileStr)
	}
}

func TestSet_CompletedDate_AutoSetOnCompleted(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	setStdout(t, repo, "001", "--status", "completed")

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("Expected status to be completed")
	}
	if !strings.Contains(s, "completed_at: ") {
		t.Errorf("Expected completed_at date to be auto-set, got:\n%s", s)
	}
}

func TestSet_CancelledDate_AutoSetOnCancelled(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	setStdout(t, repo, "001", "--status", "cancelled")

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	s := string(content)
	if !strings.Contains(s, "status: cancelled") {
		t.Error("Expected status to be cancelled")
	}
	if !strings.Contains(s, "cancelled_at: ") {
		t.Errorf("Expected cancelled_at date to be auto-set, got:\n%s", s)
	}
	if strings.Contains(s, "completed_at:") {
		t.Error("Expected no completed_at field when cancelling")
	}
}

func TestSet_CompletedDate_ClearedOnReopen(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"070-completed.md": `---
id: "070"
title: "Completed task"
status: completed
priority: medium
completed_at: 2026-03-01
created: 2026-02-08
---

# Completed task
`,
	})

	output := setStdout(t, repo, "070", "--status", "pending")

	if !strings.Contains(output, "completed_at:") {
		t.Errorf("Expected completed_at change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("070-completed.md"))
	s := string(content)
	if !strings.Contains(s, "status: pending") {
		t.Error("Expected status to be pending")
	}
	if strings.Contains(s, "completed_at:") {
		t.Errorf("Expected completed_at field to be removed, got:\n%s", s)
	}
}

func TestSet_CompletedDate_DoneFlag(t *testing.T) {
	repo := newTaskRepo(t, setTestFiles())

	setStdout(t, repo, "001", "--done")

	content, _ := os.ReadFile(repo.Path("001-setup.md"))
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("Expected status to be completed via --done")
	}
	if !strings.Contains(s, "completed_at: ") {
		t.Errorf("Expected completed_at date to be auto-set via --done, got:\n%s", s)
	}
}

func TestSet_CompletedDate_NonStatusChangePreserves(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"071-completed.md": `---
id: "071"
title: "Completed task"
status: completed
priority: medium
completed_at: 2026-03-01
created: 2026-02-08
---

# Completed task
`,
	})

	setStdout(t, repo, "071", "--priority", "high")

	content, _ := os.ReadFile(repo.Path("071-completed.md"))
	s := string(content)
	if !strings.Contains(s, "priority: high") {
		t.Error("Expected priority to be updated")
	}
	if !strings.Contains(s, "completed_at: 2026-03-01") {
		t.Errorf("Expected completed_at date to be preserved when not changing status, got:\n%s", s)
	}
}

func TestSet_CancelledDate_ClearedOnReopen(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"072-cancelled.md": `---
id: "072"
title: "Cancelled task"
status: cancelled
priority: medium
cancelled_at: 2026-03-01
created: 2026-02-08
---

# Cancelled task
`,
	})

	output := setStdout(t, repo, "072", "--status", "pending")

	if !strings.Contains(output, "cancelled_at:") {
		t.Errorf("Expected cancelled_at change in output, got: %s", output)
	}

	content, _ := os.ReadFile(repo.Path("072-cancelled.md"))
	s := string(content)
	if !strings.Contains(s, "status: pending") {
		t.Error("Expected status to be pending")
	}
	if strings.Contains(s, "cancelled_at:") {
		t.Errorf("Expected cancelled_at field to be removed, got:\n%s", s)
	}
}

func TestSet_CompletedDate_ClearsOnCancel(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"073-task.md": `---
id: "073"
title: "Completed then cancelled"
status: completed
priority: medium
completed_at: 2026-03-01
created: 2026-02-08
---

# Completed then cancelled
`,
	})

	setStdout(t, repo, "073", "--status", "cancelled")

	content, _ := os.ReadFile(repo.Path("073-task.md"))
	s := string(content)
	if !strings.Contains(s, "cancelled_at: ") {
		t.Errorf("Expected cancelled_at to be set, got:\n%s", s)
	}
	if strings.Contains(s, "completed_at:") {
		t.Errorf("Expected completed_at to be cleared when cancelling, got:\n%s", s)
	}
}
