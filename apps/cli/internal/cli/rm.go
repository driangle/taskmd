package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/worklog"
)

var (
	rmForce  bool
	rmDryRun bool
)

// rmStdinReader is the reader used for interactive confirmation.
// Override in tests to simulate user input.
var rmStdinReader io.Reader = os.Stdin

var rmCmd = &cobra.Command{
	Use:        "rm <task-id>",
	SuggestFor: []string{"delete", "remove"},
	Short:      "Delete a task file permanently",
	Long: `Remove permanently deletes a task file by ID.

The command looks up the task, displays its details, and asks for confirmation
before deleting. Use --force to skip the confirmation prompt.

Examples:
  taskmd rm 042
  taskmd rm 042 --force
  taskmd rm 042 -f
  taskmd rm 042 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "skip confirmation prompt")
	rmCmd.Flags().BoolVar(&rmDryRun, "dry-run", false, "preview what would be deleted without acting")
}

func runRm(_ *cobra.Command, args []string) error {
	taskID := args[0]

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	task := findExactMatch(taskID, tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if dupes := findDuplicatesByID(taskID, tasks); len(dupes) > 1 {
		return fmt.Errorf("refusing to delete task %s: found %d files with this ID\n%s\nRun 'taskmd deduplicate' to fix",
			taskID, len(dupes), formatDuplicatePaths(dupes))
	}

	if !rmForce {
		if err := checkTaskReferences(taskID, tasks); err != nil {
			return err
		}
	}

	relPath := resolveRelPath(scanDir, task.FilePath)

	r := getRenderer()
	fmt.Printf("Delete 1 task:\n")
	fmt.Printf("  %s  %s  %s\n", formatTaskID(task.ID, r), task.Title, formatDim("("+relPath+")", r))

	if rmDryRun {
		fmt.Println("\n" + formatWarning("Dry run — no changes made.", r))
		return nil
	}

	if !rmForce {
		fmt.Printf("\nConfirm delete? [y/N] ")
		reader := bufio.NewReader(rmStdinReader)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := os.Remove(task.FilePath); err != nil {
		return fmt.Errorf("failed to delete %s: %w", relPath, err)
	}

	cleanupWorklog(task.FilePath, task.ID)

	fmt.Println(formatSuccess("Deleted 1 task.", r))
	return nil
}

// checkTaskReferences returns an error if any other task references the given ID.
func checkTaskReferences(taskID string, tasks []*model.Task) error {
	refs := findReferencingTasks(taskID, tasks)
	if len(refs) > 0 {
		return fmt.Errorf("task %s is referenced by other tasks (use --force to delete anyway):\n%s",
			taskID, formatReferencingTasks(refs))
	}
	return nil
}

// resolveRelPath returns the task file path relative to the scan directory.
func resolveRelPath(scanDir, filePath string) string {
	absScanDir, err := filepath.Abs(scanDir)
	if err != nil {
		return filePath
	}
	relPath, err := filepath.Rel(absScanDir, filePath)
	if err != nil {
		return filePath
	}
	return relPath
}

// cleanupWorklog removes the worklog file and its parent directory if empty.
func cleanupWorklog(taskFilePath, taskID string) {
	wlPath := worklog.WorklogPath(taskFilePath, taskID)
	if !worklog.Exists(wlPath) {
		return
	}
	if err := os.Remove(wlPath); err != nil {
		fmt.Printf("Warning: failed to delete worklog %s: %v\n", wlPath, err)
		return
	}
	fmt.Println("Deleted worklog: " + filepath.Base(wlPath))
	_ = os.Remove(filepath.Dir(wlPath)) // only succeeds if empty
}

// findReferencingTasks returns tasks that reference the given task ID
// via their Dependencies or Parent fields.
func findReferencingTasks(taskID string, tasks []*model.Task) []*model.Task {
	var refs []*model.Task
	for _, t := range tasks {
		if t.ID == taskID {
			continue
		}
		if t.Parent == taskID {
			refs = append(refs, t)
			continue
		}
		for _, dep := range t.Dependencies {
			if dep == taskID {
				refs = append(refs, t)
				break
			}
		}
	}
	return refs
}

// formatReferencingTasks formats a list of referencing tasks for display.
func formatReferencingTasks(tasks []*model.Task) string {
	var lines []string
	for _, t := range tasks {
		lines = append(lines, fmt.Sprintf("  - %s %s", t.ID, t.Title))
	}
	return strings.Join(lines, "\n")
}
