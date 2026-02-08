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
