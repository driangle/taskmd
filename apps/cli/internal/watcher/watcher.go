package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a set of directories for .md file changes and calls
// onChange. Meta directories additionally trigger onChange on any event —
// they cover state whose shape matters more than its content, like the git
// common dir's worktrees/ registry.
type Watcher struct {
	onChange func()
	debounce time.Duration
	done     chan struct{}

	mu       sync.Mutex
	dirs     []string
	metaDirs []string
	watcher  *fsnotify.Watcher
}

// New creates a Watcher that monitors dir for markdown file changes.
func New(dir string, onChange func(), debounce time.Duration) *Watcher {
	return &Watcher{
		dirs:     []string{dir},
		onChange: onChange,
		debounce: debounce,
		done:     make(chan struct{}),
	}
}

// SetDirs replaces the watched directory sets. Stale watches are removed and
// new roots added; safe to call before Start (Start picks the sets up) or
// from the onChange callback while running.
func (w *Watcher) SetDirs(dirs, metaDirs []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dirs = dirs
	w.metaDirs = metaDirs
	if w.watcher != nil {
		w.syncWatches()
	}
}

// Start begins watching. It blocks until Stop is called or an error occurs.
func (w *Watcher) Start() error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.mu.Lock()
	w.watcher = fsw
	w.syncWatches()
	w.mu.Unlock()

	defer fsw.Close()

	return w.eventLoop(fsw)
}

// syncWatches reconciles fsnotify's watch list with the configured roots.
// Callers must hold w.mu. A meta dir that does not exist yet is covered by
// watching its parent, so its creation is still observed.
func (w *Watcher) syncWatches() {
	roots := make([]string, 0, len(w.dirs)+2*len(w.metaDirs))
	roots = append(roots, w.dirs...)
	for _, m := range w.metaDirs {
		roots = append(roots, m, filepath.Dir(m))
	}

	for _, watched := range w.watcher.WatchList() {
		if !underAny(watched, roots) {
			_ = w.watcher.Remove(watched)
		}
	}

	for _, dir := range w.dirs {
		_ = w.addRecursive(w.watcher, dir)
	}
	for _, m := range w.metaDirs {
		if _, err := os.Stat(m); err == nil {
			_ = w.addRecursive(w.watcher, m)
		} else {
			// Watch the parent (non-recursively) so the dir's creation fires.
			_ = w.watcher.Add(filepath.Dir(m))
		}
	}
}

// eventLoop dispatches filesystem events until the watcher is stopped or a
// channel closes.
func (w *Watcher) eventLoop(fsw *fsnotify.Watcher) error {
	var timer *time.Timer
	for {
		select {
		case event, ok := <-fsw.Events:
			if !ok {
				return nil
			}
			timer = w.handleEvent(fsw, event, timer)

		case _, ok := <-fsw.Errors:
			if !ok {
				return nil
			}
			// errors are ignored; log in verbose mode if needed

		case <-w.done:
			return nil
		}
	}
}

// handleEvent processes a single event, returning the updated debounce timer.
func (w *Watcher) handleEvent(fsw *fsnotify.Watcher, event fsnotify.Event, timer *time.Timer) *time.Timer {
	if w.isMetaEvent(event.Name) {
		return w.resetTimer(timer)
	}
	if !isMarkdown(event.Name) {
		w.watchNewDir(fsw, event)
		return timer
	}
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return timer
	}
	return w.resetTimer(timer)
}

// resetTimer debounces onChange.
func (w *Watcher) resetTimer(timer *time.Timer) *time.Timer {
	if timer != nil {
		timer.Stop()
	}
	return time.AfterFunc(w.debounce, w.onChange)
}

// isMetaEvent reports whether path is a configured meta dir or inside one.
func (w *Watcher) isMetaEvent(path string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	cleaned := filepath.Clean(path)
	for _, m := range w.metaDirs {
		m = filepath.Clean(m)
		if cleaned == m || strings.HasPrefix(cleaned, m+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// underAny reports whether path equals or lies under any of the roots.
func underAny(path string, roots []string) bool {
	cleaned := filepath.Clean(path)
	for _, root := range roots {
		root = filepath.Clean(root)
		if cleaned == root || strings.HasPrefix(cleaned, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// watchNewDir adds newly created directories under a watched markdown root to
// the watcher so nested markdown files are picked up recursively. Creations
// elsewhere (e.g. inside a meta dir's watched parent) are ignored.
func (w *Watcher) watchNewDir(fsw *fsnotify.Watcher, event fsnotify.Event) {
	if event.Op&fsnotify.Create == 0 {
		return
	}
	w.mu.Lock()
	inScope := underAny(event.Name, w.dirs)
	w.mu.Unlock()
	if !inScope {
		return
	}
	if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
		_ = w.addRecursive(fsw, event.Name)
	}
}

// Stop signals the watcher to stop.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	select {
	case <-w.done:
		// already closed
	default:
		close(w.done)
	}
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *Watcher) addRecursive(fsw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor") {
				return filepath.SkipDir
			}
			return fsw.Add(path)
		}
		return nil
	})
}

func isMarkdown(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".md")
}
