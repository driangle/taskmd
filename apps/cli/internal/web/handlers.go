package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"os/exec"

	"github.com/driangle/taskmd/sdk/go/board"
	"github.com/driangle/taskmd/sdk/go/feed"
	"github.com/driangle/taskmd/sdk/go/graph"
	"github.com/driangle/taskmd/sdk/go/metrics"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/next"
	"github.com/driangle/taskmd/sdk/go/search"
	"github.com/driangle/taskmd/sdk/go/taskfile"
	"github.com/driangle/taskmd/sdk/go/tracks"
	"github.com/driangle/taskmd/sdk/go/validator"
	"github.com/driangle/taskmd/sdk/go/worklog"
)

// ConfigResponse is the JSON response for GET /api/config.
type ConfigResponse struct {
	ReadOnly bool        `json:"readonly"`
	Version  string      `json:"version"`
	Phases   []PhaseInfo `json:"phases"`
}

func handleProjects(listFn func() ([]ProjectEntry, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if listFn == nil {
			writeJSON(w, []ProjectEntry{})
			return
		}
		projects, err := listFn()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if projects == nil {
			projects = []ProjectEntry{}
		}
		writeJSON(w, projects)
	}
}

func handleConfig(cfg Config) http.HandlerFunc {
	phases := cfg.Phases
	if phases == nil {
		phases = []PhaseInfo{}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		p := effectivePhases(r, phases)
		if p == nil {
			p = []PhaseInfo{}
		}
		writeJSON(w, ConfigResponse{
			ReadOnly: cfg.ReadOnly,
			Version:  cfg.Version,
			Phases:   p,
		})
	}
}

// getFilteredTasks returns tasks from the provider, optionally filtered by a
// "phase" query parameter.
func getFilteredTasks(dp *DataProvider, r *http.Request) ([]*model.Task, error) {
	tasks, err := dp.GetTasks()
	if err != nil {
		return nil, err
	}
	phase := r.URL.Query().Get("phase")
	if phase == "" {
		return tasks, nil
	}
	filtered := make([]*model.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Phase == phase {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TaskDetail includes the body field for individual task detail views
type TaskDetail struct {
	*model.Task
	Body           string `json:"body"`
	WorklogEntries int    `json:"worklog_entries,omitempty"`
	WorklogUpdated string `json:"worklog_updated,omitempty"`
}

func handleSearch(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results := search.Search(tasks, q)
		if results == nil {
			results = []search.Result{}
		}

		writeJSON(w, results)
	}
}

func handleTasks(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, tasks)
	}
}

func handleTaskByID(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		taskID := r.PathValue("id")
		if taskID == "" {
			http.Error(w, "task ID is required", http.StatusBadRequest)
			return
		}

		tasks, err := dp.GetTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Find task by ID
		var foundTask *model.Task
		for _, task := range tasks {
			if task.ID == taskID {
				foundTask = task
				break
			}
		}

		if foundTask == nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		// Return task with body and worklog metadata
		detail := TaskDetail{
			Task: foundTask,
			Body: foundTask.Body,
		}

		wlPath := worklog.WorklogPath(foundTask.FilePath, taskID)
		if worklog.Exists(wlPath) {
			if wl, err := worklog.ParseWorklog(wlPath); err == nil && len(wl.Entries) > 0 {
				detail.WorklogEntries = len(wl.Entries)
				last := wl.Entries[len(wl.Entries)-1]
				detail.WorklogUpdated = last.Timestamp.Format("2006-01-02T15:04:05Z07:00")
			}
		}

		writeJSON(w, detail)
	}
}

func handleBoard(dp *DataProvider, phases []PhaseInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		phases := effectivePhases(r, phases)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		groupBy := r.URL.Query().Get("groupBy")
		if groupBy == "" {
			groupBy = "status"
		}

		grouped, err := board.GroupTasks(tasks, groupBy)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if groupBy == "phase" && len(phases) > 0 {
			phaseOrder := make([]string, len(phases))
			for i, p := range phases {
				phaseOrder[i] = p.ID
			}
			board.ReorderKeys(grouped, phaseOrder)
		}

		writeJSON(w, board.ToJSON(grouped))
	}
}

func handleGraph(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		g := graph.NewGraph(tasks)
		writeJSON(w, g.ToJSON())
	}
}

func handleGraphMermaid(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		g := graph.NewGraph(tasks)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(g.ToMermaid(""))) //nolint:errcheck
	}
}

func handleStats(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		m := metrics.Calculate(tasks)
		writeJSON(w, m)
	}
}

func handleNext(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		archivedTasks, err := dp.GetArchivedTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		limit := 5
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}

		filters := r.URL.Query()["filter"]

		recs, err := next.Recommend(tasks, next.Options{
			Limit:         limit,
			Filters:       filters,
			ArchivedTasks: archivedTasks,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, recs)
	}
}

func handleTracks(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		archivedTasks, err := dp.GetArchivedTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filters := r.URL.Query()["filter"]
		scope := r.URL.Query().Get("scope")

		result, err := tracks.Assign(tasks, tracks.Options{
			Filters:       filters,
			ArchivedTasks: archivedTasks,
			Scope:         scope,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if v := r.URL.Query().Get("limit"); v != "" {
			if n, parseErr := strconv.Atoi(v); parseErr == nil && n > 0 && n < len(result.Tracks) {
				result.Tracks = result.Tracks[:n]
			}
		}

		writeJSON(w, result)
	}
}

func handleValidate(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		tasks, err := getFilteredTasks(dp, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		v := validator.NewValidator(false)
		result := v.Validate(tasks)
		writeJSON(w, result)
	}
}

// TaskUpdateRequest is the JSON request body for PUT /api/tasks/{id}.
type TaskUpdateRequest struct {
	Title    *string   `json:"title"`
	Status   *string   `json:"status"`
	Priority *string   `json:"priority"`
	Effort   *string   `json:"effort"`
	Type     *string   `json:"type"`
	Owner    *string   `json:"owner"`
	Parent   *string   `json:"parent"`
	Tags     *[]string `json:"tags"`
	Body     *string   `json:"body"`
}

// ErrorResponse is a structured JSON error response.
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, status int, msg string, details []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg, Details: details}) //nolint:errcheck
}

func findTaskByID(tasks []*model.Task, id string) *model.Task {
	for _, t := range tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func handleUpdateTask(dp *DataProvider, readonly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		if readonly {
			writeError(w, http.StatusForbidden, "server is in read-only mode", nil)
			return
		}

		taskID := r.PathValue("id")
		if taskID == "" {
			writeError(w, http.StatusBadRequest, "task ID is required", nil)
			return
		}

		var body TaskUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body", []string{err.Error()})
			return
		}

		req := toUpdateRequest(body)

		if errs := taskfile.ValidateUpdateRequest(req); len(errs) > 0 {
			writeError(w, http.StatusBadRequest, "validation failed", errs)
			return
		}

		tasks, err := dp.GetTasks()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load tasks", nil)
			return
		}

		found := findTaskByID(tasks, taskID)
		if found == nil {
			writeError(w, http.StatusNotFound, "task not found: "+taskID, nil)
			return
		}

		if err := taskfile.UpdateTaskFileLocked(dp.ScanDir(), found.FilePath, found.ID, req); err != nil {
			handleFileUpdateError(w, err)
			return
		}

		dp.Invalidate()

		updated, err := reloadTask(dp, taskID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reload tasks", nil)
			return
		}

		writeJSON(w, TaskDetail{Task: updated, Body: updated.Body})
	}
}

func toUpdateRequest(body TaskUpdateRequest) taskfile.UpdateRequest {
	return taskfile.UpdateRequest{
		Title:    body.Title,
		Status:   body.Status,
		Priority: body.Priority,
		Effort:   body.Effort,
		Type:     body.Type,
		Owner:    body.Owner,
		Parent:   body.Parent,
		Tags:     body.Tags,
		Body:     body.Body,
	}
}

func handleFileUpdateError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "no valid frontmatter") {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeError(w, http.StatusInternalServerError, "failed to update task file", []string{err.Error()})
}

func reloadTask(dp *DataProvider, taskID string) (*model.Task, error) {
	tasks, err := dp.GetTasks()
	if err != nil {
		return nil, err
	}
	found := findTaskByID(tasks, taskID)
	if found == nil {
		return nil, fmt.Errorf("task not found after update: %s", taskID)
	}
	return found, nil
}

// WorklogEntryJSON is a single worklog entry for the API.
type WorklogEntryJSON struct {
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

func handleWorklog(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		taskID := r.PathValue("id")
		if taskID == "" {
			http.Error(w, "task ID is required", http.StatusBadRequest)
			return
		}

		tasks, err := dp.GetTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		found := findTaskByID(tasks, taskID)
		if found == nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		wlPath := worklog.WorklogPath(found.FilePath, taskID)
		if !worklog.Exists(wlPath) {
			writeJSON(w, []WorklogEntryJSON{})
			return
		}

		wl, err := worklog.ParseWorklog(wlPath)
		if err != nil {
			writeJSON(w, []WorklogEntryJSON{})
			return
		}

		entries := make([]WorklogEntryJSON, len(wl.Entries))
		for i, e := range wl.Entries {
			entries[i] = WorklogEntryJSON{
				Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
				Content:   e.Content,
			}
		}

		writeJSON(w, entries)
	}
}

func handleFeed(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		q := r.URL.Query()

		limit := 20
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}

		source := q.Get("source")
		if source == "" {
			source = "all"
		}
		validSources := map[string]bool{"all": true, "git": true, "worklog": true}
		if !validSources[source] {
			http.Error(w, "invalid source: must be all, git, or worklog", http.StatusBadRequest)
			return
		}

		entries, err := feed.Query(feed.Options{
			TasksDir:  dp.ScanDir(),
			Limit:     limit,
			Since:     q.Get("since"),
			Scope:     q.Get("scope"),
			Source:    source,
			GitLogFn:  webGitLog,
			GitShowFn: webGitShow,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if entries == nil {
			entries = []feed.FeedEntry{}
		}
		writeJSON(w, entries)
	}
}

func webGitLog(_ string, args []string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func webGitShow(hash, path string) (string, error) {
	cmd := exec.Command("git", "show", hash+":"+path)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
