package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/board"
	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/validator"
)

func createTestTaskDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	task1 := `---
id: "001"
title: "Task One"
status: pending
priority: high
effort: small
tags:
  - setup
---
# Task One
`
	task2 := `---
id: "002"
title: "Task Two"
status: in-progress
priority: medium
effort: medium
dependencies:
  - "001"
tags:
  - core
---
# Task Two
`
	os.WriteFile(filepath.Join(dir, "001-task-one.md"), []byte(task1), 0644)
	os.WriteFile(filepath.Join(dir, "002-task-two.md"), []byte(task2), 0644)
	return dir
}

func TestHandleTasks(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()

	handleTasks(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var tasks []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestHandleTaskByID_Success(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/001", nil)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleTaskByID(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var task map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if task["id"] != "001" {
		t.Fatalf("expected task ID 001, got %v", task["id"])
	}

	if task["title"] != "Task One" {
		t.Fatalf("expected title 'Task One', got %v", task["title"])
	}

	// Verify body is included
	body, ok := task["body"]
	if !ok {
		t.Fatal("expected body field in response")
	}
	if body == "" || body == nil {
		t.Fatal("expected non-empty body")
	}
}

func TestHandleTaskByID_NotFound(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleTaskByID(dp)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleBoard(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/board?groupBy=status", nil)
	rec := httptest.NewRecorder()

	handleBoard(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var groups []board.JSONGroup
	if err := json.Unmarshal(rec.Body.Bytes(), &groups); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(groups) == 0 {
		t.Fatal("expected at least one group")
	}
}

func TestHandleBoardDefaultGroupBy(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/board", nil)
	rec := httptest.NewRecorder()

	handleBoard(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleBoardInvalidGroupBy(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/board?groupBy=invalid", nil)
	rec := httptest.NewRecorder()

	handleBoard(dp)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleGraph(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()

	handleGraph(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := result["nodes"]; !ok {
		t.Fatal("expected 'nodes' in graph response")
	}
	if _, ok := result["edges"]; !ok {
		t.Fatal("expected 'edges' in graph response")
	}
}

func TestHandleGraphMermaid(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/graph/mermaid", nil)
	rec := httptest.NewRecorder()

	handleGraphMermaid(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty mermaid output")
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain" {
		t.Fatalf("expected text/plain, got %s", ct)
	}
}

func TestHandleStats(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()

	handleStats(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var m metrics.Metrics
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if m.TotalTasks != 2 {
		t.Fatalf("expected 2 total tasks, got %d", m.TotalTasks)
	}
}

func TestHandleValidate(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/validate", nil)
	rec := httptest.NewRecorder()

	handleValidate(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result validator.ValidationResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}
