package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/driangle/taskmd/sdk/go/model"
)

// runListOverlay renders list output through the worktree overlay: filters and
// sorting see effective status, and rows carry sibling-worktree provenance.
func runListOverlay(overlay *worktreeOverlay) error {
	// Filter and sort shallow copies with the effective status substituted
	// (so --status matches the merged view), then map the surviving copies
	// back to their overlay tasks — the same pointer-identity pattern as
	// runListAllProjects.
	copies := make([]*model.Task, len(overlay.Tasks))
	index := make(map[*model.Task]*OverlayTask, len(overlay.Tasks))
	for i, ot := range overlay.Tasks {
		effective := *ot.Task
		effective.Status = ot.EffectiveStatus
		copies[i] = &effective
		index[&effective] = ot
	}

	filtered, err := applyListFiltersAndSort(copies)
	if err != nil {
		return err
	}

	display := make([]*OverlayTask, 0, len(filtered))
	for _, t := range filtered {
		display = append(display, index[t])
	}

	switch listFormat {
	case "json":
		if len(display) == 0 {
			return WriteJSON(os.Stdout, []*OverlayTask{})
		}
		return WriteJSON(os.Stdout, display)
	case "yaml":
		return WriteYAML(os.Stdout, display)
	case "table":
		return outputOverlayTable(display, listColumns)
	default:
		return ValidateFormat(listFormat, []string{"table", "json", "yaml"})
	}
}

// outputOverlayTable renders the list table with effective statuses. The
// WORKTREE column appears only when at least one displayed task carries
// sibling provenance.
func outputOverlayTable(display []*OverlayTask, columnsStr string) error {
	if len(display) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}
	if displayAnnotated(display) {
		columns = injectWorktreeColumn(columns)
	}

	r := getRenderer()
	tw := NewTableWriter()
	tw.AddHeader(columns)
	tw.AddSeparator()

	for _, ot := range display {
		plain := make([]string, len(columns))
		colored := make([]string, len(columns))
		for i, col := range columns {
			plain[i] = overlayColumnValue(ot, col)
			colored[i] = colorizeOverlayColumn(ot, col, r)
		}
		tw.AddRow(plain, colored)
	}

	tw.Flush(os.Stdout)
	return nil
}

// displayAnnotated reports whether any displayed task carries worktree
// provenance (a remote winner or a sibling-only task).
func displayAnnotated(display []*OverlayTask) bool {
	for _, ot := range display {
		if ot.Worktree != "" || ot.RemoteOnly {
			return true
		}
	}
	return false
}

// injectWorktreeColumn adds a "worktree" column after "status" (or at the end
// when there is no status column) unless it is already present.
func injectWorktreeColumn(columns []string) []string {
	insertIdx := len(columns)
	for i, col := range columns {
		if col == "worktree" {
			return columns
		}
		if col == "status" {
			insertIdx = i + 1
		}
	}
	result := make([]string, 0, len(columns)+1)
	result = append(result, columns[:insertIdx]...)
	result = append(result, "worktree")
	return append(result, columns[insertIdx:]...)
}

// overlayColumnValue extracts a column value from an overlay task: status
// shows the effective status, worktree shows provenance (sibling-only tasks
// are marked with a trailing *), everything else comes from the base copy.
func overlayColumnValue(ot *OverlayTask, column string) string {
	switch column {
	case "status":
		return string(ot.EffectiveStatus)
	case "worktree":
		cell := ot.Worktree
		if ot.RemoteOnly {
			cell += "*"
		}
		return cell
	default:
		return getColumnValue(ot.Task, column)
	}
}

// colorizeOverlayColumn returns the overlay column value with color applied.
func colorizeOverlayColumn(ot *OverlayTask, column string, r *lipgloss.Renderer) string {
	value := overlayColumnValue(ot, column)
	switch column {
	case "id":
		return formatTaskID(value, r)
	case "status":
		return formatStatus(value, r)
	case "priority":
		return formatPriority(value, r)
	case "effort":
		return formatEffort(value, r)
	default:
		return value
	}
}
