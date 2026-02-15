package search

import (
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestSearch_TitleMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Implement authentication", Status: model.StatusPending, Body: "Some body"},
		{ID: "002", Title: "Deploy service", Status: model.StatusInProgress, Body: "Other content"},
	}

	results := Search(tasks, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
	if results[0].MatchLocation != "title" {
		t.Errorf("expected 'title', got %s", results[0].MatchLocation)
	}
}

func TestSearch_BodyMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task one", Status: model.StatusPending, Body: "Contains deployment keyword"},
		{ID: "002", Title: "Task two", Status: model.StatusPending, Body: "Nothing relevant"},
	}

	results := Search(tasks, "deployment")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].MatchLocation != "body" {
		t.Errorf("expected 'body', got %s", results[0].MatchLocation)
	}
}

func TestSearch_TitleAndBodyMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Authentication system", Status: model.StatusPending, Body: "Implement authentication with JWT"},
	}

	results := Search(tasks, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].MatchLocation != "title,body" {
		t.Errorf("expected 'title,body', got %s", results[0].MatchLocation)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "AUTHENTICATION Module", Status: model.StatusPending},
	}

	results := Search(tasks, "authentication")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearch_NoMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Some task", Status: model.StatusPending, Body: "Nothing here"},
	}

	results := Search(tasks, "nonexistent")

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_EmptyTasks(t *testing.T) {
	results := Search([]*model.Task{}, "anything")

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_MultipleResults(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Alpha feature", Status: model.StatusPending, Body: "Build alpha"},
		{ID: "002", Title: "Beta release", Status: model.StatusInProgress, Body: "Prepare beta"},
		{ID: "003", Title: "Gamma fix", Status: model.StatusCompleted, Body: "Fix the alpha regression"},
	}

	results := Search(tasks, "alpha")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected first result 001, got %s", results[0].ID)
	}
	if results[1].ID != "003" {
		t.Errorf("expected second result 003, got %s", results[1].ID)
	}
}

func TestSearch_StatusField(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Auth task", Status: model.StatusInProgress, Body: "auth work"},
	}

	results := Search(tasks, "auth")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "in-progress" {
		t.Errorf("expected status 'in-progress', got %s", results[0].Status)
	}
}

func TestExtractBodySnippet_Short(t *testing.T) {
	snippet := ExtractBodySnippet("small body with keyword here", "keyword")

	if !strings.Contains(snippet, "keyword") {
		t.Errorf("expected snippet to contain 'keyword', got %q", snippet)
	}
	if strings.HasPrefix(snippet, "...") {
		t.Errorf("short body should not have prefix ellipsis, got %q", snippet)
	}
	if strings.HasSuffix(snippet, "...") {
		t.Errorf("short body should not have suffix ellipsis, got %q", snippet)
	}
}

func TestExtractBodySnippet_LongWithEllipsis(t *testing.T) {
	body := strings.Repeat("word ", 20) + "target" + strings.Repeat(" word", 20)
	snippet := ExtractBodySnippet(body, "target")

	if !strings.Contains(snippet, "target") {
		t.Errorf("expected snippet to contain 'target', got %q", snippet)
	}
	if !strings.HasPrefix(snippet, "...") {
		t.Errorf("expected prefix ellipsis, got %q", snippet)
	}
	if !strings.HasSuffix(snippet, "...") {
		t.Errorf("expected suffix ellipsis, got %q", snippet)
	}
}

func TestExtractBodySnippet_MatchAtStart(t *testing.T) {
	body := "target is at the start" + strings.Repeat(" word", 20)
	snippet := ExtractBodySnippet(body, "target")

	if !strings.Contains(snippet, "target") {
		t.Errorf("expected snippet to contain 'target', got %q", snippet)
	}
	if strings.HasPrefix(snippet, "...") {
		t.Errorf("match at start should not have prefix ellipsis, got %q", snippet)
	}
	if !strings.HasSuffix(snippet, "...") {
		t.Errorf("expected suffix ellipsis, got %q", snippet)
	}
}

func TestExtractBodySnippet_NoMatch(t *testing.T) {
	snippet := ExtractBodySnippet("some body text", "nonexistent")

	if snippet != "" {
		t.Errorf("expected empty snippet for no match, got %q", snippet)
	}
}
