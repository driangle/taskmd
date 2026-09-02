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

All runs: `claude-sonnet-5`, 8 samples per variant per eval, 160 samples, taskmd 0.5.0.

## Current headline

From run 1, the baseline of the skills as shipped:

- **The `lite-skill` is the strongest at 100%**, but buys it with effort — it re-implements the
  CLI in the agent loop (241 `Read` + 57 `Glob` calls across 40 samples) and costs the same as
  having no skill at all.
- **The retired `benchmark/` harness's "0% delta" verdict was wrong.** 14 of 31 failures never
  looked at the project: each ran one command, `cat ~/.claude/projects/<slug>/memory/MEMORY.md`,
  and asked the user where their tasks live. It is phrasing-dependent — the same `no-skill`
  variant scores 8/8 on "show me all my tasks" and 2/8 on "which of my tasks are still
  pending?". Either skill eliminates it.
- **`plugin-skill` scores 0/8 on "as JSON"** — it fetches correct JSON and re-renders it as a
  markdown table. Excluding that one eval it is 97% against `no-skill`'s 81%, so the 77.5%
  headline mis-ranks it in both directions.

Recommendations and their status live in [SUGGESTIONS.md](SUGGESTIONS.md).
