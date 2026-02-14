package cli

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/spf13/viper"
)

// forceColor is a test hook to bypass TTY detection.
var forceColor bool

// colorsEnabled checks if color output should be enabled based on flags, env, config, and TTY detection.
func colorsEnabled() bool {
	if noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if viper.GetBool("no-color") {
		return false
	}
	if forceColor {
		return true
	}
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return false
	}
	return true
}

// getRenderer returns a lipgloss renderer with the appropriate color profile.
func getRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	if colorsEnabled() {
		r.SetColorProfile(termenv.ANSI256)
	} else {
		r.SetColorProfile(termenv.Ascii)
	}
	return r
}

// getStatusColor returns the appropriate color style for a task status.
func getStatusColor(status string, r *lipgloss.Renderer) lipgloss.Style {
	switch strings.ToLower(status) {
	case "completed":
		return r.NewStyle().Foreground(lipgloss.Color("2")) // Green
	case "in-progress":
		return r.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "blocked":
		return r.NewStyle().Foreground(lipgloss.Color("1")) // Red
	default: // pending or other
		return r.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	}
}

// getPriorityColor returns the appropriate color style for a priority level.
func getPriorityColor(priority string, r *lipgloss.Renderer) lipgloss.Style {
	switch strings.ToLower(priority) {
	case "critical":
		return r.NewStyle().Foreground(lipgloss.Color("1")) // Red
	case "high":
		return r.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "medium":
		return r.NewStyle().Foreground(lipgloss.Color("4")) // Blue
	default: // low or other
		return r.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	}
}

// getEffortColor returns the appropriate color style for an effort level.
func getEffortColor(effort string, r *lipgloss.Renderer) lipgloss.Style {
	switch strings.ToLower(effort) {
	case "small":
		return r.NewStyle().Foreground(lipgloss.Color("2")) // Green
	case "medium":
		return r.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "large":
		return r.NewStyle().Foreground(lipgloss.Color("1")) // Red
	default:
		return r.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	}
}

// formatTaskID formats task IDs with a distinct color.
func formatTaskID(id string, r *lipgloss.Renderer) string {
	style := r.NewStyle().Foreground(lipgloss.Color("6")).Bold(true) // Cyan, bold
	return style.Render(id)
}

// formatTaskTitle formats task titles with status-based coloring.
func formatTaskTitle(title, status string, r *lipgloss.Renderer) string {
	style := getStatusColor(status, r)
	return style.Render(title)
}

// formatHeading colors a heading based on the group-by field and key value.
func formatHeading(key, groupBy string, r *lipgloss.Renderer) string {
	var style lipgloss.Style
	switch groupBy {
	case "status":
		style = getStatusColor(key, r)
	case "priority":
		style = getPriorityColor(key, r)
	case "effort":
		style = getEffortColor(key, r)
	default:
		style = r.NewStyle().Bold(true)
	}
	return style.Render(key)
}

// formatStatus formats status text with status-based color.
func formatStatus(status string, r *lipgloss.Renderer) string {
	style := getStatusColor(status, r)
	return style.Render(status)
}

// formatPriority formats priority text with priority-based color.
func formatPriority(priority string, r *lipgloss.Renderer) string {
	style := getPriorityColor(priority, r)
	return style.Render(priority)
}

// formatEffort formats effort text with effort-based color.
func formatEffort(effort string, r *lipgloss.Renderer) string {
	style := getEffortColor(effort, r)
	return style.Render(effort)
}

// formatSuccess formats a success message in green.
func formatSuccess(msg string, r *lipgloss.Renderer) string {
	return r.NewStyle().Foreground(lipgloss.Color("2")).Render(msg)
}

// formatError formats an error message in red.
func formatError(msg string, r *lipgloss.Renderer) string {
	return r.NewStyle().Foreground(lipgloss.Color("1")).Render(msg)
}

// formatWarning formats a warning message in yellow.
func formatWarning(msg string, r *lipgloss.Renderer) string {
	return r.NewStyle().Foreground(lipgloss.Color("3")).Render(msg)
}

// formatLabel formats a label in bold.
func formatLabel(label string, r *lipgloss.Renderer) string {
	return r.NewStyle().Bold(true).Render(label)
}

// formatDim formats text in gray (dimmed).
func formatDim(text string, r *lipgloss.Renderer) string {
	return r.NewStyle().Foreground(lipgloss.Color("8")).Render(text)
}
