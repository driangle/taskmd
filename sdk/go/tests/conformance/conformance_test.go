package conformance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/filter"
	"github.com/driangle/taskmd/sdk/go/graph"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/next"
	"github.com/driangle/taskmd/sdk/go/scanner"
	"github.com/driangle/taskmd/sdk/go/search"
	"github.com/driangle/taskmd/sdk/go/validator"
)

const fixturesDir = "../../../../tests/conformance/fixtures/tasks"

// scanFixtureTasks scans the conformance fixtures directory and returns all tasks.
func scanFixtureTasks(t *testing.T) []*model.Task {
	t.Helper()
	s := scanner.NewScanner(fixturesDir, false, nil)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	return result.Tasks
}

// loadJSON reads a JSON file from the expected/ directory and unmarshals it into v.
func loadJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}

// taskIDs extracts sorted task IDs from a slice of tasks.
func taskIDs(tasks []*model.Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	sort.Strings(ids)
	return ids
}

// --- Scan ---

type expectedScanTask struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Priority string   `json:"priority"`
	Effort   string   `json:"effort"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
	Group    string   `json:"group"`
	Parent   string   `json:"parent"`
	FilePath string   `json:"file_path"`
}

func TestConformance_Scan(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var expected []expectedScanTask
	loadJSON(t, "../../../../tests/conformance/expected/scan/default.json", &expected)

	if len(tasks) != len(expected) {
		t.Fatalf("scan returned %d tasks, want %d", len(tasks), len(expected))
	}

	// Sort both by file path for comparison
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].FilePath < tasks[j].FilePath
	})

	absFixtures, _ := filepath.Abs(fixturesDir)
	for i, exp := range expected {
		task := tasks[i]
		relPath, _ := filepath.Rel(absFixtures, task.FilePath)
		if relPath == "" {
			relPath = task.FilePath
		}

		if task.ID != exp.ID {
			t.Errorf("task[%d] ID = %q, want %q", i, task.ID, exp.ID)
		}
		if task.Title != exp.Title {
			t.Errorf("task[%d] Title = %q, want %q", i, task.Title, exp.Title)
		}
		if string(task.Status) != exp.Status {
			t.Errorf("task[%d] Status = %q, want %q", i, task.Status, exp.Status)
		}
		if string(task.Priority) != exp.Priority {
			t.Errorf("task[%d] Priority = %q, want %q", i, task.Priority, exp.Priority)
		}
		if string(task.Effort) != exp.Effort {
			t.Errorf("task[%d] Effort = %q, want %q", i, task.Effort, exp.Effort)
		}
		if string(task.Type) != exp.Type {
			t.Errorf("task[%d] Type = %q, want %q", i, task.Type, exp.Type)
		}
		if task.Group != exp.Group {
			t.Errorf("task[%d] Group = %q, want %q", i, task.Group, exp.Group)
		}
		if task.Parent != exp.Parent {
			t.Errorf("task[%d] Parent = %q, want %q", i, task.Parent, exp.Parent)
		}
		if relPath != exp.FilePath {
			t.Errorf("task[%d] FilePath = %q, want %q", i, relPath, exp.FilePath)
		}
		if !slicesEqual(task.Tags, exp.Tags) {
			t.Errorf("task[%d] Tags = %v, want %v", i, task.Tags, exp.Tags)
		}
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return slices.Equal(a, b)
}

// --- Filter ---

type expectedFilter struct {
	Description   string   `json:"description"`
	Filter        string   `json:"filter"`
	ExpectedIDs   []string `json:"expected_ids"`
	ExpectedCount int      `json:"expected_count"`
}

func TestConformance_FilterStatus(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var exp expectedFilter
	loadJSON(t, "../../../../tests/conformance/expected/filter/status-pending.json", &exp)

	filtered, err := filter.Apply(tasks, []string{exp.Filter}, effort.Default())
	if err != nil {
		t.Fatalf("filter: %v", err)
	}

	gotIDs := taskIDs(filtered)
	sort.Strings(exp.ExpectedIDs)

	if len(filtered) != exp.ExpectedCount {
		t.Errorf("count = %d, want %d", len(filtered), exp.ExpectedCount)
	}
	if !slices.Equal(gotIDs, exp.ExpectedIDs) {
		t.Errorf("IDs = %v, want %v", gotIDs, exp.ExpectedIDs)
	}
}

func TestConformance_FilterTag(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var exp expectedFilter
	loadJSON(t, "../../../../tests/conformance/expected/filter/tag-api.json", &exp)

	filtered, err := filter.Apply(tasks, []string{exp.Filter}, effort.Default())
	if err != nil {
		t.Fatalf("filter: %v", err)
	}

	gotIDs := taskIDs(filtered)
	sort.Strings(exp.ExpectedIDs)

	if len(filtered) != exp.ExpectedCount {
		t.Errorf("count = %d, want %d", len(filtered), exp.ExpectedCount)
	}
	if !slices.Equal(gotIDs, exp.ExpectedIDs) {
		t.Errorf("IDs = %v, want %v", gotIDs, exp.ExpectedIDs)
	}
}

// --- Validate ---

type expectedValidation struct {
	Description string `json:"description"`
	Errors      int    `json:"errors"`
	Warnings    int    `json:"warnings"`
	TaskCount   int    `json:"task_count"`
	ExitCode    int    `json:"exit_code"`
}

func TestConformance_Validate(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var exp expectedValidation
	loadJSON(t, "../../../../tests/conformance/expected/validate/valid.json", &exp)

	v := validator.NewValidator(false)
	result := v.Validate(tasks)

	if len(tasks) != exp.TaskCount {
		t.Errorf("task count = %d, want %d", len(tasks), exp.TaskCount)
	}

	errorCount := 0
	warningCount := 0
	for _, issue := range result.Issues {
		if issue.Level == validator.LevelError {
			errorCount++
		} else {
			warningCount++
		}
	}

	if errorCount != exp.Errors {
		t.Errorf("errors = %d, want %d", errorCount, exp.Errors)
		for _, issue := range result.Issues {
			t.Logf("  issue: [%s] %s: %s", issue.Level, issue.TaskID, issue.Message)
		}
	}
	if warningCount != exp.Warnings {
		t.Errorf("warnings = %d, want %d", warningCount, exp.Warnings)
	}
}

// --- Next ---

type expectedNext struct {
	Description string `json:"description"`
	Results     []struct {
		Rank            int      `json:"rank"`
		ID              string   `json:"id"`
		Title           string   `json:"title"`
		Score           int      `json:"score"`
		Reasons         []string `json:"reasons"`
		DownstreamCount int      `json:"downstream_count"`
		OnCriticalPath  bool     `json:"on_critical_path"`
	} `json:"results"`
}

func TestConformance_Next(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var exp expectedNext
	loadJSON(t, "../../../../tests/conformance/expected/next/default.json", &exp)

	recs, err := next.Recommend(tasks, next.Options{Limit: 10})
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}

	if len(recs) != len(exp.Results) {
		t.Fatalf("got %d recommendations, want %d", len(recs), len(exp.Results))
	}

	for i, expRec := range exp.Results {
		got := recs[i]

		if got.Rank != expRec.Rank {
			t.Errorf("rec[%d] Rank = %d, want %d", i, got.Rank, expRec.Rank)
		}
		if got.ID != expRec.ID {
			t.Errorf("rec[%d] ID = %q, want %q", i, got.ID, expRec.ID)
		}
		if got.Title != expRec.Title {
			t.Errorf("rec[%d] Title = %q, want %q", i, got.Title, expRec.Title)
		}
		if got.Score != expRec.Score {
			t.Errorf("rec[%d] Score = %d, want %d (id=%s)", i, got.Score, expRec.Score, got.ID)
		}
		if got.DownstreamCount != expRec.DownstreamCount {
			t.Errorf("rec[%d] DownstreamCount = %d, want %d", i, got.DownstreamCount, expRec.DownstreamCount)
		}
		if got.OnCriticalPath != expRec.OnCriticalPath {
			t.Errorf("rec[%d] OnCriticalPath = %v, want %v", i, got.OnCriticalPath, expRec.OnCriticalPath)
		}

		// Compare reasons (treat nil and [] as equivalent)
		gotReasons := got.Reasons
		if gotReasons == nil {
			gotReasons = []string{}
		}
		wantReasons := expRec.Reasons
		if wantReasons == nil {
			wantReasons = []string{}
		}
		if !slices.Equal(gotReasons, wantReasons) {
			t.Errorf("rec[%d] Reasons = %v, want %v", i, gotReasons, wantReasons)
		}
	}
}

// --- Search ---

type expectedSearch struct {
	Description   string   `json:"description"`
	Query         string   `json:"query"`
	ExpectedIDs   []string `json:"expected_ids"`
	ExpectedCount int      `json:"expected_count"`
	Results       []struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		MatchLocation string `json:"match_location"`
	} `json:"results"`
}

func TestConformance_Search(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var exp expectedSearch
	loadJSON(t, "../../../../tests/conformance/expected/search/query-api.json", &exp)

	results := search.Search(tasks, exp.Query)

	if len(results) != exp.ExpectedCount {
		t.Fatalf("got %d results, want %d", len(results), exp.ExpectedCount)
	}

	gotIDs := make([]string, len(results))
	for i, r := range results {
		gotIDs[i] = r.ID
	}
	sort.Strings(gotIDs)
	sort.Strings(exp.ExpectedIDs)

	if !slices.Equal(gotIDs, exp.ExpectedIDs) {
		t.Errorf("IDs = %v, want %v", gotIDs, exp.ExpectedIDs)
	}

	for i, expResult := range exp.Results {
		// Find matching result by ID
		var got *search.Result
		for j := range results {
			if results[j].ID == expResult.ID {
				got = &results[j]
				break
			}
		}
		if got == nil {
			t.Errorf("result[%d] ID %q not found", i, expResult.ID)
			continue
		}
		if got.Title != expResult.Title {
			t.Errorf("result %q Title = %q, want %q", got.ID, got.Title, expResult.Title)
		}
		if got.MatchLocation != expResult.MatchLocation {
			t.Errorf("result %q MatchLocation = %q, want %q", got.ID, got.MatchLocation, expResult.MatchLocation)
		}
	}
}

// --- Graph ---

type expectedGraph struct {
	Description string `json:"description"`
	Nodes       []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Status   string `json:"status"`
		Priority string `json:"priority"`
		Group    string `json:"group"`
	} `json:"nodes"`
	Edges []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"edges"`
}

func TestConformance_Graph(t *testing.T) {
	tasks := scanFixtureTasks(t)

	var exp expectedGraph
	loadJSON(t, "../../../../tests/conformance/expected/graph/default.json", &exp)

	// Exclude completed tasks (default behavior)
	var nonCompleted []*model.Task
	for _, task := range tasks {
		if task.Status != model.StatusCompleted {
			nonCompleted = append(nonCompleted, task)
		}
	}

	g := graph.NewGraph(nonCompleted)
	jsonData := g.ToJSON()

	// Check nodes
	nodes, ok := jsonData["nodes"].([]map[string]any)
	if !ok {
		t.Fatalf("nodes not found or wrong type")
	}

	if len(nodes) != len(exp.Nodes) {
		t.Fatalf("got %d nodes, want %d", len(nodes), len(exp.Nodes))
	}

	for i, expNode := range exp.Nodes {
		got := nodes[i]
		if got["id"] != expNode.ID {
			t.Errorf("node[%d] id = %v, want %q", i, got["id"], expNode.ID)
		}
		if got["title"] != expNode.Title {
			t.Errorf("node[%d] title = %v, want %q", i, got["title"], expNode.Title)
		}
		if got["status"] != expNode.Status {
			t.Errorf("node[%d] status = %v, want %q", i, got["status"], expNode.Status)
		}
	}

	// Check edges
	edges, ok := jsonData["edges"].([]map[string]string)
	if !ok {
		t.Fatalf("edges not found or wrong type")
	}

	// Build node ID set for filtering edges with missing endpoints
	nodeIDSet := make(map[string]bool)
	for _, n := range nodes {
		if id, ok := n["id"].(string); ok {
			nodeIDSet[id] = true
		}
	}

	// Build edge sets for comparison (order may vary), filtering out edges with missing endpoints
	type edge struct{ from, to string }
	gotEdges := make(map[edge]bool)
	for _, e := range edges {
		if nodeIDSet[e["from"]] && nodeIDSet[e["to"]] {
			gotEdges[edge{e["from"], e["to"]}] = true
		}
	}
	wantEdges := make(map[edge]bool)
	for _, e := range exp.Edges {
		wantEdges[edge{e.From, e.To}] = true
	}

	for e := range wantEdges {
		if !gotEdges[e] {
			t.Errorf("missing edge: %s -> %s", e.from, e.to)
		}
	}
	for e := range gotEdges {
		if !wantEdges[e] {
			t.Errorf("unexpected edge: %s -> %s", e.from, e.to)
		}
	}
}
