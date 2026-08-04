package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runInit executes the `init` command against repo, rooted at repo.Dir with the
// task directory also at repo.Dir (so files land in the repo root), in non-TTY
// mode. projectInitRoot and projectInitIsTTY are non-flag globals the harness
// reset does not cover, so they are seeded in the RunWith configure hook.
func runInit(repo *taskRepo, args ...string) cliResult {
	return runInitIn(repo, repo.Dir, repo.Dir, args...)
}

// runInitIn is runInit with explicit project root and task directory, for tests
// that place the task directory apart from the project root.
func runInitIn(repo *taskRepo, root, taskDir string, args ...string) cliResult {
	return repo.RunWith(func() {
		projectInitRoot = root
		projectInitTaskDir = taskDir
		projectInitIsTTY = func() bool { return false }
	}, append([]string{"init"}, args...)...)
}

func TestProjectInit_DefaultWritesBothFiles(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create CLAUDE.md (default agent) in root
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if !bytes.Equal(content, claudeTemplate) {
		t.Error("CLAUDE.md content does not match template")
	}

	// Should create TASKMD_SPEC.md in task dir (same as root in this test)
	specPath := filepath.Join(repo.Dir, specFilename)
	content, err = os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", specFilename, err)
	}
	if !bytes.Equal(content, initSpecTemplate) {
		t.Error("TASKMD_SPEC.md content does not match template")
	}
}

func TestProjectInit_GeminiFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--gemini")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create GEMINI.md
	geminiPath := filepath.Join(repo.Dir, "GEMINI.md")
	content, err := os.ReadFile(geminiPath)
	if err != nil {
		t.Fatalf("failed to read GEMINI.md: %v", err)
	}
	if !bytes.Equal(content, geminiTemplate) {
		t.Error("GEMINI.md content does not match template")
	}

	// Should create TASKMD_SPEC.md
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}

	// Should NOT create CLAUDE.md
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created when --gemini is specified")
	}
}

func TestProjectInit_MultipleAgentFlags(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--gemini")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create CLAUDE.md, GEMINI.md, and TASKMD_SPEC.md
	for _, name := range []string{"CLAUDE.md", "GEMINI.md", specFilename} {
		path := filepath.Join(repo.Dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s should have been created", name)
		}
	}
}

func TestProjectInit_NoSpecFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--no-spec")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create CLAUDE.md
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should have been created")
	}

	// Should NOT create TASKMD_SPEC.md
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); err == nil {
		t.Error("TASKMD_SPEC.md should not have been created with --no-spec")
	}
}

func TestProjectInit_NoAgentFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--no-agent")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create TASKMD_SPEC.md
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}

	// Should NOT create CLAUDE.md
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created with --no-agent")
	}
}

func TestProjectInit_NoSpecAndNoAgentAndNoTemplatesIsError(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--no-spec", "--no-agent", "--no-templates")
	if res.Err == nil {
		t.Fatal("expected error when all --no-* flags are set")
	}

	if !strings.Contains(res.Err.Error(), "nothing to do") {
		t.Errorf("error message %q should contain 'nothing to do'", res.Err.Error())
	}
}

func TestProjectInit_ForceOverwritesExistingFiles(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create existing files
	claudePath := repo.Write("CLAUDE.md", "old claude")
	specPath := repo.Write(specFilename, "old spec")

	res := runInit(repo, "--force")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Verify files were overwritten
	claudeContent, _ := os.ReadFile(claudePath)
	if !bytes.Equal(claudeContent, claudeTemplate) {
		t.Error("CLAUDE.md should have been overwritten")
	}

	specContent, _ := os.ReadFile(specPath)
	if !bytes.Equal(specContent, initSpecTemplate) {
		t.Error("TASKMD_SPEC.md should have been overwritten")
	}
}

func TestProjectInit_ExistingFilesSkippedWithoutForce(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create an existing CLAUDE.md
	claudePath := repo.Write("CLAUDE.md", "existing claude")

	res := runInit(repo)
	if res.Err != nil {
		t.Fatalf("expected no error when skipping existing files, got: %v", res.Err)
	}

	// Should have a warning about skipping
	if !strings.Contains(res.Stderr, "Skipped") {
		t.Errorf("expected skip warning on stderr, got: %q", res.Stderr)
	}

	// Original file should be unchanged
	content, _ := os.ReadFile(claudePath)
	if string(content) != "existing claude" {
		t.Error("existing CLAUDE.md should not have been overwritten")
	}

	// TASKMD_SPEC.md should still be created (it didn't exist)
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created even though CLAUDE.md was skipped")
	}
}

func TestProjectInit_DirFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)
	subDir := filepath.Join(repo.Dir, "my-project")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	res := runInitIn(repo, subDir, subDir)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	claudePath := filepath.Join(subDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should have been created in subdirectory")
	}

	specPath := filepath.Join(subDir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created in subdirectory")
	}
}

func TestProjectInit_StdoutPrintsWithoutCreatingFiles(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--stdout")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should contain both agent template and spec template
	if !strings.Contains(res.Stdout, string(claudeTemplate)) {
		t.Error("stdout output should contain Claude template")
	}
	if !strings.Contains(res.Stdout, string(initSpecTemplate)) {
		t.Error("stdout output should contain spec template")
	}

	// No files should have been created
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created with --stdout")
	}
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); err == nil {
		t.Error("TASKMD_SPEC.md should not have been created with --stdout")
	}
}

func TestProjectInit_CodexFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--codex")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create AGENTS.md and TASKMD_SPEC.md
	agentsPath := filepath.Join(repo.Dir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		t.Error("AGENTS.md should have been created")
	}

	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}

	// Should NOT create CLAUDE.md
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created when --codex is specified")
	}
}

func TestProjectInit_AllAgentFlags(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--gemini", "--codex")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	expected := []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md", specFilename}
	for _, name := range expected {
		path := filepath.Join(repo.Dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s should have been created", name)
		}
	}
}

func TestProjectInit_PartialSkipStillCreatesOthers(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create only CLAUDE.md as existing
	claudePath := repo.Write("CLAUDE.md", "existing")

	res := runInit(repo, "--claude", "--gemini")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// CLAUDE.md should be unchanged (skipped)
	content, _ := os.ReadFile(claudePath)
	if string(content) != "existing" {
		t.Error("existing CLAUDE.md should not have been overwritten")
	}

	// GEMINI.md should be created
	geminiPath := filepath.Join(repo.Dir, "GEMINI.md")
	if _, err := os.Stat(geminiPath); os.IsNotExist(err) {
		t.Error("GEMINI.md should have been created")
	}

	// TASKMD_SPEC.md should be created
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}
}

// --- New tests for interactive init ---

func TestProjectInit_SeparateDirectories(t *testing.T) {
	repo := newTaskRepo(t, nil)
	rootDir := repo.Dir
	taskDirPath := filepath.Join(repo.Dir, "my-tasks")

	res := runInitIn(repo, rootDir, taskDirPath, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Agent config should be in task directory
	claudePath := filepath.Join(taskDirPath, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should have been created in task directory")
	}

	// Agent config should NOT be in project root
	claudeInRoot := filepath.Join(rootDir, "CLAUDE.md")
	if _, err := os.Stat(claudeInRoot); err == nil {
		t.Error("CLAUDE.md should not have been created in project root")
	}

	// Spec should be in task directory
	specPath := filepath.Join(taskDirPath, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created in task directory")
	}

	// Spec should NOT be in project root
	specInRoot := filepath.Join(rootDir, specFilename)
	if _, err := os.Stat(specInRoot); err == nil {
		t.Error("TASKMD_SPEC.md should not have been created in project root")
	}
}

func TestProjectInit_ConfigFileContent(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	configPath := filepath.Join(repo.Dir, configFilename)
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", configFilename, err)
	}

	expected := "dir: " + repo.Dir + "\n"
	if string(content) != expected {
		t.Errorf("config content = %q, want %q", string(content), expected)
	}
}

func TestProjectInit_NonTTY_DefaultsClaude(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// No agent flags set, non-TTY (default from runInit)
	res := runInit(repo)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should create CLAUDE.md (the non-TTY default)
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should have been created as non-TTY default")
	}

	// Should NOT create GEMINI.md or AGENTS.md
	geminiPath := filepath.Join(repo.Dir, "GEMINI.md")
	if _, err := os.Stat(geminiPath); err == nil {
		t.Error("GEMINI.md should not have been created")
	}
	agentsPath := filepath.Join(repo.Dir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		t.Error("AGENTS.md should not have been created")
	}
}

func TestProjectInit_ExistingConfig_SkippedWithoutForce(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create existing config
	configPath := repo.Write(configFilename, "dir: ./old-tasks\n")

	res := runInit(repo, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should warn about skipping config
	if !strings.Contains(res.Stderr, "Skipped") || !strings.Contains(res.Stderr, configFilename) {
		t.Errorf("expected skip warning for %s, got: %q", configFilename, res.Stderr)
	}

	// Config should be unchanged
	content, _ := os.ReadFile(configPath)
	if string(content) != "dir: ./old-tasks\n" {
		t.Error("existing config should not have been overwritten")
	}
}

func TestProjectInit_ExistingConfig_OverwrittenWithForce(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create existing config
	configPath := repo.Write(configFilename, "dir: ./old-tasks\n")

	res := runInit(repo, "--force", "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Config should be overwritten
	content, _ := os.ReadFile(configPath)
	expected := "dir: " + repo.Dir + "\n"
	if string(content) != expected {
		t.Errorf("config content = %q, want %q", string(content), expected)
	}
}

func TestProjectInit_ExistingTaskDir_Graceful(t *testing.T) {
	repo := newTaskRepo(t, nil)
	taskDirPath := filepath.Join(repo.Dir, "tasks")
	if err := os.MkdirAll(taskDirPath, 0755); err != nil {
		t.Fatalf("failed to create task dir: %v", err)
	}

	res := runInitIn(repo, repo.Dir, taskDirPath, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Should note existing directory
	if !strings.Contains(res.Stderr, "already exists") {
		t.Errorf("expected 'already exists' note, got: %q", res.Stderr)
	}

	// Spec should still be created in the existing task dir
	specPath := filepath.Join(taskDirPath, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created in existing task directory")
	}
}

func TestProjectInit_CreatesTaskDir(t *testing.T) {
	repo := newTaskRepo(t, nil)
	taskDirPath := filepath.Join(repo.Dir, "new-tasks")

	// Verify task dir doesn't exist yet
	if _, err := os.Stat(taskDirPath); err == nil {
		t.Fatal("task directory should not exist before init")
	}

	res := runInitIn(repo, repo.Dir, taskDirPath, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Task directory should now exist
	info, err := os.Stat(taskDirPath)
	if err != nil {
		t.Fatalf("task directory should have been created: %v", err)
	}
	if !info.IsDir() {
		t.Error("task directory path should be a directory")
	}

	// Spec should be inside the new task directory
	specPath := filepath.Join(taskDirPath, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created in new task directory")
	}
}

func TestProjectInit_Stdout_NoSideEffects(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--stdout", "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// No config file should have been created
	configPath := filepath.Join(repo.Dir, configFilename)
	if _, err := os.Stat(configPath); err == nil {
		t.Error(".taskmd.yaml should not have been created with --stdout")
	}

	// No agent files should have been created
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created with --stdout")
	}

	// No spec file should have been created
	specPath := filepath.Join(repo.Dir, specFilename)
	if _, err := os.Stat(specPath); err == nil {
		t.Error("TASKMD_SPEC.md should not have been created with --stdout")
	}
}

func TestProjectInit_CreatesTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	tmplDir := filepath.Join(repo.Dir, ".taskmd", "templates")
	info, err := os.Stat(tmplDir)
	if err != nil {
		t.Fatalf("expected .taskmd/templates/ to be created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .taskmd/templates/ to be a directory")
	}

	// Check that built-in template files exist
	for _, name := range []string{"feature.md", "bug.md", "chore.md"} {
		path := filepath.Join(tmplDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected template %s to be created", name)
		}
	}
}

func TestProjectInit_NoTemplatesFlag(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--no-templates")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	tmplDir := filepath.Join(repo.Dir, ".taskmd", "templates")
	if _, err := os.Stat(tmplDir); err == nil {
		t.Error(".taskmd/templates/ should not have been created with --no-templates")
	}
}

func TestProjectInit_TemplatesNotOverwrittenWithoutForce(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create an existing template
	featurePath := repo.Write(filepath.Join(".taskmd", "templates", "feature.md"), "custom content")

	res := runInit(repo, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Custom template should be unchanged
	content, _ := os.ReadFile(featurePath)
	if string(content) != "custom content" {
		t.Error("existing template should not have been overwritten without --force")
	}
}

func TestProjectInit_TemplatesOverwrittenWithForce(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create an existing template
	featurePath := repo.Write(filepath.Join(".taskmd", "templates", "feature.md"), "custom content")

	res := runInit(repo, "--claude", "--force")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Template should be overwritten with built-in content
	content, _ := os.ReadFile(featurePath)
	if string(content) == "custom content" {
		t.Error("template should have been overwritten with --force")
	}
	if !strings.Contains(string(content), "_template:") {
		t.Error("overwritten template should contain built-in template content")
	}
}

func TestProjectInit_EnsureTaskDir_PathIsFile(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create a file where the task directory should be
	filePath := repo.Write("tasks", "not a directory")

	res := runInitIn(repo, repo.Dir, filePath, "--claude")
	if res.Err == nil {
		t.Fatal("expected error when task-dir path is a file")
	}

	if !strings.Contains(res.Err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got: %v", res.Err)
	}
}

// --- ID Strategy tests ---

func TestProjectInit_IDStrategy_ULID_ConfigAndTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--id-strategy", "ulid")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Check config file includes id section
	configPath := filepath.Join(repo.Dir, configFilename)
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	configStr := string(configContent)
	if !strings.Contains(configStr, "strategy: ulid") {
		t.Errorf("config should contain 'strategy: ulid', got: %s", configStr)
	}
	if !strings.Contains(configStr, "length: 9") {
		t.Errorf("config should contain 'length: 9', got: %s", configStr)
	}

	// Check CLAUDE.md has ULID examples
	claudePath := filepath.Join(repo.Dir, "CLAUDE.md")
	claudeContent, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	claudeStr := string(claudeContent)
	if !strings.Contains(claudeStr, `id: "01h5a3mpk"`) {
		t.Errorf("CLAUDE.md should contain ULID example ID, got: %s", claudeStr)
	}
	if !strings.Contains(claudeStr, "01h5a3mpk-add-user-auth.md") {
		t.Errorf("CLAUDE.md should contain ULID example filename")
	}
	if strings.Contains(claudeStr, `id: "001"`) {
		t.Error("CLAUDE.md should not contain sequential example ID")
	}

	// Check spec has ULID section
	specPath := filepath.Join(repo.Dir, specFilename)
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}
	if !strings.Contains(string(specContent), "ULID") {
		t.Error("spec should contain ULID documentation section")
	}
}

func TestProjectInit_IDStrategy_Random_ConfigAndTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--id-strategy", "random")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	configContent, _ := os.ReadFile(filepath.Join(repo.Dir, configFilename))
	configStr := string(configContent)
	if !strings.Contains(configStr, "strategy: random") {
		t.Errorf("config should contain 'strategy: random', got: %s", configStr)
	}
	if !strings.Contains(configStr, "length: 6") {
		t.Errorf("config should contain 'length: 6', got: %s", configStr)
	}

	claudeContent, _ := os.ReadFile(filepath.Join(repo.Dir, "CLAUDE.md"))
	if !strings.Contains(string(claudeContent), `id: "a3f9x2"`) {
		t.Error("CLAUDE.md should contain random example ID")
	}
	if !strings.Contains(string(claudeContent), "a3f9x2-add-user-auth.md") {
		t.Error("CLAUDE.md should contain random example filename")
	}
}

func TestProjectInit_IDStrategy_Prefixed_ConfigAndTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--id-strategy", "prefixed", "--id-prefix", "dr")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	configContent, _ := os.ReadFile(filepath.Join(repo.Dir, configFilename))
	configStr := string(configContent)
	if !strings.Contains(configStr, "strategy: prefixed") {
		t.Errorf("config should contain 'strategy: prefixed', got: %s", configStr)
	}
	if !strings.Contains(configStr, "prefix: dr") {
		t.Errorf("config should contain 'prefix: dr', got: %s", configStr)
	}

	claudeContent, _ := os.ReadFile(filepath.Join(repo.Dir, "CLAUDE.md"))
	claudeStr := string(claudeContent)
	if !strings.Contains(claudeStr, `id: "dr-001"`) {
		t.Errorf("CLAUDE.md should contain prefixed example ID, got relevant part missing")
	}
	if !strings.Contains(claudeStr, "dr-015-add-user-auth.md") {
		t.Error("CLAUDE.md should contain prefixed example filename")
	}

	specContent, _ := os.ReadFile(filepath.Join(repo.Dir, specFilename))
	if !strings.Contains(string(specContent), "prefixed") {
		t.Error("spec should contain prefixed documentation section")
	}
}

func TestProjectInit_IDStrategy_Sequential_NoIDConfig(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// No --id-strategy flag set, should default to sequential
	res := runInit(repo, "--claude")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	configContent, _ := os.ReadFile(filepath.Join(repo.Dir, configFilename))
	configStr := string(configContent)
	// Sequential is default, so no id: section should be written
	if strings.Contains(configStr, "id:") {
		t.Errorf("config should not contain id: section for sequential strategy, got: %s", configStr)
	}

	// Templates should remain unchanged (sequential examples)
	claudeContent, _ := os.ReadFile(filepath.Join(repo.Dir, "CLAUDE.md"))
	if !bytes.Equal(claudeContent, claudeTemplate) {
		t.Error("CLAUDE.md should match raw template for sequential strategy")
	}
}

func TestProjectInit_IDStrategy_InvalidStrategy(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--id-strategy", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid strategy")
	}
	if !strings.Contains(res.Err.Error(), "invalid --id-strategy") {
		t.Errorf("error should mention invalid strategy, got: %v", res.Err)
	}
}

func TestProjectInit_IDStrategy_PrefixedWithoutPrefix_NonTTY(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--id-strategy", "prefixed")
	if res.Err == nil {
		t.Fatal("expected error for prefixed strategy without prefix in non-TTY mode")
	}
	if !strings.Contains(res.Err.Error(), "id-prefix") {
		t.Errorf("error should mention id-prefix, got: %v", res.Err)
	}
}

func TestProjectInit_IDStrategy_ULID_AllAgents(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--claude", "--gemini", "--codex", "--id-strategy", "ulid")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// All agent files should have ULID examples
	for _, name := range []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md"} {
		content, err := os.ReadFile(filepath.Join(repo.Dir, name))
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		if !strings.Contains(string(content), `id: "01h5a3mpk"`) {
			t.Errorf("%s should contain ULID example ID", name)
		}
	}
}

func TestProjectInit_IDStrategy_Stdout_ShowsReplacedContent(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--stdout", "--claude", "--id-strategy", "ulid")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, `id: "01h5a3mpk"`) {
		t.Error("stdout should contain ULID example ID")
	}
	if strings.Contains(res.Stdout, `id: "001"`) {
		t.Error("stdout should not contain sequential example ID")
	}
}

func TestBuildIDConfigYAML(t *testing.T) {
	tests := []struct {
		name     string
		cfg      idStrategyConfig
		expected string
	}{
		{
			name:     "sequential produces no output",
			cfg:      idStrategyConfig{strategy: "sequential"},
			expected: "",
		},
		{
			name:     "ulid",
			cfg:      idStrategyConfig{strategy: "ulid"},
			expected: "id:\n  strategy: ulid\n  length: 9\n",
		},
		{
			name:     "random",
			cfg:      idStrategyConfig{strategy: "random"},
			expected: "id:\n  strategy: random\n  length: 6\n",
		},
		{
			name:     "prefixed",
			cfg:      idStrategyConfig{strategy: "prefixed", prefix: "cli"},
			expected: "id:\n  strategy: prefixed\n  prefix: cli\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildIDConfigYAML(tt.cfg)
			if got != tt.expected {
				t.Errorf("buildIDConfigYAML() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetIDStrategyExamples(t *testing.T) {
	tests := []struct {
		name            string
		cfg             idStrategyConfig
		wantID          string
		wantFilename    string
		wantFilePattern string
	}{
		{"sequential", idStrategyConfig{strategy: "sequential"}, "001", "015-add-user-auth.md", "NNN-descriptive-title.md"},
		{"ulid", idStrategyConfig{strategy: "ulid"}, "01h5a3mpk", "01h5a3mpk-add-user-auth.md", "ID-descriptive-title.md"},
		{"random", idStrategyConfig{strategy: "random"}, "a3f9x2", "a3f9x2-add-user-auth.md", "ID-descriptive-title.md"},
		{"prefixed", idStrategyConfig{strategy: "prefixed", prefix: "dr"}, "dr-001", "dr-015-add-user-auth.md", "DR-NNN-descriptive-title.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := getIDStrategyExamples(tt.cfg)
			if ex.exampleID != tt.wantID {
				t.Errorf("exampleID = %q, want %q", ex.exampleID, tt.wantID)
			}
			if ex.exampleFilename != tt.wantFilename {
				t.Errorf("exampleFilename = %q, want %q", ex.exampleFilename, tt.wantFilename)
			}
			if ex.filePattern != tt.wantFilePattern {
				t.Errorf("filePattern = %q, want %q", ex.filePattern, tt.wantFilePattern)
			}
		})
	}
}

func TestApplyIDStrategyReplacements(t *testing.T) {
	input := []byte(`id: "001"` + "\n" + `Files follow the pattern ` + "`NNN-descriptive-title.md`" + ` (e.g., ` + "`015-add-user-auth.md`" + `).`)

	examples := getIDStrategyExamples(idStrategyConfig{strategy: "ulid"})
	result := string(applyIDStrategyReplacements(input, examples))

	if !strings.Contains(result, `id: "01h5a3mpk"`) {
		t.Error("should replace example ID")
	}
	if !strings.Contains(result, "01h5a3mpk-add-user-auth.md") {
		t.Error("should replace example filename")
	}
	if !strings.Contains(result, "ID-descriptive-title.md") {
		t.Error("should replace file pattern")
	}
}

func TestApplyIDStrategyReplacements_Sequential_NoOp(t *testing.T) {
	input := []byte(`id: "001"` + "\n" + "015-add-user-auth.md")
	examples := getIDStrategyExamples(idStrategyConfig{strategy: "sequential"})
	result := applyIDStrategyReplacements(input, examples)
	if !bytes.Equal(result, input) {
		t.Error("sequential strategy should return unchanged content")
	}
}

func TestProjectInit_NoSpecNoAgentStillCreatesTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runInit(repo, "--no-spec", "--no-agent")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	tmplDir := filepath.Join(repo.Dir, ".taskmd", "templates")
	if _, err := os.Stat(tmplDir); os.IsNotExist(err) {
		t.Error("expected .taskmd/templates/ to be created even with --no-spec and --no-agent")
	}
}
