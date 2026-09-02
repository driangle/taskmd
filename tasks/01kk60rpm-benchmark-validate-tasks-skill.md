---
id: "01kk60rpm"
title: "Benchmark validate-tasks skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark validate-tasks skill

## Objective

Benchmark the `validate-tasks` skill with skival, comparing four variants (`no-skill`,
`plugin-skill`, `lite-skill`, `bare-project`) on deterministically-graded outcomes plus cost
and latency, so we can tell whether the skill actually helps an agent find and fix broken task
files — or whether it only restates what `taskmd validate` already does.

## Prerequisites

- **Follow the `/new-eval` skill** (`.claude/skills/new-eval/SKILL.md`) — it encodes the whole
  procedure: fixture fork, verifier choice, grader style, proving every check can fail, the
  smoke run before any paid run, and the report/suggestions artifacts. This is a **read-only**
  skill, so `evals/list-tasks/` is the reference to copy — correctness is graded from the
  agent's reported output via `check_output` (two-sided, matching IDs rather than phrasing),
  with a `no-mutation` check alongside it. Note this suite needs its **own** fixture with
  seeded defects rather than a fork of `evals/fixtures/`, which is valid on purpose.
- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, the four variants, and the hermeticity note:
  `allowed_tools` is passed through as `--allowedTools` and **does not gate built-ins**, so
  `disallowed_tools` must be pinned too (`Skill`, `TaskCreate`, …) or every variant silently
  runs the installed plugin skill
- `evals/list-tasks/suite.yaml` is the reference implementation — `defaults`, `ranking`,
  `evals`, the `&variants` anchor, and the `check_output` / `tool_not_used` steps this suite
  will also want. `evals/add-task/suite.yaml` is a second example of the same skeleton
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`) — do not add evals there

## Tasks

- [ ] Build a skival suite for validate-tasks (`evals/validate-tasks/suite.yaml`)
- [ ] Build a **deliberately broken** fixture workspace — this suite cannot reuse `add-task`'s
      clean `workspace/` as-is, because a validator has nothing to find in a valid project.
      Seed `evals/validate-tasks/workspace/` (and the matching `workspace-bare/`) with:
  - a task missing a required field (`title`)
  - a task with an invalid enum value (e.g. `status: done`)
  - two tasks sharing the same ID (duplicate ID)
  - a circular dependency (two tasks depending on each other)
  - a dangling dependency reference (a `dependencies` entry with no such task)
- [ ] Write deterministic graders as a stdlib-only Go module under `workspace/.verify/`,
      in the style of `evals/add-task/workspace/.verify/` (`checks.go`, `assert.go`, `taskmd.go`)
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, token usage and cost per sample
- [ ] Write the results report (`evals/validate-tasks/REPORT.md`)
- [ ] Write improvement suggestions (`evals/validate-tasks/SUGGESTIONS.md`)

### Proposed evals

One behavior per eval — the add-task suite deliberately split `add-bug-template` from
`add-group-routing` because a single eval asserting two things lets one failure mask the other.

- [ ] `validate-report` — "are there any problems with my tasks?" Read-only; the agent must
      name all five seeded defects and invent none. (See Grading notes: report-only.)
- [ ] `validate-fix-all` — "fix any validation errors in my tasks." Mutating; graded by
      re-inspecting the repaired files.
- [ ] `validate-fix-cycle` — "task X and task Y depend on each other, sort that out." Isolates
      the one defect where the correct repair is ambiguous (which edge to drop).
- [ ] `validate-scoped` — "check the tasks under `tasks/cli` only." Tests that the agent scopes
      validation to the requested directory instead of reporting the whole project.
- [ ] `validate-clean-project` — a workspace with no defects: "check my task files." Guards
      against false positives; the agent must report PASS and change nothing.

## Grading notes

The two eval shapes are not equally gradable, and the suite should say so rather than pretend
otherwise.

**Report-only evals** (`validate-report`, `validate-scoped`, `validate-clean-project`) are
read-only. There is no filesystem delta to assert on, so grading means asserting on the agent's
stdout: that it named every seeded defect and did not hallucinate extras.

`evals/README.md` and `evals/add-task/suite.yaml` only show `agent_exits_ok` and `check`, but
that is the add-task suite's usage, not skival's capability. skival registers ten verifier types
(`internal/verifier/pipeline.go`); the two that matter here are:

- **`check_output`** — runs a shell command with the agent's final text output piped to
  **stdin**, exit 0 = pass. This is the direct analogue of `check`: the same stdlib-only Go
  grader, reading `os.Stdin` instead of the task file.
- **`output_contains`** — asserts every literal in `values` appears in the output.

So report-only evals are gradable today, no new capability needed. Assert **two-sided**: each
seeded defect's task ID is named *and* no ID outside the seeded set is. A positive-only check
passes trivially for an agent that dumps every task. The real hazard is brittleness, not
impossibility — free-text phrasing means matching on stable facts (IDs, field names) rather than
sentences, and hand-testing each grader in the failing direction before spending tokens.

**Fix-it evals** (`validate-fix-all`, `validate-fix-cycle`) mutate files and grade cleanly:

- `taskmd validate` exits 0 afterwards (the `validates()` helper in
  `evals/add-task/workspace/.verify/taskmd.go` already shells out to it)
- every seeded defect is individually repaired — the missing `title` is present and meaningful,
  the enum value is a legal one, the duplicate ID was renumbered, the cycle broken, the dangling
  dependency either pointed at a real task or removed
- **critically**, the agent did not "fix" things by deleting or gutting the offending tasks. A
  grader that only checks `validate` exits 0 passes an agent that `rm`'d every broken file. The
  checks must assert the expected task count is unchanged, that each defective task still exists
  by ID, and that its title/body content survived.

**Known risk to this suite's value:** `taskmd validate` does most of the work here, so the skill
may add little over an agent that simply runs the CLI — the plugin skill is barely more than
"run `taskmd validate`". If the variants tie, that is a legitimate finding to report, not a
reason to skew the evals toward cases the skill happens to cover.

## Acceptance Criteria

- A skival suite exists for validate-tasks and passes `skival validate`
- The fixture workspace contains all five seeded defects and is reproducible from the checked-in
  files (each eval sets `isolate: true`, so samples never mutate the fixture)
- All variants are executed with per-sample isolation and pinned tool access
  (`allowed_tools` *and* `disallowed_tools`)
- Fix-it graders fail an agent that deletes broken tasks rather than repairing them — verified by
  hand in both directions before spending tokens on a run
- Per-sample duration, token usage and cost are recorded
- `evals/validate-tasks/REPORT.md` contains a results table, cost/latency comparison, analysis
  and recommendations
- Report-only graders assert two-sided (every seeded defect named, nothing outside the set) via
  `check_output`, so an agent that lists every task fails
- `evals/validate-tasks/SUGGESTIONS.md` is written and grounded in observed failures
