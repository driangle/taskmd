package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driangle/taskmd/sdk/go/effort"
)

// putStatus sends PUT /api/tasks/{id} with a status change and returns the
// task file's contents afterwards.
func putStatus(t *testing.T, dir, id, file, status string) string {
	t.Helper()
	dp := NewDataProvider(dir, false)

	req := httptest.NewRequest(http.MethodPut, "/api/tasks/"+id, strings.NewReader(`{"status":"`+status+`"}`))
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()

	handleUpdateTask(dp, false, effort.Default())(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	content, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		t.Fatalf("read task file: %v", err)
	}
	return string(content)
}

func TestHandleUpdateTask_CompletedSetsCompletedAt(t *testing.T) {
	dir := createTestTaskDir(t)
	today := time.Now().Format("2006-01-02")

	s := putStatus(t, dir, "001", "001-task-one.md", "completed")

	if !strings.Contains(s, "completed_at: "+today) && !strings.Contains(s, `completed_at: "`+today+`"`) {
		t.Errorf("expected completed_at %s in file, got:\n%s", today, s)
	}
}

func TestHandleUpdateTask_CancelledSetsCancelledAt(t *testing.T) {
	dir := createTestTaskDir(t)
	today := time.Now().Format("2006-01-02")

	s := putStatus(t, dir, "001", "001-task-one.md", "cancelled")

	if !strings.Contains(s, "cancelled_at: "+today) && !strings.Contains(s, `cancelled_at: "`+today+`"`) {
		t.Errorf("expected cancelled_at %s in file, got:\n%s", today, s)
	}
	if strings.Contains(s, "completed_at") {
		t.Errorf("expected no completed_at in file, got:\n%s", s)
	}
}

func TestHandleUpdateTask_ReopenClearsTerminalDates(t *testing.T) {
	dir := createTestTaskDir(t)

	putStatus(t, dir, "001", "001-task-one.md", "completed")
	s := putStatus(t, dir, "001", "001-task-one.md", "pending")

	if strings.Contains(s, "completed_at") || strings.Contains(s, "cancelled_at") {
		t.Errorf("expected terminal dates cleared on reopen, got:\n%s", s)
	}
}
