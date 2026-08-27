package cli

import (
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
)

// The overlay merge layer lives in internal/worktree so the MCP and web
// servers can share it (internal/cli imports both, so it cannot host code
// they need). These aliases keep the CLI's historical names.
type (
	// OverlayTask decorates a task with cross-worktree provenance.
	OverlayTask = worktree.Task
	// worktreeOverlay is the merged cross-worktree task view.
	worktreeOverlay = worktree.Overlay
	// worktreeCopyEntry is one copy of a task in one worktree, for get's
	// Worktrees section.
	worktreeCopyEntry = worktree.CopyEntry
	// worktreeExclusion is one task the overlay keeps out of next's
	// recommendations, with the provenance that explains why.
	worktreeExclusion = worktree.Exclusion
)
