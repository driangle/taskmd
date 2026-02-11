package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
)

const (
	viewList   = 0
	viewDetail = 1
)

// Options holds configuration options for the TUI.
type Options struct {
	FocusTaskID string
	Filter      string
	GroupBy     string
	ReadOnly    bool
}

// App is the root bubbletea model for the TUI shell.
type App struct {
	width              int
	height             int
	scanDir            string
	tasks              []*model.Task
	ready              bool
	showHelp           bool
	selectedIndex      int
	scrollOffset       int
	searchMode         bool
	searchQuery        string
	groupBy            string
	readonly           bool
	viewMode           int
	detailScrollOffset int
	renderedDetail     string
}

// New creates a new TUI app shell.
func New(scanDir string, tasks []*model.Task) App {
	return NewWithOptions(scanDir, tasks, Options{})
}

// NewWithOptions creates a new TUI app shell with configuration options.
func NewWithOptions(scanDir string, tasks []*model.Task, opts Options) App {
	app := App{
		scanDir:  scanDir,
		tasks:    tasks,
		groupBy:  opts.GroupBy,
		readonly: opts.ReadOnly,
	}

	// Apply initial filter if provided
	if opts.Filter != "" {
		app.searchQuery = parseFilter(opts.Filter)
	}

	// Find and focus on specific task if requested
	if opts.FocusTaskID != "" {
		filteredTasks := app.getFilteredTasks()
		for i, task := range filteredTasks {
			if task.ID == opts.FocusTaskID {
				app.selectedIndex = i
				break
			}
		}
	}

	return app
}

// parseFilter converts a filter expression like "status=pending" to a search query.
// For now, this is simplified - we just extract the value after '='.
// Future: support more complex filter expressions.
func parseFilter(filter string) string {
	parts := strings.Split(filter, "=")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return filter
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
		// Handle search mode input
		if m.searchMode {
			switch msg.Type {
			case tea.KeyEscape:
				// Exit search mode and clear query
				m.searchMode = false
				m.searchQuery = ""
				m.selectedIndex = 0
				m.scrollOffset = 0
			case tea.KeyEnter:
				// Exit search mode but keep query
				m.searchMode = false
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.selectedIndex = 0
					m.scrollOffset = 0
				}
			case tea.KeyRunes:
				// Add character to search query
				m.searchQuery += string(msg.Runes)
				m.selectedIndex = 0
				m.scrollOffset = 0
			}
			return m, nil
		}

		// Detail view keybindings
		if m.viewMode == viewDetail {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc", "escape", "backspace":
				m.viewMode = viewList
			case "j", "down":
				m.detailScrollOffset++
			case "k", "up":
				if m.detailScrollOffset > 0 {
					m.detailScrollOffset--
				}
			case "g":
				m.detailScrollOffset = 0
			case "G":
				lines := strings.Split(m.renderedDetail, "\n")
				if len(lines) > 0 {
					m.detailScrollOffset = len(lines) - 1
				}
			}
			return m, nil
		}

		// Normal mode keybindings
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
		case "/":
			// Enter search mode
			m.searchMode = true
			m.searchQuery = ""
		case "escape":
			// Clear search query if active
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.selectedIndex = 0
				m.scrollOffset = 0
			}
		case "enter":
			filteredTasks := m.getFilteredTasks()
			if len(filteredTasks) > 0 && m.selectedIndex < len(filteredTasks) {
				m.viewMode = viewDetail
				m.detailScrollOffset = 0
				m.renderedDetail = m.buildDetailContent(filteredTasks[m.selectedIndex])
			}
		case "j", "down":
			filteredTasks := m.getFilteredTasks()
			if len(filteredTasks) > 0 && m.selectedIndex < len(filteredTasks)-1 {
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
			filteredTasks := m.getFilteredTasks()
			if len(filteredTasks) > 0 {
				m.selectedIndex = len(filteredTasks) - 1
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

	var content string
	switch m.viewMode {
	case viewDetail:
		content = m.renderDetailView(contentHeight)
	default:
		content = m.renderContent(contentHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m App) getFilteredTasks() []*model.Task {
	if m.searchQuery == "" {
		return m.tasks
	}

	query := strings.ToLower(m.searchQuery)
	var filtered []*model.Task

	for _, task := range m.tasks {
		// Search in ID and title
		if strings.Contains(strings.ToLower(task.ID), query) ||
			strings.Contains(strings.ToLower(task.Title), query) {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

func (m App) renderHeader() string {
	title := "taskmd"
	dir := m.scanDir
	text := fmt.Sprintf(" %s  %s", title, dir)

	return headerStyle.Width(m.width).Render(text)
}

func (m App) renderFooter() string {
	var text string
	if m.viewMode == viewDetail {
		text = " Esc: back  j/k or ↓/↑: scroll  g/G: top/bottom  q: quit"
	} else if m.searchMode {
		text = fmt.Sprintf(" Search: %s_  (Esc: cancel, Enter: apply)", m.searchQuery)
	} else if m.showHelp {
		text = " q: quit  ?: close help  /: search  j/k or ↓/↑: navigate  g/G: top/bottom"
	} else if m.searchQuery != "" {
		text = fmt.Sprintf(" q: quit  ?: help  /: search  Esc: clear filter [%s]", m.searchQuery)
	} else {
		text = " q: quit  ?: help  /: search  j/k or ↓/↑: navigate"
	}

	return footerStyle.Width(m.width).Render(text)
}

func (m App) renderContent(height int) string {
	if height < 0 {
		height = 0
	}

	// Get filtered tasks based on search query
	filteredTasks := m.getFilteredTasks()

	// Handle empty state
	if len(m.tasks) == 0 {
		msg := "No tasks found in this directory.\n\nCreate a task file with frontmatter to get started!"
		return contentStyle.Width(m.width).Height(height).Render(helpStyle.Render(msg))
	}

	// Handle no search results
	if len(filteredTasks) == 0 && m.searchQuery != "" {
		msg := fmt.Sprintf("No tasks match '%s'\n\nPress Esc to clear search.", m.searchQuery)
		return contentStyle.Width(m.width).Height(height).Render(helpStyle.Render(msg))
	}

	var lines []string

	// Calculate task counts for summary (from all tasks)
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

	// Render visible tasks from filtered list
	for i := m.scrollOffset; i < len(filteredTasks) && i < m.scrollOffset+listHeight; i++ {
		task := filteredTasks[i]
		lines = append(lines, m.renderTaskRow(task, i == m.selectedIndex))
	}

	// Pad to fill height
	for len(lines) < listHeight {
		lines = append(lines, "")
	}

	// Add summary line
	var summary string
	if m.searchQuery != "" {
		summary = fmt.Sprintf("Showing %d of %d tasks: %d pending, %d in-progress, %d completed",
			len(filteredTasks), len(m.tasks), pending, inProgress, completed)
	} else {
		summary = fmt.Sprintf("%d tasks: %d pending, %d in-progress, %d completed",
			len(m.tasks), pending, inProgress, completed)
	}
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

func (m App) buildDetailContent(task *model.Task) string {
	var sb strings.Builder

	// Title
	sb.WriteString(detailTitleStyle.Render(task.Title))
	sb.WriteString("\n\n")

	// Metadata
	sb.WriteString(detailLabelStyle.Render("ID:       ") + detailValueStyle.Render(task.ID) + "\n")
	sb.WriteString(detailLabelStyle.Render("Status:   ") + m.statusString(task.Status) + "\n")

	if task.Priority != "" {
		sb.WriteString(detailLabelStyle.Render("Priority: ") + detailValueStyle.Render(string(task.Priority)) + "\n")
	}
	if task.Effort != "" {
		sb.WriteString(detailLabelStyle.Render("Effort:   ") + detailValueStyle.Render(string(task.Effort)) + "\n")
	}
	if len(task.Tags) > 0 {
		sb.WriteString(detailLabelStyle.Render("Tags:     ") + detailValueStyle.Render(strings.Join(task.Tags, ", ")) + "\n")
	}
	if len(task.Dependencies) > 0 {
		sb.WriteString(detailLabelStyle.Render("Depends:  ") + detailValueStyle.Render(strings.Join(task.Dependencies, ", ")) + "\n")
	}
	if !task.Created.IsZero() {
		sb.WriteString(detailLabelStyle.Render("Created:  ") + detailValueStyle.Render(task.Created.Format("2006-01-02")) + "\n")
	}

	// Separator
	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", 40)) + "\n\n")

	// Body rendered as markdown
	if task.Body != "" {
		width := m.width - 6 // Account for content padding
		if width < 40 {
			width = 40
		}
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			rendered, err := renderer.Render(task.Body)
			if err == nil {
				sb.WriteString(rendered)
			} else {
				sb.WriteString(task.Body)
			}
		} else {
			sb.WriteString(task.Body)
		}
	} else {
		sb.WriteString(helpStyle.Render("(no description)"))
	}

	// File path
	if task.FilePath != "" {
		sb.WriteString("\n" + detailFilePathStyle.Render(task.FilePath))
	}

	return sb.String()
}

func (m App) statusString(status model.Status) string {
	icon := " "
	color := lipgloss.Color("240")
	switch status {
	case model.StatusPending:
		icon = "○"
		color = lipgloss.Color("yellow")
	case model.StatusInProgress:
		icon = "◐"
		color = lipgloss.Color("blue")
	case model.StatusCompleted:
		icon = "●"
		color = lipgloss.Color("green")
	case model.StatusBlocked:
		icon = "✖"
		color = lipgloss.Color("red")
	}
	return lipgloss.NewStyle().Foreground(color).Render(icon+" "+string(status))
}

func (m App) renderDetailView(height int) string {
	if height < 0 {
		height = 0
	}

	lines := strings.Split(m.renderedDetail, "\n")

	// Clamp scroll offset
	maxOffset := len(lines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.detailScrollOffset > maxOffset {
		m.detailScrollOffset = maxOffset
	}

	// Slice visible lines
	start := m.detailScrollOffset
	end := start + height
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[start:end]

	// Pad to fill height
	for len(visible) < height {
		visible = append(visible, "")
	}

	body := strings.Join(visible, "\n")
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
