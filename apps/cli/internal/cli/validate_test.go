package cli

import (
	"testing"
)

func TestParseScopeEntries_WithDescription(t *testing.T) {
	scopeMap := map[string]any{
		"cli/graph": map[string]any{
			"description": "Graph visualization",
			"paths":       []any{"apps/cli/internal/graph/"},
		},
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["cli/graph"]
	if !ok {
		t.Fatal("expected scope cli/graph to exist")
	}
	if sc.Description != "Graph visualization" {
		t.Errorf("Description = %q, want %q", sc.Description, "Graph visualization")
	}
	if len(sc.Paths) != 1 || sc.Paths[0] != "apps/cli/internal/graph/" {
		t.Errorf("Paths = %v, want [apps/cli/internal/graph/]", sc.Paths)
	}
}

func TestParseScopeEntries_WithoutDescription(t *testing.T) {
	scopeMap := map[string]any{
		"cli/output": map[string]any{
			"paths": []any{"apps/cli/internal/cli/format.go"},
		},
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["cli/output"]
	if !ok {
		t.Fatal("expected scope cli/output to exist")
	}
	if sc.Description != "" {
		t.Errorf("Description = %q, want empty string", sc.Description)
	}
	if len(sc.Paths) != 1 {
		t.Errorf("Paths = %v, want 1 element", sc.Paths)
	}
}
