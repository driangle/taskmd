package cli

import (
	"github.com/driangle/taskmd/sdk/go/model"
)

// The status view's output model: the lightweight metadata shape status
// renders in every format, and the projection from a scanned task onto it.
// Rendering lives in status_text.go; cross-worktree provenance in
// status_overlay.go.

// statusChild represents a child task in the recursive children tree.
type statusChild struct {
	ID       string        `json:"id" yaml:"id"`
	Title    string        `json:"title" yaml:"title"`
	Status   string        `json:"status" yaml:"status"`
	Children []statusChild `json:"children,omitempty" yaml:"children,omitempty"`
}

// statusOutput is the lightweight metadata struct for JSON/YAML output.
type statusOutput struct {
	ID           string   `json:"id" yaml:"id"`
	Title        string   `json:"title" yaml:"title"`
	Status       string   `json:"status" yaml:"status"`
	Priority     string   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort       string   `json:"effort,omitempty" yaml:"effort,omitempty"`
	Tags         []string `json:"tags" yaml:"tags"`
	Owner        string   `json:"owner,omitempty" yaml:"owner,omitempty"`
	Parent       string   `json:"parent,omitempty" yaml:"parent,omitempty"`
	Created      string   `json:"created,omitempty" yaml:"created,omitempty"`
	Dependencies []string `json:"dependencies" yaml:"dependencies"`
	Blocked      *bool    `json:"blocked,omitempty" yaml:"blocked,omitempty"`
	BlockedBy    []string `json:"blocked_by,omitempty" yaml:"blocked_by,omitempty"`
	Group        string   `json:"group,omitempty" yaml:"group,omitempty"`
	FilePath     string   `json:"file_path" yaml:"file_path"`
	// Cross-worktree provenance, all absent unless the overlay is active and a
	// sibling copy won the merge. Status above stays the local copy's, as in
	// list and the MCP tools; EffectiveStatus is the merged one. See
	// status_overlay.go.
	EffectiveStatus string              `json:"effective_status,omitempty" yaml:"effective_status,omitempty"`
	Worktree        string              `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch          string              `json:"branch,omitempty" yaml:"branch,omitempty"`
	RemoteOnly      bool                `json:"remote_only,omitempty" yaml:"remote_only,omitempty"`
	Worktrees       []worktreeCopyEntry `json:"worktrees,omitempty" yaml:"worktrees,omitempty"`
	Children        []statusChild       `json:"children,omitempty" yaml:"children,omitempty"`
}

func buildStatusOutputFromTask(
	task *model.Task,
	childrenIndex map[string][]*model.Task,
	tasksByID map[string]*model.Task,
) statusOutput {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	out := statusOutput{
		ID:           task.ID,
		Title:        task.Title,
		Status:       string(task.Status),
		Priority:     string(task.Priority),
		Effort:       string(task.Effort),
		Tags:         task.Tags,
		Owner:        task.Owner,
		Parent:       task.Parent,
		Created:      created,
		Dependencies: task.Dependencies,
		Group:        task.Group,
		FilePath:     task.FilePath,
	}
	if len(task.Dependencies) > 0 {
		out.BlockedBy = resolveBlockingDeps(task.Dependencies, tasksByID)
		blocked := len(out.BlockedBy) > 0
		out.Blocked = &blocked
		if !blocked {
			out.BlockedBy = nil
		}
	}
	if childrenIndex != nil {
		out.Children = collectChildrenTree(task.ID, childrenIndex, map[string]bool{task.ID: true})
	}
	return out
}

func resolveBlockingDeps(deps []string, tasksByID map[string]*model.Task) []string {
	var blocking []string
	for _, depID := range deps {
		dep, ok := tasksByID[depID]
		if !ok || dep.Status != model.StatusCompleted {
			blocking = append(blocking, depID)
		}
	}
	return blocking
}

func buildTasksByIDMap(tasks []*model.Task) map[string]*model.Task {
	m := make(map[string]*model.Task, len(tasks))
	for _, t := range tasks {
		m[t.ID] = t
	}
	return m
}

func buildChildrenIndex(tasks []*model.Task) map[string][]*model.Task {
	index := make(map[string][]*model.Task)
	for _, t := range tasks {
		if t.Parent != "" {
			index[t.Parent] = append(index[t.Parent], t)
		}
	}
	return index
}

func collectChildrenTree(taskID string, index map[string][]*model.Task, visited map[string]bool) []statusChild {
	children := index[taskID]
	if len(children) == 0 {
		return nil
	}
	result := make([]statusChild, 0, len(children))
	for _, child := range children {
		if visited[child.ID] {
			continue
		}
		visited[child.ID] = true
		sc := statusChild{
			ID:     child.ID,
			Title:  child.Title,
			Status: string(child.Status),
		}
		sc.Children = collectChildrenTree(child.ID, index, visited)
		result = append(result, sc)
	}
	return result
}
