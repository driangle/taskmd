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
| _no full run yet_ | — | — | — | — | — | — | — |

## Current headline

The suite is built and its graders are proven in both directions
(`workspace/.verify/output_test.go`), but no full run has been recorded yet. Nothing here
should be cited as a measurement until a row appears above.

One prediction worth writing down *before* measuring, so it can be scored honestly
afterwards: `taskmd get <keyword>` and `taskmd get <missing-id>` both fuzzy-match, open an
interactive selection prompt, and exit 1 with no tty. `--exact` is the non-interactive form
and **neither skill mentions it**, so `plugin-skill` — the variant whose single instruction
is `taskmd get $ARGUMENTS` — is the one most exposed on `get-by-keyword` and `get-missing`.

Recommendations and their status will live in `SUGGESTIONS.md`, written from observed
failures only.
