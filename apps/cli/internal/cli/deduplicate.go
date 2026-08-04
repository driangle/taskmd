package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/taskfile"
)

var (
	dedupDryRun        bool
	dedupFormat        string
	dedupNoInteractive bool
)

// dedupStdinReader is the reader used for interactive disambiguation.
// Override in tests to simulate user input.
var dedupStdinReader io.Reader = os.Stdin

// dedupIsTTY checks whether stdin is a terminal. Override in tests.
var dedupIsTTY = func() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

var deduplicateCmd = &cobra.Command{
	Use:        "deduplicate [path]",
	SuggestFor: []string{"dedup"},
	Short:      "Detect and resolve duplicate task IDs",
	Long: `Deduplicate finds tasks with colliding IDs and reassigns new IDs to resolve conflicts.

When multiple contributors create tasks on separate branches, IDs can collide after merge.
This command detects duplicates and assigns new IDs to the newer tasks (by created date).

For each collision:
  - The oldest task keeps its original ID
  - Newer tasks get reassigned a fresh ID
  - File is renamed to match the new ID
  - Cross-references (dependencies, parent) in all tasks are updated

When references to a duplicate ID are ambiguous, you will be prompted to choose
which task each reference should resolve to. Use --no-interactive to skip prompts
and fall back to automatic behavior (all references rewrite to the new ID).

Use --dry-run to preview changes without modifying files.

Examples:
  taskmd deduplicate
  taskmd deduplicate ./tasks
  taskmd deduplicate --dry-run
  taskmd deduplicate --no-interactive
  taskmd deduplicate --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDeduplicate,
}

func init() {
	rootCmd.AddCommand(deduplicateCmd)

	deduplicateCmd.Flags().BoolVar(&dedupDryRun, "dry-run", false, "preview changes without modifying files")
	deduplicateCmd.Flags().StringVar(&dedupFormat, "format", "text", "output format (text, json)")
	deduplicateCmd.Flags().BoolVar(&dedupNoInteractive, "no-interactive", false, "skip interactive prompts for ambiguous references")
}

type reassignment struct {
	OldID       string `json:"old_id"`
	NewID       string `json:"new_id"`
	OldFilePath string `json:"old_file_path"`
	NewFilePath string `json:"new_file_path"`
	Title       string `json:"title"`
}

type deduplicateResult struct {
	DryRun        bool           `json:"dry_run"`
	Duplicates    int            `json:"duplicates"`
	Reassignments []reassignment `json:"reassignments"`
	AmbiguousRefs []ambiguousRef `json:"ambiguous_refs,omitempty"`
}

// ambiguousRef represents a reference to a duplicate ID that needs disambiguation.
type ambiguousRef struct {
	ReferencingTask *model.Task   `json:"-"`
	ReferencingFile string        `json:"referencing_file"`
	Field           string        `json:"field"`
	DuplicateID     string        `json:"duplicate_id"`
	Candidates      []*model.Task `json:"-"`
	CandidateFiles  []string      `json:"candidate_files"`
}

func runDeduplicate(cmd *cobra.Command, args []string) error {
	if err := ValidateFormat(dedupFormat, []string{"text", "json"}); err != nil {
		return err
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	taskScanner := newTaskScanner(scanDir, flags)
	scanResult, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	reassignments, err := planReassignments(scanResult.Tasks)
	if err != nil {
		return err
	}

	idMap := buildIDMap(scanResult.Tasks)
	ambRefs := findAmbiguousRefs(reassignments, scanResult.Tasks, idMap)

	interactive := !dedupNoInteractive && !dedupDryRun && dedupIsTTY()

	var refOverrides map[string]string
	if len(ambRefs) > 0 && interactive {
		refOverrides, err = promptDisambiguation(ambRefs, reassignments)
		if err != nil {
			return err
		}
	}

	if !dedupDryRun {
		if err := applyReassignments(reassignments, scanResult.Tasks, refOverrides); err != nil {
			return err
		}
	}

	result := deduplicateResult{
		DryRun:        dedupDryRun,
		Duplicates:    len(reassignments),
		Reassignments: reassignments,
		AmbiguousRefs: ambRefs,
	}

	return outputDeduplicateResult(result, flags.Quiet)
}

// findAmbiguousRefs finds references to duplicate IDs from tasks outside the collision group.
func findAmbiguousRefs(reassignments []reassignment, allTasks []*model.Task, idMap map[string][]*model.Task) []ambiguousRef {
	// Collect which old IDs are being reassigned (these are the collision IDs).
	collisionIDs := make(map[string]bool)
	for _, r := range reassignments {
		collisionIDs[r.OldID] = true
	}

	// Build set of file paths that are part of each collision group.
	collisionFiles := make(map[string]bool)
	for id := range collisionIDs {
		for _, t := range idMap[id] {
			collisionFiles[t.FilePath] = true
		}
	}

	var refs []ambiguousRef
	for _, task := range allTasks {
		if collisionFiles[task.FilePath] {
			continue
		}

		for _, depID := range task.Dependencies {
			if collisionIDs[depID] {
				refs = append(refs, ambiguousRef{
					ReferencingTask: task,
					ReferencingFile: task.FilePath,
					Field:           "dependencies",
					DuplicateID:     depID,
					Candidates:      idMap[depID],
					CandidateFiles:  taskFilePaths(idMap[depID]),
				})
			}
		}

		if task.Parent != "" && collisionIDs[task.Parent] {
			refs = append(refs, ambiguousRef{
				ReferencingTask: task,
				ReferencingFile: task.FilePath,
				Field:           "parent",
				DuplicateID:     task.Parent,
				Candidates:      idMap[task.Parent],
				CandidateFiles:  taskFilePaths(idMap[task.Parent]),
			})
		}
	}

	return refs
}

func taskFilePaths(tasks []*model.Task) []string {
	paths := make([]string, len(tasks))
	for i, t := range tasks {
		paths[i] = t.FilePath
	}
	return paths
}

// refOverrideKey builds a map key for a reference override: "filePath\x00oldID\x00field".
func refOverrideKey(filePath, oldID, field string) string {
	return filePath + "\x00" + oldID + "\x00" + field
}

// promptDisambiguation interactively asks the user which candidate each ambiguous reference should resolve to.
// Returns a map from refOverrideKey → target ID.
func promptDisambiguation(ambRefs []ambiguousRef, reassignments []reassignment) (map[string]string, error) {
	// Build a lookup: oldID → newID from reassignments.
	oldToNew := make(map[string]string)
	for _, r := range reassignments {
		oldToNew[r.OldID] = r.NewID
	}

	overrides := make(map[string]string)
	reader := bufio.NewReader(dedupStdinReader)

	fmt.Println("\nAmbiguous references detected. Please choose the correct target for each:")

	for _, ref := range ambRefs {
		fmt.Printf("\nTask %q (%s) references ID %q in %s.\n",
			ref.ReferencingTask.Title, ref.ReferencingFile, ref.DuplicateID, ref.Field)
		fmt.Println("Which task should this reference point to?")
		fmt.Println()

		for i, candidate := range ref.Candidates {
			created := candidate.Created.Format("2006-01-02")
			label := ""
			if i == 0 {
				label = " (keeps ID)"
			}
			fmt.Printf("  [%d] %q (%s, created %s)%s\n", i+1, candidate.Title, candidate.FilePath, created, label)
		}

		fmt.Printf("\nChoice [1-%d]: ", len(ref.Candidates))

		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		choice := strings.TrimSpace(line)
		idx := 0
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(ref.Candidates) {
			// Default to first candidate (oldest, keeps ID).
			idx = 1
			fmt.Printf("  Invalid choice, defaulting to [1].\n")
		}

		chosen := ref.Candidates[idx-1]
		key := refOverrideKey(ref.ReferencingFile, ref.DuplicateID, ref.Field)

		if idx == 1 {
			// User chose the oldest task (which keeps the original ID).
			// The reference should stay as-is, so we override to the original ID.
			overrides[key] = ref.DuplicateID
		} else {
			// User chose a reassigned task. Find its new ID.
			for _, r := range reassignments {
				if r.OldFilePath == chosen.FilePath {
					overrides[key] = r.NewID
					break
				}
			}
		}
	}

	return overrides, nil
}

// planReassignments detects duplicate IDs and plans which tasks need new IDs.
func planReassignments(tasks []*model.Task) ([]reassignment, error) {
	idMap := buildIDMap(tasks)
	allIDs := collectAllIDs(tasks)
	cfg := resolveIDConfig()

	var reassignments []reassignment

	for id, group := range idMap {
		if len(group) < 2 {
			continue
		}

		sortByCreated(group)

		for _, task := range group[1:] {
			newID, err := generateID(allIDs, cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to generate new ID for duplicate %q: %w", id, err)
			}

			allIDs = append(allIDs, newID)

			reassignments = append(reassignments, reassignment{
				OldID:       task.ID,
				NewID:       newID,
				OldFilePath: task.FilePath,
				NewFilePath: buildNewFilePath(task.FilePath, task.ID, newID),
				Title:       task.Title,
			})
		}
	}

	sort.Slice(reassignments, func(i, j int) bool {
		return reassignments[i].OldFilePath < reassignments[j].OldFilePath
	})

	return reassignments, nil
}

// buildIDMap groups tasks by their ID.
func buildIDMap(tasks []*model.Task) map[string][]*model.Task {
	m := make(map[string][]*model.Task)
	for _, t := range tasks {
		m[t.ID] = append(m[t.ID], t)
	}
	return m
}

// collectAllIDs returns a slice of all task IDs.
func collectAllIDs(tasks []*model.Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}

// sortByCreated sorts tasks by Created date ascending (oldest first).
// Falls back to filepath alphabetical order for equal dates.
func sortByCreated(tasks []*model.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		if !tasks[i].Created.Equal(tasks[j].Created.Time) {
			return tasks[i].Created.Before(tasks[j].Created.Time)
		}
		return tasks[i].FilePath < tasks[j].FilePath
	})
}

// buildNewFilePath constructs the renamed file path by replacing the old ID prefix with the new ID.
func buildNewFilePath(oldPath, oldID, newID string) string {
	dir := filepath.Dir(oldPath)
	base := filepath.Base(oldPath)

	// Replace the ID prefix in the filename: "001-some-slug.md" → "abc123-some-slug.md"
	if strings.HasPrefix(base, oldID+"-") {
		base = newID + base[len(oldID):]
	} else if strings.HasPrefix(base, oldID+".") {
		base = newID + base[len(oldID):]
	} else {
		// Fallback: prepend new ID
		base = newID + "-" + base
	}

	return filepath.Join(dir, base)
}

// applyReassignments performs the actual file modifications.
// refOverrides maps refOverrideKey → target ID for references that the user explicitly chose.
func applyReassignments(reassignments []reassignment, allTasks []*model.Task, refOverrides map[string]string) error {
	for _, r := range reassignments {
		if err := taskfile.ReplaceID(r.OldFilePath, r.NewID); err != nil {
			return fmt.Errorf("failed to update ID in %s: %w", r.OldFilePath, err)
		}

		if r.OldFilePath != r.NewFilePath {
			if err := os.Rename(r.OldFilePath, r.NewFilePath); err != nil {
				return fmt.Errorf("failed to rename %s → %s: %w", r.OldFilePath, r.NewFilePath, err)
			}
		}
	}

	return updateCrossReferences(reassignments, allTasks, refOverrides)
}

// updateCrossReferences rewrites dependency and parent references in all task files.
func updateCrossReferences(reassignments []reassignment, allTasks []*model.Task, refOverrides map[string]string) error {
	renamedPaths := make(map[string]string, len(reassignments))
	for _, r := range reassignments {
		renamedPaths[r.OldFilePath] = r.NewFilePath
	}

	for _, task := range allTasks {
		filePath := task.FilePath
		if newPath, ok := renamedPaths[filePath]; ok {
			filePath = newPath
		}

		for _, r := range reassignments {
			if shouldSkipRewrite(refOverrides, task.FilePath, r.OldID) {
				continue
			}

			if err := taskfile.ReplaceReference(filePath, r.OldID, r.NewID); err != nil {
				return fmt.Errorf("failed to update references in %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// shouldSkipRewrite checks if an override exists that keeps the original ID for this file+oldID.
func shouldSkipRewrite(refOverrides map[string]string, origFilePath, oldID string) bool {
	if refOverrides == nil {
		return false
	}
	for _, field := range []string{"dependencies", "parent"} {
		key := refOverrideKey(origFilePath, oldID, field)
		if targetID, ok := refOverrides[key]; ok && targetID == oldID {
			return true
		}
	}
	return false
}

func outputDeduplicateResult(result deduplicateResult, quiet bool) error {
	if dedupFormat == "json" {
		return WriteJSON(os.Stdout, result)
	}

	if result.Duplicates == 0 {
		if !quiet {
			fmt.Println("No duplicate IDs found.")
		}
		return nil
	}

	prefix := ""
	if result.DryRun {
		prefix = "[dry-run] "
	}

	fmt.Printf("%sFound %d duplicate(s) to resolve:\n\n", prefix, result.Duplicates)
	for _, r := range result.Reassignments {
		fmt.Printf("  %s → %s  %s\n", r.OldID, r.NewID, r.Title)
		fmt.Printf("    %s → %s\n", r.OldFilePath, r.NewFilePath)
	}

	if result.DryRun && len(result.AmbiguousRefs) > 0 {
		fmt.Println("\n[dry-run] Ambiguous references detected:")
		for _, ref := range result.AmbiguousRefs {
			fmt.Printf("  %s references %q in %s (%d candidates)\n",
				ref.ReferencingFile, ref.DuplicateID, ref.Field, len(ref.Candidates))
		}
	}

	if result.DryRun {
		fmt.Println("\nNo changes made (dry-run mode).")
	} else {
		fmt.Printf("\nResolved %d duplicate(s).\n", result.Duplicates)
	}

	return nil
}
