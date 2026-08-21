---
id: "01kk60rnr"
title: "Benchmark do-task skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark do-task skill

## Objective

Benchmark the `do-task` skill with skival, comparing four variants (`no-skill`, `plugin-skill`,
`lite-skill`, `bare-project`) on deterministically-graded outcomes — did the task actually move
through its lifecycle, did the bookkeeping get done, did the pinned code change land — plus cost
and latency per sample.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built (fixture workspace, `isolate: true`, the four
  variants, the Go verifier convention) and for the hermeticity note before adding variants:
  `allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned too
  (`Skill`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `ToolSearch`)
- `evals/add-task/` is the reference implementation — copy its shape, not its checks
- Read both skill files under test: `claude-code-plugin/skills/do-task/SKILL.md` (CLI-driven) and
  `claude-code-plugin-lite/skills/do-task/SKILL.md` (file-editing, no CLI)
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`) — no new evals there

## Tasks

- [ ] Build a skival suite for do-task (`evals/do-task/suite.yaml`)
  - Raise `timeout` well above `add-task`'s `240` — do-task does real work and runs longer than
    any other skill
- [ ] Set up the fixture workspaces, reusing or extending `evals/add-task/workspace/` and
      `evals/add-task/workspace-bare/`
  - `workspace/` — full `taskmd init` project (`.taskmd.yaml`, `tasks/CLAUDE.md`,
    `tasks/TASKMD_SPEC.md`), tasks grouped across `tasks/cli/`, `tasks/web/` and the root
  - `workspace-bare/` — no config, no docs, `taskmd` shadowed off PATH via `.shadow/taskmd`
  - Add at least one fixture task whose subtasks describe a **small, deterministic** change in
    `src/` (e.g. resolving the existing `// TODO: implement graceful shutdown` in `src/app.go`
    with a named function), so "did the work actually get done" is mechanically checkable
  - No `worklogs` config is needed: `taskmd worklog` writes unconditionally and there is **no
    `worklogs` key in `.taskmd.yaml`**, despite the repo's `CLAUDE.md` documenting one (verified
    against the installed CLI — see the note in `evals/fixtures/README.md`)
- [ ] Write deterministic graders as a stdlib-only Go module under the suite's
      `workspace/.verify/` (`checks.go` + `assert.go` + `taskmd.go`, one entry per check in the
      `checks` map), invoked as `cd .verify && GOWORK=off go run . <check-id>`
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, tokens and cost per sample
- [ ] Audit a run's conversation logs for tool leaks before trusting the numbers
      (see the hermeticity snippet in `evals/README.md`)
- [ ] Write the results report (`evals/do-task/REPORT.md`)
- [ ] Write improvement suggestions (`evals/do-task/SUGGESTIONS.md`)

### Proposed evals (one behavior per eval)

- `do-status-lifecycle` — "pick up task 003 and start working on it": the named task ends in a
  terminal status (`completed`), no other fixture task's status moved, and no new task was created.
- `do-subtask-checkoff` — "work through task 003": every `- [ ]` in that task's `## Tasks` section
  is now `- [x]` in the file on disk.
- `do-code-change` — "do task 006": the pinned change landed — the named function exists in
  `src/app.go`, the resolved TODO comment is gone, and the fixture's `go test ./src/...` passes.
- `do-worklog` — "start task 001 and keep a worklog": `tasks/cli/.worklogs/001.md` exists and
  carries at least one timestamped `##` heading.
- `do-lookup-by-name` — "work on the full-text search task": the agent resolves the name to task
  `002` and moves *that* task, not a keyword-adjacent one — catches lookup failures that the
  ID-based evals hide.

## Grading notes

This is the hardest suite in the set. `do-task` is the broadest skill — it looks a task up, sets it
`in-progress`, does real work, checks off subtasks, writes a worklog, and completes it — so the
gradable and ungradable parts have to be separated deliberately.

**Deterministically gradable (the bookkeeping).** All of this is a file/CLI assertion and belongs in
the suite:

- status transitioned to `in-progress` and ended `completed` (via `taskmd -d tasks list --format json`)
- `- [ ]` subtasks checked off as `- [x]` in the task file
- worklog file created at `tasks/<group>/.worklogs/<ID>.md` (or `tasks/.worklogs/<ID>.md` for root
  tasks) with a timestamped heading. `taskmd worklog` writes unconditionally — there is no
  `worklogs` config key to gate it — so a missing worklog file is a genuine miss, not a
  config-dependent maybe.
- `taskmd validate` still passes over the whole task set
- no stray new task was created, and no unrelated fixture task changed status

**Not deterministically gradable (the actual code change).** Judging arbitrary implementation
quality is exactly what the retired LLM-graded `benchmark/` harness did badly — it produced scores
that moved with the grader's mood rather than the skill. So the fixture must pin the work down to
something mechanically checkable: a named function existing, a specific string present or absent, or
a test committed in the fixture that must pass. Anything looser — "is the implementation good", "is
the worklog entry insightful", "did it pick a sensible approach" — should be **left out of the suite
entirely** rather than graded by vibes. If a behavior can't be expressed as a check that can also
fail, it is a finding for `SUGGESTIONS.md`, not a score.

**Cost and timeout risk.** Every sample here runs a real implementation loop, so this suite is far
more expensive per sample than `add-task` (which is 4 evals x 4 variants x 5 samples at
`timeout: 240`). Start at 1–2 samples with `skival run suite.yaml --samples 1` to shake out the
graders, raise the timeout substantially, and only go to the full sample count once a smoke run
passes end to end. Expect variance to be higher than `add-task`'s, so budget for more samples on the
final run, not fewer.

## Acceptance Criteria

- A skival suite exists for do-task and passes `skival validate`
- All four variants are executed with per-sample isolation and pinned tool access
      (`allowed_tools` **and** `disallowed_tools`)
- Every check is verified to fail as well as pass before a full run is paid for
- Per-sample duration, token usage and cost are recorded
- `evals/do-task/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/do-task/SUGGESTIONS.md` is written and grounded in observed failures
- `evals/README.md`'s suite table lists the new suite
