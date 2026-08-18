package main

import (
	"fmt"
	"regexp"
	"strings"
)

// placeholderPatterns are the template leftovers a filled-in task must not keep.
var placeholderPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)<!--`),
	regexp.MustCompile(`(?m)^\s*(?:-\s*\[[ x]\]\s*)?TODO\s*$`),
	regexp.MustCompile(`(?m)^\s*-\s*TODO\s*$`),
	regexp.MustCompile(`(?m)^\s*1\.\s*\.\.\.\s*$`),
}

// noPlaceholders fails if the task file still carries unedited template content.
func noPlaceholders(body string) error {
	for _, re := range placeholderPatterns {
		if match := re.FindString(body); match != "" {
			return fmt.Errorf("task file still contains placeholder content: %q", strings.TrimSpace(match))
		}
	}
	return nil
}

// summaryHeadings are the interchangeable names agents give the "what is this
// task about" section. The spec only constrains frontmatter, so grading the
// heading word instead of the content would penalize valid output.
var summaryHeadings = []string{"Objective", "Description", "Summary", "Overview", "Context"}

// summaryFilled asserts the task explains itself somewhere near the top, under
// whichever of the conventional headings the agent chose.
func summaryFilled(body string, minChars int) error {
	for _, heading := range summaryHeadings {
		if err := sectionFilled(body, heading, minChars); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no filled summary section (looked for %v with >= %d chars)", summaryHeadings, minChars)
}

// sectionFilled asserts a `## <heading>` section exists and carries at least
// minChars of non-whitespace prose.
func sectionFilled(body, heading string, minChars int) error {
	content, ok := section(body, heading)
	if !ok {
		return fmt.Errorf("task file has no %q section", "## "+heading)
	}

	dense := strings.Join(strings.Fields(content), "")
	if len(dense) < minChars {
		return fmt.Errorf("section %q is too thin (%d chars, want >= %d)", heading, len(dense), minChars)
	}
	return nil
}

// section returns the body of the named `##` heading, case-insensitively.
func section(body, heading string) (string, bool) {
	var (
		out    []string
		inside bool
		want   = strings.ToLower("## " + heading)
	)

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(trimmed, want):
			inside = true
		case strings.HasPrefix(trimmed, "## "):
			if inside {
				return strings.Join(out, "\n"), true
			}
		case inside:
			out = append(out, line)
		}
	}

	if inside {
		return strings.Join(out, "\n"), true
	}
	return "", false
}

// hasSubtasks asserts the Tasks section contains at least n checklist items.
func hasSubtasks(body string, n int) error {
	content, ok := section(body, "Tasks")
	if !ok {
		return fmt.Errorf("task file has no %q section", "## Tasks")
	}

	items := regexp.MustCompile(`(?m)^\s*-\s*\[[ x]\]\s*\S+`).FindAllString(content, -1)
	if len(items) < n {
		return fmt.Errorf("Tasks section has %d checklist items, want >= %d", len(items), n)
	}
	return nil
}

func hasTags(t Task, want ...string) error {
	have := make(map[string]bool, len(t.Tags))
	for _, tag := range t.Tags {
		have[strings.ToLower(tag)] = true
	}

	for _, tag := range want {
		if !have[strings.ToLower(tag)] {
			return fmt.Errorf("task tags %v are missing %q", t.Tags, tag)
		}
	}
	return nil
}

func hasDependency(t Task, want string) error {
	for _, dep := range t.Dependencies {
		if dep == want {
			return nil
		}
	}
	return fmt.Errorf("task dependencies %v do not include %q", t.Dependencies, want)
}

func equals(field, got, want string) error {
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("%s is %q, want %q", field, got, want)
	}
	return nil
}

// titleMentions asserts the title contains at least one of the given keywords,
// so a task about the wrong subject fails even when the metadata is right.
func titleMentions(t Task, keywords ...string) error {
	title := strings.ToLower(t.Title)
	for _, kw := range keywords {
		if strings.Contains(title, strings.ToLower(kw)) {
			return nil
		}
	}
	return fmt.Errorf("title %q mentions none of %v", t.Title, keywords)
}
