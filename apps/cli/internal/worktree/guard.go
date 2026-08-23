package worktree

import (
	"fmt"
	"os"
	"path/filepath"
)

// SiblingOnlyError is the mutation guard error for a task that exists only in
// a sibling worktree. Writes are never redirected — the mutation must fail and
// tell the user where the task actually lives.
type SiblingOnlyError struct {
	TaskID   string
	Worktree string // display path of the sibling worktree root
	Branch   string // "" when detached
}

func (e *SiblingOnlyError) Error() string {
	branch := e.Branch
	if branch == "" {
		branch = "detached"
	}
	return fmt.Sprintf("task %s exists only in worktree %s (branch %s); run taskmd there",
		e.TaskID, e.Worktree, branch)
}

// SiblingGuard returns the guard error when taskID exists only in a sibling
// worktree, and nil when it has a local copy or is unknown. Callers invoke it
// after a local lookup misses, so mutations name where the task actually lives
// instead of reporting "not found".
func (o *Overlay) SiblingGuard(taskID string) error {
	if o == nil {
		return nil
	}
	ot := o.byID[taskID]
	if ot == nil || !ot.RemoteOnly {
		return nil
	}
	for _, c := range o.copies[taskID] {
		if c.wt != nil {
			return &SiblingOnlyError{
				TaskID:   taskID,
				Worktree: DisplayPath(c.wt.Root),
				Branch:   c.wt.Branch,
			}
		}
	}
	return nil
}

// SiblingGuard scans the builder's sibling worktrees for taskID and returns
// the guard error when a copy exists there; nil otherwise. It is the
// overlay-less variant for mutation paths that never built the full merge.
func (b Builder) SiblingGuard(taskID, scanDir string) error {
	siblings, err := b.Siblings(scanDir)
	if err != nil {
		return nil
	}
	for _, st := range b.scanSiblings(siblings) {
		for _, t := range st.Tasks {
			if t.ID == taskID {
				return &SiblingOnlyError{
					TaskID:   taskID,
					Worktree: DisplayPath(st.WT.Root),
					Branch:   st.WT.Branch,
				}
			}
		}
	}
	return nil
}

// DisplayPath renders a worktree root relative to the current directory when
// possible (matching how users navigate between worktrees), falling back to
// the absolute path.
func DisplayPath(root string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return root
	}
	rel, err := filepath.Rel(cwd, root)
	if err != nil {
		return root
	}
	return rel
}
