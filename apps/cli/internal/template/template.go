package template

import (
	"fmt"
	"sort"
	"strings"
)

// Template represents a parsed task template.
type Template struct {
	Name        string
	Description string
	Source      string // "built-in", "user", or "project"
	Content     string // raw template content including frontmatter
}

// ParseTemplate extracts the _template metadata from a template file's frontmatter.
// The content must have YAML frontmatter delimited by "---".
func ParseTemplate(content, source string) (Template, error) {
	lines := strings.Split(content, "\n")
	openIdx, closeIdx := findFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return Template{}, fmt.Errorf("template has no valid frontmatter")
	}

	name, desc := extractTemplateMetadata(lines, openIdx, closeIdx)
	if name == "" {
		return Template{}, fmt.Errorf("template missing _template.name in frontmatter")
	}

	return Template{
		Name:        name,
		Description: desc,
		Source:      source,
		Content:     content,
	}, nil
}

// RenderTask renders a template into task file content by stripping the _template
// block from frontmatter and substituting variables.
func RenderTask(tmpl Template, vars map[string]string) string {
	lines := strings.Split(tmpl.Content, "\n")
	openIdx, closeIdx := findFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return substituteVars(tmpl.Content, vars)
	}

	stripped := stripTemplateBlock(lines, openIdx, closeIdx)
	result := strings.Join(stripped, "\n")
	return substituteVars(result, vars)
}

// ApplyOverrides replaces frontmatter field values using line-based matching.
// Only fields explicitly provided in overrides are replaced.
func ApplyOverrides(content string, overrides map[string]string) string {
	if len(overrides) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	openIdx, closeIdx := findFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return content
	}

	matched := make(map[string]bool, len(overrides))
	for i := openIdx + 1; i < closeIdx; i++ {
		for key, value := range overrides {
			prefix := key + ":"
			if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
				lines[i] = key + ": " + value
				matched[key] = true
				break
			}
		}
	}

	// Append overrides that had no matching line so they aren't silently dropped
	// (e.g. --phase against a template whose frontmatter declares no phase field).
	appended := make([]string, 0, len(overrides))
	for key := range overrides {
		if !matched[key] {
			appended = append(appended, key)
		}
	}
	sort.Strings(appended)

	insert := make([]string, len(appended))
	for i, key := range appended {
		insert[i] = key + ": " + overrides[key]
	}
	lines = append(lines[:closeIdx], append(insert, lines[closeIdx:]...)...)

	return strings.Join(lines, "\n")
}

// substituteVars performs simple {{var}} replacement.
func substituteVars(content string, vars map[string]string) string {
	for key, value := range vars {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return content
}

// extractTemplateMetadata reads _template.name and _template.description from frontmatter.
func extractTemplateMetadata(lines []string, openIdx, closeIdx int) (name, description string) {
	inBlock := false
	for i := openIdx + 1; i < closeIdx; i++ {
		trimmed := strings.TrimSpace(lines[i])

		if trimmed == "_template:" {
			inBlock = true
			continue
		}

		if inBlock {
			// Check if we've left the _template block (line not indented)
			if !strings.HasPrefix(lines[i], " ") && !strings.HasPrefix(lines[i], "\t") {
				break
			}
			if strings.HasPrefix(trimmed, "name:") {
				name = unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")))
			}
			if strings.HasPrefix(trimmed, "description:") {
				description = unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "description:")))
			}
		}
	}
	return name, description
}

// stripTemplateBlock removes the _template: block lines from frontmatter.
func stripTemplateBlock(lines []string, openIdx, closeIdx int) []string {
	var result []string
	inBlock := false

	for i, line := range lines {
		if i > openIdx && i < closeIdx {
			trimmed := strings.TrimSpace(line)
			if trimmed == "_template:" {
				inBlock = true
				continue
			}
			if inBlock {
				if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
					continue
				}
				inBlock = false
			}
		}
		result = append(result, line)
	}

	return result
}

// unquote strips surrounding quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

// findFrontmatterBounds returns the line indices of the opening and closing "---".
func findFrontmatterBounds(lines []string) (int, int) {
	openIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if openIdx < 0 {
				openIdx = i
			} else {
				return openIdx, i
			}
		}
	}
	return -1, -1
}
