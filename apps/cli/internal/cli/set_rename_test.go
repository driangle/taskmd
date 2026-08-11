package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

// renameTestFiles returns a task whose body heading matches its title, which is
// the shape `taskmd add` produces and the one --title is designed around.
func renameTestFiles() map[string]string {
	return map[string]string{
		"001-setup-project.md": `---
id: "001"
title: "Setup project"
status: pending
priority: high
effort: small
tags: ["infra"]
created: 2026-02-08
---

# Setup project

Initial project setup with build tooling.
`,
	}
}

// readTask reads a repo-relative task file, failing the test if it is missing.
func readTask(t *testing.T, repo *taskRepo, name string) string {
	t.Helper()
	content, err := os.ReadFile(repo.Path(name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return string(content)
}

// assertMissing fails if a repo-relative file exists.
func assertMissing(t *testing.T, repo *taskRepo, name string) {
	t.Helper()
	if _, err := os.Stat(repo.Path(name)); err == nil {
		t.Errorf("expected %s not to exist", name)
	}
}

func TestSet_Title(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	output := setStdout(t, repo, "001", "--title", "Bootstrap the project")

	if !strings.Contains(output, "title: Setup project -> Bootstrap the project") {
		t.Errorf("Expected title change in output, got: %s", output)
	}

	content := readTask(t, repo, "001-setup-project.md")
	if !strings.Contains(content, `title: "Bootstrap the project"`) {
		t.Errorf("Expected updated frontmatter title, got:\n%s", content)
	}
	if !strings.Contains(content, "# Bootstrap the project") {
		t.Errorf("Expected updated body heading, got:\n%s", content)
	}
	if !strings.Contains(content, "Initial project setup with build tooling.") {
		t.Errorf("Expected body prose to be preserved, got:\n%s", content)
	}
}

// Without --rename the file must stay put, even though its slug is now stale.
func TestSet_Title_DoesNotRenameFileByDefault(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	output := setStdout(t, repo, "001", "--title", "Bootstrap the project")

	if strings.Contains(output, "file:") {
		t.Errorf("Expected no file change reported, got: %s", output)
	}
	if _, err := os.Stat(repo.Path("001-setup-project.md")); err != nil {
		t.Errorf("Expected the original file to remain: %v", err)
	}
	assertMissing(t, repo, "001-bootstrap-the-project.md")
}

func TestSet_Title_Rename(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	output := setStdout(t, repo, "001", "--title", "Bootstrap the project", "--rename")

	if !strings.Contains(output, "file: 001-setup-project.md -> 001-bootstrap-the-project.md") {
		t.Errorf("Expected rename in output, got: %s", output)
	}

	assertMissing(t, repo, "001-setup-project.md")
	content := readTask(t, repo, "001-bootstrap-the-project.md")
	if !strings.Contains(content, `title: "Bootstrap the project"`) {
		t.Errorf("Expected updated frontmatter title, got:\n%s", content)
	}
	if !strings.Contains(content, "# Bootstrap the project") {
		t.Errorf("Expected updated body heading, got:\n%s", content)
	}
}

// The task must still be addressable by ID after its file moves.
func TestSet_Title_RenameKeepsTaskResolvable(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	setStdout(t, repo, "001", "--title", "Bootstrap the project", "--rename")

	res := repo.Run("get", "001", "--format", "json")
	if res.Err != nil {
		t.Fatalf("get after rename failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "Bootstrap the project") {
		t.Errorf("Expected renamed task to be resolvable by ID, got: %s", res.Stdout)
	}
}

func TestSet_Title_CustomHeadingLeftAlone(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-setup-project.md": `---
id: "001"
title: "Setup project"
status: pending
---

# Task 001: a deliberately custom heading

Body text.
`,
	})

	output := setStdout(t, repo, "001", "--title", "Bootstrap the project")

	if !strings.Contains(output, "does not match the old title") {
		t.Errorf("Expected a note about the skipped heading, got: %s", output)
	}

	content := readTask(t, repo, "001-setup-project.md")
	if !strings.Contains(content, "# Task 001: a deliberately custom heading") {
		t.Errorf("Expected the custom heading to survive, got:\n%s", content)
	}
	if !strings.Contains(content, `title: "Bootstrap the project"`) {
		t.Errorf("Expected the frontmatter title to change anyway, got:\n%s", content)
	}
}

func TestSet_Title_NoHeadingInBody(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-setup-project.md": `---
id: "001"
title: "Setup project"
status: pending
---

Just prose, no heading at all.
`,
	})

	output := setStdout(t, repo, "001", "--title", "Bootstrap the project")

	if strings.Contains(output, "does not match") {
		t.Errorf("Expected no heading note when there is no heading, got: %s", output)
	}
	content := readTask(t, repo, "001-setup-project.md")
	if !strings.Contains(content, `title: "Bootstrap the project"`) {
		t.Errorf("Expected the frontmatter title to change, got:\n%s", content)
	}
}

func TestSet_Title_DryRunWritesNothing(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())
	before := readTask(t, repo, "001-setup-project.md")

	output := setStdout(t, repo, "001", "--title", "Bootstrap the project", "--rename", "--dry-run")

	if !strings.Contains(output, "title: Setup project -> Bootstrap the project") {
		t.Errorf("Expected the title change to be previewed, got: %s", output)
	}
	if !strings.Contains(output, "file: 001-setup-project.md -> 001-bootstrap-the-project.md") {
		t.Errorf("Expected the rename to be previewed, got: %s", output)
	}
	if !strings.Contains(output, "Dry run") {
		t.Errorf("Expected a dry-run warning, got: %s", output)
	}

	assertMissing(t, repo, "001-bootstrap-the-project.md")
	if after := readTask(t, repo, "001-setup-project.md"); after != before {
		t.Errorf("Dry run modified the file:\n%s", after)
	}
}

func TestSet_Rename_RequiresTitle(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	res := repo.Run("set", "001", "--rename")

	if res.Err == nil {
		t.Fatal("Expected an error for --rename without --title")
	}
	if !strings.Contains(res.Err.Error(), "--rename requires --title") {
		t.Errorf("Expected an actionable error, got: %v", res.Err)
	}
}

func TestSet_Title_Empty(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	res := repo.Run("set", "001", "--title", "   ")

	if res.Err == nil {
		t.Fatal("Expected an error for an empty --title")
	}
	if !strings.Contains(res.Err.Error(), "--title cannot be empty") {
		t.Errorf("Expected an empty-title error, got: %v", res.Err)
	}
}

// resolveRenamePath refuses to clobber an existing destination. This guard is
// exercised directly rather than through the CLI because any *task* file at the
// destination would share this task's ID and trip the duplicate-ID check first —
// the guard exists for the files the scanner does not own.
func TestResolveRenamePath_DestinationExists(t *testing.T) {
	dir := t.TempDir()
	taken := filepath.Join(dir, "001-taken-name.md")
	if err := os.WriteFile(taken, []byte("placeholder\n"), 0644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	task := &model.Task{ID: "001", FilePath: filepath.Join(dir, "001-setup-project.md")}
	setRename = true
	t.Cleanup(func() { setRename = false })

	_, err := resolveRenamePath(task, "Taken name")
	if err == nil {
		t.Fatal("Expected an error when the destination file exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected a collision error, got: %v", err)
	}
}

// The collision check runs before any write, so a failure leaves the task file
// exactly as it was rather than half-applied.
func TestSet_Title_RenameFailureLeavesFileUntouched(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())
	before := readTask(t, repo, "001-setup-project.md")

	res := repo.Run("set", "001", "--title", "!!!", "--rename")

	if res.Err == nil {
		t.Fatal("Expected the rename to fail")
	}
	if after := readTask(t, repo, "001-setup-project.md"); after != before {
		t.Errorf("Expected the file to be untouched after a failed rename, got:\n%s", after)
	}
}

// A title whose slug is empty has no valid filename to move to.
func TestSet_Title_RenameEmptySlug(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	res := repo.Run("set", "001", "--title", "!!!", "--rename")

	if res.Err == nil {
		t.Fatal("Expected an error for a title with no slug characters")
	}
	if !strings.Contains(res.Err.Error(), "empty filename slug") {
		t.Errorf("Expected an empty-slug error, got: %v", res.Err)
	}
}

// Renaming to the slug the file already has is a no-op, not a collision.
func TestSet_Title_RenameToSameSlug(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	output := setStdout(t, repo, "001", "--title", "Setup  project", "--rename")

	if strings.Contains(output, "file:") {
		t.Errorf("Expected no file change when the slug is unchanged, got: %s", output)
	}
	content := readTask(t, repo, "001-setup-project.md")
	if !strings.Contains(content, `title: "Setup  project"`) {
		t.Errorf("Expected the frontmatter title to change, got:\n%s", content)
	}
}

func TestSet_Title_WithOtherFields(t *testing.T) {
	repo := newTaskRepo(t, renameTestFiles())

	output := setStdout(t, repo, "001",
		"--title", "Bootstrap the project", "--status", "in-progress", "--priority", "low")

	for _, want := range []string{
		"title: Setup project -> Bootstrap the project",
		"status: pending -> in-progress",
		"priority: high -> low",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("Expected %q in output, got: %s", want, output)
		}
	}

	content := readTask(t, repo, "001-setup-project.md")
	for _, want := range []string{`title: "Bootstrap the project"`, "status: in-progress", "priority: low"} {
		if !strings.Contains(content, want) {
			t.Errorf("Expected %q in file, got:\n%s", want, content)
		}
	}
}

// Renaming a task in a subdirectory must keep it in that subdirectory.
func TestSet_Title_RenameStaysInGroupDir(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"cli/001-setup-project.md": `---
id: "001"
title: "Setup project"
status: pending
---

# Setup project
`,
	})

	setStdout(t, repo, "001", "--title", "Bootstrap the project", "--rename")

	assertMissing(t, repo, "cli/001-setup-project.md")
	if _, err := os.Stat(repo.Path("cli/001-bootstrap-the-project.md")); err != nil {
		t.Errorf("Expected the renamed file to stay in cli/: %v", err)
	}
}
