---
id: "01kk60raz"
title: "Benchmark update-task skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark update-task skill

## Objective

Benchmark the `update-task` skill with [skival](https://github.com/driangle/skival): run the
same realistic edit requests across four variants (`no-skill`, `plugin-skill`, `lite-skill`,
`bare-project`) and compare them on deterministically-graded outcomes — the state of the task
file after the edit — plus cost and latency per sample.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins, so `disallowed_tools` must be pinned too —
  otherwise `Skill` loads the installed taskmd plugin in every variant and the suite measures
  nothing)
- `evals/add-task/` is the reference implementation — `suite.yaml`, the `&variants` anchor, and
  the stdlib-only Go grader in `workspace/.verify/`
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`); do not add to it

## Tasks

- [ ] Build a skival suite for update-task (`evals/update-task/suite.yaml`)
- [ ] Set up the fixture workspaces
  - `workspace/` — full `taskmd init` project. The `add-task` fixture already carries what this
    skill needs: five baseline tasks (`001`–`005`) with mixed statuses, priorities, efforts,
    types and tags, grouped across `tasks/cli/`, `tasks/web/` and the root. Copy it rather
    than inventing new content, so results stay comparable across suites.
  - `workspace-bare/` — same tasks, stripped of `.taskmd.yaml`, `.taskmd/`, `tasks/CLAUDE.md`
    and `tasks/TASKMD_SPEC.md`, with `.shadow/taskmd` first on `PATH`
  - No fixture beyond `add-task`'s is required: every proposed eval mutates an existing
    baseline task, and the baseline values are checked in, so the grader can compare against
    them as constants (`isolate: true` means each sample starts from the pristine copy)
- [ ] Write deterministic graders as a stdlib-only Go module under `evals/update-task/workspace/.verify/`
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, token usage and cost per sample
- [ ] Write the results report (`evals/update-task/REPORT.md`)
- [ ] Write improvement suggestions (`evals/update-task/SUGGESTIONS.md`)

### Proposed evals

One behavior per eval — the `add-task` suite split `add-bug-template` from `add-group-routing`
because a single eval asserting two things lets one failure mask the other.

- [ ] `update-fields` — "set task 004 to high priority and mark it in progress"; asserts
  `priority: high` and `status: in-progress` on task `004`, with its `effort`, `type`, `tags`
  and body left as they were
- [ ] `update-tags` — "add the tag backend to task 003 and drop the urgent tag"; asserts the
  tag array is `security` + `backend` without `urgent`, and that no other field moved
- [ ] `update-dependency` — "task 004 can't be started until task 002 is finished"; asserts
  `dependencies: ["002"]` in frontmatter on `004`, not a sentence added to the body
- [ ] `update-title` — "rename task 002 to 'Add full-text search with filters'"; asserts the
  frontmatter `title` and the `#` heading both change while the `## Objective` text survives
- [ ] `update-missing-task` — "mark task 099 as completed"; there is no `099`, so asserts the
  agent creates nothing and mutates nothing — the five baseline tasks come out byte-identical

## Grading notes

`update-task` mutates existing files, so the `add-task` grading model transfers directly: assert
on the resulting file. The mutation-specific hazards a grader has to catch, each covered by the
evals above:

- **Edit, don't create.** The agent must modify the existing task, not file a new one. Snapshot
  the task ID set before and after and require it unchanged. `newTask()` in
  `evals/add-task/workspace/.verify/taskmd.go` already had to detect the inverse case — a new
  task that reuses a fixture ID shows up as "no new task", which it reports explicitly — so the
  ID-set comparison must be paired with a file count, not just an ID diff.
- **Untouched fields stay untouched.** Anything the request did not mention — other frontmatter
  keys and the whole markdown body — must survive verbatim. Compare against the checked-in
  fixture values.
- **Frontmatter, not prose.** The change has to land as structured frontmatter. A body line
  reading "depends on task 002" is a failure, and is exactly what the `no-skill` variant is
  expected to produce.
- **Still valid.** `taskmd validate` must pass on the whole task set afterwards, catching
  broken YAML from hand-edited frontmatter (the `lite-skill` path edits files directly).

## Acceptance Criteria

- A skival suite exists for update-task and passes `skival validate`
- All four variants are executed with per-sample isolation (`isolate: true`) and pinned tool
  access (`allowed_tools` **and** `disallowed_tools`)
- Per-sample duration, token usage and cost are recorded
- Graders verify the edit landed on the existing task and that unrelated fields and body
  content are unchanged
- `evals/update-task/REPORT.md` contains a results table, cost/latency comparison, analysis and
  recommendations
- `evals/update-task/SUGGESTIONS.md` is written and grounded in observed failures
- `evals/README.md`'s suite table lists the update-task suite
