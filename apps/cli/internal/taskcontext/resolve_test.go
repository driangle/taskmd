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

func TestResolve_DirectoryExpansionRecursive(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pkg/a.go", "package pkg\n")
	writeFile(t, root, "pkg/sub/b.go", "package sub\n")
	writeFile(t, root, "pkg/sub/deep/c.go", "package deep\n")

	task := &model.Task{
		ID:      "016",
		Title:   "Recursive dir expansion",
		Context: []string{"pkg/"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot: root,
		Resolve:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 3 {
		t.Fatalf("expected 3 files after recursive expansion, got %d", len(result.Files))
	}

	paths := make(map[string]bool)
	for _, f := range result.Files {
		paths[f.Path] = true
		if !f.Exists {
			t.Errorf("expected file %s to exist", f.Path)
		}
	}
	for _, expected := range []string{"pkg/a.go", "pkg/sub/b.go", "pkg/sub/deep/c.go"} {
		if !paths[expected] {
			t.Errorf("expected %s in expanded files, got %v", expected, paths)
		}
	}
}

func TestResolve_GeneratedFilesSkipped(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n")
	writeFile(t, root, "go.sum", "github.com/foo/bar v1.0.0 h1:abc123=\n")
	writeFile(t, root, "package-lock.json", `{"lockfileVersion": 3}`)

	task := &model.Task{
		ID:      "018",
		Title:   "Generated skip test",
		Context: []string{"src/main.go", "go.sum", "package-lock.json"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot:    root,
		IncludeContent: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result.Files))
	}

	// Text file should have content
	if result.Files[0].Content == "" {
		t.Error("expected main.go to have content")
	}
	if result.Files[0].Generated {
		t.Error("expected main.go to not be marked generated")
	}

	// go.sum should be marked generated, no content
	if result.Files[1].Content != "" {
		t.Error("expected go.sum to have no content")
	}
	if !result.Files[1].Generated {
		t.Error("expected go.sum to be marked generated")
	}

	// package-lock.json should be marked generated, no content
	if result.Files[2].Content != "" {
		t.Error("expected package-lock.json to have no content")
	}
	if !result.Files[2].Generated {
		t.Error("expected package-lock.json to be marked generated")
	}
}

func TestResolve_ExpandSkipsJunkDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pkg/a.go", "package pkg\n")
	writeFile(t, root, "pkg/node_modules/dep/index.js", "module.exports = {}\n")
	writeFile(t, root, "pkg/.git/config", "[core]\n")
	writeFile(t, root, "pkg/sub/b.go", "package sub\n")

	task := &model.Task{
		ID:      "019",
		Title:   "Junk dir skip test",
		Context: []string{"pkg/"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot: root,
		Resolve:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := make(map[string]bool)
	for _, f := range result.Files {
		paths[f.Path] = true
	}

	if !paths["pkg/a.go"] {
		t.Error("expected pkg/a.go to be included")
	}
	if !paths["pkg/sub/b.go"] {
		t.Error("expected pkg/sub/b.go to be included")
	}
	if paths["pkg/node_modules/dep/index.js"] {
		t.Error("expected node_modules files to be skipped")
	}
	if paths["pkg/.git/config"] {
		t.Error("expected .git files to be skipped")
	}
}

func TestResolve_BinaryFilesSkipped(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n")
	// Write a binary file (contains null bytes)
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x00, 0x1A, 0x0A}
	fullPath := filepath.Join(root, "src", "image.png")
	if err := os.WriteFile(fullPath, binaryContent, 0o644); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	task := &model.Task{
		ID:      "017",
		Title:   "Binary skip test",
		Context: []string{"src/main.go", "src/image.png"},
	}

	result, err := Resolve(task, Options{
		ProjectRoot:    root,
		IncludeContent: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}

	textFile := result.Files[0]
	if textFile.Content == "" {
		t.Error("expected text file to have content inlined")
	}
	if textFile.Binary {
		t.Error("expected text file to not be marked binary")
	}

	binaryFile := result.Files[1]
	if binaryFile.Content != "" {
		t.Error("expected binary file to have no content inlined")
	}
	if !binaryFile.Binary {
		t.Error("expected binary file to be marked binary")
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

func TestResolve_BodyAlwaysIncluded(t *testing.T) {
	root := t.TempDir()

	task := &model.Task{
		ID:    "014",
		Title: "Body without include-content",
		Body:  "## Description\n\nThis body should always appear.",
	}

	result, err := Resolve(task, Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TaskBody != "## Description\n\nThis body should always appear." {
		t.Errorf("expected task body to be included without --include-content, got %q", result.TaskBody)
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

func TestResolve_IsDirFlag(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main.go", "package main\n")
	// "src/" is a directory, "src/main.go" is a file

	task := &model.Task{
		ID:      "015",
		Title:   "IsDir check",
		Context: []string{"src/", "src/main.go"},
	}

	result, err := Resolve(task, Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}

	dirEntry := result.Files[0]
	if !dirEntry.IsDir {
		t.Errorf("expected src/ to have IsDir=true, got false")
	}
	if !dirEntry.Exists {
		t.Errorf("expected src/ to exist")
	}

	fileEntry := result.Files[1]
	if fileEntry.IsDir {
		t.Errorf("expected src/main.go to have IsDir=false, got true")
	}
	if !fileEntry.Exists {
		t.Errorf("expected src/main.go to exist")
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
