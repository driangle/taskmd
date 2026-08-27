package cli

import (
	"fmt"
	"os"

	"github.com/driangle/taskmd/sdk/go/model"
)

// ProjectTask wraps a task with its originating project ID. When that
// project's worktree overlay is active the embedded task carries the effective
// (cross-worktree) status and the provenance fields name the winning copy;
// they are omitted entirely otherwise, so single-worktree and non-git projects
// serialize exactly as before.
type ProjectTask struct {
	ProjectID string `json:"project" yaml:"project"`
	*model.Task
	Worktree   string `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch     string `json:"branch,omitempty" yaml:"branch,omitempty"`
	RemoteOnly bool   `json:"remote_only,omitempty" yaml:"remote_only,omitempty"`
}

// QualifiedID returns the task ID prefixed with the project ID.
func (pt *ProjectTask) QualifiedID() string {
	return pt.ProjectID + ":" + pt.Task.ID
}

// scanAllProjects loads tasks from all registered projects, qualifying IDs.
// Unreachable projects are skipped with a warning to stderr.
func scanAllProjects() ([]*ProjectTask, error) {
	entries, err := LoadGlobalRegistry()
	if err != nil {
		return nil, fmt.Errorf("load global registry: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no projects registered in global registry")
	}

	// Scan each repository once, from the current worktree when the command
	// runs inside it (ADR 0005 §1).
	entries = dedupeRepoEntries(entries)

	var all []*ProjectTask
	for _, entry := range entries {
		scan, err := scanProject(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping project %q: %v\n", entry.ID, err)
			continue
		}
		all = append(all, projectTasks(entry.ID, scan)...)
	}

	return all, nil
}

// projectTasks wraps one project's scanned tasks, attaching worktree
// provenance when that project's overlay is active.
func projectTasks(projectID string, scan projectScan) []*ProjectTask {
	tasks := make([]*ProjectTask, 0, len(scan.tasks))
	for _, t := range scan.tasks {
		pt := &ProjectTask{ProjectID: projectID, Task: t}
		if ot := scan.provenance(t.ID); ot != nil {
			pt.Worktree = ot.Worktree
			pt.Branch = ot.Branch
			pt.RemoteOnly = ot.RemoteOnly
		}
		tasks = append(tasks, pt)
	}
	return tasks
}

// projectWorktreeCell renders a task's worktree provenance, marking
// sibling-only tasks with a trailing * exactly as the single-project list does.
func projectWorktreeCell(pt *ProjectTask) string {
	cell := pt.Worktree
	if pt.RemoteOnly {
		cell += "*"
	}
	return cell
}

// annotated reports whether any task carries sibling-worktree provenance,
// deciding whether the WORKTREE column is worth rendering.
func annotated(ptasks []*ProjectTask) bool {
	for _, pt := range ptasks {
		if pt.Worktree != "" || pt.RemoteOnly {
			return true
		}
	}
	return false
}

// injectProjectColumn adds a "project" column after the first column if not already present.
func injectProjectColumn(columns []string) []string {
	for _, col := range columns {
		if col == "project" {
			return columns
		}
	}
	insertIdx := 1
	if len(columns) < 2 {
		insertIdx = len(columns)
	}
	result := make([]string, 0, len(columns)+1)
	result = append(result, columns[:insertIdx]...)
	result = append(result, "project")
	result = append(result, columns[insertIdx:]...)
	return result
}
