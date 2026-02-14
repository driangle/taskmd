package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func outputReportMarkdown(data *reportData, w io.Writer) error {
	fmt.Fprintln(w, "# Project Report")
	fmt.Fprintln(w)

	writeMarkdownSummary(data.Metrics, w)
	writeMarkdownGroups(data, w)
	writeMarkdownCriticalPath(data.CriticalPath, w)
	writeMarkdownBlockedTasks(data, w)

	if data.IncludeGraph {
		writeMarkdownGraph(data.GraphMermaid, w)
	}

	return nil
}

func writeMarkdownSummary(m *metrics.Metrics, w io.Writer) {
	fmt.Fprintln(w, "## Summary")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "| Metric | Value |\n")
	fmt.Fprintf(w, "|--------|-------|\n")
	fmt.Fprintf(w, "| Total Tasks | %d |\n", m.TotalTasks)
	fmt.Fprintf(w, "| Blocked Tasks | %d |\n", m.BlockedTasksCount)
	fmt.Fprintf(w, "| Critical Path Length | %d |\n", m.CriticalPathLength)
	fmt.Fprintf(w, "| Avg Dependencies | %.1f |\n", m.AvgDependenciesPerTask)
	fmt.Fprintln(w)

	writeMarkdownStatusBreakdown(m, w)
	writeMarkdownPriorityBreakdown(m, w)
}

func writeMarkdownStatusBreakdown(m *metrics.Metrics, w io.Writer) {
	statusOrder := []model.Status{
		model.StatusPending, model.StatusInProgress, model.StatusBlocked,
		model.StatusCompleted, model.StatusCancelled,
	}

	fmt.Fprintln(w, "### By Status")
	fmt.Fprintln(w)
	for _, s := range statusOrder {
		if count, ok := m.TasksByStatus[s]; ok && count > 0 {
			fmt.Fprintf(w, "- **%s**: %d\n", s, count)
		}
	}
	fmt.Fprintln(w)
}

func writeMarkdownPriorityBreakdown(m *metrics.Metrics, w io.Writer) {
	priorityOrder := []model.Priority{
		model.PriorityCritical, model.PriorityHigh,
		model.PriorityMedium, model.PriorityLow,
	}

	hasAny := false
	for _, p := range priorityOrder {
		if count, ok := m.TasksByPriority[p]; ok && count > 0 {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return
	}

	fmt.Fprintln(w, "### By Priority")
	fmt.Fprintln(w)
	for _, p := range priorityOrder {
		if count, ok := m.TasksByPriority[p]; ok && count > 0 {
			fmt.Fprintf(w, "- **%s**: %d\n", p, count)
		}
	}
	fmt.Fprintln(w)
}

func writeMarkdownGroups(data *reportData, w io.Writer) {
	fmt.Fprintf(w, "## Tasks by %s\n", capitalizeFirst(data.GroupBy))
	fmt.Fprintln(w)

	for _, key := range data.GroupedTasks.Keys {
		tasks := data.GroupedTasks.Groups[key]
		fmt.Fprintf(w, "### %s (%d)\n", key, len(tasks))
		fmt.Fprintln(w)
		for _, t := range tasks {
			line := fmt.Sprintf("- [%s] %s", t.ID, t.Title)
			if t.Priority != "" {
				line += fmt.Sprintf(" (priority: %s)", t.Priority)
			}
			fmt.Fprintln(w, line)
		}
		fmt.Fprintln(w)
	}
}

func writeMarkdownCriticalPath(cpTasks []reportTask, w io.Writer) {
	fmt.Fprintln(w, "## Critical Path")
	fmt.Fprintln(w)

	if len(cpTasks) == 0 {
		fmt.Fprintln(w, "No dependency chains found.")
		fmt.Fprintln(w)
		return
	}

	for i, t := range cpTasks {
		fmt.Fprintf(w, "%d. [%s] %s (%s)\n", i+1, t.ID, t.Title, t.Status)
	}
	fmt.Fprintln(w)
}

func writeMarkdownBlockedTasks(data *reportData, w io.Writer) {
	fmt.Fprintln(w, "## Blocked Tasks")
	fmt.Fprintln(w)

	if len(data.BlockedTasks) == 0 {
		fmt.Fprintln(w, "No blocked tasks.")
		fmt.Fprintln(w)
		return
	}

	taskMap := make(map[string]reportTask)
	for _, t := range data.CriticalPath {
		taskMap[t.ID] = t
	}
	for _, t := range data.BlockedTasks {
		taskMap[t.ID] = t
	}
	// Also build from grouped tasks for status lookups
	allTaskMap := buildTaskMapFromGroups(data)

	for _, t := range data.BlockedTasks {
		waitingOn := formatWaitingOn(t.Dependencies, allTaskMap)
		fmt.Fprintf(w, "- [%s] %s\n  Waiting on: %s\n", t.ID, t.Title, waitingOn)
	}
	fmt.Fprintln(w)
}

func buildTaskMapFromGroups(data *reportData) map[string]reportTask {
	m := make(map[string]reportTask)
	for _, key := range data.GroupedTasks.Keys {
		for _, t := range data.GroupedTasks.Groups[key] {
			m[t.ID] = reportTask{
				ID:     t.ID,
				Title:  t.Title,
				Status: string(t.Status),
			}
		}
	}
	return m
}

func formatWaitingOn(deps []string, taskMap map[string]reportTask) string {
	parts := make([]string, len(deps))
	for i, depID := range deps {
		if t, ok := taskMap[depID]; ok {
			parts[i] = fmt.Sprintf("%s (%s)", depID, t.Status)
		} else {
			parts[i] = depID
		}
	}
	return strings.Join(parts, ", ")
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func writeMarkdownGraph(mermaid string, w io.Writer) {
	fmt.Fprintln(w, "## Dependency Graph")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "```mermaid")
	fmt.Fprint(w, mermaid)
	fmt.Fprintln(w, "```")
	fmt.Fprintln(w)
}
