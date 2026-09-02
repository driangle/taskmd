# list-tasks skill — results

Cross-run index for [`suite.yaml`](suite.yaml). **Each run is measured against a specific
commit** — the skills under test are files in this repo, so a pass rate means nothing without
the revision it was produced from.

Run a suite with the wrapper, which stamps the commit into the results directory:

```bash
evals/run-eval.sh list-tasks --samples 1     # smoke run first, always
evals/run-eval.sh list-tasks                 # full run
```

Then commit the write-up as `reports/<date>-<commit>.md` and add a row below. Raw output under
`results/` is gitignored and regenerable; these reports are the durable record.

## Runs

| Run | Commit | Date | `no-skill` | `plugin-skill` | `lite-skill` | `bare-project` | Cost |
|-----|--------|------|-----------|----------------|--------------|----------------|------|
| [1 — baseline](reports/2026-09-02-565f740.md) | `565f740` | 2026-09-02 | 80% | 77.5% | 100% | 65% | $22.11 |
| [2 — after the fixes](reports/2026-09-02-95e136c.md) | `95e136c` | 2026-09-02 | 62.5% | **100%** | 95% | 57.5% | $21.41 |

All runs: `claude-sonnet-5`, 8 samples per variant per eval, 160 samples, taskmd 0.5.0.

One further run was **discarded, not reported**: 35 of its 160 samples errored with zero
duration, leaving 3 of 20 cells empty while skival still printed a full rankings table.
`../run-eval.sh` now fails on any run containing an errored sample.

## Current headline

- **The format fix works.** `plugin-skill` went 0/8 → **8/8** on the JSON eval and has zero
  failures across all 40 samples in run 2. It is now #1 on correctness *and* cost *and* latency
  at once: 100%, $0.11, 10.9s. In run 1 it needed an "excluding the JSON eval" caveat to look
  good; it needs none now.
- **The retired `benchmark/` harness's "0% delta" verdict was wrong.** The dominant failure in
  both runs is the agent resolving "my tasks" to its own memory — running one command,
  `cat ~/.claude/projects/<slug>/memory/MEMORY.md`, and asking the user where their tasks live
  (14 occurrences in run 1, 17 in run 2). Both skills eliminate it. Finding `taskmd` on `PATH`
  was never the hard part.
- **Read cross-run deltas with care.** `no-skill` and `bare-project` contain no skill file, so
  their inputs were identical across both runs — and they still moved −17.5 and −7.5 points. The
  reported 95% intervals are within-run and do not model that drift. A cross-run difference under
  ~20 points is unresolved without a mechanism; side-by-side variants within one run are fine.

Recommendations and their status live in [SUGGESTIONS.md](SUGGESTIONS.md).
