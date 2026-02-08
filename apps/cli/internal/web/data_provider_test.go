package web

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataProvider_GetTasks(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	tasks, err := dp.GetTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestDataProvider_Cache(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	// First call should scan
	tasks1, err := dp.GetTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should return cached result
	tasks2, err := dp.GetTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks1) != len(tasks2) {
		t.Fatalf("cache returned different result")
	}
}

func TestDataProvider_Invalidate(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	// Initial scan
	tasks1, err := dp.GetTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks1) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks1))
	}

	// Add a new task file
	task3 := `---
id: "003"
title: "Task Three"
status: pending
---
# Task Three
`
	os.WriteFile(filepath.Join(dir, "003-task-three.md"), []byte(task3), 0644)

	// Without invalidation, should still return cached 2 tasks
	tasks2, err := dp.GetTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks2) != 2 {
		t.Fatalf("expected 2 cached tasks, got %d", len(tasks2))
	}

	// After invalidation, should re-scan and find 3 tasks
	dp.Invalidate()
	tasks3, err := dp.GetTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks3) != 3 {
		t.Fatalf("expected 3 tasks after invalidation, got %d", len(tasks3))
	}
}
