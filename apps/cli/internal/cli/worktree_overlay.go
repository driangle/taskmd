package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/sdk/go/model"
)

// statusLadderRank orders statuses by how advanced they are, per the worktree
// spec ladder: pending < blocked < in-progress < in-review < cancelled <
// completed. The most advanced copy of a task across worktrees wins. Unknown
// statuses rank lowest, alongside pending.
func statusLadderRank(s model.Status) int {
	switch s {
	case model.StatusBlocked:
		return 1
	case model.StatusInProgress:
		return 2
	case model.StatusInReview:
		return 3
	case model.StatusCancelled:
		return 4
	case model.StatusCompleted:
		return 5
	default:
		return 0
	}
}

// OverlayTask decorates a task with cross-worktree provenance. The embedded
// task is the local copy when one exists; content always comes from it, only
// coordination state (status, owner) is merged across worktrees.
type OverlayTask struct {
	*model.Task
	EffectiveStatus model.Status `json:"effective_status" yaml:"effective_status"`
	EffectiveOwner  string       `json:"effective_owner,omitempty" yaml:"effective_owner,omitempty"`
	Worktree        string       `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch          string       `json:"branch,omitempty" yaml:"branch,omitempty"`
	LocalOnly       bool         `json:"-" yaml:"-"`
	RemoteOnly      bool         `json:"remote_only,omitempty" yaml:"remote_only,omitempty"`
}

// worktreeOverlay is the merged cross-worktree task view: local tasks in scan
// order, then tasks that exist only in sibling worktrees, sorted by ID.
type worktreeOverlay struct {
	Tasks    []*OverlayTask
	byID     map[string]*OverlayTask
	Warnings []overlayWarning
	// local is the local task list after sibling-root attribution (§8):
	// copies scanned from a checkout nested inside the scan root belong to
	// that worktree and are absent here.
	local []*model.Task
	// copies indexes every copy of every task by ID, local copy first.
	copies map[string][]taskCopy
}

// overlayWarning is a cross-worktree consistency warning tied to a task.
type overlayWarning struct {
	TaskID  string
	Message string
}

// branchLabel names the winning copy's branch for messages, standing in for
// an empty branch on a detached HEAD.
func (ot *OverlayTask) branchLabel() string {
	if ot.Branch == "" {
		return "detached"
	}
	return ot.Branch
}

// exclusionReason returns why next must not recommend this task, or "" when
// the overlay imposes no exclusion. A task is excluded when it exists only in
// a sibling worktree, or when a sibling copy advanced it beyond the local
// status — a locally in-progress task keeps today's resume semantics even if
// a sibling also claims it.
func (ot *OverlayTask) exclusionReason() string {
	if ot.RemoteOnly {
		return fmt.Sprintf("exists only in worktree %s (branch %s)", ot.Worktree, ot.branchLabel())
	}
	if ot.Worktree == "" || statusLadderRank(ot.EffectiveStatus) <= statusLadderRank(ot.Task.Status) {
		return ""
	}
	return fmt.Sprintf("%s in worktree %s (branch %s)", ot.EffectiveStatus, ot.Worktree, ot.branchLabel())
}

// recommendationInputs returns the task list and exclusion map to feed the
// next recommender. Effective statuses are substituted on copies (so a task
// completed in a sibling unblocks its local dependents) and sibling-suppressed
// tasks are mapped to a human-readable exclusion reason; they stay in the task
// list so dependency, children, and critical-path resolution still see them.
func (o *worktreeOverlay) recommendationInputs() ([]*model.Task, map[string]string) {
	excluded := make(map[string]string)
	for _, ot := range o.Tasks {
		if reason := ot.exclusionReason(); reason != "" {
			excluded[ot.Task.ID] = reason
		}
	}
	return o.effectiveTasks(), excluded
}

// effectiveTasks returns the merged task list with the effective status
// substituted on shallow copies: local tasks in scan order, then sibling-only
// tasks. Status-aggregating views (board, stats, graph, report, tracks,
// phases) render this; content still comes from the base copy.
func (o *worktreeOverlay) effectiveTasks() []*model.Task {
	tasks := make([]*model.Task, 0, len(o.Tasks))
	for _, ot := range o.Tasks {
		task := ot.Task
		if ot.EffectiveStatus != task.Status {
			effective := *task
			effective.Status = ot.EffectiveStatus
			task = &effective
		}
		tasks = append(tasks, task)
	}
	return tasks
}

// worktreeCopyEntry is one copy of a task in one worktree, shaped for get's
// Worktrees section and its JSON/YAML output.
type worktreeCopyEntry struct {
	Worktree string `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch   string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Status   string `json:"status" yaml:"status"`
	Owner    string `json:"owner,omitempty" yaml:"owner,omitempty"`
	Local    bool   `json:"local,omitempty" yaml:"local,omitempty"`
}

// worktreeCopies returns an entry for every copy of id across worktrees when
// the copies disagree on status or owner, local copy first; nil when there is
// a single copy or all copies agree (get renders the section only on
// divergence).
func (o *worktreeOverlay) worktreeCopies(id string) []worktreeCopyEntry {
	copies := o.copies[id]
	if len(copies) < 2 {
		return nil
	}
	first := copies[0].task
	differ := false
	for _, c := range copies[1:] {
		if c.task.Status != first.Status || c.task.Owner != first.Owner {
			differ = true
			break
		}
	}
	if !differ {
		return nil
	}

	entries := make([]worktreeCopyEntry, 0, len(copies))
	for _, c := range copies {
		entry := worktreeCopyEntry{
			Status: string(c.task.Status),
			Owner:  c.task.Owner,
			Local:  c.wt == nil,
		}
		if c.wt != nil {
			entry.Worktree = filepath.Base(c.wt.Root)
			entry.Branch = c.wt.Branch
		}
		entries = append(entries, entry)
	}
	return entries
}

// siblingTasks pairs a sibling worktree with the tasks scanned from it.
type siblingTasks struct {
	wt    gitmeta.Worktree
	tasks []*model.Task
}

// taskCopy is one copy of a task in one worktree; wt is nil for the local copy.
type taskCopy struct {
	task *model.Task
	wt   *gitmeta.Worktree
}

// mergeOverlay merges the local task list with sibling worktree scans. It is
// pure with respect to git: worktree discovery and scanning happen upstream.
func mergeOverlay(local []*model.Task, siblings []siblingTasks) *worktreeOverlay {
	copies := indexCopies(local, siblings)

	overlay := &worktreeOverlay{
		byID:   make(map[string]*OverlayTask, len(copies)),
		local:  local,
		copies: copies,
	}
	add := func(ot *OverlayTask) {
		overlay.Tasks = append(overlay.Tasks, ot)
		overlay.byID[ot.Task.ID] = ot
		if warning := divergentTerminalWarning(ot.Task.ID, copies[ot.Task.ID]); warning != "" {
			overlay.Warnings = append(overlay.Warnings, overlayWarning{TaskID: ot.Task.ID, Message: warning})
		}
	}

	for _, task := range local {
		add(mergeCopies(task, copies[task.ID]))
	}

	var remoteOnlyIDs []string
	for id := range copies {
		if _, seen := overlay.byID[id]; !seen {
			remoteOnlyIDs = append(remoteOnlyIDs, id)
		}
	}
	sort.Strings(remoteOnlyIDs)
	for _, id := range remoteOnlyIDs {
		ot := mergeCopies(nil, copies[id])
		ot.RemoteOnly = true
		add(ot)
	}

	return overlay
}

// indexCopies groups every copy of every task by ID, local copy first.
func indexCopies(local []*model.Task, siblings []siblingTasks) map[string][]taskCopy {
	copies := make(map[string][]taskCopy)
	for _, task := range local {
		copies[task.ID] = append(copies[task.ID], taskCopy{task: task})
	}
	for i := range siblings {
		wt := &siblings[i].wt
		for _, task := range siblings[i].tasks {
			copies[task.ID] = append(copies[task.ID], taskCopy{task: task, wt: wt})
		}
	}
	return copies
}

// mergeCopies builds the OverlayTask for one ID. The base is the local copy
// when given, else the winning sibling copy; effective status and owner come
// from the winner. LocalOnly/RemoteOnly are set by the caller's context.
func mergeCopies(localTask *model.Task, all []taskCopy) *OverlayTask {
	winner := all[0]
	for _, c := range all[1:] {
		if beatsCopy(c, winner) {
			winner = c
		}
	}

	base := localTask
	if base == nil {
		base = winner.task
	}

	ot := &OverlayTask{
		Task:            base,
		EffectiveStatus: winner.task.Status,
		EffectiveOwner:  winner.task.Owner,
		LocalOnly:       localTask != nil && len(all) == 1,
	}
	if winner.wt != nil {
		ot.Worktree = filepath.Base(winner.wt.Root)
		ot.Branch = winner.wt.Branch
	}
	return ot
}

// beatsCopy reports whether challenger should replace current as the winning
// copy: a more advanced status wins outright; the same status falls back to
// file mtime, newest wins (a stat failure never wins a tie).
func beatsCopy(challenger, current taskCopy) bool {
	cr, wr := statusLadderRank(challenger.task.Status), statusLadderRank(current.task.Status)
	if cr != wr {
		return cr > wr
	}
	return fileMTime(challenger.task.FilePath).After(fileMTime(current.task.FilePath))
}

// fileMTime returns the file's modification time, or the zero time when the
// file cannot be stat'ed.
func fileMTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// divergentTerminalWarning returns a warning line when copies of a task sit in
// different terminal states (completed in one worktree, cancelled in another).
func divergentTerminalWarning(id string, all []taskCopy) string {
	var completedIn, cancelledIn string
	for _, c := range all {
		where := "this worktree"
		if c.wt != nil {
			where = "worktree " + filepath.Base(c.wt.Root)
		}
		switch c.task.Status {
		case model.StatusCompleted:
			completedIn = where
		case model.StatusCancelled:
			cancelledIn = where
		}
	}
	if completedIn == "" || cancelledIn == "" {
		return ""
	}
	return fmt.Sprintf("task %s is completed in %s but cancelled in %s", id, completedIn, cancelledIn)
}
