package web

import (
	"net/http"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
)

// Overlay-aware handler helpers: how the web API surfaces the merged
// cross-worktree view (spec §9). Everything here is a no-op when the overlay
// is inactive, so single-worktree behavior is unchanged.

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
