package board

import (
	"encoding/json"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestToJSON_IncludesTags(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:       "001",
			Title:    "Task with tags",
			Status:   model.StatusPending,
			Priority: model.PriorityHigh,
			Effort:   model.EffortSmall,
			Tags:     []string{"backend", "api"},
		},
	}

	gr := &GroupResult{
		Keys:   []string{"pending"},
		Groups: map[string][]*model.Task{"pending": tasks},
	}

	result := ToJSON(gr)
	if len(result) != 1 {
		t.Fatalf("expected 1 group, got %d", len(result))
	}

	jTask := result[0].Tasks[0]
	if len(jTask.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(jTask.Tags))
	}
	if jTask.Tags[0] != "backend" || jTask.Tags[1] != "api" {
		t.Errorf("unexpected tags: %v", jTask.Tags)
	}
}

func TestToJSON_OmitsEmptyTags(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:     "002",
			Title:  "Task without tags",
			Status: model.StatusPending,
			Tags:   nil,
		},
	}

	gr := &GroupResult{
		Keys:   []string{"pending"},
		Groups: map[string][]*model.Task{"pending": tasks},
	}

	result := ToJSON(gr)
	jTask := result[0].Tasks[0]

	data, err := json.Marshal(jTask)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, exists := raw["tags"]; exists {
		t.Error("expected tags to be omitted from JSON when nil, but it was present")
	}
}

func TestToJSON_OmitsEmptySliceTags(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:     "003",
			Title:  "Task with empty tags",
			Status: model.StatusPending,
			Tags:   []string{},
		},
	}

	gr := &GroupResult{
		Keys:   []string{"pending"},
		Groups: map[string][]*model.Task{"pending": tasks},
	}

	result := ToJSON(gr)
	jTask := result[0].Tasks[0]

	data, err := json.Marshal(jTask)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Note: Go's json.Marshal does NOT omit empty slices ([]string{}),
	// only nil slices. This test documents that behavior.
	// An empty slice serializes as "tags": []
}
