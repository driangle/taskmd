package web

import (
	"path/filepath"
	"sync"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// DataProvider caches scan results and invalidates on file changes. When a
// worktree overlay builder is configured, every rescan re-discovers sibling
// worktrees and rebuilds the merged view, so membership changes are picked up
// on any invalidation.
type DataProvider struct {
	scanDir string
	verbose bool
	wt      worktree.Builder

	mu      sync.RWMutex
	tasks   []*model.Task
	overlay *worktree.Overlay
	dirty   bool
}

// NewDataProvider creates a DataProvider with the worktree overlay disabled.
func NewDataProvider(scanDir string, verbose bool) *DataProvider {
	return NewDataProviderWithWorktrees(scanDir, verbose, worktree.Builder{})
}

// NewDataProviderWithWorktrees creates a DataProvider that builds the
// cross-worktree overlay per the given builder.
func NewDataProviderWithWorktrees(scanDir string, verbose bool, wt worktree.Builder) *DataProvider {
	return &DataProvider{
		scanDir: scanDir,
		verbose: verbose,
		wt:      wt,
		dirty:   true,
	}
}

// GetTasks returns the cached local tasks (after sibling-root attribution
// when the overlay is active), rescanning if dirty.
func (dp *DataProvider) GetTasks() ([]*model.Task, error) {
	tasks, _, err := dp.refresh()
	return tasks, err
}

// GetOverlay returns the cached worktree overlay, or nil when it is inactive.
func (dp *DataProvider) GetOverlay() (*worktree.Overlay, error) {
	_, overlay, err := dp.refresh()
	return overlay, err
}

// OverlayInfo reports the active overlay's shape for /api/config: the local
// worktree's name and the sibling worktree count. Nil when the overlay is
// inactive, so single-worktree responses are unchanged.
func (dp *DataProvider) OverlayInfo() (*WorktreeOverlayInfo, error) {
	if dp == nil {
		return nil, nil
	}
	overlay, err := dp.GetOverlay()
	if err != nil || overlay == nil {
		return nil, err
	}
	info := &WorktreeOverlayInfo{}
	if id, err := gitmeta.Resolve(dp.scanDir); err == nil && id != nil {
		info.Name = filepath.Base(id.WorktreeRoot)
	}
	if siblings, err := dp.wt.Siblings(dp.scanDir); err == nil {
		info.Siblings = len(siblings)
	}
	return info, nil
}

// GetEffectiveTasks returns the merged task list with effective statuses when
// the overlay is active, and the local tasks otherwise. Status-aggregating
// endpoints (board, graph, stats, next, tracks, validate, search) serve this.
func (dp *DataProvider) GetEffectiveTasks() ([]*model.Task, error) {
	tasks, overlay, err := dp.refresh()
	if err != nil {
		return nil, err
	}
	if overlay != nil {
		return overlay.EffectiveTasks(), nil
	}
	return tasks, nil
}

// refresh returns the cached tasks and overlay, rescanning when dirty.
func (dp *DataProvider) refresh() ([]*model.Task, *worktree.Overlay, error) {
	dp.mu.RLock()
	if !dp.dirty && dp.tasks != nil {
		defer dp.mu.RUnlock()
		return dp.tasks, dp.overlay, nil
	}
	dp.mu.RUnlock()

	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Double-check after acquiring write lock
	if !dp.dirty && dp.tasks != nil {
		return dp.tasks, dp.overlay, nil
	}

	s := scanner.NewScanner(dp.scanDir, dp.verbose, nil)
	result, err := s.Scan()
	if err != nil {
		return nil, nil, err
	}

	tasks := result.Tasks
	overlay, err := dp.wt.Build(dp.scanDir, tasks)
	if err != nil {
		return nil, nil, err
	}
	if overlay != nil {
		tasks = overlay.Local()
	}

	dp.tasks = tasks
	dp.overlay = overlay
	dp.dirty = false
	return dp.tasks, dp.overlay, nil
}

// GetArchivedTasks scans archive directories for tasks used in dependency resolution.
func (dp *DataProvider) GetArchivedTasks() ([]*model.Task, error) {
	s := scanner.NewScanner(dp.scanDir, dp.verbose, nil)
	return s.ScanArchive()
}

// ScanDir returns the directory being scanned.
func (dp *DataProvider) ScanDir() string {
	return dp.scanDir
}

// WatchDirs returns the directories the live-refresh watcher should cover for
// markdown changes: the scan dir plus, with the overlay enabled, each sibling
// worktree's tasks dir.
func (dp *DataProvider) WatchDirs() []string {
	dirs := []string{dp.scanDir}
	siblings, err := dp.wt.Siblings(dp.scanDir)
	if err != nil {
		return dirs
	}
	for _, wt := range siblings {
		dirs = append(dirs, wt.TasksDir)
	}
	return dirs
}

// WatchMetaDirs returns directories whose any change (not just markdown)
// signals a worktree membership change: the repo's <common-dir>/worktrees.
// Empty when the overlay is disabled or the scan dir is not in a git repo.
func (dp *DataProvider) WatchMetaDirs() []string {
	if !dp.wt.Enabled {
		return nil
	}
	id, err := gitmeta.Resolve(dp.scanDir)
	if err != nil || id == nil {
		return nil
	}
	return []string{filepath.Join(id.CommonDir, "worktrees")}
}

// Invalidate marks cached data as stale.
func (dp *DataProvider) Invalidate() {
	dp.mu.Lock()
	dp.dirty = true
	dp.mu.Unlock()
}
