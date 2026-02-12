package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/watcher"
)

const (
	viewList   = 0
	viewDetail = 1
)

// fileChangeMsg signals that task files have changed on disk.
type fileChangeMsg struct{}

// tasksRefreshedMsg contains the newly scanned tasks after a file change.
type tasksRefreshedMsg struct {
	tasks []*model.Task
}

// hideRefreshIndicatorMsg signals to hide the refresh indicator.
type hideRefreshIndicatorMsg struct{}

// Options holds configuration options for the TUI.
type Options struct {
	FocusTaskID string
	Filter      string
	GroupBy     string
	ReadOnly    bool
	Verbose     bool
}

// App is the root bubbletea model for the TUI shell.
type App struct {
	width                int
	height               int
	scanDir              string
	tasks                []*model.Task
	ready                bool
	showHelp             bool
	selectedIndex        int
	scrollOffset         int
	searchMode           bool
	searchQuery          string
	groupBy              string
	readonly             bool
	viewMode             int
	detailScrollOffset   int
	renderedDetail       string
	watcher              *watcher.Watcher
	verbose              bool
	showRefreshIndicator bool
	fileChangeChan       chan struct{}
}

// New creates a new TUI app shell.
func New(scanDir string, tasks []*model.Task) App {
	return NewWithOptions(scanDir, tasks, Options{})
}

// NewWithOptions creates a new TUI app shell with configuration options.
func NewWithOptions(scanDir string, tasks []*model.Task, opts Options) App {
	app := App{
		scanDir:        scanDir,
		tasks:          tasks,
		groupBy:        opts.GroupBy,
		readonly:       opts.ReadOnly,
		verbose:        opts.Verbose,
		fileChangeChan: make(chan struct{}, 1),
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
	// Start the file watcher and begin listening for changes
	go m.startWatcher()
	return m.waitForFileChange()
}

// startWatcher initializes and runs the file watcher in a goroutine.
func (m App) startWatcher() {
	m.watcher = watcher.New(m.scanDir, func() {
		// Send notification through channel when files change
		select {
		case m.fileChangeChan <- struct{}{}:
		default:
			// Channel full, skip (already have a pending change)
		}
	}, 200*time.Millisecond)

	// Start watching (blocks until stopped)
	_ = m.watcher.Start()
}

// waitForFileChange creates a command that waits for file change notifications.
func (m App) waitForFileChange() tea.Cmd {
	return func() tea.Msg {
		<-m.fileChangeChan
		return fileChangeMsg{}
	}
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case fileChangeMsg:
		// Files changed, trigger re-scan
		return m, m.refreshTasks()

	case tasksRefreshedMsg:
		// New tasks have been scanned, update the model
		m.updateTasksPreservingState(msg.tasks)
		m.showRefreshIndicator = true
		return m, tea.Batch(
			m.waitForFileChange(), // Continue listening for changes
			m.hideRefreshIndicator(),
		)

	case hideRefreshIndicatorMsg:
		m.showRefreshIndicator = false
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			m.handleSearchKey(msg)
			return m, nil
		}
		if m.viewMode == viewDetail {
			return m.handleDetailKey(msg)
		}
		return m.handleListKey(msg)
	}
	return m, nil
}

// refreshTasks triggers a re-scan of task files.
func (m App) refreshTasks() tea.Cmd {
	return func() tea.Msg {
		taskScanner := scanner.NewScanner(m.scanDir, m.verbose)
		result, err := taskScanner.Scan()
		if err != nil {
			// On error, return current tasks (no change)
			return tasksRefreshedMsg{tasks: m.tasks}
		}
		return tasksRefreshedMsg{tasks: result.Tasks}
	}
}

// updateTasksPreservingState updates tasks while preserving selection and scroll.
func (m *App) updateTasksPreservingState(newTasks []*model.Task) {
	// Save currently selected task ID
	var selectedTaskID string
	filteredTasks := m.getFilteredTasks()
	if len(filteredTasks) > 0 && m.selectedIndex < len(filteredTasks) {
		selectedTaskID = filteredTasks[m.selectedIndex].ID
	}

	// Update tasks
	m.tasks = newTasks

	// Try to restore selection
	if selectedTaskID != "" {
		filteredTasks = m.getFilteredTasks()
		for i, task := range filteredTasks {
			if task.ID == selectedTaskID {
				m.selectedIndex = i
				return
			}
		}
	}

	// If we couldn't restore selection, clamp to valid range
	filteredTasks = m.getFilteredTasks()
	if m.selectedIndex >= len(filteredTasks) {
		m.selectedIndex = len(filteredTasks) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// hideRefreshIndicator returns a command that hides the refresh indicator after a delay.
func (m App) hideRefreshIndicator() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(_ time.Time) tea.Msg {
		return hideRefreshIndicatorMsg{}
	})
}

func (m *App) handleSearchKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEscape:
		m.searchMode = false
		m.searchQuery = ""
		m.selectedIndex = 0
		m.scrollOffset = 0
	case tea.KeyEnter:
		m.searchMode = false
	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.selectedIndex = 0
			m.scrollOffset = 0
		}
	case tea.KeyRunes:
		m.searchQuery += string(msg.Runes)
		m.selectedIndex = 0
		m.scrollOffset = 0
	}
}

func (m App) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m *App) navigateList(key string) {
	filteredTasks := m.getFilteredTasks()
	switch key {
	case "j", "down":
		if len(filteredTasks) > 0 && m.selectedIndex < len(filteredTasks)-1 {
			m.selectedIndex++
		}
	case "k", "up":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "g":
		m.selectedIndex = 0
		m.scrollOffset = 0
	case "G":
		if len(filteredTasks) > 0 {
			m.selectedIndex = len(filteredTasks) - 1
		}
	}
}

func (m App) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
	case "/":
		m.searchMode = true
		m.searchQuery = ""
	case "escape":
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
	default:
		m.navigateList(msg.String())
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

	if m.showRefreshIndicator {
		refreshIcon := lipgloss.NewStyle().
			Foreground(lipgloss.Color("green")).
			Render(" ⟳")
		text += refreshIcon
	}

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

type statusCounts struct {
	pending, inProgress, completed, blocked, cancelled int
}

func countTaskStatuses(tasks []*model.Task) statusCounts {
	var c statusCounts
	for _, t := range tasks {
		switch t.Status {
		case model.StatusPending:
			c.pending++
		case model.StatusInProgress:
			c.inProgress++
		case model.StatusCompleted:
			c.completed++
		case model.StatusBlocked:
			c.blocked++
		case model.StatusCancelled:
			c.cancelled++
		}
	}
	return c
}

func (m App) buildSummaryLine(filteredCount int, counts statusCounts) string {
	var summary string
	if m.searchQuery != "" {
		summary = fmt.Sprintf("Showing %d of %d tasks: %d pending, %d in-progress, %d completed",
			filteredCount, len(m.tasks), counts.pending, counts.inProgress, counts.completed)
	} else {
		summary = fmt.Sprintf("%d tasks: %d pending, %d in-progress, %d completed",
			len(m.tasks), counts.pending, counts.inProgress, counts.completed)
	}
	if counts.blocked > 0 {
		summary += fmt.Sprintf(", %d blocked", counts.blocked)
	}
	if counts.cancelled > 0 {
		summary += fmt.Sprintf(", %d cancelled", counts.cancelled)
	}
	return summary
}

func (m App) renderContent(height int) string {
	if height < 0 {
		height = 0
	}

	filteredTasks := m.getFilteredTasks()

	if len(m.tasks) == 0 {
		msg := "No tasks found in this directory.\n\nCreate a task file with frontmatter to get started!"
		return contentStyle.Width(m.width).Height(height).Render(helpStyle.Render(msg))
	}
	if len(filteredTasks) == 0 && m.searchQuery != "" {
		msg := fmt.Sprintf("No tasks match '%s'\n\nPress Esc to clear search.", m.searchQuery)
		return contentStyle.Width(m.width).Height(height).Render(helpStyle.Render(msg))
	}

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
	var lines []string
	for i := m.scrollOffset; i < len(filteredTasks) && i < m.scrollOffset+listHeight; i++ {
		lines = append(lines, m.renderTaskRow(filteredTasks[i], i == m.selectedIndex))
	}
	for len(lines) < listHeight {
		lines = append(lines, "")
	}

	counts := countTaskStatuses(m.tasks)
	lines = append(lines, "", summaryStyle.Render(m.buildSummaryLine(len(filteredTasks), counts)))

	if m.showHelp {
		lines = append(lines, "", helpStyle.Render("Key Bindings:"),
			helpStyle.Render("  q / ctrl+c    Quit"),
			helpStyle.Render("  ? Toggle help"),
			helpStyle.Render("  j / ↓         Move down"),
			helpStyle.Render("  k / ↑         Move up"),
			helpStyle.Render("  g             Go to top"),
			helpStyle.Render("  G             Go to bottom"),
		)
	}

	return contentStyle.Width(m.width).Height(height).Render(strings.Join(lines, "\n"))
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

func statusIconAndColor(status model.Status) (string, lipgloss.Color) {
	switch status {
	case model.StatusPending:
		return "○", lipgloss.Color("yellow")
	case model.StatusInProgress:
		return "◐", lipgloss.Color("blue")
	case model.StatusCompleted:
		return "●", lipgloss.Color("green")
	case model.StatusBlocked:
		return "✖", lipgloss.Color("red")
	case model.StatusCancelled:
		return "⊘", lipgloss.Color("240")
	default:
		return " ", lipgloss.Color("240")
	}
}

func priorityColor(p model.Priority) lipgloss.Color {
	switch p {
	case model.PriorityCritical:
		return lipgloss.Color("red")
	case model.PriorityHigh:
		return lipgloss.Color("yellow")
	case model.PriorityMedium:
		return lipgloss.Color("blue")
	default:
		return lipgloss.Color("240")
	}
}

func (m App) statusString(status model.Status) string {
	icon, color := statusIconAndColor(status)
	return lipgloss.NewStyle().Foreground(color).Render(icon + " " + string(status))
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
	icon, color := statusIconAndColor(task.Status)
	id := fmt.Sprintf("%-6s", task.ID)
	status := lipgloss.NewStyle().Foreground(color).Render(icon)

	parts := []string{id, status, task.Title}

	if task.Priority != "" {
		pColor := priorityColor(task.Priority)
		parts = append(parts, lipgloss.NewStyle().Foreground(pColor).Render(fmt.Sprintf("[%s]", task.Priority)))
	}
	if len(task.Tags) > 0 {
		tagStr := strings.Join(task.Tags, ", ")
		if len(tagStr) > 20 {
			tagStr = tagStr[:17] + "..."
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("(%s)", tagStr)))
	}

	row := strings.Join(parts, " ")
	if selected {
		row = selectedRowStyle.Render(row)
	} else if task.Status == model.StatusCompleted {
		row = completedRowStyle.Render(row)
	}
	return row
}
