package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `sync:
  sources:
    - name: github
      project: "owner/repo"
      token_env: "GITHUB_TOKEN"
      output_dir: "tasks/github"
      field_map:
        status:
          open: pending
          closed: completed
        priority:
          p0: critical
        labels_to_tags: true
        assignee_to_owner: true
      filters:
        labels: ["task"]
`
	writeFile(t, filepath.Join(dir, ".taskmd.yaml"), content)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}

	s := cfg.Sources[0]
	if s.Name != "github" {
		t.Errorf("expected name=github, got %q", s.Name)
	}
	if s.Project != "owner/repo" {
		t.Errorf("expected project=owner/repo, got %q", s.Project)
	}
	if s.OutputDir != "tasks/github" {
		t.Errorf("expected output_dir=tasks/github, got %q", s.OutputDir)
	}
	if s.FieldMap.Status["open"] != "pending" {
		t.Errorf("expected status.open=pending, got %q", s.FieldMap.Status["open"])
	}
	if !s.FieldMap.LabelsToTags {
		t.Error("expected labels_to_tags=true")
	}
	if !s.FieldMap.AssigneeToOwner {
		t.Error("expected assignee_to_owner=true")
	}
}

func TestLoadConfig_MultipleSources(t *testing.T) {
	dir := t.TempDir()
	content := `sync:
  sources:
    - name: github
      output_dir: "tasks/github"
    - name: jira
      output_dir: "tasks/jira"
`
	writeFile(t, filepath.Join(dir, ".taskmd.yaml"), content)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(cfg.Sources))
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".taskmd.yaml"), "{{invalid yaml")

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoadConfig_NoSources(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".taskmd.yaml"), "sync:\n  sources: []\n")

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestLoadConfig_MissingName(t *testing.T) {
	dir := t.TempDir()
	content := `sync:
  sources:
    - output_dir: "tasks/foo"
`
	writeFile(t, filepath.Join(dir, ".taskmd.yaml"), content)

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for source without name")
	}
}

func TestLoadConfig_MissingOutputDir(t *testing.T) {
	dir := t.TempDir()
	content := `sync:
  sources:
    - name: github
`
	writeFile(t, filepath.Join(dir, ".taskmd.yaml"), content)

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for source without output_dir")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}
