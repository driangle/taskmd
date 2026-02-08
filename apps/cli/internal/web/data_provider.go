package web

import (
	"sync"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
)

// DataProvider caches scan results and invalidates on file changes.
type DataProvider struct {
	scanDir string
	verbose bool

	mu    sync.RWMutex
	tasks []*model.Task
	dirty bool
}

// NewDataProvider creates a DataProvider for the given directory.
func NewDataProvider(scanDir string, verbose bool) *DataProvider {
	return &DataProvider{
		scanDir: scanDir,
		verbose: verbose,
		dirty:   true,
	}
}

// GetTasks returns cached tasks, rescanning if dirty.
func (dp *DataProvider) GetTasks() ([]*model.Task, error) {
	dp.mu.RLock()
	if !dp.dirty && dp.tasks != nil {
		defer dp.mu.RUnlock()
		return dp.tasks, nil
	}
	dp.mu.RUnlock()

	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Double-check after acquiring write lock
	if !dp.dirty && dp.tasks != nil {
		return dp.tasks, nil
	}

	s := scanner.NewScanner(dp.scanDir, dp.verbose)
	result, err := s.Scan()
	if err != nil {
		return nil, err
	}

	dp.tasks = result.Tasks
	dp.dirty = false
	return dp.tasks, nil
}

// Invalidate marks cached data as stale.
func (dp *DataProvider) Invalidate() {
	dp.mu.Lock()
	dp.dirty = true
	dp.mu.Unlock()
}
