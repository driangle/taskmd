package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/tracks"
)

var (
	tracksFormat  string
	tracksFilters []string
	tracksLimit   int
	tracksScope   string
)

var tracksCmd = &cobra.Command{
	Use:        "tracks [path]",
	SuggestFor: []string{"lanes", "parallel"},
	Short:      "Show parallel work tracks based on scope overlap",
	Long: `Tracks assigns actionable tasks to parallel tracks based on the "touches"
frontmatter field. Tasks sharing a scope are placed in separate tracks so they
can be worked on without merge conflicts.

Tasks without a "touches" field are shown as "flexible" and can be assigned
to any track.

Use --scope to focus on a single scope: only tasks touching that scope (and
their dependency-connected tasks) are shown as a single ordered track.

Scope definitions can be configured in .taskmd.yaml under the "scopes" key.
Unknown scopes produce warnings when scopes are configured.

Output formats: table (default), json, yaml

Examples:
  taskmd tracks
  taskmd tracks ./tasks
  taskmd tracks --format json
  taskmd tracks --filter tag=cli
  taskmd tracks --limit 3
  taskmd tracks --scope web/graph`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTracks,
}

func init() {
	rootCmd.AddCommand(tracksCmd)

	tracksCmd.Flags().StringVar(&tracksFormat, "format", "table", "output format (table, json, yaml)")
	tracksCmd.Flags().StringArrayVar(&tracksFilters, "filter", []string{}, "filter tasks (e.g., --filter tag=cli)")
	tracksCmd.Flags().IntVar(&tracksLimit, "limit", 0, "maximum number of tracks to show (0 = unlimited)")
	tracksCmd.Flags().StringVar(&tracksScope, "scope", "", "focus on a single scope; supports wildcards (e.g. web, web*)")
}

func runTracks(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	allTasks, archivedTasks, overlay, err := scanActiveAndArchivedWithOverlay(scanDir, flags)
	if err != nil {
		return err
	}
	if overlay != nil {
		allTasks = overlay.effectiveTasks()
	}
	makeFilePathsRelative(allTasks, scanDir)

	knownScopes := loadScopesConfig()

	result, err := tracks.Assign(allTasks, tracks.Options{
		Filters:       tracksFilters,
		KnownScopes:   knownScopes,
		ArchivedTasks: archivedTasks,
		Scope:         tracksScope,
		Efforts:       resolveEffortScale(),
	})
	if err != nil {
		return err
	}

	if tracksLimit > 0 && len(result.Tracks) > tracksLimit {
		result.Tracks = result.Tracks[:tracksLimit]
	}

	switch tracksFormat {
	case "json":
		return WriteJSON(os.Stdout, result)
	case "yaml":
		return WriteYAML(os.Stdout, result)
	case "table":
		return outputTracksTable(result)
	default:
		return ValidateFormat(tracksFormat, []string{"table", "json", "yaml"})
	}
}

func outputTracksTable(result *tracks.Result) error {
	r := getRenderer()

	if len(result.Tracks) == 0 && len(result.Flexible) == 0 {
		fmt.Println("No actionable tasks found.")
		return nil
	}

	for _, warn := range result.Warnings {
		fmt.Println(formatWarning("Warning: "+warn, r))
	}
	if len(result.Warnings) > 0 {
		fmt.Println()
	}

	for _, track := range result.Tracks {
		var header string
		if len(track.Scopes) > 0 {
			header = fmt.Sprintf("Track %d (%s):", track.ID, strings.Join(track.Scopes, ", "))
		} else {
			header = fmt.Sprintf("Track %d:", track.ID)
		}
		fmt.Println(formatLabel(header, r))

		tw := NewTableWriter()
		for i, task := range track.Tasks {
			num := fmt.Sprintf("  %d.", i+1)
			plain := []string{num, task.ID, task.Title, task.Priority}
			colored := []string{num, formatTaskID(task.ID, r), task.Title, formatPriority(task.Priority, r)}
			tw.AddRow(plain, colored)
		}
		tw.Flush(os.Stdout)
		fmt.Println()
	}

	if len(result.Flexible) > 0 {
		fmt.Println(formatLabel("Flexible (no declared overlaps):", r))

		tw := NewTableWriter()
		for i, task := range result.Flexible {
			num := fmt.Sprintf("  %d.", i+1)
			plain := []string{num, task.ID, task.Title, task.Priority}
			colored := []string{num, formatTaskID(task.ID, r), task.Title, formatPriority(task.Priority, r)}
			tw.AddRow(plain, colored)
		}
		tw.Flush(os.Stdout)
	}

	return nil
}
