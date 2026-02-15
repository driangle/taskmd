package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type graphOutput struct {
	Nodes  []graphNode `json:"nodes"`
	Edges  []graphEdge `json:"edges"`
	Cycles [][]string  `json:"cycles"`
}

type graphNode struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority,omitempty"`
	Group    string `json:"group,omitempty"`
}

type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func callGraph(t *testing.T, session *gomcp.ClientSession, args map[string]any) graphOutput {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "graph",
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %+v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var out graphOutput
	if err := json.Unmarshal([]byte(text.Text), &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	return out
}

func callGraphExpectError(t *testing.T, session *gomcp.ClientSession, args map[string]any) {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "graph",
		Arguments: args,
	})
	if err != nil {
		return
	}
	if !result.IsError {
		t.Fatal("expected error but tool succeeded")
	}
}

func TestGraphTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(out.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(out.Nodes))
	}

	// Tasks 002 and 004 depend on 001, so we should have 2 edges
	if len(out.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(out.Edges))
	}
}

func TestGraphTool_NodesHaveExpectedFields(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	nodeMap := make(map[string]graphNode)
	for _, n := range out.Nodes {
		nodeMap[n.ID] = n
	}

	node001, ok := nodeMap["001"]
	if !ok {
		t.Fatal("expected node 001")
	}
	if node001.Title != "Setup project" {
		t.Errorf("expected title 'Setup project', got %q", node001.Title)
	}
	if node001.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", node001.Status)
	}
}

func TestGraphTool_EdgesReflectDependencies(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	edgeSet := make(map[string]bool)
	for _, e := range out.Edges {
		edgeSet[e.From+"->"+e.To] = true
	}

	// 002 depends on 001, so edge from 001 to 002
	if !edgeSet["001->002"] {
		t.Error("expected edge 001->002")
	}
	// 004 depends on 001, so edge from 001 to 004
	if !edgeSet["001->004"] {
		t.Error("expected edge 001->004")
	}
}

func TestGraphTool_ExcludeStatus(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir":       tmpDir,
		"exclude_status": []string{"completed"},
	})

	for _, n := range out.Nodes {
		if n.Status == "completed" {
			t.Errorf("expected completed tasks to be excluded, found %s", n.ID)
		}
	}

	// Task 001 is completed, so should be excluded (leaving 3)
	if len(out.Nodes) != 3 {
		t.Errorf("expected 3 nodes after excluding completed, got %d", len(out.Nodes))
	}
}

func TestGraphTool_RootTaskID(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir":     tmpDir,
		"root_task_id": "001",
	})

	// 001 has downstream deps 002 and 004, so subgraph includes 001, 002, 004
	nodeIDs := make(map[string]bool)
	for _, n := range out.Nodes {
		nodeIDs[n.ID] = true
	}

	if !nodeIDs["001"] {
		t.Error("expected root task 001 in subgraph")
	}
	if !nodeIDs["002"] {
		t.Error("expected dependent 002 in subgraph")
	}
	if !nodeIDs["004"] {
		t.Error("expected dependent 004 in subgraph")
	}
	// 003 has no relation to 001
	if nodeIDs["003"] {
		t.Error("task 003 should not be in subgraph of 001")
	}
}

func TestGraphTool_RootTaskNotFound(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callGraphExpectError(t, session, map[string]any{
		"task_dir":     tmpDir,
		"root_task_id": "999",
	})
}

func TestGraphTool_Filter(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir": tmpDir,
		"filters":  []string{"priority=high"},
	})

	for _, n := range out.Nodes {
		if n.Priority != "high" && n.Priority != "" {
			t.Errorf("expected only high priority tasks, found %s with priority %s", n.ID, n.Priority)
		}
	}
}

func TestGraphTool_DetectsCycles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create circular dependencies: A depends on B, B depends on A
	tasks := map[string]string{
		"001-a.md": `---
id: "001"
title: "Task A"
status: pending
priority: high
effort: small
dependencies: ["002"]
tags: []
created: 2026-01-01
---

# Task A
`,
		"002-b.md": `---
id: "002"
title: "Task B"
status: pending
priority: high
effort: small
dependencies: ["001"]
tags: []
created: 2026-01-01
---

# Task B
`,
	}

	for name, content := range tasks {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	session := setupTestServer(t)
	out := callGraph(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(out.Cycles) == 0 {
		t.Error("expected cycles to be detected")
	}
}

func TestGraphTool_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	session := setupTestServer(t)

	out := callGraph(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(out.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(out.Nodes))
	}
	if len(out.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(out.Edges))
	}
}

func TestGraphTool_InvalidFilter(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callGraphExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"filters":  []string{"bad-filter"},
	})
}

func TestGraphTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "graph" {
			found = true
			if tool.Description == "" {
				t.Error("graph tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("graph tool not found in tools list")
	}
}
