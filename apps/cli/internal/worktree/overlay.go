// Package worktree implements the cross-worktree overlay: a merged task view
// where coordination state (status, owner) is combined across all worktrees of
// one repository while content always comes from the local copy. It is shared
// by the CLI read views, the MCP server, and the web server's data layer.
package worktree

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

// Task decorates a task with cross-worktree provenance. The embedded task is
// the local copy when one exists; content always comes from it, only
// coordination state (status, owner) is merged across worktrees.
type Task struct {
	*model.Task
	EffectiveStatus model.Status `json:"effective_status" yaml:"effective_status"`
	EffectiveOwner  string       `json:"effective_owner,omitempty" yaml:"effective_owner,omitempty"`
	Worktree        string       `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch          string       `json:"branch,omitempty" yaml:"branch,omitempty"`
	LocalOnly       bool         `json:"-" yaml:"-"`
	RemoteOnly      bool         `json:"remote_only,omitempty" yaml:"remote_only,omitempty"`
	// origin is the sibling worktree the winning copy was scanned from; nil
	// when the winning copy is the local one.
	origin *gitmeta.Worktree
}

// Origin returns the sibling worktree the winning copy of this task came
// from, and false when that copy is local. Detail views use it to resolve
// paths — a worklog, say — inside the checkout that actually holds the file.
func (t *Task) Origin() (gitmeta.Worktree, bool) {
	if t.origin == nil {
		return gitmeta.Worktree{}, false
	}
	return *t.origin, true
}

// Overlay is the merged cross-worktree task view: local tasks in scan order,
// then tasks that exist only in sibling worktrees, sorted by ID.
type Overlay struct {
	Tasks    []*Task
	byID     map[string]*Task
	Warnings []Warning
	// local is the local task list after sibling-root attribution (§8):
	// copies scanned from a checkout nested inside the scan root belong to
	// that worktree and are absent here.
	local []*model.Task
	// copies indexes every copy of every task by ID, local copy first.
	copies map[string][]taskCopy
}

// Warning is a cross-worktree consistency warning tied to a task.
type Warning struct {
	TaskID  string
	Message string
}

// branchLabel names the winning copy's branch for messages, standing in for
// an empty branch on a detached HEAD.
func (t *Task) branchLabel() string {
	if t.Branch == "" {
		return "detached"
	}
	return t.Branch
}

// ExclusionReason returns why next must not recommend this task, or "" when
// the overlay imposes no exclusion. A task is excluded when it exists only in
// a sibling worktree, or when a sibling copy advanced it beyond the local
// status — a locally in-progress task keeps today's resume semantics even if
// a sibling also claims it.
func (t *Task) ExclusionReason() string {
	if t.RemoteOnly {
		return fmt.Sprintf("exists only in worktree %s (branch %s)", t.Worktree, t.branchLabel())
	}
	if t.Worktree == "" || statusLadderRank(t.EffectiveStatus) <= statusLadderRank(t.Task.Status) {
		return ""
	}
	return fmt.Sprintf("%s in worktree %s (branch %s)", t.EffectiveStatus, t.Worktree, t.branchLabel())
}

// Exclusion is one task the overlay keeps out of next's recommendations,
// together with the provenance that explains why. Reason is the same sentence
// the table view prints; the remaining fields carry it structurally for
// json/yaml consumers.
type Exclusion struct {
	ID       string `json:"id" yaml:"id"`
	Reason   string `json:"reason" yaml:"reason"`
	Worktree string `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch   string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Status   string `json:"status,omitempty" yaml:"status,omitempty"`
}

// Exclusions returns every overlay-imposed exclusion, sorted by task ID.
func (o *Overlay) Exclusions() []Exclusion {
	var exclusions []Exclusion
	for _, ot := range o.Tasks {
		reason := ot.ExclusionReason()
		if reason == "" {
			continue
		}
		exclusions = append(exclusions, Exclusion{
			ID:       ot.Task.ID,
			Reason:   reason,
			Worktree: ot.Worktree,
			Branch:   ot.Branch,
			Status:   string(ot.EffectiveStatus),
		})
	}
	sort.Slice(exclusions, func(i, j int) bool { return exclusions[i].ID < exclusions[j].ID })
	return exclusions
}

// RecommendationInputs returns the task list and exclusion map to feed the
// next recommender. Effective statuses are substituted on copies (so a task
// completed in a sibling unblocks its local dependents) and sibling-suppressed
// tasks are mapped to a human-readable exclusion reason; they stay in the task
// list so dependency, children, and critical-path resolution still see them.
func (o *Overlay) RecommendationInputs() ([]*model.Task, map[string]string) {
	exclusions := o.Exclusions()
	excluded := make(map[string]string, len(exclusions))
	for _, e := range exclusions {
		excluded[e.ID] = e.Reason
	}
	return o.EffectiveTasks(), excluded
}

// EffectiveTasks returns the merged task list with the effective status
// substituted on shallow copies: local tasks in scan order, then sibling-only
// tasks. Status-aggregating views (board, stats, graph, report, tracks,
// phases) render this; content still comes from the base copy.
func (o *Overlay) EffectiveTasks() []*model.Task {
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

// Local returns the local task list after sibling-root attribution.
func (o *Overlay) Local() []*model.Task {
	return o.local
}

// Get returns the overlay task for id, or nil when the id is unknown.
func (o *Overlay) Get(id string) *Task {
	return o.byID[id]
}

// CopyEntry is one copy of a task in one worktree, shaped for get's
// Worktrees section and its JSON/YAML output.
type CopyEntry struct {
	Worktree string `json:"worktree,omitempty" yaml:"worktree,omitempty"`
	Branch   string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Status   string `json:"status" yaml:"status"`
	Owner    string `json:"owner,omitempty" yaml:"owner,omitempty"`
	Local    bool   `json:"local,omitempty" yaml:"local,omitempty"`
}

// Copies returns an entry for every copy of id across worktrees when the
// copies disagree on status or owner, local copy first; nil when there is a
// single copy or all copies agree (detail views render the section only on
// divergence).
func (o *Overlay) Copies(id string) []CopyEntry {
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

	entries := make([]CopyEntry, 0, len(copies))
	for _, c := range copies {
		entry := CopyEntry{
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

// SiblingTasks pairs a sibling worktree with the tasks scanned from it.
type SiblingTasks struct {
	WT    gitmeta.Worktree
	Tasks []*model.Task
}

// taskCopy is one copy of a task in one worktree; wt is nil for the local copy.
type taskCopy struct {
	task *model.Task
	wt   *gitmeta.Worktree
}

// Merge merges the local task list with sibling worktree scans. It is pure
// with respect to git: worktree discovery and scanning happen upstream.
func Merge(local []*model.Task, siblings []SiblingTasks) *Overlay {
	copies := indexCopies(local, siblings)

	overlay := &Overlay{
		byID:   make(map[string]*Task, len(copies)),
		local:  local,
		copies: copies,
	}
	add := func(ot *Task) {
		overlay.Tasks = append(overlay.Tasks, ot)
		overlay.byID[ot.Task.ID] = ot
		if warning := divergentTerminalWarning(ot.Task.ID, copies[ot.Task.ID]); warning != "" {
			overlay.Warnings = append(overlay.Warnings, Warning{TaskID: ot.Task.ID, Message: warning})
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
func indexCopies(local []*model.Task, siblings []SiblingTasks) map[string][]taskCopy {
	copies := make(map[string][]taskCopy)
	for _, task := range local {
		copies[task.ID] = append(copies[task.ID], taskCopy{task: task})
	}
	for i := range siblings {
		wt := &siblings[i].WT
		for _, task := range siblings[i].Tasks {
			copies[task.ID] = append(copies[task.ID], taskCopy{task: task, wt: wt})
		}
	}
	return copies
}

// mergeCopies builds the overlay Task for one ID. The base is the local copy
// when given, else the winning sibling copy; effective status and owner come
// from the winner. LocalOnly/RemoteOnly are set by the caller's context.
func mergeCopies(localTask *model.Task, all []taskCopy) *Task {
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

	ot := &Task{
		Task:            base,
		EffectiveStatus: winner.task.Status,
		EffectiveOwner:  winner.task.Owner,
		LocalOnly:       localTask != nil && len(all) == 1,
	}
	if winner.wt != nil {
		ot.Worktree = filepath.Base(winner.wt.Root)
		ot.Branch = winner.wt.Branch
		ot.origin = winner.wt
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

// RelativizeSiblingPaths rewrites each sibling copy's file path relative to
// its own worktree's tasks dir, for display. Call it only after the merge —
// status ties are broken by stat'ing the absolute paths. Server surfaces skip
// this so file paths stay resolvable (e.g. for worklog lookups).
func (o *Overlay) RelativizeSiblingPaths() {
	for _, copies := range o.copies {
		for _, c := range copies {
			if c.wt == nil {
				continue
			}
			absBase, err := filepath.Abs(c.wt.TasksDir)
			if err != nil {
				continue
			}
			if rel, err := filepath.Rel(absBase, c.task.FilePath); err == nil {
				c.task.FilePath = rel
			}
		}
	}
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
