package cli

import (
	"path/filepath"
	"testing"
)

func TestScanAllProjects_MultipleProjects(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj1 := createProjectWithTasks(t, "tasks", map[string]string{
		"001-task-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: high
created: 2026-01-01
---
# Task A
`,
	})

	proj2 := createProjectWithTasks(t, "tasks", map[string]string{
		"001-task-b.md": `---
id: "001"
title: "Task B"
status: in-progress
priority: medium
created: 2026-01-01
---
# Task B
`,
		"002-task-c.md": `---
id: "002"
title: "Task C"
status: pending
priority: low
created: 2026-01-01
---
# Task C
`,
	})

	setupProjectFlagRegistry(t,
		"  - id: alpha\n    name: Alpha\n    path: "+proj1+"\n"+
			"  - id: beta\n    name: Beta\n    path: "+proj2+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ptasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(ptasks))
	}

	// Verify project IDs are set
	projectCounts := map[string]int{}
	for _, pt := range ptasks {
		projectCounts[pt.ProjectID]++
	}
	if projectCounts["alpha"] != 1 {
		t.Errorf("expected 1 task from alpha, got %d", projectCounts["alpha"])
	}
	if projectCounts["beta"] != 2 {
		t.Errorf("expected 2 tasks from beta, got %d", projectCounts["beta"])
	}
}

func TestScanAllProjects_QualifiedID(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := createProjectWithTasks(t, "tasks", map[string]string{
		"042-task.md": `---
id: "042"
title: "Test Task"
status: pending
priority: medium
created: 2026-01-01
---
# Test Task
`,
	})

	setupProjectFlagRegistry(t, "  - id: myproj\n    name: My Project\n    path: "+proj+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ptasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(ptasks))
	}

	got := ptasks[0].QualifiedID()
	want := "myproj:042"
	if got != want {
		t.Errorf("QualifiedID() = %q, want %q", got, want)
	}
}

func TestScanAllProjects_SkipsUnreachable(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	goodProj := createProjectWithTasks(t, "tasks", map[string]string{
		"001-task.md": `---
id: "001"
title: "Good Task"
status: pending
priority: high
created: 2026-01-01
---
# Good Task
`,
	})

	badPath := filepath.Join(t.TempDir(), "does-not-exist")

	setupProjectFlagRegistry(t,
		"  - id: good\n    name: Good\n    path: "+goodProj+"\n"+
			"  - id: bad\n    name: Bad\n    path: "+badPath+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ptasks) != 1 {
		t.Fatalf("expected 1 task (bad project skipped), got %d", len(ptasks))
	}

	if ptasks[0].ProjectID != "good" {
		t.Errorf("expected task from 'good' project, got %q", ptasks[0].ProjectID)
	}
}

func TestScanAllProjects_EmptyRegistry(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	t.Setenv("TASKMD_HOME_CONFIG", filepath.Join(t.TempDir(), "nonexistent.yaml"))

	_, err := scanAllProjects()
	if err == nil {
		t.Fatal("expected error for empty/missing registry, got nil")
	}
}
