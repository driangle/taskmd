//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// customEffortConfig is the vocabulary used across these tests: five values,
// none of which overlap the built-in small/medium/large.
const customEffortConfig = "effort: [xs, s, m, l, xl]\n"

// writeEffortTask writes a task file carrying an effort value.
func writeEffortTask(t *testing.T, dir, filename, id, title, effortValue string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}

	content := fmt.Sprintf(`---
id: %q
title: %q
status: pending
priority: medium
effort: %s
---

# %s
`, id, title, effortValue, title)

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write task file %s: %v", path, err)
	}
}

// newEffortProject creates a project configured with the custom vocabulary and
// three tasks spanning its range.
func newEffortProject(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	writeConfig(t, root, customEffortConfig)
	writeEffortTask(t, root, "001-tiny.md", "001", "Tiny Task", "xs")
	writeEffortTask(t, root, "002-mid.md", "002", "Mid Task", "m")
	writeEffortTask(t, root, "003-huge.md", "003", "Huge Task", "xl")
	return root
}

func TestEffort_CustomVocabularyValidates(t *testing.T) {
	root := newEffortProject(t)

	result := mustRun(t, root, "validate")

	if !strings.Contains(result.Stdout, "valid") {
		t.Errorf("expected validation to pass, got:\n%s", result.Stdout)
	}
}

// A file using the built-in vocabulary must be rejected once the project has
// configured its own, and the error must name the configured values.
func TestEffort_DefaultValuesRejectedUnderCustomVocabulary(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, customEffortConfig)
	writeEffortTask(t, root, "001-stale.md", "001", "Stale Task", "medium")

	result := run(t, root, "validate")

	if result.ExitCode == 0 {
		t.Fatalf("expected validation to fail, got exit 0:\n%s", result.Stdout)
	}
	combined := result.Stdout + result.Stderr
	if !strings.Contains(combined, "invalid effort") {
		t.Errorf("expected an invalid effort error, got:\n%s", combined)
	}
	if !strings.Contains(combined, "xs, s, m, l, xl") {
		t.Errorf("expected the configured values in the message, got:\n%s", combined)
	}
}

func TestEffort_FilterAcceptsCustomValues(t *testing.T) {
	root := newEffortProject(t)

	tests := []struct {
		name      string
		filter    string
		wantTitle string
		absent    string
	}{
		{"equality", "effort=xs", "Tiny Task", "Huge Task"},
		{"greater than", "effort>m", "Huge Task", "Tiny Task"},
		{"less or equal", "effort<=m", "Mid Task", "Huge Task"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mustRun(t, root, "list", "--filter", tt.filter)

			if !strings.Contains(result.Stdout, tt.wantTitle) {
				t.Errorf("filter %q: expected %q in output:\n%s", tt.filter, tt.wantTitle, result.Stdout)
			}
			if strings.Contains(result.Stdout, tt.absent) {
				t.Errorf("filter %q: did not expect %q in output:\n%s", tt.filter, tt.absent, result.Stdout)
			}
		})
	}
}

func TestEffort_AddAcceptsCustomValue(t *testing.T) {
	root := newEffortProject(t)

	mustRun(t, root, "add", "Brand New", "--effort", "l")

	result := mustRun(t, root, "list", "--filter", "effort=l")
	if !strings.Contains(result.Stdout, "Brand New") {
		t.Errorf("expected the added task to carry effort l, got:\n%s", result.Stdout)
	}
}

func TestEffort_AddRejectsValueOutsideVocabulary(t *testing.T) {
	root := newEffortProject(t)

	result := run(t, root, "add", "Nope", "--effort", "medium")

	if result.ExitCode == 0 {
		t.Fatal("expected add to fail for a value outside the vocabulary")
	}
	combined := result.Stdout + result.Stderr
	if !strings.Contains(combined, "xs, s, m, l, xl") {
		t.Errorf("expected the configured values in the error, got:\n%s", combined)
	}
}

func TestEffort_SetAcceptsCustomValue(t *testing.T) {
	root := newEffortProject(t)

	mustRun(t, root, "set", "003", "--effort", "s")

	result := mustRun(t, root, "list", "--filter", "effort=s")
	if !strings.Contains(result.Stdout, "Huge Task") {
		t.Errorf("expected task 003 to have effort s, got:\n%s", result.Stdout)
	}
}

// --quick-wins must select the lowest configured value, not the literal "small".
func TestEffort_QuickWinsUsesLowestConfiguredValue(t *testing.T) {
	root := newEffortProject(t)

	result := mustRun(t, root, "next", "--quick-wins")

	if !strings.Contains(result.Stdout, "Tiny Task") {
		t.Errorf("expected the xs task as a quick win, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Huge Task") {
		t.Errorf("did not expect the xl task among quick wins, got:\n%s", result.Stdout)
	}
}

// Help text is rendered before cobra's initializers run, so this guards the
// help-func hook that loads config first.
func TestEffort_HelpShowsConfiguredValues(t *testing.T) {
	root := newEffortProject(t)

	for _, cmd := range []string{"add", "set"} {
		t.Run(cmd, func(t *testing.T) {
			result := mustRun(t, root, cmd, "--help")

			if !strings.Contains(result.Stdout, "xs, s, m, l, xl") {
				t.Errorf("expected %s --help to list the configured values, got:\n%s", cmd, result.Stdout)
			}
		})
	}
}

func TestEffort_BoardOrdersColumnsByVocabulary(t *testing.T) {
	root := newEffortProject(t)

	result := mustRun(t, root, "board", "--group-by", "effort")

	xs := strings.Index(result.Stdout, "xs")
	m := strings.Index(result.Stdout, "## m")
	xl := strings.Index(result.Stdout, "xl")
	if xs < 0 || m < 0 || xl < 0 {
		t.Fatalf("expected xs, m and xl groups, got:\n%s", result.Stdout)
	}
	if !(xs < m && m < xl) {
		t.Errorf("expected groups ordered xs, m, xl, got:\n%s", result.Stdout)
	}
}

// --- Invalid configuration ---

func TestEffort_InvalidConfigIsReported(t *testing.T) {
	tests := []struct {
		name       string
		config     string
		wantErrSub string
	}{
		{"empty list", "effort: []\n", "at least one value"},
		{"duplicates", "effort: [a, a]\n", "duplicate effort value"},
		{"blank entry", "effort: [a, \"\"]\n", "is empty"},
		{"mapping instead of list", "effort:\n  values: [a, b]\n", "must be a list of effort values"},
		{"non-string item", "effort: [a, 3]\n", "must be a string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeConfig(t, root, tt.config)
			writeEffortTask(t, root, "001-task.md", "001", "A Task", "small")

			result := run(t, root, "validate")

			if result.ExitCode == 0 {
				t.Fatalf("expected validate to fail for %s, got exit 0:\n%s", tt.name, result.Stdout)
			}
			combined := result.Stdout + result.Stderr
			if !strings.Contains(combined, tt.wantErrSub) {
				t.Errorf("expected %q in output, got:\n%s", tt.wantErrSub, combined)
			}
		})
	}
}

// A project with no effort config must behave exactly as before this feature.
func TestEffort_NoConfigKeepsDefaultVocabulary(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "dir: .\n")
	writeEffortTask(t, root, "001-small.md", "001", "Small Task", "small")
	writeEffortTask(t, root, "002-large.md", "002", "Large Task", "large")

	mustRun(t, root, "validate")

	result := mustRun(t, root, "next", "--quick-wins")
	if !strings.Contains(result.Stdout, "Small Task") {
		t.Errorf("expected the small task as a quick win, got:\n%s", result.Stdout)
	}

	bad := run(t, root, "add", "Nope", "--effort", "xs")
	if bad.ExitCode == 0 {
		t.Error("expected a value outside the default vocabulary to be rejected")
	}
}
