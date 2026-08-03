package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// claude-code-plugin-lite/SPEC_REFERENCE.md is a condensed, hand-written subset
// of the canonical spec (docs/taskmd_specification.md). Unlike the embedded CLI
// template and the docs-site copy (guarded byte-for-byte by
// TestSpecTemplate_MatchesCanonicalSpec), it cannot be compared verbatim because
// it restructures and omits content. But its machine-critical facts — the enum
// value sets and the required-field list — MUST NOT contradict the canonical
// spec. These tests guard against that specific drift.
//
// If canonical enums or required fields change, update SPEC_REFERENCE.md in the
// same commit so these tests pass.

var (
	// A table cell that is purely a comma-separated list of backticked tokens,
	// e.g. "`pending`, `in-progress`, `completed`". Used to locate the enum
	// column independently of column order between the two differently-shaped
	// spec tables.
	backtickListCell = regexp.MustCompile("^(`[^`]+`(,\\s*|\\s*$))+$")
	backtickToken    = regexp.MustCompile("`([^`]+)`")
)

// enumFields are the frontmatter fields whose allowed values validation enforces.
var enumFields = []string{"status", "priority", "effort", "type"}

func TestSpecReference_EnumsMatchCanonical(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	canonical := readSpecFile(t, filepath.Join(repoRoot, "docs", "taskmd_specification.md"))
	lite := readSpecFile(t, filepath.Join(repoRoot, "claude-code-plugin-lite", "SPEC_REFERENCE.md"))

	for _, field := range enumFields {
		canonicalValues := enumValuesForField(canonical, field)
		liteValues := enumValuesForField(lite, field)

		if len(canonicalValues) == 0 {
			t.Errorf("no enum values found for %q in canonical spec; the extractor or spec layout changed", field)
			continue
		}
		if len(liteValues) == 0 {
			t.Errorf("no enum values found for %q in SPEC_REFERENCE.md; the extractor or reference layout changed", field)
			continue
		}
		if !equalStringSets(canonicalValues, liteValues) {
			t.Errorf("enum drift for %q:\n  canonical: %v\n  lite:      %v\n"+
				"claude-code-plugin-lite/SPEC_REFERENCE.md must list the same %s values as docs/taskmd_specification.md.",
				field, canonicalValues, liteValues, field)
		}
	}
}

func TestSpecReference_RequiredFieldsMatchCanonical(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	canonical := readSpecFile(t, filepath.Join(repoRoot, "docs", "taskmd_specification.md"))
	lite := readSpecFile(t, filepath.Join(repoRoot, "claude-code-plugin-lite", "SPEC_REFERENCE.md"))

	canonicalReq := canonicalRequiredFields(canonical)
	liteReq := liteRequiredFields(lite)

	if len(canonicalReq) == 0 {
		t.Fatal("no required fields found in canonical spec; the extractor or spec layout changed")
	}
	if !equalStringSets(canonicalReq, liteReq) {
		t.Errorf("required-field drift:\n  canonical: %v\n  lite:      %v\n"+
			"claude-code-plugin-lite/SPEC_REFERENCE.md must mark the same fields required as docs/taskmd_specification.md.",
			canonicalReq, liteReq)
	}
}

// enumValuesForField locates a field's definition row (first cell == `field`)
// and returns the backticked values from the cell that is a comma-separated list
// of backticked tokens. This is independent of column order, so it works against
// both the canonical Field Summary table and the lite Optional Fields table.
func enumValuesForField(content, field string) []string {
	target := "`" + field + "`"
	for _, line := range strings.Split(content, "\n") {
		cells := tableCells(line)
		if len(cells) < 2 || cells[0] != target {
			continue
		}
		for _, cell := range cells[1:] {
			if backtickListCell.MatchString(cell) {
				return backtickTokens(cell)
			}
		}
	}
	return nil
}

// canonicalRequiredFields reads the "Field Summary" table and returns the fields
// whose Required column contains "Yes".
func canonicalRequiredFields(content string) []string {
	sec := section(content, "### Field Summary", "## ")
	var out []string
	for _, line := range strings.Split(sec, "\n") {
		cells := tableCells(line)
		if len(cells) < 3 {
			continue
		}
		name := strings.Trim(cells[0], "`")
		if name == cells[0] || name == "" { // not a backticked field row
			continue
		}
		if strings.Contains(cells[2], "Yes") {
			out = append(out, name)
		}
	}
	return out
}

// liteRequiredFields reads the "Required Fields" section table and returns every
// field it lists (the whole section enumerates required fields).
func liteRequiredFields(content string) []string {
	sec := section(content, "### Required Fields", "### ")
	var out []string
	for _, line := range strings.Split(sec, "\n") {
		cells := tableCells(line)
		if len(cells) == 0 {
			continue
		}
		name := strings.Trim(cells[0], "`")
		if name == cells[0] || name == "" { // not a backticked field row
			continue
		}
		out = append(out, name)
	}
	return out
}

// section returns the lines after startHeading up to (but excluding) the next
// line whose trimmed text starts with endPrefix. Returns "" if not found.
func section(content, startHeading, endPrefix string) string {
	lines := strings.Split(content, "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == startHeading {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return ""
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), endPrefix) {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

// tableCells splits a markdown table row into trimmed cell values, dropping the
// empty strings produced by the leading and trailing pipes. Returns nil for
// non-table lines.
func tableCells(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") {
		return nil
	}
	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	if len(cells) > 0 && cells[0] == "" {
		cells = cells[1:]
	}
	if len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

func readSpecFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("skipping: spec file not found at %s", path)
	}
	return string(data)
}

func backtickTokens(s string) []string {
	matches := backtickToken.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

func equalStringSets(a, b []string) bool {
	as, bs := toStringSet(a), toStringSet(b)
	if len(as) != len(bs) {
		return false
	}
	for k := range as {
		if !bs[k] {
			return false
		}
	}
	return true
}

func toStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}
