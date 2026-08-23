package watcher

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcher_DetectsChanges(t *testing.T) {
	dir := t.TempDir()

	// Create initial file
	initial := filepath.Join(dir, "task.md")
	os.WriteFile(initial, []byte("# initial"), 0644)

	var count atomic.Int32
	w := New(dir, func() {
		count.Add(1)
	}, 50*time.Millisecond)

	go func() {
		if err := w.Start(); err != nil {
			t.Logf("watcher error: %v", err)
		}
	}()

	// Wait for watcher to initialize
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	os.WriteFile(initial, []byte("# modified"), 0644)

	// Wait for debounce
	time.Sleep(200 * time.Millisecond)

	w.Stop()

	if count.Load() < 1 {
		t.Fatal("expected onChange to be called at least once")
	}
}

func TestWatcher_IgnoresNonMarkdown(t *testing.T) {
	dir := t.TempDir()

	var count atomic.Int32
	w := New(dir, func() {
		count.Add(1)
	}, 50*time.Millisecond)

	go func() {
		if err := w.Start(); err != nil {
			t.Logf("watcher error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Create a non-markdown file
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not markdown"), 0644)

	time.Sleep(200 * time.Millisecond)

	w.Stop()

	if count.Load() != 0 {
		t.Fatalf("expected no onChange calls for non-md files, got %d", count.Load())
	}
}

func TestWatcher_Debounces(t *testing.T) {
	dir := t.TempDir()

	var count atomic.Int32
	w := New(dir, func() {
		count.Add(1)
	}, 100*time.Millisecond)

	go func() {
		if err := w.Start(); err != nil {
			t.Logf("watcher error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Rapid writes - should debounce to one call
	f := filepath.Join(dir, "rapid.md")
	for i := 0; i < 5; i++ {
		os.WriteFile(f, []byte("# version "+string(rune('0'+i))), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to settle
	time.Sleep(300 * time.Millisecond)

	w.Stop()

	// Should have debounced to 1-2 calls, not 5
	if count.Load() > 2 {
		t.Fatalf("expected debouncing (1-2 calls), got %d", count.Load())
	}
	if count.Load() < 1 {
		t.Fatal("expected at least 1 onChange call")
	}
}

func TestWatcher_Stop(t *testing.T) {
	dir := t.TempDir()

	w := New(dir, func() {}, 50*time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- w.Start()
	}()

	time.Sleep(50 * time.Millisecond)
	w.Stop()

	select {
	case <-done:
		// OK - stopped successfully
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop within timeout")
	}
}

func TestWatcher_SetDirs_WatchesAdditionalRoots(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	var count atomic.Int32
	w := New(dirA, func() { count.Add(1) }, 30*time.Millisecond)
	w.SetDirs([]string{dirA, dirB}, nil)

	go func() { _ = w.Start() }()
	defer w.Stop()
	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(filepath.Join(dirB, "task.md"), []byte("# b"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, &count, 1, "markdown change in a SetDirs-added root")
}

func TestWatcher_SetDirs_WhileRunningAddsRoot(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	var count atomic.Int32
	w := New(dirA, func() { count.Add(1) }, 30*time.Millisecond)

	go func() { _ = w.Start() }()
	defer w.Stop()
	time.Sleep(100 * time.Millisecond)

	w.SetDirs([]string{dirA, dirB}, nil)
	time.Sleep(50 * time.Millisecond)

	if err := os.WriteFile(filepath.Join(dirB, "task.md"), []byte("# b"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, &count, 1, "markdown change in a root added while running")
}

func TestWatcher_MetaDir_AnyEventTriggers(t *testing.T) {
	dir := t.TempDir()
	meta := filepath.Join(t.TempDir(), "worktrees")
	if err := os.MkdirAll(meta, 0o755); err != nil {
		t.Fatal(err)
	}

	var count atomic.Int32
	w := New(dir, func() { count.Add(1) }, 30*time.Millisecond)
	w.SetDirs([]string{dir}, []string{meta})

	go func() { _ = w.Start() }()
	defer w.Stop()
	time.Sleep(100 * time.Millisecond)

	// A non-markdown change inside the meta dir must trigger onChange.
	if err := os.Mkdir(filepath.Join(meta, "agent-b"), 0o755); err != nil {
		t.Fatal(err)
	}
	waitFor(t, &count, 1, "directory creation in a meta dir")
}

func TestWatcher_MetaDir_CreatedLaterIsObserved(t *testing.T) {
	dir := t.TempDir()
	parent := t.TempDir()
	meta := filepath.Join(parent, "worktrees")

	var count atomic.Int32
	w := New(dir, func() { count.Add(1) }, 30*time.Millisecond)
	w.SetDirs([]string{dir}, []string{meta}) // meta does not exist yet

	go func() { _ = w.Start() }()
	defer w.Stop()
	time.Sleep(100 * time.Millisecond)

	// Creating the meta dir itself (git's first worktree add) must trigger.
	if err := os.Mkdir(meta, 0o755); err != nil {
		t.Fatal(err)
	}
	waitFor(t, &count, 1, "creation of a previously missing meta dir")
}

// waitFor polls until count reaches at least want or times out.
func waitFor(t *testing.T, count *atomic.Int32, want int32, what string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if count.Load() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
