---
id: "01kk60rrh"
title: "Benchmark import-todos skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark import-todos skill

## Objective

Benchmark the `import-todos` skill with skival: run the same realistic "turn the code
comments into tasks" requests across four variants (`no-skill`, `plugin-skill`,
`lite-skill`, `bare-project`) against a fixture whose TODO/FIXME corpus is known exactly,
and compare them on deterministically-graded outcomes plus cost and latency per sample.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins — `disallowed_tools` must be pinned
  too, or every variant silently loads the installed plugin skill / files work into
  Claude Code's own tracker)
- `evals/add-task/` is the reference implementation: `suite.yaml` structure (`defaults`,
  `ranking`, the `&variants` anchor) and the stdlib-only Go grader under
  `workspace/.verify/`
- `benchmark/` is deprecated — its fixtures were carried over into
  `evals/add-task/workspace/`, and `src/app.go` there already carries three markers
  (`TODO: add rate limiting to API endpoints`, `FIXME: connection leak on error path`,
  `TODO: implement graceful shutdown`), which this suite builds on
- Know what the CLI actually does before grading it: `taskmd todos list` supports
  `--dir`, `--marker` (repeatable), `--include`, `--exclude` (repeatable, additive with
  `todos.exclude` in `.taskmd.yaml`), `--format table|json|yaml`, `--raw-text`, `--rich`;
  markers are `TODO`, `FIXME`, `HACK`, `XXX`, `NOTE`, `BUG`, `OPTIMIZE`; parsing is
  language-aware (Go, JS, TS, Python, Ruby, Shell, CSS, HTML, Rust, YAML, TOML) and it
  respects `.gitignore` and skips `vendor/`, `node_modules/`, `.git/`. The scanner is
  **capped** by `docs/adr/0001-core-scope-boundary.md` — do not propose new markers or
  languages as part of this benchmark.

## Tasks

- [ ] Build a skival suite for import-todos (`evals/import-todos/suite.yaml`), modelled on
      `evals/add-task/suite.yaml` (same four variants via a `&variants` anchor, same
      `allowed_tools` / `disallowed_tools` pinning, `isolate: true` per eval)
- [ ] Extend the fixture source tree into a **known-answer corpus** under
      `evals/import-todos/workspace/src/` (plus the matching `workspace-bare/`), carrying
      over the three markers already in `evals/add-task/workspace/src/app.go`:
  - [ ] `src/app.go` — the existing 3 markers (2 `TODO`, 1 `FIXME`)
  - [ ] `src/cache.go` — 3 further markers (2 `TODO`, 1 `FIXME`)
  - [ ] `src/db.go` — 1 marker, `// TODO: add connection pooling`, which is **already
        tracked** by a fixture task in `tasks/` (the idempotence trap)
  - [ ] Near-miss 1: a `TODO:` inside a **string literal** in `src/app.go` — a naive grep
        imports it; `taskmd todos list` correctly does not
  - [ ] Near-miss 2: a `TODO` in `vendor/legacy/client.go` — skipped by the scanner's
        vendored-directory rule
  - [ ] Near-miss 3: an additional excluded path covered by `todos.exclude` in the
        fixture's `.taskmd.yaml`, verified with `taskmd todos list --format json`
  - [ ] **Exact expected answer:** `taskmd todos list` reports **7** markers; the correct
        import creates **6** new tasks (the 7th is the already-tracked `src/db.go` one).
        The three near-misses must never appear as tasks.
- [ ] Write deterministic graders as a stdlib-only Go module under
      `evals/import-todos/workspace/.verify/` (`checks.go` + helpers, registered in the
      `checks` map, run as `cd .verify && GOWORK=off go run . <name>`)
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, token usage and cost per sample
- [ ] Audit a run's conversations for tool leaks per the `evals/README.md` snippet
- [ ] Write the results report (`evals/import-todos/REPORT.md`)
- [ ] Write improvement suggestions (`evals/import-todos/SUGGESTIONS.md`)
- [ ] Add the suite to the table in `evals/README.md`

### Proposed evals (one behavior each)

One eval per behavior — the add-task suite deliberately split `add-bug-template` from
`add-group-routing` because a single eval asserting two things lets one failure mask the
other.

- [ ] `todos-import-all` — "turn all the TODO and FIXME comments in the code into tasks".
      Grades the **exact set**: 6 new tasks, each seeded marker represented exactly once,
      no near-miss imported, no hallucinated extras.
- [ ] `todos-idempotent` — "bring the task list up to date with the TODOs in the code, we
      already filed some of these". Grades that the already-tracked `add connection
      pooling` marker gets **no second task** while the other 6 are each present once.
- [ ] `todos-provenance` — "import the code TODOs as tasks, and record where each one came
      from". Grades that every created task names its source file and line.
- [ ] `todos-marker-filter` — "only turn the FIXME comments into tasks, leave the TODOs
      alone". Grades exactly 2 new tasks, both FIXME-sourced, and zero TODO-sourced ones.
- [ ] `todos-source-untouched` — "import the code TODOs into the task list". Grades that
      every file under `src/` is byte-identical to the fixture and `taskmd validate`
      still passes.

## Grading notes

This skill grades unusually well, because it **mutates the project** (creates N tasks from
source comments) and — unlike `divide-and-conquer` — the correct answer is essentially
**known**: the fixture defines exactly which markers exist.

- **Exact-set grading.** Created tasks must correspond one-to-one with the seeded markers.
  Assert both the **count** and that each expected marker's text is represented, so misses
  and hallucinated extras both fail. Baseline task IDs are excluded the way
  `workspace/.verify/taskmd.go` does it, but the helper must return the *set* of new tasks
  rather than `newTask()`'s single-task shape.
- **Provenance.** Each created task should record where it came from (file path / line).
  Check what the CLI and the two skills actually emit before asserting a format — both
  `SKILL.md` files specify `<marker>: <text> (from <file>:<line>)`, and `taskmd todos list
  --format json` returns `file`, `line`, `column`, `tag`, `language`, `text`. Grade for the
  path and line number appearing somewhere in the task, not for one exact rendering.
- **Idempotence.** Re-running must not duplicate already-imported TODOs. This is the most
  valuable eval in the suite and the one most likely to fail: both skills only do a
  case-insensitive substring match against existing titles, and neither is required to act
  on the duplicate it flags.
- **Near-misses.** An agent that greps naively imports the string-literal TODO and the
  vendored one. `taskmd todos list` filters both (verified: language-aware parsing skips
  string literals; `vendor/` is skipped; `todos.exclude` in `.taskmd.yaml` is honored),
  so this is exactly where the CLI-driven `plugin-skill` should separate from the
  Grep-driven `lite-skill` and from `no-skill`.
- **Non-destructive import.** `taskmd validate` must still pass **and** the source files
  must be unmodified. An agent that "resolves" a TODO by deleting the comment has changed
  the code, which is wrong for an import operation — grade it, don't assume it.
- Only use skival features the reference suite actually uses: `agent_exits_ok` and
  `type: check` with a `run:` command. Do not invent verifier types.

## Acceptance Criteria

- A skival suite exists for import-todos and passes `skival validate`
- The fixture corpus is a known-answer set with the expected counts documented in the
  suite, including the three near-miss cases
- All four variants are executed with per-sample isolation and pinned tool access
- Per-sample duration, token usage and cost are recorded
- `evals/import-todos/REPORT.md` contains a results table, cost/latency comparison,
  analysis and recommendations
- `evals/import-todos/SUGGESTIONS.md` is written and grounded in observed failures
