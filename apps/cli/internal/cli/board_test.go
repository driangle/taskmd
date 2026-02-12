package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/board"
)

// createBoardTestFiles creates test task files with varied statuses, priorities, tags, and efforts.
func createBoardTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create a subdirectory for group testing
	subDir := filepath.Join(tmpDir, "backend")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	tasks := map[string]string{
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

	for filename, content := range tasks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return tmpDir
}

// resetBoardFlags resets board command flags to defaults before each test.
func resetBoardFlags() {
	boardGroupBy = "status"
	boardFormat = "md"
	boardOut = ""
}

// captureBoardOutput runs the board command and captures stdout.
func captureBoardOutput(t *testing.T, dir string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runBoard(boardCmd, []string{dir})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("runBoard failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestBoardCommand_DefaultGroupByStatus(t *testing.T) {
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardGroupBy = "priority"

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardGroupBy = "effort"

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardGroupBy = "tag"

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardGroupBy = "group"

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardFormat = "json"

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardFormat = "txt"

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	outFile := filepath.Join(t.TempDir(), "board.md")

	resetBoardFlags()
	boardOut = outFile

	// Run command (output goes to file, not stdout)
	err := runBoard(boardCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runBoard failed: %v", err)
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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()

	output := captureBoardOutput(t, tmpDir)

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
	tmpDir := createBoardTestFiles(t)
	resetBoardFlags()
	boardGroupBy = "nonexistent"

	err := runBoard(boardCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for invalid group-by field")
	}

	if !strings.Contains(err.Error(), "unsupported group-by field") {
		t.Errorf("Expected 'unsupported group-by field' error, got: %v", err)
	}
}

func TestBoardCommand_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	resetBoardFlags()

	output := captureBoardOutput(t, tmpDir)

	// With no tasks, output should be empty
	if strings.TrimSpace(output) != "" {
		t.Errorf("Expected empty output for empty directory, got: %q", output)
	}
}
