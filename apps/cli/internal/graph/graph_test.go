package graph

import (
	"strings"
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func createTestTasks() []*model.Task {
	return []*model.Task{
		{
			ID:           "T1",
			Title:        "Task 1",
			Status:       model.StatusCompleted,
			Dependencies: []string{},
		},
		{
			ID:           "T2",
			Title:        "Task 2",
			Status:       model.StatusCompleted,
			Dependencies: []string{"T1"},
		},
		{
			ID:           "T3",
			Title:        "Task 3",
			Status:       model.StatusInProgress,
			Dependencies: []string{"T1", "T2"},
		},
		{
			ID:           "T4",
			Title:        "Task 4",
			Status:       model.StatusPending,
			Dependencies: []string{"T2"},
		},
		{
			ID:           "T5",
			Title:        "Task 5",
			Status:       model.StatusPending,
			Dependencies: []string{"T3", "T4"},
		},
	}
}

func TestNewGraph(t *testing.T) {
	tasks := createTestTasks()
	g := NewGraph(tasks)

	if len(g.Tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(g.Tasks))
	}

	if len(g.TaskMap) != 5 {
		t.Errorf("Expected 5 tasks in map, got %d", len(g.TaskMap))
	}

	// Check adjacency (T1 blocks T2 and T3)
	if len(g.Adjacency["T1"]) != 2 {
		t.Errorf("Expected T1 to block 2 tasks, got %d", len(g.Adjacency["T1"]))
	}

	// Check reverse adjacency (T3 depends on T1 and T2)
	if len(g.RevAdjacency["T3"]) != 2 {
		t.Errorf("Expected T3 to depend on 2 tasks, got %d", len(g.RevAdjacency["T3"]))
	}
}

func TestGetDownstream(t *testing.T) {
	tasks := createTestTasks()
	g := NewGraph(tasks)

	// T1 blocks T2, T3, T4 (via T2), T5 (via T3, T4)
	downstream := g.GetDownstream("T1")

	expectedDownstream := map[string]bool{
		"T2": true,
		"T3": true,
		"T4": true,
		"T5": true,
	}

	if len(downstream) != len(expectedDownstream) {
		t.Errorf("Expected %d downstream tasks for T1, got %d", len(expectedDownstream), len(downstream))
	}

	for taskID := range expectedDownstream {
		if !downstream[taskID] {
			t.Errorf("Expected %s to be downstream of T1", taskID)
		}
	}
}

func TestGetUpstream(t *testing.T) {
	tasks := createTestTasks()
	g := NewGraph(tasks)

	// T5 depends on T3, T4, T2 (via T3, T4), T1 (via T2, T3)
	upstream := g.GetUpstream("T5")

	expectedUpstream := map[string]bool{
		"T1": true,
		"T2": true,
		"T3": true,
		"T4": true,
	}

	if len(upstream) != len(expectedUpstream) {
		t.Errorf("Expected %d upstream tasks for T5, got %d", len(expectedUpstream), len(upstream))
	}

	for taskID := range expectedUpstream {
		if !upstream[taskID] {
			t.Errorf("Expected %s to be upstream of T5", taskID)
		}
	}
}

func TestDetectCycles(t *testing.T) {
	// Create tasks with a cycle: T1 -> T2 -> T3 -> T1
	cyclicTasks := []*model.Task{
		{
			ID:           "T1",
			Title:        "Task 1",
			Dependencies: []string{"T3"},
		},
		{
			ID:           "T2",
			Title:        "Task 2",
			Dependencies: []string{"T1"},
		},
		{
			ID:           "T3",
			Title:        "Task 3",
			Dependencies: []string{"T2"},
		},
	}

	g := NewGraph(cyclicTasks)
	cycles := g.DetectCycles()

	if len(cycles) == 0 {
		t.Error("Expected to detect at least one cycle")
	}
}

func TestDetectNoCycles(t *testing.T) {
	tasks := createTestTasks()
	g := NewGraph(tasks)

	cycles := g.DetectCycles()

	if len(cycles) != 0 {
		t.Errorf("Expected no cycles, found %d", len(cycles))
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := createTestTasks()
	g := NewGraph(tasks)

	// Filter to only T1 and T2
	filtered := g.FilterTasks(map[string]bool{
		"T1": true,
		"T2": true,
	})

	if len(filtered.Tasks) != 2 {
		t.Errorf("Expected 2 filtered tasks, got %d", len(filtered.Tasks))
	}

	if _, exists := filtered.TaskMap["T1"]; !exists {
		t.Error("Expected T1 in filtered graph")
	}

	if _, exists := filtered.TaskMap["T2"]; !exists {
		t.Error("Expected T2 in filtered graph")
	}

	if _, exists := filtered.TaskMap["T3"]; exists {
		t.Error("Did not expect T3 in filtered graph")
	}
}

func TestToMermaid(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:           "T1",
			Title:        "Task 1",
			Status:       model.StatusCompleted,
			Dependencies: []string{},
		},
		{
			ID:           "T2",
			Title:        "Task 2",
			Status:       model.StatusPending,
			Dependencies: []string{"T1"},
		},
	}

	g := NewGraph(tasks)
	output := g.ToMermaid("T2")

	if !strings.Contains(output, "graph TD") {
		t.Error("Expected mermaid output to contain 'graph TD'")
	}

	if !strings.Contains(output, "T1") {
		t.Error("Expected mermaid output to contain T1")
	}

	if !strings.Contains(output, "T2") {
		t.Error("Expected mermaid output to contain T2")
	}

	if !strings.Contains(output, "T1 --> T2") {
		t.Error("Expected mermaid output to contain edge T1 --> T2")
	}

	if !strings.Contains(output, ":::focus") {
		t.Error("Expected mermaid output to highlight focus task")
	}
}

func TestToDot(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:           "T1",
			Title:        "Task 1",
			Status:       model.StatusCompleted,
			Dependencies: []string{},
		},
		{
			ID:           "T2",
			Title:        "Task 2",
			Status:       model.StatusPending,
			Dependencies: []string{"T1"},
		},
	}

	g := NewGraph(tasks)
	output := g.ToDot("T2")

	if !strings.Contains(output, "digraph tasks") {
		t.Error("Expected DOT output to contain 'digraph tasks'")
	}

	if !strings.Contains(output, "T1") {
		t.Error("Expected DOT output to contain T1")
	}

	if !strings.Contains(output, "T2") {
		t.Error("Expected DOT output to contain T2")
	}

	if !strings.Contains(output, "T1 -> T2") {
		t.Error("Expected DOT output to contain edge T1 -> T2")
	}

	if !strings.Contains(output, "fillcolor=red") {
		t.Error("Expected DOT output to highlight focus task in red")
	}
}

func TestToASCII(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:           "T1",
			Title:        "Task 1",
			Status:       model.StatusCompleted,
			Dependencies: []string{},
		},
		{
			ID:           "T2",
			Title:        "Task 2",
			Status:       model.StatusInProgress,
			Dependencies: []string{"T1"},
		},
	}

	g := NewGraph(tasks)
	output := g.ToASCII("T1", true)

	if !strings.Contains(output, "T1") {
		t.Error("Expected ASCII output to contain T1")
	}

	if !strings.Contains(output, "T2") {
		t.Error("Expected ASCII output to contain T2")
	}

	if !strings.Contains(output, "✓") {
		t.Error("Expected ASCII output to contain completed checkmark")
	}

	if !strings.Contains(output, "⋯") {
		t.Error("Expected ASCII output to contain in-progress indicator")
	}
}

func TestToJSON(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:           "T1",
			Title:        "Task 1",
			Status:       model.StatusCompleted,
			Priority:     model.PriorityHigh,
			Dependencies: []string{},
			Created:      time.Now(),
		},
		{
			ID:           "T2",
			Title:        "Task 2",
			Status:       model.StatusPending,
			Dependencies: []string{"T1"},
			Created:      time.Now(),
		},
	}

	g := NewGraph(tasks)
	jsonOutput := g.ToJSON()

	nodes, ok := jsonOutput["nodes"].([]map[string]any)
	if !ok {
		t.Fatal("Expected nodes to be a slice of maps")
	}

	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes in JSON output, got %d", len(nodes))
	}

	edges, ok := jsonOutput["edges"].([]map[string]string)
	if !ok {
		t.Fatal("Expected edges to be a slice of maps")
	}

	if len(edges) != 1 {
		t.Errorf("Expected 1 edge in JSON output, got %d", len(edges))
	}

	// Check edge structure
	if edges[0]["from"] != "T1" || edges[0]["to"] != "T2" {
		t.Error("Expected edge from T1 to T2")
	}
}
