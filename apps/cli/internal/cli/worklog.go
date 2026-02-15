package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/worklog"
)

var (
	worklogTaskID string
	worklogAdd    string
	worklogFormat string
)

var worklogCmd = &cobra.Command{
	Use:   "worklog",
	Short: "View or add worklog entries for a task",
	Long: `View or add worklog entries for a task.

Examples:
  taskmd worklog --task-id 015           # view worklog entries
  taskmd worklog --task-id 015 --add "Started implementation"
  taskmd worklog --task-id 015 --format json`,
	RunE: runWorklog,
}

func init() {
	rootCmd.AddCommand(worklogCmd)

	worklogCmd.Flags().StringVar(&worklogTaskID, "task-id", "", "task ID (required)")
	worklogCmd.Flags().StringVar(&worklogAdd, "add", "", "append a new worklog entry")
	worklogCmd.Flags().StringVar(&worklogFormat, "format", "text", "output format (text, json, yaml)")

	_ = worklogCmd.MarkFlagRequired("task-id")
}

func runWorklog(cmd *cobra.Command, _ []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Find the task by ID
	var taskFilePath string
	for _, t := range result.Tasks {
		if t.ID == worklogTaskID {
			taskFilePath = t.FilePath
			break
		}
	}
	if taskFilePath == "" {
		return fmt.Errorf("task not found: %s", worklogTaskID)
	}

	wlPath := worklog.WorklogPath(taskFilePath, worklogTaskID)

	// Add mode
	if worklogAdd != "" {
		if err := worklog.AppendEntry(wlPath, worklogAdd); err != nil {
			return fmt.Errorf("failed to add worklog entry: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Added worklog entry for task %s\n", worklogTaskID)
		return nil
	}

	// View mode
	if !worklog.Exists(wlPath) {
		fmt.Fprintf(os.Stderr, "No worklog found for task %s\n", worklogTaskID)
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
