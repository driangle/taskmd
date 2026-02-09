package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
)

// App is the root bubbletea model for the TUI shell.
type App struct {
	width         int
	height        int
	scanDir       string
	tasks         []*model.Task
	ready         bool
	showHelp      bool
	selectedIndex int
	scrollOffset  int
}

// New creates a new TUI app shell.
func New(scanDir string, tasks []*model.Task) App {
	return App{
		scanDir: scanDir,
		tasks:   tasks,
	}
}

func (m App) Init() tea.Cmd {
	return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
		case "j", "down":
			if len(m.tasks) > 0 && m.selectedIndex < len(m.tasks)-1 {
				m.selectedIndex++
			}
		case "k", "up":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "g":
			// Go to top
			m.selectedIndex = 0
			m.scrollOffset = 0
		case "G":
			// Go to bottom
			if len(m.tasks) > 0 {
				m.selectedIndex = len(m.tasks) - 1
			}
		}
	}
	return m, nil
}

func (m App) View() string {
	if !m.ready {
		return "Initializing..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := m.height - headerHeight - footerHeight

	content := m.renderContent(contentHeight)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m App) renderHeader() string {
	title := "taskmd"
	dir := m.scanDir
	text := fmt.Sprintf(" %s  %s", title, dir)

	return headerStyle.Width(m.width).Render(text)
}

func (m App) renderFooter() string {
	var text string
	if m.showHelp {
		text = " q: quit  ?: close help  j/k or ↓/↑: navigate  g/G: top/bottom"
	} else {
		text = " q: quit  ?: help  j/k or ↓/↑: navigate"
	}

	return footerStyle.Width(m.width).Render(text)
}

func (m App) renderContent(height int) string {
	if height < 0 {
		height = 0
	}

	// Handle empty state
	if len(m.tasks) == 0 {
		msg := "No tasks found in this directory.\n\nCreate a task file with frontmatter to get started!"
		return contentStyle.Width(m.width).Height(height).Render(helpStyle.Render(msg))
	}

	var lines []string

	// Calculate task counts for summary
	var pending, inProgress, completed, blocked int
	for _, t := range m.tasks {
		switch t.Status {
		case model.StatusPending:
			pending++
		case model.StatusInProgress:
			inProgress++
		case model.StatusCompleted:
			completed++
		case model.StatusBlocked:
			blocked++
		}
	}

	// Reserve 1 line for summary at bottom
	listHeight := height - 2
	if listHeight < 1 {
		listHeight = 1
	}

	// Adjust scroll offset to keep selected item visible
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	}
	if m.selectedIndex >= m.scrollOffset+listHeight {
		m.scrollOffset = m.selectedIndex - listHeight + 1
	}

	// Render visible tasks
	for i := m.scrollOffset; i < len(m.tasks) && i < m.scrollOffset+listHeight; i++ {
		task := m.tasks[i]
		lines = append(lines, m.renderTaskRow(task, i == m.selectedIndex))
	}

	// Pad to fill height
	for len(lines) < listHeight {
		lines = append(lines, "")
	}

	// Add summary line
	summary := fmt.Sprintf("%d tasks: %d pending, %d in-progress, %d completed",
		len(m.tasks), pending, inProgress, completed)
	if blocked > 0 {
		summary += fmt.Sprintf(", %d blocked", blocked)
	}
	lines = append(lines, "")
	lines = append(lines, summaryStyle.Render(summary))

	if m.showHelp {
		lines = append(lines,
			"",
			helpStyle.Render("Key Bindings:"),
			helpStyle.Render("  q / ctrl+c    Quit"),
			helpStyle.Render("  ? Toggle help"),
			helpStyle.Render("  j / ↓         Move down"),
			helpStyle.Render("  k / ↑         Move up"),
			helpStyle.Render("  g             Go to top"),
			helpStyle.Render("  G             Go to bottom"),
		)
	}

	body := strings.Join(lines, "\n")

	return contentStyle.Width(m.width).Height(height).Render(body)
}

func (m App) renderTaskRow(task *model.Task, selected bool) string {
	// Status indicator
	statusIcon := " "
	statusColor := lipgloss.Color("240")
	switch task.Status {
	case model.StatusPending:
		statusIcon = "○"
		statusColor = lipgloss.Color("yellow")
	case model.StatusInProgress:
		statusIcon = "◐"
		statusColor = lipgloss.Color("blue")
	case model.StatusCompleted:
		statusIcon = "●"
		statusColor = lipgloss.Color("green")
	case model.StatusBlocked:
		statusIcon = "✖"
		statusColor = lipgloss.Color("red")
	}

	// Format ID (fixed width)
	id := fmt.Sprintf("%-6s", task.ID)

	// Status indicator with color
	status := lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon)

	// Priority badge
	priority := ""
	if task.Priority != "" {
		priorityColor := lipgloss.Color("240")
		switch task.Priority {
		case model.PriorityCritical:
			priorityColor = lipgloss.Color("red")
		case model.PriorityHigh:
			priorityColor = lipgloss.Color("yellow")
		case model.PriorityMedium:
			priorityColor = lipgloss.Color("blue")
		case model.PriorityLow:
			priorityColor = lipgloss.Color("240")
		}
		priority = lipgloss.NewStyle().Foreground(priorityColor).Render(fmt.Sprintf("[%s]", task.Priority))
	}

	// Tags
	tags := ""
	if len(task.Tags) > 0 {
		tagStr := strings.Join(task.Tags, ", ")
		if len(tagStr) > 20 {
			tagStr = tagStr[:17] + "..."
		}
		tags = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("(%s)", tagStr))
	}

	// Build row
	var parts []string
	parts = append(parts, id, status, task.Title)
	if priority != "" {
		parts = append(parts, priority)
	}
	if tags != "" {
		parts = append(parts, tags)
	}

	row := strings.Join(parts, " ")

	// Apply selection style
	if selected {
		row = selectedRowStyle.Render(row)
	} else if task.Status == model.StatusCompleted {
		row = completedRowStyle.Render(row)
	}

	return row
}
