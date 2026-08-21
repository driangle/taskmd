package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/worklog"
)

// worklogsEnabled reports whether the project allows writing worklog entries.
// Worklogs are off unless `worklogs: true` is set in .taskmd.yaml, so an absent
// key means disabled. Reading existing worklogs is never gated.
func worklogsEnabled() bool {
	return viper.GetBool("worklogs")
}

var (
	worklogAdd    string
	worklogFormat string
)

var worklogCmd = &cobra.Command{
	Use:        "worklog [task-id]",
	SuggestFor: []string{"journal", "notes"},
	Short:      "View or add worklog entries for a task",
	Long: `View or add worklog entries for a task.

Examples:
  taskmd worklog 015                        # view worklog entries
  taskmd worklog 015 --add "Started implementation"
  taskmd worklog 015 --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runWorklog,
}

func init() {
	rootCmd.AddCommand(worklogCmd)

	worklogCmd.Flags().StringVar(&worklogAdd, "add", "", "append a new worklog entry")
	worklogCmd.Flags().StringVar(&worklogFormat, "format", "text", "output format (text, json, yaml)")
}

func runWorklog(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	// Find the task by ID
	var taskFilePath string
	for _, t := range tasks {
		if t.ID == taskID {
			taskFilePath = t.FilePath
			break
		}
	}
	if taskFilePath == "" {
		return fmt.Errorf("task not found: %s", taskID)
	}

	wlPath := worklog.WorklogPath(taskFilePath, taskID)

	// Add mode
	if worklogAdd != "" {
		if !worklogsEnabled() {
			return fmt.Errorf("worklogs are disabled for this project, so no entry was written for task %s; "+
				"set `worklogs: true` in .taskmd.yaml to enable them", taskID)
		}
		if err := worklog.AppendEntry(wlPath, worklogAdd); err != nil {
			return fmt.Errorf("failed to add worklog entry: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Added worklog entry for task %s\n", taskID)
		return nil
	}

	// View mode
	if !worklog.Exists(wlPath) {
		fmt.Fprintf(os.Stderr, "No worklog found for task %s\n", taskID)
		return nil
	}

	wl, err := worklog.ParseWorklog(wlPath)
	if err != nil {
		return fmt.Errorf("failed to read worklog: %w", err)
	}

	return outputWorklog(wl, worklogFormat, os.Stdout)
}

func outputWorklog(wl *worklog.Worklog, format string, w io.Writer) error {
	switch format {
	case "text":
		return outputWorklogText(wl, w)
	case "json":
		return WriteJSON(w, wl)
	case "yaml":
		return WriteYAML(w, wl)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", format)
	}
}

func outputWorklogText(wl *worklog.Worklog, w io.Writer) error {
	r := getRenderer()
	fmt.Fprintf(w, "%s %s\n", formatLabel("Worklog:", r), formatTaskID(wl.TaskID, r))
	fmt.Fprintf(w, "%s %d\n", formatLabel("Entries:", r), len(wl.Entries))

	if len(wl.Entries) == 0 {
		return nil
	}

	fmt.Fprintln(w)
	for i, entry := range wl.Entries {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s %s\n", formatLabel("##", r), formatDim(entry.Timestamp.Format("2006-01-02T15:04:05Z07:00"), r))
		if entry.Content != "" {
			fmt.Fprintf(w, "\n%s\n", entry.Content)
		}
	}

	return nil
}
