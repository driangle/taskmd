---
id: "01kk60rr1"
title: "Benchmark divide-and-conquer skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark divide-and-conquer skill

## Objective

Benchmark the `divide-and-conquer` skill with skival: build a suite under `evals/divide-and-conquer/`
that runs the same realistic requests against four variants (`no-skill`, `plugin-skill`, `lite-skill`,
`bare-project`), grades the outcome deterministically from the resulting task files, worklogs and git
branches, and reports correctness alongside cost and latency per sample.

## Prerequisites

- **Follow the `/new-eval` skill** (`.claude/skills/new-eval/SKILL.md`) — it encodes the whole
  procedure: fixture fork, verifier choice, grader style, proving every check can fail, the
  smoke run before any paid run, and the report/suggestions artifacts. This skill **writes**
  task files, so `evals/add-task/` is the reference to copy — correctness is graded from the
  filesystem with a `check` step.
- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned too)
- `evals/add-task/` is the reference implementation — copy its `suite.yaml` shape, its
  `&variants` anchor, and its `workspace/.verify/` grader layout
- This skill drives **subagents in git worktrees**, so the fixture must be a real git repo at
  sample time and the variants must allow the `Agent`/`Task` and `EnterPlanMode` tools that
  both SKILL.md files declare. Confirm how skival seeds a git repo into an isolated workdir
  (a checked-in nested `.git/` is not possible) before writing the fixture
- `benchmark/` is deprecated — see the banner in `benchmark/README.md`; do not add to it

## Tasks

- [ ] Build a skival suite for divide-and-conquer (`evals/divide-and-conquer/suite.yaml`)
  - reuse the `defaults` / `ranking` / `&variants` structure from `evals/add-task/suite.yaml`
  - widen `allowed_tools` to include the agent-spawning tools this skill needs, and keep
    `disallowed_tools` pinned (`Skill`, `TaskCreate`, …) so no variant loads the installed plugin
- [ ] Build the fixture workspaces
  - `workspace/` — a git-initialized `taskmd init` project whose baseline tasks include one
    large, clearly-decomposable task: `touches: [cli, web, docs]`, `effort: large`, and six
    subtasks that map one-to-one onto three non-overlapping directories (`apps/cli/`,
    `apps/web/`, `docs/`). `.taskmd.yaml` defines the `scopes` map for those three paths, so
    the natural partition is dictated by the fixture rather than by the agent's taste and is
    therefore stable across runs
  - a second baseline task whose two obvious workstreams both land in `apps/cli/` — the
    overlap case that must be serialized, not parallelized
  - a third baseline task that is `effort: small` with two coupled subtasks — the negative
    control that should not be fanned out at all
  - `workspace-bare/` — same repo with no `.taskmd.yaml`, no `.taskmd/`, no `tasks/CLAUDE.md`,
    no `TASKMD_SPEC.md`, and a `.shadow/taskmd` stub first on `PATH`
- [ ] Write deterministic graders as a stdlib-only Go module under `workspace/.verify/`
  - follow `evals/add-task/workspace/.verify/{checks.go,assert.go,taskmd.go}`; register each
    check in the `checks` map and invoke it from `suite.yaml` as
    `cd .verify && GOWORK=off go run . <name>`
  - `taskmd.go` needs the N-task generalization of `newTask()` — a helper returning *every*
    task added since the baseline, plus git helpers that shell out to `git branch`,
    `git worktree list`, `git rev-parse` and `git merge-base`
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Record correctness, duration, tokens and cost per sample
- [ ] Audit the run's conversations for tool leaks before trusting the numbers (see the
      hermeticity snippet in `evals/README.md`)
- [ ] Write the results report (`evals/divide-and-conquer/REPORT.md`)
- [ ] Write improvement suggestions (`evals/divide-and-conquer/SUGGESTIONS.md`)
- [ ] Add the suite to the table in `evals/README.md`

### Proposed evals

One behavior per eval — the add-task suite deliberately split `add-bug-template` from
`add-group-routing` because a single eval asserting two things lets one failure mask the other.

- [ ] `dnc-partition` — "pick up task 002 and work on it using parallel subagents". Asserts one
      branch per workstream exists and that the file sets the branches touch are disjoint.
- [ ] `dnc-serialize-overlap` — "work on task 004 with subagents, in parallel where you can".
      Both workstreams live in `apps/cli/`; asserts the second branch's merge-base is the first
      workstream's branch, not the fixture's base branch.
- [ ] `dnc-task-bookkeeping` — "start task 002 using subagents". Asserts the task moved to
      `in-progress`, that completed subtasks are checked off, and that a worklog entry was
      appended at the path the spec defines for that task's group.
- [ ] `dnc-no-autocommit` — "pick up task 002 with parallel subagents and get it done". Asserts
      the base branch HEAD is unmoved and the primary checkout's tree is clean: both SKILL.md
      files require asking the user before integrating, so a merged base branch is a failure.
- [ ] `dnc-single-workstream` — "work on task 005 using subagents where it helps". The negative
      control: task 005 is small and coupled, so asserts no `dnc/` branches or extra worktrees
      were created and the work was simply done in place.

## Grading notes

This skill mutates the project, so file and git inspection works — but *which* partition of the
work is correct is a judgement call, and the retired `benchmark/` harness graded exactly that
kind of thing with an LLM and did it badly. Keep the line explicit:

**Gradable structurally, not semantically**

- The number of workstream branches falls in a sensible range (e.g. 2–4 for the fixture's
  three-scope task) rather than matching an exact count
- Every workstream branch is created off the recorded base branch — or, for a serialized
  workstream, off its predecessor's branch, checked with `git merge-base --is-ancestor`
- The branches' changed file sets are pairwise disjoint for workstreams claimed as parallel
- Each branch carries at least one commit with a non-placeholder message
- Task bookkeeping: status is `in-progress`, subtasks are checked off, a worklog entry exists
  at the spec's path, and the task file is **updated rather than duplicated** — no second task
  file for the same work, no orphan tasks left behind
- The base branch HEAD is unchanged and the primary checkout has no leftover dirty state
- `taskmd validate` still passes over the whole task set

**Not gradable**

Whether the chosen decomposition is the *best* one. Two runs can split the same task three
different sensible ways and all three are fine. Do not put "was this a good breakdown?" in the
suite in any form — no LLM judge, no rubric.

**Shape-based middle ground**

A workable proxy is to require that the workstream branches *collectively* touch a set of
required file areas drawn from the fixture task's `touches` scopes — e.g. all of `apps/cli/`,
`apps/web/` and `docs/` are modified somewhere across the branches. That catches an agent that
silently dropped a third of the task. Be honest in `REPORT.md` that this measures coverage, not
quality: a run can touch all three areas and still have partitioned them badly.

**Verification surface**

skival's verify block supports `agent_exits_ok` and `check` (an arbitrary command), which is
enough — the Go grader shells out to `git` and `taskmd` the same way `add-task`'s does. Do not
reach for features beyond those two.

**Scope note**

`divide-and-conquer` *executes* a task with parallel subagents; it does not create child tasks.
Creating sub-task files — including the `parent: "<id>"` frontmatter link the spec defines for
hierarchical grouping — is the sibling `split-task` skill's behavior. Any eval about N children
linked by `parent` belongs to the `split-task` suite, tracked as task `01m0j06p2`, not this one.
