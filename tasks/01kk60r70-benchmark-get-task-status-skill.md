---
id: "01kk60r70"
title: "Benchmark get-task-status skill"
status: pending
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
---

# Benchmark get-task-status skill

## Objective

Benchmark the `get-task-status` skill with skival: run the same status-lookup requests
against four variants (`no-skill`, `plugin-skill`, `lite-skill`, `bare-project`) in an
isolated fixture project, grade each sample deterministically, and compare correctness
alongside cost and latency so we can tell whether the skill file earns its context budget.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins, so `disallowed_tools` must pin
  `Skill`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `ToolSearch` too — otherwise
  every variant silently loads the installed taskmd plugin skill)
- `evals/add-task/` is the reference suite — copy its `suite.yaml` shape, its `&variants`
  anchor, and its `.verify/` Go module layout
- `benchmark/` is deprecated (see the banner in `benchmark/README.md`); do not add to it
- The two skills under test differ in mechanism, which is the point of the comparison:
  `claude-code-plugin/skills/get-task-status/SKILL.md` shells out to `taskmd status
  $ARGUMENTS`, while `claude-code-plugin-lite/skills/get-task-status/SKILL.md` globs and
  reads frontmatter directly with no CLI

## Tasks

- [ ] Build a skival suite for get-task-status (`evals/get-task-status/suite.yaml`),
      mirroring add-task's `defaults`, `ranking`, `isolate: true`, and `&variants` anchor
- [ ] Set up the fixture workspaces, starting from copies of `evals/add-task/workspace/`
      and `evals/add-task/workspace-bare/`
  - `workspace/` — full `taskmd init` project (`.taskmd.yaml`, `.taskmd/`, `tasks/CLAUDE.md`,
    `tasks/TASKMD_SPEC.md`) with tasks grouped across `cli/`, `web/` and the root
  - `workspace-bare/` — no config, no docs, `taskmd` shadowed off PATH via `.shadow/taskmd`
  - **This skill needs more fixture than add-task has.** add-task's five tasks cover
    `pending` / `in-progress` / `completed` and a partially-checked subtask list
    (`cli/001` is 1 of 4), but *no* task carries `owner` or `dependencies` — both of which
    the plugin skill's output format promises. Add at least one task (e.g. `006`) with a
    populated `owner` and a `dependencies` list so those fields can actually be graded,
    and keep the spread of statuses/priorities/groups so a lookup can be wrong in a
    visible way
- [ ] Write deterministic graders as a stdlib-only Go module under
      `evals/get-task-status/workspace/.verify/` (`main.go` dispatching a `checks` map,
      as in add-task), reading the agent's output from **stdin** — see *Grading notes*
- [ ] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [ ] Grade correctness and record duration, token usage and cost per sample
- [ ] Audit one run's conversation logs for tool leaks before trusting the numbers
      (`evals/README.md` § Hermeticity has the one-liner)
- [ ] Write the results report (`evals/get-task-status/REPORT.md`)
- [ ] Write improvement suggestions (`evals/get-task-status/SUGGESTIONS.md`)

### Proposed evals

One behavior per eval — add-task split `add-bug-template` from `add-group-routing`
precisely so a failure in one cannot mask the other.

- `status-by-id` — "what's the status of task 002?" — must report `pending`, `medium`
  priority, and the real title, for the right task.
- `status-by-name` — "is the SSO login fix done yet?" — no ID given; must resolve the
  fuzzy name to `001` and answer `in-progress` rather than guessing "done".
- `status-metadata-fields` — "who owns 006 and what is it waiting on?" — must report the
  owner and the exact dependency IDs from frontmatter, not prose approximations.
- `status-missing-task` — "what's the status of task 042?" — no such task; must say so
  and point at listing the tasks, and must not invent a status for it.
- `status-no-body-dump` — "give me the status of task 003" — must return metadata only;
  this is what separates `get-task-status` from `get-task`, and it is the behavior most
  likely to regress when the skill is absent.

## Acceptance Criteria

- A skival suite exists for get-task-status and passes `skival validate suite.yaml`
- All four variants are executed with per-sample isolation and pinned tool access
  (`allowed_tools` **and** `disallowed_tools`)
- Correctness is graded deterministically, with each grader demonstrated to fail on wrong
  output before any full run is paid for
- Per-sample duration, token usage and cost are recorded
- `evals/get-task-status/REPORT.md` contains a results table, cost/latency comparison,
  analysis and recommendations
- `evals/get-task-status/SUGGESTIONS.md` is written and grounded in observed failures
- `evals/README.md`'s suite table gains a row for this suite

## Grading notes

`get-task-status` is read-only — it mutates nothing, so add-task's approach (run the agent,
then inspect the task file it created) does not transfer. `agent_exits_ok` alone is
worthless here, and a `check` step that only looks at the workspace would pass no matter
what the agent said, because the workspace is identical before and after.

skival does support grading the agent's answer, and this is verified, not assumed:

- `check_output` runs a command with **the agent's text output piped to its stdin**, in the
  eval working directory, exit 0 = pass (`skival` docs `verifiers.md` § `check_output`).
  This is the mechanism to use: the `.verify` module reads stdin, and cross-checks the
  claims against ground truth it derives itself by shelling out to
  `taskmd -d tasks status <id> --format json` (the `--format json` flag exists on
  `status`), exactly as add-task's verifier shells out to `taskmd list --format json`.
  Note `check_output` pipes stdin, so the verifier must read stdin *before* it chdirs and
  runs the CLI.
- `output_contains` asserts case-sensitive substrings in the agent's output. Useful as a
  cheap smoke assertion, too brittle to carry an eval on its own.

So correctness is judged by parsing the transcript against a known fixture and requiring
reported values to match it exactly. Two honest caveats:

1. **Prose is not a data format.** The agent may write "Status: pending", "it's still
   pending", or a markdown table. Graders must normalise (lowercase, collapse whitespace)
   and assert on field/value co-occurrence within a window, not on the plugin skill's exact
   pretty-printed layout — grading the layout would score formatting compliance, not
   correctness, and would unfairly fail `no-skill` and `lite-skill`.
2. **Negative assertions are the weak spot.** `status-missing-task` and
   `status-no-body-dump` are naturally phrased as "the output must *not* contain X", and a
   forbidden-substring check is one-directional: an agent that correctly says "task 042
   does not exist, so it has no status" can trip a naive `pending|completed` scan. Write
   those two graders conservatively and hand-test both directions (a right answer and a
   wrong one) before spending tokens, per `evals/README.md` § Adding a check.

No new skival capability is required. If, in practice, the prose-matching graders prove too
noisy to separate the variants, the fallback is a `judge` step — skival supports it — but
that reintroduces LLM grading, which is the thing the move off `benchmark/` was meant to
eliminate, so treat it as a last resort and say so in the report.
