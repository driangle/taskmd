package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI (Terminal User Interface)",
	Long:  "Launch an interactive terminal UI for browsing and managing tasks",
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

type tuiModel struct {
	ready bool
}

func initialTUIModel() tuiModel {
	return tuiModel{
		ready: true,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	if !m.ready {
		return "Initializing..."
	}
	return "taskmd - Markdown Task Tracker\n\nPress 'q' to quit.\n\n(TUI implementation coming soon...)"
}

func runTUI(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(initialTUIModel())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}
