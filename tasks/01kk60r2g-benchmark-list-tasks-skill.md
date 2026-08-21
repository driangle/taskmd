---
id: "01kk60r2g"
title: "Benchmark list-tasks skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark list-tasks skill

## Objective

Benchmark the `list-tasks` skill with skival: run a set of realistic "show me my tasks" requests
against four variants (`no-skill`, `plugin-skill`, `lite-skill`, `bare-project`) over a fixed
fixture project, grade each sample on deterministically-checked outcomes, and report the
per-variant pass rate alongside cost and latency so the skill's value can be judged against what
it costs to load.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned too —
  otherwise every variant silently invokes the installed taskmd plugin skill and the suite
  measures nothing)
- `evals/add-task/suite.yaml` is the reference implementation to copy structure from
  (`defaults`, `ranking`, the `&variants` anchor)
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`) — don't add evals there

## Tasks

- [ ] Build a skival suite for list-tasks (`evals/list-tasks/suite.yaml`) and get it through
      `skival validate suite.yaml`
- [ ] Set up the fixture workspaces under `evals/list-tasks/`
  - `workspace/` — full `taskmd init` project. The `add-task` fixture already covers everything
    this skill needs to read: five tasks (`001`–`005`) with mixed statuses (`in-progress`,
    `pending`, `completed`), priorities (`critical`/`high`/`medium`/`low`), types and tags, split
    across `tasks/cli/`, `tasks/web/` and the root. Copy it rather than pointing `dir` at
    `../add-task/workspace`, so the two suites stay independently editable — but keep the task
    content identical so results remain comparable.
  - `workspace-bare/` — the stripped variant: no `.taskmd.yaml`, no `.taskmd/`, no
    `tasks/CLAUDE.md`, no `TASKMD_SPEC.md`, and a `.shadow/taskmd` stub first on `PATH`
  - No fixture beyond `add-task`'s is required. One optional extension: none of the fixture
    tasks set `phase` or `owner`, so evals over those fields are not gradable today — add them
    only if a phase/owner eval is wanted.
- [ ] Write deterministic graders as a stdlib-only Go module under `workspace/.verify/`
      (`main.go` + `checks.go` + `assert.go`, run as `cd .verify && GOWORK=off go run . <name>`)
- [ ] Resolve the read-only grading question first — see **Grading notes** below; the prompt
      wording of every eval depends on which answer is chosen
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, token usage and cost per sample
- [ ] Audit the run's conversation logs for tool leakage before trusting the numbers
- [ ] Write the results report (`evals/list-tasks/REPORT.md`)
- [ ] Write improvement suggestions (`evals/list-tasks/SUGGESTIONS.md`)

### Proposed evals

One behavior per eval — the `add-task` suite deliberately split `add-bug-template` from
`add-group-routing` because a single eval asserting two things lets one failure mask the other.

- [ ] `list-all` — "show me all my tasks". Asserts the full set `001`–`005` is reported,
      including the `completed` task `005` and the ones nested under `tasks/cli/` and
      `tasks/web/`. This is the eval that catches an agent that only globs the root directory.
- [ ] `list-status-filter` — "which of my tasks are still pending?". Asserts exactly
      `002`, `003`, `004` — `001` is `in-progress` and `005` is `completed`, so both must be
      excluded. Requires a negative assertion (see Grading notes).
- [ ] `list-scope-filter` — "what's on the plate for the CLI?". Asserts exactly `001` and `005`,
      the two tasks under `tasks/cli/`.
- [ ] `list-json-format` — "list my high priority tasks as JSON". Asserts a parseable JSON array
      containing exactly `001` and `005`, each with at least `id` and `title`. Grades output
      *shape* as well as content, which the table-format evals cannot.

## Grading notes

`list-tasks` is read-only: it prints tasks and mutates nothing. The `add-task` graders work by
inspecting the task file the agent created, and that approach does not transfer — there is no
new file to inspect. Correctness has to be judged from what the agent *reported*.

`evals/README.md` documents only `agent_exits_ok` and `check`, because those are all `add-task`
needed. skival actually registers ten verifier types (`internal/verifier/pipeline.go`); the one
that solves this is **`check_output`** — it runs a shell command with the agent's final text
output piped to **stdin**, exit 0 = pass. It is the direct analogue of `check`: same stdlib-only
Go grader, reading `os.Stdin` instead of the task file. Verified against the installed binary,
which accepts both `check_output` and `output_contains`.

That gives three routes, in preference order:

1. **`check_output` with an exact-set Go grader.** Parse the agent's output for task IDs, compare
   against the set the real CLI returns for the same query, and fail on both missing *and* extra
   IDs. Fully deterministic, expresses exact-set and shape assertions, reuses the existing grader
   style, and keeps the prompt natural ("show me my tasks") — no write step bolted onto a
   read-only skill. Match on stable tokens (IDs, field names), never on sentence phrasing.
2. **`output_contains` for cheap positive assertions.** It asserts substrings are present in the
   agent's output, which covers `list-all` well (all five IDs must appear). It is
   presence-only — there is no `output_not_contains` (confirmed in `internal/verifier/output.go`)
   — so it **cannot** express `list-status-filter` or `list-scope-filter` correctly: an agent that
   dumps all five tasks would pass a "pending tasks" eval because the three expected IDs are all
   present. Use it as a smoke assertion; the two-sided grade belongs in `check_output`.
3. **A no-mutation check.** Independently of correctness, every eval should run a Go check
   asserting the fixture is untouched: `taskmd -d tasks list --format json` still returns exactly
   the five baseline tasks with unchanged fields, and `taskmd validate` passes. That catches the
   real failure mode of a read skill — an agent that "helpfully" edits or reorganizes files — and
   it is the only filesystem assertion worth making here.

`judge` is rejected: it costs tokens per criterion and is non-deterministic, which is exactly what
moving off the `benchmark/` harness was meant to eliminate.

**The honest gap:** grading a read-only skill *on its actual output*, with exact-set semantics, is
not something skival's current verifiers do. Route 1 works around it by changing the task rather
than the harness. Doing it properly would need a new capability — either an
`output_not_contains`/exact-match verifier, or a `check` that receives the agent's final output
(on stdin, or via a documented path). The skival binary does set `SKIVAL_PROMPT`,
`SKIVAL_RUN_DIR` and `SKIVAL_EVENTS_PATH` in the environment and writes
`run-N.conversation.jsonl` per sample, so a check reading the transcript *may* already be
possible — but that is undocumented, unverified, and not mentioned in `evals/README.md`. Do not
build a grader on it without confirming it first; if it works, document it in `evals/README.md`.

## Acceptance Criteria

- A skival suite exists for list-tasks and passes `skival validate`
- All four variants are executed with per-sample isolation (`isolate: true`) and pinned tool
  access (`allowed_tools` **and** `disallowed_tools`)
- Each eval asserts one behavior, and every check is verified to fail as well as pass before a
  full run is paid for
- Per-sample duration, token usage and cost are recorded
- `evals/list-tasks/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/list-tasks/SUGGESTIONS.md` is written and grounded in observed failures
- `evals/README.md`'s suite table lists the new suite

---

## Prior results (deprecated harness)

The following was recorded under the retired `benchmark/` harness (bash `run_eval.sh`, LLM-graded
`with_skill` vs `without_skill` pairs) and is kept for reference only — it is **not** a current
result, and the numbers are not comparable to a skival run.

**0% delta** across all 5 evals. The list-tasks skill provided no measurable improvement because
Claude discovers `taskmd` via the system PATH even without any project context. See
`benchmark/suggestions/list-tasks.md` for the improvement recommendations written at the time.

Additional work done under that harness, beyond its original scope:

- Added 4 list-tasks evals to `evals.json` (IDs 14-17): filter bugs, top 3 as JSON, not-done
  sorted, custom columns
- Ran all 5 evals (10 total runs) with control conditions (baseline had no `CLAUDE.md`,
  `.taskmd.yaml`, or `TASKMD_SPEC.md`)
- Created `benchmark/suggestions/list-tasks.md` with improvement recommendations
- Created `benchmark/CLAUDE.md` with lessons learned for future benchmarks
- Created `benchmark/iteration-1/snapshot.json` capturing the git commit for traceability
