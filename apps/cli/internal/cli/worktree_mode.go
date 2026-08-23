package cli

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
)

// Worktree overlay activation modes (the "worktrees" config key / --worktrees
// flag / TASKMD_WORKTREES env). auto activates the overlay only when the scan
// dir is inside a git repo with sibling worktrees; false disables it entirely.
const (
	worktreeModeAuto  = worktree.ModeAuto
	worktreeModeTrue  = worktree.ModeTrue
	worktreeModeFalse = worktree.ModeFalse
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

// resolveWorktreeMode reads the overlay activation mode. Viper gives the
// intended precedence for free once the --worktrees flag is bound: changed
// flag > TASKMD_WORKTREES env > config key > flag default (auto).
func resolveWorktreeMode() (string, error) {
	v := viper.GetString("worktrees")
	// An explicitly passed --worktrees wins even when the viper binding is
	// absent (tests reset viper), mirroring resolveTaskDir.
	if f := rootCmd.PersistentFlags().Lookup("worktrees"); f != nil && f.Changed {
		v = worktreesFlag
	}
	switch v {
	case "", worktreeModeAuto:
		return worktreeModeAuto, nil
	case worktreeModeTrue:
		return worktreeModeTrue, nil
	case worktreeModeFalse:
		return worktreeModeFalse, nil
	default:
		return "", invalidValueError("worktrees", v, worktree.ValidModes)
	}
}

// discoverSiblingWorktrees lists the sibling worktrees of the repo containing
// scanDir (nil when scanDir is not in a repo). It is a seam so overlay tests
// can inject worktrees without git.
var discoverSiblingWorktrees worktree.Discoverer = worktree.DiscoverSiblings

// worktreeBuilder resolves the activation mode into an overlay builder wired
// to the CLI's discovery seam. Commands hand it to the shared overlay code
// and to the MCP/web servers.
func worktreeBuilder(flags GlobalFlags) (worktree.Builder, error) {
	mode, err := resolveWorktreeMode()
	if err != nil {
		return worktree.Builder{}, err
	}
	return worktree.Builder{
		Mode:       mode,
		Discover:   discoverSiblingWorktrees,
		Verbose:    flags.Verbose,
		IgnoreDirs: flags.IgnoreDirs,
	}, nil
}

// buildWorktreeOverlay builds the cross-worktree overlay for the local task
// list, or returns nil when the overlay is inactive: mode false, scanDir not
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
