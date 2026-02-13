package taskfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

const inlineTagsTask = `---
id: "001"
title: "Setup project"
status: pending
priority: high
effort: small
tags: ["infra", "setup"]
created: 2026-02-08
---

# Setup project

Initial project setup.
`

const multilineTagsTask = `---
id: "002"
title: "Auth system"
status: in-progress
priority: critical
effort: large
tags:
  - backend
  - security
created: 2026-02-08
---

# Auth system

JWT-based auth.
`

const noTagsTask = `---
id: "003"
title: "No tags task"
status: pending
priority: medium
effort: medium
created: 2026-02-08
---

# No tags task

Body content here.
`

func strPtr(s string) *string { return &s }

func TestUpdateTaskFile_SingleScalarField(t *testing.T) {
	path := createTestFile(t, inlineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{Status: strPtr("completed")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("expected status to be updated")
	}
	// Other fields preserved
	if !strings.Contains(s, "priority: high") {
		t.Error("expected priority to be preserved")
	}
	if !strings.Contains(s, "# Setup project") {
		t.Error("expected body to be preserved")
	}
}

func TestUpdateTaskFile_MultipleScalarFields(t *testing.T) {
	path := createTestFile(t, inlineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{
		Status:   strPtr("completed"),
		Priority: strPtr("low"),
		Effort:   strPtr("large"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("expected status update")
	}
	if !strings.Contains(s, "priority: low") {
		t.Error("expected priority update")
	}
	if !strings.Contains(s, "effort: large") {
		t.Error("expected effort update")
	}
}

func TestUpdateTaskFile_Title(t *testing.T) {
	path := createTestFile(t, inlineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{Title: strPtr("New Title")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, `title: "New Title"`) {
		t.Errorf("expected title update, got:\n%s", s)
	}
}

func TestUpdateTaskFile_ReplaceTags(t *testing.T) {
	path := createTestFile(t, inlineTagsTask)

	newTags := []string{"new-a", "new-b"}
	err := UpdateTaskFile(path, UpdateRequest{Tags: &newTags})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, `tags: ["new-a", "new-b"]`) {
		t.Errorf("expected tags replacement, got:\n%s", s)
	}
}

func TestUpdateTaskFile_AddRemoveTags(t *testing.T) {
	path := createTestFile(t, inlineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{
		AddTags: []string{"new-tag"},
		RemTags: []string{"setup"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, `tags: ["infra", "new-tag"]`) {
		t.Errorf("expected tag add/remove, got:\n%s", s)
	}
}

func TestUpdateTaskFile_MultilineTags(t *testing.T) {
	path := createTestFile(t, multilineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{
		AddTags: []string{"api"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "tags:\n  - backend\n  - security\n  - api") {
		t.Errorf("expected multiline tags preserved with addition, got:\n%s", s)
	}
}

func TestUpdateTaskFile_Body(t *testing.T) {
	path := createTestFile(t, inlineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{Body: strPtr("# New heading\n\nNew body content.")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "# New heading") {
		t.Error("expected new body heading")
	}
	if !strings.Contains(s, "New body content.") {
		t.Error("expected new body content")
	}
	if strings.Contains(s, "Initial project setup.") {
		t.Error("expected old body to be replaced")
	}
	// Frontmatter preserved
	if !strings.Contains(s, "status: pending") {
		t.Error("expected frontmatter to be preserved")
	}
}

func TestUpdateTaskFile_PartialUpdatePreservesOtherFields(t *testing.T) {
	path := createTestFile(t, multilineTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{Status: strPtr("completed")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "status: completed") {
		t.Error("expected status update")
	}
	if !strings.Contains(s, "priority: critical") {
		t.Error("expected priority preserved")
	}
	if !strings.Contains(s, "effort: large") {
		t.Error("expected effort preserved")
	}
	if !strings.Contains(s, "# Auth system") {
		t.Error("expected body preserved")
	}
}

func TestUpdateTaskFile_NoTags_AddTags(t *testing.T) {
	path := createTestFile(t, noTagsTask)

	err := UpdateTaskFile(path, UpdateRequest{
		AddTags: []string{"new-tag"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, `tags: ["new-tag"]`) {
		t.Errorf("expected tags to be added, got:\n%s", s)
	}
}

func TestUpdateTaskFile_NoFrontmatter(t *testing.T) {
	path := createTestFile(t, "# Just a heading\n\nNo frontmatter here.")

	err := UpdateTaskFile(path, UpdateRequest{Status: strPtr("completed")})
	if err == nil {
		t.Fatal("expected error for file without frontmatter")
	}
	if !strings.Contains(err.Error(), "no valid frontmatter") {
		t.Errorf("expected 'no valid frontmatter' error, got: %v", err)
	}
}

func TestValidateUpdateRequest_Valid(t *testing.T) {
	errs := ValidateUpdateRequest(UpdateRequest{
		Status:   strPtr("completed"),
		Priority: strPtr("high"),
		Effort:   strPtr("small"),
	})
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateUpdateRequest_InvalidStatus(t *testing.T) {
	errs := ValidateUpdateRequest(UpdateRequest{Status: strPtr("invalid")})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %s", errs[0])
	}
}

func TestValidateUpdateRequest_InvalidPriority(t *testing.T) {
	errs := ValidateUpdateRequest(UpdateRequest{Priority: strPtr("urgent")})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "invalid priority") {
		t.Errorf("expected 'invalid priority' error, got: %s", errs[0])
	}
}

func TestValidateUpdateRequest_InvalidEffort(t *testing.T) {
	errs := ValidateUpdateRequest(UpdateRequest{Effort: strPtr("huge")})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "invalid effort") {
		t.Errorf("expected 'invalid effort' error, got: %s", errs[0])
	}
}

func TestValidateUpdateRequest_MultipleErrors(t *testing.T) {
	errs := ValidateUpdateRequest(UpdateRequest{
		Status: strPtr("bad"),
		Effort: strPtr("bad"),
	})
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateUpdateRequest_NilFieldsSkipped(t *testing.T) {
	errs := ValidateUpdateRequest(UpdateRequest{})
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty request, got: %v", errs)
	}
}

func TestComputeNewTags(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		add     []string
		remove  []string
		want    []string
	}{
		{"add to empty", nil, []string{"a"}, nil, []string{"a"}},
		{"remove from list", []string{"a", "b"}, nil, []string{"a"}, []string{"b"}},
		{"add and remove", []string{"a", "b"}, []string{"c"}, []string{"a"}, []string{"b", "c"}},
		{"add duplicate", []string{"a"}, []string{"a"}, nil, []string{"a"}},
		{"remove nonexistent", []string{"a"}, nil, []string{"z"}, []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeNewTags(tt.current, tt.add, tt.remove)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindFrontmatterBounds(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		wantOpen  int
		wantClose int
	}{
		{"standard", []string{"---", "id: foo", "---", "body"}, 0, 2},
		{"no frontmatter", []string{"# Heading"}, -1, -1},
		{"unclosed", []string{"---", "id: foo"}, -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, closeIdx := FindFrontmatterBounds(tt.lines)
			if open != tt.wantOpen || closeIdx != tt.wantClose {
				t.Errorf("got (%d, %d), want (%d, %d)", open, closeIdx, tt.wantOpen, tt.wantClose)
			}
		})
	}
}
