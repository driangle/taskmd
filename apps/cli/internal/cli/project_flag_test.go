package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// setupProjectFlagRegistry creates a temp global config file with the given project entries
// and sets TASKMD_HOME_CONFIG to point to it.
func setupProjectFlagRegistry(t *testing.T, projectsYAML string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".taskmd.yaml")
	content := "projects:\n" + projectsYAML
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	t.Setenv("TASKMD_HOME_CONFIG", cfgPath)
}

func TestResolveProjectDir_ValidProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	projectDir := t.TempDir()
	tasksDir := filepath.Join(projectDir, "tasks")
	os.MkdirAll(tasksDir, 0755)
	os.WriteFile(filepath.Join(projectDir, ".taskmd.yaml"), []byte("dir: ./tasks\n"), 0644)

	setupProjectFlagRegistry(t, "  - id: myproj\n    name: My Project\n    path: "+projectDir+"\n")

	got, err := resolveProjectDir("myproj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(projectDir, "tasks")
	if got != want {
		t.Errorf("resolveProjectDir(\"myproj\") = %q, want %q", got, want)
	}
}

func TestResolveProjectDir_NotFound(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	setupProjectFlagRegistry(t, "  - id: other\n    path: /tmp\n")

	_, err := resolveProjectDir("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project, got nil")
	}

	want := `project "nonexistent" not found in global registry`
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestResolveProjectDir_PathNotExist(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	badPath := filepath.Join(t.TempDir(), "does-not-exist")
	setupProjectFlagRegistry(t, "  - id: badproj\n    name: Bad\n    path: "+badPath+"\n")

	_, err := resolveProjectDir("badproj")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}

	expected := `project "badproj" path does not exist: ` + badPath
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}

func TestResolveProjectDir_DefaultDir(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, ".taskmd.yaml"), []byte("verbose: true\n"), 0644)

	setupProjectFlagRegistry(t, "  - id: nodir\n    name: No Dir\n    path: "+projectDir+"\n")

	got, err := resolveProjectDir("nodir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != projectDir {
		t.Errorf("resolveProjectDir(\"nodir\") = %q, want %q (project root)", got, projectDir)
	}
}

func TestResolveProjectDir_EmptyRegistry(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	t.Setenv("TASKMD_HOME_CONFIG", filepath.Join(t.TempDir(), "nonexistent.yaml"))

	_, err := resolveProjectDir("anything")
	if err == nil {
		t.Fatal("expected error when registry file is missing, got nil")
	}
}

func TestResolveProjectDir_TaskDirKey(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	projectDir := t.TempDir()
	os.MkdirAll(filepath.Join(projectDir, "my-tasks"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".taskmd.yaml"), []byte("task-dir: ./my-tasks\n"), 0644)

	setupProjectFlagRegistry(t, "  - id: tdkey\n    name: TaskDir Key\n    path: "+projectDir+"\n")

	got, err := resolveProjectDir("tdkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(projectDir, "my-tasks")
	if got != want {
		t.Errorf("resolveProjectDir(\"tdkey\") = %q, want %q", got, want)
	}
}

func TestResolveTaskDir_ProjectFlagTakesPrecedence(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	projectDir := t.TempDir()
	os.MkdirAll(filepath.Join(projectDir, "tasks"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".taskmd.yaml"), []byte("dir: ./tasks\n"), 0644)

	setupProjectFlagRegistry(t, "  - id: precedence\n    name: Precedence Test\n    path: "+projectDir+"\n")

	projectFlag = "precedence"
	taskDir = "/some/other/dir"
	viper.SetConfigType("yaml")

	got := resolveTaskDir()
	want := filepath.Join(projectDir, "tasks")
	if got != want {
		t.Errorf("resolveTaskDir() with --project = %q, want %q", got, want)
	}
}
