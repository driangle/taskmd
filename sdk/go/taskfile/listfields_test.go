package taskfile

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Regression coverage for the multi-line flow sequence corruption: a list field
// whose "[" opens on a line *after* its key used to fall through to the block
// branch, which spliced new "- item" lines in ahead of the untouched flow node.
//
// Every case asserts three things, per the bug report:
//
//  1. the frontmatter still parses as YAML;
//  2. the field reads back as the expected list;
//  3. no fragment of the old value survives — checked by asserting the
//     neighbouring scalars are untouched, since an orphan left behind by a
//     key-removing path folds into the scalar above it and silently corrupts it.
//
// Assertion 3 is the one that catches the "file still parses, status is now
// `pending [\"a\"]`" variant, which assertion 1 alone sails straight past.

// buildTaskFile frames a field block with scalars on both sides so an orphan has
// something to corrupt.
func buildTaskFile(fieldBlock string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("id: \"001\"\n")
	b.WriteString("title: \"Target\"\n")
	b.WriteString("status: pending\n")
	if fieldBlock != "" {
		b.WriteString(fieldBlock + "\n")
	}
	b.WriteString("created_at: \"2026-09-08\"\n")
	b.WriteString("---\n\n# Target\n")
	return b.String()
}

// readFrontmatter parses the frontmatter of a written task file, failing the
// test if it is no longer valid YAML.
func readFrontmatter(t *testing.T, path string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	lines := strings.Split(string(content), "\n")
	openIdx, closeIdx := FindFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		t.Fatalf("frontmatter delimiters missing after update:\n%s", content)
	}

	var fm map[string]any
	raw := strings.Join(lines[openIdx+1:closeIdx], "\n")
	if err := yaml.Unmarshal([]byte(raw), &fm); err != nil {
		t.Fatalf("frontmatter is not valid YAML after update: %v\n--- written file ---\n%s", err, content)
	}
	return fm
}

// assertListField checks the parsed value of a list field against want. A nil
// want means the key must be absent.
func assertListField(t *testing.T, fm map[string]any, field string, want []string, path string) {
	t.Helper()
	raw, present := fm[field]
	if want == nil {
		if present {
			t.Errorf("expected %q to be absent, got %#v", field, raw)
		}
		return
	}
	if !present {
		content, _ := os.ReadFile(path)
		t.Fatalf("expected %q to be present, frontmatter was:\n%s", field, content)
	}

	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected %q to parse as a list, got %T (%#v)", field, raw, raw)
	}
	got := make([]string, len(items))
	for i, it := range items {
		got[i] = fmt.Sprintf("%v", it)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %#v, want %#v", field, got, want)
	}
}

// assertNeighboursIntact is invariant 3: a leftover flow node folds into the
// scalar above it, so unchanged neighbours prove nothing was orphaned.
func assertNeighboursIntact(t *testing.T, fm map[string]any, path string) {
	t.Helper()
	for key, want := range map[string]string{
		"id":         "001",
		"title":      "Target",
		"status":     "pending",
		"created_at": "2026-09-08",
	} {
		if got := fmt.Sprintf("%v", fm[key]); got != want {
			content, _ := os.ReadFile(path)
			t.Errorf("neighbouring field %q was corrupted: got %q, want %q\n--- written file ---\n%s",
				key, got, want, content)
		}
	}
}

func TestUpdateTaskFile_ListFieldShapes_Dependencies(t *testing.T) {
	deps := func(v ...string) *[]string { s := v; return &s }

	tests := []struct {
		name  string
		block string
		set   *[]string
		want  []string
	}{
		{
			name:  "single-line flow",
			block: `dependencies: ["a"]`,
			set:   deps("a", "b"),
			want:  []string{"a", "b"},
		},
		{
			name:  "multi-line flow",
			block: "dependencies:\n  [\"a\"]",
			set:   deps("a", "b"),
			want:  []string{"a", "b"},
		},
		{
			name:  "multi-line flow with trailing comma (prettier)",
			block: "dependencies:\n  [\n    \"a\",\n  ]",
			set:   deps("a", "b"),
			want:  []string{"a", "b"},
		},
		{
			name:  "multi-line flow spanning several entries",
			block: "dependencies:\n  [\n    \"a\",\n    \"b\",\n    \"c\",\n  ]",
			set:   deps("a", "d"),
			want:  []string{"a", "d"},
		},
		{
			name:  "block sequence",
			block: "dependencies:\n  - a",
			set:   deps("a", "b"),
			want:  []string{"a", "b"},
		},
		{
			name:  "empty flow",
			block: `dependencies: []`,
			set:   deps("a"),
			want:  []string{"a"},
		},
		{
			name:  "key absent",
			block: "",
			set:   deps("a"),
			want:  []string{"a"},
		},
		{
			name:  "clear single-line flow",
			block: `dependencies: ["a"]`,
			set:   deps(),
			want:  nil,
		},
		{
			name:  "clear multi-line flow leaves no orphan",
			block: "dependencies:\n  [\"a\"]",
			set:   deps(),
			want:  nil,
		},
		{
			name:  "clear multi-line flow with trailing comma leaves no orphan",
			block: "dependencies:\n  [\n    \"a\",\n    \"b\",\n  ]",
			set:   deps(),
			want:  nil,
		},
		{
			name:  "clear block sequence",
			block: "dependencies:\n  - a\n  - b",
			set:   deps(),
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTestFile(t, buildTaskFile(tt.block))

			if err := UpdateTaskFile(path, UpdateRequest{Dependencies: tt.set}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			fm := readFrontmatter(t, path)
			assertListField(t, fm, "dependencies", tt.want, path)
			assertNeighboursIntact(t, fm, path)
		})
	}
}

// The block form is the one shape whose layout is preserved rather than
// normalised to flow, so assert the written text and not just the parsed value.
func TestUpdateTaskFile_BlockSequencePreservesShape(t *testing.T) {
	path := createTestFile(t, buildTaskFile("dependencies:\n  - a"))

	deps := []string{"a", "b"}
	if err := UpdateTaskFile(path, UpdateRequest{Dependencies: &deps}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "dependencies:\n  - a\n  - b") {
		t.Errorf("expected block sequence to be preserved, got:\n%s", content)
	}
}

// touches, pr and tags reach the writer through the add/remove merge path, which
// has to see the existing multi-line flow value or it silently drops it.
func TestUpdateTaskFile_ListFieldShapes_AddRemove(t *testing.T) {
	tests := []struct {
		name  string
		field string
		block string
		req   UpdateRequest
		want  []string
	}{
		{
			name:  "add to multi-line flow touches keeps existing value",
			field: "touches",
			block: "touches:\n  [\"cli\"]",
			req:   UpdateRequest{AddTouches: []string{"web"}},
			want:  []string{"cli", "web"},
		},
		{
			name:  "add to prettier-wrapped touches keeps existing values",
			field: "touches",
			block: "touches:\n  [\n    \"cli\",\n    \"web\",\n  ]",
			req:   UpdateRequest{AddTouches: []string{"docs"}},
			want:  []string{"cli", "web", "docs"},
		},
		{
			name:  "remove from multi-line flow touches",
			field: "touches",
			block: "touches:\n  [\"cli\", \"web\"]",
			req:   UpdateRequest{RemTouches: []string{"cli"}},
			want:  []string{"web"},
		},
		{
			name:  "remove last multi-line flow touches leaves no orphan",
			field: "touches",
			block: "touches:\n  [\"cli\"]",
			req:   UpdateRequest{RemTouches: []string{"cli"}},
			want:  nil,
		},
		{
			name:  "add to multi-line flow pr keeps existing value",
			field: "pr",
			block: "pr:\n  [\"u1\"]",
			req:   UpdateRequest{AddPRs: []string{"u2"}},
			want:  []string{"u1", "u2"},
		},
		{
			name:  "add to multi-line flow tags keeps existing value",
			field: "tags",
			block: "tags:\n  [\"alpha\"]",
			req:   UpdateRequest{AddTags: []string{"beta"}},
			want:  []string{"alpha", "beta"},
		},
		{
			name:  "add to prettier-wrapped tags keeps existing values",
			field: "tags",
			block: "tags:\n  [\n    \"alpha\",\n    \"beta\",\n  ]",
			req:   UpdateRequest{AddTags: []string{"gamma"}},
			want:  []string{"alpha", "beta", "gamma"},
		},
		{
			name:  "remove from multi-line flow tags",
			field: "tags",
			block: "tags:\n  [\"alpha\", \"beta\"]",
			req:   UpdateRequest{RemTags: []string{"alpha"}},
			want:  []string{"beta"},
		},
		{
			name:  "add to single-line flow touches is unchanged",
			field: "touches",
			block: `touches: ["cli"]`,
			req:   UpdateRequest{AddTouches: []string{"web"}},
			want:  []string{"cli", "web"},
		},
		{
			name:  "add to block sequence touches is unchanged",
			field: "touches",
			block: "touches:\n  - cli",
			req:   UpdateRequest{AddTouches: []string{"web"}},
			want:  []string{"cli", "web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTestFile(t, buildTaskFile(tt.block))

			if err := UpdateTaskFile(path, tt.req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			fm := readFrontmatter(t, path)
			assertListField(t, fm, tt.field, tt.want, path)
			assertNeighboursIntact(t, fm, path)
		})
	}
}

// An empty tags result keeps the key as "tags: []" rather than removing it,
// unlike the other list fields.
func TestUpdateTaskFile_ClearTagsKeepsEmptyKey(t *testing.T) {
	path := createTestFile(t, buildTaskFile("tags:\n  [\"alpha\"]"))

	if err := UpdateTaskFile(path, UpdateRequest{RemTags: []string{"alpha"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fm := readFrontmatter(t, path)
	assertNeighboursIntact(t, fm, path)

	raw, present := fm["tags"]
	if !present {
		t.Fatal("expected tags key to be kept as an empty list")
	}
	if items, ok := raw.([]any); !ok || len(items) != 0 {
		t.Errorf("tags = %#v, want an empty list", raw)
	}
}

func TestFindListField_Shapes(t *testing.T) {
	tests := []struct {
		name       string
		block      string
		wantFound  bool
		wantShape  listShape
		wantValues []string
		wantSpan   int // lines the field owns, keyIdx..endIdx
	}{
		{"absent", "", false, shapeNone, nil, 0},
		{"single-line flow", `dependencies: ["a", "b"]`, true, shapeFlow, []string{"a", "b"}, 1},
		{"empty flow", `dependencies: []`, true, shapeFlow, nil, 1},
		{"multi-line flow", "dependencies:\n  [\"a\"]", true, shapeFlow, []string{"a"}, 2},
		{"prettier flow", "dependencies:\n  [\n    \"a\",\n    \"b\",\n  ]", true, shapeFlow, []string{"a", "b"}, 5},
		{"block sequence", "dependencies:\n  - a\n  - b", true, shapeBlock, []string{"a", "b"}, 3},
		{"bare key", "dependencies:", true, shapeNone, nil, 1},
		{"value containing a bracket", `dependencies: ["a[1]", "b"]`, true, shapeFlow, []string{"a[1]", "b"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(buildTaskFile(tt.block), "\n")
			openIdx, closeIdx := FindFrontmatterBounds(lines)

			f, found := findListField(lines, openIdx, closeIdx, "dependencies")
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}
			if !found {
				return
			}
			if f.shape != tt.wantShape {
				t.Errorf("shape = %v, want %v", f.shape, tt.wantShape)
			}
			if !reflect.DeepEqual(f.values, tt.wantValues) {
				t.Errorf("values = %#v, want %#v", f.values, tt.wantValues)
			}
			if span := f.endIdx - f.keyIdx; span != tt.wantSpan {
				t.Errorf("span = %d lines, want %d", span, tt.wantSpan)
			}
		})
	}
}

// The write-back guard is a backstop for writer bugs of this class: if an update
// would render the frontmatter unreadable, the write is refused rather than
// leaving a task that silently vanishes from every view.
func TestVerifyFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid frontmatter",
			content: "---\nid: \"001\"\nstatus: pending\n---\n\n# T\n",
		},
		{
			name:    "valid list field",
			content: "---\nid: \"001\"\ndependencies: [\"a\", \"b\"]\n---\n\n# T\n",
		},
		{
			// Exactly the corruption this bug produced.
			name:    "orphaned flow node after a block sequence",
			content: "---\nid: \"001\"\ndependencies:\n  - a\n  [\"a\"]\n---\n\n# T\n",
			wantErr: true,
		},
		{
			name:    "missing delimiters",
			content: "# Just a heading\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyFrontmatter(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
