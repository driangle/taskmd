package worklog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Entry is a single timestamped worklog entry.
type Entry struct {
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	Content   string    `json:"content" yaml:"content"`
}

// Worklog is the parsed worklog for a task.
type Worklog struct {
	TaskID   string  `json:"task_id" yaml:"task_id"`
	FilePath string  `json:"file_path" yaml:"file_path"`
	Entries  []Entry `json:"entries" yaml:"entries"`
}

// timestampHeader matches "## <ISO-8601 timestamp>" headings.
var timestampHeader = regexp.MustCompile(`^## (\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:\d{2}))`)

// ParseWorklog reads a worklog file and splits it into entries on ## timestamp headings.
func ParseWorklog(filePath string) (*Worklog, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	taskID := deriveTaskID(filePath)
	wl := &Worklog{
		TaskID:   taskID,
		FilePath: filePath,
		Entries:  parseEntries(string(data)),
	}
	return wl, nil
}

func parseEntries(content string) []Entry {
	var entries []Entry
	lines := strings.Split(content, "\n")

	var current *Entry
	var contentLines []string

	for _, line := range lines {
		if m := timestampHeader.FindStringSubmatch(line); len(m) == 2 {
			// Flush previous entry
			if current != nil {
				current.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
				entries = append(entries, *current)
			}
			ts, err := time.Parse(time.RFC3339, m[1])
			if err != nil {
				// Skip entries with unparseable timestamps
				current = nil
				contentLines = nil
				continue
			}
			current = &Entry{Timestamp: ts}
			contentLines = nil
		} else if current != nil {
			contentLines = append(contentLines, line)
		}
	}

	// Flush last entry
	if current != nil {
		current.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
		entries = append(entries, *current)
	}

	return entries
}

// deriveTaskID extracts the task ID from a worklog file path.
// e.g. "tasks/cli/.worklogs/015.md" -> "015"
func deriveTaskID(filePath string) string {
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// WorklogPath computes the worklog file path for a given task file path.
// e.g. "tasks/cli/015-auth.md" with taskID "015" -> "tasks/cli/.worklogs/015.md"
func WorklogPath(taskFilePath string, taskID string) string {
	dir := filepath.Dir(taskFilePath)
	return filepath.Join(dir, ".worklogs", taskID+".md")
}

// AppendEntry appends a new timestamped entry to a worklog file,
// creating the file and .worklogs/ directory if needed.
func AppendEntry(filePath string, message string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create worklogs directory: %w", err)
	}

	entry := fmt.Sprintf("\n## %s\n\n%s\n", time.Now().UTC().Format(time.RFC3339), message)

	// If file doesn't exist, create it without leading newline
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		entry = strings.TrimLeft(entry, "\n")
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open worklog file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write worklog entry: %w", err)
	}

	return nil
}

// Exists checks whether a worklog file exists for the given path.
func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
