package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/feed"
)

// feedFuzzyThreshold is the fuzzy-match sensitivity used when resolving a
// task-id positional (mirrors the `get` command default).
const feedFuzzyThreshold = 0.6

var (
	feedFormat string
	feedLimit  int
	feedSince  string
	feedScope  string
	feedSource string
	feedField  string
)

// gitLogFunc is the function used to run git log.
// Override in tests to avoid running actual git commands.
var gitLogFunc = runGitLog

// gitShowFunc is the function used to run git show.
// Override in tests to avoid running actual git commands.
var gitShowFunc = runGitShow

var feedCmd = &cobra.Command{
	Use:        "feed [task-id]",
	SuggestFor: []string{"activity", "log", "history"},
	Short:      "Show a chronological activity feed of task changes",
	Long: `Show a chronological activity feed of recent changes to task files.

Uses git log to detect task creation, modification, and renames,
presenting them as a time-ordered feed.

Pass a task id to scope the feed to a single task's history — a timeline
of its status transitions (including the initial "created" event). Use
--field to track a different frontmatter field (e.g. priority).

Examples:
  taskmd feed
  taskmd feed --since 7d
  taskmd feed --limit 10
  taskmd feed --scope cli
  taskmd feed --format json
  taskmd feed --source worklog
  taskmd feed cli-049                # status timeline for one task
  taskmd feed cli-049 --field priority`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFeed,
}

func init() {
	rootCmd.AddCommand(feedCmd)

	feedCmd.Flags().StringVar(&feedFormat, "format", "text", "output format (text, json)")
	feedCmd.Flags().IntVar(&feedLimit, "limit", 20, "maximum number of entries to show")
	feedCmd.Flags().StringVar(&feedSince, "since", "", "show changes since (e.g. 2d, 1w, 2026-02-28)")
	feedCmd.Flags().StringVar(&feedScope, "scope", "", "filter to a tasks subdirectory; supports wildcards (e.g. cli, cli*)")
	feedCmd.Flags().StringVar(&feedSource, "source", "all", "filter by event source (all, git, worklog)")
	feedCmd.Flags().StringVar(&feedField, "field", "status", "in single-task mode, the frontmatter field whose transitions to show")
}

func runFeed(_ *cobra.Command, args []string) error {
	if err := ValidateFormat(feedFormat, []string{"text", "json"}); err != nil {
		return err
	}
	if err := validateFeedSource(feedSource); err != nil {
		return err
	}

	flags := GetGlobalFlags()

	var taskFile, taskID string
	if len(args) == 1 {
		tf, id, err := resolveFeedTaskFile(args[0])
		if err != nil {
			return err
		}
		taskFile, taskID = tf, id
	}

	entries, err := feed.Query(feed.Options{
		TasksDir:  flags.TaskDir,
		Limit:     feedLimit,
		Since:     feedSince,
		Scope:     feedScope,
		Source:    feedSource,
		TaskFile:  taskFile,
		Verbose:   flags.Verbose,
		GitLogFn:  gitLogFunc,
		GitShowFn: gitShowFunc,
	})
	if err != nil {
		return fmt.Errorf("failed to read git history (is this a git repository?): %w", err)
	}

	if taskFile != "" {
		entries = filterEntriesByField(entries, feedField)
	}

	return writeFeedOutput(entries, taskID)
}

func validateFeedSource(source string) error {
	validSources := map[string]bool{"all": true, "git": true, "worklog": true}
	if !validSources[source] {
		return fmt.Errorf("unsupported source: %q (supported: all, git, worklog)", source)
	}
	return nil
}

// resolveFeedTaskFile resolves a task-id/query to its file path (relative to
// the current directory) and task ID, reusing the same matching as `get`.
func resolveFeedTaskFile(query string) (string, string, error) {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return "", "", err
	}
	makeFilePathsRelative(tasks, scanDir)

	task, err := resolveTask(query, tasks, false, feedFuzzyThreshold)
	if err != nil {
		return "", "", err
	}

	return filepath.Join(scanDir, task.FilePath), task.ID, nil
}

func writeFeedOutput(entries []feed.FeedEntry, taskID string) error {
	if len(entries) == 0 {
		return writeEmptyFeed(taskID)
	}

	switch feedFormat {
	case "json":
		return WriteJSON(os.Stdout, entries)
	default:
		return writeFeedText(entries)
	}
}

func writeEmptyFeed(taskID string) error {
	if feedFormat != "text" {
		fmt.Print("[]\n")
		return nil
	}
	if taskID != "" {
		fmt.Printf("No %s changes found for task %s.\n", feedField, taskID)
	} else {
		fmt.Println("No recent task changes.")
	}
	return nil
}

// filterEntriesByField narrows single-task entries to changes in one field.
// Git entries keep structural events (created/renamed) plus modifications that
// touched the requested field; worklog entries pass through unchanged.
func filterEntriesByField(entries []feed.FeedEntry, field string) []feed.FeedEntry {
	var out []feed.FeedEntry
	for _, e := range entries {
		if e.Source == "worklog" {
			out = append(out, e)
			continue
		}
		if files := filterFilesByField(e.Files, field); len(files) > 0 {
			e.Files = files
			out = append(out, e)
		}
	}
	return out
}

func filterFilesByField(files []feed.FileChange, field string) []feed.FileChange {
	var out []feed.FileChange
	for _, f := range files {
		changes := pickFieldChanges(f.FieldChanges, field)
		structural := f.Status == "created" || f.Status == "renamed"
		if !structural && len(changes) == 0 {
			continue
		}
		f.FieldChanges = changes
		f.SubtaskChanges = nil
		out = append(out, f)
	}
	return out
}

func pickFieldChanges(changes []feed.FieldChange, field string) []feed.FieldChange {
	var out []feed.FieldChange
	for _, c := range changes {
		if c.Field == field {
			out = append(out, c)
		}
	}
	return out
}

func runGitLog(_ string, args []string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func runGitShow(hash, path string) (string, error) {
	cmd := exec.Command("git", "show", hash+":"+path)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func writeFeedText(entries []feed.FeedEntry) error {
	r := getRenderer()

	fmt.Println(formatDim("Recent task activity", r))
	fmt.Println()

	for i, entry := range entries {
		if i > 0 {
			fmt.Println()
		}

		if entry.Source == "worklog" {
			writeWorklogEntryText(entry, r)
			continue
		}

		date := formatDim(entry.Timestamp.Format("2006-01-02 15:04"), r)
		author := formatLabel(entry.Author, r)
		fmt.Printf("%s %s: %s\n", date, author, entry.Message)

		for _, f := range entry.Files {
			writeFileChangeText(f, r)
		}
	}

	return nil
}

func writeWorklogEntryText(entry feed.FeedEntry, r *lipgloss.Renderer) {
	date := formatDim(entry.Timestamp.Format("2006-01-02 15:04"), r)
	taskRef := ""
	if entry.TaskID != "" {
		taskRef = fmt.Sprintf(" (%s)", formatTaskID(entry.TaskID, r))
	}
	fmt.Printf("%s [Worklog]%s %s\n", date, taskRef, entry.Message)
}

func writeFileChangeText(f feed.FileChange, r *lipgloss.Renderer) {
	statusTag := fileStatusTag(f)
	taskRef := ""
	if f.TaskID != "" {
		taskRef = fmt.Sprintf(" (%s)", formatTaskID(f.TaskID, r))
	}

	summary := formatChangeSummary(f)
	if summary != "" {
		fmt.Printf("  %s %s%s: %s\n", statusTag, f.Path, taskRef, summary)
	} else {
		fmt.Printf("  %s %s%s\n", statusTag, f.Path, taskRef)
	}
}

// formatChangeSummary builds a compact one-line summary of field and subtask changes.
func formatChangeSummary(f feed.FileChange) string {
	var parts []string
	for _, fc := range f.FieldChanges {
		parts = append(parts, fmt.Sprintf("%s %s \u2192 %s", fc.Field, fc.OldValue, fc.NewValue))
	}
	done := 0
	undone := 0
	for _, sc := range f.SubtaskChanges {
		if sc.Done {
			done++
		} else {
			undone++
		}
	}
	if done > 0 {
		parts = append(parts, fmt.Sprintf("%d subtask(s) completed", done))
	}
	if undone > 0 {
		parts = append(parts, fmt.Sprintf("%d subtask(s) unchecked", undone))
	}
	return strings.Join(parts, ", ")
}

func fileStatusTag(fc feed.FileChange) string {
	if fc.TaskStatus == "completed" {
		return "[Completed]"
	}
	if fc.TaskStatus == "cancelled" {
		return "[Cancelled]"
	}
	switch fc.Status {
	case "created":
		return "[Added]"
	case "modified":
		return "[Modified]"
	case "renamed":
		return "[Renamed]"
	default:
		return "[?]"
	}
}
