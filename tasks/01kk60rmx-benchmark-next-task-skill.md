---
id: "01kk60rmx"
title: "Benchmark next-task skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark next-task skill

## Objective

Benchmark the `next-task` skill with skival: run the same "what should I work on next?"
style requests across four variants (`no-skill`, `plugin-skill`, `lite-skill`,
`bare-project`) against a fixture whose dependency graph and priority spread admit exactly
one defensible answer, and compare them on deterministically-graded correctness plus cost
and latency per sample.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, the four variants, per-sample isolation
  (`isolate: true`), and the hermeticity note before adding variants — `allowed_tools` is
  passed through as `--allowedTools` and **does not gate built-ins**, so `disallowed_tools`
  must be pinned too (`Skill`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`,
  `ToolSearch`), or every variant silently runs the installed plugin skill
- `evals/add-task/suite.yaml` is the reference suite (`defaults`, `ranking`, the `&variants`
  anchor); `evals/add-task/workspace/.verify/` is the reference grader style
- Read both skill files under test — `claude-code-plugin/skills/next-task/SKILL.md` (drives
  `taskmd next`) and `claude-code-plugin-lite/skills/next-task/SKILL.md` (globs task files
  and ranks by priority → effort → created date, no CLI)
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`) — don't add evals there

## Tasks

- [ ] Build a skival suite for next-task (`evals/next-task/suite.yaml`), mirroring
      `evals/add-task/suite.yaml`: `runner: claude-code`, pinned `allowed_tools` /
      `disallowed_tools`, `ranking` weights, one eval per behavior
- [ ] Build the fixture workspaces (`evals/next-task/workspace/`, `evals/next-task/workspace-bare/`)
  - Start from the `evals/add-task/` fixtures for shape (`taskmd init` output, `tasks/CLAUDE.md`,
    `TASKMD_SPEC.md`, `tasks/cli/`, `tasks/web/`, root-level tasks, `src/app.go`), and
    `workspace-bare/` stripped the same way — no `.taskmd.yaml`, no `.taskmd/`, no
    `tasks/CLAUDE.md`, no `TASKMD_SPEC.md`, and `.shadow/taskmd` first on `PATH`
  - **This skill needs a purpose-built task set, not the add-task one.** The add-task fixture
    has no `dependencies` and no `phase` spread, so "what's next?" has several defensible
    answers there. Author the dependency graph, priority ladder, statuses and phases so that
    **each eval has exactly one correct answer and every runner-up is wrong for a stated,
    checkable reason** (blocked by an incomplete dependency, wrong status, out of scope).
    Without that, the grader measures taste rather than correctness.
  - Keep the answer stable across variants: the lite skill ranks priority → effort → created,
    the plugin skill defers to `taskmd next`. If those two ranking rules could disagree on the
    fixture, the eval measures the tie-break, not the skill.
- [ ] Write deterministic graders as a stdlib-only Go module under `evals/next-task/workspace/.verify/`
      (`main.go` + `checks.go` + `assert.go`, `GOWORK=off`), registered in the `checks` map and
      referenced from `suite.yaml`
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Record correctness, duration, tokens and cost per sample; audit a run's
      `*.conversation.jsonl` for tool leaks before trusting the numbers
- [ ] Write the results report (`evals/next-task/REPORT.md`)
- [ ] Write improvement suggestions (`evals/next-task/SUGGESTIONS.md`)
- [ ] Proposed evals — one behavior each, so a failure in one can't mask another (the add-task
      suite split `add-bug-template` from `add-group-routing` for exactly this reason):
  - `next-obvious` — "what should I work on next?" with one unambiguously highest-priority,
    unblocked, `pending` task. The floor case: if a variant misses this, nothing else matters.
  - `next-skips-blocked` — same question, but the highest-priority pending task has a
    dependency that is not `completed`. The correct answer is the next-best *unblocked* task;
    recommending the blocked one fails.
  - `next-none-available` — "anything I can pick up?" with every pending task blocked and
    everything else `in-progress` or `completed`. The correct answer is that nothing is
    available; an agent that invents a recommendation fails.
  - `next-scoped-group` — "what should I do next in the cli group?" where a
    higher-priority task exists in another group. Recommending the out-of-group task fails.
  - `next-scoped-phase` — "what's next in the `<phase>` phase?" where the global winner sits
    in a different phase. Split from the group case deliberately: `--phase` and group scoping
    are separate code paths in the CLI and separate instructions in the lite skill.

## Grading notes

`next-task` is read-only — it recommends and mutates nothing, so the add-task approach of
inspecting the created task file does not transfer. Correctness has to be judged from what
the agent *said*.

skival supports this today; no new capability is required. Two verifier types read the agent's
output rather than the workspace:

- `output_contains` with `values: [...]` — every listed substring must appear in the agent's
  final output. Positive-only; there is no "must not contain".
- `check_output` with `run: ...` — runs a command with the agent's output piped to **stdin**,
  exit 0 = pass. This is the one to use, because the interesting assertion is two-sided.

So each eval verifies with `agent_exits_ok` plus a `check_output` step whose grader reads
stdin and asserts both directions:

```yaml
verify:
  - type: agent_exits_ok
  - type: check_output
    run: "cd .verify && GOWORK=off go run . next-skips-blocked"
```

The grader asserts the expected task ID (and/or its title) appears in the transcript, and that
each competing ID does **not** — a one-sided "mentions the right ID" check passes trivially for
an agent that lists every task. For `next-none-available` the assertion inverts: no task ID may
be presented as a recommendation, and some form of "nothing available / all blocked" must be.

Two caveats to design against:

1. **The fixture carries the fairness.** A two-sided ID assertion is only fair if exactly one
   answer is correct. This is the fixture-design constraint above, and it is the main work of
   this task.
2. **Substring matching on free prose is brittle.** IDs are matchable; phrasings like "nothing
   is available" are not. Prefer numeric-ID assertions, keep the negative set to IDs, and match
   the "nothing available" case on a small set of alternatives rather than one exact string. If
   that proves too noisy in practice, record it in `SUGGESTIONS.md` rather than loosening the
   check until it can't fail.

## Acceptance Criteria

- A skival suite exists at `evals/next-task/suite.yaml` and passes `skival validate suite.yaml`
- The fixture admits exactly one correct answer per eval, and each runner-up is excluded for a
  stated, checkable reason
- Every grader is shown to fail as well as pass, by hand, before a full run is paid for
- All four variants are executed with per-sample isolation and pinned tool access
  (`allowed_tools` **and** `disallowed_tools`), with no tool leaks in the conversation logs
- Per-sample correctness, duration, token usage and cost are recorded
- `evals/next-task/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/next-task/SUGGESTIONS.md` is written and grounded in observed failures
- `evals/README.md`'s suite table lists the new suite
