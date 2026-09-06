package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// withIgnore returns a RunWith configure hook that seeds the `ignore` config key.
func withIgnore(dirs ...string) func() {
	return func() { viper.Set("ignore", dirs) }
}

// assertNoTaskFiles fails if any .md file exists under the repo-relative dir.
func assertNoTaskFiles(t *testing.T, repo *taskRepo, dir string) {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(repo.Path(dir), "*.md"))
	if len(matches) != 0 {
		t.Errorf("expected no task files in %s, found %v", dir, matches)
	}
}

func TestAdd_GroupIgnored_Refused(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.RunWith(withIgnore("content"), "add", "Probe task", "--group", "content")

	if res.Err == nil {
		t.Fatal("expected add to refuse an ignored group, got nil error")
	}
	msg := res.Err.Error()
	for _, want := range []string{`--group "content"`, "ignore", ".taskmd.yaml", "list, get or validate"} {
		if !strings.Contains(msg, want) {
			t.Errorf("expected error to mention %q, got: %s", want, msg)
		}
	}
	if _, err := os.Stat(repo.Path("content")); !os.IsNotExist(err) {
		t.Error("expected no directory created for the refused group")
	}
}

func TestAdd_GroupIgnoredNested_Refused(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// `ignore` entries match a bare directory name at any depth, so a nested
	// group whose basename is ignored is just as unreachable as a top-level one.
	res := repo.RunWith(withIgnore("content"), "add", "Probe task", "--group", filepath.Join("a", "content"))

	if res.Err == nil {
		t.Fatal("expected add to refuse a nested group whose basename is ignored")
	}
	if !strings.Contains(res.Err.Error(), `"content"`) {
		t.Errorf("expected error to name the offending segment, got: %s", res.Err.Error())
	}
	assertNoTaskFiles(t, repo, filepath.Join("a", "content"))
}

func TestAdd_GroupNotIgnored_Unaffected(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.RunWith(withIgnore("content"), "add", "CLI task", "--group", "cli")

	if res.Err != nil {
		t.Fatalf("expected add to succeed for a non-ignored group: %v", res.Err)
	}
	files, _ := filepath.Glob(filepath.Join(repo.Path("cli"), "*.md"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file in cli/, got %d", len(files))
	}
}

func TestAdd_GroupDefaultSkipDir_Refused(t *testing.T) {
	repo := newTaskRepo(t, nil)

	// node_modules is skipped by the scanner regardless of config, so the same
	// silent-invisibility failure applies — but 'ignore' is not the culprit.
	res := repo.RunWith(nil, "add", "Probe task", "--group", "node_modules")

	if res.Err == nil {
		t.Fatal("expected add to refuse a group the scanner always skips")
	}
	msg := res.Err.Error()
	if !strings.Contains(msg, "always skipped by the scanner") {
		t.Errorf("expected always-skipped wording, got: %s", msg)
	}
	if strings.Contains(msg, "Remove it from 'ignore'") {
		t.Errorf("expected no 'ignore' remedy for a default skip dir, got: %s", msg)
	}
}

func TestAdd_GroupHidden_Refused(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.RunWith(nil, "add", "Probe task", "--group", ".drafts")

	if res.Err == nil {
		t.Fatal("expected add to refuse a hidden group")
	}
	if !strings.Contains(res.Err.Error(), "hidden") {
		t.Errorf("expected hidden-directory wording, got: %s", res.Err.Error())
	}
}
