package taskfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestWriteFileAtomic_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.md")

	if err := writeFileAtomic(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("content = %q, want %q", got, "hello")
	}

	// Overwrite replaces content.
	if err := writeFileAtomic(path, []byte("world"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic overwrite: %v", err)
	}
	got, _ = os.ReadFile(path)
	if string(got) != "world" {
		t.Fatalf("content after overwrite = %q, want %q", got, "world")
	}
}

func TestWriteFileAtomic_NoTempLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.md")
	if err := writeFileAtomic(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Fatalf("leftover temp file: %s", e.Name())
		}
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 file, got %d", len(entries))
	}
}

// TestUpdateTaskFileLocked_ConcurrentAddTags verifies that many concurrent
// writers each adding a distinct tag all succeed with no lost updates — the
// core guarantee of the per-task lock plus atomic write. Without locking, the
// read-modify-write cycles would interleave and drop tags.
func TestUpdateTaskFileLocked_ConcurrentAddTags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "010-task.md")
	seed := `---
id: "010"
title: "Concurrent task"
status: pending
priority: medium
effort: small
tags: []
created: 2026-02-08
---

# Concurrent task
`
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	const workers = 30
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			tag := fmt.Sprintf("t%02d", n)
			req := UpdateRequest{AddTags: []string{tag}}
			if err := UpdateTaskFileLocked(dir, path, "010", req); err != nil {
				t.Errorf("UpdateTaskFileLocked: %v", err)
			}
		}(i)
	}
	wg.Wait()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	for i := 0; i < workers; i++ {
		tag := fmt.Sprintf("t%02d", i)
		if !strings.Contains(string(content), tag) {
			t.Errorf("tag %q missing — lost update under concurrency", tag)
		}
	}
}
