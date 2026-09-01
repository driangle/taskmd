package worktree

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// taskMD renders a minimal task file for overlay tests.
func taskMD(id, title, status string, extraFrontmatter ...string) string {
	extra := ""
	if len(extraFrontmatter) > 0 {
		extra = strings.Join(extraFrontmatter, "\n") + "\n"
	}
	return fmt.Sprintf(`---
id: %q
title: %q
status: %s
priority: medium
created_at: 2026-01-01
%s---

# %s
`, id, title, status, extra, title)
}

// writeTaskDir lays out a directory of task files and returns its path.
func writeTaskDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for fname, content := range files {
		if err := os.WriteFile(filepath.Join(dir, fname), []byte(content), 0o644); err != nil {
			t.Fatalf("write task %s: %v", fname, err)
		}
	}
	return dir
}

// newSiblingWorktree lays out a fake sibling worktree (root + tasks dir with
// the given files) and returns its descriptor.
func newSiblingWorktree(t *testing.T, name, branch string, files map[string]string) gitmeta.Worktree {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	tasksDir := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir sibling tasks dir: %v", err)
	}
	for fname, content := range files {
		if err := os.WriteFile(filepath.Join(tasksDir, fname), []byte(content), 0o644); err != nil {
			t.Fatalf("write sibling task %s: %v", fname, err)
		}
	}
	return gitmeta.Worktree{Root: root, Branch: branch, TasksDir: tasksDir}
}

// scanDirTasks scans a directory into model tasks for direct merge tests.
func scanDirTasks(t *testing.T, dir string) []*model.Task {
	t.Helper()
	result, err := scanner.NewScanner(dir, false, nil).Scan()
	if err != nil {
		t.Fatalf("scan %s: %v", dir, err)
	}
	return result.Tasks
}

func TestStatusLadderRank_Ordering(t *testing.T) {
	t.Parallel()
	ladder := []model.Status{
		model.StatusPending, model.StatusBlocked, model.StatusInProgress,
		model.StatusInReview, model.StatusCancelled, model.StatusCompleted,
	}
	for i := 1; i < len(ladder); i++ {
		if statusLadderRank(ladder[i-1]) >= statusLadderRank(ladder[i]) {
			t.Errorf("statusLadderRank(%s) should be below %s", ladder[i-1], ladder[i])
		}
	}
	if statusLadderRank(model.Status("bogus")) != statusLadderRank(model.StatusPending) {
		t.Error("unknown status should rank alongside pending")
	}
}

func TestMerge_RemoteWinnerProvenance(t *testing.T) {
	localDir := writeTaskDir(t, map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "pending"),
	})
	sibling := newSiblingWorktree(t, "agent-b", "dnc/001/parser", map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "in-progress", `owner: "agent-b"`),
	})

	local := scanDirTasks(t, localDir)
	overlay := Merge(local, []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})

	ot := overlay.Get("001")
	if ot == nil {
		t.Fatal("task 001 missing from overlay")
		return
	}
	if ot.EffectiveStatus != model.StatusInProgress {
		t.Errorf("EffectiveStatus = %s, want in-progress", ot.EffectiveStatus)
	}
	if ot.EffectiveOwner != "agent-b" {
		t.Errorf("EffectiveOwner = %q, want agent-b", ot.EffectiveOwner)
	}
	if ot.Worktree != "agent-b" || ot.Branch != "dnc/001/parser" {
		t.Errorf("provenance = (%q, %q), want (agent-b, dnc/001/parser)", ot.Worktree, ot.Branch)
	}
	if ot.Status != model.StatusPending {
		t.Errorf("base copy status = %s, want the local pending", ot.Status)
	}
	if ot.LocalOnly || ot.RemoteOnly {
		t.Errorf("LocalOnly/RemoteOnly = %v/%v, want false/false", ot.LocalOnly, ot.RemoteOnly)
	}
}

func TestMerge_LocalWinnerHasNoProvenance(t *testing.T) {
	localDir := writeTaskDir(t, map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "completed"),
		"002-solo.md":  taskMD("002", "Solo", "pending"),
	})
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "pending"),
	})

	local := scanDirTasks(t, localDir)
	overlay := Merge(local, []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})

	if ot := overlay.Get("001"); ot.EffectiveStatus != model.StatusCompleted || ot.Worktree != "" {
		t.Errorf("001 = (%s, worktree %q), want local completed winner with no provenance",
			ot.EffectiveStatus, ot.Worktree)
	}
	if ot := overlay.Get("002"); !ot.LocalOnly {
		t.Error("002 has no sibling copy and should be LocalOnly")
	}
	if len(overlay.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", overlay.Warnings)
	}
}

func TestMerge_MtimeBreaksStatusTies(t *testing.T) {
	localDir := writeTaskDir(t, map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "in-progress", `owner: "local"`),
	})
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "in-progress", `owner: "remote"`),
	})

	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now().Add(-1 * time.Minute)
	localPath := filepath.Join(localDir, "001-alpha.md")
	siblingPath := filepath.Join(sibling.TasksDir, "001-alpha.md")

	// Sibling copy newer → remote wins the tie.
	if err := os.Chtimes(localPath, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(siblingPath, recent, recent); err != nil {
		t.Fatal(err)
	}
	overlay := Merge(scanDirTasks(t, localDir), []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})
	if ot := overlay.Get("001"); ot.Worktree != "agent-b" || ot.EffectiveOwner != "remote" {
		t.Errorf("newer sibling copy should win tie, got worktree %q owner %q", ot.Worktree, ot.EffectiveOwner)
	}

	// Local copy newer → local wins the tie.
	if err := os.Chtimes(localPath, recent, recent); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(siblingPath, old, old); err != nil {
		t.Fatal(err)
	}
	overlay = Merge(scanDirTasks(t, localDir), []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})
	if ot := overlay.Get("001"); ot.Worktree != "" || ot.EffectiveOwner != "local" {
		t.Errorf("newer local copy should win tie, got worktree %q owner %q", ot.Worktree, ot.EffectiveOwner)
	}
}

func TestMerge_RemoteOnlyAppendedSorted(t *testing.T) {
	localDir := writeTaskDir(t, map[string]string{
		"002-local.md": taskMD("002", "Local", "pending"),
	})
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"003-remote.md": taskMD("003", "Remote C", "pending"),
		"001-remote.md": taskMD("001", "Remote A", "in-progress"),
	})

	overlay := Merge(scanDirTasks(t, localDir), []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})

	var order []string
	for _, ot := range overlay.Tasks {
		order = append(order, ot.Task.ID)
	}
	// Local scan order first, then remote-only sorted by ID.
	if len(order) != 3 || order[0] != "002" || order[1] != "001" || order[2] != "003" {
		t.Fatalf("overlay order = %v, want [002 001 003]", order)
	}
	for _, id := range []string{"001", "003"} {
		ot := overlay.Get(id)
		if !ot.RemoteOnly || ot.Worktree != "agent-b" {
			t.Errorf("%s: RemoteOnly=%v worktree=%q, want remote-only with provenance", id, ot.RemoteOnly, ot.Worktree)
		}
	}
}

func TestMerge_DivergentTerminalStatesWarn(t *testing.T) {
	localDir := writeTaskDir(t, map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "cancelled"),
	})
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "completed"),
	})

	overlay := Merge(scanDirTasks(t, localDir), []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})

	if ot := overlay.Get("001"); ot.EffectiveStatus != model.StatusCompleted {
		t.Errorf("EffectiveStatus = %s, want completed (ladder winner)", ot.EffectiveStatus)
	}
	if len(overlay.Warnings) != 1 || !strings.Contains(overlay.Warnings[0].Message, "completed in worktree agent-b but cancelled in this worktree") {
		t.Errorf("Warnings = %v, want one divergent-terminal warning", overlay.Warnings)
	}
	if overlay.Warnings[0].TaskID != "001" {
		t.Errorf("warning TaskID = %q, want 001", overlay.Warnings[0].TaskID)
	}
}

func TestBuilder_Build_ActivationMatrix(t *testing.T) {
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": taskMD("001", "Alpha", "in-progress"),
	})}
	none := func(string) ([]gitmeta.Worktree, error) { return nil, nil }
	some := func(string) ([]gitmeta.Worktree, error) { return siblings, nil }
	failing := func(string) ([]gitmeta.Worktree, error) { return nil, fmt.Errorf("boom") }

	cases := []struct {
		name       string
		enabled    bool
		discover   Discoverer
		wantActive bool
	}{
		{name: "enabled with siblings", enabled: true, discover: some, wantActive: true},
		{name: "enabled without siblings", enabled: true, discover: none},
		{name: "disabled with siblings", discover: some},
		{name: "zero value is disabled", discover: some},
		{name: "discovery failure deactivates", enabled: true, discover: failing, wantActive: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			localDir := writeTaskDir(t, map[string]string{"001-alpha.md": taskMD("001", "Alpha", "pending")})
			b := Builder{Enabled: tc.enabled, Discover: tc.discover}
			overlay, err := b.Build(localDir, scanDirTasks(t, localDir))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (overlay != nil) != tc.wantActive {
				t.Errorf("overlay active = %v, want %v", overlay != nil, tc.wantActive)
			}
		})
	}
}

func TestOverlay_SiblingGuard(t *testing.T) {
	localDir := writeTaskDir(t, map[string]string{
		"001-local.md": taskMD("001", "Local", "pending"),
	})
	sibling := newSiblingWorktree(t, "agent-b", "dnc/099/parser", map[string]string{
		"001-local.md":   taskMD("001", "Local", "in-progress"),
		"099-sibling.md": taskMD("099", "Sibling only", "pending"),
	})

	overlay := Merge(scanDirTasks(t, localDir), []SiblingTasks{{WT: sibling, Tasks: scanDirTasks(t, sibling.TasksDir)}})

	if err := overlay.SiblingGuard("001"); err != nil {
		t.Errorf("guard on a task with a local copy should pass, got %v", err)
	}
	if err := overlay.SiblingGuard("404"); err != nil {
		t.Errorf("guard on an unknown id should pass, got %v", err)
	}
	err := overlay.SiblingGuard("099")
	if err == nil {
		t.Fatal("guard on a sibling-only id should fail")
	}
	msg := err.Error()
	if !strings.Contains(msg, "exists only in worktree") ||
		!strings.Contains(msg, "agent-b") ||
		!strings.Contains(msg, "dnc/099/parser") ||
		!strings.Contains(msg, "run taskmd there") {
		t.Errorf("guard message = %q, want worktree + branch + guidance", msg)
	}
}

func TestBuilder_SiblingGuard_ScansSiblings(t *testing.T) {
	sibling := newSiblingWorktree(t, "agent-b", "", map[string]string{
		"099-sibling.md": taskMD("099", "Sibling only", "pending"),
	})
	b := Builder{Enabled: true, Discover: func(string) ([]gitmeta.Worktree, error) {
		return []gitmeta.Worktree{sibling}, nil
	}}

	err := b.SiblingGuard("099", t.TempDir())
	if err == nil {
		t.Fatal("guard should find the sibling copy")
	}
	if !strings.Contains(err.Error(), "branch detached") {
		t.Errorf("empty branch should render as detached, got %q", err.Error())
	}
	if guardErr := b.SiblingGuard("404", t.TempDir()); guardErr != nil {
		t.Errorf("guard on unknown id should pass, got %v", guardErr)
	}
}

// TestBuilder_ScanSiblings_PreservesWorktreeOrder pins the invariant that makes
// concurrent sibling scanning safe: scans run in parallel and finish in
// arbitrary order, but results must still come back in worktree order. Each
// sibling carries a distinct task count so a reordered result is detectable.
func TestBuilder_ScanSiblings_PreservesWorktreeOrder(t *testing.T) {
	const siblingCount = 12
	siblings := make([]gitmeta.Worktree, 0, siblingCount)
	for i := range siblingCount {
		files := map[string]string{}
		for j := 0; j <= i; j++ {
			id := fmt.Sprintf("wt%02d-task%02d", i, j)
			files[id+".md"] = taskMD(id, "Task", "pending")
		}
		siblings = append(siblings, newSiblingWorktree(t, fmt.Sprintf("agent-%02d", i), "", files))
	}

	b := Builder{Enabled: true}
	for range 5 { // repeat: a scheduling-order bug would be intermittent
		scanned := b.scanSiblings(siblings)
		if len(scanned) != siblingCount {
			t.Fatalf("scanned %d siblings, want %d", len(scanned), siblingCount)
		}
		for i, s := range scanned {
			if s.WT.Root != siblings[i].Root {
				t.Fatalf("position %d = %s, want %s", i, s.WT.Root, siblings[i].Root)
			}
			if len(s.Tasks) != i+1 {
				t.Errorf("sibling %d has %d tasks, want %d", i, len(s.Tasks), i+1)
			}
		}
	}
}

// TestBuilder_ScanSiblings_MissingTasksDirIsEmptyNotDropped pins how a sibling
// with no tasks dir behaves. A missing root is not a scan failure: WalkDir
// surfaces it through the walk callback, which the scanner records as a scan
// error and still returns success, so the sibling stays in the list carrying
// zero tasks. It contributes nothing to the merge either way; what matters
// here is that it does not shift its neighbours' positions.
func TestBuilder_ScanSiblings_MissingTasksDirIsEmptyNotDropped(t *testing.T) {
	good := newSiblingWorktree(t, "agent-good", "", map[string]string{
		"001-a.md": taskMD("001", "A", "pending"),
	})
	missing := gitmeta.Worktree{Root: "/nonexistent-root", TasksDir: "/nonexistent-root/tasks"}
	after := newSiblingWorktree(t, "agent-after", "", map[string]string{
		"002-b.md": taskMD("002", "B", "pending"),
	})

	scanned := Builder{Enabled: true}.scanSiblings([]gitmeta.Worktree{good, missing, after})
	if len(scanned) != 3 {
		t.Fatalf("scanned %d siblings, want 3", len(scanned))
	}
	for i, want := range []string{good.Root, missing.Root, after.Root} {
		if scanned[i].WT.Root != want {
			t.Errorf("position %d = %s, want %s", i, scanned[i].WT.Root, want)
		}
	}
	if len(scanned[1].Tasks) != 0 {
		t.Errorf("missing tasks dir yielded %d tasks, want 0", len(scanned[1].Tasks))
	}
}

// captureStderr redirects os.Stderr for the duration of fn and returns what was
// written. The scanner logs verbose progress straight to os.Stderr, so this is
// the only way to observe the ordering guarantee the serial path exists to
// provide. It swaps a package-level global, so callers must not run in
// parallel with other stderr writers.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w

	// Drain concurrently: the scanner can emit more than a pipe buffer holds,
	// and a blocked write would deadlock the scan.
	done := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	os.Stderr = orig
	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("close stderr pipe: %v", closeErr)
	}
	out := <-done
	if closeErr := r.Close(); closeErr != nil {
		t.Fatalf("close read pipe: %v", closeErr)
	}
	return out
}

// orderedSiblings builds n siblings, each holding one uniquely-named task.
func orderedSiblings(t *testing.T, n int) []gitmeta.Worktree {
	t.Helper()
	siblings := make([]gitmeta.Worktree, 0, n)
	for i := range n {
		id := fmt.Sprintf("wt%02d", i)
		siblings = append(siblings, newSiblingWorktree(t, "agent-"+id, "branch-"+id,
			map[string]string{id + ".md": taskMD(id, "Task "+id, "pending")}))
	}
	return siblings
}

// TestBuilder_ScanSiblings_VerboseLogsInWorktreeOrder covers the serial branch
// of scanSiblings. Verbose scanning is deliberately not concurrent so its
// stderr progress lines stay readable, so each sibling's "Scanning directory"
// line must appear in worktree order and be stable run to run.
func TestBuilder_ScanSiblings_VerboseLogsInWorktreeOrder(t *testing.T) {
	siblings := orderedSiblings(t, 8)
	b := Builder{Enabled: true, Verbose: true}

	first := captureStderr(t, func() { b.scanSiblings(siblings) })
	second := captureStderr(t, func() { b.scanSiblings(siblings) })

	if first != second {
		t.Error("verbose output differs between runs; serial scanning should be deterministic")
	}

	prev := -1
	for i, wt := range siblings {
		at := strings.Index(first, "Scanning directory: "+wt.TasksDir+"\n")
		if at < 0 {
			t.Fatalf("no scan line for sibling %d (%s)", i, wt.TasksDir)
		}
		if at < prev {
			t.Errorf("sibling %d logged out of worktree order", i)
		}
		prev = at
	}
}

// TestBuilder_ScanSiblings_VerboseAndConcurrentAgree guards the refactor that
// split scanSiblings into two paths: the serial verbose path and the concurrent
// quiet path must return the same worktrees, in the same order, with the same
// tasks. Only the logging differs.
func TestBuilder_ScanSiblings_VerboseAndConcurrentAgree(t *testing.T) {
	siblings := orderedSiblings(t, 10)

	var verbose []SiblingTasks
	captureStderr(t, func() {
		verbose = Builder{Enabled: true, Verbose: true}.scanSiblings(siblings)
	})
	concurrent := Builder{Enabled: true}.scanSiblings(siblings)

	if len(verbose) != len(concurrent) {
		t.Fatalf("verbose scanned %d siblings, concurrent %d", len(verbose), len(concurrent))
	}
	for i := range verbose {
		if verbose[i].WT.Root != concurrent[i].WT.Root {
			t.Errorf("position %d: verbose %s, concurrent %s",
				i, verbose[i].WT.Root, concurrent[i].WT.Root)
		}
		if len(verbose[i].Tasks) != len(concurrent[i].Tasks) {
			t.Fatalf("position %d: verbose has %d tasks, concurrent %d",
				i, len(verbose[i].Tasks), len(concurrent[i].Tasks))
		}
		for j := range verbose[i].Tasks {
			if verbose[i].Tasks[j].ID != concurrent[i].Tasks[j].ID {
				t.Errorf("position %d task %d: verbose %s, concurrent %s",
					i, j, verbose[i].Tasks[j].ID, concurrent[i].Tasks[j].ID)
			}
		}
	}
}

// TestBuilder_Overlay_DeterministicAcrossRuns checks that concurrency does not
// leak into the merged view: with many siblings contributing both shared and
// sibling-only tasks, the overlay's task order and provenance must be identical
// on every run.
func TestBuilder_Overlay_DeterministicAcrossRuns(t *testing.T) {
	const siblingCount = 12
	siblings := make([]gitmeta.Worktree, 0, siblingCount)
	for i := range siblingCount {
		id := fmt.Sprintf("wt%02d", i)
		siblings = append(siblings, newSiblingWorktree(t, "agent-"+id, "branch-"+id, map[string]string{
			// A shared task every sibling advances, plus one only it has.
			"shared.md": taskMD("shared", "Shared", "in-progress"),
			id + ".md":  taskMD(id, "Only "+id, "pending"),
		}))
	}
	local := scanDirTasks(t, writeTaskDir(t, map[string]string{
		"shared.md": taskMD("shared", "Shared", "pending"),
		"local.md":  taskMD("local", "Local only", "pending"),
	}))

	b := Builder{Enabled: true}
	fingerprint := func() string {
		overlay := b.Overlay(siblings, local)
		if overlay == nil {
			t.Fatal("overlay should be active with siblings present")
		}
		var sb strings.Builder
		for _, ot := range overlay.Tasks {
			fmt.Fprintf(&sb, "%s|%s|%s|%t\n",
				ot.Task.ID, ot.EffectiveStatus, ot.Worktree, ot.RemoteOnly)
		}
		return sb.String()
	}

	want := fingerprint()
	if !strings.Contains(want, "local|pending||false") {
		t.Errorf("local-only task missing from overlay:\n%s", want)
	}
	if !strings.Contains(want, "wt00|pending|agent-wt00|true") {
		t.Errorf("sibling-only task missing or mis-attributed:\n%s", want)
	}
	for range 20 { // a data race here would show up intermittently
		if got := fingerprint(); got != want {
			t.Fatalf("overlay not deterministic:\nfirst:\n%s\nlater:\n%s", want, got)
		}
	}
}
