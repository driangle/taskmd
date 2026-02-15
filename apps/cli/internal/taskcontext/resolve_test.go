package taskcontext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestResolve_TouchesOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n")
	writeFile(t, root, "src/util.go", "package main\n")

	task := &model.Task{
		ID:      "001",
		Title:   "Test task",
		Touches: []string{"cli"},
	}
	scopes := ScopeMap{
		"cli": {"src/main.go", "src/util.go"},
	}

	result, err := Resolve(task, Options{Scopes: scopes, ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
	assertFileEntry(t, result.Files[0], "src/main.go", "scope:cli", true)
	assertFileEntry(t, result.Files[1], "src/util.go", "scope:cli", true)
}

func TestResolve_ContextOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docs/readme.md", "# README\n")

	task := &model.Task{
		ID:      "002",
		Title:   "Context only task",
		Context: []string{"docs/readme.md"},
	}

	result, err := Resolve(task, Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	assertFileEntry(t, result.Files[0], "docs/readme.md", "explicit", true)
}

func TestResolve_BothSources(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n")
	writeFile(t, root, "docs/notes.md", "# Notes\n")

	task := &model.Task{
		ID:      "003",
		Title:   "Both sources",
		Touches: []string{"cli"},
		Context: []string{"docs/notes.md"},
	}
	scopes := ScopeMap{"cli": {"src/main.go"}}

	result, err := Resolve(task, Options{Scopes: scopes, ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
	assertFileEntry(t, result.Files[0], "src/main.go", "scope:cli", true)
	assertFileEntry(t, result.Files[1], "docs/notes.md", "explicit", true)
}

func TestResolve_Deduplication(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n")

	task := &model.Task{
		ID:      "004",
		Title:   "Dedup task",
		Touches: []string{"cli"},
		Context: []string{"src/main.go"},
	}
	scopes := ScopeMap{"cli": {"src/main.go"}}

	result, err := Resolve(task, Options{Scopes: scopes, ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file after dedup, got %d", len(result.Files))
	}
	// Scope entry comes first, so it wins
	assertFileEntry(t, result.Files[0], "src/main.go", "scope:cli", true)
}

func TestResolve_DirectoryExpansion(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pkg/a.go", "package pkg\n")
	writeFile(t, root, "pkg/b.go", "package pkg\n")

	task := &model.Task{
		ID:      "005",
		Title:   "Dir expansion",
		Touches: []string{"pkg"},
	}
	scopes := ScopeMap{"pkg": {"pkg/"}}

	result, err := Resolve(task, Options{
		Scopes:      scopes,
		ProjectRoot: root,
		Resolve:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files after expansion, got %d", len(result.Files))
	}
	// Both files should exist and have scope source
	for _, f := range result.Files {
		if f.Source != "scope:pkg" {
			t.Errorf("expected source scope:pkg, got %s", f.Source)
		}
		if !f.Exists {
			t.Errorf("expected file %s to exist", f.Path)
		}
	}
}

func TestResolve_IncludeContent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n\nfunc main() {}\n")

	task := &model.Task{
		ID:      "006",
		Title:   "Content task",
		Body:    "## Description\n\nSome task body.",
		Context: []string{"src/main.go"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot:    root,
		IncludeContent: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TaskBody != "## Description\n\nSome task body." {
		t.Errorf("unexpected task body: %q", result.TaskBody)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	f := result.Files[0]
	if f.Content != "package main\n\nfunc main() {}\n" {
		t.Errorf("unexpected content: %q", f.Content)
	}
	if f.Lines != 3 {
		t.Errorf("expected 3 lines, got %d", f.Lines)
	}
}

func TestResolve_NonExistentFiles(t *testing.T) {
	root := t.TempDir()

	task := &model.Task{
		ID:      "007",
		Title:   "Missing files",
		Context: []string{"does/not/exist.go"},
	}

	result, err := Resolve(task, Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if result.Files[0].Exists {
		t.Error("expected file to not exist")
	}
}

func TestResolve_MaxFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.go", "a")
	writeFile(t, root, "b.go", "b")
	writeFile(t, root, "c.go", "c")

	task := &model.Task{
		ID:      "008",
		Title:   "Max files",
		Context: []string{"a.go", "b.go", "c.go"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot: root,
		MaxFiles:    2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files (capped), got %d", len(result.Files))
	}
}

func TestResolve_EmptyTask(t *testing.T) {
	root := t.TempDir()

	task := &model.Task{
		ID:    "009",
		Title: "Empty",
	}

	result, err := Resolve(task, Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(result.Files))
	}
}

func TestResolve_UnknownScope(t *testing.T) {
	root := t.TempDir()

	task := &model.Task{
		ID:      "010",
		Title:   "Unknown scope",
		Touches: []string{"nonexistent"},
	}
	scopes := ScopeMap{"cli": {"src/main.go"}}

	result, err := Resolve(task, Options{Scopes: scopes, ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 0 {
		t.Fatalf("expected 0 files for unknown scope, got %d", len(result.Files))
	}
}

func TestResolve_IncludeContentNoBody(t *testing.T) {
	root := t.TempDir()

	task := &model.Task{
		ID:    "011",
		Title: "No body",
	}

	result, err := Resolve(task, Options{
		ProjectRoot:    root,
		IncludeContent: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TaskBody != "" {
		t.Errorf("expected empty task body, got %q", result.TaskBody)
	}
}

func TestResolve_ContentNotInlinedForDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pkg/a.go", "package pkg\n")

	task := &model.Task{
		ID:      "012",
		Title:   "Dir content",
		Context: []string{"pkg/"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot:    root,
		IncludeContent: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The directory entry should have no content inlined
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if result.Files[0].Content != "" {
		t.Error("expected no content for directory entry")
	}
}

func TestResolve_MultipleScopes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "web/app.js", "console.log('app')\n")
	writeFile(t, root, "cli/main.go", "package main\n")

	task := &model.Task{
		ID:      "013",
		Title:   "Multi scope",
		Touches: []string{"web", "cli"},
	}
	scopes := ScopeMap{
		"web": {"web/app.js"},
		"cli": {"cli/main.go"},
	}

	result, err := Resolve(task, Options{Scopes: scopes, ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
}

// helpers

func writeFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func assertFileEntry(t *testing.T, f FileEntry, path, source string, exists bool) {
	t.Helper()
	if f.Path != path {
		t.Errorf("path: want %q, got %q", path, f.Path)
	}
	if f.Source != source {
		t.Errorf("source: want %q, got %q", source, f.Source)
	}
	if f.Exists != exists {
		t.Errorf("exists: want %v, got %v", exists, f.Exists)
	}
}
