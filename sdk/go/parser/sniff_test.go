package parser

import "testing"

func TestHasTaskSignature(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter string
		want        bool
	}{
		{"id key", `id: "003"`, true},
		{"status key", "status: pending", true},
		{"priority key", "priority: medium", true},
		{"dependencies key", "dependencies: []", true},
		{"indented key", "  status: pending", true},
		{"key with no space", "priority:high", true},
		{"jekyll frontmatter", "layout: post\nauthor: someone", false},
		{"key as substring", "grid: true\nvalidity: ok", false},
		{"key mid-line", "note: the id: field", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasTaskSignature([]byte(tt.frontmatter)); got != tt.want {
				t.Errorf("HasTaskSignature(%q) = %v, want %v", tt.frontmatter, got, tt.want)
			}
		})
	}
}

func TestParseError_MalformedTaskFrontmatter(t *testing.T) {
	withFrontmatter := &ParseError{Frontmatter: []byte("id: \"003\"\ntags: [\n  - broken\n]")}
	if !withFrontmatter.MalformedTaskFrontmatter() {
		t.Error("expected task-like malformed frontmatter to be detected")
	}

	foreign := &ParseError{Frontmatter: []byte("layout: [\n  - broken\n]")}
	if foreign.MalformedTaskFrontmatter() {
		t.Error("expected foreign frontmatter not to be flagged")
	}

	noFrontmatter := &ParseError{Message: "file is empty"}
	if noFrontmatter.MalformedTaskFrontmatter() {
		t.Error("expected error without frontmatter not to be flagged")
	}
}
