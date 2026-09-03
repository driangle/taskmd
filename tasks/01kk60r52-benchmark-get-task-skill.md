---
id: "01kk60r52"
title: "Benchmark get-task skill"
status: completed
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
completed_at: 2026-09-03
---

# Benchmark get-task skill

## Objective

Benchmark the `get-task` skill with [skival](https://github.com/driangle/skival): run the same
task-lookup requests across four variants (`no-skill`, `plugin-skill`, `lite-skill`,
`bare-project`) and compare them on deterministically-graded correctness plus cost and latency,
so we can tell whether the skill file actually helps an agent retrieve and present a task, or
whether `taskmd init`'s own docs already carry the behavior.

## Prerequisites

- **Follow the `/new-eval` skill** (`.claude/skills/new-eval/SKILL.md`) — it encodes the whole
  procedure: fixture fork, verifier choice, grader style, proving every check can fail, the
  smoke run before any paid run, and the report/suggestions artifacts. This is a **read-only**
  skill, so `evals/list-tasks/` is the reference to copy — correctness is graded from the
  agent's reported output via `check_output` (two-sided, matching IDs rather than phrasing),
  with a `no-mutation` check alongside it.
- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned too —
  otherwise every variant silently loads the installed plugin skill and the suite measures
  nothing)
- `evals/list-tasks/suite.yaml` is the reference implementation to copy the shape from — it
  is the newer of the two and adds what a read-only suite needs (`check_output`,
  `tool_not_used`, `no-mutation`). `evals/add-task/suite.yaml` shows the same `defaults` /
  `ranking` / `&variants` skeleton if you want a second example
- `benchmark/` is deprecated — do not add to it

## Tasks

- [x] Build a skival suite for get-task (`evals/get-task/suite.yaml`), reusing add-task's
      `defaults`, `ranking`, and `&variants` shape
- [x] Set up the fixture workspaces
  - `workspace/` — full `taskmd init` project, same grouped fixture tasks as add-task
    (`001`/`005` in `cli/`, `003` in `web/`, `002`/`004` at the root)
  - `workspace-bare/` — no config, no docs, `taskmd` shadowed off PATH via `.shadow/taskmd`
  - Fork `evals/fixtures/` (shared base) rather than add-task's: `003` now depends on pending
    `002`, giving the blocked-state eval a real target, and `001`/`005` are both auth tasks, so
    "the auth task" is a genuinely ambiguous keyword query without a contrived near-duplicate.
    `evals/fixtures/README.md` records the verified ground truth
- [x] Write deterministic graders as a stdlib-only Go module under `workspace/.verify/`,
      following `evals/add-task/workspace/.verify/{checks.go,assert.go}`
- [x] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [x] Grade correctness and record duration, token usage and cost per sample
- [x] Write the results report (`evals/get-task/REPORT.md`)
- [x] Write improvement suggestions (`evals/get-task/SUGGESTIONS.md`)

### Proposed evals

One behavior per eval — add-task deliberately split `add-bug-template` from
`add-group-routing` because a single eval asserting two things lets one failure mask the other.

- [x] `get-by-id` — "show me the details of task 003"; the reported title, status, priority and
      type must match the fixture file, with no fabricated fields
- [x] `get-by-keyword` — "what's the task about the SSO login bug?"; must resolve the keyword to
      `001` rather than asking the user to supply an ID
- [x] `get-missing` — "show me task 042"; must state plainly that no such task exists (and offer
      the available tasks) rather than inventing one or reporting the nearest match as if it were
      the one asked for
- [x] `get-blocked-state` — "can I start task 006?"; must report the unmet dependency on `002`
      and that the task is not startable (fixture task `003`, blocked by pending `002`)
- [x] `get-json-format` — "give me task 004 as JSON"; output must contain a parseable JSON object
      carrying at least `id`, `title` and `status` with the fixture's values

## Grading notes

`get-task` is read-only: it presents one task and mutates nothing, so add-task's approach —
inspect the created file — does not transfer. Correctness has to be judged from what the agent
*said*.

skival already supports this; no new capability is needed. Verified against
`evals/add-task/suite.yaml` and skival's verifier pipeline:

- `output_contains` — asserts the agent's final output contains each of `values:`. Case-sensitive
  literal substrings; fine for exact fixture values, too brittle for prose.
- `check_output` — runs `run:` with the agent's final text output **piped to the script's stdin**,
  passing on exit 0. This is the read-only analogue of add-task's `check`: the same stdlib-only
  Go grader style, reading stdin instead of the task file. Use
  `run: "cd .verify && GOWORK=off go run . get-by-id"`.

So the graders should read stdin and assert on *facts*, not wording: that the fixture's title,
status, priority and type appear; that a wrong task's identifiers do **not** appear (the cheapest
way to catch "close enough" answers); and for `get-json-format`, extract the fenced/embedded JSON
and `encoding/json`-decode it, asserting field values rather than string-matching.

Two honest caveats to record in the report:

1. **False negatives from phrasing.** Free-text output is not a file. Match case-insensitively
   over keyword sets (the way `titleMentions` does) rather than exact sentences, and hand-test
   both directions before spending tokens — a check that cannot fail measures nothing.
2. **`bare-project` has no CLI.** With `taskmd` shadowed off `PATH`, that variant can only read
   markdown, so graders must never require CLI-shaped output. Verifiers themselves run outside
   the agent's env and can still use the real CLI to derive expected values.

## Acceptance Criteria

- A skival suite exists for get-task and passes `skival validate`
- The forked fixture keeps the dependency edge and the ambiguous auth-task pair intact
- All four variants are executed with per-sample isolation and pinned tool access
  (`allowed_tools` **and** `disallowed_tools`)
- Every eval's grader is verified to fail on wrong output, not just pass on right output
- Per-sample duration, token usage and cost are recorded
- `evals/get-task/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/get-task/SUGGESTIONS.md` is written and grounded in observed failures
