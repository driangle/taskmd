package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSpecCommand_WritesToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	specForce = false
	specStdout = false
	dir = tmpDir

	err := runSpec(specCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputPath := filepath.Join(tmpDir, specFilename)
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, specTemplate) {
		t.Error("written content does not match embedded spec")
	}
}

func TestSpecCommand_RefusesOverwriteWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()

	existingPath := filepath.Join(tmpDir, specFilename)
	if err := os.WriteFile(existingPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	specForce = false
	specStdout = false
	dir = tmpDir

	err := runSpec(specCmd, []string{})
	if err == nil {
		t.Fatal("expected error when file already exists")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("already exists")) {
		t.Errorf("error message %q should contain 'already exists'", err.Error())
	}

	// Verify original content was not overwritten
	content, _ := os.ReadFile(existingPath)
	if string(content) != "existing content" {
		t.Error("existing file should not have been overwritten")
	}
}

func TestSpecCommand_OverwritesWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	existingPath := filepath.Join(tmpDir, specFilename)
	if err := os.WriteFile(existingPath, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	specForce = true
	specStdout = false
	dir = tmpDir

	err := runSpec(specCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}

	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(content, specTemplate) {
		t.Error("file should have been overwritten with spec content")
	}
}

func TestSpecCommand_StdoutPrintsWithoutCreatingFile(t *testing.T) {
	tmpDir := t.TempDir()

	specForce = false
	specStdout = true
	dir = tmpDir

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSpec(specCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output != string(specTemplate) {
		t.Error("stdout output does not match embedded spec")
	}

	// Verify no file was created
	outputPath := filepath.Join(tmpDir, specFilename)
	if _, err := os.Stat(outputPath); err == nil {
		t.Errorf("%s should not have been created with --stdout", specFilename)
	}
}

func TestSpecCommand_DirWritesToSpecifiedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "docs")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	specForce = false
	specStdout = false
	dir = subDir

	err := runSpec(specCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputPath := filepath.Join(subDir, specFilename)
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, specTemplate) {
		t.Error("written content does not match embedded spec")
	}
}

func TestSpecCommand_NonExistentDirectoryReturnsError(t *testing.T) {
	specForce = false
	specStdout = false
	dir = "/nonexistent/path/that/does/not/exist"

	err := runSpec(specCmd, []string{})
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("directory does not exist")) {
		t.Errorf("error message %q should contain 'directory does not exist'", err.Error())
	}
}

func TestSpecCommand_ContentMatchesTemplate(t *testing.T) {
	if len(specTemplate) == 0 {
		t.Fatal("embedded spec should not be empty")
	}

	if !bytes.HasPrefix(specTemplate, []byte("# taskmd Specification")) {
		t.Error("spec should start with expected header")
	}
}
