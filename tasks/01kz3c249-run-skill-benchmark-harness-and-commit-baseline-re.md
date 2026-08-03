---
id: "01kz3c249"
title: "Run skill benchmark harness and commit baseline results"
status: pending
priority: medium
effort: medium
type: chore
phase: skill-benchmarks
dependencies: []
tags: [benchmark, skills, evidence]
created_at: 2026-08-03
---

# Run skill benchmark harness and commit baseline results

## Objective

The `benchmark/` directory ships a thoughtful skill-effectiveness methodology (`evals.json`,
`run_eval.sh`, control-case PATH shadowing, variance caveats) but has **never produced a
committed result**: there are no `iteration-N/` directories, `benchmark/suggestions/` holds
only `.gitkeep`, and `benchmark/CLAUDE.md` repeatedly references an `iteration-1/report.md`
that does not exist. For a project whose central pitch is "designed for AI assistants," there
is currently zero committed evidence that the skills add value. Either produce that evidence or
stop shipping the empty apparatus.

## Tasks

- [ ] Complete the missing runner/aggregation/grading pieces listed in `benchmark/README.md`
      "What's next" (or confirm they already work)
- [ ] Run at least one full iteration and commit `iteration-1/report.md` in the referenced format
- [ ] Fix `benchmark/CLAUDE.md` references so they point at real, committed artifacts
- [ ] If the harness is not ready to run, move `benchmark/` to a branch and remove the
      dangling references from `main`
- [ ] Decide whether benchmark runs should be wired into CI or run manually per release

## Acceptance Criteria

- Either a real benchmark result exists in-repo, or the incomplete harness is off `main`
- No committed doc references benchmark artifacts that do not exist
- There is a documented, working way to reproduce a benchmark iteration
