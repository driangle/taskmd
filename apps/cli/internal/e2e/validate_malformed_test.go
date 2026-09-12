//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRoot copies a fixture set from testdata/ into a fresh temp dir and
// returns it. Copying (rather than pointing the binary at testdata directly)
// keeps runs isolated from the repository's own .taskmd.yaml, which config
// discovery would otherwise find by walking up from the fixture path.
func fixtureRoot(t *testing.T, set string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS(filepath.Join("testdata", set))); err != nil {
		t.Fatalf("failed to copy fixture %s: %v", set, err)
	}
	return root
}

// testdata/malformed-frontmatter reproduces GitHub issue #19: a task file
// whose frontmatter has a YAML syntax error must fail validation instead of
// being silently skipped.
func TestValidate_MalformedTaskFrontmatterFails(t *testing.T) {
	root := fixtureRoot(t, "malformed-frontmatter")

	result := run(t, root, "validate")

	if result.ExitCode == 0 {
		t.Fatalf("expected validation to fail, got exit 0:\n%s", result.Stdout)
	}
	combined := result.Stdout + result.Stderr
	if !strings.Contains(combined, "003-lorem.md") {
		t.Errorf("expected the broken file to be named, got:\n%s", combined)
	}
	if !strings.Contains(combined, "YAML") {
		t.Errorf("expected a YAML parse error, got:\n%s", combined)
	}
}

// testdata/foreign-frontmatter pairs a good task with foreign markdown whose
// frontmatter is broken but carries no task-signature keys — validation must
// still pass.
func TestValidate_ForeignMarkdownWithBrokenFrontmatterIgnored(t *testing.T) {
	root := fixtureRoot(t, "foreign-frontmatter")

	result := mustRun(t, root, "validate")

	if !strings.Contains(result.Stdout, "valid") {
		t.Errorf("expected validation to pass, got:\n%s", result.Stdout)
	}
}

// list must warn about the broken file on stderr (without --verbose), so the
// task does not just silently vanish from views.
func TestList_WarnsAboutMalformedTaskFile(t *testing.T) {
	root := fixtureRoot(t, "malformed-frontmatter")

	result := mustRun(t, root, "list")

	if !strings.Contains(result.Stderr, "003-lorem.md") {
		t.Errorf("expected a stderr warning naming the broken file, got:\n%s", result.Stderr)
	}
}
