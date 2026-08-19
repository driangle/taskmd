package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// addStdout runs `add <args...>` against repo, fails on error, and returns stdout.
func addStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"add"}, args...)...)
	if res.Err != nil {
		t.Fatalf("add %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestAdd_HappyPath(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := addStdout(t, repo, "My first task")

	if !strings.Contains(output, "Created task 001") {
		t.Errorf("expected 'Created task 001' in output, got: %s", output)
	}

	// Verify file was created
	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file matching 001-*.md, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	// Check frontmatter
	if !strings.Contains(fileStr, `id: "001"`) {
		t.Error("expected id: \"001\" in frontmatter")
	}
	if !strings.Contains(fileStr, `title: "My first task"`) {
		t.Error("expected title in frontmatter")
	}
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("expected status: pending in frontmatter")
	}
	if !strings.Contains(fileStr, "priority: medium") {
		t.Error("expected priority: medium in frontmatter")
	}
	if !strings.Contains(fileStr, "dependencies: []") {
		t.Error("expected dependencies: [] in frontmatter")
	}
	if !strings.Contains(fileStr, "tags: []") {
		t.Error("expected tags: [] in frontmatter")
	}
	if !strings.Contains(fileStr, "created_at: ") {
		t.Error("expected created_at date in frontmatter")
	}

	// Check body template
	if !strings.Contains(fileStr, "# My first task") {
		t.Error("expected heading in body")
	}
	if !strings.Contains(fileStr, "## Objective") {
		t.Error("expected Objective section in body")
	}
	if !strings.Contains(fileStr, "## Tasks") {
		t.Error("expected Tasks section in body")
	}
	if !strings.Contains(fileStr, "- [ ] TODO") {
		t.Error("expected TODO checkbox in body")
	}
	if !strings.Contains(fileStr, "## Acceptance Criteria") {
		t.Error("expected Acceptance Criteria section in body")
	}
}

func TestAdd_AllFlags(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Full featured task",
		"--priority", "high",
		"--effort", "large",
		"--tags", "backend,api",
		"--status", "in-progress",
		"--owner", "alice",
		"--depends-on", "001,002",
		"--parent", "010",
	)

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	if !strings.Contains(fileStr, "priority: high") {
		t.Error("expected priority: high")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("expected effort: large")
	}
	if !strings.Contains(fileStr, `tags: ["backend", "api"]`) {
		t.Error("expected tags with backend and api")
	}
	if !strings.Contains(fileStr, "status: in-progress") {
		t.Error("expected status: in-progress")
	}
	if !strings.Contains(fileStr, "owner: alice") {
		t.Error("expected owner: alice")
	}
	if !strings.Contains(fileStr, `dependencies: ["001", "002"]`) {
		t.Error("expected dependencies with 001 and 002")
	}
	if !strings.Contains(fileStr, `parent: "010"`) {
		t.Error("expected parent: \"010\"")
	}
}

func TestAdd_GroupFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := addStdout(t, repo, "CLI task", "--group", "cli")

	// Verify file created in subdirectory
	files, _ := filepath.Glob(filepath.Join(repo.Dir, "cli", "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file in cli/, got %d", len(files))
	}

	if !strings.Contains(output, filepath.Join("cli", "001-")) {
		t.Errorf("expected path with cli/ in output, got: %s", output)
	}
}

func TestAdd_JSONOutput(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := addStdout(t, repo, "JSON task", "--format", "json")

	var result addResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if result.ID != "001" {
		t.Errorf("expected id 001, got %s", result.ID)
	}
	if result.Title != "JSON task" {
		t.Errorf("expected title 'JSON task', got %s", result.Title)
	}
	if result.Status != "pending" {
		t.Errorf("expected status pending, got %s", result.Status)
	}
	if result.Priority != "medium" {
		t.Errorf("expected priority medium, got %s", result.Priority)
	}
	if result.FilePath == "" {
		t.Error("expected non-empty file_path")
	}
}

func TestAdd_InvalidPriority(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("add", "Bad priority", "--priority", "urgent")
	if res.Err == nil {
		t.Fatal("expected error for invalid priority")
	}
	if !strings.Contains(res.Err.Error(), "invalid priority") {
		t.Errorf("expected 'invalid priority' error, got: %v", res.Err)
	}
}

func TestAdd_InvalidEffort(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("add", "Bad effort", "--effort", "huge")
	if res.Err == nil {
		t.Fatal("expected error for invalid effort")
	}
	if !strings.Contains(res.Err.Error(), "invalid effort") {
		t.Errorf("expected 'invalid effort' error, got: %v", res.Err)
	}
}

func TestAdd_InvalidStatus(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("add", "Bad status", "--status", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(res.Err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %v", res.Err)
	}
}

func TestAdd_SpecialCharactersInTitle(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Fix bug: login/auth (urgent!)")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	filename := filepath.Base(files[0])
	// Slug should only contain lowercase alphanumeric and hyphens
	if strings.ContainsAny(filename, ":/(!)") {
		t.Errorf("filename should not contain special chars, got: %s", filename)
	}
}

func TestAdd_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := addStdout(t, repo, "First task ever")

	if !strings.Contains(output, "Created task 001") {
		t.Errorf("expected ID 001 for first task, got: %s", output)
	}
}

func TestAdd_SequentialIDs(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"005-existing.md": `---
id: "005"
title: "Existing task"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-02-16
---

# Existing task
`,
	})

	output := addStdout(t, repo, "Next task")

	if !strings.Contains(output, "Created task 006") {
		t.Errorf("expected ID 006 (next after 005), got: %s", output)
	}
}

func TestAdd_DependsOnParsing(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Dependent task", "--depends-on", "001, 002, 003")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	if !strings.Contains(string(content), `dependencies: ["001", "002", "003"]`) {
		t.Errorf("expected dependencies list, got:\n%s", string(content))
	}
}

func TestAdd_TagsParsing(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Tagged task", "--tags", "frontend, backend, api")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	if !strings.Contains(string(content), `tags: ["frontend", "backend", "api"]`) {
		t.Errorf("expected tags list, got:\n%s", string(content))
	}
}

func TestAdd_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("add", "Bad format", "--format", "xml")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestAdd_EditorNotSet(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Ensure EDITOR is not set
	origEditor := os.Getenv("EDITOR")
	os.Unsetenv("EDITOR")
	defer func() {
		if origEditor != "" {
			os.Setenv("EDITOR", origEditor)
		}
	}()

	res := repo.Run("add", "Edit task", "--edit")
	if res.Err == nil {
		t.Fatal("expected error when $EDITOR is not set")
	}
	if !strings.Contains(res.Err.Error(), "$EDITOR is not set") {
		t.Errorf("expected '$EDITOR is not set' error, got: %v", res.Err)
	}
}

func TestAdd_EffortOmittedWhenEmpty(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "No effort task")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	if strings.Contains(string(content), "effort:") {
		t.Error("effort should not appear in frontmatter when not set")
	}
}

func TestAdd_SuggestionOnTypo(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("add", "Typo task", "--priority", "hihg")
	if res.Err == nil {
		t.Fatal("expected error for invalid priority")
	}
	if !strings.Contains(res.Err.Error(), `did you mean "high"`) {
		t.Errorf("expected suggestion for 'high', got: %v", res.Err)
	}
}

func TestAdd_GroupCreatesDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Group task", "--group", "new-group")

	// Verify directory was created
	info, err := os.Stat(repo.Path("new-group"))
	if err != nil {
		t.Fatalf("expected group directory to be created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected new-group to be a directory")
	}
}

func TestAdd_TemplateFlag_Bug(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := addStdout(t, repo, "Login fails on Safari", "--template", "bug")

	if !strings.Contains(output, "Created task 001") {
		t.Errorf("expected 'Created task 001' in output, got: %s", output)
	}

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	// Should have template content
	if !strings.Contains(fileStr, "## Steps to Reproduce") {
		t.Error("expected bug template's Steps to Reproduce section")
	}
	if !strings.Contains(fileStr, "## Expected Behavior") {
		t.Error("expected bug template's Expected Behavior section")
	}
	if !strings.Contains(fileStr, "type: bug") {
		t.Error("expected type: bug from template")
	}
	if !strings.Contains(fileStr, "priority: high") {
		t.Error("expected priority: high from bug template")
	}

	// Should have substituted variables
	if !strings.Contains(fileStr, `id: "001"`) {
		t.Error("expected id substituted")
	}
	if !strings.Contains(fileStr, `title: "Login fails on Safari"`) {
		t.Error("expected title substituted")
	}
	if !strings.Contains(fileStr, "# Login fails on Safari") {
		t.Error("expected heading substituted")
	}

	// Should NOT have _template block
	if strings.Contains(fileStr, "_template:") {
		t.Error("_template block should have been stripped")
	}
}

func TestAdd_TemplateFlag_Feature(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Dark mode support", "--template", "feature")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	if !strings.Contains(fileStr, "## Objective") {
		t.Error("expected feature template's Objective section")
	}
	if !strings.Contains(fileStr, "type: feature") {
		t.Error("expected type: feature from template")
	}
	if !strings.Contains(fileStr, "priority: medium") {
		t.Error("expected priority: medium from feature template")
	}
}

func TestAdd_TemplateFlag_WithPriorityOverride(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Explicit --priority flag overrides the bug template's default.
	addStdout(t, repo, "Critical bug", "--template", "bug", "--priority", "critical")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	// Priority should be overridden from bug template's "high" to "critical"
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("expected priority: critical (overridden by flag)")
	}
}

func TestAdd_TemplateFlag_DefaultPriorityNotOverridden(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Do NOT set --priority flag, so it should use template's default (high)
	addStdout(t, repo, "Normal bug", "--template", "bug")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	// Should keep bug template's default priority (high), not the add command default (medium)
	if !strings.Contains(fileStr, "priority: high") {
		t.Errorf("expected priority: high from bug template, got:\n%s", fileStr)
	}
}

func TestAdd_TemplateFlag_NotFound(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("add", "No template", "--template", "nonexistent")
	if res.Err == nil {
		t.Fatal("expected error for nonexistent template")
	}
	if !strings.Contains(res.Err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", res.Err)
	}
	// Should list available templates
	if !strings.Contains(res.Err.Error(), "feature") {
		t.Errorf("expected available templates in error, got: %v", res.Err)
	}
}

func TestAdd_TemplateFlag_Chore(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Update dependencies", "--template", "chore")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	if !strings.Contains(fileStr, "type: chore") {
		t.Error("expected type: chore from template")
	}
	if !strings.Contains(fileStr, "priority: low") {
		t.Error("expected priority: low from chore template")
	}
}

func TestAdd_CustomSlug(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := addStdout(t, repo, "Fix the login bug", "--slug", "fix-login")

	if !strings.Contains(output, "Created task 001") {
		t.Errorf("expected 'Created task 001' in output, got: %s", output)
	}

	// Verify the file uses the custom filename
	expectedFile := repo.Path("001-fix-login.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", expectedFile)
	}

	content, _ := os.ReadFile(expectedFile)
	fileStr := string(content)

	// Title in frontmatter should still be the full title
	if !strings.Contains(fileStr, `title: "Fix the login bug"`) {
		t.Error("expected original title in frontmatter")
	}
}

func TestAdd_CustomSlug_NotUsedWhenEmpty(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "My great task")

	// Should fall back to slugified title
	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-my-great-task.md"))
	if len(files) != 1 {
		t.Fatalf("expected file 001-my-great-task.md, got files: %v", files)
	}
}

func TestAdd_TemplateFlag_WithTagsOverride(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Explicit --tags flag overrides the feature template's tags.
	addStdout(t, repo, "Tagged feature", "--template", "feature", "--tags", "ui,frontend")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	if !strings.Contains(fileStr, `["ui", "frontend"]`) {
		t.Errorf("expected tags override, got:\n%s", fileStr)
	}
}

func TestAdd_WithPhase(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "Phase task", "--phase", "v0.2")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	fileStr := string(content)

	if !strings.Contains(fileStr, "phase: v0.2") {
		t.Errorf("expected phase: v0.2 in frontmatter, got:\n%s", fileStr)
	}
}

func TestAdd_PhaseOmittedWhenEmpty(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "No phase task")

	files, _ := filepath.Glob(filepath.Join(repo.Dir, "001-*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	content, _ := os.ReadFile(files[0])
	if strings.Contains(string(content), "phase:") {
		t.Error("phase should not appear in frontmatter when not set")
	}
}

// TestAdd_DoesNotReuseArchivedID reproduces the reported bug: after a task is
// archived, `add` handed its ID out again because ID generation only scanned
// active tasks. Walks the full reported flow: add, add, complete, archive, add.
func TestAdd_DoesNotReuseArchivedID(t *testing.T) {
	repo := newTaskRepo(t, nil)

	addStdout(t, repo, "First task")
	addStdout(t, repo, "Second task")

	if res := repo.Run("set", "002", "--status", "completed"); res.Err != nil {
		t.Fatalf("set failed: %v", res.Err)
	}
	if res := repo.Run("archive", "--all-completed", "-y"); res.Err != nil {
		t.Fatalf("archive failed: %v", res.Err)
	}

	output := addStdout(t, repo, "Third task")

	if !strings.Contains(output, "Created task 003") {
		t.Errorf("expected 'Created task 003' (002 is archived), got: %s", output)
	}
}
