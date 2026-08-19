package jira

import (
	"encoding/base64"
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
	maxResults = 100
	maxPages   = 50
)

// JiraSource fetches tasks from Jira Cloud issues.
type JiraSource struct {
	// HTTPClient allows injecting a custom client for testing. Nil uses http.DefaultClient.
	HTTPClient *http.Client
}

func init() {
	sync.Register(&JiraSource{})
}

func (j *JiraSource) Name() string { return "jira" }

func (j *JiraSource) ValidateConfig(cfg sync.SourceConfig) error {
	if cfg.Project == "" {
		return fmt.Errorf("project is required (Jira project key, e.g. PROJ)")
	}
	if cfg.TokenEnv == "" {
		return fmt.Errorf("token_env is required (Jira API token)")
	}
	if cfg.UserEnv == "" {
		return fmt.Errorf("user_env is required (Jira account email)")
	}
	return nil
}

func (j *JiraSource) FetchTasks(cfg sync.SourceConfig) ([]sync.ExternalTask, error) {
	token, err := resolveEnv(cfg.TokenEnv)
	if err != nil {
		return nil, err
	}
	user, err := resolveEnv(cfg.UserEnv)
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required for Jira source")
	}

	jql := buildJQL(cfg.Project, cfg.Filters)
	auth := basicAuth(user, token)

	client := j.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	var tasks []sync.ExternalTask

	nextPageToken := ""
	for page := 0; page < maxPages; page++ {
		issues, token, err := fetchPage(client, baseURL, jql, auth, nextPageToken)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			tasks = append(tasks, issueToExternalTask(issue, baseURL))
		}

		if token == "" || len(issues) == 0 {
			break
		}
		nextPageToken = token
	}

	return tasks, nil
}

// JSON response types

type searchResponse struct {
	Issues        []jiraIssue `json:"issues"`
	NextPageToken string      `json:"nextPageToken"`
}

type jiraIssue struct {
	Key    string     `json:"key"`
	Fields jiraFields `json:"fields"`
}

type jiraFields struct {
	Summary     string        `json:"summary"`
	Description any           `json:"description"` // ADF JSON
	Status      *jiraStatus   `json:"status"`
	Priority    *jiraPriority `json:"priority"`
	Assignee    *jiraUser     `json:"assignee"`
	Labels      []string      `json:"labels"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
}

type jiraStatus struct {
	Name string `json:"name"`
}

type jiraPriority struct {
	Name string `json:"name"`
}

type jiraUser struct {
	DisplayName string `json:"displayName"`
}

// helpers

func resolveEnv(envName string) (string, error) {
	val := os.Getenv(envName)
	if val == "" {
		return "", fmt.Errorf("environment variable %q is not set", envName)
	}
	return val, nil
}

func basicAuth(user, token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+token))
}

func buildJQL(project string, filters map[string]any) string {
	jql := fmt.Sprintf("project = %q", project)

	if extra, ok := filters["jql"]; ok {
		if s, ok := extra.(string); ok && s != "" {
			jql += " AND " + s
		}
	}

	return jql
}

// issueFields is the list of fields requested from the Jira API.
const issueFields = "summary,description,status,priority,assignee,labels,created,updated"

func fetchPage(client *http.Client, baseURL, jql, auth, pageToken string) ([]jiraIssue, string, error) {
	endpoint := baseURL + "/rest/api/3/search/jql"

	params := url.Values{}
	params.Set("jql", jql)
	params.Set("maxResults", strconv.Itoa(maxResults))
	params.Set("fields", issueFields)
	if pageToken != "" {
		params.Set("nextPageToken", pageToken)
	}

	reqURL := endpoint + "?" + params.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("jira: API returned %d: %s", resp.StatusCode, string(body))
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Issues, result.NextPageToken, nil
}

func issueToExternalTask(issue jiraIssue, baseURL string) sync.ExternalTask {
	task := sync.ExternalTask{
		ExternalID: issue.Key,
		Title:      issue.Fields.Summary,
		URL:        baseURL + "/browse/" + issue.Key,
		Labels:     issue.Fields.Labels,
	}

	if issue.Fields.Description != nil {
		task.Description = adfToMarkdown(issue.Fields.Description)
	}

	if issue.Fields.Status != nil {
		task.Status = issue.Fields.Status.Name
	}

	if issue.Fields.Priority != nil {
		task.Priority = issue.Fields.Priority.Name
	}

	if issue.Fields.Assignee != nil {
		task.Assignee = issue.Fields.Assignee.DisplayName
	}

	if t, err := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Created); err == nil {
		task.CreatedAt = t
	}

	if t, err := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Updated); err == nil {
		task.UpdatedAt = t
	}

	return task
}

// ADF (Atlassian Document Format) to Markdown conversion

func adfToMarkdown(doc any) string {
	m, ok := doc.(map[string]any)
	if !ok {
		return fmt.Sprint(doc)
	}

	content, _ := m["content"].([]any)
	var sb strings.Builder
	renderNodes(&sb, content, "")
	return strings.TrimRight(sb.String(), "\n")
}

func renderNodes(sb *strings.Builder, nodes []any, listPrefix string) {
	for _, node := range nodes {
		m, ok := node.(map[string]any)
		if !ok {
			continue
		}
		renderNode(sb, m, listPrefix)
	}
}

func renderNode(sb *strings.Builder, m map[string]any, listPrefix string) {
	nodeType, _ := m["type"].(string)
	content, _ := m["content"].([]any)

	switch nodeType {
	case "paragraph":
		renderInline(sb, content)
		sb.WriteString("\n\n")
	case "heading":
		renderHeading(sb, m, content)
	case "bulletList":
		renderList(sb, content, func(_ int) string { return "- " })
	case "orderedList":
		renderList(sb, content, func(idx int) string { return strconv.Itoa(idx+1) + ". " })
	case "listItem":
		sb.WriteString(listPrefix)
		renderInline(sb, content)
		sb.WriteString("\n")
	case "codeBlock":
		renderCodeBlock(sb, m, content)
	case "blockquote":
		renderBlockquote(sb, content)
	case "text":
		renderTextNode(sb, m)
	default:
		if len(content) > 0 {
			renderNodes(sb, content, listPrefix)
		}
	}
}

func renderHeading(sb *strings.Builder, m map[string]any, content []any) {
	level := toInt(m["attrs"], "level", 1)
	sb.WriteString(strings.Repeat("#", level))
	sb.WriteString(" ")
	renderInline(sb, content)
	sb.WriteString("\n\n")
}

func renderList(sb *strings.Builder, items []any, prefixFn func(int) string) {
	for idx, item := range items {
		renderNodes(sb, []any{item}, prefixFn(idx))
	}
}

func renderCodeBlock(sb *strings.Builder, m map[string]any, content []any) {
	lang := attrString(m["attrs"], "language")
	sb.WriteString("```")
	sb.WriteString(lang)
	sb.WriteString("\n")
	renderInline(sb, content)
	sb.WriteString("\n```\n\n")
}

func renderBlockquote(sb *strings.Builder, content []any) {
	var inner strings.Builder
	renderNodes(&inner, content, "")
	for line := range strings.SplitSeq(strings.TrimRight(inner.String(), "\n"), "\n") {
		sb.WriteString("> ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func renderInline(sb *strings.Builder, nodes []any) {
	for _, node := range nodes {
		m, ok := node.(map[string]any)
		if !ok {
			continue
		}

		nodeType, _ := m["type"].(string)
		switch nodeType {
		case "text":
			renderTextNode(sb, m)
		case "hardBreak":
			sb.WriteString("\n")
		default:
			content, _ := m["content"].([]any)
			if len(content) > 0 {
				renderInline(sb, content)
			}
		}
	}
}

func renderTextNode(sb *strings.Builder, m map[string]any) {
	text, _ := m["text"].(string)
	marks, _ := m["marks"].([]any)

	for _, mark := range marks {
		mm, ok := mark.(map[string]any)
		if !ok {
			continue
		}
		markType, _ := mm["type"].(string)
		switch markType {
		case "strong":
			sb.WriteString("**")
		case "em":
			sb.WriteString("*")
		case "code":
			sb.WriteString("`")
		case "link":
			sb.WriteString("[")
		}
	}

	sb.WriteString(text)

	// Close marks in reverse order
	for i := len(marks) - 1; i >= 0; i-- {
		mm, ok := marks[i].(map[string]any)
		if !ok {
			continue
		}
		markType, _ := mm["type"].(string)
		switch markType {
		case "strong":
			sb.WriteString("**")
		case "em":
			sb.WriteString("*")
		case "code":
			sb.WriteString("`")
		case "link":
			href := attrString(mm["attrs"], "href")
			sb.WriteString("](")
			sb.WriteString(href)
			sb.WriteString(")")
		}
	}
}

func toInt(attrs any, key string, fallback int) int {
	m, ok := attrs.(map[string]any)
	if !ok {
		return fallback
	}
	v, ok := m[key]
	if !ok {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return fallback
	}
}

func attrString(attrs any, key string) string {
	m, ok := attrs.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m[key].(string)
	return s
}
