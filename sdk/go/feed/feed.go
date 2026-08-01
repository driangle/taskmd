package feed

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/driangle/taskmd/sdk/go/worklog"
)

// FeedEntry represents a single event in the activity feed.
type FeedEntry struct {
	Source    string       `json:"source"`
	Hash      string       `json:"hash,omitempty"`
	Author    string       `json:"author,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Message   string       `json:"message"`
	TaskID    string       `json:"taskID,omitempty"`
	Files     []FileChange `json:"files,omitempty"`
}

// FileChange represents a file changed in a commit.
type FileChange struct {
	Path           string          `json:"path"`
	Status         string          `json:"status"`
	TaskID         string          `json:"taskID,omitempty"`
	TaskStatus     string          `json:"taskStatus,omitempty"`
	FieldChanges   []FieldChange   `json:"fieldChanges,omitempty"`
	SubtaskChanges []SubtaskChange `json:"subtaskChanges,omitempty"`
}

// FieldChange represents a frontmatter field that changed between two versions.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

// SubtaskChange represents a subtask checkbox that was toggled.
type SubtaskChange struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// GitShowFunc is a function that retrieves file content at a given commit.
type GitShowFunc func(hash, path string) (string, error)

// GitLogFunc is a function that runs git log with the given args.
type GitLogFunc func(tasksDir string, args []string) (string, error)

var taskIDFromFilenameRegex = regexp.MustCompile(`(?:^|/)(\w+)-`)
var statusLineRegex = regexp.MustCompile(`(?m)^status:\s*(\S+)`)

// Options configures a feed query.
type Options struct {
	TasksDir string
	Limit    int
	Since    string
	Scope    string
	Source   string
	// TaskFile, when set, scopes the query to a single task file (path
	// relative to where git runs). Enables per-task history via --follow.
	TaskFile  string
	Verbose   bool
	GitLogFn  GitLogFunc
	GitShowFn GitShowFunc
}

// Query executes a feed query and returns merged, sorted entries.
func Query(opts Options) ([]FeedEntry, error) {
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if opts.Source == "" {
		opts.Source = "all"
	}

	var gitEntries, worklogEntries []FeedEntry

	if opts.Source != "worklog" {
		if opts.GitLogFn == nil {
			return nil, fmt.Errorf("GitLogFn is required for git source")
		}
		var args []string
		if opts.TaskFile != "" {
			args = BuildGitLogArgsForFile(opts.Limit, opts.Since, opts.TaskFile)
		} else {
			args = BuildGitLogArgs(opts.TasksDir, opts.Limit, opts.Since, opts.Scope)
		}
		output, err := opts.GitLogFn(opts.TasksDir, args)
		if err != nil {
			return nil, fmt.Errorf("failed to read git history: %w", err)
		}
		gitEntries = ParseGitLogOutput(output)
		for i := range gitEntries {
			gitEntries[i].Source = "git"
		}
		if opts.GitShowFn != nil {
			EnrichEntriesWithDiffAnalysis(gitEntries, opts.GitShowFn)
		}
		if opts.TaskFile != "" {
			// --follow can trace history back into unrelated look-alike files;
			// keep only changes that actually belong to the target task.
			gitEntries = filterEntriesByTaskID(gitEntries, ExtractTaskIDFromPath(opts.TaskFile))
		}
	}

	if opts.Source != "git" {
		worklogEntries = ScanWorklogEntries(opts.TasksDir, opts.Scope, opts.Since, opts.Verbose)
		if opts.TaskFile != "" {
			worklogEntries = filterWorklogsByTask(worklogEntries, ExtractTaskIDFromPath(opts.TaskFile))
		}
	}

	entries := MergeEntries(gitEntries, worklogEntries)

	if opts.Limit > 0 && len(entries) > opts.Limit {
		entries = entries[:opts.Limit]
	}

	return entries, nil
}

// BuildGitLogArgs constructs git log arguments for feed queries.
func BuildGitLogArgs(tasksDir string, limit int, since, scope string) []string {
	args := []string{
		"log",
		"--format=%H%n%an%n%ai%n%s",
		"--name-status",
		"--diff-filter=ACMR",
		fmt.Sprintf("-%d", limit),
	}

	if since != "" {
		args = append(args, "--since="+NormalizeSince(since))
	}

	args = append(args, "--")

	if scope != "" && containsGlobChars(scope) {
		matches, _ := filepath.Glob(filepath.Join(tasksDir, scope))
		for _, m := range matches {
			args = append(args, filepath.Join(m, "**", "*.md"))
		}
		if len(matches) == 0 {
			args = append(args, filepath.Join(tasksDir, scope, "**", "*.md"))
		}
	} else if scope != "" {
		args = append(args, filepath.Join(tasksDir, scope, "**", "*.md"))
	} else {
		args = append(args, filepath.Join(tasksDir, "**", "*.md"))
	}

	return args
}

// BuildGitLogArgsForFile constructs git log arguments scoped to a single task
// file. It uses --follow so the file's history is threaded across renames, and
// the initial "A" commit surfaces the task's creation event.
func BuildGitLogArgsForFile(limit int, since, taskFile string) []string {
	args := []string{
		"log",
		"--format=%H%n%an%n%ai%n%s",
		"--name-status",
		"--diff-filter=ACMR",
		fmt.Sprintf("-%d", limit),
		"--follow",
	}

	if since != "" {
		args = append(args, "--since="+NormalizeSince(since))
	}

	args = append(args, "--", taskFile)

	return args
}

// filterEntriesByTaskID drops git file changes that don't belong to the given
// task ID (and any entry thereby left empty). Renames preserve a task's ID, so
// genuine history is kept while --follow false positives are discarded.
func filterEntriesByTaskID(entries []FeedEntry, taskID string) []FeedEntry {
	if taskID == "" {
		return entries
	}
	var out []FeedEntry
	for _, e := range entries {
		var files []FileChange
		for _, f := range e.Files {
			if f.TaskID == taskID {
				files = append(files, f)
			}
		}
		if len(files) > 0 {
			e.Files = files
			out = append(out, e)
		}
	}
	return out
}

// filterWorklogsByTask keeps only worklog entries belonging to the given task ID.
func filterWorklogsByTask(entries []FeedEntry, taskID string) []FeedEntry {
	if taskID == "" {
		return entries
	}
	var out []FeedEntry
	for _, e := range entries {
		if e.TaskID == taskID {
			out = append(out, e)
		}
	}
	return out
}

func containsGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// NormalizeSince converts shorthand durations like "2d" or "1w" into
// git-compatible relative date strings like "2.days.ago" or "1.weeks.ago".
func NormalizeSince(s string) string {
	unitMap := map[byte]string{
		'd': "days",
		'w': "weeks",
		'm': "months",
		'y': "years",
	}

	if len(s) < 2 {
		return s
	}

	unit := s[len(s)-1]
	numPart := s[:len(s)-1]

	word, ok := unitMap[unit]
	if !ok {
		return s
	}

	for _, c := range numPart {
		if c < '0' || c > '9' {
			return s
		}
	}

	return numPart + "." + word + ".ago"
}

// ParseGitLogOutput parses the output of a git log command into FeedEntry values.
func ParseGitLogOutput(output string) []FeedEntry {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	var entries []FeedEntry
	var current *FeedEntry
	lineAfterHash := 0

	s := bufio.NewScanner(strings.NewReader(output))
	for s.Scan() {
		line := s.Text()

		if len(line) == 40 && isHexString(line) {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &FeedEntry{Hash: line}
			lineAfterHash = 1
			continue
		}

		if current == nil || len(line) == 0 {
			continue
		}

		lineAfterHash = parseEntryLine(current, line, lineAfterHash)
	}

	if current != nil {
		entries = append(entries, *current)
	}

	return entries
}

func parseEntryLine(entry *FeedEntry, line string, pos int) int {
	switch pos {
	case 1:
		entry.Author = line
		return 2
	case 2:
		t, err := time.Parse("2006-01-02 15:04:05 -0700", line)
		if err == nil {
			entry.Timestamp = t
		}
		return 3
	case 3:
		entry.Message = line
		return 4
	default:
		if fc := parseFileChangeLine(line); fc != nil {
			entry.Files = append(entry.Files, *fc)
		}
		return pos
	}
}

func parseFileChangeLine(line string) *FileChange {
	parts := strings.Split(line, "\t")
	if len(parts) < 2 {
		return nil
	}

	statusCode := parts[0]
	var path, status string

	switch {
	case statusCode == "A":
		status = "created"
		path = parts[1]
	case statusCode == "M":
		status = "modified"
		path = parts[1]
	case strings.HasPrefix(statusCode, "R"):
		status = "renamed"
		path = destPath(parts)
	case strings.HasPrefix(statusCode, "C"):
		// A copy (e.g. a task file created from a template under --follow) is,
		// for the destination file, a creation event.
		status = "created"
		path = destPath(parts)
	default:
		return nil
	}

	taskID := ExtractTaskIDFromPath(path)

	return &FileChange{
		Path:   path,
		Status: status,
		TaskID: taskID,
	}
}

// destPath returns the destination path of a rename/copy name-status line
// (format "<status>\t<src>\t<dst>"), falling back to the first path field.
func destPath(parts []string) string {
	if len(parts) >= 3 {
		return parts[2]
	}
	return parts[1]
}

// ExtractTaskIDFromPath extracts a task ID from a file path.
func ExtractTaskIDFromPath(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	match := taskIDFromFilenameRegex.FindStringSubmatch(base)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// EnrichEntriesWithDiffAnalysis reads task files at each commit and its parent
// to detect field-level changes and subtask completions.
func EnrichEntriesWithDiffAnalysis(entries []FeedEntry, gitShowFn GitShowFunc) {
	for i := range entries {
		for j := range entries[i].Files {
			enrichFileChange(&entries[i].Files[j], entries[i].Hash, gitShowFn)
		}
	}
}

func enrichFileChange(fc *FileChange, hash string, gitShowFn GitShowFunc) {
	newContent, err := gitShowFn(hash, fc.Path)
	if err != nil {
		return
	}

	if fc.Status != "modified" {
		setTerminalStatus(fc, newContent)
		return
	}

	oldContent, err := gitShowFn(hash+"^", fc.Path)
	if err != nil {
		setTerminalStatus(fc, newContent)
		return
	}

	fieldChanges, subtaskChanges := AnalyzeDiff(oldContent, newContent)
	fc.FieldChanges = fieldChanges
	fc.SubtaskChanges = subtaskChanges

	for _, change := range fieldChanges {
		if change.Field == "status" && (change.NewValue == "completed" || change.NewValue == "cancelled") {
			fc.TaskStatus = change.NewValue
		}
	}
}

// extractStatusFromContent extracts the status field from task file frontmatter.
func extractStatusFromContent(content string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return ""
	}
	match := statusLineRegex.FindStringSubmatch(parts[1])
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

func setTerminalStatus(fc *FileChange, content string) {
	status := extractStatusFromContent(content)
	if status == "completed" || status == "cancelled" {
		fc.TaskStatus = status
	}
}

// ScanWorklogEntries finds .worklogs/*.md files under tasksDir and converts
// their entries into FeedEntry values with Source "worklog".
func ScanWorklogEntries(tasksDir, scope, since string, verbose bool) []FeedEntry {
	var sinceTime time.Time
	if since != "" {
		sinceTime = ParseSinceTime(since)
	}

	pattern := buildWorklogGlobPattern(tasksDir, scope)
	files, _ := filepath.Glob(pattern)

	var entries []FeedEntry
	for _, f := range files {
		wl, err := worklog.ParseWorklog(f)
		if err != nil {
			continue
		}

		for _, e := range wl.Entries {
			if !sinceTime.IsZero() && e.Timestamp.Before(sinceTime) {
				continue
			}
			entries = append(entries, FeedEntry{
				Source:    "worklog",
				TaskID:    wl.TaskID,
				Timestamp: e.Timestamp,
				Message:   truncateFirstLine(e.Content),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	return entries
}

func buildWorklogGlobPattern(tasksDir, scope string) string {
	if scope != "" && containsGlobChars(scope) {
		return filepath.Join(tasksDir, scope, ".worklogs", "*.md")
	}
	if scope != "" {
		return filepath.Join(tasksDir, scope, ".worklogs", "*.md")
	}
	return filepath.Join(tasksDir, "*", ".worklogs", "*.md")
}

func truncateFirstLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// ParseSinceTime converts a since string into a time.Time cutoff.
func ParseSinceTime(since string) time.Time {
	unitDurations := map[byte]time.Duration{
		'd': 24 * time.Hour,
		'w': 7 * 24 * time.Hour,
		'm': 30 * 24 * time.Hour,
		'y': 365 * 24 * time.Hour,
	}

	if len(since) >= 2 {
		unit := since[len(since)-1]
		numPart := since[:len(since)-1]
		if d, ok := unitDurations[unit]; ok {
			allDigits := true
			for _, c := range numPart {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				n := 0
				for _, c := range numPart {
					n = n*10 + int(c-'0')
				}
				return time.Now().Add(-time.Duration(n) * d)
			}
		}
	}

	if t, err := time.Parse("2006-01-02", since); err == nil {
		return t
	}

	return time.Time{}
}

// MergeEntries merges two slices of FeedEntry sorted by timestamp descending.
func MergeEntries(a, b []FeedEntry) []FeedEntry {
	if len(b) == 0 {
		return a
	}
	if len(a) == 0 {
		return b
	}

	result := make([]FeedEntry, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Timestamp.After(b[j].Timestamp) || a[i].Timestamp.Equal(b[j].Timestamp) {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}
	result = append(result, a[i:]...)
	result = append(result, b[j:]...)
	return result
}
