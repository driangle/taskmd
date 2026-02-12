package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/tui"
)

var (
	tuiFocus    string
	tuiFilter   string
	tuiGroupBy  string
	tuiReadonly bool
)

var tuiCmd = &cobra.Command{
	Use:        "tui",
	SuggestFor: []string{"ui", "interactive", "dashboard"},
	Short:      "Launch interactive TUI (Terminal User Interface)",
	Long:       "Launch an interactive terminal UI for browsing and managing tasks",
	Args:       cobra.MaximumNArgs(1),
	RunE:       runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	tuiCmd.Flags().StringVar(&tuiFocus, "focus", "", "Start with a specific task selected (task ID)")
	tuiCmd.Flags().StringVar(&tuiFilter, "filter", "", "Pre-apply a filter (e.g., 'status=pending', 'priority=high')")
	tuiCmd.Flags().StringVar(&tuiGroupBy, "group-by", "", "Group tasks by field (status, priority, group)")
	tuiCmd.Flags().BoolVar(&tuiReadonly, "readonly", false, "Launch in read-only mode (future: disable editing)")
}

func runTUI(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Create TUI options from flags
	opts := tui.Options{
		FocusTaskID: tuiFocus,
		Filter:      tuiFilter,
		GroupBy:     tuiGroupBy,
		ReadOnly:    tuiReadonly,
		Verbose:     flags.Verbose,
	}

	app := tui.NewWithOptions(scanDir, result.Tasks, opts)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}
