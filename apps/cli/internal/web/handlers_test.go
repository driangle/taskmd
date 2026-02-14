package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/board"
	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/next"
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

	// Verify tags are included (test fixtures have "setup" and "core")
	if len(m.TagsByCount) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(m.TagsByCount))
	}
	// Both have count 1, so alphabetical: core first, setup second
	if m.TagsByCount[0].Tag != "core" {
		t.Errorf("expected first tag 'core', got %q", m.TagsByCount[0].Tag)
	}
	if m.TagsByCount[1].Tag != "setup" {
		t.Errorf("expected second tag 'setup', got %q", m.TagsByCount[1].Tag)
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

// PUT /api/tasks/{id} tests

func TestHandleUpdateTask_Success(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"status":"completed"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var task map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if task["status"] != "completed" {
		t.Errorf("expected status completed, got %v", task["status"])
	}

	// Verify file was actually updated
	content, _ := os.ReadFile(filepath.Join(dir, "001-task-one.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("expected file to contain updated status")
	}
}

func TestHandleUpdateTask_NotFound(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"status":"completed"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/999", body)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleUpdateTask_InvalidStatus(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"status":"invalid"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if errResp.Error != "validation failed" {
		t.Errorf("expected 'validation failed', got %q", errResp.Error)
	}
}

func TestHandleUpdateTask_InvalidJSON(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleUpdateTask_Title(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"title":"New Title"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	content, _ := os.ReadFile(filepath.Join(dir, "001-task-one.md"))
	if !strings.Contains(string(content), `title: "New Title"`) {
		t.Errorf("expected title update in file, got:\n%s", string(content))
	}
}

func TestHandleUpdateTask_Body(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"body":"# Updated\n\nNew body content."}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	content, _ := os.ReadFile(filepath.Join(dir, "001-task-one.md"))
	s := string(content)
	if !strings.Contains(s, "New body content.") {
		t.Error("expected new body content in file")
	}
	if strings.Contains(s, "# Task One") {
		t.Error("expected old body to be replaced")
	}
}

func TestHandleUpdateTask_Tags(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"tags":["new-a","new-b"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var task map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	tags, ok := task["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags array, got %T", task["tags"])
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
}

func TestHandleUpdateTask_PartialUpdate(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	// Only update priority, everything else should be preserved
	body := strings.NewReader(`{"priority":"low"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	content, _ := os.ReadFile(filepath.Join(dir, "001-task-one.md"))
	s := string(content)
	if !strings.Contains(s, "priority: low") {
		t.Error("expected priority to be updated")
	}
	if !strings.Contains(s, "status: pending") {
		t.Error("expected status to be preserved")
	}
	if !strings.Contains(s, "effort: small") {
		t.Error("expected effort to be preserved")
	}
}

// GET /api/config tests

func TestHandleConfig(t *testing.T) {
	cfg := Config{ReadOnly: false, Version: "1.2.3-abc1234"}

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	handleConfig(cfg)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.ReadOnly {
		t.Error("expected readonly to be false")
	}
	if resp.Version != "1.2.3-abc1234" {
		t.Errorf("expected version '1.2.3-abc1234', got %q", resp.Version)
	}
}

func TestHandleConfig_ReadOnly(t *testing.T) {
	cfg := Config{ReadOnly: true}

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	handleConfig(cfg)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp ConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !resp.ReadOnly {
		t.Error("expected readonly to be true")
	}
}

func TestHandleUpdateTask_ReadOnly(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	body := strings.NewReader(`{"status":"completed"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/001", body)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, true)(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if errResp.Error != "server is in read-only mode" {
		t.Errorf("expected 'server is in read-only mode', got %q", errResp.Error)
	}
}

// GET /api/next tests

func TestHandleNext(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/next", nil)
	rec := httptest.NewRecorder()

	handleNext(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &recs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Task 001 is pending (actionable), 002 depends on 001 (blocked)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}

	if recs[0].ID != "001" {
		t.Errorf("expected recommendation for task 001, got %s", recs[0].ID)
	}

	if recs[0].Score <= 0 {
		t.Errorf("expected positive score, got %d", recs[0].Score)
	}

	if recs[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", recs[0].Rank)
	}
}

func TestHandleNext_WithLimit(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/next?limit=1", nil)
	rec := httptest.NewRecorder()

	handleNext(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &recs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(recs) > 1 {
		t.Fatalf("expected at most 1 recommendation with limit=1, got %d", len(recs))
	}
}

func TestHandleNext_EmptyResult(t *testing.T) {
	dir := t.TempDir()

	// Create only completed tasks
	task := `---
id: "001"
title: "Done"
status: completed
priority: high
---
`
	os.WriteFile(filepath.Join(dir, "001.md"), []byte(task), 0644)

	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/next", nil)
	rec := httptest.NewRecorder()

	handleNext(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &recs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(recs) != 0 {
		t.Fatalf("expected 0 recommendations for all-completed tasks, got %d", len(recs))
	}
}

func TestHandleNext_DefaultLimit(t *testing.T) {
	dir := t.TempDir()

	// Create 7 pending tasks
	for i := 1; i <= 7; i++ {
		task := fmt.Sprintf(`---
id: "%03d"
title: "Task %d"
status: pending
priority: medium
---
`, i, i)
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("%03d.md", i)), []byte(task), 0644)
	}

	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodGet, "/api/next", nil)
	rec := httptest.NewRecorder()

	handleNext(dp)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &recs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Default limit is 5
	if len(recs) != 5 {
		t.Fatalf("expected 5 recommendations (default limit), got %d", len(recs))
	}
}
