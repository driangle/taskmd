package mcp

import (
	"fmt"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// resolveTaskDir applies the tools' shared default for an omitted task_dir.
func resolveTaskDir(taskDir string) string {
	if taskDir == "" {
		return "."
	}
	return taskDir
}

// scanWithOverlay scans taskDir and builds the worktree overlay when it is
// active (nil otherwise). The returned tasks are the local tasks after
// sibling-root attribution, so read tools that ignore the overlay behave
// exactly as before when it is inactive.
func scanWithOverlay(taskDir string, wt worktree.Builder) ([]*model.Task, *worktree.Overlay, error) {
	dir := resolveTaskDir(taskDir)
	result, err := scanner.NewScanner(dir, false, nil).Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}

	overlay, err := wt.Build(dir, result.Tasks)
	if err != nil {
		return nil, nil, err
	}

	tasks := result.Tasks
	if overlay != nil {
		tasks = overlay.Local()
	}
	return tasks, overlay, nil
}

// effectiveOrLocal returns the merged task list with effective statuses when
// the overlay is active, and the local tasks otherwise.
func effectiveOrLocal(tasks []*model.Task, overlay *worktree.Overlay) []*model.Task {
	if overlay != nil {
		return overlay.EffectiveTasks()
	}
	return tasks
}

// provenanceFields is the additive overlay block embedded in per-task tool
// outputs. All fields are empty (and omitted from JSON) when the overlay is
// inactive, so existing clients see an unchanged shape.
type provenanceFields struct {
	EffectiveStatus string               `json:"effective_status,omitempty"`
	EffectiveOwner  string               `json:"effective_owner,omitempty"`
	Worktree        string               `json:"worktree,omitempty"`
	Branch          string               `json:"branch,omitempty"`
	RemoteOnly      bool                 `json:"remote_only,omitempty"`
	Worktrees       []worktree.CopyEntry `json:"worktrees,omitempty"`
}

// provenanceFor extracts the additive overlay fields for a task, including
// the per-worktree copies section when copies diverge. Returns the zero value
// when the overlay is inactive or does not know the task.
func provenanceFor(overlay *worktree.Overlay, taskID string) provenanceFields {
	if overlay == nil {
		return provenanceFields{}
	}
	ot := overlay.Get(taskID)
	if ot == nil {
		return provenanceFields{}
	}
	return provenanceFields{
		EffectiveStatus: string(ot.EffectiveStatus),
		EffectiveOwner:  ot.EffectiveOwner,
		Worktree:        ot.Worktree,
		Branch:          ot.Branch,
		RemoteOnly:      ot.RemoteOnly,
		Worktrees:       overlay.Copies(taskID),
	}
}
