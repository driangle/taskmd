package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/driangle/taskmd/sdk/go/effort"
)

// mockStaticFS creates a minimal embedded FS for testing.
func mockStaticFS() fstest.MapFS {
	return fstest.MapFS{
		"static/dist/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>taskmd</title>
  <script type="module" crossorigin src="/assets/index-abc123.js"></script>
  <link rel="stylesheet" crossorigin href="/assets/index-abc123.css">
</head>
<body>
  <div id="root"></div>
</body>
</html>`),
		},
		"static/dist/assets/index-abc123.js": &fstest.MapFile{
			Data: []byte(`console.log("app")`),
		},
		"static/dist/assets/index-abc123.css": &fstest.MapFile{
			Data: []byte(`body { margin: 0; }`),
		},
		"static/dist/favicon.ico": &fstest.MapFile{
			Data: []byte("icon"),
		},
	}
}

// exportWithMockFS runs Export after temporarily replacing StaticFiles.
// Since StaticFiles() returns an empty FS in non-embed builds, we
// use ExportWithFS which accepts an explicit FS for testing.
func exportWithMockFS(t *testing.T, cfg ExportConfig) error {
	t.Helper()
	return ExportWithFS(cfg, mockStaticFS())
}

func TestExport_OutputStructure(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test-1.0.0",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check core files exist
	expected := []string{
		"index.html",
		"404.html",
		"assets/index-abc123.js",
		"assets/index-abc123.css",
		"api/config.json",
		"api/tasks.json",
		"api/tasks/001.json",
		"api/tasks/002.json",
		"api/tasks/001/worklog.json",
		"api/tasks/002/worklog.json",
		"api/board/status.json",
		"api/board/priority.json",
		"api/board/effort.json",
		"api/board/type.json",
		"api/board/group.json",
		"api/board/tag.json",
		"api/graph.json",
		"api/stats.json",
		"api/next.json",
		"api/tracks.json",
		"api/validate.json",
	}

	for _, f := range expected {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}
}

func TestExport_TasksJSON(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "tasks.json"))
	if err != nil {
		t.Fatalf("failed to read tasks.json: %v", err)
	}

	var tasks []map[string]any
	if err := json.Unmarshal(data, &tasks); err != nil {
		t.Fatalf("invalid JSON in tasks.json: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestExport_TaskDetailJSON(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "tasks", "001.json"))
	if err != nil {
		t.Fatalf("failed to read task detail: %v", err)
	}

	var detail map[string]any
	if err := json.Unmarshal(data, &detail); err != nil {
		t.Fatalf("invalid JSON in task detail: %v", err)
	}

	if detail["id"] != "001" {
		t.Errorf("expected id '001', got %v", detail["id"])
	}

	if _, ok := detail["body"]; !ok {
		t.Error("expected body field in task detail")
	}
}

func TestExport_BoardJSON(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	groupByValues := []string{"status", "priority", "effort", "type", "group", "tag"}
	for _, gb := range groupByValues {
		path := filepath.Join(outDir, "api", "board", gb+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read board/%s.json: %v", gb, err)
			continue
		}

		var groups []map[string]any
		if err := json.Unmarshal(data, &groups); err != nil {
			t.Errorf("invalid JSON in board/%s.json: %v", gb, err)
		}
	}
}

func TestExport_ConfigJSON(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test-1.0.0",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "config.json"))
	if err != nil {
		t.Fatalf("failed to read config.json: %v", err)
	}

	var config ConfigResponse
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("invalid JSON in config.json: %v", err)
	}

	if !config.ReadOnly {
		t.Error("expected readonly to be true in exported config")
	}

	if config.Version != "test-1.0.0" {
		t.Errorf("expected version 'test-1.0.0', got %q", config.Version)
	}

	want := []string{"small", "medium", "large"}
	if !reflect.DeepEqual(config.Efforts, want) {
		t.Errorf("expected default efforts %v, got %v", want, config.Efforts)
	}
}

// The exported site must carry the same effort vocabulary the live server serves,
// so an exported site's pickers match the server's.
func TestExport_ConfigJSON_CustomEfforts(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	scale, err := effort.NewScale([]string{"xs", "s", "m", "l", "xl"})
	if err != nil {
		t.Fatalf("NewScale failed: %v", err)
	}

	if err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test-1.0.0",
		Efforts:   scale,
	}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "config.json"))
	if err != nil {
		t.Fatalf("failed to read config.json: %v", err)
	}

	var config ConfigResponse
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("invalid JSON in config.json: %v", err)
	}

	want := []string{"xs", "s", "m", "l", "xl"}
	if !reflect.DeepEqual(config.Efforts, want) {
		t.Errorf("expected efforts %v, got %v", want, config.Efforts)
	}
}

func TestExport_IndexHTML_FetchInterceptor(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}

	html := string(data)

	if !strings.Contains(html, "window.fetch") {
		t.Error("expected fetch interceptor in index.html")
	}

	if !strings.Contains(html, "window.EventSource") {
		t.Error("expected EventSource stub in index.html")
	}

	// Asset paths should be relative
	if strings.Contains(html, `="/assets/`) {
		t.Error("expected absolute asset paths to be rewritten to relative")
	}

	if !strings.Contains(html, `="./assets/`) {
		t.Error("expected relative asset paths in index.html")
	}
}

func TestExport_IndexHTML_BasePath(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/demo/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}

	html := string(data)

	if !strings.Contains(html, `<base href="/demo/">`) {
		t.Error("expected <base> tag with /demo/ in index.html")
	}
}

func TestExport_IndexHTML_BasePathRoot(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}

	if !strings.Contains(string(data), `<base href="/">`) {
		t.Error("expected <base href=\"/\"> tag for root base-path")
	}
}

func TestExport_SPARouteFiles(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Top-level SPA routes
	routes := []string{"tasks", "board", "graph", "next", "stats", "validate", "tracks"}
	for _, route := range routes {
		path := filepath.Join(outDir, route, "index.html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected SPA fallback at %s/index.html: %v", route, err)
		}
	}

	// Per-task routes
	for _, id := range []string{"001", "002"} {
		path := filepath.Join(outDir, "tasks", id, "index.html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected task route at tasks/%s/index.html: %v", id, err)
		}
	}

	// 404.html
	if _, err := os.Stat(filepath.Join(outDir, "404.html")); err != nil {
		t.Error("expected 404.html to exist")
	}
}

func TestExport_NoEmbeddedAssets(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	// Use empty FS to simulate no embedded assets
	err := ExportWithFS(ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	}, fstest.MapFS{})

	if err == nil {
		t.Fatal("expected error when no embedded assets")
	}

	if !strings.Contains(err.Error(), "no embedded web assets") {
		t.Errorf("expected 'no embedded web assets' error, got: %v", err)
	}
}

func TestExport_MissingIndexHTML(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	// FS with static/dist/ directory but no index.html
	missingIndexFS := fstest.MapFS{
		"static/dist/assets/app.js": &fstest.MapFile{
			Data: []byte(`console.log("app")`),
		},
	}

	err := ExportWithFS(ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	}, missingIndexFS)

	if err == nil {
		t.Fatal("expected error when index.html is missing")
	}

	if !strings.Contains(err.Error(), "no embedded web assets") {
		t.Errorf("expected 'no embedded web assets' error, got: %v", err)
	}
}

func TestExport_CustomOutput(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "custom-dir", "nested")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify output is in the custom directory
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Error("expected index.html in custom output directory")
	}
}

func TestExport_GraphJSON(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "graph.json"))
	if err != nil {
		t.Fatalf("failed to read graph.json: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON in graph.json: %v", err)
	}

	if _, ok := result["nodes"]; !ok {
		t.Error("expected 'nodes' in graph.json")
	}
	if _, ok := result["edges"]; !ok {
		t.Error("expected 'edges' in graph.json")
	}
}

func TestExport_StatsJSON(t *testing.T) {
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "api", "stats.json"))
	if err != nil {
		t.Fatalf("failed to read stats.json: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON in stats.json: %v", err)
	}

	totalTasks, ok := result["total_tasks"].(float64)
	if !ok {
		t.Fatal("expected total_tasks in stats.json")
	}
	if int(totalTasks) != 2 {
		t.Errorf("expected 2 total tasks, got %d", int(totalTasks))
	}
}

func TestPatchIndexHTML_RelativePaths(t *testing.T) {
	input := `<head><script src="/assets/app.js"></script></head>`
	result := patchIndexHTML(input, "/")

	if strings.Contains(result, `"/assets/`) {
		t.Error("expected absolute asset paths to be rewritten")
	}
}

func TestPatchIndexHTML_BasePathTrailingSlash(t *testing.T) {
	input := `<head></head>`
	result := patchIndexHTML(input, "/demo")

	if !strings.Contains(result, `<base href="/demo/">`) {
		t.Error("expected base-path to get trailing slash")
	}
}

func TestPatchIndexHTML_AlwaysHasBaseTag(t *testing.T) {
	input := `<head></head>`
	result := patchIndexHTML(input, "/")

	if !strings.Contains(result, `<base href="/">`) {
		t.Error("expected <base href=\"/\"> even for root base-path")
	}
}

func TestExport_TaskWithWorklog(t *testing.T) {
	taskDir := createTestTaskDir(t)

	// Create a worklog file for task 001
	wlDir := filepath.Join(taskDir, ".worklogs")
	if err := os.MkdirAll(wlDir, 0755); err != nil {
		t.Fatalf("failed to create worklogs dir: %v", err)
	}
	worklogContent := `## 2025-01-15T10:00:00Z

Started working on task one.

## 2025-01-15T14:30:00Z

Completed initial implementation.
`
	if err := os.WriteFile(filepath.Join(wlDir, "001.md"), []byte(worklogContent), 0644); err != nil {
		t.Fatalf("failed to write worklog: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify task detail includes worklog metadata
	detailData, err := os.ReadFile(filepath.Join(outDir, "api", "tasks", "001.json"))
	if err != nil {
		t.Fatalf("failed to read task detail: %v", err)
	}

	var detail map[string]any
	if err := json.Unmarshal(detailData, &detail); err != nil {
		t.Fatalf("invalid JSON in task detail: %v", err)
	}

	entryCount, ok := detail["worklog_entries"].(float64)
	if !ok || int(entryCount) != 2 {
		t.Errorf("expected worklog_entries=2, got %v", detail["worklog_entries"])
	}

	updated, ok := detail["worklog_updated"].(string)
	if !ok || updated == "" {
		t.Error("expected non-empty worklog_updated field")
	}
	if !strings.Contains(updated, "2025-01-15") {
		t.Errorf("expected worklog_updated to contain '2025-01-15', got %q", updated)
	}

	// Verify worklog.json was generated
	wlData, err := os.ReadFile(filepath.Join(outDir, "api", "tasks", "001", "worklog.json"))
	if err != nil {
		t.Fatalf("failed to read worklog.json: %v", err)
	}

	var entries []WorklogEntryJSON
	if err := json.Unmarshal(wlData, &entries); err != nil {
		t.Fatalf("invalid JSON in worklog.json: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 worklog entries, got %d", len(entries))
	}
	if entries[0].Timestamp != "2025-01-15T10:00:00Z" {
		t.Errorf("expected first timestamp, got %q", entries[0].Timestamp)
	}
	if !strings.Contains(entries[0].Content, "Started working") {
		t.Errorf("expected first entry content, got %q", entries[0].Content)
	}

	// Verify task 002 (no worklog) has empty worklog.json
	wl2Data, err := os.ReadFile(filepath.Join(outDir, "api", "tasks", "002", "worklog.json"))
	if err != nil {
		t.Fatalf("failed to read task 002 worklog.json: %v", err)
	}

	var entries2 []WorklogEntryJSON
	if err := json.Unmarshal(wl2Data, &entries2); err != nil {
		t.Fatalf("invalid JSON in task 002 worklog.json: %v", err)
	}
	if len(entries2) != 0 {
		t.Errorf("expected empty worklog for task 002, got %d entries", len(entries2))
	}
}

func TestExport_WithoutEmbeddedAssets(t *testing.T) {
	// Export() uses StaticFiles() which returns an empty FS in test builds.
	taskDir := createTestTaskDir(t)
	outDir := filepath.Join(t.TempDir(), "export")

	err := Export(ExportConfig{
		OutputDir: outDir,
		ScanDir:   taskDir,
		BasePath:  "/",
		Version:   "test",
	})

	if err == nil {
		t.Fatal("expected error from Export() without embedded assets")
	}
	if !strings.Contains(err.Error(), "no embedded web assets") {
		t.Errorf("expected 'no embedded web assets' error, got: %v", err)
	}
}

func TestWriteJSONFile_MarshalError(t *testing.T) {
	dir := t.TempDir()

	// Channels cannot be marshaled to JSON
	err := writeJSONFile(dir, "bad.json", make(chan int))
	if err == nil {
		t.Fatal("expected error when marshaling unmarshalable value")
	}
	if !strings.Contains(err.Error(), "failed to marshal") {
		t.Errorf("expected 'failed to marshal' error, got: %v", err)
	}
}

func TestWriteJSONFile_InvalidDir(t *testing.T) {
	// Create a file where a directory is expected
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "notadir")
	os.WriteFile(blocker, []byte("x"), 0644)

	err := writeJSONFile(filepath.Join(blocker, "sub"), "test.json", "data")
	if err == nil {
		t.Fatal("expected error when directory path is a file")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("expected 'failed to create directory' error, got: %v", err)
	}
}

func TestFetchInterceptor_QueryParamsStripped(t *testing.T) {
	script := fetchInterceptorScript()

	// The interceptor should split path and query, so /api/next?limit=5
	// becomes a fetch for ./api/next.json, not ./api/next?limit=5.json
	if !strings.Contains(script, "qi = full.indexOf('?')") {
		t.Error("expected fetch interceptor to parse query params separately")
	}
}
