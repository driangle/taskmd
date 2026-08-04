package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/validator"
)

var phasesFormat string

var phasesCmd = &cobra.Command{
	Use:        "phases",
	SuggestFor: []string{"milestones", "phase"},
	Short:      "List project phases with progress stats",
	Long: `Phases displays all configured phases from .taskmd.yaml with summary stats:
- Task count and completion percentage per phase
- Status breakdown per phase
- Due dates

Phases are configured in .taskmd.yaml under the "phases" key.

Tasks with no phase are grouped into a synthetic "(unassigned)" category
(id: unassigned), shown last and only when such tasks exist.

Examples:
  taskmd phases
  taskmd phases ./tasks
  taskmd phases --format json
  taskmd phases --format yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPhases,
}

func init() {
	rootCmd.AddCommand(phasesCmd)

	phasesCmd.Flags().StringVar(&phasesFormat, "format", "table", "output format (table, json, yaml)")
}

// PhaseSummary holds computed stats for a single phase.
type PhaseSummary struct {
	ID       string         `json:"id" yaml:"id"`
	Name     string         `json:"name" yaml:"name"`
	Tasks    int            `json:"tasks" yaml:"tasks"`
	Done     int            `json:"done" yaml:"done"`
	Progress string         `json:"progress" yaml:"progress"`
	Due      string         `json:"due,omitempty" yaml:"due,omitempty"`
	ByStatus map[string]int `json:"by_status" yaml:"by_status"`
}

func runPhases(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	allPhases := parsePhasesConfig(viper.Get("phases"))
	phases := filterValidPhases(allPhases)
	if len(phases) == 0 {
		fmt.Fprintln(os.Stderr, "No phases configured. Add phases to your .taskmd.yaml file:")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  phases:")
		fmt.Fprintln(os.Stderr, "    - id: mvp")
		fmt.Fprintln(os.Stderr, "      name: \"MVP\"")
		fmt.Fprintln(os.Stderr, "      due: 2026-06-01")
		return nil
	}

	summaries := computePhaseSummaries(phases, tasks)
	if unassigned := computeUnassignedSummary(tasks); unassigned.Tasks > 0 {
		summaries = append(summaries, unassigned)
	}
	warnOrphanedPhases(phases, tasks)

	switch phasesFormat {
	case "json":
		return WriteJSON(os.Stdout, summaries)
	case "yaml":
		return WriteYAML(os.Stdout, summaries)
	case "table":
		return outputPhasesTable(summaries)
	default:
		return ValidateFormat(phasesFormat, []string{"table", "json", "yaml"})
	}
}

// filterValidPhases returns only phases that have an id, warning about those that don't.
func filterValidPhases(phases []validator.PhaseConfig) []validator.PhaseConfig {
	valid := make([]validator.PhaseConfig, 0, len(phases))
	for _, p := range phases {
		if p.ID == "" {
			label := p.Name
			if label == "" {
				label = "(unnamed)"
			}
			fmt.Fprintf(os.Stderr, "Warning: phase %q is missing \"id\" field and will be skipped\n", label)
			continue
		}
		valid = append(valid, p)
	}
	return valid
}

// Stable identifiers for the synthetic category aggregating tasks with no phase.
// Parentheses aren't valid phase-id characters, so the display name can't
// collide with a real phase, and the id is safe for programmatic consumers.
const (
	unassignedPhaseID   = "unassigned"
	unassignedPhaseName = "(unassigned)"
)

func computePhaseSummaries(phases []validator.PhaseConfig, tasks []*model.Task) []PhaseSummary {
	summaries := make([]PhaseSummary, 0, len(phases))
	for _, phase := range phases {
		due := ""
		if !phase.Due.IsZero() {
			due = phase.Due.Format("2006-01-02")
		}
		summary := PhaseSummary{
			ID:       phase.ID,
			Name:     phase.Name,
			Due:      due,
			ByStatus: make(map[string]int),
		}
		for _, task := range tasks {
			if task.Phase != phase.ID {
				continue
			}
			countTask(&summary, task)
		}
		finalizeProgress(&summary)
		summaries = append(summaries, summary)
	}
	return summaries
}

// computeUnassignedSummary aggregates every task with no phase (task.Phase == "")
// into a synthetic PhaseSummary, using the same counting rules as configured
// phases. This is distinct from the orphaned-phase warning, which flags tasks
// pointing at a phase id that doesn't exist.
func computeUnassignedSummary(tasks []*model.Task) PhaseSummary {
	summary := PhaseSummary{
		ID:       unassignedPhaseID,
		Name:     unassignedPhaseName,
		ByStatus: make(map[string]int),
	}
	for _, task := range tasks {
		if task.Phase != "" {
			continue
		}
		countTask(&summary, task)
	}
	finalizeProgress(&summary)
	return summary
}

// countTask folds a single task into the summary's status breakdown and
// active/done counts. Cancelled tasks are recorded in ByStatus but excluded
// from the active task and completion counts.
func countTask(summary *PhaseSummary, task *model.Task) {
	summary.ByStatus[string(task.Status)]++
	if task.Status == model.StatusCancelled {
		return
	}
	summary.Tasks++
	if task.Status == model.StatusCompleted {
		summary.Done++
	}
}

// finalizeProgress sets the human-readable progress percentage from the
// summary's active task and done counts.
func finalizeProgress(summary *PhaseSummary) {
	if summary.Tasks > 0 {
		pct := float64(summary.Done) / float64(summary.Tasks) * 100
		summary.Progress = fmt.Sprintf("%.0f%%", pct)
	} else {
		summary.Progress = "0%"
	}
}

func warnOrphanedPhases(phases []validator.PhaseConfig, tasks []*model.Task) {
	knownIDs := make(map[string]bool, len(phases))
	for _, p := range phases {
		knownIDs[p.ID] = true
	}
	orphaned := make(map[string]int)
	for _, task := range tasks {
		if task.Phase != "" && !knownIDs[task.Phase] {
			orphaned[task.Phase]++
		}
	}
	sortedPhases := make([]string, 0, len(orphaned))
	for phase := range orphaned {
		sortedPhases = append(sortedPhases, phase)
	}
	sort.Strings(sortedPhases)
	for _, phase := range sortedPhases {
		fmt.Fprintf(os.Stderr, "Warning: %d task(s) reference undefined phase %q\n", orphaned[phase], phase)
	}
}

func outputPhasesTable(summaries []PhaseSummary) error {
	tw := NewTableWriter()
	tw.AddHeader([]string{"ID", "Name", "Tasks", "Done", "Progress", "Due"})
	tw.AddSeparator()
	for _, s := range summaries {
		due := s.Due
		if due == "" {
			due = "-"
		}
		row := []string{
			s.ID,
			s.Name,
			fmt.Sprintf("%d", s.Tasks),
			fmt.Sprintf("%d", s.Done),
			s.Progress,
			due,
		}
		tw.AddRow(row, row)
	}
	tw.Flush(os.Stdout)
	return nil
}
