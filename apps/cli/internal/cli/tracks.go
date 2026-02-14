package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/tracks"
)

var (
	tracksFormat  string
	tracksFilters []string
	tracksLimit   int
)

var tracksCmd = &cobra.Command{
	Use:   "tracks [path]",
	Short: "Show parallel work tracks based on scope overlap",
	Long: `Tracks assigns actionable tasks to parallel tracks based on the "touches"
frontmatter field. Tasks sharing a scope are placed in separate tracks so they
can be worked on without merge conflicts.

Tasks without a "touches" field are shown as "flexible" and can be assigned
to any track.

Scope definitions can be configured in .taskmd.yaml under the "scopes" key.
Unknown scopes produce warnings when scopes are configured.

Output formats: table (default), json, yaml

Examples:
  taskmd tracks
  taskmd tracks ./tasks
  taskmd tracks --format json
  taskmd tracks --filter tag=cli
  taskmd tracks --limit 3`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTracks,
}

func init() {
	rootCmd.AddCommand(tracksCmd)

	tracksCmd.Flags().StringVar(&tracksFormat, "format", "table", "output format (table, json, yaml)")
	tracksCmd.Flags().StringArrayVar(&tracksFilters, "filter", []string{}, "filter tasks (e.g., --filter tag=cli)")
	tracksCmd.Flags().IntVar(&tracksLimit, "limit", 0, "maximum number of tracks to show (0 = unlimited)")
}

func runTracks(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	scanResult, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	allTasks := scanResult.Tasks
	makeFilePathsRelative(allTasks, scanDir)

	knownScopes := loadScopesConfig()

	result, err := tracks.Assign(allTasks, tracks.Options{
		Filters:     tracksFilters,
		KnownScopes: knownScopes,
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

func loadScopesConfig() map[string]bool {
	raw := viper.Get("scopes")
	if raw == nil {
		return nil
	}

	scopeMap, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	known := make(map[string]bool, len(scopeMap))
	for name := range scopeMap {
		known[name] = true
	}
	return known
}

func outputTracksTable(result *tracks.Result) error {
	r := getRenderer()

	if len(result.Tracks) == 0 && len(result.Flexible) == 0 {
		fmt.Println("No actionable tasks found.")
		return nil
	}

	// Print warnings.
	for _, w := range result.Warnings {
		fmt.Println(formatWarning("Warning: "+w, r))
	}
	if len(result.Warnings) > 0 {
		fmt.Println()
	}

	for _, track := range result.Tracks {
		scopeLabel := strings.Join(track.Scopes, ", ")
		header := fmt.Sprintf("Track %d (%s):", track.ID, scopeLabel)
		fmt.Println(formatLabel(header, r))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for i, task := range track.Tasks {
			fmt.Fprintf(w, "  %d.\t%s\t%s\t%s\n",
				i+1,
				formatTaskID(task.ID, r),
				task.Title,
				formatPriority(task.Priority, r),
			)
		}
		w.Flush()
		fmt.Println()
	}

	if len(result.Flexible) > 0 {
		fmt.Println(formatLabel("Flexible (no declared overlaps):", r))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for i, task := range result.Flexible {
			fmt.Fprintf(w, "  %d.\t%s\t%s\t%s\n",
				i+1,
				formatTaskID(task.ID, r),
				task.Title,
				formatPriority(task.Priority, r),
			)
		}
		w.Flush()
	}

	return nil
}
