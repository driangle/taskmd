package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand_WritesToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	initForce = false
	initStdout = false
	dir = tmpDir

	err := runInit(initCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
