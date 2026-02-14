package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/sync"
)

func TestGitHubSource_Name(t *testing.T) {
	src := &GitHubSource{}
	if src.Name() != "github" {
		t.Fatalf("expected name %q, got %q", "github", src.Name())
	}
}

func TestGitHubSource_ValidateConfig_Valid(t *testing.T) {
	src := &GitHubSource{}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		TokenEnv: "TEST_GH_TOKEN_VALID",
	}
	t.Setenv("TEST_GH_TOKEN_VALID", "tok123")

	if err := src.ValidateConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitHubSource_ValidateConfig_MissingProject(t *testing.T) {
	src := &GitHubSource{}
	cfg := sync.SourceConfig{
		Project:  "",
		TokenEnv: "TEST_GH_TOKEN",
	}
	err := src.ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestGitHubSource_ValidateConfig_InvalidFormat(t *testing.T) {
	src := &GitHubSource{}
	tests := []string{"noslash", "/repo", "owner/"}

	for _, proj := range tests {
		cfg := sync.SourceConfig{
			Project:  proj,
			TokenEnv: "TEST_GH_TOKEN",
		}
		if err := src.ValidateConfig(cfg); err == nil {
			t.Errorf("expected error for project %q", proj)
		}
	}
}

func TestGitHubSource_ValidateConfig_MissingTokenEnv(t *testing.T) {
	src := &GitHubSource{}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		TokenEnv: "",
	}
	err := src.ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing token_env")
	}
}

func TestGitHubSource_ValidateConfig_UnsetEnvVar(t *testing.T) {
	// ValidateConfig does NOT check if the env var is set — only FetchTasks does.
	// ValidateConfig only checks that the config fields are non-empty and well-formed.
	src := &GitHubSource{}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		TokenEnv: "UNSET_VAR_12345",
	}
	os.Unsetenv("UNSET_VAR_12345")

	// ValidateConfig should pass (it validates config shape, not runtime env).
	// The token resolution happens in FetchTasks.
	if err := src.ValidateConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitHubSource_FetchTasks_Basic(t *testing.T) {
	issues := []ghIssue{
		{
			Number:  1,
			Title:   "Fix bug",
			Body:    "Description here",
			State:   "open",
			HTMLURL: "https://github.com/owner/repo/issues/1",
			Labels:  []ghLabel{{Name: "bug"}, {Name: "urgent"}},
			Assignee: &ghAssignee{
				Login: "alice",
			},
			Milestone: &ghMilestone{
				Title: "v1.0",
			},
		},
		{
			Number:  2,
			Title:   "Add feature",
			Body:    "Feature desc",
			State:   "closed",
			HTMLURL: "https://github.com/owner/repo/issues/2",
			Labels:  []ghLabel{{Name: "enhancement"}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("unexpected accept header: %s", r.Header.Get("Accept"))
		}
		json.NewEncoder(w).Encode(issues)
	}))
	defer srv.Close()

	t.Setenv("TEST_FETCH_TOKEN", "test-token")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_FETCH_TOKEN",
	}

	tasks, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	task := tasks[0]
	if task.ExternalID != "1" {
		t.Errorf("expected ExternalID %q, got %q", "1", task.ExternalID)
	}
	if task.Title != "Fix bug" {
		t.Errorf("expected Title %q, got %q", "Fix bug", task.Title)
	}
	if task.Description != "Description here" {
		t.Errorf("expected Description %q, got %q", "Description here", task.Description)
	}
	if task.Status != "open" {
		t.Errorf("expected Status %q, got %q", "open", task.Status)
	}
	if task.Assignee != "alice" {
		t.Errorf("expected Assignee %q, got %q", "alice", task.Assignee)
	}
	if task.Priority != "v1.0" {
		t.Errorf("expected Priority %q, got %q", "v1.0", task.Priority)
	}
	if task.URL != "https://github.com/owner/repo/issues/1" {
		t.Errorf("expected URL %q, got %q", "https://github.com/owner/repo/issues/1", task.URL)
	}
	if len(task.Labels) != 2 || task.Labels[0] != "bug" || task.Labels[1] != "urgent" {
		t.Errorf("expected labels [bug urgent], got %v", task.Labels)
	}

	task2 := tasks[1]
	if task2.Assignee != "" {
		t.Errorf("expected empty Assignee, got %q", task2.Assignee)
	}
	if task2.Priority != "" {
		t.Errorf("expected empty Priority, got %q", task2.Priority)
	}
}

func TestGitHubSource_FetchTasks_SkipsPullRequests(t *testing.T) {
	issues := []ghIssue{
		{
			Number: 1,
			Title:  "Real issue",
			State:  "open",
		},
		{
			Number:      2,
			Title:       "A pull request",
			State:       "open",
			PullRequest: &struct{}{},
		},
		{
			Number: 3,
			Title:  "Another issue",
			State:  "open",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(issues)
	}))
	defer srv.Close()

	t.Setenv("TEST_PR_TOKEN", "tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_PR_TOKEN",
	}

	tasks, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (PRs filtered), got %d", len(tasks))
	}
	if tasks[0].ExternalID != "1" || tasks[1].ExternalID != "3" {
		t.Errorf("expected issues 1 and 3, got %s and %s", tasks[0].ExternalID, tasks[1].ExternalID)
	}
}

func TestGitHubSource_FetchTasks_Pagination(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		page := r.URL.Query().Get("page")

		switch page {
		case "1", "":
			// Return exactly perPage items to trigger next page
			issues := make([]ghIssue, perPage)
			for i := range issues {
				issues[i] = ghIssue{
					Number: i + 1,
					Title:  fmt.Sprintf("Issue %d", i+1),
					State:  "open",
				}
			}
			json.NewEncoder(w).Encode(issues)
		case "2":
			// Return fewer than perPage to signal last page
			issues := []ghIssue{
				{Number: 101, Title: "Issue 101", State: "open"},
			}
			json.NewEncoder(w).Encode(issues)
		default:
			json.NewEncoder(w).Encode([]ghIssue{})
		}
	}))
	defer srv.Close()

	t.Setenv("TEST_PAGE_TOKEN", "tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_PAGE_TOKEN",
	}

	tasks, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != perPage+1 {
		t.Fatalf("expected %d tasks, got %d", perPage+1, len(tasks))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestGitHubSource_FetchTasks_Filters(t *testing.T) {
	var receivedQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode([]ghIssue{})
	}))
	defer srv.Close()

	t.Setenv("TEST_FILTER_TOKEN", "tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_FILTER_TOKEN",
		Filters: map[string]any{
			"labels":   "bug,critical",
			"assignee": "alice",
		},
	}

	_, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedQuery == "" {
		t.Fatal("expected query params, got empty")
	}

	params, _ := url.ParseQuery(receivedQuery)
	if params.Get("labels") != "bug,critical" {
		t.Errorf("expected labels=bug,critical, got %q", params.Get("labels"))
	}
	if params.Get("assignee") != "alice" {
		t.Errorf("expected assignee=alice, got %q", params.Get("assignee"))
	}
	if params.Get("state") != "all" {
		t.Errorf("expected default state=all, got %q", params.Get("state"))
	}
}

func TestGitHubSource_FetchTasks_StateFilter(t *testing.T) {
	var receivedQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode([]ghIssue{})
	}))
	defer srv.Close()

	t.Setenv("TEST_STATE_TOKEN", "tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_STATE_TOKEN",
		Filters: map[string]any{
			"state": "open",
		},
	}

	_, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params, _ := url.ParseQuery(receivedQuery)
	if params.Get("state") != "open" {
		t.Errorf("expected state=open, got %q", params.Get("state"))
	}
}

func TestGitHubSource_FetchTasks_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"Bad credentials"}`)
	}))
	defer srv.Close()

	t.Setenv("TEST_ERR_TOKEN", "bad-tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_ERR_TOKEN",
	}

	_, err := src.FetchTasks(cfg)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to contain status code 403, got: %v", err)
	}
}

func TestGitHubSource_FetchTasks_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]ghIssue{})
	}))
	defer srv.Close()

	t.Setenv("TEST_EMPTY_TOKEN", "tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_EMPTY_TOKEN",
	}

	tasks, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestGitHubSource_FetchTasks_NilFields(t *testing.T) {
	issues := []ghIssue{
		{
			Number:    1,
			Title:     "No assignee or milestone",
			State:     "open",
			Assignee:  nil,
			Milestone: nil,
			Labels:    nil,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(issues)
	}))
	defer srv.Close()

	t.Setenv("TEST_NIL_TOKEN", "tok")

	src := &GitHubSource{HTTPClient: srv.Client()}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		BaseURL:  srv.URL,
		TokenEnv: "TEST_NIL_TOKEN",
	}

	tasks, err := src.FetchTasks(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Assignee != "" {
		t.Errorf("expected empty assignee, got %q", task.Assignee)
	}
	if task.Priority != "" {
		t.Errorf("expected empty priority, got %q", task.Priority)
	}
	if task.Labels != nil {
		t.Errorf("expected nil labels, got %v", task.Labels)
	}
}

func TestGitHubSource_FetchTasks_UnsetToken(t *testing.T) {
	os.Unsetenv("UNSET_TOKEN_VAR")

	src := &GitHubSource{}
	cfg := sync.SourceConfig{
		Project:  "owner/repo",
		TokenEnv: "UNSET_TOKEN_VAR",
	}

	_, err := src.FetchTasks(cfg)
	if err == nil {
		t.Fatal("expected error for unset token env var")
	}
	if !strings.Contains(err.Error(), "UNSET_TOKEN_VAR") {
		t.Errorf("expected error to mention env var name, got: %v", err)
	}
}
