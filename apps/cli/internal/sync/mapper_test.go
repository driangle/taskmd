package sync

import (
	"testing"
)

func TestMapExternalTask_StatusMapping(t *testing.T) {
	fm := FieldMap{
		Status: map[string]string{
			"open":   "pending",
			"closed": "completed",
		},
	}

	ext := ExternalTask{Title: "Test", Status: "open"}
	mapped := MapExternalTask(ext, fm)

	if mapped.Status != "pending" {
		t.Errorf("expected status=pending, got %q", mapped.Status)
	}

	ext.Status = "closed"
	mapped = MapExternalTask(ext, fm)

	if mapped.Status != "completed" {
		t.Errorf("expected status=completed, got %q", mapped.Status)
	}
}

func TestMapExternalTask_StatusFallback(t *testing.T) {
	fm := FieldMap{
		Status: map[string]string{
			"open": "pending",
		},
	}

	ext := ExternalTask{Title: "Test", Status: "unknown"}
	mapped := MapExternalTask(ext, fm)

	if mapped.Status != "pending" {
		t.Errorf("expected fallback status=pending, got %q", mapped.Status)
	}
}

func TestMapExternalTask_PriorityMapping(t *testing.T) {
	fm := FieldMap{
		Priority: map[string]string{
			"p0": "critical",
			"p1": "high",
		},
	}

	ext := ExternalTask{Title: "Test", Status: "open", Priority: "p0"}
	mapped := MapExternalTask(ext, fm)

	if mapped.Priority != "critical" {
		t.Errorf("expected priority=critical, got %q", mapped.Priority)
	}
}

func TestMapExternalTask_PriorityFallback(t *testing.T) {
	fm := FieldMap{
		Priority: map[string]string{
			"p0": "critical",
		},
	}

	ext := ExternalTask{Title: "Test", Status: "open", Priority: "unknown"}
	mapped := MapExternalTask(ext, fm)

	if mapped.Priority != "" {
		t.Errorf("expected empty priority fallback, got %q", mapped.Priority)
	}
}

func TestMapExternalTask_LabelsToTags(t *testing.T) {
	fm := FieldMap{LabelsToTags: true}

	ext := ExternalTask{
		Title:  "Test",
		Status: "open",
		Labels: []string{"bug", "backend"},
	}
	mapped := MapExternalTask(ext, fm)

	if len(mapped.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(mapped.Tags))
	}
	if mapped.Tags[0] != "bug" || mapped.Tags[1] != "backend" {
		t.Errorf("unexpected tags: %v", mapped.Tags)
	}
}

func TestMapExternalTask_LabelsNotMappedWhenDisabled(t *testing.T) {
	fm := FieldMap{LabelsToTags: false}

	ext := ExternalTask{
		Title:  "Test",
		Status: "open",
		Labels: []string{"bug"},
	}
	mapped := MapExternalTask(ext, fm)

	if len(mapped.Tags) != 0 {
		t.Errorf("expected no tags when labels_to_tags=false, got %v", mapped.Tags)
	}
}

func TestMapExternalTask_AssigneeToOwner(t *testing.T) {
	fm := FieldMap{AssigneeToOwner: true}

	ext := ExternalTask{Title: "Test", Status: "open", Assignee: "alice"}
	mapped := MapExternalTask(ext, fm)

	if mapped.Owner != "alice" {
		t.Errorf("expected owner=alice, got %q", mapped.Owner)
	}
}

func TestMapExternalTask_AssigneeNotMappedWhenDisabled(t *testing.T) {
	fm := FieldMap{AssigneeToOwner: false}

	ext := ExternalTask{Title: "Test", Status: "open", Assignee: "alice"}
	mapped := MapExternalTask(ext, fm)

	if mapped.Owner != "" {
		t.Errorf("expected empty owner when assignee_to_owner=false, got %q", mapped.Owner)
	}
}

func TestMapExternalTask_TitleAndDescription(t *testing.T) {
	fm := FieldMap{}

	ext := ExternalTask{
		Title:       "My Task",
		Description: "Some details",
		Status:      "open",
		URL:         "https://example.com/1",
	}
	mapped := MapExternalTask(ext, fm)

	if mapped.Title != "My Task" {
		t.Errorf("expected title=My Task, got %q", mapped.Title)
	}
	if mapped.Description != "Some details" {
		t.Errorf("expected description=Some details, got %q", mapped.Description)
	}
	if mapped.URL != "https://example.com/1" {
		t.Errorf("expected URL, got %q", mapped.URL)
	}
}

func TestMapExternalTask_EmptyFieldMap(t *testing.T) {
	fm := FieldMap{}

	ext := ExternalTask{
		Title:    "Test",
		Status:   "open",
		Priority: "high",
		Assignee: "bob",
		Labels:   []string{"label1"},
	}
	mapped := MapExternalTask(ext, fm)

	if mapped.Status != "pending" {
		t.Errorf("expected default status=pending, got %q", mapped.Status)
	}
	if mapped.Priority != "" {
		t.Errorf("expected empty priority, got %q", mapped.Priority)
	}
	if mapped.Owner != "" {
		t.Errorf("expected empty owner, got %q", mapped.Owner)
	}
	if len(mapped.Tags) != 0 {
		t.Errorf("expected no tags, got %v", mapped.Tags)
	}
}
