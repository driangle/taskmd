package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGlobalRegistry creates a temporary global config file with the given YAML content
// and sets TASKMD_HOME_CONFIG to point to it.
func setupGlobalRegistry(t *testing.T, configYAML string) {
	t.Helper()
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, ".taskmd.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to create global config: %v", err)
	}
	t.Setenv("TASKMD_HOME_CONFIG", configPath)
}

// setupEmptyGlobalRegistry points TASKMD_HOME_CONFIG to a non-existent file.
func setupEmptyGlobalRegistry(t *testing.T) {
	t.Helper()
	configDir := t.TempDir()
	t.Setenv("TASKMD_HOME_CONFIG", filepath.Join(configDir, "nonexistent.yaml"))
}

// createProjectWithTasks creates a project directory (a taskRepo) with a
// .taskmd.yaml pointing at taskSubDir and the given task files, returning the
// project's absolute path.
func createProjectWithTasks(t *testing.T, taskSubDir string, tasks map[string]string) string {
	t.Helper()
	repo := newTaskRepo(t, nil)
	repo.Write(configFilename, "task-dir: "+taskSubDir+"\n")
	for filename, content := range tasks {
		repo.Write(filepath.Join(taskSubDir, filename), content)
	}
	return repo.Dir
}

// runProjectsCmd runs `projects <args...>` against a throwaway repo (the command
// reads the global registry, not a task dir) and returns the result.
func runProjectsCmd(t *testing.T, args ...string) cliResult {
	t.Helper()
	repo := newTaskRepo(t, nil)
	return repo.Run(append([]string{"projects"}, args...)...)
}

func TestProjects_NoProjectsRegistered(t *testing.T) {
	setupEmptyGlobalRegistry(t)

	res := runProjectsCmd(t)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stderr, "No projects registered") {
		t.Errorf("expected 'No projects registered' message, got:\n%s", res.Stderr)
	}
}

func TestProjects_EmptyProjectsList(t *testing.T) {
	setupGlobalRegistry(t, "projects: []\n")

	res := runProjectsCmd(t)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stderr, "No projects registered") {
		t.Errorf("expected 'No projects registered' message, got:\n%s", res.Stderr)
	}
}

func TestProjects_ValidProjectsTable(t *testing.T) {
	projectDir := createProjectWithTasks(t, "tasks", map[string]string{
		"001.md": taskFile("001", "Task A", "pending"),
		"002.md": taskFile("002", "Task B", "in-progress"),
		"003.md": taskFile("003", "Task C", "completed"),
		"004.md": taskFile("004", "Task D", "pending"),
	})

	setupGlobalRegistry(t, "projects:\n"+
		"  - id: proj1\n"+
		"    name: \"Test Project\"\n"+
		"    path: "+projectDir+"\n")

	res := runProjectsCmd(t)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	for _, expected := range []string{"PROJECT", "PATH", "TASKS", "PENDING", "IN-PROGRESS", "COMPLETED"} {
		if !strings.Contains(res.Stdout, expected) {
			t.Errorf("table output missing header %q:\n%s", expected, res.Stdout)
		}
	}

	if !strings.Contains(res.Stdout, "Test Project") {
		t.Errorf("table output missing project name:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, projectDir) {
		t.Errorf("table output missing project path:\n%s", res.Stdout)
	}
}

func TestProjects_JSONOutput(t *testing.T) {
	projectDir := createProjectWithTasks(t, "tasks", map[string]string{
		"001.md": taskFile("001", "Task A", "pending"),
		"002.md": taskFile("002", "Task B", "in-progress"),
		"003.md": taskFile("003", "Task C", "completed"),
	})

	setupGlobalRegistry(t, "projects:\n"+
		"  - id: proj1\n"+
		"    name: \"My Project\"\n"+
		"    path: "+projectDir+"\n")

	res := runProjectsCmd(t, "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	var summaries []ProjectSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if len(summaries) != 1 {
		t.Fatalf("expected 1 project, got %d", len(summaries))
	}

	s := summaries[0]
	if s.ID != "proj1" {
		t.Errorf("expected id 'proj1', got %q", s.ID)
	}
	if s.Name != "My Project" {
		t.Errorf("expected name 'My Project', got %q", s.Name)
	}
	if s.Tasks != 3 {
		t.Errorf("expected 3 tasks, got %d", s.Tasks)
	}
	if s.Pending != 1 {
		t.Errorf("expected 1 pending, got %d", s.Pending)
	}
	if s.InProgress != 1 {
		t.Errorf("expected 1 in-progress, got %d", s.InProgress)
	}
	if s.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", s.Completed)
	}
}

func TestProjects_UnreachablePath(t *testing.T) {
	// Create one valid project
	projectDir := createProjectWithTasks(t, "tasks", map[string]string{
		"001.md": taskFile("001", "Task A", "pending"),
	})

	setupGlobalRegistry(t, "projects:\n"+
		"  - id: missing\n"+
		"    name: \"Missing Project\"\n"+
		"    path: /nonexistent/path/that/does/not/exist\n"+
		"  - id: valid\n"+
		"    name: \"Valid Project\"\n"+
		"    path: "+projectDir+"\n")

	res := runProjectsCmd(t)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stderr, "Warning") {
		t.Errorf("expected warning for missing project, got stderr:\n%s", res.Stderr)
	}
	if !strings.Contains(res.Stderr, "Missing Project") {
		t.Errorf("expected warning to mention 'Missing Project', got stderr:\n%s", res.Stderr)
	}

	// Valid project should still appear
	if !strings.Contains(res.Stdout, "Valid Project") {
		t.Errorf("expected valid project in output, got:\n%s", res.Stdout)
	}
}

func TestProjects_YAMLOutput(t *testing.T) {
	projectDir := createProjectWithTasks(t, "tasks", map[string]string{
		"001.md": taskFile("001", "Task A", "completed"),
	})

	setupGlobalRegistry(t, "projects:\n"+
		"  - id: proj1\n"+
		"    name: \"YAML Project\"\n"+
		"    path: "+projectDir+"\n")

	res := runProjectsCmd(t, "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "name: YAML Project") {
		t.Errorf("expected YAML output with project name, got:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "completed: 1") {
		t.Errorf("expected YAML output with completed count, got:\n%s", res.Stdout)
	}
}

func TestProjects_DefaultTaskDir(t *testing.T) {
	// Create project without .taskmd.yaml — should default to ./tasks
	repo := newTaskRepo(t, map[string]string{
		"tasks/001.md": taskFile("001", "Default Dir Task", "pending"),
	})

	setupGlobalRegistry(t, "projects:\n"+
		"  - id: proj1\n"+
		"    name: \"Default Dir Project\"\n"+
		"    path: "+repo.Dir+"\n")

	res := runProjectsCmd(t, "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	var summaries []ProjectSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if len(summaries) != 1 {
		t.Fatalf("expected 1 project, got %d", len(summaries))
	}
	if summaries[0].Tasks != 1 {
		t.Errorf("expected 1 task (from default tasks/ dir), got %d", summaries[0].Tasks)
	}
}

func TestProjects_InvalidFormat(t *testing.T) {
	// A valid project is needed so the format switch (which runs after
	// collectProjectSummaries) is reached.
	projectDir := createProjectWithTasks(t, "tasks", map[string]string{
		"001.md": taskFile("001", "Task", "pending"),
	})
	setupGlobalRegistry(t, "projects:\n"+
		"  - id: proj1\n"+
		"    name: \"Test\"\n"+
		"    path: "+projectDir+"\n")

	res := runProjectsCmd(t, "--format", "csv")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

// taskFile is a helper to generate task markdown content.
func taskFile(id, title, status string) string {
	return "---\nid: \"" + id + "\"\ntitle: \"" + title + "\"\nstatus: " + status + "\n---\n"
}
