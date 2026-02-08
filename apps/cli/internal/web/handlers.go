package web

import (
	"encoding/json"
	"net/http"

	"github.com/driangle/md-task-tracker/apps/cli/internal/board"
	"github.com/driangle/md-task-tracker/apps/cli/internal/graph"
	"github.com/driangle/md-task-tracker/apps/cli/internal/metrics"
	"github.com/driangle/md-task-tracker/apps/cli/internal/validator"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		w.Write([]byte(g.ToMermaid("")))
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
