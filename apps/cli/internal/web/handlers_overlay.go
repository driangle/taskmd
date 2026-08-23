package web

import (
	"net/http"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/board"
	"github.com/driangle/taskmd/sdk/go/model"
)

// Overlay-aware handler helpers: how the web API surfaces the merged
// cross-worktree view (spec §9). Everything here is a no-op when the overlay
// is inactive, so single-worktree behavior is unchanged.

// WorktreeOverlayInfo describes the active overlay on /api/config so the
// frontend can render the header indicator ("worktree agent-b — 3 siblings").
// Absent from the response when the overlay is inactive.
type WorktreeOverlayInfo struct {
	// Name is the local worktree's directory basename ("" when the scan dir
	// cannot be resolved to a git worktree, e.g. injected test discovery).
	Name string `json:"name,omitempty"`
	// Siblings is the number of sibling worktrees merged into the view.
	Siblings int `json:"siblings"`
}

// applyProvenance fills the detail's overlay fields. No-op when the overlay
// is inactive or does not know the task.
func (d *TaskDetail) applyProvenance(overlay *worktree.Overlay, taskID string) {
	if overlay == nil {
		return
	}
	ot := overlay.Get(taskID)
	if ot == nil {
		return
	}
	d.EffectiveStatus = string(ot.EffectiveStatus)
	d.EffectiveOwner = ot.EffectiveOwner
	d.Worktree = ot.Worktree
	d.Branch = ot.Branch
	d.RemoteOnly = ot.RemoteOnly
	d.Worktrees = overlay.Copies(taskID)
}

// findMergedTask looks a task up in the local list, falling back to the
// overlay's merged view so sibling-only tasks resolve on read endpoints.
func findMergedTask(dp *DataProvider, taskID string) (*model.Task, *worktree.Overlay, error) {
	tasks, err := dp.GetTasks()
	if err != nil {
		return nil, nil, err
	}
	overlay, err := dp.GetOverlay()
	if err != nil {
		return nil, nil, err
	}
	found := findTaskByID(tasks, taskID)
	if found == nil && overlay != nil {
		if ot := overlay.Get(taskID); ot != nil {
			found = ot.Task
		}
	}
	return found, overlay, nil
}

// overlayBoardGroup mirrors board.JSONGroup with provenance-carrying tasks.
type overlayBoardGroup struct {
	Group string             `json:"group"`
	Count int                `json:"count"`
	Tasks []overlayBoardTask `json:"tasks"`
}

// overlayBoardTask adds the provenance fields board cards render (the badge
// shows when the winning copy is remote, i.e. Worktree is non-empty).
type overlayBoardTask struct {
	board.JSONTask
	Worktree   string `json:"worktree,omitempty"`
	RemoteOnly bool   `json:"remote_only,omitempty"`
}

// annotateBoard decorates board groups with per-task worktree provenance.
// With the overlay inactive it returns the groups untouched, so the payload
// shape is unchanged in single-worktree repos.
func annotateBoard(groups []board.JSONGroup, overlay *worktree.Overlay) any {
	if overlay == nil {
		return groups
	}
	out := make([]overlayBoardGroup, len(groups))
	for i, g := range groups {
		tasks := make([]overlayBoardTask, len(g.Tasks))
		for j, t := range g.Tasks {
			bt := overlayBoardTask{JSONTask: t}
			if ot := overlay.Get(t.ID); ot != nil {
				bt.Worktree = ot.Worktree
				bt.RemoteOnly = ot.RemoteOnly
			}
			tasks[j] = bt
		}
		out[i] = overlayBoardGroup{Group: g.Group, Count: g.Count, Tasks: tasks}
	}
	return out
}

// writeMutationMiss reports a mutation whose task ID resolved to no local
// task. Writes stay strictly local: a task that exists only in a sibling
// worktree gets a visible 409 guard error, never a write; anything else is a
// plain 404.
func writeMutationMiss(w http.ResponseWriter, dp *DataProvider, taskID string) {
	if overlay, err := dp.GetOverlay(); err == nil {
		if guardErr := overlay.SiblingGuard(taskID); guardErr != nil {
			writeError(w, http.StatusConflict, guardErr.Error(), nil)
			return
		}
	}
	writeError(w, http.StatusNotFound, "task not found: "+taskID, nil)
}
