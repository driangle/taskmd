package lock

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestTaskLockPath(t *testing.T) {
	got := TaskLockPath("/repo/tasks", "042")
	want := filepath.Join("/repo/tasks", ".taskmd", "locks", "042.lock")
	if got != want {
		t.Fatalf("TaskLockPath = %q, want %q", got, want)
	}
}

func TestAcquireRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "locks", "x.lock")

	l, err := Acquire(path, DefaultTimeout)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("lock file not created: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	// Double release is a no-op.
	if err := l.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}

	// Lock is reusable after release.
	l2, err := Acquire(path, DefaultTimeout)
	if err != nil {
		t.Fatalf("re-Acquire: %v", err)
	}
	_ = l2.Release()
}

func TestAcquireTimesOutWhenHeld(t *testing.T) {
	path := filepath.Join(t.TempDir(), "held.lock")

	l, err := Acquire(path, DefaultTimeout)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	defer l.Release()

	start := time.Now()
	if _, err := Acquire(path, 50*time.Millisecond); err == nil {
		t.Fatal("expected timeout error acquiring a held lock, got nil")
	}
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Fatalf("returned too quickly (%v); did not wait for timeout", elapsed)
	}
}

// TestMutualExclusion runs many goroutines that each acquire the same lock,
// read a shared counter file, increment it, and write it back. With correct
// locking the final value equals the number of goroutines (no lost updates).
func TestMutualExclusion(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "counter.lock")
	counterPath := filepath.Join(dir, "counter")
	if err := os.WriteFile(counterPath, []byte("0"), 0o644); err != nil {
		t.Fatalf("seed counter: %v", err)
	}

	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			l, err := Acquire(lockPath, 5*time.Second)
			if err != nil {
				t.Errorf("Acquire: %v", err)
				return
			}
			defer l.Release()
			incrementCounter(t, counterPath)
		}()
	}
	wg.Wait()

	if got := readCounter(t, counterPath); got != workers {
		t.Fatalf("counter = %d, want %d (lost updates indicate broken locking)", got, workers)
	}
}

func incrementCounter(t *testing.T, path string) {
	t.Helper()
	n := readCounter(t, path)
	// Widen the race window so a missing lock would reliably corrupt the count.
	time.Sleep(time.Millisecond)
	if err := os.WriteFile(path, []byte(strconv.Itoa(n+1)), 0o644); err != nil {
		t.Errorf("write counter: %v", err)
	}
}

func readCounter(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read counter: %v", err)
		return 0
	}
	n, err := strconv.Atoi(string(data))
	if err != nil {
		t.Errorf("parse counter %q: %v", data, err)
		return 0
	}
	return n
}
