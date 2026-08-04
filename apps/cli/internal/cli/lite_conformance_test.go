package cli

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/nextid"
	"github.com/driangle/taskmd/sdk/go/slug"
	"github.com/driangle/taskmd/sdk/go/validator"
)

// claude-code-plugin-lite reimplements the CLI's file-generation behavior as
// English prose so the plugin can run without the Go binary. That prose is a
// SECOND copy of algorithms that already live in Go (ID generation, slug rules,
// the new-task frontmatter template, the validation enums), with no compiler to
// keep the two in step. These tests make the CLI the single source of truth:
// each fact is DERIVED from real CLI code (an SDK function, an in-package enum
// list, the actual frontmatter writer) and then asserted to be documented in the
// lite prose. If the CLI's behavior changes, the derived fact changes, the prose
// no longer matches, and the test fails — telling the contributor to update the
// prose in the same change.
//
// This complements spec_reference_test.go, which guards the lite SPEC_REFERENCE.md
// against the canonical spec document. Here we guard the lite *skills* against the
// CLI's runtime behavior.
//
// Duplicated behaviors covered (see docs/adr for the scope rationale):
//   1. ID generation — the four strategies and their shapes/charsets  (add-task)
//   2. ID defaults    — sequential zero-padding width                 (add-task)
//   3. Slug generation — lowercase / hyphenation / max length         (add-task)
//   4. Frontmatter template — always-written fields and defaults      (add-task)
//   5. Validation enums — status/priority/effort/type value sets      (validate-tasks, update-task)
//
// If a check fails, fix the named SKILL.md so its prose matches current CLI
// behavior — the CLI is authoritative (see claude-code-plugin-lite/README.md).

// liteProse holds the text of the lite skill files a conformance check reads.
// Files are loaded via readSpecFile, which t.Skip()s the whole test when the
// plugin directory is absent (mirrors spec_reference_test.go's tolerance).
type liteProse struct {
	addTask       string
	validateTasks string
	updateTask    string
}

func loadLiteProse(t *testing.T) liteProse {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..")
	skill := func(name string) string {
		return readSpecFile(t, filepath.Join(root, "claude-code-plugin-lite", "skills", name, "SKILL.md"))
	}
	return liteProse{
		addTask:       skill("add-task"),
		validateTasks: skill("validate-tasks"),
		updateTask:    skill("update-task"),
	}
}

// containsFold reports whether haystack contains needle, case-insensitively.
func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// assertProseDocuments fails when any of tokens is absent from prose.
func assertProseDocuments(t *testing.T, prose, file, behavior string, tokens ...string) {
	t.Helper()
	for _, tok := range tokens {
		if !containsFold(prose, tok) {
			t.Errorf("lite/CLI drift in %s (%s): prose does not document %q.\n"+
				"The CLI is authoritative — update the SKILL.md prose to match current CLI behavior.",
				file, behavior, tok)
		}
	}
}

// TestLiteConformance_IDStrategies exercises every real ID strategy through the
// same code path the `add` command uses (generateID → sdk/nextid), asserts the
// generated ID's shape/charset, and asserts the add-task prose documents that
// strategy. Running the actual generator makes the strategy set and charsets
// load-bearing rather than a hand-copied list.
func TestLiteConformance_IDStrategies(t *testing.T) {
	prose := loadLiteProse(t)

	// Crockford Base32 lowercase (sdk/go/nextid: excludes i, l, o, u).
	const crockford = "0123456789abcdefghjkmnpqrstvwxyz"

	cases := []struct {
		strategy string
		cfg      validator.IDConfig
		existing []string
		shape    *regexp.Regexp
		// tokens the add-task prose must document for this strategy.
		proseTokens []string
	}{
		{
			strategy:    "sequential",
			cfg:         validator.IDConfig{Strategy: "sequential", Padding: 3},
			existing:    []string{"005"},
			shape:       regexp.MustCompile(`^\d{3,}$`),
			proseTokens: []string{"sequential", "zero-pad"},
		},
		{
			strategy:    "prefixed",
			cfg:         validator.IDConfig{Strategy: "prefixed", Prefix: "dr-", Padding: 3},
			existing:    nil,
			shape:       regexp.MustCompile(`^dr-\d{3,}$`),
			proseTokens: []string{"prefixed", "prefix"},
		},
		{
			strategy:    "random",
			cfg:         validator.IDConfig{Strategy: "random", Length: 6},
			existing:    nil,
			shape:       regexp.MustCompile(`^[0-9a-z]{6}$`),
			proseTokens: []string{"random", "alphanumeric"},
		},
		{
			strategy:    "ulid",
			cfg:         validator.IDConfig{Strategy: "ulid", Length: 10},
			existing:    nil,
			shape:       regexp.MustCompile("^[" + crockford + "]{10}$"),
			proseTokens: []string{"ulid", "crockford"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.strategy, func(t *testing.T) {
			id, err := generateID(tc.existing, tc.cfg)
			if err != nil {
				t.Fatalf("generateID(%s) failed: %v", tc.strategy, err)
			}
			if !tc.shape.MatchString(id) {
				t.Fatalf("CLI %s ID %q does not match expected shape %s", tc.strategy, id, tc.shape)
			}
			assertProseDocuments(t, prose.addTask, "skills/add-task/SKILL.md",
				"ID strategy "+tc.strategy, tc.proseTokens...)
		})
	}
}

// TestLiteConformance_IDDefaults derives the sequential zero-padding width from
// the SDK (nextid.Calculate on an empty set yields the default padding) and
// asserts the add-task prose documents that same default. Sourcing the number
// from the SDK — not a literal — is what ties the prose to the CLI.
func TestLiteConformance_IDDefaults(t *testing.T) {
	prose := loadLiteProse(t)

	defaultPadding := nextid.Calculate(nil).Padding
	if defaultPadding <= 0 {
		t.Fatalf("nextid.Calculate(nil).Padding = %d, want > 0", defaultPadding)
	}

	assertProseDocuments(t, prose.addTask, "skills/add-task/SKILL.md",
		"default sequential padding width",
		"default "+strconv.Itoa(defaultPadding))
}

// TestLiteConformance_Slug pins the CLI's slug algorithm with golden fixtures
// (input → expected), then asserts the add-task prose documents the load-bearing
// rules. The max-length fact is derived from slug.Slugify itself so the prose
// stays tied to the real cap rather than a hand-copied number.
func TestLiteConformance_Slug(t *testing.T) {
	prose := loadLiteProse(t)

	fixtures := []struct{ in, want string }{
		{"Fix the Login Bug!", "fix-the-login-bug"},
		{"Hello,   World @ 2026", "hello-world-2026"},
		{"  Trailing & Leading  ", "trailing-leading"},
		{"UPPER_snake.Mix", "upper-snake-mix"},
	}
	for _, f := range fixtures {
		if got := slug.Slugify(f.in); got != f.want {
			t.Errorf("slug.Slugify(%q) = %q, want %q (CLI slug behavior changed)", f.in, got, f.want)
		}
	}

	// Derive the max slug length from CLI behavior rather than hardcoding it.
	// Use a hyphen-free input so trailing-hyphen trimming can't shorten the cap.
	maxLen := len(slug.Slugify(strings.Repeat("a", 200)))
	if maxLen == 0 {
		t.Fatal("could not derive slug max length from slug.Slugify")
	}

	assertProseDocuments(t, prose.addTask, "skills/add-task/SKILL.md",
		"slug generation rules", "lowercase", "hyphen", strconv.Itoa(maxLen))
}

// TestLiteConformance_Frontmatter generates a task file with the CLI's own
// writer (buildTaskFileContent, the no-template path used by `add`) and asserts
// the add-task prose documents every field the CLI ALWAYS writes, plus the
// hardcoded status default. If the CLI starts always-writing a new field, the
// derived key set grows and the prose must catch up.
func TestLiteConformance_Frontmatter(t *testing.T) {
	prose := loadLiteProse(t)

	// buildTaskFileContent reads add* flag globals; reset them to flag defaults
	// (status=pending, priority=medium, the rest empty) so the output reflects a
	// plain `taskmd add "<title>"`. Not parallel-safe: mutates shared globals.
	resetCLIState()
	content := buildTaskFileContent("001", "Sample Title")

	keys := frontmatterKeys(content)
	if len(keys) == 0 {
		t.Fatal("no frontmatter keys parsed from buildTaskFileContent output")
	}

	for _, key := range keys {
		behavior := "always-written frontmatter field " + key
		// The CLI writes created_at; the lite prose documents its sanctioned
		// alias `created` (both are read by the CLI parser), so accept either.
		if key == "created_at" {
			if !containsFold(prose.addTask, "created:") && !containsFold(prose.addTask, "created_at:") {
				t.Errorf("lite/CLI drift in skills/add-task/SKILL.md (%s): "+
					"prose documents neither `created:` nor `created_at:`.", behavior)
			}
			continue
		}
		assertProseDocuments(t, prose.addTask, "skills/add-task/SKILL.md", behavior, key+":")
	}

	// The status default is hardcoded by the CLI writer, not just a flag default.
	assertProseDocuments(t, prose.addTask, "skills/add-task/SKILL.md",
		"default status for a new task", "status: pending")
}

// frontmatterKeys returns the top-level YAML keys inside the leading `---` block.
func frontmatterKeys(content string) []string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}
	keyRe := regexp.MustCompile(`^([a-z_]+):`)
	var keys []string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		if m := keyRe.FindStringSubmatch(line); m != nil {
			keys = append(keys, m[1])
		}
	}
	return keys
}

// TestLiteConformance_Enums asserts that every enum value list the lite skills
// spell out in prose is EXACTLY the value set the CLI accepts (the in-package
// valid*Values lists used by `add`/`set` validation). It set-compares each
// prose list against the CLI set, so both a value dropped from the prose and a
// value added only to the CLI are caught — a whole-file substring check would
// miss a value that still appears in an example elsewhere in the file.
func TestLiteConformance_Enums(t *testing.T) {
	prose := loadLiteProse(t)

	enums := []struct {
		field  string
		values []string
	}{
		{"status", validStatusValues},
		{"priority", validPriorityValues},
		{"effort", validEffortValues},
		{"type", validTypeValues},
	}

	files := []struct {
		name  string
		prose string
	}{
		{"skills/validate-tasks/SKILL.md", prose.validateTasks},
		{"skills/update-task/SKILL.md", prose.updateTask},
	}

	for _, e := range enums {
		if len(e.values) == 0 {
			t.Fatalf("no CLI values for enum %q; the source list moved", e.field)
		}
		for _, f := range files {
			lists := proseEnumLists(f.prose, e.field)
			if len(lists) == 0 {
				t.Errorf("lite/CLI drift in %s: no %q enum value list found in prose.\n"+
					"The CLI is authoritative — document its %s values.", f.name, e.field, e.field)
				continue
			}
			for _, got := range lists {
				if !equalStringSets(got, e.values) {
					t.Errorf("lite/CLI drift in %s (%s enum):\n  prose: %v\n  CLI:   %v\n"+
						"Every %s value list in the prose must match the CLI set.",
						f.name, e.field, got, e.values, e.field)
				}
			}
		}
	}
}

// enumMarkerRe captures the comma-separated list following an enum marker phrase
// ("one of:", "valid values:", "valid:") on a single prose line.
var enumMarkerRe = regexp.MustCompile(`(?i)(?:one of|valid values|valid)\s*:\s*(.+)$`)

// enumTokenRe matches a real enum value (lowercase, hyphen-joined), so template
// placeholders like "<list>" and prose fragments are ignored.
var enumTokenRe = regexp.MustCompile(`^[a-z][a-z-]*$`)

// proseEnumLists returns every comma-separated value list found on lines that
// mention field and carry an enum marker. Each returned list is the parsed set
// of real enum tokens on that line (only lists of 2+ tokens are kept).
func proseEnumLists(prose, field string) [][]string {
	var out [][]string
	for _, line := range strings.Split(prose, "\n") {
		if !containsFold(line, field) {
			continue
		}
		m := enumMarkerRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		var vals []string
		for _, part := range strings.Split(m[1], ",") {
			tok := strings.Trim(strings.TrimSpace(part), "`\"'.")
			if enumTokenRe.MatchString(tok) {
				vals = append(vals, tok)
			}
		}
		if len(vals) >= 2 {
			out = append(out, vals)
		}
	}
	return out
}
