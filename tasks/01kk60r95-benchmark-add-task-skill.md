---
id: "01kk60r95"
title: "Benchmark add-task skill"
status: completed
priority: medium
dependencies: []
tags: ["benchmark", "skill-eval"]
created: 2026-03-08
phase: skill-benchmarks
completed_at: 2026-08-18
---

# Benchmark add-task skill

## Objective

Benchmark the add-task skill by running it with and without the skill loaded, comparing quality, timing, token usage, and cost.

## Prerequisites

- Harness: [skival](https://github.com/driangle/skival); suites live in `evals/`
- See `evals/README.md` for how a suite is built, and the hermeticity note before adding
  variants (`allowed_tools` does not gate built-ins)
- `benchmark/` is deprecated — its fixtures were carried over into `evals/add-task/workspace/`

## Tasks

- [x] Build a skival suite for add-task (`evals/add-task/suite.yaml`)
- [x] Set up the fixture workspaces
  - `workspace/` — full `taskmd init` project (config, templates, CLAUDE.md, TASKMD_SPEC.md),
    tasks grouped across `cli/`, `web/` and the root
  - `workspace-bare/` — no config, no docs, `taskmd` shadowed off PATH via `.shadow/taskmd`
- [x] Write deterministic graders (`workspace/.verify/`, stdlib-only Go module)
- [x] Run all four variants: `no-skill`, `plugin-skill`, `lite-skill`, `bare-project`
- [x] Grade correctness and record timing, tokens and cost per sample
- [x] Write the results report (`evals/add-task/REPORT.md`)
- [x] Write improvement suggestions (`evals/add-task/SUGGESTIONS.md`)

## Acceptance Criteria

- A skival suite exists for add-task and passes `skival validate`
- All variants are executed with per-sample isolation and pinned tool access
- Per-sample duration, token usage and cost are recorded
- The report contains a results table, cost/latency comparison, analysis and recommendations
- Improvement suggestions are written and grounded in observed failures
