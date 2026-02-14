package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/sync"
)

const (
	defaultBaseURL = "https://api.github.com"
	perPage        = 100
	maxPages       = 50
)

// GitHubSource fetches tasks from GitHub Issues.
type GitHubSource struct {
	// HTTPClient allows injecting a custom client for testing. Nil uses http.DefaultClient.
	HTTPClient *http.Client
}

func init() {
	sync.Register(&GitHubSource{})
}

func (g *GitHubSource) Name() string { return "github" }

func (g *GitHubSource) ValidateConfig(cfg sync.SourceConfig) error {
	if cfg.Project == "" {
		return fmt.Errorf("project is required")
	}
	if _, _, err := splitProject(cfg.Project); err != nil {
		return err
	}
	if cfg.TokenEnv == "" {
		return fmt.Errorf("token_env is required")
	}
	return nil
}

func (g *GitHubSource) FetchTasks(cfg sync.SourceConfig) ([]sync.ExternalTask, error) {
	owner, repo, err := splitProject(cfg.Project)
	if err != nil {
		return nil, err
	}

	token, err := resolveToken(cfg.TokenEnv)
	if err != nil {
		return nil, err
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/issues", baseURL, owner, repo)
	params := buildQueryParams(cfg.Filters)

	client := g.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	var tasks []sync.ExternalTask

	for page := 1; page <= maxPages; page++ {
		issues, hasMore, err := fetchPage(client, apiURL, params, token, page)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			if issue.PullRequest != nil {
				continue
			}
			tasks = append(tasks, issueToExternalTask(issue))
		}

		if !hasMore {
			break
		}
	}

	return tasks, nil
}

// JSON response types

type ghIssue struct {
	Number      int          `json:"number"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	State       string       `json:"state"`
	HTMLURL     string       `json:"html_url"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Labels      []ghLabel    `json:"labels"`
	Assignee    *ghAssignee  `json:"assignee"`
	Milestone   *ghMilestone `json:"milestone"`
	PullRequest *struct{}    `json:"pull_request"`
}

type ghLabel struct {
	Name string `json:"name"`
}

type ghAssignee struct {
	Login string `json:"login"`
}

type ghMilestone struct {
	Title string `json:"title"`
}

// helpers

func splitProject(project string) (owner, repo string, err error) {
	parts := strings.SplitN(project, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("project must be in owner/repo format, got %q", project)
	}
	return parts[0], parts[1], nil
}

func resolveToken(envName string) (string, error) {
	token := os.Getenv(envName)
	if token == "" {
		return "", fmt.Errorf("environment variable %q is not set", envName)
	}
	return token, nil
}

func buildQueryParams(filters map[string]any) url.Values {
	params := url.Values{}
	params.Set("per_page", strconv.Itoa(perPage))

	hasState := false
	for key, val := range filters {
		switch key {
		case "labels":
			params.Set("labels", toCommaSeparated(val))
		case "milestone":
			params.Set("milestone", fmt.Sprint(val))
		case "assignee":
			params.Set("assignee", fmt.Sprint(val))
		case "state":
			params.Set("state", fmt.Sprint(val))
			hasState = true
		default:
			params.Set(key, fmt.Sprint(val))
		}
	}

	if !hasState {
		params.Set("state", "all")
	}

	return params
}

func toCommaSeparated(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprint(item)
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(v, ",")
	default:
		return fmt.Sprint(v)
	}
}

func fetchPage(client *http.Client, apiURL string, params url.Values, token string, page int) ([]ghIssue, bool, error) {
	params.Set("page", strconv.Itoa(page))

	reqURL := apiURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var issues []ghIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	hasMore := len(issues) == perPage
	return issues, hasMore, nil
}

func issueToExternalTask(issue ghIssue) sync.ExternalTask {
	task := sync.ExternalTask{
		ExternalID:  strconv.Itoa(issue.Number),
		Title:       issue.Title,
		Description: issue.Body,
		Status:      issue.State,
		URL:         issue.HTMLURL,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
	}

	for _, l := range issue.Labels {
		task.Labels = append(task.Labels, l.Name)
	}

	if issue.Assignee != nil {
		task.Assignee = issue.Assignee.Login
	}

	if issue.Milestone != nil {
		task.Priority = issue.Milestone.Title
	}

	return task
}
