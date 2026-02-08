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

	var timer *time.Timer
	for {
		select {
		case event, ok := <-fsw.Events:
			if !ok {
				return nil
			}
			if !isMarkdown(event.Name) {
				// Watch new directories for recursive support
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						_ = w.addRecursive(fsw, event.Name)
					}
				}
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			// Debounce
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, w.onChange)

		case err, ok := <-fsw.Errors:
			if !ok {
				return nil
			}
			_ = err // log in verbose mode if needed

		case <-w.done:
			return nil
		}
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
