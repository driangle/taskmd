# get-task skill — results

Cross-run index for [`suite.yaml`](suite.yaml). **Each run is measured against a specific
commit** — the skills under test are files in this repo, so a pass rate means nothing without
the revision it was produced from.

Run a suite with the wrapper, which stamps the commit into the results directory:

```bash
evals/run-eval.sh get-task --samples 1     # smoke run first, always
evals/run-eval.sh get-task                 # full run
```

Then commit the write-up as `reports/<date>-<commit>.md` and add a row below. Raw output under
`results/` is gitignored and regenerable; these reports are the durable record.

## Runs

| Run | Commit | Date | `no-skill` | `plugin-skill` | `lite-skill` | `bare-project` | Cost |
|-----|--------|------|-----------|----------------|--------------|----------------|------|
| [1 — baseline](reports/2026-09-03-3081927.md) | `3081927` | 2026-09-03 | 77.5% | **100%** | **100%** | 80% | $20.34 |

`claude-sonnet-5`, 8 samples per variant per eval, 160 samples, 160 completed, taskmd 0.5.0.

Run 1's rates are a **re-grade of its recorded outputs** with three `get-missing` grader bugs
fixed; `check_output` sees only the agent's final text, so this is exact. As run, the numbers
were 72% / 98% / 100% / 80%. Both are in the report.

## Current headline

- **Both skills score 100%, and one eval is the entire story.** `plugin-skill` and
  `lite-skill` are 8/8 on `get-by-keyword` where `no-skill` and `bare-project` are **0/8**.
  Every other eval is 8/8 for all four variants. Remove that one eval and the four variants
  are indistinguishable.
- **The failure is the same one `list-tasks` found**, in a second suite with a different
  skill: the agent answers from `~/.claude/.../memory/MEMORY.md` instead of the project. One
  `bare-project` sample called the workspace "a fresh, non-git working directory with no
  files in it" while six task files sat in `tasks/`. What the skill supplies is not CLI
  knowledge — it is the instruction to look at the project at all.
- **Question phrasing decides the outcome, not difficulty.** *"show me the details of task
  003"* is 8/8 for every variant including the one with no CLI and no docs. *"what's the
  task about the SSO login bug?"* is 0/8 without a skill. Same fixture, same tools.
- **`plugin-skill` wins on cost and latency** — $0.1115 and 12.5s median against `no-skill`'s
  $0.1354 and 17.6s — while tying `lite-skill` on correctness. Unlike in `list-tasks`, it
  does not lose the JSON eval; it scores 8/8 there.
- **The CLI's fuzzy matcher costs a round trip on every miss.** All 8 `plugin-skill`
  `get-missing` samples hit `taskmd get 042`'s interactive selection prompt and its
  `Error: invalid selection`, then recovered via `taskmd list`. `--exact` avoids it and
  neither skill mentions it.

Recommendations and their status live in [SUGGESTIONS.md](SUGGESTIONS.md).
