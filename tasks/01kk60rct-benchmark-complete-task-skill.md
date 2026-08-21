---
id: "01kk60rct"
title: "Benchmark complete-task skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark complete-task skill

## Objective

Benchmark the `complete-task` skill with skival, comparing four variants (`no-skill`,
`plugin-skill`, `lite-skill`, `bare-project`) on deterministically-graded outcomes plus
cost and latency. Unlike a read-only skill, `complete-task` **mutates** task files, so
grading asserts on the resulting file: the status transition, `completed_at`, how
remaining subtasks were handled, and that nothing else in the task was disturbed.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned
  too — otherwise every variant silently loads the installed taskmd plugin skill)
- `evals/add-task/` is the reference implementation: `suite.yaml`, the `&variants` anchor,
  and the stdlib-only Go grader in `workspace/.verify/`
- Read both skill files under test — `claude-code-plugin/skills/complete-task/SKILL.md`
  (CLI-driven) and `claude-code-plugin-lite/skills/complete-task/SKILL.md` (file-editing) —
  they must both be satisfiable by the same graders
- The completion workflow (solo vs `pr-review`) is documented in `tasks/CLAUDE.md` and
  `docs/taskmd_specification.md`
- `benchmark/` is deprecated — see the banner in `benchmark/README.md`

## Tasks

- [ ] Build a skival suite for complete-task (`evals/complete-task/suite.yaml`)
- [ ] Set up the fixture workspaces, reusing the add-task fixtures as the base
  - [ ] `workspace/` — full `taskmd init` project (`.taskmd.yaml`, `.taskmd/`,
        `tasks/CLAUDE.md`, `TASKMD_SPEC.md`), tasks grouped across `cli/`, `web/` and the root
  - [ ] `workspace-bare/` — no config, no docs, `taskmd` shadowed off PATH via `.shadow/taskmd`
  - [ ] Extend beyond the add-task fixture with the states this skill needs:
        an `in-progress` task whose subtasks are **all** checked (the clean happy path),
        the existing `001` (`in-progress`, three unchecked subtasks) for the
        incomplete-work case, and a new `pending` task with an unmet
        `dependencies: ["002"]` entry
  - [ ] `workspace-pr/` — same project with `workflow: pr-review` in `.taskmd.yaml`
        (the `pr-review` eval needs its own variant list: the shared `*variants` anchor
        pins `bare-project` to `./workspace-bare`)
- [ ] Write deterministic graders as a stdlib-only Go module under `workspace/.verify/`
      (one function per check, registered in the `checks` map, invoked as
      `cd .verify && GOWORK=off go run . <name>`)
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, token usage and cost per sample
- [ ] Write the results report (`evals/complete-task/REPORT.md`)
- [ ] Write improvement suggestions (`evals/complete-task/SUGGESTIONS.md`)

### Proposed evals

One behavior per eval — the add-task suite deliberately split `add-bug-template` from
`add-group-routing` because a single eval asserting two things lets one failure mask the
other.

- [ ] `complete-clean` — "mark task 006 as done", where 006 is `in-progress` with every
      subtask already checked off. The happy path: status transition and file integrity only.
- [ ] `complete-unchecked-subtasks` — "task 001 is finished, close it out", where 001 still
      has three `- [ ]` subtasks. Two outcomes are acceptable (check them off, or refuse and
      report); silently completing with items left unchecked is the failure.
- [ ] `complete-unmet-dependency` — "mark task 007 as complete", where 007 is `pending` with
      an unmet dependency on the still-`pending` 002. Asserts the blocker is not
      cascade-completed and the project still validates.
- [ ] `complete-pr-review` — in `workspace-pr/`: "I opened
      https://github.com/acme/app/pull/42 for task 003, close it out". The correct outcome is
      `in-review` with the PR URL recorded, not `completed`.
- [ ] `complete-no-worklog` — "mark task 004 as done" in a project where `worklogs` is not
      enabled. Asserts the task completes and that no `.worklogs/` file is invented.

## Grading notes

`complete-task` mutates an existing file, so the add-task grading model transfers directly:
assert on the resulting file rather than on the transcript. The mutation-specific hazards
the graders must catch:

- **Status and date together.** `status: completed` must be accompanied by `completed_at`.
  Verify against what the CLI actually writes rather than assuming — `taskmd set <id>
  --status completed` sets `completed_at` to today and clears `cancelled_at`
  (`applyTerminalDateLogic` in `apps/cli/internal/cli/set.go`). `taskmd list --format json`
  exposes both `completed_at` and `pr`, so graders can read them without parsing YAML.
- **Leftover subtasks.** A task marked `completed` while `- [ ]` items remain in its body is
  a failure — the skill is supposed to check them off or refuse.
- **Collateral damage.** Unrelated frontmatter fields (id, title, priority, tags, type,
  dependencies) and unrelated body sections must be byte-identical to the fixture, and the
  task count must not change — some agents "complete" a task by writing a new file.
- **Still valid.** `taskmd validate` must pass over the whole task set afterwards.
- **`pr-review` inverts the expected status.** A grader that only checks for `completed`
  scores the correct `in-review` + `pr:` outcome as a failure. Assert `in-review`, the PR URL
  present in the `pr` array, and `completed_at` **absent**.
- **Scope.** Only the requested task may change; the other fixture tasks must be untouched.
