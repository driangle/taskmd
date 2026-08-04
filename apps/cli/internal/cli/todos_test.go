package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/todos"
)

// todosSourceFiles is the canonical set of source files with TODO-style comments
// scanned by the todos command tests. Kept inline (these are source files, not
// task files) because their comment markers are the subject of the tests.
func todosSourceFiles() map[string]string {
	return map[string]string{
		"main.go": `package main

// TODO: implement main logic
func main() {}

// FIXME: handle error case
func process() error { return nil }
`,
		"app.py": `# HACK: workaround for upstream bug
import os
`,
		"style.css": `/* NOTE: using hardcoded values */
.container { width: 100%; }
`,
	}
}

// todosTableStdout renders items through outputTodosTable and returns stdout,
// failing on error. Color is disabled for deterministic output.
func todosTableStdout(t *testing.T, items []todos.TodoItem, columns []string, rich bool) string {
	t.Helper()
	noColor = true
	var runErr error
	stdout, _ := captureOutput(t, func() {
		runErr = outputTodosTable(items, columns, rich)
	})
	if runErr != nil {
		t.Fatalf("outputTodosTable failed: %v", runErr)
	}
	return stdout
}

// todosListJSON runs `todos list` for a JSON result, fails on error, and parses.
func todosListJSON(t *testing.T, res cliResult) []todos.TodoItem {
	t.Helper()
	if res.Err != nil {
		t.Fatalf("todos list failed: %v", res.Err)
	}
	var parsed []todos.TodoItem
	if err := json.Unmarshal([]byte(res.Stdout), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}
	return parsed
}

func TestTodosList_TableOutput(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	items, err := todos.Scan(todos.ScanOptions{Dir: repo.Dir})
	if err != nil {
		t.Fatal(err)
	}

	output := todosTableStdout(t, items, defaultColumns, false)

	if !strings.Contains(output, "FILE") || !strings.Contains(output, "LINE") {
		t.Error("expected header with FILE and LINE")
	}
	if !strings.Contains(output, "TAG") || !strings.Contains(output, "TEXT") {
		t.Error("expected header with TAG and TEXT")
	}
	if !strings.Contains(output, "ID") {
		t.Error("expected header with ID")
	}
	if !strings.Contains(output, "TODO") {
		t.Error("expected TODO marker in output")
	}
	if !strings.Contains(output, "FIXME") {
		t.Error("expected FIXME marker in output")
	}
}

func TestTodosList_TableOutputEmpty(t *testing.T) {
	output := todosTableStdout(t, nil, defaultColumns, false)
	if !strings.Contains(output, "No TODO comments found") {
		t.Error("expected 'No TODO comments found' message")
	}
}

func TestTodosList_JSONOutput(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	items, err := todos.Scan(todos.ScanOptions{Dir: repo.Dir})
	if err != nil {
		t.Fatal(err)
	}

	var runErr error
	stdout, _ := captureOutput(t, func() { runErr = WriteJSON(os.Stdout, items) })
	if runErr != nil {
		t.Fatalf("WriteJSON failed: %v", runErr)
	}

	var parsed []todos.TodoItem
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, stdout)
	}

	if len(parsed) == 0 {
		t.Fatal("expected items in JSON output")
	}

	// Verify fields are present
	for _, item := range parsed {
		if item.FilePath == "" {
			t.Error("expected non-empty file path")
		}
		if item.Line == 0 {
			t.Error("expected non-zero line number")
		}
		if item.Marker == "" {
			t.Error("expected non-empty marker")
		}
	}
}

func TestTodosList_YAMLOutput(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	items, err := todos.Scan(todos.ScanOptions{Dir: repo.Dir})
	if err != nil {
		t.Fatal(err)
	}

	var runErr error
	output, _ := captureOutput(t, func() { runErr = WriteYAML(os.Stdout, items) })
	if runErr != nil {
		t.Fatalf("WriteYAML failed: %v", runErr)
	}

	if !strings.Contains(output, "file:") || !strings.Contains(output, "line:") {
		t.Error("expected YAML with file and line fields")
	}
	if !strings.Contains(output, "tag:") || !strings.Contains(output, "text:") {
		t.Error("expected YAML with tag and text fields")
	}
}

func TestTodosList_MarkerFilter(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	items, err := todos.Scan(todos.ScanOptions{
		Dir:     repo.Dir,
		Markers: []string{"TODO"},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if item.Marker != "TODO" {
			t.Errorf("expected only TODO markers, got %s", item.Marker)
		}
	}
}

func TestTodosList_InvalidMarker(t *testing.T) {
	err := validateMarkers([]string{"INVALID"})
	if err == nil {
		t.Fatal("expected error for invalid marker")
	}

	if !strings.Contains(err.Error(), "invalid marker") {
		t.Errorf("expected 'invalid marker' in error, got: %s", err.Error())
	}
}

func TestTodosList_ValidMarkers(t *testing.T) {
	err := validateMarkers(todos.DefaultMarkers)
	if err != nil {
		t.Fatalf("expected no error for valid markers, got: %v", err)
	}
}

func TestTodosList_RunCommand(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	res := repo.Run("todos", "list", "--dir", repo.Dir, "--format", "json")
	parsed := todosListJSON(t, res)

	if len(parsed) == 0 {
		t.Fatal("expected items from runTodosList")
	}
}

func TestTodosList_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("todos", "list", "--dir", repo.Dir, "--format", "json")
	parsed := todosListJSON(t, res)

	if len(parsed) != 0 {
		t.Fatalf("expected 0 items for empty dir, got %d", len(parsed))
	}
}

func TestTodosList_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("todos", "list", "--dir", repo.Dir, "--format", "xml")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %s", res.Err.Error())
	}
}

func TestMergeConfigExcludes_ConfigOnly(t *testing.T) {
	viper.Set("todos.exclude", []string{"*_test.go", "*.generated.*"})
	defer viper.Set("todos.exclude", nil)

	result := mergeConfigExcludes(nil)

	if len(result) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(result))
	}
	if result[0] != "*_test.go" || result[1] != "*.generated.*" {
		t.Errorf("unexpected patterns: %v", result)
	}
}

func TestMergeConfigExcludes_CLIOnly(t *testing.T) {
	viper.Set("todos.exclude", nil)

	result := mergeConfigExcludes([]string{"vendor/*"})

	if len(result) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(result))
	}
	if result[0] != "vendor/*" {
		t.Errorf("unexpected pattern: %v", result)
	}
}

func TestMergeConfigExcludes_BothMerged(t *testing.T) {
	viper.Set("todos.exclude", []string{"*_test.go"})
	defer viper.Set("todos.exclude", nil)

	result := mergeConfigExcludes([]string{"vendor/*"})

	if len(result) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(result))
	}
	if result[0] != "vendor/*" {
		t.Errorf("expected CLI pattern first, got %s", result[0])
	}
	if result[1] != "*_test.go" {
		t.Errorf("expected config pattern second, got %s", result[1])
	}
}

func TestMergeConfigExcludes_NeitherSet(t *testing.T) {
	viper.Set("todos.exclude", nil)

	result := mergeConfigExcludes(nil)

	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestTodosList_ConfigExcludePattern(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"main.go": `package main
// TODO: keep this
`,
		"main_test.go": `package main
// TODO: exclude this via config
`,
	})

	res := repo.RunWith(func() { viper.Set("todos.exclude", []string{"*_test.go"}) },
		"todos", "list", "--dir", repo.Dir, "--format", "json")
	parsed := todosListJSON(t, res)

	if len(parsed) != 1 {
		t.Fatalf("expected 1 item (test file excluded by config), got %d", len(parsed))
	}
	if parsed[0].FilePath != "main.go" {
		t.Errorf("expected main.go, got %s", parsed[0].FilePath)
	}
}

func TestTodosList_ConfigAndCLIExcludeCombine(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"main.go": `package main
// TODO: keep this
`,
		"main_test.go": `package main
// TODO: exclude via config
`,
		"app.py": `# TODO: exclude via CLI flag
`,
	})

	res := repo.RunWith(func() { viper.Set("todos.exclude", []string{"*_test.go"}) },
		"todos", "list", "--dir", repo.Dir, "--format", "json", "--exclude", "*.py")
	parsed := todosListJSON(t, res)

	if len(parsed) != 1 {
		t.Fatalf("expected 1 item (test+py excluded), got %d", len(parsed))
	}
	if parsed[0].FilePath != "main.go" {
		t.Errorf("expected main.go, got %s", parsed[0].FilePath)
	}
}

func TestTodosList_ConfigExcludePathPattern(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"main.go": `package main
// TODO: keep this
`,
		"sub/deep.go": `package sub
// TODO: exclude via path pattern
`,
	})

	res := repo.RunWith(func() { viper.Set("todos.exclude", []string{"sub/*.go"}) },
		"todos", "list", "--dir", repo.Dir, "--format", "json")
	parsed := todosListJSON(t, res)

	if len(parsed) != 1 {
		t.Fatalf("expected 1 item (sub/*.go excluded by config), got %d", len(parsed))
	}
	if parsed[0].FilePath != "main.go" {
		t.Errorf("expected main.go, got %s", parsed[0].FilePath)
	}
}

func TestTodosList_NoConfigExcludeUnchangedBehavior(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"main.go": `package main
// TODO: one
`,
		"main_test.go": `package main
// TODO: two
`,
	})

	res := repo.Run("todos", "list", "--dir", repo.Dir, "--format", "json")
	parsed := todosListJSON(t, res)

	if len(parsed) != 2 {
		t.Fatalf("expected 2 items (no config excludes), got %d", len(parsed))
	}
}

func TestTodosList_JSONOutputNewFields(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	items, err := todos.Scan(todos.ScanOptions{Dir: repo.Dir})
	if err != nil {
		t.Fatal(err)
	}

	var runErr error
	output, _ := captureOutput(t, func() { runErr = WriteJSON(os.Stdout, items) })
	if runErr != nil {
		t.Fatalf("WriteJSON failed: %v", runErr)
	}

	// Check that new fields are present in JSON
	if !strings.Contains(output, `"id"`) {
		t.Error("expected 'id' field in JSON output")
	}
	if !strings.Contains(output, `"column"`) {
		t.Error("expected 'column' field in JSON output")
	}
	if !strings.Contains(output, `"language"`) {
		t.Error("expected 'language' field in JSON output")
	}
	if !strings.Contains(output, `"tag"`) {
		t.Error("expected 'tag' field in JSON output")
	}

	// Verify fields parse correctly
	var parsed []todos.TodoItem
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	for _, item := range parsed {
		if item.ID == "" {
			t.Error("expected non-empty ID")
		}
		if len(item.ID) != 12 {
			t.Errorf("expected 12-char ID, got %d: %q", len(item.ID), item.ID)
		}
		if item.Language == "" {
			t.Error("expected non-empty language")
		}
	}
}

func TestTodosList_RawTextFlag(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	res := repo.Run("todos", "list", "--dir", repo.Dir, "--format", "json", "--raw-text")
	if res.Err != nil {
		t.Fatalf("runTodosList failed: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, `"raw_text"`) {
		t.Error("expected 'raw_text' field in JSON output when --raw-text is set")
	}

	var parsed []todos.TodoItem
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	for _, item := range parsed {
		if item.RawText == "" {
			t.Errorf("expected non-empty raw_text for %s:%d", item.FilePath, item.Line)
		}
	}
}

func TestTodosList_NoRawTextByDefault(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	res := repo.Run("todos", "list", "--dir", repo.Dir, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runTodosList failed: %v", res.Err)
	}

	if strings.Contains(res.Stdout, `"raw_text"`) {
		t.Error("did not expect 'raw_text' in JSON output by default")
	}
}

func TestTodosList_RichTableOutput(t *testing.T) {
	repo := newTaskRepo(t, todosSourceFiles())

	items, err := todos.Scan(todos.ScanOptions{Dir: repo.Dir})
	if err != nil {
		t.Fatal(err)
	}

	output := todosTableStdout(t, items, richColumns, true)

	if !strings.Contains(output, "SCOPE") {
		t.Error("expected SCOPE column in rich table output")
	}
	if !strings.Contains(output, "AGE") {
		t.Error("expected AGE column in rich table output")
	}
	if !strings.Contains(output, "AUTHOR") {
		t.Error("expected AUTHOR column in rich table output")
	}
}
