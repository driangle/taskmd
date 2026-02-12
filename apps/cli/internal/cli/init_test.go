package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand_WritesToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Reset all flags
	initForce = false
	initStdout = false
	initClaude = false
	initGemini = false
	initCodex = false
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create CLAUDE.md by default
	outputPath := filepath.Join(tmpDir, "CLAUDE.md")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, claudeTemplate) {
		t.Error("written content does not match embedded template")
	}
}

func TestInitCommand_RefusesOverwriteWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an existing CLAUDE.md
	existingPath := filepath.Join(tmpDir, "CLAUDE.md")
	err := os.WriteFile(existingPath, []byte("existing content"), 0644)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	initForce = false
	initStdout = false
	initClaude = false
	initGemini = false
	initCodex = false
	dir = tmpDir

	err = runInit(initCmd, []string{})
	if err == nil {
		t.Fatal("expected error when CLAUDE.md already exists")
	}

	expectedMsg := "CLAUDE.md already exists"
	if !bytes.Contains([]byte(err.Error()), []byte(expectedMsg)) {
		t.Errorf("error message %q should contain %q", err.Error(), expectedMsg)
	}

	// Verify original content was not overwritten
	content, _ := os.ReadFile(existingPath)
	if string(content) != "existing content" {
		t.Error("existing file should not have been overwritten")
	}
}

func TestInitCommand_OverwritesWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an existing CLAUDE.md
	existingPath := filepath.Join(tmpDir, "CLAUDE.md")
	err := os.WriteFile(existingPath, []byte("old content"), 0644)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	initForce = true
	initStdout = false
	initClaude = false
	initGemini = false
	initCodex = false
	dir = tmpDir

	err = runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}

	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(content, claudeTemplate) {
		t.Error("file should have been overwritten with template content")
	}
}

func TestInitCommand_StdoutPrintsWithoutCreatingFile(t *testing.T) {
	tmpDir := t.TempDir()

	initForce = false
	initStdout = true
	initClaude = false
	initGemini = false
	initCodex = false
	dir = tmpDir

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(initCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should output Claude template by default
	if output != string(claudeTemplate) {
		t.Error("stdout output does not match embedded template")
	}

	// Verify no file was created
	outputPath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(outputPath); err == nil {
		t.Error("CLAUDE.md should not have been created with --stdout")
	}
}

func TestInitCommand_DirWritesToSpecifiedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "my-project")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	initForce = false
	initStdout = false
	initClaude = false
	initGemini = false
	initCodex = false
	dir = subDir

	err = runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputPath := filepath.Join(subDir, "CLAUDE.md")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, claudeTemplate) {
		t.Error("written content does not match embedded template")
	}
}

func TestInitCommand_NonExistentDirectoryReturnsError(t *testing.T) {
	initForce = false
	initStdout = false
	initClaude = false
	initGemini = false
	initCodex = false
	dir = "/nonexistent/path/that/does/not/exist"

	err := runInit(initCmd, []string{})
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}

	expectedMsg := "directory does not exist"
	if !bytes.Contains([]byte(err.Error()), []byte(expectedMsg)) {
		t.Errorf("error message %q should contain %q", err.Error(), expectedMsg)
	}
}

func TestInitCommand_ContentMatchesTemplate(t *testing.T) {
	// Verify the embedded template is not empty
	if len(claudeTemplate) == 0 {
		t.Fatal("embedded template should not be empty")
	}

	// Verify it starts with expected content
	if !bytes.HasPrefix(claudeTemplate, []byte("# Working with Taskmd Tasks")) {
		t.Error("template should start with expected header")
	}
}

func TestInitCommand_GeminiFlag(t *testing.T) {
	tmpDir := t.TempDir()

	initForce = false
	initStdout = false
	initClaude = false
	initGemini = true
	initCodex = false
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create GEMINI.md
	outputPath := filepath.Join(tmpDir, "GEMINI.md")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, geminiTemplate) {
		t.Error("written content does not match Gemini template")
	}

	// Should not create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created when only --gemini is specified")
	}
}

func TestInitCommand_CodexFlag(t *testing.T) {
	tmpDir := t.TempDir()

	initForce = false
	initStdout = false
	initClaude = false
	initGemini = false
	initCodex = true
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create AGENTS.md
	outputPath := filepath.Join(tmpDir, "AGENTS.md")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, geminiTemplate) {
		t.Error("written content does not match template")
	}

	// Should not create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should not have been created when only --codex is specified")
	}
}

func TestInitCommand_MultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()

	initForce = false
	initStdout = false
	initClaude = true
	initGemini = true
	initCodex = true
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create CLAUDE.md, GEMINI.md, and AGENTS.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	claudeContent, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if !bytes.Equal(claudeContent, claudeTemplate) {
		t.Error("CLAUDE.md content does not match template")
	}

	geminiPath := filepath.Join(tmpDir, "GEMINI.md")
	geminiContent, err := os.ReadFile(geminiPath)
	if err != nil {
		t.Fatalf("failed to read GEMINI.md: %v", err)
	}
	if !bytes.Equal(geminiContent, geminiTemplate) {
		t.Error("GEMINI.md content does not match template")
	}

	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}
	if !bytes.Equal(agentsContent, codexTemplate) {
		t.Error("AGENTS.md content does not match template")
	}
}

func TestInitCommand_MultipleAgentsWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing files
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	geminiPath := filepath.Join(tmpDir, "GEMINI.md")
	os.WriteFile(claudePath, []byte("old claude"), 0644)
	os.WriteFile(geminiPath, []byte("old gemini"), 0644)

	initForce = true
	initStdout = false
	initClaude = true
	initGemini = true
	initCodex = false
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify files were overwritten
	claudeContent, _ := os.ReadFile(claudePath)
	if !bytes.Equal(claudeContent, claudeTemplate) {
		t.Error("CLAUDE.md should have been overwritten")
	}

	geminiContent, _ := os.ReadFile(geminiPath)
	if !bytes.Equal(geminiContent, geminiTemplate) {
		t.Error("GEMINI.md should have been overwritten")
	}
}

func TestInitCommand_ClaudeExplicitFlag(t *testing.T) {
	tmpDir := t.TempDir()

	initForce = false
	initStdout = false
	initClaude = true
	initGemini = false
	initCodex = false
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create only CLAUDE.md
	outputPath := filepath.Join(tmpDir, "CLAUDE.md")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, claudeTemplate) {
		t.Error("written content does not match Claude template")
	}

	// Should not create other files
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		t.Error("AGENTS.md should not have been created")
	}
}
