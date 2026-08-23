package cli

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/next"
)

// Recommendation is re-exported from the shared package.
type Recommendation = next.Recommendation

// ScoreComponent is re-exported from the shared package.
type ScoreComponent = next.ScoreComponent

const nextDefaultColumns = "rank,id,title,priority,effort,file,reason"

var (
	nextFormat         string
	nextLimit          int
	nextFilters        []string
	nextQuickWins      bool
	nextCritical       bool
	nextScope          string
	nextExact          bool
	nextRoot           string
	nextPhase          string
	nextStrictPhases   bool
	nextStrictPriority bool
	nextColumns        string
	nextStatus         string
	nextPriority       string
	nextExplain        bool
)

var nextCmd = &cobra.Command{
	Use:        "next",
	SuggestFor: []string{"pick", "suggest", "what"},
	Short:      "Recommend what task to work on next",
	Long: `Next analyzes all tasks and recommends the best ones to work on next.

Tasks are scored based on priority, critical path position, downstream impact,
effort, and phase ordering (from .taskmd.yaml). Only actionable tasks
(pending or in-progress with all dependencies completed) are shown.

--strict-priority guarantees priority is the primary sort key: no
lower-priority task is ranked above an actionable higher-priority one
(critical > high > medium > low/unset), with the existing score breaking ties
within each tier. Unlike --priority <value> (which filters out non-matching
tasks), --strict-priority keeps all actionable tasks and only reorders them.
When combined with --strict-phases, phase is the primary sort key and priority
is secondary: earlier-phase tasks rank first, and within a phase, higher
priority ranks first.

--explain prints, beneath each recommendation, an itemized breakdown of every
scoring component (priority, phase, critical path, downstream, effort) with its
point value and a total equal to the task's score. Scaled bonuses show their
base and multiplier so the scaling is visible. The structured score_breakdown
is always present in json/yaml output.

Output formats: table (default), json, yaml

Examples:
  taskmd next
  taskmd next ./tasks
  taskmd next --explain
  taskmd next --limit 3
  taskmd next --priority high
  taskmd next --priority high --format json
  taskmd next --filter tag=cli
  taskmd next --quick-wins
  taskmd next --critical --limit 1
  taskmd next --scope web/graph
  taskmd next --scope web/graph --exact
  taskmd next --root 022
  taskmd next --phase v0.2
  taskmd next --strict-phases
  taskmd next --strict-priority
  taskmd next --strict-phases --strict-priority
  taskmd next --columns rank,id,title,reason`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNext,
}

func init() {
	rootCmd.AddCommand(nextCmd)

	nextCmd.Flags().StringVar(&nextFormat, "format", "table", "output format (table, json, yaml)")
	nextCmd.Flags().IntVar(&nextLimit, "limit", 5, "maximum number of recommendations")
	nextCmd.Flags().StringArrayVar(&nextFilters, "filter", []string{}, "filter tasks (e.g., --filter tag=cli)")
	nextCmd.Flags().BoolVar(&nextQuickWins, "quick-wins", false, "show only quick wins (tasks at the lowest configured effort)")
	nextCmd.Flags().BoolVar(&nextCritical, "critical", false, "show only critical path tasks")
	nextCmd.Flags().StringVar(&nextScope, "scope", "", "filter by scope; supports wildcards (e.g. cli, cli*)")
	nextCmd.Flags().BoolVar(&nextExact, "exact", false, "disable dependency expansion for --scope (only direct matches)")
	nextCmd.Flags().StringVar(&nextRoot, "root", "", "limit recommendations to tasks reachable from an ID (its upstream deps + subtasks)")
	nextCmd.Flags().StringVar(&nextPhase, "phase", "", "filter by phase")
	nextCmd.Flags().BoolVar(&nextStrictPhases, "strict-phases", false, "enforce strict phase ordering (earlier phases always rank first)")
	nextCmd.Flags().BoolVar(&nextStrictPriority, "strict-priority", false, "enforce strict priority ordering (higher priority always ranks first, score breaks ties within a tier; with --strict-phases, phase is primary and priority secondary)")
	nextCmd.Flags().StringVar(&nextColumns, "columns", nextDefaultColumns, "comma-separated columns for table output (e.g. rank,id,title,reason)")
	nextCmd.Flags().StringVar(&nextStatus, "status", "", "shortcut for --filter status=<value>")
	nextCmd.Flags().StringVar(&nextPriority, "priority", "", "shortcut for --filter priority=<value>")
	nextCmd.Flags().BoolVar(&nextExplain, "explain", false, "show an itemized score breakdown beneath each recommendation (table format)")
}

func runNext(cmd *cobra.Command, args []string) error {
	if allProjectsFlag {
		return runNextAllProjects()
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	allTasks, archivedTasks, overlay, err := scanActiveAndArchivedWithOverlay(scanDir, flags)
	if err != nil {
		return err
	}
	makeFilePathsRelative(allTasks, scanDir)

	recTasks := allTasks
	var worktreeExcluded map[string]string
	if overlay != nil {
		recTasks, worktreeExcluded = overlay.RecommendationInputs()
	}

	expandNextShortcutFilters()

	recs, err := next.Recommend(recTasks, next.Options{
		Limit:          nextLimit,
		Filters:        nextFilters,
		QuickWins:      nextQuickWins,
		Critical:       nextCritical,
		Scope:          nextScope,
		ScopeExact:     nextExact,
		Root:           nextRoot,
		ArchivedTasks:  archivedTasks,
		Phase:          nextPhase,
		PhaseOrder:     loadPhaseOrder(),
		StrictPhases:   nextStrictPhases,
		StrictPriority: nextStrictPriority,
		Efforts:        resolveEffortScale(),
		Excluded:       worktreeExcluded,
	})
	if err != nil {
		return err
	}

	return outputNext(recs, worktreeExcluded)
}

// outputNext renders recommendations in the requested format, appending the
// worktree exclusion section to --explain table output.
func outputNext(recs []Recommendation, worktreeExcluded map[string]string) error {
	switch nextFormat {
	case "json":
		return outputNextJSON(recs)
	case "yaml":
		return outputNextYAML(recs)
	case "table":
		if err := outputNextTable(recs); err != nil {
			return err
		}
		printWorktreeExclusions(worktreeExcluded)
		return nil
	default:
		return ValidateFormat(nextFormat, []string{"table", "json", "yaml"})
	}
}

// printWorktreeExclusions lists, under --explain, tasks the worktree overlay
// kept out of the recommendations and which sibling worktree excludes them.
func printWorktreeExclusions(excluded map[string]string) {
	if !nextExplain || len(excluded) == 0 {
		return
	}

	ids := make([]string, 0, len(excluded))
	for id := range excluded {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	r := getRenderer()
	fmt.Println()
	fmt.Println(formatLabel("Excluded by sibling worktrees:", r))
	for _, id := range ids {
		fmt.Printf("  %s  %s\n", formatTaskID(id, r), excluded[id])
	}
}

// ProjectRecommendation wraps a recommendation with project context.
type ProjectRecommendation struct {
	ProjectID string `json:"project" yaml:"project"`
	Recommendation
}

func runNextAllProjects() error {
	expandNextShortcutFilters()

	allRecs, err := collectAllProjectRecs()
	if err != nil {
		return err
	}

	// Sort by score descending, re-assign ranks
	sort.Slice(allRecs, func(i, j int) bool {
		return allRecs[i].Score > allRecs[j].Score
	})
	if nextLimit > 0 && nextLimit < len(allRecs) {
		allRecs = allRecs[:nextLimit]
	}
	for i := range allRecs {
		allRecs[i].Rank = i + 1
	}

	switch nextFormat {
	case "json":
		return WriteJSON(os.Stdout, allRecs)
	case "yaml":
		return WriteYAML(os.Stdout, allRecs)
	case "table":
		return outputNextAllProjectsTable(allRecs)
	default:
		return ValidateFormat(nextFormat, []string{"table", "json", "yaml"})
	}
}

// collectAllProjectRecs gathers recommendations from every registered project.
func collectAllProjectRecs() ([]ProjectRecommendation, error) {
	entries, err := LoadGlobalRegistry()
	if err != nil {
		return nil, fmt.Errorf("load global registry: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no projects registered in global registry")
	}

	var allRecs []ProjectRecommendation
	for _, entry := range entries {
		recs, recErr := recommendForProject(entry)
		if recErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping project %q: %v\n", entry.ID, recErr)
			continue
		}
		for _, rec := range recs {
			allRecs = append(allRecs, ProjectRecommendation{ProjectID: entry.ID, Recommendation: rec})
		}
	}
	return allRecs, nil
}

// recommendForProject scans a project and returns recommendations.
func recommendForProject(entry GlobalProjectEntry) ([]Recommendation, error) {
	tasks, err := scanProjectTasks(entry)
	if err != nil {
		return nil, err
	}
	return next.Recommend(tasks, next.Options{
		Limit:          0, // get all, we'll limit after merging
		Filters:        nextFilters,
		QuickWins:      nextQuickWins,
		Critical:       nextCritical,
		Scope:          nextScope,
		ScopeExact:     nextExact,
		Phase:          nextPhase,
		StrictPhases:   nextStrictPhases,
		StrictPriority: nextStrictPriority,
		Efforts:        resolveEffortScale(),
	})
}

func outputNextAllProjectsTable(recs []ProjectRecommendation) error {
	if len(recs) == 0 {
		fmt.Println("No actionable tasks found across projects.")
		return nil
	}

	columns, err := parseNextColumns(nextColumns)
	if err != nil {
		return err
	}
	columns = injectProjectColumn(columns)

	r := getRenderer()
	fmt.Println(formatLabel("Recommended tasks (all projects):", r))
	fmt.Println()

	tw := NewTableWriter()
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = nextColumnDisplayName(col)
	}
	tw.AddHeader(headers)
	tw.AddSeparator()

	for _, rec := range recs {
		plain, colored := projectRecRow(rec, columns, r)
		tw.AddRow(plain, colored)
	}

	tw.Flush(os.Stdout)
	return nil
}

// projectRecRow builds plain and colored column values for a project recommendation.
func projectRecRow(rec ProjectRecommendation, columns []string, r *lipgloss.Renderer) ([]string, []string) {
	plain := make([]string, len(columns))
	colored := make([]string, len(columns))
	for i, col := range columns {
		switch col {
		case "project":
			plain[i] = rec.ProjectID
			colored[i] = rec.ProjectID
		case "id":
			qualID := rec.ProjectID + ":" + rec.ID
			plain[i] = qualID
			colored[i] = formatTaskID(qualID, r)
		default:
			plain[i] = getNextColumnValue(&rec.Recommendation, col)
			colored[i] = colorizeNextColumn(&rec.Recommendation, col, r)
		}
	}
	return plain, colored
}

// expandNextShortcutFilters appends --status and --priority values to nextFilters.
func expandNextShortcutFilters() {
	if nextStatus != "" {
		nextFilters = append(nextFilters, "status="+nextStatus)
	}
	if nextPriority != "" {
		nextFilters = append(nextFilters, "priority="+nextPriority)
	}
}

// loadPhaseOrder reads phase identifiers from the viper config, preserving order.
// It uses the phase id when present, falling back to name for backwards compatibility.
func loadPhaseOrder() []string {
	raw := viper.Get("phases")
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var ids []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := m["id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func outputNextJSON(recs []Recommendation) error {
	return WriteJSON(os.Stdout, recs)
}

func outputNextYAML(recs []Recommendation) error {
	return WriteYAML(os.Stdout, recs)
}

// validNextColumns lists all valid column names for the next command.
var validNextColumns = []string{
	"rank", "id", "title", "status", "priority", "effort", "phase",
	"tags", "file", "deps", "reason", "score", "project",
}

func outputNextTable(recs []Recommendation) error {
	r := getRenderer()

	if len(recs) == 0 {
		if nextRoot != "" {
			fmt.Printf("No actionable tasks found reachable from %q.\n", nextRoot)
		} else if nextScope != "" {
			fmt.Printf("No actionable tasks found for scope %q.\n", nextScope)
		} else if nextQuickWins {
			fmt.Println("No quick wins available.")
		} else if nextCritical {
			fmt.Println("No critical path tasks available.")
		} else {
			fmt.Println("No actionable tasks found.")
		}
		return nil
	}

	label := "Recommended tasks:"
	if nextScope != "" {
		label = fmt.Sprintf("Recommended tasks (scope: %s):", nextScope)
	}
	if nextQuickWins {
		label = "Recommended quick wins:"
	}
	if nextCritical {
		label = "Recommended critical path tasks:"
	}
	fmt.Println(formatLabel(label, r))
	fmt.Println()

	if nextExplain {
		return outputNextExplain(recs, r)
	}

	columns, err := parseNextColumns(nextColumns)
	if err != nil {
		return err
	}

	tw := NewTableWriter()
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = nextColumnDisplayName(col)
	}
	tw.AddHeader(headers)
	tw.AddSeparator()

	for _, rec := range recs {
		plain := make([]string, len(columns))
		colored := make([]string, len(columns))
		for i, col := range columns {
			plain[i] = getNextColumnValue(&rec, col)
			colored[i] = colorizeNextColumn(&rec, col, r)
		}
		tw.AddRow(plain, colored)
	}

	tw.Flush(os.Stdout)
	return nil
}

// explainLabelMinWidth is the minimum width for the component label column
// in --explain output, so points stay aligned across short and long labels.
const explainLabelMinWidth = 30

// outputNextExplain renders an itemized score breakdown block beneath each
// recommendation. Each block lists every scoring component with its points
// (showing base × multiplier for scaled bonuses) and a total equal to the
// task's score.
func outputNextExplain(recs []Recommendation, r *lipgloss.Renderer) error {
	for i, rec := range recs {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("%d. %s  %s\n", rec.Rank, formatTaskID(rec.ID, r), rec.Title)

		width := explainLabelWidth(rec.ScoreBreakdown)
		for _, c := range rec.ScoreBreakdown {
			line := fmt.Sprintf("     %-*s %+4d", width, c.Label, c.Points)
			if c.Multiplier != 0 {
				line += "  " + formatDim(fmt.Sprintf("(%d × %.2f)", c.Base, c.Multiplier), r)
			}
			fmt.Println(line)
		}

		fmt.Printf("     %s\n", strings.Repeat("─", width+5))
		fmt.Printf("     %-*s %4d\n", width, "total", rec.Score)
	}
	return nil
}

// explainLabelWidth returns the label column width for a breakdown block:
// the longest component label (and "total"), floored at explainLabelMinWidth.
func explainLabelWidth(components []ScoreComponent) int {
	width := explainLabelMinWidth
	for _, c := range components {
		if len(c.Label) > width {
			width = len(c.Label)
		}
	}
	return width
}

// parseNextColumns splits the columns string and validates each column name.
func parseNextColumns(columnsStr string) ([]string, error) {
	parts := strings.Split(columnsStr, ",")
	columns := make([]string, 0, len(parts))
	for _, col := range parts {
		col = strings.ToLower(strings.TrimSpace(col))
		if col == "" {
			continue
		}
		if !slices.Contains(validNextColumns, col) {
			return nil, invalidValueError("column", col, validNextColumns)
		}
		columns = append(columns, col)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("no valid columns specified")
	}
	return columns, nil
}

// getNextColumnValue extracts the value for a column from a Recommendation.
func getNextColumnValue(rec *Recommendation, column string) string {
	switch column {
	case "rank":
		return fmt.Sprintf("%d", rec.Rank)
	case "id":
		return rec.ID
	case "title":
		return rec.Title
	case "status":
		return rec.Status
	case "priority":
		return rec.Priority
	case "effort":
		return rec.Effort
	case "file":
		return rec.FilePath
	case "reason":
		return strings.Join(rec.Reasons, ", ")
	case "score":
		return fmt.Sprintf("%d", rec.Score)
	case "phase", "tags", "deps":
		return ""
	default:
		return ""
	}
}

// nextColumnDisplayName returns the display header for a column.
// Special columns get custom names; others are title-cased.
func nextColumnDisplayName(col string) string {
	switch col {
	case "rank":
		return "#"
	case "id":
		return "ID"
	case "deps":
		return "Deps"
	case "project":
		return "Project"
	default:
		if len(col) == 0 {
			return col
		}
		return strings.ToUpper(col[:1]) + col[1:]
	}
}

// colorizeNextColumn returns the column value with color formatting applied.
func colorizeNextColumn(rec *Recommendation, column string, r *lipgloss.Renderer) string {
	value := getNextColumnValue(rec, column)
	switch column {
	case "id":
		return formatTaskID(value, r)
	case "priority":
		return formatPriority(value, r)
	case "effort":
		return formatEffort(value, r)
	case "file":
		return formatDim(value, r)
	default:
		return value
	}
}
