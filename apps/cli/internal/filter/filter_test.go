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

func TestApply_ParentFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Parent"},
		{ID: "002", Title: "Child A", Parent: "001"},
		{ID: "003", Title: "Child B", Parent: "001"},
		{ID: "004", Title: "Orphan"},
	}

	t.Run("filter by parent ID", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=001"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(filtered))
		}
	})

	t.Run("filter parent=true", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=true"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks with parent, got %d", len(filtered))
		}
	})

	t.Run("filter parent=false", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=false"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks without parent, got %d", len(filtered))
		}
	})
}
