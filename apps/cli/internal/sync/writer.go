package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/taskfile"
)

// WriteTaskFile creates a new task markdown file and returns the file path.
func WriteTaskFile(dir, id string, mapped MappedTask, externalID, sourceName string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	slug := slugify(mapped.Title)
	filename := fmt.Sprintf("%s-%s.md", id, slug)
	path := filepath.Join(dir, filename)

	content := renderTaskFile(id, mapped, externalID, sourceName)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write task file: %w", err)
	}

	return path, nil
}

// UpdateSyncedTaskFile updates an existing synced task file.
func UpdateSyncedTaskFile(filePath string, mapped MappedTask) error {
	req := taskfile.UpdateRequest{
		Status: &mapped.Status,
		Title:  &mapped.Title,
	}
	if mapped.Priority != "" {
		req.Priority = &mapped.Priority
	}
	if mapped.Owner != "" {
		req.Owner = &mapped.Owner
	}
	if len(mapped.Tags) > 0 {
		req.Tags = &mapped.Tags
	}
	if mapped.Description != "" {
		req.Body = &mapped.Description
	}
	return taskfile.UpdateTaskFile(filePath, req)
}

func renderTaskFile(id string, mapped MappedTask, externalID, sourceName string) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %q\n", id)
	fmt.Fprintf(&b, "title: %q\n", mapped.Title)
	fmt.Fprintf(&b, "status: %s\n", mapped.Status)
	if mapped.Priority != "" {
		fmt.Fprintf(&b, "priority: %s\n", mapped.Priority)
	}
	if mapped.Owner != "" {
		fmt.Fprintf(&b, "owner: %s\n", mapped.Owner)
	}
	b.WriteString("dependencies: []\n")
	b.WriteString(taskfile.FormatInlineTags(mapped.Tags) + "\n")
	fmt.Fprintf(&b, "sync_source: %s\n", sourceName)
	fmt.Fprintf(&b, "sync_id: %q\n", externalID)
	b.WriteString("---\n")

	if mapped.Description != "" {
		b.WriteString("\n")
		b.WriteString(mapped.Description)
		b.WriteString("\n")
	}

	return b.String()
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
		s = strings.TrimRight(s, "-")
	}
	return s
}
