package cli

import (
	"fmt"

	"github.com/driangle/taskmd/sdk/go/model"
)

// scanTasksWithOverlay scans scanDir and builds the worktree overlay when it
// is active (nil otherwise). Unlike scanTasks, duplicate-ID warnings fire
// after copies nested in a sibling worktree's root are attributed to that
// worktree (spec §8), so a non-hidden checkout inside the scan root does not
// read as duplicates. The returned tasks are the local tasks after that
// attribution.
func scanTasksWithOverlay(scanDir string, flags GlobalFlags) ([]*model.Task, *worktreeOverlay, error) {
	result, err := newTaskScanner(scanDir, flags).Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}
	reportScanErrors(result.Errors, flags.Verbose)

	overlay, err := buildWorktreeOverlay(scanDir, result.Tasks, flags)
	if err != nil {
		return nil, nil, err
	}

	tasks := result.Tasks
	if overlay != nil {
		tasks = overlay.local
	}
	warnDuplicateIDs(tasks)
	printOverlayWarnings(overlay, flags)
	return tasks, overlay, nil
}

// scanTasksEffective returns the task list that status-aggregating views
// (board, stats, graph, report, phases) render: with the overlay active,
// statuses are effective across worktrees and sibling-only tasks are
// included; otherwise it is exactly the local scan.
func scanTasksEffective(scanDir string, flags GlobalFlags) ([]*model.Task, error) {
	tasks, overlay, err := scanTasksWithOverlay(scanDir, flags)
	if err != nil {
		return nil, err
	}
	if overlay != nil {
		return overlay.effectiveTasks(), nil
	}
	return tasks, nil
}

// scanActiveAndArchivedWithOverlay is scanTasksWithOverlay plus the archive
// scan, for commands that need both sets (next, tracks). The overlay covers
// active tasks only; the archive stays local.
func scanActiveAndArchivedWithOverlay(scanDir string, flags GlobalFlags) (active, archived []*model.Task, overlay *worktreeOverlay, err error) {
	active, overlay, err = scanTasksWithOverlay(scanDir, flags)
	if err != nil {
		return nil, nil, nil, err
	}

	archived, err = newTaskScanner(scanDir, flags).ScanArchive()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("archive scan failed: %w", err)
	}
	return active, archived, overlay, nil
}
