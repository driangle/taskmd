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
  - Fork `evals/fixtures/` (shared base), which now carries the dependency edge, phases and
    owners the add-task fixture lacked. Read `evals/fixtures/README.md` first — it records the
    verified ground truth for `taskmd next` on that fixture.
  - **`taskmd next` returns a ranked list, not a single answer, and #1 is not the
    highest-priority task.** On the shared fixture it ranks `002` first for being on the critical
    path, ahead of two `high` tasks. So "which task is #1" is *not* a crisp assertion: an agent
    answering `001` or `006` is reasoning defensibly even though the CLI disagrees. Grade the
    robust fact — a blocked task must never be recommended — and report first-place agreement as
    prose, not pass/fail. Extend the fixture only where an eval needs a case it lacks.
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
  - `next-skips-blocked` — **the anchor eval.** "What should I work on next?" On the shared
    fixture `003` is `critical` priority but blocked by pending `002`, and `taskmd next` omits it
    entirely (verified). An agent sorting by priority alone recommends it and fails. This is the
    one assertion here that is genuinely two-sided and stable.
  - `next-obvious` — same question, graded loosely: the answer must be one of the *unblocked*
    pending tasks. Do not assert a specific winner; see the ranked-list note above.
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

1. **The fixture carries the fairness.** A two-sided ID assertion is only fair when the excluded
   set is genuinely wrong — which holds for blocked tasks and not for ranking order. Assert the
   negative set (blocked IDs must not appear); keep the positive side loose.
2. **Substring matching on free prose is brittle.** IDs are matchable; phrasings like "nothing
   is available" are not. Prefer numeric-ID assertions, keep the negative set to IDs, and match
   the "nothing available" case on a small set of alternatives rather than one exact string. If
   that proves too noisy in practice, record it in `SUGGESTIONS.md` rather than loosening the
   check until it can't fail.

## Acceptance Criteria

- A skival suite exists at `evals/next-task/suite.yaml` and passes `skival validate suite.yaml`
- Every eval's negative set (IDs that must not be recommended) is excluded for a stated,
  checkable reason — not merely out-ranked
- Every grader is shown to fail as well as pass, by hand, before a full run is paid for
- All four variants are executed with per-sample isolation and pinned tool access
  (`allowed_tools` **and** `disallowed_tools`), with no tool leaks in the conversation logs
- Per-sample correctness, duration, token usage and cost are recorded
- `evals/next-task/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/next-task/SUGGESTIONS.md` is written and grounded in observed failures
- `evals/README.md`'s suite table lists the new suite
