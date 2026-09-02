package main

import "testing"

// These tests are the "a check that can't fail measures nothing" guardrail from
// evals/README.md, made permanent. Every output grader is exercised in both
// directions against realistic agent answers, at zero token cost:
//
//	cd .verify && GOWORK=off go test ./...
//
// The `no-mutation` check is not covered here because it shells out to taskmd
// against a real project; test it by hand per the "Adding a check" recipe.

const (
	// A plain `taskmd list` table — the format the CLI itself prints.
	tableAll = `Here are all 6 tasks in your project:

id   title                                 status       priority  file
---  ------------------------------------  -----------  --------  --------------------------------
002  Add full-text search                  pending      medium    002-add-search-feature.md
004  Update README with setup instructions pending      low       004-update-readme.md
001  Fix login SSO bug                     in-progress  high      cli/001-fix-login-sso-bug.md
005  Refactor authentication module        completed    high      cli/005-refactor-auth-module.md
003  Patch XSS vulnerability in comments   pending      critical  web/003-critical-security-patch.md
006  Export reports to CSV                 pending      high      web/006-export-reports-csv.md

That's 6 tasks across the cli, web and root groups.`

	// A markdown table, which is what agents tend to write unprompted.
	tablePending = `You have 4 pending tasks:

| ID  | Priority | Title                                 |
|-----|----------|---------------------------------------|
| 003 | critical | Patch XSS vulnerability in comments   |
| 006 | high     | Export reports to CSV                 |
| 002 | medium   | Add full-text search                  |
| 004 | low      | Update README with setup instructions |`

	// The case the grader exists for: a correct answer that also explains what
	// it left out. The IDs of the excluded tasks appear, with their titles, and
	// must not be counted as reported. The explanation wraps across two lines,
	// so a same-line cue filter would miss it.
	chattyPending = `I checked all six tasks and filtered to status ` + "`pending`" + `. Four match:

- **003** — Patch XSS vulnerability in comments (critical, web) — blocked on 002
- **006** — Export reports to CSV (high, web, owner jordan)
- **002** — Add full-text search (medium)
- **004** — Update README with setup instructions (low)

Note that 001 (Fix login SSO bug) is in-progress and 005 (Refactor authentication
module) is completed, so both are excluded from the list above.`

	// The failure list-all exists to catch: only the root task directory globbed,
	// missing everything nested under tasks/cli/ and tasks/web/.
	rootOnly = `I found 2 tasks in the tasks/ directory:

- 002 — Add full-text search (pending, medium)
- 004 — Update README with setup instructions (pending, low)`

	// A block layout, where the ID and the title land on different lines.
	blocksAll = `Your tasks:

### 001
Title: Fix login SSO bug
Status: in-progress

### 002
Title: Add full-text search
Status: pending

### 003
Title: Patch XSS vulnerability in comments
Status: pending

### 004
Title: Update README with setup instructions
Status: pending

### 005
Title: Refactor authentication module
Status: completed

### 006
Title: Export reports to CSV
Status: pending`

	cliGroup = `There are 2 tasks in the ` + "`cli`" + ` group:

| ID  | Status      | Priority | Title                          |
|-----|-------------|----------|--------------------------------|
| 001 | in-progress | high     | Fix login SSO bug              |
| 005 | completed   | high     | Refactor authentication module |`

	// Verbatim shape of a real pilot answer (list-phase-filter, no-skill). The
	// table is correct; the closing sentence names the two excluded tasks *with*
	// their titles and calls them "not MVP" — no exclusion cue word in sight.
	// An earlier grader scored this correct answer as a failure.
	mvpPhaseWithTrailingProse = `The **MVP** phase (due 2026-04-01, "Ship the first usable release") contains 4 tasks:

| ID | Title | Status | Priority | Owner |
|---|---|---|---|---|
| 001 | Fix login SSO bug | in-progress | high | alex |
| 002 | Add full-text search | pending | medium | — |
| 003 | Patch XSS vulnerability in comments | pending | critical | sam (depends on 002) |
| 005 | Refactor authentication module | completed | high | alex |

The other two tasks (006 export CSV, 004 update README) are in the **Polish** phase, not MVP.`

	// Verbatim shape of a real run answer (list-status-filter, bare-project).
	// Three rows label the task by group path, and an earlier anchor that only
	// accepted a bare ID silently dropped them — turning a wrong answer into a
	// differently-wrong answer, and risking the reverse elsewhere.
	groupPrefixedIDs = `Here are your pending tasks, sorted by priority:

| ID | Title | Status | Priority |
|---|---|---|---|
| web/003 | Patch XSS vulnerability in comments | pending | **critical** |
| cli/001 | Fix login SSO bug | **in-progress** | high |
| web/006 | Export reports to CSV | pending | high |
| 002 | Add full-text search | pending | medium |
| 004 | Update README with setup instructions | pending | low |`

	// No list markup at all, so the prose fallback has to carry the match.
	proseOnlyPending = `You currently have four pending tasks: 002 (Add full-text search, medium),
003 (Patch XSS vulnerability in comments, critical), 004 (Update README with setup
instructions, low), and 006 (Export reports to CSV, high).`

	// The trailing sentence names 004 and 006 without their titles, so the
	// title-token requirement must keep them out of the reported set.
	mvpPhase = `The **MVP** phase (due 2026-04-01) has 4 tasks:

- 001 Fix login SSO bug — in-progress, high, alex
- 002 Add full-text search — pending, medium
- 003 Patch XSS vulnerability in comments — pending, critical, sam
- 005 Refactor authentication module — completed, high, alex

The other two tasks (004, 006) are in the polish phase.`
)

func TestReportedIDs_TwoSided(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   []string
	}{
		{"cli table", tableAll, []string{"001", "002", "003", "004", "005", "006"}},
		{"markdown table", tablePending, []string{"002", "003", "004", "006"}},
		{"blocks", blocksAll, []string{"001", "002", "003", "004", "005", "006"}},
		{"root only", rootOnly, []string{"002", "004"}},
		{"cli group", cliGroup, []string{"001", "005"}},

		// The prose hazards, isolated.
		{"excluded tasks are not reported", chattyPending, []string{"002", "003", "004", "006"}},
		{"bare ID mentions are not reported", mvpPhase, []string{"001", "002", "003", "005"}},
		{"named excluded tasks in trailing prose", mvpPhaseWithTrailingProse, []string{"001", "002", "003", "005"}},
		{"prose-only answer still counts", proseOnlyPending, []string{"002", "003", "004", "006"}},
		{"group-prefixed IDs count", groupPrefixedIDs, []string{"001", "002", "003", "004", "006"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := assertReported(tc.output, tc.want...); err != nil {
				t.Errorf("expected exactly %v: %v", tc.want, err)
			}
		})
	}
}

func TestChecks_FailWhenTheyShould(t *testing.T) {
	cases := []struct {
		name   string
		output string
		assert func(string) error
	}{
		{"list-all rejects a root-only glob", rootOnly, listAll},
		{"list-all rejects a filtered list", tablePending, listAll},
		{"status filter rejects a full dump", tableAll, statusFilter},
		{"status filter rejects a short list", rootOnly, statusFilter},
		{"scope filter rejects a full dump", tableAll, scopeFilter},
		{"phase filter rejects a full dump", tableAll, phaseFilter},
		{"empty output is not a pass", "   \n  ", listAll},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.assert(tc.output); err == nil {
				t.Error("expected failure, got pass — this check measures nothing")
			}
		})
	}
}

func TestJSONFormat(t *testing.T) {
	fenced := "```json\n" + `[
  {"id": "001", "title": "Fix login SSO bug", "priority": "high", "tags": ["auth", "bug"]},
  {"id": "005", "title": "Refactor authentication module", "priority": "high"},
  {"id": "006", "title": "Export reports to CSV", "priority": "high"}
]` + "\n```"

	wrapped := `I ran taskmd list --priority high --format json. Result:

{
  "count": 3,
  "tasks": [
    {"id": "006", "title": "Export reports to CSV", "priority": "high"},
    {"id": "001", "title": "Fix login SSO bug", "priority": "high"},
    {"id": "005", "title": "Refactor authentication module", "priority": "high"}
  ]
}`

	allSix := "```json\n" + `[
  {"id": "001", "title": "Fix login SSO bug"},
  {"id": "002", "title": "Add full-text search"},
  {"id": "003", "title": "Patch XSS vulnerability in comments"},
  {"id": "004", "title": "Update README with setup instructions"},
  {"id": "005", "title": "Refactor authentication module"},
  {"id": "006", "title": "Export reports to CSV"}
]` + "\n```"

	noTitle := `[{"id": "001"}, {"id": "005"}, {"id": "006"}]`

	t.Run("passes", func(t *testing.T) {
		for name, out := range map[string]string{"fenced": fenced, "wrapped in an object": wrapped} {
			if err := jsonFormat(out); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		}
	})

	t.Run("fails", func(t *testing.T) {
		for name, out := range map[string]string{
			"unfiltered":      allSix,
			"missing titles":  noTitle,
			"not JSON at all": tableAll,
			"empty array":     "[]",
		} {
			if err := jsonFormat(out); err == nil {
				t.Errorf("%s: expected failure, got pass", name)
			}
		}
	})
}
