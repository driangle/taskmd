package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
)

// Cross-worktree behavior of the status views. Shared helpers (overlayTaskMD,
// newSiblingWorktree, stubSiblings) live in worktree_overlay_test.go.

// claimedElsewhereRepo builds the canonical divergence fixture: 001 is pending
// locally but in-progress in sibling agent-b, 002 is pending everywhere, and
// 003 exists only in the sibling (in-progress there).
func claimedElsewhereRepo(t *testing.T) (*taskRepo, []gitmeta.Worktree) {
	t.Helper()
	repo := newTaskRepo(t, map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
		"003-sibling.md": overlayTaskMD("003", "Sibling only", "in-progress"),
	})}
	return repo, siblings
}

// statusRows parses status --format json output.
func statusRows(t *testing.T, out string) []map[string]any {
	t.Helper()
	var rows []map[string]any
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("parse status json: %v\n%s", err, out)
	}
	return rows
}

// rowByID finds the parsed row with the given id.
func rowByID(t *testing.T, rows []map[string]any, id string) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["id"] == id {
			return row
		}
	}
	t.Fatalf("no row with id %q in %v", id, rows)
	return nil
}

func TestStatusCommand_WorktreeOverlay_ListsSiblingClaims(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "status")
	if res.Err != nil {
		t.Fatalf("status failed: %v", res.Err)
	}

	// The whole point: a task in-progress only in a sibling worktree is no
	// longer invisible here.
	if !strings.Contains(res.Stdout, "001") {
		t.Errorf("001 is in-progress in agent-b but missing from status:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "Worktree: agent-b (branch dnc/001)") {
		t.Errorf("status should name the worktree the claim came from:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "Worktrees:") {
		t.Errorf("diverging copies should render the Worktrees section:\n%s", res.Stdout)
	}
	// 002 is pending in every worktree, so it must not appear.
	if strings.Contains(res.Stdout, "Free") {
		t.Errorf("pending task 002 should not be listed:\n%s", res.Stdout)
	}
}

func TestStatusCommand_WorktreeOverlay_JSONProvenance(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "--format", "json")
	if res.Err != nil {
		t.Fatalf("status failed: %v", res.Err)
	}
	rows := statusRows(t, res.Stdout)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2 (001 claimed in sibling, 003 sibling-only)", len(rows))
	}

	claimed := rowByID(t, rows, "001")
	// Structured output follows the house convention: status is the local
	// copy's, effective_status is the merged one — the same shape list and the
	// MCP status tool emit. Text output renders the effective status instead;
	// see status_overlay.go.
	if claimed["status"] != "pending" {
		t.Errorf("status = %v, want the local copy's pending", claimed["status"])
	}
	if claimed["effective_status"] != "in-progress" {
		t.Errorf("effective_status = %v, want in-progress", claimed["effective_status"])
	}
	if claimed["worktree"] != "agent-b" || claimed["branch"] != "dnc/001" {
		t.Errorf("missing provenance fields: %v", claimed)
	}
	if claimed["remote_only"] != nil {
		t.Errorf("001 has a local copy, remote_only should be absent: %v", claimed)
	}
	copies, ok := claimed["worktrees"].([]any)
	if !ok || len(copies) != 2 {
		t.Fatalf("worktrees = %v, want an entry per diverging copy", claimed["worktrees"])
	}

	sibling := rowByID(t, rows, "003")
	if sibling["remote_only"] != true {
		t.Errorf("003 exists only in agent-b, want remote_only true: %v", sibling)
	}
}

func TestStatusCommand_WorktreeOverlay_YAMLProvenance(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("status failed: %v", res.Err)
	}
	var rows []map[string]any
	if err := yaml.Unmarshal([]byte(res.Stdout), &rows); err != nil {
		t.Fatalf("parse status yaml: %v\n%s", err, res.Stdout)
	}
	claimed := rowByID(t, rows, "001")
	if claimed["worktree"] != "agent-b" || claimed["effective_status"] != "in-progress" {
		t.Errorf("yaml row missing overlay fields: %v", claimed)
	}
}

func TestStatusCommand_WorktreeOverlay_SiblingOnlyIDResolves(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "003", "--exact")
	if res.Err != nil {
		t.Fatalf("status 003 failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "Sibling only") {
		t.Errorf("sibling-only 003 should resolve:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "remote-only: no copy in this worktree") {
		t.Errorf("sibling-only task should carry the remote-only warning:\n%s", res.Stdout)
	}
}

func TestStatusCommand_WorktreeOverlay_SingleTaskShowsCopies(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "001", "--exact")
	if res.Err != nil {
		t.Fatalf("status 001 failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "Worktrees:") {
		t.Errorf("diverging copies should render the Worktrees section:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "this worktree: pending") {
		t.Errorf("Worktrees section should disclose the local copy's status:\n%s", res.Stdout)
	}
}

// The statusline answers "what am I working on in *this* checkout", so it must
// not surface a task another agent claimed in a sibling worktree.
func TestStatusCommand_Statusline_StaysLocalWithOverlay(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)
	repo.Write("004-mine.md", overlayTaskMD("004", "Mine", "in-progress"))

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "--statusline")
	if res.Err != nil {
		t.Fatalf("status --statusline failed: %v", res.Err)
	}
	line := strings.TrimSpace(res.Stdout)
	if line != "#004 Mine" {
		t.Errorf("statusline = %q, want only the locally claimed task", line)
	}
}

func TestStatusCommand_Statusline_EmptyWhenOnlySiblingsClaim(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "--statusline")
	if res.Err != nil {
		t.Fatalf("status --statusline failed: %v", res.Err)
	}
	if strings.TrimSpace(res.Stdout) != "" {
		t.Errorf("statusline = %q, want empty: nothing is in-progress in this checkout", res.Stdout)
	}
}

// With the overlay off, output must be byte-identical to today's: no extra
// keys, no extra lines.
func TestStatusCommand_ShapeUnchangedWithoutOverlay(t *testing.T) {
	repo, siblings := claimedElsewhereRepo(t)
	repo.Write("004-mine.md", overlayTaskMD("004", "Mine", "in-progress"))

	res := repo.RunWith(stubSiblings(siblings, nil),
		"status", "--worktree-scope", "isolated", "--format", "json")
	if res.Err != nil {
		t.Fatalf("status failed: %v", res.Err)
	}
	rows := statusRows(t, res.Stdout)
	if len(rows) != 1 || rows[0]["id"] != "004" {
		t.Fatalf("rows = %v, want only the local in-progress 004", rows)
	}
	for _, key := range []string{"effective_status", "worktree", "branch", "remote_only", "worktrees"} {
		if _, present := rows[0][key]; present {
			t.Errorf("overlay key %q leaked into isolated-scope output: %v", key, rows[0])
		}
	}
}

func TestStatusCommand_ScopeFilterAppliesToSiblingTasks(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"cli/001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"cli/001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
		"web/002-other.md":   overlayTaskMD("002", "Other group", "in-progress"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "status", "--scope", "cli", "--format", "json")
	if res.Err != nil {
		t.Fatalf("status failed: %v", res.Err)
	}
	rows := statusRows(t, res.Stdout)
	if len(rows) != 1 || rows[0]["id"] != "001" {
		t.Errorf("rows = %v, want only the cli-group task", rows)
	}
}
