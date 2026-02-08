package parser

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	frontmatterDelimiter = "---"
)

// ParseError represents an error during parsing
type ParseError struct {
	FilePath string
	Message  string
	Err      error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("parse error in %s: %s: %v", e.FilePath, e.Message, e.Err)
	}
	return fmt.Sprintf("parse error in %s: %s", e.FilePath, e.Message)
}

// ParseTaskFile reads and parses a markdown file with YAML frontmatter
func ParseTaskFile(filePath string) (*model.Task, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, &ParseError{
			FilePath: filePath,
			Message:  "failed to read file",
			Err:      err,
		}
	}

	return ParseTaskContent(filePath, content)
}

// ParseTaskContent parses task content from bytes
func ParseTaskContent(filePath string, content []byte) (*model.Task, error) {
	if len(content) == 0 {
		return nil, &ParseError{
			FilePath: filePath,
			Message:  "file is empty",
		}
	}

	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		return nil, &ParseError{
			FilePath: filePath,
			Message:  "failed to extract frontmatter",
			Err:      err,
		}
	}

	task := &model.Task{
		FilePath: filePath,
		Body:     body,
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, task); err != nil {
			return nil, &ParseError{
				FilePath: filePath,
				Message:  "failed to parse YAML frontmatter",
				Err:      err,
			}
		}
	}

	if !task.IsValid() {
		return nil, &ParseError{
			FilePath: filePath,
			Message:  "task is missing required fields (id or title)",
		}
	}

	return task, nil
}

// extractFrontmatter splits content into frontmatter and body
func extractFrontmatter(content []byte) (frontmatter []byte, body string, err error) {
	lines := bytes.Split(content, []byte("\n"))

	// Check if content starts with frontmatter delimiter
	if len(lines) == 0 || string(bytes.TrimSpace(lines[0])) != frontmatterDelimiter {
		// No frontmatter, entire content is body
		return nil, string(content), nil
	}

	// Find closing delimiter
	closingIndex := -1
	for i := 1; i < len(lines); i++ {
		if string(bytes.TrimSpace(lines[i])) == frontmatterDelimiter {
			closingIndex = i
			break
		}
	}

	if closingIndex == -1 {
		return nil, "", fmt.Errorf("unclosed frontmatter delimiter")
	}

	// Extract frontmatter (between delimiters)
	frontmatterLines := lines[1:closingIndex]
	frontmatter = bytes.Join(frontmatterLines, []byte("\n"))

	// Extract body (after closing delimiter)
	if closingIndex+1 < len(lines) {
		bodyLines := lines[closingIndex+1:]
		body = strings.TrimSpace(string(bytes.Join(bodyLines, []byte("\n"))))
	}

	return frontmatter, body, nil
}
