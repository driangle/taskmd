package cli

import (
	"github.com/driangle/taskmd/sdk/go/model"
)

// Cross-worktree resolution for the task detail view. Spec §3 appends
// sibling-only tasks to the read views, and get is a read view: an ID that
// list just displayed must also be addressable by get, showing the sibling
// copy's content with its provenance. Mutations keep the local-only guard.

// getProvenance is the sibling worktree a task lives in when this checkout has
// no copy of it. It is nil for local tasks, so their output is unchanged.
type getProvenance struct {
	// Worktree is the worktree root's basename, as the Worktrees: section and
	// the list view's worktree column name it.
	Worktree string
	// Branch is empty on a detached HEAD.
	Branch string
	// TasksDir is that worktree's absolute tasks directory, which its task
	// file paths are relative to.
	TasksDir string
}

// resolvableTasks returns the task list get resolves a query against: the
// local tasks plus the tasks that exist only in a sibling worktree.
//
// Both sets are resolved in a single pass rather than falling back to siblings
// after a local miss, so exact, filename, and fuzzy matching — and the
// ambiguity errors they raise — behave identically no matter which worktree
// the candidates came from. Sibling-only IDs cannot collide with local ones by
// construction, so exact-ID lookups of local tasks are unaffected.
func resolvableTasks(local []*model.Task, overlay *worktreeOverlay) []*model.Task {
	if overlay == nil {
		return local
	}
	tasks := make([]*model.Task, 0, len(local)+len(overlay.Tasks))
	tasks = append(tasks, local...)
	for _, ot := range overlay.Tasks {
		if ot.RemoteOnly {
			tasks = append(tasks, ot.Task)
		}
	}
	return tasks
}

// remoteProvenance returns where a task lives when it exists only in a sibling
// worktree, and nil when this checkout has its own copy.
func remoteProvenance(overlay *worktreeOverlay, taskID string) *getProvenance {
	if overlay == nil {
		return nil
	}
	ot := overlay.Get(taskID)
	if ot == nil || !ot.RemoteOnly {
		return nil
	}
	p := &getProvenance{Worktree: ot.Worktree, Branch: ot.Branch}
	if wt, ok := ot.Origin(); ok {
		p.TasksDir = wt.TasksDir
	}
	return p
}

// branchLabel names the provenance branch for display, standing in for an
// empty branch on a detached HEAD.
func (p *getProvenance) branchLabel() string {
	if p.Branch == "" {
		return "detached"
	}
	return p.Branch
}
