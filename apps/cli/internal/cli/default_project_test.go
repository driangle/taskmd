package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// setupDefaultProjectConfig creates a global config with default_project and project entries.
func setupDefaultProjectConfig(t *testing.T, defaultProject string, projectsYAML string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".taskmd.yaml")
	content := "default_project: " + defaultProject + "\nprojects:\n" + projectsYAML
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	t.Setenv("TASKMD_HOME_CONFIG", cfgPath)
}

func TestResolveDefaultProject_ValidProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := createProjectWithTasks(t, "tasks", map[string]string{
		"001-task.md": `---
id: "001"
title: "Default Task"
status: pending
priority: high
created: 2026-01-01
---
# Default Task
`,
	})

	setupDefaultProjectConfig(t, "mydefault",
		"  - id: mydefault\n    name: My Default\n    path: "+proj+"\n")

	got := resolveDefaultProject()
	want := filepath.Join(proj, "tasks")
	if got != want {
		t.Errorf("resolveDefaultProject() = %q, want %q", got, want)
	}
}

func TestResolveDefaultProject_NotSet(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	// Global config with no default_project
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".taskmd.yaml")
	os.WriteFile(cfgPath, []byte("projects:\n  - id: foo\n    path: /tmp\n"), 0644)
	t.Setenv("TASKMD_HOME_CONFIG", cfgPath)

	got := resolveDefaultProject()
	if got != "" {
		t.Errorf("resolveDefaultProject() = %q, want empty string", got)
	}
}

func TestResolveDefaultProject_InvalidProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	setupDefaultProjectConfig(t, "nonexistent",
		"  - id: other\n    name: Other\n    path: /tmp\n")

	got := resolveDefaultProject()
	if got != "" {
		t.Errorf("resolveDefaultProject() = %q, want empty (fallback on invalid)", got)
	}
}

func TestResolveTaskDir_ProjectFlagOverridesDefault(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	defaultProj := createProjectWithTasks(t, "tasks", map[string]string{
		"001-task.md": `---
id: "001"
title: "Default Task"
status: pending
created: 2026-01-01
---
`,
	})

	flagProj := createProjectWithTasks(t, "tasks", map[string]string{
		"002-task.md": `---
id: "002"
title: "Flag Task"
status: pending
created: 2026-01-01
---
`,
	})

	setupDefaultProjectConfig(t, "defproj",
		"  - id: defproj\n    name: Default\n    path: "+defaultProj+"\n"+
			"  - id: flagproj\n    name: Flag\n    path: "+flagProj+"\n")

	// --project flag should take precedence over default_project
	projectFlag = "flagproj"

	got := resolveTaskDir()
	want := filepath.Join(flagProj, "tasks")
	if got != want {
		t.Errorf("resolveTaskDir() with --project = %q, want %q", got, want)
	}
}

func TestResolveTaskDir_DefaultProjectUsedWhenNoLocalConfig(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := createProjectWithTasks(t, "tasks", map[string]string{
		"001-task.md": `---
id: "001"
title: "Task"
status: pending
created: 2026-01-01
---
`,
	})

	setupDefaultProjectConfig(t, "myproj",
		"  - id: myproj\n    name: My Project\n    path: "+proj+"\n")

	// Simulate no local config: taskDir empty, no flags changed
	taskDir = ""

	got := resolveTaskDir()
	want := filepath.Join(proj, "tasks")
	if got != want {
		t.Errorf("resolveTaskDir() = %q, want %q (from default_project)", got, want)
	}
}

func TestLoadDefaultProject_ReadsFromConfig(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".taskmd.yaml")
	os.WriteFile(cfgPath, []byte("default_project: myproj\n"), 0644)
	t.Setenv("TASKMD_HOME_CONFIG", cfgPath)

	got := LoadDefaultProject()
	if got != "myproj" {
		t.Errorf("LoadDefaultProject() = %q, want %q", got, "myproj")
	}
}

func TestLoadDefaultProject_EmptyWhenMissing(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	t.Setenv("TASKMD_HOME_CONFIG", filepath.Join(t.TempDir(), "nonexistent.yaml"))

	got := LoadDefaultProject()
	if got != "" {
		t.Errorf("LoadDefaultProject() = %q, want empty", got)
	}
}
