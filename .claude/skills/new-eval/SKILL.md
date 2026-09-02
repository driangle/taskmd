---
name: new-eval
description: Build a skival eval suite that benchmarks one taskmd skill — fixture, deterministic graders, four variants, smoke run, and a committed report. Use when adding a suite under evals/, benchmarking a skill, or measuring whether a skill actually helps.
metadata:
  internal: true
---

# New eval suite

Build a suite under `evals/<skill>/` that answers one question: **does this skill make an
agent better than a vanilla session in the same project, and is it worth what it costs?**

`$ARGUMENTS` is the skill to benchmark (e.g. `next-task`).

Read [`evals/CLAUDE.md`](../../../evals/CLAUDE.md) and
[`evals/README.md`](../../../evals/README.md) first. Two reference implementations, and you
should copy from whichever matches your skill:

| | Skill kind | Grades | Copy for |
|---|---|---|---|
| `evals/add-task/` | **writes** task files | the filesystem (`check`) | mutating skills |
| `evals/list-tasks/` | **reads** and reports | the agent's output (`check_output`) | read-only skills |

**Runs cost ~$20 and 20 minutes. Ask before starting a full one.**

## Phase 0 — Decide what you are measuring

1. Read the shipped skill in **both** plugins — `claude-code-plugin/skills/<skill>/SKILL.md`
   and `claude-code-plugin-lite/skills/<skill>/SKILL.md`. The suite compares them, so know
   how they differ (the full plugin drives the CLI; lite reads files directly).
2. Classify it: does it **write** or **report**? That decides the grading route and
   everything downstream.
3. Check `tasks/` for an existing task describing the benchmark — it may already specify the
   evals, the fixture, and the expected sets.
4. Sketch 4–6 evals before writing any YAML. **One behavior per eval** — an eval asserting
   two things lets one failure mask the other. Each needs a crisp, deterministic answer; if
   you cannot state the expected result as a set of task IDs, the eval is not gradable.

## Phase 1 — Fixture

Fork the shared base — copy it, do not point `dir:` at it, so suites stay independently
editable:

```bash
cp -R evals/fixtures/workspace       evals/<skill>/workspace
cp -R evals/fixtures/workspace-bare  evals/<skill>/workspace-bare
```

- Bring your **own** fixture only if the skill needs adversarial state (`validate-tasks`
  needs seeded defects, `verify-task` needs `verify:` blocks, `complete-task` needs
  `workflow: pr-review`). Say so in the suite's task file.
- **Never add tasks to `evals/add-task/workspace/`** — it is frozen at five.
- **Verify ground truth from a copy outside this repo.** Inside it, `taskmd list` returns
  taskmd's own 200+ tasks while `taskmd validate` correctly reports six. Record the exact
  queries and expected IDs; `evals/fixtures/README.md` has the current set.

## Phase 2 — Choose verifiers

Every eval gets `agent_exits_ok` plus `tool_not_used` (hermeticity), then:

- **Write skill** → `check`, a Go grader inspecting the created task file.
- **Read-only skill** → `check_output`, which pipes the agent's final text to a grader on
  **stdin**, plus a `no-mutation` `check` asserting the fixture is untouched. That is the
  real failure mode of a read skill: an agent that "helpfully" edits while answering.

Three rules that decide whether the grader measures anything:

- **Assert two-sided.** `output_contains` is presence-only, so an agent that ignores the
  filter and dumps everything passes a "pending only" eval.
- **Match IDs and field names, never phrasing or layout.** Grading the plugin skill's
  pretty-printed table measures formatting compliance and unfairly fails `no-skill`.
- **Let structure separate results from commentary.** Agents end correct filtered answers by
  naming what they excluded. Count only list-item lines when the answer has two or more;
  treat prose as commentary. Do not blocklist cue words — that was tried and lost.

## Phase 3 — Write the grader module

Start from `evals/list-tasks/workspace/.verify/` — stdlib only, its own `go.mod` (kept out
of `go.work`, hence `GOWORK=off`), a `checks` map of name → func, one func per eval.

Then, before spending anything:

```bash
cd evals/<skill>/workspace/.verify
GOWORK=off go test ./...                     # both directions, zero tokens
cp -R . ../../workspace-bare/.verify/        # the two copies must stay identical
```

Write `output_test.go` covering realistic agent answers in **both** directions — a table, a
bullet list, a chatty answer that explains its exclusions, and a dump-everything answer that
must *fail*. **A check that cannot fail measures nothing.** For filesystem checks, hand-test
from a temp copy outside the repo: build the expected state, confirm PASS, break one thing,
confirm FAIL.

## Phase 4 — suite.yaml

Copy the structure from `evals/list-tasks/suite.yaml`:

- `defaults`: runner, model, `samples`, `timeout`, `parallel`, and `runner_config` pinning
  **`allowed_tools` and `disallowed_tools`** — `allowed_tools` does not gate built-ins, so
  without the deny list every variant silently invokes your own installed plugin skill and
  the suite measures nothing.
- `ranking` weights, and the four variants via a `&variants` anchor: `no-skill`,
  `plugin-skill`, `lite-skill`, `bare-project` (which overrides `dir:` and shadows `taskmd`
  on `PATH`).
- `isolate: true` on every eval.
- Verify order: `agent_exits_ok` → `tool_not_used` → `no-mutation` → correctness. The
  pipeline short-circuits, so put the safety checks first.
- Write `run:` as `/bin/sh -c '…'`.

Then `skival validate suite.yaml` — free, and catches structural mistakes.

## Phase 5 — Smoke run

```bash
./evals/run-eval.sh <skill> --samples 1
```

~5% of the cost, and **this is where grader bugs surface** — the `list-tasks` smoke run
caught two (one harness, one grader), and a third turned up later while classifying
failures. Read the actual agent outputs, not just the pass/fail column. Re-grade
failures by hand and ask whether the agent was wrong or the grader was. Fix and re-smoke
until every verdict is one you can defend.

## Phase 6 — Full run

1. **Commit first** — a dirty tree is recorded against a commit it did not measure.
2. Report the smoke result and the estimated cost, and **ask before spending**.
3. `./evals/run-eval.sh <skill>`
4. **Never report a run that did not finish.** Errored samples are recorded `"pass": null`
   and dropped from the math while the rankings still print.

## Phase 7 — Audit, then write it up

- Read every failing output and confirm it is a genuine agent failure. Graders have bugs too.
- Cross-check the tool census against `allowed_tools` for leakage.
- `evals/<skill>/reports/<date>-<commit>.md` — the durable per-run record. Never overwrite a
  past run's report.
- `evals/<skill>/REPORT.md` — cross-run index, one row per run with its commit.
- `evals/<skill>/SUGGESTIONS.md` — grounded in observed failures only. Each item names the
  failure it fixes and how many samples it would flip. No speculative polish.
- Add the suite to the table in `evals/README.md`, and a per-suite `CLAUDE.md` if it has
  traps a future agent would hit.

## Reporting honestly

- **Do not attribute a delta to your change without a mechanism.** Variants with unchanged
  inputs have drifted 17.5 points between runs. A pass-rate swing plus a named cause plus a
  targeted eval moving is evidence; a swing alone is not.
- If you edited a skill and its score moved, **A/B it against the previous version** before
  calling it a regression or a win.
- State grading judgment calls explicitly — where an agent's answer was defensible but
  graded wrong, say so and say how many samples it affects.
