# Working in evals/

Skill benchmarks on [skival](https://github.com/driangle/skival). Read [README.md](README.md)
before changing a suite — this file is only the rules that cost money or produce wrong
numbers when broken.

## Runs cost real money

A full run is ~160 samples, ~$20, ~20 minutes.

- **Always `--samples 1` first.** It is ~5% of the cost and it is where harness and grader
  bugs surface. Every grader in this directory has been wrong in a way only a real agent
  answer exposed.
- **Ask before a full run.** Report the smoke result first.

## Use the wrapper, not `skival run`

```bash
./run-eval.sh <suite> --samples 1     # smoke
./run-eval.sh <suite>                 # full
```

It stamps the commit into the results dir and fails on incomplete runs. **Commit first** —
a run against a dirty tree is recorded against a commit it did not measure.

## Never report a run that did not finish

skival records an errored sample as `"pass": null`, drops it from the pass-rate math, and
still prints confident rankings and confidence intervals. One run lost 35 of 160 samples
and left 3 of 20 cells empty while looking entirely normal. The wrapper now catches this —
if it warns, re-run, do not publish.

## Test graders outside this repo

`taskmd list` inside `*/workspace/` returns taskmd's own 230+ tasks — project resolution
walks up to the repo root, and `-d tasks` does not stop it. `taskmd validate` is unaffected,
so the two disagree and only one is lying. Copy the workspace to a temp dir outside the repo
before testing anything by hand. Real runs are immune (`isolate: true`).

## Conventions

- One behavior per eval. Two assertions in one eval lets one failure mask the other.
- Pin `allowed_tools` **and** `disallowed_tools` — `allowed_tools` does not gate built-ins.
- `results/` is gitignored. The durable record is `<suite>/reports/<date>-<commit>.md`;
  `REPORT.md` is the cross-run index. Never overwrite a past run's report.
- Editing [`fixtures/`](fixtures) means re-verifying the ground truth in its README — from
  a copy outside the repo.
