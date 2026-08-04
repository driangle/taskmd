package cli

import (
	"fmt"
	"os"

	"github.com/driangle/taskmd/sdk/go/model"
)

// ProjectTask wraps a task with its originating project ID.
type ProjectTask struct {
	ProjectID string `json:"project" yaml:"project"`
	*model.Task
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

	var all []*ProjectTask
	for _, entry := range entries {
		tasks, err := scanProjectTasks(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping project %q: %v\n", entry.ID, err)
			continue
		}
		for _, t := range tasks {
			all = append(all, &ProjectTask{ProjectID: entry.ID, Task: t})
		}
	}

	return all, nil
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

// scanProjectTasks scans a single project and returns its tasks with relative file paths.
func scanProjectTasks(entry GlobalProjectEntry) ([]*model.Task, error) {
	info, err := os.Stat(entry.Path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("path %q is not accessible", entry.Path)
	}

	scanDir := resolveProjectScanDir(entry.Path)
	taskScanner := newTaskScanner(scanDir, GlobalFlags{})
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	makeFilePathsRelative(result.Tasks, scanDir)
	return result.Tasks, nil
}
