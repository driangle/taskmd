package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
)

// Per-project overlay wiring for --all-projects (spec §7). The merge rules
// themselves are covered in internal/worktree; these tests cover activation
// resolved per project, provenance on ProjectTask, and the next/list surfaces.

// stubSiblingsByScanDir installs a discovery stub that returns siblings keyed
// by the scanned directory, so one test can give different projects different
// worktrees (or none).
func stubSiblingsByScanDir(byScanDir map[string][]gitmeta.Worktree) {
	discoverSiblingWorktrees = func(scanDir string) ([]gitmeta.Worktree, error) {
		return byScanDir[filepath.Clean(scanDir)], nil
	}
}

// overlayProject creates a registered-project directory holding task 001 as
// pending, optionally writing a worktree_scope into its .taskmd.yaml.
func overlayProject(t *testing.T, scope string) string {
	t.Helper()
	dir := createProjectWithTasks(t, "tasks", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	if scope != "" {
		cfg := filepath.Join(dir, configFilename)
		if err := os.WriteFile(cfg, []byte("task-dir: tasks\nworktree_scope: "+scope+"\n"), 0o644); err != nil {
			t.Fatalf("write project config: %v", err)
		}
	}
	return dir
}

// claimingSibling returns a sibling worktree that has task 001 in-progress.
func claimingSibling(t *testing.T, name string) []gitmeta.Worktree {
	t.Helper()
	return []gitmeta.Worktree{newSiblingWorktree(t, name, "agent/"+name, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
	})}
}

// statusOf returns the status --all-projects reports for a task ID.
func statusOf(t *testing.T, ptasks []*ProjectTask, id string) string {
	t.Helper()
	for _, pt := range ptasks {
		if pt.Task.ID == id {
			return string(pt.Task.Status)
		}
	}
	t.Fatalf("task %q not found in all-projects scan", id)
	return ""
}

func TestScanAllProjects_PerProjectActivationMatrix(t *testing.T) {
	cases := []struct {
		name       string
		scope      string // written to the project's own .taskmd.yaml
		siblings   bool
		wantStatus string
	}{
		{name: "default scope with sibling claim merges", scope: "", siblings: true, wantStatus: "in-progress"},
		{name: "unified with sibling claim merges", scope: "unified", siblings: true, wantStatus: "in-progress"},
		{name: "isolated with sibling claim stays local", scope: "isolated", siblings: true, wantStatus: "pending"},
		{name: "default scope without siblings stays local", scope: "", siblings: false, wantStatus: "pending"},
		{name: "unified without siblings stays local", scope: "unified", siblings: false, wantStatus: "pending"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetCLIState()
			defer resetCLIState()

			proj := overlayProject(t, tc.scope)
			if tc.siblings {
				stubSiblingsByScanDir(map[string][]gitmeta.Worktree{
					filepath.Join(proj, "tasks"): claimingSibling(t, "agent-b"),
				})
			}
			setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+proj+"\n")

			ptasks, err := scanAllProjects()
			if err != nil {
				t.Fatalf("scanAllProjects: %v", err)
			}
			if got := statusOf(t, ptasks, "001"); got != tc.wantStatus {
				t.Errorf("status = %q, want %q", got, tc.wantStatus)
			}
		})
	}
}

func TestScanAllProjects_ScopeIsPerProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	// Two repos, each with a sibling claiming 001; only beta opted out.
	unified := overlayProject(t, "unified")
	isolated := overlayProject(t, "isolated")
	stubSiblingsByScanDir(map[string][]gitmeta.Worktree{
		filepath.Join(unified, "tasks"):  claimingSibling(t, "agent-b"),
		filepath.Join(isolated, "tasks"): claimingSibling(t, "agent-c"),
	})
	setupProjectFlagRegistry(t,
		"  - id: alpha\n    path: "+unified+"\n"+
			"  - id: beta\n    path: "+isolated+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("scanAllProjects: %v", err)
	}

	byProject := map[string]*ProjectTask{}
	for _, pt := range ptasks {
		byProject[pt.ProjectID] = pt
	}
	if got := string(byProject["alpha"].Task.Status); got != "in-progress" {
		t.Errorf("unified project status = %q, want in-progress", got)
	}
	if got := string(byProject["beta"].Task.Status); got != "pending" {
		t.Errorf("isolated project must stay local, got %q", got)
	}
	if byProject["beta"].Worktree != "" {
		t.Errorf("isolated project must carry no provenance, got %q", byProject["beta"].Worktree)
	}
}

func TestScanAllProjects_InvocationScopeOverridesEveryProject(t *testing.T) {
	for _, tc := range []struct {
		name  string
		apply func(t *testing.T)
	}{
		{name: "env", apply: func(t *testing.T) { t.Setenv("TASKMD_WORKTREE_SCOPE", "isolated") }},
		{name: "flag", apply: func(t *testing.T) {
			f := rootCmd.PersistentFlags().Lookup("worktree-scope")
			if err := f.Value.Set(worktreeScopeIsolated); err != nil {
				t.Fatalf("set flag: %v", err)
			}
			f.Changed = true
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetCLIState()
			defer resetCLIState()

			// The project's own config says unified; the invocation says no.
			proj := overlayProject(t, "unified")
			stubSiblingsByScanDir(map[string][]gitmeta.Worktree{
				filepath.Join(proj, "tasks"): claimingSibling(t, "agent-b"),
			})
			setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+proj+"\n")
			tc.apply(t)

			ptasks, err := scanAllProjects()
			if err != nil {
				t.Fatalf("scanAllProjects: %v", err)
			}
			if got := statusOf(t, ptasks, "001"); got != "pending" {
				t.Errorf("invocation-wide isolated scope ignored: status = %q, want pending", got)
			}
		})
	}
}

func TestScanAllProjects_ProvenanceAndSiblingOnlyTasks(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := overlayProject(t, "")
	sibling := newSiblingWorktree(t, "agent-b", "dnc/001/parser", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
		"002-beta.md":  overlayTaskMD("002", "Beta", "pending"),
	})
	stubSiblingsByScanDir(map[string][]gitmeta.Worktree{
		filepath.Join(proj, "tasks"): {sibling},
	})
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+proj+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("scanAllProjects: %v", err)
	}

	byID := map[string]*ProjectTask{}
	for _, pt := range ptasks {
		byID[pt.Task.ID] = pt
	}

	claimed, ok := byID["001"]
	if !ok {
		t.Fatal("task 001 missing")
	}
	if claimed.Worktree != "agent-b" || claimed.Branch != "dnc/001/parser" {
		t.Errorf("provenance = %q/%q, want agent-b/dnc/001/parser", claimed.Worktree, claimed.Branch)
	}
	if claimed.RemoteOnly {
		t.Error("001 exists locally, must not be marked remote-only")
	}

	// A task created only on the sibling's branch is visible and marked.
	siblingOnly, ok := byID["002"]
	if !ok {
		t.Fatal("sibling-only task 002 missing from --all-projects")
	}
	if !siblingOnly.RemoteOnly {
		t.Error("002 should be marked remote-only")
	}
	if siblingOnly.ProjectID != "alpha" {
		t.Errorf("sibling-only task project = %q, want alpha", siblingOnly.ProjectID)
	}
}

func TestScanAllProjects_InactiveOverlayCarriesNoProvenance(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := overlayProject(t, "")
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+proj+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("scanAllProjects: %v", err)
	}

	// Serialized output must be byte-identical to pre-overlay behavior: the
	// provenance fields are omitempty and every one of them is empty here.
	data, err := json.Marshal(ptasks)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, field := range []string{"worktree", "branch", "remote_only", "effective_status"} {
		if strings.Contains(string(data), `"`+field+`"`) {
			t.Errorf("inactive overlay leaked %q into output: %s", field, data)
		}
	}
}

func TestScanAllProjects_InvalidProjectScopeSkipsProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	good := overlayProject(t, "unified")
	bad := overlayProject(t, "sometimes")
	setupProjectFlagRegistry(t,
		"  - id: good\n    path: "+good+"\n"+
			"  - id: bad\n    path: "+bad+"\n")

	ptasks, err := scanAllProjects()
	if err != nil {
		t.Fatalf("one bad project must not fail the whole run: %v", err)
	}
	for _, pt := range ptasks {
		if pt.ProjectID == "bad" {
			t.Error("project with invalid worktree_scope should be skipped")
		}
	}
	if len(ptasks) != 1 {
		t.Errorf("expected the good project's single task, got %d", len(ptasks))
	}
}

func TestListAllProjects_WorktreeColumnAndEffectiveFilter(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := overlayProject(t, "")
	siblings := claimingSibling(t, "agent-b")
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+proj+"\n")

	repo := newTaskRepo(t, nil)
	res := repo.RunWith(stubSiblingsByScanDirHook(map[string][]gitmeta.Worktree{
		filepath.Join(proj, "tasks"): siblings,
	}), "list", "--all-projects")
	if res.Err != nil {
		t.Fatalf("list --all-projects: %v", res.Err)
	}
	if !strings.Contains(strings.ToLower(res.Stdout), "worktree") {
		t.Errorf("expected WORKTREE column in output:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "agent-b") {
		t.Errorf("expected worktree provenance in output:\n%s", res.Stdout)
	}
}

func TestNextAllProjects_ExcludesSiblingClaims(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	proj := createProjectWithTasks(t, "tasks", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
		"002-beta.md":  overlayTaskMD("002", "Beta", "pending"),
	})
	sibling := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
		"002-beta.md":  overlayTaskMD("002", "Beta", "pending"),
	})}
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+proj+"\n")

	repo := newTaskRepo(t, nil)
	res := repo.RunWith(stubSiblingsByScanDirHook(map[string][]gitmeta.Worktree{
		filepath.Join(proj, "tasks"): sibling,
	}), "next", "--all-projects", "--format", "json")
	if res.Err != nil {
		t.Fatalf("next --all-projects: %v", res.Err)
	}

	var recs []struct {
		Project string `json:"project"`
		ID      string `json:"id"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("parse next output: %v\n%s", err, res.Stdout)
	}

	ids := make([]string, 0, len(recs))
	for _, rec := range recs {
		ids = append(ids, rec.ID)
		if rec.ID == "001" {
			t.Errorf("001 is claimed in a sibling worktree and must not be recommended:\n%s", res.Stdout)
		}
	}
	if len(ids) != 1 || ids[0] != "002" {
		t.Errorf("expected exactly 002 to be recommended, got %v:\n%s", ids, res.Stdout)
	}
}

// stubSiblingsByScanDirHook adapts stubSiblingsByScanDir to RunWith's configure
// hook, which runs after resetCLIState clears the discovery stub.
func stubSiblingsByScanDirHook(byScanDir map[string][]gitmeta.Worktree) func() {
	return func() { stubSiblingsByScanDir(byScanDir) }
}
