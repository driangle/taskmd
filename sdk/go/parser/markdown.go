package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/sdk/go/model"
)

const (
	frontmatterDelimiter = "---"
)

// ParseError represents an error during parsing
type ParseError struct {
	FilePath string
	Message  string
	Err      error
	// Frontmatter holds the raw frontmatter text when the error came from a
	// frontmatter block that was present but could not be parsed. It is nil
	// for other errors (missing file, no frontmatter, missing fields).
	Frontmatter []byte
}

// MalformedTaskFrontmatter reports whether this error came from a file whose
// frontmatter was present but invalid, and whose raw text carries
// task-signature keys (see HasTaskSignature). Scanners use it to distinguish
// a broken task file, which must be reported, from foreign markdown, which is
// silently skipped.
func (e *ParseError) MalformedTaskFrontmatter() bool {
	return len(e.Frontmatter) > 0 && HasTaskSignature(e.Frontmatter)
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
		// The only extraction error is an unclosed frontmatter block, so the
		// file did open a frontmatter delimiter; expose the raw content for
		// task-signature sniffing.
		return nil, &ParseError{
			FilePath:    filePath,
			Message:     "failed to extract frontmatter",
			Err:         err,
			Frontmatter: content,
		}
	}

	task := &model.Task{
		FilePath: filePath,
		Body:     body,
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, task); err != nil {
			return nil, &ParseError{
				FilePath:    filePath,
				Message:     "failed to parse YAML frontmatter",
				Err:         err,
				Frontmatter: frontmatter,
			}
		}
		// Backward compatibility: accept deprecated "created" field
		if task.Created.IsZero() && !task.CreatedDeprecated.IsZero() {
			task.Created = task.CreatedDeprecated
		}
	}

	// Derive missing fields from filename
	if task.ID == "" || task.Title == "" {
		deriveFieldsFromFilename(task)
	}

	if !task.IsValid() {
		return nil, &ParseError{
			FilePath: filePath,
			Message:  "task is missing required fields (id or title)",
		}
	}

	return task, nil
}

// deriveFieldsFromFilename extracts ID and title from a filename.
// Supports these patterns:
//  1. Sequential: "009-add-feature.md" → ID="009"
//  2. Prefixed:   "dr-001-fix-login.md" → ID="dr-001"
//  3. Random:     "a3f9x2-slug-title.md" → ID="a3f9x2"
//  4. UUID:       "f47ac10b-fix-bug.md" → ID="f47ac10b"
//  5. Full UUID:  "f47ac10b-58cc-4372-a567-0e02b2c3d479-slug.md" → ID="f47ac10b-58cc-4372-a567-0e02b2c3d479"
func deriveFieldsFromFilename(task *model.Task) {
	base := filepath.Base(task.FilePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	if len(name) == 0 {
		return
	}

	id, slug := splitFilenameID(name)
	if id == "" {
		return
	}

	if task.ID == "" {
		task.ID = id
	}
	if task.Title == "" && slug != "" {
		task.Title = strings.ReplaceAll(slug, "-", " ")
	}
}

// splitFilenameID identifies the ID portion and remaining slug from a filename stem.
// Returns ("", "") if no ID pattern matches.
func splitFilenameID(name string) (id, slug string) {
	// Pattern: Full UUID — 8-4-4-4-12 hex (e.g. "f47ac10b-58cc-4372-a567-0e02b2c3d479-slug")
	if uid, rest, ok := matchFullUUID(name); ok {
		return uid, rest
	}

	parts := strings.SplitN(name, "-", 2)

	// Pattern 1: Sequential — starts with digit (e.g. "009-add-feature")
	if name[0] >= '0' && name[0] <= '9' {
		slug := ""
		if len(parts) == 2 {
			slug = parts[1]
		}
		return parts[0], slug
	}

	if len(parts) < 2 {
		return "", ""
	}

	// Pattern 2: Prefixed — alpha prefix + hyphen + digits (e.g. "dr-001-fix-login")
	if isAlpha(parts[0]) {
		restParts := strings.SplitN(parts[1], "-", 2)
		if isNumeric(restParts[0]) {
			slug := ""
			if len(restParts) == 2 {
				slug = restParts[1]
			}
			return parts[0] + "-" + restParts[0], slug
		}
	}

	// Pattern 3: Random — 3-8 lowercase alphanumeric with at least one digit
	if isAlphanumericID(parts[0]) {
		return parts[0], parts[1]
	}

	// Pattern 4: Hex ID — 9-32 hex chars (truncated UUID longer than random range)
	if isHexID(parts[0]) {
		return parts[0], parts[1]
	}

	return "", ""
}

// isNumeric returns true if s is non-empty and all characters are digits.
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isAlpha returns true if s is non-empty and all characters are lowercase letters.
func isAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

// matchFullUUID checks if name starts with a full UUID (8-4-4-4-12 hex pattern).
// Returns the UUID, remaining slug, and whether it matched.
func matchFullUUID(name string) (id, slug string, ok bool) {
	// Full UUID is 36 chars: 8-4-4-4-12 with hyphens
	// Minimum filename: UUID alone (36 chars) or UUID-slug (37+ chars)
	if len(name) < 36 {
		return "", "", false
	}

	segments := []int{8, 4, 4, 4, 12}
	pos := 0
	for i, segLen := range segments {
		if i > 0 {
			if pos >= len(name) || name[pos] != '-' {
				return "", "", false
			}
			pos++
		}
		end := pos + segLen
		if end > len(name) {
			return "", "", false
		}
		if !isHexString(name[pos:end]) {
			return "", "", false
		}
		pos = end
	}

	uuid := name[:pos]
	rest := ""
	if pos < len(name) {
		if name[pos] != '-' {
			return "", "", false
		}
		rest = name[pos+1:]
	}
	return uuid, rest, true
}

// isHexID returns true if s is 9-32 lowercase hex chars.
// This catches truncated UUIDs longer than what isAlphanumericID handles (max 8).
func isHexID(s string) bool {
	if len(s) < 9 || len(s) > 32 {
		return false
	}
	return isHexString(s)
}

// isHexString returns true if every character in s is a lowercase hex digit (0-9, a-f).
func isHexString(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return len(s) > 0
}

// isAlphanumericID returns true if s is 3-8 lowercase alphanumeric chars with at least one digit.
// This avoids false-positives on English words like "readme".
func isAlphanumericID(s string) bool {
	if len(s) < 3 || len(s) > 8 {
		return false
	}
	hasDigit := false
	for _, c := range s {
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c < 'a' || c > 'z' {
			return false
		}
	}
	return hasDigit
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
