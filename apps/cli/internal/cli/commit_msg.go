package cli

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	commitMsgTaskID string
	commitMsgType   string
	commitMsgBody   bool
	commitMsgShort  bool
)

var allowedTypes = map[string]bool{
	"feat":     true,
	"fix":      true,
	"chore":    true,
	"docs":     true,
	"test":     true,
	"refactor": true,
}

// gitDiffFunc is the function used to get git diff output.
// Override in tests to avoid running actual git commands.
var gitDiffFunc = runGitDiffCached

var commitMsgCmd = &cobra.Command{
	Use:   "commit-msg",
	Short: "Generate a conventional commit message from task metadata",
	Long: `Generate a conventional commit message derived from one or more tasks.

When --task-id is provided, the message is generated from that task.
When no --task-id is provided, the command inspects staged changes (git diff --cached)
to find task files whose status changed to completed.

Examples:
  taskmd commit-msg --task-id 042
  taskmd commit-msg --task-id 042 --type feat
  taskmd commit-msg --task-id 042 --type feat --body
  taskmd commit-msg --task-id 042 --short
  git commit -m "$(taskmd commit-msg --task-id 042)"`,
	Args: cobra.NoArgs,
	RunE: runCommitMsg,
}

func init() {
	rootCmd.AddCommand(commitMsgCmd)

	commitMsgCmd.Flags().StringVar(&commitMsgTaskID, "task-id", "", "task ID to generate message for")
	commitMsgCmd.Flags().StringVar(&commitMsgType, "type", "chore", "commit type prefix (feat, fix, chore, docs, test, refactor)")
	commitMsgCmd.Flags().BoolVar(&commitMsgBody, "body", false, "include completed subtasks as bullet points")
	commitMsgCmd.Flags().BoolVar(&commitMsgShort, "short", false, "subject line only (no body or footer)")
}

func runCommitMsg(_ *cobra.Command, _ []string) error {
	if !allowedTypes[commitMsgType] {
		return fmt.Errorf("invalid commit type %q (allowed: feat, fix, chore, docs, test, refactor)", commitMsgType)
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	if commitMsgTaskID != "" {
		task := findExactMatch(commitMsgTaskID, tasks)
		if task == nil {
			return fmt.Errorf("task not found: %s", commitMsgTaskID)
		}
		fmt.Print(buildCommitMessage([]*model.Task{task}, commitMsgType, commitMsgBody, commitMsgShort))
		return nil
	}

	changes, err := findTaskChangesFromDiff(tasks, scanDir)
	if err != nil {
		return err
	}
	if changes.IsEmpty() {
		return fmt.Errorf("no task changes found in staged changes")
	}

	fmt.Print(buildMessageFromChanges(changes, commitMsgType, commitMsgBody, commitMsgShort))
	return nil
}

// buildMessageFromChanges routes to the appropriate message builder based on
// which change categories are present. Single-category changes use dedicated
// formats for backward compatibility; mixed changes use the combined format.
func buildMessageFromChanges(changes TaskChanges, commitType string, includeBody, short bool) string {
	onlyCompleted := len(changes.Completed) > 0 && len(changes.Added) == 0 &&
		len(changes.Started) == 0 && len(changes.Blocked) == 0 && len(changes.Cancelled) == 0
	if onlyCompleted {
		return buildCommitMessage(changes.Completed, commitType, includeBody, short)
	}

	onlyAdded := len(changes.Added) > 0 && len(changes.Completed) == 0 &&
		len(changes.Started) == 0 && len(changes.Blocked) == 0 && len(changes.Cancelled) == 0
	if onlyAdded {
		return buildAddedTaskMessage(changes.Added, commitType)
	}

	return buildMixedCommitMessage(changes, commitType, includeBody, short)
}

func buildCommitMessage(tasks []*model.Task, commitType string, includeBody, short bool) string {
	if len(tasks) == 1 {
		return buildSingleTaskMessage(tasks[0], commitType, includeBody, short)
	}
	return buildMultiTaskMessage(tasks, commitType, includeBody, short)
}

func buildSingleTaskMessage(task *model.Task, commitType string, includeBody, short bool) string {
	subject := formatSubjectLine(task, commitType)
	if short {
		return subject + "\n"
	}

	var parts []string
	parts = append(parts, subject)

	if includeBody {
		subtasks := extractCompletedSubtasks(task.Body)
		if len(subtasks) > 0 {
			body := formatSubtaskBullets(subtasks)
			parts = append(parts, body)
		}
	}

	return strings.Join(parts, "\n\n") + "\n"
}

func buildMultiTaskMessage(tasks []*model.Task, commitType string, includeBody, short bool) string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}

	subject := fmt.Sprintf("%s: complete tasks %s", commitType, strings.Join(ids, ", "))
	if short {
		return subject + "\n"
	}

	var parts []string
	parts = append(parts, subject)

	if includeBody {
		var bodyParts []string
		for _, t := range tasks {
			subtasks := extractCompletedSubtasks(t.Body)
			if len(subtasks) > 0 {
				section := fmt.Sprintf("%s:\n%s", t.Title, formatSubtaskBullets(subtasks))
				bodyParts = append(bodyParts, section)
			}
		}
		if len(bodyParts) > 0 {
			parts = append(parts, strings.Join(bodyParts, "\n\n"))
		}
	}

	return strings.Join(parts, "\n\n") + "\n"
}

func formatSubjectLine(task *model.Task, commitType string) string {
	title := lowerFirst(task.Title)
	group := task.GetGroup()
	if group != "" {
		return fmt.Sprintf("%s(%s): %s (task %s)", commitType, group, title, task.ID)
	}
	return fmt.Sprintf("%s: %s (task %s)", commitType, title, task.ID)
}

func extractCompletedSubtasks(body string) []string {
	var subtasks []string
	s := bufio.NewScanner(strings.NewReader(body))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "- [x]") || strings.HasPrefix(line, "- [X]") {
			text := strings.TrimSpace(line[5:])
			if text != "" {
				subtasks = append(subtasks, text)
			}
		}
	}
	return subtasks
}

func formatSubtaskBullets(subtasks []string) string {
	lines := make([]string, len(subtasks))
	for i, s := range subtasks {
		lines[i] = "- " + s
	}
	return strings.Join(lines, "\n")
}

// buildAddedTaskMessage generates a commit message for newly added pending tasks.
func buildAddedTaskMessage(tasks []*model.Task, commitType string) string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}

	var subject string
	if len(tasks) == 1 {
		subject = fmt.Sprintf("%s: added task %s", commitType, ids[0])
	} else {
		subject = fmt.Sprintf("%s: added tasks %s", commitType, strings.Join(ids, ", "))
	}
	return subject + "\n"
}

// buildMixedCommitMessage generates a commit message covering multiple change types.
// Subject line uses semicolon-separated segments: "chore: complete task 042; add task 045"
func buildMixedCommitMessage(changes TaskChanges, commitType string, includeBody, short bool) string {
	var segments []string

	if len(changes.Completed) > 0 {
		segments = append(segments, changeSegment("complete", changes.Completed))
	}
	if len(changes.Added) > 0 {
		segments = append(segments, changeSegment("add", changes.Added))
	}
	if len(changes.Started) > 0 {
		segments = append(segments, changeSegment("start", changes.Started))
	}
	if len(changes.Blocked) > 0 {
		segments = append(segments, changeSegment("block", changes.Blocked))
	}
	if len(changes.Cancelled) > 0 {
		segments = append(segments, changeSegment("cancel", changes.Cancelled))
	}

	subject := fmt.Sprintf("%s: %s", commitType, strings.Join(segments, "; "))
	if short {
		return subject + "\n"
	}

	var parts []string
	parts = append(parts, subject)

	if includeBody {
		body := buildMixedBody(changes)
		if body != "" {
			parts = append(parts, body)
		}
	}

	return strings.Join(parts, "\n\n") + "\n"
}

// changeSegment returns a string like "complete task 042" or "add tasks 045, 046".
func changeSegment(verb string, tasks []*model.Task) string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	if len(tasks) == 1 {
		return fmt.Sprintf("%s task %s", verb, ids[0])
	}
	return fmt.Sprintf("%s tasks %s", verb, strings.Join(ids, ", "))
}

// buildMixedBody creates the body section for mixed-type commit messages.
// Each category gets a section with task titles (and subtasks for completed).
func buildMixedBody(changes TaskChanges) string {
	type categoryGroup struct {
		label string
		tasks []*model.Task
	}

	categories := []categoryGroup{
		{"Completed", changes.Completed},
		{"Added", changes.Added},
		{"Started", changes.Started},
		{"Blocked", changes.Blocked},
		{"Cancelled", changes.Cancelled},
	}

	var sections []string
	for _, cat := range categories {
		if len(cat.tasks) == 0 {
			continue
		}
		var lines []string
		for _, t := range cat.tasks {
			if cat.label == "Completed" {
				subtasks := extractCompletedSubtasks(t.Body)
				if len(subtasks) > 0 {
					lines = append(lines, fmt.Sprintf("%s:\n%s", t.Title, formatSubtaskBullets(subtasks)))
				} else {
					lines = append(lines, fmt.Sprintf("- %s (task %s)", t.Title, t.ID))
				}
			} else {
				lines = append(lines, fmt.Sprintf("- %s (task %s)", t.Title, t.ID))
			}
		}
		section := fmt.Sprintf("%s:\n%s", cat.label, strings.Join(lines, "\n"))
		sections = append(sections, section)
	}

	return strings.Join(sections, "\n\n")
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// DiffResult categorizes files from a unified diff by their status change.
type DiffResult struct {
	Completed []string // files with +status: completed
	Added     []string // new files (--- /dev/null) with +status: pending
	Started   []string // files with +status: in-progress
	Blocked   []string // files with +status: blocked
	Cancelled []string // files with +status: cancelled
}

// IsEmpty returns true if no changes were detected.
func (d DiffResult) IsEmpty() bool {
	return len(d.Completed) == 0 && len(d.Added) == 0 && len(d.Started) == 0 &&
		len(d.Blocked) == 0 && len(d.Cancelled) == 0
}

// parseDiffResult parses unified diff output and categorizes files by their
// status change. New files (--- /dev/null) with +status: pending are "Added";
// all others are categorized by their +status: line.
func parseDiffResult(diff string) DiffResult {
	var result DiffResult
	var currentFile string
	var isNewFile bool

	s := bufio.NewScanner(strings.NewReader(diff))
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "--- ") {
			isNewFile = line == "--- /dev/null"
		} else if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
		} else if currentFile != "" && strings.HasPrefix(line, "+status: ") {
			status := strings.TrimPrefix(line, "+status: ")
			status = strings.TrimSpace(status)
			switch status {
			case "completed":
				result.Completed = append(result.Completed, currentFile)
			case "pending":
				if isNewFile {
					result.Added = append(result.Added, currentFile)
				}
			case "in-progress":
				result.Started = append(result.Started, currentFile)
			case "blocked":
				result.Blocked = append(result.Blocked, currentFile)
			case "cancelled":
				result.Cancelled = append(result.Cancelled, currentFile)
			}
			currentFile = "" // avoid duplicates from same file
		}
	}
	return result
}

// TaskChanges holds tasks categorized by their change type.
type TaskChanges struct {
	Completed []*model.Task
	Added     []*model.Task
	Started   []*model.Task
	Blocked   []*model.Task
	Cancelled []*model.Task
}

// IsEmpty returns true if no task changes were found.
func (tc TaskChanges) IsEmpty() bool {
	return len(tc.Completed) == 0 && len(tc.Added) == 0 && len(tc.Started) == 0 &&
		len(tc.Blocked) == 0 && len(tc.Cancelled) == 0
}

// findTaskChangesFromDiff runs git diff --cached, parses all status changes,
// and matches them against scanned tasks.
func findTaskChangesFromDiff(tasks []*model.Task, scanDir string) (TaskChanges, error) {
	var changes TaskChanges

	diffOutput, err := gitDiffFunc(scanDir)
	if err != nil {
		return changes, fmt.Errorf("git diff failed: %w", err)
	}

	dr := parseDiffResult(diffOutput)
	if dr.IsEmpty() {
		return changes, nil
	}

	gitRoot, err := resolveGitRoot(scanDir)
	if err != nil {
		return changes, err
	}

	// Build absolute path sets for each category.
	absMap := func(files []string) map[string]bool {
		m := make(map[string]bool, len(files))
		for _, f := range files {
			m[filepath.Clean(filepath.Join(gitRoot, f))] = true
		}
		return m
	}
	completedSet := absMap(dr.Completed)
	addedSet := absMap(dr.Added)
	startedSet := absMap(dr.Started)
	blockedSet := absMap(dr.Blocked)
	cancelledSet := absMap(dr.Cancelled)

	for _, t := range tasks {
		cleaned := filepath.Clean(t.FilePath)
		switch {
		case completedSet[cleaned]:
			changes.Completed = append(changes.Completed, t)
		case addedSet[cleaned]:
			changes.Added = append(changes.Added, t)
		case startedSet[cleaned]:
			changes.Started = append(changes.Started, t)
		case blockedSet[cleaned]:
			changes.Blocked = append(changes.Blocked, t)
		case cancelledSet[cleaned]:
			changes.Cancelled = append(changes.Cancelled, t)
		}
	}
	return changes, nil
}

// gitRootFunc resolves the git repository root. Override in tests.
var gitRootFunc = defaultGitRoot

func resolveGitRoot(scanDir string) (string, error) {
	return gitRootFunc(scanDir)
}

func defaultGitRoot(scanDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = scanDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find git root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func runGitDiffCached(scanDir string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--unified=0")
	cmd.Dir = scanDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
