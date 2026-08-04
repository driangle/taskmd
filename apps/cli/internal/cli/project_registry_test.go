package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// createProjectDir creates a temp repo containing a .taskmd.yaml file and
// returns its symlink-resolved path (e.g. /var -> /private/var on macOS) for
// reliable path comparisons against the registered entries.
func createProjectDir(t *testing.T) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}
	repo := &taskRepo{t: t, Dir: resolved}
	repo.Write(configFilename, "dir: .\n")
	return resolved
}

func TestProjectRegister_CWD(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	// chdir to the project directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })
	if err := os.Chdir(projDir); err != nil {
		t.Fatal(err)
	}

	err = runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != filepath.Base(projDir) {
		t.Errorf("expected ID %q, got %q", filepath.Base(projDir), entries[0].ID)
	}
	if entries[0].Path != projDir {
		t.Errorf("expected path %q, got %q", projDir, entries[0].Path)
	}
}

func TestProjectRegister_ExplicitPath(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	projectRegisterPath = projDir
	projectRegisterID = "my-project"
	projectRegisterName = "My Project"

	err := runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "my-project" {
		t.Errorf("expected ID %q, got %q", "my-project", entries[0].ID)
	}
	if entries[0].Name != "My Project" {
		t.Errorf("expected name %q, got %q", "My Project", entries[0].Name)
	}
	if entries[0].Path != projDir {
		t.Errorf("expected path %q, got %q", projDir, entries[0].Path)
	}
}

func TestProjectRegister_NoConfig(t *testing.T) {
	resetCLIState()
	emptyDir := t.TempDir() // no .taskmd.yaml
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	projectRegisterPath = emptyDir

	err := runProjectRegister(projectRegisterCmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
	if !strings.Contains(err.Error(), configFilename) {
		t.Errorf("error should mention %s, got: %v", configFilename, err)
	}
}

func TestProjectRegister_DuplicateID(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	projectRegisterPath = projDir
	projectRegisterID = "dup-id"

	// Register once
	err := runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Register again with same ID but different path
	projDir2 := createProjectDir(t)
	projectRegisterPath = projDir2
	projectRegisterID = "dup-id"

	err = runProjectRegister(projectRegisterCmd, []string{})
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestProjectRegister_CreatesFile(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfgDir := t.TempDir()
	globalCfg := filepath.Join(globalCfgDir, ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	// Confirm file does not exist yet
	if _, err := os.Stat(globalCfg); !os.IsNotExist(err) {
		t.Fatal("global config should not exist yet")
	}

	projectRegisterPath = projDir
	projectRegisterID = "new-proj"

	err := runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should now exist
	if _, err := os.Stat(globalCfg); err != nil {
		t.Fatalf("global config was not created: %v", err)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "new-proj" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestProjectRegister_PreservesExistingConfig(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	// Write existing config with other keys
	existing := "dir: ./tasks\nphases:\n  - id: alpha\n    name: Alpha\n"
	if err := os.WriteFile(globalCfg, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	projectRegisterPath = projDir
	projectRegisterID = "preserved"

	err := runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and check existing keys are preserved
	data, err := os.ReadFile(globalCfg)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "dir:") {
		t.Error("existing 'dir' key was lost")
	}
	if !strings.Contains(content, "phases:") {
		t.Error("existing 'phases' key was lost")
	}
	if !strings.Contains(content, "projects:") {
		t.Error("projects key was not added")
	}
	if !strings.Contains(content, "preserved") {
		t.Error("project entry not found in output")
	}
}

func TestProjectUnregister_ByID(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	// Register a project first
	projectRegisterPath = projDir
	projectRegisterID = "to-remove"
	err := runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Unregister by ID
	resetCLIState()
	projectUnregisterID = "to-remove"
	err = runProjectUnregister(projectUnregisterCmd, []string{})
	if err != nil {
		t.Fatalf("unregister failed: %v", err)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after unregister, got %d", len(entries))
	}
}

func TestProjectUnregister_ByCWD(t *testing.T) {
	resetCLIState()
	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	// Register with explicit path
	projectRegisterPath = projDir
	projectRegisterID = "cwd-proj"
	err := runProjectRegister(projectRegisterCmd, []string{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Unregister by cwd (no --id flag)
	resetCLIState()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })
	if err := os.Chdir(projDir); err != nil {
		t.Fatal(err)
	}

	err = runProjectUnregister(projectUnregisterCmd, []string{})
	if err != nil {
		t.Fatalf("unregister failed: %v", err)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestProjectUnregister_NotFound(t *testing.T) {
	resetCLIState()
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)

	projectUnregisterID = "nonexistent"

	err := runProjectUnregister(projectUnregisterCmd, []string{})
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !strings.Contains(err.Error(), "no project found") {
		t.Errorf("error should mention 'no project found', got: %v", err)
	}
}

func TestReadGlobalConfigNode_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.yaml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := readGlobalConfigNode(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Kind != yaml.DocumentNode {
		t.Errorf("expected DocumentNode, got %v", doc.Kind)
	}
}

func TestReadGlobalConfigNode_NonExistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")
	doc, err := readGlobalConfigNode(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Kind != yaml.DocumentNode {
		t.Errorf("expected DocumentNode, got %v", doc.Kind)
	}
}

func TestGlobalConfigPath_EnvOverride(t *testing.T) {
	expected := "/tmp/test-config.yaml"
	t.Setenv("TASKMD_HOME_CONFIG", expected)
	path, err := globalConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
