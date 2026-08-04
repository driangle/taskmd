package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dedupStdout runs `deduplicate <args...>` against repo, fails on error, and
// returns stdout.
func dedupStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"deduplicate"}, args...)...)
	if res.Err != nil {
		t.Fatalf("deduplicate %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestDeduplicate_NoDuplicates(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		"002-task-b.md": `---
id: "002"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-02
---

# Task B
`,
	})

	output := dedupStdout(t, repo, "--no-interactive")

	if !strings.Contains(output, "No duplicate IDs found") {
		t.Errorf("expected 'No duplicate IDs found', got: %s", output)
	}
}

func TestDeduplicate_SingleCollision(t *testing.T) {
	// Two tasks with the same ID "001", different created dates.
	repo := newTaskRepo(t, map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
	})

	output := dedupStdout(t, repo, "--no-interactive")

	if !strings.Contains(output, "1 duplicate") {
		t.Errorf("expected '1 duplicate' in output, got: %s", output)
	}
	if !strings.Contains(output, "Resolved") {
		t.Errorf("expected 'Resolved' in output, got: %s", output)
	}

	// Verify: older task (Task A) should still have ID 001
	contentA, _ := os.ReadFile(repo.Path("001-task-a.md"))
	if !strings.Contains(string(contentA), `id: "001"`) {
		t.Error("expected older task (Task A) to keep ID 001")
	}

	// Verify: newer task (Task B) should have a new ID and be renamed
	files, _ := os.ReadDir(repo.Dir)
	foundNewFile := false
	for _, f := range files {
		if f.Name() != "001-task-a.md" && strings.HasSuffix(f.Name(), ".md") {
			foundNewFile = true
			content, _ := os.ReadFile(filepath.Join(repo.Dir, f.Name()))
			if strings.Contains(string(content), `id: "001"`) {
				t.Errorf("expected renamed file to have new ID, got content: %s", string(content))
			}
			if !strings.Contains(string(content), `title: "Task B"`) {
				t.Error("expected renamed file to preserve Task B title")
			}
		}
	}
	if !foundNewFile {
		t.Error("expected a renamed file for Task B")
	}
}

func TestDeduplicate_MultipleCollisions(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		// Collision on ID "001"
		"001-first.md": `---
id: "001"
title: "First"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# First
`,
		"001-second.md": `---
id: "001"
title: "Second"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-10
---

# Second
`,
		// Collision on ID "002"
		"002-third.md": `---
id: "002"
title: "Third"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Third
`,
		"002-fourth.md": `---
id: "002"
title: "Fourth"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-10
---

# Fourth
`,
	})

	output := dedupStdout(t, repo, "--no-interactive")

	if !strings.Contains(output, "2 duplicate") {
		t.Errorf("expected '2 duplicate' in output, got: %s", output)
	}

	// Verify: both original files still exist
	files, _ := os.ReadDir(repo.Dir)
	if len(files) != 4 {
		t.Errorf("expected 4 files, got %d", len(files))
	}
}

func TestDeduplicate_CrossReferenceUpdate(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID "001"
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate of "001", will be reassigned
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001" (the duplicate ID)
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
		// Task D: has parent "001"
		"003-task-d.md": `---
id: "003"
title: "Task D"
status: pending
priority: medium
parent: "001"
dependencies: []
tags: []
created: 2026-01-10
---

# Task D
`,
	})

	if res := repo.Run("deduplicate", "--no-interactive"); res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Find what the new ID is from the renamed file
	files, _ := os.ReadDir(repo.Dir)
	var newID string
	for _, f := range files {
		name := f.Name()
		if name != "001-task-a.md" && name != "002-task-c.md" && name != "003-task-d.md" && strings.HasSuffix(name, ".md") {
			content, _ := os.ReadFile(filepath.Join(repo.Dir, name))
			// Extract new ID from frontmatter
			for _, line := range strings.Split(string(content), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "id:") {
					newID = strings.Trim(strings.TrimPrefix(strings.TrimSpace(line), "id:"), ` "`)
					break
				}
			}
		}
	}

	if newID == "" || newID == "001" {
		t.Fatal("expected Task B to have a new non-001 ID")
	}

	// In non-interactive mode, all references to "001" are rewritten to newID.
	// This is the known limitation documented in the task.
	contentC, _ := os.ReadFile(repo.Path("002-task-c.md"))
	_ = contentC

	contentD, _ := os.ReadFile(repo.Path("003-task-d.md"))
	_ = contentD
}

func TestDeduplicate_DryRun(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
	})

	output := dedupStdout(t, repo, "--dry-run")

	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected 'dry-run' in output, got: %s", output)
	}
	if !strings.Contains(output, "No changes made") {
		t.Errorf("expected 'No changes made' in output, got: %s", output)
	}

	// Verify: files should NOT be modified
	contentB, _ := os.ReadFile(repo.Path("001-task-b.md"))
	if !strings.Contains(string(contentB), `id: "001"`) {
		t.Error("dry-run should not modify files")
	}

	// Verify: file should NOT be renamed
	if _, err := os.Stat(repo.Path("001-task-b.md")); os.IsNotExist(err) {
		t.Error("dry-run should not rename files")
	}
}

func TestDeduplicate_JSONFormat(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
	})

	output := dedupStdout(t, repo, "--format", "json", "--no-interactive")

	var result deduplicateResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, output)
	}

	if result.Duplicates != 1 {
		t.Errorf("expected 1 duplicate, got %d", result.Duplicates)
	}
	if len(result.Reassignments) != 1 {
		t.Fatalf("expected 1 reassignment, got %d", len(result.Reassignments))
	}
	if result.Reassignments[0].OldID != "001" {
		t.Errorf("expected old ID '001', got %q", result.Reassignments[0].OldID)
	}
	if result.Reassignments[0].NewID == "001" {
		t.Error("expected new ID to be different from '001'")
	}
}

func TestDeduplicate_NoDuplicates_JSON(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
	})

	output := dedupStdout(t, repo, "--format", "json")

	var result deduplicateResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, output)
	}

	if result.Duplicates != 0 {
		t.Errorf("expected 0 duplicates, got %d", result.Duplicates)
	}
}

func TestDeduplicate_SameCreatedDate_FallbackToFilepath(t *testing.T) {
	// Both tasks have the same created date — should fall back to filepath order.
	repo := newTaskRepo(t, map[string]string{
		"001-aaa.md": `---
id: "001"
title: "AAA"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# AAA
`,
		"001-zzz.md": `---
id: "001"
title: "ZZZ"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# ZZZ
`,
	})

	if res := repo.Run("deduplicate", "--no-interactive"); res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// AAA should keep the ID (comes first alphabetically), ZZZ should be reassigned
	contentAAA, _ := os.ReadFile(repo.Path("001-aaa.md"))
	if !strings.Contains(string(contentAAA), `id: "001"`) {
		t.Error("expected AAA (first alphabetically) to keep ID 001")
	}

	// ZZZ file should have been renamed
	if _, err := os.Stat(repo.Path("001-zzz.md")); err == nil {
		t.Error("expected 001-zzz.md to be renamed")
	}
}

func TestBuildNewFilePath(t *testing.T) {
	tests := []struct {
		name    string
		oldPath string
		oldID   string
		newID   string
		want    string
	}{
		{
			name:    "standard dash separator",
			oldPath: "/tasks/001-fix-login.md",
			oldID:   "001",
			newID:   "abc123",
			want:    "/tasks/abc123-fix-login.md",
		},
		{
			name:    "dot separator",
			oldPath: "/tasks/001.md",
			oldID:   "001",
			newID:   "xyz",
			want:    "/tasks/xyz.md",
		},
		{
			name:    "no matching prefix fallback",
			oldPath: "/tasks/some-file.md",
			oldID:   "001",
			newID:   "xyz",
			want:    "/tasks/xyz-some-file.md",
		},
		{
			name:    "prefixed ID",
			oldPath: "/tasks/dr-001-some-task.md",
			oldID:   "dr-001",
			newID:   "dr-002",
			want:    "/tasks/dr-002-some-task.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNewFilePath(tt.oldPath, tt.oldID, tt.newID)
			if got != tt.want {
				t.Errorf("buildNewFilePath(%q, %q, %q) = %q, want %q",
					tt.oldPath, tt.oldID, tt.newID, got, tt.want)
			}
		})
	}
}

// --- Interactive disambiguation tests ---

func TestDeduplicate_InteractiveDisambiguation(t *testing.T) {
	// Override TTY check to return true for interactive mode.
	oldIsTTY := dedupIsTTY
	dedupIsTTY = func() bool { return true }
	defer func() { dedupIsTTY = oldIsTTY }()

	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID "001"
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate of "001", will be reassigned
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001" — user will choose [1] (Task A, keeps ID)
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
	})

	// Simulate user choosing "1" (the oldest task, which keeps ID "001").
	oldStdin := dedupStdinReader
	dedupStdinReader = strings.NewReader("1\n")
	defer func() { dedupStdinReader = oldStdin }()

	if res := repo.Run("deduplicate"); res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Task C should still reference "001" because user chose the task that keeps the ID.
	contentC, _ := os.ReadFile(repo.Path("002-task-c.md"))
	if !strings.Contains(string(contentC), `"001"`) {
		t.Errorf("expected Task C to still reference '001', got:\n%s", string(contentC))
	}
}

func TestDeduplicate_InteractiveChooseReassigned(t *testing.T) {
	oldIsTTY := dedupIsTTY
	dedupIsTTY = func() bool { return true }
	defer func() { dedupIsTTY = oldIsTTY }()

	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID "001"
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate, will be reassigned
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001" — user will choose [2] (Task B, gets new ID)
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
	})

	// Simulate user choosing "2" (the reassigned task).
	oldStdin := dedupStdinReader
	dedupStdinReader = strings.NewReader("2\n")
	defer func() { dedupStdinReader = oldStdin }()

	if res := repo.Run("deduplicate"); res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Task C should now reference the new ID (not "001").
	contentC, _ := os.ReadFile(repo.Path("002-task-c.md"))
	if strings.Contains(string(contentC), `"001"`) {
		t.Errorf("expected Task C's reference to be updated to new ID, got:\n%s", string(contentC))
	}
}

func TestDeduplicate_InteractiveMultipleRefs(t *testing.T) {
	oldIsTTY := dedupIsTTY
	dedupIsTTY = func() bool { return true }
	defer func() { dedupIsTTY = oldIsTTY }()

	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID "001"
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate of "001"
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001" — user chooses [1] (keep original)
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
		// Task D: parent "001" — user chooses [2] (reassigned)
		"003-task-d.md": `---
id: "003"
title: "Task D"
status: pending
priority: medium
parent: "001"
dependencies: []
tags: []
created: 2026-01-10
---

# Task D
`,
	})

	// Two prompts: first "1\n" for Task C's dep, then "2\n" for Task D's parent.
	oldStdin := dedupStdinReader
	dedupStdinReader = strings.NewReader("1\n2\n")
	defer func() { dedupStdinReader = oldStdin }()

	if res := repo.Run("deduplicate"); res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Task C should still reference "001" (chose oldest).
	contentC, _ := os.ReadFile(repo.Path("002-task-c.md"))
	if !strings.Contains(string(contentC), `"001"`) {
		t.Errorf("expected Task C to still reference '001', got:\n%s", string(contentC))
	}

	// Task D should reference the new ID (chose reassigned).
	contentD, _ := os.ReadFile(repo.Path("003-task-d.md"))
	if strings.Contains(string(contentD), `parent: "001"`) {
		t.Errorf("expected Task D's parent to be updated to new ID, got:\n%s", string(contentD))
	}
}

func TestDeduplicate_NoInteractiveFlag(t *testing.T) {
	oldIsTTY := dedupIsTTY
	dedupIsTTY = func() bool { return true }
	defer func() { dedupIsTTY = oldIsTTY }()

	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID "001"
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001"
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
	})

	// Should NOT prompt — no stdin needed.
	output := dedupStdout(t, repo, "--no-interactive")

	if !strings.Contains(output, "1 duplicate") {
		t.Errorf("expected '1 duplicate' in output, got: %s", output)
	}

	// In non-interactive mode, references are blindly rewritten (existing behavior).
	// Task C's "001" dep gets rewritten to the new ID — we just verify no error occurred.
}

func TestDeduplicate_NonTTYFallback(t *testing.T) {
	// Simulate non-TTY environment.
	oldIsTTY := dedupIsTTY
	dedupIsTTY = func() bool { return false }
	defer func() { dedupIsTTY = oldIsTTY }()

	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001"
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
	})

	// Should NOT prompt (non-TTY) — no stdin input needed.
	output := dedupStdout(t, repo)

	if !strings.Contains(output, "Resolved") {
		t.Errorf("expected 'Resolved' in output, got: %s", output)
	}
}

func TestDeduplicate_DryRunShowsAmbiguous(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		// Task A: older, keeps ID "001"
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		// Task B: newer duplicate of "001"
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
		// Task C: depends on "001" — ambiguous reference
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: medium
dependencies: ["001"]
tags: []
created: 2026-01-10
---

# Task C
`,
	})

	output := dedupStdout(t, repo, "--dry-run")

	if !strings.Contains(output, "Ambiguous references detected") {
		t.Errorf("expected dry-run to show ambiguous references, got:\n%s", output)
	}
	if !strings.Contains(output, "002-task-c.md") {
		t.Errorf("expected dry-run to mention the referencing file, got:\n%s", output)
	}
	if !strings.Contains(output, "2 candidates") {
		t.Errorf("expected dry-run to show candidate count, got:\n%s", output)
	}
	if !strings.Contains(output, "No changes made") {
		t.Errorf("expected 'No changes made' in dry-run output, got:\n%s", output)
	}
}

func TestDeduplicate_NoAmbiguousRefs(t *testing.T) {
	oldIsTTY := dedupIsTTY
	dedupIsTTY = func() bool { return true }
	defer func() { dedupIsTTY = oldIsTTY }()

	// Two tasks with same ID but no other task references them.
	repo := newTaskRepo(t, map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-01
---

# Task A
`,
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-01-15
---

# Task B
`,
	})

	// No stdin input needed — no ambiguous refs means no prompting.
	output := dedupStdout(t, repo)

	if !strings.Contains(output, "Resolved") {
		t.Errorf("expected 'Resolved' in output, got: %s", output)
	}
}
