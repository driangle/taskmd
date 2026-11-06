package taskfile

import (
	"reflect"
	"testing"
	"time"

	"github.com/driangle/taskmd/sdk/go/model"
)

func TestApplyStatusTransitionDates(t *testing.T) {
	now := time.Date(2026, 10, 23, 15, 4, 5, 0, time.UTC)
	const today = "2026-10-23"

	tests := []struct {
		name          string
		from          model.Status
		to            string // "" means the request does not change status
		wantCompleted *string
		wantCancelled *string
		wantRemove    []string
	}{
		{"pending to completed", model.StatusPending, "completed", ptr(today), nil, []string{"cancelled_at"}},
		{"pending to cancelled", model.StatusPending, "cancelled", nil, ptr(today), []string{"completed_at"}},
		{"cancelled to completed", model.StatusCancelled, "completed", ptr(today), nil, []string{"cancelled_at"}},
		{"completed to cancelled", model.StatusCompleted, "cancelled", nil, ptr(today), []string{"completed_at"}},
		{"completed to pending reopens", model.StatusCompleted, "pending", nil, nil, []string{"completed_at", "cancelled_at"}},
		{"cancelled to in-progress reopens", model.StatusCancelled, "in-progress", nil, nil, []string{"completed_at", "cancelled_at"}},
		{"pending to in-progress leaves dates alone", model.StatusPending, "in-progress", nil, nil, nil},
		{"no status change leaves dates alone", model.StatusCompleted, "", nil, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &model.Task{Status: tt.from}
			var req UpdateRequest
			if tt.to != "" {
				req.Status = ptr(tt.to)
			}

			ApplyStatusTransitionDates(task, &req, now)

			if !reflect.DeepEqual(req.Completed, tt.wantCompleted) {
				t.Errorf("Completed = %v, want %v", deref(req.Completed), deref(tt.wantCompleted))
			}
			if !reflect.DeepEqual(req.CancelledAt, tt.wantCancelled) {
				t.Errorf("CancelledAt = %v, want %v", deref(req.CancelledAt), deref(tt.wantCancelled))
			}
			if !reflect.DeepEqual(req.RemoveFields, tt.wantRemove) {
				t.Errorf("RemoveFields = %v, want %v", req.RemoveFields, tt.wantRemove)
			}
		})
	}
}

func ptr(s string) *string { return &s }

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
