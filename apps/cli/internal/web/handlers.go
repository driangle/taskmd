package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/board"
	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/next"
	"github.com/driangle/taskmd/apps/cli/internal/taskfile"
	"github.com/driangle/taskmd/apps/cli/internal/validator"
)

// ConfigResponse is the JSON response for GET /api/config.
type ConfigResponse struct {
	ReadOnly bool   `json:"readonly"`
	Version  string `json:"version"`
}

func handleConfig(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, ConfigResponse{
			ReadOnly: cfg.ReadOnly,
			Version:  cfg.Version,
		})
	}
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
	Body string `json:"body"`
}

func handleTasks(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := dp.GetTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, tasks)
	}
}

func handleTaskByID(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// Return task with body
		detail := TaskDetail{
			Task: foundTask,
			Body: foundTask.Body,
		}

		writeJSON(w, detail)
	}
}

func handleBoard(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := dp.GetTasks()
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

		writeJSON(w, board.ToJSON(grouped))
	}
}

func handleGraph(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := dp.GetTasks()
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
		tasks, err := dp.GetTasks()
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
		tasks, err := dp.GetTasks()
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
		tasks, err := dp.GetTasks()
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
			Limit:   limit,
			Filters: filters,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, recs)
	}
}

func handleValidate(dp *DataProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := dp.GetTasks()
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
