package taskfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTaskFile writes content to a temp file and returns its path.
func writeTaskFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "001-task.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return path
}

const headingFixture = `---
id: "001"
title: "Setup project"
---

# Setup project

Initial setup.
`

func TestFindTitleHeading(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantFound bool
		wantText  string
	}{
		{
			name:      "heading matching title",
			content:   headingFixture,
			wantFound: true,
			wantText:  "Setup project",
		},
		{
			name: "no heading in body",
			content: `---
id: "001"
title: "Setup project"
---

Just prose, no heading.
`,
			wantFound: false,
		},
		{
			name: "only sub-headings",
			content: `---
id: "001"
title: "Setup project"
---

## Objective

Details.
`,
			wantFound: false,
		},
		{
			name: "heading inside a fenced block is skipped",
			content: `---
id: "001"
title: "Setup project"
---

` + "```bash" + `
# not a heading, a shell comment
` + "```" + `

# Real heading
`,
			wantFound: true,
			wantText:  "Real heading",
		},
		{
			name: "tilde fence is skipped too",
			content: `---
id: "001"
title: "Setup project"
---

~~~
# fenced
~~~

# Real heading
`,
			wantFound: true,
			wantText:  "Real heading",
		},
		{
			name: "trailing whitespace is trimmed",
			content: `---
id: "001"
title: "Setup project"
---

# Setup project
`,
			wantFound: true,
			wantText:  "Setup project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindTitleHeading(writeTaskFile(t, tt.content))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Found != tt.wantFound {
				t.Errorf("Found = %v, want %v", got.Found, tt.wantFound)
			}
			if got.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", got.Text, tt.wantText)
			}
			if !got.Found && got.Line != -1 {
				t.Errorf("Line = %d, want -1 when not found", got.Line)
			}
		})
	}
}

func TestFindTitleHeading_NoFrontmatter(t *testing.T) {
	path := writeTaskFile(t, "# Just a heading\n")
	if _, err := FindTitleHeading(path); err == nil {
		t.Fatal("expected an error for a file without frontmatter")
	}
}

func TestRewriteTitleHeading_Matching(t *testing.T) {
	path := writeTaskFile(t, headingFixture)

	rewritten, err := RewriteTitleHeading(path, "Setup project", "Bootstrap the project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rewritten {
		t.Fatal("expected the heading to be rewritten")
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "# Bootstrap the project") {
		t.Errorf("heading not rewritten, got:\n%s", content)
	}
	// The frontmatter is UpdateTaskFile's job — this call must not touch it.
	if !strings.Contains(string(content), `title: "Setup project"`) {
		t.Errorf("frontmatter should be untouched, got:\n%s", content)
	}
	if !strings.Contains(string(content), "Initial setup.") {
		t.Errorf("body content should be preserved, got:\n%s", content)
	}
}

func TestRewriteTitleHeading_NonMatchingIsLeftAlone(t *testing.T) {
	path := writeTaskFile(t, `---
id: "001"
title: "Setup project"
---

# Task 001: a deliberately custom heading

Initial setup.
`)

	rewritten, err := RewriteTitleHeading(path, "Setup project", "Bootstrap the project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rewritten {
		t.Fatal("expected a non-matching heading to be left alone")
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "# Task 001: a deliberately custom heading") {
		t.Errorf("custom heading was clobbered, got:\n%s", content)
	}
}

func TestRewriteTitleHeading_NoHeading(t *testing.T) {
	path := writeTaskFile(t, `---
id: "001"
title: "Setup project"
---

Just prose.
`)
	before, _ := os.ReadFile(path)

	rewritten, err := RewriteTitleHeading(path, "Setup project", "Bootstrap the project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rewritten {
		t.Fatal("expected no rewrite when the body has no heading")
	}

	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("file changed despite no heading:\n%s", after)
	}
}
