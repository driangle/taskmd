package cli

import (
	"html/template"
	"io"

	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/model"
)

type htmlReportData struct {
	Metrics      *metrics.Metrics
	Groups       []htmlGroupData
	GroupByLabel string
	CriticalPath []htmlTaskData
	BlockedTasks []htmlBlockedTaskData
	IncludeGraph bool
	MermaidSrc   string
}

type htmlGroupData struct {
	Name  string
	Count int
	Tasks []htmlTaskData
}

type htmlTaskData struct {
	ID       string
	Title    string
	Status   string
	Priority string
}

type htmlBlockedTaskData struct {
	ID        string
	Title     string
	WaitingOn string
}

func toHTMLData(data *reportData) htmlReportData {
	allTaskMap := buildTaskMapFromGroups(data)

	groups := make([]htmlGroupData, len(data.GroupedTasks.Keys))
	for i, key := range data.GroupedTasks.Keys {
		tasks := data.GroupedTasks.Groups[key]
		htasks := make([]htmlTaskData, len(tasks))
		for j, t := range tasks {
			htasks[j] = htmlTaskData{
				ID:       t.ID,
				Title:    t.Title,
				Status:   string(t.Status),
				Priority: string(t.Priority),
			}
		}
		groups[i] = htmlGroupData{
			Name:  key,
			Count: len(tasks),
			Tasks: htasks,
		}
	}

	cpTasks := make([]htmlTaskData, len(data.CriticalPath))
	for i, t := range data.CriticalPath {
		cpTasks[i] = htmlTaskData{
			ID:       t.ID,
			Title:    t.Title,
			Status:   t.Status,
			Priority: t.Priority,
		}
	}

	blocked := make([]htmlBlockedTaskData, len(data.BlockedTasks))
	for i, t := range data.BlockedTasks {
		blocked[i] = htmlBlockedTaskData{
			ID:        t.ID,
			Title:     t.Title,
			WaitingOn: formatWaitingOn(t.Dependencies, allTaskMap),
		}
	}

	return htmlReportData{
		Metrics:      data.Metrics,
		Groups:       groups,
		GroupByLabel: capitalizeFirst(data.GroupBy),
		CriticalPath: cpTasks,
		BlockedTasks: blocked,
		IncludeGraph: data.IncludeGraph,
		MermaidSrc:   data.GraphMermaid,
	}
}

func outputReportHTML(data *reportData, w io.Writer) error {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"statusClass": func(s string) string {
			switch model.Status(s) {
			case model.StatusCompleted:
				return "completed"
			case model.StatusInProgress:
				return "in-progress"
			case model.StatusBlocked:
				return "blocked"
			case model.StatusPending:
				return "pending"
			case model.StatusCancelled:
				return "cancelled"
			default:
				return ""
			}
		},
	}).Parse(reportHTMLTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, toHTMLData(data))
}

const reportHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Project Report</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 900px; margin: 0 auto; padding: 2rem; color: #24292f; line-height: 1.6; }
  h1 { border-bottom: 2px solid #d0d7de; padding-bottom: 0.5rem; }
  h2 { border-bottom: 1px solid #d0d7de; padding-bottom: 0.3rem; margin-top: 2rem; }
  table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
  th, td { border: 1px solid #d0d7de; padding: 0.5rem 0.75rem; text-align: left; }
  th { background: #f6f8fa; }
  .badge { display: inline-block; padding: 0.15rem 0.5rem; border-radius: 1rem; font-size: 0.85rem; font-weight: 500; }
  .badge.completed { background: #dafbe1; color: #116329; }
  .badge.in-progress { background: #fff8c5; color: #4d2d00; }
  .badge.blocked { background: #f6f8fa; color: #57606a; }
  .badge.pending { background: #ddf4ff; color: #0550ae; }
  .badge.cancelled { background: #ffebe9; color: #82071e; }
  ul { list-style: none; padding-left: 0; }
  ul li { padding: 0.3rem 0; }
  .task-id { font-family: monospace; font-weight: 600; }
  .waiting-on { color: #57606a; font-size: 0.9rem; margin-left: 1.5rem; display: block; }
  ol li { padding: 0.3rem 0; }
  .mermaid { background: #f6f8fa; padding: 1rem; border-radius: 6px; margin: 1rem 0; }
</style>
</head>
<body>
<h1>Project Report</h1>

<h2>Summary</h2>
<table>
  <tr><th>Metric</th><th>Value</th></tr>
  <tr><td>Total Tasks</td><td>{{.Metrics.TotalTasks}}</td></tr>
  <tr><td>Blocked Tasks</td><td>{{.Metrics.BlockedTasksCount}}</td></tr>
  <tr><td>Critical Path Length</td><td>{{.Metrics.CriticalPathLength}}</td></tr>
  <tr><td>Avg Dependencies</td><td>{{printf "%.1f" .Metrics.AvgDependenciesPerTask}}</td></tr>
</table>

<h2>Tasks by {{.GroupByLabel}}</h2>
{{range .Groups}}
<h3>{{.Name}} ({{.Count}})</h3>
<ul>
{{range .Tasks}}  <li><span class="task-id">[{{.ID}}]</span> {{.Title}}{{if .Priority}} <span class="badge {{statusClass .Status}}">{{.Priority}}</span>{{end}}</li>
{{end}}</ul>
{{end}}

<h2>Critical Path</h2>
{{if .CriticalPath}}
<ol>
{{range .CriticalPath}}  <li><span class="task-id">[{{.ID}}]</span> {{.Title}} <span class="badge {{statusClass .Status}}">{{.Status}}</span></li>
{{end}}</ol>
{{else}}
<p>No dependency chains found.</p>
{{end}}

<h2>Blocked Tasks</h2>
{{if .BlockedTasks}}
<ul>
{{range .BlockedTasks}}  <li><span class="task-id">[{{.ID}}]</span> {{.Title}}<span class="waiting-on">Waiting on: {{.WaitingOn}}</span></li>
{{end}}</ul>
{{else}}
<p>No blocked tasks.</p>
{{end}}

{{if .IncludeGraph}}
<h2>Dependency Graph</h2>
<div class="mermaid">
{{.MermaidSrc}}
</div>
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>mermaid.initialize({startOnLoad:true});</script>
<noscript><pre>{{.MermaidSrc}}</pre></noscript>
{{end}}

</body>
</html>
`
