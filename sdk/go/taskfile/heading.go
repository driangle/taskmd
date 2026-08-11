package taskfile

import (
	"fmt"
	"os"
	"strings"
)

// Heading describes the level-1 markdown heading in a task file's body — the
// line that conventionally mirrors the frontmatter title.
type Heading struct {
	Found bool   // whether the body has a level-1 heading at all
	Line  int    // 0-based line index of the heading, or -1 when not found
	Text  string // heading text with the leading "# " stripped
}

// FindTitleHeading returns the first level-1 heading in a task file's body.
//
// Headings inside fenced code blocks are skipped: a `# comment` line in a shell
// snippet is not a title.
func FindTitleHeading(filePath string) (Heading, error) {
	lines, closeIdx, err := readBodyLines(filePath)
	if err != nil {
		return Heading{Line: -1}, err
	}
	return findHeading(lines, closeIdx+1), nil
}

// RewriteTitleHeading replaces the body's level-1 heading with newTitle, but
// only when the existing heading matches oldTitle. It reports whether the
// heading was rewritten.
//
// A heading that has diverged from the frontmatter title is a deliberate choice
// by the author, so it is left alone rather than silently overwritten.
func RewriteTitleHeading(filePath, oldTitle, newTitle string) (bool, error) {
	lines, closeIdx, err := readBodyLines(filePath)
	if err != nil {
		return false, err
	}

	h := findHeading(lines, closeIdx+1)
	if !h.Found || h.Text != oldTitle {
		return false, nil
	}

	lines[h.Line] = "# " + newTitle
	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return false, fmt.Errorf("failed to write task file: %w", err)
	}
	return true, nil
}

// readBodyLines reads a task file and returns its lines plus the index of the
// closing frontmatter delimiter.
func readBodyLines(filePath string) ([]string, int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	openIdx, closeIdx := FindFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return nil, 0, fmt.Errorf("task file has no valid frontmatter: %s", filePath)
	}
	return lines, closeIdx, nil
}

// findHeading scans from start for the first level-1 heading outside any fenced
// code block.
func findHeading(lines []string, start int) Heading {
	inFence := false
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence || !strings.HasPrefix(lines[i], "# ") {
			continue
		}
		return Heading{
			Found: true,
			Line:  i,
			Text:  strings.TrimSpace(strings.TrimPrefix(lines[i], "# ")),
		}
	}
	return Heading{Line: -1}
}
