package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetProjectInitFlags(tmpDir string) {
	projectInitForce = false
	projectInitStdout = false
	projectInitClaude = false
	projectInitGemini = false
	projectInitCodex = false
	projectInitNoSpec = false
	projectInitNoAgent = false
	dir = tmpDir
}

func TestProjectInit_DefaultWritesBothFiles(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create CLAUDE.md (default agent)
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if !bytes.Equal(content, claudeTemplate) {
		t.Error("CLAUDE.md content does not match template")
	}

	// Should create TASKMD_SPEC.md
	specPath := filepath.Join(tmpDir, specFilename)
	content, err = os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", specFilename, err)
	}
	if !bytes.Equal(content, specTemplate) {
		t.Error("TASKMD_SPEC.md content does not match template")
	}
}

func TestProjectInit_GeminiFlag(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitGemini = true

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create GEMINI.md
	geminiPath := filepath.Join(tmpDir, "GEMINI.md")
	content, err := os.ReadFile(geminiPath)
	if err != nil {
		t.Fatalf("failed to read GEMINI.md: %v", err)
	}
	if !bytes.Equal(content, geminiTemplate) {
		t.Error("GEMINI.md content does not match template")
	}

	// Should create TASKMD_SPEC.md
	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}

	// Should NOT create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created when --gemini is specified")
	}
}

func TestProjectInit_MultipleAgentFlags(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitClaude = true
	projectInitGemini = true

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create CLAUDE.md, GEMINI.md, and TASKMD_SPEC.md
	for _, name := range []string{"CLAUDE.md", "GEMINI.md", specFilename} {
		path := filepath.Join(tmpDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s should have been created", name)
		}
	}
}

func TestProjectInit_NoSpecFlag(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitNoSpec = true

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should have been created")
	}

	// Should NOT create TASKMD_SPEC.md
	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); err == nil {
		t.Error("TASKMD_SPEC.md should not have been created with --no-spec")
	}
}

func TestProjectInit_NoAgentFlag(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitNoAgent = true

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create TASKMD_SPEC.md
	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}

	// Should NOT create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created with --no-agent")
	}
}

func TestProjectInit_NoSpecAndNoAgentIsError(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitNoSpec = true
	projectInitNoAgent = true

	err := runProjectInit(projectInitCmd, []string{})
	if err == nil {
		t.Fatal("expected error when both --no-spec and --no-agent are set")
	}

	if !strings.Contains(err.Error(), "nothing to do") {
		t.Errorf("error message %q should contain 'nothing to do'", err.Error())
	}
}

func TestProjectInit_ForceOverwritesExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitForce = true

	// Create existing files
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	specPath := filepath.Join(tmpDir, specFilename)
	os.WriteFile(claudePath, []byte("old claude"), 0644)
	os.WriteFile(specPath, []byte("old spec"), 0644)

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify files were overwritten
	claudeContent, _ := os.ReadFile(claudePath)
	if !bytes.Equal(claudeContent, claudeTemplate) {
		t.Error("CLAUDE.md should have been overwritten")
	}

	specContent, _ := os.ReadFile(specPath)
	if !bytes.Equal(specContent, specTemplate) {
		t.Error("TASKMD_SPEC.md should have been overwritten")
	}
}

func TestProjectInit_ExistingFilesSkippedWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)

	// Create an existing CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(claudePath, []byte("existing claude"), 0644)

	// Capture stderr for the warning
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := runProjectInit(projectInitCmd, []string{})

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("expected no error when skipping existing files, got: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	// Should have a warning about skipping
	if !strings.Contains(stderrOutput, "Skipped") {
		t.Errorf("expected skip warning on stderr, got: %q", stderrOutput)
	}

	// Original file should be unchanged
	content, _ := os.ReadFile(claudePath)
	if string(content) != "existing claude" {
		t.Error("existing CLAUDE.md should not have been overwritten")
	}

	// TASKMD_SPEC.md should still be created (it didn't exist)
	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created even though CLAUDE.md was skipped")
	}
}

func TestProjectInit_DirFlag(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "my-project")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	resetProjectInitFlags(subDir)

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

func TestProjectInit_NonExistentDirReturnsError(t *testing.T) {
	resetProjectInitFlags("/nonexistent/path/that/does/not/exist")

	err := runProjectInit(projectInitCmd, []string{})
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}

	if !strings.Contains(err.Error(), "directory does not exist") {
		t.Errorf("error message %q should contain 'directory does not exist'", err.Error())
	}
}

func TestProjectInit_StdoutPrintsWithoutCreatingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitStdout = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runProjectInit(projectInitCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain both agent template and spec template
	if !strings.Contains(output, string(claudeTemplate)) {
		t.Error("stdout output should contain Claude template")
	}
	if !strings.Contains(output, string(specTemplate)) {
		t.Error("stdout output should contain spec template")
	}

	// No files should have been created
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created with --stdout")
	}
	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); err == nil {
		t.Error("TASKMD_SPEC.md should not have been created with --stdout")
	}
}

func TestProjectInit_CodexFlag(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitCodex = true

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create AGENTS.md and TASKMD_SPEC.md
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		t.Error("AGENTS.md should have been created")
	}

	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}

	// Should NOT create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created when --codex is specified")
	}
}

func TestProjectInit_AllAgentFlags(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitClaude = true
	projectInitGemini = true
	projectInitCodex = true

	err := runProjectInit(projectInitCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md", specFilename}
	for _, name := range expected {
		path := filepath.Join(tmpDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s should have been created", name)
		}
	}
}

func TestProjectInit_PartialSkipStillCreatesOthers(t *testing.T) {
	tmpDir := t.TempDir()
	resetProjectInitFlags(tmpDir)
	projectInitClaude = true
	projectInitGemini = true

	// Create only CLAUDE.md as existing
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	os.WriteFile(claudePath, []byte("existing"), 0644)

	// Suppress stderr warnings
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	err := runProjectInit(projectInitCmd, []string{})

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CLAUDE.md should be unchanged (skipped)
	content, _ := os.ReadFile(claudePath)
	if string(content) != "existing" {
		t.Error("existing CLAUDE.md should not have been overwritten")
	}

	// GEMINI.md should be created
	geminiPath := filepath.Join(tmpDir, "GEMINI.md")
	if _, err := os.Stat(geminiPath); os.IsNotExist(err) {
		t.Error("GEMINI.md should have been created")
	}

	// TASKMD_SPEC.md should be created
	specPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("TASKMD_SPEC.md should have been created")
	}
}
