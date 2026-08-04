package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/template"
)

// templatesListStdout runs `templates list <args...>`, fails on error, returns stdout.
func templatesListStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"templates", "list"}, args...)...)
	if res.Err != nil {
		t.Fatalf("templates list %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestTemplatesList_BuiltinTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	output := templatesListStdout(t, repo)

	// Should list built-in templates
	if !strings.Contains(output, "feature") {
		t.Error("expected 'feature' template in output")
	}
	if !strings.Contains(output, "bug") {
		t.Error("expected 'bug' template in output")
	}
	if !strings.Contains(output, "chore") {
		t.Error("expected 'chore' template in output")
	}
	if !strings.Contains(output, "built-in") {
		t.Error("expected 'built-in' source in output")
	}
}

func TestTemplatesList_JSONFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := templatesListStdout(t, repo, "--format", "json")

	var items []templateListItem
	if err := json.Unmarshal([]byte(output), &items); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if len(items) < 3 {
		t.Fatalf("expected at least 3 templates, got %d", len(items))
	}

	names := make(map[string]bool)
	for _, item := range items {
		names[item.Name] = true
		if item.Source == "" {
			t.Error("expected non-empty source")
		}
	}
	if !names["feature"] || !names["bug"] || !names["chore"] {
		t.Error("expected feature, bug, and chore templates")
	}
}

func TestTemplatesList_YAMLFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	output := templatesListStdout(t, repo, "--format", "yaml")

	if !strings.Contains(output, "name: feature") {
		t.Error("expected YAML output with feature template")
	}
	if !strings.Contains(output, "source: built-in") {
		t.Error("expected YAML output with built-in source")
	}
}

func TestTemplatesList_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("templates", "list", "--format", "xml")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestTemplatesList_ProjectTemplatesShown(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// Create a project-level template in a known project root.
	// Use the template package Discover directly to avoid viper dependency.
	customTemplate := `---
_template:
  name: custom
  description: "Custom project template"
title: "{{title}}"
id: "{{id}}"
status: pending
---

# {{title}}
`
	repo.Write(".taskmd/templates/custom.md", customTemplate)

	// Discover templates directly with known paths to verify project templates work.
	templates := template.Discover(repo.Dir, "")
	foundCustom := false
	for _, tmpl := range templates {
		if tmpl.Name == "custom" && tmpl.Source == "project" {
			foundCustom = true
		}
	}
	if !foundCustom {
		t.Error("expected 'custom' template from project source")
	}
}

func TestTemplatesList_NoTemplates(t *testing.T) {
	repo := newTaskRepo(t, nil)
	chdirTo(t, repo.Dir)

	// Clear built-in templates
	oldBuiltins := template.BuiltinTemplates
	template.BuiltinTemplates = nil
	defer func() { template.BuiltinTemplates = oldBuiltins }()

	output := templatesListStdout(t, repo)

	if !strings.Contains(output, "No templates found") {
		t.Errorf("expected 'No templates found', got: %s", output)
	}
}

func TestTemplatesList_TableHeaders(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := templatesListStdout(t, repo)

	if !strings.Contains(output, "NAME") {
		t.Error("expected NAME header")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("expected DESCRIPTION header")
	}
	if !strings.Contains(output, "SOURCE") {
		t.Error("expected SOURCE header")
	}
}
