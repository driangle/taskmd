package main

import "testing"

// Every output grader, exercised in both directions with answers shaped like the
// ones real agents produce. A check that cannot fail measures nothing, so each
// grader has at least one case that must fail and one that must pass — and the
// failing cases are the specific wrong answers this fixture makes reachable
// (the near-match the CLI offers for a missing ID, the second auth task, the
// dependency dump with no verdict, JSON re-rendered as a table).
//
//	cd .verify && GOWORK=off go test ./...
//
// It costs nothing and runs without an agent.

// cliGet003 is the verbatim text `taskmd get 003` prints, which is what the
// plugin-skill variant has in hand when it answers.
const cliGet003 = `Task: 003
Title: Patch XSS vulnerability in comments
Status: pending
Priority: critical
Effort: small
Type: bug
Phase: mvp
Tags: security, urgent
Created: 2026-03-03
File: web/003-critical-security-patch.md

Description:
─────────────────────────────────────────────────
Patch XSS vulnerability in comments

Objective

Fix a reflected XSS vulnerability in the comments section. User input is not properly sanitized before rendering in the DOM.

Tasks

☐ Sanitize all user input in comment rendering
☐ Add CSP headers to prevent inline script execution
☐ Write regression tests for XSS vectors
─────────────────────────────────────────────────

Dependencies:
  Depends on: 002 (Add full-text search)
`

type outputCase struct {
	name string
	out  string
	pass bool
}

func run(t *testing.T, grader func(string) error, cases []outputCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := grader(c.out)
			if c.pass && err != nil {
				t.Fatalf("expected PASS, got: %v", err)
			}
			if !c.pass && err == nil {
				t.Fatalf("expected FAIL, but the grader passed this answer")
			}
		})
	}
}

func TestByID(t *testing.T) {
	run(t, byID, []outputCase{
		{
			name: "verbatim CLI output",
			out:  cliGet003,
			pass: true,
		},
		{
			name: "markdown summary",
			out: `Here are the details for task **003**:

| Field | Value |
|-------|-------|
| ID | 003 |
| Title | Patch XSS vulnerability in comments |
| Status | pending |
| Priority | critical |
| Type | bug |
| Owner | sam |

It depends on 002 (Add full-text search).`,
			pass: true,
		},
		{
			name: "prose answer",
			out: `Task 003 is "Patch XSS vulnerability in comments" — a critical-priority bug
in the web group, currently pending and owned by sam. It's blocked behind 002
(Add full-text search).`,
			pass: true,
		},
		{
			// The "close enough" failure the fixture is built to catch.
			name: "details of the wrong task",
			out: `Task 001: Fix login SSO bug
Status: in-progress
Priority: high
Type: bug`,
			pass: false,
		},
		{
			name: "dumped every task instead of one",
			out: `| ID | Title | Status | Priority | Type |
|----|-------|--------|----------|------|
| 001 | Fix login SSO bug | in-progress | high | bug |
| 002 | Add full-text search | pending | medium | feature |
| 003 | Patch XSS vulnerability in comments | pending | critical | bug |
| 004 | Update README with setup instructions | pending | low | docs |
| 005 | Refactor authentication module | completed | high | improvement |
| 006 | Export reports to CSV | pending | high | feature |`,
			pass: false,
		},
		{
			name: "right task, fields omitted",
			out:  `Task 003 is "Patch XSS vulnerability in comments". It's still pending.`,
			pass: false,
		},
	})
}

func TestByKeyword(t *testing.T) {
	run(t, byKeyword, []outputCase{
		{
			name: "resolved to 001",
			out: `The SSO login bug is task **001 — Fix login SSO bug** (cli group).

- Status: in-progress
- Priority: high
- Owner: alex`,
			pass: true,
		},
		{
			// Regression: 001's body says "users authenticate via SSO". Matching
			// 005 on the token "auth" would fire on *authenticate* and score this
			// correct answer as having leaked the second auth task.
			name: "full body mentioning authentication",
			out: `Task 001 — Fix login SSO bug

Fix the login flow when users authenticate via SSO. Currently, the callback
handler does not properly validate the SAML response, causing intermittent
login failures for enterprise users.`,
			pass: true,
		},
		{
			name: "punted the ambiguity back to the user",
			out: `I found two auth-related tasks:

- 001 — Fix login SSO bug (in-progress)
- 005 — Refactor authentication module (completed)

Which one did you mean?`,
			pass: false,
		},
		{
			name: "asked for an ID instead of resolving",
			out:  `Which task ID would you like? Run ` + "`taskmd list`" + ` to see them all.`,
			pass: false,
		},
	})
}

func TestMissing(t *testing.T) {
	run(t, missing, []outputCase{
		{
			name: "stated plainly and offered the real tasks",
			out: `There is no task 042 in this project. The six tasks that exist are:

- 001 — Fix login SSO bug (in-progress)
- 002 — Add full-text search (pending)
- 003 — Patch XSS vulnerability in comments (pending)
- 004 — Update README with setup instructions (pending)
- 005 — Refactor authentication module (completed)
- 006 — Export reports to CSV (pending)`,
			pass: true,
		},
		{
			// The CLI's own miss output. It offers 002 as a candidate but does not
			// claim it *is* 042, so this is a defensible answer.
			name: "CLI near-match offered as a candidate",
			out: `No exact match found for "042". Did you mean:

  1. 002: Add full-text search (67% match)`,
			pass: true,
		},
		{
			// Verbatim from the smoke run (plugin-skill, 2026-09-03, c017794),
			// where a line-scoped mislabel check called this correct answer a
			// failure: it is one line, so 002's title shared a "line" with 042.
			// The near-match is named in a later sentence, after the answer has
			// already said the task does not exist.
			name: "near-match named in a later sentence",
			out:  `There's no task 042. Available task IDs are 001–006. Closest match was task 002 ("Add full-text search"), but none exactly matches "042" — could you confirm which task you meant?`,
			pass: true,
		},
		{
			// The failure the 67% fuzzy match makes reachable.
			name: "near-match presented as the answer",
			out: `Task 042: Add full-text search
Status: pending
Priority: medium`,
			pass: false,
		},
		{
			name: "answered about the near-match without mentioning 042",
			out: `Here's the task you asked about:

002 — Add full-text search (pending, medium priority, phase mvp)`,
			pass: false,
		},
		{
			name: "mentions 042 but never says it does not exist",
			out: `I looked up 042.
The closest task is 002 (Add full-text search), which is pending.`,
			pass: false,
		},
	})
}

func TestBlockedState(t *testing.T) {
	run(t, blockedState, []outputCase{
		{
			name: "blocked, dependency named",
			out: `Not yet — 003 is blocked. It depends on 002 (Add full-text search),
which is still pending.`,
			pass: true,
		},
		{
			name: "phrased as an ordering constraint",
			out: `You can't start 003 right now: its prerequisite, task 002, has to be
completed first.`,
			pass: true,
		},
		{
			// The discriminating case: the CLI hands the agent "Depends on: 002"
			// for free, so the dependency half alone proves nothing about whether
			// the agent answered the question.
			name: "dependency dump with no verdict",
			out:  cliGet003,
			pass: false,
		},
		{
			name: "said go ahead",
			out:  `Yes — 003 is a critical-priority bug, so it's the right thing to pick up next.`,
			pass: false,
		},
		{
			name: "blocked but dependency not named",
			out:  `That one is blocked at the moment, so I'd pick something else.`,
			pass: false,
		},
	})
}

func TestJSONFormat(t *testing.T) {
	run(t, jsonFormat, []outputCase{
		{
			name: "fenced CLI object",
			out: "Here is task 004 as JSON:\n\n```json\n" + `{
  "id": "004",
  "title": "Update README with setup instructions",
  "status": "pending",
  "priority": "low",
  "effort": "small",
  "type": "docs",
  "phase": "polish",
  "tags": ["docs"],
  "created": "2026-03-04",
  "file_path": "004-update-readme.md",
  "dependencies": {"depends_on": null, "blocks": null}
}` + "\n```",
			pass: true,
		},
		{
			name: "wrapped in a one-element array",
			out:  `[{"id": "004", "title": "Update README with setup instructions", "status": "pending"}]`,
			pass: true,
		},
		{
			// The exact failure list-tasks found in plugin-skill: correct data,
			// re-rendered as a table instead of the requested format.
			name: "re-rendered as a markdown table",
			out: `| Field | Value |
|-------|-------|
| id | 004 |
| title | Update README with setup instructions |
| status | pending |`,
			pass: false,
		},
		{
			name: "dumped every task as JSON",
			out: `[{"id": "001", "title": "Fix login SSO bug", "status": "in-progress"},
{"id": "004", "title": "Update README with setup instructions", "status": "pending"}]`,
			pass: false,
		},
		{
			name: "wrong status",
			out:  `{"id": "004", "title": "Update README with setup instructions", "status": "completed"}`,
			pass: false,
		},
		{
			name: "wrong task",
			out:  `{"id": "002", "title": "Add full-text search", "status": "pending"}`,
			pass: false,
		},
	})
}

// TestTaskObjectsSkipsNestedDependency pins the scan-jump in taskObjects. Without
// it, `taskmd get 003 --format json` — whose dependencies.depends_on carries an
// object with id "002" — would read as two tasks, and any dependent task would
// fail the "only one task" assertion for a correct answer.
func TestTaskObjectsSkipsNestedDependency(t *testing.T) {
	out := `{
  "id": "003",
  "title": "Patch XSS vulnerability in comments",
  "status": "pending",
  "dependencies": {
    "depends_on": [{"id": "002", "title": "Add full-text search"}],
    "blocks": null
  }
}`

	objs, err := taskObjects(out)
	if err != nil {
		t.Fatalf("taskObjects: %v", err)
	}
	if len(objs) != 1 {
		t.Fatalf("got %d task objects, want 1 (the nested dependency must not count)", len(objs))
	}
	if got := field(objs[0], "id"); got != "003" {
		t.Fatalf("got id %q, want 003", got)
	}
}
