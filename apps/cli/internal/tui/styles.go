package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	completedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	summaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("62"))

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("255"))

	detailFilePathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true)
)
