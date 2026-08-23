package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"os/exec"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/board"
	"github.com/driangle/taskmd/sdk/go/effort"
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
	// Efforts is the project's effort vocabulary, lowest to highest. The web UI
	// uses it to populate the effort filter and the edit-form dropdown.
	Efforts []string `json:"efforts"`
	// Worktree describes the active worktree overlay; absent when inactive.
	Worktree *WorktreeOverlayInfo `json:"worktree,omitempty"`
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

func handleConfig(cfg Config, dp *DataProvider) http.HandlerFunc {
	phases := cfg.Phases
	if phases == nil {
		phases = []PhaseInfo{}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		p := effectivePhases(r, phases)
		if p == nil {
			p = []PhaseInfo{}
		}
		// Best effort: config must keep serving even if the overlay scan
		// fails; task endpoints report that error instead.
		info, _ := effectiveDP(r, dp).OverlayInfo()
		writeJSON(w, ConfigResponse{
			ReadOnly: cfg.ReadOnly,
			Version:  cfg.Version,
			Phases:   p,
			Efforts:  cfg.Efforts.Values(),
			Worktree: info,
		})
	}
}

// filterByPhase returns the tasks matching the phase, or all tasks when the
// phase is empty.
func filterByPhase(tasks []*model.Task, phase string) []*model.Task {
	if phase == "" {
		return tasks
	}
	filtered := make([]*model.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Phase == phase {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// getFilteredTasks returns effective tasks (merged across worktrees when the
// overlay is active) from the provider, optionally filtered by a "phase"
// query parameter. Status-aggregating endpoints serve this view.
func getFilteredTasks(dp *DataProvider, r *http.Request) ([]*model.Task, error) {
	tasks, err := dp.GetEffectiveTasks()
	if err != nil {
		return nil, err
	}
	return filterByPhase(tasks, r.URL.Query().Get("phase")), nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TaskDetail includes the body field for individual task detail views. The
// provenance fields are populated only when the worktree overlay is active,
// keeping the shape unchanged otherwise.
type TaskDetail struct {
	*model.Task
	Body           string `json:"body"`
	WorklogEntries int    `json:"worklog_entries,omitempty"`
	WorklogUpdated string `json:"worklog_updated,omitempty"`

	EffectiveStatus string               `json:"effective_status,omitempty"`
	EffectiveOwner  string               `json:"effective_owner,omitempty"`
	Worktree        string               `json:"worktree,omitempty"`
	Branch          string               `json:"branch,omitempty"`
	RemoteOnly      bool                 `json:"remote_only,omitempty"`
	Worktrees       []worktree.CopyEntry `json:"worktrees,omitempty"`
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
		overlay, err := dp.GetOverlay()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		phase := r.URL.Query().Get("phase")

		if overlay != nil {
			rows := make([]*worktree.Task, 0, len(overlay.Tasks))
			for _, ot := range overlay.Tasks {
				if phase == "" || ot.Phase == phase {
					rows = append(rows, ot)
				}
			}
			writeJSON(w, rows)
			return
		}

		tasks, err := dp.GetTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, filterByPhase(tasks, phase))
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

		foundTask, overlay, err := findMergedTask(dp, taskID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
		detail.applyProvenance(overlay, taskID)

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

func handleBoard(dp *DataProvider, phases []PhaseInfo, efforts effort.Scale) http.HandlerFunc {
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

		grouped, err := board.GroupTasks(tasks, groupBy, efforts)
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

		overlay, err := dp.GetOverlay()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, annotateBoard(board.ToJSON(grouped), overlay))
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

func handleNext(dp *DataProvider, efforts effort.Scale) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		overlay, err := dp.GetOverlay()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Recommend against the merged view: sibling-claimed and sibling-only
		// tasks are excluded but still resolve dependencies.
		var tasks []*model.Task
		var excluded map[string]string
		if overlay != nil {
			tasks, excluded = overlay.RecommendationInputs()
		} else if tasks, err = dp.GetTasks(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = filterByPhase(tasks, r.URL.Query().Get("phase"))

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
			Efforts:       efforts,
			Excluded:      excluded,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, recs)
	}
}

func handleTracks(dp *DataProvider, efforts effort.Scale) http.HandlerFunc {
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
			Efforts:       efforts,
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

func handleValidate(dp *DataProvider, efforts effort.Scale) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dp := effectiveDP(r, dp)
		// Validate the local tasks (after sibling-root attribution), not the
		// merged view — a sibling's copy of a task is never a duplicate here.
		tasks, err := dp.GetTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = filterByPhase(tasks, r.URL.Query().Get("phase"))

		v := validator.NewValidator(false)
		v.SetEffortScale(efforts)
		result := v.Validate(tasks)

		if overlay, oErr := dp.GetOverlay(); oErr == nil && overlay != nil {
			for _, warning := range overlay.Warnings {
				result.AddIssue(validator.LevelWarning, warning.TaskID, "", warning.Message)
			}
		}
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

func handleUpdateTask(dp *DataProvider, readonly bool, efforts effort.Scale) http.HandlerFunc {
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

		if errs := taskfile.ValidateUpdateRequest(req, efforts); len(errs) > 0 {
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
			writeMutationMiss(w, dp, taskID)
			return
		}

		if err := taskfile.UpdateTaskFile(found.FilePath, req); err != nil {
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

		found, _, err := findMergedTask(dp, taskID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
