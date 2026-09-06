package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/driangle/taskmd/sdk/go/model"
)

// Terminal rendering for the status views: the compact --statusline form and
// the field-per-line detail block used for both the list and single-task
// views. The output model lives in status_output.go.

// outputStatusline renders the compact one-line form for shell and agent
// statuslines. Its input is always the local task list — see status_overlay.go
// for why the overlay stops here.
func outputStatusline(tasks []*model.Task, w io.Writer) error {
	task := tasks[0]
	line := fmt.Sprintf("#%s %s", task.ID, task.Title)
	if len(tasks) > 1 {
		line += fmt.Sprintf(" (+%d more)", len(tasks)-1)
	}

	fmt.Fprintln(w, line)
	return nil
}

func outputStatusListText(outputs []statusOutput, w io.Writer) error {
	for i, out := range outputs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if err := outputStatusText(out, w); err != nil {
			return err
		}
	}
	return nil
}

func outputStatusText(out statusOutput, w io.Writer) error {
	r := getRenderer()

	fmt.Fprintf(w, "%s %s\n", formatLabel("Task:", r), formatTaskID(out.ID, r))
	fmt.Fprintf(w, "%s %s\n", formatLabel("Title:", r), out.Title)
	fmt.Fprintf(w, "%s %s\n", formatLabel("Status:", r), formatStatus(displayStatus(out), r))
	printStatusProvenance(w, out, r)
	printStatusOptionalField(w, "Priority", out.Priority, r)
	printStatusOptionalField(w, "Effort", out.Effort, r)
	if len(out.Tags) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Tags:", r), strings.Join(out.Tags, ", "))
	}
	if out.Owner != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Owner:", r), out.Owner)
	}
	if out.Parent != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Parent:", r), out.Parent)
	}
	if out.Created != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Created:", r), out.Created)
	}
	if len(out.Dependencies) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Dependencies:", r), strings.Join(out.Dependencies, ", "))
	}
	if out.Blocked != nil {
		if *out.Blocked {
			fmt.Fprintf(w, "%s Yes (blocked by: %s)\n", formatLabel("Blocked:", r), strings.Join(out.BlockedBy, ", "))
		} else {
			fmt.Fprintf(w, "%s No\n", formatLabel("Blocked:", r))
		}
	}
	if out.Group != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Group:", r), out.Group)
	}
	if len(out.Children) > 0 {
		fmt.Fprintf(w, "%s\n", formatLabel("Children:", r))
		writeChildrenTree(w, out.Children, "  ", r)
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel("File:", r), formatDim(out.FilePath, r))
	printWorktreeCopies(w, out.Worktrees, r)
	return nil
}

func writeChildrenTree(w io.Writer, children []statusChild, prefix string, r *lipgloss.Renderer) {
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├─"
		if isLast {
			connector = "└─"
		}
		fmt.Fprintf(w, "%s%s %s [%s] %s\n", prefix, connector,
			formatTaskID(child.ID, r), formatStatus(child.Status, r), child.Title)
		if len(child.Children) > 0 {
			childPrefix := prefix + "│  "
			if isLast {
				childPrefix = prefix + "   "
			}
			writeChildrenTree(w, child.Children, childPrefix, r)
		}
	}
}

func printStatusOptionalField(w io.Writer, label, value string, r *lipgloss.Renderer) {
	if value == "" {
		return
	}
	var colored string
	switch label {
	case "Priority":
		colored = formatPriority(value, r)
	case "Effort":
		colored = formatEffort(value, r)
	default:
		colored = value
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel(label+":", r), colored)
}
