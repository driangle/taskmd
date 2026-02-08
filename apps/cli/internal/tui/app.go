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
	width    int
	height   int
	scanDir  string
	tasks    []*model.Task
	ready    bool
	showHelp bool
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
		text = " q: quit  ?: close help  (more keys coming soon)"
	} else {
		text = " q: quit  ?: help"
	}

	return footerStyle.Width(m.width).Render(text)
}

func (m App) renderContent(height int) string {
	if height < 0 {
		height = 0
	}

	var lines []string

	// Task summary
	total := len(m.tasks)
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

	lines = append(lines,
		fmt.Sprintf("Tasks: %d total", total),
		"",
		fmt.Sprintf("  Pending:     %d", pending),
		fmt.Sprintf("  In Progress: %d", inProgress),
		fmt.Sprintf("  Completed:   %d", completed),
		fmt.Sprintf("  Blocked:     %d", blocked),
	)

	if m.showHelp {
		lines = append(lines,
			"",
			helpStyle.Render("Key Bindings:"),
			helpStyle.Render("  q / ctrl+c  Quit"),
			helpStyle.Render("  ?           Toggle help"),
		)
	}

	body := strings.Join(lines, "\n")

	return contentStyle.Width(m.width).Height(height).Render(body)
}
