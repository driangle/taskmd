package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
	"github.com/driangle/md-task-tracker/apps/cli/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI (Terminal User Interface)",
	Long:  "Launch an interactive terminal UI for browsing and managing tasks",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	app := tui.New(scanDir, result.Tasks)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}
