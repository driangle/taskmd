package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a directory for .md file changes and calls onChange.
type Watcher struct {
	dir      string
	onChange func()
	debounce time.Duration
	done     chan struct{}
	watcher  *fsnotify.Watcher
	mu       sync.Mutex
}

// New creates a Watcher that monitors dir for markdown file changes.
func New(dir string, onChange func(), debounce time.Duration) *Watcher {
	return &Watcher{
		dir:      dir,
		onChange: onChange,
		debounce: debounce,
		done:     make(chan struct{}),
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
	w.mu.Unlock()

	defer fsw.Close()

	if err := w.addRecursive(fsw, w.dir); err != nil {
		return err
	}

	return w.eventLoop(fsw)
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
	if !isMarkdown(event.Name) {
		w.watchNewDir(fsw, event)
		return timer
	}
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return timer
	}

	// Debounce
	if timer != nil {
		timer.Stop()
	}
	return time.AfterFunc(w.debounce, w.onChange)
}

// watchNewDir adds newly created directories to the watcher so nested markdown
// files are picked up recursively.
func (w *Watcher) watchNewDir(fsw *fsnotify.Watcher, event fsnotify.Event) {
	if event.Op&fsnotify.Create == 0 {
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
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
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
