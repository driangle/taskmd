package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/sdk/go/model"
)

// Worktree overlay activation modes (the "worktrees" config key / --worktrees
// flag / TASKMD_WORKTREES env). auto activates the overlay only when the scan
// dir is inside a git repo with sibling worktrees; false disables it entirely.
const (
	worktreeModeAuto  = "auto"
	worktreeModeTrue  = "true"
	worktreeModeFalse = "false"
)

var validWorktreeModes = []string{worktreeModeAuto, worktreeModeTrue, worktreeModeFalse}

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
		return "", invalidValueError("worktrees", v, validWorktreeModes)
	}
}

// discoverSiblingWorktrees lists the sibling worktrees of the repo containing
// scanDir (nil when scanDir is not in a repo). It is a seam so overlay tests
// can inject worktrees without git.
var discoverSiblingWorktrees = defaultDiscoverSiblings

func defaultDiscoverSiblings(scanDir string) ([]gitmeta.Worktree, error) {
	id, err := gitmeta.Resolve(scanDir)
	if err != nil || id == nil {
		return nil, err
	}
	worktrees, err := gitmeta.ListWorktrees(id)
	if err != nil {
		return nil, err
	}
	var siblings []gitmeta.Worktree
	for _, wt := range worktrees {
		if !wt.IsLocal {
			siblings = append(siblings, wt)
		}
	}
	return siblings, nil
}

// buildWorktreeOverlay builds the cross-worktree overlay for the local task
// list, or returns nil when the overlay is inactive: mode false, scanDir not
// in a git repo, or no sibling worktrees to merge (in which case behavior is
// identical to today's). Call it before makeFilePathsRelative — the merge
// stats task files by their absolute paths to break status ties by mtime.
func buildWorktreeOverlay(scanDir string, localTasks []*model.Task, flags GlobalFlags) (*worktreeOverlay, error) {
	mode, err := resolveWorktreeMode()
	if err != nil {
		return nil, err
	}
	if mode == worktreeModeFalse {
		return nil, nil
	}

	siblings, err := discoverSiblingWorktrees(scanDir)
	if err != nil {
		debugLog("worktree discovery failed, overlay inactive: %v", err)
		return nil, nil
	}
	if len(siblings) == 0 {
		return nil, nil
	}

	scanned := scanSiblingWorktrees(siblings, flags)
	overlay := mergeOverlay(localTasks, scanned)
	for _, st := range scanned {
		makeFilePathsRelative(st.tasks, st.wt.TasksDir)
	}

	if !flags.Quiet {
		for _, warning := range overlay.Warnings {
			fmt.Fprintln(os.Stderr, warning)
		}
	}
	return overlay, nil
}

// scanSiblingWorktrees scans each sibling's tasks dir, skipping siblings that
// fail to scan. Sibling duplicate-ID warnings are that worktree's own concern
// and are not reported here.
func scanSiblingWorktrees(siblings []gitmeta.Worktree, flags GlobalFlags) []siblingTasks {
	var scanned []siblingTasks
	for _, wt := range siblings {
		result, err := newTaskScanner(wt.TasksDir, flags).Scan()
		if err != nil {
			if flags.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: skipping worktree %s: %v\n", wt.Root, err)
			}
			continue
		}
		scanned = append(scanned, siblingTasks{wt: wt, tasks: result.Tasks})
	}
	return scanned
}

// siblingCopyGuard returns the guard error when taskID exists in a sibling
// worktree; callers invoke it after a local lookup misses, so mutations name
// where the task actually lives instead of reporting "not found". Writes are
// never redirected — the caller must fail.
func siblingCopyGuard(taskID, scanDir string, flags GlobalFlags) error {
	mode, err := resolveWorktreeMode()
	if err != nil || mode == worktreeModeFalse {
		return nil
	}
	siblings, discoverErr := discoverSiblingWorktrees(scanDir)
	if discoverErr != nil {
		return nil
	}

	for _, st := range scanSiblingWorktrees(siblings, flags) {
		for _, t := range st.tasks {
			if t.ID == taskID {
				branch := st.wt.Branch
				if branch == "" {
					branch = "detached"
				}
				return fmt.Errorf("task %s exists only in worktree %s (branch %s); run taskmd there",
					taskID, displayWorktreePath(st.wt.Root), branch)
			}
		}
	}
	return nil
}

// displayWorktreePath renders a worktree root relative to the current
// directory when possible (matching how users navigate between worktrees),
// falling back to the absolute path.
func displayWorktreePath(root string) string {
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
