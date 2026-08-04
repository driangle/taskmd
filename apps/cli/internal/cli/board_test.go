package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/board"
)

// boardFixtures returns task files with varied statuses, priorities, tags,
// efforts, and a grouped subdirectory task. Kept inline (not promoted to a
// shared testdata set) because the board tests depend on this specific mix —
// a blocked status, three pending tasks, and a group: field.
func boardFixtures() map[string]string {
	return map[string]string{
		"001-setup.md": `---
id: "001"
title: "Setup project"
status: completed
priority: high
effort: small
dependencies: []
tags: ["infra", "setup"]
created: 2026-02-08
---

# Setup project
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
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: pending
priority: medium
effort: medium
dependencies: []
tags: ["frontend"]
created: 2026-02-08
---

# Build UI components
`,
		"004-api.md": `---
id: "004"
title: "Design API endpoints"
status: blocked
priority: high
effort: medium
dependencies: ["002"]
tags: ["backend"]
created: 2026-02-08
---

# Design API endpoints
`,
		"005-deploy.md": `---
id: "005"
title: "Setup deployment pipeline"
status: pending
priority: low
effort: large
dependencies: []
tags: ["infra"]
created: 2026-02-08
---

# Setup deployment pipeline
`,
		"backend/006-db.md": `---
id: "006"
title: "Database migrations"
status: pending
priority: medium
effort: small
dependencies: []
tags: []
group: "backend"
created: 2026-02-08
---

# Database migrations
`,
	}
}

// boardStdout runs `board <args...>` against repo, fails on error, and returns stdout.
func boardStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"board"}, args...)...)
	if res.Err != nil {
		t.Fatalf("board %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestBoardCommand_DefaultGroupByStatus(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo)

	// Should have status group headers
	if !strings.Contains(output, "## pending") {
		t.Error("Expected output to contain '## pending' header")
	}
	if !strings.Contains(output, "## in-progress") {
		t.Error("Expected output to contain '## in-progress' header")
	}
	if !strings.Contains(output, "## blocked") {
		t.Error("Expected output to contain '## blocked' header")
	}
	if !strings.Contains(output, "## completed") {
		t.Error("Expected output to contain '## completed' header")
	}
}

func TestBoardCommand_GroupByPriority(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--group-by", "priority")

	if !strings.Contains(output, "## critical") {
		t.Error("Expected output to contain '## critical' header")
	}
	if !strings.Contains(output, "## high") {
		t.Error("Expected output to contain '## high' header")
	}
	if !strings.Contains(output, "## medium") {
		t.Error("Expected output to contain '## medium' header")
	}
	if !strings.Contains(output, "## low") {
		t.Error("Expected output to contain '## low' header")
	}
}

func TestBoardCommand_GroupByEffort(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--group-by", "effort")

	if !strings.Contains(output, "## small") {
		t.Error("Expected output to contain '## small' header")
	}
	if !strings.Contains(output, "## medium") {
		t.Error("Expected output to contain '## medium' header")
	}
	if !strings.Contains(output, "## large") {
		t.Error("Expected output to contain '## large' header")
	}
}

func TestBoardCommand_GroupByTag(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--group-by", "tag")

	// Tasks with tags should appear under their tag groups
	if !strings.Contains(output, "## backend") {
		t.Error("Expected output to contain '## backend' header")
	}
	if !strings.Contains(output, "## infra") {
		t.Error("Expected output to contain '## infra' header")
	}
	if !strings.Contains(output, "## frontend") {
		t.Error("Expected output to contain '## frontend' header")
	}

	// Task 006 has no tags, should appear under (none)
	if !strings.Contains(output, "## (none)") {
		t.Error("Expected output to contain '## (none)' header for tagless tasks")
	}

	// Task 001 has tags ["infra", "setup"], should appear under both
	infraSection := output[strings.Index(output, "## infra"):]
	if !strings.Contains(infraSection, "001") {
		t.Error("Expected task 001 to appear under 'infra' group")
	}

	setupSection := output[strings.Index(output, "## setup"):]
	if !strings.Contains(setupSection, "001") {
		t.Error("Expected task 001 to appear under 'setup' group")
	}
}

func TestBoardCommand_GroupByGroup(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--group-by", "group")

	// Task 006 has group: "backend"
	if !strings.Contains(output, "## backend") {
		t.Error("Expected output to contain '## backend' header")
	}

	// Tasks without group should appear under (none)
	if !strings.Contains(output, "## (none)") {
		t.Error("Expected output to contain '## (none)' header for ungrouped tasks")
	}
}

func TestBoardCommand_JSONFormat(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--format", "json")

	var result []board.JSONGroup
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Should have groups for each status
	if len(result) == 0 {
		t.Fatal("Expected at least one group in JSON output")
	}

	// Verify structure
	for _, g := range result {
		if g.Group == "" {
			t.Error("Expected group name to be non-empty")
		}
		if g.Count != len(g.Tasks) {
			t.Errorf("Group %s: count %d doesn't match task count %d", g.Group, g.Count, len(g.Tasks))
		}
		for _, task := range g.Tasks {
			if task.ID == "" || task.Title == "" {
				t.Errorf("Group %s: task missing ID or title", g.Group)
			}
		}
	}
}

func TestBoardCommand_TextFormat(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--format", "txt")

	// Text format should have divider lines
	if !strings.Contains(output, "---") {
		t.Error("Expected text format to contain divider lines")
	}

	// Should have status groups
	if !strings.Contains(output, "pending") {
		t.Error("Expected text output to contain 'pending'")
	}
	if !strings.Contains(output, "completed") {
		t.Error("Expected text output to contain 'completed'")
	}

	// Should NOT have markdown ## headers
	if strings.Contains(output, "## ") {
		t.Error("Expected text format to NOT contain markdown headers")
	}
}

func TestBoardCommand_OutputToFile(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	outFile := filepath.Join(t.TempDir(), "board.md")

	// Run command (output goes to file, not stdout)
	res := repo.Run("board", "--out", outFile)
	if res.Err != nil {
		t.Fatalf("board failed: %v", res.Err)
	}

	// Verify file was created
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Fatal("Expected output file to be created")
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "## pending") {
		t.Error("Expected file content to contain '## pending' header")
	}
}

func TestBoardCommand_TaskCounts(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo)

	// We have: 1 completed, 1 in-progress, 3 pending, 1 blocked
	if !strings.Contains(output, "## completed (1)") {
		t.Error("Expected 'completed (1)' header")
	}
	if !strings.Contains(output, "## in-progress (1)") {
		t.Error("Expected 'in-progress (1)' header")
	}
	if !strings.Contains(output, "## pending (3)") {
		t.Error("Expected 'pending (3)' header")
	}
	if !strings.Contains(output, "## blocked (1)") {
		t.Error("Expected 'blocked (1)' header")
	}
}

func TestBoardCommand_GroupOrdering(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo)

	// Status ordering: pending < in-progress < blocked < completed
	pendingIdx := strings.Index(output, "## pending")
	inProgressIdx := strings.Index(output, "## in-progress")
	blockedIdx := strings.Index(output, "## blocked")
	completedIdx := strings.Index(output, "## completed")

	if pendingIdx == -1 || inProgressIdx == -1 || blockedIdx == -1 || completedIdx == -1 {
		t.Fatal("Expected all status groups to be present")
	}

	if pendingIdx > inProgressIdx {
		t.Error("Expected 'pending' before 'in-progress'")
	}
	if inProgressIdx > blockedIdx {
		t.Error("Expected 'in-progress' before 'blocked'")
	}
	if blockedIdx > completedIdx {
		t.Error("Expected 'blocked' before 'completed'")
	}
}

func TestBoardCommand_InvalidGroupBy(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	res := repo.Run("board", "--group-by", "nonexistent")
	if res.Err == nil {
		t.Fatal("Expected error for invalid group-by field")
	}

	if !strings.Contains(res.Err.Error(), "unsupported group-by field") {
		t.Errorf("Expected 'unsupported group-by field' error, got: %v", res.Err)
	}
}

func TestBoardCommand_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := boardStdout(t, repo)

	// With no tasks, output should be empty
	if strings.TrimSpace(output) != "" {
		t.Errorf("Expected empty output for empty directory, got: %q", output)
	}
}

func TestBoardCommand_ColorEnabled(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	forceColor = true
	defer func() { forceColor = false }()

	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "txt")

	// With colors enabled, output should contain ANSI escape codes
	// ANSI codes start with \x1b[ or \033[
	if !strings.Contains(output, "\x1b[") && !strings.Contains(output, "\033[") {
		t.Error("Expected colored output to contain ANSI escape codes")
	}
}

func TestBoardCommand_NoColorFlag(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "txt", "--no-color")

	// With --no-color flag, output should NOT contain ANSI escape codes
	if strings.Contains(output, "\x1b[") || strings.Contains(output, "\033[") {
		t.Error("Expected --no-color output to NOT contain ANSI escape codes")
	}

	// Output should still contain task information
	if !strings.Contains(output, "001") {
		t.Error("Expected output to contain task IDs even without colors")
	}
}

func TestBoardCommand_NoColorEnvVar(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	forceColor = true
	defer func() { forceColor = false }()

	// Set NO_COLOR environment variable
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "txt")

	// With NO_COLOR env var, output should NOT contain ANSI escape codes
	if strings.Contains(output, "\x1b[") || strings.Contains(output, "\033[") {
		t.Error("Expected output with NO_COLOR env var to NOT contain ANSI escape codes")
	}

	// Output should still contain task information
	if !strings.Contains(output, "002") {
		t.Error("Expected output to contain task IDs even with NO_COLOR set")
	}
}

func TestBoardCommand_ColorMarkdownFormat(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	forceColor = true
	defer func() { forceColor = false }()

	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "md")

	// Markdown format should also support colors
	if !strings.Contains(output, "\x1b[") && !strings.Contains(output, "\033[") {
		t.Error("Expected colored markdown output to contain ANSI escape codes")
	}

	// Should still have markdown structure (## prefix before colored heading)
	if !strings.Contains(output, "## ") {
		t.Error("Expected markdown headers in output")
	}
}

func TestBoardCommand_ColorJSONFormat(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "json")

	// JSON format should NOT have colors (would break JSON parsing)
	if strings.Contains(output, "\x1b[") || strings.Contains(output, "\033[") {
		t.Error("Expected JSON output to NOT contain ANSI escape codes (would break JSON)")
	}

	// Verify it's valid JSON
	var result []board.JSONGroup
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
}

func TestBoardCommand_ColorStatusBased(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	forceColor = true
	defer func() { forceColor = false }()

	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "txt")

	// Different statuses should have different colors
	// We can't easily test the exact colors, but we can verify that:
	// 1. Colors are present in the output
	// 2. Task IDs and titles are in the output
	if !strings.Contains(output, "001") {
		t.Error("Expected output to contain task 001 (completed)")
	}
	if !strings.Contains(output, "002") {
		t.Error("Expected output to contain task 002 (in-progress)")
	}
	if !strings.Contains(output, "003") {
		t.Error("Expected output to contain task 003 (pending)")
	}
	if !strings.Contains(output, "004") {
		t.Error("Expected output to contain task 004 (blocked)")
	}
}

func TestBoardCommand_ColorOutputToFile(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	outFile := filepath.Join(t.TempDir(), "board.txt")

	forceColor = true
	defer func() { forceColor = false }()

	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	// Run command (output goes to file)
	res := repo.Run("board", "--format", "txt", "--out", outFile)
	if res.Err != nil {
		t.Fatalf("board failed: %v", res.Err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Output to file should still have colors (unless --no-color is set)
	if !strings.Contains(string(content), "\x1b[") && !strings.Contains(string(content), "\033[") {
		t.Error("Expected file output to contain ANSI escape codes")
	}
}

func TestBoardCommand_HeadingColoredByStatus(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	forceColor = true
	defer func() { forceColor = false }()

	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--format", "md")

	// Headings should contain ANSI codes (since they are now colored by status)
	if !strings.Contains(output, "\x1b[") {
		t.Error("Expected colored heading output to contain ANSI escape codes")
	}

	// The ## prefix should still be present (uncolored markdown structure)
	if !strings.Contains(output, "## ") {
		t.Error("Expected markdown ## prefix in output")
	}
}

func TestBoardCommand_HeadingColoredByPriority(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())
	forceColor = true
	defer func() { forceColor = false }()

	os.Unsetenv("NO_COLOR")

	output := boardStdout(t, repo, "--group-by", "priority", "--format", "txt")

	// Should have colored headings
	if !strings.Contains(output, "\x1b[") {
		t.Error("Expected colored heading output for priority grouping")
	}

	// The raw priority names should still be embedded in the output
	if !strings.Contains(output, "critical") {
		t.Error("Expected 'critical' in colored output")
	}
	if !strings.Contains(output, "high") {
		t.Error("Expected 'high' in colored output")
	}
}

func TestBoardCommand_GroupByType(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-feat.md": `---
id: "001"
title: "New feature"
status: pending
type: feature
---
`,
		"002-bug.md": `---
id: "002"
title: "Fix crash"
status: pending
type: bug
---
`,
		"003-notype.md": `---
id: "003"
title: "No type task"
status: pending
---
`,
	})

	output := boardStdout(t, repo, "--group-by", "type")

	if !strings.Contains(output, "## feature") {
		t.Error("Expected output to contain '## feature' header")
	}
	if !strings.Contains(output, "## bug") {
		t.Error("Expected output to contain '## bug' header")
	}
	if !strings.Contains(output, "(none)") {
		t.Error("Expected output to contain '(none)' for tasks without type")
	}
}

func TestBoardCommand_GroupByPhase(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "V0.2 feature"
status: pending
phase: v0.2
---
`,
		"002.md": `---
id: "002"
title: "V0.3 feature"
status: pending
phase: v0.3
---
`,
		"003.md": `---
id: "003"
title: "No phase"
status: pending
---
`,
	})

	output := boardStdout(t, repo, "--group-by", "phase")

	if !strings.Contains(output, "## v0.2") {
		t.Error("Expected output to contain '## v0.2' header")
	}
	if !strings.Contains(output, "## v0.3") {
		t.Error("Expected output to contain '## v0.3' header")
	}
	if !strings.Contains(output, "## (none)") {
		t.Error("Expected output to contain '## (none)' header for tasks without phase")
	}
}

func TestBoardCommand_HeadingNoColorFlag(t *testing.T) {
	repo := newTaskRepo(t, boardFixtures())

	output := boardStdout(t, repo, "--format", "md", "--no-color")

	// With no-color, headings should be plain text
	if strings.Contains(output, "\x1b[") {
		t.Error("Expected no ANSI codes in no-color output")
	}

	// Should still have properly formatted headings
	if !strings.Contains(output, "## pending (3)") {
		t.Error("Expected plain '## pending (3)' header")
	}
}
