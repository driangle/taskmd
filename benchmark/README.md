# Skill Benchmarks

Measures how much the taskmd Claude Code skills improve task management compared to a vanilla Claude Code session (no skills loaded).

## How it works

Each eval runs the same natural-language prompt **twice** in an isolated project:

1. **without_skill** — `claude -p "<prompt>"` with no taskmd skills
2. **with_skill** — `claude -p "<prompt>" --allowedTools "taskmd:*"` with skills loaded

Both runs are graded against a shared set of assertions, and we record pass rate, token usage, and latency so we can compute deltas.

## Directory structure

```
benchmark/
├── README.md
├── evals.json              # eval definitions (prompts, assertions, setup)
├── fixtures/
│   ├── setup.sh            # creates an isolated project with fixture data
│   ├── tasks/              # 5 baseline task files (001-005)
│   └── src/                # source files with TODO comments
└── iteration-N/            # output from each benchmark run
    └── eval-{id}-{skill}/
        ├── with_skill/
        │   └── outputs/
        │       ├── result.md       # raw model output
        │       └── timing.json     # tokens & duration
        ├── without_skill/
        │   └── outputs/
        │       ├── result.md
        │       └── timing.json
        ├── eval_metadata.json      # assertions copied from evals.json
        └── grading.json            # per-assertion pass/fail verdicts
```

After all evals complete, an aggregation step produces:

- `iteration-N/benchmark.json` — machine-readable comparison deltas
- `iteration-N/benchmark.md` — human-readable summary table

## evals.json

Defines evals for each plugin skill:

| ID | Skill | Prompt (abbreviated) |
|----|-------|---------------------|
| 1 | list-tasks | "show me all my tasks" |
| 2 | get-task | "show me the details of task 001" |
| 3 | get-task-status | "what's the status of task 002?" |
| 4 | add-task | "create a new task to implement user notifications…" |
| 5 | update-task | "change task 002 to high priority…" |
| 6 | complete-task | "mark task 001 as done" |
| 7 | next-task | "what should I work on next?" |
| 8 | do-task | "pick up task 003 and start working on it" |
| 9 | validate-tasks | "check if all my task files are valid" |
| 10 | verify-task | "verify the acceptance criteria for task 003" |
| 11 | split-task | "task 002 is too big, can you split it into smaller pieces?" |
| 12 | import-todos | "find all the TODO comments in the code and turn them into tasks" |
| 13 | divide-and-conquer | "pick up task 002 and work on it using parallel subagents" |
| 18 | safe-queue | "can task 006 safely start alongside the current work?" |

Each eval includes:
- **prompt** — the exact user message sent to Claude
- **assertions** — what the output must demonstrate (e.g. "Runs taskmd list command")
- **setup_extra** (optional) — additional fixture mutations for that eval

## Fixtures

`fixtures/setup.sh <target-dir>` creates a fresh project:

1. Runs `taskmd init`
2. Copies 5 baseline tasks (mixed statuses, priorities, types)
3. Copies `src/app.go` with 3 TODO/FIXME comments

The baseline tasks:

| ID | Title | Status | Priority | Type |
|----|-------|--------|----------|------|
| 001 | Fix login SSO bug | in-progress | high | bug |
| 002 | Add full-text search | pending | medium | feature |
| 003 | Patch XSS vulnerability in comments | pending | critical | bug |
| 004 | Update README with setup instructions | pending | low | docs |
| 005 | Refactor authentication module | completed | high | improvement |

## What's next

- **Runner script** — automate the full eval loop (setup, run both variants, grade, aggregate)
- **Aggregation script** (`aggregate_benchmark.py`) — compute pass-rate deltas, token/latency comparisons, produce `benchmark.json` and `benchmark.md`
- **Grading** — LLM-as-judge or rule-based grading against assertions
- **CI integration** — run benchmarks on skill changes to catch regressions
