package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/search"
)

// searchTestFiles is the canonical search-integration fixture: three tasks whose
// titles/bodies mention "authentication" in different places. Kept inline (not a
// shared fixture) because the body prose is what the search matcher is tested
// against.
func searchTestFiles() map[string]string {
	return map[string]string{
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
}

// searchOutput searches tasks and renders results in the given format directly
// through the formatter, returning stdout. Color is disabled for determinism.
func searchOutput(t *testing.T, tasks []*model.Task, query, format string) (string, error) {
	t.Helper()
	noColor = true
	results := search.Search(tasks, query)

	var runErr error
	stdout, _ := captureOutput(t, func() {
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
	})
	return stdout, runErr
}

func TestSearch_MatchInTitle(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Implement authentication system", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "Some body text"},
		{ID: "002", Title: "Set up deployment", Status: model.StatusInProgress, Priority: model.PriorityMedium, Body: "Other content"},
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
	tasks := []*model.Task{
		{ID: "001", Title: "Task one", Status: model.StatusPending, Priority: model.PriorityMedium, Body: "Contains the keyword deployment here"},
		{ID: "002", Title: "Task two", Status: model.StatusPending, Priority: model.PriorityLow, Body: "Nothing relevant"},
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
	tasks := []*model.Task{
		{ID: "001", Title: "AUTHENTICATION Module", Status: model.StatusPending, Priority: model.PriorityHigh},
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
	tasks := []*model.Task{
		{ID: "001", Title: "Add authentication", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "JWT auth"},
		{ID: "002", Title: "Deploy service", Status: model.StatusInProgress, Priority: model.PriorityMedium, Body: "Deploy the authentication service"},
		{ID: "003", Title: "Write docs", Status: model.StatusCompleted, Priority: model.PriorityLow, Body: "No match here"},
	}

	results := search.Search(tasks, "authentication")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_MatchInTitleAndBody(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Authentication system", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "Implement authentication with JWT"},
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
	tasks := []*model.Task{
		{ID: "001", Title: "Some task", Status: model.StatusPending, Priority: model.PriorityMedium, Body: "Nothing here"},
	}

	results := search.Search(tasks, "zzzznonexistent")

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_JSONFormat(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "tasks/001.md", Body: "JWT authentication"},
	}

	output, err := searchOutput(t, tasks, "authentication", "json")
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
	if parsed[0].Priority != "high" {
		t.Errorf("expected priority 'high', got %s", parsed[0].Priority)
	}
}

func TestSearch_YAMLFormat(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, Priority: model.PriorityMedium, FilePath: "tasks/001.md", Body: "JWT authentication logic"},
	}

	output, err := searchOutput(t, tasks, "authentication", "yaml")
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
	if !strings.Contains(output, "priority:") {
		t.Error("expected YAML output to contain 'priority:'")
	}
}

func TestSearch_UnsupportedFormat(t *testing.T) {
	err := ValidateFormat("xml", []string{"table", "json", "yaml"})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %s", err.Error())
	}
}

func TestSearch_EmptyTaskList(t *testing.T) {
	results := search.Search([]*model.Task{}, "anything")

	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty task list, got %d", len(results))
	}
}

func TestSearch_TableOutput(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "tasks/001.md", Body: "Implement authentication"},
	}

	output, err := searchOutput(t, tasks, "authentication", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "ID") || !strings.Contains(output, "TITLE") {
		t.Error("expected table header with ID and TITLE")
	}
	if !strings.Contains(output, "PRIORITY") {
		t.Error("expected table header with PRIORITY")
	}
	if !strings.Contains(output, "001") {
		t.Error("expected task ID 001 in output")
	}
	if !strings.Contains(output, "Auth system") {
		t.Error("expected task title in output")
	}
	if !strings.Contains(output, "high") {
		t.Error("expected priority 'high' in output")
	}
}

func TestSearchTasks_Unit(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Alpha feature", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "Build the alpha feature with tests"},
		{ID: "002", Title: "Beta release", Status: model.StatusInProgress, Priority: model.PriorityMedium, Body: "Prepare the beta release notes"},
		{ID: "003", Title: "Gamma fix", Status: model.StatusCompleted, Priority: model.PriorityLow, Body: "Fix the alpha regression"},
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

func TestSearch_FilterByPriority(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "JWT authentication"},
		{ID: "002", Title: "Auth docs", Status: model.StatusPending, Priority: model.PriorityLow, Body: "Document authentication endpoints"},
	}

	// Both match "authentication", but only 001 is high priority
	filtered, err := applyFilters(tasks, []string{"priority=high"})
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}
	results := search.Search(filtered, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
}

func TestSearch_FilterByStatus(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth login", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "Login endpoint"},
		{ID: "002", Title: "Auth logout", Status: model.StatusCompleted, Priority: model.PriorityMedium, Body: "Logout endpoint"},
	}

	filtered, err := applyFilters(tasks, []string{"status=pending"})
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}
	results := search.Search(filtered, "auth")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
}

func TestSearch_MultipleFilters(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth high pending", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "security feature"},
		{ID: "002", Title: "Auth high done", Status: model.StatusCompleted, Priority: model.PriorityHigh, Body: "security fix"},
		{ID: "003", Title: "Auth low pending", Status: model.StatusPending, Priority: model.PriorityLow, Body: "security docs"},
	}

	filtered, err := applyFilters(tasks, []string{"status=pending", "priority=high"})
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}
	results := search.Search(filtered, "security")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
}

func TestSearch_SortByPriority(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Deploy low", Status: model.StatusPending, Priority: model.PriorityLow, Body: "deploy staging"},
		{ID: "002", Title: "Deploy critical", Status: model.StatusPending, Priority: model.PriorityCritical, Body: "deploy production"},
		{ID: "003", Title: "Deploy high", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "deploy canary"},
	}

	if err := sortTasks(tasks, "priority"); err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}
	results := search.Search(tasks, "deploy")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// Critical first, then high, then low
	if results[0].ID != "002" {
		t.Errorf("expected first result to be 002 (critical), got %s", results[0].ID)
	}
	if results[1].ID != "003" {
		t.Errorf("expected second result to be 003 (high), got %s", results[1].ID)
	}
	if results[2].ID != "001" {
		t.Errorf("expected third result to be 001 (low), got %s", results[2].ID)
	}
}

func TestSearch_LimitResults(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "API endpoint one", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "first endpoint"},
		{ID: "002", Title: "API endpoint two", Status: model.StatusPending, Priority: model.PriorityMedium, Body: "second endpoint"},
		{ID: "003", Title: "API endpoint three", Status: model.StatusPending, Priority: model.PriorityLow, Body: "third endpoint"},
	}

	results := search.Search(tasks, "endpoint")

	if len(results) != 3 {
		t.Fatalf("expected 3 results before limit, got %d", len(results))
	}

	// Apply limit
	limit := 2
	if limit < len(results) {
		results = results[:limit]
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results after limit, got %d", len(results))
	}
}

func TestSearch_AllFlagsCombined(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Bug fix auth", Status: model.StatusPending, Priority: model.PriorityLow, Body: "fix login bug"},
		{ID: "002", Title: "Bug fix deploy", Status: model.StatusPending, Priority: model.PriorityHigh, Body: "fix deploy bug"},
		{ID: "003", Title: "Bug fix cache", Status: model.StatusPending, Priority: model.PriorityCritical, Body: "fix cache bug"},
		{ID: "004", Title: "Bug fix API", Status: model.StatusCompleted, Priority: model.PriorityCritical, Body: "fix api bug"},
	}

	// Filter: only pending tasks
	filtered, err := applyFilters(tasks, []string{"status=pending"})
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}

	// Sort by priority (critical first)
	if err := sortTasks(filtered, "priority"); err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}

	// Search for "bug"
	results := search.Search(filtered, "bug")

	// Limit to 2
	if len(results) > 2 {
		results = results[:2]
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Critical pending first, then high pending
	if results[0].ID != "003" {
		t.Errorf("expected first result 003 (critical), got %s", results[0].ID)
	}
	if results[1].ID != "002" {
		t.Errorf("expected second result 002 (high), got %s", results[1].ID)
	}
}

func TestSearch_FilterNoMatchingTasks(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth system", Status: model.StatusCompleted, Priority: model.PriorityHigh, Body: "JWT authentication"},
	}

	// Filter for pending, but task is completed
	filtered, err := applyFilters(tasks, []string{"status=pending"})
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}
	results := search.Search(filtered, "authentication")

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_InvalidFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Test", Status: model.StatusPending, Priority: model.PriorityMedium},
	}

	_, err := applyFilters(tasks, []string{"badfilter"})
	if err == nil {
		t.Fatal("expected error for invalid filter syntax")
	}
}

func TestSearch_InvalidSort(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Test", Status: model.StatusPending, Priority: model.PriorityMedium},
	}

	err := sortTasks(tasks, "nonexistent")
	if err == nil {
		t.Fatal("expected error for invalid sort field")
	}
}

func TestSearch_IntegrationWithFiles(t *testing.T) {
	repo := newTaskRepo(t, searchTestFiles())

	res := repo.Run("search", "authentication", "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	output := res.Stdout
	if output == "" {
		t.Fatalf("expected JSON output, got empty. stderr: %s", res.Stderr)
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
