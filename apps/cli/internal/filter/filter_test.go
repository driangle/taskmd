package filter

import (
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestApply_OwnerFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Owner: "alice"},
		{ID: "002", Title: "Task B", Owner: "bob"},
		{ID: "003", Title: "Task C", Owner: ""},
	}

	filtered, err := Apply(tasks, []string{"owner=alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestApply_MultipleFilters(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Status: model.StatusPending, Owner: "alice"},
		{ID: "002", Title: "Task B", Status: model.StatusPending, Owner: "bob"},
		{ID: "003", Title: "Task C", Status: model.StatusCompleted, Owner: "alice"},
	}

	filtered, err := Apply(tasks, []string{"status=pending", "owner=alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestApply_InvalidFilterFormat(t *testing.T) {
	tasks := []*model.Task{{ID: "001"}}

	_, err := Apply(tasks, []string{"badfilter"})
	if err == nil {
		t.Fatal("expected error for invalid filter format")
	}
}
