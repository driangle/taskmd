package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/driangle/taskmd/sdk/go/board"
	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/graph"
	"github.com/driangle/taskmd/sdk/go/metrics"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/next"
	"github.com/driangle/taskmd/sdk/go/tracks"
	"github.com/driangle/taskmd/sdk/go/validator"
	"github.com/driangle/taskmd/sdk/go/worklog"
)

// ExportConfig holds configuration for the static site export.
type ExportConfig struct {
	OutputDir string
	ScanDir   string
	BasePath  string
	Verbose   bool
	Version   string
	// Efforts is the project's effort vocabulary. The zero value means the
	// default small, medium, large.
	Efforts effort.Scale
}

// Export generates a self-contained static site from task data.
func Export(cfg ExportConfig) error {
	return ExportWithFS(cfg, StaticFiles())
}

// ExportWithFS generates a static site using the provided embedded filesystem.
// This is separated from Export to allow tests to inject a mock FS.
func ExportWithFS(cfg ExportConfig, embeddedFS fs.FS) error {
	// Validate embedded assets
	staticFS, err := fs.Sub(embeddedFS, "static/dist")
	if err != nil {
		return fmt.Errorf("no embedded web assets: rebuild with `make build-full`")
	}
	indexHTML, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		return fmt.Errorf("no embedded web assets: rebuild with `make build-full`")
	}

	// Clean/create output directory
	if err := os.RemoveAll(cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Scan tasks
	dp := NewDataProvider(cfg.ScanDir, cfg.Verbose)
	tasks, err := dp.GetTasks()
	if err != nil {
		return fmt.Errorf("failed to scan tasks: %w", err)
	}

	archivedTasks, err := dp.GetArchivedTasks()
	if err != nil {
		return fmt.Errorf("failed to scan archived tasks: %w", err)
	}

	// Generate static JSON data files
	if err := generateDataFiles(cfg, tasks, archivedTasks); err != nil {
		return err
	}

	// Copy static assets (everything except index.html)
	if err := copyStaticAssets(staticFS, cfg.OutputDir); err != nil {
		return err
	}

	// Patch and write index.html
	patched := patchIndexHTML(string(indexHTML), cfg.BasePath)
	if err := os.WriteFile(filepath.Join(cfg.OutputDir, "index.html"), []byte(patched), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	// Generate SPA route fallback files
	if err := generateSPAFallbacks(cfg.OutputDir, patched, tasks); err != nil {
		return err
	}

	fmt.Printf("Exported static site to %s\n", cfg.OutputDir)
	return nil
}

func generateDataFiles(cfg ExportConfig, tasks []*model.Task, archivedTasks []*model.Task) error {
	apiDir := filepath.Join(cfg.OutputDir, "api")
	efforts := cfg.Efforts

	if err := writeJSONFile(apiDir, "config.json", ConfigResponse{
		ReadOnly: true,
		Version:  cfg.Version,
	}); err != nil {
		return err
	}

	if err := writeJSONFile(apiDir, "tasks.json", tasks); err != nil {
		return err
	}

	if err := generateTaskDetailFiles(filepath.Join(apiDir, "tasks"), tasks); err != nil {
		return err
	}

	if err := generateBoardFiles(filepath.Join(apiDir, "board"), tasks, efforts); err != nil {
		return err
	}

	return generateAnalyticsFiles(apiDir, tasks, archivedTasks, efforts)
}

func generateTaskDetailFiles(tasksDir string, tasks []*model.Task) error {
	for _, t := range tasks {
		detail := buildTaskDetail(t)
		if err := writeJSONFile(tasksDir, t.ID+".json", detail); err != nil {
			return err
		}

		entries := buildWorklogEntries(t)
		if err := writeJSONFile(filepath.Join(tasksDir, t.ID), "worklog.json", entries); err != nil {
			return err
		}
	}
	return nil
}

func buildTaskDetail(t *model.Task) TaskDetail {
	detail := TaskDetail{Task: t, Body: t.Body}
	wlPath := worklog.WorklogPath(t.FilePath, t.ID)
	if worklog.Exists(wlPath) {
		if wl, err := worklog.ParseWorklog(wlPath); err == nil && len(wl.Entries) > 0 {
			detail.WorklogEntries = len(wl.Entries)
			last := wl.Entries[len(wl.Entries)-1]
			detail.WorklogUpdated = last.Timestamp.Format("2006-01-02T15:04:05Z07:00")
		}
	}
	return detail
}

func buildWorklogEntries(t *model.Task) []WorklogEntryJSON {
	wlPath := worklog.WorklogPath(t.FilePath, t.ID)
	if !worklog.Exists(wlPath) {
		return []WorklogEntryJSON{}
	}
	wl, err := worklog.ParseWorklog(wlPath)
	if err != nil {
		return []WorklogEntryJSON{}
	}
	entries := make([]WorklogEntryJSON, len(wl.Entries))
	for i, e := range wl.Entries {
		entries[i] = WorklogEntryJSON{
			Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Content:   e.Content,
		}
	}
	return entries
}

func generateBoardFiles(boardDir string, tasks []*model.Task, efforts effort.Scale) error {
	for _, groupBy := range []string{"status", "priority", "effort", "type", "group", "tag"} {
		grouped, err := board.GroupTasks(tasks, groupBy, efforts)
		if err != nil {
			return fmt.Errorf("failed to group tasks by %s: %w", groupBy, err)
		}
		if err := writeJSONFile(boardDir, groupBy+".json", board.ToJSON(grouped)); err != nil {
			return err
		}
	}
	return nil
}

func generateAnalyticsFiles(apiDir string, tasks []*model.Task, archivedTasks []*model.Task, efforts effort.Scale) error {
	if err := writeJSONFile(apiDir, "graph.json", graph.NewGraph(tasks).ToJSON()); err != nil {
		return err
	}

	if err := writeJSONFile(apiDir, "stats.json", metrics.Calculate(tasks)); err != nil {
		return err
	}

	recs, err := next.Recommend(tasks, next.Options{Limit: 5, ArchivedTasks: archivedTasks, Efforts: efforts})
	if err != nil {
		return fmt.Errorf("failed to generate recommendations: %w", err)
	}
	if err := writeJSONFile(apiDir, "next.json", recs); err != nil {
		return err
	}

	tracksResult, err := tracks.Assign(tasks, tracks.Options{ArchivedTasks: archivedTasks, Efforts: efforts})
	if err != nil {
		return fmt.Errorf("failed to generate tracks: %w", err)
	}
	if err := writeJSONFile(apiDir, "tracks.json", tracksResult); err != nil {
		return err
	}

	v := validator.NewValidator(false)
	v.SetEffortScale(efforts)
	return writeJSONFile(apiDir, "validate.json", v.Validate(tasks))
}

func writeJSONFile(dir, filename string, v any) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", filename, err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func copyStaticAssets(staticFS fs.FS, outputDir string) error {
	return fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "index.html" {
			return nil // Skip — we patch and write it separately
		}

		dest := filepath.Join(outputDir, path)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		src, err := staticFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}
		defer src.Close()

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		out, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", dest, err)
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			return fmt.Errorf("failed to copy %s: %w", dest, err)
		}

		return nil
	})
}

// patchIndexHTML injects the fetch interceptor and adjusts paths for static hosting.
func patchIndexHTML(html string, basePath string) string {
	// Rewrite absolute asset paths to relative
	html = strings.ReplaceAll(html, `="/assets/`, `="./assets/`)
	html = strings.ReplaceAll(html, `="/favicon`, `="./favicon`)

	// Ensure trailing slash on basePath
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	// Always inject <base href> so relative paths resolve from root,
	// even when served from subdirectory SPA fallbacks like /board/index.html.
	baseTag := fmt.Sprintf(`<base href="%s">`, basePath)
	html = strings.Replace(html, "<head>", "<head>\n    "+baseTag, 1)

	// Inject fetch interceptor and EventSource stub before </head>
	html = strings.Replace(html, "</head>", fetchInterceptorScript()+"</head>", 1)

	return html
}

func fetchInterceptorScript() string {
	return `<script>
(function() {
  var origFetch = window.fetch;
  function json(body) { return new Response(body, {status:200, headers:{'Content-Type':'application/json'}}); }
  function redir(p, init) { return origFetch('./api/' + p + '.json', init); }
  window.fetch = function(input, init) {
    var url = (typeof input === 'string') ? input : input.url;
    if (url.indexOf('/api/') !== 0 && url.indexOf('./api/') !== 0) return origFetch.apply(this, arguments);
    var full = url.replace(/^\.\//, '/');
    var qi = full.indexOf('?'); var p = qi >= 0 ? full.substring(0, qi) : full; var q = qi >= 0 ? full.substring(qi) : '';
    if (p.indexOf('/api/search') === 0) return Promise.resolve(json('[]'));
    if (p === '/api/events') return new Promise(function() {});
    if (p === '/api/board') { var bm = q.match(/groupBy=(\w+)/); return redir('board/' + (bm ? bm[1] : 'status'), init); }
    var wm = p.match(/^\/api\/tasks\/([^/?]+)\/worklog/);
    if (wm) return redir('tasks/' + wm[1] + '/worklog', init);
    var tm = p.match(/^\/api\/tasks\/([^/?]+)$/);
    if (tm) return redir('tasks/' + tm[1], init);
    if (p === '/api/graph/mermaid') return Promise.resolve(new Response('', {status:200, headers:{'Content-Type':'text/plain'}}));
    return redir(p.replace(/^\/api\//, ''), init);
  };
  window.EventSource = function() { this.close = this.addEventListener = this.removeEventListener = function() {}; };
})();
</script>
`
}

func generateSPAFallbacks(outputDir, patchedHTML string, tasks []*model.Task) error {
	// Top-level SPA routes
	routes := []string{"tasks", "board", "graph", "next", "stats", "validate", "tracks"}
	for _, route := range routes {
		dir := filepath.Join(outputDir, route)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create route directory %s: %w", route, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(patchedHTML), 0644); err != nil {
			return fmt.Errorf("failed to write fallback for %s: %w", route, err)
		}
	}

	// Per-task detail routes: /tasks/{id}/index.html
	for _, t := range tasks {
		dir := filepath.Join(outputDir, "tasks", t.ID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create task route directory %s: %w", t.ID, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(patchedHTML), 0644); err != nil {
			return fmt.Errorf("failed to write fallback for task %s: %w", t.ID, err)
		}
	}

	// 404.html for GitHub Pages fallback
	if err := os.WriteFile(filepath.Join(outputDir, "404.html"), []byte(patchedHTML), 0644); err != nil {
		return fmt.Errorf("failed to write 404.html: %w", err)
	}

	return nil
}
