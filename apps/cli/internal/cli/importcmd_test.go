package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/sync"
)

// chdirTo switches the working directory to dir for the duration of the test.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })
}

// printImportTableStdout renders printImportTable and returns captured stdout.
func printImportTableStdout(t *testing.T, result *sync.ImportResult, summary importSummary, quietMode bool) string {
	t.Helper()
	var err error
	stdout, _ := captureOutput(t, func() { err = printImportTable(result, summary, quietMode) })
	if err != nil {
		t.Fatalf("printImportTable failed: %v", err)
	}
	return stdout
}

func TestImportCommand_NonInteractive(t *testing.T) {
	sourceName := "test-import-cli"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "CLI-1", Title: "Import task one", Status: "open", URL: "https://example.com/1"},
			{ExternalID: "CLI-2", Title: "Import task two", Status: "open"},
		},
	})

	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	res := repo.Run("import", "--source", sourceName, "--output-dir", repo.Path("tasks"))
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// Verify files were created
	entries, err := os.ReadDir(repo.Path("tasks"))
	if err != nil {
		t.Fatalf("failed to read tasks dir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 task files, got %d", len(entries))
	}
}

func TestImportCommand_DryRun(t *testing.T) {
	sourceName := "test-import-cli-dry"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "DRY-1", Title: "Dry run task", Status: "open"},
		},
	})

	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	res := repo.Run("import", "--source", sourceName, "--output-dir", repo.Path("tasks"), "--dry-run")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// No files should be created
	_, statErr := os.Stat(repo.Path("tasks"))
	if statErr == nil {
		t.Error("expected no tasks directory in dry-run mode")
	}
}

func TestImportCommand_JSONOutput(t *testing.T) {
	sourceName := "test-import-cli-json"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "JSON-1", Title: "JSON output task", Status: "open"},
		},
	})

	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	res := repo.Run("import", "--source", sourceName, "--output-dir", repo.Path("tasks"), "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	// Verify valid JSON
	var data importResultData
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput:\n%s", err, output)
	}

	if data.Summary.Created != 1 {
		t.Errorf("expected 1 created in summary, got %d", data.Summary.Created)
	}
	if len(data.Created) != 1 {
		t.Errorf("expected 1 created item, got %d", len(data.Created))
	}
	if data.Created[0].ExternalID != "JSON-1" {
		t.Errorf("expected external_id JSON-1, got %s", data.Created[0].ExternalID)
	}
}

func TestImportCommand_DuplicateSkip(t *testing.T) {
	sourceName := "test-import-cli-dedup"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "DUP-1", Title: "Already exists", Status: "open"},
			{ExternalID: "NEW-1", Title: "Brand new", Status: "open"},
		},
	})

	// Create an existing task with external_id DUP-1
	repo := newTaskRepo(t, map[string]string{
		"tasks/001-existing.md": "---\nid: \"001\"\ntitle: \"Existing\"\nstatus: pending\nexternal_id: \"DUP-1\"\n---\n",
	})
	chdirTo(t, repo.Dir)

	res := repo.Run("import", "--source", sourceName, "--output-dir", repo.Path("tasks"), "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	var data importResultData
	if err := json.Unmarshal([]byte(res.Stdout), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data.Summary.Created != 1 {
		t.Errorf("expected 1 created, got %d", data.Summary.Created)
	}
	if data.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", data.Summary.Skipped)
	}
	if len(data.Skipped) != 1 || data.Skipped[0].ExternalID != "DUP-1" {
		t.Errorf("expected DUP-1 to be skipped, got: %+v", data.Skipped)
	}
}

func TestImportCommand_FilterParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:     "single filter",
			input:    "state:open",
			expected: map[string]any{"state": "open"},
		},
		{
			name:     "multiple filters",
			input:    "state:open labels:bug",
			expected: map[string]any{"state": "open", "labels": "bug"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]any{},
		},
		{
			name:     "no colon",
			input:    "nocolon",
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseImportFilters(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("expected %d filters, got %d: %v", len(tt.expected), len(got), got)
				return
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("filter %q: expected %v, got %v", k, v, got[k])
				}
			}
		})
	}
}

func TestImportCommand_OutputDirFlag(t *testing.T) {
	sourceName := "test-import-cli-outdir"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "DIR-1", Title: "Custom dir task", Status: "open"},
		},
	})

	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	customDir := repo.Path("custom/output")

	res := repo.Run("import", "--source", sourceName, "--output-dir", customDir)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	entries, err := os.ReadDir(customDir)
	if err != nil {
		t.Fatalf("failed to read custom dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 task file in custom dir, got %d", len(entries))
	}
}

func TestImportCommand_FilterFlagPopulatesConfig(t *testing.T) {
	sourceName := "test-import-cli-filter-cfg"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "F-1", Title: "Filtered", Status: "open"},
		},
	})

	resetCLIState()
	importSource = sourceName
	importProject = "owner/repo"
	importTokenEnv = "GITHUB_TOKEN"
	importFilter = "state:open labels:bug"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters == nil {
		t.Fatal("expected filters to be set")
	}
	if cfg.SourceCfg.Filters["state"] != "open" {
		t.Errorf("expected state filter 'open', got %v", cfg.SourceCfg.Filters["state"])
	}
	if cfg.SourceCfg.Filters["labels"] != "bug" {
		t.Errorf("expected labels filter 'bug', got %v", cfg.SourceCfg.Filters["labels"])
	}
}

func TestImportCommand_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("import", "--source", "something", "--format", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("unexpected error message: %v", res.Err)
	}
}

func TestImportCommand_RepoFlagAliasesProject(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "myorg/myrepo"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Project != "myorg/myrepo" {
		t.Errorf("expected project=myorg/myrepo, got %q", cfg.SourceCfg.Project)
	}
}

func TestImportCommand_ProjectTakesPrecedenceOverRepo(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importProject = "explicit/project"
	importRepo = "fallback/repo"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Project != "explicit/project" {
		t.Errorf("expected project=explicit/project, got %q", cfg.SourceCfg.Project)
	}
}

func TestImportCommand_RepoFlagIgnoredForNonGitHub(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importRepo = "should-be-ignored"
	importProject = "PROJ"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Project != "PROJ" {
		t.Errorf("expected project=PROJ, got %q", cfg.SourceCfg.Project)
	}
}

func TestImportCommand_GitHubDefaultTokenEnv(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.TokenEnv != "GITHUB_TOKEN" {
		t.Errorf("expected default token env=GITHUB_TOKEN, got %q", cfg.SourceCfg.TokenEnv)
	}
}

func TestImportCommand_GitHubDefaultStateOpen(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters == nil {
		t.Fatal("expected filters to be set")
	}
	if cfg.SourceCfg.Filters["state"] != "open" {
		t.Errorf("expected default state=open, got %v", cfg.SourceCfg.Filters["state"])
	}
}

func TestImportCommand_GitHubStateNotOverridden(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importFilter = "state:closed"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters["state"] != "closed" {
		t.Errorf("expected state=closed (from --filter), got %v", cfg.SourceCfg.Filters["state"])
	}
}

func TestImportCommand_LabelsShortcutFlag(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importLabels = "bug,critical"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters["labels"] != "bug,critical" {
		t.Errorf("expected labels=bug,critical, got %v", cfg.SourceCfg.Filters["labels"])
	}
}

func TestImportCommand_MilestoneShortcutFlag(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importMilestone = "v1.0"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters["milestone"] != "v1.0" {
		t.Errorf("expected milestone=v1.0, got %v", cfg.SourceCfg.Filters["milestone"])
	}
}

func TestImportCommand_AssigneeShortcutFlag(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importAssignee = "alice"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters["assignee"] != "alice" {
		t.Errorf("expected assignee=alice, got %v", cfg.SourceCfg.Filters["assignee"])
	}
}

func TestImportCommand_FilterFlagTakesPrecedenceOverShortcut(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importFilter = "labels:from-filter"
	importLabels = "from-shortcut"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// --filter labels take precedence over --labels shortcut
	if cfg.SourceCfg.Filters["labels"] != "from-filter" {
		t.Errorf("expected labels=from-filter (from --filter), got %v", cfg.SourceCfg.Filters["labels"])
	}
}

func TestImportCommand_AllShortcutFlags(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importLabels = "bug"
	importMilestone = "v2.0"
	importAssignee = "bob"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters["labels"] != "bug" {
		t.Errorf("expected labels=bug, got %v", cfg.SourceCfg.Filters["labels"])
	}
	if cfg.SourceCfg.Filters["milestone"] != "v2.0" {
		t.Errorf("expected milestone=v2.0, got %v", cfg.SourceCfg.Filters["milestone"])
	}
	if cfg.SourceCfg.Filters["assignee"] != "bob" {
		t.Errorf("expected assignee=bob, got %v", cfg.SourceCfg.Filters["assignee"])
	}
	if cfg.SourceCfg.Filters["state"] != "open" {
		t.Errorf("expected default state=open, got %v", cfg.SourceCfg.Filters["state"])
	}
}

func TestImportCommand_NonGitHubNoDefaultState(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"
	importTokenEnv = "JIRA_TOKEN"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Non-github sources should not get default state filter
	if cfg.SourceCfg.Filters != nil {
		t.Errorf("expected nil filters for non-github source, got %v", cfg.SourceCfg.Filters)
	}
}

func TestImportCommand_JiraURLFlagAliasesBaseURL(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"
	importURL = "https://company.atlassian.net"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.BaseURL != "https://company.atlassian.net" {
		t.Errorf("expected base-url from --url, got %q", cfg.SourceCfg.BaseURL)
	}
}

func TestImportCommand_BaseURLTakesPrecedenceOverURL(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"
	importBaseURL = "https://explicit.atlassian.net"
	importURL = "https://fallback.atlassian.net"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.BaseURL != "https://explicit.atlassian.net" {
		t.Errorf("expected --base-url to take precedence, got %q", cfg.SourceCfg.BaseURL)
	}
}

func TestImportCommand_URLFlagIgnoredForNonJira(t *testing.T) {
	resetCLIState()
	importSource = "github"
	importRepo = "owner/repo"
	importURL = "https://should-be-ignored.com"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.BaseURL != "" {
		t.Errorf("expected empty base-url for non-jira source, got %q", cfg.SourceCfg.BaseURL)
	}
}

func TestImportCommand_JQLFlagPopulatesFilters(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"
	importJQL = "assignee = currentUser()"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.Filters == nil {
		t.Fatal("expected filters to be set")
	}
	if cfg.SourceCfg.Filters["jql"] != "assignee = currentUser()" {
		t.Errorf("expected jql filter, got %v", cfg.SourceCfg.Filters["jql"])
	}
}

func TestImportCommand_FilterFlagTakesPrecedenceOverJQL(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"
	importFilter = "jql:from-filter"
	importJQL = "from-shortcut"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// --filter jql takes precedence over --jql shortcut
	if cfg.SourceCfg.Filters["jql"] != "from-filter" {
		t.Errorf("expected jql from --filter, got %v", cfg.SourceCfg.Filters["jql"])
	}
}

func TestImportCommand_JiraDefaultTokenEnv(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.TokenEnv != "JIRA_TOKEN" {
		t.Errorf("expected default token env=JIRA_TOKEN, got %q", cfg.SourceCfg.TokenEnv)
	}
}

func TestImportCommand_JiraDefaultUserEnv(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.UserEnv != "JIRA_USER" {
		t.Errorf("expected default user env=JIRA_USER, got %q", cfg.SourceCfg.UserEnv)
	}
}

func TestImportCommand_JiraExplicitEnvOverridesDefaults(t *testing.T) {
	resetCLIState()
	importSource = "jira"
	importProject = "PROJ"
	importTokenEnv = "MY_JIRA_TOKEN"
	importUserEnv = "MY_JIRA_USER"

	cfg, err := buildImportConfigFromFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SourceCfg.TokenEnv != "MY_JIRA_TOKEN" {
		t.Errorf("expected explicit token env, got %q", cfg.SourceCfg.TokenEnv)
	}
	if cfg.SourceCfg.UserEnv != "MY_JIRA_USER" {
		t.Errorf("expected explicit user env, got %q", cfg.SourceCfg.UserEnv)
	}
}

func TestProjectHint(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"github", "GitHub repository (e.g. owner/repo)"},
		{"jira", "Jira project key (e.g. PROJ)"},
		{"unknown", "Project identifier"},
		{"", "Project identifier"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			got := projectHint(tt.source)
			if got != tt.expected {
				t.Errorf("projectHint(%q) = %q, want %q", tt.source, got, tt.expected)
			}
		})
	}
}

func TestPrintImportTable_QuietMode(t *testing.T) {
	result := &sync.ImportResult{
		Created: []sync.ImportAction{{ExternalID: "1", LocalID: "001", Title: "Task"}},
	}
	summary := importSummary{Total: 1, Created: 1}

	output := printImportTableStdout(t, result, summary, true)

	if output != "" {
		t.Errorf("expected no output in quiet mode, got %q", output)
	}
}

func TestPrintImportTable_WithCreated(t *testing.T) {
	result := &sync.ImportResult{
		Created: []sync.ImportAction{
			{ExternalID: "GH-1", LocalID: "001", Title: "First task"},
			{ExternalID: "GH-2", LocalID: "002", Title: "Second task"},
		},
	}
	summary := importSummary{Total: 2, Created: 2}

	output := printImportTableStdout(t, result, summary, false)

	if !strings.Contains(output, "Created 2 task(s)") {
		t.Errorf("expected 'Created 2 task(s)', got %q", output)
	}
	if !strings.Contains(output, "[001] First task") {
		t.Error("expected first task ID and title in output")
	}
	if !strings.Contains(output, "[002] Second task") {
		t.Error("expected second task ID and title in output")
	}
}

func TestPrintImportTable_WithSkipped(t *testing.T) {
	result := &sync.ImportResult{
		Skipped: []sync.ImportAction{
			{ExternalID: "GH-5", Title: "Duplicate task"},
		},
	}
	summary := importSummary{Total: 1, Skipped: 1}

	output := printImportTableStdout(t, result, summary, false)

	if !strings.Contains(output, "Skipped 1 task(s)") {
		t.Errorf("expected 'Skipped 1 task(s)', got %q", output)
	}
	if !strings.Contains(output, "[GH-5] Duplicate task") {
		t.Error("expected skipped task external ID and title in output")
	}
}

func TestPrintImportTable_WithErrors(t *testing.T) {
	result := &sync.ImportResult{
		Errors: []sync.SyncError{
			{ExternalID: "GH-9", Title: "Bad task", Err: fmt.Errorf("write failed")},
		},
	}
	summary := importSummary{Total: 1, Errors: 1}

	output := printImportTableStdout(t, result, summary, false)

	if !strings.Contains(output, "Errors 1 task(s)") {
		t.Errorf("expected 'Errors 1 task(s)', got %q", output)
	}
	if !strings.Contains(output, "[GH-9] Bad task") {
		t.Error("expected error task external ID and title in output")
	}
	if !strings.Contains(output, "write failed") {
		t.Error("expected error message in output")
	}
}

func TestPrintImportTable_Summary(t *testing.T) {
	result := &sync.ImportResult{
		Created: []sync.ImportAction{
			{ExternalID: "1", LocalID: "001", Title: "Created"},
		},
		Skipped: []sync.ImportAction{
			{ExternalID: "2", Title: "Skipped"},
		},
		Errors: []sync.SyncError{
			{ExternalID: "3", Title: "Error", Err: fmt.Errorf("fail")},
		},
	}
	summary := importSummary{Total: 3, Created: 1, Skipped: 1, Errors: 1}

	output := printImportTableStdout(t, result, summary, false)

	if !strings.Contains(output, "Done: 3 total, 1 created, 1 skipped, 1 errors") {
		t.Errorf("expected summary line, got %q", output)
	}
}
