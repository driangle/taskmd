package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	archiveIDs          []string
	archiveStatus       string
	archiveAllCompleted bool
	archiveAllCancelled bool
	archiveTag          string
	archiveDryRun       bool
	archiveYes          bool
	archiveDelete       bool
	archiveForce        bool
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive or delete completed/cancelled tasks",
	Long: `Archive moves task files into an archive/ subdirectory, keeping your
main task list clean while preserving history. Use --delete to permanently
remove tasks instead of archiving them.

Tasks are selected by ID, status, or tag. Multiple filters use AND logic.

Examples:
  taskmd archive --all-completed
  taskmd archive --all-cancelled --dry-run
  taskmd archive --id 042 --id 043
  taskmd archive --status completed --tag backend
  taskmd archive --all-completed --delete
  taskmd archive --all-completed -y`,
	Args: cobra.NoArgs,
	RunE: runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)

	archiveCmd.Flags().StringArrayVar(&archiveIDs, "id", nil, "archive task(s) by ID (repeatable)")
	archiveCmd.Flags().StringVar(&archiveStatus, "status", "", "archive tasks matching this status")
	archiveCmd.Flags().BoolVar(&archiveAllCompleted, "all-completed", false, "archive all completed tasks")
	archiveCmd.Flags().BoolVar(&archiveAllCancelled, "all-cancelled", false, "archive all cancelled tasks")
	archiveCmd.Flags().StringVar(&archiveTag, "tag", "", "archive tasks with this tag")
	archiveCmd.Flags().BoolVar(&archiveDryRun, "dry-run", false, "preview changes without making them")
	archiveCmd.Flags().BoolVarP(&archiveYes, "yes", "y", false, "skip confirmation prompt")
	archiveCmd.Flags().BoolVar(&archiveDelete, "delete", false, "permanently delete instead of archive")
	archiveCmd.Flags().BoolVarP(&archiveForce, "force", "f", false, "skip confirmation for delete")
}

func runArchive(_ *cobra.Command, _ []string) error {
	if err := validateArchiveFlags(); err != nil {
		return err
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	selected := filterArchiveTasks(result.Tasks)

	if len(selected) == 0 {
		return fmt.Errorf("no tasks match the given criteria")
	}

	absScanDir, err := filepath.Abs(scanDir)
	if err != nil {
		return fmt.Errorf("failed to resolve scan directory: %w", err)
	}

	action := "Archive"
	if archiveDelete {
		action = "Delete"
	}

	printArchivePreview(selected, action, absScanDir)

	if archiveDryRun {
		r := getRenderer()
		fmt.Println("\n" + formatWarning("Dry run — no changes made.", r))
		return nil
	}

	if !archiveYes && !(archiveDelete && archiveForce) {
		return fmt.Errorf("use --yes (-y) to confirm, or --dry-run to preview")
	}

	if archiveDelete {
		return executeDelete(selected)
	}
	return executeArchive(selected, absScanDir)
}

func validateArchiveFlags() error {
	hasFilter := len(archiveIDs) > 0 || archiveStatus != "" ||
		archiveAllCompleted || archiveAllCancelled || archiveTag != ""

	if !hasFilter {
		return fmt.Errorf("specify tasks to archive: --id, --status, --all-completed, --all-cancelled, or --tag")
	}

	if archiveAllCompleted && archiveStatus != "" {
		return fmt.Errorf("--all-completed and --status are mutually exclusive")
	}
	if archiveAllCancelled && archiveStatus != "" {
		return fmt.Errorf("--all-cancelled and --status are mutually exclusive")
	}
	if archiveAllCompleted && archiveAllCancelled {
		return fmt.Errorf("--all-completed and --all-cancelled are mutually exclusive")
	}

	if archiveStatus != "" {
		valid := []string{"pending", "in-progress", "completed", "blocked", "cancelled"}
		if !slices.Contains(valid, archiveStatus) {
			return fmt.Errorf("invalid status %q (valid: %s)", archiveStatus, strings.Join(valid, ", "))
		}
	}

	return nil
}

func filterArchiveTasks(tasks []*model.Task) []*model.Task {
	// Resolve effective status filter
	effectiveStatus := archiveStatus
	if archiveAllCompleted {
		effectiveStatus = string(model.StatusCompleted)
	} else if archiveAllCancelled {
		effectiveStatus = string(model.StatusCancelled)
	}

	// Build ID set for fast lookup
	idSet := make(map[string]bool, len(archiveIDs))
	for _, id := range archiveIDs {
		idSet[id] = true
	}

	var selected []*model.Task
	for _, task := range tasks {
		if len(idSet) > 0 && !idSet[task.ID] {
			continue
		}
		if effectiveStatus != "" && string(task.Status) != effectiveStatus {
			continue
		}
		if archiveTag != "" && !slices.Contains(task.Tags, archiveTag) {
			continue
		}
		selected = append(selected, task)
	}

	return selected
}

func printArchivePreview(tasks []*model.Task, action, absScanDir string) {
	r := getRenderer()
	fmt.Printf("%s %d task(s):\n", action, len(tasks))
	for _, task := range tasks {
		relPath, err := filepath.Rel(absScanDir, task.FilePath)
		if err != nil {
			relPath = task.FilePath
		}
		fmt.Printf("  %s  %s  %s\n", formatTaskID(task.ID, r), task.Title, formatDim("("+relPath+")", r))
	}
}

func executeDelete(tasks []*model.Task) error {
	r := getRenderer()
	for _, task := range tasks {
		if err := os.Remove(task.FilePath); err != nil {
			return fmt.Errorf("failed to delete %s: %w", task.FilePath, err)
		}
	}
	fmt.Println(formatSuccess(fmt.Sprintf("Deleted %d task(s).", len(tasks)), r))
	return nil
}

func executeArchive(tasks []*model.Task, absScanDir string) error {
	r := getRenderer()
	archiveDir := filepath.Join(absScanDir, "archive")

	for _, task := range tasks {
		relPath, err := filepath.Rel(absScanDir, task.FilePath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %s: %w", task.FilePath, err)
		}

		destPath := filepath.Join(archiveDir, relPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create archive directory: %w", err)
		}

		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("archive destination already exists: %s", destPath)
		}

		if err := os.Rename(task.FilePath, destPath); err != nil {
			return fmt.Errorf("failed to move %s: %w", task.FilePath, err)
		}
	}

	fmt.Println(formatSuccess(fmt.Sprintf("Archived %d task(s).", len(tasks)), r))
	return nil
}
