---
id: "01m0j06p2"
title: "Benchmark split-task skill"
status: pending
priority: medium
phase: skill-benchmarks
dependencies: []
tags: ["benchmark", "skill-eval"]
created_at: 2026-08-21
---

# Benchmark split-task skill

## Objective

Benchmark the `split-task` skill with [skival](https://github.com/driangle/skival), comparing
`no-skill`, `plugin-skill`, `lite-skill` and `bare-project` on deterministically-graded outcomes
plus cost and latency.

This suite exists because of a gap found while migrating the other benchmark tasks: `split-task`
is the skill that decomposes one task into sibling sub-tasks with `parent:` set, and it had no
benchmark task at all. `divide-and-conquer` — which *does* have one (`01kk60rr1`) — is a different
thing: it executes an existing task via subagents in per-task git worktrees. The two were
conflated, partly because `claude-code-plugin/skills/split-task/SKILL.md` carried the H1
"Divide and Conquer" (fixed in the same change as this task).

## Prerequisites

- Harness: skival; suites live in `evals/`
- `evals/README.md` — suite construction, the four variants, the verifier table, and the
  hermeticity note (`allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned
  too; `tool_not_used` is the hard backstop once skival is updated)
- `evals/add-task/` is the reference implementation — `suite.yaml` for structure, and
  `workspace/.verify/` for the stdlib-only Go grader style
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`)
- Read both `claude-code-plugin/skills/split-task/SKILL.md` and the lite variant — they differ in
  tool access, so graders must not assume CLI output

## Tasks

- [ ] Build `evals/split-task/suite.yaml`
- [ ] Build the fixture: one large task whose natural decomposition is stable across runs
      (a `large`-effort task with 5+ subtasks spanning clearly distinct concerns), plus a small
      coupled task that should **not** be split — the negative control
- [ ] Write deterministic graders as a stdlib-only Go module under `workspace/.verify/`
- [ ] Run the four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Record correctness, duration, tokens and cost per sample
- [ ] Write `evals/split-task/REPORT.md`
- [ ] Write `evals/split-task/SUGGESTIONS.md`
- [ ] Add a `split-task` row to the suite table in `evals/README.md`

### Proposed evals

One behavior per eval — the `add-task` suite split `add-bug-template` from `add-group-routing`
because a single eval asserting two things lets one failure mask the other.

- [ ] `split-creates-children` — "break task 010 into smaller tasks"; children exist as sibling
      files, each with `parent: "010"`
- [ ] `split-parent-updated` — the original task is amended (per the skill, a `## Sub-tasks`
      section) rather than duplicated or replaced
- [ ] `split-declines-small` — negative control: "split task 004"; a small coupled task should be
      left alone, and an agent that fans it out anyway fails
- [ ] `split-forced` — `--force` skips the complexity check and splits regardless
- [ ] `split-preserves-content` — the parent's original body, tags and acceptance criteria survive

## Grading notes

This skill mutates the project, so file inspection works — but *which* decomposition is correct is
a judgement call, and LLM-grading exactly that kind of thing is what the retired `benchmark/`
harness did badly. Draw the line explicitly:

**Gradable structurally:**

- N children created within a sensible range, not an exact count
- every child carries `parent: "<id>"` — verified against the spec, which defines `parent` as a
  single task ID, at most one per task, with children computed dynamically rather than stored
  (`docs/taskmd_specification.md`); there is no `--parent` flag involved, the skill writes the
  field directly
- each child has a distinct, non-placeholder title and body (reuse `noPlaceholders` from
  `evals/add-task/workspace/.verify/assert.go`)
- the parent still exists, is updated rather than duplicated, and its original content survives
- `taskmd validate` passes — it already flags `parent` self-references and cycles
- no orphan tasks left behind

**Not gradable:** whether the chosen decomposition is the *best* one. Keep it out of the suite
rather than reaching for `judge`.

**Middle ground:** a shape-based check that the children collectively mention a required set of
keywords drawn from the parent's scope. Be honest in the report that this is a coverage proxy,
not a measure of decomposition quality.

**Negative control matters most.** `split-declines-small` is the eval most likely to separate the
variants, because the failure mode in the wild is over-eager splitting. Grade it by asserting the
task count is unchanged and no task gained a `parent` — and hand-test it in both directions before
spending tokens, since a check that cannot fail measures nothing.

## Acceptance Criteria

- A skival suite exists for split-task and passes `skival validate`
- The fixture is checked in and reproducible; each eval sets `isolate: true`
- All variants are executed with per-sample isolation and pinned tool access
  (`allowed_tools` *and* `disallowed_tools`)
- Graders assert `parent` linkage as the spec actually defines it, and were hand-tested in both
  directions before the first paid run
- The negative-control eval fails an agent that splits a task it should have left alone
- Per-sample duration, token usage and cost are recorded
- `evals/split-task/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/split-task/SUGGESTIONS.md` is written and grounded in observed failures
