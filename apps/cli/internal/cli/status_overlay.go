package cli

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"

	"github.com/driangle/taskmd/sdk/go/model"
)

// Cross-worktree behavior for the status views (spec §4).
//
// status selects and displays against effective status — the most advanced
// copy across worktrees — and names the worktree that copy came from.
//
// It splits the two the same way list already does, which is the house
// convention for a merged read view:
//
//   - Text output renders the effective status, because a command that just
//     selected a task as in-progress printing "Status: pending" is incoherent.
//     The Worktree: line and Worktrees: section below it disclose the local
//     copy, so nothing is hidden.
//   - json/yaml keep status as the *local* copy's and add effective_status
//     alongside, matching list, the MCP status/get/list tools, and the web
//     API. Structured consumers get one stable meaning for `status` across
//     every taskmd surface.
//
// --statusline is deliberately exempt from all of this: it feeds a shell or
// agent statusline answering "what am I working on in *this* checkout", so it
// reads local files only. Surfacing a task a sibling worktree claimed would
// assert work this agent is not doing. runStatusList holds that branch.

// statusViewTasks returns the task list the status views render: with the
// overlay active, statuses are effective across worktrees and sibling-only
// tasks are included (so they are addressable by `status <id>`, exactly as
// get resolves them); otherwise it is the local scan unchanged.
//
// Unlike get, this needs no resolvableTasks call — EffectiveTasks already
// includes the sibling-only tasks, and appending them again would duplicate
// every one of them.
func statusViewTasks(local []*model.Task, overlay *worktreeOverlay) []*model.Task {
	if overlay == nil {
		return local
	}
	return overlay.EffectiveTasks()
}

// annotateStatusProvenance fills in where a status entry's winning copy came
// from. It is a no-op for local-winning tasks, which keeps every field absent
// from json/yaml and every extra line out of text output.
func annotateStatusProvenance(out *statusOutput, overlay *worktreeOverlay) {
	if overlay == nil {
		return
	}
	ot := overlay.Get(out.ID)
	if ot == nil || ot.Worktree == "" {
		return
	}
	// The task was built from the effective-status list, so restore the local
	// copy's status and carry the merged one separately. ot.Task is the local
	// copy whenever this checkout has one, and the sibling's otherwise — in
	// which case the two are identically that single copy's status.
	out.Status = string(ot.Task.Status)
	out.EffectiveStatus = string(ot.EffectiveStatus)
	out.Worktree = ot.Worktree
	out.Branch = ot.Branch
	out.RemoteOnly = ot.RemoteOnly
	out.Worktrees = overlay.Copies(out.ID)
}

// displayStatus is the status text output shows: the merged one when the
// overlay annotated this entry, else the task's own.
func displayStatus(out statusOutput) string {
	if out.EffectiveStatus != "" {
		return out.EffectiveStatus
	}
	return out.Status
}

// printStatusProvenance names the worktree the reported status came from,
// directly under the Status line it explains. Sibling-only tasks additionally
// carry get's remote-only warning, since a mutation has to be run over there.
func printStatusProvenance(w io.Writer, out statusOutput, r *lipgloss.Renderer) {
	if out.Worktree == "" {
		return
	}
	line := fmt.Sprintf("%s %s", formatLabel("Worktree:", r),
		fmt.Sprintf("%s (branch %s)", out.Worktree, statusBranchLabel(out.Branch)))
	if out.RemoteOnly {
		line += " " + formatWarning("(remote-only: no copy in this worktree)", r)
	}
	fmt.Fprintln(w, line)
}

// statusBranchLabel names a provenance branch for display, standing in for an
// empty branch on a detached HEAD.
func statusBranchLabel(branch string) string {
	if branch == "" {
		return "detached"
	}
	return branch
}
