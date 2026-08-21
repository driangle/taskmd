# Shared eval fixtures

The base project every suite except `add-task` starts from. `add-task/workspace/` is deliberately
**not** built on this: its graders hardcode `baselineIDs` as `001`–`005`, and its committed
[REPORT.md](../add-task/REPORT.md) baseline was measured against that exact fixture. Adding a task
there would break both. This directory is the place to grow the fixture instead.

- **`workspace/`** — a full `taskmd init` project: `.taskmd.yaml` (with phases), `.taskmd/templates/`,
  `tasks/CLAUDE.md`, `tasks/TASKMD_SPEC.md`, six tasks, and `src/app.go` with TODO comments.
- **`workspace-bare/`** — the same six tasks and `src/`, with no config, no templates, no docs, and
  a `.shadow/taskmd` stub first on `PATH`. Backs the `bare-project` variant.

Keep the two task sets **byte-identical**. They differ only in what surrounds the tasks; if the
tasks themselves drift, `bare-project` stops being a controlled comparison.

## The task set

| ID | Title | Status | Priority | Group | Phase | Owner | Depends on |
|----|-------|--------|----------|-------|-------|-------|------------|
| 001 | Fix login SSO bug | in-progress | high | cli | mvp | alex | — |
| 002 | Add full-text search | pending | medium | *(root)* | mvp | — | — |
| 003 | Patch XSS vulnerability in comments | pending | critical | web | mvp | sam | **002** |
| 004 | Update README with setup instructions | pending | low | *(root)* | polish | — | — |
| 005 | Refactor authentication module | completed | high | cli | mvp | alex | — |
| 006 | Export reports to CSV | pending | high | web | polish | jordan | — |

Every field is load-bearing:

- **Groups** (`cli/`, `web/`, and two tasks at the root) make group routing *discoverable* by
  reading the project. An eval that asserts "route this to `cli`" against a flat project measures
  whether the skill hardcodes taskmd's taxonomy, not whether it read anything.
- **`003` depends on `002`, which is pending**, so `003` is blocked. It is *critical* priority on
  purpose: an agent that sorts by priority alone recommends it, and is wrong.
- **Phases and owners** exist so `--phase` filters and owner-reporting evals have something to
  bite on. The earlier fixture had neither, so those evals were ungradable.
- **`005` is completed** so "not done" filters have something to exclude, and `001` is
  `in-progress` so status filters have three distinct values.
- **`001` and `005` are both auth tasks**, giving keyword lookup a genuinely ambiguous query
  ("the auth task") without needing a contrived near-duplicate.

## Ground truth, verified against the installed CLI

Re-verify after any fixture change — these are the values graders assert on.

```
taskmd next                 → ranked: 002, 001, 006, 004   (003 absent: blocked)
taskmd next --limit 1       → 002
taskmd list --status pending → 002, 003, 004, 006
```

**`next` returns a ranked list, not a single answer, and #1 is not the highest-priority task.**
`002` ranks first for being on the critical path (it unblocks `003`), ahead of the two `high`
tasks. This matters for grading: "which task is #1" is *not* a crisp assertion, because an agent
that answers `001` or `006` is reasoning defensibly even though the CLI disagrees. Grade the
robust fact instead — **`003` must never be recommended** — and treat first-place agreement as a
soft signal reported in prose, not a pass/fail check.

## What is deliberately not here

Suites needing an adversarial or specialised project bring their own fixture and say so in their
task file:

- **`validate-tasks`** needs seeded defects (missing field, bad enum, duplicate ID, dependency
  cycle, dangling reference). This fixture is valid on purpose.
- **`verify-task`** needs tasks carrying `verify:` blocks that pass and fail.
- **`complete-task`** needs a `workflow: pr-review` config variant.
- **`do-task`** needs a task pinned to a mechanically checkable code change.

## Note on worklogs

Worklogs are **opt-in**: with no `worklogs` key in `.taskmd.yaml` — the state every fixture
workspace is in — `taskmd worklog <ID> --add` exits non-zero with "worklogs are disabled for this
project" and writes nothing. Only `worklogs: true` permits a write. Reads (`taskmd worklog <ID>`,
`get`, `feed --source worklog`) work either way.

So a fixture asserting "no worklog file" needs no config at all, and one exercising worklog
writes must add `worklogs: true`. The key validates cleanly in both states.
