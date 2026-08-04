package cli

import (
	"strings"
	"testing"
)

func TestWebStart_NonExistentDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("web", "start", "--task-dir", "/nonexistent/path/that/does/not/exist")
	if res.Err == nil {
		t.Fatal("expected error for non-existent directory")
	}

	if !strings.Contains(res.Err.Error(), "not a valid directory") {
		t.Errorf("expected 'not a valid directory' error, got: %v", res.Err)
	}
}

func TestWebStart_FileInsteadOfDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)
	filePath := repo.Write("not-a-dir.txt", "hello")

	res := repo.Run("web", "start", "--task-dir", filePath)
	if res.Err == nil {
		t.Fatal("expected error when task-dir points to a file")
	}

	if !strings.Contains(res.Err.Error(), "not a valid directory") {
		t.Errorf("expected 'not a valid directory' error, got: %v", res.Err)
	}
}
