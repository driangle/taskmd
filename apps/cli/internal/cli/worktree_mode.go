package cli

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
)

func init() {
	// gitmeta stays free of flag/viper state; route its logging through the
	// CLI's --debug and --verbose channels.
	gitmeta.Debugf = debugLog
	gitmeta.Warnf = func(format string, args ...any) {
		if verbose || viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}
}

// Worktree scope values (the "worktree_scope" config key / --worktree-scope
// flag / TASKMD_WORKTREE_SCOPE env). unified merges task state across sibling
// worktrees; isolated reads only the current checkout's files.
const (
	worktreeScopeUnified  = "unified"
	worktreeScopeIsolated = "isolated"
)

// resolveWorktreeScope reads the worktree scope and reports whether the
// overlay is enabled. Viper gives the intended precedence for free once the
// --worktree-scope flag is bound: changed flag > TASKMD_WORKTREE_SCOPE env >
// config key > flag default (unified). Enabled does not mean active: the
// overlay still only forms in a multi-worktree git repo.
func resolveWorktreeScope() (bool, error) {
	v := viper.GetString("worktree_scope")
	// An explicitly passed --worktree-scope wins even when the viper binding
	// is absent (tests reset viper), mirroring resolveTaskDir.
	if f := rootCmd.PersistentFlags().Lookup("worktree-scope"); f != nil && f.Changed {
		v = worktreeScopeFlag
	}
	return worktreeScopeEnabled(v)
}

// discoverSiblingWorktrees lists the sibling worktrees of the repo containing
// scanDir (nil when scanDir is not in a repo). It is a seam so overlay tests
// can inject worktrees without git.
var discoverSiblingWorktrees worktree.Discoverer = worktree.DiscoverSiblings

// worktreeBuilder resolves the worktree scope into an overlay builder wired
// to the CLI's discovery seam. Commands hand it to the shared overlay code
// and to the MCP/web servers.
func worktreeBuilder(flags GlobalFlags) (worktree.Builder, error) {
	enabled, err := resolveWorktreeScope()
	if err != nil {
		return worktree.Builder{}, err
	}
	return worktree.Builder{
		Enabled:    enabled,
		Discover:   discoverSiblingWorktrees,
		Verbose:    flags.Verbose,
		IgnoreDirs: flags.IgnoreDirs,
	}, nil
}

// buildWorktreeOverlay builds the cross-worktree overlay for the local task
// list, or returns nil when the overlay is inactive: disabled, scanDir not
// in a git repo, or no sibling worktrees to merge (in which case behavior is
// identical to today's). Call it before makeFilePathsRelative — the merge
// stats task files by their absolute paths to break status ties by mtime.
func buildWorktreeOverlay(scanDir string, localTasks []*model.Task, flags GlobalFlags) (*worktreeOverlay, error) {
	builder, err := worktreeBuilder(flags)
	if err != nil {
		return nil, err
	}
	overlay, err := builder.Build(scanDir, localTasks)
	if err != nil || overlay == nil {
		return nil, err
	}
	overlay.RelativizeSiblingPaths()
	return overlay, nil
}

// printOverlayWarnings reports cross-worktree consistency warnings to stderr.
// Callers that surface them elsewhere (validate merges them into the
// validation result) skip this.
func printOverlayWarnings(overlay *worktreeOverlay, flags GlobalFlags) {
	if overlay == nil || flags.Quiet {
		return
	}
	for _, warning := range overlay.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning.Message)
	}
}

// siblingCopyGuard returns the guard error when taskID exists in a sibling
// worktree; callers invoke it after a local lookup misses, so mutations name
// where the task actually lives instead of reporting "not found". Writes are
// never redirected — the caller must fail.
func siblingCopyGuard(taskID, scanDir string, flags GlobalFlags) error {
	builder, err := worktreeBuilder(flags)
	if err != nil {
		return nil
	}
	return builder.SiblingGuard(taskID, scanDir)
}
