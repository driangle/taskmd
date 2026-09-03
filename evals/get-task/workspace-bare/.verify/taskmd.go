package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Task mirrors the subset of `taskmd list --format json` we assert on.
type Task struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority"`
	Effort       string   `json:"effort"`
	Type         string   `json:"type"`
	Group        string   `json:"group"`
	Phase        string   `json:"phase"`
	Owner        string   `json:"owner"`
	Tags         []string `json:"tags"`
	Dependencies []string `json:"dependencies"`
	FilePath     string   `json:"file_path"`
}

// fingerprint is the field set the no-mutation check compares. Timestamps are
// deliberately excluded: `created_at` is stable but the zero-valued
// completed/cancelled stamps add nothing, and comparing them would make the
// check brittle for no gain.
func (t Task) fingerprint() string {
	deps := append([]string(nil), t.Dependencies...)
	sort.Strings(deps)
	tags := append([]string(nil), t.Tags...)
	sort.Strings(tags)

	return strings.Join([]string{
		t.ID, t.Title, t.Status, t.Priority, t.Effort, t.Type,
		t.Group, t.Phase, t.Owner,
		strings.Join(tags, "+"), strings.Join(deps, "+"), t.FilePath,
	}, "|")
}

// baseline is the fixture exactly as it is checked in, verified against the
// installed CLI outside the taskmd repo. See ../../../fixtures/README.md.
//
// Verify it from a copy of the fixture placed *outside* this repo — run from
// inside the repo, `taskmd list` picks up taskmd's own task set and returns
// 120+ tasks even with an explicit `-d tasks`.
var baseline = []Task{
	{ID: "001", Title: "Fix login SSO bug", Status: "in-progress", Priority: "high", Effort: "medium", Type: "bug", Group: "cli", Phase: "mvp", Owner: "alex", Tags: []string{"auth", "urgent"}, FilePath: "cli/001-fix-login-sso-bug.md"},
	{ID: "002", Title: "Add full-text search", Status: "pending", Priority: "medium", Effort: "large", Type: "feature", Phase: "mvp", Tags: []string{"search", "backend", "frontend"}, FilePath: "002-add-search-feature.md"},
	{ID: "003", Title: "Patch XSS vulnerability in comments", Status: "pending", Priority: "critical", Effort: "small", Type: "bug", Group: "web", Phase: "mvp", Owner: "sam", Tags: []string{"security", "urgent"}, Dependencies: []string{"002"}, FilePath: "web/003-critical-security-patch.md"},
	{ID: "004", Title: "Update README with setup instructions", Status: "pending", Priority: "low", Effort: "small", Type: "docs", Phase: "polish", Tags: []string{"docs"}, FilePath: "004-update-readme.md"},
	{ID: "005", Title: "Refactor authentication module", Status: "completed", Priority: "high", Effort: "medium", Type: "improvement", Group: "cli", Phase: "mvp", Owner: "alex", Tags: []string{"auth", "refactor"}, FilePath: "cli/005-refactor-auth-module.md"},
	{ID: "006", Title: "Export reports to CSV", Status: "pending", Priority: "high", Effort: "medium", Type: "feature", Group: "web", Phase: "polish", Owner: "jordan", Tags: []string{"reports", "export"}, FilePath: "web/006-export-reports-csv.md"},
}

// listTasks returns every task taskmd can see in the working directory.
//
// `-d tasks` is explicit rather than relying on `.taskmd.yaml`: the bare-project
// variant has no config file, and without the flag taskmd would scan from `.`
// and report root-level tasks as group "tasks". The flag makes every variant
// grade identically.
func listTasks() ([]Task, error) {
	out, err := exec.Command("taskmd", "-d", "tasks", "list", "--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("`taskmd -d tasks list --format json` failed: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, fmt.Errorf("could not parse task list: %w", err)
	}
	return tasks, nil
}

// fixtureUnchanged asserts the agent left the task files exactly as it found
// them. This is the failure mode a read-only skill actually has: an agent that
// "helpfully" edits, reorganizes or completes tasks while answering a question
// about them.
func fixtureUnchanged() error {
	tasks, err := listTasks()
	if err != nil {
		return err
	}

	got := map[string]string{}
	for _, t := range tasks {
		got[t.ID] = t.fingerprint()
	}

	if len(tasks) != len(baseline) {
		return fmt.Errorf("fixture was mutated: %d tasks present, want the %d baseline tasks (%s)",
			len(tasks), len(baseline), strings.Join(sortedKeys(got), ", "))
	}

	for _, want := range baseline {
		have, ok := got[want.ID]
		if !ok {
			return fmt.Errorf("fixture was mutated: task %s is gone", want.ID)
		}
		if have != want.fingerprint() {
			return fmt.Errorf("fixture was mutated: task %s is now\n  %s\nwant\n  %s", want.ID, have, want.fingerprint())
		}
	}

	return validates()
}

// validates asserts that the whole task set still passes `taskmd validate`.
func validates() error {
	out, err := exec.Command("taskmd", "-d", "tasks", "validate").CombinedOutput()
	if err != nil {
		return fmt.Errorf("`taskmd validate` failed: %s", out)
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
