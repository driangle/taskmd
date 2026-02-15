package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/search"
)

func resetSearchFlags() {
	searchFormat = "table"
	taskDir = "."
	noColor = true
}

func createSearchTestFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	files := map[string]string{
		"001-auth.md": `---
id: "001"
title: "Implement authentication system"
status: pending
priority: high
tags:
  - security
created: 2026-01-01
---

# Implement authentication system

Add JWT-based authentication with login and logout endpoints.
Include token refresh and session management.
`,
		"002-deploy.md": `---
id: "002"
title: "Set up deployment pipeline"
status: in-progress
priority: medium
tags:
  - devops
created: 2026-01-02
---

# Set up deployment pipeline

Configure CI/CD with automated testing and staging environment.
The authentication service should be deployed first.
`,
		"003-docs.md": `---
id: "003"
title: "Write API documentation"
status: completed
priority: low
tags:
  - docs
created: 2026-01-03
---

# Write API documentation

Document all REST endpoints including request and response schemas.
`,
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}

	return tmpDir
}

func captureSearchOutput(t *testing.T, tasks []*model.Task, query, format string) (string, error) {
	t.Helper()

	results := search.Search(tasks, query)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var runErr error
	switch format {
	case "json":
		if len(results) > 0 {
			runErr = WriteJSON(os.Stdout, results)
		}
	case "yaml":
		if len(results) > 0 {
			runErr = WriteYAML(os.Stdout, results)
		}
	case "table":
		if len(results) > 0 {
			runErr = outputSearchTable(results, query)
		}
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), runErr
}

func TestSearch_MatchInTitle(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Implement authentication system", Status: model.StatusPending, Body: "Some body text"},
		{ID: "002", Title: "Set up deployment", Status: model.StatusInProgress, Body: "Other content"},
	}

	results := search.Search(tasks, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
	if results[0].MatchLocation != "title" {
		t.Errorf("expected match location 'title', got %s", results[0].MatchLocation)
	}
}

func TestSearch_MatchInBody(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Task one", Status: model.StatusPending, Body: "Contains the keyword deployment here"},
		{ID: "002", Title: "Task two", Status: model.StatusPending, Body: "Nothing relevant"},
	}

	results := search.Search(tasks, "deployment")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
	if results[0].MatchLocation != "body" {
		t.Errorf("expected match location 'body', got %s", results[0].MatchLocation)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "AUTHENTICATION Module", Status: model.StatusPending},
	}

	results := search.Search(tasks, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result for case-insensitive match, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
}

func TestSearch_MultipleResults(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Add authentication", Status: model.StatusPending, Body: "JWT auth"},
		{ID: "002", Title: "Deploy service", Status: model.StatusInProgress, Body: "Deploy the authentication service"},
		{ID: "003", Title: "Write docs", Status: model.StatusCompleted, Body: "No match here"},
	}

	results := search.Search(tasks, "authentication")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_MatchInTitleAndBody(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Authentication system", Status: model.StatusPending, Body: "Implement authentication with JWT"},
	}

	results := search.Search(tasks, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].MatchLocation != "title,body" {
		t.Errorf("expected match location 'title,body', got %s", results[0].MatchLocation)
	}
}

func TestSearch_NoResults(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Some task", Status: model.StatusPending, Body: "Nothing here"},
	}

	results := search.Search(tasks, "zzzznonexistent")

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_JSONFormat(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, FilePath: "tasks/001.md", Body: "JWT authentication"},
	}

	output, err := captureSearchOutput(t, tasks, "authentication", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []search.Result
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, output)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 result in JSON, got %d", len(parsed))
	}
	if parsed[0].ID != "001" {
		t.Errorf("expected id '001', got %s", parsed[0].ID)
	}
	if parsed[0].MatchLocation != "body" {
		t.Errorf("expected match_location 'body', got %s", parsed[0].MatchLocation)
	}
	if parsed[0].Snippet == "" {
		t.Error("expected non-empty snippet")
	}
}

func TestSearch_YAMLFormat(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, FilePath: "tasks/001.md", Body: "JWT authentication logic"},
	}

	output, err := captureSearchOutput(t, tasks, "authentication", "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "id:") {
		t.Error("expected YAML output to contain 'id:'")
	}
	if !strings.Contains(output, "match_location:") {
		t.Error("expected YAML output to contain 'match_location:'")
	}
	if !strings.Contains(output, "snippet:") {
		t.Error("expected YAML output to contain 'snippet:'")
	}
}

func TestSearch_UnsupportedFormat(t *testing.T) {
	resetSearchFlags()

	err := ValidateFormat("xml", []string{"table", "json", "yaml"})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %s", err.Error())
	}
}

func TestSearch_EmptyTaskList(t *testing.T) {
	resetSearchFlags()

	results := search.Search([]*model.Task{}, "anything")

	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty task list, got %d", len(results))
	}
}

func TestSearch_TableOutput(t *testing.T) {
	resetSearchFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, FilePath: "tasks/001.md", Body: "Implement authentication"},
	}

	output, err := captureSearchOutput(t, tasks, "authentication", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "ID") || !strings.Contains(output, "TITLE") {
		t.Error("expected table header with ID and TITLE")
	}
	if !strings.Contains(output, "001") {
		t.Error("expected task ID 001 in output")
	}
	if !strings.Contains(output, "Auth system") {
		t.Error("expected task title in output")
	}
}

func TestSearchTasks_Unit(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Alpha feature", Status: model.StatusPending, Body: "Build the alpha feature with tests"},
		{ID: "002", Title: "Beta release", Status: model.StatusInProgress, Body: "Prepare the beta release notes"},
		{ID: "003", Title: "Gamma fix", Status: model.StatusCompleted, Body: "Fix the alpha regression"},
	}

	tests := []struct {
		name     string
		query    string
		expected int
		ids      []string
	}{
		{"title match", "beta", 1, []string{"002"}},
		{"body match", "regression", 1, []string{"003"}},
		{"multiple matches", "alpha", 2, []string{"001", "003"}},
		{"no match", "omega", 0, nil},
		{"case insensitive", "BETA", 1, []string{"002"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := search.Search(tasks, tt.query)
			if len(results) != tt.expected {
				t.Fatalf("expected %d results, got %d", tt.expected, len(results))
			}
			for i, id := range tt.ids {
				if results[i].ID != id {
					t.Errorf("result[%d]: expected ID %s, got %s", i, id, results[i].ID)
				}
			}
		})
	}
}

func TestHighlightMatch(t *testing.T) {
	noColor = true
	r := getRenderer()

	result := highlightMatch("contains keyword here", "keyword", r)
	if !strings.Contains(result, "keyword") {
		t.Error("expected highlighted text to contain the match")
	}

	// When no match found, return original
	result = highlightMatch("no match", "xyz", r)
	if result != "no match" {
		t.Errorf("expected original text, got %q", result)
	}
}

func TestSearch_IntegrationWithFiles(t *testing.T) {
	resetSearchFlags()

	tmpDir := createSearchTestFiles(t)
	taskDir = tmpDir

	// Capture stderr for "no results" case
	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	searchFormat = "json"
	err := runSearch(searchCmd, []string{"authentication"})

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(stdoutR)

	var stderrBuf bytes.Buffer
	stderrBuf.ReadFrom(stderrR)

	output := stdoutBuf.String()
	if output == "" {
		t.Fatalf("expected JSON output, got empty. stderr: %s", stderrBuf.String())
	}

	var parsed []search.Result
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	// "authentication" appears in task 001 title+body and task 002 body
	if len(parsed) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(parsed))
	}

	foundIDs := make(map[string]bool)
	for _, r := range parsed {
		foundIDs[r.ID] = true
	}

	if !foundIDs["001"] {
		t.Error("expected task 001 to match (title contains 'authentication')")
	}
}
