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

## Writing a new eval

Mechanics are in [README.md](README.md) — "How a suite is put together", "Verifier types",
"Adding a check". These are the judgment calls it does not make for you:

- **Pick the verifier by what the skill does.** Writes files → `check` (grades the
  filesystem). Reports something → `check_output` (grades the agent's final text on stdin).
  A read-only skill also gets a `no-mutation` check, since its real failure mode is an agent
  that "helpfully" edits while answering.
- **Assert two-sided.** `output_contains` is presence-only, so an agent that ignores the
  filter and dumps everything passes a "pending only" eval. Expected values present *and*
  competing ones absent.
- **Write the prompt as a user would ask it**, and grade IDs and field names — never
  phrasing or table layout. Grading the plugin skill's format measures formatting
  compliance and unfairly fails `no-skill`.
- **Prompt vocabulary is a variable, not a detail.** Project words ("the mvp phase") ground
  the agent; generic assistant words ("still pending") get resolved against its own memory.
  Same variant, same workspace, 8/8 versus 2/8. Vary it deliberately, and if a prompt is
  ambiguous the eval measures interpretation rather than the skill.
- **New suite → fork [`fixtures/`](fixtures)**, do not point `dir` at the shared copy.
- **Prove every check can fail before paying for a run.** A check that cannot fail measures
  nothing. Prefer a Go test over a manual pass (see `list-tasks/workspace/.verify/`) — it
  costs no tokens and pins the case for the next person.

## Test graders outside this repo

`taskmd list` inside `*/workspace/` returns taskmd's own 230+ tasks — project resolution
walks up to the repo root, and `-d tasks` does not stop it. `taskmd validate` is unaffected,
so the two disagree and only one is lying. Copy the workspace to a temp dir outside the repo
before testing anything by hand. Real runs are immune (`isolate: true`).

## Conventions

- One behavior per eval. Two assertions in one eval lets one failure mask the other.
- **`allowed_tools` is what makes a suite hermetic.** skival compiles it into the CLI's
  `--tools` flag, an exclusive whitelist over built-ins, so anything unlisted — `Write`,
  `Skill`, `TaskCreate` — is denied at registration. `disallowed_tools` is advisory and
  upstream plans to deprecate it. This README said the reverse until 2026-09-03; see
  [README.md](README.md#hermeticity). Corollary: a `no-mutation` check is vacuous unless
  `Write`/`Edit` are in `allowed_tools`.
- `results/` is gitignored. The durable record is `<suite>/reports/<date>-<commit>.md`;
  `REPORT.md` is the cross-run index. Never overwrite a past run's report.
- Editing [`fixtures/`](fixtures) means re-verifying the ground truth in its README — from
  a copy outside the repo.
